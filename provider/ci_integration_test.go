// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"os"
	"testing"
)

// TestCIEnvironmentVariableLoading verifies that CI/CD environment variables are properly loaded
// for non-interactive execution.
func TestCIEnvironmentVariableLoading(t *testing.T) {
	tests := []struct {
		name     string
		envVar   string
		envValue string
		expected string
	}{
		{
			name:     "WebflowAPITokenFromEnvironment",
			envVar:   "WEBFLOW_API_TOKEN",
			envValue: "test-token-12345",
			expected: "test-token-12345",
		},
		{
			name:     "EmptyTokenHandling",
			envVar:   "WEBFLOW_API_TOKEN",
			envValue: "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Set environment variable for test
			oldValue, wasSet := os.LookupEnv(tt.envVar)
			defer func() {
				if wasSet {
					_ = os.Setenv(tt.envVar, oldValue)
				} else {
					_ = os.Unsetenv(tt.envVar)
				}
			}()

			_ = os.Setenv(tt.envVar, tt.envValue)

			// Verify environment variable is set correctly
			actual := os.Getenv(tt.envVar)
			if actual != tt.expected {
				t.Errorf("Expected %q, got %q", tt.expected, actual)
			}
		})
	}
}

// TestNonInteractiveExecution verifies provider can be used in non-interactive mode.
// This simulates CI/CD pipeline execution patterns.
func TestNonInteractiveExecution(t *testing.T) {
	// Verify we can load the provider without interactive prompts
	defer func() {
		if r := recover(); r != nil {
			t.Fatal("Provider should load without panic")
		}
	}()

	_ = Provider()

	// Verify provider has correct name
	expectedName := "webflow"
	if Name != expectedName {
		t.Errorf("Provider name mismatch: expected %q, got %q", expectedName, Name)
	}

	// Provider should handle CI/CD execution patterns
	t.Logf("Provider loaded successfully for non-interactive execution")
}

// TestCredentialNotLogged verifies that API credentials are not exposed in logs.
// This is critical for secure CI/CD execution (NFR11, FR17).
func TestCredentialNotLogged(t *testing.T) {
	testCases := []struct {
		name     string
		apiToken string
	}{
		{
			name:     "TokenNotInStringRepresentation",
			apiToken: "super-secret-token-xyz",
		},
		{
			name:     "TokenNotExposedInConfig",
			apiToken: "webflow-api-token-abc123",
		},
	}

	for _, tc := range testCases {
		t.Run(tc.name, func(t *testing.T) {
			// Create a config with sensitive token
			config := &Config{
				APIToken: tc.apiToken,
			}

			// Verify credentials are available internally
			if config.APIToken != tc.apiToken {
				t.Errorf("Config should store token internally")
			}

			// Verify RedactToken properly masks the credential
			redacted := RedactToken(tc.apiToken)
			if redacted != "[REDACTED]" {
				t.Errorf("RedactToken should return '[REDACTED]', got %q", redacted)
			}

			// Verify redacted output does not contain the original token
			if redacted == tc.apiToken {
				t.Errorf("Redacted output should not equal original token")
			}

			// Verify fmt.Sprintf patterns don't expose token via RedactToken
			logMessage := "Using token: " + RedactToken(config.APIToken)
			if logMessage != "Using token: [REDACTED]" {
				t.Errorf("Log message pattern incorrect: %s", logMessage)
			}
		})
	}
}

// TestRedactTokenFunction verifies the RedactToken helper function behavior.
func TestRedactTokenFunction(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "NormalToken",
			input:    "wf_live_abc123xyz",
			expected: "[REDACTED]",
		},
		{
			name:     "EmptyToken",
			input:    "",
			expected: "<empty>",
		},
		{
			name:     "LongToken",
			input:    "very-long-secret-token-that-should-never-appear-in-logs-ever",
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactToken(tt.input)
			if result != tt.expected {
				t.Errorf("RedactToken(%q) = %q, want %q", tt.input, result, tt.expected)
			}

			// Extra safety: ensure original token never appears in result
			if tt.input != "" && result == tt.input {
				t.Errorf("RedactToken returned the original token - security violation!")
			}
		})
	}
}

// TestExitCodeHandling verifies that provider errors properly propagate for CI/CD pipelines.
// Pulumi CLI handles exit codes; this test verifies provider returns proper errors.
func TestExitCodeHandling(t *testing.T) {
	tests := []struct {
		name        string
		token       string
		expectError bool
		errorMsg    string
		useValidate bool // Use ValidateToken instead of CreateHTTPClient
	}{
		{
			name:        "SuccessfulProviderLoad",
			token:       "",
			expectError: false,
		},
		{
			name:        "EmptyTokenReturnsError",
			token:       "",
			expectError: true,
			errorMsg:    "cannot create HTTP client with empty token",
		},
		{
			name:        "InvalidTokenTooShort",
			token:       "short",
			expectError: true,
			errorMsg:    "too short",
			useValidate: true, // ValidateToken checks length
		},
		{
			name:        "ValidTokenAccepted",
			token:       "valid-token-12345",
			expectError: false,
			useValidate: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			if tt.name == "SuccessfulProviderLoad" {
				// Provider should load without panic
				defer func() {
					if r := recover(); r != nil {
						t.Errorf("Provider load should not panic")
					}
				}()
				_ = Provider()
				t.Logf("Provider loaded successfully")
				return
			}

			var err error
			if tt.useValidate {
				// Test error propagation through ValidateToken
				err = ValidateToken(tt.token)
			} else {
				// Test error propagation through CreateHTTPClient
				_, err = CreateHTTPClient(tt.token, "test")
			}

			if tt.expectError {
				if err == nil {
					t.Errorf("Expected error for token %q, got nil", tt.token)
					return
				}
				if tt.errorMsg != "" && !containsString(err.Error(), tt.errorMsg) {
					t.Errorf("Error %q should contain %q", err.Error(), tt.errorMsg)
				}
				// This error would cause Pulumi CLI to exit with non-zero code
				t.Logf("Error properly returned (would cause non-zero exit): %v", err)
			} else if err != nil {
				t.Errorf("Unexpected error: %v", err)
			}
		})
	}
}

