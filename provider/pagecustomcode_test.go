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

func TestValidatePageID(t *testing.T) {
	tests := []struct {
		name    string
		pageID  string
		wantErr bool
	}{
		{
			name:    "valid page ID",
			pageID:  "63c720f9347c2139b248e552",
			wantErr: false,
		},
		{
			name:    "empty page ID",
			pageID:  "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidatePageID(tt.pageID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidatePageID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScriptID(t *testing.T) {
	tests := []struct {
		name    string
		scriptID string
		wantErr bool
	}{
		{
			name:    "valid script ID",
			scriptID: "cms_slider",
			wantErr: false,
		},
		{
			name:    "empty script ID",
			scriptID: "",
			wantErr: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptID(tt.scriptID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScriptVersion(t *testing.T) {
	tests := []struct {
		name     string
		version  string
		wantErr  bool
	}{
		{
			name:     "valid version",
			version:  "1.0.0",
			wantErr:  false,
		},
		{
			name:     "another valid version",
			version:  "2.5.3",
			wantErr:  false,
		},
		{
			name:     "empty version",
			version:  "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptVersion(tt.version)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestValidateScriptLocation(t *testing.T) {
	tests := []struct {
		name     string
		location string
		wantErr  bool
	}{
		{
			name:     "valid location header",
			location: "header",
			wantErr:  false,
		},
		{
			name:     "valid location footer",
			location: "footer",
			wantErr:  false,
		},
		{
			name:     "invalid location",
			location: "body",
			wantErr:  true,
		},
		{
			name:     "empty location",
			location: "",
			wantErr:  true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptLocation(tt.location)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeneratePageCustomCodeResourceID(t *testing.T) {
	tests := []struct {
		name     string
		pageID   string
		expected string
	}{
		{
			name:     "standard page ID",
			pageID:   "5f0c8c9e1c9d440000e8d8c4",
			expected: "5f0c8c9e1c9d440000e8d8c4/custom-code",
		},
		{
			name:     "another page ID",
			pageID:   "abc123def456789012345678",
			expected: "abc123def456789012345678/custom-code",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePageCustomCodeResourceID(tt.pageID)
			if result != tt.expected {
				t.Errorf("GeneratePageCustomCodeResourceID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractPageIDFromPageCustomCodeResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		wantPageID string
		wantErr    bool
	}{
		{
			name:       "valid resource ID",
			resourceID: "5f0c8c9e1c9d440000e8d8c4/custom-code",
			wantPageID: "5f0c8c9e1c9d440000e8d8c4",
			wantErr:    false,
		},
		{
			name:       "empty resource ID",
			resourceID: "",
			wantPageID: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - missing suffix",
			resourceID: "5f0c8c9e1c9d440000e8d8c4",
			wantPageID: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - wrong suffix",
			resourceID: "5f0c8c9e1c9d440000e8d8c4/content",
			wantPageID: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - only suffix",
			resourceID: "/custom-code",
			wantPageID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, err := ExtractPageIDFromPageCustomCodeResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPageIDFromPageCustomCodeResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pageID != tt.wantPageID {
				t.Errorf("ExtractPageIDFromPageCustomCodeResourceID() pageID = %v, want %v", pageID, tt.wantPageID)
			}
		})
	}
}

func TestGetPageCustomCode(t *testing.T) {
	tests := []struct {
		name           string
		pageID         string
		mockStatusCode int
		mockResponse   PageCustomCodeResponse
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "successful GET",
			pageID:         "63c720f9347c2139b248e552",
			mockStatusCode: 200,
			mockResponse: PageCustomCodeResponse{
				Scripts: []CustomCodeScript{
					{
						ID:       "cms_slider",
						Version:  "1.0.0",
						Location: "header",
						Attributes: map[string]interface{}{
							"my-attribute": "some-value",
						},
					},
				},
				LastUpdated: "2022-10-26T00:28:54.191Z",
				CreatedOn:   "2022-10-26T00:28:54.191Z",
			},
			wantErr: false,
		},
		{
			name:           "empty scripts",
			pageID:         "63c720f9347c2139b248e552",
			mockStatusCode: 200,
			mockResponse: PageCustomCodeResponse{
				Scripts:     []CustomCodeScript{},
				LastUpdated: "2022-10-26T00:28:54.191Z",
				CreatedOn:   "2022-10-26T00:28:54.191Z",
			},
			wantErr: false,
		},
		{
			name:           "page not found",
			pageID:         "invalid-page-id",
			mockStatusCode: 404,
			mockResponse:   PageCustomCodeResponse{},
			wantErr:        true,
			errorContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "GET" {
					t.Errorf("Expected GET, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
				if tt.mockStatusCode == 200 {
					w.Header().Set("Content-Type", "application/json")
					json.NewEncoder(w).Encode(tt.mockResponse)
				}
			}))
			defer server.Close()

			// Override base URL for testing
			oldURL := getPageCustomCodeBaseURL
			getPageCustomCodeBaseURL = server.URL
			defer func() { getPageCustomCodeBaseURL = oldURL }()

			// Test
			client := &http.Client{}
			resp, err := GetPageCustomCode(context.Background(), client, tt.pageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("GetPageCustomCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Errorf("GetPageCustomCode() expected non-nil response")
			}
		})
	}
}

func TestPutPageCustomCode(t *testing.T) {
	tests := []struct {
		name           string
		pageID         string
		request        *PageCustomCodeRequest
		mockStatusCode int
		mockResponse   PageCustomCodeResponse
		wantErr        bool
	}{
		{
			name:   "successful PUT",
			pageID: "63c720f9347c2139b248e552",
			request: &PageCustomCodeRequest{
				Scripts: []CustomCodeScript{
					{
						ID:       "cms_slider",
						Version:  "1.0.0",
						Location: "header",
					},
				},
			},
			mockStatusCode: 200,
			mockResponse: PageCustomCodeResponse{
				Scripts: []CustomCodeScript{
					{
						ID:       "cms_slider",
						Version:  "1.0.0",
						Location: "header",
					},
				},
				LastUpdated: "2022-10-26T00:28:54.191Z",
				CreatedOn:   "2022-10-26T00:28:54.191Z",
			},
			wantErr: false,
		},
		{
			name:           "empty request",
			pageID:         "63c720f9347c2139b248e552",
			request:        &PageCustomCodeRequest{Scripts: []CustomCodeScript{}},
			mockStatusCode: 200,
			mockResponse: PageCustomCodeResponse{
				Scripts:     []CustomCodeScript{},
				LastUpdated: "2022-10-26T00:28:54.191Z",
				CreatedOn:   "2022-10-26T00:28:54.191Z",
			},
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "PUT" {
					t.Errorf("Expected PUT, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
				w.Header().Set("Content-Type", "application/json")
				json.NewEncoder(w).Encode(tt.mockResponse)
			}))
			defer server.Close()

			// Override base URL for testing
			oldURL := putPageCustomCodeBaseURL
			putPageCustomCodeBaseURL = server.URL
			defer func() { putPageCustomCodeBaseURL = oldURL }()

			// Test
			client := &http.Client{}
			resp, err := PutPageCustomCode(context.Background(), client, tt.pageID, tt.request)

			if (err != nil) != tt.wantErr {
				t.Errorf("PutPageCustomCode() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if !tt.wantErr && resp == nil {
				t.Errorf("PutPageCustomCode() expected non-nil response")
			}
		})
	}
}

func TestDeletePageCustomCode(t *testing.T) {
	tests := []struct {
		name           string
		pageID         string
		mockStatusCode int
		wantErr        bool
	}{
		{
			name:           "successful DELETE",
			pageID:         "63c720f9347c2139b248e552",
			mockStatusCode: 204,
			wantErr:        false,
		},
		{
			name:           "idempotent delete - 404",
			pageID:         "nonexistent-page",
			mockStatusCode: 404,
			wantErr:        false, // 404 is treated as success for delete
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				if r.Method != "DELETE" {
					t.Errorf("Expected DELETE, got %s", r.Method)
				}
				w.WriteHeader(tt.mockStatusCode)
			}))
			defer server.Close()

			// Override base URL for testing
			oldURL := deletePageCustomCodeBaseURL
			deletePageCustomCodeBaseURL = server.URL
			defer func() { deletePageCustomCodeBaseURL = oldURL }()

			// Test
			client := &http.Client{}
			err := DeletePageCustomCode(context.Background(), client, tt.pageID)

			if (err != nil) != tt.wantErr {
				t.Errorf("DeletePageCustomCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}
