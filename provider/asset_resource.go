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
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// Asset is the resource controller for managing Webflow assets.
// It implements the infer.CustomResource interface for CRUD operations.
// Note: Assets are immutable - updates require replacement.
type Asset struct{}

// AssetArgs defines the input properties for the Asset resource.
type AssetArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// FileName is the name of the file to upload.
	// Must include the file extension (e.g., "logo.png", "hero.jpg").
	// Examples: "logo.png", "hero-image.jpg", "document.pdf"
	FileName string `pulumi:"fileName"`
	// FileHash is the optional MD5 hash of the file for deduplication.
	// If a file with the same hash already exists, Webflow may reuse it.
	// Optional - leave empty to always upload a new file.
	FileHash string `pulumi:"fileHash,optional"`
	// ParentFolder is the optional folder ID where the asset will be placed.
	// If not specified, the asset will be placed at the root level.
	// Example: "5f0c8c9e1c9d440000e8d8c4"
	ParentFolder string `pulumi:"parentFolder,optional"`
	// FileSource is the source of the file to upload.
	// For MVP, this tracks that an asset exists but doesn't handle actual upload.
	// Future versions may support URL or local file path.
	// Example: "https://example.com/logo.png" or "/path/to/local/file.png"
	FileSource string `pulumi:"fileSource,optional"`
}

// AssetState defines the output properties for the Asset resource.
// It embeds AssetArgs to include input properties in the output.
type AssetState struct {
	AssetArgs
	// AssetID is the Webflow-assigned asset ID (read-only).
	AssetID string `pulumi:"assetId,optional"`
	// HostedURL is the CDN URL where the asset is hosted (read-only).
	// This is the URL you can use to reference the asset in your site.
	HostedURL string `pulumi:"hostedUrl,optional"`
	// ContentType is the MIME type of the asset (read-only).
	// Examples: "image/png", "image/jpeg", "application/pdf"
	ContentType string `pulumi:"contentType,optional"`
	// Size is the size of the asset in bytes (read-only).
	Size int `pulumi:"size,optional"`
	// CreatedOn is the timestamp when the asset was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastUpdated is the timestamp when the asset was last modified (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
}

// Annotate adds descriptions and constraints to the Asset resource.
func (r *Asset) Annotate(a infer.Annotator) {
	a.SetToken("index", "Asset")
	a.Describe(r, "Manages assets (images, files, documents) for a Webflow site. "+
		"This resource allows you to upload and manage files that can be used in your Webflow site. "+
		"Note: Assets are immutable - changing any property will delete and recreate the asset.")
}

// Annotate adds descriptions to the AssetArgs fields.
func (args *AssetArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.FileName,
		"The name of the file to upload, including the extension. "+
			"Examples: 'logo.png', 'hero-image.jpg', 'document.pdf'. "+
			"The file name must not exceed 255 characters and should not contain "+
			"invalid characters (<, >, :, \", |, ?, *).")

	a.Describe(&args.FileHash,
		"Optional MD5 hash of the file content for deduplication. "+
			"If provided and a file with the same hash already exists in your site, "+
			"Webflow may reuse the existing file instead of uploading a duplicate. "+
			"Leave empty to always upload a new file.")

	a.Describe(&args.ParentFolder,
		"Optional folder ID where the asset will be organized in the Webflow Assets panel. "+
			"If not specified, the asset will be placed at the root level. "+
			"Example: '5f0c8c9e1c9d440000e8d8c4'.")

	a.Describe(&args.FileSource,
		"The source of the file to upload. "+
			"For the current implementation, this is a reference field. "+
			"In future versions, this may support URLs or local file paths for automatic upload. "+
			"Examples: 'https://example.com/logo.png', '/path/to/local/file.png'.")
}

// Annotate adds descriptions to the AssetState fields.
func (state *AssetState) Annotate(a infer.Annotator) {
	a.Describe(&state.AssetID,
		"The Webflow-assigned asset ID (read-only). "+
			"This unique identifier can be used to reference the asset in API calls.")

	a.Describe(&state.HostedURL,
		"The CDN URL where the asset is hosted (read-only). "+
			"Use this URL to reference the asset in your Webflow site or externally. "+
			"Example: 'https://assets.website-files.com/.../logo.png'.")

	a.Describe(&state.ContentType,
		"The MIME type of the uploaded asset (read-only). "+
			"Examples: 'image/png', 'image/jpeg', 'application/pdf'. "+
			"Automatically detected by Webflow based on the file content.")

	a.Describe(&state.Size,
		"The size of the asset in bytes (read-only). "+
			"This is the actual size of the uploaded file.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the asset was created (RFC3339 format, read-only). "+
			"This is automatically set when the asset is uploaded.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the asset was last modified (RFC3339 format, read-only). "+
			"For most assets, this will be the same as createdOn since assets are immutable.")
}

