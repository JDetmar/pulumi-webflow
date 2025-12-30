// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

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

// RobotsTxt is the resource controller for managing robots.txt configuration.
// It implements the infer.CustomResource interface for full CRUD operations.
type RobotsTxt struct{}

// RobotsTxtArgs defines the input properties for the RobotsTxt resource.
type RobotsTxtArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	SiteID string `pulumi:"siteId"`
	// Content is the robots.txt content in traditional format.
	Content string `pulumi:"content"`
}

// RobotsTxtState defines the output properties for the RobotsTxt resource.
// It embeds RobotsTxtArgs to include input properties in the output.
type RobotsTxtState struct {
	RobotsTxtArgs
	// LastModified is the RFC3339 timestamp of the last modification.
	LastModified string `pulumi:"lastModified"`
}

// Annotate adds descriptions and constraints to the RobotsTxt resource.
func (r *RobotsTxt) Annotate(a infer.Annotator) {
	a.SetToken("index", "RobotsTxt")
	a.Describe(r, "Manages robots.txt configuration for a Webflow site. "+
		"This resource allows you to define crawler access rules and sitemap references.")
}

// Annotate adds descriptions to the RobotsTxtArgs fields.
func (args *RobotsTxtArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3').")
	a.Describe(&args.Content, "The robots.txt content in traditional format. "+
		"Supports User-agent, Allow, Disallow, and Sitemap directives.")
}

// Annotate adds descriptions to the RobotsTxtState fields.
func (state *RobotsTxtState) Annotate(a infer.Annotator) {
	a.Describe(&state.LastModified, "RFC3339 timestamp of the last modification.")
}

// Diff determines what changes need to be made to the resource.
// siteId changes trigger replacement; content changes trigger update.
func (r *RobotsTxt) Diff(
	ctx context.Context, req infer.DiffRequest[RobotsTxtArgs, RobotsTxtState],
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

	// Check for content change (in-place update)
	if req.State.Content != req.Inputs.Content {
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"content": {Kind: p.Update},
		}
	}

	return diff, nil
}

