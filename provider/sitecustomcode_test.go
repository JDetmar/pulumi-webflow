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

	"github.com/pulumi/pulumi-go-provider/infer"
)

func TestGenerateSiteCustomCodeResourceID(t *testing.T) {
	siteID := "5f0c8c9e1c9d440000e8d8c3"
	expected := "5f0c8c9e1c9d440000e8d8c3/custom_code"

	result := GenerateSiteCustomCodeResourceID(siteID)
	if result != expected {
		t.Errorf("GenerateSiteCustomCodeResourceID() = %v, want %v", result, expected)
	}
}

func TestExtractSiteIDFromSiteCustomCodeResourceID(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantID  string
		wantErr bool
	}{
		{"valid", "5f0c8c9e1c9d440000e8d8c3/custom_code", "5f0c8c9e1c9d440000e8d8c3", false},
		{"empty", "", "", true},
		{"invalid suffix", "5f0c8c9e1c9d440000e8d8c3/robots.txt", "", true},
		{"missing suffix", "5f0c8c9e1c9d440000e8d8c3", "", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			result, err := ExtractSiteIDFromSiteCustomCodeResourceID(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ExtractSiteIDFromSiteCustomCodeResourceID() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if result != tt.wantID {
				t.Errorf("ExtractSiteIDFromSiteCustomCodeResourceID() = %v, want %v", result, tt.wantID)
			}
		})
	}
}

func TestGetSiteCustomCode(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(SiteCustomCodeResponse{
			Scripts: []CustomScript{
				{
					ID:       "cms_slider",
					Location: "header",
					Version:  "1.0.0",
					Attributes: map[string]interface{}{
						"data-config": "my-value",
					},
				},
			},
			LastUpdated: "2025-01-03T00:00:00Z",
			CreatedOn:   "2025-01-03T00:00:00Z",
		})
	}))
	defer server.Close()

	// Override base URL for testing
	getSiteCustomCodeBaseURL = server.URL
	defer func() { getSiteCustomCodeBaseURL = "" }()

	// Test
	client := &http.Client{}
	resp, err := GetSiteCustomCode(context.Background(), client, "test-site-id")
	if err != nil {
		t.Fatalf("GetSiteCustomCode() error = %v", err)
	}

	if len(resp.Scripts) != 1 {
		t.Errorf("Expected 1 script, got %d", len(resp.Scripts))
	}

	if resp.Scripts[0].ID != "cms_slider" {
		t.Errorf("Expected ID 'cms_slider', got '%s'", resp.Scripts[0].ID)
	}

	if resp.Scripts[0].Location != "header" {
		t.Errorf("Expected location 'header', got '%s'", resp.Scripts[0].Location)
	}
}

func TestPutSiteCustomCode(t *testing.T) {
	// Create mock server
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PUT" {
			t.Errorf("Expected PUT, got %s", r.Method)
		}

		// Verify request body
		var req SiteCustomCodeRequest
		if err := json.NewDecoder(r.Body).Decode(&req); err != nil {
			t.Errorf("Failed to decode request: %v", err)
		}

		if len(req.Scripts) != 1 {
			t.Errorf("Expected 1 script in request, got %d", len(req.Scripts))
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		_ = json.NewEncoder(w).Encode(SiteCustomCodeResponse{
			Scripts: []CustomScript{
				{
					ID:       "cms_slider",
					Location: "header",
					Version:  "1.0.0",
				},
			},
			LastUpdated: "2025-01-03T00:00:00Z",
			CreatedOn:   "2025-01-03T00:00:00Z",
		})
	}))
	defer server.Close()

	// Override base URL for testing
	putSiteCustomCodeBaseURL = server.URL
	defer func() { putSiteCustomCodeBaseURL = "" }()

	// Test
	client := &http.Client{}
	scripts := []CustomScript{
		{
			ID:       "cms_slider",
			Location: "header",
			Version:  "1.0.0",
		},
	}

	resp, err := PutSiteCustomCode(context.Background(), client, "test-site-id", scripts)
	if err != nil {
		t.Fatalf("PutSiteCustomCode() error = %v", err)
	}

	if len(resp.Scripts) != 1 {
		t.Errorf("Expected 1 script in response, got %d", len(resp.Scripts))
	}
}

