---
name: implement-resource
description: Implement a single Webflow API resource. Usage: /implement-resource Collection
allowed-tools: Bash, Read, Write, Grep, Glob, Task
---

# Single Resource Implementation

Implement the Webflow **$ARGUMENTS** resource for the Pulumi provider.

## Step 1: Gather Context

First, read the reference implementation to understand the pattern:

```bash
# Read the redirect resource as the reference pattern
cat provider/redirect_resource.go
cat provider/redirect.go
cat provider/redirect_test.go
```

## Step 2: Fetch Schema from OpenAPI Spec

Get the exact request/response schemas from the official OpenAPI spec:

```bash
# Download spec and extract endpoint schema
curl -s https://raw.githubusercontent.com/webflow/openapi-spec/refs/heads/main/openapi/v2.yml | \
  yq '.paths["/sites/{site_id}/$ARGUMENTS_LOWER"]'
```

Common endpoint patterns from the spec:
- Collection: `/v2/sites/{site_id}/collections`
- Page: `/v2/sites/{site_id}/pages`
- Webhook: `/v2/sites/{site_id}/webhooks`
- Asset: `/v2/sites/{site_id}/assets`
- CustomCode: `/v2/sites/{site_id}/custom_code`

## Step 3: Create Implementation Files

Create these files following the redirect pattern:

### 3.1 API Client: `provider/{resource_lower}.go`

Include:
- Request/Response structs matching Webflow API JSON
- Validation functions with actionable error messages
- Resource ID generation/extraction helpers
- GET, POST, PATCH, DELETE functions with:
  - Context cancellation support
  - Rate limit handling (429) with exponential backoff
  - Proper error handling

### 3.2 Pulumi Resource: `provider/{resource_lower}_resource.go`

Include:
- `{Resource}` controller struct
- `{Resource}Args` input struct with pulumi tags
- `{Resource}State` output struct (embeds Args)
- `Annotate()` methods for descriptions
- `Diff()` - identify replacement vs update
- `Create()` - validate, handle DryRun, call API
- `Read()` - fetch current state
- `Update()` - apply changes
- `Delete()` - remove (404 = success)

### 3.3 Tests: `provider/{resource_lower}_test.go`

Include:
- Validation function tests
- Mock HTTP server tests for each API function
- Error scenario tests (400, 401, 404, 429, 500)

### 3.4 Register in Provider

Add to `provider/provider.go`:
```go
infer.Resource(&{Resource}{}),
```

## Step 4: Verify

```bash
# Build
go build ./provider/...

# Test
go test -v ./provider/... -run {Resource}

# Lint
golangci-lint run ./provider/...
```

## Step 5: Commit

```bash
git add provider/{resource_lower}*.go provider/provider.go
git commit -m "feat({resource_lower}): implement {Resource} resource

- Add {Resource} Pulumi resource with CRUD support
- Add API client for Webflow {Resource} endpoints
- Add validation and comprehensive error handling
- Add test coverage"
```

## Quality Checklist

Before completing:
- [ ] All inputs validated before API calls
- [ ] Error messages explain what's wrong AND how to fix it
- [ ] Rate limiting handled with exponential backoff
- [ ] Delete is idempotent (404 = success)
- [ ] DryRun returns early without API calls
- [ ] Resource ID format: `{siteId}/{type}/{resourceId}`
- [ ] All struct fields have `pulumi:"fieldName"` tags
- [ ] Tests pass
- [ ] Lint passes
