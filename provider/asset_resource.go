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
	// FileHash is the MD5 hash of the file content (required).
	// Webflow uses this to identify and deduplicate assets.
	// Generate using: md5sum <filename> (Linux) or md5 <filename> (macOS)
	FileHash string `pulumi:"fileHash"`
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
	// UploadURL is the presigned S3 URL for uploading the file content (read-only).
	// Use this URL with UploadDetails to complete the asset upload via S3 POST.
	// See: https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPOST.html
	UploadURL string `pulumi:"uploadUrl,optional"`
	// UploadDetails contains AWS S3 POST form fields required for upload (read-only).
	// Keys include: acl, bucket, key, Content-Type, X-Amz-Algorithm, X-Amz-Credential,
	// X-Amz-Date, Policy, X-Amz-Signature, success_action_status, Cache-Control.
	UploadDetails map[string]string `pulumi:"uploadDetails,optional"`
	// AssetURL is the direct S3 URL for the asset (read-only).
	AssetURL string `pulumi:"assetUrl,optional"`
	// HostedURL is the Webflow CDN URL where the asset will be hosted (read-only).
	// This URL becomes accessible after completing the S3 upload.
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
		"MD5 hash of the file content (required). "+
			"Webflow uses this hash to identify and deduplicate assets. "+
			"Generate using: md5sum <filename> (Linux) or md5 <filename> (macOS). "+
			"Example: 'd41d8cd98f00b204e9800998ecf8427e'.")

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

	a.Describe(&state.UploadURL,
		"The presigned S3 URL for uploading the file content (read-only). "+
			"Use this URL along with uploadDetails to complete the asset upload. "+
			"See AWS S3 POST documentation: https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPOST.html")

	a.Describe(&state.UploadDetails,
		"AWS S3 POST form fields required to complete the upload (read-only). "+
			"Include these as form fields when POSTing the file to uploadUrl. "+
			"Keys: acl, bucket, key, Content-Type, X-Amz-Algorithm, X-Amz-Credential, "+
			"X-Amz-Date, Policy, X-Amz-Signature, success_action_status, Cache-Control.")

	a.Describe(&state.AssetURL,
		"The direct S3 URL for the asset (read-only). "+
			"This is the raw S3 location where the file is stored.")

	a.Describe(&state.HostedURL,
		"The Webflow CDN URL where the asset will be hosted (read-only). "+
			"This URL becomes accessible after completing the S3 upload. "+
			"Example: 'https://assets.website-files.com/.../logo.png'.")

	a.Describe(&state.ContentType,
		"The MIME type of the asset (read-only). "+
			"Examples: 'image/png', 'image/jpeg', 'application/pdf'. "+
			"Determined by the fileName extension.")

	a.Describe(&state.Size,
		"The size of the asset in bytes (read-only). "+
			"This is the actual size of the uploaded file.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the asset metadata was created (RFC3339 format, read-only). "+
			"This is set when the asset is registered with Webflow.")

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
// The Webflow API returns an asset ID and presigned S3 upload URL.
// Note: The actual file upload to S3 must be done separately using the uploadUrl and uploadDetails.
func (r *Asset) Create(
	ctx context.Context, req infer.CreateRequest[AssetArgs],
) (infer.CreateResponse[AssetState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("validation failed for Asset resource: %w", err)
	}
	if err := ValidateFileName(req.Inputs.FileName); err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("validation failed for Asset resource: %w", err)
	}
	if err := ValidateFileHash(req.Inputs.FileHash); err != nil {
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

	// Call Webflow API to create asset metadata and get upload URL
	uploadResp, err := PostAssetUploadURL(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.FileName, req.Inputs.FileHash, req.Inputs.ParentFolder,
	)
	if err != nil {
		return infer.CreateResponse[AssetState]{}, fmt.Errorf("failed to create asset: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid asset ID
	if uploadResp.ID == "" {
		return infer.CreateResponse[AssetState]{}, errors.New(
			"webflow API returned empty asset ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Populate state from API response
	state.AssetID = uploadResp.ID
	state.UploadURL = uploadResp.UploadURL
	state.UploadDetails = uploadResp.UploadDetails
	state.AssetURL = uploadResp.AssetURL
	state.HostedURL = uploadResp.HostedURL
	state.ContentType = uploadResp.ContentType
	state.CreatedOn = uploadResp.CreatedOn
	state.LastUpdated = uploadResp.LastUpdated

	// Note: The actual file must be uploaded to S3 using uploadUrl and uploadDetails.
	// See: https://docs.aws.amazon.com/AmazonS3/latest/API/RESTObjectPOST.html

	resourceID := GenerateAssetResourceID(req.Inputs.SiteID, uploadResp.ID)

	return infer.CreateResponse[AssetState]{
		ID:     resourceID,
		Output: state,
	}, nil
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
		"assets are immutable and cannot be updated in place. " +
			"Any changes will trigger a replacement (delete and recreate)")
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
