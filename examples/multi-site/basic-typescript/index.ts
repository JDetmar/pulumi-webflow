import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Example: Create multiple Webflow sites in a single Pulumi program
// This demonstrates the basic pattern for managing multiple sites programmatically

// Define site configurations as an array
const siteConfigs = [
  {
    name: "marketing-site",
    displayName: "Marketing Site",
    shortName: "marketing-site",
    timeZone: "America/Los_Angeles",
  },
  {
    name: "docs-site",
    displayName: "Documentation Site",
    shortName: "docs-site",
    timeZone: "America/New_York",
  },
  {
    name: "blog-site",
    displayName: "Blog Site",
    shortName: "blog-site",
    timeZone: "America/Chicago",
  },
];

// Create all sites using map pattern
const sites = siteConfigs.map((config) => {
  const site = new webflow.Site(config.name, {
    displayName: config.displayName,
    shortName: config.shortName,
    timeZone: config.timeZone,
  });

  // Configure robots.txt for each site
  new webflow.RobotsTxt(`${config.name}-robots`, {
    siteID: site.id,
    content: `User-agent: *
Allow: /`,
  });

  return site;
});

// Export site IDs for reference
siteConfigs.forEach((config, index) => {
  pulumi.export(`${config.name}-id`, sites[index].id);
});

// Export all site IDs as a list
pulumi.export("all-site-ids", sites.map((site) => site.id));
