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
	"time"
)

// Site represents a Webflow site configuration.
// This struct maps to the Webflow API v2 Site object.
type Site struct {
	// ID is the unique identifier for the site (read-only).
	ID string `json:"id,omitempty"`
	// WorkspaceID is the workspace that contains this site (read-only).
	WorkspaceID string `json:"workspaceId,omitempty"`
	// DisplayName is the human-readable name of the site.
	DisplayName string `json:"displayName"`
	// ShortName is the slugified version of the site name (lowercase alphanumeric with hyphens).
	ShortName string `json:"shortName,omitempty"`
	// TimeZone is the IANA timezone identifier for the site.
	TimeZone string `json:"timeZone,omitempty"`
	// LastPublished is the timestamp of the last site publish (read-only).
	LastPublished string `json:"lastPublished,omitempty"`
	// LastUpdated is the timestamp of the last site update (read-only).
	LastUpdated string `json:"lastUpdated,omitempty"`
	// PreviewURL is the URL to a preview image of the site (read-only).
	PreviewURL string `json:"previewUrl,omitempty"`
	// ParentFolderID is the folder where the site is organized (optional).
	ParentFolderID string `json:"parentFolderId,omitempty"`
	// CustomDomains is the list of custom domains attached to the site (read-only for now).
	CustomDomains []string `json:"customDomains,omitempty"`
	// DataCollectionEnabled indicates if data collection is enabled for the site (read-only).
	DataCollectionEnabled bool `json:"dataCollectionEnabled,omitempty"`
	// DataCollectionType is the type of data collection enabled (read-only).
	DataCollectionType string `json:"dataCollectionType,omitempty"`
}

// SiteResponse represents the API response structure for site list/get operations.
type SiteResponse struct {
	Sites []Site `json:"sites,omitempty"`
}

// SiteCreateRequest represents the request body for creating a new site.
// Note: The Webflow API uses "name" in the request, but returns "displayName" in the response.
type SiteCreateRequest struct {
	// Name is the name of the site (maps to displayName in response).
	Name string `json:"name"`
	// TemplateName is the optional template to use for site creation.
	TemplateName string `json:"templateName,omitempty"`
	// ParentFolderID is the optional folder where the site will be organized.
	ParentFolderID string `json:"parentFolderId,omitempty"`
}

// SiteUpdateRequest represents the request body for updating a site.
// The Webflow PATCH API accepts "name" (not "displayName") and "parentFolderId".
// Note: shortName is read-only (auto-generated from name) and cannot be set via API.
// Note: TimeZone is read-only and cannot be updated via API.
type SiteUpdateRequest struct {
	Name           string `json:"name,omitempty"`
	ParentFolderID string `json:"parentFolderId,omitempty"`
}

// SitePublishRequest represents the request body for publishing a site.
// All fields are optional - if domains not specified, publishes to all configured domains.
type SitePublishRequest struct {
	Domains []string `json:"domains,omitempty"`
}

// SitePublishResponse represents the API response from publishing a site.
type SitePublishResponse struct {
	Published bool   `json:"published,omitempty"`
	Queued    bool   `json:"queued,omitempty"`
	Message   string `json:"message,omitempty"`
}

// ValidateDisplayName validates that displayName meets Webflow requirements.
// Actionable error messages explain: what's wrong, expected format, and how to fix it.
func ValidateDisplayName(displayName string) error {
	if displayName == "" {
		return errors.New("displayName is required but was not provided. " +
			"Expected format: A non-empty string representing your site's name. " +
			"Fix: Provide a name for your site (e.g., 'My Marketing Site', 'Company Blog', 'Product Landing Page')")
	}

	// Webflow site names typically have a practical length limit
	if len(displayName) > 255 {
		return fmt.Errorf("displayName is too long: '%s' exceeds maximum length of 255 characters. "+
			"Expected format: A string with 1-255 characters. "+
			"Fix: Use a shorter, more concise site name", displayName)
	}

	return nil
}

// ValidateShortName validates that shortName meets Webflow's slug requirements.
// Webflow's shortName must be lowercase alphanumeric with hyphens, no leading/trailing hyphens.
// If shortName is empty, that's OK - Webflow will generate one from displayName.
// Actionable error messages explain: what's wrong, expected format, and how to fix it.
func ValidateShortName(shortName string) error {
	// shortName is optional - if empty, Webflow will auto-generate from displayName
	if shortName == "" {
		return nil
	}

	// Webflow shortName must be lowercase alphanumeric with hyphens only
	// Pattern: start with letter/number, can have hyphens in middle, end with letter/number
	shortNameRegex := regexp.MustCompile(`^[a-z0-9]+(-[a-z0-9]+)*$`)
	if !shortNameRegex.MatchString(shortName) {
		return fmt.Errorf("invalid shortName format: '%s' contains invalid characters. "+
			"Expected format: lowercase letters (a-z), numbers (0-9), and hyphens (-) only. "+
			"Must start and end with a letter or number (e.g., 'my-site', 'company-blog-2024', 'product-1'). "+
			"Fix: Use only lowercase letters, numbers, and hyphens. No spaces, underscores, or special characters. "+
			"No leading/trailing hyphens", shortName)
	}

	return nil
}

