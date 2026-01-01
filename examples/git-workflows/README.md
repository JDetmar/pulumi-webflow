# Git Workflow Examples

Complete examples of Git workflows for managing Webflow infrastructure changes.

## Table of Contents

1. [Basic Feature Branch Workflow](#basic-feature-branch-workflow)
2. [Multi-Environment Workflow](#multi-environment-workflow)
3. [Release Management](#release-management)
4. [Hotfix Workflow](#hotfix-workflow)
5. [Commit Message Examples](#commit-message-examples)

---

## Basic Feature Branch Workflow

The recommended workflow for most infrastructure changes.

### Step-by-Step Example

```bash
# 1. Start with a clean main branch
git checkout main
git pull origin main

# 2. Create a feature branch
git checkout -b feat/add-gdpr-redirects

# 3. Make your infrastructure changes
# Edit index.ts, Pulumi.dev.yaml, etc.
vim index.ts

# 4. Test locally
pulumi preview

# 5. Commit your changes with a clear message
git add .
git commit -m "feat(redirects): add GDPR-compliant redirect rules

- Redirect /privacy to privacy-policy.webflow.io
- Redirect /terms to terms-of-service.webflow.io
- Adds audit logging for compliance tracking

Resolves: Story 2.2 - Redirect CRUD Operations"

# 6. Push your branch
git push origin feat/add-gdpr-redirects

# 7. Create a Pull Request on GitHub
# (GitHub will prompt you with a link)
```

### Pull Request Template

When you create the PR, fill in this template:

```markdown
## Description

Adds GDPR-compliant redirect rules to the production stack.

## Type of Change

- [x] New feature or resource
- [ ] Bug fix
- [ ] Documentation update
- [ ] Configuration change

## Related Story

Story: 2.2 - Redirect CRUD Operations
Requirement: FR16 - GDPR Compliance Redirects

## Acceptance Criteria Satisfied

- [x] Redirects from /privacy and /terms are working
- [x] Audit logging is configured
- [x] Infrastructure change preview is accurate

## Testing Performed

```
$ pulumi preview
Previewing update (production):

     Type                      Name            Plan       Info
 ~   webflow:Redirect          privacy         update
 +   webflow:Redirect          terms           create

Resources:
    + 1 to create
    ~ 1 to update
```

## Changes Summary

- Modified: Pulumi.production.yaml (added 2 new redirects)
- No deleted resources
- No breaking changes
```

### After Code Review

Once your PR is reviewed and approved:

```bash
# 1. GitHub will notify you of approval
# 2. You or a maintainer merges the PR
# 3. This creates a merge commit with all context

# 4. Delete the feature branch (optional, GitHub can do this)
git branch -d feat/add-gdpr-redirects
git push origin --delete feat/add-gdpr-redirects

# 5. Pull the latest main to get the merge commit
git checkout main
git pull origin main
```

---

## Multi-Environment Workflow

Managing separate dev, staging, and production environments.

### Scenario: Deploy to Multiple Environments

```bash
# 1. Create feature branch
git checkout -b feat/add-sitemap-redirect

# 2. Make changes and test in development
git add .
git commit -m "feat(redirects): add sitemap redirect"

# 3. Deploy to development stack
pulumi stack select dev
pulumi up

# 4. Deploy to staging stack
pulumi stack select staging
pulumi up

# 5. Create PR for code review before production
git push origin feat/add-sitemap-redirect
# Create PR on GitHub

# 6. After approval, merge to main
# GitHub: Click "Merge pull request"

# 7. Deploy to production
git checkout main
git pull origin main
pulumi stack select production
pulumi preview
pulumi up
```

### Stack-Specific Configuration Files

```bash
# Each stack has its own configuration file
# These are all committed to Git

Pulumi.dev.yaml         # Development configuration
Pulumi.staging.yaml     # Staging configuration
Pulumi.production.yaml  # Production configuration
```

### Viewing Environment-Specific Changes

```bash
# See what changed in production stack
git log --oneline -- Pulumi.production.yaml

# Compare production vs staging config
git diff Pulumi.staging.yaml Pulumi.production.yaml

# See all changes across all stacks
git log --oneline -- Pulumi.*.yaml
```

---

## Release Management

Tagging releases for version tracking and audit trails.

### Creating a Release

```bash
# 1. Complete all changes and commits
# 2. Make sure everything is tested

# 3. Create a release tag
git tag -a v1.2.0 -m "Release 1.2.0: GDPR Compliance Features

- Added GDPR-compliant redirects
- Implemented audit logging
- Updated documentation

Date: 2025-12-15
Deployed to production by: Alice Johnson"

# 4. Push the tag
git push origin v1.2.0

# 5. GitHub automatically creates a release with the tag message
```

### Viewing Release History

```bash
# See all releases
git tag -l

# See changes since last release
git log v1.1.0..v1.2.0 --oneline

# Show release tag details
git show v1.2.0

# List releases with dates
git log --tags --simplify-by-decoration --pretty="format:%d %ai" | head -10
```

### Rollback to Previous Release

```bash
# If something breaks in production
git log --oneline --all | grep -i "v1.1" | head -5

# Revert to previous release tag
git checkout v1.1.0 -- Pulumi.production.yaml

# Or check out the full state from a release
git checkout v1.1.0

# Then recreate the branch and deploy
git checkout -b rollback-v1.1.0
git push origin rollback-v1.1.0
```

---

## Hotfix Workflow

Handling urgent production issues.

### Scenario: Production Issue Needs Immediate Fix

```bash
# 1. Create a hotfix branch from production tag
git checkout -b hotfix/critical-redirect-fix v1.2.0

# 2. Make the urgent fix
vim index.ts
git commit -m "fix(redirects): correct GDPR redirect URL

The /privacy redirect was pointing to wrong URL.
Updated to correct privacy policy page.

Fixes: BUG-123 - Production hotfix"

# 3. Test the fix
pulumi preview
pulumi up

# 4. Push hotfix branch
git push origin hotfix/critical-redirect-fix

# 5. Create PR with high priority
# Add label: "urgent", "hotfix", "production"

# 6. After rapid review and approval, merge
# Merge hotfix PR to both main and release branches

git checkout main
git pull origin main
# (Pull request is merged)

# 7. Also apply to previous release branch if needed
git checkout release-v1.2
git merge main
git push origin release-v1.2

# 8. Tag as patch release
git tag -a v1.2.1 -m "Hotfix 1.2.1: Fix critical redirect bug"
git push origin v1.2.1
```

---

## Commit Message Examples

### Example 1: New Feature

```
feat(redirects): add international redirect rules

- Adds redirects for /de (German), /fr (French), /es (Spanish)
- Routes to locale-specific landing pages
- Implements country detection for automatic routing
- Improves international SEO compliance

Story: 3.1 - Multi-Language Site Support
Related-To: FR18 - International Compliance
Tested-With: pulumi preview (verified 3 new redirects)
```

### Example 2: Bug Fix

```
fix(robotstxt): correct bot exclusion pattern

The previous regex pattern was too broad and excluded
legitimate search engine crawlers. Updated pattern:

Old: /admin.*
New: /admin(/.*)?$ with strict validation

Fixes: BUG-42 - Googlebot blocked from indexing
Resolves: Story 1.5 - RobotsTxt CRUD Operations
```

### Example 3: Compliance/Security Change

```
feat(compliance): implement audit logging for all changes

- Log all infrastructure modifications
- Record timestamp, author, and change description
- Enable compliance officer review
- Supports SOC 2 and HIPAA audit requirements

Resolves: FR37 - Version Control Audit Trail
Related-To: Story 7.1 - Version Control Integration
Compliance: SOC 2, HIPAA, GDPR
```

### Example 4: Documentation Update

```
docs(version-control): add audit trail guide

- Add comprehensive version control integration guide
- Document Git workflow best practices
- Add examples of audit report generation
- Include compliance templates for auditors

Story: 7.1 - Version Control Integration for Audit
See-Also: docs/audit-trail.md, docs/version-control.md
```

### Example 5: Refactoring

```
refactor(infrastructure): simplify redirect configuration

Consolidate multiple redirect definitions into a single
configurable structure for better maintainability.

Before:
  const redirects = [
    {source: '/old', target: '/new'},
    {source: '/old2', target: '/new2'},
    // ... 20 more entries
  ]

After:
  const redirects = loadRedirectsFromConfig()

- Reduced code duplication
- Makes configuration management easier
- No functional changes
- All tests pass

Co-Authored-By: Bob Smith <bob@company.com>
```

---

## Advanced Git Commands for Infrastructure

### Finding Changes Related to Specific Resources

```bash
# Find all redirect-related changes
git log --grep="redirect" --oneline

# Find changes by a specific author
git log --author="Alice Johnson" --oneline -- Pulumi.*.yaml

# Find changes in a specific date range
git log --since="2025-12-01" --until="2025-12-31" --oneline
```

### Reviewing Code Before Committing

```bash
# See what changed since last commit
git diff

# See what was staged for commit
git diff --staged

# See full change history of a file
git log -p -- Pulumi.production.yaml

# Compare two commits
git diff abc1234 def5678
```

### Undoing Changes

```bash
# Unstage a file
git reset Pulumi.production.yaml

# Discard changes in working directory
git checkout -- Pulumi.production.yaml

# Undo last commit (keep changes)
git reset --soft HEAD~1

# Undo last commit (discard changes)
git reset --hard HEAD~1

# Revert a commit (creates new commit undoing it)
git revert abc1234
```

### Cherry-Picking Changes

```bash
# Apply a specific commit to current branch
git cherry-pick abc1234

# Apply multiple commits
git cherry-pick abc1234 def5678 ghi9012

# Apply commits from a range
git cherry-pick abc1234..ghi9012
```

---

## CI/CD Integration Examples

### GitHub Actions Workflow

See `.github/workflows/` for complete examples of:
- Automatic testing on PR
- Infrastructure preview in PR comments
- Automated deployment on merge to main
- Audit log generation

### Manual Deployment Commands

```bash
# Deploy from main branch
git checkout main
git pull origin main
pulumi stack select production
pulumi preview
pulumi up

# Verify deployment
git log -1 --oneline -- Pulumi.production.yaml
pulumi stack

# Tag the deployment
git tag -a deployment-$(date +%Y-%m-%d-%H%M%S) -m "Deployment to production"
git push origin $(git describe --tags)
```

---

## Best Practices Summary

✅ **DO:**
- Use descriptive commit messages
- Create feature branches for changes
- Use pull requests for code review
- Test changes before committing
- Reference stories/requirements in commits
- Tag releases for version tracking
- Keep commits focused and atomic

❌ **DON'T:**
- Commit directly to main
- Mix multiple unrelated changes in one commit
- Use vague commit messages like "updates"
- Skip code review
- Deploy without testing
- Commit sensitive data
- Force push to shared branches

---

## Additional Resources

- [Version Control Integration Guide](../../docs/version-control.md) - Complete guide
- [Audit Trail Documentation](../../docs/audit-trail.md) - Compliance reporting
- [Git Documentation](https://git-scm.com/docs) - Official Git reference
- [Conventional Commits](https://www.conventionalcommits.org/) - Standard commit format
