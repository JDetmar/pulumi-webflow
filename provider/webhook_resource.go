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

// Webhook is the resource controller for managing Webflow webhooks.
// It implements the infer.CustomResource interface for full CRUD operations.
type Webhook struct{}

// WebhookArgs defines the input properties for the Webhook resource.
type WebhookArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// TriggerType is the Webflow event that triggers this webhook.
	// Valid values: form_submission, site_publish, page_created, page_metadata_updated,
	// page_deleted, ecomm_new_order, ecomm_order_changed, ecomm_inventory_changed,
	// memberships_user_account_added, memberships_user_account_updated, memberships_user_account_deleted,
	// collection_item_created, collection_item_changed, collection_item_deleted, collection_item_unpublished
	TriggerType string `pulumi:"triggerType"`
	// URL is the HTTPS endpoint where Webflow will send webhook events.
	// Must be a valid HTTPS URL (e.g., "https://example.com/webhooks/webflow")
	URL string `pulumi:"url"`
	// Filter is an optional map for filtering webhook events.
	// The structure depends on the triggerType and allows you to receive only specific events.
	Filter map[string]interface{} `pulumi:"filter,optional"`
}

// WebhookState defines the output properties for the Webhook resource.
// It embeds WebhookArgs to include input properties in the output.
type WebhookState struct {
	WebhookArgs
	// CreatedOn is the timestamp when the webhook was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastTriggered is the timestamp when the webhook was last triggered (read-only).
	LastTriggered string `pulumi:"lastTriggered,optional"`
}

// Annotate adds descriptions and constraints to the Webhook resource.
func (w *Webhook) Annotate(a infer.Annotator) {
	a.SetToken("index", "Webhook")
	a.Describe(w, "Manages webhooks for a Webflow site. "+
		"Webhooks allow you to receive real-time notifications when events occur in your Webflow site, "+
		"such as form submissions, page updates, e-commerce orders, and more. "+
		"Note: Webhooks cannot be updated in-place; any change to triggerType, url, or filter requires replacement.")
}

// Annotate adds descriptions to the WebhookArgs fields.
func (args *WebhookArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.TriggerType,
		"The Webflow event that triggers this webhook. "+
			"Valid values: form_submission, site_publish, page_created, page_metadata_updated, "+
			"page_deleted, ecomm_new_order, ecomm_order_changed, ecomm_inventory_changed, "+
			"memberships_user_account_added, memberships_user_account_updated, memberships_user_account_deleted, "+
			"collection_item_created, collection_item_changed, collection_item_deleted, collection_item_unpublished. "+
			"Example: 'form_submission' to receive notifications when forms are submitted.")

	a.Describe(&args.URL,
		"The HTTPS endpoint where Webflow will send webhook events "+
			"(e.g., 'https://example.com/webhooks/webflow', 'https://api.example.com/events'). "+
			"Must be a valid HTTPS URL. Webflow requires HTTPS for security. "+
			"Your endpoint should accept POST requests with JSON payloads containing event data.")

	a.Describe(&args.Filter,
		"Optional filter for webhook events. "+
			"The structure depends on the triggerType and allows you to receive only specific events. "+
			"For example, for collection_item_created, you can filter by collection ID. "+
			"Refer to Webflow API documentation for filter options for each trigger type.")
}

// Annotate adds descriptions to the WebhookState fields.
func (state *WebhookState) Annotate(a infer.Annotator) {
	a.Describe(&state.CreatedOn,
		"The timestamp when the webhook was created (RFC3339 format). "+
			"This is automatically set by Webflow when the webhook is created and is read-only.")

	a.Describe(&state.LastTriggered,
		"The timestamp when the webhook was last triggered (RFC3339 format). "+
			"This is automatically updated by Webflow when the webhook fires and is read-only. "+
			"Will be empty if the webhook has never been triggered.")
}

