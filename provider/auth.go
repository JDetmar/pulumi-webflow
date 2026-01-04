// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"crypto/tls"
	"errors"
	"net/http"
	"os"
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
// The client enforces HTTPS/TLS, includes authentication headers, and has appropriate timeouts.
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

	// Create HTTP client with timeout
	client := &http.Client{
		Timeout:   30 * time.Second,
		Transport: authTransport,
	}

	return client, nil
}
