import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * Ecommerce Settings Example - Reading Webflow Ecommerce Configuration
 *
 * This example demonstrates how to import and track Webflow ecommerce settings
 * as infrastructure state. This is a read-only resource that reflects the
 * ecommerce configuration set up in the Webflow dashboard.
 *
 * Prerequisites:
 * - Ecommerce must be enabled on your Webflow site through the dashboard
 * - Your API token must have the 'ecommerce:read' scope
 *
 * Use Cases:
 * - Verify ecommerce is enabled on a site before deploying ecommerce-related resources
 * - Reference the site's default currency in other resources
 * - Track ecommerce configuration as part of your infrastructure state
 * - Audit when ecommerce was enabled on your site
 */

// Example 1: Basic Ecommerce Settings Import
// Import the ecommerce settings for a site to track them in Pulumi state
const ecommerceSettings = new webflow.EcommerceSettings("site-ecommerce", {
  siteId: siteId,
});

// Export settings information for reference
export const deployedSiteId = siteId;

// Ecommerce settings outputs
export const ecommerceSiteId = ecommerceSettings.siteId;
export const defaultCurrency = ecommerceSettings.defaultCurrency;
export const ecommerceCreatedOn = ecommerceSettings.createdOn;

// Print deployment success message
const message = pulumi.interpolate`Ecommerce Settings imported for site ${siteId}:
  - Default Currency: ${ecommerceSettings.defaultCurrency}
  - Ecommerce Enabled On: ${ecommerceSettings.createdOn}`;
message.apply((m) => console.log(m));
