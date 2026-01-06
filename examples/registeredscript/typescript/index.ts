import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * RegisteredScript Example - Managing Custom JavaScript Scripts
 *
 * This example demonstrates how to register and manage externally hosted JavaScript
 * scripts in your Webflow site's script registry. Registered scripts can then be
 * deployed across your site using the SiteCustomCode or PageCustomCode resources.
 *
 * Scripts must be:
 * - Externally hosted (CDN or your server)
 * - Include Sub-Resource Integrity (SRI) hash for security
 * - Follow semantic versioning
 */

// Example 1: Register a CDN-hosted analytics script
const analyticsScript = new webflow.RegisteredScript("analytics-script", {
  siteId: siteId,
  displayName: "AnalyticsTracker",
  hostedLocation: "https://cdn.example.com/analytics-tracker.js",
  integrityHash: "sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8wC",
  version: "1.0.0",
  canCopy: true, // Allow copying when site is duplicated
});

// Example 2: Register a custom CMS slider script
const cmsSliderScript = new webflow.RegisteredScript("cms-slider", {
  siteId: siteId,
  displayName: "CmsSlider",
  hostedLocation: "https://cdn.example.com/cms-slider.min.js",
  integrityHash: "sha384-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567ABC890DEF123",
  version: "2.1.5",
  canCopy: false, // Don't copy when duplicating site
});

// Example 3: Register a custom widget with SHA-256 hash
const customWidget = new webflow.RegisteredScript("custom-widget", {
  siteId: siteId,
  displayName: "MyCustomWidget123",
  hostedLocation: "https://widgets.example.com/custom-widget-v3.js",
  integrityHash: "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  version: "3.0.0",
  canCopy: true,
});

// Example 4: Register multiple versions of the same script
// Useful for gradual rollouts or A/B testing
const marketingScriptV1 = new webflow.RegisteredScript("marketing-v1", {
  siteId: siteId,
  displayName: "MarketingPixel",
  hostedLocation: "https://cdn.example.com/marketing-v1.0.0.js",
  integrityHash: "sha384-v1hash000000000000000000000000000000000000000000000000000000000",
  version: "1.0.0",
  canCopy: true,
});

const marketingScriptV2 = new webflow.RegisteredScript("marketing-v2", {
  siteId: siteId,
  displayName: "MarketingPixel",
  hostedLocation: "https://cdn.example.com/marketing-v2.0.0.js",
  integrityHash: "sha384-v2hash000000000000000000000000000000000000000000000000000000000",
  version: "2.0.0",
  canCopy: true,
});

// Export the script IDs and details for use in other resources
export const deployedSiteId = siteId;
export const analyticsScriptId = analyticsScript.id;
export const cmsSliderScriptId = cmsSliderScript.id;
export const customWidgetScriptId = customWidget.id;
export const marketingV1ScriptId = marketingScriptV1.id;
export const marketingV2ScriptId = marketingScriptV2.id;

// Export created and updated timestamps
export const analyticsScriptCreatedOn = analyticsScript.createdOn;
export const analyticsScriptLastUpdated = analyticsScript.lastUpdated;

// Print deployment success message
const message = pulumi.interpolate`Successfully registered 5 custom scripts to site ${siteId}`;
message.apply((m) => console.log(m));
