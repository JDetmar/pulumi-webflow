// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"fmt"

	"github.com/pulumi/pulumi-go-provider/infer"
)

// GetTokenInfo is a Pulumi Function that retrieves information about the current API token.
// It calls the Webflow /token/introspect endpoint to get authorization details.
type GetTokenInfo struct{}

// GetTokenInfoInput defines the input parameters for the GetTokenInfo function.
// This function has no required inputs as it operates on the configured provider token.
type GetTokenInfoInput struct {
	// No inputs required - uses the provider's configured API token
}

// GetTokenInfoAuthorizedTo defines the authorization scope in the output.
type GetTokenInfoAuthorizedTo struct {
	// SiteIDs is the list of site IDs this token is authorized to access.
	SiteIDs []string `pulumi:"siteIds"`
	// WorkspaceIDs is the list of workspace IDs this token is authorized to access.
	WorkspaceIDs []string `pulumi:"workspaceIds"`
	// UserIDs is the list of user IDs this token is authorized to access.
	UserIDs []string `pulumi:"userIds"`
}

// GetTokenInfoAuthorization defines the authorization details in the output.
type GetTokenInfoAuthorization struct {
	// ID is the unique identifier for this authorization.
	ID string `pulumi:"id"`
	// CreatedOn is the timestamp when this authorization was created.
	CreatedOn string `pulumi:"createdOn"`
	// LastUsed is the timestamp when this token was last used.
	LastUsed string `pulumi:"lastUsed"`
	// GrantType is the OAuth grant type used to obtain this token.
	GrantType string `pulumi:"grantType"`
	// RateLimit is the rate limit for this token (requests per minute).
	RateLimit int `pulumi:"rateLimit"`
	// Scope is the OAuth scopes granted to this token.
	Scope string `pulumi:"scope"`
	// AuthorizedTo contains the resources this token can access.
	AuthorizedTo GetTokenInfoAuthorizedTo `pulumi:"authorizedTo"`
}

// GetTokenInfoApplication defines the application details in the output.
type GetTokenInfoApplication struct {
	// ID is the unique identifier for the application.
	ID string `pulumi:"id"`
	// Description is the application description.
	Description string `pulumi:"description"`
	// Homepage is the application homepage URL.
	Homepage string `pulumi:"homepage"`
	// DisplayName is the human-readable name of the application.
	DisplayName string `pulumi:"displayName"`
}

// GetTokenInfoOutput defines the output of the GetTokenInfo function.
type GetTokenInfoOutput struct {
	// Authorization contains details about the token authorization.
	Authorization GetTokenInfoAuthorization `pulumi:"authorization"`
	// Application contains details about the application that owns this token.
	Application GetTokenInfoApplication `pulumi:"application"`
}

// Annotate adds descriptions to the GetTokenInfo function.
func (f *GetTokenInfo) Annotate(a infer.Annotator) {
	a.Describe(f, "Retrieves information about the current Webflow API token, "+
		"including authorization details, scopes, rate limits, and the authorized resources. "+
		"This is useful for validating your API token configuration and understanding what resources it can access.")
}

// Annotate adds descriptions to the GetTokenInfoInput fields.
func (i *GetTokenInfoInput) Annotate(a infer.Annotator) {
	// No inputs to describe
}

// Annotate adds descriptions to the GetTokenInfoOutput fields.
func (o *GetTokenInfoOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.Authorization, "Authorization details for the API token, including scopes and authorized resources.")
	a.Describe(&o.Application, "Application details for the token owner.")
}

// Annotate adds descriptions to the GetTokenInfoAuthorization fields.
func (auth *GetTokenInfoAuthorization) Annotate(a infer.Annotator) {
	a.Describe(&auth.ID, "The unique identifier for this authorization.")
	a.Describe(&auth.CreatedOn, "The timestamp when this authorization was created (RFC3339 format).")
	a.Describe(&auth.LastUsed, "The timestamp when this token was last used (RFC3339 format).")
	a.Describe(&auth.GrantType, "The OAuth grant type used to obtain this token (e.g., 'authorization_code').")
	a.Describe(&auth.RateLimit, "The rate limit for this token in requests per minute.")
	a.Describe(&auth.Scope, "The OAuth scopes granted to this token (space or comma separated).")
	a.Describe(&auth.AuthorizedTo, "The resources this token is authorized to access.")
}

// Annotate adds descriptions to the GetTokenInfoAuthorizedTo fields.
func (authTo *GetTokenInfoAuthorizedTo) Annotate(a infer.Annotator) {
	a.Describe(&authTo.SiteIDs, "List of site IDs this token is authorized to access.")
	a.Describe(&authTo.WorkspaceIDs, "List of workspace IDs this token is authorized to access.")
	a.Describe(&authTo.UserIDs, "List of user IDs this token is authorized to access.")
}

// Annotate adds descriptions to the GetTokenInfoApplication fields.
func (app *GetTokenInfoApplication) Annotate(a infer.Annotator) {
	a.Describe(&app.ID, "The unique identifier for the application.")
	a.Describe(&app.Description, "The application description.")
	a.Describe(&app.Homepage, "The application homepage URL.")
	a.Describe(&app.DisplayName, "The human-readable name of the application.")
}

// Invoke implements the infer.Fn interface to retrieve token information.
func (f *GetTokenInfo) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetTokenInfoInput],
) (infer.FunctionResponse[GetTokenInfoOutput], error) {
	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.FunctionResponse[GetTokenInfoOutput]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := GetTokenIntrospect(ctx, client)
	if err != nil {
		return infer.FunctionResponse[GetTokenInfoOutput]{}, fmt.Errorf("failed to get token info: %w", err)
	}

	// Convert API response to output
	output := GetTokenInfoOutput{
		Authorization: GetTokenInfoAuthorization{
			ID:        response.Authorization.ID,
			CreatedOn: response.Authorization.CreatedOn,
			LastUsed:  response.Authorization.LastUsed,
			GrantType: response.Authorization.GrantType,
			RateLimit: response.Authorization.RateLimit,
			Scope:     response.Authorization.Scope,
			AuthorizedTo: GetTokenInfoAuthorizedTo{
				SiteIDs:      response.Authorization.AuthorizedTo.SiteIDs,
				WorkspaceIDs: response.Authorization.AuthorizedTo.WorkspaceIDs,
				UserIDs:      response.Authorization.AuthorizedTo.UserIDs,
			},
		},
		Application: GetTokenInfoApplication{
			ID:          response.Application.ID,
			Description: response.Application.Description,
			Homepage:    response.Application.Homepage,
			DisplayName: response.Application.DisplayName,
		},
	}

	// Ensure slice fields are not nil (Pulumi prefers empty slices)
	if output.Authorization.AuthorizedTo.SiteIDs == nil {
		output.Authorization.AuthorizedTo.SiteIDs = []string{}
	}
	if output.Authorization.AuthorizedTo.WorkspaceIDs == nil {
		output.Authorization.AuthorizedTo.WorkspaceIDs = []string{}
	}
	if output.Authorization.AuthorizedTo.UserIDs == nil {
		output.Authorization.AuthorizedTo.UserIDs = []string{}
	}

	return infer.FunctionResponse[GetTokenInfoOutput]{Output: output}, nil
}
