# Webflow Provider Logging & Troubleshooting Guide

## Table of Contents

1. [Introduction](#introduction)
2. [Quick Start](#quick-start)
3. [Pulumi Logging Levels](#pulumi-logging-levels)
4. [Credential Redaction](#credential-redaction)
5. [Common Troubleshooting Scenarios](#common-troubleshooting-scenarios)
6. [CI/CD Logging Configuration](#cicd-logging-configuration)
7. [Log Analysis Techniques](#log-analysis-techniques)
8. [Performance Considerations](#performance-considerations)
9. [Troubleshooting](#troubleshooting)

---

## Introduction

Detailed logging is critical when troubleshooting issues with the Webflow Pulumi provider. This guide explains how to enable comprehensive logging, understand log output, verify credential safety, and troubleshoot common problems.

### Why Detailed Logging Matters

Verbose logging reveals:

- **API Calls**: What requests the provider sends to Webflow API
- **Responses**: How the API responds (HTTP status, response data)
- **Errors**: Detailed error messages and stack traces
- **Configuration**: What configuration the provider loaded
- **Resource Creation**: Step-by-step progress of resource operations

### When to Enable Verbose Logging

**✅ DO enable verbose logging:**

- Troubleshooting deployment failures
- Diagnosing authentication or permission issues
- Debugging resource creation or update failures
- Creating bug reports for support
- Development and testing environments
- Initial deployment validation in new environments

**❌ DON'T enable verbose logging:**

- Production deployments (performance impact: 5-10% overhead)
- Automated CI/CD pipelines (unless debugging failures)
- When credentials might be exposed (verify redaction!)
- High-frequency deployments

### Prerequisites

- Pulumi CLI installed (v3.0 or later)
- Webflow Pulumi provider installed
- Understanding of Pulumi concepts (stacks, resources, configuration)
- Familiarity with your terminal/command line

---

## Quick Start

### 1. Enable Verbose Logging for Local Development

```bash
# Option A: Command-line flag (simplest)
pulumi up --verbose

# Option B: Environment variable (affects all commands)
export PULUMI_LOG_LEVEL=debug
pulumi up

# Option C: Capture logs to file for analysis
pulumi up --verbose 2>&1 | tee deployment.log
```

### 2. View Log Output

The verbose flag adds detailed output to your terminal:

```
✅ Running new deployment on stack 'dev'

🔍 Loading configuration for troubleshooting example
🔐 Verifying Webflow API authentication
Token source: Pulumi config (credentials redacted in logs)
🏗️  Creating Webflow site

API Call: GET https://api.webflow.com/v2/sites
Response: 200 OK
...
✅ Site created successfully: <site-id>
🤖 Configuring robots.txt
✅ Robots.txt configured successfully
📤 Exported site ID for reference
```

### 3. Verify Credential Redaction

Critical: Verify that your credentials are NOT exposed in logs:

```bash
# Search logs for any token mentions
pulumi up --verbose 2>&1 | grep -i "token\|bearer\|authorization"

# Expected output:
# ✅ Should ONLY see: "[REDACTED]"
# ❌ Should NEVER see: actual token values like "wf_xyz..."
```

### 4. Disable Verbose Logging

Once you've resolved the issue, disable verbose logging for better performance:

```bash
# Clear the environment variable
unset PULUMI_LOG_LEVEL

# Or use command-line flags without --verbose
pulumi up
```

---

## Pulumi Logging Levels

Pulumi provides multiple logging levels for different levels of detail:

### Logging Levels

| Level | Flag | Env Variable | Description |
|-------|------|-------------|-------------|
| **Info** | (default) | `PULUMI_LOG_LEVEL=info` | Normal operational messages (recommended for production) |
| **Debug** | `--verbose` | `PULUMI_LOG_LEVEL=debug` | Detailed diagnostic information (troubleshooting) |
| **Warning** | N/A | `PULUMI_LOG_LEVEL=warn` | Potential issues that don't prevent execution |
| **Error** | N/A | `PULUMI_LOG_LEVEL=error` | Failures that prevent operations |

### Command-Line Flags

```bash
# Enable verbose/debug logging
pulumi up --verbose

# Output logs to stderr instead of stdout
pulumi up --logtostderr

# Enable detailed workflow logging (very verbose)
pulumi up --logflow

# Combine flags for maximum detail
pulumi up --verbose --logtostderr
```

### Environment Variables

```bash
# Set logging level for all Pulumi operations
export PULUMI_LOG_LEVEL=debug

# Enable gRPC debug logging (provider internals)
export PULUMI_DEBUG_GRPC=true

# Direct all logs to stderr
export PULUMI_LOG_TO_STDERR=true
```

### Log File Locations

Pulumi automatically saves logs to `~/.pulumi/logs/`:

```bash
# View today's logs
cat ~/.pulumi/logs/pulumi-$(date +%Y%m%d).log

# View last 50 lines of logs
tail -50 ~/.pulumi/logs/pulumi-*.log

# Search logs for specific errors
grep -i "error\|failed" ~/.pulumi/logs/pulumi-*.log
```

---

## Credential Redaction

**CRITICAL: The Webflow provider implements automatic credential redaction.**

### How Credential Redaction Works

The provider uses the `RedactToken()` function to ensure sensitive values are never logged:

```go
// Every token reference in logs becomes:
Token: [REDACTED]

// Authorization headers become:
Authorization: Bearer [REDACTED]

// Connection strings become:
webflow:apiToken: [REDACTED]
```

### Verifying Redaction is Working

Before enabling verbose logging in production, verify redaction is functioning:

```bash
# 1. Create a test with verbose logging
export WEBFLOW_API_TOKEN="wf_test123456789"
pulumi up --verbose 2>&1 > test-logs.txt

# 2. Search for your token pattern - should find NOTHING
grep -i "wf_test123456789" test-logs.txt
# (no output = good ✅)

# 3. Search for redaction placeholder - should find MATCHES
grep "\[REDACTED\]" test-logs.txt
# (shows matches = good ✅)

# 4. Clean up
rm test-logs.txt
unset WEBFLOW_API_TOKEN
```

### Security Best Practices

✅ **DO:**

- Always use `--secret` flag when setting credentials:
  ```bash
  pulumi config set webflow:apiToken $TOKEN --secret
  ```

- Verify encrypted storage in stack config:
  ```bash
  # Should show "secure:" not plain text
  grep "webflow:apiToken" Pulumi.*.yaml
  ```

- Review CI/CD logs before committing to version control

- Use different tokens for different environments (dev/staging/prod)

- Rotate credentials regularly

❌ **DON'T:**

- Use `--show-secrets` flag in production logs
- Store plain-text tokens in configuration files
- Commit unencrypted credentials to version control
- Use same token across multiple environments
- Expose logs containing credentials publicly

---

## Common Troubleshooting Scenarios

### Authentication Failures

**Symptom**: "API token not configured" or 401 Unauthorized

**Debug steps**:

```bash
# 1. Check token is configured
pulumi config get webflow:apiToken
# Should show "[secret]" not empty

# 2. Verify token source
pulumi config --json | grep webflow

# 3. Run with verbose logging
pulumi up --verbose 2>&1 | grep -i "token\|auth"

# 4. Confirm redaction (token should show as [REDACTED])
pulumi up --verbose 2>&1 | grep -i "bearer\|authorization"
```

**Common causes**:

- Token not set: `pulumi config set webflow:apiToken $TOKEN --secret`
- Token expired: Regenerate token in Webflow dashboard
- Token lacks permissions: Verify token has required scopes
- Wrong environment: Verify `pulumi stack select` is correct

### API Connection Issues

**Symptom**: Timeout errors or "connection refused"

**Debug steps**:

```bash
# 1. Check network connectivity to Webflow API
curl -v https://api.webflow.com/v2/

# 2. Review verbose logs for connection errors
pulumi up --verbose 2>&1 | grep -i "connect\|timeout\|dns"

# 3. Check for firewall/proxy blocking
# (your network team can verify)
```

**Common causes**:

- Network firewall blocking `api.webflow.com`
- HTTP proxy interference
- DNS resolution failures
- Webflow API temporarily unavailable

### Rate Limiting

**Symptom**: "429 Too Many Requests" errors

**Debug steps**:

```bash
# Check logs for rate limiting indicators
pulumi up --verbose 2>&1 | grep -i "429\|rate"

# Verify request frequency
# (slow down concurrent deployments)
```

**Solutions**:

- Run deployments sequentially, not in parallel
- Increase delays between resource creation
- Contact Webflow support for rate limit increases

### Resource Creation Failures

**Symptom**: "Failed to create resource" errors

**Debug steps**:

```bash
# 1. Enable maximum verbosity
pulumi up --verbose --logtostderr

# 2. Look for specific API error messages
pulumi up --verbose 2>&1 | grep -A 5 "error\|failed"

# 3. Check resource configuration in code
# (verify displayName, shortName, timezone are valid)
```

**Common causes**:

- Invalid timezone value
- Duplicate short name (must be unique)
- Invalid characters in display name
- Permission denied (token lacks create permission)

### State Management Issues

**Symptom**: Resource exists in Webflow but not in Pulumi state

**Debug steps**:

```bash
# 1. Check local state file exists
ls Pulumi.*.json

# 2. View state contents
pulumi stack export > state-backup.json
cat state-backup.json | jq '.'

# 3. Run import to add orphaned resources
pulumi import <resource-type> <resource-name> <resource-id>
```

**Recovery**:

- Use `pulumi import` for orphaned resources
- Use `pulumi refresh` to sync state with actual resources
- Backup state before making state changes

---

## CI/CD Logging Configuration

### Environment Detection

The Python CI/CD example demonstrates environment-aware logging:

```python
# Detect CI/CD environment
is_ci = os.getenv("CI") == "true"
environment = os.getenv("PULUMI_STACK", "unknown")

# Configure logging based on environment
if is_ci:
    pulumi.log.info(f"🤖 Running in CI/CD environment: {environment}")
    pulumi.log.debug("Verbose logging enabled for CI/CD troubleshooting")
else:
    pulumi.log.info(f"💻 Running in local environment: {environment}")
```

### GitHub Actions Example

```yaml
name: Deploy Webflow Infrastructure

on: [push]

env:
  PULUMI_CONFIG_PASSPHRASE: ${{ secrets.PULUMI_PASSPHRASE }}
  WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}

jobs:
  deploy:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - uses: pulumi/actions@v4
        with:
          command: up
          stack-name: prod
          # Enable verbose logging for CI/CD troubleshooting
          args: --verbose

      # Capture logs for debugging
      - name: Save logs if failure
        if: failure()
        run: |
          tar -czf logs.tar.gz ~/.pulumi/logs/
          echo "::notice::Logs saved to logs.tar.gz"
```

### GitLab CI Example

```yaml
stages:
  - deploy

deploy_webflow:
  stage: deploy
  script:
    - export PULUMI_CONFIG_PASSPHRASE=$PULUMI_PASSPHRASE
    - pulumi stack select $CI_COMMIT_BRANCH
    # Enable verbose logging for troubleshooting
    - pulumi up --verbose --yes
  only:
    - main
    - develop
```

### Log Retention & Analysis

```bash
# In CI/CD, capture logs in artifact directories
mkdir -p logs
pulumi up --verbose 2>&1 | tee logs/deployment.log

# Archive logs for historical analysis
tar -czf logs-$(date +%Y%m%d-%H%M%S).tar.gz logs/

# Upload to artifact storage (GitHub, GitLab, etc.)
# Allows review of failures without re-running
```

---

## Log Analysis Techniques

### Parsing Log Output

Extract specific information from verbose logs:

```bash
# Find all API calls made by provider
pulumi up --verbose 2>&1 | grep "API Call\|GET\|POST\|PUT\|DELETE"

# Find all responses
pulumi up --verbose 2>&1 | grep "Response:" | head -20

# Find all errors
pulumi up --verbose 2>&1 | grep -i "error\|failed\|❌"

# Extract timing information
pulumi up --verbose 2>&1 | grep "Duration\|elapsed"
```

### Filtering Relevant Entries

```bash
# Find resource creation logs only
pulumi up --verbose 2>&1 | grep "Creating\|✅\|❌"

# Find authentication-related logs
pulumi up --verbose 2>&1 | grep -i "token\|auth\|credential\|\[redacted\]"

# Find performance-related logs
pulumi up --verbose 2>&1 | grep -i "timeout\|slow\|performance"
```

### Creating Support Tickets with Logs

When reporting issues:

1. **Capture full logs:**
   ```bash
   pulumi up --verbose 2>&1 | tee full-logs.txt
   ```

2. **Verify no credentials are exposed:**
   ```bash
   grep -i "wf_\|token=\|Bearer" full-logs.txt
   # Should return NO results (only [REDACTED])
   ```

3. **Include relevant sections in support ticket:**
   - Error messages (lines with ERROR or ❌)
   - API responses (HTTP status codes)
   - Resource IDs and names
   - Your environment (Pulumi version, provider version)

4. **Redact any remaining sensitive data:**
   ```bash
   # Remove any non-redacted sensitive information
   sed 's/your-secret/[REDACTED]/g' full-logs.txt > logs-for-support.txt
   ```

---

## Performance Considerations

### Impact of Verbose Logging

Verbose logging adds overhead:

| Aspect | Impact | Notes |
|--------|--------|-------|
| **Execution Speed** | 5-10% slower | Noticeable on large deployments |
| **Memory Usage** | 10-20% higher | Logs held in memory during execution |
| **Log File Size** | 2-5x larger | Disk space for ~/.pulumi/logs/ |
| **Network I/O** | Minimal | Logs stay local, not transmitted |

### Production Recommendations

**Default logging (Recommended for Production):**

```bash
# Use default (info) level
pulumi up  # No --verbose flag

# Result:
# ✅ Minimal performance impact
# ✅ Small log files
# ✅ Only essential messages shown
# ✅ Credentials never exposed
```

**Verbose logging (Troubleshooting Only):**

```bash
# Enable only when debugging
pulumi up --verbose

# Remember to disable after troubleshooting:
unset PULUMI_LOG_LEVEL
```

### Log File Management

```bash
# Check log directory size
du -sh ~/.pulumi/logs/

# Remove old logs (older than 30 days)
find ~/.pulumi/logs/ -name "*.log" -mtime +30 -delete

# Archive logs instead of deleting
tar -czf pulumi-logs-archive-$(date +%Y%m).tar.gz ~/.pulumi/logs/
```

---

## Troubleshooting

### Logs Not Appearing

**Problem**: Verbose flag isn't producing detailed output

**Solutions**:

1. **Verify the flag:**
   ```bash
   pulumi up --verbose  # Double-check spelling
   ```

2. **Check environment variable:**
   ```bash
   echo $PULUMI_LOG_LEVEL  # Should show "debug" or similar
   ```

3. **Redirect stderr:**
   ```bash
   pulumi up --verbose 2>&1 | head -100
   # (some systems hide stderr by default)
   ```

### Credentials Visible in Logs

**⚠️ SECURITY ISSUE: If you see actual tokens in logs**

1. **Immediately stop the operation**
2. **Revoke the exposed token in Webflow dashboard**
3. **Generate a new token**
4. **Update Pulumi config:**
   ```bash
   pulumi config set webflow:apiToken $NEW_TOKEN --secret
   ```
5. **Delete any logs containing the old token**
6. **Report to Webflow support**

### Log File Size Issues

**Problem**: `~/.pulumi/logs/` is consuming too much disk space

**Solutions**:

```bash
# Check which files are largest
du -sh ~/.pulumi/logs/* | sort -h | tail -10

# Remove logs older than 60 days
find ~/.pulumi/logs/ -name "*.log" -mtime +60 -delete

# Compress recent logs
gzip ~/.pulumi/logs/pulumi-20231101.log

# Limit future log retention (if configurable in your version)
# Check Pulumi documentation for retention policies
```

### Common Mistakes and Solutions

| Mistake | Solution |
|---------|----------|
| Running with `--verbose` in production | Use `--verbose` only for troubleshooting, disable for normal deployments |
| Committing logs with credentials | Always verify logs with `grep token` before committing |
| Disabling logging entirely | Keep default logging enabled (minimal overhead) |
| Not capturing logs for failures | Add `2>&1 | tee logs.txt` to capture both stdout and stderr |
| Forgetting to rotate old logs | Set up cron job to clean logs older than 30 days |

---

## Examples in This Directory

### TypeScript Troubleshooting Example

`typescript-troubleshooting/` - Demonstrates verbose logging in TypeScript:

```bash
cd typescript-troubleshooting
npm install
pulumi up --verbose
```

**Features:**
- Logging at multiple levels (info, debug)
- Resource creation tracking
- Error handling with logging
- Credential redaction verification

### Python CI/CD Logging Example

`python-cicd-logging/` - Shows environment-aware logging:

```bash
cd python-cicd-logging
pip install -r requirements.txt
pulumi up --verbose
```

**Features:**
- CI/CD environment detection
- Logging based on environment (dev/prod)
- Token source reporting without exposure
- Stack-specific configuration logging

### Go Log Analysis Example

`go-log-analysis/` - Demonstrates structured logging in Go:

```bash
cd go-log-analysis
pulumi up --verbose
```

**Features:**
- Structured logging patterns
- Resource lifecycle tracking
- Diagnostic information logging
- Go best practices for Pulumi

---

## Additional Resources

- [Pulumi Logging Documentation](https://www.pulumi.com/docs/reference/cli/logs/)
- [Webflow API Documentation](https://developers.webflow.com/)
- [Pulumi Troubleshooting Guide](https://www.pulumi.com/docs/troubleshooting/)

---

## Summary

**Key Takeaways:**

1. ✅ **Enable verbose logging** when troubleshooting: `pulumi up --verbose`
2. ✅ **Verify credential redaction** always works: credentials should show as `[REDACTED]`
3. ✅ **Disable verbose logging** for production deployments
4. ✅ **Capture logs** for support tickets and historical analysis
5. ✅ **Use environment-aware logging** in CI/CD pipelines
6. ✅ **Manage log files** to prevent disk space issues

**Remember:** The provider implements automatic credential redaction, so your tokens stay safe even with verbose logging enabled!
