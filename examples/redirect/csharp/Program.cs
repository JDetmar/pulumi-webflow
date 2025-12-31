using System;
using System.Collections.Generic;
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

        // Example 1: Permanent Redirect (301) - Best for content moves
        // Use 301 redirects when content has permanently moved to preserve SEO value
        var permanentRedirect = new Redirect("old-blog-to-new-blog", new RedirectArgs
        {
            SiteId = siteId,
            SourcePath = "/blog/old-article",
            DestinationPath = "/blog/articles/updated-article",
            StatusCode = 301,
        });

        // Example 2: Temporary Redirect (302) - Use for temporary changes
        // Use 302 redirects for seasonal content or A/B testing
        var temporaryRedirect = new Redirect("temporary-landing-page", new RedirectArgs
        {
            SiteId = siteId,
            SourcePath = "/old-campaign",
            DestinationPath = "/new-campaign-2025",
            StatusCode = 302,
        });

        // Example 3: External Redirect (301) - Redirect to another domain
        // Useful for partner links or moved subdomains
        var externalRedirect = new Redirect("external-partner-link", new RedirectArgs
        {
            SiteId = siteId,
            SourcePath = "/partner",
            DestinationPath = "https://partner-site.com",
            StatusCode = 301,
        });

        // Example 4: Bulk Redirects using Loop
        // Efficient pattern for migrating multiple URLs at once
        var redirectMappings = new[]
        {
            new { Old = "/product-a", New = "/products/product-a" },
            new { Old = "/product-b", New = "/products/product-b" },
            new { Old = "/product-c", New = "/products/product-c" },
        };

        var bulkRedirects = new List<Redirect>();
        for (int i = 0; i < redirectMappings.Length; i++)
        {
            var mapping = redirectMappings[i];
            var redirect = new Redirect($"bulk-redirect-{i}", new RedirectArgs
            {
                SiteId = siteId,
                SourcePath = mapping.Old,
                DestinationPath = mapping.New,
                StatusCode = 301,
            });
            bulkRedirects.Add(redirect);
        }

        // Export the redirect resources for reference
        return new Dictionary<string, object?>
        {
            ["deployedSiteId"] = siteId,
            ["permanentRedirectId"] = permanentRedirect.Id,
            ["temporaryRedirectId"] = temporaryRedirect.Id,
            ["externalRedirectId"] = externalRedirect.Id,
            ["bulkRedirectIds"] = Output.All(bulkRedirects.Select(r => r.Id).ToArray()),
        };
    });
}
