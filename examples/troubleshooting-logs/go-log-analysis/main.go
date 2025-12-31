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

	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		// Log initialization with detailed debug information
		ctx.Log.Info("🔍 Initializing Webflow site creation with logging analysis", nil)
		ctx.Log.Debug("Context initialized for deployment analysis", nil)

		// Demonstrate structured logging for troubleshooting
		ctx.Log.Info("🏗️  Creating Webflow site resource", nil)
		ctx.Log.Debug("Resource type: webflow:index:Site", nil)

		// Create a site with logging at each phase
		site, err := webflow.NewSite(ctx, "go-logging-analysis", &webflow.SiteArgs{
			DisplayName: pulumi.String("Go Logging Analysis Example"),
			ShortName:   pulumi.String("go-logging-demo"),
			TimeZone:    pulumi.String("America/Los_Angeles"),
		})
		if err != nil {
			ctx.Log.Error(fmt.Sprintf("❌ Failed to create site: %v", err), nil)
			return err
		}

		ctx.Log.Info("✅ Site resource created successfully", nil)

		// Log resource IDs and diagnostic information
		site.ID().ApplyT(func(id string) error {
			ctx.Log.Debug(fmt.Sprintf("Site ID: %s", id), nil)
			ctx.Log.Info(fmt.Sprintf("✅ Site provisioned: %s", id), nil)
			return nil
		})

		// Create robots.txt with logging
		ctx.Log.Info("🤖 Configuring robots.txt resource", nil)

		site.ID().ApplyT(func(siteID string) error {
			_, err := webflow.NewRobotsTxt(ctx, "go-robots", &webflow.RobotsTxtArgs{
				SiteID:  pulumi.String(siteID),
				Content: pulumi.String("User-agent: *\nAllow: /\n"),
			})
			if err != nil {
				ctx.Log.Error(fmt.Sprintf("❌ Failed to create robots.txt: %v", err), nil)
				return err
			}

			ctx.Log.Info("✅ Robots.txt configured successfully", nil)
			return nil
		})

		// Export stack outputs
		ctx.Export("siteID", site.ID())
		ctx.Log.Info("📤 Stack outputs exported", nil)

		// Final log summary
		ctx.Log.Info("🎉 Webflow infrastructure provisioned successfully", nil)
		ctx.Log.Debug("Deployment complete - all resources created and configured", nil)

		return nil
	})
}
