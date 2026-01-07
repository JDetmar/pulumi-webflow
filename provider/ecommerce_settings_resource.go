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

// EcommerceSettings is the resource controller for managing Webflow ecommerce settings.
// It implements the infer.CustomResource interface.
//
// Note: This is a read-only resource. Ecommerce must be enabled through the Webflow dashboard.
// This resource allows you to import and track existing ecommerce settings as infrastructure state.
type EcommerceSettings struct{}

// EcommerceSettingsArgs defines the input properties for the EcommerceSettings resource.
type EcommerceSettingsArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
}

// EcommerceSettingsState defines the output properties for the EcommerceSettings resource.
// It embeds EcommerceSettingsArgs to include input properties in the output.
type EcommerceSettingsState struct {
	EcommerceSettingsArgs
	// DefaultCurrency is the three-letter ISO 4217 currency code for the site (read-only).
	// Examples: "USD", "EUR", "GBP"
	DefaultCurrency string `pulumi:"defaultCurrency"`
	// CreatedOn is the timestamp when ecommerce was enabled on the site (read-only, ISO 8601 format).
	CreatedOn string `pulumi:"createdOn,optional"`
}

// Annotate adds descriptions and constraints to the EcommerceSettings resource.
func (r *EcommerceSettings) Annotate(a infer.Annotator) {
	a.SetToken("index", "EcommerceSettings")
	a.Describe(r, "Manages (imports) ecommerce settings for a Webflow site. "+
		"This is a read-only resource that allows you to track and reference existing ecommerce settings. "+
		"Ecommerce must be enabled through the Webflow dashboard before this resource can be used. "+
		"Use this resource to access the site's default currency and verify ecommerce is enabled.")
}

// Annotate adds descriptions to the EcommerceSettingsArgs fields.
func (args *EcommerceSettingsArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"The site must have ecommerce enabled through the Webflow dashboard. "+
			"You can find your site ID in the Webflow dashboard under Site Settings.")
}

// Annotate adds descriptions to the EcommerceSettingsState fields.
func (state *EcommerceSettingsState) Annotate(a infer.Annotator) {
	a.Describe(&state.DefaultCurrency,
		"The three-letter ISO 4217 currency code for the site (e.g., 'USD', 'EUR', 'GBP'). "+
			"This is the default currency used for ecommerce transactions on this site. "+
			"This value is set in the Webflow dashboard and is read-only.")

	a.Describe(&state.CreatedOn,
		"The timestamp when ecommerce was enabled on the site (ISO 8601 format). "+
			"This is automatically set when ecommerce is enabled and is read-only.")
}

// Diff determines what changes need to be made to the ecommerce settings resource.
// Only siteId changes trigger replacement since this is a read-only resource.
func (r *EcommerceSettings) Diff(
	ctx context.Context, req infer.DiffRequest[EcommerceSettingsArgs, EcommerceSettingsState],
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

	// No other changes are possible since this is a read-only resource
	return diff, nil
}

// Create "creates" an ecommerce settings resource by reading the existing settings from Webflow.
// This is a read-only resource - ecommerce must be enabled through the Webflow dashboard first.
func (r *EcommerceSettings) Create(
	ctx context.Context, req infer.CreateRequest[EcommerceSettingsArgs],
) (infer.CreateResponse[EcommerceSettingsState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[EcommerceSettingsState]{},
			fmt.Errorf("validation failed for EcommerceSettings resource: %w", err)
	}

	state := EcommerceSettingsState{
		EcommerceSettingsArgs: req.Inputs,
		DefaultCurrency:       "",
		CreatedOn:             "",
	}
	resourceID := GenerateEcommerceSettingsResourceID(req.Inputs.SiteID)

	// During preview, return expected state without making API calls
	if req.DryRun {
		// During preview, we can't know the actual values, so we set placeholders
		state.DefaultCurrency = "USD" // Placeholder - actual value will be fetched on apply
		return infer.CreateResponse[EcommerceSettingsState]{
			ID:     resourceID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[EcommerceSettingsState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to read existing ecommerce settings
	response, err := GetEcommerceSettings(ctx, client, req.Inputs.SiteID)
	if err != nil {
		return infer.CreateResponse[EcommerceSettingsState]{},
			fmt.Errorf("failed to read ecommerce settings: %w. "+
				"Ensure ecommerce is enabled on this site through the Webflow dashboard", err)
	}

	// Update state with response
	state.DefaultCurrency = response.DefaultCurrency
	state.CreatedOn = response.CreatedOn

	return infer.CreateResponse[EcommerceSettingsState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of ecommerce settings from Webflow.
// Used for drift detection and import operations.
func (r *EcommerceSettings) Read(
	ctx context.Context, req infer.ReadRequest[EcommerceSettingsArgs, EcommerceSettingsState],
) (infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState], error) {
	// Extract siteID from resource ID
	siteID, err := ExtractSiteIDFromEcommerceSettingsResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{},
			fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{},
			fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := GetEcommerceSettings(ctx, client, siteID)
	if err != nil {
		// Propagate context cancellation errors
		if errors.Is(err, context.Canceled) {
			return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{}, err
		}
		// Treat "not found" or "ecommerce not enabled" as resource deletion
		// This means ecommerce has been disabled on the site
		if strings.Contains(strings.ToLower(err.Error()), "not found") ||
			strings.Contains(strings.ToLower(err.Error()), "ecommerce not enabled") {
			return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{
				ID: "",
			}, nil
		}
		// For other errors (network issues, rate limiting, etc.), propagate the error
		return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{},
			fmt.Errorf("failed to read ecommerce settings: %w", err)
	}

	// Build current state from API response
	currentInputs := EcommerceSettingsArgs{
		SiteID: siteID,
	}
	currentState := EcommerceSettingsState{
		EcommerceSettingsArgs: currentInputs,
		DefaultCurrency:       response.DefaultCurrency,
		CreatedOn:             response.CreatedOn,
	}

	return infer.ReadResponse[EcommerceSettingsArgs, EcommerceSettingsState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update is a no-op for this read-only resource.
// Ecommerce settings cannot be modified through the Webflow API - they must be changed
// through the Webflow dashboard.
func (r *EcommerceSettings) Update(
	ctx context.Context, req infer.UpdateRequest[EcommerceSettingsArgs, EcommerceSettingsState],
) (infer.UpdateResponse[EcommerceSettingsState], error) {
	// This should never be called since Diff only reports changes for siteId (which triggers replace)
	// But we implement it to return the current state
	state := EcommerceSettingsState{
		EcommerceSettingsArgs: req.Inputs,
		DefaultCurrency:       req.State.DefaultCurrency,
		CreatedOn:             req.State.CreatedOn,
	}

	return infer.UpdateResponse[EcommerceSettingsState]{
		Output: state,
	}, nil
}

// Delete removes the ecommerce settings from Pulumi state.
// Note: This does NOT disable ecommerce on the site - that must be done through
// the Webflow dashboard. This only removes the resource from Pulumi management.
func (r *EcommerceSettings) Delete(
	ctx context.Context, req infer.DeleteRequest[EcommerceSettingsState],
) (infer.DeleteResponse, error) {
	// This is a no-op - we can't delete ecommerce settings via API.
	// The resource is simply removed from Pulumi state.
	// Ecommerce must be disabled through the Webflow dashboard if needed.
	return infer.DeleteResponse{}, nil
}
