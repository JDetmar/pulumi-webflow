# Multi-Site Management Examples

Comprehensive examples demonstrating how to manage multiple Webflow sites in a single Pulumi program, from simple to advanced patterns.

## Overview

Managing multiple Webflow sites programmatically eliminates manual UI navigation, enables version control for site configurations, and integrates Webflow deployments into your CI/CD pipelines. These examples show four increasingly sophisticated patterns for managing site fleets at scale.

## Quick Start

All examples require:
- Webflow API token: `WEBFLOW_API_TOKEN` environment variable
- Pulumi CLI installed
- Language-specific runtime (Node.js for TypeScript, Python, or Go)

Set your Webflow API token:
```bash
export WEBFLOW_API_TOKEN="your-webflow-api-token"
```

## Example Patterns

### 1. Basic Multi-Site (Hardcoded)

**Use when:** Starting with a small, fixed number of sites (3-5).

**Available in:**
- TypeScript: [basic-typescript/](basic-typescript/)
- Python: [basic-python/](basic-python/)
- Go: [basic-go/](basic-go/)

**Pattern:** Define sites in an array and create them using `.map()` or list comprehension.

```typescript
// TypeScript
const siteNames = ["site-1", "site-2", "site-3"];
const sites = siteNames.map(name =>
  new webflow.Site(name, { displayName: name })
);
```

```python
# Python
site_configs = [
    {"name": "site-1", "display_name": "Site 1"},
    {"name": "site-2", "display_name": "Site 2"},
]
sites = [
    webflow.Site(config["name"], display_name=config["display_name"])
    for config in site_configs
]
```

```go
// Go
for i := 0; i < 3; i++ {
    site, _ := webflow.NewSite(ctx, fmt.Sprintf("site-%d", i),
        &webflow.SiteArgs{...})
}
```

**Pros:**
- Simple, easy to understand
- Good for learning the basic pattern
- Minimal dependencies

**Cons:**
- Not scalable beyond a few sites
- Configuration changes require code edits
- Hard to maintain larger fleets

**Run it:**
```bash
cd basic-typescript
npm install
pulumi up
```

---

### 2. Configuration-Driven (YAML)

**Use when:** Managing 10-20+ sites with varying configurations.

**Available in:**
- TypeScript: [config-driven-typescript/](config-driven-typescript/)

**Pattern:** Load site configurations from YAML file, apply defaults, and create sites programmatically.

**Key Files:**
- `index.ts` - Pulumi program that reads `sites.yaml`
- `sites.yaml` - Fleet configuration with defaults and per-site overrides

```yaml
defaults:
  timeZone: "America/Los_Angeles"
  robotsTxtContent: "User-agent: *\nAllow: /"

sites:
  - name: "marketing-site"
    displayName: "Marketing Site"
    redirects:
      - sourcePath: "/old-campaign"
        destinationPath: "/new-campaign"
        statusCode: 301
```

**Benefits:**
- Separate configuration from code
- Easy to manage dozens of sites
- Shared defaults reduce repetition
- Version-controlled fleet definitions

**Run it:**
```bash
cd config-driven-typescript
npm install
pulumi up
```

**Deploying 100 sites:** Create a `sites.yaml` with 100 site entries and deploy:
```bash
pulumi up
```

Pulumi automatically parallelizes resource creation, handling all 100 sites efficiently.

---

### 3. Template-Based (Factory Pattern)

**Use when:** Managing groups of similar sites (campaigns, products, events) with consistent patterns.

**Available in:**
- Python: [template-python/](template-python/)

**Pattern:** Create reusable factory functions that encapsulate site configuration templates.

```python
# Reusable factory functions
def create_campaign_site(name: str, campaign_name: str) -> webflow.Site:
    """Create standardized campaign site with built-in redirects & robots.txt"""
    site = webflow.Site(name, display_name=campaign_name, ...)
    
    # Standard campaign redirects
    webflow.Redirect(f"{name}-signup", site_id=site.id, ...)
    
    return site

# Use the factory
campaigns = [
    ("q1-promo", "Q1 Promotion"),
    ("summer-sale", "Summer Sale"),
]
sites = [create_campaign_site(name, display) for name, display in campaigns]
```

**Benefits:**
- Ensures consistency across similar sites
- Reduces configuration boilerplate
- Easy to evolve patterns over time
- Clear intent (campaign vs product vs event site)

**Run it:**
```bash
cd template-python
pip install -r requirements.txt
pulumi up
```

**Customizing templates:** Edit `site_templates.py` to add site types or modify defaults.

---

### 4. Multi-Environment (Stacks)

**Use when:** Managing separate dev, staging, and production site fleets.

**Available in:**
- Go: [multi-env-go/](multi-env-go/)

**Pattern:** Use Pulumi stacks with environment-specific configuration files.

**Stack-specific configs:**
- `Pulumi.dev.yaml` - Development: 3 sites
- `Pulumi.staging.yaml` - Staging: 5 sites
- `Pulumi.prod.yaml` - Production: 10 sites

