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
	"time"
)

// CustomScript represents a registered script applied to a site or page.
// This struct matches the Webflow API v2 response format for custom code scripts.
type CustomScript struct {
	// ID is the unique identifier of the registered custom code script.
	ID string `json:"id"`
	// Version is the semantic version string for the registered script (e.g., "0.0.1").
	Version string `json:"version"`
	// Location is where the script is placed - either "header" or "footer".
	Location string `json:"location"`
	// Attributes are developer-specified key/value pairs applied as attributes to the script.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// SiteCustomCodeResponse represents the Webflow API response for site custom code.
// This struct matches the GET /v2/sites/{site_id}/custom_code endpoint response.
type SiteCustomCodeResponse struct {
	// Scripts is a list of scripts applied to the site.
	Scripts []CustomScript `json:"scripts"`
	// LastUpdated is the date when the site's scripts were last updated (read-only).
	LastUpdated string `json:"lastUpdated,omitempty"`
	// CreatedOn is the date when the site's scripts were created (read-only).
	CreatedOn string `json:"createdOn,omitempty"`
}

// SiteCustomCodeRequest represents the request body for PUT /sites/{site_id}/custom_code.
// Used for adding or updating custom code scripts on a site.
type SiteCustomCodeRequest struct {
	// Scripts is a list of scripts to apply to the site.
	Scripts []CustomScript `json:"scripts,omitempty"`
}

// ValidateScriptID validates that a script ID is non-empty.
// Script IDs are required to identify which registered script to apply.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptID(scriptID string) error {
	if scriptID == "" {
		return errors.New("script id is required but was not provided. " +
			"Please provide a valid registered script ID. " +
			"Script IDs are registered using the Register Script endpoint. " +
			"Example: 'cms_slider', 'analytics', etc.")
	}
	return nil
}

// ValidateScriptVersion validates that a script version is a valid semantic version string.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptVersion(version string) error {
	if version == "" {
		return errors.New("version is required but was not provided. " +
			"Please provide a valid semantic version string (e.g., '1.0.0', '0.1.2'). " +
			"The version must match a version of the registered script.")
	}
	return nil
}

// ValidateScriptLocation validates that a script location is either "header" or "footer".
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptLocation(location string) error {
	if location != "header" && location != "footer" {
		return fmt.Errorf("location must be either 'header' or 'footer': got '%s'. "+
			"header = script placed in the <head> section of the page. "+
			"footer = script placed at the end of the <body> section. "+
			"Please choose one of these two locations.", location)
	}
	return nil
}

// GenerateSiteCustomCodeResourceID generates a Pulumi resource ID for a SiteCustomCode resource.
// Format: {siteID}/custom_code
// Note: SiteCustomCode is a 1:1 relationship with a site, so we use a simple suffix.
func GenerateSiteCustomCodeResourceID(siteID string) string {
	return siteID + "/custom_code"
}

// ExtractSiteIDFromSiteCustomCodeResourceID extracts the siteID from a SiteCustomCode resource ID.
// Expected format: {siteID}/custom_code
func ExtractSiteIDFromSiteCustomCodeResourceID(resourceID string) (string, error) {
	if resourceID == "" {
		return "", errors.New("resourceId cannot be empty")
	}

	suffix := "/custom_code"
	if len(resourceID) <= len(suffix) {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/custom_code, got: %s", resourceID)
	}

	if resourceID[len(resourceID)-len(suffix):] != suffix {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/custom_code, got: %s", resourceID)
	}

	siteID := resourceID[:len(resourceID)-len(suffix)]
	if siteID == "" {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/custom_code, got: %s", resourceID)
	}

	return siteID, nil
}

// getSiteCustomCodeBaseURL is used internally for testing to override the API base URL.
var getSiteCustomCodeBaseURL = ""

// GetSiteCustomCode retrieves all custom code scripts applied to a Webflow site.
// It calls GET /v2/sites/{site_id}/custom_code endpoint.
// Returns the parsed response or an error if the request fails.
func GetSiteCustomCode(ctx context.Context, client *http.Client, siteID string) (*SiteCustomCodeResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getSiteCustomCodeBaseURL != "" {
		baseURL = getSiteCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/custom_code", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
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

			// Wait before next retry if we haven't exhausted retries
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

		var response SiteCustomCodeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// putSiteCustomCodeBaseURL is used internally for testing to override the API base URL.
var putSiteCustomCodeBaseURL = ""

// PutSiteCustomCode creates or updates custom code scripts on a Webflow site.
// It calls PUT /v2/sites/{site_id}/custom_code endpoint.
// Returns the updated response or an error if the request fails.
func PutSiteCustomCode(
	ctx context.Context, client *http.Client,
	siteID string, scripts []CustomScript,
) (*SiteCustomCodeResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if putSiteCustomCodeBaseURL != "" {
		baseURL = putSiteCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/custom_code", baseURL, siteID)

	requestBody := SiteCustomCodeRequest{
		Scripts: scripts,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
			backoff := time.Duration(1<<(attempt-1)) * time.Second
			select {
			case <-ctx.Done():
				return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
			case <-time.After(backoff):
			}
		}

		req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(bodyBytes))
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

			// Wait before next retry if we haven't exhausted retries
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

		var response SiteCustomCodeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteSiteCustomCodeBaseURL is used internally for testing to override the API base URL.
var deleteSiteCustomCodeBaseURL = ""

// DeleteSiteCustomCode removes all custom code scripts from a Webflow site.
// It calls DELETE /v2/sites/{site_id}/custom_code endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteSiteCustomCode(ctx context.Context, client *http.Client, siteID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteSiteCustomCodeBaseURL != "" {
		baseURL = deleteSiteCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/custom_code", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff
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

			// Wait before next retry if we haven't exhausted retries
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
