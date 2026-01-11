// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"errors"
	"fmt"
	"strings"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Redirect is the resource controller for managing Webflow redirects.
// It implements the infer.CustomResource interface for full CRUD operations.
type Redirect struct{}

// RedirectArgs defines the input properties for the Redirect resource.
type RedirectArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// SourcePath is the URL path to redirect from (e.g., "/old-page").
	// Must start with "/" and contain only valid URL characters.
	// Examples: "/old-page", "/blog/2023", "/products/item-1"
	SourcePath string `pulumi:"sourcePath"`
	// DestinationPath is the URL path to redirect to (e.g., "/new-page").
	// Must start with "/" and contain only valid URL characters.
	// Examples: "/new-page", "/home", "/products/item-1"
	DestinationPath string `pulumi:"destinationPath"`
	// StatusCode is the HTTP status code for the redirect (301 or 302).
	// 301 = permanent redirect (for pages moved permanently)
	// 302 = temporary redirect (for temporary page moves or maintenance)
	StatusCode int `pulumi:"statusCode"`
}

// RedirectState defines the output properties for the Redirect resource.
// It embeds RedirectArgs to include input properties in the output.
type RedirectState struct {
	RedirectArgs
	// CreatedOn is the timestamp when the redirect was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
}

// Annotate adds descriptions and constraints to the Redirect resource.
func (r *Redirect) Annotate(a infer.Annotator) {
	a.SetToken("index", "Redirect")
	a.Describe(r, "Manages HTTP redirects for a Webflow site. "+
		"This resource allows you to define redirect rules for old URLs to new locations, "+
		"supporting both permanent (301) and temporary (302) redirects.")
}

// Annotate adds descriptions to the RedirectArgs fields.
func (args *RedirectArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.SourcePath,
		"The URL path to redirect from (e.g., '/old-page', '/blog/2023'). "+
			"Must start with '/' and contain only valid URL characters "+
			"(letters, numbers, hyphens, underscores, slashes, dots). "+
			"Query strings and fragments are not allowed in the source path.")

	a.Describe(&args.DestinationPath,
		"The URL path to redirect to (e.g., '/new-page', '/home'). "+
			"Must start with '/' and contain only valid URL characters. "+
			"This is the location where users will be redirected when they visit the source path.")

	a.Describe(&args.StatusCode,
		"The HTTP status code for the redirect. Must be either 301 or 302. "+
			"301 = permanent redirect (use when a page has moved permanently; search engines update their index). "+
			"302 = temporary redirect (use for maintenance or temporary page moves).")
}

// Annotate adds descriptions to the RedirectState fields.
func (state *RedirectState) Annotate(a infer.Annotator) {
	a.Describe(&state.CreatedOn,
		"The timestamp when the redirect was created (RFC3339 format). "+
			"This is automatically set when the redirect is created and is read-only.")
}

