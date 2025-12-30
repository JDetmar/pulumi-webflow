// Copyright 2025, Pulumi Corporation.
//
// Licensed under the Apache License, Version 2.0 (the "License");
// you may not use this file except in compliance with the License.
// You may obtain a copy of the License at
//
//     http://www.apache.org/licenses/LICENSE-2.0
//
// Unless required by applicable law or agreed to in writing, software
// distributed under the License is distributed on an "AS IS" BASIS,
// WITHOUT WARRANTIES OR CONDITIONS OF ANY KIND, either express or implied.
// See the License for the specific language governing permissions and
// limitations under the License.

package main

import (
	"fmt"
	"slices"
	"strings"

	"github.com/pulumi/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

// SiteConfig represents configuration for a single site
type SiteConfig struct {
	DisplayName   string           `json:"displayName"`
	ShortName     string           `json:"shortName"`
	TimeZone      string           `json:"timeZone"`
	AllowIndexing bool             `json:"allowIndexing"`
	Redirects     []RedirectConfig `json:"redirects,omitempty"`
}

// RedirectConfig represents a redirect rule for a site
type RedirectConfig struct {
	SourcePath      string `json:"sourcePath"`
	DestinationPath string `json:"destinationPath"`
	StatusCode      int    `json:"statusCode"`
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// Load stack-specific configuration
		environmentName := cfg.Require("environmentName") // "dev", "staging", "prod"

		// Validate environment configuration
		validEnvironments := []string{"dev", "staging", "prod"}
		if !slices.Contains(validEnvironments, environmentName) {
			return fmt.Errorf(
				"invalid environment '%s'. must be one of: %s",
				environmentName, strings.Join(validEnvironments, ", "))
		}

		// Production safety check
		isProd := environmentName == "prod"
		if isProd {
			confirmation := cfg.Get("prodDeploymentConfirmed")
			if confirmation != "yes" {
				return fmt.Errorf(
					"⚠️  production deployment requires explicit confirmation\n" +
						"run: pulumi config set prodDeploymentConfirmed yes\n" +
						"this prevents accidental production deployments")
			}
		}

		// Load site configurations from stack config
		// Each site is a distinct entity with its own purpose
		var sitesConfig map[string]SiteConfig
		cfg.RequireObject("sites", &sitesConfig)

		ctx.Log.Info(
			fmt.Sprintf("🚀 deploying %d sites to %s environment",
				len(sitesConfig), environmentName), nil)

		// Create each configured site
		siteExports := make(map[string]pulumi.StringOutput)
		siteNames := make([]string, 0, len(sitesConfig))

		for siteKey, siteConfig := range sitesConfig {
			siteNames = append(siteNames, siteKey)

			// Create site with its specific configuration
			site, err := webflow.NewSite(ctx, siteKey, &webflow.SiteArgs{
				DisplayName: pulumi.String(siteConfig.DisplayName),
				ShortName:   pulumi.String(siteConfig.ShortName),
				TimeZone:    pulumi.String(siteConfig.TimeZone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site %s: %w", siteKey, err)
			}

			siteExports[fmt.Sprintf("%s-id", siteKey)] = site.ID

			// Configure robots.txt based on site's indexing preference
			robotsContent := "User-agent: *\nAllow: /\n"
			if !siteConfig.AllowIndexing {
				robotsContent = fmt.Sprintf(
					"User-agent: *\nDisallow: /\n\n# %s - Do not index",
					strings.ToUpper(environmentName))
			}

			_, err = webflow.NewRobotsTxt(ctx, fmt.Sprintf("%s-robots", siteKey),
				&webflow.RobotsTxtArgs{
					SiteID:  site.ID,
					Content: pulumi.String(robotsContent),
				})
			if err != nil {
				return fmt.Errorf("failed to create robots.txt for site %s: %w", siteKey, err)
			}

			// Create any configured redirects for this site
			for _, redirect := range siteConfig.Redirects {
				redirectName := fmt.Sprintf("%s-redirect%s",
					siteKey, strings.ReplaceAll(redirect.SourcePath, "/", "-"))

				_, err = webflow.NewRedirect(ctx, redirectName,
					&webflow.RedirectArgs{
						SiteID:          site.ID,
						SourcePath:      pulumi.String(redirect.SourcePath),
						DestinationPath: pulumi.String(redirect.DestinationPath),
						StatusCode:      pulumi.Int(redirect.StatusCode),
					})
				if err != nil {
					return fmt.Errorf("failed to create redirect %s for site %s: %w",
						redirect.SourcePath, siteKey, err)
				}
			}

			ctx.Log.Info(fmt.Sprintf("✅ configured site: %s (%s)",
				siteKey, siteConfig.DisplayName), nil)
		}

		// Export site IDs for reference
		for key, value := range siteExports {
			ctx.Export(key, value)
		}

		// Export summary information
		ctx.Export("environment", pulumi.String(environmentName))
		ctx.Export("siteCount", pulumi.Int(len(sitesConfig)))
		ctx.Export("siteNames", pulumi.ToStringArray(siteNames))
		ctx.Export("stackName", pulumi.String(ctx.Stack()))
		ctx.Export("projectName", pulumi.String(ctx.Project()))

		ctx.Log.Info(
			fmt.Sprintf("✅ deployment complete: %d sites configured for %s",
				len(sitesConfig), environmentName), nil)

		return nil
	})
}
