import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values (these should be set via `pulumi config set`)
const siteId = config.requireSecret("siteId");

/**
 * Deploy a RobotsTxt resource to your Webflow site
 *
 * This example creates a robots.txt file that:
 * - Allows all search engine crawlers (User-agent: *)
 * - Allows Google's bot (Googlebot) to crawl all pages
 *
 * You can customize the robots.txt content by modifying the content string below
 */
const robotsTxt = new webflow.RobotsTxt("my-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

User-agent: Googlebot
Allow: /`,
});

// Export the site ID for reference
export const deployedSiteId = siteId;
export const robotsTxtId = robotsTxt.id;

// Print a success message
const message = pulumi.interpolate`✅ Successfully deployed robots.txt to site ${siteId}`;
message.apply((m) => console.log(m));
