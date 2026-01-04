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

// TestValidateFieldType tests the field type validation function.
func TestValidateFieldType(t *testing.T) {
	tests := []struct {
		name      string
		fieldType string
		wantErr   bool
	}{
		{"valid PlainText", FieldTypePlainText, false},
		{"valid RichText", FieldTypeRichText, false},
		{"valid Image", FieldTypeImage, false},
		{"valid Number", FieldTypeNumber, false},
		{"valid DateTime", FieldTypeDateTime, false},
		{"valid Reference", FieldTypeReference, false},
		{"empty", "", true},
		{"invalid lowercase", "plaintext", true},
		{"invalid unknown", "InvalidType", true},
		{"invalid mixed case", "plainText", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldType(tt.fieldType)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldType() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "type") {
				t.Errorf("ValidateFieldType() error should mention 'type': %v", err)
			}
		})
	}
}

// TestValidateFieldDisplayName tests the field displayName validation function.
func TestValidateFieldDisplayName(t *testing.T) {
	tests := []struct {
		name        string
		displayName string
		wantErr     bool
	}{
		{"valid short", "Title", false},
		{"valid long", strings.Repeat("a", 255), false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 256), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateFieldDisplayName(tt.displayName)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateFieldDisplayName() error = %v, wantErr %v", err, tt.wantErr)
			}
			if err != nil && !strings.Contains(err.Error(), "displayName") {
				t.Errorf("ValidateFieldDisplayName() error should mention 'displayName': %v", err)
			}
		})
	}
}

// TestGenerateCollectionFieldResourceID tests resource ID generation.
func TestGenerateCollectionFieldResourceID(t *testing.T) {
	tests := []struct {
		name         string
		collectionID string
		fieldID      string
		want         string
	}{
		{
			name:         "standard IDs",
			collectionID: "5f0c8c9e1c9d440000e8d8c3",
			fieldID:      "6789abcdef0123456789abcd",
			want:         "5f0c8c9e1c9d440000e8d8c3/fields/6789abcdef0123456789abcd",
		},
		{
			name:         "with slashes in fieldID",
			collectionID: "abc123",
			fieldID:      "field/with/slashes",
			want:         "abc123/fields/field/with/slashes",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			got := GenerateCollectionFieldResourceID(tt.collectionID, tt.fieldID)
			if got != tt.want {
				t.Errorf("GenerateCollectionFieldResourceID() = %v, want %v", got, tt.want)
			}
		})
	}
}

// TestExtractIDsFromCollectionFieldResourceID tests resource ID extraction.
func TestExtractIDsFromCollectionFieldResourceID(t *testing.T) {
	tests := []struct {
		name              string
		resourceID        string
		wantCollectionID  string
		wantFieldID       string
		wantErr           bool
	}{
		{
			name:              "valid resource ID",
			resourceID:        "5f0c8c9e1c9d440000e8d8c3/fields/6789abcdef0123456789abcd",
			wantCollectionID:  "5f0c8c9e1c9d440000e8d8c3",
			wantFieldID:       "6789abcdef0123456789abcd",
			wantErr:           false,
		},
		{
			name:              "fieldID with slashes",
			resourceID:        "abc123/fields/field/with/slashes",
			wantCollectionID:  "abc123",
			wantFieldID:       "field/with/slashes",
			wantErr:           false,
		},
		{
			name:              "empty resource ID",
			resourceID:        "",
			wantCollectionID:  "",
			wantFieldID:       "",
			wantErr:           true,
		},
		{
			name:              "invalid format - missing fields",
			resourceID:        "5f0c8c9e1c9d440000e8d8c3",
			wantCollectionID:  "",
			wantFieldID:       "",
			wantErr:           true,
		},
		{
			name:              "invalid format - wrong separator",
			resourceID:        "5f0c8c9e1c9d440000e8d8c3/wrongkey/6789abcdef0123456789abcd",
			wantCollectionID:  "",
			wantFieldID:       "",
			wantErr:           true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			collectionID, fieldID, err := ExtractIDsFromCollectionFieldResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractIDsFromCollectionFieldResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if collectionID != tt.wantCollectionID {
				t.Errorf("ExtractIDsFromCollectionFieldResourceID() collectionID = %v, want %v", collectionID, tt.wantCollectionID)
			}
			if fieldID != tt.wantFieldID {
				t.Errorf("ExtractIDsFromCollectionFieldResourceID() fieldID = %v, want %v", fieldID, tt.wantFieldID)
			}
		})
	}
}

