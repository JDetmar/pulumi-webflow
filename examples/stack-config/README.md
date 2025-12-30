# Multi-Environment Stack Configuration Guide

This guide demonstrates how to safely manage multiple environments (dev, staging, production) using Pulumi's stack configuration system with the Webflow provider.

## Table of Contents

1. [Quick Start](#quick-start)
2. [Stack Configuration Concepts](#stack-configuration-concepts)
3. [Credential Management](#credential-management)
4. [Implementation Examples](#implementation-examples)
5. [Stack Promotion Workflow](#stack-promotion-workflow)
6. [State Management](#state-management)
7. [Security Best Practices](#security-best-practices)
8. [Troubleshooting](#troubleshooting)

## Quick Start

### Prerequisites

- Pulumi CLI installed (`pulumi` command available)
- Webflow API token for each environment
- This repository cloned locally
- Node.js, Python, or Go runtime (depending on which example you use)

### Setup Your First Stack (Development)

```bash
# Navigate to an example directory
cd examples/stack-config/typescript-complete

# Initialize a new development stack
pulumi stack init dev

# Configure the API token (encrypted in config)
pulumi config set webflow:apiToken <your-dev-webflow-token> --secret

# The example already has Pulumi.dev.yaml with sites configured
# You can modify the sites in Pulumi.dev.yaml or set environmentName:
pulumi config set environmentName dev

# Verify configuration
pulumi config  # API token shows as [secret]

# Preview changes
pulumi preview

# Deploy to dev
pulumi up
```

> **Note:** Site definitions are configured in the `Pulumi.<stack>.yaml` files as a `sites` object,
> not individual config values. See the example YAML files for the structure.

### Setup Staging Stack

```bash
# Create staging stack
pulumi stack init staging

# Configure staging credentials (different token!)
pulumi config set webflow:apiToken <your-staging-webflow-token> --secret

# The example already has Pulumi.staging.yaml with sites configured
pulumi config set environmentName staging

# Preview and deploy
pulumi preview
pulumi up
```

### Setup Production Stack (With Safety Confirmation)

```bash
# Create production stack
pulumi stack init prod

# Configure production credentials (different token!)
pulumi config set webflow:apiToken <your-prod-webflow-token> --secret

# The example already has Pulumi.prod.yaml with sites configured
pulumi config set environmentName prod

# IMPORTANT: Require safety confirmation to prevent accidents
pulumi config set prodDeploymentConfirmed yes

# Preview changes (ALWAYS review before production)
pulumi preview

# Deploy to production
pulumi up
```

## Stack Configuration Concepts

### What Is a Stack?

A **stack** is an independent instance of your infrastructure. Each stack has:
- Its own configuration file (`Pulumi.<stack>.yaml`)
- Its own state file (stored locally or in Pulumi backend)
- Its own set of resources deployed to your infrastructure

### Stack Files

```
examples/stack-config/typescript-complete/
├── Pulumi.yaml              # Project definition (shared by all stacks)
├── Pulumi.dev.yaml          # Dev stack configuration
├── Pulumi.staging.yaml      # Staging stack configuration
├── Pulumi.prod.yaml         # Production stack configuration
├── index.ts                 # Infrastructure code (same for all stacks)
└── package.json
```

### How Stack Configuration Works

1. **Pulumi.yaml** - Project definition (same for all stacks)
   ```yaml
   name: webflow-multi-environment
   runtime: nodejs
   ```

2. **Pulumi.<stack>.yaml** - Stack-specific configuration
   ```yaml
   config:
     webflow:apiToken:
       secure: AAABAPxyz...  # Encrypted API token
     environmentName: dev
     siteCount: 3
   ```

3. **Code Logic** - Same infrastructure code adapts based on stack config
   ```typescript
   const config = new pulumi.Config();
   const environmentName = config.require("environmentName");  // "dev" or "prod"
   const siteCount = config.requireNumber("siteCount");        // 3, 5, or 10
   ```

### Stack Switching

```bash
# List all stacks
pulumi stack ls

# Switch to a different stack
pulumi stack select dev
pulumi up           # Deploys to dev with dev configuration

# Switch to production
pulumi stack select prod
pulumi up           # Deploys to prod with prod configuration
```

## Credential Management

### Setting Encrypted Credentials Per Stack

Each environment needs its own Webflow API token:

```bash
# Development token
pulumi stack select dev
pulumi config set webflow:apiToken <dev-token> --secret

# Staging token (different from dev!)
pulumi stack select staging
pulumi config set webflow:apiToken <staging-token> --secret

# Production token (different from staging!)
pulumi stack select prod
pulumi config set webflow:apiToken <prod-token> --secret
```

### Why Use `--secret` Flag?

The `--secret` flag encrypts the token using Pulumi's encryption:

```yaml
# With --secret (CORRECT - encrypted)
config:
  webflow:
    apiToken:
      secure: AAABAPxyz...encrypted...

# Without --secret (WRONG - plain text)
config:
  webflow:
    apiToken: "raw-token-here"  # NEVER do this!
```

### Verifying Encrypted Credentials

```bash
# Configuration shows [secret] (correct)
$ pulumi config
KEY               VALUE
webflow:apiToken  [secret]
environmentName   dev

# Never use --show-secrets to display the token
# This prevents accidental token leakage
```

### Token Isolation

Each stack's credentials are completely isolated:

```bash
# Dev stack uses dev token
pulumi stack select dev
pulumi preview  # Uses dev API token

# Switch to prod - prod token is used automatically
pulumi stack select prod
pulumi preview  # Uses prod API token (different!)

# No risk of dev token being used in prod
# No risk of token leakage between stacks
```

## Implementation Examples

This guide includes three complete example implementations:

### TypeScript Example (typescript-complete/)

Best for: Teams using TypeScript/Node.js

**Features:**
- Type-safe stack configuration
- Production safety checks (requires confirmation)
- Environment-specific redirects and robots.txt
- Comprehensive error handling

**Key concepts:**
```typescript
// Production safety check
if (environmentName === "prod") {
    const confirmation = config.get("prodDeploymentConfirmed");
    if (confirmation !== "yes") {
        throw new Error("Production deployment requires confirmation");
    }
}

// Environment-specific resources
if (isProd) {
    new webflow.Redirect(...);  // Only in production
}
```

**Run the example:**
```bash
cd typescript-complete
npm install
pulumi stack init dev
pulumi config set webflow:apiToken <token> --secret
pulumi config set environmentName dev
pulumi config set siteCount 3
pulumi up
```

### Python Example (python-workflow/)

Best for: Teams using Python

**Features:**
- Pythonic configuration loading
- Configuration validation with clear error messages
- Time zone mapping based on region
- Clean, readable infrastructure code

**Key concepts:**
```python
# Validate environment
valid_environments = ["dev", "staging", "prod"]
if environment_name not in valid_environments:
    raise ValueError(f"Invalid environment '{environment_name}'")

# Time zone mapping
timezone_map = {
    "us-west-2": "America/Los_Angeles",
    "us-east-1": "America/New_York",
}
timezone = timezone_map.get(deployment_region)
```

**Run the example:**
```bash
cd python-workflow
pip install -r requirements.txt
pulumi stack init dev
pulumi config set webflow:apiToken <token> --secret
pulumi config set environmentName dev
pulumi config set siteCount 3
pulumi up
```

### Go Example (go-advanced/)

Best for: Teams using Go, advanced patterns

**Features:**
- Idiomatic Go patterns
- Configuration validation using slices
- Advanced error handling
- Production-grade code

**Key concepts:**
```go
// Validate using slices package (idiomatic Go)
validEnvironments := []string{"dev", "staging", "prod"}
if !slices.Contains(validEnvironments, environmentName) {
    return fmt.Errorf("invalid environment '%s'", environmentName)
}

// Timezone map
timezoneMap := map[string]string{
    "us-west-2": "America/Los_Angeles",
    "us-east-1": "America/New_York",
}
```

**Run the example:**
```bash
cd go-advanced
go mod download
pulumi stack init dev
pulumi config set webflow:apiToken <token> --secret
pulumi config set environmentName dev
pulumi config set siteCount 3
pulumi up
```

## Stack Promotion Workflow

This workflow safely promotes infrastructure changes from dev → staging → prod:

### Step 1: Develop in Dev

```bash
# Make code changes
# Edit index.ts or __main__.py

pulumi stack select dev
pulumi preview      # Review changes
pulumi up           # Deploy to dev
```

### Step 2: Test in Staging

```bash
# Code is unchanged, only configuration differs
pulumi stack select staging
pulumi preview      # Preview uses staging config
pulumi up           # Deploy to staging with staging config
```

### Step 3: Deploy to Production

```bash
# Same code, production configuration
pulumi stack select prod

# Always review production changes carefully
pulumi preview

# Production deployment requires explicit confirmation
pulumi up
```

### Key Points

- **Same code** - All stacks run the same infrastructure code
- **Different configuration** - Stack-specific settings control behavior
- **Different credentials** - Each stack uses its own API token
- **Independent state** - Each stack maintains separate state

## State Management

### Stack State Files

Each stack has independent state:

```bash
# Dev stack state (local file)
~/.pulumi/stacks/webflow-multi-environment/dev/

# Staging stack state
~/.pulumi/stacks/webflow-multi-environment/staging/

# Production stack state
~/.pulumi/stacks/webflow-multi-environment/prod/
```

### Inspecting State

```bash
# Export current stack's state
pulumi stack export

# Export specific stack's state
pulumi stack export --stack dev

# Compare stacks
pulumi stack export --stack dev > dev-state.json
pulumi stack export --stack prod > prod-state.json
diff dev-state.json prod-state.json
```

### State Independence

States are completely independent - changes in one stack don't affect others:

```bash
# Modify and delete resources in dev
pulumi stack select dev
pulumi destroy      # Destroys only dev resources

# Prod is unaffected
pulumi stack select prod
pulumi refresh      # Prod resources still exist
```

## Security Best Practices

### ✅ DO (Correct Practices)

```bash
# Always use --secret for API tokens
pulumi config set webflow:apiToken <token> --secret

# Different tokens per environment
pulumi stack select dev
pulumi config set webflow:apiToken <dev-token> --secret

pulumi stack select prod
pulumi config set webflow:apiToken <prod-token> --secret

# Add .gitignore for backup files
echo "*.backup" >> .gitignore
echo "Pulumi.*.yaml.bak" >> .gitignore

# Audit configured credentials per stack
pulumi stack select dev && pulumi config
pulumi stack select staging && pulumi config
pulumi stack select prod && pulumi config
```

### ❌ DON'T (Avoid These)

```bash
# Never commit plain-text tokens
pulumi config set webflow:apiToken <token>  # WRONG - no --secret

# Never use same token across environments
# (Easy to accidentally expose prod credentials if dev is compromised)

# Never commit Pulumi.*.yaml files with plain-text tokens
git add Pulumi.*.yaml  # May contain tokens if not using --secret

# Never use --show-secrets to display tokens
pulumi config get webflow:apiToken --show-secrets
```

### Security Checklist

- [ ] Always use `--secret` flag for API tokens
- [ ] Different API token for each environment
- [ ] Verify tokens are encrypted in Pulumi.*.yaml files
- [ ] Add .gitignore for backup files
- [ ] Audit configured credentials per stack
- [ ] Never commit plain-text secrets
- [ ] Use strong Pulumi passphrases
- [ ] Rotate credentials periodically

## Troubleshooting

### Problem: "Webflow API token not configured"

**Cause:** Token not set for current stack

**Solution:**
```bash
pulumi config set webflow:apiToken <token> --secret
```

### Problem: Wrong environment deployed

**Cause:** Forgot to switch stacks before deploying

**Solution:**
```bash
# Check current stack
pulumi stack ls  # Shows * next to current stack

# Switch to correct stack
pulumi stack select prod

# Verify before deploying
pulumi config  # Verify environmentName
pulumi preview
```

### Problem: "Invalid environment 'staging' (or similar)"

**Cause:** Configuration has typo in environmentName

**Solution:**
```bash
pulumi config set environmentName staging  # Correct value

# Verify
pulumi config  # Should show correct value
```

### Problem: "Production deployment requires confirmation"

**Cause:** `prodDeploymentConfirmed` not set to "yes"

**Solution:**
```bash
pulumi config set prodDeploymentConfirmed yes

# Verify
pulumi config  # Should show prodDeploymentConfirmed: yes
```

### Problem: Sites are missing or wrong per environment

**Cause:** `sites` object in stack config doesn't match expected sites

**Solution:**
Edit the appropriate `Pulumi.<stack>.yaml` file directly to add/modify sites:

```yaml
# In Pulumi.dev.yaml (or staging/prod)
config:
  sites:
    marketing:
      displayName: "Marketing Site"
      shortName: "marketing-dev"
      timeZone: "America/New_York"
      allowIndexing: false
```

Then verify:

```bash
pulumi preview  # Should show correct sites
```

### Problem: Credentials work in dev but not prod

**Cause:** Different tokens not set for prod

**Solution:**
```bash
pulumi stack select prod
pulumi config set webflow:apiToken <prod-specific-token> --secret

# Test connectivity
pulumi preview
```

### Debug: View full configuration

```bash
# Show all config for current stack
pulumi config

# Show stack-specific config file
cat Pulumi.$(pulumi stack).yaml
```

You can also add logging in your code:

```typescript
// TypeScript
console.log("Environment:", environmentName);
console.log("Sites:", Object.keys(sitesConfig));
```

## Advanced Patterns

### Configuration Hierarchy

```typescript
// Code reads config in this order:
// 1. Pulumi.<stack>.yaml (highest priority)
// 2. Pulumi.yaml (default values)
// 3. Fallback in code (lowest priority)

const config = new pulumi.Config();
const region = config.get("deploymentRegion") || "us-west-2";  // Fallback
```

### Environment Detection

```typescript
// Code can adapt based on environment
const isProd = environmentName === "prod";
const isStaging = environmentName === "staging";

if (isProd) {
    // Production-only resources
}
if (isStaging) {
    // Staging-only resources
}
```

### Stack-Aware Resource Naming

```typescript
// Resources automatically include stack name in state
const siteName = `${environmentName}-site-1`;  // "dev-site-1", "prod-site-1"

// Pulumi tracks these separately per stack
// No naming conflicts when managing multiple stacks
```

## Next Steps

1. **Choose an example** - TypeScript, Python, or Go
2. **Initialize first stack** - Set up dev environment
3. **Test stack switching** - Verify isolation works
4. **Promote to staging** - Test promotion workflow
5. **Deploy to production** - With safety checks enabled

## References

- [Pulumi Stack Configuration Documentation](https://www.pulumi.com/docs/iac/concepts/config/)
- [Webflow Provider Documentation](https://www.pulumi.com/registry/packages/webflow/)
- [Best Practices for Managing Secrets](https://www.pulumi.com/docs/iac/concepts/secrets/)
