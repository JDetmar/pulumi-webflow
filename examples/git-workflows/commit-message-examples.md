# Commit Message Examples

Reference examples of well-written commit messages for infrastructure changes.

## Conventional Commit Format

All commits follow this format:

```
type(scope): subject line (max 50 characters)

Detailed description (max 72 characters per line):
- What changed
- Why it changed
- How it affects the system

Story: X.Y
Requirement: FR## (optional)
```

## Examples by Type

### Feature Commits

#### Example 1: New Resource

```
feat(robots): add robots.txt configuration

Implements initial robots.txt resource to control search engine bot access.

Changes:
- Allow Googlebot, Bingbot, and major search engines
- Block aggressive crawlers and scrapers
- Enable page-specific bot blocking rules

Testing:
- Verified with pulumi preview
- Tested bot access rules match Webflow settings
- Confirmed search engines can still crawl public pages

Story: 1.5 - RobotsTxt CRUD Operations
Requirement: FR14 - Search Engine Bot Management
```

#### Example 2: Infrastructure Enhancement

```
feat(redirects): add wildcard redirect support

Adds support for pattern-based redirects to handle URL migrations.

Features:
- Supports wildcards: /old-* → /new-*
- Pattern matching for complex redirect rules
- Backward compatible with exact-match redirects

Breaking Changes: None
Migration Path: Existing redirects continue to work unchanged

Story: 2.2 - Redirect CRUD Operations
See-Also: docs/version-control.md#redirect-patterns
```

#### Example 3: Compliance Feature

```
feat(compliance): implement audit logging for state changes

Adds comprehensive audit logging to track all infrastructure modifications
for compliance and security requirements.

Implementation:
- Logs all resource create/update/delete operations
- Records timestamp, operator, and change details
- Stores logs in immutable format
- Supports filtering and querying by resource/date/operator

Compliance:
- Satisfies FR37: Version Control Audit Trail
- Enables FR38: Configuration Change Auditing
- Supports SOC 2, HIPAA, GDPR requirements

Testing:
- Unit tests for log generation
- Integration tests for change tracking
- Tested with 1000+ simulated changes

Story: 7.1 - Version Control Integration for Audit
Related-To: Story 7.2, Story 7.3
```

### Bug Fix Commits

#### Example 1: Critical Bug Fix

```
fix(auth): resolve token rotation infinite loop

Fixed a critical issue where token rotation was triggering infinite
loop when token expiration time fell on certain seconds.

Root Cause:
The expiration time comparison was using < instead of <=, causing
tokens that expired on exact second boundaries to never rotate.

Fix:
Changed comparison operator to properly handle boundary conditions.

Impact:
- Critical: Tokens no longer get stuck in retry loops
- High: Eliminates CPU spike from failed rotation attempts
- Medium: Slightly improves error message clarity

Testing:
- Added test for boundary condition times
- Verified with tokens at second boundaries
- Performance test confirms no CPU spike

Fixes: BUG-89 - Token Rotation Loop
Related-To: Story 1.2 - API Authentication
```

#### Example 2: Configuration Bug

```
fix(config): correct site ID validation regex

The site ID validation regex was too restrictive and rejected valid
24-character hex IDs in edge cases.

Issue:
Regex pattern: ^[a-f0-9]{24}$
Problem: Didn't account for uppercase letters that Webflow sometimes returns
Result: Valid IDs were rejected as invalid

Solution:
Updated regex to: ^[a-fA-F0-9]{24}$

Testing:
- Tested with 500+ real Webflow site IDs
- Verified uppercase/lowercase/mixed case
- Confirmed backward compatible

Fixes: BUG-42 - Invalid Site ID Error
Story: 1.2 - Webflow API Authentication
```

### Documentation Commits

#### Example 1: New Documentation

```
docs(version-control): add comprehensive Git workflow guide

Adds detailed documentation for Git workflows, commit message standards,
and best practices for infrastructure changes.

Contents:
- Git workflow diagram with step-by-step instructions
- Commit message conventions and examples
- Pull request template for code review
- Troubleshooting guide for common Git issues
- Multi-environment management examples

File Changes:
- docs/version-control.md (NEW - 800+ lines)
- README.md (UPDATED - added version control section)
- examples/git-workflows/ (NEW - workflow examples)

Story: 7.1 - Version Control Integration for Audit
```

