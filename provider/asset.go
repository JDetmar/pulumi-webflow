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

// AssetVariant represents different size variants of an uploaded asset.
type AssetVariant struct {
	URL    string `json:"url"`
	Width  int    `json:"width,omitempty"`
	Height int    `json:"height,omitempty"`
	Size   int    `json:"size,omitempty"`
}

// AssetResponse represents a Webflow asset from the API.
type AssetResponse struct {
	ID               string                  `json:"id"`
	ContentType      string                  `json:"contentType"`
	Size             int                     `json:"size"`
	SiteID           string                  `json:"siteId"`
	HostedURL        string                  `json:"hostedUrl"`
	OriginalFileName string                  `json:"originalFileName"`
	DisplayName      string                  `json:"displayName,omitempty"`
	CreatedOn        string                  `json:"createdOn"`
	LastUpdated      string                  `json:"lastUpdated"`
	Variants         map[string]AssetVariant `json:"variants,omitempty"`
}

// AssetListResponse represents the response from listing assets.
type AssetListResponse struct {
	Assets     []AssetResponse `json:"assets"`
	Pagination struct {
		Total  int `json:"total"`
		Limit  int `json:"limit"`
		Offset int `json:"offset"`
	} `json:"pagination,omitempty"`
}

// AssetUploadDetails contains additional signed parameters for S3 upload.
type AssetUploadDetails struct {
	URL         string `json:"url"`
	ContentType string `json:"contentType"`
	ACL         string `json:"acl,omitempty"`
}

// AssetUploadResponse represents the response from requesting an asset upload URL.
type AssetUploadResponse struct {
	UploadURL     string             `json:"uploadUrl"`
	UploadDetails AssetUploadDetails `json:"uploadDetails"`
}

// AssetUploadRequest represents the request body for initiating an asset upload.
type AssetUploadRequest struct {
	FileName     string `json:"fileName"`
	FileHash     string `json:"fileHash,omitempty"`
	ParentFolder string `json:"parentFolder,omitempty"`
}

// assetIDPattern is the regex pattern for validating Webflow asset IDs.
// Asset IDs are typically 24-character hexadecimal strings.
var assetIDPattern = regexp.MustCompile(`^[a-f0-9]{24}$`)

// ValidateAssetID validates that an assetID matches the Webflow asset ID format.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateAssetID(assetID string) error {
	if assetID == "" {
		return errors.New("assetId is required but was not provided; " +
			"please provide a valid Webflow asset ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c3'); " +
			"you can find asset IDs in the Webflow dashboard under Assets")
	}
	if !assetIDPattern.MatchString(assetID) {
		return fmt.Errorf("assetId has invalid format: got '%s'; "+
			"expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c3'); "+
			"please check your asset ID in the Webflow dashboard "+
			"and ensure it contains only lowercase letters (a-f) and digits (0-9)", assetID)
	}
	return nil
}

// ValidateFileName validates that a fileName is non-empty and has a reasonable format.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateFileName(fileName string) error {
	if fileName == "" {
		return errors.New("fileName is required but was not provided; " +
			"please provide a valid file name with extension " +
			"(e.g., 'logo.png', 'hero-image.jpg', 'document.pdf')")
	}

	// Check for reasonable length
	if len(fileName) > 255 {
		return fmt.Errorf("fileName is too long: '%s' exceeds maximum length of 255 characters, "+
			"please use a shorter file name", fileName)
	}

	// Check for common invalid characters (most filesystems disallow these)
	invalidChars := []string{"<", ">", ":", "\"", "|", "?", "*"}
	for _, char := range invalidChars {
		if strings.Contains(fileName, char) {
			return fmt.Errorf("fileName contains invalid character '%s': got '%s', "+
				"please remove invalid characters from the file name; "+
				"valid characters: letters, numbers, hyphens, underscores, dots, spaces", char, fileName)
		}
	}

	return nil
}

// GenerateAssetResourceID generates a Pulumi resource ID for an Asset resource.
// Format: {siteID}/assets/{assetID}
func GenerateAssetResourceID(siteID, assetID string) string {
	return fmt.Sprintf("%s/assets/%s", siteID, assetID)
}

// ExtractIDsFromAssetResourceID extracts the siteID and assetID from an Asset resource ID.
// Expected format: {siteID}/assets/{assetID}
func ExtractIDsFromAssetResourceID(resourceID string) (siteID, assetID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "assets" {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/assets/{assetId}, got: %s", resourceID)
	}

	siteID = parts[0]
	assetID = strings.Join(parts[2:], "/") // Handle assetID that might contain slashes

	return siteID, assetID, nil
}

// getAssetBaseURL is used internally for testing to override the API base URL.
var getAssetBaseURL = ""

// GetAsset retrieves a single asset by ID from Webflow.
// It calls GET /v2/assets/{asset_id} endpoint.
// Returns the parsed response or an error if the request fails.
func GetAsset(ctx context.Context, client *http.Client, assetID string) (*AssetResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getAssetBaseURL != "" {
		baseURL = getAssetBaseURL
	}

	url := fmt.Sprintf("%s/v2/assets/%s", baseURL, assetID)

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

		var asset AssetResponse
		if err := json.Unmarshal(body, &asset); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &asset, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// listAssetsBaseURL is used internally for testing to override the API base URL.
var listAssetsBaseURL = ""

// ListAssets retrieves all assets for a Webflow site.
// It calls GET /v2/sites/{site_id}/assets endpoint.
// Returns the parsed response or an error if the request fails.
func ListAssets(ctx context.Context, client *http.Client, siteID string) (*AssetListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if listAssetsBaseURL != "" {
		baseURL = listAssetsBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/assets", baseURL, siteID)

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

		var response AssetListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postAssetUploadURLBaseURL is used internally for testing to override the API base URL.
var postAssetUploadURLBaseURL = ""

// PostAssetUploadURL requests a presigned upload URL from Webflow for uploading an asset.
// This is step 1 of the 2-step asset upload process.
// It calls POST /v2/sites/{site_id}/assets endpoint.
// Returns the upload URL and details for uploading to S3.
func PostAssetUploadURL(
	ctx context.Context, client *http.Client,
	siteID, fileName, fileHash, parentFolder string,
) (*AssetUploadResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postAssetUploadURLBaseURL != "" {
		baseURL = postAssetUploadURLBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/assets", baseURL, siteID)

	requestBody := AssetUploadRequest{
		FileName:     fileName,
		FileHash:     fileHash,
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

		// Handle error responses (accept both 200 and 201 as success)
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var uploadResp AssetUploadResponse
		if err := json.Unmarshal(body, &uploadResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &uploadResp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteAssetBaseURL is used internally for testing to override the API base URL.
var deleteAssetBaseURL = ""

// DeleteAsset deletes an asset from Webflow.
// It calls DELETE /v2/assets/{asset_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteAsset(ctx context.Context, client *http.Client, assetID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteAssetBaseURL != "" {
		baseURL = deleteAssetBaseURL
	}

	url := fmt.Sprintf("%s/v2/assets/%s", baseURL, assetID)

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