// ValidateWorkspaceID validates that workspaceID is a non-empty string.
// Workspace IDs are required for site creation via the Webflow API.
// Actionable error messages explain: what's wrong, expected format, and how to fix it.
func ValidateWorkspaceID(workspaceID string) error {
	if workspaceID == "" {
		return errors.New("workspaceId is required but was not provided. " +
			"Expected format: Your Webflow workspace ID (a 24-character hexadecimal string). " +
			"Fix: Provide your workspace ID. You can find it in your Webflow dashboard under Account Settings > Workspace. " +
			"Note: Creating sites via API requires an Enterprise workspace")
	}

	return nil
}

// postSiteBaseURL is used internally for testing to override the API base URL.
var postSiteBaseURL = ""

// patchSiteBaseURL is used internally for testing to override the API base URL.
var patchSiteBaseURL = ""

// publishSiteBaseURL is used internally for testing to override the API base URL.
var publishSiteBaseURL = ""

// deleteSiteBaseURL is used internally for testing to override the API base URL.
var deleteSiteBaseURL = ""

// getSiteBaseURL is used internally for testing to override the API base URL.
var getSiteBaseURL = ""

// PostSite creates a new site in the specified Webflow workspace.
// Enterprise workspace is required for site creation via API.
// Note: API request uses "name" but response returns "displayName".
// Returns the created Site or an error if the request fails.
func PostSite(
	ctx context.Context, client *http.Client,
	workspaceID, displayName, parentFolderID, templateName string,
) (*Site, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if postSiteBaseURL != "" {
		baseURL = postSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/workspaces/%s/sites", baseURL, workspaceID)

	// Map displayName → name for API request
	requestBody := SiteCreateRequest{
		Name:           displayName,
		TemplateName:   templateName,   // Optional, empty string OK
		ParentFolderID: parentFolderID, // Optional, empty string OK
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
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

		// Accept both 200 and 201 as success
		if resp.StatusCode != 200 && resp.StatusCode != 201 {
			return nil, handleSiteError(resp.StatusCode, body)
		}

		var site Site
		if err := json.Unmarshal(body, &site); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &site, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// PatchSite updates an existing site's configuration.
// Only changed fields should be sent in the request to minimize API payload.
// Note: shortName is read-only (auto-generated from name) and cannot be set via API.
// Note: TimeZone is read-only and cannot be updated via API.
// Returns the updated Site or an error if the request fails.
func PatchSite(
	ctx context.Context, client *http.Client,
	siteID, displayName, parentFolderID string,
) (*Site, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if patchSiteBaseURL != "" {
		baseURL = patchSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s", baseURL, siteID)

	// Build request with provided fields (empty strings = not changed)
	// Note: Webflow API accepts "name" (not "displayName") for PATCH
	// Note: shortName is read-only and cannot be set via API
	// Note: TimeZone is read-only and cannot be updated via API
	requestBody := SiteUpdateRequest{
		Name:           displayName,
		ParentFolderID: parentFolderID,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
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
			return nil, handleSiteError(resp.StatusCode, body)
		}

		var site Site
		if err := json.Unmarshal(body, &site); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &site, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// PublishSite publishes a site to production, making it live on configured domains.
// This operation is asynchronous - the API returns immediately with job status.
// The actual publish completion happens asynchronously and can be monitored via
// subsequent Read operations that check the lastPublished timestamp.
// Returns publish status or an error if the request fails.
func PublishSite(
	ctx context.Context, client *http.Client, siteID string, domains []string,
) (*SitePublishResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if publishSiteBaseURL != "" {
		baseURL = publishSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s/publish", baseURL, siteID)

	// Build request with optional domains
	requestBody := SitePublishRequest{
		Domains: domains,
	}

	bodyBytes, err := json.Marshal(requestBody)
	if err != nil {
		return nil, fmt.Errorf("failed to marshal request body: %w", err)
	}

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
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
		// Accept both 200 and 202 (Accepted) as success for async publish
		if resp.StatusCode != 200 && resp.StatusCode != 202 {
			return nil, handleSiteError(resp.StatusCode, body)
		}

		var publishResp SitePublishResponse
		if err := json.Unmarshal(body, &publishResp); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &publishResp, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DeleteSite permanently deletes a site from Webflow.
// This operation cannot be undone - the site and all its content will be permanently removed.
// Returns nil on success (204 No Content), or an error if the request fails.
// Note: 404 responses are treated as success (idempotent - site already deleted).
func DeleteSite(ctx context.Context, client *http.Client, siteID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if deleteSiteBaseURL != "" {
		baseURL = deleteSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
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
					return fmt.Errorf("context cancelled during retry: %w", ctx.Err())
				case <-time.After(waitTime):
				}
			}
			continue
		}

		// Handle 404 as success (idempotent deletion)
		if resp.StatusCode == 404 {
			// Site doesn't exist - deletion already complete
			return nil
		}

		// Handle error responses
		// 204 No Content is the success status for deletion
		if resp.StatusCode != 204 {
			return handleSiteError(resp.StatusCode, body)
		}

		// Success - 204 No Content
		return nil
	}

	return fmt.Errorf("max retries exceeded: %w", lastErr)
}

// GetSite retrieves the current state of a site from Webflow.
// Returns the site data if successful, or an error if the request fails.
// Note: Returns nil, nil (not an error) when site is not found (404) - caller handles appropriately.
func GetSite(ctx context.Context, client *http.Client, siteID string) (*Site, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getSiteBaseURL != "" {
		baseURL = getSiteBaseURL
	}

	url := fmt.Sprintf("%s/v2/sites/%s", baseURL, siteID)

	var lastErr error
	for attempt := 0; attempt <= maxRetries; attempt++ {
		// Exponential backoff on retry
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

		// Handle 404 Not Found - site was deleted externally
		// This is NOT an error in the context of Read - caller will handle appropriately
		if resp.StatusCode == 404 {
			return nil, nil // Return nil, nil to signal "site not found"
		}

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleSiteError(resp.StatusCode, body)
		}

		// Parse successful response (200 OK)
		var siteData Site
		if err := json.Unmarshal(body, &siteData); err != nil {
			return nil, fmt.Errorf("failed to parse site response: %w", err)
		}

		// Success - return site data
		return &siteData, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// handleSiteError provides actionable error messages for Site API operations.
// Error messages explain what went wrong, why it happened, and how to fix it.
func handleSiteError(statusCode int, body []byte) error {
	switch statusCode {
	case 400:
		return fmt.Errorf("bad request: the site creation/update request was "+
			"incorrectly formatted. Details: %s. "+
			"Please check your site configuration and ensure all required fields "+
			"(workspaceId, displayName) are provided with valid values. "+
			"Verify that optional fields (shortName, timeZone, templateName) "+
			"follow the correct format", string(body))
	case 401:
		return errors.New("unauthorized: authentication failed. " +
			"Your Webflow API token is invalid or has expired. " +
			"To fix this: 1) Verify your token in the Webflow dashboard " +
			"(Settings > Integrations > API Access), " +
			"2) Ensure the token has the required scopes for site management, " +
			"3) Update your Pulumi config with: " +
			"'pulumi config set webflow:apiToken <your-token> --secret'")
	case 403:
		return errors.New("forbidden: access denied. " +
			"Your API token does not have permission to create/manage sites, " +
			"OR your workspace is not an Enterprise workspace. " +
			"To fix this: 1) Verify you have an Enterprise Webflow workspace " +
			"(site creation via API requires Enterprise), " +
			"2) Ensure your API token has the required scopes for site management, " +
			"3) Check that the workspace ID is correct and belongs to your account")
	case 404:
		return errors.New("not found: the Webflow site, workspace, or template " +
			"does not exist. " +
			"To fix this: 1) Verify the workspace ID is correct " +
			"(24-character lowercase hex string), " +
			"2) If using templateName, verify the template name is valid in your " +
			"Webflow account (check available templates in Webflow dashboard), " +
			"3) Check that the workspace exists in your Webflow dashboard, " +
			"4) Try creating without templateName to isolate the issue")
	case 429:
		return errors.New("rate limited: too many requests to Webflow API. " +
			"The provider will automatically retry with exponential backoff. " +
			"If this error persists, please wait a few minutes before trying again. " +
			"Consider reducing the frequency of operations or contact Webflow support if rate limits are consistently exceeded")
	case 500:
		return fmt.Errorf("server error: Webflow API encountered an internal error. "+
			"Details: %s. "+
			"This is a temporary issue on Webflow's side. "+
			"The provider will automatically retry. If the error persists, please contact Webflow support", string(body))
	default:
		return fmt.Errorf("unexpected error (HTTP %d): %s. "+
			"This may indicate a temporary issue with the Webflow API or an unhandled error condition. "+
			"Please check the Webflow API status and try again", statusCode, string(body))
	}
}
