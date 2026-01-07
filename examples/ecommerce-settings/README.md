# Ecommerce Settings Resource Examples

This directory contains examples demonstrating how to import and track Webflow ecommerce settings using Pulumi.

## What You'll Learn

- Import existing ecommerce settings into Pulumi state
- Access the site's default currency for use in other resources
- Verify that ecommerce is enabled on a site
- Track when ecommerce was enabled on your site

## What Are Webflow Ecommerce Settings?

Webflow Ecommerce Settings represent the configuration for a site's ecommerce functionality. These settings include:

- **Default Currency**: The three-letter ISO 4217 currency code (e.g., "USD", "EUR", "GBP")
- **Created On**: When ecommerce was enabled on the site

**Important**: This is a **read-only** resource. Ecommerce must be enabled and configured through the Webflow dashboard. This resource allows you to:
- Import and track existing settings as infrastructure state
- Reference the currency in other ecommerce-related resources
- Verify ecommerce is enabled before deploying dependent resources

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
pulumi config set siteId your-site-id-here --secret
pulumi up
```

## Examples Included

### 1. Basic Ecommerce Settings Import

Import the ecommerce settings for a site to track them in Pulumi state.

```typescript
const ecommerceSettings = new webflow.EcommerceSettings("site-ecommerce", {
  siteId: siteId,
});

// Access the default currency
export const currency = ecommerceSettings.defaultCurrency;
```

**Use Case:** Track ecommerce configuration as part of your infrastructure state and reference the currency in other resources.

## Configuration

Each example requires the following configuration:

| Config Key        | Required | Description                              |
|-------------------|----------|------------------------------------------|
| `siteId`          | Yes      | Your Webflow site ID (must have ecommerce enabled) |
| `environment`     | No       | Deployment environment (default: development) |

**Finding Your Site ID:**
1. Log in to Webflow
2. Go to Site Settings > General
3. Copy the Site ID (24-character lowercase hexadecimal string)

**Enabling Ecommerce:**
1. Go to your Webflow site dashboard
2. Navigate to Ecommerce tab
3. Follow the setup wizard to enable ecommerce
4. Configure your payment provider and currency

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    deployedSiteId           : "abc123..."
    ecommerceSiteId          : "abc123..."
    defaultCurrency          : "USD"
    ecommerceCreatedOn       : "2024-01-15T10:30:00Z"
```

## Important Notes

### Read-Only Resource

This resource is **read-only**. You cannot:
- Enable ecommerce via the API (must be done in Webflow dashboard)
- Change the default currency via the API
- Disable ecommerce via the API

The resource simply imports and tracks the existing settings.

### Ecommerce Must Be Enabled First

Before using this resource, ecommerce must be enabled on your site:

1. Go to your Webflow site dashboard
2. Click on "Ecommerce" in the left sidebar
3. Complete the ecommerce setup wizard
4. Configure your payment provider and currency settings

If ecommerce is not enabled, you'll see a 409 error:
```
ecommerce not enabled: the site does not have ecommerce enabled
```

### Required API Scope

Your Webflow API token must have the `ecommerce:read` scope to access ecommerce settings.

## Cleanup

To remove the resource from Pulumi state:

```bash
pulumi destroy
pulumi stack rm dev
```

**Note:** This only removes the resource from Pulumi state. It does NOT disable ecommerce on your site (that must be done through the Webflow dashboard).

## Troubleshooting

### "Site not found" Error

1. Verify your site ID in Webflow: Settings > General
2. Ensure correct format: 24-character lowercase hexadecimal
3. Check API token has access to the site

### "Ecommerce not enabled" Error (409)

1. Log into your Webflow dashboard
2. Go to the Ecommerce tab for your site
3. Complete the ecommerce setup wizard
4. Retry the Pulumi deployment

### "Unauthorized" Error (401)

1. Verify your API token is valid
2. Ensure the token has `ecommerce:read` scope
3. Check the token hasn't expired

### "Forbidden" Error (403)

1. Verify the API token has access to this specific site
2. Check that the site belongs to the workspace associated with your token
3. Ensure you have the correct permissions in Webflow

## Related Resources

- [Main Examples Index](../README.md)
- [Webflow Ecommerce Documentation](https://university.webflow.com/lesson/intro-to-ecommerce)
- [Webflow API Reference](https://developers.webflow.com/data/reference/ecommerce/settings/get-settings)

## Next Steps

After importing ecommerce settings, consider:
- Using the `defaultCurrency` output in other ecommerce-related resources
- Setting up webhooks for ecommerce events (orders, inventory changes)
- Creating product collections and items
- Building automated ecommerce workflows
