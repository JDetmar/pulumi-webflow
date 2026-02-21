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

const testScriptID = "test-script-123"

// TestRegisteredScriptCreate_ValidationErrors tests input validation in Create
func TestRegisteredScriptCreate_ValidationErrors(t *testing.T) {
	resource := &RegisteredScriptResource{}

	tests := []struct {
		name   string
		inputs RegisteredScriptResourceArgs
		want   string
	}{
		{
			name: "invalid siteId",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         "invalid", // Too short
				DisplayName:    "TestScript",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "validation failed",
		},
		{
			name: "missing displayName",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "displayName is required",
		},
		{
			name: "displayName too long",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "ThisIsAVeryLongNameThatExceedsTheMaximumLengthOfFiftyCharacters",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "too long",
		},
		{
			name: "displayName with special chars",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "Script-With-Dashes",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "invalid characters",
		},
		{
			name: "missing hostedLocation",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "hostedLocation is required",
		},
		{
			name: "hostedLocation without https",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "ftp://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1.0.0",
			},
			want: "must start with 'http://' or 'https://'",
		},
		{
			name: "missing integrityHash",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "",
				Version:        "1.0.0",
			},
			want: "integrityHash is required",
		},
		{
			name: "integrityHash invalid format",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "md5-abc123",
				Version:        "1.0.0",
			},
			want: "must start with 'sha'",
		},
		{
			name: "missing version",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "",
			},
			want: "version is required",
		},
		{
			name: "invalid version format",
			inputs: RegisteredScriptResourceArgs{
				SiteID:         testSiteID,
				DisplayName:    "TestScript",
				HostedLocation: "https://example.com/script.js",
				IntegrityHash:  "sha384-abc123",
				Version:        "1",
			},
			want: "Semantic Version format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			ctx := context.Background()
			resp, err := resource.Create(ctx, infer.CreateRequest[RegisteredScriptResourceArgs]{
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

// TestRegisteredScriptCreate_DryRun tests dry-run behavior
func TestRegisteredScriptCreate_DryRun(t *testing.T) {
	resource := &RegisteredScriptResource{}

	inputs := RegisteredScriptResourceArgs{
		SiteID:         testSiteID,
		DisplayName:    "TestScript",
		HostedLocation: "https://example.com/script.js",
		IntegrityHash:  "sha384-abc123",
		Version:        "1.0.0",
		CanCopy:        true,
	}

	ctx := context.Background()
	resp, err := resource.Create(ctx, infer.CreateRequest[RegisteredScriptResourceArgs]{
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

// TestPostRegisteredScript_Success tests successful creation via API
func TestPostRegisteredScript_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "POST" {
			t.Errorf("Expected POST, got %s", r.Method)
		}

		if !containsStr(r.URL.Path, "/registered_scripts/hosted") {
			t.Errorf("Expected /registered_scripts/hosted path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusCreated)
		_ = json.NewEncoder(w).Encode(RegisteredScript{
			ID:             testScriptID,
			DisplayName:    "TestScript",
			HostedLocation: "https://example.com/script.js",
			IntegrityHash:  "sha384-abc123",
			Version:        "1.0.0",
			CanCopy:        true,
			CreatedOn:      time.Now().Format(time.RFC3339),
			LastUpdated:    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Override base URL for testing
	postRegisteredScriptBaseURL = server.URL

	client := &http.Client{}
	resp, err := PostRegisteredScript(
		context.Background(), client, testSiteID,
		"TestScript", "https://example.com/script.js",
		"sha384-abc123", "1.0.0", true,
	)
	if err != nil {
		t.Fatalf("PostRegisteredScript() failed: %v", err)
	}

	if resp.ID != testScriptID {
		t.Errorf("PostRegisteredScript() ID = %s, want %s", resp.ID, testScriptID)
	}

	if resp.DisplayName != "TestScript" {
		t.Errorf("PostRegisteredScript() DisplayName = %s, want TestScript", resp.DisplayName)
	}

	if resp.CanCopy != true {
		t.Errorf("PostRegisteredScript() CanCopy = %v, want true", resp.CanCopy)
	}
}

// TestPatchRegisteredScript_Success tests successful update via API
func TestPatchRegisteredScript_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "PATCH" {
			t.Errorf("Expected PATCH, got %s", r.Method)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RegisteredScript{
			ID:             testScriptID,
			DisplayName:    "TestScript",
			HostedLocation: "https://cdn.example.com/script-v2.js",
			IntegrityHash:  "sha384-def456",
			Version:        "2.0.0",
			CanCopy:        false,
			CreatedOn:      time.Now().Add(-24 * time.Hour).Format(time.RFC3339),
			LastUpdated:    time.Now().Format(time.RFC3339),
		})
	}))
	defer server.Close()

	// Override base URL for testing
	patchRegisteredScriptBaseURL = server.URL

	client := &http.Client{}
	resp, err := PatchRegisteredScript(
		context.Background(), client, testSiteID, testScriptID,
		"TestScript", "https://cdn.example.com/script-v2.js",
		"sha384-def456", "2.0.0", false,
	)
	if err != nil {
		t.Fatalf("PatchRegisteredScript() failed: %v", err)
	}

	if resp.Version != "2.0.0" {
		t.Errorf("PatchRegisteredScript() Version = %s, want 2.0.0", resp.Version)
	}

	if resp.HostedLocation != "https://cdn.example.com/script-v2.js" {
		t.Errorf("PatchRegisteredScript() HostedLocation = %s, want https://cdn.example.com/script-v2.js",
			resp.HostedLocation)
	}
}

// TestDeleteRegisteredScript_Success tests successful deletion via API
func TestDeleteRegisteredScript_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "DELETE" {
			t.Errorf("Expected DELETE, got %s", r.Method)
		}
		w.WriteHeader(http.StatusNoContent)
	}))
	defer server.Close()

	// Override base URL for testing
	deleteRegisteredScriptBaseURL = server.URL

	client := &http.Client{}
	err := DeleteRegisteredScript(
		context.Background(), client, testSiteID, testScriptID,
	)
	if err != nil {
		t.Fatalf("DeleteRegisteredScript() failed: %v", err)
	}
}

// TestDeleteRegisteredScript_NotFound tests idempotent deletion (404 as success)
func TestDeleteRegisteredScript_NotFound(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusNotFound)
		_ = json.NewEncoder(w).Encode(map[string]string{"error": "not found"})
	}))
	defer server.Close()

	deleteRegisteredScriptBaseURL = server.URL

	client := &http.Client{}
	err := DeleteRegisteredScript(
		context.Background(), client, testSiteID, testScriptID,
	)
	if err != nil {
		t.Fatalf("DeleteRegisteredScript() should handle 404 gracefully: %v", err)
	}
}

