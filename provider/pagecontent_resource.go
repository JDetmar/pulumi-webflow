// Copyright 2025, Justin Detmar.
// SPDX-License-Identifier: MIT
//
// This is an unofficial, community-maintained Pulumi provider for Webflow.
// Not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.

package provider

import (
	"context"
	"fmt"
	"time"

	p "github.com/pulumi/pulumi-go-provider"
	"github.com/pulumi/pulumi-go-provider/infer"
)

// PageContent is the resource controller for managing Webflow page content.
// It allows updating static content (text) within existing page DOM nodes.
// Note: This does NOT manage page structure/layout, only content within existing nodes.
type PageContent struct{}

// NodeContentUpdate represents a single node content update.
type NodeContentUpdate struct {
	// NodeID is the unique identifier for the DOM node to update (required).
	NodeID string `pulumi:"nodeId"`
	// Text is the new text content for the node (required for text nodes).
	Text string `pulumi:"text"`
}

// PageContentArgs defines the input properties for the PageContent resource.
type PageContentArgs struct {
	// PageID is the Webflow page ID to update (24-character lowercase hexadecimal string).
	// Example: "5f0c8c9e1c9d440000e8d8c4"
	PageID string `pulumi:"pageId"`
	// Nodes is the list of node content updates to apply.
	// Each node update specifies the nodeId and the new text content.
	Nodes []NodeContentUpdate `pulumi:"nodes"`
}

// PageContentState defines the output properties for the PageContent resource.
// It embeds PageContentArgs to include input properties in the output.
type PageContentState struct {
	PageContentArgs
	// LastUpdated is the timestamp when the content was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
}

// Annotate adds descriptions and constraints to the PageContent resource.
func (r *PageContent) Annotate(a infer.Annotator) {
	a.SetToken("index", "PageContent")
	a.Describe(r, "Manages static content (text) for a Webflow page. "+
		"This resource allows you to update text content within existing DOM nodes on a page. "+
		"It does NOT manage page structure or layout - only content within existing nodes. "+
		"To find node IDs, you must first retrieve the page DOM structure using the Webflow API. "+
		"\n\n**IMPORTANT LIMITATION:** This resource does NOT support drift detection for content changes. "+
		"If content is modified outside of Pulumi (via Webflow UI or API), those changes will NOT be detected "+
		"during 'pulumi refresh' or 'pulumi up'. The resource only verifies that the page still exists. "+
		"This is due to the complexity of extracting and comparing specific node text from the full DOM structure.")
}

// Annotate adds descriptions to the PageContentArgs fields.
func (args *PageContentArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.PageID,
		"The Webflow page ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c4'). "+
			"You can find page IDs using the Pages API list endpoint or in the Webflow designer. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.Nodes,
		"List of node content updates to apply. "+
			"Each update specifies the nodeId (from the page's DOM structure) and the new text content. "+
			"Node IDs can be retrieved by fetching the page DOM using GET /pages/{page_id}/dom. "+
			"Only text content in existing nodes can be updated via this resource.")
}

// Annotate adds descriptions to NodeContentUpdate fields.
func (ncu *NodeContentUpdate) Annotate(a infer.Annotator) {
	a.Describe(&ncu.NodeID,
		"The unique identifier for the DOM node to update. "+
			"This ID comes from the page's DOM structure and must exist on the page. "+
			"Retrieve node IDs using GET /pages/{page_id}/dom endpoint.")

	a.Describe(&ncu.Text,
		"The new text content for the node. "+
			"This will replace the existing text content in the specified node. "+
			"Only applicable to text nodes or elements containing text.")
}

// Annotate adds descriptions to the PageContentState fields.
func (state *PageContentState) Annotate(a infer.Annotator) {
	a.Describe(&state.LastUpdated,
		"The timestamp when the page content was last updated (RFC3339 format). "+
			"This is automatically set when content is updated and is read-only.")
}

