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

// DOMNode represents a node in the page DOM structure.
// This struct matches the Webflow API v2 response format for page DOM nodes.
type DOMNode struct {
	// NodeID is the unique identifier for this DOM node (required for updates).
	NodeID string `json:"nodeId,omitempty"`
	// Type is the node type (e.g., "text", "element", "image").
	Type string `json:"type,omitempty"`
	// Text is the text content for text nodes (updatable).
	Text string `json:"text,omitempty"`
	// Tag is the HTML tag name for element nodes (e.g., "div", "p", "h1").
	Tag string `json:"tag,omitempty"`
	// Attributes contains the node's HTML attributes (updatable for some node types).
	Attributes map[string]interface{} `json:"attributes,omitempty"`
	// Children contains child nodes (nested structure).
	Children []DOMNode `json:"children,omitempty"`
	// V is the version field used by Webflow for tracking changes.
	V string `json:"v,omitempty"`
}

// PageContentResponse represents the Webflow API response for GET /pages/{page_id}/dom.
// It contains the full DOM structure of the page.
type PageContentResponse struct {
	// PageID is the unique identifier for the page.
	PageID string `json:"pageId,omitempty"`
	// Nodes is the array of root-level DOM nodes for the page.
	Nodes []DOMNode `json:"nodes,omitempty"`
}

// PageContentRequest represents the request body for PUT /pages/{page_id}/dom.
// Used for updating static content (text and simple attributes) in the page.
type PageContentRequest struct {
	// Nodes is the array of node updates to apply.
	// Each node must have a NodeID and the fields to update (e.g., Text).
	Nodes []DOMNodeUpdate `json:"nodes,omitempty"`
}

// DOMNodeUpdate represents a single node update in the page content request.
// Only includes fields that can be updated via the API.
type DOMNodeUpdate struct {
	// NodeID is the unique identifier for the node to update (required).
	NodeID string `json:"nodeId"`
	// Text is the new text content for text nodes (optional).
	Text *string `json:"text,omitempty"`
}

// ValidateNodeID validates that a nodeID is non-empty.
// Node IDs are required for updating page content.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateNodeID(nodeID string) error {
	if nodeID == "" {
		return errors.New("nodeId is required but was not provided. " +
			"Please provide a valid node ID from the page's DOM structure. " +
			"You can retrieve node IDs by fetching the page content first using the Webflow API " +
			"GET /pages/{page_id}/dom endpoint.")
	}
	return nil
}

// GeneratePageContentResourceID generates a Pulumi resource ID for a PageContent resource.
// Format: {pageID}/content
// Note: PageContent is a 1:1 relationship with a page, so we use a simple suffix.
func GeneratePageContentResourceID(pageID string) string {
	return pageID + "/content"
}

// ExtractPageIDFromPageContentResourceID extracts the pageID from a PageContent resource ID.
// Expected format: {pageID}/content
func ExtractPageIDFromPageContentResourceID(resourceID string) (string, error) {
	if resourceID == "" {
		return "", errors.New("resourceId cannot be empty")
	}

	// Simple suffix removal
	suffix := "/content"
	if len(resourceID) <= len(suffix) {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/content, got: %s", resourceID)
	}

	if resourceID[len(resourceID)-len(suffix):] != suffix {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/content, got: %s", resourceID)
	}

	pageID := resourceID[:len(resourceID)-len(suffix)]
	if pageID == "" {
		return "", fmt.Errorf("invalid resource ID format: expected {pageId}/content, got: %s", resourceID)
	}

	return pageID, nil
}

// getPageContentBaseURL is used internally for testing to override the API base URL.
var getPageContentBaseURL = ""

// GetPageContent retrieves the DOM structure of a page from Webflow.
// It calls GET /v2/pages/{page_id}/dom endpoint.
// Returns the parsed response or an error if the request fails.
func GetPageContent(ctx context.Context, client *http.Client, pageID string) (*PageContentResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getPageContentBaseURL != "" {
		baseURL = getPageContentBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s/dom", baseURL, pageID)

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

		var response PageContentResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// putPageContentBaseURL is used internally for testing to override the API base URL.
var putPageContentBaseURL = ""

// PutPageContent updates the static content of a page in Webflow.
// It calls PUT /v2/pages/{page_id}/dom endpoint.
// Returns the updated response or an error if the request fails.
func PutPageContent(
	ctx context.Context, client *http.Client,
	pageID string, nodes []DOMNodeUpdate,
) (*PageContentResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if putPageContentBaseURL != "" {
		baseURL = putPageContentBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s/dom", baseURL, pageID)

	requestBody := PageContentRequest{
		Nodes: nodes,
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

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response PageContentResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
