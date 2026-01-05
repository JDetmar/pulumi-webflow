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
)

// Note: ValidateDisplayName tests are in site_test.go since it's a shared function.

// TestValidateAssetFolderID_Valid tests valid asset folder IDs
func TestValidateAssetFolderID_Valid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"typical ID", "5f0c8c9e1c9d440000e8d8c3"},
		{"all lowercase hex", "abcdef1234567890abcdef12"},
		{"all digits", "123456789012345678901234"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssetFolderID(tt.id)
			if err != nil {
				t.Errorf("ValidateAssetFolderID(%q) = %v, want nil", tt.id, err)
			}
		})
	}
}

// TestValidateAssetFolderID_Empty tests empty asset folder ID
func TestValidateAssetFolderID_Empty(t *testing.T) {
	err := ValidateAssetFolderID("")
	if err == nil {
		t.Error("ValidateAssetFolderID(\"\") = nil, want error")
	}
	if !strings.Contains(err.Error(), "required") {
		t.Errorf("Expected error to mention 'required', got: %v", err)
	}
}

// TestValidateAssetFolderID_Invalid tests invalid asset folder IDs
func TestValidateAssetFolderID_Invalid(t *testing.T) {
	tests := []struct {
		name string
		id   string
	}{
		{"too short", "5f0c8c9e1c9d44"},
		{"too long", "5f0c8c9e1c9d440000e8d8c3extra"},
		{"uppercase letters", "5F0C8C9E1C9D440000E8D8C3"},
		{"invalid characters", "5f0c8c9e1c9d440000e8d8g3"},
		{"with spaces", "5f0c8c9e1c9d 40000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssetFolderID(tt.id)
			if err == nil {
				t.Errorf("ValidateAssetFolderID(%q) = nil, want error", tt.id)
			}
		})
	}
}

