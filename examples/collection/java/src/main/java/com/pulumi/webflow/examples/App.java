package com.pulumi.webflow.examples;

import com.pulumi.Pulumi;
import com.pulumi.core.Output;
import com.pulumi.webflow.Collection;
import com.pulumi.webflow.CollectionArgs;

public class App {
    public static void main(String[] args) {
        Pulumi.run(ctx -> {
            // Get configuration values
            var config = ctx.config();
            var siteId = config.requireSecret("siteId");
            var environment = config.get("environment").orElse("development");

            // Example 1: Blog Posts Collection
            // A common pattern for blog content with all required fields
            var blogCollection = new Collection("blog-posts-collection",
                CollectionArgs.builder()
                    .siteId(siteId)
                    .displayName("Blog Posts")
                    .singularName("Blog Post")
                    .slug("blog-posts")
                    .build());

            // Example 2: Products Collection with Auto-Generated Slug
            // Omit the slug to let Webflow auto-generate it from the displayName
            var productsCollection = new Collection("products-collection",
                CollectionArgs.builder()
                    .siteId(siteId)
                    .displayName("Products")
                    .singularName("Product")
                    // slug is optional - Webflow will generate automatically
                    .build());

            // Example 3: Team Members Collection
            // Demonstrates custom slug different from display name
            var teamCollection = new Collection("team-members-collection",
                CollectionArgs.builder()
                    .siteId(siteId)
                    .displayName("Team Members")
                    .singularName("Team Member")
                    .slug("team")
                    .build());

            // Example 4: Portfolio Items Collection
            // Another common use case for showcasing work
            var portfolioCollection = new Collection("portfolio-collection",
                CollectionArgs.builder()
                    .siteId(siteId)
                    .displayName("Portfolio Items")
                    .singularName("Portfolio Item")
                    .slug("portfolio")
                    .build());

            // Example 5: Dynamic Collections Based on Config
            // Create collections based on configuration for multi-environment setups
            var testCollection = new Collection("test-collection-" + environment,
                CollectionArgs.builder()
                    .siteId(siteId)
                    .displayName("Test Collection (" + environment + ")")
                    .singularName("Test Item")
                    .slug("test-" + environment)
                    .build());

            // Export collection details for reference
            ctx.export("deployedSiteId", siteId);
            ctx.export("blogCollectionId", blogCollection.id());
            ctx.export("blogCollectionName", blogCollection.displayName());
            ctx.export("blogCollectionSlug", blogCollection.slug());
            ctx.export("blogCollectionCreatedOn", blogCollection.createdOn());

            ctx.export("productsCollectionId", productsCollection.id());
            ctx.export("teamCollectionId", teamCollection.id());
            ctx.export("portfolioCollectionId", portfolioCollection.id());
            ctx.export("testCollectionId", testCollection.id());

            // Export a summary of all collections
            ctx.export("allCollections", Output.tuple(
                blogCollection.displayName(),
                productsCollection.displayName(),
                teamCollection.displayName(),
                portfolioCollection.displayName(),
                testCollection.displayName()
            ).applyValue(t -> String.join(", ", t.t1, t.t2, t.t3, t.t4, t.t5)));
        });
    }
}
