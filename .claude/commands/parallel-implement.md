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

## Token Optimization Strategy

To prevent "Prompt is too long" errors:

1. **Use specialized agents** (`api-implementer`, `api-reviewer`) instead of generic subagents with long prompts
2. **Use haiku model** for all subagents - implementation and review tasks are straightforward
3. **Limit parallelism** to 2-3 resources at a time (not 4+)
4. **Monitor token usage** - if approaching 100k tokens, complete current batch before continuing

**Why this matters:** Each subagent accumulates context (file reads, code generation, test output). With 4 subagents on sonnet model using long embedded prompts, you can hit 200k token limit. With 2-3 subagents on haiku model using specialized agents, you stay well under limits.

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

For each resource, spawn the **api-implementer** subagent using the Task tool:

```
Use Task tool with:
  subagent_type: "api-implementer"
  model: "haiku"  # Use haiku for efficiency - implementation is straightforward
  prompt: "Implement the {RESOURCE} resource for the Pulumi Webflow provider in the worktree at ../pulumi-webflow-{resource_lower}. When complete, commit your changes with an appropriate commit message and report status."

The api-implementer agent already knows:
- Pulumi provider patterns and boilerplate
- Webflow API conventions and error handling
- Testing patterns with mock HTTP servers
- How to structure the three-file pattern (API client, resource, tests)
- How to update provider.go registration
```

**Why use api-implementer instead of generic subagent?**
- ✅ Reduces token usage (no need to repeat instructions)
- ✅ Uses specialized agent's built-in knowledge
- ✅ Easier to maintain (update api-implementer once, benefits everywhere)
- ✅ Consistent implementation patterns across all resources

### Phase 3: Code Review

After each implementation subagent completes, spawn the **api-reviewer** subagent:

```
Use Task tool with:
  subagent_type: "api-reviewer"
  model: "haiku"  # Use haiku for efficiency - reviews are focused
  prompt: "Review the {RESOURCE} implementation in ../pulumi-webflow-{resource_lower}. Verify correctness, consistency with existing patterns, run tests, and report APPROVED or CHANGES_REQUESTED with specific issues."

The api-reviewer agent already knows:
- Pulumi provider best practices and patterns
- Webflow API implementation requirements
- Testing standards and coverage expectations
- How to run builds, tests, and linting
- What to look for in code quality and correctness
```

**Why use api-reviewer instead of generic subagent?**
- ✅ Consistent review standards across all resources
- ✅ Reduced token usage (review checklist is built-in)
- ✅ Specialized knowledge of what good implementations look like

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
