// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// State migration support for the version → scriptVersion rename (v0.9.x → v0.10.x).
//
// In v0.10.0 (commit 4795ae3), the `version` field was renamed to `scriptVersion` in
// all Args and State structs to avoid a collision with the pulumi-go-provider `infer`
// framework, which strips properties named "version" during Diff.
//
// This file provides state migrations so that existing Pulumi state (which stores the
// old field name "version") can be automatically upgraded to "scriptVersion" on first
// access, without requiring manual state surgery.
//
// Affected resources: InlineScript, RegisteredScript, SiteCustomCode, PageCustomCode.

// --- InlineScript migration (v0.9.x state shape) ---

// inlineScriptStateV09 represents the InlineScript state shape from v0.9.x,
// where the version field used the pulumi property name "version".
type inlineScriptStateV09 struct {
	SiteID         string `pulumi:"siteId"`
	SourceCode     string `pulumi:"sourceCode"`
	Version        string `pulumi:"version,optional"`
	DisplayName    string `pulumi:"displayName"`
	CanCopy        bool   `pulumi:"canCopy,optional"`
	IntegrityHash  string `pulumi:"integrityHash,optional"`
	ScriptID       string `pulumi:"scriptId"`
	HostedLocation string `pulumi:"hostedLocation,optional"`
	CreatedOn      string `pulumi:"createdOn,optional"`
	LastUpdated    string `pulumi:"lastUpdated,optional"`
}

// StateMigrations implements infer.CustomStateMigrations for InlineScript.
func (*InlineScript) StateMigrations(_ context.Context) []infer.StateMigrationFunc[InlineScriptState] {
	return []infer.StateMigrationFunc[InlineScriptState]{
		infer.StateMigration(migrateInlineScriptFromV09),
	}
}

func migrateInlineScriptFromV09(
	_ context.Context, old inlineScriptStateV09,
) (infer.MigrationResult[InlineScriptState], error) {
	return infer.MigrationResult[InlineScriptState]{
		Result: &InlineScriptState{
			InlineScriptArgs: InlineScriptArgs{
				SiteID:        old.SiteID,
				SourceCode:    old.SourceCode,
				Version:       old.Version,
				DisplayName:   old.DisplayName,
				CanCopy:       old.CanCopy,
				IntegrityHash: old.IntegrityHash,
			},
			ScriptID:       old.ScriptID,
			HostedLocation: old.HostedLocation,
			CreatedOn:      old.CreatedOn,
			LastUpdated:    old.LastUpdated,
		},
	}, nil
}

// --- RegisteredScript migration (v0.9.x state shape) ---

// registeredScriptStateV09 represents the RegisteredScript state shape from v0.9.x.
type registeredScriptStateV09 struct {
	SiteID         string `pulumi:"siteId"`
	DisplayName    string `pulumi:"displayName"`
	HostedLocation string `pulumi:"hostedLocation"`
	IntegrityHash  string `pulumi:"integrityHash"`
	Version        string `pulumi:"version,optional"`
	CanCopy        bool   `pulumi:"canCopy,optional"`
	ScriptID       string `pulumi:"scriptId"`
	CreatedOn      string `pulumi:"createdOn,optional"`
	LastUpdated    string `pulumi:"lastUpdated,optional"`
}

// StateMigrations implements infer.CustomStateMigrations for RegisteredScriptResource.
func (*RegisteredScriptResource) StateMigrations(
	_ context.Context,
) []infer.StateMigrationFunc[RegisteredScriptResourceState] {
	return []infer.StateMigrationFunc[RegisteredScriptResourceState]{
		infer.StateMigration(migrateRegisteredScriptFromV09),
	}
}

func migrateRegisteredScriptFromV09(
	_ context.Context, old registeredScriptStateV09,
) (infer.MigrationResult[RegisteredScriptResourceState], error) {
	return infer.MigrationResult[RegisteredScriptResourceState]{
		Result: &RegisteredScriptResourceState{
			RegisteredScriptResourceArgs: RegisteredScriptResourceArgs{
				SiteID:         old.SiteID,
				DisplayName:    old.DisplayName,
				HostedLocation: old.HostedLocation,
				IntegrityHash:  old.IntegrityHash,
				Version:        old.Version,
				CanCopy:        old.CanCopy,
			},
			ScriptID:    old.ScriptID,
			CreatedOn:   old.CreatedOn,
			LastUpdated: old.LastUpdated,
		},
	}, nil
}