#### Example 2: Update Existing Documentation

```
docs(troubleshooting): add workaround for token timeout

Updates troubleshooting guide with solution for slow API token validation.

Changes:
- Added "Token Validation Timeout" section
- Documented timeout increase workaround
- Added environment variable configuration option
- Linked to related issues

Previous: Did not address timeout scenarios
Updated: Now covers 3 timeout scenarios with solutions

Story: 5.4 - Detailed Logging for Troubleshooting
Relates-To: BUG-67
```

### Refactoring Commits

#### Example 1: Code Organization

```
refactor(config): consolidate stack configuration handling

Refactors stack configuration management to reduce duplication and
improve maintainability without changing functionality.

Before:
- Config loading scattered across 3 files
- Validation logic repeated in multiple places
- Error messages inconsistent

After:
- Single config.ts handles all loading
- Centralized validation logic
- Consistent error messages

Impact:
- Code duplication reduced by 40%
- Configuration tests are more comprehensive
- Error handling is now uniform
- Zero functional changes

Breaking Changes: None
Migration: No migration required

Co-Authored-By: Alice Johnson <alice@company.com>
Co-Authored-By: Bob Smith <bob@company.com>
```

#### Example 2: Performance Improvement

```
refactor(performance): optimize redirect lookup

Refactors redirect lookup to use hash map instead of linear search,
improving performance from O(n) to O(1).

Performance Impact:
- Before: 50ms lookup time for 1000 redirects
- After: <1ms lookup time for 1000 redirects
- 50x faster lookups

Changes:
- Replaced array iteration with Map
- Added index building during initialization
- Updated tests for new structure

Memory Impact:
- Slight increase in memory (< 1% for typical deployments)
- Justified by dramatic performance improvement

All Tests Pass: ✓ 100% (unit, integration, e2e)
Backward Compatible: ✓ Yes

Story: 4.1 - SDK Generation Pipeline
Relates-To: Performance Initiative
```

### Security Commits

#### Example 1: Security Hardening

```
feat(security): implement token rotation policy

Adds automatic API token rotation policy to limit exposure from
compromised tokens.

Security Improvements:
- Tokens rotated every 30 days
- Old tokens kept valid for 7 days during transition
- Automatic rotation with no manual intervention required
- Rotation failures trigger alerts

Implementation:
- New rotation scheduler service
- Token versioning support
- Backward compatibility for old tokens
- Detailed audit logging of rotations

Testing:
- Unit tests for rotation logic
- Integration tests with multiple token versions
- Security review completed (approved by InfoSec team)
- Load testing confirms no performance impact

Compliance:
- Resolves FR25 - API Token Rotation Policy
- Supports HIPAA token rotation requirements
- Aligns with SOC 2 access control standards

Story: 1.2 - Webflow API Authentication
Security-Reviewed-By: InfoSec Team
```

#### Example 2: Security Fix

```
fix(security): prevent credential exposure in logs

Fixes security vulnerability where API credentials could appear in
debug logs under certain error conditions.

Vulnerability:
- Type: Credential Exposure
- Severity: High
- Impact: API tokens could be logged in plaintext
- Discovery: Internal security audit

Fix:
- Implemented credential masking in all logging
- Redact tokens: keep first 4 and last 4 characters only
- Applied to logs, error messages, and debug output
- Added tests to verify masking works

Example:
Before: "token: sk_live_abc123def456ghi789"
After:  "token: sk_l...i789"

Testing:
- Unit tests verify masking pattern
- Integration tests confirm no tokens in logs
- Manual testing with real credentials
- Security review approved

Fixes: VULN-12 - Credential Exposure in Logs
Security-Severity: High
Requires-Deployment: Immediate

Story: 1.2 - Webflow API Authentication
Security-Reviewed-By: Security Team
```

### Merge Commits

