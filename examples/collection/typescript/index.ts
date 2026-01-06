import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * Collection Example - Creating and Managing CMS Collections
 *
 * This example demonstrates how to create CMS collections for your Webflow sites.
 * Collections are containers for structured content items (blog posts, products, etc.).
 *
 * Important: Webflow collections do not support updates via API.
 * Any changes to collection properties will require replacement (delete + recreate).
 */

// Example 1: Blog Posts Collection
// A common pattern for blog content with all required fields
const blogCollection = new webflow.Collection("blog-posts-collection", {
  siteId: siteId,
  displayName: "Blog Posts",
  singularName: "Blog Post",
  slug: "blog-posts",
});

// Example 2: Products Collection with Auto-Generated Slug
// Omit the slug to let Webflow auto-generate it from the displayName
const productsCollection = new webflow.Collection("products-collection", {
  siteId: siteId,
  displayName: "Products",
  singularName: "Product",
  // slug is optional - Webflow will generate "products" automatically
});

// Example 3: Team Members Collection
// Demonstrates custom slug different from display name
const teamCollection = new webflow.Collection("team-members-collection", {
  siteId: siteId,
  displayName: "Team Members",
  singularName: "Team Member",
  slug: "team",
});

// Example 4: Portfolio Items Collection
// Another common use case for showcasing work
const portfolioCollection = new webflow.Collection("portfolio-collection", {
  siteId: siteId,
  displayName: "Portfolio Items",
  singularName: "Portfolio Item",
  slug: "portfolio",
});

// Example 5: Dynamic Collections Based on Config
// Create collections based on configuration for multi-environment setups
const environment = config.get("environment") || "development";
const testCollection = new webflow.Collection(`test-collection-${environment}`, {
  siteId: siteId,
  displayName: `Test Collection (${environment})`,
  singularName: "Test Item",
  slug: `test-${environment}`,
});

// Export collection details for reference
export const deployedSiteId = siteId;
export const blogCollectionId = blogCollection.id;
export const blogCollectionName = blogCollection.displayName;
export const blogCollectionSlug = blogCollection.slug;
export const blogCollectionCreatedOn = blogCollection.createdOn;

export const productsCollectionId = productsCollection.id;
export const teamCollectionId = teamCollection.id;
export const portfolioCollectionId = portfolioCollection.id;
export const testCollectionId = testCollection.id;

// Export a summary of all collections
export const allCollections = pulumi.all([
  blogCollection.displayName,
  productsCollection.displayName,
  teamCollection.displayName,
  portfolioCollection.displayName,
  testCollection.displayName,
]).apply(names => names.join(", "));

// Print deployment success message
const message = pulumi.interpolate`✅ Successfully deployed 5 collections to site ${siteId}`;
message.apply((m) => console.log(m));
