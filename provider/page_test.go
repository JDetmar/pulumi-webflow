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

	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestValidatePageID_Valid tests valid page IDs
func TestValidatePageID_Valid(t *testing.T) {
	tests := []struct {
		name   string
		pageID string
	}{
		{"valid page ID", "5f0c8c9e1c9d440000e8d8c4"},
		{"another valid ID", "507f1f77bcf86cd799439011"},
		{"all lowercase hex", "abcdef0123456789abcdef01"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageID(tt.pageID)
			if err != nil {
				t.Errorf("ValidatePageID(%q) = %v, want nil", tt.pageID, err)
			}
		})
	}
}

// TestValidatePageID_Empty tests empty page ID
func TestValidatePageID_Empty(t *testing.T) {
	err := ValidatePageID("")
	if err == nil {
		t.Error("ValidatePageID(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidatePageID_InvalidFormat tests invalid page ID formats
func TestValidatePageID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name   string
		pageID string
	}{
		{"too short", "5f0c8c9e1c9d"},
		{"too long", "5f0c8c9e1c9d440000e8d8c4000"},
		{"uppercase letters", "5F0C8C9E1C9D440000E8D8C4"},
		{"invalid characters", "5g0c8c9e1c9d440000e8d8c4"},
		{"with spaces", "5f0c8c9e 1c9d440000e8d8c4"},
		{"with dashes", "5f0c8c9e-1c9d-4400-00e8d8c4"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageID(tt.pageID)
			if err == nil {
				t.Errorf("ValidatePageID(%q) = nil, want error", tt.pageID)
			}
			if !strings.Contains(err.Error(), "invalid format") {
				t.Errorf("Expected error to mention 'invalid format', got: %v", err)
			}
		})
	}
}

// TestGeneratePageResourceID tests resource ID generation
func TestGeneratePageResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	pageID := "5f0c8c9e1c9d440000e8d8c4"

	resourceID := GeneratePageResourceID(siteID, pageID)
	expected := "5f0c8c9e1c9d440000e8d8c3/pages/5f0c8c9e1c9d440000e8d8c4"

	if resourceID != expected {
		t.Errorf("GeneratePageResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromPageResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromPageResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/pages/5f0c8c9e1c9d440000e8d8c4"

	siteID, pageID, err := ExtractIDsFromPageResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromPageResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromPageResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if pageID != "5f0c8c9e1c9d440000e8d8c4" {
		t.Errorf("ExtractIDsFromPageResourceID() pageID = %q, want %q", pageID, "5f0c8c9e1c9d440000e8d8c4")
	}
}

// TestExtractIDsFromPageResourceID_Empty tests empty resource ID
func TestExtractIDsFromPageResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromPageResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromPageResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromPageResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromPageResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing pages part", "5f0c8c9e1c9d440000e8d8c3/5f0c8c9e1c9d440000e8d8c4"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/redirects/5f0c8c9e1c9d440000e8d8c4"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
		{"empty site ID", "/pages/5f0c8c9e1c9d440000e8d8c4"},
		{"empty page ID", "5f0c8c9e1c9d440000e8d8c3/pages/"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromPageResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromPageResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestErrorMessagesAreActionable verifies error messages contain guidance
func TestPageErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidatePageID empty",
			func() error { return ValidatePageID("") },
			[]string{"required", "24-character"},
		},
		{
			"ValidatePageID invalid format",
			func() error { return ValidatePageID("invalid") },
			[]string{"invalid format", "24-character", "lowercase"},
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

// TestGetPages_Valid tests retrieving pages successfully
func TestGetPages_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/pages") {
			t.Errorf("Expected /pages in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := PagesResponse{
			Pages: []Page{
				{
					ID:          "page1",
					SiteID:      "5f0c8c9e1c9d440000e8d8c3",
					Title:       "Home",
					Slug:        "home",
					CreatedOn:   "2024-01-01T00:00:00Z",
					LastUpdated: "2024-01-02T00:00:00Z",
					Archived:    false,
					Draft:       false,
				},
				{
					ID:          "page2",
					SiteID:      "5f0c8c9e1c9d440000e8d8c3",
					Title:       "About",
					Slug:        "about",
					CreatedOn:   "2024-01-01T00:00:00Z",
					LastUpdated: "2024-01-02T00:00:00Z",
					Archived:    false,
					Draft:       false,
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetPages(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetPages failed: %v", err)
	}

	if len(result.Pages) != 2 {
		t.Errorf("Expected 2 pages, got %d", len(result.Pages))
	}
	if result.Pages[0].ID != "page1" {
		t.Errorf("Expected page1, got %s", result.Pages[0].ID)
	}
	if result.Pages[0].Title != "Home" {
		t.Errorf("Expected title 'Home', got %s", result.Pages[0].Title)
	}
}

// TestGetPages_NotFound tests 404 handling
func TestGetPages_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetPages(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetPages_EmptyList tests successful response with no pages
func TestGetPages_EmptyList(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := PagesResponse{
			Pages: []Page{},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetPages(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetPages failed: %v", err)
	}

	if len(result.Pages) != 0 {
		t.Errorf("Expected 0 pages, got %d", len(result.Pages))
	}
}

// TestGetPage_Valid tests retrieving a single page successfully
func TestGetPage_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/pages/") {
			t.Errorf("Expected /pages/ in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		page := Page{
			ID:           "page1",
			SiteID:       "5f0c8c9e1c9d440000e8d8c3",
			Title:        "Home",
			Slug:         "home",
			ParentID:     "",
			CollectionID: "",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-02T00:00:00Z",
			Archived:     false,
			Draft:        false,
		}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	oldURL := getPageBaseURL
	getPageBaseURL = server.URL
	defer func() { getPageBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetPage(ctx, client, "page1")
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	if result.ID != "page1" {
		t.Errorf("Expected ID page1, got %s", result.ID)
	}
	if result.Title != "Home" {
		t.Errorf("Expected title 'Home', got %s", result.Title)
	}
	if result.Slug != "home" {
		t.Errorf("Expected slug 'home', got %s", result.Slug)
	}
}

// TestGetPage_NotFound tests 404 handling for single page
func TestGetPage_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("page not found"))
	}))
	defer server.Close()

	oldURL := getPageBaseURL
	getPageBaseURL = server.URL
	defer func() { getPageBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetPage(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetPage_WithParentAndCollection tests page with optional fields
func TestGetPage_WithParentAndCollection(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		page := Page{
			ID:           "page1",
			SiteID:       "5f0c8c9e1c9d440000e8d8c3",
			Title:        "Nested Page",
			Slug:         "nested",
			ParentID:     "parent123",
			CollectionID: "collection456",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-02T00:00:00Z",
			Archived:     false,
			Draft:        true,
		}
		_ = json.NewEncoder(w).Encode(page)
	}))
	defer server.Close()

	oldURL := getPageBaseURL
	getPageBaseURL = server.URL
	defer func() { getPageBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetPage(ctx, client, "page1")
	if err != nil {
		t.Fatalf("GetPage failed: %v", err)
	}

	if result.ParentID != "parent123" {
		t.Errorf("Expected ParentID parent123, got %s", result.ParentID)
	}
	if result.CollectionID != "collection456" {
		t.Errorf("Expected CollectionID collection456, got %s", result.CollectionID)
	}
	if !result.Draft {
		t.Error("Expected Draft to be true, got false")
	}
}

// TestGetPages_Unauthorized tests 401 handling
func TestGetPages_Unauthorized(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusUnauthorized)
		_, _ = w.Write([]byte("unauthorized"))
	}))
	defer server.Close()

	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetPages(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 401, got nil")
	}
	if !strings.Contains(err.Error(), "unauthorized") {
		t.Errorf("Expected 'unauthorized' in error, got: %v", err)
	}
}

