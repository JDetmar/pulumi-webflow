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

// CollectionFieldResponse represents the Webflow API response for a collection field.
// This struct matches the Webflow API v2 response format for collection fields.
type CollectionFieldResponse struct {
	ID          string                 `json:"id"`                    // Webflow-assigned field ID (read-only)
	IsEditable  bool                   `json:"isEditable"`            // Whether the field can be edited (read-only)
	IsRequired  bool                   `json:"isRequired"`            // Whether the field is required
	Type        string                 `json:"type"`                  // Field type (PlainText, RichText, Image, etc.)
	Slug        string                 `json:"slug"`                  // URL-friendly slug for the field
	DisplayName string                 `json:"displayName"`           // Human-readable name of the field
	HelpText    string                 `json:"helpText,omitempty"`    // Optional help text for the field
	Validations map[string]interface{} `json:"validations,omitempty"` // Type-specific validations
}

// CollectionFieldRequest represents the request body for POST/PUT collection field.
type CollectionFieldRequest struct {
	Type        string                 `json:"type"`                  // Field type (required for POST)
	DisplayName string                 `json:"displayName"`           // Human-readable name (required)
	Slug        string                 `json:"slug,omitempty"`        // Optional URL slug
	IsRequired  bool                   `json:"isRequired,omitempty"`  // Whether the field is required
	HelpText    string                 `json:"helpText,omitempty"`    // Optional help text
	Validations map[string]interface{} `json:"validations,omitempty"` // Type-specific validations
}

// Valid field types for Webflow collection fields.
const (
	FieldTypePlainText      = "PlainText"
	FieldTypeRichText       = "RichText"
	FieldTypeImage          = "Image"
	FieldTypeMultiImage     = "MultiImage"
	FieldTypeVideo          = "Video"
	FieldTypeLink           = "Link"
	FieldTypeEmail          = "Email"
	FieldTypePhone          = "Phone"
	FieldTypeNumber         = "Number"
	FieldTypeDateTime       = "DateTime"
	FieldTypeSwitch         = "Switch"
	FieldTypeColor          = "Color"
	FieldTypeOption         = "Option"
	FieldTypeFile           = "File"
	FieldTypeReference      = "Reference"
	FieldTypeMultiReference = "MultiReference"
)

// ValidFieldTypes is a map of all valid field types for validation.
var ValidFieldTypes = map[string]bool{
	FieldTypePlainText:      true,
	FieldTypeRichText:       true,
	FieldTypeImage:          true,
	FieldTypeMultiImage:     true,
	FieldTypeVideo:          true,
	FieldTypeLink:           true,
	FieldTypeEmail:          true,
	FieldTypePhone:          true,
	FieldTypeNumber:         true,
	FieldTypeDateTime:       true,
	FieldTypeSwitch:         true,
	FieldTypeColor:          true,
	FieldTypeOption:         true,
	FieldTypeFile:           true,
	FieldTypeReference:      true,
	FieldTypeMultiReference: true,
}

// ValidateFieldType validates that a field type is one of the supported types.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateFieldType(fieldType string) error {
	if fieldType == "" {
		return errors.New("type is required but was not provided. " +
			"Please provide a valid field type (e.g., 'PlainText', 'RichText', 'Image'). " +
			"Supported types: PlainText, RichText, Image, MultiImage, Video, Link, Email, Phone, " +
			"Number, DateTime, Switch, Color, Option, File, Reference, MultiReference.")
	}
	if !ValidFieldTypes[fieldType] {
		return fmt.Errorf("type has invalid value: got '%s'. "+
			"Supported types: PlainText, RichText, Image, MultiImage, Video, Link, Email, Phone, "+
			"Number, DateTime, Switch, Color, Option, File, Reference, MultiReference. "+
			"Please use one of the supported field types.", fieldType)
	}
	return nil
}

// ValidateFieldDisplayName validates that displayName is non-empty and reasonable length.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateFieldDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("displayName is required but was not provided. " +
			"Please provide a name for your field (e.g., 'Title', 'Description', 'Author'). " +
			"The display name is shown in the Webflow CMS interface.")
	}
	if len(displayName) > 255 {
		return fmt.Errorf("displayName is too long: '%s' exceeds maximum length of 255 characters. "+
			"Please use a shorter, more concise name for your field.", displayName)
	}
	return nil
}

// GenerateCollectionFieldResourceID generates a Pulumi resource ID for a CollectionField resource.
// Format: {collectionID}/fields/{fieldID}
func GenerateCollectionFieldResourceID(collectionID, fieldID string) string {
	return fmt.Sprintf("%s/fields/%s", collectionID, fieldID)
}

