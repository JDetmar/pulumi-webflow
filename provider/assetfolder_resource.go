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

// AssetFolder is the resource controller for managing Webflow asset folders.
// It implements the infer.CustomResource interface for CRUD operations.
// Note: The Webflow API does not support delete or update operations for asset folders.
// Deleting this resource will only remove it from Pulumi state, not from Webflow.
type AssetFolder struct{}

// AssetFolderArgs defines the input properties for the AssetFolder resource.
type AssetFolderArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// DisplayName is the human-readable name for the asset folder.
	// This name appears in the Webflow Assets panel.
	// Examples: "Images", "Documents", "Icons"
	DisplayName string `pulumi:"displayName"`
	// ParentFolder is the optional ID of the parent folder.
	// If not specified, the folder will be created at the root level.
	// Example: "5f0c8c9e1c9d440000e8d8c4"
	ParentFolder string `pulumi:"parentFolder,optional"`
}

// AssetFolderState defines the output properties for the AssetFolder resource.
// It embeds AssetFolderArgs to include input properties in the output.
type AssetFolderState struct {
	AssetFolderArgs
	// FolderID is the Webflow-assigned folder ID (read-only).
	FolderID string `pulumi:"folderId,optional"`
	// Assets is the list of asset IDs contained in this folder (read-only).
	Assets []string `pulumi:"assets,optional"`
	// CreatedOn is the timestamp when the folder was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastUpdated is the timestamp when the folder was last modified (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
}

// Annotate adds descriptions and constraints to the AssetFolder resource.
func (r *AssetFolder) Annotate(a infer.Annotator) {
	a.SetToken("index", "AssetFolder")
	a.Describe(r, "Manages asset folders for organizing files in a Webflow site. "+
		"This resource allows you to create folders to organize your assets (images, documents, etc.) "+
		"in the Webflow Assets panel. "+
		"NOTE: The Webflow API does not support deleting or updating asset folders. "+
		"Deleting this resource will only remove it from Pulumi state, not from Webflow. "+
		"Any changes to folder properties will require creating a new folder.")
}

// Annotate adds descriptions to the AssetFolderArgs fields.
func (args *AssetFolderArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.DisplayName,
		"The human-readable name for the asset folder. "+
			"This name appears in the Webflow Assets panel and helps organize your files. "+
			"Examples: 'Images', 'Documents', 'Icons', 'Hero Backgrounds'. "+
			"Maximum length: 255 characters.")

	a.Describe(&args.ParentFolder,
		"Optional ID of the parent folder for creating nested folder structures. "+
			"If not specified, the folder will be created at the root level of the Assets panel. "+
			"Example: '5f0c8c9e1c9d440000e8d8c4'.")
}

// Annotate adds descriptions to the AssetFolderState fields.
func (state *AssetFolderState) Annotate(a infer.Annotator) {
	a.Describe(&state.FolderID,
		"The Webflow-assigned folder ID (read-only). "+
			"This unique identifier can be used to reference the folder in other resources, "+
			"such as when uploading assets to this folder.")

	a.Describe(&state.Assets,
		"List of asset IDs currently contained in this folder (read-only). "+
			"This is automatically populated by Webflow when assets are added to the folder.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the folder was created (RFC3339 format, read-only). "+
			"This is automatically set when the folder is created.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the folder was last modified (RFC3339 format, read-only). "+
			"This is updated when assets are added or removed from the folder.")
}