// Diff determines what changes need to be made to the asset resource.
// Assets are immutable - any change requires replacement (delete + recreate).
func (r *Asset) Diff(
	ctx context.Context, req infer.DiffRequest[AssetArgs, AssetState],
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

	// FileName change requires replacement (assets are immutable)
	if req.State.FileName != req.Inputs.FileName {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"fileName": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// FileHash change requires replacement
	if req.State.FileHash != req.Inputs.FileHash {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"fileHash": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// ParentFolder change requires replacement
	if req.State.ParentFolder != req.Inputs.ParentFolder {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"parentFolder": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// FileSource change requires replacement
	if req.State.FileSource != req.Inputs.FileSource {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"fileSource": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	return diff, nil
}

// Create creates a new asset by requesting an upload URL from Webflow.
// Note: For MVP, this creates the upload URL but doesn't perform the actual upload.
// Future versions will support automatic file upload.
func (r *Asset) Create(
	ctx context.Context, req infer.CreateRequest[AssetArgs],
) (infer.CreateResponse[AssetState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("validation failed for Asset resource: %w", err)
	}
	if err := ValidateFileName(req.Inputs.FileName); err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("validation failed for Asset resource: %w", err)
	}

	state := AssetState{
		AssetArgs: req.Inputs,
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set preview values
		state.AssetID = fmt.Sprintf("preview-%d", time.Now().Unix())
		state.HostedURL = "https://assets.website-files.com/preview/" + state.FileName
		state.CreatedOn = time.Now().Format(time.RFC3339)
		previewID := GenerateAssetResourceID(req.Inputs.SiteID, state.AssetID)
		return infer.CreateResponse[AssetState]{
			ID:     previewID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get upload URL
	// Note: For MVP, we're getting the upload URL but not performing the actual upload
	// Future versions will support automatic file upload from FileSource
	uploadResp, err := PostAssetUploadURL(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.FileName, req.Inputs.FileHash, req.Inputs.ParentFolder,
	)
	if err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("failed to request asset upload URL: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid upload URL
	if uploadResp.UploadURL == "" {
		return infer.CreateResponse[AssetState]{}, errors.New(
			"webflow API returned empty upload URL; " +
				"this is unexpected and may indicate an API issue")
	}

	// For MVP: We have the upload URL but don't perform the actual upload
	// The user would need to upload the file manually using the upload URL
	// or we track an existing asset
	//
	// For now, we'll return an error explaining this limitation
	return infer.CreateResponse[AssetState]{}, errors.New(
		"asset upload is not yet fully implemented; " +
			"this resource currently supports only tracking existing assets; " +
			"to upload assets, use the Webflow dashboard or API directly; " +
			"future versions will support automatic file upload")
}

// Read retrieves the current state of an asset from Webflow.
// Used for drift detection and import operations.
func (r *Asset) Read(
	ctx context.Context, req infer.ReadRequest[AssetArgs, AssetState],
) (infer.ReadResponse[AssetArgs, AssetState], error) {
	// Extract siteID and assetID from resource ID
	siteID, assetID, err := ExtractIDsFromAssetResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[AssetArgs, AssetState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[AssetArgs, AssetState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get asset details
	asset, err := GetAsset(ctx, client, assetID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if errors.Is(err, errors.New("not found")) {
			return infer.ReadResponse[AssetArgs, AssetState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[AssetArgs, AssetState]{}, fmt.Errorf("failed to read asset: %w", err)
	}

	// Build current state from API response
	currentInputs := AssetArgs{
		SiteID:   siteID,
		FileName: asset.OriginalFileName,
		// FileHash and ParentFolder are not returned by GET API
		FileHash:     req.State.FileHash,
		ParentFolder: req.State.ParentFolder,
		FileSource:   req.State.FileSource,
	}
	currentState := AssetState{
		AssetArgs:   currentInputs,
		AssetID:     asset.ID,
		HostedURL:   asset.HostedURL,
		ContentType: asset.ContentType,
		Size:        asset.Size,
		CreatedOn:   asset.CreatedOn,
		LastUpdated: asset.LastUpdated,
	}

	return infer.ReadResponse[AssetArgs, AssetState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update is not supported for assets - they are immutable.
// Any changes will trigger a replacement (delete + create).
func (r *Asset) Update(
	ctx context.Context, req infer.UpdateRequest[AssetArgs, AssetState],
) (infer.UpdateResponse[AssetState], error) {
	// This should never be called because Diff always returns UpdateReplace
	// But implement it defensively
	return infer.UpdateResponse[AssetState]{}, errors.New(
		"assets are immutable and cannot be updated in place; " +
			"any changes will trigger a replacement (delete and recreate)")
}

// Delete removes an asset from the Webflow site.
func (r *Asset) Delete(ctx context.Context, req infer.DeleteRequest[AssetState]) (infer.DeleteResponse, error) {
	// Extract siteID and assetID from resource ID
	_, assetID, err := ExtractIDsFromAssetResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteAsset(ctx, client, assetID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete asset: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
