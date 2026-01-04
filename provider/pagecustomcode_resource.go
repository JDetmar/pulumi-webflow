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

// PageCustomCode is the resource controller for managing custom code scripts on a Webflow page.
// It allows applying registered custom code scripts (JavaScript) to pages.
type PageCustomCode struct{}

// PageCustomCodeScript represents a single script to apply to a page.
type PageCustomCodeScript struct {
	// ID is the unique identifier of a registered custom code script (required).
	// Must be a script that was previously registered via the RegisteredScript resource.
	ID string `pulumi:"id"`
	// Version is the semantic version string for the registered script (required).
	// Example: "1.0.0" or "2.5.3"
	Version string `pulumi:"version"`
	// Location is where the script should be applied (required).
	// Must be either "header" (loaded before body closes) or "footer" (loaded at end of page).
	Location string `pulumi:"location"`
	// Attributes is an optional map of developer-specified key/value pairs for script attributes.
	Attributes map[string]interface{} `pulumi:"attributes,omitempty"`
}

// PageCustomCodeArgs defines the input properties for the PageCustomCode resource.
type PageCustomCodeArgs struct {
	// PageID is the Webflow page ID to apply scripts to (24-character hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c4"
	PageID string `pulumi:"pageId"`
	// Scripts is the list of custom code scripts to apply to the page.
	// All scripts in this list will be applied; scripts not listed will be removed.
	Scripts []PageCustomCodeScript `pulumi:"scripts"`
}

// PageCustomCodeState defines the output properties for the PageCustomCode resource.
// It embeds PageCustomCodeArgs to include input properties in the output.
type PageCustomCodeState struct {
	PageCustomCodeArgs
	// LastUpdated is the timestamp when the custom code was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
	// CreatedOn is the timestamp when the custom code was first created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
}

// Annotate adds descriptions and constraints to the PageCustomCode resource.
func (r *PageCustomCode) Annotate(a infer.Annotator) {
	a.SetToken("index", "PageCustomCode")
	a.Describe(r, "Manages custom code (JavaScript) scripts applied to a Webflow page. "+
		"This resource allows you to apply registered custom code scripts to specific pages. "+
		"Scripts must first be registered using the RegisteredScript resource before they can be applied. "+
		"All scripts in the configuration will be applied to the page; scripts not listed will be removed.")
}

// Annotate adds descriptions to the PageCustomCodeArgs fields.
func (args *PageCustomCodeArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.PageID,
		"The Webflow page ID (24-character hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c4'). "+
			"You can find page IDs using the Webflow Pages API or in the Webflow designer. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.Scripts,
		"List of custom code scripts to apply to the page. "+
			"Each script must reference a script ID and version that have been previously registered. "+
			"All scripts in this list will be applied to the page; "+
			"any scripts not listed will be removed from the page.")
}

// Annotate adds descriptions to PageCustomCodeScript fields.
func (s *PageCustomCodeScript) Annotate(a infer.Annotator) {
	a.Describe(&s.ID,
		"The unique identifier of a registered custom code script. "+
			"This must be a script that was previously registered using the RegisteredScript resource. "+
			"Script IDs are assigned by Webflow when the script is registered.")

	a.Describe(&s.Version,
		"The semantic version string for the registered script (e.g., '1.0.0'). "+
			"This version must match a registered version of the script. "+
			"You can have multiple versions of the same script registered.")

	a.Describe(&s.Location,
		"Where the script should be applied on the page. "+
			"Must be either 'header' (loaded in page header) or 'footer' (loaded at end of page). "+
			"Use 'header' for scripts that don't depend on DOM elements. "+
			"Use 'footer' for scripts that need to run after DOM is fully loaded.")

	a.Describe(&s.Attributes,
		"Optional developer-specified key/value pairs for script attributes. "+
			"These attributes can be used by the script to customize its behavior on this page.")
}

// Annotate adds descriptions to the PageCustomCodeState fields.
func (state *PageCustomCodeState) Annotate(a infer.Annotator) {
	a.Describe(&state.LastUpdated,
		"The timestamp when the page custom code was last updated (RFC3339 format). "+
			"This is automatically set when the configuration is updated and is read-only.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the page custom code was first created (RFC3339 format). "+
			"This is automatically set when the configuration is first created and is read-only.")
}

