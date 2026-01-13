# Frequently Asked Questions (FAQ)

Quick answers to common questions about the Webflow Pulumi Provider.

## Table of Contents

1. [Getting Started](#getting-started)
2. [Authentication](#authentication)
3. [Configuration](#configuration)
4. [State Management](#state-management)
5. [Resources](#resources)
6. [Multi-Site Management](#multi-site-management)
7. [CI/CD Integration](#cicd-integration)
8. [Performance](#performance)
9. [Security](#security)
10. [Troubleshooting](#troubleshooting)

## Getting Started

### What is the Webflow Pulumi Provider?

The Webflow Pulumi Provider enables you to manage your Webflow sites and resources using Infrastructure as Code with Pulumi. You can define your Webflow resources in code (TypeScript, Python, Go, .NET, or Java), version control them, and deploy with `pulumi up`.

**Resources managed:**
- **Sites** - Create, update, and delete Webflow sites
- **Redirects** - Manage URL redirects
- **RobotsTxt** - Manage robots.txt files

### How do I get started?

1. **Install the provider:**
   ```bash
   pulumi plugin install resource webflow v0.1.0
   ```

2. **Create a new Pulumi project:**
   ```bash
   pulumi new
   ```

3. **Install SDK:**
   ```bash
   npm install @jdetmar/pulumi-webflow  # TypeScript/JavaScript
   pip install pulumi-webflow   # Python
   ```

4. **Set API token:**
   ```bash
   pulumi config set webflow:apiToken --secret
   ```

5. **Create your first resource:**
   ```python
   import pulumi
   import pulumi_webflow as webflow

   site = webflow.Site("my-site",
       site_id="your_site_id_here"
   )
   ```

6. **Deploy:**
   ```bash
   pulumi up
   ```

### Which languages are supported?

The provider supports:
- **TypeScript/JavaScript** - Native support via @jdetmar/pulumi-webflow
- **Python** - Via pulumi-webflow package
- **Go** - Via pulumi-webflow Go SDK
- **.NET** - Via Pulumi.Webflow package
- **Java** - Via pulumi-webflow Maven package

All languages have the same capabilities; choose what's most comfortable for you.

### What's included in the provider?

The provider includes:
- **Provider Plugin** - Binary that manages resources
- **Language SDKs** - Bindings for TypeScript, Python, Go, .NET, Java
- **Documentation** - API reference, examples, guides
- **Examples** - Real-world usage patterns
- **CLI Integration** - Works seamlessly with `pulumi` command

### Is there an example project I can start from?

Yes! Check the `examples/` directory in the GitHub repository:
- `examples/basic-site/` - Simple site management
- `examples/multi-site/` - Managing multiple sites
- `examples/with-redirects/` - Site with URL redirects
- `examples/multi-environment/` - Dev, staging, production stacks
- `examples/troubleshooting-logs/` - Logging and debugging examples

Each example includes a README with setup instructions.

### How much does this cost?

The provider itself is free (open source).

**Costs depend on:**
- Your Webflow plan (Designer, Team, Agency, Enterprise)
- AWS/Pulumi Cloud if using managed backend
- Infrastructure where you run deployments (typically minimal)

The provider doesn't add licensing fees on top of your Webflow subscription.

## Authentication

### How do I get my Webflow API token?

1. Go to Webflow Dashboard
2. Click your account icon (top right)
3. Select "Account Settings"
4. Navigate to "API & Webhooks"
5. Click "Generate API Token" or copy existing token
6. Copy the token (shown only once)
7. Save it securely

**Important:** The token is only shown once. If you lose it, generate a new one.

### Where should I store my API token?

**Don't:**
- Commit to version control
- Put in code files
- Share via email
- Log to console or files

**Do:**
- Store in environment variables
- Use `--secret` flag with Pulumi config
- Use Pulumi Cloud for secrets management
- Use CI/CD secrets management (GitHub Actions, etc.)

### How do I configure the API token in Pulumi?

**Option 1: Environment Variable**
```bash
export WEBFLOW_API_TOKEN="your_token_here"
pulumi up
```

**Option 2: Pulumi Config (Recommended)**
```bash
pulumi config set webflow:apiToken --secret
# Enter token when prompted
```

**Option 3: CI/CD Secrets**
```yaml
# GitHub Actions
env:
  WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
```

### Can I use different tokens for different stacks?

Yes! Set different tokens per stack:

```bash
# For development stack
pulumi stack select dev
pulumi config set webflow:apiToken --secret

# For production stack
pulumi stack select production
pulumi config set webflow:apiToken --secret
```

Each stack has its own secrets.

### What permissions does the API token need?

The token needs access to:
- **Sites** - View and manage your sites
- **Collections** - If managing collections (for future features)

Tokens can be restricted to specific sites. Use the Webflow Dashboard to configure granular permissions.

### What if my token expires or is compromised?

1. **Immediately regenerate:** Webflow Dashboard → Account Settings → API & Webhooks
2. **Delete old token** - Invalidates it immediately
3. **Update Pulumi config:**
   ```bash
   pulumi config set webflow:apiToken --secret
   ```
4. **Monitor Webflow** - Check for unauthorized activity in your dashboard

### Can I rotate my API token?

Yes, it's recommended to rotate periodically:

1. Generate new token (Webflow Dashboard → API & Webhooks)
2. Update Pulumi config: `pulumi config set webflow:apiToken --secret`
3. Verify deployment works with new token
4. Delete old token

For CI/CD, update secrets in your platform (GitHub, GitLab, etc.) simultaneously.

## Configuration

### What is a site ID?

A **site ID** is the unique identifier for your Webflow site. It's a 24-character hexadecimal string that looks like: `507f1f77bcf86cd799439011`

### How do I find my site ID?

1. Go to Webflow Designer
2. Click your site name (top left)
3. Select "Project Settings"
4. Go to "API & Webhooks"
5. Copy your "Site ID" (24-character string)

### Do I need a site ID for each resource?

Yes. Each resource (Site, Redirect, RobotsTxt) is associated with a site ID.

```python
# Redirect for a specific site
redirect = webflow.Redirect("my-redirect",
    site_id="507f1f77bcf86cd799439011",  # Required
    source="/old-page",
    target="/new-page"
)
```

### Can I use the same site ID for multiple stacks?

**No.** Each stack should manage a different site or different collections.

**Bad:**
```
Dev stack → Site A
Prod stack → Site A  # Don't do this!
```

**Good:**
```
Dev stack → Site A
Prod stack → Site B
```

**If you need multiple stacks on one site:**
- Use different resource names: `redirect-v1`, `redirect-v2`
- Document which stack manages which resources
- Implement locks to prevent simultaneous deployments

### What should my stack structure be?

Common structure:

```
myproject/
├── Pulumi.yaml              # Project file
├── Pulumi.dev.yaml          # Dev stack config
├── Pulumi.staging.yaml      # Staging stack config
├── Pulumi.prod.yaml         # Production stack config
└── __main__.py              # Your code
```

Each stack has its own:
- Site ID (different Webflow site)
- API token (different Webflow account or restricted token)
- Resource definitions

### Can I have multiple projects using the same API token?

Yes, but **not recommended** for security:

**Better approach:**
- Create separate Webflow accounts per environment
- Use different API tokens per project
- Restrict token scope to necessary sites

### How do I organize resources across teams?

For team/multi-tenant setups:

1. **Separate Webflow accounts** - One per team/tenant
2. **Separate API tokens** - Different token per account
3. **Separate Pulumi stacks** - One stack per team
4. **Pulumi organization** - If using Pulumi Cloud

### What's the naming convention for resources?

Use descriptive names for your Pulumi resources:

```python
# Good
site = webflow.Site("production-site", site_id="...")
redirect = webflow.Redirect("old-domain-redirect", site_id="...")
robots = webflow.RobotsTxt("disallow-bots", site_id="...")

# Avoid
site1 = webflow.Site("s1", site_id="...")
r = webflow.Redirect("r", site_id="...")
```

Names should be:
- Descriptive
- Lowercase with hyphens
- Unique within your stack
- Memorable for your team

## State Management

### What is Pulumi state?

Pulumi state is the record of:
- What resources exist in Webflow
- Their properties and configuration
- How your code maps to those resources

State is stored in:
- **Local file** - Default (`.pulumi/stacks/`)
- **Pulumi Cloud** - Recommended for teams
- **Custom backend** - Azure Blob Storage, S3, etc.

### How do I check my state?

```bash
# Show state summary
pulumi stack

# Export full state (may contain secrets!)
pulumi stack export

# Show without secrets
pulumi state export
```

### When should I run `pulumi refresh`?

Run `pulumi refresh` to:
1. **Detect drift** - After manual changes in Webflow Designer
2. **Update state** - After external modifications
3. **Troubleshoot** - To verify resources actually exist

```bash
pulumi refresh
```

### What is state drift?

**Drift** = difference between:
- **Actual state** (what exists in Webflow)
- **Desired state** (what your Pulumi code says should exist)

**Examples:**
- You manually changed a redirect in Webflow Designer
- Another team member deployed different changes
- External system modified a resource

**Detect drift:**
```bash
pulumi refresh
```

**Handle drift:**
```bash
pulumi refresh        # Accept actual state
# or
pulumi up --force     # Overwrite with desired state
```

### How do I import existing Webflow resources?

Use `pulumi import` to add existing resources to state:

```bash
# Import an existing site
pulumi import webflow:index:Site my-imported-site site-id-123abc

# Import a redirect
pulumi import webflow:index:Redirect my-redirect redirect-id-456def
```

Then add the resource to your code:

```python
site = webflow.Site("my-imported-site",
    site_id="507f1f77bcf86cd799439011"
)
```

### What if I delete resources manually in Webflow?

If you delete resources directly in Webflow Dashboard:

1. Run `pulumi refresh` to detect
2. Your state will be out of sync
3. Next `pulumi up` will recreate the resource

**To avoid:**
- Only modify resources through Pulumi
- Use `pulumi destroy` when deleting
- Restrict direct Webflow Designer access for managed sites

### How do I delete resources?

1. **Remove from code:**
   ```python
   # Delete this line
   # redirect = webflow.Redirect("my-redirect", site_id="...")
   ```

2. **Deploy:**
   ```bash
   pulumi up
   ```
   Pulumi will detect and delete the resource.

3. **Or use destroy:**
   ```bash
   pulumi destroy
   ```
   Destroys all resources in the stack.

### Can I rename a resource in Pulumi?

**Simple rename in code:**
```python
# Before
site = webflow.Site("old-name", site_id="...")

# After
site = webflow.Site("new-name", site_id="...")
```

Then:
```bash
pulumi up
```

Pulumi will see it as:
- Deleting "old-name"
- Creating "new-name"

The actual resource in Webflow stays the same, but state is recreated.

**To keep existing state:**
```bash
pulumi state mv urn:pulumi:...:old-name urn:pulumi:...:new-name
```

## Resources

### What resource types are available?

Currently supported:
- **Site** - Webflow site management
- **Redirect** - URL redirects
- **RobotsTxt** - robots.txt management

More resources coming in future versions.

### How do I create a Site resource?

```python
import pulumi
import pulumi_webflow as webflow

site = webflow.Site("my-site",
    site_id="507f1f77bcf86cd799439011",  # Required
    display_name="My Webflow Site"        # Optional
)

# Export outputs
pulumi.export("site_id", site.site_id)
```

### How do I create a Redirect resource?

```python
import pulumi
import pulumi_webflow as webflow

redirect = webflow.Redirect("old-to-new",
    site_id="507f1f77bcf86cd799439011",  # Required
    source="/old-page",                   # Required
    target="/new-page"                    # Required
)

# Redirect with permanent status (301)
redirect_301 = webflow.Redirect("old-domain",
    site_id="507f1f77bcf86cd799439011",
    source="/",
    target="https://new-domain.com"
)
```

### How do I manage robots.txt?

```python
import pulumi
import pulumi_webflow as webflow

robots = webflow.RobotsTxt("my-robots",
    site_id="507f1f77bcf86cd799439011",
    content="""User-agent: *
Disallow: /admin/
Disallow: /private/
Allow: /public/"""
)
```

### Can I use outputs from one resource in another?

Yes! Outputs are values exposed by resources:

```python
import pulumi
import pulumi_webflow as webflow

site = webflow.Site("my-site",
    site_id="507f1f77bcf86cd799439011"
)

# Use site ID in redirect
redirect = webflow.Redirect("my-redirect",
    site_id=site.site_id,  # Use output from site
    source="/old",
    target="/new"
)

# Export for later use
pulumi.export("site_id", site.site_id)
```

### How do I reference resources from other projects?

Use stack references:

```python
import pulumi

# Reference another project's stack
infra = pulumi.StackReference(f"organization/project/stack")

# Get outputs
site_id = infra.get_output("site_id")

# Use in your resources
redirect = webflow.Redirect("my-redirect",
    site_id=site_id,
    source="/old",
    target="/new"
)
```

## Multi-Site Management

### How do I manage multiple Webflow sites?

**Option 1: Multiple resources in one stack:**
```python
site1 = webflow.Site("site-1", site_id="id-1")
site2 = webflow.Site("site-2", site_id="id-2")
site3 = webflow.Site("site-3", site_id="id-3")
```

**Option 2: Separate stacks per site:**
```
project/
├── Pulumi.dev.yaml       # Dev site
├── Pulumi.staging.yaml   # Staging site
└── Pulumi.prod.yaml      # Production site
```

**Option 3: Separate projects per site:**
```
webflow-org/
├── site-1-project/
├── site-2-project/
└── site-3-project/
```

### What's the best way to manage fleets of sites?

For 10+ sites, structure like:

1. **Shared configuration:**
   ```python
   # config.py
   SITES = {
       "client-a": "507f1f77bcf86cd799439011",
       "client-b": "507f1f77bcf86cd799439012",
       "client-c": "507f1f77bcf86cd799439013",
   }
   ```

2. **Loop through sites:**
   ```python
   for client_name, site_id in SITES.items():
       site = webflow.Site(f"site-{client_name}",
           site_id=site_id
       )
   ```

3. **Organize stacks:**
   ```
   Pulumi.dev.yaml        # All sites in dev
   Pulumi.prod.yaml       # All sites in prod
   ```

### How do I manage site naming conventions?

Use clear, descriptive names:

```python
# Good: Clear what each site is
site_client_a_production = webflow.Site("client-a-prod", site_id="...")
site_client_b_staging = webflow.Site("client-b-staging", site_id="...")

# Avoid: Ambiguous
site1 = webflow.Site("site-1", site_id="...")
site2 = webflow.Site("site-2", site_id="...")
```

Document mappings:
```python
"""
Site mappings for our deployment:
- client-a-prod: https://client-a.com (site ID: ...)
- client-b-staging: https://staging.client-b.com (site ID: ...)
"""
```

### Can I manage sites across different Webflow accounts?

Yes, using multiple API tokens:

```python
import pulumi

# Get tokens from config
token1 = pulumi.Config().require_secret("webflow_token_account1")
token2 = pulumi.Config().require_secret("webflow_token_account2")

# Configure per stack
# Pulumi.stack1.yaml: webflow:apiToken = token1
# Pulumi.stack2.yaml: webflow:apiToken = token2

# Stacks manage different Webflow accounts
```

### How do I handle naming conflicts across multiple sites?

For sites managed in the same stack, use prefixes:

```python
# Group by site
site_a_redirect = webflow.Redirect("site-a-old-domain",
    site_id="site-a-id",
    source="/",
    target="https://site-a.new.com"
)

site_b_redirect = webflow.Redirect("site-b-old-domain",
    site_id="site-b-id",
    source="/",
    target="https://site-b.new.com"
)
```

Or organize by type:

```python
# All redirects together
redirects = {
    "site-a": webflow.Redirect(...),
    "site-b": webflow.Redirect(...),
}

# All robots files together
robots = {
    "site-a": webflow.RobotsTxt(...),
    "site-b": webflow.RobotsTxt(...),
}
```

## CI/CD Integration

### How do I deploy from GitHub Actions?

1. **Set up secrets:**
   ```
   Settings → Secrets and variables → Actions
   Add: PULUMI_ACCESS_TOKEN
   Add: WEBFLOW_API_TOKEN
   ```

2. **Create workflow file:**
   ```yaml
   name: Deploy
   on: [push]
   jobs:
     deploy:
       runs-on: ubuntu-latest
       steps:
         - uses: actions/checkout@v3

         - uses: actions/setup-python@v4
           with:
             python-version: 3.9

         - uses: pulumi/actions@v4
           with:
             command: up
           env:
             PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
             WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
   ```

3. **Customize for your language:**
   - TypeScript: Use `setup-node` instead
   - Go: Use `setup-go`
   - .NET: Use `setup-dotnet`

### How do I deploy from GitLab CI?

```yaml
stages:
  - deploy

deploy:
  stage: deploy
  image: python:3.9
  script:
    - pip install pulumi pulumi-webflow
    - pulumi stack select production
    - pulumi up --yes
  env:
    PULUMI_ACCESS_TOKEN: $PULUMI_ACCESS_TOKEN
    WEBFLOW_API_TOKEN: $WEBFLOW_API_TOKEN
```

### How do I deploy to multiple stacks in CI/CD?

Use a matrix strategy (GitHub Actions):

```yaml
strategy:
  matrix:
    stack: [dev, staging, production]

steps:
  - uses: pulumi/actions@v4
    with:
      command: up
      stack-name: ${{ matrix.stack }}
    env:
      PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
      WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
```

Or use loops (GitLab CI):

```yaml
script:
  - for stack in dev staging production; do
      pulumi stack select $stack
      pulumi up --yes
    done
```

### How do I prevent accidental production deployments?

Use branch protection:

1. **GitHub:** Settings → Branches → Branch protection rules
   - Require reviews for production deployments
   - Limit who can push to main

2. **In CI/CD:**
   ```yaml
   if: github.ref == 'refs/heads/main'
   ```

3. **In Pulumi:**
   ```python
   import pulumi

   stack_name = pulumi.get_stack()
   if stack_name == "production":
       # Add extra warnings
       print("⚠️ Deploying to PRODUCTION")
   ```

### How do I prevent concurrent deployments?

Use concurrency locks:

```yaml
concurrency:
  group: deployment-${{ matrix.stack }}
  cancel-in-progress: false
```

Or in Pulumi:
```bash
pulumi up --lock-file=deploy.lock
```

### How do I handle approvals in CI/CD?

```yaml
deploy:
  runs-on: ubuntu-latest
  environment:
    name: production
    # GitHub automatically requires approval
  steps:
    - uses: pulumi/actions@v4
      with:
        command: up
```

### What environment variables should I set?

Essential:
```yaml
PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
PULUMI_SKIP_UPDATE_CHECK: true
```

Optional:
```yaml
PULUMI_DEBUG: false          # Enable for debugging
LOGLEVEL: info               # Log level
TZ: America/Chicago          # Timezone for timestamps
```

## Performance

### How long does a deployment take?

Typical times:
- **Small deployments** (1-5 resources): 10-30 seconds
- **Medium deployments** (5-20 resources): 30-60 seconds
- **Large deployments** (20+ resources): 1-5 minutes

Factors affecting speed:
- Number of resources
- Webflow API response time
- Network latency
- Resource type complexity

### How can I speed up deployments?

1. **Reduce parallelism for stability:**
   ```bash
   pulumi up --parallelism=1
   ```

2. **For faster, use parallelism (if stable):**
   ```bash
   pulumi up --parallelism=5
   ```

3. **Skip preview:**
   ```bash
   pulumi up --skip-preview
   ```

4. **Optimize resource count:**
   - Only manage necessary resources with Pulumi
   - Use Webflow UI for resources you don't need to version control

### Is there a rate limit?

Webflow API has rate limits:
- **API calls:** Limited per second (exact limit not published)
- **Rate limit:** Returned as HTTP 429 error

**If you hit limits:**
```bash
# Reduce parallelism
pulumi up --parallelism=1

# Add delays between operations
# (implement in code if needed)

# Reduce resource count
```

### How can I optimize large deployments?

For 50+ resources:

1. **Split into multiple stacks:**
   ```python
   # stack1.py: Resources 1-25
   # stack2.py: Resources 26-50
   ```

2. **Use stack references:**
   ```python
   # Deploy stack1 first, then stack2
   stack1_outputs = pulumi.StackReference(...)
   ```

3. **Add explicit dependencies:**
   ```python
   redirect = webflow.Redirect(...,
       opts=pulumi.ResourceOptions(depends_on=[site])
   )
   ```

4. **Deploy in stages:**
   ```bash
   pulumi up --refresh  # Verify state
   pulumi preview       # Check changes
   pulumi up           # Deploy
   ```

## Security

### How do I keep my API tokens secure?

**Best Practices:**

1. **Never commit tokens to version control:**
   ```bash
   # Use Pulumi secrets
   pulumi config set webflow:apiToken --secret
   ```

2. **Use environment variables in CI/CD:**
   ```yaml
   env:
     WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
   ```

3. **Rotate tokens regularly:**
   - Generate new token in Webflow Dashboard
   - Update Pulumi config and CI/CD secrets
   - Delete old token immediately

4. **Use minimal permissions:**
   - Create tokens with only required scopes
   - Separate tokens for dev/staging/production

### What happens if my token is compromised?

**Immediate Actions:**

1. **Revoke the token immediately:**
   - Webflow Dashboard → Account Settings → API & Webhooks
   - Delete the compromised token

2. **Generate a new token:**
   - Create new token with same permissions
   - Update all configurations

3. **Audit for unauthorized access:**
   - Check Webflow Dashboard for unexpected changes
   - Review deployment logs
   - Check site content for modifications

4. **Update all references:**
   ```bash
   pulumi config set webflow:apiToken --secret
   # Update CI/CD secrets
   ```

### How do I prevent credential leakage in logs?

The provider automatically redacts sensitive data, but follow these practices:

1. **Always use `--secret` flag:**
   ```bash
   pulumi config set webflow:apiToken --secret
   ```

2. **Never print tokens in code:**
   ```python
   # BAD - Never do this
   print(f"Token: {api_token}")

   # GOOD - Just use the token
   provider = webflow.Provider("provider", api_token=api_token)
   ```

3. **Review logs before sharing:**
   ```bash
   # Check for leaked secrets
   grep -i "token\|secret\|password" deployment.log
   ```

4. **Use Pulumi's secret handling:**
   - Secrets are encrypted in state
   - Not displayed in console output
   - Redacted in stack exports

### How do I manage secrets across environments?

**Per-Environment Secrets:**

```bash
# Development
pulumi stack select dev
pulumi config set webflow:apiToken --secret

# Staging
pulumi stack select staging
pulumi config set webflow:apiToken --secret

# Production
pulumi stack select production
pulumi config set webflow:apiToken --secret
```

Each stack has isolated secrets - they don't cross over.

### What security features does the provider include?

**Built-in Security:**

- **TLS 1.2+ enforced** - All API calls use secure HTTPS
- **Token redaction** - Tokens are never logged
- **Secret encryption** - Pulumi encrypts sensitive config values
- **No credential storage** - Tokens passed at runtime, not stored in provider

**Error Code Reference:**

- `WEBFLOW_AUTH_001` - Token not configured
- `WEBFLOW_AUTH_002` - Token empty
- `WEBFLOW_AUTH_003` - Token format invalid

See [Troubleshooting Guide](./troubleshooting.md#authentication--credentials) for details.

## Troubleshooting

### Where do I find help?

1. **Check this FAQ** - Most questions answered here
2. **See Troubleshooting Guide** - For error-specific help (docs/troubleshooting.md)
3. **GitHub Issues** - https://github.com/jdetmar/pulumi-webflow/issues
4. **GitHub Discussions** - https://github.com/jdetmar/pulumi-webflow/discussions
5. **Webflow Support** - For Webflow-specific issues

### How do I enable verbose logging?

```bash
pulumi up --debug 2>&1 | tee deployment.log
```

Check `deployment.log` for detailed error messages.

### How do I report a bug?

1. **Search existing issues** - Might be already reported
2. **Create new issue:**
   - Describe what happened
   - Include error output
   - Share `pulumi version` output
   - Share minimal code to reproduce
   - **Don't include credentials!**

3. **Example issue:**
   ```
   Title: Site creation fails with "invalid site ID"

   Error:
   Error creating resource: invalid site ID 'abc': must be 24-character hex

   Steps to reproduce:
   1. Create site with site_id="abc"
   2. Run `pulumi up`

   Expected: Site should be created
   Actual: Error about invalid format

   Environment:
   - pulumi --version
   - pulumi plugin ls | grep webflow
   - OS: macOS 13.0
   ```

### What information should I include in bug reports?

Always include:
- **Error message** - Full output
- **Steps to reproduce** - Minimal example
- **Environment:**
  ```bash
  pulumi version
  pulumi plugin ls
  <language> --version
  uname -a
  ```
- **Config (no secrets):**
  ```bash
  pulumi config
  ```

Helpful:
- **Debug output:**
  ```bash
  pulumi up --debug 2>&1 | tail -100
  ```
- **Minimal code example**
- **Expected vs. actual behavior**

### How do I get support?

**For Pulumi provider issues:**
- GitHub Issues: https://github.com/jdetmar/pulumi-webflow/issues
- GitHub Discussions: https://github.com/jdetmar/pulumi-webflow/discussions

**For Webflow API issues:**
- Webflow Support: https://webflow.com/support
- Webflow API Docs: https://developers.webflow.com

**For Pulumi general questions:**
- Pulumi Community: https://www.pulumi.com/community/
- Pulumi Docs: https://www.pulumi.com/docs/

### What's the best way to report a security issue?

**Don't** open a public GitHub issue for security vulnerabilities.

Instead:
1. Email security@webflow.com if Webflow API related
2. Email info@pulumi.com if Pulumi platform related
3. Check SECURITY.md for responsible disclosure process

### How do I stay updated on new features?

1. **Watch releases:** GitHub → Releases tab
2. **Join discussions:** GitHub → Discussions tab
3. **Follow on social media:**
   - Pulumi: @PulumiCorp
   - Webflow: @webflow

---

## Still have questions?

- **Search this FAQ** with Ctrl+F
- **Check Troubleshooting Guide** - docs/troubleshooting.md
- **Review examples** - examples/ directory
- **Open a GitHub issue** - https://github.com/jdetmar/pulumi-webflow/issues
