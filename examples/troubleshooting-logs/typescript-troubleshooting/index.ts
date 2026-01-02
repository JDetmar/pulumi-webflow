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

import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Log configuration loading (INFO level - always shown)
pulumi.log.info("🔍 Loading configuration for troubleshooting example");

// Log detailed config (DEBUG level - only with --verbose)
pulumi.log.debug("Pulumi context initialized for troubleshooting");

// Example: Troubleshooting authentication
pulumi.log.info("🔐 Verifying Webflow API authentication");
pulumi.log.debug("Token source: Pulumi config (credentials redacted in logs)");

// Create a site with logging at each step
pulumi.log.info("🏗️  Creating Webflow site");

const site = new webflow.Site("troubleshooting-site", {
    displayName: "Troubleshooting Example Site",
    shortName: "troubleshoot-demo",
    timeZone: "America/Los_Angeles",
});

// Log resource creation (ID only shown after deployment)
site.id.apply((id: string) => {
    pulumi.log.info(`✅ Site created successfully: ${id}`);
    pulumi.log.debug(`Site details: displayName=Troubleshooting Example Site`);
    return id;
});

// Example: Error handling with logging
site.id.apply((id: string) => {
    pulumi.log.info("🤖 Configuring robots.txt");

    const robots = new webflow.RobotsTxt("troubleshoot-robots", {
        siteId: id,
        content: "User-agent: *\nAllow: /\n",
    });

    pulumi.log.info("✅ Robots.txt configured successfully");
    return robots;
});

// Export with logging
pulumi.export("siteId", site.id);
pulumi.log.info("📤 Exported site ID for reference");
