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

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// CollectionField is the resource controller for managing Webflow CMS collection fields.
// It implements the infer.CustomResource interface for full CRUD operations.
type CollectionField struct{}

// CollectionFieldArgs defines the input properties for the CollectionField resource.
type CollectionFieldArgs struct {
	// CollectionID is the Webflow collection ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	CollectionID string `pulumi:"collectionId"`
	// Type is the field type (PlainText, RichText, Image, etc.).
	// Cannot be changed after creation - requires replacement.
	Type string `pulumi:"type"`
	// DisplayName is the human-readable name of the field.
	// Example: "Title", "Description", "Author"
	DisplayName string `pulumi:"displayName"`
	// Slug is the URL-friendly slug for the field (optional).
	// If not provided, Webflow will auto-generate from displayName.
	// Example: "title", "description"
	Slug string `pulumi:"slug,optional"`
	// IsRequired indicates whether the field is required (optional, defaults to false).
	IsRequired bool `pulumi:"isRequired,optional"`
	// HelpText is optional help text shown in the CMS interface.
	HelpText string `pulumi:"helpText,optional"`
	// Validations contains type-specific validation rules (optional).
	// Example for Number: {"min": 0, "max": 100}
	Validations map[string]interface{} `pulumi:"validations,optional"`
}

// CollectionFieldState defines the output properties for the CollectionField resource.
// It embeds CollectionFieldArgs to include input properties in the output.
type CollectionFieldState struct {
	CollectionFieldArgs
	// FieldID is the Webflow-assigned field ID (read-only).
	FieldID string `pulumi:"fieldId,optional"`
	// IsEditable indicates whether the field can be edited (read-only).
	IsEditable bool `pulumi:"isEditable,optional"`
}

// Annotate adds descriptions and constraints to the CollectionField resource.
func (f *CollectionField) Annotate(a infer.Annotator) {
	a.SetToken("index", "CollectionField")
	a.Describe(f, "Manages fields for a Webflow CMS collection. "+
		"Collection fields define the structure of content items in a collection. "+
		"Note: The field type cannot be changed after creation - changing it requires replacement (delete + recreate).")
}

// Annotate adds descriptions to the CollectionFieldArgs fields.
func (args *CollectionFieldArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.CollectionID,
		"The Webflow collection ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find collection IDs via the Webflow API or dashboard. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.Type,
		"The field type (e.g., 'PlainText', 'RichText', 'Image', 'Number'). "+
			"Supported types: PlainText, RichText, Image, MultiImage, Video, Link, Email, Phone, "+
			"Number, DateTime, Switch, Color, Option, File, Reference, MultiReference. "+
			"IMPORTANT: Cannot be changed after creation - changing this requires replacement.")

	a.Describe(&args.DisplayName,
		"The human-readable name of the field (e.g., 'Title', 'Description', 'Author'). "+
			"This name appears in the Webflow CMS interface. "+
			"Maximum length: 255 characters.")

	a.Describe(&args.Slug,
		"The URL-friendly slug for the field (optional, e.g., 'title', 'description'). "+
			"If not provided, Webflow will auto-generate a slug from the displayName. "+
			"The slug is used in API requests and exports.")

	a.Describe(&args.IsRequired,
		"Whether the field is required (optional, defaults to false). "+
			"When true, content items must provide a value for this field.")

	a.Describe(&args.HelpText,
		"Optional help text shown in the CMS interface (e.g., 'Enter the article title'). "+
			"Helps content editors understand what to enter in this field.")

	a.Describe(&args.Validations,
		"Type-specific validation rules (optional). "+
			"Different field types support different validations. "+
			"Example for Number type: {\"min\": 0, \"max\": 100}. "+
			"Example for PlainText type: {\"maxLength\": 500}. "+
			"Refer to Webflow API documentation for validation options for each field type.")
}

// Annotate adds descriptions to the CollectionFieldState fields.
func (state *CollectionFieldState) Annotate(a infer.Annotator) {
	a.Describe(&state.FieldID,
		"The Webflow-assigned field ID (read-only). "+
			"This ID is automatically generated when the field is created.")

	a.Describe(&state.IsEditable,
		"Whether the field can be edited (read-only). "+
			"System fields may not be editable.")
}

