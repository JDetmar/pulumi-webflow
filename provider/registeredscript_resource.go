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

// RegisteredScriptResource is the resource controller for managing Webflow registered scripts.
// It implements the infer.CustomResource interface for full CRUD operations.
type RegisteredScriptResource struct{}

// RegisteredScriptResourceArgs defines the input properties for the RegisteredScript resource.
type RegisteredScriptResourceArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// DisplayName is the user-facing name for the script (1-50 alphanumeric characters).
	// Example: "CmsSlider", "AnalyticsScript", "MyCustomScript123"
	DisplayName string `pulumi:"displayName"`
	// HostedLocation is the URI for the externally hosted script.
	// Example: "https://cdn.example.com/my-script.js"
	HostedLocation string `pulumi:"hostedLocation"`
	// IntegrityHash is the Sub-Resource Integrity Hash (SRI) for the script.
	// Format: "sha384-<hash>", "sha256-<hash>", or "sha512-<hash>"
	// You can generate an SRI hash using https://www.srihash.org/
	IntegrityHash string `pulumi:"integrityHash"`
	// Version is the Semantic Version (SemVer) string for the script.
	// Format: "major.minor.patch" (e.g., "1.0.0", "2.3.1")
	// See https://semver.org/ for more information.
	Version string `pulumi:"version"`
	// CanCopy indicates whether the script can be copied on site duplication.
	// Default: false
	CanCopy bool `pulumi:"canCopy,optional"`
}

// RegisteredScriptResourceState defines the output properties for the RegisteredScript resource.
// It embeds RegisteredScriptResourceArgs to include input properties in the output.
type RegisteredScriptResourceState struct {
	RegisteredScriptResourceArgs
	// CreatedOn is the timestamp when the script was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastUpdated is the timestamp when the script was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
}

// Annotate adds descriptions and constraints to the RegisteredScript resource.
func (r *RegisteredScriptResource) Annotate(a infer.Annotator) {
	a.SetToken("index", "RegisteredScript")
	a.Describe(r, "Manages custom code scripts in the Webflow script registry. "+
		"This resource allows you to register and manage externally hosted scripts that can be "+
		"deployed across your Webflow site with version control and integrity verification.")
}

// Annotate adds descriptions to the RegisteredScriptResourceArgs fields.
func (args *RegisteredScriptResourceArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.DisplayName,
		"The user-facing name for the script (1-50 alphanumeric characters). "+
			"This name is used to identify the script in the Webflow interface. "+
			"Only letters (A-Z, a-z) and numbers (0-9) are allowed. "+
			"Example valid names: 'CmsSlider', 'AnalyticsScript', 'MyCustomScript123'.")

	a.Describe(&args.HostedLocation,
		"The URI for the externally hosted script (e.g., 'https://cdn.example.com/my-script.js'). "+
			"Must be a valid HTTP or HTTPS URL. "+
			"The script should be publicly accessible and properly configured for cross-origin requests.")

	a.Describe(&args.IntegrityHash,
		"The Sub-Resource Integrity (SRI) hash for the script. "+
			"Format: 'sha384-<hash>', 'sha256-<hash>', or 'sha512-<hash>'. "+
			"SRI hashes help ensure that the script hasn't been modified in transit. "+
			"You can generate an SRI hash using https://www.srihash.org/")

	a.Describe(&args.Version,
		"The Semantic Version (SemVer) string for the script "+
			"(e.g., '1.0.0', '2.3.1'). "+
			"This helps track different versions of your script. "+
			"See https://semver.org/ for more information on semantic versioning.")

	a.Describe(&args.CanCopy,
		"Indicates whether the script can be copied when the site is duplicated. "+
			"Default: false. "+
			"When true, the script will be included when creating a copy of the site.")
}

// Annotate adds descriptions to the RegisteredScriptResourceState fields.
func (state *RegisteredScriptResourceState) Annotate(a infer.Annotator) {
	a.Describe(&state.CreatedOn,
		"The timestamp when the script was created (RFC3339 format). "+
			"This is automatically set by Webflow when the script is created and is read-only.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the script was last updated (RFC3339 format). "+
			"This is automatically updated by Webflow when the script is modified and is read-only.")
}

