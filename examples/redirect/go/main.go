package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		siteID := cfg.RequireSecret("siteId")
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "development"
		}

		// Example 1: Permanent Redirect (301)
		permanentRedirect, err := webflow.NewRedirect(ctx, "old-blog-to-new-blog", &webflow.RedirectArgs{
			SiteId:          siteID,
			SourcePath:      pulumi.String("/blog/old-article"),
			DestinationPath: pulumi.String("/blog/articles/updated-article"),
			StatusCode:      pulumi.Int(301),
		})
		if err != nil {
			return fmt.Errorf("failed to create permanent redirect: %w", err)
		}

		// Example 2: Temporary Redirect (302)
		temporaryRedirect, err := webflow.NewRedirect(ctx, "temporary-landing-page", &webflow.RedirectArgs{
			SiteId:          siteID,
			SourcePath:      pulumi.String("/old-campaign"),
			DestinationPath: pulumi.String("/new-campaign-2025"),
			StatusCode:      pulumi.Int(302),
		})
		if err != nil {
			return fmt.Errorf("failed to create temporary redirect: %w", err)
		}

		// Example 3: External Redirect
		externalRedirect, err := webflow.NewRedirect(ctx, "external-partner-link", &webflow.RedirectArgs{
			SiteId:          siteID,
			SourcePath:      pulumi.String("/partner"),
			DestinationPath: pulumi.String("https://partner-site.com"),
			StatusCode:      pulumi.Int(301),
		})
		if err != nil {
			return fmt.Errorf("failed to create external redirect: %w", err)
		}

		// Example 4: Bulk Redirects
		redirectMappings := []struct {
			old string
			new string
		}{
			{"/product-a", "/products/product-a"},
			{"/product-b", "/products/product-b"},
			{"/product-c", "/products/product-c"},
		}

		bulkRedirectIds := []pulumi.StringOutput{}
		for i, mapping := range redirectMappings {
			redirect, err := webflow.NewRedirect(ctx, fmt.Sprintf("bulk-redirect-%d", i), &webflow.RedirectArgs{
				SiteId:          siteID,
				SourcePath:      pulumi.String(mapping.old),
				DestinationPath: pulumi.String(mapping.new),
				StatusCode:      pulumi.Int(301),
			})
			if err != nil {
				return fmt.Errorf("failed to create bulk redirect %d: %w", i, err)
			}
			bulkRedirectIds = append(bulkRedirectIds, redirect.ID().ToStringOutput())
		}

		// Export values
		ctx.Export("deployedSiteId", siteID)
		ctx.Export("permanentRedirectId", permanentRedirect.ID())
		ctx.Export("temporaryRedirectId", temporaryRedirect.ID())
		ctx.Export("externalRedirectId", externalRedirect.ID())
		ctx.Export("bulkRedirectIds", pulumi.StringArray(bulkRedirectIds))

		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully deployed %d redirects in %s environment", len(redirectMappings)+3, environment),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
