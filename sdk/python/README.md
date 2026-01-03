# Webflow Pulumi Provider

[![Build Status](https://img.shields.io/github/actions/workflow/status/jdetmar/pulumi-webflow/ci.yml?branch=main)](https://github.com/jdetmar/pulumi-webflow/actions)
[![License: MIT](https://img.shields.io/badge/License-MIT-yellow.svg)](https://opensource.org/licenses/MIT)

> **⚠️ Unofficial Community Provider**
>
> This is an **unofficial, community-maintained** Pulumi provider for Webflow. It is **not affiliated with, endorsed by, or supported by Pulumi Corporation or Webflow, Inc.** This project is an independent effort to bring infrastructure-as-code capabilities to Webflow using Pulumi.
>
> - **Not an official product** - Created and maintained by the community
> - **No warranties** - Provided "as-is" under the MIT License
> - **Community support only** - Issues and questions via [GitHub](https://github.com/jdetmar/pulumi-webflow/issues)

**Manage your Webflow sites and resources as code using Pulumi**

The Webflow Pulumi Provider lets you programmatically manage Webflow resources (sites, redirects, robots.txt, and more) using the same Pulumi infrastructure-as-code approach you use for cloud resources. Deploy, preview, and destroy Webflow infrastructure alongside your other cloud deployments.

## What You Can Do

- **Deploy Webflow resources as code** - Define sites, redirects, robots.txt and other resources in TypeScript, Python, Go, C#, or Java
- **Preview before deploying** - Use `pulumi preview` to see exactly what will change
- **Manage multiple environments** - Create separate stacks for dev, staging, and production
- **Version control your infrastructure** - Track all changes in Git
- **Integrate with CI/CD** - Automate deployments in your GitHub Actions, GitLab CI, or other pipelines

## Table of Contents

1. [Prerequisites](#prerequisites)
2. [Installation](#installation)
3. [Quick Start](#quick-start) ← **Start here** (20 minutes to your first deployment)
4. [Authentication](#authentication)
5. [Verification](#verification)
6. [Troubleshooting](#troubleshooting)
7. [Version Control & Audit Trail](#version-control--audit-trail)
8. [Multi-Language Examples](#multi-language-examples)
9. [Next Steps](#next-steps)
10. [Contributing](#contributing)

---

## Prerequisites

Before you begin, make sure you have:

### 1. **Pulumi CLI**
- Download and install from [pulumi.com/docs/install](https://www.pulumi.com/docs/install/)
- Verify installation: `pulumi version` (requires v3.0 or later)

### 2. **Programming Language Runtime** (choose at least one)
- **TypeScript**: [Node.js](https://nodejs.org/) 14.x or later
- **Python**: [Python](https://www.python.org/downloads/) 3.8 or later
- **Go**: [Go](https://golang.org/dl/) 1.21 or later
- **C#**: [.NET](https://dotnet.microsoft.com/download) 6.0 or later
- **Java**: [Java](https://adoptopenjdk.net/) 11 or later

### 3. **Webflow Account**
- A Webflow account with API access enabled
- Access to at least one Webflow site (where you'll deploy your first resource)

### 4. **Webflow API Token**
- Your Webflow API token (see [Authentication](#authentication) section below)

---

## Installation

The Webflow provider installs automatically when you first run `pulumi up`. For manual installation:

```bash
# Automatic installation (recommended - happens on first pulumi up/preview)
# Just run the Quick Start below, and the provider will install automatically

# OR manual installation if you prefer
pulumi plugin install resource webflow

# Verify installation
pulumi plugin ls | grep webflow
```

---

## Quick Start

### Deploy Your First Webflow Resource in Under 20 Minutes

This quick start walks you through deploying a robots.txt resource to your Webflow site using TypeScript. The entire process takes about 5 minutes once prerequisites are met.

### Step 1: Create a New Pulumi Project (2 minutes)

```bash
# Create a new directory for your Pulumi project
mkdir my-webflow-project
cd my-webflow-project

# Initialize a new Pulumi project
pulumi new --template typescript

# When prompted:
# - Enter a project name: my-webflow-project
# - Enter a stack name: dev
# - Enter a passphrase (or leave empty for no encryption): <press enter>
```

This creates:
- `Pulumi.yaml` - Project configuration
- `Pulumi.dev.yaml` - Stack-specific settings
- `index.ts` - Your infrastructure code
- `package.json` - Node.js dependencies

### Step 2: Configure Webflow Authentication (3 minutes)

```bash
# Get your Webflow API token (see Authentication section below if you don't have one)

# Set your token in Pulumi config (encrypted in Pulumi.dev.yaml)
pulumi config set webflow:apiToken --secret

# When prompted, paste your Webflow API token and press Enter
```

**What's happening:** Your token is encrypted and stored locally in `Pulumi.dev.yaml` (which is in .gitignore). It's never stored in plain text.

### Step 3: Write Your First Resource (5 minutes)

Replace the contents of `index.ts` with:

```typescript
import * as pulumi from "@pulumi/pulumi";
import * as webflow from "pulumi-webflow";

// Get config values
const config = new pulumi.Config();
const siteId = config.requireSecret("siteId"); // We'll set this next

// Deploy a robots.txt resource
const robotsTxt = new webflow.RobotsTxt("my-robots", {
  siteId: siteId,
  content: `User-agent: *
Allow: /

User-agent: Googlebot
Allow: /
`, // Standard robots.txt allowing all crawlers
});

// Export the site ID for reference
export const deployedSiteId = siteId;
```

### Step 4: Configure Your Site ID (2 minutes)

```bash
# Find your Webflow site ID (24-character hex string) from Webflow Designer
# You can find it in: Project Settings > API & Webhooks > Site ID

# Set it in your Pulumi config
pulumi config set siteId --secret

# When prompted, paste your 24-character site ID and press Enter
```

**Need help finding your site ID?**
- In Webflow Designer, go to **Project Settings** (bottom of sidebar)
- Click **API & Webhooks**
- Your **Site ID** is displayed as a 24-character hex string (e.g., `5f0c8c9e1c9d440000e8d8c3`)

### Step 5: Preview Your Deployment (2 minutes)

```bash
# Install dependencies
npm install

# Preview the changes Pulumi will make
pulumi preview
```

Expected output:
```
Previewing update (dev):

     Type                           Name         Plan       Info
 +   webflow:RobotsTxt             my-robots    create

Resources:
    + 1 to create

Do you want to perform this update?
  > yes
    no
    details
```

### Step 6: Deploy! (2 minutes)

```bash
# Deploy to your Webflow site
pulumi up
```

When prompted, select **yes** to confirm the deployment.

Expected output:
```
     Type                           Name         Plan       Status
 +   webflow:RobotsTxt             my-robots    create     created

Outputs:
    deployedSiteId: "5f0c8c9e1c9d440000e8d8c3"

Resources:
    + 1 created

Duration: 3s
```

### Step 7: Verify in Webflow (2 minutes)

1. Open Webflow Designer
2. Go to **Project Settings** → **SEO** → **robots.txt**
3. You should see the robots.txt content you deployed!

### Step 8: Clean Up (Optional)

```bash
# Remove the resource from Webflow
pulumi destroy

# When prompted, select 'yes' to confirm
```

**Congratulations!** You've successfully deployed your first Webflow resource using Pulumi! 🎉

---

## Authentication

### Getting Your Webflow API Token

1. Log in to [Webflow](https://webflow.com)
2. Go to **Account Settings** (bottom left of screen)
3. Click **API Tokens** in the left sidebar
4. Click **Create New Token**
5. Name it something descriptive (e.g., "Pulumi Provider")
6. Grant the following permissions:
   - **Sites**: Read & Write
   - **Redirects**: Read & Write (if using Redirect resources)
   - **Robots.txt**: Read & Write (if using RobotsTxt resources)
7. Click **Create Token**
8. **Copy the token immediately** - Webflow won't show it again

### Setting Up Your Token in Pulumi

```bash
# Option 1: Pulumi config (recommended - encrypted in Pulumi.dev.yaml)
pulumi config set webflow:apiToken --secret

# Option 2: Environment variable
export WEBFLOW_API_TOKEN="your-token-here"

# Option 3: Code (NOT RECOMMENDED for production - security risk)
# Don't do this in production code!
```

### Security Best Practices

- ✅ **DO** use Pulumi config with `--secret` flag (encrypts locally)
- ✅ **DO** use environment variables in CI/CD pipelines
- ✅ **DO** keep tokens in `.env` files (never commit to Git)
- ❌ **DON'T** commit tokens to Git
- ❌ **DON'T** hardcode tokens in your Pulumi program
- ❌ **DON'T** share tokens via email or chat
- 🔐 **Rotate tokens regularly** - Create new tokens and retire old ones monthly

### CI/CD Configuration

For GitHub Actions or other CI/CD:

```yaml
# .github/workflows/deploy.yml
env:
  WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
  PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: pulumi/actions@v4
        with:
          command: up
```

---

## Verification

### Confirm Your Installation

```bash
# Check Pulumi is installed
pulumi version

# Check the Webflow provider is available
pulumi plugin ls | grep webflow

# Should output something like:
# resource webflow 1.0.0-alpha.0+dev
```

### Verify Authentication Works

```bash
# Inside a Pulumi project directory:
pulumi preview

# If authentication fails, you'll see an error like:
# Error: Unauthorized - Check your Webflow API token
```

### After Deployment

1. **In Webflow Designer:**
   - Check that your resource appears in the appropriate settings (robots.txt, redirects, etc.)
   - Verify the configuration matches what you deployed

2. **Via Pulumi:**
   ```bash
   pulumi stack output deployedSiteId
   # Should output your 24-character site ID
   ```

3. **Via Command Line:**
   ```bash
   # View your stack's resources
   pulumi stack select dev
   pulumi stack

   # View detailed resource information
   pulumi stack export | jq .
   ```

---

## Troubleshooting

### Installation Issues

**Problem: "Plugin not found: webflow" error**

```
Error: Plugin webflow not found. Run `pulumi plugin install` to make sure the plugin is installed.
```

**Solution:**
```bash
# Manual plugin installation
pulumi plugin install resource webflow

# Or just run pulumi up - it will install automatically
pulumi up
```

---

### Authentication Issues

**Problem: "Unauthorized" error during pulumi up**

```
Error: Unauthorized - Invalid or expired Webflow API token
```

**Solutions:**
1. Verify your token was set correctly:
   ```bash
   pulumi config get webflow:apiToken --show-secrets
   # Should show your actual token (masked normally, shown with --show-secrets)
   ```

2. Check if token is expired - regenerate it in Webflow:
   - Account Settings → API Tokens
   - Click the refresh icon next to your token
   - Update Pulumi: `pulumi config set webflow:apiToken --secret`

3. Verify token permissions:
   - Go to Webflow Account Settings → API Tokens
   - Check that the token has "Sites: Read & Write" permission

---

### Configuration Issues

**Problem: "Invalid site ID" error**

```
Error: Invalid or malformed siteId. Must be a 24-character hex string.
```

**Solutions:**
1. Get the correct site ID:
   - Open Webflow Designer
   - Go to Project Settings → API & Webhooks
   - Copy your Site ID (24-character hex string)

2. Update your config:
   ```bash
   pulumi config set siteId --secret
   # Paste your correct site ID
   ```

3. Verify it was set:
   ```bash
   pulumi config get siteId --show-secrets
   ```

---

### Network Issues

**Problem: "Connection timeout" or "Connection refused"**

**Solutions:**
1. Check your internet connection
2. Check if Webflow API is accessible:
   ```bash
   curl -H "Authorization: Bearer YOUR_TOKEN" https://api.webflow.com/v2/sites
   ```
3. Check for corporate firewalls or proxy issues
4. Try again - may be temporary API issue

---

### Get Help

- 📖 **Full Troubleshooting Guide**: [docs/troubleshooting.md](./docs/troubleshooting.md) - Comprehensive error reference and diagnostic procedures
- ❓ **FAQ**: [docs/faq.md](./docs/faq.md) - Answers to common questions
- 📚 **Full Documentation**: [docs/](./docs/)
- 🔧 **Examples**: Check [examples/](./examples/) for comprehensive examples
- 🐛 **Report Bugs**: [GitHub Issues](https://github.com/jdetmar/pulumi-webflow/issues)
- 💬 **Ask Questions**: [GitHub Discussions](https://github.com/jdetmar/pulumi-webflow/discussions)

---

## Version Control & Audit Trail

All infrastructure changes can be tracked in Git to create an immutable audit trail for compliance.

### Key Features

- ✅ **Automatic Audit Trail** - Every change tracked with author, timestamp, and reason
- ✅ **Code Review Process** - Pull requests enable peer review before deployment
- ✅ **Compliance Ready** - Generate audit reports for SOC 2, HIPAA, GDPR compliance
- ✅ **CI/CD Integration** - GitHub Actions validates changes before merge
- ✅ **Multi-Environment** - Separate tracking for dev, staging, production

### Quick Example

```bash
# Create a feature branch for your change
git checkout -b feat/add-redirects

# Make infrastructure changes
# ... update your Pulumi code ...

# Preview the changes
pulumi preview

# Commit with a clear message
git commit -m "feat(redirects): add GDPR-compliant redirect rules

- Redirect /privacy to privacy-policy.webflow.io
- Redirect /terms to terms-of-service.webflow.io
- Requires approval before deployment

Story: 2.2 - Redirect CRUD Operations"

# Push and create a Pull Request for code review
git push origin feat/add-redirects
```

### Compliance Documentation

- 📖 **[Version Control Integration Guide](./docs/version-control.md)** - Complete Git workflow guide with best practices

### Generating Audit Reports

```bash
# View recent changes
git log --oneline -- Pulumi.*.yaml

# Generate detailed audit report for a time period
git log --since="2025-12-01" --until="2025-12-31" \
  --format="%ai | %an | %s" \
  -- Pulumi.*.yaml

# Export audit trail for compliance review
git log -p -- Pulumi.*.yaml > audit-report.txt
```

See the [Version Control guide](./docs/version-control.md) for complete instructions on setting up Git workflows and compliance reporting.

---

## Multi-Language Examples

The quickstart above uses TypeScript. Complete examples for other languages are in the [examples/quickstart/](./examples/quickstart/) directory:

- **TypeScript**: [examples/quickstart/typescript/](./examples/quickstart/typescript/) - Recommended for quick onboarding
- **Python**: [examples/quickstart/python/](./examples/quickstart/python/) - Pythonic approach
- **Go**: [examples/quickstart/go/](./examples/quickstart/go/) - High-performance deployment

Each example includes:
- Complete, copy-pasteable code
- Language-specific setup instructions
- Pulumi.yaml configuration
- Package manager files (package.json, requirements.txt, go.mod)
- Comprehensive README

---

## Next Steps

Once you've completed the Quick Start:

### 1. **Explore More Resources**
- Deploy multiple resource types (Redirects, Sites, etc.)
- Use the [examples/](./examples/) directory for real-world patterns
- Check [docs/](./docs/) for comprehensive reference documentation

### 2. **Multi-Environment Setup**
- Create separate stacks for dev, staging, and production
- Use different site IDs for each environment
- See: [examples/stack-config/](./examples/stack-config/)

### 3. **Advanced Patterns**
- Multi-site management: [examples/multi-site/](./examples/multi-site/)
- CI/CD integration: [examples/ci-cd/](./examples/ci-cd/)
- Logging and troubleshooting: [examples/troubleshooting-logs/](./examples/troubleshooting-logs/)

### 4. **Learn Pulumi Concepts**
- [Pulumi Documentation](https://www.pulumi.com/docs/)
- [Getting Started with Pulumi](https://www.pulumi.com/docs/iac/getting-started/)
- [Pulumi Best Practices](https://www.pulumi.com/docs/using-pulumi/best-practices/)

### 5. **Connect with the Community**
- [Pulumi Community Slack](https://pulumi-community.slack.com/)
- [Pulumi GitHub Discussions](https://github.com/pulumi/pulumi/discussions)

---

## Contributing

We welcome contributions! Please see [CONTRIBUTING.md](./CONTRIBUTING.md) for guidelines.

### Ways to Contribute

- **Report bugs** - Found an issue? [Create a GitHub issue](https://github.com/jdetmar/pulumi-webflow/issues)
- **Submit improvements** - Have an idea? [Create a discussion](https://github.com/jdetmar/pulumi-webflow/discussions)
- **Contribute code** - Fork the repo, make changes, and submit a pull request
- **Improve documentation** - Help us document features and patterns

---

## License

This project is licensed under the MIT License - see [LICENSE](./LICENSE) file for details.

---

## Troubleshooting Quick Reference

| Problem | Solution |
|---------|----------|
| Plugin not found | `pulumi plugin install resource webflow` |
| Invalid token | Check Webflow Account Settings → API Tokens |
| Invalid site ID | Verify in Webflow Designer → Project Settings → API & Webhooks |
| Deployment times out | Check internet connection, try again |
| Token format error | Ensure you're using the full API token (not just a prefix) |
| Site not found | Verify site ID matches the site where you want to deploy |

---

**Ready to get started?** Jump to [Quick Start](#quick-start) above! 🚀