// Diff determines what changes need to be made to the page custom code resource.
// PageID changes trigger replacement (different page).
// Scripts changes trigger in-place update.
func (r *PageCustomCode) Diff(
	ctx context.Context, req infer.DiffRequest[PageCustomCodeArgs, PageCustomCodeState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}

	// Check for pageId change (requires replacement)
	if req.State.PageID != req.Inputs.PageID {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"pageId": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check if scripts list has changed
	if !pageCustomCodeScriptsEqual(req.State.Scripts, req.Inputs.Scripts) {
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"scripts": {Kind: p.Update},
		}
		return diff, nil
	}

	return diff, nil
}

// pageCustomCodeScriptsEqual checks if two script lists are equal.
func pageCustomCodeScriptsEqual(stateScripts, inputScripts []PageCustomCodeScript) bool {
	if len(stateScripts) != len(inputScripts) {
		return false
	}

	// Create maps for easier comparison
	stateMap := make(map[string]PageCustomCodeScript)
	for _, s := range stateScripts {
		stateMap[s.ID] = s
	}

	// Check if all input scripts exist in state with same values
	for _, inputScript := range inputScripts {
		stateScript, exists := stateMap[inputScript.ID]
		if !exists {
			return false
		}
		if stateScript.Version != inputScript.Version ||
			stateScript.Location != inputScript.Location {
			return false
		}
		// Compare attributes maps
		if !pageCustomCodeAttributesEqual(stateScript.Attributes, inputScript.Attributes) {
			return false
		}
	}

	return true
}

// pageCustomCodeAttributesEqual checks if two attribute maps are equal.
func pageCustomCodeAttributesEqual(stateAttrs, inputAttrs map[string]interface{}) bool {
	if len(stateAttrs) != len(inputAttrs) {
		return false
	}
	for key, value := range stateAttrs {
		if inputAttrs[key] != value {
			return false
		}
	}
	return true
}

