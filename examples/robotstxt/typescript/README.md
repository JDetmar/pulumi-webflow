# RobotsTxt Example - TypeScript

This example demonstrates how to manage `robots.txt` files for your Webflow sites using Pulumi and the Webflow provider.

## What is robots.txt?

The `robots.txt` file is a simple text file placed in your site's root directory that tells search engine crawlers which pages they can and cannot access. This is essential for SEO and managing crawler load on your site.

## Example Contents

This example shows three different `robots.txt` configuration patterns:

1. **Allow All** - Standard configuration allowing all crawlers (most common for public sites)
2. **Selective Blocking** - Blocks specific directories and crawlers (useful for staging/development)
3. **Restrict Directories** - Protects sensitive directories like `/api/` and `/internal/`

## Prerequisites

- Node.js 14+ installed
- Pulumi CLI 3.0+ installed
- Webflow site ID (available in Webflow settings)
- Webflow API token for authentication

## Setup Instructions

### 1. Install Dependencies

```bash
npm install
```

This will install:
- `@pulumi/pulumi`: Pulumi framework
- `pulumi-webflow`: Webflow provider for Pulumi

### 2. Configure Pulumi Stack

Set up your Pulumi stack with the required configuration:

```bash
# Set your Webflow site ID (will be prompted for secret)
pulumi config set webflow:siteId your-site-id --secret

# Optionally set the environment
pulumi config set environment production
```

You can also create a `Pulumi.dev.yaml` file:

```yaml
config:
  webflow:siteId: your-site-id
  environment: development
```

### 3. Deploy

Preview the changes:
```bash
pulumi preview
```

Deploy to your Webflow site:
```bash
pulumi up
```

### 4. Verify Deployment

After successful deployment, your `robots.txt` files will be created on your Webflow site. You can verify by:

1. Visiting `https://your-site.com/robots.txt`
2. Checking the Webflow dashboard for the RobotsTxt resource
3. Running Pulumi to check the resource status:
   ```bash
   pulumi stack output
   ```

### 5. Cleanup

Remove all deployed resources:

```bash
pulumi destroy
```

## Expected Output

When you run `pulumi up`, you should see output similar to:

```
     Type                   Name                        Status
 +   pulumi:pulumi:Stack   webflow-robotstxt-example   created
 +   webflow:RobotsTxt     allow-all-robots            created
 +   webflow:RobotsTxt     selective-block-robots      created
 +   webflow:RobotsTxt     restrict-directories-robots created

Outputs:
    deployedSiteId: "your-site-id"
    allowAllRobotsId: "resource-id-1"
    allowAllRobotsLastModified: "2025-01-01T12:00:00Z"
    ...

✅ Successfully deployed RobotsTxt resources to site your-site-id
```

## Code Example Breakdown

### Basic RobotsTxt Creation

```typescript
const robotsTxt = new webflow.RobotsTxt("my-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`,
});
```

Key properties:
- `siteId` (required): Your Webflow site ID
- `content` (required): The robots.txt file content

### Configuration Values

The example uses Pulumi config to manage sensitive values:

```typescript
const config = new pulumi.Config();
const siteId = config.requireSecret("siteId");  // Secret value
const environment = config.get("environment");   // Regular config
```

## Common Patterns

### Allow All Search Engines (Default)

```typescript
const robotsTxt = new webflow.RobotsTxt("allow-all", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

Sitemap: https://example.com/sitemap.xml`,
});
```

### Block Specific Bots

```typescript
const robotsTxt = new webflow.RobotsTxt("block-bad-bots", {
  siteId: siteId,
  content: `User-agent: AhrefsBot
Disallow: /

User-agent: SemrushBot
Disallow: /

User-agent: *
Allow: /`,
});
```

### Restrict API Endpoints

```typescript
const robotsTxt = new webflow.RobotsTxt("api-restricted", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

Disallow: /api/
Disallow: /internal/
Disallow: /admin/`,
});
```

## Troubleshooting

### Issue: "Site not found" Error

**Solution**: Verify that your `siteId` is correct by:
1. Logging into Webflow
2. Going to Settings → General
3. Copying the exact Site ID

### Issue: "Authentication failed" Error

**Solution**: Ensure your Webflow API token is properly configured and has permission to manage the site.

### Issue: robots.txt Not Updating

**Solution**:
1. Search engines cache `robots.txt` for 24-48 hours
2. Use the Google Search Console to request immediate re-crawl
3. Check that Pulumi reported successful resource creation

### Issue: Import Errors

**Solution**: Ensure you've run `npm install`:
```bash
npm install
npm run build
```

## Advanced Usage

### Environment-Specific Robots Files

```typescript
const environment = config.get("environment") || "development";

const robotsContent = environment === "production"
  ? "User-agent: *\nAllow: /"  // Allow all in production
  : "User-agent: *\nDisallow: /";  // Block all in development

const robotsTxt = new webflow.RobotsTxt("env-specific-robots", {
  siteId: siteId,
  content: robotsContent,
});
```

### Multiple Sites

```typescript
const sites = ["site-1", "site-2", "site-3"];

for (let i = 0; i < sites.length; i++) {
  new webflow.RobotsTxt(`robots-${i}`, {
    siteId: sites[i],
    content: "User-agent: *\nAllow: /",
  });
}
```

## See Also

- [Webflow Provider Documentation](../../../docs/api/robotstxt.md)
- [RobotsTxt API Reference](../../../docs/api/robotstxt.md)
- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [robots.txt Standard](https://www.robotstxt.org/)

## Next Steps

- Explore other Webflow provider resources: [Redirect](../redirect) and [Site](../site)
- Learn about [Redirect resources](../redirect/typescript) for URL redirects
- See [Multi-Site Management](../../multi-site/basic-typescript) for managing multiple sites
- Check out [CI/CD Integration](../../ci-cd/typescript) examples
