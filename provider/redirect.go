// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"strings"
	"time"
)

// RedirectRule represents a redirect configuration in Webflow.
// This struct matches the Webflow API v2 response format for redirect rules.
type RedirectRule struct {
	ID              string `json:"id,omitempty"` // Webflow-assigned redirect ID
	SourcePath      string `json:"fromUrl"`      // Path to redirect from (e.g., "/old-page")
	DestinationPath string `json:"toUrl"`        // Path to redirect to (e.g., "/new-page")
	StatusCode      int    `json:"statusCode"`   // 301 (permanent) or 302 (temporary)
}

// RedirectResponse represents the Webflow API response for redirects.
type RedirectResponse struct {
	Redirects []RedirectRule `json:"redirects"` // List of redirect rules
}

// RedirectRequest represents the request body for POST/PATCH redirects.
type RedirectRequest struct {
	SourcePath      string `json:"fromUrl,omitempty"`    // Path to redirect from
	DestinationPath string `json:"toUrl,omitempty"`      // Path to redirect to
	StatusCode      int    `json:"statusCode,omitempty"` // 301 or 302
}

// pathPattern is the regex pattern for validating URL paths.
// Valid paths: start with "/" followed by alphanumeric, hyphens, underscores, slashes, dots
var pathPattern = regexp.MustCompile(`^/[a-zA-Z0-9\-_/.]*$`)

// ValidateSourcePath validates that a sourcePath is a valid URL path.
// Webflow redirects expect paths to start with "/" and contain only valid URL characters.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateSourcePath(path string) error {
	if path == "" {
		return errors.New("sourcePath is required but was not provided. " +
			"Please provide a valid URL path starting with '/' (e.g., '/old-page', '/blog/2023'). " +
			"Example valid paths: '/about-us', '/products/item-1', '/news/2024'.")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("sourcePath must start with '/': got '%s'. "+
			"Example valid paths: '/old-page', '/blog/2023', '/products/item-1'. "+
			"Please ensure the path begins with a forward slash.", path)
	}
	if !pathPattern.MatchString(path) {
		return fmt.Errorf("sourcePath contains invalid characters: got '%s'. "+
			"Allowed characters: A-Z, a-z, 0-9, hyphens (-), underscores (_), forward slashes (/), and dots (.). "+
			"Example valid paths: '/old-page', '/blog/2023', '/products/item-1'. "+
			"Please remove any invalid characters.", path)
	}
	return nil
}

// ValidateDestinationPath validates that a destinationPath is a valid URL path.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateDestinationPath(path string) error {
	if path == "" {
		return errors.New("destinationPath is required but was not provided. " +
			"Please provide a valid URL path starting with '/' (e.g., '/new-page', '/home'). " +
			"Example valid paths: '/about-us', '/products/item-1', '/news/2024'.")
	}
	if !strings.HasPrefix(path, "/") {
		return fmt.Errorf("destinationPath must start with '/': got '%s'. "+
			"Example valid paths: '/new-page', '/home', '/products/item-1'. "+
			"Please ensure the path begins with a forward slash.", path)
	}
	if !pathPattern.MatchString(path) {
		return fmt.Errorf("destinationPath contains invalid characters: got '%s'. "+
			"Allowed characters: A-Z, a-z, 0-9, hyphens (-), underscores (_), forward slashes (/), and dots (.). "+
			"Example valid paths: '/new-page', '/home', '/products/item-1'. "+
			"Please remove any invalid characters.", path)
	}
	return nil
}

// ValidateStatusCode validates that a statusCode is either 301 or 302.
// 301 = permanent redirect, 302 = temporary redirect
// Returns actionable error messages explaining redirect types and accepted values.
func ValidateStatusCode(statusCode int) error {
	if statusCode != 301 && statusCode != 302 {
		return fmt.Errorf("statusCode must be either 301 or 302: got %d. "+
			"301 = permanent redirect (use for pages moved permanently). "+
			"302 = temporary redirect (use for temporary page moves or maintenance). "+
			"Example: statusCode=301 for permanent moves, statusCode=302 for temporary redirects.", statusCode)
	}
	return nil
}

