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

// WebhookResponse represents a webhook configuration in Webflow.
// This struct matches the Webflow API v2 response format for webhooks.
type WebhookResponse struct {
	ID            string                 `json:"id,omitempty"`            // Webflow-assigned webhook ID
	TriggerType   string                 `json:"triggerType"`             // Event that triggers the webhook
	URL           string                 `json:"url"`                     // HTTPS endpoint to receive webhook
	WorkspaceID   string                 `json:"workspaceId,omitempty"`   // Workspace ID (read-only)
	SiteID        string                 `json:"siteId"`                  // Site ID
	LastTriggered string                 `json:"lastTriggered,omitempty"` // Last trigger timestamp (read-only)
	CreatedOn     string                 `json:"createdOn,omitempty"`     // Creation timestamp (read-only)
	Filter        map[string]interface{} `json:"filter,omitempty"`        // Optional event filter
}

// WebhooksListResponse represents the Webflow API response for listing webhooks.
type WebhooksListResponse struct {
	Webhooks []WebhookResponse `json:"webhooks"` // List of webhooks
}

// WebhookRequest represents the request body for POST webhooks.
type WebhookRequest struct {
	TriggerType string                 `json:"triggerType"`      // Event that triggers the webhook
	URL         string                 `json:"url"`              // HTTPS endpoint to receive webhook
	Filter      map[string]interface{} `json:"filter,omitempty"` // Optional event filter
}

// webhookIDPattern is the regex pattern for validating Webflow webhook IDs.
// Webhook IDs are 24-character lowercase hexadecimal strings (same format as site IDs).
var webhookIDPattern = regexp.MustCompile(`^[a-f0-9]{24}$`)

// Valid trigger types from Webflow API documentation
var validTriggerTypes = map[string]bool{
	"form_submission":                  true,
	"site_publish":                     true,
	"page_created":                     true,
	"page_metadata_updated":            true,
	"page_deleted":                     true,
	"ecomm_new_order":                  true,
	"ecomm_order_changed":              true,
	"ecomm_inventory_changed":          true,
	"memberships_user_account_added":   true,
	"memberships_user_account_updated": true,
	"memberships_user_account_deleted": true,
	"collection_item_created":          true,
	"collection_item_changed":          true,
	"collection_item_deleted":          true,
	"collection_item_unpublished":      true,
}

// ValidateWebhookID validates that a webhookID matches the Webflow webhook ID format.
// Webhook IDs are 24-character lowercase hexadecimal strings.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateWebhookID(webhookID string) error {
	if webhookID == "" {
		return errors.New("webhookId is required but was not provided. " +
			"Please provide a valid Webflow webhook ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c3'). " +
			"Webhook IDs are assigned by Webflow when a webhook is created.")
	}
	if !webhookIDPattern.MatchString(webhookID) {
		return fmt.Errorf("webhookId has invalid format: got '%s'. "+
			"Expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"Please check the webhook ID and ensure it contains only lowercase letters (a-f) and digits (0-9).", webhookID)
	}
	return nil
}

// ValidateWebhookURL validates that a webhook URL is a valid HTTPS endpoint.
// Webflow requires webhook URLs to use HTTPS for security.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateWebhookURL(url string) error {
	if url == "" {
		return errors.New("url is required but was not provided. " +
			"Please provide a valid HTTPS URL where Webflow should send webhook events " +
			"(e.g., 'https://example.com/webhooks/webflow', 'https://api.example.com/events'). " +
			"Note: Webflow requires HTTPS URLs for security.")
	}
	if !strings.HasPrefix(url, "https://") {
		return fmt.Errorf("url must use HTTPS protocol: got '%s'. "+
			"Webflow requires webhook URLs to use HTTPS for security. "+
			"Example valid URLs: 'https://example.com/webhooks/webflow', 'https://api.example.com/events'. "+
			"Please update the URL to use HTTPS instead of HTTP.", url)
	}
	// Basic URL format validation
	if !strings.Contains(url[8:], ".") {
		return fmt.Errorf("url appears to be invalid: got '%s'. "+
			"Expected format: https://domain.com/path. "+
			"Example valid URLs: 'https://example.com/webhooks', 'https://api.example.com/events'. "+
			"Please provide a valid HTTPS URL.", url)
	}
	return nil
}

