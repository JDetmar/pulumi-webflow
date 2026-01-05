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
	"regexp"
	"strings"
	"time"
)

// RegisteredScript represents a registered custom code script in Webflow.
// This struct matches the Webflow API v2 response format for registered scripts.
type RegisteredScript struct {
	ID             string `json:"id,omitempty"`          // Human-readable ID derived from display name (read-only)
	DisplayName    string `json:"displayName"`           // User-facing name for the script (1-50 alphanumeric chars)
	HostedLocation string `json:"hostedLocation"`        // URI for externally hosted script
	IntegrityHash  string `json:"integrityHash"`         // Sub-Resource Integrity Hash (SRI)
	CanCopy        bool   `json:"canCopy"`               // Whether script can be copied on site duplication
	Version        string `json:"version"`               // Semantic Version (SemVer) string
	CreatedOn      string `json:"createdOn,omitempty"`   // Timestamp when created (read-only)
	LastUpdated    string `json:"lastUpdated,omitempty"` // Timestamp when last updated (read-only)
}

// RegisteredScriptsResponse represents the Webflow API response for listing registered scripts.
type RegisteredScriptsResponse struct {
	RegisteredScripts []RegisteredScript `json:"registeredScripts"`
	Pagination        PaginationInfo     `json:"pagination,omitempty"`
}

// PaginationInfo represents pagination metadata from the API response.
type PaginationInfo struct {
	Limit  int `json:"limit"`
	Offset int `json:"offset"`
	Total  int `json:"total"`
}

// RegisteredScriptRequest represents the request body for POST/PATCH registered scripts.
type RegisteredScriptRequest struct {
	DisplayName    string `json:"displayName"`
	HostedLocation string `json:"hostedLocation"`
	IntegrityHash  string `json:"integrityHash"`
	CanCopy        bool   `json:"canCopy,omitempty"`
	Version        string `json:"version"`
}

// displayNamePattern validates that display names are between 1-50 alphanumeric characters.
var displayNamePattern = regexp.MustCompile(`^[a-zA-Z0-9]{1,50}$`)

// ValidateScriptDisplayName validates that a displayName is a valid Webflow script name.
// Must be 1-50 alphanumeric characters.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateScriptDisplayName(name string) error {
	if name == "" {
		return errors.New("displayName is required but was not provided; " +
			"please provide a user-facing name for the script between 1 and 50 alphanumeric characters; " +
			"example valid names: 'CmsSlider', 'AnalyticsScript', 'MyCustomScript123'")
	}
	if len(name) > 50 {
		return fmt.Errorf("displayName is too long: got %d characters, maximum is 50; "+
			"please shorten the name; "+
			"example valid names: 'CmsSlider', 'AnalyticsScript', 'MyCustomScript123'", len(name))
	}
	if !displayNamePattern.MatchString(name) {
		return fmt.Errorf("displayName contains invalid characters: got '%s', "+
			"allowed characters: A-Z, a-z, 0-9; spaces and special characters are not allowed; "+
			"example valid names: 'CmsSlider', 'AnalyticsScript', 'MyCustomScript123'; "+
			"please use only alphanumeric characters", name)
	}
	return nil
}

// ValidateHostedLocation validates that a hostedLocation is a valid HTTP/HTTPS URL.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateHostedLocation(url string) error {
	if url == "" {
		return errors.New("hostedLocation is required but was not provided; " +
			"please provide a valid HTTP or HTTPS URL where your script is hosted; " +
			"example: 'https://cdn.example.com/my-script.js'")
	}
	if !strings.HasPrefix(url, "http://") && !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("hostedLocation must start with 'http://' or 'https://': got '%s', "+
			"example valid URLs: 'https://cdn.example.com/my-script.js', 'https://cdnjs.cloudflare.com/...'; "+
			"please ensure the URL is properly formatted with a scheme", url)
	}
	return nil
}

// ValidateIntegrityHash validates that an integrityHash is a properly formatted SRI hash.
// Should be in format: sha384-<hash> or sha256-<hash> or sha512-<hash>
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateIntegrityHash(hash string) error {
	if hash == "" {
		return errors.New("integrityHash is required but was not provided; " +
			"please provide a Sub-Resource Integrity (SRI) hash for your hosted script; " +
			"format: 'sha384-<hash>', 'sha256-<hash>', or 'sha512-<hash>'; " +
			"you can generate an SRI hash using: https://www.srihash.org/")
	}
	if !strings.HasPrefix(hash, "sha") {
		return fmt.Errorf("integrityHash must start with 'sha': got '%s', "+
			"supported algorithms: sha256, sha384, sha512; "+
			"format: 'sha384-<hash>', 'sha256-<hash>', or 'sha512-<hash>'; "+
			"you can generate an SRI hash using: https://www.srihash.org/", hash)
	}
	return nil
}

