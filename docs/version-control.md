# Version Control Integration for Audit & Compliance

Track all infrastructure changes in Git to create an immutable audit trail for compliance requirements like SOC 2, HIPAA, and GDPR.

## Quick Reference

| Task | Command | Purpose |
|------|---------|---------|
| View infrastructure changes | `git log --oneline -- Pulumi.*` | See all stack config changes |
| Show exact infrastructure diff | `git show <commit>` | View specific change details |
| Generate audit report | `./examples/audit-reports/generate-audit-log.sh` | Create compliance audit report |
| Review by resource type | `git log --grep="resource-type"` | Filter changes by resource |
| Find who changed what | `git log -p -- <file>` | Full change history with diffs |

## Table of Contents

1. [Quick Start](#quick-start)
2. [Git Workflow for Infrastructure](#git-workflow-for-infrastructure)
3. [Commit Message Conventions](#commit-message-conventions)
4. [Pull Request Workflow](#pull-request-workflow)
5. [Audit Trail & Compliance](#audit-trail--compliance)
6. [Generating Audit Reports](#generating-audit-reports)
7. [Multi-Environment Management](#multi-environment-management)
8. [CI/CD Integration](#cicd-integration)
9. [Best Practices](#best-practices)

---

## Quick Start

### Set Up Your Pulumi Project in Git

```bash
# Initialize a new Git repository
git init my-webflow-infrastructure
cd my-webflow-infrastructure

# Create your Pulumi project
pulumi new webflow --dir .

# Create initial commit with proper message format
git add .
git commit -m "feat(infrastructure): initialize Pulumi project for Webflow provider"

# Create a stack-specific config
pulumi stack init production
git add Pulumi.production.yaml
git commit -m "feat(stacks): add production stack configuration"

# Set your Webflow API token (encrypted in stack file)
pulumi config set webflow:apiToken --secret
# The encrypted token is stored in Pulumi.production.yaml and is safe to commit to Git
git add Pulumi.production.yaml
git commit -m "feat(auth): configure Webflow API credentials for production"
```

### Deploy and Track Changes

```bash
# Make an infrastructure change
# Example: Add a new robots.txt resource

# Review the changes that will be made
pulumi preview

# Deploy the changes
pulumi up

# Create a Git commit to record the infrastructure change
git add Pulumi.production.yaml
git commit -m "feat(robotstxt): deploy robots.txt for SEO crawlers (Story 1.5)

- Allows Googlebot, Bingbot, and other major search engines
- Blocks aggressive scrapers and bad bots
- Enables indexing of public pages only

Resolves compliance requirement FR37: Version control audit trail"

# Push to remote repository
git push origin main
```

---

## Git Workflow for Infrastructure

### Recommended Workflow for Teams

```
┌─────────────────────────────────────────────────┐
│          1. Create Feature Branch              │
│    git checkout -b feat/add-new-redirects     │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   2. Make Infrastructure Changes               │
│   - Update Pulumi code in your stack          │
│   - Test with: pulumi preview                 │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   3. Commit Changes with Clear Messages        │
│   git add .                                    │
│   git commit -m "feat(scope): description"    │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   4. Push Branch and Open Pull Request         │
│   git push origin feat/add-new-redirects      │
│   (Create PR on GitHub for review)            │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   5. Code Review & CI/CD Validation            │
│   - GitHub Actions runs test suite             │
│   - Peer review ensures compliance             │
│   - Pulumi preview output visible in PR        │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   6. Merge to Main (Admin approval)            │
│   - Merged commit creates immutable record     │
│   - Git history shows who approved changes     │
└──────────────────────┬──────────────────────────┘
                      │
┌─────────────────────▼──────────────────────────┐
│   7. Production Deployment                     │
│   - CD pipeline triggers on main merge         │
│   - Deployment is traceable to Git commit      │
└─────────────────────────────────────────────────┘
```

### Step-by-Step Example

#### Step 1: Create a Feature Branch

```bash
# Create and switch to a new feature branch
git checkout -b feat/add-compliance-redirects

# Branch name format: type/description
# Types: feat, fix, docs, refactor
```

#### Step 2: Make Your Infrastructure Change

```bash
# Edit your Pulumi code to add resources
vim index.ts

# Preview the changes (see what will happen)
pulumi preview --stack production

# This shows:
# - Which resources will be created/updated/deleted
# - Exact property changes
# - Estimated costs (if applicable)
```

#### Step 3: Commit Your Changes

```bash
# Stage all changes
git add .

# Create a commit with a descriptive message
git commit -m "feat(compliance): add GDPR-compliant redirect rules

- Redirect /privacy to https://privacy-policy.webflow.io
- Redirect /terms to https://terms-of-service.webflow.io
- Add audit logging for all redirects
- Ensures compliance with GDPR cookie policy requirements

Relates to Story 2.2: Redirect CRUD Operations"
```

#### Step 4: Push and Create a Pull Request

```bash
# Push your branch to remote
git push origin feat/add-compliance-redirects

# On GitHub:
# 1. GitHub detects the new branch and suggests creating a PR
# 2. Click "Create Pull Request"
# 3. Fill in the PR template with:
#    - What changed
#    - Why it changed (compliance requirement, bug fix, etc.)
#    - Testing done
#    - Relevant acceptance criteria
```

#### Step 5: Code Review and CI/CD

```bash
# GitHub Actions automatically:
# 1. Runs the test suite
# 2. Checks code quality (linting)
# 3. Runs pulumi preview to show infrastructure changes
# 4. Requires peer review approval

# Team members can:
# - Review the code changes
# - Request modifications if needed
# - Approve when satisfied
```

#### Step 6: Merge and Deploy

```bash
# After approval, merge the PR (done via GitHub web interface)
# This creates a merge commit with all context

# The merge commit includes:
# - Who approved the change
# - Code review comments
# - CI/CD validation results
# - Pulumi preview output
```

---

## Commit Message Conventions

Following commit message conventions makes your Git history readable and useful for audits.

### Format

```
type(scope): subject line (max 50 characters)

Detailed description of the change (max 72 characters per line):
- What changed
- Why it changed
- How it affects compliance
- Related story/issue numbers

Related Story: X.Y (optional)
Resolves: FR## (requirement number, optional)
```

### Types

- **feat**: New feature or infrastructure resource
- **fix**: Bug fix in infrastructure code
- **docs**: Documentation changes
- **refactor**: Code restructuring without behavior change
- **test**: Adding or updating tests
- **chore**: Dependency updates, CI/CD changes

### Examples

#### Example 1: Adding a New Resource

```
feat(robotstxt): add robots.txt for search engine management

- Allows Googlebot, Bingbot, and major search engines
- Blocks scrapers and malicious bots from crawling
- Improves SEO by guiding crawler behavior
- Enables robots.txt updates without manual Webflow configuration

Related Story: 1.5 - RobotsTxt CRUD Operations
Resolves: FR14 - Search engine bot management
```

#### Example 2: Fixing a Configuration Issue

```
fix(auth): correct API token path resolution

- Fixed token path resolution in non-standard environments
- Adds fallback to WEBFLOW_API_TOKEN environment variable
- Improves error messaging for missing credentials

Related Story: 1.2 - Webflow API Authentication
Resolves: BUG-42
```

#### Example 3: Updating Documentation for Compliance

```
docs(audit-trail): document Git history as audit trail

- Added guide for auditors to review infrastructure changes
- Documented commit message standards for compliance
- Added examples of generating audit reports from Git log

Related Story: 7.1 - Version Control Integration for Audit
Resolves: FR37 - Track configuration changes in version control
```

---

## Pull Request Workflow

### Create a Pull Request for Code Review

Pull requests (PRs) enable peer review and CI/CD validation before changes reach production.

### Pull Request Template

Use the PR template when creating pull requests:

```markdown
## Description

What does this PR do? Provide a summary of the changes.

## Type of Change

- [ ] New feature or resource
- [ ] Bug fix
- [ ] Documentation update
- [ ] Configuration change
- [ ] Other (describe)

## Related Story / Requirement

- Story: 7.1 - Version Control Integration for Audit
- Requirement: FR37 - Version control as audit trail
- Issue: #123 (if applicable)

## Acceptance Criteria Satisfied

- [ ] AC1: Git History as Audit Trail
- [ ] AC2: Auditor Review Capability

## Testing Performed

- [ ] Tested with: `pulumi preview --stack <name>`
- [ ] Reviewed Pulumi output for expected changes
- [ ] Verified no unintended resource changes
- [ ] Tested with CI/CD pipeline (if applicable)

## Infrastructure Changes Summary

Show the Pulumi preview output:

```
(Paste the output of `pulumi preview --stack production`)
```

## Checklist

- [ ] Commit messages follow conventional format
- [ ] Code follows project coding standards
- [ ] Documentation is updated
- [ ] All tests pass locally
- [ ] No unintended changes in stack config files
```

### PR Review Process

**For Code Reviewers:**

1. **Review the PR description** - Understand what changed and why
2. **Check the infrastructure impact** - Review the Pulumi preview output
3. **Review the code** - Ensure changes follow standards
4. **Verify compliance** - Confirm changes meet requirements
5. **Request changes** if needed or **Approve**

**Example review comment:**

```
This PR looks good, but I have one question:

The new redirect rule redirects /old-page to /new-page.
Can you confirm:
1. Old page has no active backlinks?
2. Analytics are configured to track the redirect?

Otherwise, this implementation is solid and meets AC2.
```

---

## Audit Trail & Compliance

### What Gets Audited in Git

Git automatically tracks:

| Item | What's Tracked | Audit Value |
|------|---|---|
| **Who** | Author name and email | Identify who made changes |
| **What** | File changes and diffs | See exact infrastructure changes |
| **When** | Timestamp of commit | Timeline of changes |
| **Why** | Commit message | Business justification |
| **Code Review** | PR approvals and comments | Evidence of review process |
| **CI/CD Validation** | Test results in PR | Validation before deployment |

### Example: Auditing a Change

Auditor wants to know what changed on December 15th at 2:30 PM:

```bash
# Find the commit
git log --since="2025-12-15 14:00" --until="2025-12-15 15:00" --oneline

# Output:
# abc1234 feat(redirects): update redirect policy for GDPR (2025-12-15 14:28:00)

# View the exact change
git show abc1234

# Output shows:
# - Author: Alice Johnson <alice@company.com>
# - Date: Dec 15 14:28:00 2025
# - Full change details
# - Commit message with justification
```

### Compliance Requirements Satisfied

#### FR37: Track configuration changes through version control

✅ **Met by:** All infrastructure changes stored in Git with full history

```bash
# All changes are tracked
git log --oneline -- Pulumi.*.yaml
```

#### FR38: Audit configuration changes through Git history

✅ **Met by:** Git history with timestamps, authors, and complete diffs

```bash
# Generate audit report (see below)
./examples/audit-reports/generate-audit-log.sh production
```

#### FR39: Detailed change previews

✅ **Met by:** Pulumi preview in pull requests and Git diffs

```bash
# See exactly what changed
git diff origin/main..HEAD -- Pulumi.*.yaml
```

---

## Generating Audit Reports

### Basic Audit Report

View all infrastructure changes for a time period:

```bash
# Get audit log for last 30 days
git log --since="30 days ago" \
  --format="%h %ai %an %s" \
  -- Pulumi.*.yaml

# Output:
# abc1234 2025-12-15 14:28:00 Alice Johnson feat(redirects): update policy
# def5678 2025-12-14 09:15:00 Bob Smith feat(robots): add crawler rules
# ghi9012 2025-12-13 16:45:00 Alice Johnson fix(auth): correct token path
```

### Detailed Audit Report with Changes

View changes made in a specific date range:

```bash
# See what actually changed
git log --since="2025-12-01" --until="2025-12-31" \
  --name-status \
  -- Pulumi.*.yaml

# Output shows:
# commit abc1234
# Author: Alice Johnson
# Date: Dec 15 14:28:00 2025
#
# M  Pulumi.production.yaml
#
# commit def5678
# ...
```

### Author-Based Audit Report

See all changes by a specific team member:

```bash
# Get all changes by Alice Johnson
git log --author="Alice Johnson" \
  --format="%h %ai %s"

# Get all changes excluding Alice
git log --not --author="Alice Johnson" \
  --format="%h %ai %an %s"
```

### Change-Type Audit Report

View only specific types of changes:

```bash
# See only new resources created
git log --grep="feat(.*resource" --format="%h %ai %an %s"

# See only security-related changes
git log --grep="security\|auth\|token" --format="%h %ai %an %s"

# See only bug fixes
git log --grep="^fix" --format="%h %ai %an %s"
```

### Full Change Audit (Resource-by-Resource)

Generate a detailed report showing exactly what changed in each resource:

```bash
# See exact changes to a specific resource (e.g., redirects)
git log -p --follow -- Pulumi.production.yaml | grep -A 20 "redirects:"

# See all changes to all stack files with author info
git log -p --format="%h %ai %an %s" -- Pulumi.*.yaml
```

### Automated Audit Report Script

Use the provided script to generate compliance-ready audit reports:

```bash
# Generate audit report for production stack
./examples/audit-reports/generate-audit-log.sh production

# Generate report for date range
./examples/audit-reports/generate-audit-log.sh production 2025-12-01 2025-12-31

# Generate report in CSV format for external auditors
./examples/audit-reports/generate-audit-log.sh production --csv

# Output: audit-log-production-2025-12-31.csv
```

---

## Multi-Environment Management

### Stack Configuration Files

Each environment has its own configuration file tracked in Git:

```
my-webflow-infrastructure/
├── Pulumi.dev.yaml          # Development stack
├── Pulumi.staging.yaml      # Staging stack
└── Pulumi.production.yaml    # Production stack
```

### Deploying to Different Environments

```bash
# View all stacks
pulumi stack ls

# Switch to development
pulumi stack select dev
pulumi up

# Switch to staging
pulumi stack select staging
pulumi up

# Switch to production
pulumi stack select production
pulumi up
```

### Auditing Per-Environment Changes

```bash
# See all production-specific changes
git log --oneline -- Pulumi.production.yaml

# See differences between environments
git diff Pulumi.staging.yaml Pulumi.production.yaml

# View when production was last updated
git log -n 1 --format="%h %ai %an" -- Pulumi.production.yaml
```

---

## CI/CD Integration

### GitHub Actions Workflow Example

Create `.github/workflows/infrastructure.yml` to automatically validate changes:

```yaml
name: Validate Infrastructure Changes

on:
  pull_request:
    paths:
      - 'Pulumi.*.yaml'
      - 'index.ts'
      - '.github/workflows/infrastructure.yml'

jobs:
  validate:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4

      - name: Setup Pulumi
        uses: pulumi/actions@v5
        with:
          command: preview
          stack-name: production
        env:
          PULUMI_ACCESS_TOKEN: ${{ secrets.PULUMI_ACCESS_TOKEN }}
          WEBFLOW_API_TOKEN: ${{ secrets.WEBFLOW_API_TOKEN }}

      - name: Comment Preview on PR
        uses: actions/github-script@v7
        with:
          script: |
            // GitHub will automatically show Pulumi preview in the PR
            // This creates an audit trail of all infrastructure changes
```

This creates an audit trail where:
- Every PR shows the exact infrastructure changes
- CI/CD validation proves the change was tested
- GitHub tracks approvals and comments
- The merge commit links to all related discussions

---

## Best Practices

### 1. Commit Frequency

```bash
# ✅ Good: Small, focused commits
git commit -m "feat(redirect): add GDPR privacy policy redirect"

# ❌ Avoid: Large commits mixing multiple changes
git commit -m "update various config files"
```

### 2. Meaningful Commit Messages

```bash
# ✅ Good: Why and what
git commit -m "feat(auth): add token rotation support

- Implements 30-day token rotation policy
- Required by compliance requirement FR25
- Includes automatic renewal before expiry"

# ❌ Avoid: Unclear messages
git commit -m "updates"
```

### 3. Review Before Commit

```bash
# ✅ Good: Review changes before committing
pulumi preview  # See what will change
git diff        # See code changes
git commit      # Commit after review

# ❌ Avoid: Committing without review
git add -A && git commit -m "updates" && git push
```

### 4. Use Pull Requests

```bash
# ✅ Good: Create PR for code review
git push origin feat/new-feature  # Create PR on GitHub
# Wait for review and CI validation

# ❌ Avoid: Direct pushes to main
git push origin main
```

### 5. Tag Releases

```bash
# ✅ Good: Mark production releases with tags
git tag -a v1.0.0 -m "Production release 1.0.0"
git push origin v1.0.0

# This creates an audit trail of production versions
git describe          # Shows current version
git log v1.0.0..HEAD  # Shows changes since last release
```

### 6. Keep Sensitive Data Out of Commits

```bash
# ✅ Good: Use Pulumi secrets for sensitive values
pulumi config set webflow:apiToken --secret
# Token is encrypted in Pulumi.*.yaml and safe to commit

# ❌ Avoid: Committing plain text secrets
# Never commit: WEBFLOW_API_TOKEN=sk_live_abc123...
```

### 7. Document Breaking Changes

```bash
# ✅ Good: Flag breaking changes clearly
git commit -m "BREAKING CHANGE: update redirect schema

- Old format: {source: ..., target: ...}
- New format: {source_url: ..., target_url: ...}
- Migration guide: see docs/migration-v2.md

Resolves: FR40 - Enhanced redirect management"
```

### 8. Link to Requirements

```bash
# ✅ Good: Reference requirements and stories
git commit -m "feat(compliance): add audit log export

Related Story: 7.1 - Version Control Integration
Resolves: FR37 - Track configuration changes
See also: docs/audit-trail.md"
```

---

## Compliance Checklist

Before deploying infrastructure changes:

- [ ] **Code Review**: Changes reviewed by at least one team member
- [ ] **CI/CD Validation**: All tests pass in GitHub Actions
- [ ] **Pulumi Preview**: Infrastructure changes reviewed and approved
- [ ] **Commit Message**: Clear message explaining why change was made
- [ ] **Requirements Link**: Commit references related story/requirement
- [ ] **Secrets**: No sensitive data in commit
- [ ] **Git History**: Complete audit trail preserved
- [ ] **PR Documentation**: PR template completed with details

---

## Summary

Git integration provides:

✅ **Automatic Audit Trail** - Every change tracked with who, what, when, why
✅ **Code Review Process** - Pull requests ensure peer review before deployment
✅ **Immutable History** - Git history cannot be altered or deleted
✅ **Compliance Evidence** - Complete trail for auditors to review
✅ **Rollback Capability** - Easy revert if changes cause issues
✅ **Multi-Environment Support** - Separate tracking for dev/staging/production

For more details on auditing and generating compliance reports, see [Audit Trail Documentation](./audit-trail.md).
