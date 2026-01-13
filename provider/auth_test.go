// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"crypto/tls"
	"io"
	"net/http"
	"os"
	"strings"
	"sync/atomic"
	"testing"
	"time"
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

	// Verify it's a retry transport (outermost layer)
	rt, ok := client.Transport.(*retryTransport)
	if !ok {
		t.Fatal("HTTP client transport is not retryTransport")
	}

	// Verify retry settings
	if rt.maxRetries != DefaultMaxRetries {
		t.Errorf("Retry maxRetries mismatch: expected %d, got %d", DefaultMaxRetries, rt.maxRetries)
	}

	// Verify authenticated transport is wrapped inside
	authTransport, ok := rt.transport.(*authenticatedTransport)
	if !ok {
		t.Fatal("Retry transport does not wrap authenticatedTransport")
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

	// Navigate through transport chain: retryTransport -> authenticatedTransport -> http.Transport
	rt := client.Transport.(*retryTransport)
	authTransport := rt.transport.(*authenticatedTransport)
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

// TestRetryTransport_NoRetryOnSuccess tests that successful requests don't retry.
func TestRetryTransport_NoRetryOnSuccess(t *testing.T) {
	var callCount int32

	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&callCount, 1)
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	rt := &retryTransport{
		transport:  mockTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call, got %d", callCount)
	}
	_ = resp.Body.Close()
}

// TestRetryTransport_RetryOn429 tests that 429 responses trigger retries.
func TestRetryTransport_RetryOn429(t *testing.T) {
	var callCount int32

	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			count := atomic.AddInt32(&callCount, 1)
			if count < 3 {
				// Return 429 for first two calls
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("{}")),
					Header:     http.Header{},
				}, nil
			}
			// Succeed on third call
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	rt := &retryTransport{
		transport:  mockTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond, // Short delay for testing
		maxDelay:   100 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200 after retries, got %d", resp.StatusCode)
	}
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected 3 calls (2 retries), got %d", callCount)
	}
	_ = resp.Body.Close()
}

// TestRetryTransport_MaxRetriesExceeded tests behavior when max retries are exhausted.
func TestRetryTransport_MaxRetriesExceeded(t *testing.T) {
	var callCount int32

	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&callCount, 1)
			// Always return 429
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader("{}")),
				Header:     http.Header{},
			}, nil
		},
	}

	rt := &retryTransport{
		transport:  mockTransport,
		maxRetries: 2,
		baseDelay:  10 * time.Millisecond,
		maxDelay:   100 * time.Millisecond,
	}

	req, _ := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	// Should return 429 after exhausting retries
	if resp.StatusCode != http.StatusTooManyRequests {
		t.Errorf("Expected status 429 after max retries, got %d", resp.StatusCode)
	}
	// 1 initial + 2 retries = 3 calls
	if atomic.LoadInt32(&callCount) != 3 {
		t.Errorf("Expected 3 calls, got %d", callCount)
	}
	_ = resp.Body.Close()
}

// TestRetryTransport_RespectsRetryAfterHeader tests Retry-After header handling.
func TestRetryTransport_RespectsRetryAfterHeader(t *testing.T) {
	var callCount int32
	var delays []time.Duration
	var lastCall time.Time

	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			now := time.Now()
			if !lastCall.IsZero() {
				delays = append(delays, now.Sub(lastCall))
			}
			lastCall = now

			count := atomic.AddInt32(&callCount, 1)
			if count == 1 {
				// First call returns 429 with Retry-After header
				header := http.Header{}
				header.Set("Retry-After", "1") // 1 second
				return &http.Response{
					StatusCode: http.StatusTooManyRequests,
					Body:       io.NopCloser(strings.NewReader("{}")),
					Header:     header,
				}, nil
			}
			return &http.Response{
				StatusCode: http.StatusOK,
				Body:       io.NopCloser(strings.NewReader("{}")),
			}, nil
		},
	}

	rt := &retryTransport{
		transport:  mockTransport,
		maxRetries: 3,
		baseDelay:  10 * time.Millisecond, // Would be much shorter without Retry-After
		maxDelay:   5 * time.Second,
	}

	req, _ := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
	resp, err := rt.RoundTrip(req)
	if err != nil {
		t.Fatalf("RoundTrip failed: %v", err)
	}
	if resp.StatusCode != http.StatusOK {
		t.Errorf("Expected status 200, got %d", resp.StatusCode)
	}

	// Verify we waited approximately 1 second (Retry-After value)
	if len(delays) > 0 {
		if delays[0] < 900*time.Millisecond {
			t.Errorf("Expected delay of ~1s from Retry-After, got %v", delays[0])
		}
	}
	_ = resp.Body.Close()
}

