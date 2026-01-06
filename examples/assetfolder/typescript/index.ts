import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * AssetFolder Example - Organizing Webflow Assets into Folders
 *
 * This example demonstrates how to create and organize asset folders in Webflow.
 * Asset folders help you organize your site's files (images, documents, etc.)
 * in the Webflow Assets panel.
 *
 * Key features:
 * - Create root-level folders
 * - Create nested folder structures (parent/child relationships)
 * - Organize multiple folders for different asset types
 *
 * NOTE: The Webflow API does not support deleting or updating asset folders.
 * Deleting this resource will only remove it from Pulumi state, not from Webflow.
 * Any changes to folder properties will require creating a new folder.
 */

// Example 1: Create root-level folders for basic organization
const imagesFolder = new webflow.AssetFolder("images-folder", {
  siteId: siteId,
  displayName: "Images",
});

const documentsFolder = new webflow.AssetFolder("documents-folder", {
  siteId: siteId,
  displayName: "Documents",
});

const iconsFolder = new webflow.AssetFolder("icons-folder", {
  siteId: siteId,
  displayName: "Icons",
});

// Example 2: Create nested folder structure (child folders)
// These folders will be organized under the Images folder
const heroImagesFolder = new webflow.AssetFolder("hero-images-subfolder", {
  siteId: siteId,
  displayName: "Hero Images",
  parentFolder: imagesFolder.folderId,
});

const thumbnailsFolder = new webflow.AssetFolder("thumbnails-subfolder", {
  siteId: siteId,
  displayName: "Thumbnails",
  parentFolder: imagesFolder.folderId,
});

const productPhotosFolder = new webflow.AssetFolder("product-photos-subfolder", {
  siteId: siteId,
  displayName: "Product Photos",
  parentFolder: imagesFolder.folderId,
});

// Example 3: Deeper nesting - create folders within subfolders
const mobileHeroFolder = new webflow.AssetFolder("mobile-hero-subfolder", {
  siteId: siteId,
  displayName: "Mobile",
  parentFolder: heroImagesFolder.folderId,
});

const desktopHeroFolder = new webflow.AssetFolder("desktop-hero-subfolder", {
  siteId: siteId,
  displayName: "Desktop",
  parentFolder: heroImagesFolder.folderId,
});

// Example 4: Bulk folder creation for multiple asset types
const assetTypesFolders: webflow.AssetFolder[] = [];
const assetTypes = [
  "Videos",
  "PDFs",
  "Fonts",
  "SVG Graphics",
];

assetTypes.forEach((typeName, index) => {
  const folder = new webflow.AssetFolder(`bulk-folder-${index}`, {
    siteId: siteId,
    displayName: typeName,
  });
  assetTypesFolders.push(folder);
});

// Export folder IDs for use in Asset resources
// These IDs can be used as the parentFolder when creating assets
export const deployedSiteId = siteId;
export const imagesFolderId = imagesFolder.folderId;
export const documentsFolderId = documentsFolder.folderId;
export const iconsFolderId = iconsFolder.folderId;

// Export nested folder IDs
export const heroImagesFolderId = heroImagesFolder.folderId;
export const thumbnailsFolderId = thumbnailsFolder.folderId;
export const productPhotosFolderId = productPhotosFolder.folderId;

// Export deeper nested folder IDs
export const mobileHeroFolderId = mobileHeroFolder.folderId;
export const desktopHeroFolderId = desktopHeroFolder.folderId;

// Export bulk created folder IDs
export const bulkFolderIds = assetTypesFolders.map((f) => f.folderId);

// Export folder details
export const imagesFolderDetails = {
  id: imagesFolder.folderId,
  name: imagesFolder.displayName,
  createdOn: imagesFolder.createdOn,
  lastUpdated: imagesFolder.lastUpdated,
};

// Print deployment success message
const totalFolders = 3 + 3 + 2 + assetTypes.length; // root + nested + deeper + bulk
const message = pulumi.interpolate`✅ Successfully created ${totalFolders} asset folders in site ${siteId}`;
message.apply((m) => console.log(m));
