// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestValidateDisplayName_Valid tests valid display name inputs.
func TestValidateDisplayName_Valid(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
	}{
		{"simple name", "My Site"},
		{"name with multiple words", "My Marketing Site"},
		{"name with numbers", "Company Blog 2024"},
		{"name with special characters", "Joe's Restaurant & Bar"},
		{"long valid name", "This is a very long but valid site name with many words"},
		{"max length name", "a" + string(make([]byte, 254))}, // 255 chars
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDisplayName(tt.displayName)
			if err != nil {
				t.Errorf("ValidateDisplayName(%q) returned unexpected error: %v", tt.displayName, err)
			}
		})
	}
}

// TestValidateDisplayName_Empty tests empty display name (required field).
func TestValidateDisplayName_Empty(t *testing.T) {
	err := ValidateDisplayName("")
	if err == nil {
		t.Error("ValidateDisplayName(\"\") should return error for empty string")
	}
	if err != nil {
		expectedErr := "displayName is required but was not provided. " +
			"Expected format: A non-empty string representing your site's name. " +
			"Fix: Provide a name for your site (e.g., 'My Marketing Site', 'Company Blog', 'Product Landing Page')"
		if err.Error() != expectedErr {
			t.Errorf("ValidateDisplayName(\"\") returned unexpected error message: %v", err)
		}
	}
}

// TestValidateDisplayName_TooLong tests display name exceeding max length.
func TestValidateDisplayName_TooLong(t *testing.T) {
	tooLongName := "a" + string(make([]byte, 255)) // 256 chars
	err := ValidateDisplayName(tooLongName)
	if err == nil {
		t.Error("ValidateDisplayName with 256 characters should return error")
	}
	if err != nil && !strings.Contains(err.Error(), "too long") {
		t.Errorf("ValidateDisplayName error should mention 'too long', got: %v", err)
	}
}

// TestValidateShortName_Valid tests valid short name inputs.
func TestValidateShortName_Valid(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
	}{
		{"simple shortName", "my-site"},
		{"shortName with numbers", "company-blog-2024"},
		{"single word", "blog"},
		{"numbers only", "123"},
		{"mixed alphanumeric", "site123abc"},
		{"multiple hyphens", "my-great-site-example"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortName(tt.shortName)
			if err != nil {
				t.Errorf("ValidateShortName(%q) returned unexpected error: %v", tt.shortName, err)
			}
		})
	}
}

// TestValidateShortName_Empty tests empty short name (optional field, should pass).
func TestValidateShortName_Empty(t *testing.T) {
	err := ValidateShortName("")
	if err != nil {
		t.Errorf("ValidateShortName(\"\") should return nil for empty string (optional field), got: %v", err)
	}
}

// TestValidateShortName_Uppercase tests uppercase characters (invalid).
func TestValidateShortName_Uppercase(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
	}{
		{"all uppercase", "MY-SITE"},
		{"mixed case", "My-Site"},
		{"single uppercase", "A"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortName(tt.shortName)
			if err == nil {
				t.Errorf("ValidateShortName(%q) should return error for uppercase", tt.shortName)
			}
			if err != nil && !strings.Contains(err.Error(), "lowercase") {
				t.Errorf("ValidateShortName error should mention 'lowercase', got: %v", err)
			}
		})
	}
}

// TestValidateShortName_InvalidCharacters tests special characters and spaces.
func TestValidateShortName_InvalidCharacters(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
	}{
		{"with spaces", "my site"},
		{"with underscores", "my_site"},
		{"with dots", "my.site"},
		{"with special chars", "my@site"},
		{"with query string", "my-site?test=1"},
		{"with hash", "my-site#anchor"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortName(tt.shortName)
			if err == nil {
				t.Errorf("ValidateShortName(%q) should return error for invalid characters", tt.shortName)
			}
			if err != nil && !strings.Contains(err.Error(), "invalid characters") {
				t.Errorf("ValidateShortName error should mention 'invalid characters', got: %v", err)
			}
		})
	}
}

// TestValidateShortName_LeadingTrailingHyphens tests hyphens at start/end (invalid).
func TestValidateShortName_LeadingTrailingHyphens(t *testing.T) {
	tests := []struct {
		name      string
		shortName string
	}{
		{"leading hyphen", "-my-site"},
		{"trailing hyphen", "my-site-"},
		{"both", "-my-site-"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateShortName(tt.shortName)
			if err == nil {
				t.Errorf("ValidateShortName(%q) should return error for leading/trailing hyphens", tt.shortName)
			}
			if err != nil && !strings.Contains(err.Error(), "leading/trailing") {
				t.Errorf("ValidateShortName error should mention leading/trailing hyphens, got: %v", err)
			}
		})
	}
}

// Note: ValidateTimeZone tests were removed because timezone is now read-only
// (output only). The Webflow API does not support setting timezone via API.

// TestValidateWorkspaceID_Valid tests valid workspace ID inputs.
func TestValidateWorkspaceID_Valid(t *testing.T) {
	tests := []struct {
		name        string
		workspaceID string
	}{
		{"hex ID", "5f0c8c9e1c9d440000e8d8c3"},
		{"another hex ID", "abc123def456789012345678"},
		{"shorter valid ID", "workspace-id-123"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateWorkspaceID(tt.workspaceID)
			if err != nil {
				t.Errorf("ValidateWorkspaceID(%q) returned unexpected error: %v", tt.workspaceID, err)
			}
		})
	}
}

// TestValidateWorkspaceID_Empty tests empty workspace ID (required field).
func TestValidateWorkspaceID_Empty(t *testing.T) {
	err := ValidateWorkspaceID("")
	if err == nil {
		t.Error("ValidateWorkspaceID(\"\") should return error for empty string")
	}
	if err != nil && !strings.Contains(err.Error(), "required") {
		t.Errorf("ValidateWorkspaceID error should mention 'required', got: %v", err)
	}
}

// Note: GenerateSiteResourceId and ExtractIdsFromSiteResourceId functions were removed
// as the Site resource now uses simple siteId as the resource ID instead of {workspaceId}/sites/{siteId}

// ============================================================================
// PostSite API Function Tests
// ============================================================================

