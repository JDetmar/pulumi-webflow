// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// containsStr checks if string contains substring (case-insensitive)
func containsStr(s, substr string) bool {
	return bytes.Contains(bytes.ToLower([]byte(s)), bytes.ToLower([]byte(substr)))
}

const testSiteID = "5f0c8c9e1c9d440000e8d8c3"

// TestRedirectCreate_ValidationErrors tests input validation in Create
func TestRedirectCreate_ValidationErrors(t *testing.T) {
	redirect := &Redirect{}

	tests := []struct {
		name   string
		inputs RedirectArgs
		want   string
	}{
		{
			name: "invalid siteId",
			inputs: RedirectArgs{
				SiteID:          "invalid", // Too short
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "validation failed",
		},
		{
			name: "missing sourcePath",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "sourcePath is required",
		},
		{
			name: "sourcePath without slash",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "old-page",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "must start with '/'",
		},
		{
			name: "missing destinationPath",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "",
				StatusCode:      301,
			},
			want: "destinationPath is required",
		},
		{
			name: "destinationPath without slash",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "new",
				StatusCode:      301,
			},
			want: "must start with '/'",
		},
		{
			name: "invalid statusCode",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      400,
			},
			want: "must be either 301 or 302",
		},
		{
			name: "sourcePath with space",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old page",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
		{
			name: "sourcePath with query string",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/page?query=value",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
		{
			name: "sourcePath with special char",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/page@name",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
		{
			name: "destinationPath with space",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "/new page",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
		{
			name: "destinationPath with hash",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "/page#anchor",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := infer.CreateRequest[RedirectArgs]{
				Inputs: tt.inputs,
				DryRun: false,
			}

			_, err := redirect.Create(context.Background(), req)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !containsStr(err.Error(), tt.want) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.want, err.Error())
			}
		})
	}
}

// TestRedirectCreate_DryRun tests preview mode for redirect creation
func TestRedirectCreate_DryRun(t *testing.T) {
	redirect := &Redirect{}

	req := infer.CreateRequest[RedirectArgs]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		DryRun: true,
	}

	// Execute
	resp, err := redirect.Create(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Create (DryRun) failed: %v", err)
	}
	if resp.ID == "" {
		t.Error("Expected non-empty ID in dry-run mode")
	}
	if !containsStr(resp.ID, "preview-") {
		t.Errorf("Expected preview ID in dry-run, got '%s'", resp.ID)
	}
}

// TestRedirectUpdate_DryRun tests preview mode for redirect update
func TestRedirectUpdate_DryRun(t *testing.T) {
	redirect := &Redirect{}

	req := infer.UpdateRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/updated-page",
			StatusCode:      302,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
		DryRun: true,
	}

	// Execute
	resp, err := redirect.Update(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Update (DryRun) failed: %v", err)
	}
	if resp.Output.DestinationPath != "/updated-page" {
		t.Errorf("Expected destination path '/updated-page', got '%s'", resp.Output.DestinationPath)
	}
	if resp.Output.StatusCode != 302 {
		t.Errorf("Expected status code 302, got %d", resp.Output.StatusCode)
	}
}

// TestRedirectUpdate_ValidationErrors tests input validation in Update
func TestRedirectUpdate_ValidationErrors(t *testing.T) {
	redirect := &Redirect{}

	tests := []struct {
		name   string
		inputs RedirectArgs
		want   string
	}{
		{
			name: "invalid siteId",
			inputs: RedirectArgs{
				SiteID:          "invalid",
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      301,
			},
			want: "validation failed",
		},
		{
			name: "missing destinationPath",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "",
				StatusCode:      301,
			},
			want: "destinationPath is required",
		},
		{
			name: "invalid statusCode",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "/new",
				StatusCode:      303,
			},
			want: "must be either 301 or 302",
		},
		{
			name: "destinationPath with invalid chars",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old",
				DestinationPath: "/new page?q=1",
				StatusCode:      301,
			},
			want: "invalid characters",
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			req := infer.UpdateRequest[RedirectArgs, RedirectState]{
				Inputs: tt.inputs,
				State: RedirectState{
					RedirectArgs: RedirectArgs{
						SiteID:          testSiteID,
						SourcePath:      "/old",
						DestinationPath: "/old-page",
						StatusCode:      301,
					},
				},
				DryRun: false,
			}

			_, err := redirect.Update(context.Background(), req)
			if err == nil {
				t.Fatal("Expected validation error, got nil")
			}
			if !containsStr(err.Error(), tt.want) {
				t.Errorf("Expected error containing '%s', got '%s'", tt.want, err.Error())
			}
		})
	}
}

