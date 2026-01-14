# Troubleshooting Guide

This guide helps you resolve common issues when using the Webflow Pulumi Provider. Find your error in the sections below, or use the quick reference table to locate solutions.

## Quick Reference

| Error | Category | Solution |
|-------|----------|----------|
| `Plugin not found` | Installation | Install provider plugin: `pulumi plugin install resource webflow` |
| `API token not configured` | Authentication | Set `WEBFLOW_API_TOKEN` environment variable or configure in stack |
| `invalid site ID` | Configuration | Get your site ID from Webflow Designer → Project Settings → API & Webhooks |
| `Site not found` | Configuration | Verify site ID is correct and account has access |
| `Connection timeout` | Network | Check network connectivity, increase timeout, verify Webflow API availability |
| `Rate limit exceeded (429)` | Network | Wait before retrying, reduce parallelism with `parallelism=1` |
| `Resource creation failed` | Runtime | Check error message for specific cause, review logs with verbose mode |
| `State drift detected` | State Management | Run `pulumi refresh` to update state, then investigate actual vs. desired |
| `Import failed` | State Management | Verify resource exists in Webflow, check resource type is correct |
| `Non-interactive mode fails` | CI/CD | Ensure `PULUMI_SKIP_UPDATE_CHECK=true` in pipeline |

## Table of Contents

