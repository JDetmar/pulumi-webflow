using System;
using System.Collections.Generic;
using System.Collections.Immutable;
using System.Linq;
using System.Threading.Tasks;
using Pulumi;
using Pulumi.Webflow;

class Program
{
    static Task<int> Main() => Deployment.RunAsync(() =>
    {
        // Create a Pulumi config object
        var config = new Config();

        // Get configuration values
        var collectionId = config.Require("collectionId");
        var environment = config.Get("environment") ?? "development";

        // Example 1: Draft Blog Post
        // Create a blog post that exists in the CMS but is NOT published to the live site
        var draftBlogPost = new CollectionItem("draft-blog-post", new CollectionItemArgs
        {
            CollectionId = collectionId,
            FieldData = new Dictionary<string, object>
            {
                ["name"] = "Getting Started with Webflow CMS",
                ["slug"] = "getting-started-webflow-cms",
                // Add your custom fields here based on your collection schema
                // Example custom fields (uncomment and modify based on your schema):
                // ["post-body"] = "Learn how to use Webflow CMS to manage your content...",
                // ["author"] = "John Doe",
                // ["publish-date"] = "2025-01-06",
                // ["featured-image"] = "https://example.com/image.jpg",
            }.ToImmutableDictionary(),
            IsDraft = true, // Not published to live site
            IsArchived = false,
        });

        // Example 2: Published Product
        // Create a product that is immediately visible on the live site
        var publishedProduct = new CollectionItem("published-product", new CollectionItemArgs
        {
            CollectionId = collectionId,
            FieldData = new Dictionary<string, object>
            {
                ["name"] = "Premium Widget",
                ["slug"] = "premium-widget",
                // Add your custom fields here based on your collection schema
                // Example custom fields (uncomment and modify based on your schema):
                // ["price"] = 99.99,
                // ["description"] = "The best widget on the market",
                // ["category"] = "Electronics",
                // ["in-stock"] = true,
            }.ToImmutableDictionary(),
            IsDraft = false, // Published to live site
            IsArchived = false,
        });

        // Example 3: Archived Content
        // Create an item that is archived (hidden but retained for records)
        var archivedItem = new CollectionItem("archived-item", new CollectionItemArgs
        {
            CollectionId = collectionId,
            FieldData = new Dictionary<string, object>
            {
                ["name"] = "Discontinued Product",
                ["slug"] = "discontinued-product-archive",
            }.ToImmutableDictionary(),
            IsDraft = true,
            IsArchived = true, // Hidden from both CMS and live site
        });

        // Example 4: Bulk Content Creation
        // Create multiple items efficiently using a loop
        var contentData = new[]
        {
            new { Name = "Introduction to TypeScript", Slug = "intro-typescript", Category = "Tutorial" },
            new { Name = "Advanced Pulumi Patterns", Slug = "advanced-pulumi-patterns", Category = "Tutorial" },
            new { Name = "Webflow API Best Practices", Slug = "webflow-api-best-practices", Category = "Guide" },
        };

        var bulkItems = new List<CollectionItem>();
        for (int i = 0; i < contentData.Length; i++)
        {
            var data = contentData[i];
            var item = new CollectionItem($"bulk-item-{i}", new CollectionItemArgs
            {
                CollectionId = collectionId,
                FieldData = new Dictionary<string, object>
                {
                    ["name"] = data.Name,
                    ["slug"] = data.Slug,
                    // Add your custom fields here
                    // ["category"] = data.Category,
                }.ToImmutableDictionary(),
                IsDraft = true, // Start as drafts
            });
            bulkItems.Add(item);
        }

        // Example 5: Localized Content (optional - only if your site uses localization)
        // Uncomment if your Webflow site has localization enabled
        // var localizedItem = new CollectionItem("localized-item", new CollectionItemArgs
        // {
        //     CollectionId = collectionId,
        //     FieldData = new Dictionary<string, object>
        //     {
        //         ["name"] = "Bienvenue",
        //         ["slug"] = "bienvenue",
        //     }.ToImmutableDictionary(),
        //     CmsLocaleId = "fr-FR", // French locale
        //     IsDraft = false,
        // });

        // Print deployment success message
        var totalItems = 3 + bulkItems.Count;
        Log.Info($"✅ Successfully deployed {totalItems} collection items to collection {collectionId}");
        Log.Info($"   Environment: {environment}");
        Log.Info($"   Draft items: {1 + bulkItems.Count}");
        Log.Info("   Published items: 1");
        Log.Info("   Archived items: 1");

        // Export the resource IDs for reference
        return new Dictionary<string, object?>
        {
            ["deployedCollectionId"] = collectionId,
            ["deployedEnvironment"] = environment,

            // Draft blog post exports
            ["draftPostId"] = draftBlogPost.Id,
            ["draftPostItemId"] = draftBlogPost.ItemId,
            ["draftPostCreatedOn"] = draftBlogPost.CreatedOn,

            // Published product exports
            ["publishedProductId"] = publishedProduct.Id,
            ["publishedProductItemId"] = publishedProduct.ItemId,
            ["publishedProductLastUpdated"] = publishedProduct.LastUpdated,

            // Archived item exports
            ["archivedItemId"] = archivedItem.Id,
            ["archivedItemItemId"] = archivedItem.ItemId,

            // Bulk items exports
            ["bulkItemIds"] = Output.All(bulkItems.Select(item => item.Id).ToArray()),
            ["bulkItemItemIds"] = Output.All(bulkItems.Select(item => item.ItemId).ToArray()),
        };
    });
}