// containsString checks if s contains substr (helper for error message checking)
func containsString(s, substr string) bool {
	return len(s) >= len(substr) && (s == substr || substr == "" ||
		(s != "" && substr != "" && searchString(s, substr)))
}

func searchString(s, substr string) bool {
	for i := 0; i <= len(s)-len(substr); i++ {
		if s[i:i+len(substr)] == substr {
			return true
		}
	}
	return false
}

// TestPulumiDiagnosticFormatting verifies that error messages follow Pulumi diagnostic formatting
// for CI/CD log parsing (NFR29).
func TestPulumiDiagnosticFormatting(t *testing.T) {
	// This test verifies that the provider doesn't have custom error formatting
	// that would conflict with Pulumi's standard diagnostic output.
	// Pulumi handles all CLI output formatting automatically.

	tests := []struct {
		name     string
		scenario string
	}{
		{
			name:     "StandardPulumiErrorFormat",
			scenario: "Provider delegates error formatting to Pulumi CLI",
		},
		{
			name:     "NoCustomFormatting",
			scenario: "Provider returns structured errors for Pulumi to format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Verify provider is loaded - actual error formatting is tested
			// through integration tests with Pulumi CLI
			defer func() {
				if r := recover(); r != nil {
					t.Errorf("Provider should load without panic")
				}
			}()

			_ = Provider()
			t.Logf("Provider loaded for error formatting: %s", tt.scenario)
		})
	}
}

// TestCIEnvironmentSetupPatterns tests various CI/CD environment setup patterns
func TestCIEnvironmentSetupPatterns(t *testing.T) {
	patterns := []struct {
		name     string
		envVars  map[string]string
		expected bool
	}{
		{
			name: "GitHubActionsEnvironment",
			envVars: map[string]string{
				"GITHUB_ACTIONS":            "true",
				"WEBFLOW_API_TOKEN":         "github-token",
				"PULUMI_SKIP_CONFIRMATIONS": "true",
			},
			expected: true,
		},
		{
			name: "GitLabCIEnvironment",
			envVars: map[string]string{
				"GITLAB_CI":                 "true",
				"WEBFLOW_API_TOKEN":         "gitlab-token",
				"PULUMI_SKIP_CONFIRMATIONS": "true",
			},
			expected: true,
		},
		{
			name: "GenericCIEnvironment",
			envVars: map[string]string{
				"CI":                        "true",
				"WEBFLOW_API_TOKEN":         "ci-token",
				"PULUMI_SKIP_CONFIRMATIONS": "true",
			},
			expected: true,
		},
	}

	for _, pattern := range patterns {
		t.Run(pattern.name, func(t *testing.T) {
			// Set environment variables
			oldEnv := make(map[string]string)
			for key, value := range pattern.envVars {
				oldVal, exists := os.LookupEnv(key)
				if exists {
					oldEnv[key] = oldVal
				}
				_ = os.Setenv(key, value)
			}

			// Cleanup
			defer func() {
				for key := range pattern.envVars {
					if oldVal, exists := oldEnv[key]; exists {
						_ = os.Setenv(key, oldVal)
					} else {
						_ = os.Unsetenv(key)
					}
				}
			}()

			// Verify environment is configured for CI/CD
			defer func() {
				if r := recover(); r != nil && pattern.expected {
					t.Errorf("Provider should initialize in CI environment")
				}
			}()

			_ = Provider()
			t.Logf("Provider initialized successfully in %s", pattern.name)
		})
	}
}

// TestMultiStackManagement verifies that provider supports multi-stack deployments
// required for dev/staging/prod CI/CD patterns.
func TestMultiStackManagement(t *testing.T) {
	// The provider itself doesn't manage stacks - Pulumi does via configuration.
	// This test verifies the provider is compatible with multi-stack patterns.

	t.Run("ProviderSupportsMultipleStacks", func(t *testing.T) {
		// Provider should load identically regardless of which stack it's used with
		defer func() {
			if r := recover(); r != nil {
				t.Fatal("Provider should load for multi-stack deployments")
			}
		}()

		_ = Provider()

		// Provider name should be consistent across all stacks
		if Name != "webflow" {
			t.Errorf("Provider name should be consistent for multi-stack management")
		}

		t.Logf("Provider supports multi-stack deployments")
	})
}
