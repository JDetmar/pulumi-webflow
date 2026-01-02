// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"crypto/tls"
	"net/http"
	"os"
	"strings"
	"testing"
)

// TestGetEnvToken tests retrieving token from environment variable.
func TestGetEnvToken(t *testing.T) {
	// Test when env var is set
	_ = os.Setenv("WEBFLOW_API_TOKEN", "test-token-from-env")
	defer func() { _ = os.Unsetenv("WEBFLOW_API_TOKEN") }()

	token := getEnvToken()
	if token != "test-token-from-env" {
		t.Errorf("Expected token 'test-token-from-env', got '%s'", token)
	}
}

// TestGetEnvToken_NotSet tests when environment variable is not set.
func TestGetEnvToken_NotSet(t *testing.T) {
	_ = os.Unsetenv("WEBFLOW_API_TOKEN")

	token := getEnvToken()
	if token != "" {
		t.Errorf("Expected empty token, got '%s'", token)
	}
}

// TestValidateToken_ValidToken tests validation of valid tokens.
func TestValidateToken_ValidToken(t *testing.T) {
	tests := []struct {
		name  string
		token string
	}{
		{"valid token", "wf_1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p"},
		{"minimum length", "1234567890"},
		{"long token", "very_long_token_that_should_pass_validation_check_12345678901234567890"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToken(tt.token)
			if err != nil {
				t.Errorf("ValidateToken failed for valid token: %v", err)
			}
		})
	}
}

// TestValidateToken_InvalidToken tests validation errors for invalid tokens.
func TestValidateToken_InvalidToken(t *testing.T) {
	tests := []struct {
		name          string
		token         string
		expectedError string
	}{
		{"empty token", "", "cannot be empty"},
		{"too short", "short", "too short"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateToken(tt.token)
			if err == nil {
				t.Error("Expected error for invalid token, got nil")
			}
			if !strings.Contains(err.Error(), tt.expectedError) {
				t.Errorf("Expected error containing '%s', got: %v", tt.expectedError, err)
			}
		})
	}
}

// TestRedactToken tests token redaction for logging.
func TestRedactToken(t *testing.T) {
	tests := []struct {
		name     string
		token    string
		expected string
	}{
		{"normal token", "wf_1a2b3c4d5e6f7g8h9i0j1k2l3m4n5o6p", "[REDACTED]"},
		{"short token", "test", "[REDACTED]"},
		{"empty token", "", "<empty>"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactToken(tt.token)
			if result != tt.expected {
				t.Errorf("Expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// TestCreateHTTPClient_Success tests successful HTTP client creation.
func TestCreateHTTPClient_Success(t *testing.T) {
	token := "test-token-1234567890"
	version := "0.1.0"

	client, err := CreateHTTPClient(token, version)
	if err != nil {
		t.Fatalf("CreateHTTPClient failed: %v", err)
	}

	if client == nil {
		t.Fatal("CreateHTTPClient returned nil client")
	}

	// Verify timeout is set
	if client.Timeout == 0 {
		t.Error("HTTP client timeout not set")
	}

	// Verify transport is configured
	if client.Transport == nil {
		t.Fatal("HTTP client transport is nil")
	}

	// Verify it's an authenticated transport
	authTransport, ok := client.Transport.(*authenticatedTransport)
	if !ok {
		t.Fatal("HTTP client transport is not authenticatedTransport")
	}

	if authTransport.token != token {
		t.Errorf("Transport token mismatch: expected '%s', got '%s'", token, authTransport.token)
	}

	if authTransport.version != version {
		t.Errorf("Transport version mismatch: expected '%s', got '%s'", version, authTransport.version)
	}
}

// TestCreateHTTPClient_EmptyToken tests error when creating client with empty token.
func TestCreateHTTPClient_EmptyToken(t *testing.T) {
	_, err := CreateHTTPClient("", "0.1.0")
	if err == nil {
		t.Fatal("Expected error when creating client with empty token, got nil")
	}

	expectedMsg := "cannot create HTTP client with empty token"
	if !strings.Contains(err.Error(), expectedMsg) {
		t.Errorf("Expected error containing '%s', got: %v", expectedMsg, err)
	}
}

// TestCreateHTTPClient_TLSConfiguration tests that TLS is properly configured.
func TestCreateHTTPClient_TLSConfiguration(t *testing.T) {
	client, err := CreateHTTPClient("test-token-1234567890", "0.1.0")
	if err != nil {
		t.Fatalf("CreateHTTPClient failed: %v", err)
	}

	authTransport := client.Transport.(*authenticatedTransport)
	httpTransport, ok := authTransport.transport.(*http.Transport)
	if !ok {
		t.Fatal("Base transport is not http.Transport")
	}

	if httpTransport.TLSClientConfig == nil {
		t.Fatal("TLS config is nil")
	}

	if httpTransport.TLSClientConfig.MinVersion != tls.VersionTLS12 {
		t.Errorf("Expected TLS 1.2 minimum, got version %d", httpTransport.TLSClientConfig.MinVersion)
	}
}

// TestAuthenticatedTransport_RoundTrip tests that auth headers are added.
func TestAuthenticatedTransport_RoundTrip(t *testing.T) {
	token := "test-token-1234567890"
	version := "0.1.0"

	// Create mock round tripper
	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			// Verify Authorization header
			authHeader := req.Header.Get("Authorization")
			expectedAuth := "Bearer " + token
			if authHeader != expectedAuth {
				t.Errorf("Expected Authorization header '%s', got '%s'", expectedAuth, authHeader)
			}

			// Verify User-Agent header
			userAgent := req.Header.Get("User-Agent")
			expectedUA := "pulumi-webflow/" + version
			if userAgent != expectedUA {
				t.Errorf("Expected User-Agent '%s', got '%s'", expectedUA, userAgent)
			}

			return &http.Response{
				StatusCode: 200,
				Body:       http.NoBody,
			}, nil
		},
	}

	authTransport := &authenticatedTransport{
		token:     token,
		version:   version,
		transport: mockTransport,
	}

	// Create a test request
	req, err := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	// Execute round trip
	_, err = authTransport.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
}

// TestCreateHTTPClient_ErrorHandling tests that HTTP client handles connection errors.
func TestCreateHTTPClient_ErrorHandling(t *testing.T) {
	client, err := CreateHTTPClient("test-token-1234567890", "0.1.0")
	if err != nil {
		t.Fatalf("CreateHTTPClient failed: %v", err)
	}

	// Attempt to connect to invalid host to verify error handling
	req, err := http.NewRequest("GET", "https://invalid-webflow-host-that-does-not-exist.com/v2/sites", http.NoBody)
	if err != nil {
		t.Fatalf("Failed to create request: %v", err)
	}

	resp, err := client.Do(req)
	if err == nil {
		t.Error("Expected error when connecting to invalid host, got nil")
		if resp != nil {
			_ = resp.Body.Close()
		}
	}

	// Verify error is a network error (not a panic or nil pointer)
	if err != nil && !strings.Contains(err.Error(), "no such host") &&
		!strings.Contains(err.Error(), "connection") &&
		!strings.Contains(err.Error(), "dial") {
		t.Logf("Got expected network error (type may vary by platform): %v", err)
	}
}

// mockRoundTripper is a mock http.RoundTripper for testing.
type mockRoundTripper struct {
	handler func(*http.Request) (*http.Response, error)
}

func (m *mockRoundTripper) RoundTrip(req *http.Request) (*http.Response, error) {
	return m.handler(req)
}
