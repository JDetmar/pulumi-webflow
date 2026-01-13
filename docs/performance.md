# Performance Guide

This guide covers performance considerations when using the Pulumi Webflow provider.

## Rate Limiting

### Webflow API Limits

Webflow enforces rate limits on their API. The exact limits depend on your plan:

| Plan | Rate Limit |
|------|------------|
| Free/Basic | 60 requests per minute |
| CMS/Business | 120 requests per minute |
| Enterprise | Custom limits |

### Automatic Retry Handling

The provider automatically handles rate limiting:

- **Retry Strategy**: Exponential backoff with up to 3 retries
- **Retry-After Header**: Respects the `Retry-After` header when present
- **Base Delay**: Starts at 1 second, doubles with each retry
- **Max Delay**: Capped at 30 seconds per retry

You don't need to configure anything - rate limit handling is built-in.

### Best Practices

1. **Batch Operations Thoughtfully**: When managing many resources, consider breaking deployments into smaller batches if you encounter rate limits.

2. **Use `pulumi preview`**: Preview changes before applying to understand the number of API calls that will be made.

3. **Off-Peak Deployments**: For large deployments, consider running during off-peak hours.

## Concurrent Operations

### Pulumi Parallelism

Pulumi executes independent operations in parallel by default. You can control this:

```bash
# Reduce parallelism to minimize rate limit issues
pulumi up --parallel 2

# Disable parallelism entirely
pulumi up --parallel 1
```

### Resource Dependencies

Resources with dependencies are created sequentially. Use `dependsOn` to create explicit dependencies when needed:

```typescript
const collection = new webflow.Collection("my-collection", {
    siteId: site.siteId,
    displayName: "My Collection",
});

const field = new webflow.CollectionField("my-field", {
    collectionId: collection.collectionId,
    displayName: "Title",
    type: "PlainText",
}, { dependsOn: [collection] });
```

## HTTP Client Configuration

The provider uses a configured HTTP client with:

| Setting | Value | Purpose |
|---------|-------|---------|
| Timeout | 30 seconds | Maximum time for a single request |
| TLS Version | 1.2+ | Security compliance |
| Retry Count | 3 | Maximum retry attempts for 429 errors |
| Backoff | Exponential | 1s, 2s, 4s delay progression |

These values are optimized for typical Webflow API usage and cannot be customized.

## Resource-Specific Considerations

### Sites

- Site reads are cached by Pulumi during a single operation
- Site updates may take time to propagate to Webflow's CDN

### Redirects

- Redirect operations are lightweight
- Due to Webflow API limitations, updates trigger delete + create

### Collections and Items

- Collection item operations can be slower for large collections
- Consider paginating if managing hundreds of items

### Assets

- Asset creation returns quickly, but actual file upload is separate
- Use presigned URLs returned by the provider for uploads

## Monitoring and Debugging

### Verbose Logging

Enable verbose logging to see API timing:

```bash
PULUMI_LOG_LEVEL=debug pulumi up
```

### Identifying Bottlenecks

1. Check `pulumi preview` output for resource count
2. Monitor for 429 status codes in logs
3. Consider reducing `--parallel` if rate limits are hit

## Recommendations by Deployment Size

### Small (< 50 resources)

- Default settings work well
- No special configuration needed

### Medium (50-200 resources)

- Consider `--parallel 4` to reduce API pressure
- Review resource dependencies for optimization

### Large (200+ resources)

- Use `--parallel 2` or lower
- Split into multiple stacks if possible
- Schedule deployments during off-peak hours
- Monitor for rate limit errors and adjust

## Troubleshooting Performance Issues

### Slow Deployments

1. Check if rate limits are being hit (look for retries in logs)
2. Reduce parallelism
3. Verify network connectivity to api.webflow.com

### Timeout Errors

1. Check Webflow API status at status.webflow.com
2. Retry the operation
3. If persistent, the resource may require investigation

### Rate Limit Exhaustion

If you consistently hit rate limits:

1. Reduce `--parallel` setting
2. Consider upgrading your Webflow plan
3. Implement deployment scheduling
