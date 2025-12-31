package main

import (
	"fmt"

	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// main creates a Webflow RobotsTxt resource using the Webflow Pulumi Provider
//
// This example demonstrates:
// - Loading configuration from Pulumi config
// - Creating a RobotsTxt resource
// - Exporting outputs for reference
func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Create a config object to access configuration values
		cfg := config.New(ctx, "")

		// Get configuration values (these should be set via `pulumi config set`)
		siteID := cfg.RequireSecret("siteId")

		// Deploy a RobotsTxt resource to your Webflow site
		//
		// This example creates a robots.txt file that:
		// - Allows all search engine crawlers (User-agent: *)
		// - Allows Google's bot (Googlebot) to crawl all pages
		//
		// You can customize the robots.txt content by modifying the content string below
		robotsTxt, err := webflow.NewRobotsTxt(ctx, "my-robots", &webflow.RobotsTxtArgs{
			SiteId: siteID,
			Content: pulumi.String(`User-agent: *
Allow: /

User-agent: Googlebot
Allow: /`),
		})
		if err != nil {
			return err
		}

		// Export values for reference
		ctx.Export("deployedSiteId", siteID)
		ctx.Export("robotsTxtId", robotsTxt.ID())

		// Print a success message
		fmt.Println("✅ Successfully deployed robots.txt to your Webflow site")

		return nil
	})
}