// ValidateTriggerType validates that a triggerType is a recognized Webflow event.
// Returns actionable error messages listing all valid trigger types.
func ValidateTriggerType(triggerType string) error {
	if triggerType == "" {
		return errors.New("triggerType is required but was not provided. " +
			"Please specify which Webflow event should trigger this webhook. " +
			"Valid trigger types: form_submission, site_publish, page_created, page_metadata_updated, " +
			"page_deleted, ecomm_new_order, ecomm_order_changed, ecomm_inventory_changed, " +
			"memberships_user_account_added, memberships_user_account_updated, memberships_user_account_deleted, " +
			"collection_item_created, collection_item_changed, collection_item_deleted, collection_item_unpublished. " +
			"Example: 'form_submission' for form submissions, 'site_publish' for site publishes.")
	}
	if !validTriggerTypes[triggerType] {
		return fmt.Errorf("triggerType '%s' is not a valid Webflow event type. "+
			"Valid trigger types are: form_submission, site_publish, page_created, page_metadata_updated, "+
			"page_deleted, ecomm_new_order, ecomm_order_changed, ecomm_inventory_changed, "+
			"memberships_user_account_added, memberships_user_account_updated, memberships_user_account_deleted, "+
			"collection_item_created, collection_item_changed, collection_item_deleted, collection_item_unpublished. "+
			"Please use one of these valid trigger types. "+
			"Example: 'form_submission' for form submissions, 'site_publish' for site publishes.", triggerType)
	}
	return nil
}

// GenerateWebhookResourceID generates a Pulumi resource ID for a Webhook resource.
// Format: {siteID}/webhooks/{webhookID}
func GenerateWebhookResourceID(siteID, webhookID string) string {
	return fmt.Sprintf("%s/webhooks/%s", siteID, webhookID)
}

// ExtractIDsFromWebhookResourceID extracts the siteID and webhookID from a Webhook resource ID.
// Expected format: {siteID}/webhooks/{webhookID}
func ExtractIDsFromWebhookResourceID(resourceID string) (siteID, webhookID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) < 3 || parts[1] != "webhooks" {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/webhooks/{webhookId}, got: %s", resourceID)
	}

	siteID = parts[0]
	webhookID = strings.Join(parts[2:], "/") // Handle webhookID that might contain slashes

	return siteID, webhookID, nil
}

// getWebhooksBaseURL is used internally for testing to override the API base URL.
var getWebhooksBaseURL = ""

// GetWebhooks retrieves all webhooks for a Webflow site.
// It calls GET /v2/sites/{site_id}/webhooks endpoint.
// Returns the parsed response or an error if the request fails.
func GetWebhooks(ctx context.Context, client *http.Client, siteID string) (*WebhooksListResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getWebhooksBaseURL != "" {
		baseURL = getWebhooksBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/webhooks", baseURL, siteID)

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

		var response WebhooksListResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getWebhookBaseURL is used internally for testing to override the API base URL.
var getWebhookBaseURL = ""

// GetWebhook retrieves a single webhook by ID from Webflow.
// It calls GET /v2/webhooks/{webhook_id} endpoint.
// Returns the webhook or an error if the request fails.
func GetWebhook(ctx context.Context, client *http.Client, webhookID string) (*WebhookResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getWebhookBaseURL != "" {
		baseURL = getWebhookBaseURL
	}

	url := fmt.Sprintf("%s/v2/webhooks/%s", baseURL, webhookID)

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

		var webhook WebhookResponse
		if err := json.Unmarshal(body, &webhook); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &webhook, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// postWebhookBaseURL is used internally for testing to override the API base URL.
var postWebhookBaseURL = ""

// PostWebhook creates a new webhook for a Webflow site.
// It calls POST /v2/sites/{site_id}/webhooks endpoint.
// Returns the created webhook or an error if the request fails.
func PostWebhook(
	ctx context.Context, client *http.Client,
	siteID, triggerType, url string, filter map[string]interface{},
) (*WebhookResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postWebhookBaseURL != "" {
		baseURL = postWebhookBaseURL
	}

	apiURL := fmt.Sprintf("%s/v2/sites/%s/webhooks", baseURL, siteID)

	requestBody := WebhookRequest{
		TriggerType: triggerType,
		URL:         url,
		Filter:      filter,
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

		req, err := http.NewRequestWithContext(ctx, "POST", apiURL, bytes.NewReader(bodyBytes))
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

		var webhook WebhookResponse
		if err := json.Unmarshal(body, &webhook); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &webhook, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// deleteWebhookBaseURL is used internally for testing to override the API base URL.
var deleteWebhookBaseURL = ""

// DeleteWebhook removes a webhook from Webflow.
// It calls DELETE /v2/webhooks/{webhook_id} endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteWebhook(ctx context.Context, client *http.Client, webhookID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteWebhookBaseURL != "" {
		baseURL = deleteWebhookBaseURL
	}

	url := fmt.Sprintf("%s/v2/webhooks/%s", baseURL, webhookID)

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