// --- SiteCustomCode migration (v0.9.x state shape) ---

// customScriptArgsV09 represents the CustomScriptArgs shape from v0.9.x,
// where the version field used the pulumi property name "version".
type customScriptArgsV09 struct {
	ID         string                 `pulumi:"id"`
	Version    string                 `pulumi:"version"`
	Location   string                 `pulumi:"location"`
	Attributes map[string]interface{} `pulumi:"attributes,optional"`
}

// siteCustomCodeStateV09 represents the SiteCustomCode state shape from v0.9.x.
type siteCustomCodeStateV09 struct {
	SiteID      string                `pulumi:"siteId"`
	Scripts     []customScriptArgsV09 `pulumi:"scripts"`
	LastUpdated string                `pulumi:"lastUpdated,optional"`
	CreatedOn   string                `pulumi:"createdOn,optional"`
}

// StateMigrations implements infer.CustomStateMigrations for SiteCustomCode.
func (*SiteCustomCode) StateMigrations(_ context.Context) []infer.StateMigrationFunc[SiteCustomCodeState] {
	return []infer.StateMigrationFunc[SiteCustomCodeState]{
		infer.StateMigration(migrateSiteCustomCodeFromV09),
	}
}

func migrateSiteCustomCodeFromV09(
	_ context.Context, old siteCustomCodeStateV09,
) (infer.MigrationResult[SiteCustomCodeState], error) {
	scripts := make([]CustomScriptArgs, len(old.Scripts))
	for i, s := range old.Scripts {
		attrs := s.Attributes
		if attrs == nil {
			attrs = map[string]interface{}{}
		}
		scripts[i] = CustomScriptArgs{
			ID:         s.ID,
			Version:    s.Version,
			Location:   s.Location,
			Attributes: attrs,
		}
	}
	return infer.MigrationResult[SiteCustomCodeState]{
		Result: &SiteCustomCodeState{
			SiteCustomCodeArgs: SiteCustomCodeArgs{
				SiteID:  old.SiteID,
				Scripts: scripts,
			},
			LastUpdated: old.LastUpdated,
			CreatedOn:   old.CreatedOn,
		},
	}, nil
}

// --- PageCustomCode migration (v0.9.x state shape) ---

// pageCustomCodeScriptV09 represents the PageCustomCodeScript shape from v0.9.x.
type pageCustomCodeScriptV09 struct {
	ID         string                 `pulumi:"id"`
	Version    string                 `pulumi:"version"`
	Location   string                 `pulumi:"location"`
	Attributes map[string]interface{} `pulumi:"attributes,optional"`
}

// pageCustomCodeStateV09 represents the PageCustomCode state shape from v0.9.x.
type pageCustomCodeStateV09 struct {
	PageID      string                    `pulumi:"pageId"`
	Scripts     []pageCustomCodeScriptV09 `pulumi:"scripts"`
	LastUpdated string                    `pulumi:"lastUpdated,optional"`
	CreatedOn   string                    `pulumi:"createdOn,optional"`
}

// StateMigrations implements infer.CustomStateMigrations for PageCustomCode.
func (*PageCustomCode) StateMigrations(_ context.Context) []infer.StateMigrationFunc[PageCustomCodeState] {
	return []infer.StateMigrationFunc[PageCustomCodeState]{
		infer.StateMigration(migratePageCustomCodeFromV09),
	}
}

func migratePageCustomCodeFromV09(
	_ context.Context, old pageCustomCodeStateV09,
) (infer.MigrationResult[PageCustomCodeState], error) {
	scripts := make([]PageCustomCodeScript, len(old.Scripts))
	for i, s := range old.Scripts {
		attrs := s.Attributes
		if attrs == nil {
			attrs = map[string]interface{}{}
		}
		scripts[i] = PageCustomCodeScript{
			ID:         s.ID,
			Version:    s.Version,
			Location:   s.Location,
			Attributes: attrs,
		}
	}
	return infer.MigrationResult[PageCustomCodeState]{
		Result: &PageCustomCodeState{
			PageCustomCodeArgs: PageCustomCodeArgs{
				PageID:  old.PageID,
				Scripts: scripts,
			},
			LastUpdated: old.LastUpdated,
			CreatedOn:   old.CreatedOn,
		},
	}, nil
}
