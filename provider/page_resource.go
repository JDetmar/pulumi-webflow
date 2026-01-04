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

	"github.com/pulumi/pulumi-go-provider/infer"
)

// PageData is the data source controller for reading Webflow pages.
// Pages cannot be created via the Webflow API - they must be created in the Webflow designer.
// This data source allows you to read page information for use in other resources.
type PageData struct{}

// PageDataArgs defines the input properties for the Page data source.
type PageDataArgs struct {
	// SiteID is the Webflow site ID (24-character lowercase hexadecimal string).
	// Required for listing all pages or looking up a specific page.
	// Example: "5f0c8c9e1c9d440000e8d8c3"
	SiteID string `pulumi:"siteId"`
	// PageID is the specific page ID to retrieve (optional).
	// If specified, retrieves only this page. If omitted, retrieves all pages for the site.
	// Example: "5f0c8c9e1c9d440000e8d8c4"
	PageID string `pulumi:"pageId,optional"`
}

// PageDataState defines the output properties for the Page data source.
type PageDataState struct {
	PageDataArgs
	// WebflowPageID is the Webflow page ID (read-only, only populated when PageID is specified).
	WebflowPageID string `pulumi:"webflowPageId,optional"`
	// Title is the page title (read-only).
	Title string `pulumi:"title,optional"`
	// Slug is the URL slug for the page (read-only).
	Slug string `pulumi:"slug,optional"`
	// ParentID is the ID of the parent page for nested pages (read-only, optional).
	ParentID string `pulumi:"parentId,optional"`
	// CollectionID is the ID of the CMS collection for collection pages (read-only, optional).
	CollectionID string `pulumi:"collectionId,optional"`
	// CreatedOn is the timestamp when the page was created (read-only).
	CreatedOn string `pulumi:"createdOn,optional"`
	// LastUpdated is the timestamp when the page was last updated (read-only).
	LastUpdated string `pulumi:"lastUpdated,optional"`
	// Archived indicates if the page is archived (read-only).
	Archived bool `pulumi:"archived,optional"`
	// Draft indicates if the page is in draft mode (read-only).
	Draft bool `pulumi:"draft,optional"`
	// Pages contains all pages when PageID is not specified (read-only).
	// This is a list of page objects with all their properties.
	Pages []PageInfo `pulumi:"pages,optional"`
}

// PageInfo represents a single page's information in the Pages list.
type PageInfo struct {
	// PageID is the Webflow page ID.
	PageID string `pulumi:"pageId"`
	// SiteID is the Webflow site ID this page belongs to.
	SiteID string `pulumi:"siteId"`
	// Title is the page title.
	Title string `pulumi:"title"`
	// Slug is the URL slug for the page.
	Slug string `pulumi:"slug"`
	// ParentID is the ID of the parent page (optional).
	ParentID string `pulumi:"parentId,optional"`
	// CollectionID is the ID of the CMS collection (optional).
	CollectionID string `pulumi:"collectionId,optional"`
	// CreatedOn is the timestamp when the page was created.
	CreatedOn string `pulumi:"createdOn"`
	// LastUpdated is the timestamp when the page was last updated.
	LastUpdated string `pulumi:"lastUpdated"`
	// Archived indicates if the page is archived.
	Archived bool `pulumi:"archived"`
	// Draft indicates if the page is in draft mode.
	Draft bool `pulumi:"draft"`
}

// Annotate adds descriptions and constraints to the PageData resource.
func (r *PageData) Annotate(a infer.Annotator) {
	a.SetToken("index", "PageData")
	a.Describe(r, "Reads page information from a Webflow site. "+
		"Pages cannot be created via the API - they must be created in the Webflow designer. "+
		"Use this data source to retrieve page metadata for use in your infrastructure code. "+
		"Specify pageId to get a single page, or omit it to list all pages in the site.")
}

// Annotate adds descriptions to the PageDataArgs fields.
func (args *PageDataArgs) Annotate(a infer.Annotator) {
	a.Describe(&args.SiteID,
		"The Webflow site ID (24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c3'). "+
			"You can find your site ID in the Webflow dashboard under Site Settings. "+
			"This field will be validated before making any API calls.")

	a.Describe(&args.PageID,
		"The specific page ID to retrieve (optional, 24-character lowercase hexadecimal string, "+
			"e.g., '5f0c8c9e1c9d440000e8d8c4'). "+
			"If specified, only this page's data will be returned. "+
			"If omitted, all pages for the site will be returned in the 'pages' output.")
}

// Annotate adds descriptions to the PageDataState fields.
func (state *PageDataState) Annotate(a infer.Annotator) {
	a.Describe(&state.WebflowPageID,
		"The Webflow page ID (read-only). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.Title,
		"The page title (read-only). "+
			"This is the title shown in browser tabs and search results. "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.Slug,
		"The URL slug for the page (read-only, e.g., 'about' for '/about'). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.ParentID,
		"The ID of the parent page (read-only, optional). "+
			"Only present for nested pages. "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.CollectionID,
		"The ID of the CMS collection (read-only, optional). "+
			"Only present for collection pages. "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.CreatedOn,
		"The timestamp when the page was created (read-only, RFC3339 format). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.LastUpdated,
		"The timestamp when the page was last updated (read-only, RFC3339 format). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.Archived,
		"Indicates if the page is archived (read-only). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.Draft,
		"Indicates if the page is in draft mode (read-only). "+
			"Only populated when pageId input is specified.")

	a.Describe(&state.Pages,
		"List of all pages in the site (read-only). "+
			"Only populated when pageId input is NOT specified. "+
			"Each page includes all metadata fields: id, siteId, title, slug, "+
			"createdOn, lastUpdated, archived, draft, and optional parentId/collectionId.")
}

