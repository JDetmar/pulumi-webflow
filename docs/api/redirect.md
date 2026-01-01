# Resource: webflow.Redirect

Create and manage HTTP redirects for a Webflow site. This resource allows you to programmatically create 301 (permanent) and 302 (temporary) redirects.

## Overview

The `Redirect` resource manages URL redirects for your Webflow site. Redirects are essential for SEO and user experience when managing URL changes.

**Use this resource when you need to:**
- Create permanent redirects (301) for SEO-friendly URL changes
- Create temporary redirects (302) for testing or temporary paths
- Consolidate multiple domains into one primary domain
- Handle legacy URLs after site restructuring
- Manage redirect rules across multiple environments
- Bulk create redirects programmatically

**Complexity Level:** Medium - Requires understanding HTTP status codes

## Example Usage

### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

const config = new pulumi.Config();
const siteId = config.requireSecret("siteId");

// Create a permanent redirect (301)
const oldToNewRedirect = new webflow.Redirect("old-to-new", {
  siteId: siteId,
  sourcePath: "/old-page",
  destinationPath: "/new-page",
  statusCode: 301,
});

// Export the redirect ID
export const redirectId = oldToNewRedirect.id;
```

### Python

```python
import pulumi
import webflow_webflow as webflow

config = pulumi.Config()
site_id = config.require_secret("site_id")

# Create a permanent redirect (301)
old_to_new = webflow.Redirect(
    "old-to-new",
    site_id=site_id,
    source_path="/old-page",
    destination_path="/new-page",
    status_code=301,
)

# Export the redirect ID
pulumi.export("redirect_id", old_to_new.id)
```

### Go

```go
package main

import (
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		siteId := cfg.RequireSecret("siteId")

		// Create a permanent redirect (301)
		oldToNew, err := webflow.NewRedirect(ctx, "old-to-new", &webflow.RedirectArgs{
			SiteId:          siteId,
			SourcePath:      pulumi.String("/old-page"),
			DestinationPath: pulumi.String("/new-page"),
			StatusCode:      pulumi.Int(301),
		})
		if err != nil {
			return err
		}

		// Export the redirect ID
		ctx.Export("redirectId", oldToNew.ID())
		return nil
	})
}
```

## Argument Reference

The following arguments are supported when creating a Redirect resource:

### Inputs (Arguments)

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `siteId` | String | Yes | **The Webflow site ID** - 24-character lowercase hexadecimal string (e.g., `5f0c8c9e1c9d440000e8d8c3`). Found in Webflow project settings. **⚠️ Changing this value triggers resource replacement.** |
| `sourcePath` | String | Yes | **Source URL path** - The old URL path to redirect from (e.g., `/old-page`). Should start with `/`. **⚠️ Changing this value triggers resource replacement.** |
| `destinationPath` | String | Yes | **Destination URL path** - The new URL path to redirect to (e.g., `/new-page`). Can be an absolute URL for external redirects (e.g., `https://example.com/page`). |
| `statusCode` | Integer | Yes | **HTTP status code** - Must be either `301` (permanent) or `302` (temporary). **301 is recommended for SEO-friendly URL changes.** |

### Status Code Reference

| Code | Type | Use Case |
|------|------|----------|
| **301** | Permanent | SEO-friendly redirect. Use when permanently moving a page. Signals to search engines that the old URL should be replaced with the new one. |
| **302** | Temporary | Temporary redirect. Use for testing or when you might change the redirect later. Tells search engines to keep indexing the original URL. |

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

| Property | Type | Description |
|----------|------|-------------|
| `id` | String | The unique identifier for this redirect resource |
| `createdOn` | String | RFC3339 timestamp when the redirect was created (read-only). Automatically set when the redirect is created. |

## Common Patterns

### Pattern 1: Permanent Page Redirect (301)

Redirect an old page to its new location (SEO-friendly):

```typescript
const pageRedirect = new webflow.Redirect("about-us-redirect", {
  siteId: siteId,
  sourcePath: "/about",
  destinationPath: "/company/about-us",
  statusCode: 301,
});
```

**When to use:** You've permanently moved a page and want search engines to update their index.

### Pattern 2: Temporary Redirect (302)

Temporary redirect while you're testing or making temporary changes:

```typescript
const testRedirect = new webflow.Redirect("test-redirect", {
  siteId: siteId,
  sourcePath: "/beta-feature",
  destinationPath: "/feature-v2",
  statusCode: 302,
});
```

**When to use:** The redirect is temporary, or you're unsure if it will be permanent.

