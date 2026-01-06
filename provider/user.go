// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"net/mail"
	"strings"
	"time"
)

// UserAccessGroup represents an access group assigned to a user.
type UserAccessGroup struct {
	Slug string `json:"slug"` // Access group identifier for APIs
	Type string `json:"type"` // "admin" (via API/designer) or "ecommerce" (via purchase)
}

// UserData represents the user's basic info and custom fields.
type UserData struct {
	Name                 string `json:"name"`                            // User's name (no omitempty to allow clearing)
	Email                string `json:"email,omitempty"`                 // User's email address
	AcceptPrivacy        bool   `json:"accept-privacy,omitempty"`        // Privacy policy acceptance
	AcceptCommunications bool   `json:"accept-communications,omitempty"` // Communications acceptance
}

// User represents a Webflow site user.
type User struct {
	ID              string            `json:"id,omitempty"`              // Webflow-assigned user ID
	IsEmailVerified bool              `json:"isEmailVerified,omitempty"` // Email verification status
	LastUpdated     string            `json:"lastUpdated,omitempty"`     // Last update timestamp
	InvitedOn       string            `json:"invitedOn,omitempty"`       // Invitation timestamp
	CreatedOn       string            `json:"createdOn,omitempty"`       // Creation timestamp
	LastLogin       string            `json:"lastLogin,omitempty"`       // Last login timestamp
	Status          string            `json:"status,omitempty"`          // invited, verified, unverified
	AccessGroups    []UserAccessGroup `json:"accessGroups,omitempty"`    // Access groups
	Data            *UserData         `json:"data,omitempty"`            // User data
}

// UserListResponse represents the Webflow API response for listing users.
type UserListResponse struct {
	Users  []User `json:"users"`  // List of users
	Count  int    `json:"count"`  // Number of users returned
	Limit  int    `json:"limit"`  // Limit specified in request
	Offset int    `json:"offset"` // Offset for pagination
	Total  int    `json:"total"`  // Total number of users
}

// InviteUserRequest represents the request body for inviting a new user.
type InviteUserRequest struct {
	Email        string   `json:"email"`                  // Email address to send invite to
	AccessGroups []string `json:"accessGroups,omitempty"` // Access group slugs
}

// UpdateUserRequest represents the request body for updating a user.
type UpdateUserRequest struct {
	AccessGroups []string  `json:"accessGroups,omitempty"` // Access group slugs
	Data         *UserData `json:"data,omitempty"`         // User data (name etc, NOT email/password)
}

// ValidateUserEmail validates that an email address is in a valid format.
// Returns actionable error messages explaining what's wrong and how to fix it.
func ValidateUserEmail(email string) error {
	if email == "" {
		return errors.New("email is required but was not provided. " +
			"Please provide a valid email address (e.g., 'user@example.com'). " +
			"The user will receive an invitation email at this address")
	}

	// Use Go's mail package for email validation
	_, err := mail.ParseAddress(email)
	if err != nil {
		return fmt.Errorf("email address is invalid: '%s'. "+
			"Please provide a valid email address in the format 'user@example.com'. "+
			"The email must contain an @ symbol and a valid domain", email)
	}

	return nil
}

// ValidateUserAccessGroups validates that access group slugs are non-empty strings.
// Returns actionable error messages explaining what's wrong and how to fix it.
func ValidateUserAccessGroups(accessGroups []string) error {
	for i, slug := range accessGroups {
		if slug == "" {
			return fmt.Errorf("accessGroups[%d] is empty. "+
				"Each access group slug must be a non-empty string. "+
				"Example: ['premium-members', 'beta-testers']. "+
				"Access group slugs can be found in the Webflow dashboard under Users > Access Groups", i)
		}
		// Check for common issues
		if strings.Contains(slug, " ") {
			return fmt.Errorf("accessGroups[%d] '%s' contains spaces. "+
				"Access group slugs should not contain spaces. "+
				"Use hyphens or underscores instead (e.g., 'premium-members' instead of 'premium members')", i, slug)
		}
	}
	return nil
}