// TestPostSite_Success tests successful site creation with all fields
func TestPostSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v2/workspaces/") || !strings.Contains(r.URL.Path, "/sites") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Parse request body
		body, _ := io.ReadAll(r.Body)
		var reqBody SiteCreateRequest
		_ = json.Unmarshal(body, &reqBody)

		// Verify request body mapping (name in request)
		if reqBody.Name != "My Test Site" {
			t.Errorf("Expected name 'My Test Site', got '%s'", reqBody.Name)
		}
		if reqBody.ParentFolderID != "folder123" {
			t.Errorf("Expected parentFolderId 'folder123', got '%s'", reqBody.ParentFolderID)
		}

		// Return mock Site response (displayName in response)
		response := Site{
			ID:          "site123",
			WorkspaceID: "workspace456",
			DisplayName: reqBody.Name, // Maps name → displayName
			ShortName:   "my-test-site",
			TimeZone:    "America/New_York",
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override API base URL for testing
	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PostSite(context.Background(), client, "workspace456", "My Test Site", "folder123", "")
	// Assertions
	if err != nil {
		t.Fatalf("PostSite failed: %v", err)
	}
	if site.ID != "site123" {
		t.Errorf("Expected site ID 'site123', got '%s'", site.ID)
	}
	if site.DisplayName != "My Test Site" {
		t.Errorf("Expected displayName 'My Test Site', got '%s'", site.DisplayName)
	}
	if site.WorkspaceID != "workspace456" {
		t.Errorf("Expected workspaceId 'workspace456', got '%s'", site.WorkspaceID)
	}
}

// TestPostSite_MinimalFields tests successful site creation with minimal required fields
func TestPostSite_MinimalFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody SiteCreateRequest
		_ = json.Unmarshal(body, &reqBody)

		// Verify minimal fields
		if reqBody.Name != "Minimal Site" {
			t.Errorf("Expected name 'Minimal Site', got '%s'", reqBody.Name)
		}
		if reqBody.ParentFolderID != "" {
			t.Errorf("Expected empty parentFolderId, got '%s'", reqBody.ParentFolderID)
		}

		response := Site{
			ID:          "site-minimal",
			WorkspaceID: "workspace123",
			DisplayName: reqBody.Name,
			ShortName:   "minimal-site",
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PostSite(context.Background(), client, "workspace123", "Minimal Site", "", "")
	if err != nil {
		t.Fatalf("PostSite failed: %v", err)
	}
	if site.ID != "site-minimal" {
		t.Errorf("Expected site ID 'site-minimal', got '%s'", site.ID)
	}
}

// TestPostSite_RateLimiting tests 429 handling with retry logic
func TestPostSite_RateLimiting(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			// First request: rate limit
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message": "Rate limit exceeded"}`))
		} else {
			// Second request: success
			response := Site{
				ID:          "site123",
				DisplayName: "My Test Site",
				WorkspaceID: "workspace456",
			}
			w.WriteHeader(201)
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PostSite(context.Background(), client, "workspace456", "My Test Site", "", "")
	if err != nil {
		t.Fatalf("PostSite should succeed after retry, got error: %v", err)
	}
	if attemptCount < 2 {
		t.Errorf("Expected at least 2 attempts due to rate limiting, got %d", attemptCount)
	}
	if site.ID != "site123" {
		t.Errorf("Expected site ID 'site123', got '%s'", site.ID)
	}
}

// TestPostSite_InvalidWorkspace tests 403 response for non-Enterprise workspace
func TestPostSite_InvalidWorkspace(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte("Forbidden: Enterprise workspace required"))
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PostSite(context.Background(), client, "non-enterprise-workspace", "My Site", "", "")

	if err == nil {
		t.Error("Expected error for 403 response, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected 'forbidden' in error message, got: %v", err)
	}
}

// TestPostSite_ContextCancellation tests context cancellation during creation
func TestPostSite_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Delay to allow cancellation
		w.WriteHeader(201)
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PostSite(ctx, client, "workspace456", "My Site", "", "")

	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}

// TestPostSite_EmptySiteID tests defensive check for empty site ID in response
func TestPostSite_EmptySiteID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Return response with empty site ID
		response := Site{
			ID:          "", // Empty ID
			DisplayName: "My Test Site",
			WorkspaceID: "workspace456",
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PostSite(context.Background(), client, "workspace456", "My Site", "", "")
	// Should succeed - defensive check is in Create method, not PostSite
	if err != nil {
		t.Errorf("PostSite should succeed even with empty ID, got error: %v", err)
	}
	if site.ID != "" {
		t.Errorf("Expected empty site ID, got '%s'", site.ID)
	}
}

// TestPostSite_InvalidJSON tests handling of invalid JSON in response
func TestPostSite_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(201)
		_, _ = w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	oldURL := postSiteBaseURL
	postSiteBaseURL = server.URL
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PostSite(context.Background(), client, "workspace456", "My Site", "", "")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected 'failed to parse response' in error, got: %v", err)
	}
}

// TestPostSite_NetworkError tests network error handling
func TestPostSite_NetworkError(t *testing.T) {
	// Use invalid URL to trigger network error
	oldURL := postSiteBaseURL
	postSiteBaseURL = "http://localhost:1" // Port 1 should be unreachable
	defer func() { postSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 1 * time.Second}
	_, err := PostSite(context.Background(), client, "workspace456", "My Site", "", "")

	if err == nil {
		t.Error("Expected network error, got nil")
	}
	// Error should mention connection or network issue
	if !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("Expected 'max retries exceeded' in error, got: %v", err)
	}
}

// ============================================================================
// Create Method Tests
// ============================================================================

// TestSiteCreate_ValidationErrors tests validation errors caught before API call
func TestSiteCreate_ValidationErrors(t *testing.T) {
	tests := []struct {
		name      string
		args      SiteArgs
		wantErr   bool
		errSubstr string
	}{
		{
			name:      "empty workspaceId",
			args:      SiteArgs{WorkspaceID: "", DisplayName: "Site"},
			wantErr:   true,
			errSubstr: "workspaceId is required",
		},
		{
			name:      "empty displayName",
			args:      SiteArgs{WorkspaceID: "ws123", DisplayName: ""},
			wantErr:   true,
			errSubstr: "displayName is required",
		},
		// Note: shortName validation tests removed - shortName is now read-only (ignored on input)
		// Note: timezone validation test removed - timezone is now read-only (output only)
	}

	resource := &SiteResource{}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := infer.CreateRequest[SiteArgs]{
				Inputs: tt.args,
			}

			_, err := resource.Create(context.Background(), req)
			if (err != nil) != tt.wantErr {
				t.Errorf("Create() error = %v, wantErr %v", err, tt.wantErr)
			}
			if tt.wantErr && !strings.Contains(err.Error(), tt.errSubstr) {
				t.Errorf("Expected error containing '%s', got: %v", tt.errSubstr, err)
			}
		})
	}
}

