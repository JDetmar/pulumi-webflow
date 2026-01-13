# Logging and Debugging Guide

This guide explains how to use the structured logging features in the Pulumi Webflow provider to debug issues and monitor operations.

## Overview

The Pulumi Webflow provider includes comprehensive structured logging to help you:
- Debug issues during development and production
- Monitor API interactions and resource operations
- Trace rate limiting and retry behavior
- Audit provider operations for compliance

All logging uses Pulumi's native logging framework and respects Pulumi's log level configuration.

## Log Levels

The provider uses standard log levels following Pulumi conventions:

### DEBUG
Detailed information useful for development and debugging:
- API request and response details (with sensitive data redacted)
- Dry-run/preview mode notifications
- Internal operation steps

**Example:**
```
DEBUG Creating Webflow site [workspaceId=abc123, displayName=My Site, shortName=my-site]
DEBUG Calling Webflow API to create site
DEBUG HTTP request completed [method=POST, url=/v2/sites, status=200, attempt=1]
```

### INFO
General operational information about resource lifecycle:
- Resource creation, updates, and deletions
- Successful API operations
- Resource state transitions

**Example:**
```
INFO Creating Webflow site [workspaceId=abc123, displayName=My Site]
INFO Site created successfully [siteId=def456]
INFO Site published successfully
```

