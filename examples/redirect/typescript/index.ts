import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * Redirect Example - Creating and Managing URL Redirects
 *
 * This example demonstrates how to manage URL redirects for your Webflow sites.
 * Redirects are useful for:
 * - Moving content to new URLs (SEO-friendly 301 redirects)
 * - Temporary redirects (302 redirects)
 * - External site redirects
 * - Handling URL pattern changes
 */

// Example 1: Permanent Redirect (301) - Best for content moves
const permanentRedirect = new webflow.Redirect("old-blog-to-new-blog", {
  siteId: siteId,
  sourcePath: "/blog/old-article",
  destinationPath: "/blog/articles/updated-article",
  statusCode: 301,
});

// Example 2: Temporary Redirect (302) - Use for temporary changes
const temporaryRedirect = new webflow.Redirect("temporary-landing-page", {
  siteId: siteId,
  sourcePath: "/old-campaign",
  destinationPath: "/new-campaign-2025",
  statusCode: 302,
});

// Example 3: External Redirect (301) - Redirect to another domain
const externalRedirect = new webflow.Redirect("external-partner-link", {
  siteId: siteId,
  sourcePath: "/partner",
  destinationPath: "https://partner-site.com",
  statusCode: 301,
});

// Example 4: Bulk Redirects using Loop
const bulkRedirects: webflow.Redirect[] = [];
const redirectMappings = [
  { old: "/product-a", new: "/products/product-a" },
  { old: "/product-b", new: "/products/product-b" },
  { old: "/product-c", new: "/products/product-c" },
];

redirectMappings.forEach((mapping, index) => {
  const redirect = new webflow.Redirect(`bulk-redirect-${index}`, {
    siteId: siteId,
    sourcePath: mapping.old,
    destinationPath: mapping.new,
    statusCode: 301,
  });
  bulkRedirects.push(redirect);
});

// Export the redirect resources for reference
export const deployedSiteId = siteId;
export const permanentRedirectId = permanentRedirect.id;
export const temporaryRedirectId = temporaryRedirect.id;
export const externalRedirectId = externalRedirect.id;
export const bulkRedirectIds = bulkRedirects.map((r) => r.id);

// Print deployment success message
const message = pulumi.interpolate`✅ Successfully deployed ${bulkRedirects.length + 3} redirects to site ${siteId}`;
message.apply((m) => console.log(m));
