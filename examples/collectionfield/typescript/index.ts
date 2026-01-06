import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const collectionId = config.requireSecret("collectionId");

/**
 * CollectionField Example - Creating and Managing CMS Collection Fields
 *
 * This example demonstrates how to create various types of fields for Webflow CMS collections.
 * Collection fields define the structure and validation rules for content items.
 *
 * Field types demonstrated:
 * - PlainText: Single-line text input
 * - RichText: Multi-line rich text editor
 * - Number: Numeric input with min/max validation
 * - DateTime: Date and time picker
 * - Switch: Boolean toggle (true/false)
 * - Image: Single image reference
 * - Email: Email address with validation
 */

// Example 1: Plain Text Field (Required)
// Best for titles, names, short descriptions
const titleField = new webflow.CollectionField("article-title", {
  collectionId: collectionId,
  type: "PlainText",
  displayName: "Article Title",
  slug: "article-title",
  isRequired: true,
  helpText: "Enter the main title for this article (required)",
  validations: {
    maxLength: 100,
  },
});

// Example 2: Rich Text Field
// Best for long-form content like blog posts or descriptions
const contentField = new webflow.CollectionField("article-content", {
  collectionId: collectionId,
  type: "RichText",
  displayName: "Article Content",
  slug: "content",
  isRequired: true,
  helpText: "Main body content of the article",
});

// Example 3: Number Field with Validations
// Best for prices, quantities, ratings, scores
const readTimeField = new webflow.CollectionField("read-time", {
  collectionId: collectionId,
  type: "Number",
  displayName: "Read Time (minutes)",
  slug: "read-time",
  isRequired: false,
  helpText: "Estimated reading time in minutes",
  validations: {
    min: 1,
    max: 120,
    decimalPlaces: 0, // Integer only
  },
});

// Example 4: DateTime Field
// Best for publish dates, event dates, deadlines
const publishDateField = new webflow.CollectionField("publish-date", {
  collectionId: collectionId,
  type: "DateTime",
  displayName: "Publish Date",
  slug: "publish-date",
  isRequired: true,
  helpText: "When this article should be published",
});

// Example 5: Switch Field (Boolean)
// Best for feature flags, visibility toggles, yes/no options
const featuredField = new webflow.CollectionField("is-featured", {
  collectionId: collectionId,
  type: "Switch",
  displayName: "Featured Article",
  slug: "is-featured",
  isRequired: false,
  helpText: "Mark as featured to display on homepage",
});

// Example 6: Email Field
// Best for contact information, author emails
const authorEmailField = new webflow.CollectionField("author-email", {
  collectionId: collectionId,
  type: "Email",
  displayName: "Author Email",
  slug: "author-email",
  isRequired: false,
  helpText: "Contact email for the article author",
});

// Example 7: Image Field
// Best for cover images, thumbnails, hero images
const coverImageField = new webflow.CollectionField("cover-image", {
  collectionId: collectionId,
  type: "Image",
  displayName: "Cover Image",
  slug: "cover-image",
  isRequired: false,
  helpText: "Main cover image for the article",
});

// Example 8: Plain Text Field with Minimal Configuration
// Demonstrates auto-generated slug (not explicitly set)
const shortDescriptionField = new webflow.CollectionField("short-description", {
  collectionId: collectionId,
  type: "PlainText",
  displayName: "Short Description",
  // slug will be auto-generated as "short-description"
  validations: {
    maxLength: 250,
  },
});

// Example 9: Phone Field
// Best for contact numbers
const phoneField = new webflow.CollectionField("contact-phone", {
  collectionId: collectionId,
  type: "Phone",
  displayName: "Contact Phone",
  slug: "contact-phone",
  isRequired: false,
  helpText: "Contact phone number",
});

// Example 10: Color Field
// Best for theme colors, branding
const accentColorField = new webflow.CollectionField("accent-color", {
  collectionId: collectionId,
  type: "Color",
  displayName: "Accent Color",
  slug: "accent-color",
  isRequired: false,
  helpText: "Custom accent color for this article",
});

// Export field IDs and metadata for reference
export const deployedCollectionId = collectionId;
export const titleFieldId = titleField.fieldId;
export const contentFieldId = contentField.fieldId;
export const readTimeFieldId = readTimeField.fieldId;
export const publishDateFieldId = publishDateField.fieldId;
export const featuredFieldId = featuredField.fieldId;
export const authorEmailFieldId = authorEmailField.fieldId;
export const coverImageFieldId = coverImageField.fieldId;
export const shortDescriptionFieldId = shortDescriptionField.fieldId;
export const phoneFieldId = phoneField.fieldId;
export const accentColorFieldId = accentColorField.fieldId;

// Summary of created fields
const fieldSummary = pulumi.all([
  titleField.displayName,
  contentField.displayName,
  readTimeField.displayName,
  publishDateField.displayName,
  featuredField.displayName,
  authorEmailField.displayName,
  coverImageField.displayName,
  shortDescriptionField.displayName,
  phoneField.displayName,
  accentColorField.displayName,
]).apply((names) => {
  return `✅ Successfully created ${names.length} collection fields:\n${names.map((n, i) => `  ${i + 1}. ${n}`).join("\n")}`;
});

export const summary = fieldSummary;
