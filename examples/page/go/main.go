package main

import (
	"fmt"

	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		siteID := cfg.RequireSecret("siteId")
		pageID := cfg.Get("pageId") // Optional: set to get a specific page

		// Example 1: Get all pages for a site
		allPages, err := webflow.NewPageData(ctx, "all-pages", &webflow.PageDataArgs{
			SiteId: siteID,
		})
		if err != nil {
			return fmt.Errorf("failed to get all pages: %w", err)
		}

		// Example 2: Get a specific page by ID (conditional on config)
		var specificPage *webflow.PageData
		if pageID != "" {
			specificPage, err = webflow.NewPageData(ctx, "specific-page", &webflow.PageDataArgs{
				SiteId: siteID,
				PageId: pulumi.StringPtr(pageID),
			})
			if err != nil {
				return fmt.Errorf("failed to get specific page: %w", err)
			}
		}

		// Export outputs for all pages scenario
		ctx.Export("sitePages", allPages.Pages.ApplyT(func(pages []webflow.PageInfo) interface{} {
			result := make([]map[string]interface{}, len(pages))
			for i, page := range pages {
				result[i] = map[string]interface{}{
					"id":       page.PageId,
					"title":    page.Title,
					"slug":     page.Slug,
					"draft":    page.Draft,
					"archived": page.Archived,
				}
			}
			return result
		}))

		ctx.Export("pageCount", allPages.Pages.ApplyT(func(pages []webflow.PageInfo) int {
			return len(pages)
		}))

		ctx.Export("pageIds", allPages.Pages.ApplyT(func(pages []webflow.PageInfo) []string {
			ids := make([]string, len(pages))
			for i, page := range pages {
				ids[i] = page.PageId
			}
			return ids
		}))

		// Export outputs for specific page scenario (if configured)
		if specificPage != nil {
			ctx.Export("pageTitle", specificPage.Title)
			ctx.Export("pageSlug", specificPage.Slug)
			ctx.Export("pageWebflowId", specificPage.WebflowPageId)
			ctx.Export("pageCreatedOn", specificPage.CreatedOn)
			ctx.Export("pageLastUpdated", specificPage.LastUpdated)
			ctx.Export("pageIsDraft", specificPage.Draft)
			ctx.Export("pageIsArchived", specificPage.Archived)
			ctx.Export("pageParentId", specificPage.ParentId)
			ctx.Export("pageCollectionId", specificPage.CollectionId)
		}

		// Print helpful information
		allPages.Pages.ApplyT(func(pages []webflow.PageInfo) interface{} {
			ctx.Log.Info(fmt.Sprintf("\n📄 Found %d pages in the site", len(pages)), nil)

			// Show a sample of pages
			sampleSize := len(pages)
			if sampleSize > 5 {
				sampleSize = 5
			}

			if sampleSize > 0 {
				ctx.Log.Info(fmt.Sprintf("\nFirst %d pages:", sampleSize), nil)
				for i := 0; i < sampleSize; i++ {
					ctx.Log.Info(fmt.Sprintf("  %d. \"%s\" (/%s)", i+1, pages[i].Title, pages[i].Slug), nil)
				}

				if len(pages) > sampleSize {
					ctx.Log.Info(fmt.Sprintf("  ... and %d more", len(pages)-sampleSize), nil)
				}
			}
			return nil
		})

		if specificPage != nil {
			specificPage.Title.ApplyT(func(title string) interface{} {
				ctx.Log.Info(fmt.Sprintf("\n✅ Retrieved page: \"%s\"", title), nil)
				return nil
			})
		}

		return nil
	})
}
