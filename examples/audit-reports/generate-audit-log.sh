#!/bin/bash
#
# Audit Log Generator for Pulumi Infrastructure
#
# Usage: ./generate-audit-log.sh <stack> [--from DATE] [--to DATE] [--csv] [--compliance]
#
# Examples:
#   ./generate-audit-log.sh production
#   ./generate-audit-log.sh production --from 2025-12-01 --to 2025-12-31
#   ./generate-audit-log.sh production --csv > audit.csv
#   ./generate-audit-log.sh production --compliance > compliance-report.txt
#

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
BLUE='\033[0;34m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default values
STACK=""
FROM_DATE=""
TO_DATE=""
CSV_MODE=false
COMPLIANCE_MODE=false
TODAY=$(date +%Y-%m-%d)
THIRTY_DAYS_AGO=$(date -d "30 days ago" +%Y-%m-%d 2>/dev/null || date -v-30d +%Y-%m-%d)

# Get repository URL dynamically
get_repo_url() {
    local remote_url
    remote_url=$(git config --get remote.origin.url 2>/dev/null)
    if [ -z "$remote_url" ]; then
        echo ""
        return
    fi
    # Convert SSH to HTTPS format and remove .git suffix
    echo "$remote_url" | sed 's/git@github.com:/https:\/\/github.com\//' | sed 's/\.git$//'
}
REPO_URL=$(get_repo_url)

# Parse command line arguments
if [ $# -lt 1 ]; then
    echo "Usage: $0 <stack> [--from DATE] [--to DATE] [--csv] [--compliance]"
    echo ""
    echo "Examples:"
    echo "  $0 production"
    echo "  $0 production --from 2025-12-01 --to 2025-12-31"
    echo "  $0 production --csv"
    echo "  $0 production --compliance"
    exit 1
fi

STACK=$1
shift

# Parse optional arguments
while [ $# -gt 0 ]; do
    case "$1" in
        --from)
            FROM_DATE="$2"
            shift 2
            ;;
        --to)
            TO_DATE="$2"
            shift 2
            ;;
        --csv)
            CSV_MODE=true
            shift
            ;;
        --compliance)
            COMPLIANCE_MODE=true
            shift
            ;;
        *)
            echo "Unknown option: $1"
            exit 1
            ;;
    esac
done

# Set default dates if not provided
if [ -z "$FROM_DATE" ]; then
    FROM_DATE=$THIRTY_DAYS_AGO
fi

if [ -z "$TO_DATE" ]; then
    TO_DATE=$TODAY
fi

# Validate stack file exists
STACK_FILE="Pulumi.${STACK}.yaml"
if [ ! -f "$STACK_FILE" ]; then
    echo -e "${RED}Error: Stack file not found: $STACK_FILE${NC}"
    echo "Available stacks:"
    ls -1 Pulumi.*.yaml 2>/dev/null | sed 's/Pulumi\./  /g' | sed 's/\.yaml//g' || echo "  (none found)"
    exit 1
fi

# Function to format table header
print_header() {
    local width=$1
    printf '%-20s | %-10s | %-20s | %-40s\n' "Date Time" "Author" "Commit" "Message"
    printf '%s\n' "$(printf '=%.0s' {1..92})"
}

# Function to format CSV header
print_csv_header() {
    echo "Date,Time,Author,CommitHash,ChangeType,Description,Stack,ChangeUrl"
}

# Function to extract commit type from subject
get_commit_type() {
    local subject="$1"
    if [[ $subject =~ ^feat ]]; then
        echo "feature"
    elif [[ $subject =~ ^fix ]]; then
        echo "bugfix"
    elif [[ $subject =~ ^docs ]]; then
        echo "documentation"
    elif [[ $subject =~ ^refactor ]]; then
        echo "refactoring"
    elif [[ $subject =~ ^test ]]; then
        echo "testing"
    else
        echo "other"
    fi
}

# Function to sanitize CSV values to prevent formula injection
# Prefixes cells starting with =, +, -, or @ with a single quote
sanitize_csv_value() {
    local value="$1"
    # Check if value starts with a dangerous character
    if [[ $value =~ ^[=+\-@] ]]; then
        echo "'$value"
    else
        echo "$value"
    fi
}

