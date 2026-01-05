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

// CollectionResource is the resource controller for managing Webflow CMS collections.
// It implements the infer.CustomResource interface for full CRUD operations.
type CollectionResource struct{}

// CollectionArgs defines the input properties for the Collection resource.
type CollectionArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// DisplayName is the human-readable name of the collection.
	// Example: "Blog Posts", "Products", "Team Members"
	DisplayName string `pulumi:"displayName"`
	// SingularName is the singular form of the collection name.
	// Example: "Blog Post" for "Blog Posts", "Product" for "Products"
	SingularName string `pulumi:"singularName"`
	// Slug is the URL-friendly slug for the collection (optional).
	// If not provided, Webflow will auto-generate from displayName.
	// Example: "blog-posts", "products"
	Slug string `pulumi:"slug,optional"`
}

// CollectionState defines the output properties for the Collection resource.
// It embeds CollectionArgs to include input properties in the output.
type CollectionState struct {
	CollectionArgs
	// CreatedOn is the timestamp when the collection was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastUpdated is the timestamp when the collection was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
}

// Annotate adds descriptions and constraints to the Collection resource.
func (c *CollectionResource) Annotate(a infer.Annotator) {
	a.SetToken("index", "Collection")
	a.Describe(c, "Manages CMS collections for a Webflow site. "+
		"Collections are containers for structured content items (blog posts, products, etc.). "+
		"Note: Webflow collections do not support updates - any changes require replacement (delete + recreate).")
}

// Annotate adds descriptions to the CollectionArgs fields.
func (args *CollectionArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.DisplayName,
		"The human-readable name of the collection (e.g., 'Blog Posts', 'Products', 'Team Members'). "+
			"This name appears in the Webflow CMS interface. "+
			"Maximum length: 255 characters.")

	a.Describe(&args.SingularName,
		"The singular form of the collection name (e.g., 'Blog Post' for 'Blog Posts', 'Product' for 'Products'). "+
			"Used in the CMS UI when referring to individual items. "+
			"Maximum length: 255 characters.")

	a.Describe(&args.Slug,
		"The URL-friendly slug for the collection (optional, e.g., 'blog-posts', 'products'). "+
			"If not provided, Webflow will auto-generate a slug from the displayName. "+
			"The slug determines the URL path for collection items.")
}

// Annotate adds descriptions to the CollectionState fields.
func (state *CollectionState) Annotate(a infer.Annotator) {
	a.Describe(&state.CreatedOn,
		"The timestamp when the collection was created (RFC3339 format). "+
			"This is automatically set by Webflow and is read-only.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the collection was last updated (RFC3339 format). "+
			"This is automatically updated by Webflow and is read-only.")
}

// Diff determines what changes need to be made to the collection resource.
// Note: Webflow collections do not support updates via API, so ALL changes require replacement.
func (c *CollectionResource) Diff(
	ctx context.Context, req infer.DiffRequest[CollectionArgs, CollectionState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}
	detailedDiff := map[string]p.PropertyDiff{}

	// Check for siteId change (requires replacement)
	if req.State.SiteID != req.Inputs.SiteID {
		detailedDiff["siteId"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for displayName change (requires replacement - no update API)
	if req.State.DisplayName != req.Inputs.DisplayName {
		detailedDiff["displayName"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for singularName change (requires replacement - no update API)
	if req.State.SingularName != req.Inputs.SingularName {
		detailedDiff["singularName"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for slug change (requires replacement - no update API)
	if req.State.Slug != req.Inputs.Slug {
		detailedDiff["slug"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Only set DetailedDiff if changes were detected
	if len(detailedDiff) > 0 {
		diff.DetailedDiff = detailedDiff
	}

	return diff, nil
}

// Create creates a new collection on the Webflow site.
func (c *CollectionResource) Create(
	ctx context.Context, req infer.CreateRequest[CollectionArgs],
) (infer.CreateResponse[CollectionState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[CollectionState]{}, fmt.Errorf("validation failed for Collection resource: %w", err)
	}
	if err := ValidateCollectionDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.CreateResponse[CollectionState]{}, fmt.Errorf("validation failed for Collection resource: %w", err)
	}
	if err := ValidateSingularName(req.Inputs.SingularName); err != nil {
		return infer.CreateResponse[CollectionState]{}, fmt.Errorf("validation failed for Collection resource: %w", err)
	}

	state := CollectionState{
		CollectionArgs: req.Inputs,
		CreatedOn:      "", // Will be populated from API response
		LastUpdated:    "", // Will be populated from API response
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set preview timestamps
		now := time.Now().Format(time.RFC3339)
		state.CreatedOn = now
		state.LastUpdated = now
		// Generate a predictable ID for dry-run
		previewID := fmt.Sprintf("preview-%d", time.Now().Unix())
		return infer.CreateResponse[CollectionState]{
			ID:     GenerateCollectionResourceID(req.Inputs.SiteID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[CollectionState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostCollection(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.DisplayName, req.Inputs.SingularName, req.Inputs.Slug,
	)
	if err != nil {
		return infer.CreateResponse[CollectionState]{}, fmt.Errorf("failed to create collection: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid collection ID
	if response.ID == "" {
		return infer.CreateResponse[CollectionState]{}, errors.New(
			"webflow API returned empty collection ID; " +
				"this is unexpected and may indicate an API issue")
	}

	// Set timestamps from API response
	state.CreatedOn = response.CreatedOn
	state.LastUpdated = response.LastUpdated

	resourceID := GenerateCollectionResourceID(req.Inputs.SiteID, response.ID)

	return infer.CreateResponse[CollectionState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a collection from Webflow.
// Used for drift detection and import operations.
func (c *CollectionResource) Read(
	ctx context.Context, req infer.ReadRequest[CollectionArgs, CollectionState],
) (infer.ReadResponse[CollectionArgs, CollectionState], error) {
	// Extract siteID and collectionID from resource ID
	siteID, collectionID, err := ExtractIDsFromCollectionResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[CollectionArgs, CollectionState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[CollectionArgs, CollectionState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get collection details
	response, err := GetCollection(ctx, client, collectionID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[CollectionArgs, CollectionState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[CollectionArgs, CollectionState]{}, fmt.Errorf("failed to read collection: %w", err)
	}

	// Build current state from API response
	currentInputs := CollectionArgs{
		SiteID:       siteID,
		DisplayName:  response.DisplayName,
		SingularName: response.SingularName,
		Slug:         response.Slug,
	}
	currentState := CollectionState{
		CollectionArgs: currentInputs,
		CreatedOn:      response.CreatedOn,
		LastUpdated:    response.LastUpdated,
	}

	return infer.ReadResponse[CollectionArgs, CollectionState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update is not supported for Webflow collections.
// The Webflow API does not provide an update endpoint for collections.
// All changes require replacement (delete + recreate).
func (c *CollectionResource) Update(
	ctx context.Context, req infer.UpdateRequest[CollectionArgs, CollectionState],
) (infer.UpdateResponse[CollectionState], error) {
	// This should never be called because Diff() always returns UpdateReplace for all changes
	return infer.UpdateResponse[CollectionState]{}, errors.New(
		"collection updates are not supported by the Webflow API; " +
			"all changes require replacement (delete + recreate); " +
			"this error should not occur - please report this as a bug")
}

// Delete removes a collection from the Webflow site.
func (c *CollectionResource) Delete(
	ctx context.Context, req infer.DeleteRequest[CollectionState],
) (infer.DeleteResponse, error) {
	// Extract siteID and collectionID from resource ID
	_, collectionID, err := ExtractIDsFromCollectionResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteCollection(ctx, client, collectionID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete collection: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
