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

// CollectionItemResource is the resource controller for managing Webflow CMS collection items.
// It implements the infer.CustomResource interface for full CRUD operations.
type CollectionItemResource struct{}

// CollectionItemArgs defines the input properties for the CollectionItem resource.
type CollectionItemArgs struct {
	// CollectionID is the Webflow collection ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	CollectionID string `pulumi:"collectionId"`
	// FieldData is a map of field slugs to values for the collection item.
	// The field slugs must match the fields defined in the collection schema.
	// Example: {"name": "My Blog Post", "slug": "my-blog-post", "content": "Post content..."}
	FieldData map[string]interface{} `pulumi:"fieldData"`
	// IsArchived indicates whether the item is archived (optional, defaults to false).
	IsArchived *bool `pulumi:"isArchived,optional"`
	// IsDraft indicates whether the item is a draft (optional, defaults to true).
	IsDraft *bool `pulumi:"isDraft,optional"`
	// CmsLocaleID is the locale ID for localized sites (optional).
	// Example: "en-US"
	CmsLocaleID string `pulumi:"cmsLocaleId,optional"`
}

// CollectionItemState defines the output properties for the CollectionItem resource.
// It embeds CollectionItemArgs to include input properties in the output.
type CollectionItemState struct {
	CollectionItemArgs
	// ItemID is the Webflow-assigned item ID (read-only).
	ItemID string `pulumi:"itemId,optional"`
	// LastPublished is the timestamp when the item was last published (read-only).
	LastPublished string `pulumi:"lastPublished,optional"`
	// LastUpdated is the timestamp when the item was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
	// CreatedOn is the timestamp when the item was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
}

// Annotate adds descriptions and constraints to the CollectionItem resource.
func (c *CollectionItemResource) Annotate(a infer.Annotator) {
	a.SetToken("index", "CollectionItem")
	a.Describe(c, "Manages CMS collection items for a Webflow collection. "+
		"Collection items represent individual content entries (blog posts, products, etc.) "+
		"within a CMS collection. Each item has dynamic field data based on the collection schema.")
}

// Annotate adds descriptions to the CollectionItemArgs fields.
func (args *CollectionItemArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.CollectionID,
		"The Webflow collection ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find collection IDs via the Webflow API or dashboard. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.FieldData,
		"A map of field slugs to values for the collection item. "+
			"The field slugs must match the fields defined in the collection schema. "+
			"Common fields include 'name' (required), 'slug' (required, URL-friendly), "+
			"and any custom fields you've added to the collection. "+
			"Example: {\"name\": \"My Blog Post\", \"slug\": \"my-blog-post\", \"content\": \"Post content...\"}")

	a.Describe(&args.IsArchived,
		"Whether the item is archived (optional, defaults to false). "+
			"Archived items are not visible on the published site but remain in the CMS.")

	a.Describe(&args.IsDraft,
		"Whether the item is a draft (optional, defaults to true). "+
			"Draft items are not published to the live site. "+
			"Set to false to publish the item immediately upon creation.")

	a.Describe(&args.CmsLocaleID,
		"The locale ID for localized sites (optional, e.g., 'en-US'). "+
			"Only required if your site uses Webflow's localization features. "+
			"Leave empty for non-localized sites.")
}

// Annotate adds descriptions to the CollectionItemState fields.
func (state *CollectionItemState) Annotate(a infer.Annotator) {
	a.Describe(&state.ItemID,
		"The Webflow-assigned item ID (read-only). "+
			"This is automatically set by Webflow when the item is created.")

	a.Describe(&state.LastPublished,
		"The timestamp when the item was last published (RFC3339 format, read-only). "+
			"This is automatically updated by Webflow when the item is published.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the item was last updated (RFC3339 format, read-only). "+
			"This is automatically updated by Webflow whenever the item is modified.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the item was created (RFC3339 format, read-only). "+
			"This is automatically set by Webflow and is immutable.")
}

