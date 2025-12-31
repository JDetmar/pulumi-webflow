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
        var environment = config.Get("environment") ?? "development";

        // Example 1: Allow All Crawlers (most common for public sites)
        var allowAllRobots = new RobotsTxt("allow-all-robots", new RobotsTxtArgs
        {
            SiteId = siteId,
            Content = @"User-agent: *
Allow: /

# Allow specific crawler access with no delays
User-agent: Googlebot
Allow: /
Crawl-delay: 0

User-agent: Bingbot
Allow: /
Crawl-delay: 1",
        });

        // Example 2: Selective Blocking (for staging/development)
        var selectiveBlockRobots = new RobotsTxt("selective-block-robots", new RobotsTxtArgs
        {
            SiteId = siteId,
            Content = @"User-agent: *
Allow: /

# Disallow admin and private sections
Disallow: /admin/
Disallow: /private/
Disallow: /staging/
Disallow: /test/

# Block specific crawlers
User-agent: AhrefsBot
Disallow: /

User-agent: SemrushBot
Disallow: /",
        });

        // Example 3: Restrict Directories (protect API and backend)
        var restrictDirectoriesRobots = new RobotsTxt("restrict-directories-robots", new RobotsTxtArgs
        {
            SiteId = siteId,
            Content = @"User-agent: *
Allow: /
Disallow: /api/
Disallow: /internal/
Disallow: /*.json$
Disallow: /*.xml$

# Specify sitemap location
Sitemap: https://example.com/sitemap.xml",
        });

        // Export the robot resources for reference
        return new Dictionary<string, object?>
        {
            ["deployedSiteId"] = siteId,
            ["allowAllRobotsId"] = allowAllRobots.Id,
            ["allowAllRobotsLastModified"] = allowAllRobots.LastModified,
            ["selectiveBlockRobotsId"] = selectiveBlockRobots.Id,
            ["restrictDirectoriesRobotsId"] = restrictDirectoriesRobots.Id,
        };
    });
}