// TestRedirectDiff_NoChanges tests diff when no properties changed
func TestRedirectDiff_NoChanges(t *testing.T) {
	redirect := &Redirect{}

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if resp.HasChanges {
		t.Error("Expected no changes")
	}
}

// TestRedirectDiff_SiteIDChange tests that siteId change requires replacement
func TestRedirectDiff_SiteIDChange(t *testing.T) {
	redirect := &Redirect{}
	newSiteID := "6f1d9d0f2d0e551111f9e9d4"

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          newSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for siteId change")
	}
	if !resp.DeleteBeforeReplace {
		t.Error("Expected DeleteBeforeReplace=true for siteId change")
	}
	if _, ok := resp.DetailedDiff["siteId"]; !ok {
		t.Error("Expected siteId in DetailedDiff")
	}
}

// TestRedirectDiff_SourcePathChange tests that sourcePath change requires replacement
func TestRedirectDiff_SourcePathChange(t *testing.T) {
	redirect := &Redirect{}

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/new-old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for sourcePath change")
	}
	if !resp.DeleteBeforeReplace {
		t.Error("Expected DeleteBeforeReplace=true for sourcePath change")
	}
}

// TestRedirectDiff_DestinationPathChange tests that destinationPath change requires replacement
// NOTE: Due to Webflow API limitation, all changes require replacement (delete + recreate)
func TestRedirectDiff_DestinationPathChange(t *testing.T) {
	redirect := &Redirect{}

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/updated-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for destinationPath change")
	}
	// Due to Webflow API limitation, all changes require replacement
	if !resp.DeleteBeforeReplace {
		t.Error("Expected DeleteBeforeReplace=true for destinationPath change (Webflow API limitation)")
	}
	if _, ok := resp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
}

// TestRedirectDiff_StatusCodeChange tests that statusCode change requires replacement
// NOTE: Due to Webflow API limitation, all changes require replacement (delete + recreate)
func TestRedirectDiff_StatusCodeChange(t *testing.T) {
	redirect := &Redirect{}

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      302,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for statusCode change")
	}
	// Due to Webflow API limitation, all changes require replacement
	if !resp.DeleteBeforeReplace {
		t.Error("Expected DeleteBeforeReplace=true for statusCode change (Webflow API limitation)")
	}
	if _, ok := resp.DetailedDiff["statusCode"]; !ok {
		t.Error("Expected statusCode in DetailedDiff")
	}
}

// TestRedirectDiff_MultipleFieldsChange tests that when both destinationPath and statusCode differ,
// Diff triggers replacement (due to Webflow API limitation, each change causes early return with replacement)
func TestRedirectDiff_MultipleFieldsChange(t *testing.T) {
	redirect := &Redirect{}

	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/updated-page",
			StatusCode:      302,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for multiple field changes")
	}
	// Due to Webflow API limitation, all changes require replacement
	// The Diff method returns early on first detected change, so only one field shows in DetailedDiff
	if !resp.DeleteBeforeReplace {
		t.Error("Expected DeleteBeforeReplace=true for changes (Webflow API limitation)")
	}
	// Since destinationPath is checked before statusCode, it appears in DetailedDiff
	if _, ok := resp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
	// Only 1 field in DetailedDiff due to early return behavior
	if len(resp.DetailedDiff) != 1 {
		t.Errorf("Expected 1 field in DetailedDiff (early return on first change), got %d", len(resp.DetailedDiff))
	}
}

// TestDiff_WithDriftedState tests Diff correctly identifies drift from Read output
func TestDiff_WithDriftedState(t *testing.T) {
	redirect := &Redirect{}

	// Simulate: code defines 301, but API returned 302 (drift detected by Read)
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      302, // Drifted value from API
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for drifted state")
	}
	if _, ok := resp.DetailedDiff["statusCode"]; !ok {
		t.Error("Expected statusCode in DetailedDiff for drift")
	}
}

