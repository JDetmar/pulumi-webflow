// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"testing"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// TestMigrateInlineScriptFromV09 verifies that old InlineScript state with
// pulumi property name "version" is correctly migrated to "scriptVersion".
func TestMigrateInlineScriptFromV09(t *testing.T) {
	old := inlineScriptStateV09{
		SiteID:         "5f0c8c9e1c9d440000e8d8c3",
		SourceCode:     "console.log('hello');",
		Version:        "1.0.0",
		DisplayName:    "TestScript",
		CanCopy:        true,
		IntegrityHash:  "sha384-abc123",
		ScriptID:       "testscript",
		HostedLocation: "https://example.com/script.js",
		CreatedOn:      "2025-01-01T00:00:00Z",
		LastUpdated:    "2025-01-02T00:00:00Z",
	}

	result, err := migrateInlineScriptFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}

	state := *result.Result
	if state.Version != "1.0.0" {
		t.Errorf("expected Version '1.0.0', got %q", state.Version)
	}
	if state.SiteID != old.SiteID {
		t.Errorf("expected SiteID %q, got %q", old.SiteID, state.SiteID)
	}
	if state.SourceCode != old.SourceCode {
		t.Errorf("expected SourceCode %q, got %q", old.SourceCode, state.SourceCode)
	}
	if state.DisplayName != old.DisplayName {
		t.Errorf("expected DisplayName %q, got %q", old.DisplayName, state.DisplayName)
	}
	if state.CanCopy != old.CanCopy {
		t.Errorf("expected CanCopy %v, got %v", old.CanCopy, state.CanCopy)
	}
	if state.IntegrityHash != old.IntegrityHash {
		t.Errorf("expected IntegrityHash %q, got %q", old.IntegrityHash, state.IntegrityHash)
	}
	if state.ScriptID != old.ScriptID {
		t.Errorf("expected ScriptID %q, got %q", old.ScriptID, state.ScriptID)
	}
	if state.HostedLocation != old.HostedLocation {
		t.Errorf("expected HostedLocation %q, got %q", old.HostedLocation, state.HostedLocation)
	}
	if state.CreatedOn != old.CreatedOn {
		t.Errorf("expected CreatedOn %q, got %q", old.CreatedOn, state.CreatedOn)
	}
	if state.LastUpdated != old.LastUpdated {
		t.Errorf("expected LastUpdated %q, got %q", old.LastUpdated, state.LastUpdated)
	}
}

// TestMigrateRegisteredScriptFromV09 verifies that old RegisteredScript state
// with "version" is correctly migrated to "scriptVersion".
func TestMigrateRegisteredScriptFromV09(t *testing.T) {
	old := registeredScriptStateV09{
		SiteID:         "5f0c8c9e1c9d440000e8d8c3",
		DisplayName:    "Analytics",
		HostedLocation: "https://cdn.example.com/analytics.js",
		IntegrityHash:  "sha384-def456",
		Version:        "2.0.0",
		CanCopy:        false,
		ScriptID:       "analytics",
		CreatedOn:      "2025-01-01T00:00:00Z",
		LastUpdated:    "2025-01-02T00:00:00Z",
	}

	result, err := migrateRegisteredScriptFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}

	state := *result.Result
	if state.Version != "2.0.0" {
		t.Errorf("expected Version '2.0.0', got %q", state.Version)
	}
	if state.SiteID != old.SiteID {
		t.Errorf("expected SiteID %q, got %q", old.SiteID, state.SiteID)
	}
	if state.DisplayName != old.DisplayName {
		t.Errorf("expected DisplayName %q, got %q", old.DisplayName, state.DisplayName)
	}
	if state.HostedLocation != old.HostedLocation {
		t.Errorf("expected HostedLocation %q, got %q", old.HostedLocation, state.HostedLocation)
	}
	if state.ScriptID != old.ScriptID {
		t.Errorf("expected ScriptID %q, got %q", old.ScriptID, state.ScriptID)
	}
}

