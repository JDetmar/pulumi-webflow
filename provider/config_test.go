// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestSafeGetConfigToken_EmptyContext(t *testing.T) {
	// Test that safeGetConfigToken doesn't panic with an empty context
	// This simulates the invoke function scenario where config isn't available
	ctx := context.Background()

	// This should NOT panic - it should return empty string
	token := safeGetConfigToken(ctx)

	assert.Equal(t, "", token, "safeGetConfigToken should return empty string when config is not in context")
}

func TestGetHTTPClient_WithEnvVar(t *testing.T) {
	// t.Setenv automatically restores the original value after the test
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Use empty context (no config) - should work because env var is set
	ctx := context.Background()

	client, err := GetHTTPClient(ctx, "test-version")

	assert.NoError(t, err, "GetHTTPClient should succeed with env var set")
	assert.NotNil(t, client, "HTTP client should not be nil")
}

func TestGetHTTPClient_NoTokenConfigured(t *testing.T) {
	// Clear the environment variable
	t.Setenv("WEBFLOW_API_TOKEN", "")

	// Use empty context (no config) and no env var
	ctx := context.Background()

	client, err := GetHTTPClient(ctx, "test-version")

	assert.Error(t, err, "GetHTTPClient should return error when no token is configured")
	assert.Nil(t, client, "HTTP client should be nil when no token is configured")
	assert.ErrorIs(t, err, ErrTokenNotConfigured, "Error should be ErrTokenNotConfigured")
}

func TestGetHTTPClient_InvalidToken(t *testing.T) {
	// Set up environment variable with invalid (too short) token
	t.Setenv("WEBFLOW_API_TOKEN", "short")

	ctx := context.Background()

	client, err := GetHTTPClient(ctx, "test-version")

	assert.Error(t, err, "GetHTTPClient should return error for invalid token")
	assert.Nil(t, client, "HTTP client should be nil for invalid token")
	assert.Contains(t, err.Error(), "WEBFLOW_AUTH_003", "Error should indicate invalid token")
}

func TestGetHTTPClient_EnvVarWorksWithoutConfig(t *testing.T) {
	// This test verifies that env var works even when config is unavailable
	// (which would cause infer.GetConfig to panic if called directly)
	t.Setenv("WEBFLOW_API_TOKEN", "env-token-12345678901234567890")

	// Use context.Background() which has no config
	ctx := context.Background()

	// This should NOT panic because env var is checked first
	client, err := GetHTTPClient(ctx, "test-version")

	assert.NoError(t, err, "GetHTTPClient should succeed with env var (config not needed)")
	assert.NotNil(t, client, "HTTP client should not be nil")
}

// TestGetHTTPClient_NoPanicOnInvokeContext tests the specific scenario that was crashing:
// invoke functions being called with a context that doesn't have provider config
func TestGetHTTPClient_NoPanicOnInvokeContext(t *testing.T) {
	// Clear the environment variable to force the config path
	t.Setenv("WEBFLOW_API_TOKEN", "")

	// Simulate invoke function context (no config attached)
	ctx := context.Background()

	// This should NOT panic - it should return an error gracefully
	defer func() {
		if r := recover(); r != nil {
			t.Errorf("GetHTTPClient panicked (this was the bug we fixed): %v", r)
		}
	}()

	client, err := GetHTTPClient(ctx, "test-version")

	// Should get an error, not a panic
	assert.Error(t, err, "Should return error when no token available")
	assert.Nil(t, client)
	assert.ErrorIs(t, err, ErrTokenNotConfigured)
}
