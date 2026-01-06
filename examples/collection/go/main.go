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

		// Example 1: Blog Posts Collection
		// A common pattern for blog content with all required fields
		blogCollection, err := webflow.NewCollection(ctx, "blog-posts-collection", &webflow.CollectionArgs{
			SiteId:       siteID,
			DisplayName:  pulumi.String("Blog Posts"),
			SingularName: pulumi.String("Blog Post"),
			Slug:         pulumi.String("blog-posts"),
		})
		if err != nil {
			return fmt.Errorf("failed to create blog collection: %w", err)
		}

		// Example 2: Products Collection with Auto-Generated Slug
		// Omit the slug to let Webflow auto-generate it from the displayName
		productsCollection, err := webflow.NewCollection(ctx, "products-collection", &webflow.CollectionArgs{
			SiteId:       siteID,
			DisplayName:  pulumi.String("Products"),
			SingularName: pulumi.String("Product"),
			// slug is optional - Webflow will generate automatically
		})
		if err != nil {
			return fmt.Errorf("failed to create products collection: %w", err)
		}

		// Example 3: Team Members Collection
		// Demonstrates custom slug different from display name
		teamCollection, err := webflow.NewCollection(ctx, "team-members-collection", &webflow.CollectionArgs{
			SiteId:       siteID,
			DisplayName:  pulumi.String("Team Members"),
			SingularName: pulumi.String("Team Member"),
			Slug:         pulumi.String("team"),
		})
		if err != nil {
			return fmt.Errorf("failed to create team collection: %w", err)
		}

		// Example 4: Portfolio Items Collection
		// Another common use case for showcasing work
		portfolioCollection, err := webflow.NewCollection(ctx, "portfolio-collection", &webflow.CollectionArgs{
			SiteId:       siteID,
			DisplayName:  pulumi.String("Portfolio Items"),
			SingularName: pulumi.String("Portfolio Item"),
			Slug:         pulumi.String("portfolio"),
		})
		if err != nil {
			return fmt.Errorf("failed to create portfolio collection: %w", err)
		}

		// Example 5: Dynamic Collections Based on Config
		// Create collections based on configuration for multi-environment setups
		testCollection, err := webflow.NewCollection(ctx, fmt.Sprintf("test-collection-%s", environment), &webflow.CollectionArgs{
			SiteId:       siteID,
			DisplayName:  pulumi.String(fmt.Sprintf("Test Collection (%s)", environment)),
			SingularName: pulumi.String("Test Item"),
			Slug:         pulumi.String(fmt.Sprintf("test-%s", environment)),
		})
		if err != nil {
			return fmt.Errorf("failed to create test collection: %w", err)
		}

		// Export collection details for reference
		ctx.Export("deployedSiteId", siteID)
		ctx.Export("blogCollectionId", blogCollection.ID())
		ctx.Export("blogCollectionName", blogCollection.DisplayName)
		ctx.Export("blogCollectionSlug", blogCollection.Slug)
		ctx.Export("blogCollectionCreatedOn", blogCollection.CreatedOn)

		ctx.Export("productsCollectionId", productsCollection.ID())
		ctx.Export("teamCollectionId", teamCollection.ID())
		ctx.Export("portfolioCollectionId", portfolioCollection.ID())
		ctx.Export("testCollectionId", testCollection.ID())

		// Export a summary of all collections
		ctx.Export("allCollections", pulumi.All(
			blogCollection.DisplayName,
			productsCollection.DisplayName,
			teamCollection.DisplayName,
			portfolioCollection.DisplayName,
			testCollection.DisplayName,
		).ApplyT(func(args []interface{}) string {
			names := make([]string, len(args))
			for i, arg := range args {
				names[i] = arg.(string)
			}
			result := ""
			for i, name := range names {
				if i > 0 {
					result += ", "
				}
				result += name
			}
			return result
		}))

		ctx.Log.Info(
			"✅ Successfully deployed 5 collections",
			&pulumi.LogOptions{},
		)

		return nil
	})
}