// Diff determines what changes need to be made to the collection field resource.
// CollectionID and Type changes trigger replacement (cannot be changed).
// Other fields can be updated in-place.
func (f *CollectionField) Diff(
	ctx context.Context, req infer.DiffRequest[CollectionFieldArgs, CollectionFieldState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}
	detailedDiff := map[string]p.PropertyDiff{}

	// Check for collectionId change (requires replacement)
	if req.State.CollectionID != req.Inputs.CollectionID {
		detailedDiff["collectionId"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for type change (requires replacement - cannot change field type)
	if req.State.Type != req.Inputs.Type {
		detailedDiff["type"] = p.PropertyDiff{Kind: p.UpdateReplace}
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
	}

	// Check for displayName change (can be updated in-place)
	if req.State.DisplayName != req.Inputs.DisplayName {
		detailedDiff["displayName"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Check for slug change (can be updated in-place)
	if req.State.Slug != req.Inputs.Slug {
		detailedDiff["slug"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Check for isRequired change (can be updated in-place)
	if req.State.IsRequired != req.Inputs.IsRequired {
		detailedDiff["isRequired"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Check for helpText change (can be updated in-place)
	if req.State.HelpText != req.Inputs.HelpText {
		detailedDiff["helpText"] = p.PropertyDiff{Kind: p.Update}
		diff.HasChanges = true
	}

	// Note: validations comparison is simplified here.
	// In a production system, you might want deep comparison of the map.
	// For now, we assume if the map reference changed, it needs updating.

	// Only set DetailedDiff if changes were detected
	if len(detailedDiff) > 0 {
		diff.DetailedDiff = detailedDiff
	}

	return diff, nil
}

// Create creates a new field for a Webflow collection.
func (f *CollectionField) Create(
	ctx context.Context, req infer.CreateRequest[CollectionFieldArgs],
) (infer.CreateResponse[CollectionFieldState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateCollectionID(req.Inputs.CollectionID); err != nil {
		return infer.CreateResponse[CollectionFieldState]{}, fmt.Errorf(
			"validation failed for CollectionField resource: %w", err)
	}
	if err := ValidateFieldType(req.Inputs.Type); err != nil {
		return infer.CreateResponse[CollectionFieldState]{}, fmt.Errorf(
			"validation failed for CollectionField resource: %w", err)
	}
	if err := ValidateFieldDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.CreateResponse[CollectionFieldState]{}, fmt.Errorf(
			"validation failed for CollectionField resource: %w", err)
	}

	state := CollectionFieldState{
		CollectionFieldArgs: req.Inputs,
		FieldID:             "", // Will be populated from API response
		IsEditable:          true,
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Generate a predictable ID for dry-run
		previewID := "preview-field"
		return infer.CreateResponse[CollectionFieldState]{
			ID:     GenerateCollectionFieldResourceID(req.Inputs.CollectionID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[CollectionFieldState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostCollectionField(
		ctx, client, req.Inputs.CollectionID,
		req.Inputs.Type, req.Inputs.DisplayName, req.Inputs.Slug,
		req.Inputs.HelpText, req.Inputs.IsRequired, req.Inputs.Validations,
	)
	if err != nil {
		return infer.CreateResponse[CollectionFieldState]{}, fmt.Errorf("failed to create collection field: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid field ID
	if response.ID == "" {
		return infer.CreateResponse[CollectionFieldState]{}, errors.New(
			"Webflow API returned empty field ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Update state with API response
	state.FieldID = response.ID
	state.IsEditable = response.IsEditable

	resourceID := GenerateCollectionFieldResourceID(req.Inputs.CollectionID, response.ID)

	return infer.CreateResponse[CollectionFieldState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a collection field from Webflow.
// Used for drift detection and import operations.
func (f *CollectionField) Read(
	ctx context.Context, req infer.ReadRequest[CollectionFieldArgs, CollectionFieldState],
) (infer.ReadResponse[CollectionFieldArgs, CollectionFieldState], error) {
	// Extract collectionID and fieldID from resource ID
	collectionID, fieldID, err := ExtractIDsFromCollectionFieldResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[CollectionFieldArgs, CollectionFieldState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[CollectionFieldArgs, CollectionFieldState]{}, fmt.Errorf(
			"failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get field details
	response, err := GetCollectionField(ctx, client, collectionID, fieldID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[CollectionFieldArgs, CollectionFieldState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[CollectionFieldArgs, CollectionFieldState]{}, fmt.Errorf(
			"failed to read collection field: %w", err)
	}

	// Build current state from API response
	currentInputs := CollectionFieldArgs{
		CollectionID: collectionID,
		Type:         response.Type,
		DisplayName:  response.DisplayName,
		Slug:         response.Slug,
		IsRequired:   response.IsRequired,
		HelpText:     response.HelpText,
		Validations:  response.Validations,
	}
	currentState := CollectionFieldState{
		CollectionFieldArgs: currentInputs,
		FieldID:             response.ID,
		IsEditable:          response.IsEditable,
	}

	return infer.ReadResponse[CollectionFieldArgs, CollectionFieldState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing collection field.
func (f *CollectionField) Update(
	ctx context.Context, req infer.UpdateRequest[CollectionFieldArgs, CollectionFieldState],
) (infer.UpdateResponse[CollectionFieldState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateCollectionID(req.Inputs.CollectionID); err != nil {
		return infer.UpdateResponse[CollectionFieldState]{}, fmt.Errorf(
			"validation failed for CollectionField resource: %w", err)
	}
	if err := ValidateFieldDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.UpdateResponse[CollectionFieldState]{}, fmt.Errorf(
			"validation failed for CollectionField resource: %w", err)
	}

	state := CollectionFieldState{
		CollectionFieldArgs: req.Inputs,
		FieldID:             req.State.FieldID,     // Preserve field ID
		IsEditable:          req.State.IsEditable, // Preserve editability flag
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		return infer.UpdateResponse[CollectionFieldState]{
			Output: state,
		}, nil
	}

	// Extract the Webflow field ID from the Pulumi resource ID
	_, fieldID, err := ExtractIDsFromCollectionFieldResourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[CollectionFieldState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[CollectionFieldState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PutCollectionField(
		ctx, client, req.Inputs.CollectionID, fieldID,
		req.Inputs.DisplayName, req.Inputs.Slug, req.Inputs.HelpText,
		req.Inputs.IsRequired, req.Inputs.Validations,
	)
	if err != nil {
		return infer.UpdateResponse[CollectionFieldState]{}, fmt.Errorf("failed to update collection field: %w", err)
	}

	// Update state with API response
	state.IsEditable = response.IsEditable

	return infer.UpdateResponse[CollectionFieldState]{
		Output: state,
	}, nil
}

// Delete removes a field from a Webflow collection.
func (f *CollectionField) Delete(
	ctx context.Context, req infer.DeleteRequest[CollectionFieldState],
) (infer.DeleteResponse, error) {
	// Extract collectionID and fieldID from resource ID
	collectionID, fieldID, err := ExtractIDsFromCollectionFieldResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteCollectionField(ctx, client, collectionID, fieldID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete collection field: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
