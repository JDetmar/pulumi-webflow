package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		workspaceID := cfg.Require("workspaceId")
		displayName := cfg.Require("displayName")
		shortName := cfg.Require("shortName")
		timezone := cfg.Get("timezone")
		if timezone == "" {
			timezone = "America/New_York"
		}

		// Example 1: Basic Site Creation
		basicSite, err := webflow.NewSite(ctx, "basic-site", &webflow.SiteArgs{
			WorkspaceId: pulumi.String(workspaceID),
			DisplayName: pulumi.String(displayName),
			ShortName:   pulumi.String(shortName),
			TimeZone:    pulumi.String(timezone),
		})
		if err != nil {
			return fmt.Errorf("failed to create basic site: %w", err)
		}

		// Example 2: Multi-Environment Site Configuration
		environments := []string{"development", "staging", "production"}
		environmentSites := make([]*webflow.Site, 0)
		for _, env := range environments {
			site, err := webflow.NewSite(ctx, fmt.Sprintf("site-%s", env), &webflow.SiteArgs{
				WorkspaceId: pulumi.String(workspaceID),
				DisplayName: pulumi.String(fmt.Sprintf("%s-%s", displayName, env)),
				ShortName:   pulumi.String(fmt.Sprintf("%s-%s", shortName, env)),
				TimeZone:    pulumi.String(timezone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site %s: %w", env, err)
			}
			environmentSites = append(environmentSites, site)
		}

		// Example 3: Site with Configuration
		configuredSite, err := webflow.NewSite(ctx, "configured-site", &webflow.SiteArgs{
			WorkspaceId: pulumi.String(workspaceID),
			DisplayName: pulumi.String(fmt.Sprintf("%s-configured", displayName)),
			ShortName:   pulumi.String(fmt.Sprintf("%s-configured", shortName)),
			TimeZone:    pulumi.String(timezone),
		})
		if err != nil {
			return fmt.Errorf("failed to create configured site: %w", err)
		}

		// Export values
		ctx.Export("basicSiteId", basicSite.ID())
		ctx.Export("basicSiteName", basicSite.DisplayName)

		siteIds := make(pulumi.StringArray, len(environmentSites))
		for i, site := range environmentSites {
			siteIds[i] = site.ID().ToStringOutput()
		}
		ctx.Export("environmentSiteIds", siteIds)
		ctx.Export("configuredSiteId", configuredSite.ID())

		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully created %d sites", len(environmentSites)+2),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