// Diff determines what changes need to be made to the registered script resource.
// SiteID and DisplayName changes trigger replacement.
// Other changes trigger in-place update.
func (r *RegisteredScriptResource) Diff(
	ctx context.Context, req infer.DiffRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}
	detailedDiff := map[string]p.PropertyDiff{}

	// Check for siteId change (requires replacement - primary key)
	if req.State.SiteID != req.Inputs.SiteID {
		detailedDiff["siteId"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Check for displayName change (requires replacement - secondary key)
	if req.State.DisplayName != req.Inputs.DisplayName {
		detailedDiff["displayName"] = p.PropertyDiff{Kind: p.UpdateReplace}
	}

	// Check for hostedLocation change (supports update)
	if req.State.HostedLocation != req.Inputs.HostedLocation {
		detailedDiff["hostedLocation"] = p.PropertyDiff{Kind: p.Update}
	}

	// Check for integrityHash change (supports update)
	if req.State.IntegrityHash != req.Inputs.IntegrityHash {
		detailedDiff["integrityHash"] = p.PropertyDiff{Kind: p.Update}
	}

	// Check for version change (supports update)
	if req.State.Version != req.Inputs.Version {
		detailedDiff["version"] = p.PropertyDiff{Kind: p.Update}
	}

	// Check for canCopy change (supports update)
	if req.State.CanCopy != req.Inputs.CanCopy {
		detailedDiff["canCopy"] = p.PropertyDiff{Kind: p.Update}
	}

	// If any changes were detected, populate the diff response
	if len(detailedDiff) > 0 {
		diff.HasChanges = true
		// Only set DeleteBeforeReplace if any replacement changes are needed
		for _, change := range detailedDiff {
			if change.Kind == p.UpdateReplace {
				diff.DeleteBeforeReplace = true
				break
			}
		}
		diff.DetailedDiff = detailedDiff
	}

	return diff, nil
}

// Create creates a new registered script on the Webflow site.
func (r *RegisteredScriptResource) Create(
	ctx context.Context, req infer.CreateRequest[RegisteredScriptResourceArgs],
) (infer.CreateResponse[RegisteredScriptResourceState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateScriptDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateHostedLocation(req.Inputs.HostedLocation); err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateIntegrityHash(req.Inputs.IntegrityHash); err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateVersion(req.Inputs.Version); err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}

	state := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: req.Inputs,
		CreatedOn:                    "", // Will be populated after creation
		LastUpdated:                  "", // Will be populated after creation
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set a preview timestamp
		now := time.Now().Format(time.RFC3339)
		state.CreatedOn = now
		state.LastUpdated = now
		// Generate a predictable ID for dry-run
		previewID := fmt.Sprintf("preview-%d", time.Now().Unix())
		return infer.CreateResponse[RegisteredScriptResourceState]{
			ID:     GenerateRegisteredScriptResourceID(req.Inputs.SiteID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostRegisteredScript(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.DisplayName, req.Inputs.HostedLocation, req.Inputs.IntegrityHash,
		req.Inputs.Version, req.Inputs.CanCopy,
	)
	if err != nil {
		return infer.CreateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("failed to create registered script: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid script ID
	if response.ID == "" {
		return infer.CreateResponse[RegisteredScriptResourceState]{}, errors.New(
			"Webflow API returned empty registered script ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Update state with values from API response
	state.CreatedOn = response.CreatedOn
	state.LastUpdated = response.LastUpdated

	resourceID := GenerateRegisteredScriptResourceID(req.Inputs.SiteID, response.ID)

	return infer.CreateResponse[RegisteredScriptResourceState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a registered script from Webflow.
// Used for drift detection and import operations.
func (r *RegisteredScriptResource) Read(
	ctx context.Context, req infer.ReadRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState],
) (infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState], error) {
	// Extract siteID and scriptID from resource ID
	siteID, scriptID, err := ExtractIDsFromRegisteredScriptResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{},
			fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{},
			fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get all scripts for this site
	response, err := GetRegisteredScripts(ctx, client, siteID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{},
			fmt.Errorf("failed to read registered scripts: %w", err)
	}

	// Find the specific script in the list
	var foundScript *RegisteredScript
	for i, script := range response.RegisteredScripts {
		if script.ID == scriptID {
			foundScript = &response.RegisteredScripts[i]
			break
		}
	}

	// If script not found, return empty ID to signal deletion
	if foundScript == nil {
		return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
			ID: "",
		}, nil
	}

	// Build current state from API response
	currentInputs := RegisteredScriptResourceArgs{
		SiteID:         siteID,
		DisplayName:    foundScript.DisplayName,
		HostedLocation: foundScript.HostedLocation,
		IntegrityHash:  foundScript.IntegrityHash,
		Version:        foundScript.Version,
		CanCopy:        foundScript.CanCopy,
	}
	currentState := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: currentInputs,
		CreatedOn:                    foundScript.CreatedOn,
		LastUpdated:                  foundScript.LastUpdated,
	}

	return infer.ReadResponse[RegisteredScriptResourceArgs, RegisteredScriptResourceState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing registered script.
func (r *RegisteredScriptResource) Update(
	ctx context.Context, req infer.UpdateRequest[RegisteredScriptResourceArgs, RegisteredScriptResourceState],
) (infer.UpdateResponse[RegisteredScriptResourceState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateScriptDisplayName(req.Inputs.DisplayName); err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateHostedLocation(req.Inputs.HostedLocation); err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateIntegrityHash(req.Inputs.IntegrityHash); err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}
	if err := ValidateVersion(req.Inputs.Version); err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("validation failed for RegisteredScript resource: %w", err)
	}

	state := RegisteredScriptResourceState{
		RegisteredScriptResourceArgs: req.Inputs,
		CreatedOn:                    req.State.CreatedOn, // Preserve the creation timestamp from current state
		LastUpdated:                  "",                  // Will be updated from API response
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		state.LastUpdated = time.Now().Format(time.RFC3339)
		return infer.UpdateResponse[RegisteredScriptResourceState]{
			Output: state,
		}, nil
	}

	// Extract the Webflow script ID from the Pulumi resource ID
	_, scriptID, err := ExtractIDsFromRegisteredScriptResourceID(req.ID)
	if err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PatchRegisteredScript(
		ctx, client, req.Inputs.SiteID, scriptID,
		req.Inputs.DisplayName, req.Inputs.HostedLocation, req.Inputs.IntegrityHash,
		req.Inputs.Version, req.Inputs.CanCopy,
	)
	if err != nil {
		return infer.UpdateResponse[RegisteredScriptResourceState]{},
			fmt.Errorf("failed to update registered script: %w", err)
	}

	// Update state with values from API response
	state.LastUpdated = response.LastUpdated

	return infer.UpdateResponse[RegisteredScriptResourceState]{
		Output: state,
	}, nil
}

// Delete removes a registered script from the Webflow site.
func (r *RegisteredScriptResource) Delete(
	ctx context.Context, req infer.DeleteRequest[RegisteredScriptResourceState],
) (infer.DeleteResponse, error) {
	// Extract siteID and scriptID from resource ID
	siteID, scriptID, err := ExtractIDsFromRegisteredScriptResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteRegisteredScript(ctx, client, siteID, scriptID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete registered script: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
