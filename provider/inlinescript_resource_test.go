// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

const testInlineScriptID = "test-inline-script-123"

// TestInlineScriptCreate_ValidationErrors tests input validation in Create
func TestInlineScriptCreate_ValidationErrors(t *testing.T) {
	resource := &InlineScript{}

	tests := []struct {
		name   string
		inputs InlineScriptArgs
		want   string
	}{
		{
			name: "invalid siteId",
			inputs: InlineScriptArgs{
				SiteID:      "invalid", // Too short
				SourceCode:  "console.log('hello');",
				Version:     "1.0.0",
				DisplayName: "TestScript",
			},
			want: "validation failed",
		},
		{
			name: "missing sourceCode",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "",
				Version:     "1.0.0",
				DisplayName: "TestScript",
			},
			want: "sourceCode is required",
		},
		{
			name: "sourceCode too long",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  strings.Repeat("a", 2001),
				Version:     "1.0.0",
				DisplayName: "TestScript",
			},
			want: "too long",
		},
		{
			name: "missing version",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "console.log('hello');",
				Version:     "",
				DisplayName: "TestScript",
			},
			want: "version is required",
		},
		{
			name: "invalid version format",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "console.log('hello');",
				Version:     "1",
				DisplayName: "TestScript",
			},
			want: "Semantic Version format",
		},
		{
			name: "missing displayName",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "console.log('hello');",
				Version:     "1.0.0",
				DisplayName: "",
			},
			want: "displayName is required",
		},
		{
			name: "displayName too long",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "console.log('hello');",
				Version:     "1.0.0",
				DisplayName: "ThisIsAVeryLongNameThatExceedsTheMaximumLengthOfFiftyCharacters",
			},
			want: "too long",
		},
		{
			name: "displayName with special chars",
			inputs: InlineScriptArgs{
				SiteID:      testSiteID,
				SourceCode:  "console.log('hello');",
				Version:     "1.0.0",
				DisplayName: "Script-With-Dashes",
			},
			want: "invalid characters",
		},
		{
			name: "invalid integrityHash format",
			inputs: InlineScriptArgs{
				SiteID:        testSiteID,
				SourceCode:    "console.log('hello');",
				Version:       "1.0.0",
				DisplayName:   "TestScript",
				IntegrityHash: "md5-abc123",
			},
			want: "must start with 'sha'",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := resource.Create(ctx, infer.CreateRequest[InlineScriptArgs]{
				Inputs: tt.inputs,
			})

			if err == nil {
				t.Fatalf("Create() expected error, got nil")
			}

			if !containsStr(err.Error(), tt.want) {
				t.Errorf("Create() error = %v, want substring %q", err, tt.want)
			}

			if resp.ID != "" {
				t.Errorf("Create() returned ID when expecting error: %s", resp.ID)
			}
		})
	}
}

