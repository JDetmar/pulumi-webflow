# PageCustomCode Resource Examples

This directory contains examples demonstrating how to apply registered custom JavaScript scripts to specific pages in your Webflow site using Pulumi.

## What You'll Learn

- Apply registered scripts to individual pages
- Override site-level scripts on specific pages
- Control script placement per page (header vs footer)
- Pass page-specific configuration via custom attributes
- Manage different scripts for different page types

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
pulumi config set landingPageId your-landing-page-id
pulumi config set productPageId your-product-page-id
pulumi up
```

## Finding Page IDs

Page IDs are 24-character hexadecimal strings. You can find them:

### Method 1: Webflow Designer
1. Open your site in the Webflow Designer
2. Open the Pages panel
3. Right-click a page → "Copy page ID"

### Method 2: Webflow Pages API
```bash
curl -X GET "https://api.webflow.com/v2/sites/{site_id}/pages" \
  -H "Authorization: Bearer YOUR_API_TOKEN"
```

### Method 3: Browser Developer Tools
1. Open your published site
2. View page source
3. Look for `data-wf-page="<page-id>"` in the `<html>` tag

## Examples Included

### 1. Landing Page Scripts

Apply conversion tracking and heatmap scripts to a landing page.

```typescript
const landingPageScripts = new webflow.PageCustomCode("landing-page-scripts", {
  pageId: landingPageId,
  scripts: [
    {
      id: conversionTrackingScript.id,
      version: "3.2.1",
      location: "header",
      attributes: {
        "data-campaign-id": "summer-2025",
        "data-conversion-type": "landing",
      },
    },
    {
      id: heatmapScript.id,
      version: "2.0.0",
      location: "footer",
      attributes: {
        "data-page-type": "landing",
        "data-track-clicks": "true",
      },
    },
  ],
});
```

### 2. Product Page with 360 Viewer

Add interactive product viewer to product pages.

```typescript
const productPageScripts = new webflow.PageCustomCode("product-page-scripts", {
  pageId: productPageId,
  scripts: [
    {
      id: productViewerScript.id,
      version: "1.5.0",
      location: "footer",
      attributes: {
        "data-viewer-container": "#product-viewer",
        "data-zoom-enabled": "true",
        "data-auto-rotate": "false",
      },
    },
  ],
});
```

### 3. Minimal Single-Script Configuration

Apply just one script to a page.

```typescript
const thankYouPageScripts = new webflow.PageCustomCode("thank-you-page-scripts", {
  pageId: thankYouPageId,
  scripts: [
    {
      id: conversionTrackingScript.id,
      version: "3.2.1",
      location: "header",
    },
  ],
});
```

## Prerequisites

Before using PageCustomCode, you must:

1. **Register your scripts** using the `RegisteredScript` resource
2. **Know the script ID** from the registered script
3. **Know the version** you want to deploy
4. **Have page IDs** for the pages you want to customize

See the [RegisteredScript example](../registeredscript/) for how to register scripts.

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `webflow:siteId`  | Yes      | Your Webflow site ID (stored as secret)  |
| `landingPageId`   | Yes*     | Page ID for landing page example        |
| `productPageId`   | Yes*     | Page ID for product page example        |
| `environment`     | No       | Deployment environment (default: development) |

*Required for the examples as written. Adjust based on your needs.

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId                  : [secret]
    landingPageScriptsCreatedOn     : "2025-01-06T12:34:56Z"
    landingPageScriptsLastUpdated   : "2025-01-06T12:34:56Z"
    productPageScriptsCreatedOn     : "2025-01-06T12:35:10Z"
    conversionTrackingScriptId      : "abc123..."
    productViewerScriptId           : "def456..."
    heatmapScriptId                 : "ghi789..."
    configuredLandingPageId         : "5f0c8c9e1c9d440000e8d8c4"
    configuredProductPageId         : "5f0c8c9e1c9d440000e8d8c5"
```

## Script Locations

### Header Location

Scripts with `location: "header"` are placed in the `<head>` section and execute before the page body loads.

**Best for:**
- Conversion tracking pixels
- Analytics scripts
- Meta tags and tracking codes
- Scripts that don't depend on DOM elements

### Footer Location

Scripts with `location: "footer"` are placed before the closing `</body>` tag and execute after page content loads.

**Best for:**
- Product viewers and interactive widgets
- Scripts that manipulate DOM elements
- Chat widgets specific to certain pages
- Scripts that can defer loading for better performance

## Custom Attributes

Use the `attributes` field to pass page-specific configuration:

```typescript
attributes: {
  "data-campaign-id": "summer-2025",
  "data-page-type": "landing",
  "data-viewer-container": "#product-viewer",
  "data-track-clicks": "true",
}
```

These become HTML attributes on the script tag:
```html
<script
  src="https://cdn.example.com/script.js"
  data-campaign-id="summer-2025"
  data-page-type="landing"
></script>
```

## Page vs Site Custom Code

### PageCustomCode (this resource)
- Applies to **specific pages** only
- Managed per-page
- Best for page-specific functionality
- Can override or supplement site-level scripts
- Example: Product viewer only on product pages

### SiteCustomCode
- Applies to **all pages** on the site
- Managed in one place
- Best for global scripts (analytics, chat, cookie consent)
- Example: Analytics on every page

