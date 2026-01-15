# Site Resource Examples

This directory contains examples demonstrating how to create and manage Webflow sites using Pulumi in all supported languages.

## What You'll Learn

- Create basic Webflow sites with required properties
- Configure workspaces and site settings
- Implement multi-environment site configurations
- Manage site lifecycle (create, update, delete)

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |
| Python     | `python/`    | `__main__.py`  | `requirements.txt`  |
| Go         | `go/`        | `main.go`      | `go.mod`            |
| C#         | `csharp/`    | `Program.cs`   | `.csproj`           |
| Java       | `java/`      | `App.java`     | `pom.xml`           |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi config set workspaceId "YOUR_WORKSPACE_ID"
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### Python

```bash
cd python
python3 -m venv venv
source venv/bin/activate
pip install -r requirements.txt
pulumi stack init dev
pulumi config set workspaceId "YOUR_WORKSPACE_ID"
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set workspaceId "YOUR_WORKSPACE_ID"
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### C# (.NET)

```bash
cd csharp
dotnet restore
pulumi stack init dev
pulumi config set workspaceId "YOUR_WORKSPACE_ID"
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### Java

```bash
cd java
mvn install
pulumi stack init dev
pulumi config set workspaceId "YOUR_WORKSPACE_ID"
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

## Examples Included

### 1. Basic Site Creation

Create a simple site with required properties.

```typescript
const basicSite = new webflow.Site("basic-site", {
    workspaceId: "YOUR_WORKSPACE_ID",
    displayName: "My Website",
    shortName: "my-website",
});

// The site's timezone is available as a read-only output:
// basicSite.timeZone
```

### 2. Multi-Environment Configuration

Create sites for different environments (development, staging, production).

```typescript
const environments = ["development", "staging", "production"];
const sites = environments.map(env =>
    new webflow.Site(`site-${env}`, {
        workspaceId: "YOUR_WORKSPACE_ID",
        displayName: `My Site - ${env}`,
        shortName: `my-site-${env}`,
    })
);
```

### 3. Full Configuration Example

Demonstrates all available configuration options.

## Configuration

Each example requires the following configuration:

| Config Key      | Required | Description                                    |
|-----------------|----------|------------------------------------------------|
| `workspaceId`   | Yes      | Webflow workspace ID where site will be created |
| `displayName`   | Yes      | Human-readable name for the site               |
| `shortName`     | Yes      | URL slug for the site                          |

**Note:** The site's `timeZone` is a read-only output field that reflects the timezone configured in Webflow. It cannot be set via the API.

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    basicSiteId           : "abc123..."
    basicSiteName         : "My Site"
    environmentSiteIds    : ["def456...", "ghi789...", "jkl012..."]
    configuredSiteId      : "mno345..."
```

## Cleanup

To remove all created sites:

```bash
pulumi destroy
pulumi stack rm dev
```

## Troubleshooting

### "Invalid short name" Error

The short name must be:
- Lowercase letters, numbers, and hyphens only
- No spaces or special characters
- Unique within your Webflow account

### "Invalid workspace ID" Error

The workspace ID must be a valid Webflow workspace ID. You can find your workspace ID in the Webflow dashboard under Account Settings > Workspace. Note that an Enterprise workspace is required for site creation via the API.

### "Quota exceeded" Error

You've reached the maximum number of sites for your Webflow plan. Upgrade your plan or delete unused sites.

## Related Resources

- [Main Examples Index](../README.md)
- [Webflow Site Settings](https://university.webflow.com/lesson/site-settings)
- [Webflow Sites API](https://developers.webflow.com/reference/sites)
