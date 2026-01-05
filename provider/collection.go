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
	"strings"
	"time"
)

// Collection represents a Webflow CMS collection.
// This struct matches the Webflow API v2 response format for collections.
type Collection struct {
	ID           string `json:"id,omitempty"`          // Webflow-assigned collection ID (read-only)
	DisplayName  string `json:"displayName"`           // Human-readable name of the collection
	SingularName string `json:"singularName"`          // Singular form of the collection name
	Slug         string `json:"slug,omitempty"`        // URL-friendly slug for the collection
	CreatedOn    string `json:"createdOn,omitempty"`   // Creation timestamp (read-only)
	LastUpdated  string `json:"lastUpdated,omitempty"` // Last update timestamp (read-only)
}

// CollectionListResponse represents the Webflow API response for listing collections.
type CollectionListResponse struct {
	Collections []Collection `json:"collections"` // List of collections
}

// CollectionRequest represents the request body for POST collection.
type CollectionRequest struct {
	DisplayName  string `json:"displayName"`    // Human-readable name
	SingularName string `json:"singularName"`   // Singular form
	Slug         string `json:"slug,omitempty"` // Optional URL slug
}

// ValidateCollectionID validates that a collectionID matches the Webflow collection ID format.
// Collection IDs are 24-character lowercase hexadecimal strings (same format as site IDs).
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateCollectionID(collectionID string) error {
	if collectionID == "" {
		return errors.New("collectionId is required but was not provided; " +
			"please provide a valid Webflow collection ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c3'); " +
			"you can find collection IDs via the Webflow API or dashboard")
	}
	if !siteIDPattern.MatchString(collectionID) {
		return fmt.Errorf("collectionId has invalid format: got '%s', "+
			"expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c3'); "+
			"please ensure the collection ID contains only lowercase letters (a-f) and digits (0-9)", collectionID)
	}
	return nil
}

// ValidateCollectionDisplayName validates that displayName is non-empty and reasonable length.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateCollectionDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("displayName is required but was not provided; " +
			"please provide a name for your collection (e.g., 'Blog Posts', 'Products', 'Team Members'); " +
			"the display name is shown in the Webflow CMS interface")
	}
	if len(displayName) > 255 {
		return fmt.Errorf("displayName is too long: '%s' exceeds maximum length of 255 characters, "+
			"please use a shorter, more concise name for your collection", displayName)
	}
	return nil
}

// ValidateSingularName validates that singularName is non-empty and reasonable length.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateSingularName(singularName string) error {
	if singularName == "" {
		return errors.New("singularName is required but was not provided; " +
			"please provide the singular form of your collection name " +
			"(e.g., 'Blog Post' for 'Blog Posts', 'Product' for 'Products'); " +
			"the singular name is used in the CMS UI when referring to individual items")
	}
	if len(singularName) > 255 {
		return fmt.Errorf("singularName is too long: '%s' exceeds maximum length of 255 characters, "+
			"please use a shorter name", singularName)
	}
	return nil
}

// GenerateCollectionResourceID generates a Pulumi resource ID for a Collection resource.
// Format: {siteID}/collections/{collectionID}
func GenerateCollectionResourceID(siteID, collectionID string) string {
	return fmt.Sprintf("%s/collections/%s", siteID, collectionID)
}

// ExtractIDsFromCollectionResourceID extracts the siteID and collectionID from a Collection resource ID.
// Expected format: {siteID}/collections/{collectionID}
func ExtractIDsFromCollectionResourceID(resourceID string) (siteID, collectionID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "collections" {
		return "", "", fmt.Errorf(
			"invalid resource ID format: expected {siteId}/collections/{collectionId}, got: %s",
			resourceID,
		)
	}

	siteID = parts[0]
	collectionID = strings.Join(parts[2:], "/") // Handle collectionID that might contain slashes

	return siteID, collectionID, nil
}

// getCollectionsBaseURL is used internally for testing to override the API base URL.
var getCollectionsBaseURL = ""

// GetCollections retrieves all collections for a Webflow site.
// It calls GET /v2/sites/{site_id}/collections endpoint.
// Returns the parsed response or an error if the request fails.
func GetCollections(ctx context.Context, client *http.Client, siteID string) (*CollectionListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getCollectionsBaseURL != "" {
		baseURL = getCollectionsBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/collections", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff on retry
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
		_ = resp.Body.Close() // Close immediately after reading
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

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var response CollectionListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getCollectionBaseURL is used internally for testing to override the API base URL.
var getCollectionBaseURL = ""

// GetCollection retrieves a single collection by ID.
// It calls GET /v2/collections/{collection_id} endpoint.
// Returns the collection or an error if the request fails.
func GetCollection(ctx context.Context, client *http.Client, collectionID string) (*Collection, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getCollectionBaseURL != "" {
		baseURL = getCollectionBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s", baseURL, collectionID)

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

		// Handle rate limiting
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var collection Collection
		if err := json.Unmarshal(body, &collection); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &collection, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postCollectionBaseURL is used internally for testing to override the API base URL.
var postCollectionBaseURL = ""

// PostCollection creates a new collection for a Webflow site.
// It calls POST /v2/sites/{site_id}/collections endpoint.
// Returns the created collection or an error if the request fails.
func PostCollection(
	ctx context.Context, client *http.Client,
	siteID, displayName, singularName, slug string,
) (*Collection, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postCollectionBaseURL != "" {
		baseURL = postCollectionBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/collections", baseURL, siteID)

	requestBody := CollectionRequest{
		DisplayName:  displayName,
		SingularName: singularName,
		Slug:         slug, // Optional, empty string OK
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

		// Handle rate limiting
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		// Accept both 200 and 201 as success
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var collection Collection
		if err := json.Unmarshal(body, &collection); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &collection, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteCollectionBaseURL is used internally for testing to override the API base URL.
var deleteCollectionBaseURL = ""

// DeleteCollection removes a collection from a Webflow site.
// It calls DELETE /v2/collections/{collection_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteCollection(ctx context.Context, client *http.Client, collectionID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteCollectionBaseURL != "" {
		baseURL = deleteCollectionBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s", baseURL, collectionID)

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

		// Handle rate limiting
		if resp.StatusCode == 429 {
			retryAfter := resp.Header.Get("Retry-After")
			var waitTime time.Duration
			if retryAfter != "" {
				waitTime = getRetryAfterDuration(retryAfter, time.Duration(1<<uint(attempt))*time.Second)
			} else {
				waitTime = time.Duration(1<<uint(attempt)) * time.Second
			}

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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
