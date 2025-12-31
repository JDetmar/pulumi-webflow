---
name: parallel-implement
description: Orchestrate parallel implementation of Webflow API resources using subagents and git worktrees
allowed-tools: Task, Bash, Read, Write, Grep
---

# Parallel API Implementation Orchestrator

You are coordinating parallel implementation of Webflow API resources for the Pulumi provider.

## Your Role

You are the **orchestrator**. You will:
1. Set up isolated git worktrees for each resource
2. Spawn subagents to implement each resource in parallel
3. Coordinate code reviews
4. Manage the merge process

## Pre-Flight Checks

Before starting, verify:
```bash
# Check we're in the pulumi-webflow repo
git remote -v | grep pulumi-webflow

# Check main branch is clean
git status --porcelain

# List existing worktrees
git worktree list
```

## Workflow

### Phase 1: Setup Worktrees

For each resource requested, create a worktree:

```bash
RESOURCE_LOWER=$(echo "$RESOURCE" | tr '[:upper:]' '[:lower:]' | tr -d ' ')
git worktree add ../pulumi-webflow-${RESOURCE_LOWER} -b feat/${RESOURCE_LOWER}-resource
```

### Phase 2: Spawn Implementation Subagents

For each resource, spawn a subagent using the Task tool with this prompt:

```
Implement the {RESOURCE} resource for the Webflow Pulumi provider.

Working Directory: ../pulumi-webflow-{resource_lower}

## Reference Files (READ THESE FIRST)
1. provider/redirect_resource.go - Reference implementation for resource structure
2. provider/redirect.go - Reference implementation for API client
3. provider/redirect_test.go - Reference implementation for tests
4. API_IMPLEMENTATION_MANIFEST.md - Implementation pattern guide

## Webflow API Documentation
Endpoint: https://developers.webflow.com/data/reference/{resource-path}

## Files to Create

### 1. provider/{resource_lower}.go
Create the API client with:
- Request/Response structs matching Webflow API JSON
- Validation functions (ValidateSiteID already exists in redirect.go)
- Helper functions for resource ID generation/extraction
- GET, POST, PATCH, DELETE functions with:
  - Context cancellation support
  - Rate limit handling (429) with exponential backoff
  - Proper error handling using handleWebflowError()
  - HTTP client from GetHTTPClient()

### 2. provider/{resource_lower}_resource.go  
Create the Pulumi resource with:
- {Resource} struct (empty, controller)
- {Resource}Args struct with pulumi tags
- {Resource}State struct embedding Args
- Annotate() methods for all structs
- Create() - validate inputs, call API, return state
- Read() - fetch current state from API
- Update() - handle changes (or replace if needed)
- Delete() - remove resource, handle 404 gracefully
- Diff() - determine what changes require replacement

### 3. provider/{resource_lower}_test.go
Create tests for:
- All validation functions
- Each API function with mock HTTP server
- Error scenarios (400, 401, 403, 404, 429, 500)

### 4. Update provider/provider.go
Add: infer.Resource(&{Resource}{})

## Implementation Checklist
- [ ] All inputs validated before API calls
- [ ] Error messages are actionable (explain what's wrong, how to fix)
- [ ] Rate limiting handled with exponential backoff
- [ ] Delete is idempotent (404 = success)
- [ ] DryRun checks in Create/Update return early
- [ ] Resource ID format: {siteId}/{resource_type}/{resourceId}
- [ ] Tests pass: go test -v ./provider/... -run {Resource}
- [ ] Linting passes: golangci-lint run ./provider/...

## When Complete
```bash
git add -A
git commit -m "feat({resource_lower}): implement {Resource} resource

- Add {Resource} Pulumi resource with CRUD support
- Add API client for Webflow {Resource} endpoints  
- Add validation and error handling
- Add test coverage"
```

Report: "IMPLEMENTATION COMPLETE" or "BLOCKED: [reason]"
```

### Phase 3: Code Review

After each implementation subagent completes, spawn a review subagent:

```
Review the {RESOURCE} implementation in ../pulumi-webflow-{resource_lower}

## Review Checklist

### Code Quality
- [ ] Follows patterns from redirect_resource.go
- [ ] Proper error handling with actionable messages
- [ ] Input validation before API calls
- [ ] No sensitive data in logs/errors

### Correctness
- [ ] CRUD operations are idempotent
- [ ] Diff() correctly identifies replacement vs update
- [ ] Delete handles 404 gracefully
- [ ] Rate limiting with exponential backoff

### Testing
- [ ] All validation functions tested
- [ ] Happy path tests for each API function
- [ ] Error scenario tests (400, 401, 404, 429, 500)
- [ ] Mock server used (no real API calls in tests)

### Build Verification
```bash
cd ../pulumi-webflow-{resource_lower}
go build ./provider/...
go test -v ./provider/... -run {Resource}
golangci-lint run ./provider/...
```

Report: "APPROVED" or "CHANGES_REQUESTED: [specific issues]"
```

### Phase 4: Merge Coordination

After all implementations are reviewed and approved:

```bash
# Return to main repo
cd /path/to/main/pulumi-webflow

# For each completed resource, merge
git merge ../pulumi-webflow-${RESOURCE_LOWER}/feat/${RESOURCE_LOWER}-resource --no-ff \
  -m "Merge feat/${RESOURCE_LOWER}-resource: Add ${RESOURCE} resource"
```

### Phase 5: Cleanup

```bash
# Remove worktrees after successful merge
git worktree remove ../pulumi-webflow-${RESOURCE_LOWER}
git branch -d feat/${RESOURCE_LOWER}-resource
```

## Conflict Prevention

To minimize merge conflicts, each subagent:
1. Only modifies files specific to their resource
2. Does NOT modify shared files (auth.go, config.go)
3. Adds new entries to provider.go (orchestrator consolidates these)

## Status Tracking

Maintain this table in your context:

| Resource | Worktree | Implement | Review | Merge |
|----------|----------|-----------|--------|-------|
| {name} | ⏳/✅ | ⏳/🔄/✅/❌ | ⏳/🔄/✅/❌ | ⏳/✅ |

## Example Usage

User: "Implement Collection, Page, and Webhook in parallel"

You:
1. Create 3 worktrees (../pulumi-webflow-collection, -page, -webhook)
2. Spawn 3 implementation subagents via Task tool
3. Track status as they complete
4. Spawn review subagents for completed implementations
5. Coordinate merges for approved implementations
6. Clean up worktrees
7. Report final status
