import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const displayName = config.require("displayName");
const shortName = config.require("shortName");
const customDomain = config.get("customDomain");
const timezone = config.get("timezone") || "America/New_York";

/**
 * Site Example - Creating and Managing Webflow Sites
 *
 * This example demonstrates how to create and manage Webflow sites using Pulumi.
 * Sites are the primary resource in the Webflow provider.
 */

// Example 1: Basic Site Creation
const basicSite = new webflow.Site("basic-site", {
  displayName: displayName,
  shortName: shortName,
  timezone: timezone,
});

// Example 2: Site with Custom Domain
let siteWithDomain: webflow.Site | undefined;
if (customDomain) {
  siteWithDomain = new webflow.Site("site-with-domain", {
    displayName: `${displayName}-domain`,
    shortName: `${shortName}-domain`,
    customDomain: customDomain,
    timezone: timezone,
  });
}

// Example 3: Multi-Environment Site Configuration
const environments = ["development", "staging", "production"];
const environmentSites = environments.map(
  (env) =>
    new webflow.Site(`site-${env}`, {
      displayName: `${displayName}-${env}`,
      shortName: `${shortName}-${env}`,
      timezone: timezone,
    })
);

// Example 4: Site with Configuration
const configuredSite = new webflow.Site("configured-site", {
  displayName: `${displayName}-configured`,
  shortName: `${shortName}-configured`,
  timezone: timezone,
});

// Export the site resources for reference
export const basicSiteId = basicSite.id;
export const basicSiteName = basicSite.displayName;
export const customDomainSiteId = siteWithDomain?.id || "not-created";
export const environmentSiteIds = environmentSites.map((s) => s.id);
export const configuredSiteId = configuredSite.id;

// Print deployment success message
const message = pulumi.interpolate`✅ Successfully created ${environmentSites.length + 1} sites`;
message.apply((m) => console.log(m));
