# Resource: webflow.RobotsTxt

Manage robots.txt configuration for a Webflow site. This resource allows you to define crawler access rules and sitemap references through infrastructure code.

## Overview

The `RobotsTxt` resource manages the robots.txt file for your Webflow site. The robots.txt file tells search engines and web crawlers which pages they can access and crawl on your site.

**Use this resource when you need to:**
- Configure crawler access rules (User-agent, Allow, Disallow directives)
- Add sitemap references for search engines
- Control SEO crawler behavior programmatically
- Manage robots.txt across multiple environments (dev, staging, prod)

**Complexity Level:** Simple - Good starting point for learning the provider

## Example Usage

### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

const config = new pulumi.Config();
const siteId = config.requireSecret("siteId");

// Create a robots.txt resource with standard configuration
const robotsTxt = new webflow.RobotsTxt("my-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`,
});

// Export the resource ID for reference
export const robotsTxtId = robotsTxt.id;
export const lastModified = robotsTxt.lastModified;
```

### Python

```python
import pulumi
import webflow_webflow as webflow

config = pulumi.Config()
site_id = config.require_secret("site_id")

# Create a robots.txt resource with standard configuration
robots_txt = webflow.RobotsTxt(
    "my-robots",
    site_id=site_id,
    content="""User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml""",
)

# Export the resource ID for reference
pulumi.export("robots_txt_id", robots_txt.id)
pulumi.export("last_modified", robots_txt.last_modified)
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

		// Create a robots.txt resource with standard configuration
		robotsTxt, err := webflow.NewRobotsTxt(ctx, "my-robots", &webflow.RobotsTxtArgs{
			SiteId: siteId,
			Content: pulumi.String(`User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`),
		})
		if err != nil {
			return err
		}

		// Export the resource ID for reference
		ctx.Export("robotsTxtId", robotsTxt.ID())
		ctx.Export("lastModified", robotsTxt.LastModified)
		return nil
	})
}
```

## Argument Reference

The following arguments are supported when creating a RobotsTxt resource:

### Inputs (Arguments)

| Property | Type   | Required | Description |
|----------|--------|----------|-------------|
| `siteId` | String | Yes | **The Webflow site ID** - 24-character lowercase hexadecimal string (e.g., `5f0c8c9e1c9d440000e8d8c3`). Found in Webflow project settings under "API & Webhooks" → "Site ID". **⚠️ Changing this value triggers resource replacement.** |
| `content` | String | Yes | **The robots.txt content** in traditional robots.txt format. Supports standard directives: `User-agent`, `Allow`, `Disallow`, and `Sitemap`. One directive per line. See examples below for proper format. |

### Example Site ID Format

Site IDs are 24 characters of lowercase hexadecimal:
```
5f0c8c9e1c9d440000e8d8c3
```

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

| Property | Type | Description |
|----------|------|-------------|
| `id` | String | The unique identifier for this robots.txt resource (typically the site ID) |
| `lastModified` | String | RFC3339 timestamp of when the robots.txt was last modified (e.g., `2025-12-30T12:34:56Z`) |

## Common Patterns

### Pattern 1: Allow All Crawlers

Standard configuration allowing all search engine crawlers:

```typescript
const robotsTxt = new webflow.RobotsTxt("allow-all", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`,
});
```

This tells all crawlers they can access your entire site and provides the sitemap location.

### Pattern 2: Block Specific Crawler

Block a specific crawler (e.g., Googlebot-Image) while allowing others:

```typescript
const robotsTxt = new webflow.RobotsTxt("selective-blocking", {
  siteId: siteId,
  content: `User-agent: Googlebot-Image
Disallow: /

User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`,
});
```

### Pattern 3: Restrict Directories

Block crawlers from accessing specific directories (e.g., staging, admin panels):

```typescript
const robotsTxt = new webflow.RobotsTxt("restrict-dirs", {
  siteId: siteId,
  content: `User-agent: *
Allow: /
Disallow: /admin
Disallow: /staging
Disallow: /private

Sitemap: https://example.com/sitemap.xml`,
});
```

### Pattern 4: Environment-Specific Configuration

Different robots.txt for different environments:

```typescript
const config = new pulumi.Config();
const environment = config.require("environment");

let robotsContent = `User-agent: *
Allow: /
Sitemap: https://example.com/sitemap.xml`;

if (environment === "staging") {
  robotsContent = `User-agent: *
Disallow: /
Crawl-delay: 10`;
}

const robotsTxt = new webflow.RobotsTxt("env-specific", {
  siteId: siteId,
  content: robotsContent,
});
```

## Robots.txt Format Guide

### Directives Reference

- **User-agent** - Specifies which crawler/bot this rule applies to
  - `User-agent: *` = all bots
  - `User-agent: Googlebot` = only Google's bot
  - `User-agent: Bingbot` = only Bing's bot

- **Allow** - Paths the crawler is allowed to access
  - `Allow: /` = allow access to entire site
  - `Allow: /public` = allow access to /public directory only

- **Disallow** - Paths the crawler cannot access
  - `Disallow: /admin` = block /admin directory
  - `Disallow: /private` = block /private directory
  - `Disallow: /` = block all access

- **Sitemap** - URL to your XML sitemap
  - `Sitemap: https://example.com/sitemap.xml`

### Example: Complete robots.txt

```
# Allow most crawlers
User-agent: *
Allow: /
Disallow: /admin
Disallow: /private
Disallow: /temp

# Specific rules for Bingbot
User-agent: Bingbot
Crawl-delay: 1

# Block bad crawlers
User-agent: BadBot
Disallow: /

# Sitemap reference
Sitemap: https://example.com/sitemap.xml
```

## Troubleshooting

### Error: "Invalid site ID"

**Cause:** Site ID is not in the correct format or doesn't exist

**Solution:**
- Verify site ID is 24 characters of lowercase hex: `5f0c8c9e1c9d440000e8d8c3`
- Check Webflow project settings → API & Webhooks for correct site ID
- Ensure the site exists in your Webflow account

### Error: "Invalid robots.txt content"

**Cause:** robots.txt format is invalid

**Solution:**
- Check syntax: one directive per line
- Use standard directives only: `User-agent`, `Allow`, `Disallow`, `Sitemap`, `Crawl-delay`
- No special characters or invalid formatting
- Refer to [robots.txt standard](https://www.robotstxt.org/) for syntax details

### Changes Rejected During Update

**Cause:** Trying to change the `siteId` - this triggers resource replacement

**Solution:**
- If you need a new site: delete this resource and create a new one with the new siteId
- Or use `pulumi destroy` followed by `pulumi up`

### robots.txt Not Updating in Webflow

**Cause:** Pulumi change was successful but Webflow hasn't reloaded

**Solution:**
- Clear browser cache
- Wait a few moments for propagation (usually instant)
- Verify the update succeeded: `pulumi stack output` to see lastModified timestamp
- Check Webflow project for the updated file

## Related Resources

- [Provider Configuration](./provider-configuration.md) - Set up API authentication
- [Redirect Resource](./redirect.md) - Manage URL redirects
- [Site Resource](./site.md) - Create and manage Webflow sites
- [Quickstart Guide](../README.md#quick-start) - Get started quickly
- [robots.txt Standard](https://www.robotstxt.org/) - Official robots.txt documentation
