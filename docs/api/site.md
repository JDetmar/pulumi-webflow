# Resource: webflow.Site

Create, configure, publish, and manage complete Webflow sites through infrastructure code. This is the most powerful resource in the provider, enabling full site lifecycle management.

## Overview

The `Site` resource represents a Webflow site and enables complete lifecycle management: creation, configuration, publishing, and deletion.

**Use this resource when you need to:**
- Create new Webflow sites programmatically
- Configure site properties (name, domain, timezone)
- Publish sites through infrastructure code
- Manage site lifecycle in CI/CD pipelines
- Deploy sites to multiple environments
- Import existing sites into managed infrastructure
- Delete sites programmatically

**Complexity Level:** Advanced - Requires understanding Webflow site concepts and complete lifecycle management

## Example Usage

### TypeScript - Basic Site Creation

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");
const environment = config.require("environment");

// Create a new Webflow site
const site = new webflow.Site("my-site", {
  workspaceId: workspaceId,
  displayName: `My Site - ${environment}`,
  shortName: `my-site-${environment}`,
  timeZone: "America/New_York",
});

// Export site information
export const siteId = site.id;
export const displayName = site.displayName;
```

### Python - Multi-Environment Deployment

```python
import pulumi
import webflow_webflow as webflow

config = pulumi.Config()
workspace_id = config.require_secret("workspace_id")
environment = config.require("environment")

# Create sites for different environments
site = webflow.Site(
    f"my-site-{environment}",
    workspace_id=workspace_id,
    display_name=f"My Site - {environment}",
    short_name=f"my-site-{environment}",
    time_zone="America/New_York",
)

# Export site information
pulumi.export("site_id", site.id)
pulumi.export("display_name", site.display_name)
```

### Go - Production Site Creation

```go
package main