// TestGetRegisteredScripts_Success tests retrieving all scripts for a site
func TestGetRegisteredScripts_Success(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != "GET" {
			t.Errorf("Expected GET, got %s", r.Method)
		}

		if !containsStr(r.URL.Path, "/registered_scripts") {
			t.Errorf("Expected /registered_scripts path, got %s", r.URL.Path)
		}

		w.Header().Set("Content-Type", "application/json")
		_ = json.NewEncoder(w).Encode(RegisteredScriptsResponse{
			RegisteredScripts: []RegisteredScript{
				{
					ID:             testScriptID,
					DisplayName:    "TestScript",
					HostedLocation: "https://example.com/script.js",
					IntegrityHash:  "sha384-abc123",
					Version:        "1.0.0",
					CanCopy:        true,
					CreatedOn:      time.Now().Add(-48 * time.Hour).Format(time.RFC3339),
					LastUpdated:    time.Now().Format(time.RFC3339),
				},
			},
			Pagination: PaginationInfo{
				Limit:  10,
				Offset: 0,
				Total:  1,
			},
		})
	}))
	defer server.Close()

	// Override base URL for testing
	getRegisteredScriptsBaseURL = server.URL

	client := &http.Client{}
	resp, err := GetRegisteredScripts(context.Background(), client, testSiteID)
	if err != nil {
		t.Fatalf("GetRegisteredScripts() failed: %v", err)
	}

	if len(resp.RegisteredScripts) != 1 {
		t.Errorf("GetRegisteredScripts() returned %d scripts, want 1", len(resp.RegisteredScripts))
	}

	if resp.RegisteredScripts[0].ID != testScriptID {
		t.Errorf("GetRegisteredScripts() ID = %s, want %s", resp.RegisteredScripts[0].ID, testScriptID)
	}

	if resp.Pagination.Total != 1 {
		t.Errorf("GetRegisteredScripts() Pagination.Total = %d, want 1", resp.Pagination.Total)
	}
}

// TestRegisteredScriptResourceID tests ID generation and extraction
func TestRegisteredScriptResourceID(t *testing.T) {
	// Test generation
	resourceID := GenerateRegisteredScriptResourceID(testSiteID, testScriptID)
	expectedID := fmt.Sprintf("%s/registered_scripts/%s", testSiteID, testScriptID)

	if resourceID != expectedID {
		t.Errorf("GenerateRegisteredScriptResourceID() = %s, want %s", resourceID, expectedID)
	}

	// Test extraction
	extracted, scriptID, err := ExtractIDsFromRegisteredScriptResourceID(resourceID)
	if err != nil {
		t.Fatalf("ExtractIDsFromRegisteredScriptResourceID() failed: %v", err)
	}

	if extracted != testSiteID {
		t.Errorf("ExtractIDsFromRegisteredScriptResourceID() siteID = %s, want %s", extracted, testSiteID)
	}

	if scriptID != testScriptID {
		t.Errorf("ExtractIDsFromRegisteredScriptResourceID() scriptID = %s, want %s", scriptID, testScriptID)
	}
}

