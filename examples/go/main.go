package main

import (
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Example: Configure robots.txt for a site
		robotsTxt, err := webflow.NewRobotsTxt(ctx, "myRobotsTxt", &webflow.RobotsTxtArgs{
			SiteId:  pulumi.String("your-site-id-here"),
			Content: pulumi.String("User-agent: *\nAllow: /"),
		})
		if err != nil {
			return err
		}

		// Example: Create a redirect
		redirect, err := webflow.NewRedirect(ctx, "myRedirect", &webflow.RedirectArgs{
			SiteId:          pulumi.String("your-site-id-here"),
			SourcePath:      pulumi.String("/old-page"),
			DestinationPath: pulumi.String("/new-page"),
			StatusCode:      pulumi.Int(301),
		})
		if err != nil {
			return err
		}

		ctx.Export("robotsTxtId", robotsTxt.ID())
		ctx.Export("redirectId", redirect.ID())
		return nil
	})
}
