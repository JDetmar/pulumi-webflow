# Webflow Pulumi Provider

[![License](https://img.shields.io/badge/License-Apache%202.0-blue.svg)](https://opensource.org/licenses/Apache-2.0)
[![Go Report Card](https://goreportcard.com/badge/github.com/jdetmar/pulumi-webflow)](https://goreportcard.com/report/github.com/jdetmar/pulumi-webflow)

> Infrastructure as Code for Webflow Sites

The Webflow Pulumi Provider enables you to manage Webflow sites programmatically through infrastructure as code. Provision, configure, and manage Webflow sites, redirects, and robots.txt configurations using your favorite programming language.

## Features

- **Multi-Language Support**: Write infrastructure code in TypeScript, Python, Go, C#, or Java
- **Webflow Resource Management**: Manage sites, redirects, and robots.txt configurations
- **Preview Changes**: See exactly what will change before applying (Story 1.7)
- **State Management**: Automatic drift detection and state synchronization
- **Infrastructure as Code**: Version control your Webflow infrastructure
- **CI/CD Integration**: Automate site deployments through your existing pipelines

## Installation

### Prerequisites

- [Pulumi CLI](https://www.pulumi.com/docs/get-started/install/) installed
- A [Webflow account](https://webflow.com/) with API access
- Webflow API token ([generate here](https://webflow.com/dashboard/account/apps))

### Install the Provider

The provider will be automatically installed when you reference it in your Pulumi program.

Alternatively, you can install it manually:

```bash
pulumi plugin install resource webflow
```

## Quick Start

> **Note:** The examples below demonstrate planned functionality. Resource implementations are in active development (see [Development Status](#development-status) section).

### TypeScript Example

```typescript
import * as webflow from "pulumi-webflow";

// Configure robots.txt for a Webflow site (Coming in Story 1.4+)
const robotsTxt = new webflow.RobotsTxt("site-robots", {
    siteId: "your-site-id",
    content: `User-agent: *
Allow: /`,
});

export const robotsId = robotsTxt.id;
```

### Python Example

```python
import pulumi_webflow as webflow

# Configure robots.txt for a Webflow site
robots_txt = webflow.RobotsTxt("site-robots",
    site_id="your-site-id",
    content="""User-agent: *
Allow: /""")

pulumi.export("robots_id", robots_txt.id)
```

### Go Example

```go
package main

import (
	"github.com/jdetmar/pulumi-webflow/sdk/go/webflow"
	"github.com/pulumi/pulumi/sdk/v3/go/pulumi"
)

func main() {
	pulumi.Run(func(ctx *pulumi.Context) error {
		robotsTxt, err := webflow.NewRobotsTxt(ctx, "site-robots", &webflow.RobotsTxtArgs{
			SiteId: pulumi.String("your-site-id"),
			Content: pulumi.String("User-agent: *\nAllow: /"),
		})
		if err != nil {
			return err
		}

		ctx.Export("robotsId", robotsTxt.ID())
		return nil
	})
}
```

## Resources

### Development Status

#### Current Phase: Story 1.7 - Preview/Plan Workflow (Completed)

The provider infrastructure is now complete with:

- ✅ Provider schema generation for multi-language SDK support
- ✅ Complete CRUD lifecycle methods (Check, Diff, Create, Read, Update, Delete)
- ✅ Preview workflow with accurate change detection (Story 1.7)
- ✅ Authentication and configuration management
- ✅ Plugin manifest and distribution setup
- ✅ Comprehensive test coverage (57+ tests)
- ✅ Performance validated (<2s startup time, <10s preview)

#### Next Phase: Resource Implementations (Stories 1.4+)

- **RobotsTxt**: Manage robots.txt configuration for Webflow sites (Story 1.4)
- **Redirect**: Create and manage redirects (301/302) (Epic 2)
- **Site**: Complete site lifecycle management (create, update, publish, delete) (Epic 3)

### Coming Soon

- Forms management
- Collections and CMS
- Custom code injection
- Webhooks integration

## Configuration

### Authentication

The provider requires a Webflow API v2 token for authentication. You can configure the token in two ways:

**Option 1: Pulumi Configuration (Recommended)**

```bash
pulumi config set webflow:token your-api-token --secret
```

The `--secret` flag ensures the token is encrypted in your Pulumi state file.

**Option 2: Environment Variable**

```bash
export WEBFLOW_API_TOKEN="your-api-token"
```

**Generating a Webflow API Token:**

1. Log in to your Webflow account
2. Navigate to Site Settings → Apps & Integrations
3. Scroll to the "API access" section
4. Click "Generate API token"
5. Choose appropriate scopes for your use case
6. Copy the token (it will only be shown once!)

**Security Best Practices:**
- Always use the `--secret` flag when setting tokens via Pulumi config
- Never commit tokens to version control
- Rotate tokens regularly (every 30-90 days recommended)
- Use scoped tokens with minimal required permissions
- Tokens expire after 365 consecutive days of inactivity

## Preview Workflow

The provider supports Pulumi's preview workflow, allowing you to see exactly what changes will be made before applying them.

### Previewing Changes

Use `pulumi preview` to see a detailed preview of changes:

```bash
pulumi preview
```

The preview will show:
- **Create operations** (`+`): New resources that will be created
- **Update operations** (`~`): Existing resources that will be modified
- **Delete operations** (`-`): Resources that will be removed
- **Property-level changes**: Detailed before/after values for each changed property

### Preview Output Example

```
Previewing update (dev):

     Type                      Name              Plan
     pulumi:pulumi:Stack       webflow-dev
 ~   webflow:index:RobotsTxt   site-robots       update
     └─ content                "User-agent: *\nAllow: /" => "User-agent: *\nAllow: /\nDisallow: /admin/"

Resources:
    ~ 1 to update
    1 unchanged
```

### Preview Features

- **Fast Performance**: Preview completes in milliseconds (no API calls made)
- **Accurate Changes**: Preview exactly matches what will be applied
- **Sensitive Data Protection**: API tokens and credentials are automatically redacted
- **Change Detection**: Automatically detects create, update, and delete operations
- **Property-Level Details**: Shows exactly which properties will change

### Preview Best Practices

1. **Always preview before applying**: Run `pulumi preview` before `pulumi up` to verify changes
2. **Review property changes**: Check that all property changes are expected
3. **Verify no unexpected modifications**: Ensure only intended changes are shown
4. **Check sensitive data**: Confirm that secrets are properly redacted (shown as `[secret]`)

### Preview Workflow

```
1. Modify your Pulumi program
2. Run `pulumi preview` to see changes
3. Review the preview output
4. Run `pulumi up` to apply changes (or `pulumi up --yes` to skip confirmation)
```

The preview workflow is fully integrated with Pulumi's standard commands and requires no additional configuration.

### Troubleshooting Preview Issues

**Preview shows "no changes" but you expect changes:**
- Verify your Pulumi program has been saved
- Check that resource properties are actually different from current state
- Run `pulumi refresh` to sync state with actual Webflow configuration
- Ensure you're previewing the correct stack (`pulumi preview --stack <stack-name>`)

**Preview is slow (>10 seconds):**
- Check network connectivity to Webflow API
- Verify API token has correct permissions
- Check for rate limiting (429 errors in verbose logs)
- Run `pulumi preview --verbose` to see detailed timing

**Preview shows unexpected changes:**
- Run `pulumi refresh` to sync state
- Check if resources were modified manually in Webflow UI
- Verify your Pulumi program matches intended configuration
- Review DetailedDiff output for property-level changes

**Preview doesn't show sensitive data redaction:**
- Verify token is set with `--secret` flag: `pulumi config set webflow:token <token> --secret`
- Check that Config struct properly marks token as secret
- Sensitive values should appear as `[secret]` in preview output

**Preview fails with authentication errors:**
- Verify API token is set: `pulumi config get webflow:token`
- Check token hasn't expired (tokens expire after 365 days of inactivity)
- Ensure token has required scopes (site_config:read for preview)
- Try regenerating token in Webflow dashboard

**Integration Testing:**
- See `examples/yaml-test/Pulumi.yaml` for integration test example
- Run `pulumi preview` in the example directory to test preview workflow
- Verify preview output matches expected changes before running `pulumi up`

## Documentation

- [API Reference](https://www.pulumi.com/registry/packages/webflow/api-docs/) _(coming soon)_
- [Examples](./examples/) _(coming soon)_
- [Webflow API Documentation](https://developers.webflow.com/reference/rest-introduction)

## SDK Generation

The provider automatically generates SDKs for multiple programming languages from the provider schema.

### Generating SDKs Locally

```bash
# Generate provider schema from binary
make gen-schema

# Generate all language SDKs from schema
make gen-sdks

# Build and test all SDKs
make build-sdks

# Build individual SDKs
make build-sdk-nodejs    # TypeScript/JavaScript SDK
make build-sdk-python    # Python SDK
make build-sdk-go        # Go SDK
make build-sdk-dotnet    # C# SDK
make build-sdk-java      # Java SDK (requires Maven)
```

### SDK Directory Structure

Generated SDKs are placed in the `sdk/` directory:

```text
sdk/
├── nodejs/               # TypeScript/JavaScript SDK (pulumi-webflow)
│   ├── package.json      # NPM package configuration
│   ├── index.ts          # Main entry point
│   ├── provider.ts       # Provider configuration
│   ├── site.ts           # Site resource
│   ├── redirect.ts       # Redirect resource
│   ├── robotsTxt.ts      # RobotsTxt resource
│   └── config/           # Configuration types
├── python/               # Python SDK (webflow_webflow)
│   ├── pyproject.toml    # Package configuration
│   └── webflow_webflow/  # Python module
├── go/                   # Go SDK
│   └── webflow/          # Go package
├── dotnet/               # C# SDK
│   └── *.cs files        # C# resource classes
└── java/                 # Java SDK
    ├── pom.xml           # Maven configuration
    └── src/              # Java sources
```

### SDK Versioning

- SDKs are version-synced with the provider version
- Version is set in `pulumi-plugin.json` and `Makefile`
- Generated during release builds automatically

## TypeScript SDK

The TypeScript/JavaScript SDK (`pulumi-webflow`) enables Node.js developers to manage Webflow infrastructure through Pulumi.

### Installation

```bash
npm install pulumi-webflow
```

### TypeScript Quick Start

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Create a robots.txt resource
const robotsTxt = new webflow.RobotsTxt("site-robots", {
    siteId: "your-site-id",
    content: `User-agent: *
Allow: /`,
});

// Create a redirect
const redirect = new webflow.Redirect("old-page-redirect", {
    siteId: "your-site-id",
    sourcePath: "/old-page",
    destinationPath: "/new-page",
    statusCode: 301,
});

// Create or manage a site
const site = new webflow.Site("my-site", {
    workspaceId: "your-workspace-id",
    displayName: "My New Site",
    shortName: "my-site",
});

export const siteId = site.id;
```

### IDE Setup & Type Hints

The TypeScript SDK includes full type definitions for excellent IDE support:

**VS Code:**
- Install [Pulumi Extension](https://marketplace.visualstudio.com/items?itemName=pulumi.pulumi)
- Enable automatic type checking in `settings.json`:
  ```json
  {
    "typescript.checkJs": true,
    "typescript.enablePromptUseWorkspaceTsdk": true
  }
  ```

**IntelliSense:**
- All resource types have complete JSDoc documentation
- Hover over any resource or property to see documentation
- Type errors are caught at development time

### TypeScript Configuration

The SDK supports Node.js 18.0.0 or later and works with both ES Modules and CommonJS:

```typescript
// ES Modules (modern)
import * as webflow from "pulumi-webflow";

// CommonJS (traditional)
const webflow = require("pulumi-webflow");
```

### Troubleshooting TypeScript SDK

**Cannot find module 'pulumi-webflow'**
- Ensure `npm install pulumi-webflow` completed successfully
- Check `node_modules/pulumi-webflow/bin/index.js` exists
- Verify `npm list pulumi-webflow` shows the package installed

**Missing type definitions**
- Ensure `types` field in package.json points to `bin/index.d.ts`
- TypeScript should automatically find types in `node_modules`
- Try: `npm reinstall pulumi-webflow`

**Version mismatch errors**
- Ensure Pulumi CLI version matches the SDK version
- SDK version should match provider version (both 0.1.0+)
- See [Pulumi docs on version compatibility](https://www.pulumi.com/docs/reference/pulumi-cli/)

## Development

See [CONTRIBUTING.md](./CONTRIBUTING.md) for information on building and contributing to this provider.

### Building from Source

```bash
# Clone the repository
git clone https://github.com/jdetmar/pulumi-webflow
cd pulumi-webflow

# Build the provider
go build -o pulumi-resource-webflow

# Install locally for testing
cp pulumi-resource-webflow $HOME/.pulumi/plugins/resource-webflow-v0.1.0/
```

## Support

- **Issues**: [GitHub Issues](https://github.com/jdetmar/pulumi-webflow/issues)
- **Community**: [Pulumi Community Slack](https://slack.pulumi.com/)
- **Documentation**: [Pulumi Registry](https://www.pulumi.com/registry/)

## License

This provider is licensed under the Apache License, Version 2.0. See [LICENSE](./LICENSE) for the full license text.

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for details on how to contribute to this project.

---

Made with ❤️ by the Pulumi community