// GenerateRedirectResourceID generates a Pulumi resource ID for a Redirect resource.
// Format: {siteID}/redirects/{redirectID}
func GenerateRedirectResourceID(siteID, redirectID string) string {
	return fmt.Sprintf("%s/redirects/%s", siteID, redirectID)
}

// ExtractIDsFromRedirectResourceID extracts the siteID and redirectID from a Redirect resource ID.
// Expected format: {siteID}/redirects/{redirectID}
func ExtractIDsFromRedirectResourceID(resourceID string) (siteID, redirectID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "redirects" {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/redirects/{redirectId}, got: %s", resourceID)
	}

	siteID = parts[0]
	redirectID = strings.Join(parts[2:], "/") // Handle redirectID that might contain slashes

	return siteID, redirectID, nil
}

// getRedirectsBaseURL is used internally for testing to override the API base URL.
var getRedirectsBaseURL = ""

// GetRedirects retrieves all redirects for a Webflow site.
// It calls GET /v2/sites/{site_id}/redirects endpoint.
// Returns the parsed response or an error if the request fails.
func GetRedirects(ctx context.Context, client *http.Client, siteID string) (*RedirectResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getRedirectsBaseURL != "" {
		baseURL = getRedirectsBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/redirects", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Check for Retry-After header from previous response, or use exponential backoff
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

			// Enhanced rate limiting error message with clear delay information
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			// Check for Retry-After header for the next retry
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

		var response RedirectResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postRedirectBaseURL is used internally for testing to override the API base URL.
var postRedirectBaseURL = ""

// PostRedirect creates a new redirect for a Webflow site.
// It calls POST /v2/sites/{site_id}/redirects endpoint.
// Returns the created redirect or an error if the request fails.
func PostRedirect(
	ctx context.Context, client *http.Client,
	siteID, sourcePath, destinationPath string, statusCode int,
) (*RedirectRule, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postRedirectBaseURL != "" {
		baseURL = postRedirectBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/redirects", baseURL, siteID)

	requestBody := RedirectRequest{
		SourcePath:      sourcePath,
		DestinationPath: destinationPath,
		StatusCode:      statusCode,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Check for Retry-After header from previous response, or use exponential backoff
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

			// Enhanced rate limiting error message with clear delay information
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			// Check for Retry-After header for the next retry
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

		var redirect RedirectRule
		if err := json.Unmarshal(body, &redirect); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &redirect, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// patchRedirectBaseURL is used internally for testing to override the API base URL.
var patchRedirectBaseURL = ""

// PatchRedirect updates an existing redirect for a Webflow site.
// It calls PATCH /v2/sites/{site_id}/redirects/{redirect_id} endpoint.
// Returns the updated redirect or an error if the request fails.
func PatchRedirect(
	ctx context.Context, client *http.Client,
	siteID, redirectID, sourcePath, destinationPath string, statusCode int,
) (*RedirectRule, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if patchRedirectBaseURL != "" {
		baseURL = patchRedirectBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/redirects/%s", baseURL, siteID, redirectID)

	// Note: According to Webflow API, PATCH does NOT accept sourcePath (fromUrl)
	// The source path is immutable - if you need to change it, delete and recreate
	requestBody := RedirectRequest{
		DestinationPath: destinationPath,
		StatusCode:      statusCode,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Check for Retry-After header from previous response, or use exponential backoff
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

			// Enhanced rate limiting error message with clear delay information
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			// Check for Retry-After header for the next retry
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

		var redirect RedirectRule
		if err := json.Unmarshal(body, &redirect); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &redirect, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteRedirectBaseURL is used internally for testing to override the API base URL.
var deleteRedirectBaseURL = ""

// DeleteRedirect removes a redirect from a Webflow site.
// It calls DELETE /v2/sites/{site_id}/redirects/{redirect_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteRedirect(ctx context.Context, client *http.Client, siteID, redirectID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteRedirectBaseURL != "" {
		baseURL = deleteRedirectBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/redirects/%s", baseURL, siteID, redirectID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Check for Retry-After header from previous response, or use exponential backoff
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

			// Enhanced rate limiting error message with clear delay information
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			// Check for Retry-After header for the next retry
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
