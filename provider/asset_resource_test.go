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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestAssetRead_NotFound tests that Read() returns empty ID when asset is not found.
// This test verifies the bug fix for the error comparison at line 320 in asset_resource.go.
func TestAssetRead_NotFound(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that returns 404 Not Found
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify we're calling the GET /v2/assets/{asset_id} endpoint
		if r.Method != "GET" || !strings.Contains(r.URL.Path, "/v2/assets/") {
			t.Errorf("Expected GET request to /v2/assets/{asset_id}, got %s %s", r.Method, r.URL.Path)
		}

		// Return 404 Not Found (asset was deleted in Webflow)
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		response := map[string]string{
			"message": "Asset not found",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override the API base URL to use our test server
	oldURL := getAssetBaseURL
	getAssetBaseURL = server.URL
	defer func() { getAssetBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}

	// Setup: Create context
	ctx := context.Background()

	// Setup: Build a read request with existing state
	resourceID := "5f0c8c9e1c9d440000e8d8c3/assets/6789abcdef1234567890abcd"
	req := infer.ReadRequest[AssetArgs, AssetState]{
		ID: resourceID,
		State: AssetState{
			AssetArgs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			AssetID: "6789abcdef1234567890abcd",
		},
	}

	// Execute: Call Read()
	resp, err := asset.Read(ctx, req)
	// Verify: No error should be returned
	if err != nil {
		t.Errorf("Read() should not return error for 404, got: %v", err)
	}

	// Verify: Empty ID signals deletion to Pulumi
	if resp.ID != "" {
		t.Errorf("Read() should return empty ID for 404 (not found), got: %s", resp.ID)
	}
}

// TestAssetRead_Success tests that Read() correctly retrieves and returns asset state.
func TestAssetRead_Success(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that returns asset details
	expectedAsset := AssetResponse{
		ID:               "6789abcdef1234567890abcd",
		ContentType:      "image/png",
		Size:             12345,
		SiteID:           "5f0c8c9e1c9d440000e8d8c3",
		HostedURL:        "https://assets.website-files.com/.../logo.png",
		OriginalFileName: "logo.png",
		DisplayName:      "logo.png",
		CreatedOn:        "2025-01-12T10:00:00Z",
		LastUpdated:      "2025-01-12T10:00:00Z",
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify we're calling the GET /v2/assets/{asset_id} endpoint
		if r.Method != "GET" || !strings.Contains(r.URL.Path, "/v2/assets/") {
			t.Errorf("Expected GET request to /v2/assets/{asset_id}, got %s %s", r.Method, r.URL.Path)
		}

		// Return 200 OK with asset details
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(expectedAsset)
	}))
	defer server.Close()

	// Override the API base URL to use our test server
	oldURL := getAssetBaseURL
	getAssetBaseURL = server.URL
	defer func() { getAssetBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}

	// Setup: Create context
	ctx := context.Background()

	// Setup: Build a read request with existing state
	resourceID := "5f0c8c9e1c9d440000e8d8c3/assets/6789abcdef1234567890abcd"
	req := infer.ReadRequest[AssetArgs, AssetState]{
		ID: resourceID,
		State: AssetState{
			AssetArgs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			AssetID: "6789abcdef1234567890abcd",
		},
	}

	// Execute: Call Read()
	resp, err := asset.Read(ctx, req)
	// Verify: No error should be returned
	if err != nil {
		t.Errorf("Read() should not return error for successful read, got: %v", err)
	}

	// Verify: Resource ID should be preserved
	if resp.ID != resourceID {
		t.Errorf("Read() should preserve resource ID, expected: %s, got: %s", resourceID, resp.ID)
	}

	// Verify: State should match API response
	if resp.State.AssetID != expectedAsset.ID {
		t.Errorf("Expected AssetID=%s, got %s", expectedAsset.ID, resp.State.AssetID)
	}
	if resp.State.HostedURL != expectedAsset.HostedURL {
		t.Errorf("Expected HostedURL=%s, got %s", expectedAsset.HostedURL, resp.State.HostedURL)
	}
	if resp.State.ContentType != expectedAsset.ContentType {
		t.Errorf("Expected ContentType=%s, got %s", expectedAsset.ContentType, resp.State.ContentType)
	}
	if resp.State.Size != expectedAsset.Size {
		t.Errorf("Expected Size=%d, got %d", expectedAsset.Size, resp.State.Size)
	}
}

