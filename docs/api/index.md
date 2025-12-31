# Webflow Pulumi Provider - API Reference

Complete API reference documentation for all Webflow Pulumi Provider resources.

## Resources

The Webflow Pulumi Provider exposes three core resources for managing Webflow infrastructure:

### 1. RobotsTxt Resource
Manage robots.txt configuration for Webflow sites to control crawler access and define sitemaps.

- **Namespace:** `webflow:index:RobotsTxt`
- **Complexity:** Simple
- **Key Use Cases:** SEO configuration, crawler control, search engine optimization
- **[Full Documentation →](./robotstxt.md)**

### 2. Redirect Resource
Create and manage HTTP redirects (301/302) for Webflow sites.

- **Namespace:** `webflow:index:Redirect`
- **Complexity:** Medium
- **Key Use Cases:** URL redirect management, legacy URL handling, domain consolidation
- **[Full Documentation →](./redirect.md)**

### 3. Site Resource
Create, configure, publish, and manage complete Webflow sites.

- **Namespace:** `webflow:index:Site`
- **Complexity:** Advanced
- **Key Use Cases:** Site lifecycle management, multi-environment deployments, site creation automation
- **[Full Documentation →](./site.md)**

## Provider Configuration

Configure the Webflow provider with your API credentials.

- **[Provider Configuration Guide →](./provider-configuration.md)**

## Language Support

All resources are available in multiple languages. Examples are provided for the primary languages:

- **TypeScript** - Primary language, most common in Pulumi (examples included)
- **Python** - Available with snake_case property naming (examples included)
- **Go** - Idiomatic Go patterns and types (examples included)
- **C#** - .NET framework support (SDK available, examples coming soon)
- **Java** - Enterprise Java support (SDK available, examples coming soon)

## Quick Links

- [Quickstart Guide](../README.md#quick-start) - Get started in under 20 minutes
- [Examples](../examples/) - Ready-to-run example programs
- [Troubleshooting](../examples/troubleshooting-logs/README.md) - Common issues and solutions

## Documentation Format

Each resource documentation includes:

- **Overview** - What the resource manages and why you'd use it
- **Example Usage** - Copy-pasteable code in all supported languages
- **Argument Reference** - Complete property documentation with types and constraints
- **Attribute Reference** - Output properties returned by the resource
- **Common Patterns** - How to use the resource in real-world scenarios
- **Troubleshooting** - Common errors and how to fix them
- **Related Resources** - Links to related documentation

## Property Naming Conventions

Properties are named consistently across languages, but follow language conventions:

| Language   | Naming Convention | Example          |
|------------|-------------------|------------------|
| TypeScript | camelCase         | `siteId`         |
| Python     | snake_case        | `site_id`        |
| Go         | PascalCase        | `SiteId`         |
| C#         | PascalCase        | `SiteId`         |
| Java       | camelCase         | `siteId`         |

All examples explicitly show the correct naming for each language.

## Schema Information

The provider schema is automatically generated from the provider implementation and published to the Pulumi Registry. The SDK packages for each language are also auto-generated to ensure consistency across the provider surface area.

---

**API Reference Version:** 1.0.0-alpha.0+dev
**Last Updated:** 2025-12-30
