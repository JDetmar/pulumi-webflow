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

// CollectionItem represents a single item in a Webflow CMS collection.
// This struct matches the Webflow API v2 response format for collection items.
type CollectionItem struct {
	ID            string                 `json:"id,omitempty"`            // Webflow-assigned item ID (read-only)
	CmsLocaleID   string                 `json:"cmsLocaleId,omitempty"`   // Locale ID for localized sites
	LastPublished string                 `json:"lastPublished,omitempty"` // Last publish timestamp (read-only)
	LastUpdated   string                 `json:"lastUpdated,omitempty"`   // Last update timestamp (read-only)
	CreatedOn     string                 `json:"createdOn,omitempty"`     // Creation timestamp (read-only)
	IsArchived    bool                   `json:"isArchived"`              // Whether the item is archived
	IsDraft       bool                   `json:"isDraft"`                 // Whether the item is a draft
	FieldData     map[string]interface{} `json:"fieldData"`               // Dynamic field data (name, slug, etc.)
}

// CollectionItemListResponse represents the Webflow API response for listing collection items.
type CollectionItemListResponse struct {
	Items      []CollectionItem `json:"items"`                // List of collection items
	Pagination *Pagination      `json:"pagination,omitempty"` // Pagination metadata (if applicable)
}

// Pagination represents pagination metadata for list responses.
type Pagination struct {
	Limit  int `json:"limit,omitempty"`  // Number of items per page
	Offset int `json:"offset,omitempty"` // Offset for pagination
	Total  int `json:"total,omitempty"`  // Total number of items
}

// CollectionItemRequest represents the request body for POST/PATCH collection items.
type CollectionItemRequest struct {
	FieldData   map[string]interface{} `json:"fieldData"`             // Dynamic field data
	IsArchived  *bool                  `json:"isArchived,omitempty"`  // Whether the item is archived
	IsDraft     *bool                  `json:"isDraft,omitempty"`     // Whether the item is a draft
	CmsLocaleID string                 `json:"cmsLocaleId,omitempty"` // Locale ID for localized sites
}

// ValidateFieldData validates that fieldData is non-empty and contains required fields.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateFieldData(fieldData map[string]interface{}) error {
	if len(fieldData) == 0 {
		return errors.New("fieldData is required but was not provided. " +
			"Please provide a map of field slugs to values (e.g., {\"name\": \"My Item\", \"slug\": \"my-item\"}). " +
			"The field slugs must match the fields defined in the collection schema")
	}
	// Note: Name and slug are typically required but may be auto-generated
	// We don't enforce this here as it depends on the collection schema
	return nil
}

// GenerateCollectionItemResourceID generates a Pulumi resource ID for a CollectionItem resource.
// Format: {collectionID}/items/{itemID}
func GenerateCollectionItemResourceID(collectionID, itemID string) string {
	return fmt.Sprintf("%s/items/%s", collectionID, itemID)
}

// ExtractIDsFromCollectionItemResourceID extracts the collectionID and itemID from a CollectionItem resource ID.
// Expected format: {collectionID}/items/{itemID}
func ExtractIDsFromCollectionItemResourceID(resourceID string) (collectionID, itemID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "items" {
		return "", "", fmt.Errorf(
			"invalid resource ID format: expected {collectionId}/items/{itemId}, got: %s",
			resourceID,
		)
	}

	collectionID = parts[0]
	itemID = strings.Join(parts[2:], "/") // Handle itemID that might contain slashes

	return collectionID, itemID, nil
}

// getCollectionItemsBaseURL is used internally for testing to override the API base URL.
var getCollectionItemsBaseURL = ""

// GetCollectionItems retrieves all items for a Webflow collection.
// It calls GET /v2/collections/{collection_id}/items endpoint.
// Returns the parsed response or an error if the request fails.
func GetCollectionItems(
	ctx context.Context, client *http.Client, collectionID string,
) (*CollectionItemListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getCollectionItemsBaseURL != "" {
		baseURL = getCollectionItemsBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/items", baseURL, collectionID)

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

		var response CollectionItemListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getCollectionItemBaseURL is used internally for testing to override the API base URL.
var getCollectionItemBaseURL = ""

// GetCollectionItem retrieves a single collection item by ID.
// It calls GET /v2/collections/{collection_id}/items/{item_id} endpoint.
// Returns the collection item or an error if the request fails.
func GetCollectionItem(ctx context.Context, client *http.Client, collectionID, itemID string) (*CollectionItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getCollectionItemBaseURL != "" {
		baseURL = getCollectionItemBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/items/%s", baseURL, collectionID, itemID)

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

		var item CollectionItem
		if err := json.Unmarshal(body, &item); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &item, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postCollectionItemBaseURL is used internally for testing to override the API base URL.
var postCollectionItemBaseURL = ""

// PostCollectionItem creates a new item in a Webflow collection.
// It calls POST /v2/collections/{collection_id}/items endpoint.
// Returns the created collection item or an error if the request fails.
func PostCollectionItem(
	ctx context.Context, client *http.Client,
	collectionID string, fieldData map[string]interface{},
	isArchived, isDraft *bool, cmsLocaleID string,
) (*CollectionItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postCollectionItemBaseURL != "" {
		baseURL = postCollectionItemBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/items", baseURL, collectionID)

	requestBody := CollectionItemRequest{
		FieldData:   fieldData,
		IsArchived:  isArchived,
		IsDraft:     isDraft,
		CmsLocaleID: cmsLocaleID,
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

		// Accept 200, 201, and 202 as success
		// 202 Accepted is returned when the item is created asynchronously
		if resp.StatusCode != 200 && resp.StatusCode != 201 && resp.StatusCode != 202 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var item CollectionItem
		if err := json.Unmarshal(body, &item); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &item, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// patchCollectionItemBaseURL is used internally for testing to override the API base URL.
var patchCollectionItemBaseURL = ""

// PatchCollectionItem updates an existing item in a Webflow collection.
// It calls PATCH /v2/collections/{collection_id}/items/{item_id} endpoint.
// Returns the updated collection item or an error if the request fails.
func PatchCollectionItem(
	ctx context.Context, client *http.Client,
	collectionID, itemID string, fieldData map[string]interface{},
	isArchived, isDraft *bool, cmsLocaleID string,
) (*CollectionItem, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if patchCollectionItemBaseURL != "" {
		baseURL = patchCollectionItemBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/items/%s", baseURL, collectionID, itemID)

	requestBody := CollectionItemRequest{
		FieldData:   fieldData,
		IsArchived:  isArchived,
		IsDraft:     isDraft,
		CmsLocaleID: cmsLocaleID,
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

		var item CollectionItem
		if err := json.Unmarshal(body, &item); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &item, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteCollectionItemBaseURL is used internally for testing to override the API base URL.
var deleteCollectionItemBaseURL = ""

// DeleteCollectionItem removes an item from a Webflow collection.
// It calls DELETE /v2/collections/{collection_id}/items/{item_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteCollectionItem(ctx context.Context, client *http.Client, collectionID, itemID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteCollectionItemBaseURL != "" {
		baseURL = deleteCollectionItemBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/items/%s", baseURL, collectionID, itemID)

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
