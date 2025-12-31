package main

import (
	"fmt"

	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := pulumi.NewConfig(ctx, "")
		displayName := cfg.Require("displayName")
		shortName := cfg.Require("shortName")
		customDomain := cfg.Get("customDomain")
		timezone := cfg.Get("timezone")
		if timezone == "" {
			timezone = "America/New_York"
		}

		// Example 1: Basic Site Creation
		basicSite, err := webflow.NewSite(ctx, "basic-site", &webflow.SiteArgs{
			DisplayName: pulumi.String(displayName),
			ShortName:   pulumi.String(shortName),
			Timezone:    pulumi.String(timezone),
		})
		if err != nil {
			return fmt.Errorf("failed to create basic site: %w", err)
		}

		// Example 2: Site with Custom Domain
		var siteWithDomain *webflow.Site
		if customDomain != "" {
			siteWithDomain, err = webflow.NewSite(ctx, "site-with-domain", &webflow.SiteArgs{
				DisplayName: pulumi.String(fmt.Sprintf("%s-domain", displayName)),
				ShortName:   pulumi.String(fmt.Sprintf("%s-domain", shortName)),
				CustomDomain: pulumi.String(customDomain),
				Timezone:    pulumi.String(timezone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site with domain: %w", err)
			}
		}

		// Example 3: Multi-Environment Site Configuration
		environments := []string{"development", "staging", "production"}
		environmentSites := make([]*webflow.Site, 0)
		for _, env := range environments {
			site, err := webflow.NewSite(ctx, fmt.Sprintf("site-%s", env), &webflow.SiteArgs{
				DisplayName: pulumi.String(fmt.Sprintf("%s-%s", displayName, env)),
				ShortName:   pulumi.String(fmt.Sprintf("%s-%s", shortName, env)),
				Timezone:    pulumi.String(timezone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site %s: %w", env, err)
			}
			environmentSites = append(environmentSites, site)
		}

		// Example 4: Site with Configuration
		configuredSite, err := webflow.NewSite(ctx, "configured-site", &webflow.SiteArgs{
			DisplayName: pulumi.String(fmt.Sprintf("%s-configured", displayName)),
			ShortName:   pulumi.String(fmt.Sprintf("%s-configured", shortName)),
			Timezone:    pulumi.String(timezone),
		})
		if err != nil {
			return fmt.Errorf("failed to create configured site: %w", err)
		}

		// Export values
		ctx.Export("basicSiteId", basicSite.ID())
		ctx.Export("basicSiteName", basicSite.DisplayName)

		if siteWithDomain != nil {
			ctx.Export("customDomainSiteId", siteWithDomain.ID())
		} else {
			ctx.Export("customDomainSiteId", pulumi.String("not-created"))
		}

		siteIds := make(pulumi.StringArray, len(environmentSites))
		for i, site := range environmentSites {
			siteIds[i] = site.ID().ToStringOutput()
		}
		ctx.Export("environmentSiteIds", siteIds)
		ctx.Export("configuredSiteId", configuredSite.ID())

		ctx.Log.Info(
			fmt.Sprintf("✅ Successfully created %d sites", len(environmentSites)+1),
			&pulumi.LogOptions{},
		)

		return nil
	})
}