// Diff determines what changes need to be made to the asset folder resource.
// Since Webflow doesn't support updates, all changes require replacement.
func (r *AssetFolder) Diff(
	ctx context.Context, req infer.DiffRequest[AssetFolderArgs, AssetFolderState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}

	// SiteID change requires replacement
	if req.State.SiteID != req.Inputs.SiteID {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"siteId": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// DisplayName change requires replacement (no update API available)
	if req.State.DisplayName != req.Inputs.DisplayName {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"displayName": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// ParentFolder change requires replacement (no update API available)
	if req.State.ParentFolder != req.Inputs.ParentFolder {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"parentFolder": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	return diff, nil
}

// Create creates a new asset folder in the Webflow site.
func (r *AssetFolder) Create(
	ctx context.Context, req infer.CreateRequest[AssetFolderArgs],
) (infer.CreateResponse[AssetFolderState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[AssetFolderState]{}, fmt.Errorf("validation failed for AssetFolder resource: %w", err)
	}
	if err := ValidateDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.CreateResponse[AssetFolderState]{}, fmt.Errorf("validation failed for AssetFolder resource: %w", err)
	}
	// Validate parent folder ID if provided
	if req.Inputs.ParentFolder != "" {
		if err := ValidateAssetFolderID(req.Inputs.ParentFolder); err != nil {
			return infer.CreateResponse[AssetFolderState]{},
				fmt.Errorf("validation failed for AssetFolder resource (parentFolder): %w", err)
		}
	}

	state := AssetFolderState{
		AssetFolderArgs: req.Inputs,
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set preview values
		state.FolderID = fmt.Sprintf("preview-%d", time.Now().Unix())
		state.Assets = []string{}
		state.CreatedOn = time.Now().Format(time.RFC3339)
		state.LastUpdated = state.CreatedOn
		previewID := GenerateAssetFolderResourceID(req.Inputs.SiteID, state.FolderID)
		return infer.CreateResponse[AssetFolderState]{
			ID:     previewID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[AssetFolderState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to create the folder
	folder, err := PostAssetFolder(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.DisplayName, req.Inputs.ParentFolder,
	)
	if err != nil {
		return infer.CreateResponse[AssetFolderState]{}, fmt.Errorf("failed to create asset folder: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid folder ID
	if folder.ID == "" {
		return infer.CreateResponse[AssetFolderState]{}, errors.New(
			"webflow API returned empty folder ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Populate state from API response
	state.FolderID = folder.ID
	state.Assets = folder.Assets
	state.CreatedOn = folder.CreatedOn
	state.LastUpdated = folder.LastUpdated

	resourceID := GenerateAssetFolderResourceID(req.Inputs.SiteID, folder.ID)

	return infer.CreateResponse[AssetFolderState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of an asset folder from Webflow.
// Used for drift detection and import operations.
func (r *AssetFolder) Read(
	ctx context.Context, req infer.ReadRequest[AssetFolderArgs, AssetFolderState],
) (infer.ReadResponse[AssetFolderArgs, AssetFolderState], error) {
	// Extract siteID and folderID from resource ID
	siteID, folderID, err := ExtractIDsFromAssetFolderResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[AssetFolderArgs, AssetFolderState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[AssetFolderArgs, AssetFolderState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get folder details
	folder, err := GetAssetFolder(ctx, client, folderID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[AssetFolderArgs, AssetFolderState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[AssetFolderArgs, AssetFolderState]{}, fmt.Errorf("failed to read asset folder: %w", err)
	}

	// Build current state from API response
	currentInputs := AssetFolderArgs{
		SiteID:       siteID,
		DisplayName:  folder.DisplayName,
		ParentFolder: folder.ParentFolder,
	}
	currentState := AssetFolderState{
		AssetFolderArgs: currentInputs,
		FolderID:        folder.ID,
		Assets:          folder.Assets,
		CreatedOn:       folder.CreatedOn,
		LastUpdated:     folder.LastUpdated,
	}

	return infer.ReadResponse[AssetFolderArgs, AssetFolderState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update is not supported for asset folders - the Webflow API doesn't have an update endpoint.
// Any changes will trigger a replacement (delete + create).
func (r *AssetFolder) Update(
	ctx context.Context, req infer.UpdateRequest[AssetFolderArgs, AssetFolderState],
) (infer.UpdateResponse[AssetFolderState], error) {
	// This should never be called because Diff always returns UpdateReplace
	// But implement it defensively
	return infer.UpdateResponse[AssetFolderState]{}, errors.New(
		"asset folders cannot be updated in place - the Webflow API does not support folder updates. " +
			"Any changes will trigger a replacement (create new folder, then remove old from state). " +
			"Note: The old folder will remain in Webflow as the API does not support deletion")
}

// Delete removes the asset folder from Pulumi state.
// NOTE: The Webflow API does not support deleting asset folders, so the folder
// will remain in Webflow even after this resource is destroyed.
func (r *AssetFolder) Delete(
	ctx context.Context, req infer.DeleteRequest[AssetFolderState],
) (infer.DeleteResponse, error) {
	// The Webflow API does not support deleting asset folders.
	// We can only remove the resource from Pulumi state.
	// Log a warning to inform the user about the limitation.
	NewLogContext(ctx).
		WithField("siteId", req.State.SiteID).
		WithField("folderName", req.State.DisplayName).
		Warn("Asset folder cannot be deleted via API - removing from Pulumi state only. " +
			"The folder will remain in Webflow and must be manually deleted from the dashboard if needed")

	return infer.DeleteResponse{}, nil
}
