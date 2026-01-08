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
		WithDescription(
			"Unofficial community-maintained Pulumi provider for managing Webflow sites, "+
				"redirects, and robots.txt. Not affiliated with Pulumi Corporation or Webflow, Inc.",
		).
		WithHomepage("https://github.com/JDetmar/pulumi-webflow").
		WithNamespace(Name).
		WithConfig(infer.Config(&Config{})).
		WithResources(
			infer.Resource(&SiteResource{}),
			infer.Resource(&Redirect{}),
			infer.Resource(&RobotsTxt{}),
			infer.Resource(&CollectionResource{}),
			infer.Resource(&CollectionField{}),
			infer.Resource(&CollectionItemResource{}),
			infer.Resource(&PageData{}),
			infer.Resource(&Webhook{}),
			infer.Resource(&Asset{}),
			infer.Resource(&AssetFolder{}),
			infer.Resource(&PageContent{}),
			infer.Resource(&SiteCustomCode{}),
			infer.Resource(&RegisteredScriptResource{}),
			infer.Resource(&PageCustomCode{}),
			infer.Resource(&UserResource{}),
			infer.Resource(&EcommerceSettings{}),
		).
		WithFunctions(
			infer.Function(&GetTokenInfo{}),
			infer.Function(&GetAuthorizedUser{}),
		).
		WithModuleMap(map[tokens.ModuleName]tokens.ModuleName{
			"provider": "index",
		}).
		WithLanguageMap(map[string]any{
			"csharp": map[string]any{
				"rootNamespace":        "Community.Pulumi",
				"respectSchemaVersion": true,
			},
			"java": map[string]any{
				"basePackage": "io.github.jdetmar.pulumi",
				"buildFiles":  "gradle",
			},
			"nodejs": map[string]any{
				"packageName": "@jdetmar/pulumi-webflow",
				"packageDescription": "Unofficial community-maintained Pulumi provider for Webflow. " +
					"Not affiliated with Pulumi Corporation or Webflow, Inc.",
			},
			"python": map[string]any{
				"packageName": "pulumi_webflow",
				"packageDescription": "Unofficial community-maintained Pulumi provider for Webflow. " +
					"Not affiliated with Pulumi Corporation or Webflow, Inc.",
			},
		}).
		Build()
	if err != nil {
		panic(fmt.Errorf("unable to build provider: %w", err))
	}
	return prov
}
