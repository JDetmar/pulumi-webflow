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
        var workspaceId = config.Require("workspaceId");
        var displayName = config.Require("displayName");
        var shortName = config.Require("shortName");

        // Example 1: Basic Site Creation
        // Create a simple site with required properties
        var basicSite = new Site("basic-site", new SiteArgs
        {
            WorkspaceId = workspaceId,
            DisplayName = displayName,
            ShortName = shortName,
        });

        // Example 2: Multi-Environment Site Configuration
        // Create sites for different environments using a loop
        var environments = new[] { "development", "staging", "production" };
        var environmentSites = new List<Site>();

        foreach (var env in environments)
        {
            var site = new Site($"site-{env}", new SiteArgs
            {
                WorkspaceId = workspaceId,
                DisplayName = $"{displayName}-{env}",
                ShortName = $"{shortName}-{env}",
            });
            environmentSites.Add(site);
        }

        // Example 3: Site with Full Configuration
        // Demonstrates all available configuration options
        var configuredSite = new Site("configured-site", new SiteArgs
        {
            WorkspaceId = workspaceId,
            DisplayName = $"{displayName}-configured",
            ShortName = $"{shortName}-configured",
        });

        // Export the site resources for reference
        return new Dictionary<string, object?>
        {
            ["basicSiteId"] = basicSite.Id,
            ["basicSiteName"] = basicSite.DisplayName,
            ["environmentSiteIds"] = Output.All(environmentSites.Select(s => s.Id).ToArray()),
            ["configuredSiteId"] = configuredSite.Id,
        };
    });
}
