import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const collectionId = config.require("collectionId");
const environment = config.get("environment") || "development";

/**
 * CollectionItem Example - Creating and Managing CMS Content
 *
 * This example demonstrates how to manage collection items (blog posts, products, etc.)
 * in your Webflow CMS. Collection items are the individual content entries within a
 * CMS collection.
 *
 * Prerequisites:
 * - A Webflow collection ID (get from your Webflow dashboard)
 * - Knowledge of your collection's field schema (field names/slugs)
 */

// Example 1: Draft Blog Post
// Create a blog post that exists in the CMS but is NOT published to the live site
const draftBlogPost = new webflow.CollectionItem("draft-blog-post", {
  collectionId: collectionId,
  fieldData: {
    name: "Getting Started with Webflow CMS",
    slug: "getting-started-webflow-cms",
    // Add your custom fields here based on your collection schema
    // Example custom fields (uncomment and modify based on your schema):
    // "post-body": "Learn how to use Webflow CMS to manage your content...",
    // "author": "John Doe",
    // "publish-date": "2025-01-06",
    // "featured-image": "https://example.com/image.jpg",
  },
  isDraft: true, // Not published to live site
  isArchived: false,
});

// Example 2: Published Product
// Create a product that is immediately visible on the live site
const publishedProduct = new webflow.CollectionItem("published-product", {
  collectionId: collectionId,
  fieldData: {
    name: "Premium Widget",
    slug: "premium-widget",
    // Add your custom fields here based on your collection schema
    // Example custom fields (uncomment and modify based on your schema):
    // "price": 99.99,
    // "description": "The best widget on the market",
    // "category": "Electronics",
    // "in-stock": true,
  },
  isDraft: false, // Published to live site
  isArchived: false,
});

// Example 3: Archived Content
// Create an item that is archived (hidden but retained for records)
const archivedItem = new webflow.CollectionItem("archived-item", {
  collectionId: collectionId,
  fieldData: {
    name: "Discontinued Product",
    slug: "discontinued-product-archive",
  },
  isDraft: true,
  isArchived: true, // Hidden from both CMS and live site
});

// Example 4: Bulk Content Creation
// Create multiple items efficiently using a loop
const bulkItems: webflow.CollectionItem[] = [];
const contentData = [
  {
    name: "Introduction to TypeScript",
    slug: "intro-typescript",
    category: "Tutorial",
  },
  {
    name: "Advanced Pulumi Patterns",
    slug: "advanced-pulumi-patterns",
    category: "Tutorial",
  },
  {
    name: "Webflow API Best Practices",
    slug: "webflow-api-best-practices",
    category: "Guide",
  },
];

contentData.forEach((data, index) => {
  const item = new webflow.CollectionItem(`bulk-item-${index}`, {
    collectionId: collectionId,
    fieldData: {
      name: data.name,
      slug: data.slug,
      // Add your custom fields here
      // "category": data.category,
    },
    isDraft: true, // Start as drafts
  });
  bulkItems.push(item);
});

// Example 5: Localized Content (optional - only if your site uses localization)
// Uncomment if your Webflow site has localization enabled
// const localizedItem = new webflow.CollectionItem("localized-item", {
//   collectionId: collectionId,
//   fieldData: {
//     name: "Bienvenue",
//     slug: "bienvenue",
//   },
//   cmsLocaleId: "fr-FR", // French locale
//   isDraft: false,
// });

// Export the resource IDs for reference
export const deployedCollectionId = collectionId;
export const deployedEnvironment = environment;

// Draft blog post exports
export const draftPostId = draftBlogPost.id;
export const draftPostItemId = draftBlogPost.itemId;
export const draftPostCreatedOn = draftBlogPost.createdOn;

// Published product exports
export const publishedProductId = publishedProduct.id;
export const publishedProductItemId = publishedProduct.itemId;
export const publishedProductLastUpdated = publishedProduct.lastUpdated;

// Archived item exports
export const archivedItemId = archivedItem.id;
export const archivedItemItemId = archivedItem.itemId;

// Bulk items exports
export const bulkItemIds = bulkItems.map((item) => item.id);
export const bulkItemItemIds = bulkItems.map((item) => item.itemId);

// Print deployment success message
const totalItems = 3 + bulkItems.length;
console.log(`✅ Successfully deployed ${totalItems} collection items to collection ${collectionId}`);
console.log(`   Environment: ${environment}`);
console.log(`   Draft items: ${1 + bulkItems.length}`);
console.log(`   Published items: 1`);
console.log(`   Archived items: 1`);
