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
	"testing"
)

// TestValidateAssetID tests the ValidateAssetID function.
func TestValidateAssetID(t *testing.T) {
	tests := []struct {
		name    string
		assetID string
		wantErr bool
	}{
		{
			name:    "valid asset ID",
			assetID: "5f0c8c9e1c9d440000e8d8c3",
			wantErr: false,
		},
		{
			name:    "empty asset ID",
			assetID: "",
			wantErr: true,
		},
		{
			name:    "invalid asset ID - too short",
			assetID: "5f0c8c9e1c9d44",
			wantErr: true,
		},
		{
			name:    "invalid asset ID - contains uppercase",
			assetID: "5F0C8C9E1C9D440000E8D8C3",
			wantErr: true,
		},
		{
			name:    "invalid asset ID - contains invalid characters",
			assetID: "5f0c8c9e1c9d440000e8d8g3",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateAssetID(tt.assetID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateAssetID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateFileHash tests the ValidateFileHash function.
func TestValidateFileHash(t *testing.T) {
	tests := []struct {
		name     string
		fileHash string
		wantErr  bool
	}{
		{
			name:     "valid MD5 hash - lowercase",
			fileHash: "d41d8cd98f00b204e9800998ecf8427e",
			wantErr:  false,
		},
		{
			name:     "valid MD5 hash - uppercase",
			fileHash: "D41D8CD98F00B204E9800998ECF8427E",
			wantErr:  false,
		},
		{
			name:     "valid MD5 hash - mixed case",
			fileHash: "d41D8cd98F00b204E9800998ecf8427E",
			wantErr:  false,
		},
		{
			name:     "empty hash",
			fileHash: "",
			wantErr:  true,
		},
		{
			name:     "too short",
			fileHash: "d41d8cd98f00b204",
			wantErr:  true,
		},
		{
			name:     "too long",
			fileHash: "d41d8cd98f00b204e9800998ecf8427e00",
			wantErr:  true,
		},
		{
			name:     "invalid characters",
			fileHash: "d41d8cd98f00b204e9800998ecf8427g",
			wantErr:  true,
		},
		{
			name:     "with spaces",
			fileHash: "d41d8cd98f00b204 9800998ecf8427e",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileHash(tt.fileHash)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateFileName tests the ValidateFileName function.
func TestValidateFileName(t *testing.T) {
	tests := []struct {
		name     string
		fileName string
		wantErr  bool
	}{
		{
			name:     "valid file name with extension",
			fileName: "logo.png",
			wantErr:  false,
		},
		{
			name:     "valid file name with multiple dots",
			fileName: "my.hero.image.jpg",
			wantErr:  false,
		},
		{
			name:     "valid file name with spaces",
			fileName: "my logo image.png",
			wantErr:  false,
		},
		{
			name:     "valid file name with hyphens and underscores",
			fileName: "my-hero_image-2024.png",
			wantErr:  false,
		},
		{
			name:     "empty file name",
			fileName: "",
			wantErr:  true,
		},
		{
			name:     "file name too long",
			fileName: string(make([]byte, 256)) + ".png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character <",
			fileName: "my<file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character >",
			fileName: "my>file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character :",
			fileName: "my:file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character \"",
			fileName: "my\"file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character |",
			fileName: "my|file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character ?",
			fileName: "my?file.png",
			wantErr:  true,
		},
		{
			name:     "file name with invalid character *",
			fileName: "my*file.png",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFileName(tt.fileName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFileName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateAssetResourceID tests the GenerateAssetResourceID function.
func TestGenerateAssetResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	assetID := "5f0c8c9e1c9d440000e8d8c4"

	expected := "5f0c8c9e1c9d440000e8d8c3/assets/5f0c8c9e1c9d440000e8d8c4"
	result := GenerateAssetResourceID(siteID, assetID)

	if result != expected {
		t.Errorf("GenerateAssetResourceID() = %v, want %v", result, expected)
	}
}

// TestExtractIDsFromAssetResourceID tests the ExtractIDsFromAssetResourceID function.
func TestExtractIDsFromAssetResourceID(t *testing.T) {
	tests := []struct {
		name        string
		resourceID  string
		wantSiteID  string
		wantAssetID string
		wantErr     bool
	}{
		{
			name:        "valid resource ID",
			resourceID:  "5f0c8c9e1c9d440000e8d8c3/assets/5f0c8c9e1c9d440000e8d8c4",
			wantSiteID:  "5f0c8c9e1c9d440000e8d8c3",
			wantAssetID: "5f0c8c9e1c9d440000e8d8c4",
			wantErr:     false,
		},
		{
			name:        "empty resource ID",
			resourceID:  "",
			wantSiteID:  "",
			wantAssetID: "",
			wantErr:     true,
		},
		{
			name:        "invalid format - missing assets segment",
			resourceID:  "5f0c8c9e1c9d440000e8d8c3/5f0c8c9e1c9d440000e8d8c4",
			wantSiteID:  "",
			wantAssetID: "",
			wantErr:     true,
		},
		{
			name:        "invalid format - wrong segment name",
			resourceID:  "5f0c8c9e1c9d440000e8d8c3/images/5f0c8c9e1c9d440000e8d8c4",
			wantSiteID:  "",
			wantAssetID: "",
			wantErr:     true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			siteID, assetID, err := ExtractIDsFromAssetResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractIDsFromAssetResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if siteID != tt.wantSiteID {
				t.Errorf("ExtractIDsFromAssetResourceID() siteID = %v, want %v", siteID, tt.wantSiteID)
			}
			if assetID != tt.wantAssetID {
				t.Errorf("ExtractIDsFromAssetResourceID() assetID = %v, want %v", assetID, tt.wantAssetID)
			}
		})
	}
}

// TestGetAsset tests the GetAsset function with a mock server.
func TestGetAsset(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Verify path
		expectedPath := "/v2/assets/5f0c8c9e1c9d440000e8d8c4"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetResponse{
			ID:               "5f0c8c9e1c9d440000e8d8c4",
			ContentType:      "image/png",
			Size:             12345,
			SiteID:           "5f0c8c9e1c9d440000e8d8c3",
			HostedURL:        "https://assets.website-files.com/example/logo.png",
			OriginalFileName: "logo.png",
			DisplayName:      "Logo",
			CreatedOn:        "2024-01-01T00:00:00Z",
			LastUpdated:      "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override base URL for testing
	getAssetBaseURL = server.URL
	defer func() { getAssetBaseURL = "" }()

	// Create HTTP client
	client := &http.Client{}

	// Test GetAsset
	asset, err := GetAsset(context.Background(), client, "5f0c8c9e1c9d440000e8d8c4")
	if err != nil {
		t.Fatalf("GetAsset() error = %v", err)
	}

	// Verify response
	if asset.ID != "5f0c8c9e1c9d440000e8d8c4" {
		t.Errorf("Expected asset ID 5f0c8c9e1c9d440000e8d8c4, got %s", asset.ID)
	}
	if asset.ContentType != "image/png" {
		t.Errorf("Expected content type image/png, got %s", asset.ContentType)
	}
	if asset.Size != 12345 {
		t.Errorf("Expected size 12345, got %d", asset.Size)
	}
	if asset.HostedURL != "https://assets.website-files.com/example/logo.png" {
		t.Errorf("Expected hosted URL https://assets.website-files.com/example/logo.png, got %s", asset.HostedURL)
	}
}

// TestListAssets tests the ListAssets function with a mock server.
func TestListAssets(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "GET" {
			t.Errorf("Expected GET request, got %s", r.Method)
		}

		// Verify path
		expectedPath := "/v2/sites/5f0c8c9e1c9d440000e8d8c3/assets"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Return mock response
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetListResponse{
			Assets: []AssetResponse{
				{
					ID:               "5f0c8c9e1c9d440000e8d8c4",
					ContentType:      "image/png",
					Size:             12345,
					SiteID:           "5f0c8c9e1c9d440000e8d8c3",
					HostedURL:        "https://assets.website-files.com/example/logo.png",
					OriginalFileName: "logo.png",
					CreatedOn:        "2024-01-01T00:00:00Z",
					LastUpdated:      "2024-01-01T00:00:00Z",
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override base URL for testing
	listAssetsBaseURL = server.URL
	defer func() { listAssetsBaseURL = "" }()

	// Create HTTP client
	client := &http.Client{}

	// Test ListAssets
	assets, err := ListAssets(context.Background(), client, "5f0c8c9e1c9d440000e8d8c3")
	if err != nil {
		t.Fatalf("ListAssets() error = %v", err)
	}

	// Verify response
	if len(assets.Assets) != 1 {
		t.Errorf("Expected 1 asset, got %d", len(assets.Assets))
	}
	if assets.Assets[0].ID != "5f0c8c9e1c9d440000e8d8c4" {
		t.Errorf("Expected asset ID 5f0c8c9e1c9d440000e8d8c4, got %s", assets.Assets[0].ID)
	}
}

// TestPostAssetUploadURL tests the PostAssetUploadURL function with a mock server.
func TestPostAssetUploadURL(t *testing.T) {
	// Create a mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		// Verify request method
		if r.Method != "POST" {
			t.Errorf("Expected POST request, got %s", r.Method)
		}

		// Verify path
		expectedPath := "/v2/sites/5f0c8c9e1c9d440000e8d8c3/assets"
		if r.URL.Path != expectedPath {
			t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
		}

		// Verify Content-Type header
		if r.Header.Get("Content-Type") != "application/json" {
			t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
		}

		// Parse request body
		var req AssetUploadRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		// Verify request body
		if req.FileName != "logo.png" {
			t.Errorf("Expected fileName logo.png, got %s", req.FileName)
		}
		if req.FileHash != "d41d8cd98f00b204e9800998ecf8427e" {
			t.Errorf("Expected fileHash d41d8cd98f00b204e9800998ecf8427e, got %s", req.FileHash)
		}

		// Return mock response matching actual Webflow API
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := AssetUploadResponse{
			ID:        "5f0c8c9e1c9d440000e8d8c4",
			UploadURL: "https://s3.amazonaws.com/bucket/upload?signature=xyz",
			UploadDetails: map[string]string{
				"acl":                   "public-read",
				"bucket":                "webflow-bucket",
				"key":                   "assets/logo.png",
				"Content-Type":          "image/png",
				"X-Amz-Algorithm":       "AWS4-HMAC-SHA256",
				"X-Amz-Credential":      "AKIAEXAMPLE/20240101/us-east-1/s3/aws4_request",
				"X-Amz-Date":            "20240101T000000Z",
				"Policy":                "base64policy",
				"X-Amz-Signature":       "signature123",
				"success_action_status": "201",
			},
			AssetURL:         "https://s3.amazonaws.com/webflow-bucket/assets/logo.png",
			HostedURL:        "https://assets.website-files.com/example/logo.png",
			ContentType:      "image/png",
			OriginalFileName: "logo.png",
			CreatedOn:        "2024-01-01T00:00:00Z",
			LastUpdated:      "2024-01-01T00:00:00Z",
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override base URL for testing
	postAssetUploadURLBaseURL = server.URL
	defer func() { postAssetUploadURLBaseURL = "" }()

	// Create HTTP client
	client := &http.Client{}

	// Test PostAssetUploadURL
	uploadResp, err := PostAssetUploadURL(
		context.Background(), client,
		"5f0c8c9e1c9d440000e8d8c3", "logo.png", "d41d8cd98f00b204e9800998ecf8427e", "",
	)
	if err != nil {
		t.Fatalf("PostAssetUploadURL() error = %v", err)
	}

	// Verify response - asset ID
	if uploadResp.ID != "5f0c8c9e1c9d440000e8d8c4" {
		t.Errorf("Expected asset ID 5f0c8c9e1c9d440000e8d8c4, got %s", uploadResp.ID)
	}

	// Verify response - upload URL
	if uploadResp.UploadURL != "https://s3.amazonaws.com/bucket/upload?signature=xyz" {
		t.Errorf("Expected upload URL https://s3.amazonaws.com/bucket/upload?signature=xyz, got %s", uploadResp.UploadURL)
	}

	// Verify response - hosted URL
	if uploadResp.HostedURL != "https://assets.website-files.com/example/logo.png" {
		t.Errorf("Expected hosted URL https://assets.website-files.com/example/logo.png, got %s", uploadResp.HostedURL)
	}

	// Verify response - content type
	if uploadResp.ContentType != "image/png" {
		t.Errorf("Expected content type image/png, got %s", uploadResp.ContentType)
	}

	// Verify response - upload details contain expected fields
	if uploadResp.UploadDetails["acl"] != "public-read" {
		t.Errorf("Expected acl public-read, got %s", uploadResp.UploadDetails["acl"])
	}
}

// TestDeleteAsset tests the DeleteAsset function with a mock server.
func TestDeleteAsset(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful delete - 204",
			statusCode: http.StatusNoContent,
			wantErr:    false,
		},
		{
			name:       "idempotent delete - 404",
			statusCode: http.StatusNotFound,
			wantErr:    false,
		},
		{
			name:       "error - 500",
			statusCode: http.StatusInternalServerError,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create a mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}

				// Verify path
				expectedPath := "/v2/assets/5f0c8c9e1c9d440000e8d8c4"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Override base URL for testing
			deleteAssetBaseURL = server.URL
			defer func() { deleteAssetBaseURL = "" }()

			// Create HTTP client
			client := &http.Client{}

			// Test DeleteAsset
			err := DeleteAsset(context.Background(), client, "5f0c8c9e1c9d440000e8d8c4")
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteAsset() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGetAssetNotFound tests GetAsset handling of 404 responses.
func TestGetAssetNotFound(t *testing.T) {
	// Create a mock server that returns 404
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_, _ = w.Write([]byte(`{"message": "Asset not found"}`))
	}))
	defer server.Close()

	// Override base URL for testing
	getAssetBaseURL = server.URL
	defer func() { getAssetBaseURL = "" }()

	// Create HTTP client
	client := &http.Client{}

	// Test GetAsset with non-existent asset
	_, err := GetAsset(context.Background(), client, "nonexistent")
	if err == nil {
		t.Error("Expected error for non-existent asset, got nil")
	}
}

// TestGetAssetRateLimited tests GetAsset handling of 429 rate limit responses.
func TestGetAssetRateLimited(t *testing.T) {
	attemptCount := 0

	// Create a mock server that returns 429 twice, then 200
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 3 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			_, _ = w.Write([]byte(`{"message": "Rate limit exceeded"}`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(http.StatusOK)
			response := AssetResponse{
				ID:          "5f0c8c9e1c9d440000e8d8c4",
				ContentType: "image/png",
				Size:        12345,
			}
			_ = json.NewEncoder(w).Encode(response)
		}
	}))
	defer server.Close()

	// Override base URL for testing
	getAssetBaseURL = server.URL
	defer func() { getAssetBaseURL = "" }()

	// Create HTTP client
	client := &http.Client{}

	// Test GetAsset - should retry and succeed
	asset, err := GetAsset(context.Background(), client, "5f0c8c9e1c9d440000e8d8c4")
	if err != nil {
		t.Fatalf("GetAsset() should succeed after retries, got error: %v", err)
	}

	if asset.ID != "5f0c8c9e1c9d440000e8d8c4" {
		t.Errorf("Expected asset ID 5f0c8c9e1c9d440000e8d8c4, got %s", asset.ID)
	}

	if attemptCount != 3 {
		t.Errorf("Expected 3 attempts, got %d", attemptCount)
	}
}
