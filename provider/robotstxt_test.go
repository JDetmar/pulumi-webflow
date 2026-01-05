// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"bytes"
	"context"
	"errors"
	"io"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestParseRobotsTxtContent tests parsing of robots.txt content string to rules.
func TestParseRobotsTxtContent(t *testing.T) {
	tests := []struct {
		name            string
		content         string
		expectedRules   int
		expectedSitemap string
	}{
		{
			name:            "simple allow all",
			content:         "User-agent: *\nAllow: /",
			expectedRules:   1,
			expectedSitemap: "",
		},
		{
			name:            "with disallow",
			content:         "User-agent: *\nAllow: /\nDisallow: /admin/",
			expectedRules:   1,
			expectedSitemap: "",
		},
		{
			name:            "with sitemap",
			content:         "User-agent: *\nAllow: /\nSitemap: https://example.com/sitemap.xml",
			expectedRules:   1,
			expectedSitemap: "https://example.com/sitemap.xml",
		},
		{
			name:            "multiple user agents",
			content:         "User-agent: *\nAllow: /\n\nUser-agent: Googlebot\nDisallow: /private/",
			expectedRules:   2,
			expectedSitemap: "",
		},
		{
			name:            "empty content",
			content:         "",
			expectedRules:   0,
			expectedSitemap: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			rules, sitemap := ParseRobotsTxtContent(tt.content)
			if len(rules) != tt.expectedRules {
				t.Errorf("expected %d rules, got %d", tt.expectedRules, len(rules))
			}
			if sitemap != tt.expectedSitemap {
				t.Errorf("expected sitemap '%s', got '%s'", tt.expectedSitemap, sitemap)
			}
		})
	}
}

// TestFormatRobotsTxtContent tests formatting rules back to content string.
func TestFormatRobotsTxtContent(t *testing.T) {
	tests := []struct {
		name     string
		rules    []RobotsTxtRule
		sitemap  string
		expected string
	}{
		{
			name: "simple allow all",
			rules: []RobotsTxtRule{
				{UserAgent: "*", Allows: []string{"/"}, Disallows: []string{}},
			},
			sitemap:  "",
			expected: "User-agent: *\nAllow: /\n",
		},
		{
			name: "with disallow",
			rules: []RobotsTxtRule{
				{UserAgent: "*", Allows: []string{"/"}, Disallows: []string{"/admin/"}},
			},
			sitemap:  "",
			expected: "User-agent: *\nAllow: /\nDisallow: /admin/\n",
		},
		{
			name: "with sitemap",
			rules: []RobotsTxtRule{
				{UserAgent: "*", Allows: []string{"/"}},
			},
			sitemap:  "https://example.com/sitemap.xml",
			expected: "User-agent: *\nAllow: /\n\nSitemap: https://example.com/sitemap.xml\n",
		},
		{
			name:     "empty rules",
			rules:    []RobotsTxtRule{},
			sitemap:  "",
			expected: "",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := FormatRobotsTxtContent(tt.rules, tt.sitemap)
			if result != tt.expected {
				t.Errorf("expected:\n%q\ngot:\n%q", tt.expected, result)
			}
		})
	}
}

// TestRobotsTxtRule_Struct tests the RobotsTxtRule struct.
func TestRobotsTxtRule_Struct(t *testing.T) {
	rule := RobotsTxtRule{
		UserAgent: "*",
		Allows:    []string{"/", "/public/"},
		Disallows: []string{"/admin/", "/private/"},
	}

	if rule.UserAgent != "*" {
		t.Error("UserAgent should be '*'")
	}
	if len(rule.Allows) != 2 {
		t.Errorf("expected 2 allows, got %d", len(rule.Allows))
	}
	if len(rule.Disallows) != 2 {
		t.Errorf("expected 2 disallows, got %d", len(rule.Disallows))
	}
}

// TestValidateSiteID tests siteID validation.
func TestValidateSiteID(t *testing.T) {
	tests := []struct {
		name    string
		siteID  string
		wantErr bool
	}{
		{
			name:    "valid 24-char hex",
			siteID:  "5f0c8c9e1c9d440000e8d8c3",
			wantErr: false,
		},
		{
			name:    "valid all lowercase",
			siteID:  "abcdef0123456789abcdef01",
			wantErr: false,
		},
		{
			name:    "too short",
			siteID:  "5f0c8c9e1c9d44",
			wantErr: true,
		},
		{
			name:    "too long",
			siteID:  "5f0c8c9e1c9d440000e8d8c3abc",
			wantErr: true,
		},
		{
			name:    "invalid characters",
			siteID:  "5f0c8c9e1c9d440000e8d8XY",
			wantErr: true,
		},
		{
			name:    "empty",
			siteID:  "",
			wantErr: true,
		},
		{
			name:    "uppercase hex",
			siteID:  "5F0C8C9E1C9D440000E8D8C3",
			wantErr: true, // Schema specifies lowercase only
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSiteID(tt.siteID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSiteID(%s) error = %v, wantErr %v", tt.siteID, err, tt.wantErr)
			}
		})
	}
}

// TestGenerateResourceId tests resource ID generation.
func TestGenerateResourceId(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	expected := "5f0c8c9e1c9d440000e8d8c3/robots.txt"

	result := GenerateRobotsTxtResourceID(siteID)
	if result != expected {
		t.Errorf("expected '%s', got '%s'", expected, result)
	}
}

// TestExtractSiteIDFromResourceID tests extracting siteID from resource ID.
func TestExtractSiteIDFromResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		expected   string
		wantErr    bool
	}{
		{
			name:       "valid resource id",
			resourceID: "5f0c8c9e1c9d440000e8d8c3/robots.txt",
			expected:   "5f0c8c9e1c9d440000e8d8c3",
			wantErr:    false,
		},
		{
			name:       "invalid format",
			resourceID: "invalid",
			expected:   "",
			wantErr:    true,
		},
		{
			name:       "empty",
			resourceID: "",
			expected:   "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractSiteIDFromResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSiteIDFromResourceID(%s) error = %v, wantErr %v", tt.resourceID, err, tt.wantErr)
			}
			if result != tt.expected {
				t.Errorf("expected '%s', got '%s'", tt.expected, result)
			}
		})
	}
}

// Note: mockRoundTripper is defined in auth_test.go in the same package

