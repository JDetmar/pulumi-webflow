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
	"strconv"
	"strings"
	"time"
)

// RobotsTxtRule represents a single user-agent rule in robots.txt.
// This struct matches the Webflow API v2 response format for robots.txt rules.
type RobotsTxtRule struct {
	UserAgent string   `json:"userAgent"` // The user-agent this rule applies to (e.g., "*", "Googlebot")
	Allows    []string `json:"allows"`    // Paths that are allowed for this user-agent
	Disallows []string `json:"disallows"` // Paths that are disallowed for this user-agent
}

// RobotsTxtResponse represents the Webflow API response for robots.txt.
type RobotsTxtResponse struct {
	Rules   []RobotsTxtRule `json:"rules"`   // List of user-agent rules
	Sitemap string          `json:"sitemap"` // URL to the sitemap
}

// RobotsTxtRequest represents the request body for PUT/PATCH robots.txt.
type RobotsTxtRequest struct {
	Rules   []RobotsTxtRule `json:"rules,omitempty"`   // List of user-agent rules
	Sitemap string          `json:"sitemap,omitempty"` // URL to the sitemap
}

// siteIDPattern is the regex pattern for validating Webflow site IDs.
// Site IDs are 24-character lowercase hexadecimal strings.
var siteIDPattern = regexp.MustCompile(`^[a-f0-9]{24}$`)

// ValidateSiteID validates that a siteID matches the Webflow site ID format.
// Webflow site IDs are 24-character lowercase hexadecimal strings.
// During Pulumi preview, placeholder IDs (starting with "preview-") are allowed.
// Returns actionable error messages that explain what's wrong and how to fix it.
func ValidateSiteID(siteID string) error {
	if siteID == "" {
		return errors.New("siteId is required but was not provided. " +
			"Please provide a valid Webflow site ID " +
			"(24-character lowercase hexadecimal string, e.g., '5f0c8c9e1c9d440000e8d8c3'). " +
			"You can find your site ID in the Webflow dashboard under Site Settings")
	}
	// During Pulumi preview, dependent resources receive placeholder IDs like "preview-1234567890"
	// These must be allowed to pass validation since the real ID isn't known yet
	if strings.HasPrefix(siteID, "preview-") {
		return nil
	}
	if !siteIDPattern.MatchString(siteID) {
		return fmt.Errorf("siteId has invalid format: got '%s'. "+
			"Expected a 24-character lowercase hexadecimal string "+
			"(e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"Please check your site ID in the Webflow dashboard "+
			"and ensure it contains only lowercase letters (a-f) and digits (0-9)", siteID)
	}
	return nil
}

// GenerateRobotsTxtResourceID generates a Pulumi resource ID for a RobotsTxt resource.
// Format: {siteID}/robots.txt
func GenerateRobotsTxtResourceID(siteID string) string {
	return siteID + "/robots.txt"
}

// ExtractSiteIDFromResourceID extracts the siteID from a RobotsTxt resource ID.
// Expected format: {siteID}/robots.txt
func ExtractSiteIDFromResourceID(resourceID string) (string, error) {
	if resourceID == "" {
		return "", errors.New("resourceId cannot be empty")
	}

	parts := strings.Split(resourceID, "/")
	if len(parts) != 2 || parts[1] != "robots.txt" {
		return "", fmt.Errorf("invalid resource ID format: expected {siteId}/robots.txt, got: %s", resourceID)
	}

	return parts[0], nil
}