// =============================================================================
// Create Operation Tests
// =============================================================================

// TestAssetCreate_Success tests successful asset creation with mocked API
func TestAssetCreate_Success(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that returns asset upload response
	expectedResponse := AssetUploadResponse{
		ID:               "6789abcdef1234567890abcd",
		UploadURL:        "https://s3.amazonaws.com/test-bucket/upload",
		AssetURL:         "https://s3.amazonaws.com/test-bucket/logo.png",
		HostedURL:        "https://assets.website-files.com/.../logo.png",
		ContentType:      "image/png",
		OriginalFileName: "logo.png",
		CreatedOn:        "2025-01-12T10:00:00Z",
		LastUpdated:      "2025-01-12T10:00:00Z",
		UploadDetails: map[string]string{
			"acl":    "public-read",
			"bucket": "test-bucket",
			"key":    "logo.png",
		},
	}

	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint and method
		if r.Method != "POST" || !strings.Contains(r.URL.Path, "/v2/sites/") || !strings.Contains(r.URL.Path, "/assets") {
			t.Errorf("Expected POST to /v2/sites/{site_id}/assets, got %s %s", r.Method, r.URL.Path)
		}

		// Return successful response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(expectedResponse)
	}))
	defer server.Close()

	// Override the API base URL
	oldURL := postAssetUploadURLBaseURL
	postAssetUploadURLBaseURL = server.URL
	defer func() { postAssetUploadURLBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}

	// Setup: Create context
	ctx := context.Background()

	// Execute: Call Create()
	req := infer.CreateRequest[AssetArgs]{
		Inputs: AssetArgs{
			SiteID:   "5f0c8c9e1c9d440000e8d8c3",
			FileName: "logo.png",
			FileHash: "d41d8cd98f00b204e9800998ecf8427e",
		},
		DryRun: false,
	}

	resp, err := asset.Create(ctx, req)
	// Verify: No error
	if err != nil {
		t.Fatalf("Create() should not return error, got: %v", err)
	}

	// Verify: Resource ID is properly formatted
	expectedID := "5f0c8c9e1c9d440000e8d8c3/assets/6789abcdef1234567890abcd"
	if resp.ID != expectedID {
		t.Errorf("Expected ID=%s, got %s", expectedID, resp.ID)
	}

	// Verify: Output state contains API response data
	if resp.Output.AssetID != expectedResponse.ID {
		t.Errorf("Expected AssetID=%s, got %s", expectedResponse.ID, resp.Output.AssetID)
	}
	if resp.Output.HostedURL != expectedResponse.HostedURL {
		t.Errorf("Expected HostedURL=%s, got %s", expectedResponse.HostedURL, resp.Output.HostedURL)
	}
	if resp.Output.ContentType != expectedResponse.ContentType {
		t.Errorf("Expected ContentType=%s, got %s", expectedResponse.ContentType, resp.Output.ContentType)
	}
}

// TestAssetCreate_ValidationFailure tests validation errors in Create
func TestAssetCreate_ValidationFailure(t *testing.T) {
	asset := &Asset{}
	ctx := context.Background()

	tests := []struct {
		name   string
		inputs AssetArgs
		want   string
	}{
		{
			name: "invalid siteId",
			inputs: AssetArgs{
				SiteID:   "invalid",
				FileName: "logo.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			want: "validation failed",
		},
		{
			name: "missing fileName",
			inputs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			want: "fileName is required",
		},
		{
			name: "fileName too long",
			inputs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: strings.Repeat("a", 256) + ".png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			want: "too long",
		},
		{
			name: "fileName with invalid characters",
			inputs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo<>.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			want: "invalid character",
		},
		{
			name: "missing fileHash",
			inputs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "",
			},
			want: "fileHash is required",
		},
		{
			name: "invalid fileHash format",
			inputs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "invalid",
			},
			want: "invalid format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := infer.CreateRequest[AssetArgs]{
				Inputs: tt.inputs,
				DryRun: false,
			}

			_, err := asset.Create(ctx, req)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !strings.Contains(strings.ToLower(err.Error()), strings.ToLower(tt.want)) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.want, err.Error())
			}
		})
	}
}