// ExtractIDsFromCollectionFieldResourceID extracts the collectionID and fieldID from a CollectionField resource ID.
// Expected format: {collectionID}/fields/{fieldID}
func ExtractIDsFromCollectionFieldResourceID(resourceID string) (collectionID, fieldID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "fields" {
		return "", "", fmt.Errorf(
			"invalid resource ID format: expected {collectionId}/fields/{fieldId}, got: %s",
			resourceID,
		)
	}

	collectionID = parts[0]
	fieldID = strings.Join(parts[2:], "/") // Handle fieldID that might contain slashes

	return collectionID, fieldID, nil
}

// getCollectionFieldBaseURL is used internally for testing to override the API base URL.
var getCollectionFieldBaseURL = ""

// GetCollectionField retrieves a single collection field by ID.
// Note: Webflow doesn't have a direct GET endpoint for individual fields.
// We fetch the entire collection and filter for the specific field.
func GetCollectionField(
	ctx context.Context, client *http.Client, collectionID, fieldID string,
) (*CollectionFieldResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getCollectionFieldBaseURL != "" {
		baseURL = getCollectionFieldBaseURL
	}

	// Get the full collection which includes all fields
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
			}
			continue
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		// Parse collection response to extract fields
		var collection struct {
			Fields []CollectionFieldResponse `json:"fields"`
		}
		if err := json.Unmarshal(body, &collection); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		// Find the specific field
		for _, field := range collection.Fields {
			if field.ID == fieldID {
				return &field, nil
			}
		}

		// Field not found in collection
		return nil, errors.New("not found: the collection field does not exist. " +
			"The field may have been deleted or the field ID is incorrect. " +
			"To fix this: 1) Verify the field ID is correct, " +
			"2) Check that the field exists in the collection")
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postCollectionFieldBaseURL is used internally for testing to override the API base URL.
var postCollectionFieldBaseURL = ""

// PostCollectionField creates a new field for a Webflow collection.
// It calls POST /v2/collections/{collection_id}/fields endpoint.
// Returns the created field or an error if the request fails.
func PostCollectionField(
	ctx context.Context, client *http.Client,
	collectionID, fieldType, displayName, slug, helpText string,
	isRequired bool, validations map[string]interface{},
) (*CollectionFieldResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postCollectionFieldBaseURL != "" {
		baseURL = postCollectionFieldBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/fields", baseURL, collectionID)

	requestBody := CollectionFieldRequest{
		Type:        fieldType,
		DisplayName: displayName,
		Slug:        slug,
		IsRequired:  isRequired,
		HelpText:    helpText,
		Validations: validations,
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
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
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

		var field CollectionFieldResponse
		if err := json.Unmarshal(body, &field); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &field, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// putCollectionFieldBaseURL is used internally for testing to override the API base URL.
var putCollectionFieldBaseURL = ""

// PutCollectionField updates an existing field for a Webflow collection.
// It calls PUT /v2/collections/{collection_id}/fields/{field_id} endpoint.
// Returns the updated field or an error if the request fails.
func PutCollectionField(
	ctx context.Context, client *http.Client,
	collectionID, fieldID, displayName, slug, helpText string,
	isRequired bool, validations map[string]interface{},
) (*CollectionFieldResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if putCollectionFieldBaseURL != "" {
		baseURL = putCollectionFieldBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/fields/%s", baseURL, collectionID, fieldID)

	requestBody := CollectionFieldRequest{
		// Note: Type is NOT included in PUT requests - it cannot be changed
		DisplayName: displayName,
		Slug:        slug,
		IsRequired:  isRequired,
		HelpText:    helpText,
		Validations: validations,
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
			}
			continue
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var field CollectionFieldResponse
		if err := json.Unmarshal(body, &field); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &field, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteCollectionFieldBaseURL is used internally for testing to override the API base URL.
var deleteCollectionFieldBaseURL = ""

// DeleteCollectionField removes a field from a Webflow collection.
// It calls DELETE /v2/collections/{collection_id}/fields/{field_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteCollectionField(ctx context.Context, client *http.Client, collectionID, fieldID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteCollectionFieldBaseURL != "" {
		baseURL = deleteCollectionFieldBaseURL
	}

	url := fmt.Sprintf("%s/v2/collections/%s/fields/%s", baseURL, collectionID, fieldID)

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
				"If this error persists, please wait a few minutes before trying again or contact Webflow support.",
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