// TestSiteCreate_DryRun tests DryRun mode returns preview without API call
func TestSiteCreate_DryRun(t *testing.T) {
	resource := &SiteResource{}
	args := SiteArgs{
		WorkspaceID: "workspace456",
		DisplayName: "Preview Site",
		ShortName:   "preview-site",
	}

	req := infer.CreateRequest[SiteArgs]{
		Inputs: args,
		DryRun: true,
	}

	resp, err := resource.Create(context.Background(), req)
	if err != nil {
		t.Fatalf("Create DryRun failed: %v", err)
	}

	// Verify preview ID format (now just preview-timestamp)
	if !strings.HasPrefix(resp.ID, "preview-") {
		t.Errorf("Expected preview ID to start with 'preview-', got '%s'", resp.ID)
	}

	// Verify state contains input values
	if resp.Output.DisplayName != "Preview Site" {
		t.Errorf("Expected displayName 'Preview Site', got '%s'", resp.Output.DisplayName)
	}
}

// TestSiteCreate_ShortNameInputIgnored verifies that when a user provides a shortName,
// it is ignored and the auto-generated value from Webflow is used instead.
// The Webflow API does not support setting shortName on create or update.
func TestSiteCreate_ShortNameInputIgnored(t *testing.T) {
	postServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var reqBody SiteCreateRequest
		_ = json.Unmarshal(body, &reqBody)
		if reqBody.Name != "My Custom Site" {
			t.Errorf("Expected name 'My Custom Site', got '%s'", reqBody.Name)
		}
		// Webflow auto-generates shortName from displayName
		response := Site{
			ID:          "site-abc",
			WorkspaceID: "ws-123",
			DisplayName: reqBody.Name,
			ShortName:   "my-custom-site", // auto-generated, NOT user's "custom-slug"
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer postServer.Close()

	oldPostURL := postSiteBaseURL
	postSiteBaseURL = postServer.URL
	defer func() { postSiteBaseURL = oldPostURL }()

	t.Setenv("WEBFLOW_API_TOKEN", "test-token-abc123def456")

	resource := &SiteResource{}
	resp, err := resource.Create(context.Background(), infer.CreateRequest[SiteArgs]{
		Inputs: SiteArgs{
			WorkspaceID: "ws-123",
			DisplayName: "My Custom Site",
			ShortName:   "custom-slug", // This should be ignored
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	// shortName should come from API response, not user input
	if resp.Output.ShortName != "my-custom-site" {
		t.Errorf("Expected auto-generated shortName 'my-custom-site', got '%s'", resp.Output.ShortName)
	}
	if resp.ID != "site-abc" {
		t.Errorf("Expected site ID 'site-abc', got '%s'", resp.ID)
	}
}

// TestSiteCreate_ShortNameFromAPIResponse verifies that shortName is always
// populated from the API response (auto-generated by Webflow).
func TestSiteCreate_ShortNameFromAPIResponse(t *testing.T) {
	postServer := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Site{
			ID:          "site-xyz",
			WorkspaceID: "ws-123",
			DisplayName: "Auto Name Site",
			ShortName:   "auto-name-site", // auto-generated by Webflow
		}
		w.WriteHeader(201)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer postServer.Close()

	oldPostURL := postSiteBaseURL
	postSiteBaseURL = postServer.URL
	defer func() { postSiteBaseURL = oldPostURL }()

	t.Setenv("WEBFLOW_API_TOKEN", "test-token-abc123def456")

	resource := &SiteResource{}
	resp, err := resource.Create(context.Background(), infer.CreateRequest[SiteArgs]{
		Inputs: SiteArgs{
			WorkspaceID: "ws-123",
			DisplayName: "Auto Name Site",
		},
	})
	if err != nil {
		t.Fatalf("Create failed: %v", err)
	}
	if resp.Output.ShortName != "auto-name-site" {
		t.Errorf("Expected auto-generated shortName 'auto-name-site', got '%s'", resp.Output.ShortName)
	}
}

// ============================================================================
// PatchSite API Function Tests
// ============================================================================

// TestPatchSite_Success tests successful site update with all fields
func TestPatchSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH request, got %s", r.Method)
		}

		response := Site{
			ID:          "site123",
			WorkspaceID: "workspace456",
			DisplayName: "Updated Site Name",
			ShortName:   "updated-slug",
			TimeZone:    "America/New_York",
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	// Note: shortName and timeZone removed - both are read-only and cannot be set via API
	site, err := PatchSite(
		context.Background(), client, "site123",
		"Updated Site Name", "",
	)
	if err != nil {
		t.Fatalf("PatchSite failed: %v", err)
	}
	if site.DisplayName != "Updated Site Name" {
		t.Errorf("Expected displayName 'Updated Site Name', got '%s'", site.DisplayName)
	}
}

// TestPatchSite_SingleFieldChange tests single field update
func TestPatchSite_SingleFieldChange(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Site{
			ID:          "site123",
			DisplayName: "New Name",
			ShortName:   "my-site",
			TimeZone:    "UTC",
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PatchSite(context.Background(), client, "site123", "New Name", "")
	if err != nil {
		t.Fatalf("PatchSite failed: %v", err)
	}
	if site.DisplayName != "New Name" {
		t.Errorf("Expected displayName 'New Name', got '%s'", site.DisplayName)
	}
}

// TestPatchSite_NoChanges tests idempotent update
func TestPatchSite_NoChanges(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := Site{
			ID:          "site123",
			DisplayName: "My Site",
			ShortName:   "my-site",
			TimeZone:    "America/New_York",
		}
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	site, err := PatchSite(context.Background(), client, "site123", "My Site", "")
	if err != nil {
		t.Fatalf("PatchSite failed: %v", err)
	}
	if site.DisplayName != "My Site" {
		t.Errorf("Expected displayName 'My Site', got '%s'", site.DisplayName)
	}
}

// TestPatchSite_RateLimiting tests 429 handling
func TestPatchSite_RateLimiting(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(429)
		} else {
			response := Site{ID: "site123", DisplayName: "Updated Name"}
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PatchSite(context.Background(), client, "site123", "Updated Name", "")
	if err != nil {
		t.Fatalf("PatchSite should succeed after retry: %v", err)
	}
	if attemptCount < 2 {
		t.Errorf("Expected at least 2 attempts, got %d", attemptCount)
	}
}

// TestPatchSite_InvalidSiteID tests 404 response
func TestPatchSite_InvalidSiteID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("Not found"))
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PatchSite(context.Background(), client, "nonexistent", "Name", "")

	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
}

// TestPatchSite_ContextCancellation tests context cancellation
func TestPatchSite_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PatchSite(ctx, client, "site123", "Name", "")

	if err == nil {
		t.Error("Expected error for cancelled context")
	}
}

// ============================================================================
// Update Method Tests
// ============================================================================

// TestSiteUpdate_ValidationError tests validation before API call
func TestSiteUpdate_ValidationError(t *testing.T) {
	resource := &SiteResource{}
	req := infer.UpdateRequest[SiteArgs, SiteState]{
		ID: "workspace456/sites/site123",
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "", // Invalid
		},
	}

	_, err := resource.Update(context.Background(), req)
	if err == nil {
		t.Error("Expected validation error for empty displayName")
	}
}