import (
	"fmt"

	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi/config"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		cfg := config.New(ctx, "")
		workspaceId := cfg.RequireSecret("workspaceId")
		environment := cfg.Require("environment")

		// Create a new Webflow site
		site, err := webflow.NewSite(ctx, "my-site", &webflow.SiteArgs{
			WorkspaceId: workspaceId,
			DisplayName: pulumi.String(fmt.Sprintf("My Site - %s", environment)),
			ShortName:   pulumi.String(fmt.Sprintf("my-site-%s", environment)),
			TimeZone:    pulumi.String("America/New_York"),
		})
		if err != nil {
			return err
		}

		// Export site information
		ctx.Export("siteId", site.ID())
		ctx.Export("displayName", site.DisplayName)
		return nil
	})
}
```

## Argument Reference

The following arguments are supported when creating a Site resource:

### Inputs (Arguments)

| Property | Type | Required | Description |
|----------|------|----------|-------------|
| `workspaceId` | String | Yes | **The Webflow workspace ID** - 24-character lowercase hexadecimal string where the site will be created. Required for site creation (Enterprise workspace required by Webflow API). Found in Webflow dashboard under Account Settings > Workspace. **Changing this triggers resource replacement.** |
| `displayName` | String | Yes | **Friendly site name** - The name displayed in Webflow dashboard (e.g., "My Company Website"). Can include spaces and special characters. This is for humans, not machines. **Changing this triggers in-place update.** |
| `shortName` | String | Optional | **Machine-readable site identifier** - Used in Webflow URLs and as site slug. If not provided, Webflow will automatically generate one from displayName. Must be lowercase, contain only alphanumeric characters and hyphens (e.g., `my-site-prod`). Cannot contain spaces. **Changing this triggers in-place update.** |
| `timeZone` | String | Optional | **Time zone for the site** - Affects scheduled publishing and timestamp display. Use IANA timezone format (e.g., `America/New_York`, `Europe/London`, `Asia/Tokyo`). Full list available on [Wikipedia IANA Timezone List](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones). **Changing this triggers in-place update.** |
| `parentFolderId` | String | Optional | **Folder ID for organization** - The folder ID where the site will be organized in the Webflow dashboard. If not specified, the site will be placed at the workspace root. |
| `publish` | Boolean | Optional | **Auto-publish after changes** - When set to true, the provider will publish the site to production after successfully creating or updating it. Default: false (manual publishing required). Note: Site must have at least one published version before automatic publishing will work. |
| `templateName` | String | Optional | **Template for site creation** - The template to use when creating the site. If not specified, Webflow will create a blank site. **WARNING: This field is IMMUTABLE. Changing this value will DELETE the existing site and CREATE a new one (DESTRUCTIVE).** |

### Constraints and Validation

- **workspaceId:** 24-character lowercase hexadecimal string (e.g., `5f0c8c9e1c9d440000e8d8c3`)
- **displayName:** 1-255 characters, can include spaces and symbols
- **shortName:** 3-50 characters, lowercase letters, numbers, hyphens only, must start with letter (optional - auto-generated if not provided)
- **timeZone:** Must be valid IANA timezone (e.g., `America/New_York`, not `EST`) - optional

## Attribute Reference

In addition to all arguments above, the following attributes are exported:

| Property | Type | Description |
|----------|------|-------------|
| `id` | String | The unique Webflow site ID (24-character hex string) - use this for other resources like RobotsTxt and Redirect |
| `lastPublished` | String | RFC3339 timestamp of the last time the site was published (read-only). Automatically set by Webflow when changes are published to production. |
| `lastUpdated` | String | RFC3339 timestamp of the last time the site configuration was updated (read-only). |
| `previewUrl` | String | URL to a preview image of the site. Automatically generated by Webflow for site thumbnails. |
| `customDomains` | Array of String | List of custom domains attached to the site (read-only in this release). Custom domain management will be available in a future release. |
| `dataCollectionEnabled` | Boolean | Indicates whether data collection is enabled for this site (read-only). Controlled by your Webflow workspace settings. |
| `dataCollectionType` | String | The type of data collection enabled. Possible values: `always`, `optOut`, `disabled`. Controlled by workspace settings. |

## Common Patterns

### Pattern 1: Create Production Site

Create a production site with proper configuration:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const productionSite = new webflow.Site("prod-site", {
  workspaceId: workspaceId,
  displayName: "My Company - Production",
  shortName: "company-prod",
  timeZone: "America/Los_Angeles",
});

export const productionSiteId = productionSite.id;
export const productionDomains = productionSite.customDomains;
```

### Pattern 2: Multi-Environment Deployment

Deploy the same site structure to dev, staging, and production:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const environments = ["dev", "staging", "production"];
const sites = environments.map((env) =>
  new webflow.Site(`site-${env}`, {
    workspaceId: workspaceId,
    displayName: `Company Site - ${env.toUpperCase()}`,
    shortName: `company-${env}`,
    timeZone: "America/New_York",
  })
);

// Export all site IDs
export const siteIds = sites.map((s) => s.id);
```

### Pattern 3: Site with Timezone Configuration

Set appropriate timezone for international sites:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const usaSite = new webflow.Site("us-site", {
  workspaceId: workspaceId,
  displayName: "US Operations",
  shortName: "us-ops",
  timeZone: "America/Chicago",
});

const euroSite = new webflow.Site("euro-site", {
  workspaceId: workspaceId,
  displayName: "European Operations",
  shortName: "euro-ops",
  timeZone: "Europe/London",
});
```

### Pattern 4: Environment-Specific Configuration

Use environment config to customize site properties:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");
const environment = config.require("environment");

const site = new webflow.Site("managed-site", {
  workspaceId: workspaceId,
  displayName: `Managed Site - ${environment}`,
  shortName: `managed-${environment}`,
  timeZone: "UTC",
});

// Export site info for downstream resources
export const siteId = site.id;
```

### Pattern 5: Site with Descriptive Naming

Use consistent naming conventions across multiple sites:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const siteConfig = {
  workspaceId: workspaceId,
  displayName: "E-commerce Platform",
  shortName: "ecommerce-platform",
  timeZone: "America/New_York",
};

const site = new webflow.Site("ecommerce", siteConfig);
```

