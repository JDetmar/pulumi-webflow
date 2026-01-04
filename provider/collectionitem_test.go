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
)

// TestValidateFieldData tests the ValidateFieldData function.
func TestValidateFieldData(t *testing.T) {
	tests := []struct {
		name      string
		fieldData map[string]interface{}
		wantErr   bool
	}{
		{
			name:      "valid field data",
			fieldData: map[string]interface{}{"name": "Test Item", "slug": "test-item"},
			wantErr:   false,
		},
		{
			name:      "nil field data",
			fieldData: nil,
			wantErr:   true,
		},
		{
			name:      "empty field data",
			fieldData: map[string]interface{}{},
			wantErr:   true,
		},
		{
			name:      "field data with multiple fields",
			fieldData: map[string]interface{}{"name": "Test", "slug": "test", "content": "Content"},
			wantErr:   false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldData(tt.fieldData)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldData() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestGenerateCollectionItemResourceID tests the GenerateCollectionItemResourceID function.
func TestGenerateCollectionItemResourceID(t *testing.T) {
	tests := []struct {
		name         string
		collectionID string
		itemID       string
		want         string
	}{
		{
			name:         "valid IDs",
			collectionID: "5f0c8c9e1c9d440000e8d8c3",
			itemID:       "6f1d9d0f2d0e550111f9e9d4",
			want:         "5f0c8c9e1c9d440000e8d8c3/items/6f1d9d0f2d0e550111f9e9d4",
		},
		{
			name:         "itemID with slashes",
			collectionID: "5f0c8c9e1c9d440000e8d8c3",
			itemID:       "6f1d9d0f/special/item",
			want:         "5f0c8c9e1c9d440000e8d8c3/items/6f1d9d0f/special/item",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCollectionItemResourceID(tt.collectionID, tt.itemID)
			if got != tt.want {
				t.Errorf("GenerateCollectionItemResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExtractIDsFromCollectionItemResourceID tests the ExtractIDsFromCollectionItemResourceID function.
func TestExtractIDsFromCollectionItemResourceID(t *testing.T) {
	tests := []struct {
		name             string
		resourceID       string
		wantCollectionID string
		wantItemID       string
		wantErr          bool
	}{
		{
			name:             "valid resource ID",
			resourceID:       "5f0c8c9e1c9d440000e8d8c3/items/6f1d9d0f2d0e550111f9e9d4",
			wantCollectionID: "5f0c8c9e1c9d440000e8d8c3",
			wantItemID:       "6f1d9d0f2d0e550111f9e9d4",
			wantErr:          false,
		},
		{
			name:             "itemID with slashes",
			resourceID:       "5f0c8c9e1c9d440000e8d8c3/items/6f1d9d0f/special/item",
			wantCollectionID: "5f0c8c9e1c9d440000e8d8c3",
			wantItemID:       "6f1d9d0f/special/item",
			wantErr:          false,
		},
		{
			name:       "empty resource ID",
			resourceID: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - no items segment",
			resourceID: "5f0c8c9e1c9d440000e8d8c3/redirects/6f1d9d0f2d0e550111f9e9d4",
			wantErr:    true,
		},
		{
			name:       "invalid format - too few parts",
			resourceID: "5f0c8c9e1c9d440000e8d8c3/items",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			gotCollectionID, gotItemID, err := ExtractIDsFromCollectionItemResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractIDsFromCollectionItemResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !tt.wantErr {
				if gotCollectionID != tt.wantCollectionID {
					t.Errorf("ExtractIDsFromCollectionItemResourceID() collectionID = %v, want %v",
						gotCollectionID, tt.wantCollectionID)
				}
				if gotItemID != tt.wantItemID {
					t.Errorf("ExtractIDsFromCollectionItemResourceID() itemID = %v, want %v",
						gotItemID, tt.wantItemID)
				}
			}
		})
	}
}

// TestGetCollectionItems tests the GetCollectionItems function.
func TestGetCollectionItems(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   CollectionItemListResponse
		wantErr    bool
	}{
		{
			name:       "successful request",
			statusCode: 200,
			response: CollectionItemListResponse{
				Items: []CollectionItem{
					{
						ID:        "6f1d9d0f2d0e550111f9e9d4",
						FieldData: map[string]interface{}{"name": "Test Item", "slug": "test-item"},
						IsDraft:   true,
					},
				},
			},
			wantErr: false,
		},
		{
			name:       "empty list",
			statusCode: 200,
			response: CollectionItemListResponse{
				Items: []CollectionItem{},
			},
			wantErr: false,
		},
		{
			name:       "404 not found",
			statusCode: 404,
			response:   CollectionItemListResponse{},
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: 500,
			response:   CollectionItemListResponse{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "GET" {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				// Verify URL path
				if !strings.Contains(r.URL.Path, "/v2/collections/") || !strings.Contains(r.URL.Path, "/items") {
					t.Errorf("Unexpected URL path: %s", r.URL.Path)
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			// Override base URL for testing
			getCollectionItemsBaseURL = server.URL
			defer func() { getCollectionItemsBaseURL = "" }()

			// Test
			client := &http.Client{}
			resp, err := GetCollectionItems(context.Background(), client, "5f0c8c9e1c9d440000e8d8c3")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCollectionItems() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp != nil {
				if len(resp.Items) != len(tt.response.Items) {
					t.Errorf("GetCollectionItems() returned %d items, want %d",
						len(resp.Items), len(tt.response.Items))
				}
			}
		})
	}
}

// TestGetCollectionItem tests the GetCollectionItem function.
func TestGetCollectionItem(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   CollectionItem
		wantErr    bool
	}{
		{
			name:       "successful request",
			statusCode: 200,
			response: CollectionItem{
				ID:        "6f1d9d0f2d0e550111f9e9d4",
				FieldData: map[string]interface{}{"name": "Test Item", "slug": "test-item"},
				IsDraft:   true,
				CreatedOn: "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name:       "404 not found",
			statusCode: 404,
			response:   CollectionItem{},
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: 500,
			response:   CollectionItem{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "GET" {
					t.Errorf("Expected GET, got %s", r.Method)
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			// Override base URL for testing
			getCollectionItemBaseURL = server.URL
			defer func() { getCollectionItemBaseURL = "" }()

			// Test
			client := &http.Client{}
			resp, err := GetCollectionItem(context.Background(), client,
				"5f0c8c9e1c9d440000e8d8c3", "6f1d9d0f2d0e550111f9e9d4")
			if (err != nil) != tt.wantErr {
				t.Errorf("GetCollectionItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp != nil {
				if resp.ID != tt.response.ID {
					t.Errorf("GetCollectionItem() ID = %v, want %v", resp.ID, tt.response.ID)
				}
			}
		})
	}
}

// TestPostCollectionItem tests the PostCollectionItem function.
func TestPostCollectionItem(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   CollectionItem
		wantErr    bool
	}{
		{
			name:       "successful creation - 201",
			statusCode: 201,
			response: CollectionItem{
				ID:        "6f1d9d0f2d0e550111f9e9d4",
				FieldData: map[string]interface{}{"name": "New Item", "slug": "new-item"},
				IsDraft:   true,
				CreatedOn: "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name:       "successful creation - 200",
			statusCode: 200,
			response: CollectionItem{
				ID:        "6f1d9d0f2d0e550111f9e9d4",
				FieldData: map[string]interface{}{"name": "New Item", "slug": "new-item"},
				IsDraft:   true,
				CreatedOn: "2024-01-01T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name:       "400 bad request",
			statusCode: 400,
			response:   CollectionItem{},
			wantErr:    true,
		},
		{
			name:       "401 unauthorized",
			statusCode: 401,
			response:   CollectionItem{},
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: 500,
			response:   CollectionItem{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "POST" {
					t.Errorf("Expected POST, got %s", r.Method)
				}

				// Verify content type
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode == 200 || tt.statusCode == 201 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			// Override base URL for testing
			postCollectionItemBaseURL = server.URL
			defer func() { postCollectionItemBaseURL = "" }()

			// Test
			client := &http.Client{}
			fieldData := map[string]interface{}{"name": "New Item", "slug": "new-item"}
			isDraft := true
			resp, err := PostCollectionItem(context.Background(), client,
				"5f0c8c9e1c9d440000e8d8c3", fieldData, nil, &isDraft, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("PostCollectionItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp != nil {
				if resp.ID != tt.response.ID {
					t.Errorf("PostCollectionItem() ID = %v, want %v", resp.ID, tt.response.ID)
				}
			}
		})
	}
}

// TestPatchCollectionItem tests the PatchCollectionItem function.
func TestPatchCollectionItem(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		response   CollectionItem
		wantErr    bool
	}{
		{
			name:       "successful update",
			statusCode: 200,
			response: CollectionItem{
				ID:          "6f1d9d0f2d0e550111f9e9d4",
				FieldData:   map[string]interface{}{"name": "Updated Item", "slug": "updated-item"},
				IsDraft:     false,
				LastUpdated: "2024-01-02T00:00:00Z",
			},
			wantErr: false,
		},
		{
			name:       "404 not found",
			statusCode: 404,
			response:   CollectionItem{},
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: 500,
			response:   CollectionItem{},
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "PATCH" {
					t.Errorf("Expected PATCH, got %s", r.Method)
				}

				// Verify content type
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type: application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
				w.Header().Set("Content-Type", "application/json")
				if tt.statusCode == 200 {
					_ = json.NewEncoder(w).Encode(tt.response)
				}
			}))
			defer server.Close()

			// Override base URL for testing
			patchCollectionItemBaseURL = server.URL
			defer func() { patchCollectionItemBaseURL = "" }()

			// Test
			client := &http.Client{}
			fieldData := map[string]interface{}{"name": "Updated Item", "slug": "updated-item"}
			isDraft := false
			resp, err := PatchCollectionItem(context.Background(), client,
				"5f0c8c9e1c9d440000e8d8c3", "6f1d9d0f2d0e550111f9e9d4",
				fieldData, nil, &isDraft, "")
			if (err != nil) != tt.wantErr {
				t.Errorf("PatchCollectionItem() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp != nil {
				if resp.ID != tt.response.ID {
					t.Errorf("PatchCollectionItem() ID = %v, want %v", resp.ID, tt.response.ID)
				}
			}
		})
	}
}

// TestDeleteCollectionItem tests the DeleteCollectionItem function.
func TestDeleteCollectionItem(t *testing.T) {
	tests := []struct {
		name       string
		statusCode int
		wantErr    bool
	}{
		{
			name:       "successful deletion - 204",
			statusCode: 204,
			wantErr:    false,
		},
		{
			name:       "idempotent deletion - 404",
			statusCode: 404,
			wantErr:    false,
		},
		{
			name:       "401 unauthorized",
			statusCode: 401,
			wantErr:    true,
		},
		{
			name:       "500 server error",
			statusCode: 500,
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE, got %s", r.Method)
				}

				// Return mock response
				w.WriteHeader(tt.statusCode)
			}))
			defer server.Close()

			// Override base URL for testing
			deleteCollectionItemBaseURL = server.URL
			defer func() { deleteCollectionItemBaseURL = "" }()

			// Test
			client := &http.Client{}
			err := DeleteCollectionItem(context.Background(), client,
				"5f0c8c9e1c9d440000e8d8c3", "6f1d9d0f2d0e550111f9e9d4")
			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCollectionItem() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
