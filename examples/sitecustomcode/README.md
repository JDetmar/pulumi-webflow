# SiteCustomCode Resource Examples

This directory contains examples demonstrating how to apply registered custom JavaScript scripts to an entire Webflow site using Pulumi.

## What You'll Learn

- Apply registered scripts to all pages of a Webflow site
- Control script placement (header vs footer)
- Pass custom attributes to scripts for configuration
- Manage multiple site-wide scripts together

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

### 1. Site-Wide Analytics and Tracking

Apply analytics scripts to track all pages automatically.

```typescript
const siteScripts = new webflow.SiteCustomCode("site-wide-scripts", {
  siteId: siteId,
  scripts: [
    {
      id: analyticsScript.id,
      version: "4.0.0",
      location: "header",
      attributes: {
        "data-site-id": "GA-123456789",
      },
    },
  ],
});
```

### 2. Multiple Scripts with Different Placements

Combine header and footer scripts for optimal performance.

```typescript
scripts: [
  {
    // Analytics in header - loads before page renders
    id: analyticsScript.id,
    version: "4.0.0",
    location: "header",
  },
  {
    // Chat widget in footer - loads after content
    id: chatWidgetScript.id,
    version: "2.5.0",
    location: "footer",
    attributes: {
      "data-widget-id": "chat-123",
    },
  },
]
```

### 3. Cookie Consent and Compliance

Apply cookie consent banners and GDPR compliance scripts.

```typescript
{
  id: cookieConsentScript.id,
  version: "1.0.0",
  location: "header",
  attributes: {
    "data-theme": "dark",
    "data-position": "bottom-right",
  },
}
```

## Prerequisites

Before using SiteCustomCode, you must:

1. **Register your scripts** using the `RegisteredScript` resource
2. **Know the script ID** from the registered script
3. **Know the version** you want to deploy

See the [RegisteredScript example](../registeredscript/) for how to register scripts.

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
    deployedSiteId          : [secret]
    siteScriptsCreatedOn    : "2025-01-06T12:34:56Z"
    siteScriptsLastUpdated  : "2025-01-06T12:34:56Z"
    appliedScriptCount      : 3
    analyticsScriptId       : "abc123..."
    chatWidgetScriptId      : "def456..."
    cookieConsentScriptId   : "ghi789..."
```

## Script Locations

### Header Location

Scripts with `location: "header"` are placed in the `<head>` section and execute before the page body loads.

**Best for:**
- Analytics and tracking codes
- Meta pixel scripts
- Cookie consent banners
- Scripts that don't depend on DOM elements

### Footer Location

Scripts with `location: "footer"` are placed before the closing `</body>` tag and execute after page content loads.

**Best for:**
- Chat widgets
- Interactive components
- Scripts that manipulate DOM elements
- Scripts that can defer loading for better performance

## Custom Attributes

Use the `attributes` field to pass configuration to your scripts:

```typescript
attributes: {
  "data-site-id": "GA-123456789",
  "data-theme": "dark",
  "data-auto-open": "false",
  "async": "true",
}
```

These become HTML attributes on the script tag:
```html
<script
  src="https://cdn.example.com/script.js"
  data-site-id="GA-123456789"
  data-theme="dark"
  data-auto-open="false"
  async="true"
></script>
```

## Managing Script Updates

### Updating a Script Version

To update to a new version of a script:

1. Register the new version with RegisteredScript
2. Update the `version` field in SiteCustomCode
3. Run `pulumi up`

```typescript
// Change from version "1.0.0" to "2.0.0"
{
  id: analyticsScript.id,
  version: "2.0.0",  // Updated
  location: "header",
}
```

### Adding New Scripts

Add scripts to the `scripts` array and run `pulumi up`:

```typescript
scripts: [
  { /* existing script */ },
  { /* existing script */ },
  {
    // New script
    id: newScript.id,
    version: "1.0.0",
    location: "footer",
  },
]
```

### Removing Scripts

Remove scripts from the `scripts` array and run `pulumi up`:

```typescript
// Remove the second script by deleting it from the array
scripts: [
  { /* keep this script */ },
  // Script removed
  { /* keep this script */ },
]
```

## Site vs Page Custom Code

### SiteCustomCode (this resource)
- Applies to **all pages** on the site
- Managed in one place
- Best for global scripts (analytics, chat, cookie consent)

### PageCustomCode
- Applies to **specific pages** only
- Managed per-page
- Best for page-specific functionality
- Can override site-level scripts

See the [PageCustomCode example](../pagecustomcode/) for page-specific scripts.

## Cleanup

To remove all site-wide custom code:

```bash
pulumi destroy
pulumi stack rm dev
```

**Note:** This only removes the custom code configuration. The registered scripts remain available for future use.

## Troubleshooting

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

// 2. Then apply it to the site
const siteCode = new webflow.SiteCustomCode("site-code", {
  siteId: siteId,
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

The version must match a registered version exactly:
```typescript
// RegisteredScript has version "1.0.0"
const script = new webflow.RegisteredScript("script", {
  version: "1.0.0",  // Registered version
  // ...
});

// SiteCustomCode must use the same version
scripts: [{
  id: script.id,
  version: "1.0.0",  // Must match
  location: "header",
}]
```

### Scripts Not Loading on Pages

1. Check browser console for errors
2. Verify the registered script URL is accessible
3. Check SRI hash matches the hosted file
4. Ensure scripts aren't blocked by Content Security Policy (CSP)
5. Verify the site has been published in Webflow

## Related Resources

- [RegisteredScript Example](../registeredscript/) - How to register scripts first
- [PageCustomCode Example](../pagecustomcode/) - Apply scripts to specific pages
- [Main Examples Index](../README.md)