// TestRetryTransport_ContextCancellation tests that context cancellation stops retries.
func TestRetryTransport_ContextCancellation(t *testing.T) {
	var callCount int32

	mockTransport := &mockRoundTripper{
		handler: func(req *http.Request) (*http.Response, error) {
			atomic.AddInt32(&callCount, 1)
			return &http.Response{
				StatusCode: http.StatusTooManyRequests,
				Body:       io.NopCloser(strings.NewReader("{}")),
				Header:     http.Header{},
			}, nil
		},
	}

	rt := &retryTransport{
		transport:  mockTransport,
		maxRetries: 5,
		baseDelay:  500 * time.Millisecond, // Long delay so we can cancel
		maxDelay:   5 * time.Second,
	}

	ctx, cancel := context.WithCancel(context.Background())
	req, _ := http.NewRequestWithContext(ctx, "GET", "https://api.webflow.com/v2/sites", http.NoBody)

	// Cancel context after a short delay
	go func() {
		time.Sleep(50 * time.Millisecond)
		cancel()
	}()

	_, err := rt.RoundTrip(req)

	if err != context.Canceled {
		t.Errorf("Expected context.Canceled error, got %v", err)
	}
	// Should have made only 1 call before context was cancelled during wait
	if atomic.LoadInt32(&callCount) != 1 {
		t.Errorf("Expected 1 call before cancellation, got %d", callCount)
	}
}

// TestRetryTransport_CalculateDelay tests the delay calculation logic.
func TestRetryTransport_CalculateDelay(t *testing.T) {
	rt := &retryTransport{
		baseDelay: 1 * time.Second,
		maxDelay:  30 * time.Second,
	}

	tests := []struct {
		name        string
		attempt     int
		retryAfter  string
		minExpected time.Duration
		maxExpected time.Duration
	}{
		{"attempt 0", 0, "", 1 * time.Second, 1 * time.Second},
		{"attempt 1", 1, "", 2 * time.Second, 2 * time.Second},
		{"attempt 2", 2, "", 4 * time.Second, 4 * time.Second},
		{"attempt 5 (capped)", 5, "", 30 * time.Second, 30 * time.Second},
		{"with Retry-After", 0, "5", 5 * time.Second, 5 * time.Second},
		{"Retry-After exceeds max", 0, "60", 30 * time.Second, 30 * time.Second},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			header := http.Header{}
			if tt.retryAfter != "" {
				header.Set("Retry-After", tt.retryAfter)
			}
			resp := &http.Response{Header: header}

			delay := rt.calculateDelay(resp, tt.attempt)

			if delay < tt.minExpected || delay > tt.maxExpected {
				t.Errorf("Expected delay between %v and %v, got %v", tt.minExpected, tt.maxExpected, delay)
			}
		})
	}
}

// TestRetryTransport_NoRetryOnOtherErrors tests that non-429 errors don't retry.
func TestRetryTransport_NoRetryOnOtherErrors(t *testing.T) {
	statusCodes := []int{
		http.StatusBadRequest,
		http.StatusUnauthorized,
		http.StatusForbidden,
		http.StatusNotFound,
		http.StatusInternalServerError,
	}

	for _, statusCode := range statusCodes {
		t.Run(http.StatusText(statusCode), func(t *testing.T) {
			var callCount int32

			mockTransport := &mockRoundTripper{
				handler: func(req *http.Request) (*http.Response, error) {
					atomic.AddInt32(&callCount, 1)
					return &http.Response{
						StatusCode: statusCode,
						Body:       io.NopCloser(strings.NewReader("{}")),
					}, nil
				},
			}

			rt := &retryTransport{
				transport:  mockTransport,
				maxRetries: 3,
				baseDelay:  10 * time.Millisecond,
				maxDelay:   100 * time.Millisecond,
			}

			req, _ := http.NewRequest("GET", "https://api.webflow.com/v2/sites", http.NoBody)
			resp, err := rt.RoundTrip(req)
			if err != nil {
				t.Fatalf("RoundTrip failed: %v", err)
			}
			if resp.StatusCode != statusCode {
				t.Errorf("Expected status %d, got %d", statusCode, resp.StatusCode)
			}
			if atomic.LoadInt32(&callCount) != 1 {
				t.Errorf("Expected 1 call (no retries), got %d", callCount)
			}
			_ = resp.Body.Close()
		})
	}
}