// TestSiteUpdate_DryRun tests DryRun mode
func TestSiteUpdate_DryRun(t *testing.T) {
	resource := &SiteResource{}
	req := infer.UpdateRequest[SiteArgs, SiteState]{
		ID: "workspace456/sites/site123",
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "Updated Site",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				DisplayName: "Old Site",
			},
		},
		DryRun: true,
	}

	resp, err := resource.Update(context.Background(), req)
	if err != nil {
		t.Fatalf("Update DryRun failed: %v", err)
	}

	if resp.Output.DisplayName != "Updated Site" {
		t.Errorf("Expected displayName 'Updated Site', got '%s'", resp.Output.DisplayName)
	}
}

// ============================================================================
// Diff Method Tests
// ============================================================================

// TestSiteDiff_NoChanges tests when no fields changed
func TestSiteDiff_NoChanges(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Site",
			ShortName:   "my-site",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "My Site",
				ShortName:   "my-site",
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if diff.HasChanges {
		t.Error("Expected no changes")
	}
}

// TestSiteDiff_DisplayNameChanged tests displayName change
func TestSiteDiff_DisplayNameChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "New Name",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "Old Name",
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !diff.HasChanges {
		t.Error("Expected HasChanges=true")
	}
	if _, ok := diff.DetailedDiff["displayName"]; !ok {
		t.Error("Expected 'displayName' in DetailedDiff")
	}
}

// TestSiteDiff_MultipleFieldsChanged tests CRITICAL: all changes accumulate
func TestSiteDiff_MultipleFieldsChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID:    "workspace456",
			DisplayName:    "New Name",
			ParentFolderID: "folder-new",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID:    "workspace456",
				DisplayName:    "Old Name",
				ParentFolderID: "folder-old",
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// CRITICAL: All changes should be accumulated, not overwritten
	// Note: shortName and timeZone are read-only (output only) and not diffed
	expectedChanges := []string{"displayName", "parentFolderId"}
	for _, field := range expectedChanges {
		if _, ok := diff.DetailedDiff[field]; !ok {
			t.Errorf("Expected '%s' in DetailedDiff", field)
		}
	}
	if len(diff.DetailedDiff) != len(expectedChanges) {
		t.Errorf("Expected %d changes, got %d", len(expectedChanges), len(diff.DetailedDiff))
	}
}

// TestSiteDiff_ImmutableFieldChanged tests workspaceId requires replace
func TestSiteDiff_ImmutableFieldChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "new-workspace",
			DisplayName: "Site",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "old-workspace",
				DisplayName: "Site",
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if _, ok := diff.DetailedDiff["workspaceId"]; !ok {
		t.Error("Expected 'workspaceId' in DetailedDiff")
	}

	propertyDiff := diff.DetailedDiff["workspaceId"]
	if propertyDiff.Kind != p.UpdateReplace {
		t.Errorf("Expected UpdateReplace for workspaceId, got %v", propertyDiff.Kind)
	}
}

// TestSiteDiff_ShortNameReadOnly tests that shortName changes are NOT detected in Diff
// because shortName is read-only (auto-generated by Webflow) and cannot be set via API.
func TestSiteDiff_ShortNameReadOnly(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Site",
			ShortName:   "new-slug", // Even if user provides different shortName...
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "My Site",
				ShortName:   "old-slug", // ...it should NOT trigger a diff
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// shortName is read-only — changes should NOT be detected
	if diff.HasChanges {
		t.Error("Expected HasChanges=false — shortName is read-only and should not trigger changes")
	}
	if _, ok := diff.DetailedDiff["shortName"]; ok {
		t.Error("shortName should NOT be in DetailedDiff — it's read-only")
	}
}

// TestSiteDiff_TimeZoneReadOnly tests that timezone differences in state don't trigger changes
// because timezone is read-only (output only) and cannot be set by users.
func TestSiteDiff_TimeZoneReadOnly(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Site",
			// Note: TimeZone is not in SiteArgs - it's read-only output
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "My Site",
			},
			TimeZone: "UTC", // This is a read-only output field
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// No changes should be detected - timezone is read-only
	if diff.HasChanges {
		t.Error("Expected HasChanges=false - timezone is read-only and should not trigger changes")
	}
	if _, ok := diff.DetailedDiff["timeZone"]; ok {
		t.Error("timeZone should NOT be in DetailedDiff - it's read-only")
	}
}

// TestSiteDiff_ParentFolderIDChanged tests parentFolderId change alone
func TestSiteDiff_ParentFolderIDChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID:    "workspace456",
			DisplayName:    "My Site",
			ParentFolderID: "folder-new",
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID:    "workspace456",
				DisplayName:    "My Site",
				ParentFolderID: "folder-old",
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !diff.HasChanges {
		t.Error("Expected HasChanges=true")
	}
	if _, ok := diff.DetailedDiff["parentFolderId"]; !ok {
		t.Error("Expected 'parentFolderId' in DetailedDiff")
	}
	if len(diff.DetailedDiff) != 1 {
		t.Errorf("Expected only 1 change (parentFolderId), got %d", len(diff.DetailedDiff))
	}
}

