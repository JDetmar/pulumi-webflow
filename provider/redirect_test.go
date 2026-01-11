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

	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestValidateSourcePath_Valid tests valid source paths
func TestValidateSourcePath_Valid(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/old-page"},
		{"nested path", "/blog/2023"},
		{"path with hyphen", "/old-page-name"},
		{"path with underscore", "/old_page"},
		{"root path", "/"},
		{"deep nested path", "/products/category/item-1"},
		{"path with dot", "/files/document.pdf"},
		{"complex path", "/blog/2023/my-post_v1"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourcePath(tt.path)
			if err != nil {
				t.Errorf("ValidateSourcePath(%q) = %v, want nil", tt.path, err)
			}
		})
	}
}

// TestValidateSourcePath_Empty tests empty source path
func TestValidateSourcePath_Empty(t *testing.T) {
	err := ValidateSourcePath("")
	if err == nil {
		t.Error("ValidateSourcePath(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateSourcePath_MissingSlash tests path missing leading slash
func TestValidateSourcePath_MissingSlash(t *testing.T) {
	err := ValidateSourcePath("old-page")
	if err == nil {
		t.Error("ValidateSourcePath(\"old-page\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "must start with '/'") {
		t.Errorf("Expected error to mention 'must start with', got: %v", err)
	}
}

// TestValidateSourcePath_InvalidCharacters tests path with invalid characters
func TestValidateSourcePath_InvalidCharacters(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"path with space", "/old page"},
		{"path with question mark", "/page?query"},
		{"path with hash", "/page#anchor"},
		{"path with special char", "/page@name"},
		{"path with backslash", "/path\\file"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourcePath(tt.path)
			if err == nil {
				t.Errorf("ValidateSourcePath(%q) = nil, want error", tt.path)
			}
			if !strings.Contains(err.Error(), "invalid characters") {
				t.Errorf("Expected error to mention 'invalid characters', got: %v", err)
			}
		})
	}
}

// TestValidateSourcePath_EdgeCases tests edge case paths
func TestValidateSourcePath_EdgeCases(t *testing.T) {
	tests := []struct {
		name    string
		path    string
		wantErr bool
	}{
		{"double slash", "//old-page", false},           // Allowed by regex (slashes are valid)
		{"double slash in middle", "/old//page", false}, // Allowed by regex
		{"trailing slash", "/old-page/", false},
		{"path with trailing dot", "/old-page.", false},
		{"path with double dots", "/old..page", false}, // Allowed by regex (dots are valid)
		{"hidden path with dot", "/.hidden", false},
		{"multiple segments", "/path/to/resource", false},
		{"path with numbers", "/page123", false},
		{"complex valid path", "/blog/2024/my-post_v2.html", false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourcePath(tt.path)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourcePath(%q) error = %v, wantErr %v", tt.path, err, tt.wantErr)
			}
		})
	}
}

// TestValidateDestinationPath_Valid tests valid destination paths
func TestValidateDestinationPath_Valid(t *testing.T) {
	tests := []struct {
		name string
		path string
	}{
		{"simple path", "/new-page"},
		{"nested path", "/home"},
		{"path with hyphen", "/new-page-name"},
		{"path with underscore", "/new_page"},
		{"root path", "/"},
		{"deep nested path", "/products/category/item-1"},
		{"path with dot", "/files/document.pdf"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateDestinationPath(tt.path)
			if err != nil {
				t.Errorf("ValidateDestinationPath(%q) = %v, want nil", tt.path, err)
			}
		})
	}
}

// TestValidateDestinationPath_Empty tests empty destination path
func TestValidateDestinationPath_Empty(t *testing.T) {
	err := ValidateDestinationPath("")
	if err == nil {
		t.Error("ValidateDestinationPath(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateDestinationPath_MissingSlash tests path missing leading slash
func TestValidateDestinationPath_MissingSlash(t *testing.T) {
	err := ValidateDestinationPath("new-page")
	if err == nil {
		t.Error("ValidateDestinationPath(\"new-page\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "must start with '/'") {
		t.Errorf("Expected error to mention 'must start with', got: %v", err)
	}
}

// TestValidateStatusCode_Valid tests valid status codes
func TestValidateStatusCode_Valid(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"permanent redirect", 301},
		{"temporary redirect", 302},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusCode(tt.status)
			if err != nil {
				t.Errorf("ValidateStatusCode(%d) = %v, want nil", tt.status, err)
			}
		})
	}
}