// Create applies custom code scripts to a page.
// Note: PageCustomCode is a configuration resource - "create" means "apply this configuration".
func (r *PageCustomCode) Create(
	ctx context.Context, req infer.CreateRequest[PageCustomCodeArgs],
) (infer.CreateResponse[PageCustomCodeState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidatePageID(req.Inputs.PageID); err != nil {
		return infer.CreateResponse[PageCustomCodeState]{},
			fmt.Errorf("validation failed for PageCustomCode resource: %w", err)
	}

	// Validate scripts
	if len(req.Inputs.Scripts) == 0 {
		return infer.CreateResponse[PageCustomCodeState]{}, errors.New(
			"validation failed for PageCustomCode resource: " +
				"at least one script is required. " +
				"Please provide a list of scripts with id, version, and location fields.")
	}

	for i, script := range req.Inputs.Scripts {
		if err := ValidateScriptID(script.ID); err != nil {
			return infer.CreateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
		if err := ValidateScriptVersion(script.Version); err != nil {
			return infer.CreateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
		if err := ValidateScriptLocation(script.Location); err != nil {
			return infer.CreateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
	}

	state := PageCustomCodeState{
		PageCustomCodeArgs: req.Inputs,
		LastUpdated:        "",
		CreatedOn:          "",
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		now := time.Now().Format(time.RFC3339)
		state.LastUpdated = now
		state.CreatedOn = now
		resourceID := GeneratePageCustomCodeResourceID(req.Inputs.PageID)
		return infer.CreateResponse[PageCustomCodeState]{
			ID:     resourceID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[PageCustomCodeState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Convert input scripts to API format
	apiScripts := make([]CustomCodeScript, len(req.Inputs.Scripts))
	for i, script := range req.Inputs.Scripts {
		apiScripts[i] = CustomCodeScript(script)
	}

	// Call Webflow API to apply scripts
	response, err := PutPageCustomCode(ctx, client, req.Inputs.PageID, &PageCustomCodeRequest{
		Scripts: apiScripts,
	})
	if err != nil {
		return infer.CreateResponse[PageCustomCodeState]{}, fmt.Errorf("failed to apply custom code to page: %w", err)
	}

	// Populate timestamps from response
	state.LastUpdated = response.LastUpdated
	state.CreatedOn = response.CreatedOn

	resourceID := GeneratePageCustomCodeResourceID(req.Inputs.PageID)

	return infer.CreateResponse[PageCustomCodeState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of page custom code from Webflow.
// Used for drift detection and refresh operations.
func (r *PageCustomCode) Read(
	ctx context.Context, req infer.ReadRequest[PageCustomCodeArgs, PageCustomCodeState],
) (infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState], error) {
	// Extract pageID from resource ID
	pageID, err := ExtractPageIDFromPageCustomCodeResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState]{},
			fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get current custom code
	response, err := GetPageCustomCode(ctx, client, pageID)
	if err != nil {
		// If page not found, return empty ID to signal deletion
		return infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState]{
			ID: "",
		}, nil
	}

	// Build current state - preserve configured scripts from state instead of reading from API
	// This is intentional: we want to track what we configured, not what might have been added externally
	currentInputs := PageCustomCodeArgs{
		PageID:  pageID,
		Scripts: req.State.Scripts, // Preserve configured scripts (NOT read from API)
	}
	currentState := PageCustomCodeState{
		PageCustomCodeArgs: currentInputs,
		LastUpdated:        response.LastUpdated,
		CreatedOn:          response.CreatedOn,
	}

	// Verify the page still exists (basic check)
	if pageID == "" {
		return infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState]{
			ID: "",
		}, nil
	}

	return infer.ReadResponse[PageCustomCodeArgs, PageCustomCodeState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies existing page custom code.
func (r *PageCustomCode) Update(
	ctx context.Context, req infer.UpdateRequest[PageCustomCodeArgs, PageCustomCodeState],
) (infer.UpdateResponse[PageCustomCodeState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidatePageID(req.Inputs.PageID); err != nil {
		return infer.UpdateResponse[PageCustomCodeState]{},
			fmt.Errorf("validation failed for PageCustomCode resource: %w", err)
	}

	// Validate scripts
	if len(req.Inputs.Scripts) == 0 {
		return infer.UpdateResponse[PageCustomCodeState]{}, errors.New(
			"validation failed for PageCustomCode resource: " +
				"at least one script is required. " +
				"Please provide a list of scripts with id, version, and location fields.")
	}

	for i, script := range req.Inputs.Scripts {
		if err := ValidateScriptID(script.ID); err != nil {
			return infer.UpdateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
		if err := ValidateScriptVersion(script.Version); err != nil {
			return infer.UpdateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
		if err := ValidateScriptLocation(script.Location); err != nil {
			return infer.UpdateResponse[PageCustomCodeState]{}, fmt.Errorf(
				"validation failed for PageCustomCode resource, scripts[%d]: %w", i, err)
		}
	}

	state := PageCustomCodeState{
		PageCustomCodeArgs: req.Inputs,
		LastUpdated:        "",
		CreatedOn:          req.State.CreatedOn, // Preserve original creation time
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		state.LastUpdated = time.Now().Format(time.RFC3339)
		return infer.UpdateResponse[PageCustomCodeState]{
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[PageCustomCodeState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Convert input scripts to API format
	apiScripts := make([]CustomCodeScript, len(req.Inputs.Scripts))
	for i, script := range req.Inputs.Scripts {
		apiScripts[i] = CustomCodeScript(script)
	}

	// Call Webflow API to update scripts
	response, err := PutPageCustomCode(ctx, client, req.Inputs.PageID, &PageCustomCodeRequest{
		Scripts: apiScripts,
	})
	if err != nil {
		return infer.UpdateResponse[PageCustomCodeState]{}, fmt.Errorf("failed to update custom code on page: %w", err)
	}

	// Update timestamp
	state.LastUpdated = response.LastUpdated

	return infer.UpdateResponse[PageCustomCodeState]{
		Output: state,
	}, nil
}

// Delete removes all custom code scripts from the page.
// This calls the DELETE endpoint to remove the custom code.
func (r *PageCustomCode) Delete(
	ctx context.Context, req infer.DeleteRequest[PageCustomCodeState],
) (infer.DeleteResponse, error) {
	// Extract pageID from resource ID
	pageID, err := ExtractPageIDFromPageCustomCodeResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to delete custom code (idempotent)
	err = DeletePageCustomCode(ctx, client, pageID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete page custom code: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
