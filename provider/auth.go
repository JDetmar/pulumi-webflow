// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"crypto/tls"
	"errors"
	"math"
	"net/http"
	"os"
	"strconv"
	"time"
)

// Error codes for programmatic error handling.
// Use these codes to identify specific error types in automation and scripts.
const (
	// ErrCodeAuthNotConfigured indicates the API token is missing.
	ErrCodeAuthNotConfigured = "WEBFLOW_AUTH_001"
	// ErrCodeAuthEmpty indicates an empty API token was provided.
	ErrCodeAuthEmpty = "WEBFLOW_AUTH_002"
	// ErrCodeAuthInvalid indicates the API token format is invalid.
	ErrCodeAuthInvalid = "WEBFLOW_AUTH_003"
)

// ErrTokenNotConfigured is returned when no API token is available.
var ErrTokenNotConfigured = errors.New("[" + ErrCodeAuthNotConfigured + "] Webflow API token not configured. " +
	"Configure using: pulumi config set webflow:apiToken <token> --secret " +
	"OR set WEBFLOW_API_TOKEN environment variable. " +
	"See: https://github.com/jdetmar/pulumi-webflow/blob/main/docs/troubleshooting.md#api-token-not-configured")

// getEnvToken retrieves the Webflow API token from the environment variable.
func getEnvToken() string {
	return os.Getenv("WEBFLOW_API_TOKEN")
}

// ValidateToken performs basic validation on the API token.
// Checks that the token is non-empty and has reasonable length.
func ValidateToken(token string) error {
	if token == "" {
		return errors.New("[" + ErrCodeAuthEmpty + "] API token cannot be empty. " +
			"Provide a valid Webflow API token via config or environment variable. " +
			"See: https://github.com/jdetmar/pulumi-webflow/blob/main/docs/troubleshooting.md#api-token-not-configured")
	}

	// Basic sanity check - Webflow tokens should be reasonably long
	if len(token) < 10 {
		return errors.New("[" + ErrCodeAuthInvalid + "] API token appears invalid (too short). " +
			"Webflow API tokens are typically 40+ characters. " +
			"See: https://github.com/jdetmar/pulumi-webflow/blob/main/docs/troubleshooting.md#invalid-or-expired-token")
	}

	return nil
}

// RedactToken returns a redacted version of the token for logging.
// Always returns "[REDACTED]" to prevent token leakage in logs.
func RedactToken(token string) string {
	if token == "" {
		return "<empty>"
	}
	return "[REDACTED]"
}

// Default retry configuration for rate limit handling.
const (
	// DefaultMaxRetries is the maximum number of retry attempts for rate-limited requests.
	DefaultMaxRetries = 3
	// DefaultBaseDelay is the initial delay before the first retry.
	DefaultBaseDelay = 1 * time.Second
	// DefaultMaxDelay caps the maximum delay between retries.
	DefaultMaxDelay = 30 * time.Second
)

// retryTransport is an http.RoundTripper that handles rate limiting with exponential backoff.
type retryTransport struct {
	transport  http.RoundTripper // Underlying transport for actual HTTP requests
	maxRetries int               // Maximum number of retry attempts
	baseDelay  time.Duration     // Initial delay before first retry
	maxDelay   time.Duration     // Maximum delay between retries
}

// RoundTrip implements http.RoundTripper with retry logic for rate-limited requests.
func (t *retryTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	var resp *http.Response
	var err error
	ctx := req.Context()

	for attempt := 0; attempt <= t.maxRetries; attempt++ {
		// Clone the request for each attempt (body may have been consumed)
		clonedReq := req.Clone(ctx)

		resp, err = t.transport.RoundTrip(clonedReq)
		if err != nil {
			// Log API request error
			NewLogContext(ctx).
				WithField("method", req.Method).
				WithField("url", req.URL.Path).
				WithField("attempt", attempt+1).
				Errorf("HTTP request failed: %v", err)
			return nil, err
		}

		// Log API request at debug level
		NewLogContext(ctx).
			WithField("method", req.Method).
			WithField("url", req.URL.Path).
			WithField("status", resp.StatusCode).
			WithField("attempt", attempt+1).
			Debug("HTTP request completed")

		// If not rate limited, return immediately
		if resp.StatusCode != http.StatusTooManyRequests {
			return resp, nil
		}

		// Don't retry if we've exhausted attempts
		if attempt == t.maxRetries {
			NewLogContext(ctx).
				WithField("method", req.Method).
				WithField("url", req.URL.Path).
				WithField("maxRetries", t.maxRetries).
				Warn("Rate limit exceeded, max retries exhausted")
			return resp, nil
		}

		// Close the response body before retrying
		_ = resp.Body.Close()

		// Calculate delay with exponential backoff
		delay := t.calculateDelay(resp, attempt)

		// Log retry attempt
		NewLogContext(ctx).
			WithField("method", req.Method).
			WithField("url", req.URL.Path).
			WithField("attempt", attempt+1).
			WithField("retryAfter", delay.String()).
			Warnf("Rate limited, retrying after %v", delay)

		// Check if context is cancelled before sleeping
		select {
		case <-ctx.Done():
			return nil, ctx.Err()
		case <-time.After(delay):
			// Continue to next retry attempt
		}
	}

	return resp, nil
}

