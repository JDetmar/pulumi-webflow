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

// GetAuthorizedUser is a Pulumi Function that retrieves information about the user
// who authorized the current API token.
// It calls the Webflow /token/authorized_by endpoint.
type GetAuthorizedUser struct{}

// GetAuthorizedUserInput defines the input parameters for the GetAuthorizedUser function.
// This function has no required inputs as it operates on the configured provider token.
type GetAuthorizedUserInput struct {
	// No inputs required - uses the provider's configured API token
}

// GetAuthorizedUserOutput defines the output of the GetAuthorizedUser function.
type GetAuthorizedUserOutput struct {
	// UserID is the unique identifier for the authorized user.
	UserID string `pulumi:"userId"`
	// Email is the email address of the authorized user.
	Email string `pulumi:"email"`
	// FirstName is the first name of the authorized user.
	FirstName string `pulumi:"firstName"`
	// LastName is the last name of the authorized user.
	LastName string `pulumi:"lastName"`
}

// Annotate adds descriptions to the GetAuthorizedUser function.
func (f *GetAuthorizedUser) Annotate(a infer.Annotator) {
	a.Describe(f, "Retrieves information about the user who authorized the current Webflow API token. "+
		"This is useful for auditing and understanding which user's credentials are being used. "+
		"Requires the 'authorized_user:read' scope.")
}

// Annotate adds descriptions to the GetAuthorizedUserInput fields.
func (i *GetAuthorizedUserInput) Annotate(a infer.Annotator) {
	// No inputs to describe
}

// Annotate adds descriptions to the GetAuthorizedUserOutput fields.
func (o *GetAuthorizedUserOutput) Annotate(a infer.Annotator) {
	a.Describe(&o.UserID, "The unique identifier for the authorized user.")
	a.Describe(&o.Email, "The email address of the authorized user.")
	a.Describe(&o.FirstName, "The first name of the authorized user.")
	a.Describe(&o.LastName, "The last name of the authorized user.")
}

// Invoke implements the infer.Fn interface to retrieve authorized user information.
func (f *GetAuthorizedUser) Invoke(
	ctx context.Context,
	req infer.FunctionRequest[GetAuthorizedUserInput],
) (infer.FunctionResponse[GetAuthorizedUserOutput], error) {
	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.FunctionResponse[GetAuthorizedUserOutput]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := GetAuthorizedBy(ctx, client)
	if err != nil {
		return infer.FunctionResponse[GetAuthorizedUserOutput]{}, fmt.Errorf("failed to get authorized user info: %w", err)
	}

	// Convert API response to output
	output := GetAuthorizedUserOutput{
		UserID:    response.ID,
		Email:     response.Email,
		FirstName: response.FirstName,
		LastName:  response.LastName,
	}

	return infer.FunctionResponse[GetAuthorizedUserOutput]{Output: output}, nil
}
