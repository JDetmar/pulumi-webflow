# Token Data Source Examples

This directory contains examples demonstrating how to use the Token data sources (functions) to retrieve information about your Webflow API token and the user who authorized it.

## What You'll Learn

- Retrieve API token authorization details (scopes, rate limits, authorized resources)
- Get information about the user who authorized the token
- Validate token configuration and permissions
- Use token info for conditional logic in your infrastructure

## Available Data Sources

| Function | Description |
|----------|-------------|
| `getTokenInfo` | Returns token authorization details, scopes, rate limits, and authorized site/workspace/user IDs |
| `getAuthorizedUser` | Returns information about the user who authorized the API token |

## Available Languages

| Language   | Directory    | Entry Point    | Dependencies        |
|------------|--------------|----------------|---------------------|
| TypeScript | `typescript/`| `index.ts`     | `package.json`      |

## Quick Start

### TypeScript

```bash
cd typescript
npm install
pulumi stack init dev
pulumi up
```

## Examples Included

### 1. Get Token Information

Retrieve details about the current API token including:
- Authorization ID and creation date
- Last used timestamp
- OAuth scopes granted
- Rate limit (requests per minute)
- Authorized site IDs, workspace IDs, and user IDs
- Application details

### 2. Get Authorized User

Retrieve information about who authorized the token:
- User ID
- Email address
- First and last name

### 3. Conditional Logic Based on Token Scopes

Use token information to conditionally create resources based on available permissions.

## Configuration

This example requires only the Webflow API token to be configured:

| Config Key          | Required | Description                              |
|---------------------|----------|------------------------------------------|
| `webflow:apiToken`  | Yes      | Your Webflow API token (set via env var or config) |

Set via environment variable:
```bash
export WEBFLOW_API_TOKEN=your-api-token-here
```

Or via Pulumi config:
```bash
pulumi config set webflow:apiToken your-api-token-here --secret
```

## Expected Output

After successful deployment, you'll see exports like:

```
Outputs:
    authorizationId     : "55818d58616600637b9a5786"
    authorizedSiteIds   : ["62f3b1f7eafac55d0c64ef91"]
    authorizedUserEmail : "user@example.com"
    authorizedUserName  : "John Doe"
    rateLimit           : 60
    scopes              : "sites:read sites:write assets:read"
```

## Use Cases

### Token Validation

Use `getTokenInfo` to validate that your token has the required permissions before creating resources:

```typescript
const tokenInfo = await webflow.getTokenInfo({});
if (!tokenInfo.authorization.scope.includes("sites:write")) {
    throw new Error("Token requires sites:write scope");
}
```

### Audit Logging

Use `getAuthorizedUser` to track who is making infrastructure changes:

```typescript
const user = await webflow.getAuthorizedUser({});
console.log(`Deployment initiated by: ${user.email}`);
```

### Multi-Site Management

Use the authorized site IDs to dynamically manage resources across all accessible sites:

```typescript
const tokenInfo = await webflow.getTokenInfo({});
tokenInfo.authorization.authorizedTo.siteIds.forEach(siteId => {
    // Create resources for each authorized site
});
```

## Cleanup

Since these are data sources (read-only), there's nothing to destroy. Simply remove the stack:

```bash
pulumi stack rm dev
```

## Related Resources

- [Main Examples Index](../README.md)
- [Webflow API Authentication](https://developers.webflow.com/data/docs/getting-started-data-clients)