1. [Installation & Setup](#installation--setup)
2. [Authentication & Credentials](#authentication--credentials)
3. [Configuration](#configuration)
4. [Runtime Errors](#runtime-errors)
5. [Resource-Specific Issues](#resource-specific-issues)
6. [Network & API Issues](#network--api-issues)
7. [State Management](#state-management)
8. [Multi-Environment](#multi-environment)
9. [CI/CD Integration](#cicd-integration)
10. [Diagnostic Procedures](#diagnostic-procedures)
11. [Logging & Debugging](#logging--debugging)

## Installation & Setup

### Plugin Installation Failures

**Error:** `Plugin not found in PATH`

**Cause:** The Webflow Pulumi plugin is not installed or not in your PATH.

**Solution:**
1. Install the plugin:
   ```bash
   pulumi plugin install resource webflow v0.1.0
   ```
2. Verify installation:
   ```bash
   pulumi plugin ls
   ```
   You should see `webflow` listed in the resource plugins.

3. If still not found, check your plugin directory:
   ```bash
   ls ~/.pulumi/plugins/resource-webflow-v*/bin/
   ```

**Prevention:** Always run `pulumi plugin install` when updating your Pulumi CLI or provider version.

---

### SDK Package Installation Failures

**Error:** `Module not found` or `Package not found`

**Cause:** Webflow SDK not installed for your language.

**Solutions:**

**TypeScript/JavaScript:**
```bash
npm install @jdetmar/pulumi-webflow
# or
yarn add @jdetmar/pulumi-webflow
```

**Python:**
```bash
pip install pulumi-webflow
```

**Go:**
```bash
go get github.com/jdetmar/pulumi-webflow/sdk/go/webflow
```

**.NET:**
```bash
dotnet add package Pulumi.Webflow
```

**Java:**
```bash
# Add to pom.xml:
<dependency>
  <groupId>com.github.jdetmar</groupId>
  <artifactId>pulumi-webflow</artifactId>
  <version>0.1.0</version>
</dependency>
```

---

### Version Compatibility Issues

**Error:** `Incompatible provider version` or `Version mismatch`

**Cause:** SDK version doesn't match provider plugin version.

**Solution:**
1. Check installed plugin version:
   ```bash
   pulumi plugin ls | grep webflow
   ```
2. Check SDK version in your project:
   - **TypeScript:** `npm list @jdetmar/pulumi-webflow`
   - **Python:** `pip show pulumi-webflow`
   - **Go:** `go list -m github.com/jdetmar/pulumi-webflow`

3. Update to matching versions:
   ```bash
   # Update SDK
   npm install @jdetmar/pulumi-webflow@0.1.0

   # Update plugin
   pulumi plugin install resource webflow v0.1.0
   ```

**Prevention:** Pin SDK versions in your lock files (package-lock.json, poetry.lock, go.mod, etc.)

---

### Platform-Specific Issues

**Windows:**
- Ensure plugin path is in `%USERPROFILE%\.pulumi\plugins\resource-webflow-v*\bin\`
- Use backslashes or forward slashes consistently in paths
- Run PowerShell as Administrator if permission errors occur

**macOS:**
- Intel: Plugin should work with x86_64 architecture
- Apple Silicon: Use arm64 builds or rosetta emulation
- Check codesigning issues: `spctl -a -v -t exec ~/.pulumi/plugins/resource-webflow-v*/bin/pulumi-resource-webflow`

**Linux:**
- Ensure execute permissions: `chmod +x ~/.pulumi/plugins/resource-webflow-v*/bin/pulumi-resource-webflow`
- For Ubuntu/Debian: May need to install libssl-dev dependencies
- SELinux: May need to adjust context for plugin binary

## Authentication & Credentials

### API Token Not Configured

**Error:** `API token not configured` or `WEBFLOW_API_TOKEN not set`

**Cause:** No Webflow API token provided to the provider.

**Solution:**
1. Get your API token from Webflow:
   - Go to Webflow Dashboard
   - Account Settings → API & Webhooks
   - Generate new token or copy existing token
   - Copy the token (it's only shown once)

2. Configure token in your Pulumi stack:
   ```bash
   pulumi config set webflow:apiToken "your_token_here" --secret
   ```
   Or set environment variable:
   ```bash
   export WEBFLOW_API_TOKEN="your_token_here"
   ```

3. For CI/CD, set as a secret:
   ```bash
   pulumi config set webflow:apiToken --secret
   # Then enter your token when prompted
   ```

**Prevention:** Never commit tokens to version control. Use `--secret` flag when setting tokens.

#### Authentication error codes

When authentication fails, you may see structured error codes in addition to the human‑readable messages documented in this section:

- `WEBFLOW_AUTH_001`: No credentials provided (for example, `WEBFLOW_API_TOKEN` not set or `webflow:apiToken` not configured). See [API Token Not Configured](#api-token-not-configured).
- `WEBFLOW_AUTH_002`: Credentials were provided but were rejected by the Webflow API (for example, invalid, revoked, or expired token). See [Invalid or Expired Token](#invalid-or-expired-token).
- `WEBFLOW_AUTH_003`: Other authentication or authorization issues (for example, token with insufficient permissions or account-level restrictions). Review the full error message and verify token scope and account access.
---

### Authentication Error Codes

When authentication fails, you may see structured error codes in addition to the human-readable messages documented in this section:

- `WEBFLOW_AUTH_001`: No credentials provided (for example, `WEBFLOW_API_TOKEN` not set or `webflow:apiToken` not configured). See [API Token Not Configured](#api-token-not-configured).
- `WEBFLOW_AUTH_002`: Credentials were provided but are empty or invalid format. See [Invalid or Expired Token](#invalid-or-expired-token).
- `WEBFLOW_AUTH_003`: Token format is invalid (for example, too short). Review the full error message and verify your token.

These error codes are designed for programmatic error handling in CI/CD pipelines and automation scripts.

---

### Invalid or Expired Token

**Error:** `Unauthorized` or `Invalid token` (401 error)

**Cause:** Token is invalid, expired, or revoked.

**Solution:**
1. Verify token format: Should be a 40+ character alphanumeric string
2. Generate a new token:
   - Webflow Dashboard → Account Settings → API & Webhooks
   - Delete old token (invalidates it)
   - Create new token and update your configuration
3. Update provider configuration:
   ```bash
   pulumi config set webflow:apiToken "new_token_here" --secret
   ```

**Prevention:**
- Rotate tokens periodically
- Revoke tokens when no longer needed
- Monitor token usage in Webflow Dashboard

---

### Insufficient Permissions

**Error:** `Forbidden` or `Access denied` (403 error)

**Cause:** Token has limited permissions, or account doesn't have access to the site.

**Solution:**
1. Verify token permissions:
   - Webflow Dashboard → Account Settings → API & Webhooks
   - Check token has "Sites" and "Collections" scopes if needed

2. Verify account has access to the site:
   - Webflow Dashboard → Sites
   - Ensure you're a member of the site's team
   - Check your role has necessary permissions (typically require Admin or Editor)

3. For team accounts, verify:
   - You're in the correct team workspace
   - Your role includes the necessary permissions
   - Site isn't restricted to specific team members

**Prevention:** Use tokens with minimal required permissions. Review token permissions before deploying.

---

### Credential Leakage in Logs

**Error:** Token or credentials visible in log output

**Cause:** Credentials accidentally logged or printed.

**Solution:**
1. Always use `--secret` when setting sensitive values:
   ```bash
   pulumi config set webflow:apiToken --secret
   ```

2. Never print credentials in code:
   ```go
   // BAD
   fmt.Println("Token:", apiToken)

   // GOOD
   // Just use token, don't print it
   ```

3. Check Pulumi logs are clean:
   ```bash
   pulumi stack export | grep -v "apiToken"
   ```

4. If credentials were exposed:
   - Immediately regenerate tokens in Webflow Dashboard
   - Rotate any credentials that may have been exposed
   - Check Webflow Dashboard for unauthorized activity

**Prevention:**
- Never log credentials
- Use environment variables instead of hardcoding
- Review logs before sharing with others
- Use `--secret` for all sensitive configuration

## Configuration

### Invalid Site ID Format

**Error:** `invalid site ID 'abc123': must be 24-character hex string`

**Cause:** Site ID format is incorrect.

**Solution:**
1. Get correct site ID from Webflow:
   - Go to Webflow Designer
   - Project Settings → API & Webhooks
   - Copy your site ID (24-character hex string like `507f1f77bcf86cd799439011`)

2. Update your configuration:
   ```bash
   pulumi config set webflow:siteId "correct_site_id"
   ```

3. For individual resources:
   ```python
   site = webflow.Site("my-site",
       site_id="507f1f77bcf86cd799439011"  # 24-character hex string
   )
   ```

**Format Requirements:**
- Must be exactly 24 characters
- Must be hexadecimal (0-9, a-f)
- Example valid IDs: `507f1f77bcf86cd799439011`, `64f1a2b3c4d5e6f7890abcde`

---

### Site Not Found

**Error:** `Site not found` or `404 Not Found` for site

**Cause:**
- Site ID is incorrect
- Site was deleted
- Account doesn't have access to site

**Solution:**
1. Verify site ID:
   - Webflow Designer → Project Settings → API & Webhooks
   - Copy site ID exactly as shown
   - Check for typos or extra spaces

2. Verify you have access:
   - Webflow Dashboard → Sites
   - Confirm site exists in your list
   - If team account, verify you're on correct team

3. Check if site was deleted:
   - Webflow Dashboard → Sites
   - If site not listed, it was deleted
   - You'll need to recreate it or use different site

4. Update configuration:
   ```bash
   pulumi config set webflow:siteId "correct_site_id"
   ```

---

### Invalid Resource Properties

**Error:** `Invalid property 'xxx'` or `Unknown field`

**Cause:** Property name or value is invalid.

**Solution:**
1. Check property names:
   - Consult API documentation for resource type
   - Property names are case-sensitive
   - Use snake_case (not camelCase) for Pulumi resources

2. Common property mistakes:
   - Resource types are case-sensitive
   - Some properties may have format requirements
   - Dates should be ISO 8601 format
   - IDs should be 24-character hex strings

3. Example: Creating a Site
   ```python
   site = webflow.Site("my-site",
       site_id="507f1f77bcf86cd799439011",  # Required, 24-char hex
       display_name="My Site"  # Optional, string
   )
   ```

4. Check API documentation for required vs. optional properties

---

### Stack Configuration Errors

**Error:** `Stack 'xxx' not found` or Config mismatch

**Cause:** Stack doesn't exist or configuration is missing.

**Solution:**
1. List available stacks:
   ```bash
   pulumi stack ls
   ```

2. Create missing stack:
   ```bash
   pulumi stack init staging
   pulumi stack init production
   ```

3. Set required configuration:
   ```bash
   pulumi config set webflow:siteId "your_site_id"
   pulumi config set webflow:apiToken "your_token" --secret
   ```

4. Verify configuration:
   ```bash
   pulumi config
   ```

---

### Environment Variable Issues

**Error:** Variables not recognized or wrong values used

**Cause:** Environment variables not set correctly.

**Solution:**
1. Set environment variables:
   ```bash
   export WEBFLOW_API_TOKEN="your_token"
   export WEBFLOW_SITE_ID="your_site_id"
   ```

2. Verify they're set:
   ```bash
   echo $WEBFLOW_API_TOKEN
   echo $WEBFLOW_SITE_ID
   ```

3. For Windows:
   ```powershell
   $env:WEBFLOW_API_TOKEN = "your_token"
   $env:WEBFLOW_SITE_ID = "your_site_id"
   ```

4. For CI/CD, set as secrets in your pipeline configuration (GitHub Actions, GitLab CI, etc.)

## Runtime Errors

### Resource Creation Failures

**Error:** `Error creating resource` with API error details

**Cause:** Resource creation failed in Webflow API.

**Solutions depend on resource type:**

**Site Creation Failure:**
- Verify site ID is available
- Check account has permission to create sites
- Verify all required properties are provided
- Check Webflow API status

**Resource Creation Failure (RobotsTxt, Redirect):**
- Verify site exists and is accessible
- Check resource-specific constraints (e.g., domain format for redirects)
- Verify no conflicting resources already exist
- Check resource properties match API requirements

**Action:**
1. Enable verbose logging to see full error:
   ```bash
   pulumi up --debug 2>&1 | grep -A 5 "Error creating"
   ```

2. Check error message for specific cause
3. Review "Diagnostic Procedures" section for systematic troubleshooting

---

### Update Conflicts

**Error:** `Conflict` or `Resource already exists` during update

**Cause:**
- Resource was modified outside Pulumi
- State is out of sync with actual resource
- Concurrent update attempt

**Solution:**
1. Refresh Pulumi state:
   ```bash
   pulumi refresh
   ```

2. If state still conflicts:
   ```bash
   pulumi refresh --force
   ```

3. Review what changed:
   ```bash
   pulumi preview
   ```

4. Understand the conflict:
   - Was the resource modified in Webflow Designer?
   - Has another Pulumi user made changes?
   - Is there a concurrent deployment?

5. Resolve by:
   - Accepting the remote changes: `pulumi refresh`
   - Or overwriting with your desired state: `pulumi up --force`

---

### Delete Failures

**Error:** `Cannot delete resource` or `Resource has dependencies`

**Cause:**
- Other resources depend on this resource
- Resource is protected or locked
- API doesn't allow deletion

**Solution:**
1. Check dependencies:
   ```bash
   pulumi export | grep -A 2 "references"
   ```

2. Delete dependent resources first:
   - Identify which resources reference the failing resource
   - Remove or update those resources first
   - Then delete the original resource

3. If protected, remove protection:
   ```bash
   pulumi config set --path resource.protection false
   ```

4. Force deletion if safe:
   ```bash
   pulumi destroy --force
   ```

---

### Validation Errors

**Error:** `Validation failed` or specific field validation error

**Cause:** Resource properties fail validation.

**Solutions:**
1. Check property formats:
   - URLs should be valid URLs
   - Email addresses should be valid
   - Numbers should be correct type (integer, float)
   - Strings should meet length requirements

2. Review error message for specific field
3. Consult API documentation for field requirements
4. Update property value and retry

---

### Type Mismatches

**Error:** `Type error` or `Cannot convert X to Y`

**Cause:** Property value is wrong type.

**Solution:**
1. Check property type requirements:
   ```python
   # WRONG - string where int expected
   site = webflow.Site("my-site", site_id=12345)

   # CORRECT - string
   site = webflow.Site("my-site", site_id="507f1f77bcf86cd799439011")
   ```

2. For each language:
   - **Python:** Use type hints to catch errors early
   - **TypeScript:** Enable strict mode: `"strict": true` in tsconfig.json
   - **Go:** Type system catches at compile time
   - **.NET:** Strong typing helps catch errors

3. Cast values if needed:
   ```typescript
   const siteId = config.get("siteId") as string;
   ```

## Resource-Specific Issues

### RobotsTxt Resource Errors

**Error:** `Failed to update robots.txt` or `robots.txt content rejected`

**Cause:**

- Invalid robots.txt syntax
- Content exceeds size limits
- Site doesn't support robots.txt customization

**Solution:**

1. Validate robots.txt syntax:

   ```text
   # Valid robots.txt format
   User-agent: *
   Disallow: /admin/
   Allow: /public/

   Sitemap: https://example.com/sitemap.xml
   ```

2. Check content requirements:
   - Must start with valid directive (User-agent, Disallow, Allow, Sitemap)
   - No HTML or script content allowed
   - Keep size under 500KB (Webflow limit)

3. Verify site access:
   ```bash
   # Confirm site ID is correct
   pulumi config get webflow:siteId
   ```

**Common Mistakes:**

- Including HTML tags in robots.txt content
- Using invalid directive names (case-sensitive)
- Missing User-agent before Disallow/Allow rules

**Example - Correct Usage:**

```python
robots = webflow.RobotsTxt("my-robots",
    site_id="507f1f77bcf86cd799439011",
    content="""User-agent: *
Disallow: /admin/
Disallow: /private/
Allow: /

Sitemap: https://example.com/sitemap.xml"""
)
```

---

### Redirect Resource Errors

**Error:** `Invalid redirect source` or `Redirect conflict`

**Cause:**
- Source path format invalid
- Duplicate redirect already exists
- Circular redirect detected

**Solution:**
1. Check source path format:
   - Must start with `/`
   - No query strings in source (use path only)
   - Case-sensitive matching

2. Check for conflicts:
   ```bash
   pulumi refresh
   pulumi preview
   ```

3. Verify no circular redirects:
   - `/a` → `/b` → `/a` is circular
   - Check your redirect chain

**Example - Correct Usage:**

```python
redirect = webflow.Redirect("old-to-new",
    site_id="507f1f77bcf86cd799439011",
    source="/old-page",      # Must start with /
    target="/new-page"       # Can be path or full URL
)
```

---

## Network & API Issues

### Connection Timeouts

**Error:** `Connection timeout` or `Request timed out`

**Cause:**
- Network connectivity issue
- Webflow API slow or unavailable
- Firewall blocking connection

**Solution:**
1. Check network connectivity:
   ```bash
   ping api.webflow.com
   ```

2. Verify Webflow API is available:
   - Check Webflow Status Page: https://status.webflow.com
   - Try accessing Webflow Designer in browser

3. Increase timeout:
   ```bash
   # In your Pulumi configuration
   PULUMI_BACKEND_CONTEXT_TIMEOUT=60s pulumi up
   ```

4. Check firewall/proxy:
   - If behind corporate proxy, configure Pulumi
   - Check firewall allows outbound HTTPS on port 443
   - Disable VPN if it's interfering

5. Retry operation:
   ```bash
   pulumi up --refresh
   ```

**Prevention:** Configure appropriate timeouts for your network environment.

---

### Rate Limiting (429 Errors)

**Error:** `Rate limit exceeded (429)` or `Too Many Requests`

**Cause:** Too many API requests in short time period.

**Solution:**
1. Wait before retrying:
   ```bash
   sleep 60
   pulumi up
   ```

2. Reduce parallelism:
   ```bash
   pulumi up --parallelism=1
   ```

3. Spread operations over time:
   - Don't deploy multiple stacks simultaneously
   - Stagger resource creation in large deployments
   - Batch operations when possible

4. Check your API token limits:
   - Webflow Dashboard → Account Settings → API & Webhooks
   - Review rate limit documentation
   - Consider using higher-tier API token if available

**Prevention:**
- Use `--parallelism=1` for initial deployments
- Monitor API usage in Webflow Dashboard
- Implement exponential backoff in custom code

---

### API Unavailability

**Error:** `500 Internal Server Error` or `Service Unavailable`

**Cause:** Webflow API is down or having issues.

**Solution:**
1. Check Webflow Status:
   - https://status.webflow.com
   - Subscribe to status updates

2. Wait for Webflow to recover:
   ```bash
   # Wait 30 seconds
   sleep 30
   pulumi up
   ```

3. Check your API token still works:
   ```bash
   # Try a simple operation
   pulumi preview
   ```

4. If persistent, contact Webflow Support:
   - Webflow Dashboard → Help → Contact Support
   - Include your site ID and error message

**Prevention:**
- Monitor Webflow Status page before deployments
- Implement retry logic in critical deployments
- Have incident response plan for API outages

---

### Network Proxy Issues

**Error:** `Connection refused` or `Proxy error`

**Cause:** Network proxy interfering with API calls.

**Solution:**
1. Configure HTTP/HTTPS proxy:
   ```bash
   # For HTTP proxy
   export HTTP_PROXY="http://proxy.company.com:8080"
   export HTTPS_PROXY="http://proxy.company.com:8080"

   # Configure no-proxy list
   export NO_PROXY="localhost,127.0.0.1"
   ```

2. For corporate proxy with authentication:
   ```bash
   export HTTPS_PROXY="http://user:password@proxy.company.com:8080"
   ```

3. Verify proxy settings:
   ```bash
   echo $HTTP_PROXY
   echo $HTTPS_PROXY
   ```

4. Test connectivity through proxy:
   ```bash
   curl -x $HTTPS_PROXY https://api.webflow.com
   ```

---

### Firewall Blocking

**Error:** `Connection refused` or `No route to host`

**Cause:** Firewall blocking outbound connections.

**Solution:**
1. Check if port 443 is open:
   ```bash
   telnet api.webflow.com 443
   ```

2. For corporate firewall:
   - Contact network team
   - Whitelist `api.webflow.com` on port 443
   - Ensure HTTPS is not intercepted

3. For personal firewall:
   - Check Windows Defender, macOS firewall
   - Allow Pulumi CLI outbound connections
   - Temporarily disable firewall to test

4. Test connectivity:
   ```bash
   curl https://api.webflow.com/info
   ```

## State Management

### State Drift Detection

**Error:** `Drift detected` when running `pulumi refresh`

**Cause:** Resources in Webflow differ from Pulumi state.

**Solution:**
1. Review the drift:
   ```bash
   pulumi refresh
   pulumi stack export
   ```

2. Understand what changed:
   - Was resource modified in Webflow Designer?
   - Was it modified by another Pulumi user?
   - Was it updated by an automated process?

3. Decide how to handle:
   ```bash
   # Accept the remote changes (update state from Webflow)
   pulumi refresh

   # Or overwrite with desired state
   pulumi up --force
   ```

4. To prevent drift:
   - Only modify resources through Pulumi
   - Use `ReadOnly` resources if monitoring only
   - Implement drift detection in CI/CD

**Prevention:**
- Run `pulumi refresh` regularly
- Monitor Pulumi deployments
- Restrict direct Webflow Designer access for managed sites

---

### State Refresh Failures

**Error:** `Failed to refresh state`

**Cause:**
- Resource no longer exists in Webflow
- Permissions changed
- API error

**Solution:**
1. Check if resource exists:
   - Webflow Dashboard → Check if resource is there
   - If deleted, remove from Pulumi: `pulumi destroy`

2. Check permissions:
   - Verify API token still has access
   - Check site still accessible
   - Verify account permissions

3. Try refresh again:
   ```bash
   pulumi refresh
   ```

4. If still failing, get details:
   ```bash
   pulumi refresh --debug 2>&1 | grep -A 10 "error"
   ```

---

### State Corruption

**Error:** Invalid state or corrupted state file

**Cause:**
- Manual state file edits
- Incomplete operations
- Backup/restore issues

**Solution:**
1. Never edit state files manually
2. If corrupted, restore from backup:
   ```bash
   pulumi stack export > backup.json
   # Restore from earlier version if available
   ```

3. Rebuild state from scratch:
   ```bash
   # Remove corrupted stack
   pulumi stack rm

   # Recreate stack
   pulumi stack init
   pulumi up
   ```

**Prevention:**
- Don't manually edit state files
- Keep regular backups
- Use managed backends (Pulumi Cloud recommended)

---

### Import Conflicts

**Error:** `Resource already exists` during import or `Import failed`

**Cause:**
- Resource ID already exists in state
- Resource not found in Webflow
- Wrong resource type

**Solution:**
1. Verify resource exists in Webflow:
   - Webflow Dashboard → Check resource
   - Verify you have access

2. Check it's not already in state:
   ```bash
   pulumi stack export | grep "resource_id"
   ```

3. Remove conflicting resource from state (if needed):
   ```bash
   pulumi state delete resources/xxx
   ```

4. Import correctly:
   ```bash
   pulumi import webflow:index:Site my-site 507f1f77bcf86cd799439011
   ```

---

### Missing State File

**Error:** `State file not found` or `Stack doesn't exist`

**Cause:**
- Stack was deleted
- Wrong backend configured
- State file path incorrect

**Solution:**
1. Create new stack:
   ```bash
   pulumi stack init
   ```

2. Or switch to existing stack:
   ```bash
   pulumi stack select <stack-name>
   ```

3. Verify backend configuration:
   ```bash
   pulumi config show
   ```

4. If using self-managed backend, verify path:
   ```bash
   export PULUMI_BACKEND_URL="file://~/.pulumi"
   ```

## Multi-Environment

### Wrong Stack Deployed

**Error:** Wrong site ID or wrong configuration deployed

**Cause:**
- Selected wrong stack
- Stack configuration incorrect
- Deployment mismatch

**Solution:**
1. Verify correct stack selected:
   ```bash
   pulumi stack
   ```

2. List all stacks:
   ```bash
   pulumi stack ls
   ```

3. Switch to correct stack:
   ```bash
   pulumi stack select staging
   ```

4. Verify configuration:
   ```bash
   pulumi config
   ```

5. Only then deploy:
   ```bash
   pulumi up
   ```

**Prevention:**
- Always verify stack name before deploying
- Use stack-specific naming conventions
- Implement CI/CD guards for production deployments

---

### Credential Mixing

**Error:** Using wrong credentials for environment

**Cause:**
- Environment variables point to wrong account
- Stack configuration has wrong token

**Solution:**
1. Verify credentials:
   ```bash
   echo $WEBFLOW_API_TOKEN
   pulumi config get webflow:apiToken
   ```

2. Separate credentials by stack:
   ```bash
   # For staging
   pulumi config set webflow:apiToken "staging_token" --secret

   # For production (select prod stack first)
   pulumi stack select production
   pulumi config set webflow:apiToken "prod_token" --secret
   ```

3. Use stack-specific secrets:
   ```bash
   # Pulumi Cloud manages secrets per stack
   pulumi config set webflow:apiToken --secret
   ```

**Prevention:**
- Never share tokens between environments
- Use different Webflow accounts for dev/staging/prod
- Rotate tokens regularly

---

### Site ID Conflicts

**Error:** Site ID used for multiple stacks or wrong site being updated

**Cause:**
- Multiple stacks managing same site
- Configuration mismatch

**Solution:**
1. Verify each stack has unique site ID:
   ```bash
   pulumi stack select dev
   pulumi config get webflow:siteId

   pulumi stack select staging
   pulumi config get webflow:siteId
   ```

2. Update if needed:
   ```bash
   pulumi stack select staging
   pulumi config set webflow:siteId "staging_site_id"
   ```

3. Document site ID mapping:
   ```
   Development: 507f1f77bcf86cd799439011
   Staging: 507f1f77bcf86cd799439012
   Production: 507f1f77bcf86cd799439013
   ```

**Prevention:**
- One stack per site
- Clear naming conventions
- Document site ID assignments

---

### Environment-Specific Errors

**Error:** Works in one environment but not another

**Cause:**
- Environment-specific configuration
- Different API token permissions
- Environment-specific firewall rules

**Solution:**
1. Compare configurations:
   ```bash
   pulumi stack select dev && pulumi config
   pulumi stack select prod && pulumi config
   ```

2. Check permissions differ:
   - Webflow Dashboard → Team Members
   - Verify token permissions in each environment
   - Check API token rotation status

3. Test connectivity in specific environment:
   ```bash
   WEBFLOW_API_TOKEN=$(pulumi config get webflow:apiToken) \
   curl https://api.webflow.com/user
   ```

## CI/CD Integration

### Non-Interactive Prompts

**Error:** `Error: input is not a tty` or prompts in pipeline

**Cause:** Pulumi expects interactive input in non-interactive CI/CD.

**Solution:**
1. Configure auto-approve mode:
   ```bash
   pulumi up --yes
   ```

2. Set skip update check:
   ```bash
   export PULUMI_SKIP_UPDATE_CHECK=true
   ```

3. Configure stack selection:
   ```bash
   export PULUMI_STACK="production"
   ```

4. In CI/CD pipeline:
   ```yaml
   # GitHub Actions example
   - name: Deploy with Pulumi
     run: |
       pulumi stack select production
       pulumi up --yes
     env:
       PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_TOKEN }}
       WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_TOKEN }}
   ```

---

### Timeout in Pipelines

**Error:** Deployment times out in CI/CD

**Cause:**
- Pipeline timeout too short
- Large deployments take longer
- Network latency in pipeline environment

**Solution:**
1. Increase pipeline timeout:
   ```yaml
   # GitHub Actions
   - name: Deploy
     run: pulumi up --yes
     timeout-minutes: 30
   ```

2. Reduce parallelism:
   ```bash
   pulumi up --yes --parallelism=1
   ```

3. Enable verbose logging to see progress:
   ```bash
   pulumi up --yes --debug
   ```

4. Check API rate limits:
   - May need to reduce number of resources
   - Implement exponential backoff

---

### Credential Injection

**Error:** Credentials not available in pipeline

**Cause:**
- Secrets not configured
- Environment variable names wrong
- Token format incorrect

**Solution:**
1. Configure secrets in CI/CD:
   ```bash
   # GitHub Actions
   Settings → Secrets → New repository secret
   WEBFLOW_API_TOKEN = "your_token"
   PULUMI_ACCESS_TOKEN = "your_pulumi_token"
   ```

2. Use secrets in pipeline:
   ```yaml
   env:
     WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}
     PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
   ```

3. Verify credentials set:
   ```bash
   - name: Verify credentials
     run: |
       [ -n "$WEBFLOW_API_TOKEN" ] && echo "✓ WEBFLOW_API_TOKEN set"
       [ -n "$PULUMI_ACCESS_TOKEN" ] && echo "✓ PULUMI_ACCESS_TOKEN set"
   ```

---

### Exit Code Handling

**Error:** Pipeline doesn't fail when deployment fails

**Cause:**
- Exit codes not checked
- Error handling not configured

**Solution:**
1. Always check exit codes:
   ```bash
   pulumi up --yes || exit 1
   ```

2. Capture output:
   ```bash
   if ! pulumi up --yes; then
       echo "Deployment failed"
       exit 1
   fi
   ```

3. In YAML pipelines (auto-failures):
   ```yaml
   - run: pulumi up --yes
     # Automatically fails if exit code != 0
   ```

## Diagnostic Procedures

Follow these step-by-step procedures to systematically troubleshoot issues.

### Procedure 1: Systematic Investigation Workflow

Use this when you're unsure where the problem is.

**Step 1: Verify Prerequisites**
1. Check Pulumi CLI installed:
   ```bash
   pulumi version
   ```

2. Check provider plugin installed:
   ```bash
   pulumi plugin ls | grep webflow
   ```

3. Check you're in Pulumi project directory:
   ```bash
   ls Pulumi.*.yaml
   ```

4. Check credentials set:
   ```bash
   [ -n "$WEBFLOW_API_TOKEN" ] && echo "Token set"
   ```

**Step 2: Check Stack Configuration**
1. Verify stack exists:
   ```bash
   pulumi stack
   ```

2. Verify configuration:
   ```bash
   pulumi config
   ```

3. Verify secrets are set:
   ```bash
   pulumi config --show-secrets
   ```

**Step 3: Verify Connectivity**
1. Check network:
   ```bash
   ping api.webflow.com
   ```

2. Check API availability:
   ```bash
   curl -I https://api.webflow.com/user
   ```
   Should return 401 (unauthorized) or 200, not connection error

3. Check Webflow Status:
   ```bash
   curl https://status.webflow.com/api/v2/components.json
   ```

**Step 4: Validate Configuration**
1. Check site ID format:
   ```bash
   pulumi config get webflow:siteId
   # Should be 24-character hex string
   ```

2. Test API token:
   ```bash
   curl -H "Authorization: Bearer $WEBFLOW_API_TOKEN" \
        https://api.webflow.com/user
   ```

3. Check site access:
   ```bash
   curl -H "Authorization: Bearer $WEBFLOW_API_TOKEN" \
        https://api.webflow.com/sites/$(pulumi config get webflow:siteId)
   ```

**Step 5: Run Preview**
1. Run dry-run:
   ```bash
   pulumi preview
   ```

2. Check output for errors:
   ```bash
   pulumi preview 2>&1 | grep -i error
   ```

**Step 6: Enable Verbose Logging**
1. Re-run with debug:
   ```bash
   pulumi preview --debug 2>&1 | tee debug.log
   ```

2. Search for errors:
   ```bash
   grep -i error debug.log
   ```

### Procedure 2: Information Gathering for Bug Reports

When filing a GitHub issue, gather:

1. **Environment Information:**
   ```bash
   pulumi version
   pulumi plugin ls | grep webflow
   go version  # or python --version, node --version, etc.
   uname -a     # OS information
   ```

2. **Configuration (without secrets):**
   ```bash
   pulumi config
   # DO NOT SHOW --show-secrets
   ```

3. **Error Output:**
   ```bash
   pulumi up --debug 2>&1 | tail -50 > error.log
   # Share error.log in issue
   ```

4. **Minimal Reproduction:**
   - Share Pulumi.yaml and Pulumi.[stack].yaml (no secrets)
   - Share main.py or main.go (simplified if large)

### Procedure 3: Isolation Techniques

To narrow down the problem:

1. **Isolate by Resource Type:**
   - Comment out all resources except the failing one
   - Test if still fails
   - Add resources back one at a time

2. **Isolate by Stack:**
   - Try in different stack
   - If works in one, configuration or state issue
   - If fails in all, code or credentials issue

3. **Isolate by Environment:**
   - Try in different environment (dev, staging, prod)
   - Check if environment-specific configuration

4. **Test Simple Case:**
   - Create minimal resource first
   - Verify it works
   - Add complexity incrementally

### Procedure 4: Validation Steps

After making changes, verify:

1. **Preview works:**
   ```bash
   pulumi preview
   # Should show desired changes, no errors
   ```

2. **Apply succeeds:**
   ```bash
   pulumi up --yes
   # Should complete with "Updating (stack-name)"
   ```

3. **Resource exists:**
   ```bash
   # Check in Webflow Dashboard
   ```

4. **State consistent:**
   ```bash
   pulumi refresh
   # Should show no drift
   ```

## Logging & Debugging

### Verbose Mode Activation

**Enable Detailed Logging:**

```bash
# Run with debug flag
pulumi up --debug

# Or set environment variable
export PULUMI_DEBUG=true
pulumi up
```

**For Specific Output:**
```bash
# Capture to file
pulumi up --debug 2>&1 | tee deployment.log

# Filter for errors
pulumi up --debug 2>&1 | grep -i error

# Filter for warnings
pulumi up --debug 2>&1 | grep -i warning
```

### Log Levels

**Log Level Descriptions:**

- **ERROR**: Critical failures that prevent operation (show first)
- **WARN**: Warnings about potential issues
- **INFO**: General informational messages (default level)
- **DEBUG**: Detailed debugging information
- **TRACE**: Very detailed trace information

**Control Log Level:**
```bash
# Set environment variable
export PULUMI_DEBUG=true          # Debug level
export PULUMI_SKIP_VERSION_CHECK=true

# For Go provider
export LOGLEVEL=debug
```

### Log Locations

**Pulumi Logs:**
```bash
# Default location
~/.pulumi/logs/

# List recent logs
ls -lrt ~/.pulumi/logs/ | tail

# View latest log
tail -f ~/.pulumi/logs/pulumi-*.log
```

**Webflow Provider Logs:**
- Check Pulumi logs first
- Provider output in `pulumi up` stdout/stderr

### Log Analysis Guide

**When troubleshooting:**

1. **Find the error:**
   ```bash
   grep -i "error\|failed" deployment.log | head -20
   ```

2. **Look for root cause:**
   - First error is usually the root cause
   - Subsequent errors may be consequences

3. **Understand error context:**
   ```bash
   # Show lines around error
   grep -i "error" -B 5 -A 5 deployment.log
   ```

4. **Check for stack traces:**
   ```bash
   grep -A 20 "panic\|traceback\|stack trace" deployment.log
   ```

5. **Identify API errors:**
   ```bash
   grep -i "api\|webflow\|401\|403\|404\|429\|500" deployment.log
   ```

### Sensitive Data Handling

**What's logged:**

- Webflow API responses (may contain site data)
- HTTP headers (may contain authorization info)
- Stack traces (may reference file paths)

**What's NOT logged:**

- API tokens (redacted in logs)
- Passwords (never logged)
- Sensitive configuration values (marked `--secret`)

**If credentials exposed:**
1. Immediately regenerate tokens
2. Review Webflow Dashboard for unauthorized access
3. Check deployment logs are cleaned up
4. Rotate any exposed credentials

**Prevent exposure:**
```bash
# Always use --secret for sensitive values
pulumi config set webflow:apiToken --secret

# Verify before sharing logs
grep -i "apitoken\|password\|secret" deployment.log
```

### Performance Impact

**Logging overhead:**
- DEBUG level: +5-10% slower
- TRACE level: +20-30% slower
- Used for troubleshooting, disable for production

**Best practices:**
- Disable debug logging in production
- Use for specific troubleshooting sessions
- Remove debug logs after troubleshooting
- Monitor log file size (logs can grow large)

---

## Getting Help

If you can't find your issue here:

1. **Check Recent Issues:** https://github.com/jdetmar/pulumi-webflow/issues
2. **Search Documentation:** Use Ctrl+F to search this page
3. **Enable Verbose Logging:** Use `--debug` flag for more details
4. **File Issue:** https://github.com/jdetmar/pulumi-webflow/issues/new
   - Include error output
   - Include steps to reproduce
   - Include output of `pulumi version` and `pulumi plugin ls`

5. **Webflow Support:** https://webflow.com/support
   - For Webflow API issues
   - For account/credential issues

Remember: Most issues are resolved by:
1. Checking error message carefully
2. Enabling verbose logging
3. Verifying credentials and configuration
4. Checking network connectivity
5. Consulting this guide
