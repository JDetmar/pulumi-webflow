# Pulumi Webflow Provider - Parallel Implementation Setup

This directory contains Claude Code commands, agents, and scripts for implementing Webflow API resources in parallel.

## Quick Start

### Option 1: Single Resource (Simplest)

In your pulumi-webflow repo with Claude Code:

```bash
# Implement one resource at a time
/implement-resource Collection
```

### Option 2: Parallel with Subagents (Recommended)

```bash
# Use the orchestrator to manage multiple implementations
/parallel-implement Collection,Page,Webhook
```

### Option 3: Manual Worktrees + Multiple Terminals

```bash
# Run the setup script to create worktrees
./scripts/setup-worktrees.sh collection page webhook

# Open separate terminals for each worktree
cd ../pulumi-webflow-collection && claude
cd ../pulumi-webflow-page && claude
cd ../pulumi-webflow-webhook && claude

# In each terminal, use
/implement-resource {ResourceName}
```

## Files Overview

```
pulumi-webflow/
├── .claude/
│   ├── commands/
│   │   ├── parallel-implement.md    # Orchestrates parallel work
│   │   └── implement-resource.md    # Single resource implementation
│   └── agents/
│       ├── api-implementer.md       # Implementation specialist
│       └── api-reviewer.md          # Code review specialist
├── scripts/
│   └── setup-worktrees.sh           # Creates git worktrees
└── docs/
    └── API_IMPLEMENTATION_MANIFEST.md   # List of APIs to implement
```

## Commands Reference

### `/implement-resource {ResourceName}`

Implements a single resource in the current directory.

**Usage:**
```
/implement-resource Collection
/implement-resource Webhook
/implement-resource CustomCode
```

**What it does:**
1. Reads reference implementation (redirect_resource.go)
2. Creates {resource}.go, {resource}_resource.go, {resource}_test.go
3. Registers resource in provider.go
4. Runs tests and linting

### `/parallel-implement {Resources}`

Orchestrates parallel implementation using subagents and worktrees.

**Usage:**
```
/parallel-implement Collection,Page,Webhook
```

**What it does:**
1. Creates git worktrees for each resource
2. Spawns subagents to implement each resource
3. Coordinates code reviews
4. Manages merge process

## Scripts Reference

### `scripts/setup-worktrees.sh`

Creates git worktrees for parallel development.

```bash
# Create worktrees for specific resources
./scripts/setup-worktrees.sh collection page webhook

# List available resources
./scripts/setup-worktrees.sh --list

# Show status of existing worktrees
./scripts/setup-worktrees.sh --status

# Clean up all worktrees
./scripts/setup-worktrees.sh --clean
```

## Available Resources

See `docs/API_IMPLEMENTATION_MANIFEST.md` for the complete list. Priority order:

| Priority | Resources | Notes |
|----------|-----------|-------|
| 1 | Collection, CollectionItem, Page | CMS foundation |
| 2 | CustomDomain, CustomCode, RegisteredScript, Webhook | Site config |
| 3 | Asset, AssetFolder | Media management |
| 4 | Form, FormSubmission, User, AccessGroup | Forms & users |
| 5 | Product, Order, Inventory | E-commerce (enterprise) |

## Implementation Pattern

Each resource follows the pattern established by the existing `Redirect` resource:

```
provider/
├── {resource}.go           # API client (HTTP calls, validation)
├── {resource}_resource.go  # Pulumi resource (CRUD, state)
└── {resource}_test.go      # Tests
```

### Key Requirements

1. **Validation before API calls** - All inputs validated with actionable error messages
2. **Rate limit handling** - Exponential backoff on 429 responses
3. **Idempotent delete** - 404 response = success (already deleted)
4. **DryRun support** - Return early during `pulumi preview`
5. **Resource ID format** - `{siteId}/{type}/{resourceId}`

## Workflow Tips

### For Best Results with Parallel Implementation

1. **Start with lower complexity resources** - Page, CustomDomain, Form (read-only)
2. **Then medium complexity** - Webhook, Asset, CustomCode
3. **Save high complexity for last** - CollectionItem (staged + live items)

### Avoiding Merge Conflicts

- Each agent only modifies files for their specific resource
- Don't modify shared files (auth.go, config.go)
- Register resources in provider.go AFTER merging implementations

### If a Subagent Gets Stuck

1. Check the worktree: `cd ../pulumi-webflow-{resource}`
2. Review what was created: `git status`
3. Run tests manually: `go test -v ./provider/... -run {Resource}`
4. Continue in that terminal with Claude Code

## Troubleshooting

### "Worktree already exists"

```bash
./scripts/setup-worktrees.sh --clean
./scripts/setup-worktrees.sh collection page webhook
```

### Tests failing with network errors

The tests should use mock HTTP servers, not real API calls. Check that:
- `get{Resource}BaseURL` variable is set in test
- `defer func() { get{Resource}BaseURL = "" }()` cleans up

### Merge conflicts in provider.go

After merging all implementations, manually consolidate the `WithResources()` calls:

```go
WithResources(
    infer.Resource(&SiteResource{}),
    infer.Resource(&Redirect{}),
    infer.Resource(&RobotsTxt{}),
    infer.Resource(&Collection{}),      // New
    infer.Resource(&Page{}),            // New
    infer.Resource(&Webhook{}),         // New
)
```