// ParseRobotsTxtContent parses a robots.txt content string into structured rules and sitemap.
// This converts the traditional robots.txt format into the Webflow API format.
//
// Example input:
//
//	User-agent: *
//	Allow: /
//	Disallow: /admin/
//	Sitemap: https://example.com/sitemap.xml
//
// Returns the parsed rules and sitemap URL.
func ParseRobotsTxtContent(content string) (rules []RobotsTxtRule, sitemap string) {
	if content == "" {
		return []RobotsTxtRule{}, ""
	}

	var currentRule *RobotsTxtRule

	lines := strings.Split(content, "\n")
	for _, line := range lines {
		line = strings.TrimSpace(line)
		if line == "" {
			continue
		}

		// Check for sitemap directive (case-insensitive)
		if strings.HasPrefix(strings.ToLower(line), "sitemap:") {
			sitemap = strings.TrimSpace(line[8:])
			continue
		}

		// Check for user-agent directive (case-insensitive)
		if strings.HasPrefix(strings.ToLower(line), "user-agent:") {
			// Save previous rule if exists
			if currentRule != nil {
				rules = append(rules, *currentRule)
			}
			// Start new rule
			userAgent := strings.TrimSpace(line[11:])
			currentRule = &RobotsTxtRule{
				UserAgent: userAgent,
				Allows:    []string{},
				Disallows: []string{},
			}
			continue
		}

		// Parse Allow/Disallow directives
		if currentRule != nil {
			if strings.HasPrefix(strings.ToLower(line), "allow:") {
				path := strings.TrimSpace(line[6:])
				if path != "" {
					currentRule.Allows = append(currentRule.Allows, path)
				}
			} else if strings.HasPrefix(strings.ToLower(line), "disallow:") {
				path := strings.TrimSpace(line[9:])
				if path != "" {
					currentRule.Disallows = append(currentRule.Disallows, path)
				}
			}
		}
	}

	// Don't forget the last rule
	if currentRule != nil {
		rules = append(rules, *currentRule)
	}

	return rules, sitemap
}

// FormatRobotsTxtContent formats structured rules and sitemap into a robots.txt content string.
// This converts the Webflow API format back to traditional robots.txt format.
func FormatRobotsTxtContent(rules []RobotsTxtRule, sitemap string) string {
	if len(rules) == 0 && sitemap == "" {
		return ""
	}

	var builder strings.Builder

	for _, rule := range rules {
		builder.WriteString(fmt.Sprintf("User-agent: %s\n", rule.UserAgent))

		for _, allow := range rule.Allows {
			builder.WriteString(fmt.Sprintf("Allow: %s\n", allow))
		}

		for _, disallow := range rule.Disallows {
			builder.WriteString(fmt.Sprintf("Disallow: %s\n", disallow))
		}
	}

	if sitemap != "" {
		if len(rules) > 0 {
			builder.WriteString("\n")
		}
		builder.WriteString(fmt.Sprintf("Sitemap: %s\n", sitemap))
	}

	return builder.String()
}

// webflowAPIBaseURL is the base URL for Webflow API v2.
const webflowAPIBaseURL = "https://api.webflow.com"

// maxRetries is the maximum number of retry attempts for rate-limited requests.
const maxRetries = 3

// getRetryAfterDuration parses the Retry-After header and returns the backoff duration.
// The header can be either a number of seconds or an HTTP date.
// Returns a default backoff if the header is invalid or not present.
func getRetryAfterDuration(retryAfter string, defaultBackoff time.Duration) time.Duration {
	if retryAfter == "" {
		return defaultBackoff
	}

	// Try parsing as seconds (most common)
	if seconds, err := strconv.Atoi(retryAfter); err == nil && seconds > 0 {
		return time.Duration(seconds) * time.Second
	}

	// If parsing fails, use the default backoff
	return defaultBackoff
}

// handleNetworkError converts network errors to actionable error messages with recovery guidance.
// Returns different messages for timeout vs connection failures vs generic network issues.
func handleNetworkError(err error) error {
	switch {
	case strings.Contains(err.Error(), "timeout") || strings.Contains(err.Error(), "deadline exceeded"):
		return fmt.Errorf("network timeout: the request to Webflow API timed out after 30 seconds. "+
			"This may indicate network connectivity issues or Webflow API is slow to respond. "+
			"To fix this: 1) Check your internet connection, 2) Verify Webflow API status, "+
			"3) Wait a few minutes and retry, 4) If the problem persists, contact Webflow support: %w", err)
	case strings.Contains(err.Error(), "connection refused") || strings.Contains(err.Error(), "no such host"):
		return fmt.Errorf("connection failed: unable to connect to Webflow API. "+
			"This may indicate network connectivity issues or DNS problems. "+
			"To fix this: 1) Check your internet connection, 2) Verify DNS resolution (try: nslookup api.webflow.com), "+
			"3) Check firewall/proxy settings, 4) If using a VPN, try disconnecting: %w", err)
	default:
		return fmt.Errorf("network error: request to Webflow API failed. "+
			"This may indicate network connectivity issues. "+
			"To fix this: 1) Check your internet connection, 2) Verify Webflow API is accessible, "+
			"3) Wait a few minutes and retry, 4) If the problem persists, check network logs: %w", err)
	}
}