// calculateDelay determines how long to wait before the next retry.
// It respects the Retry-After header if present, otherwise uses exponential backoff.
func (t *retryTransport) calculateDelay(resp *http.Response, attempt int) time.Duration {
	// Check for Retry-After header (can be seconds or HTTP-date)
	if retryAfter := resp.Header.Get("Retry-After"); retryAfter != "" {
		// Try parsing as seconds first
		if seconds, err := strconv.ParseInt(retryAfter, 10, 64); err == nil {
			delay := time.Duration(seconds) * time.Second
			if delay > t.maxDelay {
				return t.maxDelay
			}
			return delay
		}
		// Could also parse HTTP-date format, but seconds is more common for APIs
	}

	// Exponential backoff: baseDelay * 2^attempt
	delay := time.Duration(float64(t.baseDelay) * math.Pow(2, float64(attempt)))
	if delay > t.maxDelay {
		return t.maxDelay
	}
	return delay
}

// authenticatedTransport is an http.RoundTripper that adds authentication headers.
type authenticatedTransport struct {
	token     string            // Webflow API token for Bearer authentication
	version   string            // Provider version for User-Agent header
	transport http.RoundTripper // Underlying transport for actual HTTP requests
}

// RoundTrip implements http.RoundTripper interface, adding authentication headers.
func (t *authenticatedTransport) RoundTrip(req *http.Request) (*http.Response, error) {
	// Clone the request to avoid modifying the original
	clonedReq := req.Clone(req.Context())

	// Add authentication header
	authHeader := "Bearer " + t.token
	clonedReq.Header.Set("Authorization", authHeader)

	// Add user agent
	clonedReq.Header.Set("User-Agent", "pulumi-webflow/"+t.version)

	// Add API version header for Webflow API v2
	clonedReq.Header.Set("Accept-Version", "2.0.0")

	// Execute the request
	return t.transport.RoundTrip(clonedReq)
}

// CreateHTTPClient creates an HTTP client configured for Webflow API v2.
// The client enforces HTTPS/TLS, includes authentication headers, has appropriate timeouts,
// and automatically retries rate-limited requests with exponential backoff.
//
// Note: This client does not set a base URL. In Pulumi provider architecture, resource
// implementations construct full URLs (e.g., "https://api.webflow.com/v2/sites/{id}")
// when making requests using this client.
func CreateHTTPClient(token, version string) (*http.Client, error) {
	if token == "" {
		return nil, errors.New("[" + ErrCodeAuthEmpty + "] cannot create HTTP client with empty token. " +
			"See: https://github.com/jdetmar/pulumi-webflow/blob/main/docs/troubleshooting.md#api-token-not-configured")
	}

	// Create TLS config with secure defaults
	tlsConfig := &tls.Config{
		MinVersion: tls.VersionTLS12, // Enforce TLS 1.2 minimum
	}

	// Create base transport with TLS config
	baseTransport := &http.Transport{
		TLSClientConfig: tlsConfig,
	}

	// Wrap with authentication transport
	authTransport := &authenticatedTransport{
		token:     token,
		version:   version,
		transport: baseTransport,
	}

	// Wrap with retry transport for rate limit handling
	retryTransport := &retryTransport{
		transport:  authTransport,
		maxRetries: DefaultMaxRetries,
		baseDelay:  DefaultBaseDelay,
		maxDelay:   DefaultMaxDelay,
	}

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: retryTransport,
	}

	return client, nil
}
