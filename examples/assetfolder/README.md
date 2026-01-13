# AssetFolder Resource Examples

This directory contains examples demonstrating how to create and organize asset folders for Webflow sites using Pulumi.

## What You'll Learn

- Create root-level folders for organizing assets
- Build nested folder structures (parent/child relationships)
- Organize multiple folders for different asset types
- Use folder IDs when uploading assets

## Important: Folder Lifecycle Limitations

The Webflow API has specific limitations for asset folders:

- **No Update Support**: Folders cannot be renamed or modified after creation
- **No Delete Support**: Folders cannot be deleted via the API
- **State Management**: Deleting this resource only removes it from Pulumi state, not from Webflow
- **Changes Require Replacement**: Any property changes will create a new folder (the old one remains in Webflow)

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set webflow:siteId your-site-id --secret
pulumi up
```

## Examples Included

### 1. Root-Level Folders

Create top-level folders for basic asset organization:

```typescript
const imagesFolder = new webflow.AssetFolder("images-folder", {
  siteId: siteId,
  displayName: "Images",
});
```

### 2. Nested Folder Structures

Build parent/child folder hierarchies:

```typescript
const heroImagesFolder = new webflow.AssetFolder("hero-images", {
  siteId: siteId,
  displayName: "Hero Images",
  parentFolder: imagesFolder.folderId, // Child of Images folder
});
```

### 3. Deep Nesting

Create multi-level folder hierarchies:

```typescript
const mobileHeroFolder = new webflow.AssetFolder("mobile-hero", {
  siteId: siteId,
  displayName: "Mobile",
  parentFolder: heroImagesFolder.folderId, // Grandchild folder
});
```

### 4. Bulk Folder Creation

Efficiently create multiple folders for different asset categories:

```typescript
const assetTypes = ["Videos", "PDFs", "Fonts", "SVG Graphics"];

assetTypes.forEach((typeName) => {
  new webflow.AssetFolder(typeName, {
    siteId: siteId,
    displayName: typeName,
  });
});
```

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `webflow:siteId`  | Yes      | Your Webflow site ID (stored as secret)  |
| `environment`     | No       | Deployment environment (default: development) |

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId           : [secret]
    imagesFolderId          : "abc123..."
    documentsFolderId       : "def456..."
    iconsFolderId           : "ghi789..."
    heroImagesFolderId      : "jkl012..."
    thumbnailsFolderId      : "mno345..."
    productPhotosFolderId   : "pqr678..."
    bulkFolderIds           : ["stu...", "vwx...", "yz1...", "234..."]
```

## Using Folder IDs with Assets

After creating folders, use their IDs when uploading assets:

```typescript
// Create folder
const imagesFolder = new webflow.AssetFolder("images", {
  siteId: siteId,
  displayName: "Images",
});

// Upload asset to folder
const logoAsset = new webflow.Asset("company-logo", {
  siteId: siteId,
  fileName: "logo.png",
  fileHash: "d41d8cd98f00b204e9800998ecf8427e",
  parentFolder: imagesFolder.folderId, // Place in Images folder
});
```

## Cleanup

To remove folders from Pulumi state (they will remain in Webflow):

```bash
pulumi destroy
pulumi stack rm dev
```

**Important**: The folders will remain in your Webflow Assets panel after running `pulumi destroy`. To remove them completely, manually delete them from the Webflow dashboard.

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings → General
2. Ensure correct format: `abc123def456` (24-character lowercase hexadecimal)
3. Check your API token has access to the site

### "Invalid parent folder ID" Error

1. Ensure the parent folder exists and has been created successfully
2. Verify the parent folder ID format is correct
3. Check that you're not creating circular dependencies (folder as its own parent)

### Folders Not Appearing in Webflow

1. Refresh the Webflow Assets panel
2. Check that the folder creation succeeded in Pulumi outputs
3. Verify you're looking at the correct site in Webflow

### "Cannot update folder" Error

Asset folders cannot be updated via the API. If you need to:
- **Rename a folder**: Create a new folder with the new name (old folder remains)
- **Change parent folder**: Create a new folder with the new parent (old folder remains)
- **Delete a folder**: Manually delete from Webflow dashboard (API doesn't support deletion)

## Folder Organization Best Practices

### Recommended Structure

```
Root Level
├── Images
│   ├── Hero Images
│   │   ├── Mobile
│   │   └── Desktop
│   ├── Thumbnails
│   └── Product Photos
├── Documents
│   ├── PDFs
│   └── Spreadsheets
├── Icons
│   ├── Navigation
│   └── Social Media
└── Videos
    ├── Hero Videos
    └── Product Demos
```

### Tips

- Keep folder names descriptive and consistent
- Use 2-3 levels of nesting maximum for easier navigation
- Group assets by function/purpose rather than file type when possible
- Plan your folder structure before creating folders (they can't be renamed)

## Related Resources

- [Asset Resource Examples](../asset/)
- [Main Examples Index](../README.md)
- [Webflow Assets Documentation](https://university.webflow.com/lesson/assets-panel)
- [Webflow Asset API](https://developers.webflow.com/reference/assets)
