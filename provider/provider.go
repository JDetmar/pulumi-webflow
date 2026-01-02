// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

// Package provider wires up the Webflow Pulumi provider and exposes the Provider constructor.
package provider

import (
	"fmt"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
	"github.com/pulumi/pulumi/sdk/v3/go/common/tokens"
)

// Version is initialized by the Go linker to contain the semver of this build.
var Version string

// Name controls how this provider is referenced in package names and elsewhere.
const Name string = "webflow"

// Provider creates a new instance of the Webflow provider with all supported resources.
func Provider() p.Provider {
	// Propagate build version into HTTP client user-agent strings.
	if Version != "" {
		SetProviderVersion(Version)
	}

	prov, err := infer.NewProviderBuilder().
		WithDisplayName("Webflow (Unofficial)").
		WithDescription("Unofficial community-maintained Pulumi provider for managing Webflow sites, redirects, and robots.txt. Not affiliated with Pulumi Corporation or Webflow, Inc.").
		WithHomepage("https://github.com/jdetmar/pulumi-webflow").
		WithNamespace(Name).
		WithConfig(infer.Config(&Config{})).
		WithResources(
			infer.Resource(&SiteResource{}),
			infer.Resource(&Redirect{}),
			infer.Resource(&RobotsTxt{}),
		).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		}).
		WithLanguageMap(map[string]any{
			"csharp": map[string]any{
				"rootNamespace": "Pulumi",
			},
			"nodejs": map[string]any{
				"packageName":        "@jdetmar/pulumi-webflow",
				"packageDescription": "Unofficial community-maintained Pulumi provider for Webflow. Not affiliated with Pulumi Corporation or Webflow, Inc.",
			},
		}).
		Build()
	if err != nil {
		panic(fmt.Errorf("unable to build provider: %w", err))
	}
	return prov
}