// GetRobotsTxt retrieves the robots.txt configuration for a Webflow site.
// It calls GET /v2/sites/{site_id}/robots_txt endpoint.
// Returns the parsed response or an error if the request fails.
func GetRobotsTxt(ctx context.Context, client *http.Client, siteID string) (*RobotsTxtResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	url := fmt.Sprintf("%s/v2/sites/%s/robots_txt", webflowAPIBaseURL, siteID)

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

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response RobotsTxtResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// PutRobotsTxt creates or updates the robots.txt configuration for a Webflow site.
// It calls PUT /v2/sites/{site_id}/robots_txt endpoint.
// Returns the updated response or an error if the request fails.
func PutRobotsTxt(
	ctx context.Context, client *http.Client,
	siteID string, rules []RobotsTxtRule, sitemap string,
) (*RobotsTxtResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	url := fmt.Sprintf("%s/v2/sites/%s/robots_txt", webflowAPIBaseURL, siteID)

	requestBody := RobotsTxtRequest{
		Rules:   rules,
		Sitemap: sitemap,
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

		// Handle error responses
		if resp.StatusCode != 200 {
			return nil, handleWebflowError(resp.StatusCode, body)
		}

		var response RobotsTxtResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// DeleteRobotsTxt removes the robots.txt configuration from a Webflow site.
// It calls DELETE /v2/sites/{site_id}/robots_txt endpoint.
// Returns nil on success (including 404 for idempotency) or an error if the request fails.
func DeleteRobotsTxt(ctx context.Context, client *http.Client, siteID string) error {
	if err := ctx.Err(); err != nil {
		return fmt.Errorf("context cancelled: %w", err)
	}

	url := fmt.Sprintf("%s/v2/sites/%s/robots_txt", webflowAPIBaseURL, siteID)

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
			lastErr = fmt.Errorf("rate limited: Webflow API rate limit exceeded (HTTP 429). "+
				"The provider will automatically retry with exponential backoff. "+
				"Retry attempt %d of %d, waiting %v before next attempt. "+
				"If this error persists, please wait a few minutes before trying again or contact Webflow support",
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

// handleWebflowError converts HTTP error responses to actionable error messages.
// Error messages follow Pulumi diagnostic formatting and include actionable guidance.
func handleWebflowError(statusCode int, body []byte) error {
	switch statusCode {
	case 400:
		return fmt.Errorf("bad request: the request to Webflow API was incorrectly formatted. "+
			"Details: %s. "+
			"Please check your resource configuration and ensure all required fields are provided with valid values. "+
			"If the error persists, verify your robots.txt content follows the correct format", string(body))
	case 401:
		return errors.New("unauthorized: authentication failed. " +
			"Your Webflow API token is invalid or has expired. " +
			"To fix this: 1) Verify your token in the Webflow dashboard (Settings > Integrations > API Access), " +
			"2) Ensure the token has 'site_config:read' and 'site_config:write' scopes, " +
			"3) Update your Pulumi config with: 'pulumi config set webflow:apiToken <your-token> --secret'")
	case 403:
		return errors.New("forbidden: access denied to this resource. " +
			"Your API token does not have permission to access this Webflow site. " +
			"To fix this: 1) Verify the site ID is correct, " +
			"2) Ensure your API token has 'site_config:read' and 'site_config:write' scopes, " +
			"3) Check that the site belongs to the Webflow workspace associated with your API token")
	case 404:
		return errors.New("not found: the Webflow site or robots.txt configuration does not exist. " +
			"To fix this: 1) Verify the site ID is correct (24-character lowercase hex string), " +
			"2) Check that the site exists in your Webflow dashboard, " +
			"3) Ensure you're using the correct site ID for your Webflow workspace")
	case 429:
		return errors.New("rate limited: too many requests to Webflow API. " +
			"The provider will automatically retry with exponential backoff. " +
			"If this error persists, please wait a few minutes before trying again. " +
			"Consider reducing the frequency of operations or contact Webflow support if rate limits are consistently exceeded")
	case 500:
		return fmt.Errorf("server error: Webflow API encountered an internal error. "+
			"Details: %s. "+
			"This is a temporary issue on Webflow's side. "+
			"Please wait a few minutes and try again. "+
			"If the problem persists, check Webflow's status page or contact Webflow support", string(body))
	default:
		return fmt.Errorf("unexpected error (HTTP %d): %s. "+
			"This is an unexpected response from the Webflow API. "+
			"Please check Webflow's status page or contact Webflow support if this error persists", statusCode, string(body))
	}
}