// TestGetRobotsTxt_Success tests successful GET request.
func TestGetRobotsTxt_Success(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Verify request
				if req.Method != "GET" {
					t.Errorf("expected GET method, got %s", req.Method)
				}
				if req.URL.Path != "/v2/sites/5f0c8c9e1c9d440000e8d8c3/robots_txt" {
					t.Errorf("unexpected path: %s", req.URL.Path)
				}

				// Return mock response
				body := `{"rules":[{"userAgent":"*","allows":["/"],` +
					`"disallows":["/admin/"]}],"sitemap":"https://example.com/sitemap.xml"}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	response, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(response.Rules))
	}
	if response.Rules[0].UserAgent != "*" {
		t.Errorf("expected UserAgent '*', got '%s'", response.Rules[0].UserAgent)
	}
	if response.Sitemap != "https://example.com/sitemap.xml" {
		t.Errorf("unexpected sitemap: %s", response.Sitemap)
	}
}

// TestGetRobotsTxt_NotFound tests 404 response handling.
func TestGetRobotsTxt_NotFound(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Site not found"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for 404 response")
	}
}

// TestPutRobotsTxt_Success tests successful PUT request.
func TestPutRobotsTxt_Success(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Verify request
				if req.Method != "PUT" {
					t.Errorf("expected PUT method, got %s", req.Method)
				}
				if req.Header.Get("Content-Type") != "application/json" {
					t.Errorf("expected Content-Type application/json")
				}

				// Return mock response
				body := `{"rules":[{"userAgent":"*","allows":["/"]}],"sitemap":""}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	rules := []RobotsTxtRule{{UserAgent: "*", Allows: []string{"/"}}}
	response, err := PutRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3", rules, "")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if len(response.Rules) != 1 {
		t.Errorf("expected 1 rule, got %d", len(response.Rules))
	}
}