// TestDiff_WithDriftedDestinationPath tests drift when destination path drifted
func TestDiff_WithDriftedDestinationPath(t *testing.T) {
	redirect := &Redirect{}

	// Simulate drift where destination was manually changed in Webflow UI
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/products",
			DestinationPath: "/new-products",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/products",
				DestinationPath: "/old-location", // Manual change in Webflow UI
				StatusCode:      301,
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for destination path drift")
	}
	if _, ok := resp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
}

// TestDiff_WithDriftedStatusCode tests drift when status code manually changed in Webflow
func TestDiff_WithDriftedStatusCode(t *testing.T) {
	redirect := &Redirect{}

	// Simulate drift where status code was manually changed in Webflow UI
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/blog",
			DestinationPath: "/news",
			StatusCode:      301, // Code says permanent
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/blog",
				DestinationPath: "/news",
				StatusCode:      302, // But API shows temporary (manual change)
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for status code drift")
	}
	if _, ok := resp.DetailedDiff["statusCode"]; !ok {
		t.Error("Expected statusCode in DetailedDiff")
	}
}

// TestDiff_WithDeletedResource tests Diff with deleted resource state
func TestDiff_WithDeletedResource(t *testing.T) {
	redirect := &Redirect{}

	// When resource is deleted, code still defines it but state is empty
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/new-page",
				StatusCode:      301,
			},
		},
	}

	// Execute - Note: In real scenario, Pulumi would call Create to recreate
	// Diff would just compare what's defined vs what was there
	resp, err := redirect.Diff(context.Background(), req)
	// Verify no changes (Diff only compares present state)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	// Since both inputs and state are identical, Diff reports no changes
	// Pulumi detects the missing resource through the Read returning empty ID
	if resp.HasChanges {
		t.Error("Expected no changes in Diff (Pulumi detects deletion via Read)")
	}
}

// TestDiff_WithBothFieldsDrifted tests drift when both destinationPath and statusCode changed in Webflow
// Due to Webflow API limitation, Diff returns early on first detected change
func TestDiff_WithBothFieldsDrifted(t *testing.T) {
	redirect := &Redirect{}

	// Simulate drift where both destination and status code were changed in Webflow UI
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-path",
			DestinationPath: "/desired-dest",
			StatusCode:      301, // Code says permanent
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-path",
				DestinationPath: "/manual-dest", // Manual change in Webflow UI
				StatusCode:      302,            // Also changed in UI
			},
		},
	}

	// Execute
	resp, err := redirect.Diff(context.Background(), req)
	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for both fields drifted")
	}
	// Due to early return behavior, only first detected field (destinationPath) appears
	if _, ok := resp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
	// Only 1 field in DetailedDiff due to early return behavior
	if len(resp.DetailedDiff) != 1 {
		t.Errorf("Expected 1 field in DetailedDiff (early return), got %d", len(resp.DetailedDiff))
	}
}

// TestDriftPerformance tests that drift detection completes within NFR3 requirement
func TestDriftPerformance(t *testing.T) {
	redirect := &Redirect{}

	// Create a request with drifted state
	req := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/new-page",
			StatusCode:      301,
		},
		State: RedirectState{
			RedirectArgs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/old-page",
				DestinationPath: "/changed-page",
				StatusCode:      302,
			},
		},
	}

	// Execute Diff and measure performance
	start := time.Now()
	resp, err := redirect.Diff(context.Background(), req)
	elapsed := time.Since(start)

	// Verify
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}
	if !resp.HasChanges {
		t.Error("Expected HasChanges=true for drifted state")
	}

	// NFR3: Drift detection must complete within 10 seconds
	// Diff is in-memory operation, should complete in milliseconds
	if elapsed > 100*time.Millisecond {
		t.Errorf("Drift detection took too long: %v (should be <100ms for in-memory Diff)", elapsed)
	}
}

