// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"net/http"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Config defines the provider configuration.
// The apiToken field is marked as a secret and will be automatically handled by Pulumi.
type Config struct {
	// APIToken is the Webflow API v2 bearer token for authentication.
	// Can be set via `pulumi config set webflow:apiToken <value> --secret` or WEBFLOW_API_TOKEN env var.
	APIToken string `pulumi:"apiToken,optional" provider:"secret"`
}

// Annotate adds descriptions to the Config fields for schema generation.
func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.APIToken, "Webflow API v2 bearer token for authentication. "+
		"Can also be set via WEBFLOW_API_TOKEN environment variable.")
}

// Configure validates the configuration and sets up the HTTP client.
// This is called after the configuration is loaded and before any resource operations.
func (c *Config) Configure(ctx context.Context) error {
	// Token validation and HTTP client creation will happen when resources need it
	// The infer package automatically handles environment variable fallback
	return nil
}

// safeGetConfigToken safely retrieves the API token from provider config.
// It uses recover() to handle the case where infer.GetConfig panics
// (which happens when config is not available in the context, e.g., during
// invoke function calls before Configure() completes).
func safeGetConfigToken(ctx context.Context) (token string) {
	defer func() {
		if r := recover(); r != nil {
			// GetConfig panicked - config not available in context
			// This can happen for invoke functions called before Configure()
			token = ""
		}
	}()

	config := infer.GetConfig[*Config](ctx)
	if config != nil {
		return config.APIToken
	}
	return ""
}

// GetHTTPClient retrieves or creates the HTTP client for Webflow API calls.
// It checks for the API token in this order:
// 1. WEBFLOW_API_TOKEN environment variable (preferred for CI/CD and invoke functions)
// 2. Provider config (pulumi config set webflow:apiToken)
func GetHTTPClient(ctx context.Context, version string) (*http.Client, error) {
	// Try environment variable first - this is safe and never panics
	// This also handles invoke functions where config may not be available
	token := getEnvToken()

	// If no env var, safely try to get from provider config
	if token == "" {
		token = safeGetConfigToken(ctx)
	}

	if token == "" {
		return nil, ErrTokenNotConfigured
	}

	// Validate token
	if err := ValidateToken(token); err != nil {
		return nil, err
	}

	// Create HTTP client with authentication
	return CreateHTTPClient(token, version)
}
