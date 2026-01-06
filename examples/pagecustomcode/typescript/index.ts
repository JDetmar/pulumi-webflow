import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");
const landingPageId = config.require("landingPageId");
const productPageId = config.require("productPageId");

/**
 * PageCustomCode Example - Managing Page-Specific Custom Scripts
 *
 * This example demonstrates how to apply registered custom JavaScript scripts
 * to specific pages in your Webflow site. This is useful when you need different
 * scripts on different pages, or want to override site-level script configuration.
 *
 * Prerequisites:
 * - Scripts must first be registered using the RegisteredScript resource
 * - You'll need page IDs from your Webflow site (24-character hex strings)
 * - Configure page IDs using: pulumi config set landingPageId <id>
 */

// Step 1: Register custom scripts (prerequisite)
const conversionTrackingScript = new webflow.RegisteredScript("conversion-tracking", {
  siteId: siteId,
  displayName: "ConversionPixel",
  hostedLocation: "https://cdn.example.com/conversion-pixel.js",
  integrityHash: "sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8wC",
  version: "3.2.1",
  canCopy: true,
});

const productViewerScript = new webflow.RegisteredScript("product-viewer", {
  siteId: siteId,
  displayName: "Product360Viewer",
  hostedLocation: "https://cdn.example.com/product-360-viewer.min.js",
  integrityHash: "sha384-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567ABC890DEF123",
  version: "1.5.0",
  canCopy: true,
});

const heatmapScript = new webflow.RegisteredScript("heatmap-tracking", {
  siteId: siteId,
  displayName: "HeatmapTracker",
  hostedLocation: "https://cdn.example.com/heatmap-v2.js",
  integrityHash: "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  version: "2.0.0",
  canCopy: false,
});

// Step 2: Apply scripts to specific pages

// Example 1: Landing page with conversion tracking and heatmap
const landingPageScripts = new webflow.PageCustomCode("landing-page-scripts", {
  pageId: landingPageId,
  scripts: [
    {
      // Conversion tracking in header
      id: conversionTrackingScript.id,
      version: "3.2.1",
      location: "header",
      attributes: {
        "data-campaign-id": "summer-2025",
        "data-conversion-type": "landing",
      },
    },
    {
      // Heatmap tracking in footer
      id: heatmapScript.id,
      version: "2.0.0",
      location: "footer",
      attributes: {
        "data-page-type": "landing",
        "data-track-clicks": "true",
        "data-track-scrolling": "true",
      },
    },
  ],
});

// Example 2: Product page with 360 viewer and conversion tracking
const productPageScripts = new webflow.PageCustomCode("product-page-scripts", {
  pageId: productPageId,
  scripts: [
    {
      // 360 product viewer in footer (needs DOM elements)
      id: productViewerScript.id,
      version: "1.5.0",
      location: "footer",
      attributes: {
        "data-viewer-container": "#product-viewer",
        "data-zoom-enabled": "true",
        "data-auto-rotate": "false",
      },
    },
    {
      // Conversion tracking in header
      id: conversionTrackingScript.id,
      version: "3.2.1",
      location: "header",
      attributes: {
        "data-campaign-id": "product-launch",
        "data-conversion-type": "product-view",
      },
    },
  ],
});

// Example 3: Minimal configuration with single script
const minimalPageScripts = new webflow.PageCustomCode("thank-you-page-scripts", {
  pageId: landingPageId, // Reusing landingPageId for example
  scripts: [
    {
      id: conversionTrackingScript.id,
      version: "3.2.1",
      location: "header",
    },
  ],
});

// Export useful information
export const deployedSiteId = siteId;
export const landingPageScriptsCreatedOn = landingPageScripts.createdOn;
export const landingPageScriptsLastUpdated = landingPageScripts.lastUpdated;
export const productPageScriptsCreatedOn = productPageScripts.createdOn;

// Export script IDs for reference
export const conversionTrackingScriptId = conversionTrackingScript.id;
export const productViewerScriptId = productViewerScript.id;
export const heatmapScriptId = heatmapScript.id;

// Export page IDs to confirm configuration
export const configuredLandingPageId = landingPageId;
export const configuredProductPageId = productPageId;

// Print deployment success message
const message = pulumi.interpolate`Successfully applied custom scripts to 2 pages in site ${siteId}`;
message.apply((m) => console.log(m));