// TestRegisteredScriptResourceID_Invalid tests error handling for invalid IDs
func TestRegisteredScriptResourceID_Invalid(t *testing.T) {
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
			inputID: fmt.Sprintf("%s/webhooks/%s", testSiteID, testScriptID),
			wantErr: "invalid resource ID format",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			_, _, err := ExtractIDsFromRegisteredScriptResourceID(tt.inputID)

			if err == nil {
				t.Fatalf("ExtractIDsFromRegisteredScriptResourceID() expected error, got nil")
			}

			if !containsStr(err.Error(), tt.wantErr) {
				t.Errorf("ExtractIDsFromRegisteredScriptResourceID() error = %v, want substring %q", err, tt.wantErr)
			}
		})
	}
}

// TestValidateScriptDisplayName tests display name validation
func TestValidateScriptDisplayName(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid alphanumeric", "TestScript123", false},
		{"valid single char", "A", false},
		{"valid 50 chars", strings.Repeat("a", 50), false},
		{"empty", "", true},
		{"too long", strings.Repeat("a", 51), true},
		{"with spaces", "Test Script", true},
		{"with dashes", "Test-Script", true},
		{"with underscores", "Test_Script", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateScriptDisplayName(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateScriptDisplayName() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateHostedLocation tests URL validation
func TestValidateHostedLocation(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid https", "https://example.com/script.js", false},
		{"valid http", "http://example.com/script.js", false},
		{"empty", "", true},
		{"missing scheme", "example.com/script.js", true},
		{"ftp", "ftp://example.com/script.js", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateHostedLocation(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateHostedLocation() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateIntegrityHash tests hash validation
func TestValidateIntegrityHash(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid sha384", "sha384-abc123def456", false},
		{"valid sha256", "sha256-abc123def456", false},
		{"valid sha512", "sha512-abc123def456", false},
		{"empty", "", true},
		{"md5", "md5-abc123", true},
		{"no algorithm", "abc123def456", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateIntegrityHash(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateIntegrityHash() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// TestValidateVersion tests version validation
func TestValidateVersion(t *testing.T) {
	tests := []struct {
		name    string
		input   string
		wantErr bool
	}{
		{"valid semver", "1.0.0", false},
		{"valid semver", "2.3.1", false},
		{"valid semver", "0.0.1", false},
		{"valid two-part version", "1.0", false},
		{"empty", "", true},
		{"no dots", "1", true},
		{"no dots long", "version123", true},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			err := ValidateVersion(tt.input)
			if (err != nil) != tt.wantErr {
				t.Errorf("ValidateVersion() error = %v, wantErr %v", err, tt.wantErr)
			}
		})
	}
}

// =============================================================================
// RegisteredScript Drift Detection Tests
// =============================================================================

// TestRegisteredScriptDiff_SameVersion_NoChange tests that Diff correctly
// reports NO changes when user input version matches state version.
// This tests the bug where pulumi preview showed "version: 1.0.0 => 1.0.0" as an update.
func TestRegisteredScriptDiff_SameVersion_NoChange(t *testing.T) {
	resource := &RegisteredScriptResource{}

	// User input and state have IDENTICAL version
	userInputs := RegisteredScriptResourceArgs{
		SiteID:         "site123",
		DisplayName:    "TestScript",
		HostedLocation: "https://cdn.example.com/script.js",
		IntegrityHash:  "sha384-abc123",
		Version:        "1.0.0",
		CanCopy:        false,
	}

	stateFromRead := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: RegisteredScriptResourceArgs{
			SiteID:         "site123",
			DisplayName:    "TestScript",
			HostedLocation: "https://cdn.example.com/script.js",
			IntegrityHash:  "sha384-abc123",
			Version:        "1.0.0", // SAME as user input
			CanCopy:        false,
		},
	}

	diffReq := infer.DiffRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
		Inputs: userInputs,
		State:  stateFromRead,
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// CRITICAL: No changes should be detected when values are identical
	if diffResp.HasChanges {
		t.Errorf("Diff() incorrectly detected changes when version is identical")
		t.Errorf("DetailedDiff: %+v", diffResp.DetailedDiff)
	}

	if diffResp.DetailedDiff != nil {
		if _, hasVersion := diffResp.DetailedDiff["scriptVersion"]; hasVersion {
			t.Errorf("Diff() incorrectly flagged version for change when values are identical")
		}
	}
}

// TestRegisteredScriptDiff_EmptyVersionInState_ShouldNotTriggerChange tests that
// when API doesn't return version (Read falls back to user input), Diff works correctly.
func TestRegisteredScriptDiff_VersionFromFallback_NoChange(t *testing.T) {
	resource := &RegisteredScriptResource{}

	// User specified version
	userInputs := RegisteredScriptResourceArgs{
		SiteID:         "site123",
		DisplayName:    "TestScript",
		HostedLocation: "https://cdn.example.com/script.js",
		IntegrityHash:  "sha384-abc123",
		Version:        "1.0.0",
		CanCopy:        false,
	}

	// Read() falls back to user input when API returns empty version
	// (as implemented in registeredscript_resource.go lines 303-322)
	stateFromRead := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: RegisteredScriptResourceArgs{
			SiteID:         "site123",
			DisplayName:    "TestScript",
			HostedLocation: "https://cdn.example.com/script.js",
			IntegrityHash:  "sha384-abc123",
			Version:        "1.0.0", // Fallback from user input
			CanCopy:        false,
		},
	}

	diffReq := infer.DiffRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
		Inputs: userInputs,
		State:  stateFromRead,
	}

	diffResp, err := resource.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff() error = %v", err)
	}

	// No changes should be detected
	if diffResp.HasChanges {
		t.Errorf("Diff() incorrectly detected changes with fallback version")
		t.Errorf("DetailedDiff: %+v", diffResp.DetailedDiff)
	}
}

// TestRegisteredScriptDiff_ChangesRequireReplacement tests that all property changes
// trigger UpdateReplace since Webflow API doesn't support PATCH for registered scripts.
func TestRegisteredScriptDiff_ChangesRequireReplacement(t *testing.T) {
	resource := &RegisteredScriptResource{}

	baseInputs := RegisteredScriptResourceArgs{
		SiteID:         "site123",
		DisplayName:    "TestScript",
		HostedLocation: "https://cdn.example.com/script.js",
		IntegrityHash:  "sha384-abc123",
		Version:        "1.0.0",
		CanCopy:        false,
	}

	baseState := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: baseInputs,
	}

	tests := []struct {
		name      string
		modifyFn  func(args *RegisteredScriptResourceArgs)
		fieldName string
	}{
		{
			name: "siteId change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.SiteID = "site456"
			},
			fieldName: "siteId",
		},
		{
			name: "displayName change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.DisplayName = "NewScriptName"
			},
			fieldName: "displayName",
		},
		{
			name: "hostedLocation change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.HostedLocation = "https://cdn.example.com/script-v2.js"
			},
			fieldName: "hostedLocation",
		},
		{
			name: "integrityHash change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.IntegrityHash = "sha384-def456"
			},
			fieldName: "integrityHash",
		},
		{
			name: "version change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.Version = "2.0.0"
			},
			fieldName: "scriptVersion",
		},
		{
			name: "canCopy change",
			modifyFn: func(args *RegisteredScriptResourceArgs) {
				args.CanCopy = true
			},
			fieldName: "canCopy",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			// Create modified inputs
			modifiedInputs := baseInputs
			tt.modifyFn(&modifiedInputs)

			diffReq := infer.DiffRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
				Inputs: modifiedInputs,
				State:  baseState,
			}

			diffResp, err := resource.Diff(context.Background(), diffReq)
			if err != nil {
				t.Fatalf("Diff() error = %v", err)
			}

			// Changes should be detected
			if !diffResp.HasChanges {
				t.Errorf("Diff() should detect changes for %s", tt.fieldName)
			}

			// DeleteBeforeReplace should be true (all changes require replacement)
			if !diffResp.DeleteBeforeReplace {
				t.Errorf("Diff() DeleteBeforeReplace should be true for %s", tt.fieldName)
			}

			// Field should be marked as UpdateReplace, not Update
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

// TestRegisteredScriptUpdate_ReturnsError tests that Update method returns an error
// since Webflow API doesn't support PATCH for registered scripts.
func TestRegisteredScriptUpdate_ReturnsError(t *testing.T) {
	resource := &RegisteredScriptResource{}

	updateReq := infer.UpdateRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
		ID: "site123/registered_scripts/script456",
		Inputs: RegisteredScriptResourceArgs{
			SiteID:         "site123",
			DisplayName:    "TestScript",
			HostedLocation: "https://cdn.example.com/script.js",
			IntegrityHash:  "sha384-abc123",
			Version:        "1.0.0",
		},
		State: RegisteredScriptResourceState{
			RegisteredScriptResourceArgs: RegisteredScriptResourceArgs{
				SiteID:         "site123",
				DisplayName:    "TestScript",
				HostedLocation: "https://cdn.example.com/old-script.js",
				IntegrityHash:  "sha384-old",
				Version:        "0.9.0",
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