// Diff determines what changes need to be made to the page content resource.
// PageID changes trigger replacement (different page).
// Nodes changes trigger in-place update.
func (r *PageContent) Diff(
	ctx context.Context, req infer.DiffRequest[PageContentArgs, PageContentState],
) (infer.DiffResponse, error) {
	diff := infer.DiffResponse{}

	// Check for pageId change (requires replacement - different page)
	if req.State.PageID != req.Inputs.PageID {
		diff.DeleteBeforeReplace = true
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"pageId": {Kind: p.UpdateReplace},
		}
		return diff, nil
	}

	// Check for nodes changes (in-place update)
	// Compare node counts first
	if len(req.State.Nodes) != len(req.Inputs.Nodes) {
		diff.HasChanges = true
		diff.DetailedDiff = map[string]p.PropertyDiff{
			"nodes": {Kind: p.Update},
		}
		return diff, nil
	}

	// Compare individual nodes
	// Create maps for easier comparison
	stateNodes := make(map[string]string)
	for _, node := range req.State.Nodes {
		stateNodes[node.NodeID] = node.Text
	}

	inputNodes := make(map[string]string)
	for _, node := range req.Inputs.Nodes {
		inputNodes[node.NodeID] = node.Text
	}

	// Check if any nodes changed
	for nodeID, inputText := range inputNodes {
		stateText, exists := stateNodes[nodeID]
		if !exists || stateText != inputText {
			diff.HasChanges = true
			diff.DetailedDiff = map[string]p.PropertyDiff{
				"nodes": {Kind: p.Update},
			}
			return diff, nil
		}
	}

	// Check if any nodes were removed
	for nodeID := range stateNodes {
		if _, exists := inputNodes[nodeID]; !exists {
			diff.HasChanges = true
			diff.DetailedDiff = map[string]p.PropertyDiff{
				"nodes": {Kind: p.Update},
			}
			return diff, nil
		}
	}

	return diff, nil
}

