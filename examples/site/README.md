# Site Resource Examples

This directory contains examples demonstrating how to create and manage Webflow sites using Pulumi in all supported languages.

## What You'll Learn

- Create basic Webflow sites with required properties
- Configure custom domains for sites
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
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### Go

```bash
cd go
go mod download
pulumi stack init dev
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### C# (.NET)

```bash
cd csharp
dotnet restore
pulumi stack init dev
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

### Java

```bash
cd java
mvn install
pulumi stack init dev
pulumi config set displayName "My Site"
pulumi config set shortName "my-site"
pulumi up
```

## Examples Included

### 1. Basic Site Creation

Create a simple site with required properties.

```typescript
const basicSite = new webflow.Site("basic-site", {
    displayName: "My Website",
    shortName: "my-website",
    timezone: "America/New_York",
});
```

### 2. Site with Custom Domain

Create a site with a custom domain configured.

```typescript
const siteWithDomain = new webflow.Site("site-with-domain", {
    displayName: "My Production Site",
    shortName: "my-prod-site",
    customDomain: "www.example.com",
    timezone: "America/New_York",
});
```

### 3. Multi-Environment Configuration

Create sites for different environments (development, staging, production).

```typescript
const environments = ["development", "staging", "production"];
const sites = environments.map(env =>
    new webflow.Site(`site-${env}`, {
        displayName: `My Site - ${env}`,
        shortName: `my-site-${env}`,
        timezone: "America/New_York",
    })
);
```

### 4. Full Configuration Example

Demonstrates all available configuration options.

## Configuration

Each example requires the following configuration:

| Config Key      | Required | Description                                    |
|-----------------|----------|------------------------------------------------|
| `displayName`   | Yes      | Human-readable name for the site               |
| `shortName`     | Yes      | URL slug for the site                          |
| `customDomain`  | No       | Custom domain (e.g., www.example.com)          |
| `timezone`      | No       | Site timezone (default: America/New_York)      |
| `environment`   | No       | Deployment environment (default: development)  |

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    basicSiteId           : "abc123..."
    basicSiteName         : "My Site"
    customDomainSiteId    : "def456..." (or "not-created")
    environmentSiteIds    : ["ghi789...", "jkl012...", "mno345..."]
    configuredSiteId      : "pqr678..."
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

### "Custom domain already in use" Error

The custom domain is already configured on another site. Either:
1. Remove it from the other site first
2. Use a different domain

### "Quota exceeded" Error

You've reached the maximum number of sites for your Webflow plan. Upgrade your plan or delete unused sites.

## Related Resources

- [Site API Reference](../../docs/api/site.md)
- [Main Examples Index](../README.md)
- [Webflow Site Settings](https://university.webflow.com/lesson/site-settings)