// TestDriftWorkflow_DetectAndCorrect tests complete drift detection and correction workflow
func TestDriftWorkflow_DetectAndCorrect(t *testing.T) {
	redirect := &Redirect{}

	// Step 1: Simulate Read detecting drift (API returned different values)
	readResp := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/drifted-dest", // Manual change in Webflow
			StatusCode:      302,             // Manual change in Webflow
		},
	}

	// Step 2: Diff identifies the changes
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/correct-dest", // Code-defined value
			StatusCode:      301,             // Code-defined value
		},
		State: readResp,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify Diff detected the drift
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect drift")
	}
	// Due to early return behavior, only first detected field appears
	if len(diffResp.DetailedDiff) != 1 {
		t.Errorf("Expected 1 field in DetailedDiff (early return), got %d", len(diffResp.DetailedDiff))
	}

	// Step 3: Update would correct the drift (simulated - actual Update requires API)
	// In real Pulumi flow: Update() calls PatchRedirect with code values
	expectedCorrectedState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/correct-dest",
			StatusCode:      301,
		},
	}

	// Verify no drift after correction
	finalCheckReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: diffReq.Inputs,
		State:  expectedCorrectedState,
	}

	finalCheckResp, err := redirect.Diff(context.Background(), finalCheckReq)
	if err != nil {
		t.Fatalf("Final check Diff failed: %v", err)
	}
	if finalCheckResp.HasChanges {
		t.Error("Expected no changes after drift correction")
	}
}

// TestDriftWorkflow_DetectAndRecreate tests drift detection and recreation after deletion
func TestDriftWorkflow_DetectAndRecreate(t *testing.T) {
	redirect := &Redirect{}

	// Step 1: Resource was deleted in Webflow (Read returns empty ID)
	// This signals to Pulumi that resource is missing

	// Step 2: Code still defines the resource
	codeDefinedInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/old-page",
		DestinationPath: "/new-page",
		StatusCode:      301,
	}

	// Step 3: Diff with deleted state
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeDefinedInputs,
		State: RedirectState{
			RedirectArgs: codeDefinedInputs,
		},
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Note: Diff reports no changes (both inputs and state are identical)
	// Pulumi detects the missing resource via the empty ID from Read
	if diffResp.HasChanges {
		t.Error("Expected no changes in Diff (Pulumi detects deletion via Read returning empty ID)")
	}

	// Step 4: After Pulumi detects missing resource, it calls Create to recreate
	// Create would call PostRedirect with the code-defined values
	// Result: Resource recreated with correct values
}

// TestReadDriftDetection tests that Read operation detects API drift correctly
func TestReadDriftDetection(t *testing.T) {
	// This test validates the integration between Read, Diff, and Update
	// Read() returns current API state (which may differ from code-defined inputs)
	// Diff() identifies the differences
	// Update() corrects them via API calls

	redirect := &Redirect{}

	// Scenario: Code defines one set of values, but API has different values
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/products",
		DestinationPath: "/shop",
		StatusCode:      301,
	}

	// Simulate API response with drifted values (changed in Webflow UI)
	apiState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/products",
			DestinationPath: "/catalog", // Manual change
			StatusCode:      302,        // Manual change
		},
	}

	// Diff detects the drift
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  apiState,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify drift detected
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect drift between code and API values")
	}

	// Due to early return behavior, only first detected field appears
	// destinationPath is checked before statusCode
	if _, ok := diffResp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
}

// STORY 2.4: State Refresh Tests
// Note: These tests focus on Diff and state logic verification
// Read() testing requires full Pulumi context with config (tested in integration scenarios)

// TestRefresh_UnchangedState_DiffDetectsNoChanges verifies Diff when state matches code
func TestRefresh_UnchangedState_DiffDetectsNoChanges(t *testing.T) {
	redirect := &Redirect{}

	// Code-defined values
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/about",
		DestinationPath: "/about-us",
		StatusCode:      301,
	}

	// State matches code (no manual changes)
	currentState := RedirectState{
		RedirectArgs: codeInputs,
	}

	// Diff should show no changes
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  currentState,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify no changes detected
	if diffResp.HasChanges {
		t.Error("Expected no changes when state matches code")
	}
}

// TestRefresh_ModifiedState_DiffDetectsChanges verifies Diff detects manual Webflow changes
func TestRefresh_ModifiedState_DiffDetectsChanges(t *testing.T) {
	redirect := &Redirect{}

	// Code-defined values
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/products",
		DestinationPath: "/shop",
		StatusCode:      301,
	}

	// Refreshed state from Webflow with manual changes
	refreshedState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/products",
			DestinationPath: "/catalog", // Manual change in Webflow
			StatusCode:      302,        // Manual change in Webflow
		},
	}

	// Diff should detect changes
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  refreshedState,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify changes detected
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect changes in refreshed state")
	}

	// Due to early return behavior, only first detected field appears
	if _, ok := diffResp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff")
	}
}