// TestAssetCreate_APIError tests handling of API errors during creation
func TestAssetCreate_APIError(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that returns an error
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusBadRequest)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Invalid file hash",
		})
	}))
	defer server.Close()

	// Override the API base URL
	oldURL := postAssetUploadURLBaseURL
	postAssetUploadURLBaseURL = server.URL
	defer func() { postAssetUploadURLBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}
	ctx := context.Background()

	// Execute: Call Create()
	req := infer.CreateRequest[AssetArgs]{
		Inputs: AssetArgs{
			SiteID:   "5f0c8c9e1c9d440000e8d8c3",
			FileName: "logo.png",
			FileHash: "d41d8cd98f00b204e9800998ecf8427e",
		},
		DryRun: false,
	}

	_, err := asset.Create(ctx, req)

	// Verify: Error is returned
	if err == nil {
		t.Fatal("Expected error from API, got nil")
	}
	if !strings.Contains(err.Error(), "failed to create asset") {
		t.Errorf("Expected error message about creation failure, got: %v", err)
	}
}

// TestAssetCreate_DryRun tests preview mode (dry-run) for asset creation
func TestAssetCreate_DryRun(t *testing.T) {
	asset := &Asset{}
	ctx := context.Background()

	req := infer.CreateRequest[AssetArgs]{
		Inputs: AssetArgs{
			SiteID:   "5f0c8c9e1c9d440000e8d8c3",
			FileName: "logo.png",
			FileHash: "d41d8cd98f00b204e9800998ecf8427e",
		},
		DryRun: true,
	}

	// Execute: Call Create() in dry-run mode
	resp, err := asset.Create(ctx, req)
	// Verify: No error
	if err != nil {
		t.Fatalf("Create() dry-run failed: %v", err)
	}

	// Verify: Preview ID is generated
	if resp.ID == "" {
		t.Error("Expected non-empty ID in dry-run mode")
	}
	if !strings.Contains(resp.ID, "5f0c8c9e1c9d440000e8d8c3/assets/") {
		t.Errorf("Expected preview ID to contain site ID and assets path, got: %s", resp.ID)
	}

	// Verify: Preview state is populated
	if resp.Output.HostedURL == "" {
		t.Error("Expected preview HostedURL to be set")
	}
	if resp.Output.CreatedOn == "" {
		t.Error("Expected preview CreatedOn timestamp to be set")
	}
}

// =============================================================================
// Delete Operation Tests
// =============================================================================

// TestAssetDelete_Success tests successful asset deletion
func TestAssetDelete_Success(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that confirms deletion
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify correct endpoint and method
		if r.Method != "DELETE" || !strings.Contains(r.URL.Path, "/v2/assets/") {
			t.Errorf("Expected DELETE to /v2/assets/{asset_id}, got %s %s", r.Method, r.URL.Path)
		}

		// Return successful deletion
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Override the API base URL
	oldURL := deleteAssetBaseURL
	deleteAssetBaseURL = server.URL
	defer func() { deleteAssetBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}
	ctx := context.Background()

	// Execute: Call Delete()
	resourceID := "5f0c8c9e1c9d440000e8d8c3/assets/6789abcdef1234567890abcd"
	req := infer.DeleteRequest[AssetState]{
		ID: resourceID,
		State: AssetState{
			AssetArgs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			AssetID: "6789abcdef1234567890abcd",
		},
	}

	_, err := asset.Delete(ctx, req)
	// Verify: No error
	if err != nil {
		t.Fatalf("Delete() should not return error, got: %v", err)
	}
}

// TestAssetDelete_NotFound tests idempotent deletion (404 handling)
func TestAssetDelete_NotFound(t *testing.T) {
	// Setup: Set environment variable for authentication
	t.Setenv("WEBFLOW_API_TOKEN", "test-token-12345678901234567890")

	// Setup: Mock HTTP server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{
			"message": "Asset not found",
		})
	}))
	defer server.Close()

	// Override the API base URL
	oldURL := deleteAssetBaseURL
	deleteAssetBaseURL = server.URL
	defer func() { deleteAssetBaseURL = oldURL }()

	// Setup: Create the Asset resource controller
	asset := &Asset{}
	ctx := context.Background()

	// Execute: Call Delete()
	resourceID := "5f0c8c9e1c9d440000e8d8c3/assets/6789abcdef1234567890abcd"
	req := infer.DeleteRequest[AssetState]{
		ID: resourceID,
		State: AssetState{
			AssetArgs: AssetArgs{
				SiteID:   "5f0c8c9e1c9d440000e8d8c3",
				FileName: "logo.png",
				FileHash: "d41d8cd98f00b204e9800998ecf8427e",
			},
			AssetID: "6789abcdef1234567890abcd",
		},
	}

	_, err := asset.Delete(ctx, req)
	// Verify: No error (404 should be handled gracefully for idempotency)
	if err != nil {
		t.Fatalf("Delete() should handle 404 gracefully, got: %v", err)
	}
}