func TestDeleteSiteCustomCode(t *testing.T) {
	// Create mock server for successful delete
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Override base URL for testing
	deleteSiteCustomCodeBaseURL = server.URL
	defer func() { deleteSiteCustomCodeBaseURL = "" }()

	// Test
	client := &http.Client{}
	err := DeleteSiteCustomCode(context.Background(), client, "test-site-id")
	if err != nil {
		t.Fatalf("DeleteSiteCustomCode() error = %v", err)
	}
}

// TestSiteCustomCodeCreate_DryRun_WithUnknownScriptIDs verifies that preview succeeds
// when script IDs are unknown (empty strings from the infer framework).
// This simulates the case where scripts[].id references a RegisteredScript.scriptId
// output that hasn't been created yet.
func TestSiteCustomCodeCreate_DryRun_WithUnknownScriptIDs(t *testing.T) {
	resource := &SiteCustomCode{}
	ctx := context.Background()

	// Simulate Pulumi unknowns: empty strings for all script fields
	req := infer.CreateRequest[SiteCustomCodeArgs]{
		Inputs: SiteCustomCodeArgs{
			SiteID: "", // unknown siteId (e.g., from Site resource output)
			Scripts: []CustomScriptArgs{
				{
					ID:       "", // unknown scriptId from RegisteredScript
					Version:  "", // unknown version
					Location: "", // unknown location
				},
			},
		},
		DryRun: true,
	}

	resp, err := resource.Create(ctx, req)
	if err != nil {
		t.Fatalf("Create() DryRun with unknown inputs should succeed, got error: %v", err)
	}

	if resp.ID == "" {
		t.Error("Create() DryRun should return a non-empty resource ID")
	}

	if resp.Output.LastUpdated == "" {
		t.Error("Create() DryRun should set LastUpdated timestamp")
	}

	if resp.Output.CreatedOn == "" {
		t.Error("Create() DryRun should set CreatedOn timestamp")
	}
}

// TestSiteCustomCodeUpdate_DryRun_WithUnknownScriptIDs verifies that preview succeeds
// for updates when script IDs are unknown.
func TestSiteCustomCodeUpdate_DryRun_WithUnknownScriptIDs(t *testing.T) {
	resource := &SiteCustomCode{}
	ctx := context.Background()

	req := infer.UpdateRequest[SiteCustomCodeArgs, SiteCustomCodeState]{
		ID: "5f0c8c9e1c9d440000e8d8c3/custom_code",
		Inputs: SiteCustomCodeArgs{
			SiteID: "5f0c8c9e1c9d440000e8d8c3",
			Scripts: []CustomScriptArgs{
				{
					ID:       "", // unknown scriptId from RegisteredScript
					Version:  "", // unknown version
					Location: "", // unknown location
				},
			},
		},
		State: SiteCustomCodeState{
			SiteCustomCodeArgs: SiteCustomCodeArgs{
				SiteID: "5f0c8c9e1c9d440000e8d8c3",
				Scripts: []CustomScriptArgs{
					{
						ID:       "old_script",
						Version:  "1.0.0",
						Location: "header",
					},
				},
			},
		},
		DryRun: true,
	}

	resp, err := resource.Update(ctx, req)
	if err != nil {
		t.Fatalf("Update() DryRun with unknown inputs should succeed, got error: %v", err)
	}

	if resp.Output.LastUpdated == "" {
		t.Error("Update() DryRun should set LastUpdated timestamp")
	}
}

func TestDeleteSiteCustomCodeIdempotent(t *testing.T) {
	// Create mock server that returns 404 (resource already deleted)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNotFound)
	}))
	defer server.Close()

	// Override base URL for testing
	deleteSiteCustomCodeBaseURL = server.URL
	defer func() { deleteSiteCustomCodeBaseURL = "" }()

	// Test - should not error even with 404
	client := &http.Client{}
	err := DeleteSiteCustomCode(context.Background(), client, "test-site-id")
	if err != nil {
		t.Fatalf("DeleteSiteCustomCode() error = %v (expected no error for 404)", err)
	}
}
