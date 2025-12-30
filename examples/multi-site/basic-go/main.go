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

	"github.com/pulumi/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

type SiteConfig struct {
	Name        string
	DisplayName string
	ShortName   string
	TimeZone    string
}

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Define site configurations
		siteConfigs := []SiteConfig{
			{
				Name:        "marketing-site",
				DisplayName: "Marketing Site",
				ShortName:   "marketing-site",
				TimeZone:    "America/Los_Angeles",
			},
			{
				Name:        "docs-site",
				DisplayName: "Documentation Site",
				ShortName:   "docs-site",
				TimeZone:    "America/New_York",
			},
			{
				Name:        "blog-site",
				DisplayName: "Blog Site",
				ShortName:   "blog-site",
				TimeZone:    "America/Chicago",
			},
		}

		// Create sites
		siteIDs := make([]pulumi.StringOutput, 0)

		for _, config := range siteConfigs {
			site, err := webflow.NewSite(ctx, config.Name, &webflow.SiteArgs{
				DisplayName: pulumi.String(config.DisplayName),
				ShortName:   pulumi.String(config.ShortName),
				TimeZone:    pulumi.String(config.TimeZone),
			})
			if err != nil {
				return fmt.Errorf("failed to create site %s: %w", config.Name, err)
			}

			siteIDs = append(siteIDs, site.ID)

			// Create robots.txt for each site
			_, err = webflow.NewRobotsTxt(ctx, fmt.Sprintf("%s-robots", config.Name),
				&webflow.RobotsTxtArgs{
					SiteID: site.ID,
					Content: pulumi.String("User-agent: *\nAllow: /"),
				})
			if err != nil {
				return fmt.Errorf("failed to create robots.txt for site %s: %w",
					config.Name, err)
			}

			// Export individual site ID
			ctx.Export(fmt.Sprintf("%s-id", config.Name), site.ID)
		}

		// Export all site IDs as array
		ctx.Export("all-site-ids", pulumi.Array(siteIDs))
		ctx.Export("site-count", pulumi.Int(len(siteConfigs)))

		return nil
	})
}