### Pattern 3: External Domain Redirect

Redirect to a completely different domain (consolidating multiple domains):

```typescript
const externalRedirect = new webflow.Redirect("external-redirect", {
  siteId: siteId,
  sourcePath: "/",
  destinationPath: "https://newdomain.com",
  statusCode: 301,
});
```

**When to use:** Consolidating multiple domains into a single primary domain for SEO.

### Pattern 4: Bulk Redirects

Create multiple redirects programmatically from a list:

```typescript
const oldToNewMapping = [
  { old: "/blog/post-1", new: "/articles/topic-1" },
  { old: "/blog/post-2", new: "/articles/topic-2" },
  { old: "/services", new: "/solutions" },
  { old: "/team", new: "/company/team" },
];

const redirects = oldToNewMapping.map((mapping, index) =>
  new webflow.Redirect(`redirect-${index}`, {
    siteId: siteId,
    sourcePath: mapping.old,
    destinationPath: mapping.new,
    statusCode: 301,
  })
);

// Export all redirect IDs
export const redirectIds = redirects.map((r) => r.id);
```

**When to use:** Migrating many URLs after a site restructuring or reorganization.

### Pattern 5: Environment-Specific Redirects

Different redirect targets for different environments:

```typescript
const config = new pulumi.Config();
const environment = config.require("environment");

let destinationPath = "/feature";
if (environment === "staging") {
  destinationPath = "/feature-beta";
} else if (environment === "production") {
  destinationPath = "https://main-site.com/feature";
}

const envRedirect = new webflow.Redirect("env-redirect", {
  siteId: siteId,
  sourcePath: "/new-feature",
  destinationPath: destinationPath,
  statusCode: 301,
});
```

**When to use:** Different redirect targets for dev, staging, and production environments.

## Path Format Guide

### Source Path Format

The source path is what visitors access (the old URL):

```
✅ /old-page          # Simple path
✅ /blog/old-article  # Nested path
✅ /                  # Root redirect
❌ https://example.com/old-page  # Don't include domain
```

### Destination Path Format

The destination can be a relative path (same domain) or absolute URL (different domain):

```
✅ /new-page                    # Relative path (same domain)
✅ /blog/new-article            # Nested relative path
✅ https://newdomain.com/page   # Absolute URL (external domain)
✅ https://newdomain.com        # External domain root
❌ newdomain.com/page           # Must include https://
```

## Troubleshooting

### Error: "Invalid status code"

**Cause:** Status code is not 301 or 302

**Solution:**
- Use `301` for permanent redirects
- Use `302` for temporary redirects
- Check syntax: must be an integer (301 not "301")

### Error: "Invalid source path"

**Cause:** Source path format is incorrect

**Solution:**
- Ensure path starts with `/` (e.g., `/old-page` not `old-page`)
- Don't include domain (e.g., `/page` not `https://example.com/page`)
- Use URL-safe characters only

### Error: "Invalid destination path"

**Cause:** Destination path format is incorrect for the target type

**Solution:**
- For same domain: start with `/` (e.g., `/new-page`)
- For external: use full URL (e.g., `https://example.com/page`)
- No spaces or invalid characters
- For external URLs, must include `https://`

### Redirect Not Working

**Cause:** Webflow hasn't reloaded or redirect is being cached

**Solution:**
- Wait a few moments for propagation (usually instant)
- Clear browser cache and cookies
- Test in private/incognito browser window to bypass cache
- Verify redirect in Webflow admin panel
- Check that source and destination paths are correct

### Conflict with Existing Redirects

**Cause:** A redirect already exists for the source path

**Solution:**
- Update or delete the existing redirect first
- Or modify the source path to be unique
- Check Webflow admin for existing redirects

## See Also

**Troubleshooting:**
- [Troubleshooting Guide](../troubleshooting.md) - Comprehensive error reference and solutions
- [FAQ](../faq.md) - Frequently asked questions
- Look for "Redirect" section in troubleshooting guide for resource-specific issues

**Related Resources:**
- [Provider Configuration](./provider-configuration.md) - Set up API authentication
- [RobotsTxt Resource](./robotstxt.md) - Manage robots.txt
- [Site Resource](./site.md) - Create and manage Webflow sites
- [Quickstart Guide](../README.md#quick-start) - Get started quickly
- [HTTP Status Codes](https://developer.mozilla.org/en-US/docs/Web/HTTP/Status) - Learn more about status codes
- [SEO Best Practices](https://support.google.com/webmasters/answer/7440203) - Google's redirect guidance