// TestValidateStatusCode_Invalid tests invalid status codes
func TestValidateStatusCode_Invalid(t *testing.T) {
	tests := []struct {
		name   string
		status int
	}{
		{"400 bad request", 400},
		{"200 ok", 200},
		{"404 not found", 404},
		{"500 server error", 500},
		{"307 temporary redirect", 307},
		{"308 permanent redirect", 308},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateStatusCode(tt.status)
			if err == nil {
				t.Errorf("ValidateStatusCode(%d) = nil, want error", tt.status)
			}
			if !strings.Contains(err.Error(), "301 or 302") {
				t.Errorf("Expected error to mention '301 or 302', got: %v", err)
			}
		})
	}
}

// TestGenerateRedirectResourceID tests resource ID generation
func TestGenerateRedirectResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	redirectID := "redir_12345"

	resourceID := GenerateRedirectResourceID(siteID, redirectID)
	expected := "5f0c8c9e1c9d440000e8d8c3/redirects/redir_12345"

	if resourceID != expected {
		t.Errorf("GenerateRedirectResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromRedirectResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromRedirectResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/redirects/redir_12345"

	siteID, redirectID, err := ExtractIDsFromRedirectResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromRedirectResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromRedirectResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if redirectID != "redir_12345" {
		t.Errorf("ExtractIDsFromRedirectResourceID() redirectID = %q, want %q", redirectID, "redir_12345")
	}
}

// TestExtractIDsFromRedirectResourceID_Empty tests empty resource ID
func TestExtractIDsFromRedirectResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromRedirectResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromRedirectResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromRedirectResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromRedirectResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing redirects part", "5f0c8c9e1c9d440000e8d8c3/redir_12345"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/robots/redir_12345"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromRedirectResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromRedirectResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestErrorMessagesAreActionable verifies error messages contain guidance
func TestErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateSourcePath empty",
			func() error { return ValidateSourcePath("") },
			[]string{"required", "Example"},
		},
		{
			"ValidateSourcePath missing slash",
			func() error { return ValidateSourcePath("old-page") },
			[]string{"must start with '/'", "Example"},
		},
		{
			"ValidateStatusCode invalid",
			func() error { return ValidateStatusCode(200) },
			[]string{"301", "302", "permanent", "temporary"},
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

// TestGetRedirects_Valid tests retrieving redirects successfully
func TestGetRedirects_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/redirects") {
			t.Errorf("Expected /redirects in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := RedirectResponse{
			Redirects: []RedirectRule{
				{ID: "redirect1", SourcePath: "/old", DestinationPath: "/new", StatusCode: 301},
				{ID: "redirect2", SourcePath: "/about", DestinationPath: "/team", StatusCode: 302},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetRedirects(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetRedirects failed: %v", err)
	}

	if len(result.Redirects) != 2 {
		t.Errorf("Expected 2 redirects, got %d", len(result.Redirects))
	}
	if result.Redirects[0].ID != "redirect1" {
		t.Errorf("Expected redirect1, got %s", result.Redirects[0].ID)
	}
}

// TestGetRedirects_NotFound tests 404 handling
func TestGetRedirects_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetRedirects(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestPostRedirect_Valid tests creating a redirect successfully
func TestPostRedirect_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req RedirectRequest
		_ = json.Unmarshal(body, &req)

		if req.SourcePath != "/old" {
			t.Errorf("Expected sourcePath /old, got %s", req.SourcePath)
		}
		if req.DestinationPath != "/new" {
			t.Errorf("Expected destinationPath /new, got %s", req.DestinationPath)
		}
		if req.StatusCode != 301 {
			t.Errorf("Expected statusCode 301, got %d", req.StatusCode)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := RedirectRule{ID: "new-redirect-1", SourcePath: "/old", DestinationPath: "/new", StatusCode: 301}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postRedirectBaseURL
	postRedirectBaseURL = server.URL
	defer func() { postRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "/old", "/new", 301)
	if err != nil {
		t.Fatalf("PostRedirect failed: %v", err)
	}

	if result.ID != "new-redirect-1" {
		t.Errorf("Expected ID new-redirect-1, got %s", result.ID)
	}
	if result.SourcePath != "/old" {
		t.Errorf("Expected sourcePath /old, got %s", result.SourcePath)
	}
}

// TestPostRedirect_ValidationError tests 400 handling
func TestPostRedirect_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid redirect configuration"))
	}))
	defer server.Close()

	oldURL := postRedirectBaseURL
	postRedirectBaseURL = server.URL
	defer func() { postRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PostRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "invalid", "/new", 301)
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestPatchRedirect_Valid tests updating a redirect successfully
func TestPatchRedirect_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req RedirectRequest
		_ = json.Unmarshal(body, &req)

		if req.DestinationPath != "/updated" {
			t.Errorf("Expected destinationPath /updated, got %s", req.DestinationPath)
		}
		if req.StatusCode != 302 {
			t.Errorf("Expected statusCode 302, got %d", req.StatusCode)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := RedirectRule{ID: "redirect1", SourcePath: "/old", DestinationPath: "/updated", StatusCode: 302}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := patchRedirectBaseURL
	patchRedirectBaseURL = server.URL
	defer func() { patchRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PatchRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "redirect1", "/old", "/updated", 302)
	if err != nil {
		t.Fatalf("PatchRedirect failed: %v", err)
	}

	if result.DestinationPath != "/updated" {
		t.Errorf("Expected destinationPath /updated, got %s", result.DestinationPath)
	}
	if result.StatusCode != 302 {
		t.Errorf("Expected statusCode 302, got %d", result.StatusCode)
	}
}

