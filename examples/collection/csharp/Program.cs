using System;
using System.Collections.Generic;
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
        var siteId = config.RequireSecret("siteId");
        var environment = config.Get("environment") ?? "development";

        // Example 1: Blog Posts Collection
        // A common pattern for blog content with all required fields
        var blogCollection = new Collection("blog-posts-collection", new CollectionArgs
        {
            SiteId = siteId,
            DisplayName = "Blog Posts",
            SingularName = "Blog Post",
            Slug = "blog-posts",
        });

        // Example 2: Products Collection with Auto-Generated Slug
        // Omit the slug to let Webflow auto-generate it from the displayName
        var productsCollection = new Collection("products-collection", new CollectionArgs
        {
            SiteId = siteId,
            DisplayName = "Products",
            SingularName = "Product",
            // slug is optional - Webflow will generate automatically
        });

        // Example 3: Team Members Collection
        // Demonstrates custom slug different from display name
        var teamCollection = new Collection("team-members-collection", new CollectionArgs
        {
            SiteId = siteId,
            DisplayName = "Team Members",
            SingularName = "Team Member",
            Slug = "team",
        });

        // Example 4: Portfolio Items Collection
        // Another common use case for showcasing work
        var portfolioCollection = new Collection("portfolio-collection", new CollectionArgs
        {
            SiteId = siteId,
            DisplayName = "Portfolio Items",
            SingularName = "Portfolio Item",
            Slug = "portfolio",
        });

        // Example 5: Dynamic Collections Based on Config
        // Create collections based on configuration for multi-environment setups
        var testCollection = new Collection($"test-collection-{environment}", new CollectionArgs
        {
            SiteId = siteId,
            DisplayName = $"Test Collection ({environment})",
            SingularName = "Test Item",
            Slug = $"test-{environment}",
        });

        // Export collection details for reference
        return new Dictionary<string, object?>
        {
            ["deployedSiteId"] = siteId,
            ["blogCollectionId"] = blogCollection.Id,
            ["blogCollectionName"] = blogCollection.DisplayName,
            ["blogCollectionSlug"] = blogCollection.Slug,
            ["blogCollectionCreatedOn"] = blogCollection.CreatedOn,
            ["productsCollectionId"] = productsCollection.Id,
            ["teamCollectionId"] = teamCollection.Id,
            ["portfolioCollectionId"] = portfolioCollection.Id,
            ["testCollectionId"] = testCollection.Id,
            ["allCollections"] = Output.Tuple(
                blogCollection.DisplayName,
                productsCollection.DisplayName,
                teamCollection.DisplayName,
                portfolioCollection.DisplayName,
                testCollection.DisplayName
            ).Apply(t => string.Join(", ", new[] { t.Item1, t.Item2, t.Item3, t.Item4, t.Item5 })),
        };
    });
}