## Timezone Configuration

### Common Timezone Examples

| Region | Timezone | Example |
|--------|----------|---------|
| **US** | `America/New_York` | Eastern Time |
| **US** | `America/Chicago` | Central Time |
| **US** | `America/Denver` | Mountain Time |
| **US** | `America/Los_Angeles` | Pacific Time |
| **UK** | `Europe/London` | GMT/BST |
| **Europe** | `Europe/Paris` | CET/CEST |
| **Europe** | `Europe/Berlin` | CET/CEST |
| **India** | `Asia/Kolkata` | IST |
| **China** | `Asia/Shanghai` | CST |
| **Japan** | `Asia/Tokyo` | JST |
| **Australia** | `Australia/Sydney` | AEDT/AEST |
| **UTC** | `UTC` | Coordinated Universal Time |

### Finding Timezone Values

Use the [IANA Timezone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) to find the correct timezone string for your location.

## Site Lifecycle Operations

### Creating a Site

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const newSite = new webflow.Site("new", {
  workspaceId: workspaceId,
  displayName: "New Site",
  shortName: "newsite",
  timeZone: "UTC",
});
```

Webflow creates the site and returns a 24-character site ID for use in other resources.

### Publishing a Site

To publish a site (make it live), use the Site resource with publish operations:

```typescript
// Publishing would typically be done through a separate workflow
// after the site is created and content is added
```

### Updating Site Properties

Modify site properties in-place:

```typescript
const config = new pulumi.Config();
const workspaceId = config.requireSecret("workspaceId");

const site = new webflow.Site("updatable", {
  workspaceId: workspaceId,
  displayName: "Original Name",
  shortName: "original",
  timeZone: "America/New_York",
});

// To update, modify properties and run `pulumi up`
// displayName, shortName, and timeZone can be updated in-place
```

### Importing Existing Sites

Import a site that already exists in Webflow:

```bash
# Import existing site into Pulumi state
pulumi import webflow:index:Site my-site <existing-site-id>
```

### Deleting a Site

Remove a site from Webflow:

```bash
# Remove the Site resource from your code and run:
pulumi up
# This will delete the site from Webflow
```

## Troubleshooting

### Error: "Invalid short name"

**Cause:** Short name doesn't meet Webflow's requirements

**Solutions:**
- Must be lowercase: `my-site` not `My-Site`
- Only alphanumeric and hyphens: `my-site-2` not `my_site-2`
- Must start with letter: `website` not `2website`
- 3-50 characters long
- No spaces allowed

### Error: "Invalid timezone"

**Cause:** Timezone string is not a valid IANA timezone

**Solutions:**
- Use exact IANA timezone format: `America/New_York` not `EST`
- Check [timezone database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) for correct value
- Common mistake: using abbreviations (EST) instead of full names (America/New_York)

### Error: "Domain already exists"

**Cause:** Custom domain is already used by another Webflow site or domain

**Solutions:**
- Ensure the domain isn't already in use
- Check domain registrar to confirm domain is available
- Verify domain DNS records aren't pointing to another service
- Try a different domain if needed

### Site Not Appearing

**Cause:** Site created but not yet visible in Webflow dashboard

**Solutions:**
- Wait a few moments for propagation
- Refresh Webflow dashboard
- Verify site ID in Pulumi output
- Check for creation errors in Pulumi output

### Cannot Delete Site

**Cause:** Site has dependent resources (redirects, robots.txt)

**Solutions:**
- Delete dependent resources first
- Or delete all resources together: `pulumi destroy`

## Related Resources

- [Provider Configuration](./provider-configuration.md) - Set up API authentication
- [RobotsTxt Resource](./robotstxt.md) - Manage robots.txt for your sites
- [Redirect Resource](./redirect.md) - Create redirects for your sites
- [Quickstart Guide](../README.md#quick-start) - Get started quickly
- [IANA Timezone Database](https://en.wikipedia.org/wiki/List_of_tz_database_time_zones) - Find your timezone
