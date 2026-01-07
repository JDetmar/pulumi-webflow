// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"
)

// TestValidateCurrencyCode_Valid tests valid currency codes
func TestValidateCurrencyCode_Valid(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"USD", "USD"},
		{"EUR", "EUR"},
		{"GBP", "GBP"},
		{"JPY", "JPY"},
		{"CAD", "CAD"},
		{"AUD", "AUD"},
		{"CHF", "CHF"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrencyCode(tt.code)
			if err != nil {
				t.Errorf("ValidateCurrencyCode(%q) = %v, want nil", tt.code, err)
			}
		})
	}
}

// TestValidateCurrencyCode_Empty tests empty currency code
func TestValidateCurrencyCode_Empty(t *testing.T) {
	err := ValidateCurrencyCode("")
	if err == nil {
		t.Error("ValidateCurrencyCode(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateCurrencyCode_Invalid tests invalid currency codes
func TestValidateCurrencyCode_Invalid(t *testing.T) {
	tests := []struct {
		name string
		code string
	}{
		{"lowercase", "usd"},
		{"too short", "US"},
		{"too long", "USDD"},
		{"with numbers", "US1"},
		{"with special char", "US$"},
		{"mixed case", "Usd"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCurrencyCode(tt.code)
			if err == nil {
				t.Errorf("ValidateCurrencyCode(%q) = nil, want error", tt.code)
			}
			if !strings.Contains(err.Error(), "invalid format") {
				t.Errorf("Expected error to mention 'invalid format', got: %v", err)
			}
		})
	}
}

// TestGenerateEcommerceSettingsResourceID tests resource ID generation
func TestGenerateEcommerceSettingsResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"

	resourceID := GenerateEcommerceSettingsResourceID(siteID)
	expected := "5f0c8c9e1c9d440000e8d8c3/ecommerce/settings"

	if resourceID != expected {
		t.Errorf("GenerateEcommerceSettingsResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractSiteIDFromEcommerceSettingsResourceID_Valid tests extracting site ID from valid resource ID
func TestExtractSiteIDFromEcommerceSettingsResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/ecommerce/settings"

	siteID, err := ExtractSiteIDFromEcommerceSettingsResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractSiteIDFromEcommerceSettingsResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractSiteIDFromEcommerceSettingsResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
}

// TestExtractSiteIDFromEcommerceSettingsResourceID_Empty tests empty resource ID
func TestExtractSiteIDFromEcommerceSettingsResourceID_Empty(t *testing.T) {
	_, err := ExtractSiteIDFromEcommerceSettingsResourceID("")
	if err == nil {
		t.Error("ExtractSiteIDFromEcommerceSettingsResourceID(\"\") error = nil, want error")
	}
}

// TestExtractSiteIDFromEcommerceSettingsResourceID_InvalidFormat tests invalid format
func TestExtractSiteIDFromEcommerceSettingsResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing suffix", "5f0c8c9e1c9d440000e8d8c3"},
		{"wrong suffix", "5f0c8c9e1c9d440000e8d8c3/ecommerce"},
		{"partial suffix", "5f0c8c9e1c9d440000e8d8c3/settings"},
		{"different resource type", "5f0c8c9e1c9d440000e8d8c3/redirects/123"},
		{"too short", "/ecommerce/settings"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, err := ExtractSiteIDFromEcommerceSettingsResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractSiteIDFromEcommerceSettingsResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestGetEcommerceSettings_Valid tests retrieving ecommerce settings successfully
func TestGetEcommerceSettings_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/ecommerce/settings") {
			t.Errorf("Expected /ecommerce/settings in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := EcommerceSettingsResponse{
			SiteID:          "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:       "2024-01-15T10:30:00Z",
			DefaultCurrency: "USD",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetEcommerceSettings failed: %v", err)
	}

	if result.SiteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("Expected siteId '5f0c8c9e1c9d440000e8d8c3', got %s", result.SiteID)
	}
	if result.DefaultCurrency != "USD" {
		t.Errorf("Expected defaultCurrency 'USD', got %s", result.DefaultCurrency)
	}
	if result.CreatedOn != "2024-01-15T10:30:00Z" {
		t.Errorf("Expected createdOn '2024-01-15T10:30:00Z', got %s", result.CreatedOn)
	}
}

// TestGetEcommerceSettings_NotFound tests 404 handling
func TestGetEcommerceSettings_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetEcommerceSettings_EcommerceNotEnabled tests 409 handling when ecommerce is not enabled
func TestGetEcommerceSettings_EcommerceNotEnabled(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusConflict)
		_, _ = w.Write([]byte(`{"code": "ecommerce_not_enabled", "message": "Site does not have ecommerce enabled"}`))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 409, got nil")
	}
	if !strings.Contains(err.Error(), "ecommerce not enabled") {
		t.Errorf("Expected 'ecommerce not enabled' in error, got: %v", err)
	}
}

// TestGetEcommerceSettings_Unauthorized tests 401 handling
func TestGetEcommerceSettings_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Expected 'unauthorized' in error, got: %v", err)
	}
}

// TestGetEcommerceSettings_Forbidden tests 403 handling
func TestGetEcommerceSettings_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected 'forbidden' in error, got: %v", err)
	}
}

// TestGetEcommerceSettings_ServerError tests 500 handling
func TestGetEcommerceSettings_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestGetEcommerceSettings_BadRequest tests 400 handling
func TestGetEcommerceSettings_BadRequest(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("bad request"))
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetEcommerceSettings(ctx, client, "invalid-site-id")
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestEcommerceSettingsErrorMessagesAreActionable verifies error messages contain guidance
func TestEcommerceSettingsErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateCurrencyCode empty",
			func() error { return ValidateCurrencyCode("") },
			[]string{"required", "ISO 4217"},
		},
		{
			"ValidateCurrencyCode invalid",
			func() error { return ValidateCurrencyCode("usd") },
			[]string{"invalid format", "uppercase"},
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := tt.testFunc()
			if err == nil {
				t.Errorf("%s: expected error, got nil", tt.name)
				return
			}

			errMsg := err.Error()
			for _, expectedStr := range tt.contains {
				if !strings.Contains(errMsg, expectedStr) {
					t.Errorf("%s: error message missing %q. Got: %s", tt.name, expectedStr, errMsg)
				}
			}
		})
	}
}

// TestHandleEcommerceNotEnabledError tests the 409 error handler
func TestHandleEcommerceNotEnabledError(t *testing.T) {
	body := []byte(`{"code": "ecommerce_not_enabled", "message": "Site lacks ecommerce"}`)
	err := handleEcommerceNotEnabledError(body)

	if err == nil {
		t.Error("Expected error, got nil")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "ecommerce not enabled") {
		t.Errorf("Expected 'ecommerce not enabled' in error, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "Webflow dashboard") {
		t.Errorf("Expected actionable guidance mentioning 'Webflow dashboard', got: %s", errMsg)
	}
}

// TestGetEcommerceSettings_ContextCancellation tests context cancellation handling
func TestGetEcommerceSettings_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(http.StatusOK)
	}))
	defer server.Close()

	oldURL := getEcommerceSettingsBaseURL
	getEcommerceSettingsBaseURL = server.URL
	defer func() { getEcommerceSettingsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	_, err := GetEcommerceSettings(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}
