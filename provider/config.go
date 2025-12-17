// Package provider implements the Webflow Pulumi Provider using the modern pulumi-go-provider SDK.
package provider

import (
	"context"
	"net/http"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// Config defines the provider configuration.
// The token field is marked as a secret and will be automatically handled by Pulumi.
type Config struct {
	// Token is the Webflow API v2 bearer token for authentication.
	// Can be set via `pulumi config set webflow:token <value> --secret` or WEBFLOW_API_TOKEN env var.
	Token string `pulumi:"token,optional" provider:"secret"`
}

// Annotate adds descriptions to the Config fields for schema generation.
func (c *Config) Annotate(a infer.Annotator) {
	a.Describe(&c.Token, "Webflow API v2 bearer token for authentication. "+
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
	if config != nil && config.Token != "" {
		token = config.Token
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