// TestRefresh_StateWithEmptyId_DiffDetectsPropertyChange verifies Diff detects when
// state with empty ID differs from code.
func TestRefresh_StateWithEmptyId_DiffDetectsPropertyChange(t *testing.T) {
	redirect := &Redirect{}

	// Code-defined values
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/old-page",
		DestinationPath: "/new-page",
		StatusCode:      301,
	}

	// Refreshed state shows different values (as if resource was modified)
	// Empty ID case is handled by Pulumi's state management, not Diff
	refreshedState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/old-page",
			DestinationPath: "/different-dest", // Different from code
			StatusCode:      302,               // Different from code
		},
	}

	// Diff should detect that state differs from inputs
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  refreshedState,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify changes detected (state differs from code)
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect that state differs from code")
	}
}

// TestRefreshWorkflow_DeleteAndRecreate tests that after deletion and refresh, Diff shows changes
func TestRefreshWorkflow_DeleteAndRecreate(t *testing.T) {
	redirect := &Redirect{}

	// Code-defined values (resource exists in code)
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/contact",
		DestinationPath: "/contact-us",
		StatusCode:      301,
	}

	// After refresh following deletion, state is empty/default
	// This represents the state after Read() returns empty ID
	emptyState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          "",
			SourcePath:      "",
			DestinationPath: "",
			StatusCode:      0,
		},
	}

	// Diff should detect that all properties differ from code
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  emptyState,
	}

	diffResp, err := redirect.Diff(context.Background(), diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify changes detected (resource needs to be created)
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect that resource needs creation after deletion")
	}
}

// TestRefreshBatch_MultipleRedirectsConsistency verifies batch refresh logic consistency
func TestRefreshBatch_MultipleRedirectsConsistency(t *testing.T) {
	redirect := &Redirect{}

	tests := []struct {
		name   string
		inputs RedirectArgs
		state  RedirectState
		expect bool // HasChanges
	}{
		{
			name: "Batch refresh with no changes",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/batch-1",
				DestinationPath: "/batch-dest-1",
				StatusCode:      301,
			},
			state: RedirectState{
				RedirectArgs: RedirectArgs{
					SiteID:          testSiteID,
					SourcePath:      "/batch-1",
					DestinationPath: "/batch-dest-1",
					StatusCode:      301,
				},
			},
			expect: false,
		},
		{
			name: "Batch refresh with changes",
			inputs: RedirectArgs{
				SiteID:          testSiteID,
				SourcePath:      "/batch-2",
				DestinationPath: "/batch-dest-2",
				StatusCode:      301,
			},
			state: RedirectState{
				RedirectArgs: RedirectArgs{
					SiteID:          testSiteID,
					SourcePath:      "/batch-2",
					DestinationPath: "/batch-dest-changed", // Manual change
					StatusCode:      302,                   // Manual change
				},
			},
			expect: true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
				Inputs: tt.inputs,
				State:  tt.state,
			}

			diffResp, err := redirect.Diff(context.Background(), diffReq)
			if err != nil {
				t.Fatalf("Diff failed: %v", err)
			}

			if diffResp.HasChanges != tt.expect {
				t.Errorf("Expected HasChanges=%v, got %v", tt.expect, diffResp.HasChanges)
			}
		})
	}
}

// TestRefreshDetectsStateChanges verifies that Diff shows changes after state refresh
func TestRefreshDetectsStateChanges(t *testing.T) {
	redirect := &Redirect{}

	// Code-defined values
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/docs",
		DestinationPath: "/documentation",
		StatusCode:      301,
	}

	// Original state
	originalState := RedirectState{
		RedirectArgs: codeInputs,
	}

	// After refresh, state reflects Webflow changes
	refreshedState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      "/docs",
			DestinationPath: "/guides", // Changed in Webflow
			StatusCode:      302,       // Changed in Webflow
		},
	}

	// Diff original vs code (before refresh)
	diffBeforeRefresh := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  originalState,
	}
	diffBeforeResp, err := redirect.Diff(context.Background(), diffBeforeRefresh)
	if err != nil {
		t.Fatalf("Diff before refresh failed: %v", err)
	}

	// Diff refreshed vs code (after refresh)
	diffAfterRefresh := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  refreshedState,
	}
	diffAfterResp, err := redirect.Diff(context.Background(), diffAfterRefresh)
	if err != nil {
		t.Fatalf("Diff after refresh failed: %v", err)
	}

	// Before refresh: no changes
	if diffBeforeResp.HasChanges {
		t.Error("Expected no changes before refresh")
	}

	// After refresh: changes detected
	if !diffAfterResp.HasChanges {
		t.Error("Expected changes after refresh to reflect Webflow modifications")
	}
}