// Read retrieves page information from Webflow.
// This is the primary operation for a data source.
func (r *PageData) Read(
	ctx context.Context, req infer.ReadRequest[PageDataArgs, PageDataState],
) (infer.ReadResponse[PageDataArgs, PageDataState], error) {
	// Validate siteID
	if err := ValidateSiteID(req.State.SiteID); err != nil {
		return infer.ReadResponse[PageDataArgs, PageDataState]{}, fmt.Errorf("validation failed for PageData: %w", err)
	}

	// Get HTTP client
	client, err := GetHTTPClient(ctx, providerVersion)
	if err != nil {
		return infer.ReadResponse[PageDataArgs, PageDataState]{}, fmt.Errorf("failed to create HTTP client: %w", err)
	}

	// If pageID is specified, get single page
	if req.State.PageID != "" {
		// Validate pageID
		if err := ValidatePageID(req.State.PageID); err != nil {
			return infer.ReadResponse[PageDataArgs, PageDataState]{}, fmt.Errorf("validation failed for PageData: %w", err)
		}

		// Call Webflow API to get single page
		page, err := GetPage(ctx, client, req.State.PageID)
		if err != nil {
			// If page not found, return empty ID to signal deletion
			if errors.Is(err, errors.New("not found")) {
				return infer.ReadResponse[PageDataArgs, PageDataState]{
					ID: "",
				}, nil
			}
			return infer.ReadResponse[PageDataArgs, PageDataState]{}, fmt.Errorf("failed to read page: %w", err)
		}

		// Build state from API response
		currentInputs := PageDataArgs{
			SiteID: req.State.SiteID,
			PageID: req.State.PageID,
		}
		currentState := PageDataState{
			PageDataArgs:  currentInputs,
			WebflowPageID: page.ID,
			Title:         page.Title,
			Slug:         page.Slug,
			ParentID:     page.ParentID,
			CollectionID: page.CollectionID,
			CreatedOn:    page.CreatedOn,
			LastUpdated:  page.LastUpdated,
			Archived:     page.Archived,
			Draft:        page.Draft,
		}

		return infer.ReadResponse[PageDataArgs, PageDataState]{
			ID:     req.ID,
			Inputs: currentInputs,
			State:  currentState,
		}, nil
	}

	// If pageID is not specified, list all pages
	response, err := GetPages(ctx, client, req.State.SiteID)
	if err != nil {
		return infer.ReadResponse[PageDataArgs, PageDataState]{}, fmt.Errorf("failed to list pages: %w", err)
	}

	// Convert API pages to PageInfo
	pages := make([]PageInfo, len(response.Pages))
	for i, p := range response.Pages {
		pages[i] = PageInfo{
			PageID:       p.ID,
			SiteID:       p.SiteID,
			Title:        p.Title,
			Slug:         p.Slug,
			ParentID:     p.ParentID,
			CollectionID: p.CollectionID,
			CreatedOn:    p.CreatedOn,
			LastUpdated:  p.LastUpdated,
			Archived:     p.Archived,
			Draft:        p.Draft,
		}
	}

	// Build state from API response
	currentInputs := PageDataArgs{
		SiteID: req.State.SiteID,
		PageID: "", // No specific page ID
	}
	currentState := PageDataState{
		PageDataArgs: currentInputs,
		Pages:        pages,
	}

	return infer.ReadResponse[PageDataArgs, PageDataState]{
		ID:     req.ID,
		Inputs: currentInputs,
		State:  currentState,
	}, nil
}

// Create is not supported for PageData (data sources are read-only).
// This satisfies the infer.CustomResource interface but will return an error if called.
func (r *PageData) Create(
	ctx context.Context, req infer.CreateRequest[PageDataArgs],
) (infer.CreateResponse[PageDataState], error) {
	return infer.CreateResponse[PageDataState]{}, errors.New(
		"PageData is a read-only data source. " +
			"Pages cannot be created via the Webflow API - they must be created in the Webflow designer. " +
			"Use this data source to read existing pages only.")
}

// Update is not supported for PageData (data sources are read-only).
func (r *PageData) Update(
	ctx context.Context, req infer.UpdateRequest[PageDataArgs, PageDataState],
) (infer.UpdateResponse[PageDataState], error) {
	return infer.UpdateResponse[PageDataState]{}, errors.New(
		"PageData is a read-only data source. " +
			"Pages cannot be updated via this data source. " +
			"To modify pages, use the Webflow designer.")
}

// Delete is not supported for PageData (data sources are read-only).
func (r *PageData) Delete(ctx context.Context, req infer.DeleteRequest[PageDataState]) (infer.DeleteResponse, error) {
	// For data sources, delete is a no-op (we don't actually delete anything in Webflow)
	return infer.DeleteResponse{}, nil
}
