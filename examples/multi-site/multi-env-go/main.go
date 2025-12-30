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
	"strings"

	"github.com/pulumi/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")

		// Load environment-specific configuration
		stackName := ctx.Stack()
		sitePrefix := cfg.Require("sitePrefix")  // e.g., "dev", "staging", "prod"
		siteCount := cfg.RequireInt("siteCount") // Number of sites to create

		ctx.Log.Info(fmt.Sprintf("Deploying %d %s sites to stack: %s", siteCount,
			sitePrefix, stackName), nil)

		// Determine environment-specific settings
		var (
			timeZones = []string{
				"America/Los_Angeles",
				"America/Chicago",
				"America/New_York",
			}
			robotsContent = fmt.Sprintf(`User-agent: *
Allow: /
X-Environment: %s`, sitePrefix)
		)

		// Create environment-specific site fleet
		siteIDs := make([]pulumi.StringOutput, 0)

		for i := 0; i < siteCount; i++ {
			// Create unique site name with environment prefix
			siteName := fmt.Sprintf("%s-site-%d", sitePrefix, i+1)
			displayName := fmt.Sprintf("%s Site %d",
				strings.ToUpper(sitePrefix), i+1)
			shortName := fmt.Sprintf("%s-site-%d",
				strings.ToLower(sitePrefix), i+1)

			// Distribute sites across time zones for realistic variety
			timeZone := timeZones[i%len(timeZones)]

			// Create the site
			site, err := webflow.NewSite(ctx, siteName, &webflow.SiteArgs{
				DisplayName: pulumi.String(displayName),
				ShortName:   pulumi.String(shortName),
				TimeZone:    pulumi.String(timeZone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site %s: %w", siteName, err)
			}

			siteIDs = append(siteIDs, site.ID)

			// Add robots.txt with environment marker
			_, err = webflow.NewRobotsTxt(ctx, fmt.Sprintf("%s-robots", siteName),
				&webflow.RobotsTxtArgs{
					SiteID:  site.ID,
					Content: pulumi.String(robotsContent),
				})
			if err != nil {
				return fmt.Errorf("failed to create robots.txt for %s: %w",
					siteName, err)
			}

			// Add environment-specific redirect
			if sitePrefix == "prod" {
				// Production: redirect old domains
				_, err = webflow.NewRedirect(ctx, fmt.Sprintf("%s-domain-redirect", siteName),
					&webflow.RedirectArgs{
						SiteID:            site.ID,
						SourcePath:        pulumi.String("/old-domain"),
						DestinationPath:   pulumi.String("/"),
						StatusCode:        pulumi.Int(301),
					})
				if err != nil {
					return fmt.Errorf("failed to create redirect for %s: %w",
						siteName, err)
				}
			}

			// Export individual site ID
			ctx.Export(fmt.Sprintf("%s-id", siteName), site.ID)
		}

		// Export summary information
		ctx.Export(fmt.Sprintf("%s-total-sites", sitePrefix), pulumi.Int(siteCount))
		ctx.Export(fmt.Sprintf("%s-site-ids", sitePrefix),
			pulumi.Array(siteIDs))

		return nil
	})
}
