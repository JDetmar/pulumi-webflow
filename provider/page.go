// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"io"
	"net/http"
	"regexp"
	"time"
)

// Page represents a single page in a Webflow site.
// This struct matches the Webflow API v2 response format for pages.
type Page struct {
	// ID is the unique identifier for the page (24-character hex string).
	ID string `json:"id"`
	// SiteID is the Webflow site ID this page belongs to.
	SiteID string `json:"siteId"`
	// Title is the page title (appears in browser tabs and search results).
	Title string `json:"title"`
	// Slug is the URL slug for the page (e.g., "about" for "/about").
	Slug string `json:"slug"`
	// ParentID is the ID of the parent page (for nested pages, optional).
	ParentID string `json:"parentId,omitempty"`
	// CollectionID is the ID of the CMS collection (for collection pages, optional).
	CollectionID string `json:"collectionId,omitempty"`
	// CreatedOn is the timestamp when the page was created.
	CreatedOn string `json:"createdOn"`
	// LastUpdated is the timestamp when the page was last updated.
	LastUpdated string `json:"lastUpdated"`
	// Archived indicates if the page is archived.
	Archived bool `json:"archived"`
	// Draft indicates if the page is in draft mode.
	Draft bool `json:"draft"`
	// CanBranch indicates if the page can be branched (read-only).
	CanBranch bool `json:"canBranch,omitempty"`
	// Locales contains locale information for the page (optional).
	Locales *PageLocales `json:"locales,omitempty"`
}

// PageLocales represents locale information for a page.
type PageLocales struct {
	// Primary is the primary locale for the page.
	Primary string `json:"primary,omitempty"`
	// Secondary is a list of secondary locales.
	Secondary []string `json:"secondary,omitempty"`
}

// PagesResponse represents the Webflow API response for GET /sites/{site_id}/pages.
type PagesResponse struct {
	// Pages is the list of pages in the site.
	Pages []Page `json:"pages"`
}

// pageIDPattern is the regex pattern for validating Webflow page IDs.
// Page IDs are 24-character lowercase hexadecimal strings (same format as site IDs).
var pageIDPattern = regexp.MustCompile(`^[a-f0-9]{24}$`)

// ValidatePageID validates that a pageID matches the Webflow page ID format.
// Webflow page IDs are 24-character lowercase hexadecimal strings.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidatePageID(pageID string) error {
	if pageID == "" {
		return errors.New("pageId is required but was not provided. " +
			"Please provide a valid Webflow page ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c4'). " +
			"You can find page IDs using the Pages API list endpoint or in the Webflow designer")
	}
	if !pageIDPattern.MatchString(pageID) {
		return fmt.Errorf("pageId has invalid format: got '%s'. "+
			"Expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c4'). "+
			"Please check your page ID and ensure it contains only "+
			"lowercase letters (a-f) and digits (0-9)", pageID)
	}
	return nil
}

// GeneratePageResourceID generates a Pulumi resource ID for a Page data source.
// Format: {siteID}/pages/{pageID}
func GeneratePageResourceID(siteID, pageID string) string {
	return fmt.Sprintf("%s/pages/%s", siteID, pageID)
}

// ExtractIDsFromPageResourceID extracts the siteID and pageID from a Page resource ID.
// Expected format: {siteID}/pages/{pageID}
func ExtractIDsFromPageResourceID(resourceID string) (siteID, pageID string, err error) {
	if resourceID == "" {
		return "", "", errors.New("resourceId cannot be empty")
	}

	// Simple split-based parsing
	// Format: {siteID}/pages/{pageID}
	// Find "/pages/" marker
	pagesMarker := "/pages/"
	idx := len(resourceID)
	for i := 0; i < len(resourceID)-len(pagesMarker)+1; i++ {
		if resourceID[i:i+len(pagesMarker)] == pagesMarker {
			idx = i
			break
		}
	}

	if idx >= len(resourceID) {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/pages/{pageId}, got: %s", resourceID)
	}

	siteID = resourceID[:idx]
	pageID = resourceID[idx+len(pagesMarker):]

	if siteID == "" || pageID == "" {
		return "", "", fmt.Errorf("invalid resource ID format: expected {siteId}/pages/{pageId}, got: %s", resourceID)
	}

	return siteID, pageID, nil
}

// getPagesBaseURL is used internally for testing to override the API base URL.
var getPagesBaseURL = ""

// GetPages retrieves all pages for a Webflow site.
// It calls GET /v2/sites/{site_id}/pages endpoint.
// Returns the parsed response or an error if the request fails.
func GetPages(ctx context.Context, client *http.Client, siteID string) (*PagesResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getPagesBaseURL != "" {
		baseURL = getPagesBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/pages", baseURL, siteID)

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
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
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

		var response PagesResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getPageBaseURL is used internally for testing to override the API base URL.
var getPageBaseURL = ""

// GetPage retrieves a single page by ID from Webflow.
// It calls GET /v2/pages/{page_id} endpoint.
// Returns the parsed page or an error if the request fails.
func GetPage(ctx context.Context, client *http.Client, pageID string) (*Page, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getPageBaseURL != "" {
		baseURL = getPageBaseURL
	}

	url := fmt.Sprintf("%s/v2/pages/%s", baseURL, pageID)

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

		var page Page
		if err := json.Unmarshal(body, &page); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &page, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}
