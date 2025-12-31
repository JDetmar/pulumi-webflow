package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Get configuration values
		cfg := pulumi.NewConfig(ctx, "")
		siteID := cfg.RequireSecret("siteId")
		environment := cfg.Get("environment")
		if environment == "" {
			environment = "development"
		}

		// Example 1: Allow All Crawlers (most common for public sites)
		allowAllRobots, err := webflow.NewRobotsTxt(ctx, "allow-all-robots", &webflow.RobotsTxtArgs{
			SiteId: siteID,
			Content: pulumi.String(`User-agent: *
Allow: /

# Allow specific crawler access with no delays
User-agent: Googlebot
Allow: /
Crawl-delay: 0

User-agent: Bingbot
Allow: /
Crawl-delay: 1`),
		})
		if err != nil {
			return fmt.Errorf("failed to create allow-all-robots: %w", err)
		}

		// Example 2: Selective Blocking (for staging/development)
		selectiveBlockRobots, err := webflow.NewRobotsTxt(ctx, "selective-block-robots", &webflow.RobotsTxtArgs{
			SiteId: siteID,
			Content: pulumi.String(`User-agent: *
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
Disallow: /`),
		})
		if err != nil {
			return fmt.Errorf("failed to create selective-block-robots: %w", err)
		}

		// Example 3: Restrict Directories (protect API and backend)
		restrictDirectoriesRobots, err := webflow.NewRobotsTxt(ctx, "restrict-directories-robots", &webflow.RobotsTxtArgs{
			SiteId: siteID,
			Content: pulumi.String(`User-agent: *
Allow: /
Disallow: /api/
Disallow: /internal/
Disallow: /*.json$
Disallow: /*.xml$

# Specify sitemap location
Sitemap: https://example.com/sitemap.xml`),
		})
		if err != nil {
			return fmt.Errorf("failed to create restrict-directories-robots: %w", err)
		}

		// Export values for reference
		ctx.Export("deployedSiteId", siteID)
		ctx.Export("allowAllRobotsId", allowAllRobots.ID())
		ctx.Export("allowAllRobotsLastModified", allowAllRobots.LastModified)
		ctx.Export("selectiveBlockRobotsId", selectiveBlockRobots.ID())
		ctx.Export("restrictDirectoriesRobotsId", restrictDirectoriesRobots.ID())

		// Print success message
		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully deployed RobotsTxt resources in %s environment", environment),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
