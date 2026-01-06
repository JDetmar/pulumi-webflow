package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		collectionID := cfg.Require("collectionId")
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "development"
		}

		// Example 1: Draft Blog Post
		// Create a blog post that exists in the CMS but is NOT published to the live site
		isDraft := true
		isNotArchived := false
		draftBlogPost, err := webflow.NewCollectionItem(ctx, "draft-blog-post", &webflow.CollectionItemArgs{
			CollectionId: pulumi.String(collectionID),
			FieldData: pulumi.Map{
				"name": pulumi.String("Getting Started with Webflow CMS"),
				"slug": pulumi.String("getting-started-webflow-cms"),
				// Add your custom fields here based on your collection schema
				// Example custom fields (uncomment and modify based on your schema):
				// "post-body": pulumi.String("Learn how to use Webflow CMS to manage your content..."),
				// "author": pulumi.String("John Doe"),
				// "publish-date": pulumi.String("2025-01-06"),
				// "featured-image": pulumi.String("https://example.com/image.jpg"),
			},
			IsDraft:    &isDraft,
			IsArchived: &isNotArchived,
		})
		if err != nil {
			return fmt.Errorf("failed to create draft blog post: %w", err)
		}

		// Example 2: Published Product
		// Create a product that is immediately visible on the live site
		isPublished := false
		publishedProduct, err := webflow.NewCollectionItem(ctx, "published-product", &webflow.CollectionItemArgs{
			CollectionId: pulumi.String(collectionID),
			FieldData: pulumi.Map{
				"name": pulumi.String("Premium Widget"),
				"slug": pulumi.String("premium-widget"),
				// Add your custom fields here based on your collection schema
				// Example custom fields (uncomment and modify based on your schema):
				// "price": pulumi.Float64(99.99),
				// "description": pulumi.String("The best widget on the market"),
				// "category": pulumi.String("Electronics"),
				// "in-stock": pulumi.Bool(true),
			},
			IsDraft:    &isPublished, // false = published
			IsArchived: &isNotArchived,
		})
		if err != nil {
			return fmt.Errorf("failed to create published product: %w", err)
		}

		// Example 3: Archived Content
		// Create an item that is archived (hidden but retained for records)
		isArchived := true
		archivedItem, err := webflow.NewCollectionItem(ctx, "archived-item", &webflow.CollectionItemArgs{
			CollectionId: pulumi.String(collectionID),
			FieldData: pulumi.Map{
				"name": pulumi.String("Discontinued Product"),
				"slug": pulumi.String("discontinued-product-archive"),
			},
			IsDraft:    &isDraft,
			IsArchived: &isArchived, // Hidden from both CMS and live site
		})
		if err != nil {
			return fmt.Errorf("failed to create archived item: %w", err)
		}

		// Example 4: Bulk Content Creation
		// Create multiple items efficiently using a loop
		contentData := []struct {
			name     string
			slug     string
			category string
		}{
			{"Introduction to TypeScript", "intro-typescript", "Tutorial"},
			{"Advanced Pulumi Patterns", "advanced-pulumi-patterns", "Tutorial"},
			{"Webflow API Best Practices", "webflow-api-best-practices", "Guide"},
		}

		bulkItemIDs := []pulumi.StringOutput{}
		bulkItemItemIDs := []pulumi.StringOutput{}
		for i, data := range contentData {
			item, err := webflow.NewCollectionItem(ctx, fmt.Sprintf("bulk-item-%d", i), &webflow.CollectionItemArgs{
				CollectionId: pulumi.String(collectionID),
				FieldData: pulumi.Map{
					"name": pulumi.String(data.name),
					"slug": pulumi.String(data.slug),
					// Add your custom fields here
					// "category": pulumi.String(data.category),
				},
				IsDraft: &isDraft, // Start as drafts
			})
			if err != nil {
				return fmt.Errorf("failed to create bulk item %d: %w", i, err)
			}
			bulkItemIDs = append(bulkItemIDs, item.ID().ToStringOutput())
			bulkItemItemIDs = append(bulkItemItemIDs, item.ItemId)
		}

		// Example 5: Localized Content (optional - only if your site uses localization)
		// Uncomment if your Webflow site has localization enabled
		// localizedItem, err := webflow.NewCollectionItem(ctx, "localized-item", &webflow.CollectionItemArgs{
		// 	CollectionId: pulumi.String(collectionID),
		// 	FieldData: pulumi.Map{
		// 		"name": pulumi.String("Bienvenue"),
		// 		"slug": pulumi.String("bienvenue"),
		// 	},
		// 	CmsLocaleId: pulumi.String("fr-FR"), // French locale
		// 	IsDraft:     &isPublished,
		// })
		// if err != nil {
		// 	return fmt.Errorf("failed to create localized item: %w", err)
		// }

		// Export the resource IDs for reference
		ctx.Export("deployedCollectionId", pulumi.String(collectionID))
		ctx.Export("deployedEnvironment", pulumi.String(environment))

		// Draft blog post exports
		ctx.Export("draftPostId", draftBlogPost.ID())
		ctx.Export("draftPostItemId", draftBlogPost.ItemId)
		ctx.Export("draftPostCreatedOn", draftBlogPost.CreatedOn)

		// Published product exports
		ctx.Export("publishedProductId", publishedProduct.ID())
		ctx.Export("publishedProductItemId", publishedProduct.ItemId)
		ctx.Export("publishedProductLastUpdated", publishedProduct.LastUpdated)

		// Archived item exports
		ctx.Export("archivedItemId", archivedItem.ID())
		ctx.Export("archivedItemItemId", archivedItem.ItemId)

		// Bulk items exports
		ctx.Export("bulkItemIds", pulumi.StringArray(bulkItemIDs))
		ctx.Export("bulkItemItemIds", pulumi.StringArray(bulkItemItemIDs))

		// Print deployment success message
		totalItems := 3 + len(contentData)
		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully deployed %d collection items to collection %s", totalItems, collectionID),
			&pulumi.LogOptions{},
		)
		ctx.Log.Info(
			fmt.Sprintf("   Environment: %s", environment),
			&pulumi.LogOptions{},
		)
		ctx.Log.Info(
			fmt.Sprintf("   Draft items: %d", 1+len(contentData)),
			&pulumi.LogOptions{},
		)
		ctx.Log.Info("   Published items: 1", &pulumi.LogOptions{})
		ctx.Log.Info("   Archived items: 1", &pulumi.LogOptions{})

		return nil
	})
}
