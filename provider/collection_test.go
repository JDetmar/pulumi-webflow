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

// TestValidateCollectionID tests collectionID validation
func TestValidateCollectionID(t *testing.T) {
	tests := []struct {
		name         string
		collectionID string
		wantErr      bool
	}{
		{"valid collection ID", "5f0c8c9e1c9d440000e8d8c3", false},
		{"valid collection ID 2", "abcdef0123456789abcdef01", false},
		{"empty collection ID", "", true},
		{"too short", "5f0c8c9e1c9d440000e8d8", true},
		{"too long", "5f0c8c9e1c9d440000e8d8c3aa", true},
		{"uppercase letters", "5F0C8C9E1C9D440000E8D8C3", true},
		{"invalid characters", "5f0c8c9e1c9d440000e8d8cg", true},
		{"with spaces", "5f0c8c9e 1c9d440000e8d8c3", true},
		{"with dashes", "5f0c8c9e-1c9d-4400-00e8d8c3", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCollectionID(tt.collectionID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCollectionID(%s) error = %v, wantErr %v", tt.collectionID, err, tt.wantErr)
			}
		})
	}
}

// TestValidateCollectionDisplayName tests displayName validation
func TestValidateCollectionDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{"valid name", "Blog Posts", false},
		{"valid name with numbers", "Products 2024", false},
		{"valid name with special chars", "Team Members & Partners", false},
		{"empty name", "", true},
		{"very long name", strings.Repeat("a", 256), true},
		{"max length name", strings.Repeat("a", 255), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateCollectionDisplayName(tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateCollectionDisplayName(%s) error = %v, wantErr %v", tt.displayName, err, tt.wantErr)
			}
		})
	}
}

// TestValidateSingularName tests singularName validation
func TestValidateSingularName(t *testing.T) {
	tests := []struct {
		name         string
		singularName string
		wantErr      bool
	}{
		{"valid singular name", "Blog Post", false},
		{"valid singular name 2", "Product", false},
		{"empty singular name", "", true},
		{"very long name", strings.Repeat("a", 256), true},
		{"max length name", strings.Repeat("a", 255), false},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSingularName(tt.singularName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSingularName(%s) error = %v, wantErr %v", tt.singularName, err, tt.wantErr)
			}
		})
	}
}

// TestGenerateCollectionResourceID tests resource ID generation
func TestGenerateCollectionResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	collectionID := "abc123def456789012345678"

	resourceID := GenerateCollectionResourceID(siteID, collectionID)
	expected := "5f0c8c9e1c9d440000e8d8c3/collections/abc123def456789012345678"

	if resourceID != expected {
		t.Errorf("GenerateCollectionResourceID() = %q, want %q", resourceID, expected)
	}
}

// TestExtractIDsFromCollectionResourceID_Valid tests extracting IDs from valid resource ID
func TestExtractIDsFromCollectionResourceID_Valid(t *testing.T) {
	resourceID := "5f0c8c9e1c9d440000e8d8c3/collections/abc123def456789012345678"

	siteID, collectionID, err := ExtractIDsFromCollectionResourceID(resourceID)
	if err != nil {
		t.Errorf("ExtractIDsFromCollectionResourceID() error = %v, want nil", err)
	}
	if siteID != "5f0c8c9e1c9d440000e8d8c3" {
		t.Errorf("ExtractIDsFromCollectionResourceID() siteID = %q, want %q", siteID, "5f0c8c9e1c9d440000e8d8c3")
	}
	if collectionID != "abc123def456789012345678" {
		t.Errorf("ExtractIDsFromCollectionResourceID() collectionID = %q, want %q", collectionID, "abc123def456789012345678")
	}
}

// TestExtractIDsFromCollectionResourceID_Empty tests empty resource ID
func TestExtractIDsFromCollectionResourceID_Empty(t *testing.T) {
	_, _, err := ExtractIDsFromCollectionResourceID("")
	if err == nil {
		t.Error("ExtractIDsFromCollectionResourceID(\"\") error = nil, want error")
	}
}

// TestExtractIDsFromCollectionResourceID_InvalidFormat tests invalid format
func TestExtractIDsFromCollectionResourceID_InvalidFormat(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
	}{
		{"missing collections part", "5f0c8c9e1c9d440000e8d8c3/abc123def456789012345678"},
		{"wrong middle part", "5f0c8c9e1c9d440000e8d8c3/redirects/abc123def456789012345678"},
		{"too few parts", "5f0c8c9e1c9d440000e8d8c3"},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromCollectionResourceID(tt.resourceID)
			if err == nil {
				t.Errorf("ExtractIDsFromCollectionResourceID(%q) error = nil, want error", tt.resourceID)
			}
		})
	}
}

// TestGetCollections_Valid tests retrieving collections successfully
func TestGetCollections_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}
		if !strings.Contains(r.URL.Path, "/collections") {
			t.Errorf("Expected /collections in path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := CollectionListResponse{
			Collections: []Collection{
				{
					ID:           "collection1",
					DisplayName:  "Blog Posts",
					SingularName: "Blog Post",
					Slug:         "blog-posts",
					CreatedOn:    "2024-01-01T00:00:00Z",
					LastUpdated:  "2024-01-02T00:00:00Z",
				},
				{
					ID:           "collection2",
					DisplayName:  "Products",
					SingularName: "Product",
					Slug:         "products",
					CreatedOn:    "2024-01-01T00:00:00Z",
					LastUpdated:  "2024-01-02T00:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL for this test
	oldURL := getCollectionsBaseURL
	getCollectionsBaseURL = server.URL
	defer func() { getCollectionsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetCollections(ctx, client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("GetCollections failed: %v", err)
	}

	if len(result.Collections) != 2 {
		t.Errorf("Expected 2 collections, got %d", len(result.Collections))
	}
	if result.Collections[0].ID != "collection1" {
		t.Errorf("Expected collection1, got %s", result.Collections[0].ID)
	}
}

// TestGetCollections_NotFound tests 404 handling
func TestGetCollections_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("site not found"))
	}))
	defer server.Close()

	oldURL := getCollectionsBaseURL
	getCollectionsBaseURL = server.URL
	defer func() { getCollectionsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := GetCollections(ctx, client, "nonexistent")
	if err == nil {
		t.Error("Expected error for 404, got nil")
	}
	if !strings.Contains(err.Error(), "not found") {
		t.Errorf("Expected 'not found' in error, got: %v", err)
	}
}

