# Webflow Pulumi Provider - State Management Guide

## Overview

The Webflow Pulumi Provider uses the modern `pulumi-go-provider` SDK v1.2.0 which handles state management automatically. This guide explains how state is stored, managed, and used to ensure your infrastructure stays in sync with code-defined configuration.

## How State Works

### State File Location

Pulumi stores resource state in stack-specific state files:

```
.pulumi/stacks/{organization}/{project}/{stack-name}.json
```

Example:
```
.pulumi/stacks/my-org/webflow-project/dev.json
.pulumi/stacks/my-org/webflow-project/prod.json
```

### State File Structure

Each resource's state includes:

```json
{
  "type": "webflow:index:RobotsTxt",
  "urn": "urn:pulumi:dev::webflow-project::webflow:index:RobotsTxt::myRobots",
  "id": "5f0c8c9e1c9d440000e8d8c3/robots.txt",
  "inputs": {
    "siteId": "5f0c8c9e1c9d440000e8d8c3",
    "content": "User-agent: *\nAllow: /\nDisallow: /admin/"
  },
  "outputs": {
    "siteId": "5f0c8c9e1c9d440000e8d8c3",
    "content": "User-agent: *\nAllow: /\nDisallow: /admin/",
    "lastModified": "2025-12-10T12:34:56Z"
  }
}
```

**Key Fields:**
- `id`: Resource identifier used for import/export and drift detection
- `inputs`: Configuration properties you specified in code
- `outputs`: Current state from Webflow API (includes computed properties like `lastModified`)

## Secrets Encryption (NFR12)

### Token Security

The Webflow API token is marked as a secret in the provider configuration:

```go
type Config struct {
    Token string `pulumi:"token,optional" provider:"secret"`
}
```

When Pulumi serializes state, the `provider:"secret"` tag tells Pulumi to **encrypt** this field using your configured secrets provider:

- **Default (passphrase)**: Uses a passphrase to encrypt
- **AWS KMS**: Uses AWS Key Management Service
- **Azure Key Vault**: Uses Azure Key Vault
- **Google Cloud KMS**: Uses Google Cloud KMS

### Verifying Token Encryption

To verify that your token is encrypted in the state file:

```bash
# Export state file
pulumi stack export > state.json

# Check if token is encrypted (should show [secret] or encrypted value)
cat state.json | grep -i token

# Should show something like:
# "token": "[secret]"
# NOT:
# "token": "my-actual-token-value"
```

## Idempotent Operations (FR12, NFR6)

### How Idempotency Works

When you run `pulumi up` multiple times without code changes:

1. **Read**: Pulumi calls `Read()` to fetch current state from Webflow
2. **Diff**: Pulumi calls `Diff()` to compare desired state (from code) vs current state (from Webflow)
3. **No-Op**: If `Diff()` returns `HasChanges: false`, Pulumi **skips the Update** and makes **zero API calls**

This ensures repeated deployments are safe and efficient - your Webflow infrastructure isn't modified unless code changes.

### Example: Idempotent Deployment

```bash
# First deployment - creates resource
$ pulumi up
Updating stack 'dev'...
  Applying changes...
    ✓ webflow:index:RobotsTxt myRobots created

# Second deployment - no code changes
$ pulumi up
Updating stack 'dev'...
  This preview has no changes.

# Zero API calls made to Webflow - only local comparison
```

## State Consistency Under Failure (NFR7)

### Atomic Operations

The SDK ensures all CRUD operations are atomic:

- **Create fails**: No state file created; resource doesn't exist in Pulumi
- **Update fails**: Old state preserved; state file contains previous values (not corrupted)
- **Delete fails**: State file preserved; resource remains in Pulumi state
- **Context cancelled**: Operation stops gracefully; no partial state

### Example: Failed Update Preserves State

```bash
# Resource exists with content "A"
$ pulumi preview
  webflow:index:RobotsTxt myRobots
    content: "User-agent: *\nAllow: /"

# Try to update to content "B" but API fails
$ pulumi up
Error: failed to update robots.txt: webflow API error 500

# State still contains original content "A" - not corrupted
$ pulumi stack export | jq '.deployment.resources[].outputs.content'
"User-agent: *\nAllow: /"
```

## Import Workflows (NFR28)

### Importing Existing Resources

Import an existing Webflow robots.txt configuration into Pulumi:

```bash
# Get your site ID (24-character hex)
SITE_ID="5f0c8c9e1c9d440000e8d8c3"

# Import existing resource
pulumi import webflow:index:RobotsTxt myRobots "$SITE_ID/robots.txt"

# Resource is now in Pulumi state with current Webflow configuration
# pulumi up will maintain this configuration going forward
```

**How It Works:**
1. Pulumi calls `Read()` with the provided ID
2. `Read()` fetches current configuration from Webflow API
3. Pulumi stores fetched configuration in state file
4. Future `pulumi up` commands manage this resource

### Resource ID Format

Resource IDs follow the format:

```
{siteId}/robots.txt
```

Example:
```
5f0c8c9e1c9d440000e8d8c3/robots.txt
```

## Refresh Workflow - Drift Detection

### Detecting Configuration Drift

When you manually change a robots.txt configuration in Webflow UI and want Pulumi to detect this:

```bash
# Your code specifies:
content: "User-agent: *\nAllow: /"

# But someone changed it manually in Webflow to:
# User-agent: *
# Disallow: /

# Refresh state to detect drift:
pulumi refresh

# Pulumi calls Read() for all resources
# State file is updated with current Webflow configuration
# You'll see the drift in preview
```

### Example: Detecting Drift

```bash
$ pulumi refresh
Refreshing stack 'dev'...
Detecting changes...

# Drift detected!
  webflow:index:RobotsTxt myRobots
    content: "User-agent: *\nDisallow: /" (was "User-agent: *\nAllow: /")

# Use pulumi up to correct drift back to code-defined state
$ pulumi up
Updating stack 'dev'...
  ✓ webflow:index:RobotsTxt myRobots updated
```

## Multi-Stack State Isolation

Each Pulumi stack maintains **separate, isolated state**:

```bash
# Dev stack
$ pulumi stack select dev
$ pulumi up  # Manages dev Webflow site

# Prod stack
$ pulumi stack select prod
$ pulumi up  # Manages prod Webflow site - completely separate state
```

State files are stored separately:
- `.pulumi/stacks/{org}/{project}/dev.json`
- `.pulumi/stacks/{org}/{project}/prod.json`

This ensures:
- Dev changes don't affect prod
- Each stack can have different configurations
- Rollback per-stack is safe and isolated

## State Export/Import

### Export State

Export entire stack state to a JSON file:

```bash
pulumi stack export > backup.json

# View exported state
cat backup.json | jq '.deployment.resources[] | select(.type == "webflow:index:RobotsTxt")'
```

### Import State

Restore a previously exported state:

```bash
# WARNING: This overwrites current state!
pulumi stack import < backup.json

# Verify restore
pulumi stack export | diff - backup.json
```

## Security Best Practices

1. **Encrypt State Files**: Always use a secrets provider (not default passphrase in production)
   ```bash
   pulumi config set --secret
   ```

2. **Protect State Files**: Store `.pulumi/` directory securely
   - Add to `.gitignore` if using git
   - Use VCS encryption for `.pulumi/` metadata
   - Restrict access to CI/CD pipelines

3. **Rotate Tokens**: Periodically rotate Webflow API tokens
   ```bash
   pulumi config set webflow:token <new-token> --secret
   pulumi up  # Redeploy with new token
   ```

4. **Audit State Changes**: Track all state modifications
   ```bash
   pulumi history
   pulumi history --show-full  # Shows what changed in each update
   ```

## Troubleshooting

### State Out of Sync

If state doesn't match actual Webflow configuration:

```bash
# Refresh to sync state with current Webflow
pulumi refresh

# Review what changed
pulumi preview

# Correct drift
pulumi up
```

### State File Corruption

If state file becomes corrupted:

```bash
# Export to backup
pulumi stack export > backup.json

# Restore from backup
pulumi stack import < backup.json
```

### Missing State File

If `.pulumi/stacks/{stack-name}.json` is deleted:

```bash
# Recreate state by importing resources
pulumi import webflow:index:RobotsTxt myRobots "$SITE_ID/robots.txt"

# State will be populated from current Webflow configuration
```

## State Management References

- [Pulumi State Management Documentation](https://www.pulumi.com/docs/concepts/state/)
- [Pulumi Secrets Management](https://www.pulumi.com/docs/concepts/secrets/)
- [Pulumi Stack Management](https://www.pulumi.com/docs/concepts/stack/)
