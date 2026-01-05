import * as pulumi from "@pulumi/pulumi";
import * as webflow from "@jdetmar/pulumi-webflow";

// Create a Pulumi config object
const config = new pulumi.Config();

// Get configuration values
const siteId = config.requireSecret("siteId");

/**
 * Asset Example - Creating and Managing Webflow Assets
 *
 * This example demonstrates how to register assets with Webflow.
 * After registration, you'll receive uploadUrl and uploadDetails
 * to complete the file upload to S3.
 *
 * Two-step process:
 * 1. Register asset metadata (this provider handles this)
 * 2. Upload file to S3 using uploadUrl + uploadDetails (done separately)
 */

// Example 1: Register a single asset
// The fileHash is the MD5 hash of your file content
// Generate with: md5sum logo.png (Linux) or md5 -q logo.png (macOS)
const logoAsset = new webflow.Asset("company-logo", {
  siteId: siteId,
  fileName: "logo.png",
  fileHash: "d41d8cd98f00b204e9800998ecf8427e",
});

// Example 2: Asset with folder organization
const heroAsset = new webflow.Asset("hero-image", {
  siteId: siteId,
  fileName: "hero-banner.jpg",
  fileHash: "a1b2c3d4e5f6789012345678abcdef12",
  // parentFolder: "folder-id-here", // Uncomment to organize in a folder
});

// Example 3: Bulk asset registration
const iconAssets: webflow.Asset[] = [];
const icons = [
  { name: "icon-home", fileName: "home.svg", fileHash: "11111111111111111111111111111111" },
  { name: "icon-settings", fileName: "settings.svg", fileHash: "22222222222222222222222222222222" },
  { name: "icon-user", fileName: "user.svg", fileHash: "33333333333333333333333333333333" },
];

icons.forEach((icon) => {
  const asset = new webflow.Asset(icon.name, {
    siteId: siteId,
    fileName: icon.fileName,
    fileHash: icon.fileHash,
  });
  iconAssets.push(asset);
});

// Export values for the logo asset
// These are needed to complete the S3 upload
export const logoAssetId = logoAsset.assetId;
export const logoUploadUrl = logoAsset.uploadUrl;
export const logoUploadDetails = logoAsset.uploadDetails;
export const logoAssetUrl = logoAsset.assetUrl;
export const logoHostedUrl = logoAsset.hostedUrl;

// Export hero asset info
export const heroAssetId = heroAsset.assetId;
export const heroHostedUrl = heroAsset.hostedUrl;

// Export icon asset IDs
export const iconAssetIds = iconAssets.map((a) => a.assetId);

// Print deployment message
const message = pulumi.interpolate`Registered ${icons.length + 2} assets. Use uploadUrl and uploadDetails to complete S3 uploads.`;
message.apply((m) => console.log(m));
