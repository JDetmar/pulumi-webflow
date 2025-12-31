#!/bin/bash
# setup-worktrees.sh
# Creates git worktrees for parallel Webflow API resource implementation

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Default resources to implement (can be overridden)
DEFAULT_RESOURCES="collection page webhook asset custom_code"

usage() {
    echo "Usage: $0 [options] [resource1 resource2 ...]"
    echo ""
    echo "Creates git worktrees for parallel API resource implementation."
    echo ""
    echo "Options:"
    echo "  -l, --list          List available resources"
    echo "  -c, --clean         Remove all worktrees"
    echo "  -s, --status        Show status of all worktrees"
    echo "  -h, --help          Show this help"
    echo ""
    echo "Examples:"
    echo "  $0 collection page webhook    # Create worktrees for specific resources"
    echo "  $0                            # Create worktrees for default resources"
    echo "  $0 --clean                    # Remove all worktrees"
    echo ""
    echo "Available resources:"
    echo "  collection, collection_item, page, custom_domain, custom_code,"
    echo "  registered_script, webhook, asset, asset_folder, form,"
    echo "  form_submission, user, access_group, product, order, inventory"
}

list_resources() {
    echo -e "${GREEN}Available Webflow API Resources:${NC}"
    echo ""
    echo "Priority 1 - Content Management:"
    echo "  collection        - CMS Collections (CRUD)"
    echo "  collection_item   - CMS Collection Items (CRUD, staged + live)"
    echo "  page              - Pages (read-only)"
    echo ""
    echo "Priority 2 - Site Configuration:"
    echo "  custom_domain     - Custom Domains (read-only)"
    echo "  custom_code       - Custom Code injection (site + page level)"
    echo "  registered_script - Registered Scripts (inline + hosted)"
    echo "  webhook           - Webhooks (CRUD)"
    echo ""
    echo "Priority 3 - Assets:"
    echo "  asset             - Assets (CRUD)"
    echo "  asset_folder      - Asset Folders (CRUD)"
    echo ""
    echo "Priority 4 - Forms & Users:"
    echo "  form              - Forms (read-only)"
    echo "  form_submission   - Form Submissions (read + update)"
    echo "  user              - Users (read-only)"
    echo "  access_group      - Access Groups (read-only)"
    echo ""
    echo "Priority 5 - E-commerce (Enterprise):"
    echo "  product           - Products (CRUD)"
    echo "  order             - Orders (read + update)"
    echo "  inventory         - Inventory (read + update)"
}

show_status() {
    echo -e "${GREEN}Current Worktrees:${NC}"
    git worktree list
    echo ""
    
    echo -e "${GREEN}Worktree Status:${NC}"
    for dir in ../pulumi-webflow-*; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            resource=${name#pulumi-webflow-}
            branch=$(git -C "$dir" branch --show-current 2>/dev/null || echo "unknown")
            status=$(git -C "$dir" status --porcelain 2>/dev/null | wc -l | tr -d ' ')
            
            if [ "$status" -eq 0 ]; then
                echo -e "  ${GREEN}✓${NC} $resource (branch: $branch) - clean"
            else
                echo -e "  ${YELLOW}●${NC} $resource (branch: $branch) - $status uncommitted changes"
            fi
        fi
    done
}

clean_worktrees() {
    echo -e "${YELLOW}Removing all pulumi-webflow worktrees...${NC}"
    
    for dir in ../pulumi-webflow-*; do
        if [ -d "$dir" ]; then
            name=$(basename "$dir")
            resource=${name#pulumi-webflow-}
            
            echo -n "  Removing $resource... "
            git worktree remove "$dir" --force 2>/dev/null || rm -rf "$dir"
            echo -e "${GREEN}done${NC}"
            
            # Also try to delete the branch
            branch="feat/${resource}-resource"
            git branch -D "$branch" 2>/dev/null || true
        fi
    done
    
    echo -e "${GREEN}Cleanup complete.${NC}"
}

create_worktree() {
    local resource=$1
    local resource_lower=$(echo "$resource" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')
    local branch="feat/${resource_lower}-resource"
    local worktree_dir="../pulumi-webflow-${resource_lower}"
    
    if [ -d "$worktree_dir" ]; then
        echo -e "  ${YELLOW}⚠${NC}  $resource - worktree already exists at $worktree_dir"
        return 0
    fi
    
    echo -n "  Creating $resource worktree... "
    
    # Create worktree with new branch
    if git worktree add "$worktree_dir" -b "$branch" 2>/dev/null; then
        echo -e "${GREEN}done${NC}"
        
        # Copy API manifest to worktree
        if [ -f "API_IMPLEMENTATION_MANIFEST.md" ]; then
            cp "API_IMPLEMENTATION_MANIFEST.md" "$worktree_dir/"
        fi
        
        return 0
    else
        # Branch might already exist
        if git worktree add "$worktree_dir" "$branch" 2>/dev/null; then
            echo -e "${GREEN}done (existing branch)${NC}"
            return 0
        else
            echo -e "${RED}failed${NC}"
            return 1
        fi
    fi
}

# Parse arguments
RESOURCES=()
while [[ $# -gt 0 ]]; do
    case $1 in
        -l|--list)
            list_resources
            exit 0
            ;;
        -c|--clean)
            clean_worktrees
            exit 0
            ;;
        -s|--status)
            show_status
            exit 0
            ;;
        -h|--help)
            usage
            exit 0
            ;;
        -*)
            echo -e "${RED}Unknown option: $1${NC}"
            usage
            exit 1
            ;;
        *)
            RESOURCES+=("$1")
            shift
            ;;
    esac
done

# Use default resources if none specified
if [ ${#RESOURCES[@]} -eq 0 ]; then
    IFS=' ' read -ra RESOURCES <<< "$DEFAULT_RESOURCES"
fi

# Verify we're in the right directory
if [ ! -f "go.mod" ] || ! grep -q "pulumi-webflow" go.mod 2>/dev/null; then
    echo -e "${RED}Error: Must be run from the pulumi-webflow repository root${NC}"
    exit 1
fi

# Verify main branch is clean
if [ -n "$(git status --porcelain)" ]; then
    echo -e "${YELLOW}Warning: Working directory has uncommitted changes${NC}"
    read -p "Continue anyway? (y/N) " -n 1 -r
    echo
    if [[ ! $REPLY =~ ^[Yy]$ ]]; then
        exit 1
    fi
fi

echo -e "${GREEN}Setting up worktrees for parallel implementation...${NC}"
echo ""

# Create worktrees
for resource in "${RESOURCES[@]}"; do
    create_worktree "$resource"
done

echo ""
echo -e "${GREEN}Setup complete!${NC}"
echo ""
echo "Next steps:"
echo "  1. cd into each worktree directory"
echo "  2. Run 'claude' to start a Claude Code session"
echo "  3. Use /parallel-implement or manually implement the resource"
echo ""
echo "Worktrees created:"
for resource in "${RESOURCES[@]}"; do
    resource_lower=$(echo "$resource" | tr '[:upper:]' '[:lower:]' | tr ' ' '_')
    echo "  ../pulumi-webflow-${resource_lower}"
done
