// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"strings"
	"testing"
)

func TestRedactSensitiveData(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		expected string
	}{
		{
			name:     "empty string",
			input:    "",
			expected: "<empty>",
		},
		{
			name:     "short sensitive data",
			input:    "secret123",
			expected: "[REDACTED]",
		},
		{
			name:     "long token",
			input:    "abcdef1234567890abcdef1234567890abcdef1234567890",
			expected: "[REDACTED]",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := RedactSensitiveData(tt.input)
			if result != tt.expected {
				t.Errorf("RedactSensitiveData(%q) = %q, expected %q", tt.input, result, tt.expected)
			}
		})
	}
}

func TestTruncateForLogging(t *testing.T) {
	tests := []struct {
		name     string
		input    string
		maxLen   int
		contains string
	}{
		{
			name:     "short string not truncated",
			input:    "hello",
			maxLen:   10,
			contains: "hello",
		},
		{
			name:     "long string truncated",
			input:    strings.Repeat("a", 100),
			maxLen:   20,
			contains: "truncated",
		},
		{
			name:     "exact length not truncated",
			input:    "exactly20characters!",
			maxLen:   20,
			contains: "exactly20characters!",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := TruncateForLogging(tt.input, tt.maxLen)
			if !strings.Contains(result, tt.contains) {
				t.Errorf("TruncateForLogging result %q does not contain %q", result, tt.contains)
			}
			if len(tt.input) > tt.maxLen && len(result) > tt.maxLen+50 {
				t.Errorf("TruncateForLogging result too long: %d chars", len(result))
			}
		})
	}
}

func TestSafeString(t *testing.T) {
	tests := []struct {
		name      string
		fieldName string
		value     interface{}
		expected  string
	}{
		{
			name:      "nil value",
			fieldName: "anything",
			value:     nil,
			expected:  "<nil>",
		},
		{
			name:      "normal field",
			fieldName: "siteId",
			value:     "12345",
			expected:  "12345",
		},
		{
			name:      "token field redacted",
			fieldName: "apiToken",
			value:     "secret-token-value",
			expected:  "[REDACTED]",
		},
		{
			name:      "password field redacted",
			fieldName: "userPassword",
			value:     "mypassword",
			expected:  "[REDACTED]",
		},
		{
			name:      "authorization field redacted",
			fieldName: "Authorization",
			value:     "Bearer token",
			expected:  "[REDACTED]",
		},
		{
			name:      "number value",
			fieldName: "count",
			value:     42,
			expected:  "42",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := SafeString(tt.fieldName, tt.value)
			if result != tt.expected {
				t.Errorf("SafeString(%q, %v) = %q, expected %q", tt.fieldName, tt.value, result, tt.expected)
			}
		})
	}
}

func TestLogContextFormatMessage(t *testing.T) {
	tests := []struct {
		name     string
		msg      string
		fields   map[string]interface{}
		contains []string
	}{
		{
			name:     "no fields",
			msg:      "test message",
			fields:   map[string]interface{}{},
			contains: []string{"test message"},
		},
		{
			name:   "single field",
			msg:    "creating resource",
			fields: map[string]interface{}{"siteId": "12345"},
			contains: []string{
				"creating resource",
				"siteId=12345",
			},
		},
		{
			name: "multiple fields",
			msg:  "API request",
			fields: map[string]interface{}{
				"method": "POST",
				"status": 200,
			},
			contains: []string{
				"API request",
				"method=POST",
				"status=200",
			},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			lc := &LogContext{
				fields: tt.fields,
			}
			result := lc.formatMessage(tt.msg)
			for _, expected := range tt.contains {
				if !strings.Contains(result, expected) {
					t.Errorf("formatMessage result %q does not contain %q", result, expected)
				}
			}
		})
	}
}
