# Audit Report Examples

Scripts and examples for generating compliance-ready audit reports from Git history.

## Quick Start

### Generate a Basic Audit Report

```bash
# View all infrastructure changes for the last 30 days
./generate-audit-log.sh production

# View changes for a specific date range
./generate-audit-log.sh production 2025-12-01 2025-12-31
```

### Generate a CSV Report for Spreadsheets

```bash
# Export as CSV for Excel/Google Sheets
./generate-audit-log.sh production --csv > audit-report.csv
```

### Generate a Compliance Report

```bash
# Create a compliance-focused report
./generate-audit-log.sh production --compliance > compliance-report.txt
```

## Available Scripts

### generate-audit-log.sh

Main script for generating audit reports from Git history.

**Usage:**

```bash
./generate-audit-log.sh <stack> [--from DATE] [--to DATE] [--csv] [--compliance]
```

**Parameters:**

- `<stack>` - Pulumi stack name (e.g., `production`, `staging`, `dev`)
- `--from DATE` - Start date (YYYY-MM-DD), defaults to 30 days ago
- `--to DATE` - End date (YYYY-MM-DD), defaults to today
- `--csv` - Output as CSV format instead of table
- `--compliance` - Generate compliance-focused report with full details

**Examples:**

```bash
# Report for production stack, last 30 days
./generate-audit-log.sh production

# Report for staging stack, specific date range
./generate-audit-log.sh staging --from 2025-12-01 --to 2025-12-31

# CSV export
./generate-audit-log.sh production --csv

# Full compliance report
./generate-audit-log.sh production --compliance
```

### compliance-report.sh

Generate a formatted compliance report for external auditors.

**Usage:**

```bash
./compliance-report.sh <stack> [--organization NAME] [--period MONTH]
```

**Output:**

Creates a professional report including:
- Executive summary
- Changes by type and date
- Author summary
- Approval chain
- Supporting evidence links

## Audit Report Examples

### Example 1: Monthly Audit Report

```bash
# Generate report for December 2025
./generate-audit-log.sh production \
  --from 2025-12-01 \
  --to 2025-12-31 \
  --compliance > audit-2025-12.txt
```

**Output format:**

```
AUDIT REPORT: Production Stack
Generated: 2025-12-31 20:30:00
Period: 2025-12-01 to 2025-12-31

SUMMARY
=======
Total Changes: 12
Total Authors: 3
Change Types: 8 features, 2 fixes, 2 docs

DETAILED CHANGES
================
2025-12-15 | 14:28:00 | Alice Johnson  | feat(redirects): add GDPR rules
  Commit: abc1234
  Message: Redirect /privacy to privacy-policy.webflow.io
           Requires approval before deployment
           Story: 2.2 - Redirect CRUD Operations

2025-12-14 | 09:15:00 | Bob Smith      | feat(robots): update crawlers
  Commit: def5678
  ...
```

### Example 2: CSV Export for Analysis

```bash
# Export as CSV
./generate-audit-log.sh production --csv > audit.csv
```

**CSV format:**

```csv
Date,Time,Author,CommitHash,Type,Description,Stack
2025-12-15,14:28:00,Alice Johnson,abc1234,feat,Add GDPR redirects,production
2025-12-14,09:15:00,Bob Smith,def5678,feat,Update search crawler rules,production
2025-12-13,16:45:00,Alice Johnson,ghi9012,fix,Correct token handling,production
```

### Example 3: Find Specific Changes

```bash
# Find all security-related changes
git log --grep="auth\|security\|token\|secret" \
  --format="%h | %ai | %an | %s" \
  -- Pulumi.production.yaml

# Find all changes by a specific person
git log --author="Alice Johnson" \
  --format="%h %ai %s" \
  -- Pulumi.*.yaml

# Find changes in a date range
git log --since="2025-12-01" --until="2025-12-31" \
  --format="%ai | %an | %s" \
  -- Pulumi.*.yaml
```

## Compliance Use Cases

### SOC 2 Compliance Audit

```bash
# Generate complete audit trail for SOC 2 review
./generate-audit-log.sh production \
  --from "$(date -d 'first day of this year' +%Y-%m-%d)" \
  --to "$(date +%Y-%m-%d)" \
  --compliance > soc2-audit-trail.txt

# Verify all changes were reviewed
echo "Changes by approval status:"
git log --format="%h %s" -- Pulumi.production.yaml | grep -c "Approved"
```

### HIPAA/GDPR Compliance Report

```bash
# Find all security and privacy-related changes
./generate-audit-log.sh production --compliance | grep -i "security\|privacy\|gdpr\|hipaa"

# Verify encryption and credentials are not exposed
git log -p -- Pulumi.*.yaml | grep -i "password\|token\|secret\|key" | grep -v "# secret" || echo "✅ No exposed secrets found"
```

### Change Management Audit

```bash
# Track all changes and their approvers
git log --format="%h | %ai | %an | %s" -- Pulumi.production.yaml

# Generate timeline of deployments
git log --reverse --format="%ai | %s" -- Pulumi.production.yaml | head -20
```

## Troubleshooting

### Script not found

```bash
# Make script executable
chmod +x generate-audit-log.sh
chmod +x compliance-report.sh

# Run from correct directory
cd /path/to/pulumi-webflow
./examples/audit-reports/generate-audit-log.sh production
```

### No changes found

```bash
# Verify you're in the Git repository root
git log --oneline -- Pulumi.*.yaml | head -5

# If no output, the stack files may not have been committed
git status
```

### Git not found

```bash
# Install Git if not available
brew install git  # macOS
apt-get install git  # Ubuntu/Debian
```

## Integration with CI/CD

### GitHub Actions Example

```yaml
# .github/workflows/compliance-audit.yml
name: Monthly Compliance Audit

on:
  schedule:
    - cron: '0 0 1 * *'  # Run on first of each month

jobs:
  audit:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v4
        with:
          fetch-depth: 0  # Full history

      - name: Generate Compliance Report
        run: |
          chmod +x ./examples/audit-reports/generate-audit-log.sh
          ./examples/audit-reports/generate-audit-log.sh production \
            --compliance > compliance-report-$(date +%Y-%m).txt

      - name: Upload Report
        uses: actions/upload-artifact@v3
        with:
          name: compliance-reports
          path: compliance-report-*.txt
```

## Advanced Queries

### Find Changes Affecting Specific Resources

```bash
# Find all changes mentioning "redirect" or "robots"
git log --grep="redirect\|robots" \
  --format="%h %ai %an %s" \
  -- Pulumi.production.yaml
```

### Generate Timeline of Changes

```bash
# Show changes in reverse chronological order with commit details
git log --reverse \
  --format="%ai | %an | %s" \
  -- Pulumi.production.yaml
```

### Statistics Report

```bash
# Count changes per author
git shortlog -sn -- Pulumi.*.yaml

# Count changes per type
git log --format="%s" -- Pulumi.*.yaml | cut -d':' -f1 | sort | uniq -c

# Average time between changes
git log --format="%ai" -- Pulumi.*.yaml | \
  awk '{cmd="date -d \""$1"\" +%s"; cmd | getline epoch; close(cmd); print epoch}' | \
  awk '{if(prev) print $1-prev; prev=$1}'
```

## See Also

- [Version Control Integration Guide](../../docs/version-control.md) - Complete Git workflow documentation
- [Audit Trail Documentation](../../docs/audit-trail.md) - Compliance reporting and best practices
- [Git Log Documentation](https://git-scm.com/docs/git-log) - Full git log reference