// TestPatchRedirect_NotFound tests 404 handling for update
func TestPatchRedirect_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("redirect not found"))
	}))
	defer server.Close()

	oldURL := patchRedirectBaseURL
	patchRedirectBaseURL = server.URL
	defer func() { patchRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PatchRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "nonexistent", "/old", "/new", 301)
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestDeleteRedirect_Valid tests deleting a redirect successfully
func TestDeleteRedirect_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	oldURL := deleteRedirectBaseURL
	deleteRedirectBaseURL = server.URL
	defer func() { deleteRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "redirect1")
	if err != nil {
		t.Fatalf("DeleteRedirect failed: %v", err)
	}
}

// TestDeleteRedirect_NotFound_Idempotent tests that 404 on delete is treated as success (idempotent)
func TestDeleteRedirect_NotFound_Idempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("redirect not found"))
	}))
	defer server.Close()

	oldURL := deleteRedirectBaseURL
	deleteRedirectBaseURL = server.URL
	defer func() { deleteRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "nonexistent")
	if err != nil {
		t.Errorf("DeleteRedirect should handle 404 as success (idempotent), got error: %v", err)
	}
}

// TestDeleteRedirect_ServerError tests error handling
func TestDeleteRedirect_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := deleteRedirectBaseURL
	deleteRedirectBaseURL = server.URL
	defer func() { deleteRedirectBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteRedirect(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "redirect1")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestRedirectDiff_StatusCodeZeroNoDrift tests that statusCode: 0 in state
// (API didn't return it) doesn't cause drift when input specifies a valid code.
func TestRedirectDiff_StatusCodeZeroNoDrift(t *testing.T) {
	r := &Redirect{}

	// Simulate state where API returned statusCode 0 (not in response)
	// but user specified statusCode 301 in their Pulumi program
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		ID: "siteid/redirects/redirectid",
		Inputs: RedirectArgs{
			SiteID:          "siteid",
			SourcePath:      "/old",
			DestinationPath: "/new",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          "siteid",
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      0, // API list endpoint doesn't return statusCode
			},
		},
	}

	ctx := context.Background()
	result, err := r.Diff(ctx, req)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	// Should NOT report changes when state.StatusCode is 0
	if result.HasChanges {
		t.Error("Expected no changes when state.StatusCode is 0, but HasChanges is true")
	}
	if _, ok := result.DetailedDiff["statusCode"]; ok {
		t.Error("Expected no statusCode diff when state is 0, but diff was reported")
	}
}

// TestRedirectDiff_StatusCodeActualChange tests that actual statusCode changes are detected.
func TestRedirectDiff_StatusCodeActualChange(t *testing.T) {
	r := &Redirect{}

	// Simulate state where API returned statusCode 301
	// but user changed their Pulumi program to 302
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		ID: "siteid/redirects/redirectid",
		Inputs: RedirectArgs{
			SiteID:          "siteid",
			SourcePath:      "/old",
			DestinationPath: "/new",
			StatusCode:      302, // User changed to 302
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          "siteid",
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      301, // Was 301
			},
		},
	}

	ctx := context.Background()
	result, err := r.Diff(ctx, req)
	if err != nil {
		t.Fatalf("Diff returned error: %v", err)
	}

	// Should report changes when statusCode actually changed
	if !result.HasChanges {
		t.Error("Expected changes when statusCode changed from 301 to 302, but HasChanges is false")
	}
	if _, ok := result.DetailedDiff["statusCode"]; !ok {
		t.Error("Expected statusCode diff when changing 301->302, but no diff was reported")
	}
}