// TestInlineScriptCreate_ValidIntegrityHash tests that a valid optional integrityHash passes
func TestInlineScriptCreate_ValidIntegrityHash(t *testing.T) {
	resource := &InlineScript{}

	inputs := InlineScriptArgs{
		SiteID:        testSiteID,
		SourceCode:    "console.log('hello');",
		Version:       "1.0.0",
		DisplayName:   "TestScript",
		IntegrityHash: "sha384-abc123",
	}

	ctx := context.Background()
	resp, err := resource.Create(ctx, infer.CreateRequest[InlineScriptArgs]{
		Inputs: inputs,
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create() dry-run with valid integrityHash failed: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("Create() dry-run returned empty ID")
	}
}

// TestInlineScriptCreate_EmptyIntegrityHashAllowed tests that empty integrityHash is valid
func TestInlineScriptCreate_EmptyIntegrityHashAllowed(t *testing.T) {
	resource := &InlineScript{}

	inputs := InlineScriptArgs{
		SiteID:        testSiteID,
		SourceCode:    "console.log('hello');",
		Version:       "1.0.0",
		DisplayName:   "TestScript",
		IntegrityHash: "", // empty is OK for inline scripts
	}

	ctx := context.Background()
	resp, err := resource.Create(ctx, infer.CreateRequest[InlineScriptArgs]{
		Inputs: inputs,
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create() dry-run with empty integrityHash should succeed: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("Create() dry-run returned empty ID")
	}
}

// TestInlineScriptCreate_DryRun tests dry-run behavior
func TestInlineScriptCreate_DryRun(t *testing.T) {
	resource := &InlineScript{}

	inputs := InlineScriptArgs{
		SiteID:      testSiteID,
		SourceCode:  "console.log('hello');",
		Version:     "1.0.0",
		DisplayName: "TestScript",
		CanCopy:     true,
	}

	ctx := context.Background()
	resp, err := resource.Create(ctx, infer.CreateRequest[InlineScriptArgs]{
		Inputs: inputs,
		DryRun: true,
	})
	if err != nil {
		t.Fatalf("Create() dry-run failed: %v", err)
	}

	if resp.ID == "" {
		t.Errorf("Create() dry-run returned empty ID")
	}

	if !containsStr(resp.ID, testSiteID) {
		t.Errorf("Create() dry-run ID should contain siteId: %s", resp.ID)
	}

	if resp.Output.CreatedOn == "" {
		t.Errorf("Create() dry-run should set CreatedOn timestamp")
	}

	if resp.Output.LastUpdated == "" {
		t.Errorf("Create() dry-run should set LastUpdated timestamp")
	}
}

// TestPostInlineScript_Success tests successful creation via API
func TestPostInlineScript_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if !containsStr(r.URL.Path, "/registered_scripts/inline") {
			t.Errorf("Expected /registered_scripts/inline path, got %s", r.URL.Path)
		}

		// Verify the request body
		var reqBody InlineScriptRequest
		if err := json.NewDecoder(r.Body).Decode(&reqBody); err != nil {
			t.Errorf("Failed to decode request body: %v", err)
		}

		if reqBody.SourceCode != "console.log('hello');" {
			t.Errorf("Expected sourceCode 'console.log('hello');', got '%s'", reqBody.SourceCode)
		}

		if reqBody.DisplayName != "TestScript" {
			t.Errorf("Expected displayName 'TestScript', got '%s'", reqBody.DisplayName)
		}

		if reqBody.Version != "1.0.0" {
			t.Errorf("Expected version '1.0.0', got '%s'", reqBody.Version)
		}

		if reqBody.CanCopy != true {
			t.Errorf("Expected canCopy true, got %v", reqBody.CanCopy)
		}

		if reqBody.IntegrityHash != "sha384-abc123" {
			t.Errorf("Expected integrityHash 'sha384-abc123', got '%s'", reqBody.IntegrityHash)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(InlineScriptResponse{
			ID:             testInlineScriptID,
			DisplayName:    "TestScript",
			SourceCode:     "console.log('hello');",
			HostedLocation: "https://cdn.webflow.com/inline/test-script.js",
			IntegrityHash:  "sha384-abc123",
			Version:        "1.0.0",
			CanCopy:        true,
			CreatedOn:      time.Now().Format(time.RFC3339),
			LastUpdated:    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Override base URL for testing
	postInlineScriptBaseURL = server.URL
	defer func() { postInlineScriptBaseURL = "" }()

	client := &http.Client{}
	resp, err := PostInlineScript(
		context.Background(), client, testSiteID,
		"console.log('hello');", "1.0.0", "TestScript", true, "sha384-abc123",
	)
	if err != nil {
		t.Fatalf("PostInlineScript() failed: %v", err)
	}

	if resp.ID != testInlineScriptID {
		t.Errorf("PostInlineScript() ID = %s, want %s", resp.ID, testInlineScriptID)
	}

	if resp.DisplayName != "TestScript" {
		t.Errorf("PostInlineScript() DisplayName = %s, want TestScript", resp.DisplayName)
	}

	if resp.SourceCode != "console.log('hello');" {
		t.Errorf("PostInlineScript() SourceCode = %s, want console.log('hello');", resp.SourceCode)
	}

	if resp.CanCopy != true {
		t.Errorf("PostInlineScript() CanCopy = %v, want true", resp.CanCopy)
	}

	if resp.HostedLocation == "" {
		t.Errorf("PostInlineScript() HostedLocation should not be empty")
	}
}

// TestPostInlineScript_RateLimit tests rate limiting handling
func TestPostInlineScript_RateLimit(t *testing.T) {
	attempt := 0
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		attempt++
		if attempt <= 1 {
			w.Header().Set("Retry-After", "1")
			w.WriteHeader(http.StatusTooManyRequests)
			return
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(InlineScriptResponse{
			ID:          testInlineScriptID,
			DisplayName: "TestScript",
			SourceCode:  "console.log('hello');",
			Version:     "1.0.0",
			CreatedOn:   time.Now().Format(time.RFC3339),
			LastUpdated: time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	postInlineScriptBaseURL = server.URL
	defer func() { postInlineScriptBaseURL = "" }()

	client := &http.Client{}
	resp, err := PostInlineScript(
		context.Background(), client, testSiteID,
		"console.log('hello');", "1.0.0", "TestScript", false, "",
	)
	if err != nil {
		t.Fatalf("PostInlineScript() should retry on rate limit: %v", err)
	}

	if resp.ID != testInlineScriptID {
		t.Errorf("PostInlineScript() ID = %s, want %s", resp.ID, testInlineScriptID)
	}
}

// TestPostInlineScript_ServerError tests server error handling
func TestPostInlineScript_ServerError(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusInternalServerError)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "internal server error"})
	}))
	defer server.Close()

	postInlineScriptBaseURL = server.URL
	defer func() { postInlineScriptBaseURL = "" }()

	client := &http.Client{}
	_, err := PostInlineScript(
		context.Background(), client, testSiteID,
		"console.log('hello');", "1.0.0", "TestScript", false, "",
	)
	if err == nil {
		t.Fatal("PostInlineScript() should fail on server error")
	}
}

// TestInlineScriptDelete_Success tests successful deletion via the shared delete endpoint
func TestInlineScriptDelete_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}

		if !containsStr(r.URL.Path, "/registered_scripts/") {
			t.Errorf("Expected /registered_scripts/ path, got %s", r.URL.Path)
		}

		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Override base URL for testing
	deleteRegisteredScriptBaseURL = server.URL
	defer func() { deleteRegisteredScriptBaseURL = "" }()

	client := &http.Client{}
	err := DeleteRegisteredScript(
		context.Background(), client, testSiteID, testInlineScriptID,
	)
	if err != nil {
		t.Fatalf("DeleteRegisteredScript() for inline script failed: %v", err)
	}
}