// TestMigrateSiteCustomCodeFromV09 verifies that old SiteCustomCode state
// with nested scripts using "version" is correctly migrated.
func TestMigrateSiteCustomCodeFromV09(t *testing.T) {
	old := siteCustomCodeStateV09{
		SiteID: "5f0c8c9e1c9d440000e8d8c3",
		Scripts: []customScriptArgsV09{
			{
				ID:       "analytics",
				Version:  "1.0.0",
				Location: "header",
				Attributes: map[string]interface{}{
					"data-config": "test",
				},
			},
			{
				ID:       "tracker",
				Version:  "2.0.0",
				Location: "footer",
			},
		},
		LastUpdated: "2025-01-02T00:00:00Z",
		CreatedOn:   "2025-01-01T00:00:00Z",
	}

	result, err := migrateSiteCustomCodeFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}

	state := *result.Result
	if state.SiteID != old.SiteID {
		t.Errorf("expected SiteID %q, got %q", old.SiteID, state.SiteID)
	}
	if len(state.Scripts) != 2 {
		t.Fatalf("expected 2 scripts, got %d", len(state.Scripts))
	}
	if state.Scripts[0].Version != "1.0.0" {
		t.Errorf("expected script[0] Version '1.0.0', got %q", state.Scripts[0].Version)
	}
	if state.Scripts[0].ID != "analytics" {
		t.Errorf("expected script[0] ID 'analytics', got %q", state.Scripts[0].ID)
	}
	if state.Scripts[0].Attributes["data-config"] != "test" {
		t.Errorf("expected script[0] attribute 'data-config'='test', got %v", state.Scripts[0].Attributes["data-config"])
	}
	if state.Scripts[1].Version != "2.0.0" {
		t.Errorf("expected script[1] Version '2.0.0', got %q", state.Scripts[1].Version)
	}
	if state.Scripts[1].Location != "footer" {
		t.Errorf("expected script[1] Location 'footer', got %q", state.Scripts[1].Location)
	}
}

// TestMigratePageCustomCodeFromV09 verifies that old PageCustomCode state
// with nested scripts using "version" is correctly migrated.
func TestMigratePageCustomCodeFromV09(t *testing.T) {
	old := pageCustomCodeStateV09{
		PageID: "5f0c8c9e1c9d440000e8d8c4",
		Scripts: []pageCustomCodeScriptV09{
			{
				ID:       "widget",
				Version:  "3.0.0",
				Location: "footer",
			},
		},
		LastUpdated: "2025-01-02T00:00:00Z",
		CreatedOn:   "2025-01-01T00:00:00Z",
	}

	result, err := migratePageCustomCodeFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}

	state := *result.Result
	if state.PageID != old.PageID {
		t.Errorf("expected PageID %q, got %q", old.PageID, state.PageID)
	}
	if len(state.Scripts) != 1 {
		t.Fatalf("expected 1 script, got %d", len(state.Scripts))
	}
	if state.Scripts[0].Version != "3.0.0" {
		t.Errorf("expected script Version '3.0.0', got %q", state.Scripts[0].Version)
	}
	if state.Scripts[0].ID != "widget" {
		t.Errorf("expected script ID 'widget', got %q", state.Scripts[0].ID)
	}
}

// TestInlineScriptImplementsStateMigrations verifies the InlineScript resource
// implements the CustomStateMigrations interface.
func TestInlineScriptImplementsStateMigrations(t *testing.T) {
	var _ infer.CustomStateMigrations[InlineScriptState] = (*InlineScript)(nil)
}

// TestRegisteredScriptImplementsStateMigrations verifies the RegisteredScriptResource
// implements the CustomStateMigrations interface.
func TestRegisteredScriptImplementsStateMigrations(t *testing.T) {
	var _ infer.CustomStateMigrations[RegisteredScriptResourceState] = (*RegisteredScriptResource)(nil)
}

// TestSiteCustomCodeImplementsStateMigrations verifies the SiteCustomCode resource
// implements the CustomStateMigrations interface.
func TestSiteCustomCodeImplementsStateMigrations(t *testing.T) {
	var _ infer.CustomStateMigrations[SiteCustomCodeState] = (*SiteCustomCode)(nil)
}

// TestPageCustomCodeImplementsStateMigrations verifies the PageCustomCode resource
// implements the CustomStateMigrations interface.
func TestPageCustomCodeImplementsStateMigrations(t *testing.T) {
	var _ infer.CustomStateMigrations[PageCustomCodeState] = (*PageCustomCode)(nil)
}

// TestMigrateInlineScriptFromV09_EmptyVersion verifies migration handles empty version.
func TestMigrateInlineScriptFromV09_EmptyVersion(t *testing.T) {
	old := inlineScriptStateV09{
		SiteID:      "5f0c8c9e1c9d440000e8d8c3",
		SourceCode:  "console.log('hello');",
		Version:     "", // empty version from old state
		DisplayName: "TestScript",
		ScriptID:    "testscript",
	}

	result, err := migrateInlineScriptFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}
	if result.Result.Version != "" {
		t.Errorf("expected empty Version, got %q", result.Result.Version)
	}
}

// TestMigrateSiteCustomCodeFromV09_EmptyScripts verifies migration handles empty script list.
func TestMigrateSiteCustomCodeFromV09_EmptyScripts(t *testing.T) {
	old := siteCustomCodeStateV09{
		SiteID:  "5f0c8c9e1c9d440000e8d8c3",
		Scripts: []customScriptArgsV09{},
	}

	result, err := migrateSiteCustomCodeFromV09(context.Background(), old)
	if err != nil {
		t.Fatalf("unexpected error: %v", err)
	}
	if result.Result == nil {
		t.Fatal("expected non-nil migration result")
	}
	if len(result.Result.Scripts) != 0 {
		t.Errorf("expected 0 scripts, got %d", len(result.Result.Scripts))
	}
}
