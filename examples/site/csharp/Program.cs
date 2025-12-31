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
        var displayName = config.Require("displayName");
        var shortName = config.Require("shortName");
        var customDomain = config.Get("customDomain");
        var timezone = config.Get("timezone") ?? "America/New_York";

        // Example 1: Basic Site Creation
        // Create a simple site with required properties
        var basicSite = new Site("basic-site", new SiteArgs
        {
            DisplayName = displayName,
            ShortName = shortName,
            Timezone = timezone,
        });

        // Example 2: Site with Custom Domain (conditional creation)
        // Only create if customDomain is provided in configuration
        Site? siteWithDomain = null;
        if (!string.IsNullOrEmpty(customDomain))
        {
            siteWithDomain = new Site("site-with-domain", new SiteArgs
            {
                DisplayName = $"{displayName}-domain",
                ShortName = $"{shortName}-domain",
                CustomDomain = customDomain,
                Timezone = timezone,
            });
        }

        // Example 3: Multi-Environment Site Configuration
        // Create sites for different environments using a loop
        var environments = new[] { "development", "staging", "production" };
        var environmentSites = new List<Site>();

        foreach (var env in environments)
        {
            var site = new Site($"site-{env}", new SiteArgs
            {
                DisplayName = $"{displayName}-{env}",
                ShortName = $"{shortName}-{env}",
                Timezone = timezone,
            });
            environmentSites.Add(site);
        }

        // Example 4: Site with Full Configuration
        // Demonstrates all available configuration options
        var configuredSite = new Site("configured-site", new SiteArgs
        {
            DisplayName = $"{displayName}-configured",
            ShortName = $"{shortName}-configured",
            Timezone = timezone,
        });

        // Export the site resources for reference
        return new Dictionary<string, object?>
        {
            ["basicSiteId"] = basicSite.Id,
            ["basicSiteName"] = basicSite.DisplayName,
            ["customDomainSiteId"] = siteWithDomain?.Id ?? Output.Create("not-created"),
            ["environmentSiteCount"] = environmentSites.Count,
            ["configuredSiteId"] = configuredSite.Id,
        };
    });
}
