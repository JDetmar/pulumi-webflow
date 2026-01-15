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

// Load configuration from current stack
const config = new pulumi.Config();

// Stack-specific configuration (defined in Pulumi.<stack>.yaml)
const environmentName = config.require("environmentName");  // "dev", "staging", "prod"

// Validate environment configuration
const validEnvironments = ["dev", "staging", "prod"];
if (!validEnvironments.includes(environmentName)) {
    throw new Error(
        `Invalid environment "${environmentName}". ` +
        `Must be one of: ${validEnvironments.join(", ")}`
    );
}

// Production safety check: require explicit confirmation
const isProd = environmentName === "prod";
if (isProd) {
    const confirmation = config.get("prodDeploymentConfirmed");
    if (confirmation !== "yes") {
        throw new Error(
            "⚠️  Production deployment requires explicit confirmation.\n" +
            "Run: pulumi config set prodDeploymentConfirmed yes\n" +
            "This prevents accidental production deployments."
        );
    }
}

// Site configuration interface - each site is a distinct entity
interface SiteConfig {
    displayName: string;
    shortName: string;
    allowIndexing: boolean;  // Whether search engines can index this site
    redirects?: Array<{
        sourcePath: string;
        destinationPath: string;
        statusCode: number;
    }>;
}

// Load site configurations from stack config
// Each environment defines its own sites with environment-specific settings
const sitesConfig = config.requireObject<Record<string, SiteConfig>>("sites");

pulumi.log.info(
    `🚀 Deploying ${Object.keys(sitesConfig).length} sites to ${environmentName} environment`
);

// Create each configured site
const siteExports: Record<string, pulumi.Output<string>> = {};

for (const [siteKey, siteConfig] of Object.entries(sitesConfig)) {
    // Create the site with its specific configuration
    const site = new webflow.Site(siteKey, {
        displayName: siteConfig.displayName,
        shortName: siteConfig.shortName,
    });

    siteExports[`${siteKey}-id`] = site.id;

    // Configure robots.txt based on site's indexing preference
    const robotsContent = siteConfig.allowIndexing
        ? "User-agent: *\nAllow: /\n"
        : `User-agent: *\nDisallow: /\n\n# ${environmentName.toUpperCase()} - Do not index`;

    new webflow.RobotsTxt(`${siteKey}-robots`, {
        siteId: site.id,
        content: robotsContent,
    });

    // Create any configured redirects for this site
    if (siteConfig.redirects) {
        for (const redirect of siteConfig.redirects) {
            const redirectName = `${siteKey}-redirect${redirect.sourcePath.replace(/\//g, "-")}`;
            new webflow.Redirect(redirectName, {
                siteId: site.id,
                sourcePath: redirect.sourcePath,
                destinationPath: redirect.destinationPath,
                statusCode: redirect.statusCode,
            });
        }
    }

    pulumi.log.info(`✅ Configured site: ${siteKey} (${siteConfig.displayName})`);
}

// Export site IDs for reference
for (const [key, value] of Object.entries(siteExports)) {
    pulumi.export(key, value);
}

// Export summary information
pulumi.export("environment", environmentName);
pulumi.export("siteCount", Object.keys(sitesConfig).length);
pulumi.export("siteNames", Object.keys(sitesConfig));
pulumi.export("stackName", pulumi.getStack());
pulumi.export("projectName", pulumi.getProject());

pulumi.log.info(
    `✅ Deployment complete: ${Object.keys(sitesConfig).length} sites configured for ${environmentName}`
);