// NFR2 Performance Tests: State refresh completes within 15 seconds for up to 100 resources

// TestRefreshPerformance_100Redirects validates NFR2: refresh 100 resources in <15 seconds
func TestRefreshPerformance_100Redirects(t *testing.T) {
	// Generate 100 redirects for the mock response
	redirects := make([]RedirectRule, 100)
	for i := 0; i < 100; i++ {
		redirects[i] = RedirectRule{
			ID:              fmt.Sprintf("redirect-%d", i),
			SourcePath:      fmt.Sprintf("/old-%d", i),
			DestinationPath: fmt.Sprintf("/new-%d", i),
			StatusCode:      301,
		}
	}

	// Create mock server returning 100 redirects
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := RedirectResponse{Redirects: redirects}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	// Override API base URL
	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	// Time the refresh operation for 100 resources
	start := time.Now()

	// Simulate refresh: GetRedirects (1 API call) + 100 Diff operations
	response, err := GetRedirects(ctx, client, testSiteID)
	if err != nil {
		t.Fatalf("GetRedirects failed: %v", err)
	}

	if len(response.Redirects) != 100 {
		t.Fatalf("Expected 100 redirects, got %d", len(response.Redirects))
	}

	// Simulate Diff for each redirect (as Pulumi would do during refresh)
	redirect := &Redirect{}
	for i, rule := range response.Redirects {
		codeInputs := RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      rule.SourcePath,
			DestinationPath: rule.DestinationPath,
			StatusCode:      rule.StatusCode,
		}
		state := RedirectState{
			RedirectArgs: codeInputs,
		}

		diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
			Inputs: codeInputs,
			State:  state,
		}

		_, err := redirect.Diff(ctx, diffReq)
		if err != nil {
			t.Fatalf("Diff failed for redirect %d: %v", i, err)
		}
	}

	elapsed := time.Since(start)

	// NFR2: Must complete within 15 seconds
	if elapsed > 15*time.Second {
		t.Errorf("NFR2 FAILED: Refresh of 100 resources took %v (limit: 15s)", elapsed)
	}

	t.Logf("NFR2 PASSED: Refreshed 100 resources in %v", elapsed)
}