// =============================================================================
// Diff Detection Tests
// =============================================================================

// TestAssetDiff_NoChanges tests that Diff correctly reports no changes
// when inputs match state
func TestAssetDiff_NoChanges(t *testing.T) {
	asset := &Asset{}
	ctx := context.Background()

	args := AssetArgs{
		SiteID:       "5f0c8c9e1c9d440000e8d8c3",
		FileName:     "logo.png",
		FileHash:     "d41d8cd98f00b204e9800998ecf8427e",
		ParentFolder: "folder123",
		FileSource:   "https://example.com/logo.png",
	}

	req := infer.DiffRequest[AssetArgs, AssetState]{
		Inputs: args,
		State: AssetState{
			AssetArgs: args,
			AssetID:   "6789abcdef1234567890abcd",
		},
	}

	// Execute: Call Diff()
	resp, err := asset.Diff(ctx, req)
	// Verify: No error
	if err != nil {
		t.Fatalf("Diff() should not return error, got: %v", err)
	}

	// Verify: No changes detected
	if resp.HasChanges {
		t.Error("Diff() should not detect changes when inputs match state")
	}
}

// TestAssetDiff_RequiresReplacement tests that any property change
// requires replacement (assets are immutable)
func TestAssetDiff_RequiresReplacement(t *testing.T) {
	asset := &Asset{}
	ctx := context.Background()

	baseArgs := AssetArgs{
		SiteID:       "5f0c8c9e1c9d440000e8d8c3",
		FileName:     "logo.png",
		FileHash:     "d41d8cd98f00b204e9800998ecf8427e",
		ParentFolder: "folder123",
		FileSource:   "https://example.com/logo.png",
	}

	baseState := AssetState{
		AssetArgs: baseArgs,
		AssetID:   "6789abcdef1234567890abcd",
	}

	tests := []struct {
		name      string
		modifyFn  func(args *AssetArgs)
		fieldName string
	}{
		{
			name: "siteId change",
			modifyFn: func(args *AssetArgs) {
				args.SiteID = "6f1d9d0f2d0e551111f9e9d4"
			},
			fieldName: "siteId",
		},
		{
			name: "fileName change",
			modifyFn: func(args *AssetArgs) {
				args.FileName = "hero.jpg"
			},
			fieldName: "fileName",
		},
		{
			name: "fileHash change",
			modifyFn: func(args *AssetArgs) {
				args.FileHash = "e9800998ecf8427ed41d8cd98f00b204"
			},
			fieldName: "fileHash",
		},
		{
			name: "parentFolder change",
			modifyFn: func(args *AssetArgs) {
				args.ParentFolder = "folder456"
			},
			fieldName: "parentFolder",
		},
		{
			name: "fileSource change",
			modifyFn: func(args *AssetArgs) {
				args.FileSource = "https://example.com/hero.png"
			},
			fieldName: "fileSource",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create modified inputs
			modifiedArgs := baseArgs
			tt.modifyFn(&modifiedArgs)

			req := infer.DiffRequest[AssetArgs, AssetState]{
				Inputs: modifiedArgs,
				State:  baseState,
			}

			// Execute: Call Diff()
			resp, err := asset.Diff(ctx, req)
			// Verify: No error
			if err != nil {
				t.Fatalf("Diff() should not return error, got: %v", err)
			}

			// Verify: Changes detected
			if !resp.HasChanges {
				t.Errorf("Diff() should detect changes for %s", tt.fieldName)
			}

			// Verify: DeleteBeforeReplace is true (assets are immutable)
			if !resp.DeleteBeforeReplace {
				t.Errorf("Diff() DeleteBeforeReplace should be true for %s (assets are immutable)", tt.fieldName)
			}

			// Verify: Field is marked as UpdateReplace
			if diff, ok := resp.DetailedDiff[tt.fieldName]; ok {
				if diff.Kind != p.UpdateReplace {
					t.Errorf("Diff() %s should be UpdateReplace, got %v", tt.fieldName, diff.Kind)
				}
			} else {
				t.Errorf("Diff() DetailedDiff should contain %s", tt.fieldName)
			}
		})
	}
}
