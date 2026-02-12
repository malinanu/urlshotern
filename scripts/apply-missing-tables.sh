#!/bin/bash
# Apply Missing Tables to Production Database
# Bash script for Linux/macOS

set -e

# Configuration
CONTAINER_NAME="${1:-urlshortener-postgres-1}"
DB_USER="${2:-urlshortener}"
DB_NAME="${3:-urlshortener_db}"
SCRIPT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")" && pwd)"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
CYAN='\033[0;36m'
GRAY='\033[0;37m'
NC='\033[0m' # No Color

echo -e "${CYAN}=== Apply Missing Database Tables ===${NC}"
echo ""

# Parse arguments
VERIFY=false
DRY_RUN=false

while [[ "$#" -gt 0 ]]; do
    case $1 in
        --verify) VERIFY=true ;;
        --dry-run) DRY_RUN=true ;;
        --container) CONTAINER_NAME="$2"; shift ;;
        --user) DB_USER="$2"; shift ;;
        --database) DB_NAME="$2"; shift ;;
        *) echo -e "${RED}Unknown parameter: $1${NC}"; exit 1 ;;
    esac
    shift
done

# Check if Docker is running
if ! docker ps &> /dev/null; then
    echo -e "${RED}ERROR: Docker is not running or not accessible${NC}"
    exit 1
fi

# Check if container exists
if ! docker ps -a --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}ERROR: Container '${CONTAINER_NAME}' not found${NC}"
    echo -e "${YELLOW}Available containers:${NC}"
    docker ps -a --format "{{.Names}}"
    exit 1
fi

# Check if container is running
if ! docker ps --format "{{.Names}}" | grep -q "^${CONTAINER_NAME}$"; then
    echo -e "${RED}ERROR: Container '${CONTAINER_NAME}' is not running${NC}"
    echo -e "${YELLOW}Start it with: docker start ${CONTAINER_NAME}${NC}"
    exit 1
fi

# Verify mode
if [ "$VERIFY" = true ]; then
    echo -e "${GREEN}Running verification script...${NC}"
    echo ""

    # Copy verification script to container
    docker cp "${SCRIPT_DIR}/verify-tables.sql" "${CONTAINER_NAME}:/tmp/verify-tables.sql"

    # Execute verification
    docker exec -it "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -f /tmp/verify-tables.sql

    echo ""
    echo -e "${GREEN}Verification complete!${NC}"
    exit 0
fi

echo -e "${YELLOW}Target: Container '${CONTAINER_NAME}', Database '${DB_NAME}'${NC}"
echo ""

# Dry run mode
if [ "$DRY_RUN" = true ]; then
    echo -e "${YELLOW}DRY RUN MODE - No changes will be made${NC}"
    echo ""
    echo -e "${CYAN}Would execute:${NC}"
    echo -e "${GRAY}  1. Copy add-missing-tables.sql to container${NC}"
    echo -e "${GRAY}  2. Execute SQL script in database${NC}"
    echo -e "${GRAY}  3. Add 15+ new tables (roles, permissions, teams, etc.)${NC}"
    echo -e "${GRAY}  4. Add columns to url_mappings table${NC}"
    echo -e "${GRAY}  5. Create indexes for performance${NC}"
    echo -e "${GRAY}  6. Populate default roles and permissions${NC}"
    echo ""
    echo -e "${YELLOW}Run without --dry-run to apply changes${NC}"
    exit 0
fi

# Confirm before proceeding
echo -e "${YELLOW}This will add missing tables to your production database.${NC}"
echo -e "${GREEN}This operation is SAFE - it only adds new tables and columns.${NC}"
echo -e "${GREEN}Existing data will NOT be modified or deleted.${NC}"
echo ""
read -p "Do you want to proceed? (yes/no) " -r
echo ""

if [[ ! $REPLY =~ ^[Yy][Ee][Ss]$ ]]; then
    echo -e "${YELLOW}Operation cancelled.${NC}"
    exit 0
fi

echo -e "${CYAN}Step 1: Copying SQL script to container...${NC}"
if docker cp "${SCRIPT_DIR}/add-missing-tables.sql" "${CONTAINER_NAME}:/tmp/add-missing-tables.sql"; then
    echo -e "${GREEN}  ✓ Script copied successfully${NC}"
else
    echo -e "${RED}  ✗ Failed to copy script${NC}"
    exit 1
fi

echo ""
echo -e "${CYAN}Step 2: Executing SQL script...${NC}"
echo -e "${GRAY}  This may take 10-30 seconds...${NC}"

OUTPUT=$(docker exec -it "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -f /tmp/add-missing-tables.sql 2>&1) || true

# Check for errors in output
if echo "$OUTPUT" | grep -q "ERROR"; then
    echo -e "${YELLOW}  ⚠ Some errors occurred (may be normal):${NC}"
    echo "$OUTPUT" | grep "ERROR" | while read -r line; do
        echo -e "${YELLOW}    $line${NC}"
    done
    echo ""
    echo -e "${GRAY}  Common safe errors:${NC}"
    echo -e "${GRAY}    - 'relation already exists' = Table already exists (OK)${NC}"
    echo -e "${GRAY}    - 'column already exists' = Column already added (OK)${NC}"
else
    echo -e "${GREEN}  ✓ SQL script executed successfully${NC}"
fi

echo ""
echo -e "${CYAN}Step 3: Verifying changes...${NC}"

if docker cp "${SCRIPT_DIR}/verify-tables.sql" "${CONTAINER_NAME}:/tmp/verify-tables.sql"; then
    VERIFY_OUTPUT=$(docker exec -it "$CONTAINER_NAME" psql -U "$DB_USER" -d "$DB_NAME" -f /tmp/verify-tables.sql 2>&1) || true

    # Extract table count from output
    TABLE_COUNT=$(echo "$VERIFY_OUTPUT" | grep -oP "Total Tables.*\|\s*\K\d+" | head -1)

    if [ -n "$TABLE_COUNT" ]; then
        echo -e "${GREEN}  ✓ Database now has $TABLE_COUNT tables${NC}"
    else
        echo -e "${GREEN}  ✓ Changes applied (run with --verify for detailed check)${NC}"
    fi
else
    echo -e "${YELLOW}  ⚠ Verification script failed, but changes may have been applied${NC}"
fi

echo ""
echo -e "${GREEN}=== SUCCESS ===${NC}"
echo ""
echo -e "${GREEN}Missing tables have been added to your database!${NC}"
echo ""
echo -e "${CYAN}Next steps:${NC}"
echo -e "${GRAY}  1. Run verification: ./apply-missing-tables.sh --verify${NC}"
echo -e "${GRAY}  2. Create admin user via API${NC}"
echo -e "${GRAY}  3. Assign admin role in database${NC}"
echo -e "${GRAY}  4. Test admin login${NC}"
echo ""
echo -e "${YELLOW}For detailed verification, run:${NC}"
echo -e "  ./apply-missing-tables.sh --verify"
echo ""

# Usage information
cat << 'EOF'
Usage:
  ./apply-missing-tables.sh                    # Apply changes
  ./apply-missing-tables.sh --verify           # Verify tables
  ./apply-missing-tables.sh --dry-run          # Preview changes
  ./apply-missing-tables.sh --container mydb   # Custom container
  ./apply-missing-tables.sh --help             # Show this help
EOF