// Diff determines what changes need to be made to the redirect resource.
// siteId and sourcePath changes trigger replacement (primary key).
// destinationPath and statusCode changes trigger in-place update.
func (r *Redirect) Diff(
	ctx context.Context, req infer.DiffRequest[RedirectArgs, RedirectState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}

	// Check for siteId change (requires replacement)
	if req.State.SiteID != req.Inputs.SiteID {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"siteId": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for sourcePath change (requires replacement - it's the primary key)
	if req.State.SourcePath != req.Inputs.SourcePath {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"sourcePath": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// NOTE: Due to a Webflow API limitation, the PATCH endpoint returns a 409 conflict error
	// when updating redirects, even for valid updates. This appears to be a bug in the Webflow API
	// where it checks for source path uniqueness but doesn't exclude the redirect being updated.
	// Therefore, ALL changes require replacement (delete + recreate) instead of in-place update.

	// Check for destinationPath change (requires replacement due to Webflow API limitation)
	if req.State.DestinationPath != req.Inputs.DestinationPath {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"destinationPath": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for statusCode change (requires replacement due to Webflow API limitation)
	// Only report change if state has valid statusCode that differs from inputs.
	// The Webflow API list endpoint doesn't return statusCode, so it comes back as 0.
	if req.State.StatusCode != 0 && req.State.StatusCode != req.Inputs.StatusCode {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"statusCode": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	return diff, nil
}

// Create creates a new redirect on the Webflow site.
func (r *Redirect) Create(
	ctx context.Context, req infer.CreateRequest[RedirectArgs],
) (infer.CreateResponse[RedirectState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}
	if err := ValidateSourcePath(req.Inputs.SourcePath); err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}
	if err := ValidateDestinationPath(req.Inputs.DestinationPath); err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}
	if err := ValidateStatusCode(req.Inputs.StatusCode); err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}

	state := RedirectState{
		RedirectArgs: req.Inputs,
		CreatedOn:    "", // Will be populated after creation
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set a preview timestamp
		state.CreatedOn = time.Now().Format(time.RFC3339)
		// Generate a predictable ID for dry-run
		previewID := fmt.Sprintf("preview-%d", time.Now().Unix())
		return infer.CreateResponse[RedirectState]{
			ID:     GenerateRedirectResourceID(req.Inputs.SiteID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostRedirect(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.SourcePath, req.Inputs.DestinationPath, req.Inputs.StatusCode,
	)
	if err != nil {
		return infer.CreateResponse[RedirectState]{}, fmt.Errorf("failed to create redirect: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid redirect ID
	if response.ID == "" {
		return infer.CreateResponse[RedirectState]{}, errors.New(
			"webflow API returned empty redirect ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Set creation timestamp
	state.CreatedOn = time.Now().Format(time.RFC3339)

	resourceID := GenerateRedirectResourceID(req.Inputs.SiteID, response.ID)

	return infer.CreateResponse[RedirectState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a redirect from Webflow.
// Used for drift detection and import operations.
func (r *Redirect) Read(
	ctx context.Context, req infer.ReadRequest[RedirectArgs, RedirectState],
) (infer.ReadResponse[RedirectArgs, RedirectState], error) {
	// Extract siteID and redirectID from resource ID
	siteID, redirectID, err := ExtractIDsFromRedirectResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[RedirectArgs, RedirectState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[RedirectArgs, RedirectState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get all redirects for this site
	response, err := GetRedirects(ctx, client, siteID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[RedirectArgs, RedirectState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[RedirectArgs, RedirectState]{}, fmt.Errorf("failed to read redirects: %w", err)
	}

	// Find the specific redirect in the list
	var foundRedirect *RedirectRule
	for _, redirect := range response.Redirects {
		if redirect.ID == redirectID {
			foundRedirect = &redirect
			break
		}
	}

	// If redirect not found, return empty ID to signal deletion
	if foundRedirect == nil {
		return infer.ReadResponse[RedirectArgs, RedirectState]{
			ID: "",
		}, nil
	}

	// Build current state from API response
	currentInputs := RedirectArgs{
		SiteID:          siteID,
		SourcePath:      foundRedirect.SourcePath,
		DestinationPath: foundRedirect.DestinationPath,
		StatusCode:      foundRedirect.StatusCode,
	}
	currentState := RedirectState{
		RedirectArgs: currentInputs,
		CreatedOn:    req.State.CreatedOn, // Preserve the creation timestamp from existing state
	}

	return infer.ReadResponse[RedirectArgs, RedirectState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing redirect.
func (r *Redirect) Update(
	ctx context.Context, req infer.UpdateRequest[RedirectArgs, RedirectState],
) (infer.UpdateResponse[RedirectState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}
	if err := ValidateDestinationPath(req.Inputs.DestinationPath); err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}
	if err := ValidateStatusCode(req.Inputs.StatusCode); err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("validation failed for Redirect resource: %w", err)
	}

	state := RedirectState{
		RedirectArgs: req.Inputs,
		CreatedOn:    req.State.CreatedOn, // Preserve the creation timestamp from current state
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		return infer.UpdateResponse[RedirectState]{
			Output: state,
		}, nil
	}

	// Extract the Webflow redirect ID from the Pulumi resource ID
	_, redirectID, err := ExtractIDsFromRedirectResourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	_, err = PatchRedirect(
		ctx, client, req.Inputs.SiteID, redirectID,
		req.Inputs.SourcePath, req.Inputs.DestinationPath, req.Inputs.StatusCode,
	)
	if err != nil {
		return infer.UpdateResponse[RedirectState]{}, fmt.Errorf("failed to update redirect: %w", err)
	}

	return infer.UpdateResponse[RedirectState]{
		Output: state,
	}, nil
}

// Delete removes a redirect from the Webflow site.
func (r *Redirect) Delete(ctx context.Context, req infer.DeleteRequest[RedirectState]) (infer.DeleteResponse, error) {
	// Extract siteID and redirectID from resource ID
	siteID, redirectID, err := ExtractIDsFromRedirectResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteRedirect(ctx, client, siteID, redirectID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete redirect: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