// TestGenerateAssetFolderResourceID tests resource ID generation
func TestGenerateAssetFolderResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	folderID := "6390c49774a71f0e3c1a08ee"

	resourceID := GenerateAssetFolderResourceID(siteID, folderID)
	expected := "5f0c8c9e1c9d440000e8d8c3/asset-folders/6390c49774a71f0e3c1a08ee"

	if resourceID != expected {
		t.Errorf("GenerateAssetFolderResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromAssetFolderResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromAssetFolderResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/asset-folders/6390c49774a71f0e3c1a08ee"

	siteID, folderID, err := ExtractIDsFromAssetFolderResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromAssetFolderResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromAssetFolderResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if folderID != "6390c49774a71f0e3c1a08ee" {
		t.Errorf("ExtractIDsFromAssetFolderResourceID() folderID = %q, want %q", folderID, "6390c49774a71f0e3c1a08ee")
	}
}

// TestExtractIDsFromAssetFolderResourceID_Empty tests empty resource ID
func TestExtractIDsFromAssetFolderResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromAssetFolderResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromAssetFolderResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromAssetFolderResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromAssetFolderResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing asset-folders part", "5f0c8c9e1c9d440000e8d8c3/6390c49774a71f0e3c1a08ee"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/folders/6390c49774a71f0e3c1a08ee"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromAssetFolderResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromAssetFolderResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestListAssetFolders_Valid tests retrieving asset folders successfully
func TestListAssetFolders_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/asset_folders") {
			t.Errorf("Expected /asset_folders in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetFolderListResponse{
			AssetFolders: []AssetFolderResponse{
				{
					ID:           "6390c49774a71f0e3c1a08ee",
					DisplayName:  "Images",
					ParentFolder: "",
					Assets:       []string{"asset1", "asset2"},
					SiteID:       "5f0c8c9e1c9d440000e8d8c3",
					CreatedOn:    "2024-01-01T00:00:00Z",
					LastUpdated:  "2024-01-02T00:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := listAssetFoldersBaseURL
	listAssetFoldersBaseURL = server.URL
	defer func() { listAssetFoldersBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := ListAssetFolders(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("ListAssetFolders failed: %v", err)
	}

	if len(result.AssetFolders) != 1 {
		t.Errorf("Expected 1 folder, got %d", len(result.AssetFolders))
	}
	if result.AssetFolders[0].DisplayName != "Images" {
		t.Errorf("Expected DisplayName 'Images', got %s", result.AssetFolders[0].DisplayName)
	}
}

// TestListAssetFolders_NotFound tests 404 handling
func TestListAssetFolders_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := listAssetFoldersBaseURL
	listAssetFoldersBaseURL = server.URL
	defer func() { listAssetFoldersBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := ListAssetFolders(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetAssetFolder_Valid tests retrieving a single asset folder successfully
func TestGetAssetFolder_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetFolderResponse{
			ID:           "6390c49774a71f0e3c1a08ee",
			DisplayName:  "Documents",
			ParentFolder: "parent123",
			Assets:       []string{"doc1", "doc2"},
			SiteID:       "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-02T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getAssetFolderBaseURL
	getAssetFolderBaseURL = server.URL
	defer func() { getAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetAssetFolder(ctx, client, "6390c49774a71f0e3c1a08ee")
	if err != nil {
		t.Fatalf("GetAssetFolder failed: %v", err)
	}

	if result.ID != "6390c49774a71f0e3c1a08ee" {
		t.Errorf("Expected ID '6390c49774a71f0e3c1a08ee', got %s", result.ID)
	}
	if result.DisplayName != "Documents" {
		t.Errorf("Expected DisplayName 'Documents', got %s", result.DisplayName)
	}
	if result.ParentFolder != "parent123" {
		t.Errorf("Expected ParentFolder 'parent123', got %s", result.ParentFolder)
	}
}

// TestGetAssetFolder_NotFound tests 404 handling
func TestGetAssetFolder_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("folder not found"))
	}))
	defer server.Close()

	oldURL := getAssetFolderBaseURL
	getAssetFolderBaseURL = server.URL
	defer func() { getAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetAssetFolder(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestPostAssetFolder_Valid tests creating an asset folder successfully
func TestPostAssetFolder_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		body, _ := io.ReadAll(r.Body)
		var req AssetFolderCreateRequest
		_ = json.Unmarshal(body, &req)

		if req.DisplayName != "New Folder" {
			t.Errorf("Expected displayName 'New Folder', got %s", req.DisplayName)
		}
		if req.ParentFolder != "parent123" {
			t.Errorf("Expected parentFolder 'parent123', got %s", req.ParentFolder)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetFolderResponse{
			ID:           "new-folder-id-123",
			DisplayName:  "New Folder",
			ParentFolder: "parent123",
			Assets:       []string{},
			SiteID:       "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postAssetFolderBaseURL
	postAssetFolderBaseURL = server.URL
	defer func() { postAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostAssetFolder(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "New Folder", "parent123")
	if err != nil {
		t.Fatalf("PostAssetFolder failed: %v", err)
	}

	if result.ID != "new-folder-id-123" {
		t.Errorf("Expected ID 'new-folder-id-123', got %s", result.ID)
	}
	if result.DisplayName != "New Folder" {
		t.Errorf("Expected DisplayName 'New Folder', got %s", result.DisplayName)
	}
}

// TestPostAssetFolder_WithoutParent tests creating a root-level folder
func TestPostAssetFolder_WithoutParent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		body, _ := io.ReadAll(r.Body)
		var req AssetFolderCreateRequest
		_ = json.Unmarshal(body, &req)

		if req.ParentFolder != "" {
			t.Errorf("Expected empty parentFolder, got %s", req.ParentFolder)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetFolderResponse{
			ID:          "root-folder-id",
			DisplayName: "Root Folder",
			SiteID:      "5f0c8c9e1c9d440000e8d8c3",
			CreatedOn:   "2024-01-01T00:00:00Z",
			LastUpdated: "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postAssetFolderBaseURL
	postAssetFolderBaseURL = server.URL
	defer func() { postAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostAssetFolder(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "Root Folder", "")
	if err != nil {
		t.Fatalf("PostAssetFolder failed: %v", err)
	}

	if result.ID != "root-folder-id" {
		t.Errorf("Expected ID 'root-folder-id', got %s", result.ID)
	}
}

// TestPostAssetFolder_ValidationError tests 400 handling
func TestPostAssetFolder_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid folder configuration"))
	}))
	defer server.Close()

	oldURL := postAssetFolderBaseURL
	postAssetFolderBaseURL = server.URL
	defer func() { postAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PostAssetFolder(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "", "")
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestPostAssetFolder_ServerError tests 500 handling
func TestPostAssetFolder_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := postAssetFolderBaseURL
	postAssetFolderBaseURL = server.URL
	defer func() { postAssetFolderBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PostAssetFolder(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "Test", "")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestErrorMessagesAreActionable_AssetFolder verifies error messages contain guidance
func TestErrorMessagesAreActionable_AssetFolder(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateAssetFolderID empty",
			func() error { return ValidateAssetFolderID("") },
			[]string{"required", "24-character"},
		},
		{
			"ValidateAssetFolderID invalid",
			func() error { return ValidateAssetFolderID("invalid") },
			[]string{"invalid format", "24-character"},
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
