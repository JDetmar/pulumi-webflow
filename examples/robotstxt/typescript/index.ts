import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * RobotsTxt Example - Creating and Managing robots.txt Files
 *
 * This example demonstrates how to manage robots.txt files for your Webflow sites.
 * The robots.txt file controls how search engine crawlers interact with your site.
 */

// Example 1: Allow All Crawlers (most common for public sites)
const allowAllRobots = new webflow.RobotsTxt("allow-all-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

# Allow specific crawler access with no delays
User-agent: Googlebot
Allow: /
Crawl-delay: 0

User-agent: Bingbot
Allow: /
Crawl-delay: 1`,
});

// Example 2: Selective Blocking (for staging/development)
const selectiveBlockRobots = new webflow.RobotsTxt("selective-block-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

# Disallow admin and private sections
Disallow: /admin/
Disallow: /private/
Disallow: /staging/
Disallow: /test/

# Block specific crawlers
User-agent: AhrefsBot
Disallow: /

User-agent: SemrushBot
Disallow: /`,
});

// Example 3: Restrict Directories (protect API and backend)
const restrictDirectoriesRobots = new webflow.RobotsTxt("restrict-directories-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /
Disallow: /api/
Disallow: /internal/
Disallow: /*.json$
Disallow: /*.xml$

# Specify sitemap location
Sitemap: https://example.com/sitemap.xml`,
});

// Export the robot resources for reference
export const deployedSiteId = siteId;
export const allowAllRobotsId = allowAllRobots.id;
export const allowAllRobotsLastModified = allowAllRobots.lastModified;
export const selectiveBlockRobotsId = selectiveBlockRobots.id;
export const restrictDirectoriesRobotsId = restrictDirectoriesRobots.id;

// Print deployment success message
const message = pulumi.interpolate`✅ Successfully deployed RobotsTxt resources to site ${siteId}`;
message.apply((m) => console.log(m));