# Function to generate compliance report
generate_compliance_report() {
    echo "=================================="
    echo "COMPLIANCE AUDIT REPORT"
    echo "=================================="
    echo ""
    echo "Stack: $STACK"
    echo "Stack File: $STACK_FILE"
    echo "Period: $FROM_DATE to $TO_DATE"
    echo "Generated: $(date '+%Y-%m-%d %H:%M:%S')"
    echo ""

    # Summary statistics
    local total_changes=$(git log --since="$FROM_DATE" --until="$TO_DATE" --oneline -- "$STACK_FILE" 2>/dev/null | wc -l)
    local unique_authors=$(git log --since="$FROM_DATE" --until="$TO_DATE" --format="%an" -- "$STACK_FILE" 2>/dev/null | sort | uniq | wc -l)

    echo "SUMMARY"
    echo "======="
    echo "Total Changes: $total_changes"
    echo "Unique Authors: $unique_authors"
    echo ""

    # Changes by type
    echo "CHANGES BY TYPE"
    echo "==============="
    git log --since="$FROM_DATE" --until="$TO_DATE" --format="%s" -- "$STACK_FILE" 2>/dev/null | \
        awk -F: '{print $1}' | sort | uniq -c | sort -rn || echo "No changes found"
    echo ""

    # Changes by author
    echo "CHANGES BY AUTHOR"
    echo "================="
    git log --since="$FROM_DATE" --until="$TO_DATE" --format="%an" -- "$STACK_FILE" 2>/dev/null | \
        sort | uniq -c | sort -rn || echo "No changes found"
    echo ""

    # Detailed changes
    echo "DETAILED CHANGES"
    echo "================"
    echo ""

    git log --since="$FROM_DATE" --until="$TO_DATE" \
        --pretty=format:"%h|%ai|%an|%s" -- "$STACK_FILE" 2>/dev/null | \
        while IFS='|' read -r hash datetime author subject; do
            # Split datetime into date and time
            date_part=$(echo "$datetime" | cut -d' ' -f1)
            time_part=$(echo "$datetime" | cut -d' ' -f2)

            change_type=$(get_commit_type "$subject")

            printf "%-12s | %-8s | %-20s | %s\n" "$date_part" "$change_type" "$author" "$subject"
            echo "  Commit: $hash"
            echo ""
        done || true

    echo ""
    echo "COMPLIANCE CHECKLIST"
    echo "===================="
    echo "✅ Audit trail captured from Git version control"
    echo "✅ All changes have author identification"
    echo "✅ Timestamps recorded for all changes"
    echo "✅ Change descriptions documented (commit messages)"
    echo "✅ Report generated: $(date '+%Y-%m-%d %H:%M:%S %Z')"
    echo ""
    echo "For compliance review: Ensure all changes were authorized and tested."
}

# Function to generate CSV report
generate_csv_report() {
    print_csv_header

    git log --since="$FROM_DATE" --until="$TO_DATE" \
        --pretty=format:"%h|%ai|%an|%s" -- "$STACK_FILE" 2>/dev/null | \
        while IFS='|' read -r hash datetime author subject; do
            # Split datetime into date and time
            date_part=$(echo "$datetime" | cut -d' ' -f1)
            time_part=$(echo "$datetime" | cut -d' ' -f2)

            change_type=$(get_commit_type "$subject")

            # Sanitize author and subject to prevent CSV formula injection
            author_safe=$(sanitize_csv_value "$author")
            subject_safe=$(sanitize_csv_value "$subject")

            # Escape quotes in subject for CSV
            subject_escaped="${subject_safe//\"/\"\"}"

            if [ -n "$REPO_URL" ]; then
                echo "$date_part,$time_part,$author_safe,$hash,$change_type,\"$subject_escaped\",$STACK,$REPO_URL/commit/$hash"
            else
                echo "$date_part,$time_part,$author_safe,$hash,$change_type,\"$subject_escaped\",$STACK,"
            fi
        done || true
}

# Function to generate table report
generate_table_report() {
    echo ""
    echo -e "${BLUE}Audit Log for Stack: $STACK${NC}"
    echo "Period: $FROM_DATE to $TO_DATE"
    echo ""

    local change_count=$(git log --since="$FROM_DATE" --until="$TO_DATE" --oneline -- "$STACK_FILE" 2>/dev/null | wc -l)

    if [ "$change_count" -eq 0 ]; then
        echo "No changes found for this period."
        echo ""
        return
    fi

    print_header

    git log --since="$FROM_DATE" --until="$TO_DATE" \
        --pretty=format:"%h|%ai|%an|%s" -- "$STACK_FILE" 2>/dev/null | \
        while IFS='|' read -r hash datetime author subject; do
            # Split datetime into date and time
            date_part=$(echo "$datetime" | cut -d' ' -f1)
            time_part=$(echo "$datetime" | cut -d' ' -f2)

            # Truncate long subjects
            subject_short="${subject:0:38}"
            if [ ${#subject} -gt 38 ]; then
                subject_short="${subject_short}..."
            fi

            printf '%-20s | %-10s | %-20s | %-40s\n' "$date_part $time_part" "$author" "$hash" "$subject_short"
        done || true

    echo ""
    echo "Total changes: $change_count"
    echo ""
}

# Main execution
if ! git log --oneline -- "$STACK_FILE" &>/dev/null; then
    echo -e "${RED}Error: Not in a Git repository or $STACK_FILE not tracked in Git${NC}"
    exit 1
fi

if [ "$COMPLIANCE_MODE" = true ]; then
    generate_compliance_report
elif [ "$CSV_MODE" = true ]; then
    generate_csv_report
else
    generate_table_report
fi

exit 0