// TestGetPages_Forbidden tests 403 handling
func TestGetPages_Forbidden(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusForbidden)
		_, _ = w.Write([]byte("forbidden"))
	}))
	defer server.Close()

	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetPages(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 403, got nil")
	}
	if !strings.Contains(err.Error(), "forbidden") {
		t.Errorf("Expected 'forbidden' in error, got: %v", err)
	}
}

// TestGetPages_ServerError tests 500 handling
func TestGetPages_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := getPagesBaseURL
	getPagesBaseURL = server.URL
	defer func() { getPagesBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetPages(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestPageDataRead_NotFound tests that Read() returns empty ID when page is not found.
// This is critical for Pulumi to detect resource deletion and trigger recreation.
func TestPageDataRead_NotFound(t *testing.T) {
	// t.Setenv automatically restores the original value after the test
	t.Setenv("WEBFLOW_API_TOKEN", "wfp_1234567890abcdef1234567890abcdef1234567890abcdef1234567890abcd")

	// Setup mock server that returns 404 for GetPage
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if strings.Contains(r.URL.Path, "/pages/") {
			w.WriteHeader(http.StatusNotFound)
			_, _ = w.Write([]byte("page not found"))
			return
		}
		t.Errorf("Unexpected request path: %s", r.URL.Path)
		w.WriteHeader(http.StatusInternalServerError)
	}))
	defer server.Close()

	// Override API base URL
	oldURL := getPageBaseURL
	getPageBaseURL = server.URL
	defer func() { getPageBaseURL = oldURL }()

	// Create PageData resource
	pageData := &PageData{}
	ctx := context.Background()

	// Create a test request as if Pulumi is calling Read()
	state := PageDataState{
		PageDataArgs: PageDataArgs{
			SiteID: "5f0c8c9e1c9d440000e8d8c3",
			PageID: "5f0c8c9e1c9d440000e8d8c4", // Valid format page ID
		},
	}

	req := infer.ReadRequest[PageDataArgs, PageDataState]{
		ID:    "5f0c8c9e1c9d440000e8d8c3/pages/5f0c8c9e1c9d440000e8d8c4",
		State: state,
	}

	// Call Read() - should return empty ID for "not found" error
	response, err := pageData.Read(ctx, req)
	// Verify no error is returned (not found is handled gracefully)
	if err != nil {
		t.Errorf("Read() returned error for not found case, expected nil: %v", err)
	}

	// Verify that ID is empty (signals deletion to Pulumi)
	if response.ID != "" {
		t.Errorf("Read() returned ID = %q, expected empty string to signal deletion", response.ID)
	}
}
