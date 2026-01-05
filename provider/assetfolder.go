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

// AssetFolderResponse represents a Webflow asset folder from the API.
type AssetFolderResponse struct {
	ID           string   `json:"id"`
	DisplayName  string   `json:"displayName"`
	ParentFolder string   `json:"parentFolder,omitempty"`
	Assets       []string `json:"assets,omitempty"`
	SiteID       string   `json:"siteId"`
	CreatedOn    string   `json:"createdOn"`
	LastUpdated  string   `json:"lastUpdated"`
}

// AssetFolderListResponse represents the response from listing asset folders.
type AssetFolderListResponse struct {
	AssetFolders []AssetFolderResponse `json:"assetFolders"`
	Pagination   struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination,omitempty"`
}

// AssetFolderCreateRequest represents the request body for creating an asset folder.
type AssetFolderCreateRequest struct {
	DisplayName  string `json:"displayName"`
	ParentFolder string `json:"parentFolder,omitempty"`
}

// ValidateAssetFolderID validates that an assetFolderID matches the Webflow asset folder ID format.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateAssetFolderID(assetFolderID string) error {
	if assetFolderID == "" {
		return errors.New("assetFolderId is required but was not provided. " +
			"Please provide a valid Webflow asset folder ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c3'). " +
			"You can find asset folder IDs in the Webflow dashboard under Assets")
	}
	if !siteIDPattern.MatchString(assetFolderID) {
		return fmt.Errorf("assetFolderId has invalid format: got '%s'. "+
			"Expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"Please check your asset folder ID in the Webflow dashboard "+
			"and ensure it contains only lowercase letters (a-f) and digits (0-9)", assetFolderID)
	}
	return nil
}

// GenerateAssetFolderResourceID generates a Pulumi resource ID for an AssetFolder resource.
// Format: {siteID}/asset-folders/{assetFolderID}
func GenerateAssetFolderResourceID(siteID, assetFolderID string) string {
	return fmt.Sprintf("%s/asset-folders/%s", siteID, assetFolderID)
}

// ExtractIDsFromAssetFolderResourceID extracts the siteID and assetFolderID from an AssetFolder resource ID.
// Expected format: {siteID}/asset-folders/{assetFolderID}
func ExtractIDsFromAssetFolderResourceID(resourceID string) (siteID, assetFolderID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "asset-folders" {
		return "", "", fmt.Errorf(
			"invalid resource ID format: expected {siteId}/asset-folders/{assetFolderId}, got: %s", resourceID)
	}

	siteID = parts[0]
	assetFolderID = strings.Join(parts[2:], "/") // Handle ID that might contain slashes

	return siteID, assetFolderID, nil
}

// listAssetFoldersBaseURL is used internally for testing to override the API base URL.
var listAssetFoldersBaseURL = ""

// ListAssetFolders retrieves all asset folders for a Webflow site.
// It calls GET /v2/sites/{site_id}/asset_folders endpoint.
// Returns the parsed response or an error if the request fails.
func ListAssetFolders(ctx context.Context, client *http.Client, siteID string) (*AssetFolderListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if listAssetFoldersBaseURL != "" {
		baseURL = listAssetFoldersBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/asset_folders", baseURL, siteID)

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

		var response AssetFolderListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getAssetFolderBaseURL is used internally for testing to override the API base URL.
var getAssetFolderBaseURL = ""

// GetAssetFolder retrieves a single asset folder by ID from Webflow.
// It calls GET /v2/asset_folders/{asset_folder_id} endpoint.
// Returns the parsed response or an error if the request fails.
func GetAssetFolder(ctx context.Context, client *http.Client, assetFolderID string) (*AssetFolderResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getAssetFolderBaseURL != "" {
		baseURL = getAssetFolderBaseURL
	}

	url := fmt.Sprintf("%s/v2/asset_folders/%s", baseURL, assetFolderID)

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

		var folder AssetFolderResponse
		if err := json.Unmarshal(body, &folder); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &folder, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postAssetFolderBaseURL is used internally for testing to override the API base URL.
var postAssetFolderBaseURL = ""

// PostAssetFolder creates a new asset folder in a Webflow site.
// It calls POST /v2/sites/{site_id}/asset_folders endpoint.
// Returns the created asset folder or an error if the request fails.
func PostAssetFolder(
	ctx context.Context, client *http.Client,
	siteID, displayName, parentFolder string,
) (*AssetFolderResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postAssetFolderBaseURL != "" {
		baseURL = postAssetFolderBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/asset_folders", baseURL, siteID)

	requestBody := AssetFolderCreateRequest{
		DisplayName:  displayName,
		ParentFolder: parentFolder,
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

		var folder AssetFolderResponse
		if err := json.Unmarshal(body, &folder); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &folder, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