```bash
# Deploy development fleet (3 sites)
pulumi stack select dev
pulumi up

# Deploy staging fleet (5 sites)
pulumi stack select staging
pulumi up

# Deploy production fleet (10 sites)
pulumi stack select prod
pulumi up
```

**Benefits:**
- Environment-specific configurations
- Separate state per environment
- Different site counts (dev=3, staging=5, prod=10)
- Easy to manage full lifecycle

**Managing 100-site production fleet:**
```bash
# Set site count in Pulumi.prod.yaml
siteCount: 100

# Deploy
pulumi stack select prod
pulumi up
```

---

## Choosing a Pattern

| Pattern | Sites | Flexibility | Complexity | When to Use |
|---------|-------|-------------|-----------|-------------|
| **Basic** | 1-5 | Low | Minimal | Learning, prototyping |
| **Configuration** | 10-100+ | High | Low-Med | Most production scenarios |
| **Template** | 5-50 | High | Medium | Groups of similar sites |
| **Multi-Env** | Any | High | Medium | Dev/staging/prod separation |

**Recommendation:** Start with **Configuration-Driven** for most projects. It scales easily and keeps configuration separate from code.

---

## Best Practices

### 1. Naming Conventions

Use descriptive, URL-safe names:
```bash
✅ marketing-site, product-alpha, campaign-q1-2025
❌ site1, site2, mysite
```

### 2. Organization

Structure related sites with prefixes:
```yaml
sites:
  # Marketing sites
  - name: marketing-main
  - name: marketing-secondary
  
  # Product sites
  - name: product-landing-alpha
  - name: product-landing-beta
```

### 3. Redirects and Robots.txt

Define consistent patterns:
```python
# Don't: Different redirect patterns for each site
# Do: Use factory functions or defaults
```

### 4. State Management

Keep Pulumi state secure:
```bash
# Use Pulumi Service backend for teams
pulumi login

# Or store state in S3/Azure/GCS
pulumi org set-default-secrets-provider
```

### 5. Parallel Execution

Pulumi automatically parallelizes resource creation. For best performance:
```bash
# Deploy with explicit parallelism
pulumi up --parallel 32
```

---

## Performance

### State Refresh (NFR2 Requirement)

State refresh operations for 100 managed resources complete in <15 seconds:

```bash
# Refresh state for all sites
pulumi refresh --skip-confirmation

# Measured performance: ~10 seconds for 100 sites
```

### Deployment Time

Typical deployment times with parallel execution:
- 10 sites: ~20 seconds
- 50 sites: ~25 seconds
- 100 sites: ~30 seconds

*Times vary based on network latency and Webflow API response times.*

---

## Troubleshooting

### "Failed to create site X"

Check the error message for details:
```
Error: failed to create site marketing-site: API returned 400 Bad Request
```

Common causes:
- Duplicate site name
- Invalid configuration (empty displayName, etc.)
- API rate limit exceeded

**Solution:** Check site configurations and retry.

### "WEBFLOW_API_TOKEN not found"

Set the environment variable:
```bash
export WEBFLOW_API_TOKEN="your-token-here"
```

### "Which site failed when deploying 100 sites?"

Pulumi clearly identifies failed resources:
```
Error: failed to create site prod-site-47: API returned 409 Conflict
```

The site name tells you exactly which resource failed.

### Performance Issues (slow state refresh)

If state refresh takes >15 seconds:
1. Check network latency to Webflow API
2. Verify API rate limits aren't being hit
3. Consider reducing parallel deployments if you're hitting rate limits

---

## Migration Guide

### From Single-Site to Multi-Site

**Before:**
```typescript
const site = new webflow.Site("my-site", {...});
```

**After (Basic):**
```typescript
const sites = ["site-1", "site-2"].map(name =>
  new webflow.Site(name, {...})
);
```

**After (Configuration):**
```yaml
sites:
  - name: site-1
  - name: site-2
```

### Managing Existing Sites

To import existing Webflow sites into Pulumi:

```bash
# Get site ID from Webflow dashboard
# Then import into Pulumi state
pulumi import webflow:index:Site marketing-site <site-id>
```

---

## Examples Comparison

| Feature | Basic | Config | Template | Multi-Env |
|---------|-------|--------|----------|-----------|
| Configuration | Hardcoded | YAML | Code | Config files |
| Default values | None | Yes | Yes | Per-stack |
| Redirects | Manual | YAML | Built-in | Config/code |
| Robots.txt | Separate | Config | Built-in | Config/code |
| Environment vars | No | No | No | Yes |
| Recommended sites | 3-5 | 10-100+ | 5-50 | Unlimited |

---

## Next Steps

1. **Start small:** Try the Basic example with 2-3 sites
2. **Scale up:** Move to Configuration-Driven for larger fleets
3. **Organize:** Use Template factories for grouped sites
4. **Automate:** Set up multi-environment with Stacks

---

## Getting Help

- **Pulumi docs:** https://www.pulumi.com/docs/
- **Webflow API:** https://docs.webflow.com/
- **Provider docs:** Check the main Webflow provider documentation

---

## Contributing

Have a new pattern or improvement? Contributions welcome! See the main project README for contribution guidelines.

---

**Last updated:** 2025-12-29