// TestPatchSite_NetworkError tests network error handling with retry
func TestPatchSite_NetworkError(t *testing.T) {
	// Use invalid URL to trigger network error
	oldURL := patchSiteBaseURL
	patchSiteBaseURL = "http://localhost:1" // Port 1 should be unreachable
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 1 * time.Second}
	_, err := PatchSite(context.Background(), client, "site123", "Name", "")

	if err == nil {
		t.Error("Expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("Expected 'max retries exceeded' in error, got: %v", err)
	}
}

// TestPatchSite_InvalidJSON tests handling of invalid JSON response
func TestPatchSite_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	oldURL := patchSiteBaseURL
	patchSiteBaseURL = server.URL
	defer func() { patchSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PatchSite(context.Background(), client, "site123", "Name", "")

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected 'failed to parse response' in error, got: %v", err)
	}
}

// Note: TestSiteUpdate_Success with real API call requires provider context infrastructure.
// The PatchSite API function is tested directly via TestPatchSite_* tests.
// The Update method's validation and DryRun logic is tested via TestSiteUpdate_ValidationError
// and TestSiteUpdate_DryRun.
// Full integration testing is done via manual testing with `pulumi up`.

// ============================================================================
// PublishSite API Function Tests
// ============================================================================

// TestPublishSite_Success tests successful site publishing with default domains
func TestPublishSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v2/sites/") || !strings.Contains(r.URL.Path, "/publish") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Parse request body
		body, _ := io.ReadAll(r.Body)
		var reqBody SitePublishRequest
		_ = json.Unmarshal(body, &reqBody)

		// Return mock publish response (202 Accepted for async operation)
		response := SitePublishResponse{
			Published: true,
			Queued:    false,
			Message:   "Site published successfully",
		}
		w.WriteHeader(202)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := PublishSite(context.Background(), client, "site123", nil)
	if err != nil {
		t.Fatalf("PublishSite failed: %v", err)
	}
	if !resp.Published {
		t.Errorf("Expected Published=true, got false")
	}
}

// TestPublishSite_SuccessWith200 tests 200 OK response (alternative to 202)
func TestPublishSite_SuccessWith200(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		response := SitePublishResponse{
			Published: true,
			Queued:    false,
		}
		w.WriteHeader(200) // 200 OK instead of 202
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := PublishSite(context.Background(), client, "site123", nil)
	if err != nil {
		t.Fatalf("PublishSite failed with 200 response: %v", err)
	}
	if !resp.Published {
		t.Errorf("Expected Published=true, got false")
	}
}

// TestPublishSite_WithSpecificDomains tests publishing to specific domains array
func TestPublishSite_WithSpecificDomains(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Parse request body to verify domains
		body, _ := io.ReadAll(r.Body)
		var reqBody SitePublishRequest
		_ = json.Unmarshal(body, &reqBody)

		if len(reqBody.Domains) != 2 {
			t.Errorf("Expected 2 domains, got %d", len(reqBody.Domains))
		}
		if len(reqBody.Domains) > 0 && reqBody.Domains[0] != "example.com" {
			t.Errorf("Expected first domain 'example.com', got '%s'", reqBody.Domains[0])
		}

		response := SitePublishResponse{
			Published: true,
			Queued:    false,
		}
		w.WriteHeader(202)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := PublishSite(context.Background(), client, "site123", []string{"example.com", "www.example.com"})
	if err != nil {
		t.Fatalf("PublishSite failed: %v", err)
	}
	if !resp.Published {
		t.Errorf("Expected Published=true, got false")
	}
}

// TestPublishSite_RateLimiting tests 429 handling with retry logic
func TestPublishSite_RateLimiting(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount == 1 {
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message": "Rate limit exceeded"}`))
		} else {
			response := SitePublishResponse{
				Published: true,
				Queued:    false,
			}
			w.WriteHeader(202)
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	resp, err := PublishSite(context.Background(), client, "site123", nil)
	if err != nil {
		t.Fatalf("PublishSite should succeed after retry: %v", err)
	}
	if attemptCount < 2 {
		t.Errorf("Expected at least 2 attempts due to rate limiting, got %d", attemptCount)
	}
	if !resp.Published {
		t.Errorf("Expected Published=true after successful retry, got false")
	}
}

// TestPublishSite_NetworkError tests network error handling and retry
func TestPublishSite_NetworkError(t *testing.T) {
	oldURL := publishSiteBaseURL
	publishSiteBaseURL = "http://localhost:1" // Unreachable port
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 1 * time.Second}
	_, err := PublishSite(context.Background(), client, "site123", nil)

	if err == nil {
		t.Error("Expected network error, got nil")
	}
	if !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("Expected 'max retries exceeded' in error, got: %v", err)
	}
}

// TestPublishSite_InvalidSiteID tests 404 response for non-existent site
func TestPublishSite_InvalidSiteID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(404)
		_, _ = w.Write([]byte("Site not found"))
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PublishSite(context.Background(), client, "nonexistent-site", nil)

	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error message, got: %v", err)
	}
}

// TestPublishSite_SiteNotReady tests 400 error when site not ready for publishing
func TestPublishSite_SiteNotReady(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(400)
		_, _ = w.Write([]byte("Bad request: site has no published version"))
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PublishSite(context.Background(), client, "site123", nil)

	if err == nil {
		t.Error("Expected error for unpublishable site, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error message, got: %v", err)
	}
}

// TestPublishSite_ContextCancellation tests context cancellation during publish
func TestPublishSite_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		time.Sleep(100 * time.Millisecond) // Delay to allow cancellation
		w.WriteHeader(202)
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PublishSite(ctx, client, "site123", nil)

	if err == nil {
		t.Error("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") && !strings.Contains(err.Error(), "canceled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}

// TestPublishSite_InvalidJSON tests handling of invalid JSON in response
func TestPublishSite_InvalidJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(202)
		_, _ = w.Write([]byte("invalid json {{{"))
	}))
	defer server.Close()

	oldURL := publishSiteBaseURL
	publishSiteBaseURL = server.URL
	defer func() { publishSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	_, err := PublishSite(context.Background(), client, "site123", nil)

	if err == nil {
		t.Error("Expected error for invalid JSON, got nil")
	}
	if !strings.Contains(err.Error(), "failed to parse response") {
		t.Errorf("Expected 'failed to parse response' in error, got: %v", err)
	}
}