// TestGetCollection_Valid tests retrieving a single collection successfully
func TestGetCollection_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := Collection{
			ID:           "collection1",
			DisplayName:  "Blog Posts",
			SingularName: "Blog Post",
			Slug:         "blog-posts",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-02T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getCollectionBaseURL
	getCollectionBaseURL = server.URL
	defer func() { getCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := GetCollection(ctx, client, "collection1")
	if err != nil {
		t.Fatalf("GetCollection failed: %v", err)
	}

	if result.ID != "collection1" {
		t.Errorf("Expected collection1, got %s", result.ID)
	}
	if result.DisplayName != "Blog Posts" {
		t.Errorf("Expected 'Blog Posts', got %s", result.DisplayName)
	}
}

// TestPostCollection_Valid tests creating a collection successfully
func TestPostCollection_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		var req CollectionRequest
		_ = json.NewDecoder(r.Body).Decode(&req)

		if req.DisplayName != "Blog Posts" {
			t.Errorf("Expected displayName 'Blog Posts', got %s", req.DisplayName)
		}
		if req.SingularName != "Blog Post" {
			t.Errorf("Expected singularName 'Blog Post', got %s", req.SingularName)
		}
		if req.Slug != "blog-posts" {
			t.Errorf("Expected slug 'blog-posts', got %s", req.Slug)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		response := Collection{
			ID:           "new-collection-1",
			DisplayName:  "Blog Posts",
			SingularName: "Blog Post",
			Slug:         "blog-posts",
			CreatedOn:    "2024-01-01T00:00:00Z",
			LastUpdated:  "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := postCollectionBaseURL
	postCollectionBaseURL = server.URL
	defer func() { postCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	result, err := PostCollection(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "Blog Posts", "Blog Post", "blog-posts")
	if err != nil {
		t.Fatalf("PostCollection failed: %v", err)
	}

	if result.ID != "new-collection-1" {
		t.Errorf("Expected ID new-collection-1, got %s", result.ID)
	}
	if result.DisplayName != "Blog Posts" {
		t.Errorf("Expected displayName 'Blog Posts', got %s", result.DisplayName)
	}
}

// TestPostCollection_ValidationError tests 400 handling
func TestPostCollection_ValidationError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusBadRequest)
		_, _ = w.Write([]byte("invalid collection configuration"))
	}))
	defer server.Close()

	oldURL := postCollectionBaseURL
	postCollectionBaseURL = server.URL
	defer func() { postCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	_, err := PostCollection(ctx, client, "5f0c8c9e1c9d440000e8d8c3", "", "Blog Post", "")
	if err == nil {
		t.Error("Expected error for 400, got nil")
	}
	if !strings.Contains(err.Error(), "bad request") {
		t.Errorf("Expected 'bad request' in error, got: %v", err)
	}
}

// TestDeleteCollection_Valid tests deleting a collection successfully
func TestDeleteCollection_Valid(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	oldURL := deleteCollectionBaseURL
	deleteCollectionBaseURL = server.URL
	defer func() { deleteCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteCollection(ctx, client, "collection1")
	if err != nil {
		t.Fatalf("DeleteCollection failed: %v", err)
	}
}

// TestDeleteCollection_NotFound_Idempotent tests that 404 on delete is treated as success (idempotent)
func TestDeleteCollection_NotFound_Idempotent(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte("collection not found"))
	}))
	defer server.Close()

	oldURL := deleteCollectionBaseURL
	deleteCollectionBaseURL = server.URL
	defer func() { deleteCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteCollection(ctx, client, "nonexistent")
	if err != nil {
		t.Errorf("DeleteCollection should handle 404 as success (idempotent), got error: %v", err)
	}
}

// TestDeleteCollection_ServerError tests error handling
func TestDeleteCollection_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_, _ = w.Write([]byte("server error"))
	}))
	defer server.Close()

	oldURL := deleteCollectionBaseURL
	deleteCollectionBaseURL = server.URL
	defer func() { deleteCollectionBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	err := DeleteCollection(ctx, client, "collection1")
	if err == nil {
		t.Error("Expected error for 500, got nil")
	}
	if !strings.Contains(err.Error(), "server error") {
		t.Errorf("Expected 'server error' in error, got: %v", err)
	}
}

// TestErrorMessagesAreActionable verifies error messages contain guidance
func TestCollectionErrorMessagesAreActionable(t *testing.T) {
	tests := []struct {
		name     string
		testFunc func() error
		contains []string
	}{
		{
			"ValidateCollectionID empty",
			func() error { return ValidateCollectionID("") },
			[]string{"required", "24-character"},
		},
		{
			"ValidateCollectionID invalid format",
			func() error { return ValidateCollectionID("invalid") },
			[]string{"invalid format", "24-character", "hexadecimal"},
		},
		{
			"ValidateCollectionDisplayName empty",
			func() error { return ValidateCollectionDisplayName("") },
			[]string{"required", "name"},
		},
		{
			"ValidateSingularName empty",
			func() error { return ValidateSingularName("") },
			[]string{"required", "singular"},
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
