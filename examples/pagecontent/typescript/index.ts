import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const pageId = config.requireSecret("pageId");

/**
 * PageContent Example - Managing Static Text Content on Webflow Pages
 *
 * This example demonstrates how to manage static text content within existing DOM nodes
 * on a Webflow page. This is useful for:
 * - Programmatically updating text content across multiple pages
 * - Maintaining consistent messaging via infrastructure-as-code
 * - Automating content updates as part of deployments
 *
 * IMPORTANT NOTES:
 * - This resource does NOT manage page structure or layout
 * - It only updates text content within existing DOM nodes
 * - Node IDs must be retrieved from the page's DOM structure first
 * - Use GET /pages/{page_id}/dom endpoint to find node IDs
 * - Drift detection is limited: only verifies the page exists, not content
 */

/**
 * Example 1: Update Hero Section Text
 *
 * Updates the main heading and subtitle in a hero section.
 * You would get these node IDs by fetching the page DOM first.
 */
const heroContent = new webflow.PageContent("hero-section-content", {
  pageId: pageId,
  nodes: [
    {
      nodeId: "hero-heading-node-id",
      text: "Welcome to Our Platform",
    },
    {
      nodeId: "hero-subtitle-node-id",
      text: "Build amazing experiences with our tools",
    },
  ],
});

/**
 * Example 2: Update Footer Copyright Text
 *
 * Keeps copyright year and company information up-to-date.
 */
const currentYear = new Date().getFullYear();
const footerContent = new webflow.PageContent("footer-content", {
  pageId: pageId,
  nodes: [
    {
      nodeId: "footer-copyright-node-id",
      text: `© ${currentYear} Your Company Name. All rights reserved.`,
    },
  ],
});

/**
 * Example 3: Update Multiple Text Blocks
 *
 * Update multiple content sections at once, such as feature descriptions.
 */
const featureContent = new webflow.PageContent("feature-section-content", {
  pageId: pageId,
  nodes: [
    {
      nodeId: "feature-1-title-node-id",
      text: "Fast Performance",
    },
    {
      nodeId: "feature-1-description-node-id",
      text: "Lightning-fast load times for the best user experience.",
    },
    {
      nodeId: "feature-2-title-node-id",
      text: "Secure & Reliable",
    },
    {
      nodeId: "feature-2-description-node-id",
      text: "Enterprise-grade security with 99.9% uptime guarantee.",
    },
  ],
});

// Export resource information for reference
export const deployedPageId = pageId;
export const heroContentId = heroContent.id;
export const heroLastUpdated = heroContent.lastUpdated;
export const footerContentId = footerContent.id;
export const featureContentId = featureContent.id;

// Print deployment success message
const message = pulumi.interpolate`✅ Successfully updated page content for page ${pageId}`;
message.apply((m) => console.log(m));
