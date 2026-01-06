package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.CollectionItem;
import com.pulumi.webflow.CollectionItemArgs;

import java.util.List;
import java.util.ArrayList;
import java.util.Map;
import java.util.stream.Collectors;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var collectionId = config.require("collectionId");
            var environment = config.get("environment").orElse("development");

            // Example 1: Draft Blog Post
            // Create a blog post that exists in the CMS but is NOT published to the live site
            var draftBlogPost = new CollectionItem("draft-blog-post",
                CollectionItemArgs.builder()
                    .collectionId(collectionId)
                    .fieldData(Map.of(
                        "name", "Getting Started with Webflow CMS",
                        "slug", "getting-started-webflow-cms"
                        // Add your custom fields here based on your collection schema
                        // Example custom fields (uncomment and modify based on your schema):
                        // "post-body", "Learn how to use Webflow CMS to manage your content...",
                        // "author", "John Doe",
                        // "publish-date", "2025-01-06",
                        // "featured-image", "https://example.com/image.jpg"
                    ))
                    .isDraft(true) // Not published to live site
                    .isArchived(false)
                    .build());

            // Example 2: Published Product
            // Create a product that is immediately visible on the live site
            var publishedProduct = new CollectionItem("published-product",
                CollectionItemArgs.builder()
                    .collectionId(collectionId)
                    .fieldData(Map.of(
                        "name", "Premium Widget",
                        "slug", "premium-widget"
                        // Add your custom fields here based on your collection schema
                        // Example custom fields (uncomment and modify based on your schema):
                        // "price", 99.99,
                        // "description", "The best widget on the market",
                        // "category", "Electronics",
                        // "in-stock", true
                    ))
                    .isDraft(false) // Published to live site
                    .isArchived(false)
                    .build());

            // Example 3: Archived Content
            // Create an item that is archived (hidden but retained for records)
            var archivedItem = new CollectionItem("archived-item",
                CollectionItemArgs.builder()
                    .collectionId(collectionId)
                    .fieldData(Map.of(
                        "name", "Discontinued Product",
                        "slug", "discontinued-product-archive"
                    ))
                    .isDraft(true)
                    .isArchived(true) // Hidden from both CMS and live site
                    .build());

            // Example 4: Bulk Content Creation
            // Create multiple items efficiently using a loop
            var contentData = List.of(
                Map.of(
                    "name", "Introduction to TypeScript",
                    "slug", "intro-typescript",
                    "category", "Tutorial"
                ),
                Map.of(
                    "name", "Advanced Pulumi Patterns",
                    "slug", "advanced-pulumi-patterns",
                    "category", "Tutorial"
                ),
                Map.of(
                    "name", "Webflow API Best Practices",
                    "slug", "webflow-api-best-practices",
                    "category", "Guide"
                )
            );

            var bulkItems = new ArrayList<CollectionItem>();
            for (int i = 0; i < contentData.size(); i++) {
                var data = contentData.get(i);
                var item = new CollectionItem("bulk-item-" + i,
                    CollectionItemArgs.builder()
                        .collectionId(collectionId)
                        .fieldData(Map.of(
                            "name", data.get("name"),
                            "slug", data.get("slug")
                            // Add your custom fields here
                            // "category", data.get("category")
                        ))
                        .isDraft(true) // Start as drafts
                        .build());
                bulkItems.add(item);
            }

            // Example 5: Localized Content (optional - only if your site uses localization)
            // Uncomment if your Webflow site has localization enabled
            // var localizedItem = new CollectionItem("localized-item",
            //     CollectionItemArgs.builder()
            //         .collectionId(collectionId)
            //         .fieldData(Map.of(
            //             "name", "Bienvenue",
            //             "slug", "bienvenue"
            //         ))
            //         .cmsLocaleId("fr-FR") // French locale
            //         .isDraft(false)
            //         .build());

            // Print deployment success message
            var totalItems = 3 + bulkItems.size();
            ctx.log().info("✅ Successfully deployed " + totalItems + " collection items to collection " + collectionId);
            ctx.log().info("   Environment: " + environment);
            ctx.log().info("   Draft items: " + (1 + bulkItems.size()));
            ctx.log().info("   Published items: 1");
            ctx.log().info("   Archived items: 1");

            // Export values for reference
            ctx.export("deployedCollectionId", Output.of(collectionId));
            ctx.export("deployedEnvironment", Output.of(environment));

            // Draft blog post exports
            ctx.export("draftPostId", draftBlogPost.id());
            ctx.export("draftPostItemId", draftBlogPost.itemId());
            ctx.export("draftPostCreatedOn", draftBlogPost.createdOn());

            // Published product exports
            ctx.export("publishedProductId", publishedProduct.id());
            ctx.export("publishedProductItemId", publishedProduct.itemId());
            ctx.export("publishedProductLastUpdated", publishedProduct.lastUpdated());

            // Archived item exports
            ctx.export("archivedItemId", archivedItem.id());
            ctx.export("archivedItemItemId", archivedItem.itemId());

            // Bulk items exports
            ctx.export("bulkItemIds", Output.all(bulkItems.stream()
                .map(CollectionItem::id)
                .collect(Collectors.toList())));
            ctx.export("bulkItemItemIds", Output.all(bulkItems.stream()
                .map(CollectionItem::itemId)
                .collect(Collectors.toList())));
        });
    }
}
