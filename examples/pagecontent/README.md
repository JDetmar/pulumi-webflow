# PageContent Resource Example

This directory contains an example demonstrating how to manage static text content on Webflow pages using Pulumi.

## What You'll Learn

- Update static text content within existing DOM nodes on a Webflow page
- Manage hero section headings and subtitles programmatically
- Automate footer copyright updates with dynamic year
- Update multiple content sections in a single resource
- Understand the limitations of PageContent drift detection

## What This Example Does

The PageContent resource allows you to programmatically update text content within existing DOM nodes on a Webflow page. This is particularly useful for:

- **Automated Content Updates**: Keep copyright years, version numbers, or status messages current
- **Multi-Page Consistency**: Update similar content across multiple pages from code
- **Infrastructure-as-Code Content**: Manage content updates as part of your deployment pipeline
- **Dynamic Text Generation**: Inject computed values (like current year) into page content

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |

## Prerequisites

Before running this example, you need:

1. **Pulumi CLI** installed ([installation guide](https://www.pulumi.com/docs/get-started/install/))
2. **Webflow API token** set as `WEBFLOW_API_TOKEN` environment variable
3. **Page ID** from your Webflow site (24-character hex string)
4. **Node IDs** from your page's DOM structure

### Finding Node IDs

To find node IDs for your page:

```bash
# Using curl (replace with your page ID and API token)
curl -X GET "https://api.webflow.com/v2/pages/{page_id}/dom" \
  -H "Authorization: Bearer YOUR_API_TOKEN" \
  -H "accept: application/json"
```

This returns the page's DOM structure. Look for the `id` field on each node to identify which nodes you want to update.

Alternatively, you can use the Webflow Designer to inspect elements and find their IDs.

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set webflow:pageId your-page-id --secret
pulumi up
```

## Examples Included

### 1. Hero Section Content

Updates the main heading and subtitle in a hero section:

```typescript
nodes: [
  {
    nodeId: "hero-heading-node-id",
    text: "Welcome to Our Platform",
  },
  {
    nodeId: "hero-subtitle-node-id",
    text: "Build amazing experiences with our tools",
  },
]
```

### 2. Footer Copyright (with Dynamic Year)

Automatically updates copyright year using JavaScript:

```typescript
const currentYear = new Date().getFullYear();
nodes: [
  {
    nodeId: "footer-copyright-node-id",
    text: `© ${currentYear} Your Company Name. All rights reserved.`,
  },
]
```

### 3. Feature Section Content

Updates multiple feature titles and descriptions at once:

```typescript
nodes: [
  {
    nodeId: "feature-1-title-node-id",
    text: "Fast Performance",
  },
  {
    nodeId: "feature-1-description-node-id",
    text: "Lightning-fast load times for the best user experience.",
  },
  // ... more nodes
]
```

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `webflow:pageId`  | Yes      | Your Webflow page ID (24-char hex string) |
| `environment`     | No       | Deployment environment (default: development) |

### Setting Configuration

```bash
# Required: Set your page ID
pulumi config set webflow:pageId 5f0c8c9e1c9d440000e8d8c4 --secret

# Optional: Set environment
pulumi config set environment production
```

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedPageId      : [secret]
    heroContentId       : "pagecontent-abc123..."
    heroLastUpdated     : "2025-01-06T10:30:00Z"
    footerContentId     : "pagecontent-def456..."
    featureContentId    : "pagecontent-ghi789..."
```

## Important Limitations

### Drift Detection

**The PageContent resource has LIMITED drift detection:**

- ✅ **Detects**: When the page is deleted or becomes inaccessible
- ❌ **Does NOT detect**: Changes to text content made outside Pulumi

If content is modified via the Webflow UI or API, those changes will NOT be detected during `pulumi refresh` or `pulumi up`. The resource only verifies that the page itself still exists.

This limitation exists because extracting and comparing specific node text from the full DOM structure would require complex recursive traversal of the entire DOM tree, matching node IDs to current text values, and handling edge cases like deleted or moved nodes.

### What This Means

- Manual changes in Webflow won't trigger Pulumi updates
- `pulumi refresh` won't detect content drift
- Next `pulumi up` will revert manual changes to your declared state
- Consider documenting which content is managed by Pulumi to avoid confusion

## Cleanup

To remove the page content management (note: this doesn't delete the content from the page, just stops managing it):

```bash
pulumi destroy
pulumi stack rm dev
```

**Note:** Deleting a PageContent resource does NOT remove the text from your page. It only stops Pulumi from managing that content. The text remains on the page as-is.

## Troubleshooting

### "Invalid page ID" Error

1. Verify your page ID is a 24-character lowercase hexadecimal string
2. Check format: `5f0c8c9e1c9d440000e8d8c4` (no dashes, no uppercase)
3. Get the correct ID from Webflow Designer or Pages API

### "Node not found" Error

The specified node ID doesn't exist on the page:

1. Fetch the current page DOM: `GET /pages/{page_id}/dom`
2. Verify the node ID exists in the response
3. Ensure you're using the correct page ID
4. Check if the page layout has changed

### "Validation failed" Error

Common causes:
- Missing required fields (`pageId` or `nodes`)
- Empty `text` field in a node update
- Empty `nodes` array (at least one node required)
- Invalid node ID format

### Content Not Updating

If your content isn't changing:
1. Verify the node ID is correct
2. Check that the node contains editable text
3. Ensure the page is published in Webflow
4. Try fetching the page DOM to confirm structure

## Best Practices

1. **Document Your Node IDs**: Keep a mapping of node IDs to their purpose
2. **Use Meaningful Resource Names**: Name resources by content purpose (e.g., "hero-content", "footer-content")
3. **Group Related Nodes**: Update related content in the same PageContent resource
4. **Version Control**: Store node ID mappings in your repository
5. **Testing**: Test content updates in a staging site first
6. **Avoid Structure Changes**: Don't change page structure while managing content with this resource

## Related Resources

- [Page Resource Example](../page/)
- [Main Examples Index](../README.md)
- [Webflow Pages API Documentation](https://developers.webflow.com/reference/pages)

## Understanding the Resource Lifecycle

### Create
Creates the PageContent resource and applies text updates to the specified nodes.

### Update
Updates node content when you change the `text` values in your configuration. If you change the `pageId`, the resource will be replaced (delete + create).

### Delete
Removes the PageContent resource from Pulumi state. **Does NOT delete content from the page** - the text remains as-is.

### Read/Refresh
Verifies the page still exists. Does NOT detect content drift (see Limitations above).

## Real-World Use Cases

### Automated Copyright Year
```typescript
const year = new Date().getFullYear();
text: `© ${year} Company Name`
```

### Build Version Display
```typescript
const version = process.env.BUILD_VERSION || "1.0.0";
text: `Version ${version}`
```

### Environment-Specific Messaging
```typescript
const env = pulumi.getStack();
text: env === "production" ? "Live Site" : "Test Environment"
```

### Status Page Updates
```typescript
text: "All systems operational as of " + new Date().toISOString()
```