// TestInlineScriptDelete_NotFound tests idempotent deletion (404 as success)
func TestInlineScriptDelete_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, _ *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	deleteRegisteredScriptBaseURL = server.URL
	defer func() { deleteRegisteredScriptBaseURL = "" }()

	client := &http.Client{}
	err := DeleteRegisteredScript(
		context.Background(), client, testSiteID, testInlineScriptID,
	)
	if err != nil {
		t.Fatalf("DeleteRegisteredScript() should handle 404 gracefully: %v", err)
	}
}

// TestInlineScriptResourceID tests ID generation and extraction
func TestInlineScriptResourceID(t *testing.T) {
	// Test generation
	resourceID := GenerateInlineScriptResourceID(testSiteID, testInlineScriptID)
	expectedID := fmt.Sprintf("%s/inline_scripts/%s", testSiteID, testInlineScriptID)

	if resourceID != expectedID {
		t.Errorf("GenerateInlineScriptResourceID() = %s, want %s", resourceID, expectedID)
	}

	// Test extraction
	extracted, scriptID, err := ExtractIDsFromInlineScriptResourceID(resourceID)
	if err != nil {
		t.Fatalf("ExtractIDsFromInlineScriptResourceID() failed: %v", err)
	}

	if extracted != testSiteID {
		t.Errorf("ExtractIDsFromInlineScriptResourceID() siteID = %s, want %s", extracted, testSiteID)
	}

	if scriptID != testInlineScriptID {
		t.Errorf("ExtractIDsFromInlineScriptResourceID() scriptID = %s, want %s", scriptID, testInlineScriptID)
	}
}

