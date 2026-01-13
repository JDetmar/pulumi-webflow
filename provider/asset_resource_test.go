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
