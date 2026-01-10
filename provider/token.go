// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"time"
)

// AuthorizedTo represents the authorization scope for a token.
type AuthorizedTo struct {
	SiteIDs      []string `json:"siteIds,omitempty"`
	WorkspaceIDs []string `json:"workspaceIds,omitempty"`
	UserIDs      []string `json:"userIds,omitempty"`
}

// Authorization represents the token authorization details.
type Authorization struct {
	ID           string       `json:"id"`
	CreatedOn    string       `json:"createdOn,omitempty"`
	LastUsed     string       `json:"lastUsed,omitempty"`
	GrantType    string       `json:"grantType,omitempty"`
	RateLimit    int          `json:"rateLimit,omitempty"`
	Scope        string       `json:"scope,omitempty"`
	AuthorizedTo AuthorizedTo `json:"authorizedTo,omitempty"`
}

// Application represents the application details for a token.
type Application struct {
	ID          string `json:"id"`
	Description string `json:"description,omitempty"`
	Homepage    string `json:"homepage,omitempty"`
	DisplayName string `json:"displayName,omitempty"`
}

// TokenIntrospectResponse represents the response from GET /token/introspect.
type TokenIntrospectResponse struct {
	Authorization Authorization `json:"authorization"`
	Application   Application   `json:"application,omitempty"`
}

// AuthorizedByResponse represents the response from GET /token/authorized_by.
type AuthorizedByResponse struct {
	ID        string `json:"id"`
	Email     string `json:"email"`
	FirstName string `json:"firstName,omitempty"`
	LastName  string `json:"lastName,omitempty"`
}

// getTokenIntrospectBaseURL is used internally for testing to override the API base URL.
var getTokenIntrospectBaseURL = ""

// GetTokenIntrospect retrieves token authorization information.
// It calls GET /token/introspect endpoint.
// Returns the parsed response or an error if the request fails.
func GetTokenIntrospect(ctx context.Context, client *http.Client) (*TokenIntrospectResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getTokenIntrospectBaseURL != "" {
		baseURL = getTokenIntrospectBaseURL
	}

	url := baseURL + "/v2/token/introspect"

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
			return nil, handleTokenError(resp.StatusCode, body)
		}

		var response TokenIntrospectResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// getAuthorizedByBaseURL is used internally for testing to override the API base URL.
var getAuthorizedByBaseURL = ""

// GetAuthorizedBy retrieves information about the user who authorized the token.
// It calls GET /token/authorized_by endpoint.
// Returns the parsed response or an error if the request fails.
func GetAuthorizedBy(ctx context.Context, client *http.Client) (*AuthorizedByResponse, error) {
	if err := ctx.Err(); err != nil {
		return nil, fmt.Errorf("context cancelled: %w", err)
	}

	baseURL := webflowAPIBaseURL
	if getAuthorizedByBaseURL != "" {
		baseURL = getAuthorizedByBaseURL
	}

	url := baseURL + "/v2/token/authorized_by"

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
			return nil, handleTokenError(resp.StatusCode, body)
		}

		var response AuthorizedByResponse
		if err := json.Unmarshal(body, &response); err != nil {
			return nil, fmt.Errorf("failed to parse response: %w", err)
		}

		return &response, nil
	}

	return nil, fmt.Errorf("max retries exceeded: %w", lastErr)
}

// handleTokenError converts HTTP error responses to actionable error messages for token endpoints.
func handleTokenError(statusCode int, body []byte) error {
	switch statusCode {
	case 401:
		return fmt.Errorf("unauthorized: authentication failed (HTTP 401). "+
			"Your Webflow API token is invalid or has expired. "+
			"To fix this: 1) Verify your token in the Webflow dashboard (Settings > Integrations > API Access), "+
			"2) Ensure the token is valid and not expired, "+
			"3) Update your Pulumi config with: 'pulumi config set webflow:apiToken <your-token> --secret'. "+
			"Details: %s", string(body))
	case 403:
		return fmt.Errorf("forbidden: access denied (HTTP 403). "+
			"Your API token does not have the required permissions. "+
			"To fix this: Ensure your API token has the 'authorized_user:read' scope "+
			"for the authorized_by endpoint. Details: %s", string(body))
	case 404:
		return fmt.Errorf("not found: the requested endpoint does not exist (HTTP 404). "+
			"This may indicate an API version mismatch. Details: %s", string(body))
	case 500:
		return fmt.Errorf("server error: Webflow API encountered an internal error (HTTP 500). "+
			"This is a temporary issue on Webflow's side. "+
			"Please wait a few minutes and try again. "+
			"If the problem persists, check Webflow's status page or contact Webflow support. "+
			"Details: %s", string(body))
	default:
		return fmt.Errorf("unexpected error (HTTP %d): %s. "+
			"This is an unexpected response from the Webflow API. "+
			"Please check Webflow's status page or contact Webflow support if this error persists",
			statusCode, string(body))
	}
}