// ============================================================================
// Diff Method Tests - Publish Property
// ============================================================================

// TestSiteDiff_PublishChanged tests publish property change detection
func TestSiteDiff_PublishChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Site",
			Publish:     true, // Changed to true
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "My Site",
				Publish:     false, // Was false
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if !diff.HasChanges {
		t.Error("Expected HasChanges=true")
	}
	if _, ok := diff.DetailedDiff["publish"]; !ok {
		t.Error("Expected 'publish' in DetailedDiff")
	}
}

// ============================================================================
// Create/Update Integration Tests - Publish Property
// ============================================================================
//
// NOTE: Full integration tests for Create/Update with actual API calls require
// the Pulumi provider context (GetHTTPClient needs infer.GetConfig). These tests
// verify Diff behavior and DryRun mode. Full integration testing with publish
// is done via manual `pulumi up` with the provider binary.
//
// The API functions (PostSite, PatchSite, PublishSite) are tested directly above,
// verifying the underlying HTTP calls work correctly.

// TestSiteCreate_DryRunWithPublish tests DryRun mode with publish=true
// Issue #5 fix: DryRun behavior test
func TestSiteCreate_DryRunWithPublish(t *testing.T) {
	apiCalled := false

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		apiCalled = true // Should NOT be called in DryRun
		w.WriteHeader(500)
	}))
	defer server.Close()

	oldPostURL := postSiteBaseURL
	oldPublishURL := publishSiteBaseURL
	postSiteBaseURL = server.URL
	publishSiteBaseURL = server.URL
	defer func() {
		postSiteBaseURL = oldPostURL
		publishSiteBaseURL = oldPublishURL
	}()

	resource := &SiteResource{}
	req := infer.CreateRequest[SiteArgs]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "Test Site",
			Publish:     true,
		},
		DryRun: true, // DryRun should skip ALL API calls
	}

	resp, err := resource.Create(context.Background(), req)

	if apiCalled {
		t.Error("API should NOT be called in DryRun mode")
	}
	if err != nil {
		t.Errorf("DryRun should succeed without error, got: %v", err)
	}
	if !strings.HasPrefix(resp.ID, "preview-") {
		t.Errorf("Expected preview ID format (preview-*), got: %s", resp.ID)
	}
	// Verify publish property is preserved in output
	if resp.Output.Publish != true {
		t.Error("Expected Publish=true in DryRun output")
	}
}

// TestSiteDiff_PublishIdempotency tests that same publish value shows no changes
// Issue #6 fix: Idempotency test
func TestSiteDiff_PublishIdempotency(t *testing.T) {
	resource := &SiteResource{}

	// Test publish=true to publish=true (no change)
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Site",
			Publish:     true,
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "My Site",
				Publish:     true, // Same value
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	if diff.HasChanges {
		t.Error("Expected no changes when publish value is the same")
	}
	if _, ok := diff.DetailedDiff["publish"]; ok {
		t.Error("publish should NOT be in DetailedDiff when unchanged")
	}
}

// TestSiteDiff_PublishAndOtherFieldsChanged tests publish change combined with other changes
func TestSiteDiff_PublishAndOtherFieldsChanged(t *testing.T) {
	resource := &SiteResource{}
	req := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "New Name",
			Publish:     true,
		},
		State: SiteState{
			SiteArgs: SiteArgs{
				WorkspaceID: "workspace456",
				DisplayName: "Old Name",
				Publish:     false,
			},
		},
	}

	diff, err := resource.Diff(context.Background(), req)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Should accumulate all changes
	// Note: timeZone removed - it's now read-only (output only)
	expectedChanges := []string{"displayName", "publish"}
	for _, field := range expectedChanges {
		if _, ok := diff.DetailedDiff[field]; !ok {
			t.Errorf("Expected '%s' in DetailedDiff", field)
		}
	}
	if len(diff.DetailedDiff) != len(expectedChanges) {
		t.Errorf("Expected %d changes, got %d", len(expectedChanges), len(diff.DetailedDiff))
	}
}

// TestDeleteSite_Success tests successful site deletion (204 No Content)
func TestDeleteSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method and path
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v2/sites/") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Return 204 No Content (success)
		w.WriteHeader(204)
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(context.Background(), client, "site123")
	if err != nil {
		t.Fatalf("DeleteSite failed: %v", err)
	}
}

// TestDeleteSite_AlreadyDeleted404 tests idempotent deletion (404 = success)
func TestDeleteSite_AlreadyDeleted404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify DELETE request
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE request, got %s", r.Method)
		}

		// Return 404 Not Found (site already deleted - treat as success)
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "Site not found"}`))
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(context.Background(), client, "site456")
	if err != nil {
		t.Fatalf("DeleteSite should treat 404 as success (idempotent), got error: %v", err)
	}
}

// TestDeleteSite_RateLimiting tests 429 rate limit with retry
func TestDeleteSite_RateLimiting(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			// First attempt: return 429 with Retry-After header
			// Note: Headers must be set BEFORE WriteHeader() in Go
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"error": "rate limited"}`))
		} else {
			// Second attempt: success
			w.WriteHeader(204)
		}
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(context.Background(), client, "site789")
	if err != nil {
		t.Fatalf("DeleteSite should retry on 429, got error: %v", err)
	}
	if attempt != 2 {
		t.Errorf("Expected 2 attempts (429 then success), got %d", attempt)
	}
}

// TestDeleteSite_PermissionError tests 403 Forbidden
func TestDeleteSite_PermissionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error": "Insufficient permissions to delete site"}`))
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(context.Background(), client, "site999")

	if err == nil {
		t.Fatal("Expected error for 403 Forbidden, got nil")
	}
	if !strings.Contains(err.Error(), "access denied") && !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected error to mention permission issue, got: %v", err)
	}
}

// TestDeleteSite_NetworkError tests network failure handling
func TestDeleteSite_NetworkError(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}

	// Use unreachable address to trigger network error
	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = "http://127.0.0.1:1/invalid"
	defer func() { deleteSiteBaseURL = oldURL }()

	err := DeleteSite(context.Background(), client, "site000")

	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}
	if !strings.Contains(err.Error(), "max retries exceeded") {
		t.Errorf("Expected 'max retries exceeded' error, got: %v", err)
	}
}