// Diff determines what changes need to be made to the collection item resource.
// collectionId change requires replacement.
// All other fields (fieldData, isArchived, isDraft, cmsLocaleId) can be updated in place.
func (c *CollectionItemResource) Diff(
	ctx context.Context, req infer.DiffRequest[CollectionItemArgs, CollectionItemState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}
	detailedDiff := map[string]p.PropertyDiff{}

	// Check for collectionId change (requires replacement)
	if req.State.CollectionID != req.Inputs.CollectionID {
		detailedDiff["collectionId"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for fieldData changes (can be updated in place)
	// Note: Deep comparison of maps is complex, so we rely on Pulumi's default diffing
	// The framework will detect changes and call Update if needed

	// Check for isArchived change (can be updated in place)
	if req.State.IsArchived != nil && req.Inputs.IsArchived != nil {
		if *req.State.IsArchived != *req.Inputs.IsArchived {
			detailedDiff["isArchived"] = p.PropertyDiff{Kind: p.Update}
			diff.HasChanges = true
		}
	} else if req.State.IsArchived != req.Inputs.IsArchived {
		// One is nil, the other is not
		detailedDiff["isArchived"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Check for isDraft change (can be updated in place)
	if req.State.IsDraft != nil && req.Inputs.IsDraft != nil {
		if *req.State.IsDraft != *req.Inputs.IsDraft {
			detailedDiff["isDraft"] = p.PropertyDiff{Kind: p.Update}
			diff.HasChanges = true
		}
	} else if req.State.IsDraft != req.Inputs.IsDraft {
		// One is nil, the other is not
		detailedDiff["isDraft"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Check for cmsLocaleId change (can be updated in place)
	if req.State.CmsLocaleID != req.Inputs.CmsLocaleID {
		detailedDiff["cmsLocaleId"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Only set DetailedDiff if changes were detected
	if len(detailedDiff) > 0 {
		diff.DetailedDiff = detailedDiff
	}

	return diff, nil
}

// Create creates a new collection item in the Webflow collection.
func (c *CollectionItemResource) Create(
	ctx context.Context, req infer.CreateRequest[CollectionItemArgs],
) (infer.CreateResponse[CollectionItemState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateCollectionID(req.Inputs.CollectionID); err != nil {
		return infer.CreateResponse[CollectionItemState]{}, fmt.Errorf(
			"validation failed for CollectionItem resource: %w", err)
	}
	if err := ValidateFieldData(req.Inputs.FieldData); err != nil {
		return infer.CreateResponse[CollectionItemState]{}, fmt.Errorf(
			"validation failed for CollectionItem resource: %w", err)
	}

	state := CollectionItemState{
		CollectionItemArgs: req.Inputs,
		ItemID:             "", // Will be populated from API response
		LastPublished:      "", // Will be populated from API response
		LastUpdated:        "", // Will be populated from API response
		CreatedOn:          "", // Will be populated from API response
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set preview timestamps
		now := time.Now().Format(time.RFC3339)
		state.CreatedOn = now
		state.LastUpdated = now
		// Generate a predictable ID for dry-run
		previewID := fmt.Sprintf("preview-%d", time.Now().Unix())
		state.ItemID = previewID
		return infer.CreateResponse[CollectionItemState]{
			ID:     GenerateCollectionItemResourceID(req.Inputs.CollectionID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[CollectionItemState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostCollectionItem(
		ctx, client, req.Inputs.CollectionID, req.Inputs.FieldData,
		req.Inputs.IsArchived, req.Inputs.IsDraft, req.Inputs.CmsLocaleID,
	)
	if err != nil {
		return infer.CreateResponse[CollectionItemState]{}, fmt.Errorf("failed to create collection item: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid item ID
	if response.ID == "" {
		return infer.CreateResponse[CollectionItemState]{}, errors.New(
			"Webflow API returned empty item ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Set timestamps and ID from API response
	state.ItemID = response.ID
	state.CreatedOn = response.CreatedOn
	state.LastUpdated = response.LastUpdated
	state.LastPublished = response.LastPublished

	resourceID := GenerateCollectionItemResourceID(req.Inputs.CollectionID, response.ID)

	return infer.CreateResponse[CollectionItemState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a collection item from Webflow.
// Used for drift detection and import operations.
func (c *CollectionItemResource) Read(
	ctx context.Context, req infer.ReadRequest[CollectionItemArgs, CollectionItemState],
) (infer.ReadResponse[CollectionItemArgs, CollectionItemState], error) {
	// Extract collectionID and itemID from resource ID
	collectionID, itemID, err := ExtractIDsFromCollectionItemResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[CollectionItemArgs, CollectionItemState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[CollectionItemArgs, CollectionItemState]{}, fmt.Errorf(
			"failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get collection item details
	response, err := GetCollectionItem(ctx, client, collectionID, itemID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[CollectionItemArgs, CollectionItemState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[CollectionItemArgs, CollectionItemState]{}, fmt.Errorf(
			"failed to read collection item: %w", err)
	}

	// Build current state from API response
	currentInputs := CollectionItemArgs{
		CollectionID: collectionID,
		FieldData:    response.FieldData,
		CmsLocaleID:  response.CmsLocaleID,
	}

	// Handle pointer fields - always create pointers for optional fields
	isArchived := response.IsArchived
	currentInputs.IsArchived = &isArchived
	isDraft := response.IsDraft
	currentInputs.IsDraft = &isDraft

	currentState := CollectionItemState{
		CollectionItemArgs: currentInputs,
		ItemID:             response.ID,
		CreatedOn:          response.CreatedOn,
		LastUpdated:        response.LastUpdated,
		LastPublished:      response.LastPublished,
	}

	return infer.ReadResponse[CollectionItemArgs, CollectionItemState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing collection item.
func (c *CollectionItemResource) Update(
	ctx context.Context, req infer.UpdateRequest[CollectionItemArgs, CollectionItemState],
) (infer.UpdateResponse[CollectionItemState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateCollectionID(req.Inputs.CollectionID); err != nil {
		return infer.UpdateResponse[CollectionItemState]{}, fmt.Errorf(
			"validation failed for CollectionItem resource: %w", err)
	}
	if err := ValidateFieldData(req.Inputs.FieldData); err != nil {
		return infer.UpdateResponse[CollectionItemState]{}, fmt.Errorf(
			"validation failed for CollectionItem resource: %w", err)
	}

	state := CollectionItemState{
		CollectionItemArgs: req.Inputs,
		ItemID:             req.State.ItemID,        // Preserve from current state
		CreatedOn:          req.State.CreatedOn,     // Preserve from current state
		LastUpdated:        "",                      // Will be updated by API
		LastPublished:      req.State.LastPublished, // Preserve from current state
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		state.LastUpdated = time.Now().Format(time.RFC3339)
		return infer.UpdateResponse[CollectionItemState]{
			Output: state,
		}, nil
	}

	// Extract the Webflow item ID from the Pulumi resource ID
	collectionID, itemID, err := ExtractIDsFromCollectionItemResourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[CollectionItemState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[CollectionItemState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PatchCollectionItem(
		ctx, client, collectionID, itemID, req.Inputs.FieldData,
		req.Inputs.IsArchived, req.Inputs.IsDraft, req.Inputs.CmsLocaleID,
	)
	if err != nil {
		return infer.UpdateResponse[CollectionItemState]{}, fmt.Errorf("failed to update collection item: %w", err)
	}

	// Update timestamps from API response
	state.LastUpdated = response.LastUpdated
	state.LastPublished = response.LastPublished

	return infer.UpdateResponse[CollectionItemState]{
		Output: state,
	}, nil
}

// Delete removes a collection item from the Webflow collection.
func (c *CollectionItemResource) Delete(
	ctx context.Context, req infer.DeleteRequest[CollectionItemState],
) (infer.DeleteResponse, error) {
	// Extract collectionID and itemID from resource ID
	collectionID, itemID, err := ExtractIDsFromCollectionItemResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteCollectionItem(ctx, client, collectionID, itemID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete collection item: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