// TestRefreshPerformance_SingleRedirect tests baseline performance
func TestRefreshPerformance_SingleRedirect(t *testing.T) {
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := RedirectResponse{
			Redirects: []RedirectRule{
				{ID: "redirect-1", SourcePath: "/old", DestinationPath: "/new", StatusCode: 301},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	start := time.Now()

	response, err := GetRedirects(ctx, client, testSiteID)
	if err != nil {
		t.Fatalf("GetRedirects failed: %v", err)
	}

	// Perform Diff
	redirect := &Redirect{}
	rule := response.Redirects[0]
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      rule.SourcePath,
		DestinationPath: rule.DestinationPath,
		StatusCode:      rule.StatusCode,
	}
	state := RedirectState{
		RedirectArgs: codeInputs,
	}

	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  state,
	}

	_, err = redirect.Diff(ctx, diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	elapsed := time.Since(start)

	// Single redirect should complete in <1 second
	if elapsed > 1*time.Second {
		t.Errorf("Single redirect refresh took %v (expected <1s)", elapsed)
	}

	t.Logf("Single redirect refreshed in %v", elapsed)
}

// TestRefreshAPI_GetRedirects_ReturnsCurrentState tests that GetRedirects API returns current Webflow state
func TestRefreshAPI_GetRedirects_ReturnsCurrentState(t *testing.T) {
	// Simulate Webflow returning a redirect with modified values (manual change in UI)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		// Return state that differs from what Pulumi expects (simulating manual UI change)
		response := RedirectResponse{
			Redirects: []RedirectRule{
				{
					ID:              "redirect-abc123",
					SourcePath:      "/products",
					DestinationPath: "/catalog-changed", // Manual change in Webflow
					StatusCode:      302,                // Manual change in Webflow
				},
			},
		}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	// This simulates what Read() does internally: fetch current state from Webflow
	response, err := GetRedirects(ctx, client, testSiteID)
	if err != nil {
		t.Fatalf("GetRedirects failed: %v", err)
	}

	// Verify API returned current Webflow state
	if len(response.Redirects) != 1 {
		t.Fatalf("Expected 1 redirect, got %d", len(response.Redirects))
	}

	apiRedirect := response.Redirects[0]

	// Verify the API returned the "drifted" state (manual changes)
	if apiRedirect.DestinationPath != "/catalog-changed" {
		t.Errorf("Expected API to return current Webflow state '/catalog-changed', got '%s'", apiRedirect.DestinationPath)
	}
	if apiRedirect.StatusCode != 302 {
		t.Errorf("Expected API to return current status code 302, got %d", apiRedirect.StatusCode)
	}

	// Now simulate what Pulumi does: compare API state with code-defined inputs
	redirect := &Redirect{}
	codeInputs := RedirectArgs{
		SiteID:          testSiteID,
		SourcePath:      "/products",
		DestinationPath: "/shop", // Code says /shop
		StatusCode:      301,     // Code says 301
	}

	// Build state from API response (what Read() would return)
	refreshedState := RedirectState{
		RedirectArgs: RedirectArgs{
			SiteID:          testSiteID,
			SourcePath:      apiRedirect.SourcePath,
			DestinationPath: apiRedirect.DestinationPath, // /catalog-changed from API
			StatusCode:      apiRedirect.StatusCode,      // 302 from API
		},
	}

	// Diff should detect the drift between code and refreshed API state
	diffReq := infer.DiffRequest[RedirectArgs, RedirectState]{
		Inputs: codeInputs,
		State:  refreshedState,
	}

	diffResp, err := redirect.Diff(ctx, diffReq)
	if err != nil {
		t.Fatalf("Diff failed: %v", err)
	}

	// Verify drift detected after refresh
	if !diffResp.HasChanges {
		t.Error("Expected Diff to detect changes between code and refreshed API state")
	}

	// Due to early return behavior, only first detected field appears
	if _, ok := diffResp.DetailedDiff["destinationPath"]; !ok {
		t.Error("Expected destinationPath in DetailedDiff (code=/shop, api=/catalog-changed)")
	}
}

// TestRefreshAPI_GetRedirects_ResourceDeleted tests that API returns empty when resource deleted
func TestRefreshAPI_GetRedirects_ResourceDeleted(t *testing.T) {
	// Simulate Webflow returning empty list (redirect was deleted in UI)
	server := httptest.NewServer(http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.Header().Set("Content-Type", "application/json")
		w.WriteHeader(http.StatusOK)
		response := RedirectResponse{Redirects: []RedirectRule{}}
		_ = json.NewEncoder(w).Encode(response)
	}))
	defer server.Close()

	oldURL := getRedirectsBaseURL
	getRedirectsBaseURL = server.URL
	defer func() { getRedirectsBaseURL = oldURL }()

	client := &http.Client{Timeout: 30 * time.Second}
	ctx := context.Background()

	response, err := GetRedirects(ctx, client, testSiteID)
	if err != nil {
		t.Fatalf("GetRedirects failed: %v", err)
	}

	// Verify empty list returned (resource deleted in Webflow)
	if len(response.Redirects) != 0 {
		t.Errorf("Expected 0 redirects (deleted), got %d", len(response.Redirects))
	}

	// Simulate what Read() does: search for specific redirect ID
	targetRedirectID := "redirect-abc123"
	var foundRedirect *RedirectRule
	for _, r := range response.Redirects {
		if r.ID == targetRedirectID {
			foundRedirect = &r
			break
		}
	}

	// Redirect not found = deleted in Webflow
	if foundRedirect != nil {
		t.Error("Expected redirect to not be found (simulating deletion)")
	}

	// In Read(), this would return empty ID to signal deletion to Pulumi
	// Pulumi then marks resource for recreation on next `pulumi up`
}