// GenerateUserResourceID generates a Pulumi resource ID for a User resource.
// Format: {siteID}/users/{userID}
func GenerateUserResourceID(siteID, userID string) string {
	return fmt.Sprintf("%s/users/%s", siteID, userID)
}

// ExtractIDsFromUserResourceID extracts the siteID and userID from a User resource ID.
// Expected format: {siteID}/users/{userID}
func ExtractIDsFromUserResourceID(resourceID string) (siteID, userID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "users" {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/users/{userId}, got: %s", resourceID)
	}

	siteID = parts[0]
	userID = strings.Join(parts[2:], "/") // Handle userID that might contain slashes

	return siteID, userID, nil
}

// getUserBaseURL is used internally for testing to override the API base URL.
var getUserBaseURL = ""

// GetUser retrieves a single user from a Webflow site.
// It calls GET /v2/sites/{site_id}/users/{user_id} endpoint.
// Returns the parsed user or an error if the request fails.
func GetUser(ctx context.Context, client *http.Client, siteID, userID string) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getUserBaseURL != "" {
		baseURL = getUserBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/users/%s", baseURL, siteID, userID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "GET", url, http.NoBody)
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &user, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// inviteUserBaseURL is used internally for testing to override the API base URL.
var inviteUserBaseURL = ""

// InviteUser creates and invites a new user to a Webflow site.
// It calls POST /v2/sites/{site_id}/users/invite endpoint.
// Returns the created user or an error if the request fails.
func InviteUser(ctx context.Context, client *http.Client, siteID, email string, accessGroups []string) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if inviteUserBaseURL != "" {
		baseURL = inviteUserBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/users/invite", baseURL, siteID)

	requestBody := InviteUserRequest{
		Email:        email,
		AccessGroups: accessGroups,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "POST", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle error responses (accept both 200 and 201 as success)
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &user, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// updateUserBaseURL is used internally for testing to override the API base URL.
var updateUserBaseURL = ""

// UpdateUser updates an existing user on a Webflow site.
// It calls PATCH /v2/sites/{site_id}/users/{user_id} endpoint.
// Note: email and password cannot be updated via this endpoint.
// Returns the updated user or an error if the request fails.
func UpdateUser(
	ctx context.Context, client *http.Client,
	siteID, userID string, accessGroups []string, data *UserData,
) (*User, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if updateUserBaseURL != "" {
		baseURL = updateUserBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/users/%s", baseURL, siteID, userID)

	requestBody := UpdateUserRequest{
		AccessGroups: accessGroups,
		Data:         data,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "PATCH", url, bytes.NewReader(bodyBytes))
		if err != nil {
			return nil, fmt.Errorf("failed to create request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var user User
		if err := json.Unmarshal(body, &user); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &user, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteUserBaseURL is used internally for testing to override the API base URL.
var deleteUserBaseURL = ""

// DeleteUser removes a user from a Webflow site.
// It calls DELETE /v2/sites/{site_id}/users/{user_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteUser(ctx context.Context, client *http.Client, siteID, userID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteUserBaseURL != "" {
		baseURL = deleteUserBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/users/%s", baseURL, siteID, userID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "DELETE", url, http.NoBody)
		if err != nil {
			return fmt.Errorf("failed to create request: %w", err)
		}

		resp, err := client.Do(req)
		if err != nil {
			lastErr = handleNetworkError(err)
			continue
		}

		body, err := io.ReadAll(resp.Body)
		_ = resp.Body.Close()
		if err != nil {
			lastErr = fmt.Errorf("failed to read response body: %w", err)
			continue
		}

		// Handle rate limiting with retry
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// 204 No Content is success
		// 404 Not Found is also success (idempotent delete)
		if resp.StatusCode == 204 || resp.StatusCode == 404 {
			return nil
		}

		// Handle other error responses
		return handleWebflowError(resp.StatusCode, body)
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
