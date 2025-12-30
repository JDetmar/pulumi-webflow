// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// GetHTTPClient retrieves or creates the HTTP client for Webflow API calls.
func GetHTTPClient(ctx context.Context, version string) (*http.Client, error) {
	// Get config from context
	config := infer.GetConfig[*Config](ctx)

	// Try to get token from config, fall back to environment variable
	token := ""
	if config != nil && config.APIToken != "" {
		token = config.APIToken
	} else {
		// Config not available or token empty, try environment variable
		token = getEnvToken()
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
