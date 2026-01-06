import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * SiteCustomCode Example - Managing Site-Wide Custom Scripts
 *
 * This example demonstrates how to apply registered custom JavaScript scripts
 * to your entire Webflow site. Scripts applied at the site level will be included
 * on all pages unless overridden at the page level.
 *
 * Prerequisites:
 * - Scripts must first be registered using the RegisteredScript resource
 * - You'll need the script ID and version from your registered scripts
 */

// Step 1: Register custom scripts (prerequisite)
const analyticsScript = new webflow.RegisteredScript("analytics-script", {
  siteId: siteId,
  displayName: "GoogleAnalytics",
  hostedLocation: "https://cdn.example.com/ga-v4.js",
  integrityHash: "sha384-oqVuAfXRKap7fdgcCY5uykM6+R9GqQ8K/uxy9rx7HNQlGYl1kPzQho1wx4JwY8wC",
  version: "4.0.0",
  canCopy: true,
});

const chatWidgetScript = new webflow.RegisteredScript("chat-widget", {
  siteId: siteId,
  displayName: "LiveChat",
  hostedLocation: "https://cdn.example.com/livechat-v2.min.js",
  integrityHash: "sha384-abc123def456ghi789jkl012mno345pqr678stu901vwx234yz567ABC890DEF123",
  version: "2.5.0",
  canCopy: true,
});

const cookieConsentScript = new webflow.RegisteredScript("cookie-consent", {
  siteId: siteId,
  displayName: "CookieConsent",
  hostedLocation: "https://cdn.example.com/cookie-consent.js",
  integrityHash: "sha256-47DEQpj8HBSa+/TImW+5JCeuQeRkm5NMpJWZG3hSuFU=",
  version: "1.0.0",
  canCopy: true,
});

// Step 2: Apply scripts to the entire site
const siteScripts = new webflow.SiteCustomCode("site-wide-scripts", {
  siteId: siteId,
  scripts: [
    {
      // Analytics in header - loads before page renders
      id: analyticsScript.id,
      version: "4.0.0",
      location: "header",
      attributes: {
        "data-site-id": "GA-123456789",
      },
    },
    {
      // Cookie consent in header - must load early
      id: cookieConsentScript.id,
      version: "1.0.0",
      location: "header",
      attributes: {
        "data-theme": "dark",
        "data-position": "bottom-right",
      },
    },
    {
      // Chat widget in footer - loads after page content
      id: chatWidgetScript.id,
      version: "2.5.0",
      location: "footer",
      attributes: {
        "data-widget-id": "chat-widget-123",
        "data-auto-open": "false",
      },
    },
  ],
});

// Example: Minimal configuration with single script
const minimalSiteScripts = new webflow.SiteCustomCode("minimal-site-scripts", {
  siteId: siteId,
  scripts: [
    {
      id: analyticsScript.id,
      version: "4.0.0",
      location: "header",
    },
  ],
});

// Export useful information
export const deployedSiteId = siteId;
export const siteScriptsCreatedOn = siteScripts.createdOn;
export const siteScriptsLastUpdated = siteScripts.lastUpdated;
export const appliedScriptCount = 3;

// Export script IDs for reference
export const analyticsScriptId = analyticsScript.id;
export const chatWidgetScriptId = chatWidgetScript.id;
export const cookieConsentScriptId = cookieConsentScript.id;

// Print deployment success message
const message = pulumi.interpolate`Successfully applied ${appliedScriptCount} site-wide custom scripts to site ${siteId}`;
message.apply((m) => console.log(m));
