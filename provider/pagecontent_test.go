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

func TestValidateNodeID(t *testing.T) {
	tests := []struct {
		name    string
		nodeID  string
		wantErr bool
	}{
		{
			name:    "valid node ID",
			nodeID:  "node-12345",
			wantErr: false,
		},
		{
			name:    "empty node ID",
			nodeID:  "",
			wantErr: true,
		},
		{
			name:    "UUID-style node ID",
			nodeID:  "550e8400-e29b-41d4-a716-446655440000",
			wantErr: false,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateNodeID(tt.nodeID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateNodeID() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

func TestGeneratePageContentResourceID(t *testing.T) {
	tests := []struct {
		name     string
		pageID   string
		expected string
	}{
		{
			name:     "standard page ID",
			pageID:   "5f0c8c9e1c9d440000e8d8c4",
			expected: "5f0c8c9e1c9d440000e8d8c4/content",
		},
		{
			name:     "another page ID",
			pageID:   "abc123def456789012345678",
			expected: "abc123def456789012345678/content",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result := GeneratePageContentResourceID(tt.pageID)
			if result != tt.expected {
				t.Errorf("GeneratePageContentResourceID() = %v, want %v", result, tt.expected)
			}
		})
	}
}

func TestExtractPageIDFromPageContentResourceID(t *testing.T) {
	tests := []struct {
		name       string
		resourceID string
		wantPageID string
		wantErr    bool
	}{
		{
			name:       "valid resource ID",
			resourceID: "5f0c8c9e1c9d440000e8d8c4/content",
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
			resourceID: "5f0c8c9e1c9d440000e8d8c4/nodes",
			wantPageID: "",
			wantErr:    true,
		},
		{
			name:       "invalid format - only suffix",
			resourceID: "/content",
			wantPageID: "",
			wantErr:    true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			pageID, err := ExtractPageIDFromPageContentResourceID(tt.resourceID)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractPageIDFromPageContentResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if pageID != tt.wantPageID {
				t.Errorf("ExtractPageIDFromPageContentResourceID() pageID = %v, want %v", pageID, tt.wantPageID)
			}
		})
	}
}

func TestGetPageContent(t *testing.T) {
	tests := []struct {
		name           string
		pageID         string
		mockStatusCode int
		mockResponse   PageContentResponse
		wantErr        bool
		errorContains  string
	}{
		{
			name:           "successful GET",
			pageID:         "5f0c8c9e1c9d440000e8d8c4",
			mockStatusCode: 200,
			mockResponse: PageContentResponse{
				PageID: "5f0c8c9e1c9d440000e8d8c4",
				Nodes: []DOMNode{
					{
						NodeID: "node-1",
						Type:   "text",
						Text:   "Hello World",
					},
				},
			},
			wantErr: false,
		},
		{
			name:           "404 not found",
			pageID:         "nonexistent",
			mockStatusCode: 404,
			mockResponse:   PageContentResponse{},
			wantErr:        true,
			errorContains:  "not found",
		},
		{
			name:           "401 unauthorized",
			pageID:         "5f0c8c9e1c9d440000e8d8c4",
			mockStatusCode: 401,
			mockResponse:   PageContentResponse{},
			wantErr:        true,
			errorContains:  "unauthorized",
		},
		{
			name:           "500 server error",
			pageID:         "5f0c8c9e1c9d440000e8d8c4",
			mockStatusCode: 500,
			mockResponse:   PageContentResponse{},
			wantErr:        true,
			errorContains:  "server error",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "GET" {
					t.Errorf("Expected GET request, got %s", r.Method)
				}

				// Verify URL path
				expectedPath := "/v2/pages/" + tt.pageID + "/dom"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Set response status
				w.WriteHeader(tt.mockStatusCode)

				// Write response body
				if tt.mockStatusCode == 200 {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					_, _ = w.Write([]byte(`{"message": "error"}`))
				}
			}))
			defer server.Close()

			// Override base URL for testing
			getPageContentBaseURL = server.URL
			defer func() { getPageContentBaseURL = "" }()

			// Create HTTP client
			client := &http.Client{}

			// Call function
			response, err := GetPageContent(context.Background(), client, tt.pageID)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("GetPageContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("GetPageContent() error = %v, should contain %q", err, tt.errorContains)
				}
				return
			}

			// Verify response
			if response == nil {
				t.Fatal("GetPageContent() returned nil response")
			}
			if response.PageID != tt.mockResponse.PageID {
				t.Errorf("GetPageContent() PageID = %v, want %v", response.PageID, tt.mockResponse.PageID)
			}
			if len(response.Nodes) != len(tt.mockResponse.Nodes) {
				t.Errorf("GetPageContent() Nodes length = %v, want %v", len(response.Nodes), len(tt.mockResponse.Nodes))
			}
		})
	}
}