// TestDeleteSite_ContextCancellation tests context cancellation during request
func TestDeleteSite_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(204)
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	// Create context and cancel immediately
	ctx, cancel := context.WithCancel(context.Background())
	cancel()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(ctx, client, "site111")

	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("Expected error mentioning context cancellation, got: %v", err)
	}
}

// TestDeleteSite_ServerError tests 500 server error handling
func TestDeleteSite_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	oldURL := deleteSiteBaseURL
	deleteSiteBaseURL = server.URL
	defer func() { deleteSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	err := DeleteSite(context.Background(), client, "site222")

	if err == nil {
		t.Fatal("Expected error for 500 server error, got nil")
	}
	if !strings.Contains(err.Error(), "server error") && !strings.Contains(err.Error(), "internal") {
		t.Errorf("Expected error to mention server error, got: %v", err)
	}
}

// Note: Site Delete method integration tests require full provider context setup.
// The DeleteSite API function is thoroughly tested above with all scenarios.
// The Delete method is a thin wrapper that:
// 1. Uses resource ID directly as siteId (no parsing needed)
// 2. Gets HTTP client (provider framework responsibility)
// 3. Calls DeleteSite (tested with TestDeleteSite_* tests)
//
// Testing the full integration would require mocking the provider's Config,
// which is tested in end-to-end Pulumi deployment scenarios.
//
// The implementation is straightforward and mirrors the proven pattern from
// Redirect.Delete and follows the exact same structure and error handling.

// TestGetSite_Success tests successful site retrieval (200 OK)
func TestGetSite_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify GET request
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/v2/sites/") {
			t.Errorf("Unexpected URL path: %s", r.URL.Path)
		}

		// Return 200 OK with site data
		response := Site{
			ID:          "site123",
			WorkspaceID: "workspace456",
			DisplayName: "Test Site",
			ShortName:   "test-site",
			TimeZone:    "America/New_York",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(context.Background(), client, "site123")
	if err != nil {
		t.Fatalf("GetSite failed: %v", err)
	}
	if siteData == nil {
		t.Fatal("Expected site data, got nil")
	}
	if siteData.DisplayName != "Test Site" {
		t.Errorf("Expected DisplayName 'Test Site', got '%s'", siteData.DisplayName)
	}
	if siteData.ShortName != "test-site" {
		t.Errorf("Expected ShortName 'test-site', got '%s'", siteData.ShortName)
	}
}

// TestGetSite_NotFound404 tests site not found (404 - site deleted externally)
func TestGetSite_NotFound404(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}
		// Return 404 Not Found (site doesn't exist - treat as nil, nil signal)
		w.WriteHeader(404)
		_, _ = w.Write([]byte(`{"error": "Site not found"}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(context.Background(), client, "nonexistent")
	if err != nil {
		t.Fatalf("GetSite should return nil, nil for 404, got error: %v", err)
	}
	if siteData != nil {
		t.Fatalf("Expected nil site data for 404, got: %v", siteData)
	}
}

// TestGetSite_RateLimiting tests 429 rate limit with retry
func TestGetSite_RateLimiting(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt == 1 {
			// First attempt: return 429 with Retry-After
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"error": "Rate limited"}`))
			return
		}
		// Second attempt: return 200 OK
		response := Site{
			ID:          "site123",
			WorkspaceID: "workspace456",
			DisplayName: "Test Site",
			ShortName:   "test-site",
			TimeZone:    "America/New_York",
		}
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(context.Background(), client, "site123")
	if err != nil {
		t.Fatalf("GetSite should retry on 429 and succeed, got error: %v", err)
	}
	if siteData == nil {
		t.Fatal("Expected site data after retry, got nil")
	}
	if siteData.DisplayName != "Test Site" {
		t.Errorf("Expected DisplayName 'Test Site', got '%s'", siteData.DisplayName)
	}
}