// TestGetCollectionField tests the GetCollectionField API function.
func TestGetCollectionField(t *testing.T) {
	tests := []struct {
		name           string
		collectionID   string
		fieldID        string
		mockStatusCode int
		mockResponse   interface{}
		wantErr        bool
		wantFieldID    string
	}{
		{
			name:           "successful get",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 200,
			mockResponse: map[string]interface{}{
				"fields": []map[string]interface{}{
					{
						"id":          "field123",
						"type":        "PlainText",
						"displayName": "Title",
						"slug":        "title",
						"isEditable":  true,
						"isRequired":  true,
					},
				},
			},
			wantErr:     false,
			wantFieldID: "field123",
		},
		{
			name:           "field not found in collection",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "nonexistent",
			mockStatusCode: 200,
			mockResponse: map[string]interface{}{
				"fields": []map[string]interface{}{
					{
						"id":          "field123",
						"type":        "PlainText",
						"displayName": "Title",
					},
				},
			},
			wantErr: true,
		},
		{
			name:           "404 not found",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 404,
			mockResponse:   map[string]string{"error": "collection not found"},
			wantErr:        true,
		},
		{
			name:           "401 unauthorized",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 401,
			mockResponse:   map[string]string{"error": "unauthorized"},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method and path
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}
				expectedPath := "/v2/collections/" + tt.collectionID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return mock response
				w.WriteHeader(tt.mockStatusCode)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Override base URL for testing
			getCollectionFieldBaseURL = server.URL
			defer func() { getCollectionFieldBaseURL = "" }()

			client := &http.Client{}
			field, err := GetCollectionField(context.Background(), client, tt.collectionID, tt.fieldID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetCollectionField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && field != nil {
				if field.ID != tt.wantFieldID {
					t.Errorf("GetCollectionField() field.ID = %v, want %v", field.ID, tt.wantFieldID)
				}
			}
		})
	}
}

// TestPostCollectionField tests the PostCollectionField API function.
func TestPostCollectionField(t *testing.T) {
	tests := []struct {
		name           string
		collectionID   string
		fieldType      string
		displayName    string
		mockStatusCode int
		mockResponse   CollectionFieldResponse
		wantErr        bool
	}{
		{
			name:           "successful create",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldType:      "PlainText",
			displayName:    "Title",
			mockStatusCode: 201,
			mockResponse: CollectionFieldResponse{
				ID:          "field123",
				Type:        "PlainText",
				DisplayName: "Title",
				Slug:        "title",
				IsEditable:  true,
				IsRequired:  false,
			},
			wantErr: false,
		},
		{
			name:           "400 bad request",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldType:      "InvalidType",
			displayName:    "Test",
			mockStatusCode: 400,
			mockResponse:   CollectionFieldResponse{},
			wantErr:        true,
		},
		{
			name:           "403 forbidden",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldType:      "PlainText",
			displayName:    "Title",
			mockStatusCode: 403,
			mockResponse:   CollectionFieldResponse{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "POST" {
					t.Errorf("Expected POST request, got %s", r.Method)
				}
				expectedPath := "/v2/collections/" + tt.collectionID + "/fields"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify Content-Type
				if r.Header.Get("Content-Type") != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", r.Header.Get("Content-Type"))
				}

				// Return mock response
				w.WriteHeader(tt.mockStatusCode)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Override base URL for testing
			postCollectionFieldBaseURL = server.URL
			defer func() { postCollectionFieldBaseURL = "" }()

			client := &http.Client{}
			field, err := PostCollectionField(
				context.Background(), client, tt.collectionID,
				tt.fieldType, tt.displayName, "", "", false, nil,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("PostCollectionField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && field != nil {
				if field.ID != tt.mockResponse.ID {
					t.Errorf("PostCollectionField() field.ID = %v, want %v", field.ID, tt.mockResponse.ID)
				}
			}
		})
	}
}

