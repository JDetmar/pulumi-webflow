# Audit Trail & Compliance Documentation

Complete guide for auditors to review infrastructure changes, verify compliance, and generate audit reports using Git history as the source of truth.

## Table of Contents

1. [Audit Trail Overview](#audit-trail-overview)
2. [For Auditors](#for-auditors)
3. [Compliance Requirements](#compliance-requirements)
4. [Generating Audit Reports](#generating-audit-reports)
5. [Change Review Workflows](#change-review-workflows)
6. [Compliance Reporting Templates](#compliance-reporting-templates)
7. [Audit Best Practices](#audit-best-practices)

---

## Audit Trail Overview

### What Is an Audit Trail?

An audit trail is a documented record of all infrastructure changes. The Webflow Pulumi Provider uses Git as the audit trail source:

| Element | Source | Example |
|---------|--------|---------|
| **Change Details** | Git commit and diffs | What configuration changed |
| **Who Made It** | Commit author | `Alice Johnson <alice@company.com>` |
| **When It Changed** | Commit timestamp | `2025-12-15 14:28:00 UTC` |
| **Why It Changed** | Commit message | `feat(compliance): add GDPR redirect rules` |
| **Code Review** | PR comments and approvals | `Approved by Bob Smith` |
| **Validation** | CI/CD pipeline results | `All tests passed` |

### Why Git for Audit Trails?

✅ **Immutable** - History cannot be deleted or modified
✅ **Authenticated** - Cryptographic signatures on commits
✅ **Complete** - Every change tracked from creation to deployment
✅ **Traceable** - Links changes to review process and deployment
✅ **Standards-Based** - Git is industry-standard for audit evidence

---

## For Auditors

### Accessing the Audit Trail

```bash
# 1. Clone the repository
git clone <repository-url>
cd my-webflow-infrastructure

# 2. View the commit history
git log --oneline

# Output shows all changes with one-line summaries:
# abc1234 feat(compliance): add GDPR-compliant redirects
# def5678 feat(auth): configure token rotation
# ghi9012 feat(robots): update search engine rules
```

### Reviewing a Specific Change

```bash
# View complete details of a specific change
git show abc1234

# Output includes:
# - Commit hash: abc1234
# - Author: Alice Johnson <alice@company.com>
# - Date: Sun Dec 15 14:28:00 2025 +0000
# - Commit message with full description
# - Exact code/configuration changes (diffs)
# - Files that were modified
```

### Example Audit Investigation

**Scenario:** An auditor asks "Who changed the authentication configuration on Dec 15?"

```bash
# Find the change
git log --since="2025-12-14" --until="2025-12-16" \
  --format="%h %ai %an %s" \
  -- Pulumi.*.yaml

# Output:
# abc1234 2025-12-15 14:28:00 +0000 Alice Johnson feat(auth): update API token handling

# View the exact change
git show abc1234

# Output shows:
# - What changed in the configuration
# - The exact diff
# - The commit message explaining why
# - Alice Johnson as the author
```

### Verifying Commit Authenticity

```bash
# GPG-signed commits provide cryptographic proof
git log --pretty=fuller --format="%h %G?"

# Output shows:
# N = not signed
# U = unverified
# G = valid signature
# B = bad signature

# View full commit details including signature
git show --show-signature abc1234
```

---

## Compliance Requirements

### Functional Requirement 37: Version Control Integration for Audit

**Requirement:** Track configuration changes through version control

**How It's Met:**

✅ All infrastructure code stored in Git repository
✅ Every configuration change creates a commit
✅ Commit history provides complete change record
✅ Changes cannot be deleted or modified retroactively

**Verification:**

```bash
# Prove all stack files are in Git
git log --follow -- Pulumi.*.yaml | head -20

# This shows the complete history of all changes
```

### Functional Requirement 38: Audit Configuration Changes (Story 7.2)

**Requirement:** Audit personnel can review what changed and who changed it

**How It's Met:**

✅ Git log shows author, timestamp, and change description
✅ Diffs show exactly what was modified
✅ Pull requests show code review and approvals
✅ CI/CD pipeline results prove validation before deployment

**Verification:**

```bash
# Generate audit report showing all changes
git log --format="%ai %an %s" -- Pulumi.*.yaml

# Generate report for specific date range
git log --since="2025-12-01" --until="2025-12-31" \
  --format="%ai %an %s" -- Pulumi.*.yaml
```

### Functional Requirement 39: Detailed Change Previews (Story 7.3)

**Requirement:** Show exact infrastructure changes before deployment

**How It's Met:**

✅ Pulumi preview output in pull requests
✅ Git diffs show configuration changes
✅ Commit history shows before/after state

**Verification:**

```bash
# View exact changes between deployments
git diff HEAD~1 HEAD -- Pulumi.production.yaml

# Show what changed in a specific commit
git show abc1234 -- Pulumi.*.yaml
```

---

## Generating Audit Reports

### Quick Audit Reports

#### Report 1: All Changes in Last 30 Days

```bash
# Simple list of all changes
git log --since="30 days ago" \
  --format="%h | %ai | %an | %s" \
  -- Pulumi.*.yaml

# Output:
# abc1234 | 2025-12-15 14:28:00 | Alice Johnson | feat(redirects): add GDPR rules
# def5678 | 2025-12-14 09:15:00 | Bob Smith | feat(robots): update crawlers
# ghi9012 | 2025-12-13 16:45:00 | Alice Johnson | fix(auth): token handling
```

#### Report 2: Changes by Author

```bash
# See all changes made by a specific person
git log --author="Alice Johnson" \
  --format="%h %ai %s" \
  -- Pulumi.*.yaml

# Count changes per person
git shortlog -sn -- Pulumi.*.yaml
```

#### Report 3: Changes by Type

```bash
# See only new feature additions
git log --grep="^feat" \
  --format="%h %ai %an %s" \
  -- Pulumi.*.yaml

# See only bug fixes
git log --grep="^fix" \
  --format="%h %ai %an %s" \
  -- Pulumi.*.yaml

# See only security-related changes
git log --grep="auth\|security\|token\|secret" \
  --format="%h %ai %an %s" \
  -- Pulumi.*.yaml
```

### Detailed Audit Reports

#### Report 4: Full Change Details (With Diffs)

```bash
# Generate detailed report showing exactly what changed
git log -p --since="2025-12-01" --until="2025-12-31" \
  --format="%h | %ai | %an | %s" \
  -- Pulumi.*.yaml \
  > audit-report-2025-12.txt

# This creates a report showing:
# - Commit hash, date, author, message
# - Exact diffs showing what changed
# - Can be printed or emailed to compliance team
```

#### Report 5: Per-Stack Changes

```bash
# See all changes to production stack
git log --format="%h %ai %an %s" \
  -- Pulumi.production.yaml

# See all changes to staging stack
git log --format="%h %ai %an %s" \
  -- Pulumi.staging.yaml

# Compare how many changes in each environment
echo "Production changes:"
git log --oneline -- Pulumi.production.yaml | wc -l

echo "Staging changes:"
git log --oneline -- Pulumi.staging.yaml | wc -l
```

#### Report 6: Change Timeline

```bash
# Show cumulative changes over time
git log --since="2025-11-01" \
  --reverse \
  --format="%ai | %an | %s" \
  -- Pulumi.*.yaml

# Output shows changes chronologically:
# 2025-11-05 | Alice Johnson | feat: initial setup
# 2025-11-07 | Bob Smith | feat: add first resource
# 2025-11-10 | Alice Johnson | feat: add second resource
```

### Automated Report Generation

Use the provided audit report script:

```bash
# Generate standard audit report
./examples/audit-reports/generate-audit-log.sh production

# Generate CSV for spreadsheet import
./examples/audit-reports/generate-audit-log.sh production --csv > audit.csv

# Generate report for specific date range
./examples/audit-reports/generate-audit-log.sh production \
  --from 2025-12-01 --to 2025-12-31

# Generate compliance report for external auditors
./examples/audit-reports/generate-audit-log.sh production --compliance
```

---

## Change Review Workflows

### Standard Change Workflow

All infrastructure changes follow this documented workflow:

```
Developer: Create feature branch
         ↓
Developer: Make code changes & commit
         ↓
Developer: Push branch and create Pull Request (PR)
         ↓
Code Reviewer: Review changes in PR
         ↓
CI/CD Pipeline: Validate changes (tests, linting, pulumi preview)
         ↓
Approver: Review and approve change
         ↓
Maintainer: Merge PR to main branch (creates merge commit)
         ↓
CD Pipeline: Deploy change to production
         ↓
Audit Trail: Change recorded in Git with full context
```

### Audit-Ready PR Template

Every pull request includes this information:

```markdown
## Description
What infrastructure change is being made and why?

## Story/Requirement
Link to the requirement this change fulfills (e.g., Story 7.1, FR37)

## Acceptance Criteria Satisfied
Which AC are met by this change?

## Testing
How was this change tested?

## Infrastructure Changes Summary
Show the pulumi preview output:
[Pulumi preview output here]

## Reviewers
Who reviewed this change?
- [ ] Security review (if applicable)
- [ ] Compliance review (if applicable)
```

### Audit-Ready Commit Message

Every commit includes:

```
type(scope): subject

Detailed description:
- What changed
- Why it changed
- Business justification
- Compliance requirement number (if applicable)

Related Story: X.Y
Resolves: FR## (requirement)
```

---

## Compliance Reporting Templates

### Template 1: Monthly Compliance Report

```markdown
# Monthly Compliance Audit Report
## December 2025

### Executive Summary
- Total infrastructure changes: 12
- All changes reviewed and approved
- Zero non-compliant changes
- Audit trail 100% complete

### Changes Summary
| Date | Author | Type | Description | Approver | Status |
|------|--------|------|---|---|---|
| 2025-12-15 | Alice | feat | Add GDPR redirects | Bob | ✅ Approved |
| 2025-12-14 | Bob | feat | Update robots.txt | Alice | ✅ Approved |
| ... | ... | ... | ... | ... | ... |

### Compliance Requirements Status
- [ ] FR37: Version control tracking ✅ Met
- [ ] FR38: Audit trail accessible ✅ Met
- [ ] FR39: Change previews available ✅ Met

### Supporting Evidence
- Git repository: [link]
- Pull request history: [link]
- CI/CD pipeline results: [link]

### Conclusion
All infrastructure changes during this period were made according to
documented change control procedures and are fully auditable.
```

### Template 2: Change Justification Report

```markdown
# Change Justification Report
## Change: Add GDPR Privacy Redirect

### Change Details
- Commit: abc1234
- Author: Alice Johnson
- Date: 2025-12-15 14:28:00
- PR: #42

### Business Justification
Required to meet GDPR compliance requirement for privacy policy redirect.

### Technical Details
Configuration change adds new redirect:
- From: /privacy
- To: https://privacy-policy.webflow.io

### Testing
- ✅ Tested with: pulumi preview --stack production
- ✅ CI/CD validation passed
- ✅ Code review approved by Bob Smith

### Approval
- ✅ Code review: Bob Smith
- ✅ Compliance review: N/A
- ✅ Deploy approval: Required before deployment

### Evidence
- Full change: `git show abc1234`
- PR discussion: [link to PR]
- Test results: [link to CI/CD]
```

### Template 3: SOC 2 / HIPAA Audit Checklist

```markdown
# Compliance Audit Checklist

## Version Control & Audit Trail
- [ ] All configuration changes stored in version control
- [ ] Complete history available and immutable
- [ ] Changes traceable to author
- [ ] Timestamps recorded for all changes
- [ ] Change reasons documented (in commit messages)

## Change Control Process
- [ ] Changes reviewed before deployment
- [ ] Peer review evidence (PR approvals)
- [ ] CI/CD validation performed
- [ ] Testing results recorded
- [ ] Approvals documented

## Access Control
- [ ] Repository access restricted to authorized personnel
- [ ] Deployment permissions verified
- [ ] Multi-factor authentication enabled (if applicable)

## Audit Trail Accessibility
- [ ] Audit trail can be retrieved in full
- [ ] Reports can be generated on demand
- [ ] External auditors can access history
- [ ] Change details are complete and readable

## Compliance Verification
- [ ] All required changes identified
- [ ] Each change mapped to requirement
- [ ] Evidence links provided
- [ ] No gaps in audit trail
```

---

## Audit Best Practices

### For Compliance Officers

#### 1. Regular Audits (Monthly)

```bash
# Run monthly audit check
cd my-webflow-infrastructure

# Linux (GNU date):
./examples/audit-reports/generate-audit-log.sh production \
  --from "$(date -d 'first day of last month' +%Y-%m-%d)" \
  --to "$(date -d 'last day of last month' +%Y-%m-%d)" \
  --csv > "audit-report-$(date +%Y-%m).csv"

# macOS (BSD date) - specify dates manually or use this script:
YEAR=$(date +%Y)
MONTH=$(date +%m)
LAST_MONTH=$((MONTH - 1))
[ $LAST_MONTH -eq 0 ] && { LAST_MONTH=12; YEAR=$((YEAR - 1)); }
FIRST_DAY=$(printf "%04d-%02d-01" $YEAR $LAST_MONTH)
LAST_DAY=$(date -j -f "%Y-%m-%d" "$FIRST_DAY" -v+1m -v-1d +%Y-%m-%d)
./examples/audit-reports/generate-audit-log.sh production \
  --from "$FIRST_DAY" \
  --to "$LAST_DAY" \
  --csv > "audit-report-$(printf '%04d-%02d' $YEAR $LAST_MONTH).csv"

# Review the report for:
# - Unusual changes
# - Changes by unauthorized persons
# - Missing documentation
```

#### 2. Track Authorization

```bash
# Verify changes were authorized
git log --format="%ai %an %s" -- Pulumi.*.yaml | while read line; do
  echo "Check if this change was authorized:"
  echo "$line"
done
```

#### 3. Investigate Anomalies

```bash
# Example: Find changes outside normal business hours
git log --format="%ai %an %s" -- Pulumi.*.yaml | grep "23:\|00:\|01:\|02:"

# Example: Find changes by contractors
git log --author="contractor" --format="%ai %s" -- Pulumi.*.yaml
```

#### 4. Maintain Audit Records

```bash
# Export audit records for archival
mkdir audit-records
git log --all --format="%ai | %an | %s" \
  -- Pulumi.*.yaml > audit-records/all-changes.txt

git log --all -p -- Pulumi.*.yaml \
  > audit-records/all-changes-detailed.txt

# Archive these files off-system for compliance retention periods
tar -czf audit-records-2025.tar.gz audit-records/
```

### For Developers

#### 1. Clear Commit Messages

When developers write clear commit messages, auditors can understand why changes were made:

```bash
# ✅ Good: Clear business justification
git commit -m "feat(compliance): add GDPR consent redirect

- Redirects users to privacy policy
- Required by GDPR Article 13 compliance requirement
- Tested with pulumi preview
- Resolves compliance requirement FR37"

# ❌ Bad: No context
git commit -m "update config"
```

#### 2. Link to Requirements

```bash
# Include requirement references
git commit -m "feat(security): add token rotation

Resolves: FR25 - API token rotation policy
Related Story: 2.1 - Authentication requirements"
```

#### 3. Reference Pull Requests

```bash
# Include PR information
git commit -m "feat: add audit trail support

PR: #42
Reviewed by: Alice Johnson, Bob Smith
See also: Story 7.1"
```

---

## Audit Commands Quick Reference

### View Recent Changes

```bash
# Show last 10 infrastructure changes
git log -n 10 --oneline -- Pulumi.*.yaml
```

### Find Changes by Date

```bash
# Changes made on specific date
git log --date=short --after="2025-12-15" --before="2025-12-16" --oneline

# Changes in last week
git log --since="7 days ago" --oneline -- Pulumi.*.yaml
```

### Find Changes by Author

```bash
# All changes by Alice Johnson
git log --author="Alice" --oneline -- Pulumi.*.yaml

# All changes NOT by Alice
git log --not --author="Alice" --oneline -- Pulumi.*.yaml
```

### Find Specific Change Types

```bash
# Security-related changes
git log --grep="auth\|security\|token\|secret" --oneline

# Feature additions
git log --grep="^feat" --oneline

# Bug fixes
git log --grep="^fix" --oneline
```

### Generate Reports

```bash
# Simple text report
git log --format="%h %ai %an %s" -- Pulumi.*.yaml > report.txt

# CSV format for Excel/Sheets (SECURE - prevents formula injection)
# Use the provided script which sanitizes output:
./examples/audit-reports/generate-audit-log.sh production --csv > report.csv

# Alternative: Manual CSV with formula injection protection
git log --format="%h|%ai|%an|%s" -- Pulumi.*.yaml | \
  awk -F'|' '{
    # Sanitize author (field 3) and subject (field 4)
    gsub(/^[=+\-@]/, "'"'&'"'", $3);
    gsub(/^[=+\-@]/, "'"'&'"'", $4);
    # Escape quotes for CSV
    gsub(/"/, "\"\"", $4);
    print $1","$2","$3",\""$4"\"";
  }' > report.csv

# Detailed report with changes
git log -p --format="%h %ai %an %s" -- Pulumi.*.yaml > report-detailed.txt
```

---

## Summary

The Git-based audit trail provides:

✅ **Complete Record** - Every change tracked with author, timestamp, and reason
✅ **Immutable History** - Cannot be deleted or altered retroactively
✅ **Code Review Evidence** - PR approvals and discussions documented
✅ **Automated Validation** - CI/CD results prove changes were tested
✅ **Compliance Ready** - Reports can be generated for auditors
✅ **Accessible** - Audit trail is transparent and reviewable

For version control best practices, see [Version Control Integration Guide](./version-control.md).
