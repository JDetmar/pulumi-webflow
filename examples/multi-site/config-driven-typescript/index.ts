import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";
import * as fs from "fs";
import * as yaml from "js-yaml";

// Example: Configuration-driven multi-site management
// Load site configurations from YAML for scalable fleet management

interface SiteConfig {
  name: string;
  displayName: string;
  shortName: string;
  redirects?: Array<{
    sourcePath: string;
    destinationPath: string;
    statusCode: number;
  }>;
  robotsTxtContent?: string;
}

interface ConfigFile {
  sites: SiteConfig[];
  defaults?: {
    robotsTxtContent?: string;
  };
}

// Load configuration from YAML file
const configPath = "./sites.yaml";
const configContent = fs.readFileSync(configPath, "utf8");
const config = yaml.load(configContent) as ConfigFile;

// Get default values
const defaultRobotsTxt =
  config.defaults?.robotsTxtContent ||
  `User-agent: *
Allow: /`;

// Create all sites from configuration
const sites = config.sites.map((siteConfig) => {
  // Use defaults where not specified
  const robotsTxtContent = siteConfig.robotsTxtContent || defaultRobotsTxt;

  // Create the site
  const site = new webflow.Site(siteConfig.name, {
    displayName: siteConfig.displayName,
    shortName: siteConfig.shortName,
  });

  // Add robots.txt configuration
  new webflow.RobotsTxt(`${siteConfig.name}-robots`, {
    siteID: site.id,
    content: robotsTxtContent,
  });

  // Add redirects if specified
  if (siteConfig.redirects && siteConfig.redirects.length > 0) {
    siteConfig.redirects.forEach((redirect, idx) => {
      new webflow.Redirect(`${siteConfig.name}-redirect-${idx}`, {
        siteID: site.id,
        sourcePath: redirect.sourcePath,
        destinationPath: redirect.destinationPath,
        statusCode: redirect.statusCode,
      });
    });
  }

  return site;
});

// Export information about deployed sites
pulumi.export("deployed-sites", config.sites.length);
pulumi.export("site-names", config.sites.map((s) => s.name));

// Export individual site IDs
config.sites.forEach((siteConfig, idx) => {
  pulumi.export(`${siteConfig.name}-id`, sites[idx].id);
});

// Export all site IDs
pulumi.export("all-site-ids", sites.map((site) => site.id));