// TestPutCollectionField tests the PutCollectionField API function.
func TestPutCollectionField(t *testing.T) {
	tests := []struct {
		name           string
		collectionID   string
		fieldID        string
		displayName    string
		mockStatusCode int
		mockResponse   CollectionFieldResponse
		wantErr        bool
	}{
		{
			name:           "successful update",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			displayName:    "Updated Title",
			mockStatusCode: 200,
			mockResponse: CollectionFieldResponse{
				ID:          "field123",
				Type:        "PlainText",
				DisplayName: "Updated Title",
				Slug:        "updated-title",
				IsEditable:  true,
				IsRequired:  true,
			},
			wantErr: false,
		},
		{
			name:           "404 not found",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "nonexistent",
			displayName:    "Test",
			mockStatusCode: 404,
			mockResponse:   CollectionFieldResponse{},
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}
				expectedPath := "/v2/collections/" + tt.collectionID + "/fields/" + tt.fieldID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return mock response
				w.WriteHeader(tt.mockStatusCode)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Override base URL for testing
			putCollectionFieldBaseURL = server.URL
			defer func() { putCollectionFieldBaseURL = "" }()

			client := &http.Client{}
			field, err := PutCollectionField(
				context.Background(), client, tt.collectionID, tt.fieldID,
				tt.displayName, "", "", false, nil,
			)

			if (err != nil) != tt.wantErr {
				t.Errorf("PutCollectionField() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && field != nil {
				if field.DisplayName != tt.mockResponse.DisplayName {
					t.Errorf("PutCollectionField() field.DisplayName = %v, want %v",
						field.DisplayName, tt.mockResponse.DisplayName)
				}
			}
		})
	}
}

// TestDeleteCollectionField tests the DeleteCollectionField API function.
func TestDeleteCollectionField(t *testing.T) {
	tests := []struct {
		name           string
		collectionID   string
		fieldID        string
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "successful delete - 204",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 204,
			wantErr:        false,
		},
		{
			name:           "successful delete - 404 (idempotent)",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "nonexistent",
			mockStatusCode: 404,
			wantErr:        false,
		},
		{
			name:           "403 forbidden",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 403,
			wantErr:        true,
		},
		{
			name:           "500 server error",
			collectionID:   "5f0c8c9e1c9d440000e8d8c3",
			fieldID:        "field123",
			mockStatusCode: 500,
			wantErr:        true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE request, got %s", r.Method)
				}
				expectedPath := "/v2/collections/" + tt.collectionID + "/fields/" + tt.fieldID
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Return mock response
				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			// Override base URL for testing
			deleteCollectionFieldBaseURL = server.URL
			defer func() { deleteCollectionFieldBaseURL = "" }()

			client := &http.Client{}
			err := DeleteCollectionField(context.Background(), client, tt.collectionID, tt.fieldID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeleteCollectionField() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestCollectionFieldRateLimiting tests rate limit handling with retry.
func TestCollectionFieldRateLimiting(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++
		if attemptCount < 2 {
			// First attempt returns 429
			w.WriteHeader(429)
			w.Header().Set("Retry-After", "1")
			return
		}
		// Second attempt succeeds
		w.WriteHeader(200)
		w.Header().Set("Content-Type", "application/json")
		json.NewEncoder(w).Encode(map[string]interface{}{
			"fields": []map[string]interface{}{
				{
					"id":          "field123",
					"type":        "PlainText",
					"displayName": "Title",
				},
			},
		})
	}))
	defer server.Close()

	getCollectionFieldBaseURL = server.URL
	defer func() { getCollectionFieldBaseURL = "" }()

	client := &http.Client{}
	field, err := GetCollectionField(context.Background(), client, "collection123", "field123")

	if err != nil {
		t.Errorf("GetCollectionField() should succeed after retry, got error: %v", err)
	}
	if field == nil || field.ID != "field123" {
		t.Errorf("GetCollectionField() should return field after retry")
	}
	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts (1 retry), got %d", attemptCount)
	}
}