// Create creates a new robots.txt configuration on the Webflow site.
func (r *RobotsTxt) Create(
	ctx context.Context, req infer.CreateRequest[RobotsTxtArgs],
) (infer.CreateResponse[RobotsTxtState], error) {
	// Validate inputs BEFORE generating resource ID (validation happens before API calls)
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.CreateResponse[RobotsTxtState]{},
			fmt.Errorf("validation failed for RobotsTxt resource: %w", err)
	}
	if req.Inputs.Content == "" {
		return infer.CreateResponse[RobotsTxtState]{}, errors.New(
			"validation failed for RobotsTxt resource: " +
				"content is required but was not provided. " +
				"Please provide robots.txt content with at least one directive " +
				"(e.g., 'User-agent: *\\nAllow: /'). " +
				"The content should follow the traditional robots.txt format " +
				"with User-agent, Allow, Disallow, and Sitemap directives.")
	}

	state := RobotsTxtState{
		RobotsTxtArgs: req.Inputs,
		LastModified:  time.Now().UTC().Format(time.RFC3339),
	}
	resourceID := GenerateRobotsTxtResourceID(req.Inputs.SiteID)

	// During preview, return expected state without making API calls
	if req.DryRun {
		return infer.CreateResponse[RobotsTxtState]{
			ID:     resourceID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[RobotsTxtState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Parse content to structured format
	rules, sitemap := ParseRobotsTxtContent(req.Inputs.Content)

	// Call Webflow API
	response, err := PutRobotsTxt(ctx, client, req.Inputs.SiteID, rules, sitemap)
	if err != nil {
		return infer.CreateResponse[RobotsTxtState]{}, fmt.Errorf("failed to create robots.txt: %w", err)
	}

	// Update state with response
	state.Content = FormatRobotsTxtContent(response.Rules, response.Sitemap)
	state.LastModified = time.Now().UTC().Format(time.RFC3339)

	return infer.CreateResponse[RobotsTxtState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of the robots.txt from Webflow.
// Used for drift detection and import operations.
func (r *RobotsTxt) Read(
	ctx context.Context, req infer.ReadRequest[RobotsTxtArgs, RobotsTxtState],
) (infer.ReadResponse[RobotsTxtArgs, RobotsTxtState], error) {
	// Extract siteID from resource ID
	siteID, err := ExtractSiteIDFromResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[RobotsTxtArgs, RobotsTxtState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[RobotsTxtArgs, RobotsTxtState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API
	response, err := GetRobotsTxt(ctx, client, siteID)
	if err != nil {
		// Resource not found - return empty ID to signal deletion
		if strings.Contains(err.Error(), "not found") {
			return infer.ReadResponse[RobotsTxtArgs, RobotsTxtState]{
				ID: "",
			}, nil
		}
		return infer.ReadResponse[RobotsTxtArgs, RobotsTxtState]{}, fmt.Errorf("failed to read robots.txt: %w", err)
	}

	// Build current state from API response
	content := FormatRobotsTxtContent(response.Rules, response.Sitemap)
	currentInputs := RobotsTxtArgs{
		SiteID:  siteID,
		Content: content,
	}
	currentState := RobotsTxtState{
		RobotsTxtArgs: currentInputs,
		// Preserve LastModified from existing state - don't regenerate it
		// The Webflow API doesn't return a last modified timestamp, and regenerating
		// it on every Read() causes false drift detection
		LastModified: req.State.LastModified,
	}

	return infer.ReadResponse[RobotsTxtArgs, RobotsTxtState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies an existing robots.txt configuration.
func (r *RobotsTxt) Update(
	ctx context.Context, req infer.UpdateRequest[RobotsTxtArgs, RobotsTxtState],
) (infer.UpdateResponse[RobotsTxtState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidateSiteID(req.Inputs.SiteID); err != nil {
		return infer.UpdateResponse[RobotsTxtState]{},
			fmt.Errorf("validation failed for RobotsTxt resource: %w", err)
	}
	if req.Inputs.Content == "" {
		return infer.UpdateResponse[RobotsTxtState]{}, errors.New(
			"validation failed for RobotsTxt resource: " +
				"content is required but was not provided. " +
				"Please provide robots.txt content with at least one directive " +
				"(e.g., 'User-agent: *\\nAllow: /'). " +
				"The content should follow the traditional robots.txt format " +
				"with User-agent, Allow, Disallow, and Sitemap directives.")
	}

	state := RobotsTxtState{
		RobotsTxtArgs: req.Inputs,
		LastModified:  time.Now().UTC().Format(time.RFC3339),
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		return infer.UpdateResponse[RobotsTxtState]{
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[RobotsTxtState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Parse content to structured format
	rules, sitemap := ParseRobotsTxtContent(req.Inputs.Content)

	// Call Webflow API
	response, err := PutRobotsTxt(ctx, client, req.Inputs.SiteID, rules, sitemap)
	if err != nil {
		return infer.UpdateResponse[RobotsTxtState]{}, fmt.Errorf("failed to update robots.txt: %w", err)
	}

	// Update state with response
	state.Content = FormatRobotsTxtContent(response.Rules, response.Sitemap)
	state.LastModified = time.Now().UTC().Format(time.RFC3339)

	return infer.UpdateResponse[RobotsTxtState]{
		Output: state,
	}, nil
}

// Delete removes the robots.txt configuration from the Webflow site.
func (r *RobotsTxt) Delete(ctx context.Context, req infer.DeleteRequest[RobotsTxtState]) (infer.DeleteResponse, error) {
	// Extract siteID from resource ID
	siteID, err := ExtractSiteIDFromResourceID(req.ID)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API (handles 404 gracefully for idempotency)
	if err := DeleteRobotsTxt(ctx, client, siteID); err != nil {
		return infer.DeleteResponse{}, fmt.Errorf("failed to delete robots.txt: %w", err)
	}

	return infer.DeleteResponse{}, nil
}

// providerVersion is set during provider initialization.
// This is a package-level variable that gets set when the provider starts.
var providerVersion = "0.0.0"

// SetProviderVersion sets the provider version for use in API calls.
func SetProviderVersion(version string) {
	providerVersion = version
}