```
Merge pull request #42 from jdetmar/feat/add-gdpr-redirects

feat(redirects): add GDPR-compliant redirect rules

Reviewed by: Alice Johnson, Bob Smith
Status: All tests passed, 2 approvals
Release: Scheduled for 1.3.0

Summary of changes:
- Added /privacy redirect to privacy policy
- Added /terms redirect to terms of service
- Implements audit logging for redirects
- Updates documentation with examples

Story: 2.2 - Redirect CRUD Operations
Requirement: FR16 - GDPR Compliance Redirects
```

## Commit Message Anti-Patterns (Examples to Avoid)

### ❌ Bad: Vague Messages

```
// Don't do this:
git commit -m "update config"
git commit -m "fixes"
git commit -m "changes"
git commit -m "WIP"
```

### ❌ Bad: Too Long Subject Line

```
// Don't do this (> 50 chars):
git commit -m "feat(infrastructure): adds new redirect rules for GDPR compliance and also updates the documentation to explain how to use them properly"

// Do this instead:
git commit -m "feat(redirects): add GDPR-compliant redirects

- Adds /privacy and /terms redirects
- Updates documentation with examples
- Implements audit logging"
```

### ❌ Bad: Missing Context

```
// Don't do this:
git commit -m "fix: it now works"

// Do this instead:
git commit -m "fix(auth): resolve token expiration validation

Token comparison logic was using < instead of <=,
causing boundary condition failures on second boundaries."
```

### ❌ Bad: Mixing Multiple Changes

```
// Don't do this:
git commit -m "update redirects, fix auth, add logging"

// Do this instead:
// Create separate commits for each change:
git commit -m "feat(redirects): add GDPR redirects"
git commit -m "fix(auth): resolve token expiration"
git commit -m "feat(logging): add request logging"
```

## Testing Commit Messages

```bash
# View commit message before committing
git commit -m "message" --dry-run

# View recent commit messages
git log --oneline -n 10

# View full commit with message and changes
git show <commit-hash>

# Search commits by message
git log --grep="GDPR"
git log --grep="Story 2.2"

# View commits from specific author
git log --author="Alice" --pretty=fuller
```

## Style Guidelines

- **Subject line**: Concise, descriptive, imperative mood
- **Body paragraphs**: Explain what and why, not how
- **Line length**: 50 chars for subject, 72 for body
- **References**: Include story/requirement numbers
- **Tone**: Professional, clear, factual
- **Co-authorship**: Include when pair programming

## Tools and Integration

### Pre-commit Hook

Use a pre-commit hook to validate commit messages:

```bash
#!/bin/bash
# .git/hooks/pre-commit

# Check if subject line is <= 50 chars
SUBJECT=$(git diff --cached --format=%s -z)
if [ ${#SUBJECT} -gt 50 ]; then
    echo "Error: Commit subject must be <= 50 characters"
    exit 1
fi

# Check if subject starts with valid type
if ! [[ $SUBJECT =~ ^(feat|fix|docs|refactor|test|chore) ]]; then
    echo "Error: Commit must start with type: feat, fix, docs, refactor, test, or chore"
    exit 1
fi

exit 0
```

### Commit Templates

Create `.gitmessage` template:

```
# <type>(<scope>): <subject>
#
# <body>
#
# <footer>

# Types:
#   feat:     A new feature
#   fix:      A bug fix
#   docs:     Documentation only changes
#   refactor: Code change that neither fixes a bug nor adds a feature
#   test:     Adding missing tests
#   chore:    Changes to build/dependencies

# Remember:
#   - Limit subject to 50 characters
#   - Reference stories/requirements in footer
#   - Explain what and why, not how
```

Use with: `git config --global commit.template ~/.gitmessage`

---

## Summary

Well-written commit messages:

✅ Explain the "why" behind the change
✅ Are searchable for future archaeology
✅ Enable fast scanning of history
✅ Support audit and compliance requirements
✅ Help other developers understand intent
✅ Create an executable history of decisions

See [Version Control Guide](../../docs/version-control.md) for complete workflow documentation.