// Diff determines what changes need to be made to the webhook resource.
// Webflow webhooks do not support updates - all changes require replacement.
func (w *Webhook) Diff(
	ctx context.Context, req infer.DiffRequest[WebhookArgs, WebhookState],
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

	// Check for triggerType change (requires replacement - webhooks cannot be updated)
	if req.State.TriggerType != req.Inputs.TriggerType {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"triggerType": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for URL change (requires replacement - webhooks cannot be updated)
	if req.State.URL != req.Inputs.URL {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"url": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for filter change (requires replacement - webhooks cannot be updated)
	// Compare filter maps - if either is nil or they differ, trigger replacement
	if !mapsEqual(req.State.Filter, req.Inputs.Filter) {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"filter": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	return diff, nil
}

// mapsEqual compares two maps for equality.
// Returns true if both maps are nil or have the same keys and values.
func mapsEqual(a, b map[string]interface{}) bool {
	if a == nil && b == nil {
		return true
	}
	if a == nil || b == nil {
		return false
	}
	if len(a) != len(b) {
		return false
	}
	for k, v := range a {
		if bv, ok := b[k]; !ok || fmt.Sprintf("%v", v) != fmt.Sprintf("%v", bv) {
			return false
		}
	}
	return true
}

// Create creates a new webhook on the Webflow site.
func (w *Webhook) Create(
	ctx context.Context, req infer.CreateRequest[WebhookArgs],
) (infer.CreateResponse[WebhookState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf("validation failed for Webhook resource: %w", err)
	}
	if err := ValidateTriggerType(req.Inputs.TriggerType); err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf("validation failed for Webhook resource: %w", err)
	}
	if err := ValidateWebhookURL(req.Inputs.URL); err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf("validation failed for Webhook resource: %w", err)
	}

	state := WebhookState{
		WebhookArgs:   req.Inputs,
		CreatedOn:     "", // Will be populated from API response
		LastTriggered: "", // Will be populated from API response if available
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set a preview timestamp
		state.CreatedOn = time.Now().Format(time.RFC3339)
		// Generate a predictable ID for dry-run
		previewID := fmt.Sprintf("preview-%d", time.Now().Unix())
		return infer.CreateResponse[WebhookState]{
			ID:     GenerateWebhookResourceID(req.Inputs.SiteID, previewID),
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := PostWebhook(
		ctx, client, req.Inputs.SiteID,
		req.Inputs.TriggerType, req.Inputs.URL, req.Inputs.Filter,
	)
	if err != nil {
		return infer.CreateResponse[WebhookState]{}, fmt.Errorf("failed to create webhook: %w", err)
	}

	// Defensive check: Ensure Webflow API returned a valid webhook ID
	if response.ID == "" {
		return infer.CreateResponse[WebhookState]{}, errors.New(
			"Webflow API returned empty webhook ID - " +
				"this is unexpected and may indicate an API issue")
	}

	// Populate state with API response data
	state.CreatedOn = response.CreatedOn
	state.LastTriggered = response.LastTriggered

	resourceID := GenerateWebhookResourceID(req.Inputs.SiteID, response.ID)

	return infer.CreateResponse[WebhookState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of a webhook from Webflow.
// Used for drift detection and import operations.
func (w *Webhook) Read(
	ctx context.Context, req infer.ReadRequest[WebhookArgs, WebhookState],
) (infer.ReadResponse[WebhookArgs, WebhookState], error) {
	// Extract siteID and webhookID from resource ID
	siteID, webhookID, err := ExtractIDsFromWebhookResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[WebhookArgs, WebhookState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[WebhookArgs, WebhookState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get all webhooks for this site
	response, err := GetWebhooks(ctx, client, siteID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[WebhookArgs, WebhookState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[WebhookArgs, WebhookState]{}, fmt.Errorf("failed to read webhooks: %w", err)
	}

	// Find the specific webhook in the list
	var foundWebhook *WebhookResponse
	for _, webhook := range response.Webhooks {
		if webhook.ID == webhookID {
			foundWebhook = &webhook
			break
		}
	}

	// If webhook not found, return empty ID to signal deletion
	if foundWebhook == nil {
		return infer.ReadResponse[WebhookArgs, WebhookState]{
			ID: "",
		}, nil
	}

	// Build current state from API response
	currentInputs := WebhookArgs{
		SiteID:      siteID,
		TriggerType: foundWebhook.TriggerType,
		URL:         foundWebhook.URL,
		Filter:      foundWebhook.Filter,
	}
	currentState := WebhookState{
		WebhookArgs:   currentInputs,
		CreatedOn:     foundWebhook.CreatedOn,
		LastTriggered: foundWebhook.LastTriggered,
	}

	return infer.ReadResponse[WebhookArgs, WebhookState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update is not supported for webhooks.
// Webflow does not provide an update endpoint for webhooks.
// All changes require replacement (delete + recreate).
func (w *Webhook) Update(
	ctx context.Context, req infer.UpdateRequest[WebhookArgs, WebhookState],
) (infer.UpdateResponse[WebhookState], error) {
	// This should never be called because Diff marks all changes as UpdateReplace
	// But we implement it defensively to return a clear error message
	return infer.UpdateResponse[WebhookState]{}, errors.New(
		"webhooks cannot be updated in-place. " +
			"Webflow does not support updating webhooks - all changes require replacement. " +
			"This is a provider bug if you're seeing this error. " +
			"Please report this issue at https://github.com/jdetmar/pulumi-webflow/issues")
}

// Delete removes a webhook from the Webflow site.
func (w *Webhook) Delete(ctx context.Context, req infer.DeleteRequest[WebhookState]) (infer.DeleteResponse, error) {
	// Extract siteID and webhookID from resource ID
	_, webhookID, err := ExtractIDsFromWebhookResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteWebhook(ctx, client, webhookID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete webhook: %w", err)
	}

	return infer.DeleteResponse{}, nil
}