func TestPutPageContent(t *testing.T) {
	tests := []struct {
		name           string
		pageID         string
		nodes          []DOMNodeUpdate
		mockStatusCode int
		mockResponse   PageContentResponse
		wantErr        bool
		errorContains  string
	}{
		{
			name:   "successful PUT",
			pageID: "5f0c8c9e1c9d440000e8d8c4",
			nodes: []DOMNodeUpdate{
				{
					NodeID: "node-1",
					Text:   stringPtr("Updated text"),
				},
			},
			mockStatusCode: 200,
			mockResponse: PageContentResponse{
				PageID: "5f0c8c9e1c9d440000e8d8c4",
				Nodes: []DOMNode{
					{
						NodeID: "node-1",
						Type:   "text",
						Text:   "Updated text",
					},
				},
			},
			wantErr: false,
		},
		{
			name:   "multiple node updates",
			pageID: "5f0c8c9e1c9d440000e8d8c4",
			nodes: []DOMNodeUpdate{
				{
					NodeID: "node-1",
					Text:   stringPtr("Text 1"),
				},
				{
					NodeID: "node-2",
					Text:   stringPtr("Text 2"),
				},
			},
			mockStatusCode: 200,
			mockResponse: PageContentResponse{
				PageID: "5f0c8c9e1c9d440000e8d8c4",
				Nodes:  []DOMNode{},
			},
			wantErr: false,
		},
		{
			name:   "400 bad request",
			pageID: "5f0c8c9e1c9d440000e8d8c4",
			nodes: []DOMNodeUpdate{
				{
					NodeID: "invalid-node",
					Text:   stringPtr("Text"),
				},
			},
			mockStatusCode: 400,
			mockResponse:   PageContentResponse{},
			wantErr:        true,
			errorContains:  "bad request",
		},
		{
			name:           "404 not found",
			pageID:         "nonexistent",
			nodes:          []DOMNodeUpdate{},
			mockStatusCode: 404,
			mockResponse:   PageContentResponse{},
			wantErr:        true,
			errorContains:  "not found",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create mock HTTP server
			server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
				// Verify request method
				if r.Method != "PUT" {
					t.Errorf("Expected PUT request, got %s", r.Method)
				}

				// Verify URL path
				expectedPath := "/v2/pages/" + tt.pageID + "/dom"
				if r.URL.Path != expectedPath {
					t.Errorf("Expected path %s, got %s", expectedPath, r.URL.Path)
				}

				// Verify Content-Type header
				contentType := r.Header.Get("Content-Type")
				if contentType != "application/json" {
					t.Errorf("Expected Content-Type application/json, got %s", contentType)
				}

				// Set response status
				w.WriteHeader(tt.mockStatusCode)

				// Write response body
				if tt.mockStatusCode == 200 {
					w.Header().Set("Content-Type", "application/json")
					_ = json.NewEncoder(w).Encode(tt.mockResponse)
				} else {
					_, _ = w.Write([]byte(`{"message": "error"}`))
				}
			}))
			defer server.Close()

			// Override base URL for testing
			putPageContentBaseURL = server.URL
			defer func() { putPageContentBaseURL = "" }()

			// Create HTTP client
			client := &http.Client{}

			// Call function
			response, err := PutPageContent(context.Background(), client, tt.pageID, tt.nodes)

			// Check error
			if (err != nil) != tt.wantErr {
				t.Errorf("PutPageContent() error = %v, wantErr %v", err, tt.wantErr)
				return
			}

			if tt.wantErr {
				if tt.errorContains != "" && !strings.Contains(err.Error(), tt.errorContains) {
					t.Errorf("PutPageContent() error = %v, should contain %q", err, tt.errorContains)
				}
				return
			}

			// Verify response
			if response == nil {
				t.Fatal("PutPageContent() returned nil response")
			}
			if response.PageID != tt.mockResponse.PageID {
				t.Errorf("PutPageContent() PageID = %v, want %v", response.PageID, tt.mockResponse.PageID)
			}
		})
	}
}

// TestGetPageContent_RateLimited_429 tests rate limiting with automatic retry
func TestGetPageContent_RateLimited_429(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// First attempt returns 429, second attempt succeeds
		if attemptCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message": "rate limited"}`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(PageContentResponse{
				PageID: "test-page-id",
				Nodes:  []DOMNode{},
			})
		}
	}))
	defer server.Close()

	// Override base URL for testing
	getPageContentBaseURL = server.URL
	defer func() { getPageContentBaseURL = "" }()

	client := &http.Client{}
	response, err := GetPageContent(context.Background(), client, "test-page-id")

	if err != nil {
		t.Errorf("GetPageContent() should succeed after retry, got error: %v", err)
	}
	if response == nil {
		t.Fatal("GetPageContent() returned nil response")
	}
	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts (1 retry), got %d", attemptCount)
	}
}

// TestPutPageContent_RateLimited_429 tests rate limiting with automatic retry
func TestPutPageContent_RateLimited_429(t *testing.T) {
	attemptCount := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attemptCount++

		// First attempt returns 429, second attempt succeeds
		if attemptCount == 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(429)
			_, _ = w.Write([]byte(`{"message": "rate limited"}`))
		} else {
			w.Header().Set("Content-Type", "application/json")
			w.WriteHeader(200)
			_ = json.NewEncoder(w).Encode(PageContentResponse{
				PageID: "test-page-id",
				Nodes:  []DOMNode{},
			})
		}
	}))
	defer server.Close()

	// Override base URL for testing
	putPageContentBaseURL = server.URL
	defer func() { putPageContentBaseURL = "" }()

	client := &http.Client{}
	nodes := []DOMNodeUpdate{{NodeID: "node-1", Text: stringPtr("test")}}
	response, err := PutPageContent(context.Background(), client, "test-page-id", nodes)

	if err != nil {
		t.Errorf("PutPageContent() should succeed after retry, got error: %v", err)
	}
	if response == nil {
		t.Fatal("PutPageContent() returned nil response")
	}
	if attemptCount != 2 {
		t.Errorf("Expected 2 attempts (1 retry), got %d", attemptCount)
	}
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}