// TestInlineScriptResourceID_Invalid tests error handling for invalid IDs
func TestInlineScriptResourceID_Invalid(t *testing.T) {
	tests := []struct {
		name    string
		inputID string
		wantErr string
	}{
		{
			name:    "empty ID",
			inputID: "",
			wantErr: "cannot be empty",
		},
		{
			name:    "invalid format",
			inputID: "invalid-format",
			wantErr: "invalid resource ID format",
		},
		{
			name:    "wrong resource type",
			inputID: fmt.Sprintf("%s/webhooks/%s", testSiteID, testInlineScriptID),
			wantErr: "invalid resource ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromInlineScriptResourceID(tt.inputID)

			if err == nil {
				t.Fatalf("ExtractIDsFromInlineScriptResourceID() expected error, got nil")
			}

			if !containsStr(err.Error(), tt.wantErr) {
				t.Errorf("ExtractIDsFromInlineScriptResourceID() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

// TestValidateSourceCode tests source code validation
func TestValidateSourceCode(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid short code", "console.log('hello');", false},
		{"valid at max length", strings.Repeat("a", 2000), false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 2001), true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateSourceCode(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateSourceCode() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// InlineScript Diff Tests
// =============================================================================

// TestInlineScriptDiff_SameVersion_NoChange tests that Diff correctly
// reports NO changes when user input version matches state version.
func TestInlineScriptDiff_SameVersion_NoChange(t *testing.T) {
	resource := &InlineScript{}

	userInputs := InlineScriptArgs{
		SiteID:      "site123",
		SourceCode:  "console.log('hello');",
		Version:     "1.0.0",
		DisplayName: "TestScript",
		CanCopy:     false,
	}

	stateFromRead := InlineScriptState{
		InlineScriptArgs: InlineScriptArgs{
			SiteID:      "site123",
			SourceCode:  "console.log('hello');",
			Version:     "1.0.0",
			DisplayName: "TestScript",
			CanCopy:     false,
		},
	}

	diffReq := infer.DiffRequest[InlineScriptArgs, InlineScriptState]{
		Inputs: userInputs,
		State:  stateFromRead,
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if diffResp.HasChanges {
		t.Errorf("Diff() incorrectly detected changes when values are identical")
		t.Errorf("DetailedDiff: %+v", diffResp.DetailedDiff)
	}

	if diffResp.DetailedDiff != nil {
		if _, hasVersion := diffResp.DetailedDiff["scriptVersion"]; hasVersion {
			t.Errorf("Diff() incorrectly flagged version for change when values are identical")
		}
	}
}

// TestInlineScriptDiff_VersionFromFallback_NoChange tests that
// when API doesn't return version, Diff works correctly with fallback.
func TestInlineScriptDiff_VersionFromFallback_NoChange(t *testing.T) {
	resource := &InlineScript{}

	userInputs := InlineScriptArgs{
		SiteID:      "site123",
		SourceCode:  "console.log('hello');",
		Version:     "1.0.0",
		DisplayName: "TestScript",
		CanCopy:     false,
	}

	stateFromRead := InlineScriptState{
		InlineScriptArgs: InlineScriptArgs{
			SiteID:      "site123",
			SourceCode:  "console.log('hello');",
			Version:     "1.0.0", // Fallback from user input
			DisplayName: "TestScript",
			CanCopy:     false,
		},
	}

	diffReq := infer.DiffRequest[InlineScriptArgs, InlineScriptState]{
		Inputs: userInputs,
		State:  stateFromRead,
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	if diffResp.HasChanges {
		t.Errorf("Diff() incorrectly detected changes with fallback version")
		t.Errorf("DetailedDiff: %+v", diffResp.DetailedDiff)
	}
}

// TestInlineScriptDiff_ChangesRequireReplacement tests that all property changes
// trigger UpdateReplace since Webflow API doesn't support PATCH for inline scripts.
func TestInlineScriptDiff_ChangesRequireReplacement(t *testing.T) {
	resource := &InlineScript{}

	baseInputs := InlineScriptArgs{
		SiteID:        "site123",
		SourceCode:    "console.log('hello');",
		Version:       "1.0.0",
		DisplayName:   "TestScript",
		CanCopy:       false,
		IntegrityHash: "sha384-abc123",
	}

	baseState := InlineScriptState{
		InlineScriptArgs: baseInputs,
	}

	tests := []struct {
		name      string
		modifyFn  func(args *InlineScriptArgs)
		fieldName string
	}{
		{
			name: "siteId change",
			modifyFn: func(args *InlineScriptArgs) {
				args.SiteID = "site456"
			},
			fieldName: "siteId",
		},
		{
			name: "sourceCode change",
			modifyFn: func(args *InlineScriptArgs) {
				args.SourceCode = "console.log('world');"
			},
			fieldName: "sourceCode",
		},
		{
			name: "displayName change",
			modifyFn: func(args *InlineScriptArgs) {
				args.DisplayName = "NewScriptName"
			},
			fieldName: "displayName",
		},
		{
			name: "integrityHash change",
			modifyFn: func(args *InlineScriptArgs) {
				args.IntegrityHash = "sha384-def456"
			},
			fieldName: "integrityHash",
		},
		{
			name: "version change",
			modifyFn: func(args *InlineScriptArgs) {
				args.Version = "2.0.0"
			},
			fieldName: "scriptVersion",
		},
		{
			name: "canCopy change",
			modifyFn: func(args *InlineScriptArgs) {
				args.CanCopy = true
			},
			fieldName: "canCopy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			modifiedInputs := baseInputs
			tt.modifyFn(&modifiedInputs)

			diffReq := infer.DiffRequest[InlineScriptArgs, InlineScriptState]{
				Inputs: modifiedInputs,
				State:  baseState,
			}

			diffResp, err := resource.Diff(context.Background(), diffReq)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}

			if !diffResp.HasChanges {
				t.Errorf("Diff() should detect changes for %s", tt.fieldName)
			}

			if !diffResp.DeleteBeforeReplace {
				t.Errorf("Diff() DeleteBeforeReplace should be true for %s", tt.fieldName)
			}

			if diff, ok := diffResp.DetailedDiff[tt.fieldName]; ok {
				if diff.Kind != p.UpdateReplace {
					t.Errorf("Diff() %s should be UpdateReplace, got %v", tt.fieldName, diff.Kind)
				}
			} else {
				t.Errorf("Diff() DetailedDiff should contain %s", tt.fieldName)
			}
		})
	}
}

// TestInlineScriptUpdate_ReturnsError tests that Update method returns an error
// since Webflow API doesn't support PATCH for inline scripts.
func TestInlineScriptUpdate_ReturnsError(t *testing.T) {
	resource := &InlineScript{}

	updateReq := infer.UpdateRequest[InlineScriptArgs, InlineScriptState]{
		ID: "site123/inline_scripts/script456",
		Inputs: InlineScriptArgs{
			SiteID:      "site123",
			SourceCode:  "console.log('new');",
			Version:     "1.0.0",
			DisplayName: "TestScript",
		},
		State: InlineScriptState{
			InlineScriptArgs: InlineScriptArgs{
				SiteID:      "site123",
				SourceCode:  "console.log('old');",
				Version:     "0.9.0",
				DisplayName: "TestScript",
			},
		},
	}

	_, err := resource.Update(context.Background(), updateReq)

	if err == nil {
		t.Fatal("Update() should return an error")
	}

	if !containsStr(err.Error(), "cannot be updated in-place") {
		t.Errorf("Update() error should mention updates not supported, got: %v", err)
	}

	if !containsStr(err.Error(), "PATCH") {
		t.Errorf("Update() error should mention PATCH not supported, got: %v", err)
	}
}

// TestGetRegisteredScripts_FindsInlineScript tests that the shared list endpoint
// can find inline scripts alongside hosted scripts.
func TestGetRegisteredScripts_FindsInlineScript(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RegisteredScriptsResponse{
			RegisteredScripts: []RegisteredScript{
				{
					ID:             "hosted-script-1",
					DisplayName:    "HostedScript",
					HostedLocation: "https://example.com/script.js",
					IntegrityHash:  "sha384-abc123",
					Version:        "1.0.0",
					CanCopy:        false,
				},
				{
					ID:          testInlineScriptID,
					DisplayName: "InlineScript",
					Version:     "1.0.0",
					CanCopy:     true,
					CreatedOn:   time.Now().Format(time.RFC3339),
					LastUpdated: time.Now().Format(time.RFC3339),
				},
			},
			Pagination: PaginationInfo{
				Limit:  10,
				Offset: 0,
				Total:  2,
			},
		})
	}))
	defer server.Close()

	getRegisteredScriptsBaseURL = server.URL
	defer func() { getRegisteredScriptsBaseURL = "" }()

	client := &http.Client{}
	resp, err := GetRegisteredScripts(context.Background(), client, testSiteID)
	if err != nil {
		t.Fatalf("GetRegisteredScripts() failed: %v", err)
	}

	if len(resp.RegisteredScripts) != 2 {
		t.Errorf("GetRegisteredScripts() returned %d scripts, want 2", len(resp.RegisteredScripts))
	}

	// Find the inline script
	found := false
	for _, script := range resp.RegisteredScripts {
		if script.ID == testInlineScriptID {
			found = true
			if script.DisplayName != "InlineScript" {
				t.Errorf("Inline script DisplayName = %s, want InlineScript", script.DisplayName)
			}
			break
		}
	}

	if !found {
		t.Errorf("GetRegisteredScripts() did not find inline script with ID %s", testInlineScriptID)
	}
}