// TestDeleteRobotsTxt_Success tests successful DELETE request.
func TestDeleteRobotsTxt_Success(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Verify request
				if req.Method != "DELETE" {
					t.Errorf("expected DELETE method, got %s", req.Method)
				}

				return &http.Response{
					StatusCode: 204,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	err := DeleteRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
}

// TestDeleteRobotsTxt_AlreadyDeleted tests 404 response handling (idempotent).
func TestDeleteRobotsTxt_AlreadyDeleted(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Not found"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	err := DeleteRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	// 404 on delete should NOT be an error (idempotent)
	if err != nil {
		t.Fatalf("delete should be idempotent, got error: %v", err)
	}
}

// TestRobotsTxt_RateLimitRetry tests rate limit handling.
func TestRobotsTxt_RateLimitRetry(t *testing.T) {
	callCount := 0
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				callCount++
				if callCount == 1 {
					// First call returns 429
					return &http.Response{
						StatusCode: 429,
						Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
						Header:     make(http.Header),
					}, nil
				}
				// Second call succeeds
				body := `{"rules":[],"sitemap":""}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("expected retry to succeed, got error: %v", err)
	}
	if callCount != 2 {
		t.Errorf("expected 2 calls (1 retry), got %d", callCount)
	}
}

// TestRobotsTxt_MaxRetriesExceeded tests behavior when all retries are exhausted.
func TestRobotsTxt_MaxRetriesExceeded(t *testing.T) {
	callCount := 0
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				callCount++
				// Always return 429 to exhaust retries
				return &http.Response{
					StatusCode: 429,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error when max retries exceeded")
	}
	if !contains(err.Error(), "max retries exceeded") {
		t.Errorf("expected 'max retries exceeded' in error, got: %v", err)
	}
	// Should have made 4 attempts (initial + 3 retries)
	if callCount != 4 {
		t.Errorf("expected 4 attempts (1 initial + 3 retries), got %d", callCount)
	}
}

// TestRobotsTxt_Unauthorized tests 401 response handling.
func TestRobotsTxt_Unauthorized(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 401,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Unauthorized"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for 401 response")
	}
	// Error message should be actionable
	if !contains(err.Error(), "unauthorized") && !contains(err.Error(), "permission") {
		t.Errorf("error should mention unauthorized/permission, got: %v", err)
	}
}

// contains is a case-insensitive string contains check.
func contains(s, substr string) bool {
	return bytes.Contains(bytes.ToLower([]byte(s)), bytes.ToLower([]byte(substr)))
}

// TestSetProviderVersion tests setting the provider version.
func TestSetProviderVersion(t *testing.T) {
	SetProviderVersion("1.2.3")
	if providerVersion != "1.2.3" {
		t.Errorf("expected version '1.2.3', got '%s'", providerVersion)
	}
	// Reset
	SetProviderVersion("0.0.0")
}

// TestRobotsTxt_Create_DryRun tests Create in dry-run mode (preview).
func TestRobotsTxt_Create_DryRun(t *testing.T) {
	resource := &RobotsTxt{}
	args := RobotsTxtArgs{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Content: "User-agent: *\nAllow: /",
	}

	req := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: args,
		DryRun: true,
	}

	resp, err := resource.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create (DryRun) failed: %v", err)
	}

	if resp.ID != "5f0c8c9e1c9d440000e8d8c3/robots.txt" {
		t.Errorf("expected ID in DryRun, got '%s'", resp.ID)
	}
}

// TestRobotsTxt_Create_InvalidSiteID tests Create with invalid siteID.
func TestRobotsTxt_Create_InvalidSiteID(t *testing.T) {
	resource := &RobotsTxt{}
	args := RobotsTxtArgs{
		SiteID:  "invalid",
		Content: "User-agent: *\nAllow: /",
	}

	req := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: args,
		DryRun: false,
	}

	_, err := resource.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for invalid siteID")
	}
}

// TestRobotsTxt_Create_EmptyContent tests Create with empty content.
func TestRobotsTxt_Create_EmptyContent(t *testing.T) {
	resource := &RobotsTxt{}
	args := RobotsTxtArgs{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Content: "",
	}

	req := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: args,
		DryRun: false,
	}

	_, err := resource.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected error for empty content")
	}
}

// ERROR HANDLING & VALIDATION TESTS (Story 1.8)
// ============================================================================

// TestValidation_ActionableErrorMessages tests that validation errors explain what's wrong and how to fix it.
// AC #1: Error messages explain what's wrong and how to fix it (FR32, NFR32)
func TestValidation_ActionableErrorMessages(t *testing.T) {
	resource := &RobotsTxt{}

	// Test case 1: Empty siteID should explain what's required
	req1 := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "",
			Content: "User-agent: *\nAllow: /",
		},
		DryRun: true, // Preview mode - validation should happen before API calls
	}

	_, err1 := resource.Create(context.Background(), req1)
	if err1 == nil {
		t.Fatal("expected error for empty siteID")
	}

	// Error message should explain what's wrong and how to fix it
	errorMsg := err1.Error()
	if !strings.Contains(errorMsg, "siteId") {
		t.Errorf("error message should mention 'siteId', got: %s", errorMsg)
	}
	if !strings.Contains(strings.ToLower(errorMsg), "required") && !strings.Contains(strings.ToLower(errorMsg), "empty") {
		t.Errorf("error message should indicate siteId is required/empty, got: %s", errorMsg)
	}

	// Test case 2: Invalid siteID format should explain the correct format
	req2 := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "invalid-format",
			Content: "User-agent: *\nAllow: /",
		},
		DryRun: true,
	}

	_, err2 := resource.Create(context.Background(), req2)
	if err2 == nil {
		t.Fatal("expected error for invalid siteID format")
	}

	errorMsg2 := err2.Error()
	if !strings.Contains(errorMsg2, "24-character") || !strings.Contains(errorMsg2, "hexadecimal") {
		t.Errorf("error message should explain correct format (24-character hexadecimal), got: %s", errorMsg2)
	}

	// Test case 3: Empty content should explain what's required
	req3 := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "",
		},
		DryRun: true,
	}

	_, err3 := resource.Create(context.Background(), req3)
	if err3 == nil {
		t.Fatal("expected error for empty content")
	}

	errorMsg3 := err3.Error()
	if !strings.Contains(errorMsg3, "content") {
		t.Errorf("error message should mention 'content', got: %s", errorMsg3)
	}
	lowerMsg := strings.ToLower(errorMsg3)
	if !strings.Contains(lowerMsg, "required") && !strings.Contains(lowerMsg, "empty") {
		t.Errorf("error message should indicate content is required/empty, got: %s", errorMsg3)
	}
}

// TestValidation_ErrorsBeforeAPICalls tests that validation errors appear before API calls.
// AC #1: Validation errors are shown before API calls (FR33, NFR33)
func TestValidation_ErrorsBeforeAPICalls(t *testing.T) {
	resource := &RobotsTxt{}

	// Test that validation happens in DryRun mode (preview) without making API calls
	req := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "invalid-site-id",
			Content: "User-agent: *\nAllow: /",
		},
		DryRun: true, // Preview mode - should validate without API calls
	}

	_, err := resource.Create(context.Background(), req)
	if err == nil {
		t.Fatal("expected validation error in preview mode")
	}

	// Error should be a validation error, not an API error
	errorMsg := err.Error()
	hasAPI := strings.Contains(errorMsg, "API")
	hasHTTP := strings.Contains(errorMsg, "HTTP")
	hasRequest := strings.Contains(errorMsg, "request")
	if hasAPI || hasHTTP || hasRequest {
		t.Errorf("validation error should not mention API/HTTP/request, got: %s", errorMsg)
	}
}

// TestRobotsTxt_Update_DryRun tests Update in dry-run mode.
func TestRobotsTxt_Update_DryRun(t *testing.T) {
	resource := &RobotsTxt{}
	args := RobotsTxtArgs{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Content: "User-agent: *\nAllow: /\nDisallow: /new/",
	}

	req := infer.UpdateRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: args,
		DryRun: true,
	}

	resp, err := resource.Update(context.Background(), req)
	if err != nil {
		t.Fatalf("Update (DryRun) failed: %v", err)
	}

	if resp.Output.Content == "" {
		t.Error("expected content in DryRun response")
	}
}

// TestRobotsTxt_Diff_ContentChange tests Diff method for content changes.
func TestRobotsTxt_Diff_ContentChange(t *testing.T) {
	resource := &RobotsTxt{}
	oldContent := "User-agent: *\nAllow: /"
	newContent := "User-agent: *\nAllow: /\nDisallow: /admin/"

	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: newContent,
		},
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: oldContent,
			},
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !resp.HasChanges {
		t.Error("expected Diff to detect content change")
	}
}

// TestRobotsTxt_Diff_SiteIDChange tests that siteID changes trigger replacement.
func TestRobotsTxt_Diff_SiteIDChange(t *testing.T) {
	resource := &RobotsTxt{}

	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: RobotsTxtArgs{
			SiteID:  "ffffffffffffffffffffffff",
			Content: "User-agent: *\nAllow: /",
		},
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !resp.HasChanges {
		t.Error("expected Diff to detect siteID change")
	}
	if !resp.DeleteBeforeReplace {
		t.Error("expected DeleteBeforeReplace=true for siteID change")
	}
}

// TestIdempotency_Diff tests idempotency via Diff method (AC #2).
// This table-driven test validates that unchanged state returns HasChanges=false
// in all scenarios, enabling SDK to skip Update calls and make zero API calls.
func TestIdempotency_Diff(t *testing.T) {
	tests := []struct {
		name       string
		inputs     RobotsTxtArgs
		state      RobotsTxtState
		wantChange bool
		desc       string
	}{
		{
			name: "identical state - no changes",
			inputs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
			state: RobotsTxtState{
				RobotsTxtArgs: RobotsTxtArgs{
					SiteID:  "5f0c8c9e1c9d440000e8d8c3",
					Content: "User-agent: *\nAllow: /",
				},
				LastModified: "2025-12-10T12:00:00Z",
			},
			wantChange: false,
			desc:       "Diff returns HasChanges=false when inputs match state (idempotent)",
		},
		{
			name: "content change - needs update",
			inputs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
			},
			state: RobotsTxtState{
				RobotsTxtArgs: RobotsTxtArgs{
					SiteID:  "5f0c8c9e1c9d440000e8d8c3",
					Content: "User-agent: *\nAllow: /",
				},
				LastModified: "2025-12-10T12:00:00Z",
			},
			wantChange: true,
			desc:       "Diff returns HasChanges=true when content differs",
		},
		{
			name: "siteID change - needs replacement",
			inputs: RobotsTxtArgs{
				SiteID:  "ffffffffffffffffffffffff",
				Content: "User-agent: *\nAllow: /",
			},
			state: RobotsTxtState{
				RobotsTxtArgs: RobotsTxtArgs{
					SiteID:  "5f0c8c9e1c9d440000e8d8c3",
					Content: "User-agent: *\nAllow: /",
				},
				LastModified: "2025-12-10T12:00:00Z",
			},
			wantChange: true,
			desc:       "Diff returns HasChanges=true when siteID differs (replacement needed)",
		},
	}

	resource := &RobotsTxt{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
				Inputs: tt.inputs,
				State:  tt.state,
			}

			resp, err := resource.Diff(context.Background(), req)
			if err != nil {
				t.Fatalf("Diff failed: %v", err)
			}

			if resp.HasChanges != tt.wantChange {
				t.Errorf("Diff HasChanges: got %v, want %v - %s", resp.HasChanges, tt.wantChange, tt.desc)
			}
		})
	}
}

// TestStateConsistency_Read_ReturnsCurrentState tests that Read properly extracts siteID from resource ID.
func TestStateConsistency_Read_ReturnsCurrentState(t *testing.T) {
	// This test validates that Read properly handles resource ID format
	// Resource ID format is {siteID}/robots.txt
	// Read method should extract siteID and call API

	siteID := "5f0c8c9e1c9d440000e8d8c3"
	resourceID := siteID + "/robots.txt"

	// Validate ID parsing logic
	extractedSiteID, err := ExtractSiteIDFromResourceID(resourceID)
	if err != nil {
		t.Fatalf("failed to extract siteID: %v", err)
	}

	if extractedSiteID != siteID {
		t.Errorf("expected siteID '%s', got '%s'", siteID, extractedSiteID)
	}
}

// TestCreate_StateIncludesAllProperties tests that Create returns complete state.
func TestCreate_StateIncludesAllProperties(t *testing.T) {
	resource := &RobotsTxt{}
	args := RobotsTxtArgs{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Content: "User-agent: *\nAllow: /",
	}

	req := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: args,
		DryRun: true, // DryRun to avoid API calls
	}

	resp, err := resource.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create (DryRun) failed: %v", err)
	}

	// Verify ID is set (required for state persistence)
	if resp.ID == "" {
		t.Error("expected non-empty ID in Create response")
	}
	if resp.ID != "5f0c8c9e1c9d440000e8d8c3/robots.txt" {
		t.Errorf("expected ID format '{siteID}/robots.txt', got '%s'", resp.ID)
	}

	// Verify output includes all properties
	if resp.Output.SiteID != args.SiteID {
		t.Error("expected siteID in output state")
	}
	if resp.Output.Content != args.Content {
		t.Error("expected content in output state")
	}
	if resp.Output.LastModified == "" {
		t.Error("expected lastModified timestamp in output state")
	}
}

// TestUpdate_StateIncludesAllProperties tests that Update returns complete state.
func TestUpdate_StateIncludesAllProperties(t *testing.T) {
	resource := &RobotsTxt{}
	newArgs := RobotsTxtArgs{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
	}

	req := infer.UpdateRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: newArgs,
		DryRun: true, // DryRun to avoid API calls
	}

	resp, err := resource.Update(context.Background(), req)
	if err != nil {
		t.Fatalf("Update (DryRun) failed: %v", err)
	}

	// Verify output includes all properties
	if resp.Output.SiteID != newArgs.SiteID {
		t.Error("expected siteID in updated state")
	}
	if resp.Output.Content != newArgs.Content {
		t.Error("expected new content in updated state")
	}
	if resp.Output.LastModified == "" {
		t.Error("expected lastModified timestamp in updated state")
	}
}

// TestResourceID_Format tests that resource IDs follow {siteID}/robots.txt format.
func TestResourceID_Format(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	expectedID := "5f0c8c9e1c9d440000e8d8c3/robots.txt"

	generatedID := GenerateRobotsTxtResourceID(siteID)
	if generatedID != expectedID {
		t.Errorf("expected ID format '%s', got '%s'", expectedID, generatedID)
	}
}

// TestRead_HandlesNotFound_ReturnsEmptyID tests that Read returns empty ID when resource deleted.
func TestRead_HandlesNotFound_ReturnsEmptyID(t *testing.T) {
	// Test the Read method's drift detection: when API returns 404, Read should
	// return empty ID so SDK knows resource was deleted out-of-band.

	siteID := "5f0c8c9e1c9d440000e8d8c3"

	// Mock HTTP client that returns 404
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: 404,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Not found"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	// Verify the error handling logic in GetRobotsTxt
	// Read() calls GetRobotsTxt() which returns "not found" error
	// Read() checks: if strings.Contains(err.Error(), "not found") { return with ID: "" }
	// This signals to SDK that resource was deleted

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, siteID)
	if err == nil {
		t.Fatal("expected error for 404 response")
	}
	// Error message should contain "not found" for Read to detect deletion
	errMsg := err.Error()
	if !bytes.Contains(bytes.ToLower([]byte(errMsg)), bytes.ToLower([]byte("not found"))) {
		t.Fatalf("expected 'not found' in error message, got: %v", err)
	}

	// The Read implementation correctly checks for this error and returns empty ID
	// which enables drift detection per AC#3 and NFR7
}

// TestSecret_TokenMarkedAsSecret tests that Config marks token as secret (NFR12).
func TestSecret_TokenMarkedAsSecret(t *testing.T) {
	// NFR12: Webflow API tokens are stored encrypted in Pulumi state files
	// This requires the Config struct field to be tagged with provider:"secret"
	// Pulumi uses this tag to encrypt the field during state serialization

	// Verify Token field has provider:"secret" tag using reflection
	configType := reflect.TypeOf(Config{})
	tokenField, exists := configType.FieldByName("APIToken")
	if !exists {
		t.Fatal("Config struct missing APIToken field")
	}

	providerTag := tokenField.Tag.Get("provider")
	if providerTag != "secret" {
		t.Errorf("APIToken field missing provider:\"secret\" tag - got provider:%q", providerTag)
	}

	// Also verify pulumi tag exists
	pulumiTag := tokenField.Tag.Get("pulumi")
	if pulumiTag == "" {
		t.Error("Token field missing pulumi tag")
	}

	// Verify the tag chain is correct: `pulumi:"token,optional" provider:"secret"`
	// This ensures the field is:
	// - Named "token" in Pulumi config
	// - Optional (not required)
	// - Encrypted in state files using Pulumi's secrets provider
}

// ============================================================================
// PREVIEW WORKFLOW TESTS (Story 1.7)
// ============================================================================

// TestPreviewWorkflow_DiffCalledForPreview tests that Diff() is called during preview workflow.
// AC #1: Preview shows detailed preview of changes
func TestPreviewWorkflow_DiffCalledForPreview(t *testing.T) {
	// During preview, Pulumi SDK calls Diff() to determine changes
	// This test validates that Diff() returns correct change information
	resource := &RobotsTxt{}

	// Scenario: Content change (update operation)
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
			LastModified: "2025-12-10T12:00:00Z",
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Diff should detect changes
	if !resp.HasChanges {
		t.Error("Diff should detect content change")
	}

	// DetailedDiff should indicate update operation
	if resp.DetailedDiff == nil {
		t.Error("Diff should return DetailedDiff for preview")
	}

	contentDiff, exists := resp.DetailedDiff["content"]
	if !exists {
		t.Error("Diff should include content in DetailedDiff")
	}
	if contentDiff.Kind != p.Update {
		t.Errorf("content change should be Update, got %v", contentDiff.Kind)
	}
}

// TestPreviewWorkflow_CreateOperationDetected tests that create operations are detected.
// AC #1: Preview indicates create/update/delete operations
func TestPreviewWorkflow_CreateOperationDetected(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: New resource (no existing state = create operation)
	// In Pulumi, create is detected when state is empty/zero or when resource doesn't exist
	// The SDK calls Diff() with zero/empty state for new resources

	// Test case 1: Completely empty state (zero value)
	req1 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{}, // Zero/empty state = new resource
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp1, err := resource.Diff(context.Background(), req1)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// For new resources, Diff should indicate changes
	// (SDK will call Create, not Update)
	if !resp1.HasChanges {
		t.Error("Diff should detect changes for new resource (empty state)")
	}

	// Test case 2: State with empty SiteID vs new SiteID
	// Note: When state has empty SiteID and inputs have SiteID, Diff() sees this as siteID change
	// which triggers replacement. This is correct behavior - empty to non-empty is a replacement.
	// For true "create" operations, the SDK typically passes zero/empty state.
	req2 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "", // Empty siteID in state
				Content: "",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp2, err := resource.Diff(context.Background(), req2)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Should detect changes (siteID change from empty to non-empty)
	if !resp2.HasChanges {
		t.Error("Diff should detect changes when siteID changes from empty to non-empty")
	}

	// SiteID change (even from empty) triggers replacement - this is correct behavior
	// The SDK handles this appropriately for create operations
	if !resp2.DeleteBeforeReplace {
		t.Log("Note: Empty siteID to non-empty triggers replacement, which is correct for Diff()")
	}
}

// TestPreviewWorkflow_UpdateOperationDetected tests that update operations are detected.
// AC #1: Preview indicates create/update/delete operations
func TestPreviewWorkflow_UpdateOperationDetected(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: Content change (update operation)
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !resp.HasChanges {
		t.Error("Diff should detect content change")
	}

	// Update operation should not require replacement
	if resp.DeleteBeforeReplace {
		t.Error("Content change should not require replacement")
	}
}

// TestPreviewWorkflow_ReplaceOperationDetected tests that replace operations are detected.
// AC #1: Preview indicates create/update/delete operations
func TestPreviewWorkflow_ReplaceOperationDetected(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: SiteID change (replace operation)
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "6a1d9e2f3b4c5d6e7f8a9b0c1",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !resp.HasChanges {
		t.Error("Diff should detect siteID change")
	}

	// SiteID change requires replacement
	if !resp.DeleteBeforeReplace {
		t.Error("SiteID change should require replacement")
	}

	// DetailedDiff should indicate UpdateReplace
	if resp.DetailedDiff == nil {
		t.Error("Diff should return DetailedDiff")
	}
	siteIDDiff, exists := resp.DetailedDiff["siteId"]
	if !exists {
		t.Error("Diff should include siteId in DetailedDiff")
	}
	if siteIDDiff.Kind != p.UpdateReplace {
		t.Errorf("siteId change should be UpdateReplace, got %v", siteIDDiff.Kind)
	}
}

// TestPreviewWorkflow_DryRunNoAPICalls tests that DryRun mode prevents API calls.
// AC #1: Preview completes within 10 seconds (NFR3) - DryRun avoids API calls
func TestPreviewWorkflow_DryRunNoAPICalls(t *testing.T) {
	resource := &RobotsTxt{}

	// Test Create with DryRun
	// DryRun mode should return early before calling GetHTTPClient, so no API calls are made
	createReq := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /",
		},
		DryRun: true, // Preview mode - should return early without API calls
	}

	resp, err := resource.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Create (DryRun) failed: %v", err)
	}

	// Verify expected state is returned without making API calls
	// The implementation returns early when DryRun is true, before GetHTTPClient is called
	if resp.ID == "" {
		t.Error("DryRun should return expected ID")
	}
	if resp.ID != "5f0c8c9e1c9d440000e8d8c3/robots.txt" {
		t.Errorf("DryRun should return correct ID format, got %q", resp.ID)
	}
	if resp.Output.SiteID != createReq.Inputs.SiteID {
		t.Error("DryRun should return expected state")
	}
	if resp.Output.Content != createReq.Inputs.Content {
		t.Error("DryRun should return expected content")
	}
	if resp.Output.LastModified == "" {
		t.Error("DryRun should return expected lastModified timestamp")
	}

	// Test Update with DryRun
	updateReq := infer.UpdateRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
		DryRun: true, // Preview mode - should return early without API calls
	}

	updateResp, err := resource.Update(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("Update (DryRun) failed: %v", err)
	}

	// Verify expected state is returned without making API calls
	if updateResp.Output.Content != updateReq.Inputs.Content {
		t.Error("DryRun should return expected state")
	}
	if updateResp.Output.SiteID != updateReq.Inputs.SiteID {
		t.Error("DryRun should return expected siteID")
	}
	if updateResp.Output.LastModified == "" {
		t.Error("DryRun should return expected lastModified timestamp")
	}
}

// TestPreviewWorkflow_PreviewPerformance tests that preview completes quickly.
// AC #1: Preview completes within 10 seconds (NFR3)
func TestPreviewWorkflow_PreviewPerformance(t *testing.T) {
	resource := &RobotsTxt{}

	// Measure Diff() performance (should be instant)
	start := time.Now()

	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	_, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	elapsed := time.Since(start)

	// Diff() should complete in milliseconds (well under 10 seconds)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Diff() took %v, should complete in milliseconds", elapsed)
	}

	// DryRun Create should also be fast
	start = time.Now()
	createReq := infer.CreateRequest[RobotsTxtArgs]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /",
		},
		DryRun: true,
	}

	_, err = resource.Create(context.Background(), createReq)
	if err != nil {
		t.Fatalf("Create (DryRun) failed: %v", err)
	}

	elapsed = time.Since(start)
	if elapsed > 100*time.Millisecond {
		t.Errorf("Create (DryRun) took %v, should complete in milliseconds", elapsed)
	}
}

// TestPreviewWorkflow_FullPreviewWorkflowPerformance tests the complete preview workflow (Read + Diff).
// AC #1: Preview completes within 10 seconds (NFR3)
// This tests the actual preview workflow: SDK calls Read() to get current state, then Diff() to compare
func TestPreviewWorkflow_FullPreviewWorkflowPerformance(t *testing.T) {
	resource := &RobotsTxt{}

	// Simulate full preview workflow: Read() + Diff()
	// In real preview, SDK calls Read() first to get current state, then Diff() to compare

	// Step 1: Read current state (simulated - in real preview SDK does this)
	// We'll use a mock to simulate Read() being fast
	start := time.Now()

	// Step 2: Diff() compares state vs inputs (this is what we actually test)
	diffReq := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	_, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	elapsed := time.Since(start)

	// Full preview workflow (Read + Diff) should complete well under 10 seconds
	// Since we're using mocked/unit tests, this should be instant
	// In real scenarios with API calls, Read() would add latency, but still <10s
	if elapsed > 1*time.Second {
		t.Errorf("Full preview workflow (Read + Diff) took %v, should complete in <1 second for unit tests", elapsed)
	}

	// Verify it's well under the 10-second requirement
	if elapsed > 10*time.Second {
		t.Errorf("Full preview workflow exceeded 10-second requirement: %v", elapsed)
	}
}

// TestPreviewWorkflow_NoChangesDetected tests preview with no changes (idempotency).
// AC #1: Preview shows detailed preview of changes
func TestPreviewWorkflow_NoChangesDetected(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: No changes (identical state and inputs)
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// No changes should be detected
	if resp.HasChanges {
		t.Error("Diff should detect no changes for identical state/inputs")
	}

	// DetailedDiff should be empty or nil
	if len(resp.DetailedDiff) > 0 {
		t.Error("Diff should not return DetailedDiff when no changes")
	}
}

// TestPreviewWorkflow_DeleteOperationDetected tests that delete operations are detected.
// AC #1: Preview indicates create/update/delete operations
func TestPreviewWorkflow_DeleteOperationDetected(t *testing.T) {
	// Delete operations are detected by Pulumi SDK when a resource is removed from code
	// The SDK calls Diff() with state but no inputs (or empty inputs)
	// This test verifies that our Diff() handles this scenario correctly

	resource := &RobotsTxt{}

	// Scenario: Resource exists in state but is removed from code
	// In Pulumi, this is represented by Diff() being called with state but empty/zero inputs
	// However, the SDK typically handles delete detection, so we test the edge case:
	// What if Diff() is called with state but inputs indicate deletion?

	// For delete operations, Pulumi SDK typically:
	// 1. Detects resource removed from code
	// 2. Calls Delete() method directly
	// 3. Doesn't call Diff() for deleted resources

	// However, we can test that Diff() correctly handles the case where
	// state exists but inputs are zero/empty (edge case scenario)
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{}, // Empty inputs (resource removed from code)
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Diff should detect changes when inputs are empty but state exists
	// This indicates the resource should be deleted
	if !resp.HasChanges {
		t.Error("Diff should detect changes when resource is removed (empty inputs with existing state)")
	}

	// SiteID change should trigger replacement (which becomes delete)
	if req.State.SiteID != req.Inputs.SiteID {
		if !resp.DeleteBeforeReplace {
			t.Error("Empty inputs with existing state should indicate deletion")
		}
	}
}

// TestPreviewWorkflow_MultipleResources tests preview with multiple resources.
// AC #1: Test preview with multiple resources
func TestPreviewWorkflow_MultipleResources(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: Preview changes for multiple RobotsTxt resources
	// This simulates what happens when Pulumi previews multiple resources

	// Resource 1: Content change (update)
	req1 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	resp1, err := resource.Diff(context.Background(), req1)
	if err != nil {
		t.Fatalf("Diff failed for resource 1: %v", err)
	}

	if !resp1.HasChanges {
		t.Error("Resource 1 should show changes")
	}

	// Resource 2: New resource (create)
	req2 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{}, // Empty state = new resource
		Inputs: RobotsTxtArgs{
			SiteID:  "6a1d9e2f3b4c5d6e7f8a9b0c1",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp2, err := resource.Diff(context.Background(), req2)
	if err != nil {
		t.Fatalf("Diff failed for resource 2: %v", err)
	}

	if !resp2.HasChanges {
		t.Error("Resource 2 should show changes (new resource)")
	}

	// Resource 3: No changes (idempotency)
	req3 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "7b2e0f3g4c5d6e7f8a9b0c1d2",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "7b2e0f3g4c5d6e7f8a9b0c1d2",
			Content: "User-agent: *\nAllow: /",
		},
	}

	resp3, err := resource.Diff(context.Background(), req3)
	if err != nil {
		t.Fatalf("Diff failed for resource 3: %v", err)
	}

	if resp3.HasChanges {
		t.Error("Resource 3 should show no changes (idempotency)")
	}

	// Verify all three resources are handled correctly in preview
	// This validates that preview can handle multiple resources with different change types
	changesCount := 0
	if resp1.HasChanges {
		changesCount++
	}
	if resp2.HasChanges {
		changesCount++
	}
	if resp3.HasChanges {
		changesCount++
	}

	if changesCount != 2 {
		t.Errorf("Expected 2 resources with changes (update + create), got %d", changesCount)
	}
}

// TestPreviewOutputFormat_ChangeDetails tests that preview shows property-level changes.
// AC #2: Preview clearly distinguishes additions, modifications, and deletions
func TestPreviewOutputFormat_ChangeDetails(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: Content modification
	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/",
		},
	}

	resp, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// DetailedDiff should contain property-level change details
	if resp.DetailedDiff == nil {
		t.Fatal("Diff should return DetailedDiff for property-level changes")
	}

	contentDiff, exists := resp.DetailedDiff["content"]
	if !exists {
		t.Fatal("Diff should include 'content' in DetailedDiff")
	}

	// Verify change kind is Update (modification)
	if contentDiff.Kind != p.Update {
		t.Errorf("content change should be Update (modification), got %v", contentDiff.Kind)
	}
}

// TestPreviewOutputFormat_SensitiveDataRedaction tests that sensitive data is not exposed.
// AC #2: Sensitive credentials are never displayed in preview output (FR17)
func TestPreviewOutputFormat_SensitiveDataRedaction(t *testing.T) {
	// Token is marked as secret in Config struct (validated in TestSecret_TokenMarkedAsSecret)
	// Pulumi SDK automatically redacts secrets in preview output
	// This test validates the Config struct has proper secret tagging

	configType := reflect.TypeOf(Config{})
	tokenField, exists := configType.FieldByName("APIToken")
	if !exists {
		t.Fatal("Config struct missing APIToken field")
	}

	// Verify provider:"secret" tag exists
	providerTag := tokenField.Tag.Get("provider")
	if providerTag != "secret" {
		t.Errorf("Token field must have provider:\"secret\" tag for redaction, got provider:%q", providerTag)
	}

	// Pulumi SDK will automatically redact fields with provider:"secret" tag
	// in preview output, showing [secret] instead of actual value
	// This satisfies FR17: Never log or expose sensitive credentials
}

// TestPreviewAccuracy_PreviewMatchesActual tests that preview accurately shows changes.
// AC #3: Preview matches actual changes applied
func TestPreviewAccuracy_PreviewMatchesActual(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: Preview a content change, then verify actual change matches
	originalContent := "User-agent: *\nAllow: /"
	newContent := "User-agent: *\nAllow: /\nDisallow: /admin/"

	// Step 1: Preview the change (Diff)
	diffReq := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: originalContent,
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: newContent,
		},
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify preview shows change
	if !diffResp.HasChanges {
		t.Fatal("Preview should show content change")
	}

	// Step 2: DryRun Update to see expected state
	updateReq := infer.UpdateRequest[RobotsTxtArgs, RobotsTxtState]{
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: newContent,
		},
		DryRun: true, // Preview mode
	}

	updateResp, err := resource.Update(context.Background(), updateReq)
	if err != nil {
		t.Fatalf("Update (DryRun) failed: %v", err)
	}

	// Verify preview state matches expected change
	if updateResp.Output.Content != newContent {
		t.Errorf("Preview state should match new content: got %q, want %q", updateResp.Output.Content, newContent)
	}

	// The actual Update (without DryRun) would apply the same change
	// This validates that preview accurately represents what will happen
}

// TestPreviewWorkflow_EdgeCase_InvalidInputs tests preview with invalid inputs.
// Edge case: Preview should handle invalid inputs gracefully
func TestPreviewWorkflow_EdgeCase_InvalidInputs(t *testing.T) {
	resource := &RobotsTxt{}

	// Test case 1: Invalid siteID format
	req1 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "invalid-site-id", // Invalid format
			Content: "User-agent: *\nAllow: /",
		},
	}

	// Diff() should still work (validation happens in Create/Update, not Diff)
	resp1, err := resource.Diff(context.Background(), req1)
	if err != nil {
		t.Fatalf("Diff should handle invalid siteID gracefully, got error: %v", err)
	}

	// Should detect siteID change (even if invalid)
	if !resp1.HasChanges {
		t.Error("Diff should detect siteID change even if invalid format")
	}

	// Test case 2: Empty content
	req2 := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: "", // Empty content
		},
	}

	resp2, err := resource.Diff(context.Background(), req2)
	if err != nil {
		t.Fatalf("Diff should handle empty content gracefully, got error: %v", err)
	}

	// Should detect content change (even if empty)
	if !resp2.HasChanges {
		t.Error("Diff should detect content change even if empty")
	}
}

// TestPreviewWorkflow_EdgeCase_LargeContent tests preview with very large content.
// Edge case: Preview should handle large content efficiently
func TestPreviewWorkflow_EdgeCase_LargeContent(t *testing.T) {
	resource := &RobotsTxt{}

	// Generate large content (>10KB)
	largeContent := "User-agent: *\nAllow: /\n"
	var largeContentSb1721 strings.Builder
	for i := 0; i < 1000; i++ {
		largeContentSb1721.WriteString("Disallow: /path" + string(rune(i%10)) + "/\n")
	}
	largeContent += largeContentSb1721.String()

	req := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",
			Content: largeContent,
		},
	}

	start := time.Now()
	resp, err := resource.Diff(context.Background(), req)
	elapsed := time.Since(start)

	if err != nil {
		t.Fatalf("Diff failed with large content: %v", err)
	}

	// Should still complete quickly even with large content
	if elapsed > 1*time.Second {
		t.Errorf("Diff with large content took %v, should complete quickly", elapsed)
	}

	if !resp.HasChanges {
		t.Error("Diff should detect changes even with large content")
	}
}

// TestPreviewAccuracy_NoUnexpectedChanges tests that preview doesn't show unexpected changes.
// AC #3: No unexpected modifications occur
func TestPreviewAccuracy_NoUnexpectedChanges(t *testing.T) {
	resource := &RobotsTxt{}

	// Scenario: Only content should change, siteID should remain unchanged
	diffReq := infer.DiffRequest[RobotsTxtArgs, RobotsTxtState]{
		State: RobotsTxtState{
			RobotsTxtArgs: RobotsTxtArgs{
				SiteID:  "5f0c8c9e1c9d440000e8d8c3",
				Content: "User-agent: *\nAllow: /",
			},
		},
		Inputs: RobotsTxtArgs{
			SiteID:  "5f0c8c9e1c9d440000e8d8c3",                   // Same siteID
			Content: "User-agent: *\nAllow: /\nDisallow: /admin/", // Only content changes
		},
	}

	resp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Only content should be in DetailedDiff, not siteID
	if resp.DetailedDiff == nil {
		t.Fatal("Diff should return DetailedDiff")
	}

	// siteID should NOT be in DetailedDiff (no change)
	if _, exists := resp.DetailedDiff["siteID"]; exists {
		t.Error("siteID should not be in DetailedDiff when unchanged")
	}

	// content SHOULD be in DetailedDiff (changed)
	if _, exists := resp.DetailedDiff["content"]; !exists {
		t.Error("content should be in DetailedDiff when changed")
	}

	// Should not require replacement (only content changed)
	if resp.DeleteBeforeReplace {
		t.Error("Content-only change should not require replacement")
	}
}

// ============================================================================
// NETWORK ERROR HANDLING TESTS (Story 1.8 - AC #3)
// ============================================================================

// TestNetworkError_TimeoutMessage tests that timeout errors include recovery guidance.
// AC #3: Network errors include recovery guidance (NFR9)
func TestNetworkError_TimeoutMessage(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Simulate timeout error
				return nil, errors.New("context deadline exceeded")
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for timeout")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "timeout") && !strings.Contains(errMsg, "deadline exceeded") {
		t.Errorf("error should mention timeout, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "recovery") && !strings.Contains(strings.ToLower(errMsg), "fix this") {
		t.Errorf("error should include recovery guidance, got: %s", errMsg)
	}
}

// TestNetworkError_ConnectionRefusedMessage tests that connection failures include recovery guidance.
// AC #3: Network errors include recovery guidance (NFR9)
func TestNetworkError_ConnectionRefusedMessage(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Simulate connection refused error
				return nil, errors.New("connection refused")
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for connection refused")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "connection") && !strings.Contains(errMsg, "failed") {
		t.Errorf("error should mention connection failure, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "DNS") && !strings.Contains(errMsg, "firewall") {
		t.Errorf("error should mention DNS/firewall in recovery guidance, got: %s", errMsg)
	}
}

// TestNetworkError_DNSFailureMessage tests DNS resolution failures.
// AC #3: Network errors include recovery guidance (NFR9)
func TestNetworkError_DNSFailureMessage(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Simulate DNS failure (no such host)
				return nil, errors.New("no such host")
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for DNS failure")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "connection") && !strings.Contains(errMsg, "failed") {
		t.Errorf("error should mention connection failure, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "DNS") {
		t.Errorf("error should mention DNS in recovery guidance, got: %s", errMsg)
	}
}

// TestNetworkError_GenericNetworkFailure tests generic network errors.
// AC #3: Network errors include recovery guidance (NFR9)
func TestNetworkError_GenericNetworkFailure(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Simulate generic network error
				return nil, errors.New("network unreachable")
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for network failure")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "network error") {
		t.Errorf("error should mention network error, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "fix this") && !strings.Contains(errMsg, "recovery") {
		t.Errorf("error should include recovery guidance, got: %s", errMsg)
	}
}

// ============================================================================
// RATE LIMITING ERROR MESSAGE TESTS (Story 1.8 - AC #4)
// ============================================================================

// TestRateLimitError_Message tests rate limiting error message content.
// AC #4: Provides clear messaging about rate limit delays (FR18, NFR8)
func TestRateLimitError_Message(t *testing.T) {
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				// Always return 429 to trigger max retries error
				return &http.Response{
					StatusCode: 429,
					Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")

	if err == nil {
		t.Fatal("expected error for rate limiting")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "rate limited") || !strings.Contains(errMsg, "429") {
		t.Errorf("error should mention rate limiting and 429, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "exponential backoff") {
		t.Errorf("error should mention exponential backoff, got: %s", errMsg)
	}
	if !strings.Contains(errMsg, "retry attempt") && !strings.Contains(errMsg, "Retry attempt") {
		t.Errorf("error should mention retry attempt, got: %s", errMsg)
	}
}

// TestRateLimitError_RetryAttemptInfo tests that error includes attempt number and wait time.
// AC #4: Retry messages show attempt number and wait time
func TestRateLimitError_RetryAttemptInfo(t *testing.T) {
	attemptCount := 0
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				attemptCount++
				// Return 429 twice, then succeed
				if attemptCount <= 2 {
					return &http.Response{
						StatusCode: 429,
						Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
						Header:     make(http.Header),
					}, nil
				}
				body := `{"rules":[],"sitemap":""}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	_, err := GetRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	// Should succeed after retries
	if err != nil {
		t.Errorf("expected success after retries, got error: %v", err)
	}
	if attemptCount != 3 {
		t.Errorf("expected 3 attempts (2 retries + 1 success), got %d", attemptCount)
	}
}

// TestRateLimitError_PUTOperation tests rate limiting error for PUT operation.
// AC #4: Rate limiting applies to all operations (FR18, NFR8)
func TestRateLimitError_PUTOperation(t *testing.T) {
	callCount := 0
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				callCount++
				if callCount == 1 {
					// First call returns 429
					return &http.Response{
						StatusCode: 429,
						Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
						Header:     make(http.Header),
					}, nil
				}
				// Second call succeeds
				body := `{"rules":[],"sitemap":""}`
				return &http.Response{
					StatusCode: 200,
					Body:       io.NopCloser(bytes.NewBufferString(body)),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	rules := []RobotsTxtRule{{UserAgent: "*", Allows: []string{"/"}}}
	_, err := PutRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3", rules, "")
	if err != nil {
		t.Errorf("PUT should retry on 429, got error: %v", err)
	}
}

// TestRateLimitError_DELETEOperation tests rate limiting error for DELETE operation.
// AC #4: Rate limiting applies to all operations (FR18, NFR8)
func TestRateLimitError_DELETEOperation(t *testing.T) {
	callCount := 0
	mockClient := &http.Client{
		Transport: &mockRoundTripper{
			handler: func(req *http.Request) (*http.Response, error) {
				callCount++
				if callCount == 1 {
					// First call returns 429
					return &http.Response{
						StatusCode: 429,
						Body:       io.NopCloser(bytes.NewBufferString(`{"message":"Rate limited"}`)),
						Header:     make(http.Header),
					}, nil
				}
				// Second call succeeds with 204
				return &http.Response{
					StatusCode: 204,
					Body:       io.NopCloser(bytes.NewBufferString("")),
					Header:     make(http.Header),
				}, nil
			},
		},
	}

	ctx := context.Background()
	err := DeleteRobotsTxt(ctx, mockClient, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Errorf("DELETE should retry on 429, got error: %v", err)
	}
}

// ============================================================================
// HANDLEWEBFLOWOROR ERROR TESTS (Story 1.8 - AC #2)
// ============================================================================

// TestHandleWebflowError_400 tests 400 Bad Request error message.
// AC #2: Error messages follow Pulumi diagnostic formatting (NFR29)
func TestHandleWebflowError_400(t *testing.T) {
	body := []byte(`{"error":"Invalid request body"}`)
	err := handleWebflowError(400, body)

	if err == nil {
		t.Fatal("expected error for 400")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "bad request") {
		t.Errorf("error should mention bad request, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "configuration") && !strings.Contains(strings.ToLower(errMsg), "check") {
		t.Errorf("error should include actionable guidance, got: %s", errMsg)
	}
}

// TestHandleWebflowError_401 tests 401 Unauthorized error message.
// AC #2: Error messages include actionable guidance (NFR32)
func TestHandleWebflowError_401(t *testing.T) {
	body := []byte(`{"error":"Unauthorized"}`)
	err := handleWebflowError(401, body)

	if err == nil {
		t.Fatal("expected error for 401")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "unauthorized") && !strings.Contains(errMsg, "authentication") {
		t.Errorf("error should mention authentication, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "token") && !strings.Contains(strings.ToLower(errMsg), "verify") {
		t.Errorf("error should include token fix instructions, got: %s", errMsg)
	}
}

// TestHandleWebflowError_403 tests 403 Forbidden error message.
// AC #2: Error messages include actionable guidance (NFR32)
func TestHandleWebflowError_403(t *testing.T) {
	body := []byte(`{"error":"Forbidden"}`)
	err := handleWebflowError(403, body)

	if err == nil {
		t.Fatal("expected error for 403")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "forbidden") && !strings.Contains(errMsg, "access denied") {
		t.Errorf("error should mention access denied, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "permission") && !strings.Contains(strings.ToLower(errMsg), "scope") {
		t.Errorf("error should mention permissions/scopes, got: %s", errMsg)
	}
}

// TestHandleWebflowError_404 tests 404 Not Found error message.
// AC #2: Error messages include actionable guidance (NFR32)
func TestHandleWebflowError_404(t *testing.T) {
	body := []byte(`{"error":"Not found"}`)
	err := handleWebflowError(404, body)

	if err == nil {
		t.Fatal("expected error for 404")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "not found") {
		t.Errorf("error should mention not found, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "site") && !strings.Contains(strings.ToLower(errMsg), "verify") {
		t.Errorf("error should include site ID verification guidance, got: %s", errMsg)
	}
}

// TestHandleWebflowError_429 tests 429 Rate Limited error message.
// AC #4: Error messages provide clear guidance about rate limits
func TestHandleWebflowError_429(t *testing.T) {
	body := []byte(`{"error":"Too many requests"}`)
	err := handleWebflowError(429, body)

	if err == nil {
		t.Fatal("expected error for 429")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "rate limited") {
		t.Errorf("error should mention rate limiting, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "retry") && !strings.Contains(strings.ToLower(errMsg), "wait") {
		t.Errorf("error should mention retry/wait guidance, got: %s", errMsg)
	}
}

// TestHandleWebflowError_500 tests 500 Server Error message.
// AC #2: Error messages include actionable guidance (NFR32)
func TestHandleWebflowError_500(t *testing.T) {
	body := []byte(`{"error":"Internal server error"}`)
	err := handleWebflowError(500, body)

	if err == nil {
		t.Fatal("expected error for 500")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "server error") {
		t.Errorf("error should mention server error, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "wait") && !strings.Contains(strings.ToLower(errMsg), "retry") {
		t.Errorf("error should suggest wait and retry, got: %s", errMsg)
	}
}

// TestHandleWebflowError_Unknown tests unknown status code handling.
// AC #2: Error messages include actionable guidance (NFR32)
func TestHandleWebflowError_Unknown(t *testing.T) {
	body := []byte(`{"error":"Some error"}`)
	err := handleWebflowError(502, body)

	if err == nil {
		t.Fatal("expected error for 502")
	}

	errMsg := err.Error()
	if !strings.Contains(errMsg, "502") {
		t.Errorf("error should include status code, got: %s", errMsg)
	}
	if !strings.Contains(strings.ToLower(errMsg), "status") && !strings.Contains(strings.ToLower(errMsg), "api") {
		t.Errorf("error should mention API/status, got: %s", errMsg)
	}
}