// Create updates page content by applying the specified node updates.
// Note: PageContent is a configuration resource - "create" means "apply this configuration".
func (r *PageContent) Create(
	ctx context.Context, req infer.CreateRequest[PageContentArgs],
) (infer.CreateResponse[PageContentState], error) {
	// Validate inputs BEFORE generating resource ID
	if err := ValidatePageID(req.Inputs.PageID); err != nil {
		return infer.CreateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource: %w", err)
	}

	// Validate nodes
	if len(req.Inputs.Nodes) == 0 {
		return infer.CreateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource: "+
			"at least one node update is required. "+
			"Please provide a list of nodes with nodeId and text fields. "+
			"Node IDs can be retrieved using GET /pages/{page_id}/dom endpoint.")
	}

	for i, node := range req.Inputs.Nodes {
		if err := ValidateNodeID(node.NodeID); err != nil {
			return infer.CreateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource, node[%d]: %w", i, err)
		}
		if node.Text == "" {
			return infer.CreateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource, node[%d]: "+
				"text is required but was not provided. "+
				"Please provide the new text content for nodeId '%s'.", i, node.NodeID)
		}
	}

	state := PageContentState{
		PageContentArgs: req.Inputs,
		LastUpdated:     "", // Will be populated after update
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		// Set a preview timestamp
		state.LastUpdated = time.Now().Format(time.RFC3339)
		// Generate resource ID
		resourceID := GeneratePageContentResourceID(req.Inputs.PageID)
		return infer.CreateResponse[PageContentState]{
			ID:     resourceID,
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.CreateResponse[PageContentState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Convert nodes to API format
	nodeUpdates := make([]DOMNodeUpdate, len(req.Inputs.Nodes))
	for i, node := range req.Inputs.Nodes {
		nodeUpdates[i] = DOMNodeUpdate{
			NodeID: node.NodeID,
			Text:   &node.Text,
		}
	}

	// Call Webflow API to update page content
	_, err = PutPageContent(ctx, client, req.Inputs.PageID, nodeUpdates)
	if err != nil {
		return infer.CreateResponse[PageContentState]{}, fmt.Errorf("failed to update page content: %w", err)
	}

	// Set update timestamp
	state.LastUpdated = time.Now().Format(time.RFC3339)

	resourceID := GeneratePageContentResourceID(req.Inputs.PageID)

	return infer.CreateResponse[PageContentState]{
		ID:     resourceID,
		Output: state,
	}, nil
}

// Read retrieves the current state of page content from Webflow.
// Used for drift detection and refresh operations.
func (r *PageContent) Read(
	ctx context.Context, req infer.ReadRequest[PageContentArgs, PageContentState],
) (infer.ReadResponse[PageContentArgs, PageContentState], error) {
	// Extract pageID from resource ID
	pageID, err := ExtractPageIDFromPageContentResourceID(req.ID)
	if err != nil {
		return infer.ReadResponse[PageContentArgs, PageContentState]{}, fmt.Errorf("invalid resource ID: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[PageContentArgs, PageContentState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Call Webflow API to get current page content
	response, err := GetPageContent(ctx, client, pageID)
	if err != nil {
		// If page not found, return empty ID to signal deletion
		// Note: We can't easily detect if specific nodes exist without deep traversal
		// For now, if the page is gone, consider the resource gone
		return infer.ReadResponse[PageContentArgs, PageContentState]{
			ID: "",
		}, nil
	}

	// Build current state
	// DRIFT DETECTION LIMITATION:
	// We preserve the configured nodes from state instead of extracting them from the API response.
	// This means drift detection does NOT work for content changes made outside of Pulumi.
	// Extracting and comparing specific node text from the full DOM structure would require:
	// 1. Complex recursive traversal of the entire DOM tree
	// 2. Matching node IDs to their current text values
	// 3. Handling edge cases (deleted nodes, moved nodes, nested structures)
	// For now, we only verify that the page itself still exists (basic drift check).
	currentInputs := PageContentArgs{
		PageID: pageID,
		Nodes:  req.State.Nodes, // Preserve configured nodes (NOT read from API)
	}
	currentState := PageContentState{
		PageContentArgs: currentInputs,
		LastUpdated:     req.State.LastUpdated, // Preserve timestamp
	}

	// Verify the page still exists (basic check)
	if response.PageID == "" {
		return infer.ReadResponse[PageContentArgs, PageContentState]{
			ID: "",
		}, nil
	}

	return infer.ReadResponse[PageContentArgs, PageContentState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Update modifies existing page content.
func (r *PageContent) Update(
	ctx context.Context, req infer.UpdateRequest[PageContentArgs, PageContentState],
) (infer.UpdateResponse[PageContentState], error) {
	// Validate inputs BEFORE making API calls
	if err := ValidatePageID(req.Inputs.PageID); err != nil {
		return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource: %w", err)
	}

	// Validate nodes
	if len(req.Inputs.Nodes) == 0 {
		return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource: "+
			"at least one node update is required. "+
			"Please provide a list of nodes with nodeId and text fields.")
	}

	for i, node := range req.Inputs.Nodes {
		if err := ValidateNodeID(node.NodeID); err != nil {
			return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource, node[%d]: %w", i, err)
		}
		if node.Text == "" {
			return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("validation failed for PageContent resource, node[%d]: "+
				"text is required but was not provided. "+
				"Please provide the new text content for nodeId '%s'.", i, node.NodeID)
		}
	}

	state := PageContentState{
		PageContentArgs: req.Inputs,
		LastUpdated:     "", // Will be updated after API call
	}

	// During preview, return expected state without making API calls
	if req.DryRun {
		state.LastUpdated = time.Now().Format(time.RFC3339)
		return infer.UpdateResponse[PageContentState]{
			Output: state,
		}, nil
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// Convert nodes to API format
	nodeUpdates := make([]DOMNodeUpdate, len(req.Inputs.Nodes))
	for i, node := range req.Inputs.Nodes {
		nodeUpdates[i] = DOMNodeUpdate{
			NodeID: node.NodeID,
			Text:   &node.Text,
		}
	}

	// Call Webflow API to update page content
	_, err = PutPageContent(ctx, client, req.Inputs.PageID, nodeUpdates)
	if err != nil {
		return infer.UpdateResponse[PageContentState]{}, fmt.Errorf("failed to update page content: %w", err)
	}

	// Set update timestamp
	state.LastUpdated = time.Now().Format(time.RFC3339)

	return infer.UpdateResponse[PageContentState]{
		Output: state,
	}, nil
}

// Delete removes the page content configuration.
// Note: For PageContent, delete is a no-op since we don't actually delete content from the page.
// The content remains on the page - we just stop managing it via Pulumi.
func (r *PageContent) Delete(ctx context.Context, req infer.DeleteRequest[PageContentState]) (infer.DeleteResponse, error) {
	// PageContent is a configuration resource - deleting it just means we stop managing it.
	// We don't actually delete content from the page (that would break the page).
	// This is idempotent and always succeeds.
	return infer.DeleteResponse{}, nil
}
