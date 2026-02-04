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

// InlineScriptResponse represents the Webflow API response for an inline script.
// This struct matches the Webflow API v2 response format for inline registered scripts.
type InlineScriptResponse struct {
	ID             string `json:"id,omitempty"`             // Human-readable ID derived from display name (read-only)
	DisplayName    string `json:"displayName"`              // User-facing name for the script (1-50 alphanumeric chars)
	SourceCode     string `json:"sourceCode"`               // The inline script source code
	HostedLocation string `json:"hostedLocation,omitempty"` // URI for the hosted version (read-only, set by Webflow)
	IntegrityHash  string `json:"integrityHash"`            // Sub-Resource Integrity Hash (SRI)
	CanCopy        bool   `json:"canCopy"`                  // Whether script can be copied on site duplication
	Version        string `json:"version"`                  // Semantic Version (SemVer) string
	CreatedOn      string `json:"createdOn,omitempty"`      // Timestamp when created (read-only)
	LastUpdated    string `json:"lastUpdated,omitempty"`    // Timestamp when last updated (read-only)
}

// InlineScriptRequest represents the request body for POST /registered_scripts/inline.
type InlineScriptRequest struct {
	SourceCode    string `json:"sourceCode"`
	Version       string `json:"version"`
	DisplayName   string `json:"displayName"`
	CanCopy       bool   `json:"canCopy,omitempty"`
	IntegrityHash string `json:"integrityHash,omitempty"`
}

// maxSourceCodeLength is the maximum number of characters allowed for inline script source code.
const maxSourceCodeLength = 2000

// ValidateSourceCode validates that a sourceCode value is valid for an inline script.
// Must be non-empty and at most 2000 characters.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateSourceCode(code string) error {
	if code == "" {
		return errors.New("sourceCode is required but was not provided. " +
			"Please provide the inline JavaScript code to register. " +
			"The code is limited to 2000 characters. " +
			"Example: 'console.log(\"Hello from Webflow\");'")
	}
	if len(code) > maxSourceCodeLength {
		return fmt.Errorf("sourceCode is too long: got %d characters, maximum is %d. "+
			"Please shorten your inline script code. "+
			"If your script is too large for inline registration, consider hosting it externally "+
			"and using the RegisteredScript resource with a hostedLocation instead",
			len(code), maxSourceCodeLength)
	}
	return nil
}

// GenerateInlineScriptResourceID generates a Pulumi resource ID for an InlineScript resource.
// Format: {siteID}/inline_scripts/{scriptID}
func GenerateInlineScriptResourceID(siteID, scriptID string) string {
	return fmt.Sprintf("%s/inline_scripts/%s", siteID, scriptID)
}

// ExtractIDsFromInlineScriptResourceID extracts the siteID and scriptID from an InlineScript resource ID.
// Expected format: {siteID}/inline_scripts/{scriptID}
func ExtractIDsFromInlineScriptResourceID(resourceID string) (siteID, scriptID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "inline_scripts" {
		return "", "",
			fmt.Errorf("invalid resource ID format: expected {siteId}/inline_scripts/{scriptId}, got: %s", resourceID)
	}

	siteID = parts[0]
	scriptID = strings.Join(parts[2:], "/") // Handle scriptID that might contain slashes

	return siteID, scriptID, nil
}

// postInlineScriptBaseURL is used internally for testing to override the API base URL.
var postInlineScriptBaseURL = ""

// PostInlineScript creates a new inline registered script for a Webflow site.
// It calls POST /v2/sites/{site_id}/registered_scripts/inline endpoint.
// Returns the created script or an error if the request fails.
func PostInlineScript(
	ctx context.Context, client *http.Client,
	siteID, sourceCode, version, displayName string, canCopy bool, integrityHash string,
) (*InlineScriptResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postInlineScriptBaseURL != "" {
		baseURL = postInlineScriptBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/registered_scripts/inline", baseURL, siteID)

	requestBody := InlineScriptRequest{
		SourceCode:    sourceCode,
		Version:       version,
		DisplayName:   displayName,
		CanCopy:       canCopy,
		IntegrityHash: integrityHash,
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
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var script InlineScriptResponse
		if err := json.Unmarshal(body, &script); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &script, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