### WARN
Potentially problematic situations that don't prevent operations:
- Rate limiting encountered with automatic retry
- Non-fatal issues or limitations
- API limitations (e.g., resources that can't be deleted)

**Example:**
```
WARN Rate limited, retrying after 2s [method=POST, url=/v2/sites, attempt=2, retryAfter=2s]
WARN Asset folder cannot be deleted via API - removing from Pulumi state only [siteId=abc123, folderName=Images]
WARN Deleting Webflow site - this is a destructive operation [siteId=abc123]
```

### ERROR
Errors that prevent an operation from completing:
- API failures with context
- Validation errors
- Authentication failures
- Unexpected API responses

**Example:**
```
ERROR Validation failed: siteId is required [workspaceId=, displayName=My Site]
ERROR Failed to create site via API: 401 Unauthorized
ERROR API returned empty site ID
```

## Enabling Verbose Logging

### Environment Variable
Set the `PULUMI_LOG_LEVEL` environment variable:

```bash
# Enable all logging (DEBUG and above)
export PULUMI_LOG_LEVEL=debug
pulumi up

# Enable INFO and above (default)
export PULUMI_LOG_LEVEL=info
pulumi up

# Enable WARN and above only
export PULUMI_LOG_LEVEL=warning
pulumi up
```

### Command Line Flag
Use the `--verbose` or `--log-level` flag:

```bash
# Maximum verbosity (DEBUG level)
pulumi up --verbose=9

# INFO level
pulumi up --verbose=3

# Specific log level
pulumi up --log-level=debug
```

### Per-Operation
Enable logging for a specific operation:

```bash
# Debug a preview operation
PULUMI_LOG_LEVEL=debug pulumi preview

# Debug an update operation
PULUMI_LOG_LEVEL=debug pulumi up

# Debug a refresh operation
PULUMI_LOG_LEVEL=debug pulumi refresh
```

## Logging in Different Scenarios

### Resource Creation
```
INFO Creating Webflow site [workspaceId=ws123, displayName=Marketing Site, shortName=marketing-site]
DEBUG Calling Webflow API to create site
DEBUG HTTP request completed [method=POST, url=/v2/sites, status=201, attempt=1]
INFO Site created successfully [siteId=site456]
```

### API Rate Limiting
```
DEBUG HTTP request completed [method=POST, url=/v2/collections, status=429, attempt=1]
WARN Rate limited, retrying after 1s [method=POST, url=/v2/collections, attempt=1, retryAfter=1s]
DEBUG HTTP request completed [method=POST, url=/v2/collections, status=201, attempt=2]
INFO Collection created successfully [collectionId=col789]
```

### Validation Errors
```
INFO Creating Webflow site [workspaceId=, displayName=, shortName=]
ERROR Validation failed: workspaceId is required [workspaceId=, displayName=]
```

### Destructive Operations
```
WARN Deleting Webflow site - this is a destructive operation [siteId=site456, displayName=Marketing Site]
DEBUG Calling Webflow API to delete site
DEBUG HTTP request completed [method=DELETE, url=/v2/sites/site456, status=204, attempt=1]
INFO Site deleted successfully [siteId=site456]
```

## Structured Log Fields

Logs include structured fields for easy parsing and filtering:

| Field | Description | Example |
|-------|-------------|---------|
| `siteId` | Webflow site ID | `5f0c8c9e1c9d440000e8d8c3` |
| `workspaceId` | Webflow workspace ID | `ws123abc` |
| `displayName` | Resource display name | `My Site` |
| `fileName` | Asset file name | `logo.png` |
| `method` | HTTP method | `POST`, `GET`, `PATCH`, `DELETE` |
| `url` | API endpoint path | `/v2/sites`, `/v2/collections` |
| `status` | HTTP status code | `200`, `201`, `429`, `401` |
| `attempt` | Retry attempt number | `1`, `2`, `3` |
| `retryAfter` | Retry delay duration | `1s`, `2s`, `4s` |

## Sensitive Data Protection

The provider automatically redacts sensitive information in logs:

### API Tokens
Always redacted as `[REDACTED]`:
```
ERROR Failed to authenticate: token=[REDACTED]
```

### Large Responses
Truncated to prevent log spam:
```
DEBUG Response body: {...}... (truncated, 5234 total chars)
```

### Field-Based Redaction
Fields containing sensitive keywords are automatically redacted:
- `token`, `apiToken`
- `password`, `secret`
- `key`, `authorization`

## Debugging Common Issues

### Issue: "Site not found" after creation
**Enable DEBUG logging to see the API response:**
```bash
PULUMI_LOG_LEVEL=debug pulumi up
```

Look for:
```
DEBUG HTTP request completed [method=POST, url=/v2/sites, status=201, attempt=1]
INFO Site created successfully [siteId=abc123]
```

Verify the `siteId` is correctly returned.

### Issue: Intermittent failures
**Enable DEBUG logging to see retry behavior:**
```bash
PULUMI_LOG_LEVEL=debug pulumi up
```

Look for:
```
WARN Rate limited, retrying after 2s [method=POST, attempt=2]
WARN Rate limit exceeded, max retries exhausted [maxRetries=3]
```

This indicates you're hitting API rate limits.

### Issue: Unexpected resource updates
**Enable INFO logging to see what changed:**
```bash
PULUMI_LOG_LEVEL=info pulumi up
```

Look for:
```
INFO Updating Webflow site [siteId=abc123, displayName=New Name]
INFO Site updated successfully
```

Compare with your code to identify the difference.

### Issue: Authentication failures
**Enable ERROR logging to see the specific error:**
```bash
PULUMI_LOG_LEVEL=error pulumi up
```

Look for:
```
ERROR Failed to create HTTP client: [WEBFLOW_AUTH_001] Webflow API token not configured
ERROR Validation failed: API token cannot be empty
```

## Programmatic Access to Logs

If you're building automation around the provider, you can parse structured logs:

### JSON Format
Pulumi can output logs in JSON format:
```bash
pulumi up --json | jq '.message'
```

### Filtering Specific Events
Filter for specific operations:
```bash
# Show only site creation events
pulumi up --json | jq 'select(.message | contains("Creating Webflow site"))'

# Show only rate limiting events
pulumi up --json | jq 'select(.message | contains("Rate limited"))'

# Show only errors
pulumi up --json | jq 'select(.type == "error")'
```

## Best Practices

1. **Development**: Use `PULUMI_LOG_LEVEL=debug` to see all operations
2. **Production**: Use `PULUMI_LOG_LEVEL=info` for operational visibility
3. **CI/CD**: Use `PULUMI_LOG_LEVEL=warning` to focus on issues
4. **Troubleshooting**: Enable DEBUG temporarily when investigating issues
5. **Audit Trail**: Capture INFO logs for compliance and audit requirements

## Performance Considerations

- **DEBUG logging**: Minimal overhead, logs are generated but may not be displayed
- **Structured fields**: Efficient - no string concatenation until needed
- **Log level**: Respects Pulumi's configuration - logs below threshold are skipped

## Related Documentation

- [Troubleshooting Guide](./troubleshooting.md) - Common issues and solutions
- [Pulumi Logging Documentation](https://www.pulumi.com/docs/support/troubleshooting/#verbose-logging) - Pulumi's logging system
- [Performance Guide](./performance.md) - Optimizing provider operations

## Feedback

If you have suggestions for improving logging or need additional log messages, please open an issue on the [GitHub repository](https://github.com/JDetmar/pulumi-webflow/issues).
