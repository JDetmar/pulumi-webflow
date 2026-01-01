# Webflow Provider Configuration

Configure the Webflow Pulumi provider with your API credentials and authentication settings.

## Overview

The Webflow provider requires an API token for authentication with the Webflow platform. This token is used for all API calls to create, read, update, and delete resources.

## Authentication

### Webflow API Token

The provider requires a Webflow API token (also called an access token). This is a secret credential that should never be committed to version control.

**Where to Get Your Token:**

1. Log in to your Webflow account at [webflow.com](https://webflow.com)
2. Go to **Account Settings** → **Integrations** → **API & Webhooks**
3. Click **Generate Access Token**
4. Copy the token (you'll only see it once)
5. Store it securely - never commit to Git

**Token Requirements:**

- Webflow Enterprise account or higher
- API access must be enabled for your workspace
- The token should have permissions for the resources you're managing

## Configuration

The API token can be configured in several ways:

### Method 1: Pulumi Configuration (Recommended for Teams)

Store the token in encrypted Pulumi configuration:

```bash
pulumi config set webflow:apiToken --secret
# When prompted, paste your Webflow API token
```

The token will be:
- Encrypted at rest in `Pulumi.<stack>.yaml`
- Marked as secret (not shown in logs or outputs)
- Protected by Pulumi's built-in encryption

### Method 2: Environment Variable

Set the `WEBFLOW_API_TOKEN` environment variable:

```bash
export WEBFLOW_API_TOKEN=your-api-token-here
```

This is useful for:
- Local development
- CI/CD pipelines (set as GitHub Actions secret, GitLab CI variable, etc.)
- Container deployments

### Method 3: Program Configuration

Explicitly pass the token in your Pulumi program:

```typescript
const config = new pulumi.Config();
const apiToken = config.requireSecret("apiToken");

const provider = new pulumi.providers.webflow.Provider("my-provider", {
  apiToken: apiToken,
});
```

## Example Usage

### TypeScript

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// The provider will automatically use webflow:apiToken from config
// No need to explicitly create a provider - resources use the default

const robotsTxt = new webflow.RobotsTxt("my-robots", {
  siteId: "5f0c8c9e1c9d440000e8d8c3",
  content: `User-agent: *
Allow: /`,
});

export const resourceId = robotsTxt.id;
```

### Python

```python
import pulumi
import webflow_webflow as webflow

# The provider will automatically use webflow:apiToken from config

robots_txt = webflow.RobotsTxt(
    "my-robots",
    site_id="5f0c8c9e1c9d440000e8d8c3",
    content="""User-agent: *
Allow: /""",
)

pulumi.export("resource_id", robots_txt.id)
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
		// The provider will automatically use WEBFLOW_API_TOKEN or webflow:apiToken from config
		cfg := config.New(ctx, "")
		siteId := cfg.RequireSecret("siteId")

		robotsTxt, err := webflow.NewRobotsTxt(ctx, "my-robots", &webflow.RobotsTxtArgs{
			SiteId: siteId,
			Content: pulumi.String(`User-agent: *
Allow: /`),
		})
		if err != nil {
			return err
		}

		ctx.Export("resourceId", robotsTxt.ID())
		return nil
	})
}
```

## Configuration Options

### webflow:apiToken (Required)

- **Type:** String (Secret)
- **Description:** Webflow API token for authentication
- **Default:** None (required)
- **Environment Variable:** `WEBFLOW_API_TOKEN`

## Security Best Practices

### DO ✅

- ✅ Store tokens in Pulumi secrets or environment variables
- ✅ Use encrypted configuration for team environments
- ✅ Rotate tokens regularly
- ✅ Restrict token permissions to only needed resources
- ✅ Audit token usage in your Webflow account
- ✅ Use separate tokens for different environments (dev, staging, prod)

### DON'T ❌

- ❌ Commit tokens to version control (even in `.gitignore` files)
- ❌ Log tokens to console or files
- ❌ Share tokens via email or chat
- ❌ Use the same token across multiple teams or environments
- ❌ Store tokens in plain text files

## Environment-Specific Configuration

### Development Environment

```bash
# Set token in local environment
export WEBFLOW_API_TOKEN=your-dev-token

# Run Pulumi
pulumi up
```

### CI/CD Pipeline (GitHub Actions Example)

```yaml
- name: Deploy with Pulumi
  env:
    WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
  run: |
    pulumi up --yes
```

### Multiple Environments (dev/staging/prod)

Create separate stacks with different tokens:

```bash
# Create separate stacks
pulumi stack init dev
pulumi stack init staging
pulumi stack init prod

# Configure each stack
pulumi config set webflow:apiToken --secret  # Prompts for token for current stack

# Switch between stacks
pulumi stack select dev
pulumi up

pulumi stack select prod
pulumi up
```

## Troubleshooting

### "Invalid API token" Error

**Cause:** Token is missing, invalid, or expired

**Solutions:**
1. Verify token is set: `pulumi config get webflow:apiToken`
2. Check token in Webflow account - may be expired
3. Generate a new token if needed
4. Ensure token is set before running `pulumi up`

### "Authentication failed" Error

**Cause:** Token lacks required permissions

**Solutions:**
1. Verify token has API permissions in your Webflow account
2. Check token is for the correct workspace
3. Ensure token is not expired
4. Try generating a new token with full permissions

### Token Not Found

**Cause:** Token not provided via config or environment variable

**Solutions:**
1. Set via Pulumi config: `pulumi config set webflow:apiToken --secret`
2. OR set environment variable: `export WEBFLOW_API_TOKEN=your-token`
3. Verify configuration before running `pulumi up`

## Related Documentation

- [Quick Start Guide](../README.md#quick-start) - Get started in 20 minutes
- [Troubleshooting Guide](../troubleshooting.md) - Comprehensive authentication troubleshooting
- [FAQ](../faq.md) - Authentication and configuration FAQs
- [RobotsTxt Resource](./robotstxt.md) - Manage robots.txt configurations
- [Redirect Resource](./redirect.md) - Create and manage redirects
- [Site Resource](./site.md) - Manage Webflow sites
- [Pulumi Secrets Documentation](https://www.pulumi.com/docs/concepts/secrets/) - Secure credential management