See the [SiteCustomCode example](../sitecustomcode/) for site-wide scripts.

## Common Use Cases

### Landing Page Tracking
```typescript
// Track conversions specifically on landing pages
{
  id: conversionPixelScript.id,
  version: "1.0.0",
  location: "header",
  attributes: {
    "data-campaign": "fb-ads-summer",
  },
}
```

### Product Page Features
```typescript
// Add 360 product viewer only to product pages
{
  id: productViewerScript.id,
  version: "2.0.0",
  location: "footer",
  attributes: {
    "data-container": ".product-viewer",
  },
}
```

### Checkout Page Analytics
```typescript
// Enhanced tracking on checkout pages
{
  id: enhancedAnalyticsScript.id,
  version: "1.5.0",
  location: "header",
  attributes: {
    "data-track-cart": "true",
    "data-track-checkout-steps": "true",
  },
}
```

### Blog Post Features
```typescript
// Social sharing widgets on blog posts
{
  id: socialShareScript.id,
  version: "3.0.0",
  location: "footer",
}
```

## Managing Page Scripts

### Updating Script Version

To update to a new version:

1. Register the new version with RegisteredScript (if not already registered)
2. Update the `version` field in PageCustomCode
3. Run `pulumi up`

```typescript
scripts: [{
  id: myScript.id,
  version: "2.0.0",  // Changed from "1.0.0"
  location: "header",
}]
```

### Adding Scripts to a Page

Add to the `scripts` array and run `pulumi up`:

```typescript
scripts: [
  { /* existing script */ },
  {
    // New script
    id: newScript.id,
    version: "1.0.0",
    location: "footer",
  },
]
```

### Removing Scripts from a Page

Remove from the `scripts` array and run `pulumi up`:

```typescript
scripts: [
  { /* keep this script */ },
  // Removed script by deleting it from array
]
```

### Applying Scripts to Multiple Pages

Create separate PageCustomCode resources:

```typescript
const landingPage1 = new webflow.PageCustomCode("landing-1", {
  pageId: "page-id-1",
  scripts: [{ /* ... */ }],
});

const landingPage2 = new webflow.PageCustomCode("landing-2", {
  pageId: "page-id-2",
  scripts: [{ /* ... */ }],
});
```

Or use a loop for many pages:

```typescript
const pageIds = ["page-1", "page-2", "page-3"];
pageIds.forEach((pageId, index) => {
  new webflow.PageCustomCode(`page-${index}`, {
    pageId: pageId,
    scripts: [{ /* same scripts for each */ }],
  });
});
```

## Cleanup

To remove custom code from specific pages:

```bash
pulumi destroy
pulumi stack rm dev
```

**Note:** This only removes the custom code configuration from pages. The registered scripts remain available for future use.

## Troubleshooting

### "Invalid pageId" Error

Page IDs must be 24-character hexadecimal strings:
- Valid: `5f0c8c9e1c9d440000e8d8c4`
- Invalid: `my-page`, `page-123`, `5f0c8c9e` (too short)

Verify your page ID:
1. Copy it from Webflow Designer
2. Or get it from the Pages API
3. Ensure it's exactly 24 characters

### "Script not found" Error

The script must be registered first:
```typescript
// 1. Register the script
const myScript = new webflow.RegisteredScript("my-script", {
  siteId: siteId,
  displayName: "MyScript",
  hostedLocation: "https://cdn.example.com/script.js",
  integrityHash: "sha384-...",
  version: "1.0.0",
});

// 2. Then apply it to pages
const pageCode = new webflow.PageCustomCode("page-code", {
  pageId: pageId,
  scripts: [{
    id: myScript.id,  // Reference the registered script
    version: "1.0.0",
    location: "header",
  }],
});
```

### "Invalid location" Error

Location must be exactly `"header"` or `"footer"`:
- Valid: `"header"`, `"footer"`
- Invalid: `"head"`, `"body"`, `"Header"`, `"FOOTER"`

### "Version not found" Error

The version must match a registered version:
```typescript
// RegisteredScript has version "1.0.0"
const script = new webflow.RegisteredScript("script", {
  version: "1.0.0",  // Registered version
  // ...
});

// PageCustomCode must use the same version
scripts: [{
  id: script.id,
  version: "1.0.0",  // Must match exactly
  location: "header",
}]
```

### "At least one script is required" Error

The `scripts` array cannot be empty:
```typescript
// Invalid
scripts: []

// Valid
scripts: [{
  id: myScript.id,
  version: "1.0.0",
  location: "header",
}]
```

### Page ID Changes Force Replacement

Changing the `pageId` will delete and recreate the resource:
```typescript
// Changing pageId triggers replacement
pageId: "old-page-id"  →  pageId: "new-page-id"
```

If you need to move scripts from one page to another:
1. Create new PageCustomCode for the new page
2. Delete the old PageCustomCode resource

### Scripts Not Loading on Page

1. Verify the page has been published in Webflow
2. Check browser console for JavaScript errors
3. Ensure the registered script URL is accessible
4. Verify SRI hash matches the hosted file
5. Check that scripts aren't blocked by CSP
6. Clear browser cache and hard refresh

## Related Resources

- [RegisteredScript Example](../registeredscript/) - How to register scripts first
- [SiteCustomCode Example](../sitecustomcode/) - Apply scripts site-wide
- [Main Examples Index](../README.md)