// ValidateVersion validates that a version is a proper semantic version string.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateVersion(version string) error {
	if version == "" {
		return errors.New("version is required but was not provided; " +
			"please provide a Semantic Version (SemVer) string for your script; " +
			"format: 'major.minor.patch' (e.g., '1.0.0', '2.3.1'); " +
			"see https://semver.org/ for more information")
	}
	if !strings.Contains(version, ".") {
		return fmt.Errorf("version must be in Semantic Version format: got '%s', "+
			"expected format: 'major.minor.patch' (e.g., '1.0.0', '2.3.1'); "+
			"see https://semver.org/ for more information", version)
	}
	return nil
}

// GenerateRegisteredScriptResourceID generates a Pulumi resource ID for a RegisteredScript resource.
// Format: {siteID}/registered_scripts/{scriptID}
func GenerateRegisteredScriptResourceID(siteID, scriptID string) string {
	return fmt.Sprintf("%s/registered_scripts/%s", siteID, scriptID)
}

// ExtractIDsFromRegisteredScriptResourceID extracts the siteID and scriptID from a RegisteredScript resource ID.
// Expected format: {siteID}/registered_scripts/{scriptID}
func ExtractIDsFromRegisteredScriptResourceID(resourceID string) (siteID, scriptID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "registered_scripts" {
		return "", "",
			fmt.Errorf("invalid resource ID format: expected {siteId}/registered_scripts/{scriptId}, got: %s", resourceID)
	}

	siteID = parts[0]
	scriptID = strings.Join(parts[2:], "/") // Handle scriptID that might contain slashes

	return siteID, scriptID, nil
}

// getRegisteredScriptsBaseURL is used internally for testing to override the API base URL.
var getRegisteredScriptsBaseURL = ""

// GetRegisteredScripts retrieves all registered scripts for a Webflow site.
// It calls GET /v2/sites/{site_id}/registered_scripts endpoint.
// Returns the parsed response or an error if the request fails.
func GetRegisteredScripts(ctx context.Context, client *http.Client, siteID string) (*RegisteredScriptsResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getRegisteredScriptsBaseURL != "" {
		baseURL = getRegisteredScriptsBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/registered_scripts", baseURL, siteID)

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
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var response RegisteredScriptsResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postRegisteredScriptBaseURL is used internally for testing to override the API base URL.
var postRegisteredScriptBaseURL = ""

// PostRegisteredScript creates a new registered script for a Webflow site.
// It calls POST /v2/sites/{site_id}/registered_scripts/hosted endpoint.
// Returns the created script or an error if the request fails.
func PostRegisteredScript(
	ctx context.Context, client *http.Client,
	siteID, displayName, hostedLocation, integrityHash, version string, canCopy bool,
) (*RegisteredScript, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postRegisteredScriptBaseURL != "" {
		baseURL = postRegisteredScriptBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/registered_scripts/hosted", baseURL, siteID)

	requestBody := RegisteredScriptRequest{
		DisplayName:    displayName,
		HostedLocation: hostedLocation,
		IntegrityHash:  integrityHash,
		CanCopy:        canCopy,
		Version:        version,
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
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var script RegisteredScript
		if err := json.Unmarshal(body, &script); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &script, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// patchRegisteredScriptBaseURL is used internally for testing to override the API base URL.
var patchRegisteredScriptBaseURL = ""

// PatchRegisteredScript updates an existing registered script for a Webflow site.
// It calls PATCH /v2/sites/{site_id}/registered_scripts/{script_id} endpoint.
// Returns the updated script or an error if the request fails.
func PatchRegisteredScript(
	ctx context.Context, client *http.Client,
	siteID, scriptID, displayName, hostedLocation, integrityHash, version string, canCopy bool,
) (*RegisteredScript, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if patchRegisteredScriptBaseURL != "" {
		baseURL = patchRegisteredScriptBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/registered_scripts/%s", baseURL, siteID, scriptID)

	requestBody := RegisteredScriptRequest{
		DisplayName:    displayName,
		HostedLocation: hostedLocation,
		IntegrityHash:  integrityHash,
		CanCopy:        canCopy,
		Version:        version,
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
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var script RegisteredScript
		if err := json.Unmarshal(body, &script); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &script, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteRegisteredScriptBaseURL is used internally for testing to override the API base URL.
var deleteRegisteredScriptBaseURL = ""

// DeleteRegisteredScript removes a registered script from a Webflow site.
// It calls DELETE /v2/sites/{site_id}/registered_scripts/{script_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteRegisteredScript(ctx context.Context, client *http.Client, siteID, scriptID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteRegisteredScriptBaseURL != "" {
		baseURL = deleteRegisteredScriptBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/registered_scripts/%s", baseURL, siteID, scriptID)

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
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429), "+
				"the provider will automatically retry with exponential backoff; "+
				"retry attempt %d of %d, waiting %v before next attempt; "+
				"if this error persists, please wait a few minutes before trying again or contact Webflow support",
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