// TestGetSite_MalformedJSON tests malformed JSON response
func TestGetSite_MalformedJSON(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{invalid json`)) // Malformed JSON
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(context.Background(), client, "site123")

	if err == nil {
		t.Fatal("Expected error for malformed JSON, got nil")
	}
	if siteData != nil {
		t.Errorf("Expected nil site data on error, got: %v", siteData)
	}
	if !strings.Contains(err.Error(), "parse") {
		t.Errorf("Expected 'parse' in error message, got: %v", err)
	}
}

// TestGetSite_PermissionError tests 403 Forbidden
func TestGetSite_PermissionError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(403)
		_, _ = w.Write([]byte(`{"error": "Forbidden"}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(context.Background(), client, "site123")

	if err == nil {
		t.Fatal("Expected error for 403 Forbidden, got nil")
	}
	if siteData != nil {
		t.Errorf("Expected nil site data on error, got: %v", siteData)
	}
}

// TestGetSite_NetworkError tests network failure handling
func TestGetSite_NetworkError(t *testing.T) {
	client := &http.Client{Timeout: 30 * time.Second}

	oldURL := getSiteBaseURL
	getSiteBaseURL = "http://localhost:99999" // Non-existent server
	defer func() { getSiteBaseURL = oldURL }()

	siteData, err := GetSite(context.Background(), client, "site123")

	if err == nil {
		t.Fatal("Expected error for network failure, got nil")
	}
	if siteData != nil {
		t.Errorf("Expected nil site data on error, got: %v", siteData)
	}
}

// TestGetSite_ContextCancellation tests context cancellation
func TestGetSite_ContextCancellation(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Simulate slow response
		time.Sleep(100 * time.Millisecond)
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	ctx, cancel := context.WithCancel(context.Background())
	cancel() // Cancel immediately

	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(ctx, client, "site123")

	if err == nil {
		t.Fatal("Expected error for cancelled context, got nil")
	}
	if siteData != nil {
		t.Errorf("Expected nil site data on error, got: %v", siteData)
	}
	if !strings.Contains(err.Error(), "context cancelled") {
		t.Errorf("Expected 'context cancelled' in error, got: %v", err)
	}
}

// TestGetSite_ServerError tests handling of 500 Internal Server Error
func TestGetSite_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(500)
		_, _ = w.Write([]byte(`{"error": "Internal server error"}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(ctx, client, "site123")

	if err == nil {
		t.Fatal("Expected error for server error, got nil")
	}
	if siteData != nil {
		t.Errorf("Expected nil site data on error, got: %v", siteData)
	}
	// The error message contains context about the server error
	if !strings.Contains(err.Error(), "server error") && !strings.Contains(err.Error(), "internal") {
		t.Errorf("Expected server/internal error message, got: %v", err)
	}
}

// =============================================================================
// GetSite API Function Tests - Additional Coverage
// =============================================================================

// TestGetSite_WithParentFolderID tests that parentFolderId is correctly parsed from API response
func TestGetSite_WithParentFolderID(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{
			"id": "site123",
			"workspaceId": "workspace456",
			"displayName": "Test Site",
			"shortName": "test-site",
			"parentFolderId": "folder789"
		}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(ctx, client, "site123")
	if err != nil {
		t.Fatalf("GetSite() returned unexpected error: %v", err)
	}
	if siteData == nil {
		t.Fatal("Expected site data, got nil")
	}
	if siteData.ParentFolderID != "folder789" {
		t.Errorf("Expected ParentFolderID 'folder789', got '%s'", siteData.ParentFolderID)
	}
}

// TestGetSite_AllFields tests that all fields are correctly parsed from API response
func TestGetSite_AllFields(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(200)
		_, _ = w.Write([]byte(`{
			"id": "site123",
			"workspaceId": "workspace456",
			"displayName": "My Test Site",
			"shortName": "my-test-site",
			"timeZone": "America/New_York",
			"parentFolderId": "folder789",
			"lastPublished": "2024-01-15T10:30:00Z",
			"lastUpdated": "2024-01-15T12:00:00Z",
			"previewUrl": "https://preview.webflow.com/site123",
			"customDomains": ["example.com", "www.example.com"],
			"dataCollectionEnabled": true,
			"dataCollectionType": "optOut"
		}`))
	}))
	defer server.Close()

	oldURL := getSiteBaseURL
	getSiteBaseURL = server.URL
	defer func() { getSiteBaseURL = oldURL }()

	ctx := context.Background()
	client := &http.Client{Timeout: 30 * time.Second}
	siteData, err := GetSite(ctx, client, "site123")
	if err != nil {
		t.Fatalf("GetSite() returned unexpected error: %v", err)
	}
	if siteData == nil {
		t.Fatal("Expected site data, got nil")
	}

	// Verify all fields are parsed correctly
	if siteData.ID != "site123" {
		t.Errorf("Expected ID 'site123', got '%s'", siteData.ID)
	}
	if siteData.WorkspaceID != "workspace456" {
		t.Errorf("Expected WorkspaceID 'workspace456', got '%s'", siteData.WorkspaceID)
	}
	if siteData.DisplayName != "My Test Site" {
		t.Errorf("Expected DisplayName 'My Test Site', got '%s'", siteData.DisplayName)
	}
	if siteData.ShortName != "my-test-site" {
		t.Errorf("Expected ShortName 'my-test-site', got '%s'", siteData.ShortName)
	}
	if siteData.TimeZone != "America/New_York" {
		t.Errorf("Expected TimeZone 'America/New_York', got '%s'", siteData.TimeZone)
	}
	if siteData.ParentFolderID != "folder789" {
		t.Errorf("Expected ParentFolderID 'folder789', got '%s'", siteData.ParentFolderID)
	}
	if siteData.LastPublished != "2024-01-15T10:30:00Z" {
		t.Errorf("Expected LastPublished '2024-01-15T10:30:00Z', got '%s'", siteData.LastPublished)
	}
	if siteData.LastUpdated != "2024-01-15T12:00:00Z" {
		t.Errorf("Expected LastUpdated '2024-01-15T12:00:00Z', got '%s'", siteData.LastUpdated)
	}
	if siteData.PreviewURL != "https://preview.webflow.com/site123" {
		t.Errorf("Expected PreviewURL 'https://preview.webflow.com/site123', got '%s'", siteData.PreviewURL)
	}
	if len(siteData.CustomDomains) != 2 {
		t.Errorf("Expected 2 custom domains, got %d", len(siteData.CustomDomains))
	}
	if siteData.DataCollectionEnabled != true {
		t.Error("Expected DataCollectionEnabled to be true")
	}
	if siteData.DataCollectionType != "optOut" {
		t.Errorf("Expected DataCollectionType 'optOut', got '%s'", siteData.DataCollectionType)
	}
}

// =============================================================================
// SiteResource.Read() Tests
// =============================================================================

// Note: Complex ID parsing tests were removed as the Site resource now uses
// simple siteId as the resource ID instead of {workspaceId}/sites/{siteId}.
// Import tests were also removed as import now just uses the siteId directly.

// TestSiteDrift_ShortNameFromAPI_ShouldNotTriggerChange tests that when
// API returns shortName (which is always auto-generated by Webflow),
// it should NOT trigger a phantom change in Diff, since shortName is read-only.
func TestSiteDrift_ShortNameFromAPI_ShouldNotTriggerChange(t *testing.T) {
	resource := &SiteResource{}

	// User's Pulumi config - they did NOT specify shortName
	userInputs := SiteArgs{
		WorkspaceID: "workspace456",
		DisplayName: "My Test Site",
		// ShortName intentionally empty - user didn't specify it
	}

	// Simulate what Read() currently returns after fetching from API
	// API always returns shortName, so Read() populates it in State
	stateFromRead := SiteState{
		SiteArgs: SiteArgs{
			WorkspaceID: "workspace456",
			DisplayName: "My Test Site",
			ShortName:   "my-test-site", // API returned this, Read() included it
		},
	}

	// Diff compares user inputs vs state from Read
	diffReq := infer.DiffRequest[SiteArgs, SiteState]{
		Inputs: userInputs,    // What user has in Pulumi config
		State:  stateFromRead, // What Read() returned (includes API values)
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// THE KEY ASSERTION: There should be NO changes detected
	// The user didn't specify shortName, and we shouldn't force them to
	// explicitly set it to empty just because the API returned a value
	if diffResp.HasChanges {
		t.Errorf("Diff() detected phantom changes - this is the bug we're fixing")
		t.Errorf("DetailedDiff: %+v", diffResp.DetailedDiff)
	}

	// Specifically check that shortName is NOT flagged
	if diffResp.DetailedDiff != nil {
		if _, hasShortName := diffResp.DetailedDiff["shortName"]; hasShortName {
			t.Errorf("Diff() incorrectly flagged shortName - user didn't specify it, shouldn't be a change")
		}
	}
}
