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

// CustomCodeScript represents a script applied to a page.
// This struct matches the Webflow API v2 response format for custom code scripts.
type CustomCodeScript struct {
	// ID is the unique identifier for the registered custom code script.
	ID string `json:"id"`
	// Version is the semantic version string for the registered script (e.g., "1.0.0").
	Version string `json:"version"`
	// Location is where the script should be applied: "header" or "footer".
	Location string `json:"location"`
	// Attributes is a map of developer-specified key/value pairs for script attributes.
	Attributes map[string]interface{} `json:"attributes,omitempty"`
}

// PageCustomCodeResponse represents the Webflow API response for GET /pages/{page_id}/custom_code.
// It contains all scripts applied to a page by the App.
type PageCustomCodeResponse struct {
	// Scripts is the list of custom code scripts applied to the page.
	Scripts []CustomCodeScript `json:"scripts"`
	// LastUpdated is the date when the page's scripts were last updated (read-only).
	LastUpdated string `json:"lastUpdated,omitempty"`
	// CreatedOn is the date when the page's scripts were created (read-only).
	CreatedOn string `json:"createdOn,omitempty"`
}

// PageCustomCodeRequest represents the request body for PUT /pages/{page_id}/custom_code.
// Used for applying or updating custom code scripts on a page.
type PageCustomCodeRequest struct {
	// Scripts is the list of scripts to apply to the page.
	Scripts []CustomCodeScript `json:"scripts"`
}

// ValidateScriptID validates that a scriptID is a non-empty string.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptID(scriptID string) error {
	if scriptID == "" {
		return errors.New("script id is required but was not provided. " +
			"Please provide the ID of a registered custom code script. " +
			"Scripts must be registered first using the RegisteredScript resource " +
			"before they can be applied to a page.")
	}
	return nil
}

// ValidateScriptVersion validates that a scriptVersion is a valid semantic version.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptVersion(version string) error {
	if version == "" {
		return errors.New("script version is required but was not provided. " +
			"Please provide a semantic version string (e.g., '1.0.0'). " +
			"Version must match a registered version of the script.")
	}
	return nil
}

// ValidateScriptLocation validates that a scriptLocation is either "header" or "footer".
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptLocation(location string) error {
	if location != "header" && location != "footer" {
		return fmt.Errorf("script location must be either 'header' or 'footer': got '%s'. "+
			"'header' = script loads in the page header (recommended for performance). "+
			"'footer' = script loads in the page footer (use for DOM-dependent scripts). "+
			"Please specify one of these two values.", location)
	}
	return nil
}

// GeneratePageCustomCodeResourceID generates a Pulumi resource ID for a PageCustomCode resource.
// Format: {pageID}/custom-code
// Note: PageCustomCode is a 1:1 relationship with a page, so we use a simple suffix.
func GeneratePageCustomCodeResourceID(pageID string) string {
	return pageID + "/custom-code"
}

// ExtractPageIDFromPageCustomCodeResourceID extracts the pageID from a PageCustomCode resource ID.
// Expected format: {pageID}/custom-code
func ExtractPageIDFromPageCustomCodeResourceID(resourceID string) (string, error) {
	if resourceID == "" {
		return "", errors.New("resourceId cannot be empty")
	}

	// Simple suffix removal
	suffix := "/custom-code"
	if len(resourceID) <= len(suffix) {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/custom-code, got: %s", resourceID)
	}

	if resourceID[len(resourceID)-len(suffix):] != suffix {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/custom-code, got: %s", resourceID)
	}

	pageID := resourceID[:len(resourceID)-len(suffix)]
	if pageID == "" {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/custom-code, got: %s", resourceID)
	}

	return pageID, nil
}

// getPageCustomCodeBaseURL is used internally for testing to override the API base URL.
var getPageCustomCodeBaseURL = ""

// GetPageCustomCode retrieves all custom code scripts applied to a page from Webflow.
// It calls GET /v2/pages/{page_id}/custom_code endpoint.
// Returns the parsed response or an error if the request fails.
func GetPageCustomCode(ctx context.Context, client *http.Client, pageID string) (*PageCustomCodeResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getPageCustomCodeBaseURL != "" {
		baseURL = getPageCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s/custom_code", baseURL, pageID)

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

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
				continue
			}
		}

		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response PageCustomCodeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// putPageCustomCodeBaseURL is used internally for testing to override the API base URL.
var putPageCustomCodeBaseURL = ""

// PutPageCustomCode updates custom code scripts applied to a page.
// It calls PUT /v2/pages/{page_id}/custom_code endpoint.
// Returns the updated response or an error if the request fails.
func PutPageCustomCode(
	ctx context.Context,
	client *http.Client,
	pageID string,
	request *PageCustomCodeRequest,
) (*PageCustomCodeResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if putPageCustomCodeBaseURL != "" {
		baseURL = putPageCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s/custom_code", baseURL, pageID)

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

		// Marshal request body
		reqBody, err := json.Marshal(request)
		if err != nil {
			return nil, fmt.Errorf("failed to marshal request: %w", err)
		}

		req, err := http.NewRequestWithContext(ctx, "PUT", url, bytes.NewReader(reqBody))
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

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return nil, fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
				continue
			}
		}

		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response PageCustomCodeResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deletePageCustomCodeBaseURL is used internally for testing to override the API base URL.
var deletePageCustomCodeBaseURL = ""

// DeletePageCustomCode removes all custom code scripts from a page.
// It calls DELETE /v2/pages/{page_id}/custom_code endpoint.
// This is idempotent: deleting a page with no custom code returns 204 (success).
func DeletePageCustomCode(ctx context.Context, client *http.Client, pageID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deletePageCustomCodeBaseURL != "" {
		baseURL = deletePageCustomCodeBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s/custom_code", baseURL, pageID)

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

			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
				attempt+1, maxRetries+1, waitTime)

			if attempt < maxRetries {
				select {
				case <-ctx.Done():
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
				continue
			}
		}

		// Handle 404 and 204 as success (idempotent delete)
		if resp.StatusCode == 204 || resp.StatusCode == 404 {
			return nil
		}

		if resp.StatusCode < 200 || resp.StatusCode >= 300 {
			return handleWebflowError(resp.StatusCode, body)
		}

		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}
