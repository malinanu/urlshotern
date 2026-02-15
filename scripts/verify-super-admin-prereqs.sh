#!/bin/bash
# ============================================================================
# Super Admin Migration Prerequisites Verification Script
# ============================================================================
# Purpose: Verify that all prerequisites are met before applying migration 033
# Usage: ./scripts/verify-super-admin-prereqs.sh

set -e

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
USER_ID="112125335771807740"
ROLE_NAME="super_admin"

echo -e "${BLUE}========================================${NC}"
echo -e "${BLUE}Super Admin Migration Prerequisites Check${NC}"
echo -e "${BLUE}========================================${NC}\n"

# Function to run SQL query
run_query() {
    docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db -t -c "$1"
}

# Check 1: PostgreSQL container is running
echo -e "${BLUE}[1/5] Checking PostgreSQL container...${NC}"
if docker ps -q -f name=postgres | grep -q .; then
    echo -e "${GREEN}✓ PostgreSQL container is running${NC}\n"
else
    echo -e "${RED}✗ PostgreSQL container not found${NC}"
    echo -e "${YELLOW}Please start your Docker containers first${NC}\n"
    exit 1
fi

# Check 2: Database connection
echo -e "${BLUE}[2/5] Testing database connection...${NC}"
if run_query "SELECT 1;" | grep -q "1"; then
    echo -e "${GREEN}✓ Database connection successful${NC}\n"
else
    echo -e "${RED}✗ Cannot connect to database${NC}\n"
    exit 1
fi

# Check 3: User exists
echo -e "${BLUE}[3/5] Checking if user $USER_ID exists...${NC}"
USER_INFO=$(run_query "SELECT id, name, email FROM users WHERE id = $USER_ID;" | xargs)
if [ -n "$USER_INFO" ]; then
    echo -e "${GREEN}✓ User found:${NC}"
    echo -e "  ${USER_INFO}\n"
else
    echo -e "${RED}✗ User $USER_ID not found in database${NC}"
    echo -e "${YELLOW}Please verify the user ID or create the user first${NC}\n"
    exit 1
fi

# Check 4: Super admin role exists
echo -e "${BLUE}[4/5] Checking if super_admin role exists...${NC}"
ROLE_INFO=$(run_query "SELECT id, name, display_name FROM roles WHERE name = '$ROLE_NAME';" | xargs)
if [ -n "$ROLE_INFO" ]; then
    echo -e "${GREEN}✓ Role found:${NC}"
    echo -e "  ${ROLE_INFO}\n"
else
    echo -e "${RED}✗ Super admin role not found${NC}"
    echo -e "${YELLOW}Please run migration 030 first:${NC}"
    echo -e "  docker exec -i \$(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs/migrations/030_add_cms_permissions_up.sql\n"
    exit 1
fi

# Check 5: Current role assignment status
echo -e "${BLUE}[5/5] Checking current role assignments for user...${NC}"
CURRENT_ROLES=$(run_query "SELECT r.name FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.user_id = $USER_ID;" | xargs)
if [ -n "$CURRENT_ROLES" ]; then
    echo -e "${YELLOW}ⓘ User currently has roles:${NC}"
    echo -e "  ${CURRENT_ROLES}"

    if echo "$CURRENT_ROLES" | grep -q "super_admin"; then
        echo -e "${YELLOW}⚠ User already has super_admin role${NC}"
        echo -e "${YELLOW}Migration 033 will be skipped (idempotent)${NC}\n"
    else
        echo -e "${GREEN}✓ User does not have super_admin role yet${NC}\n"
    fi
else
    echo -e "${YELLOW}ⓘ User has no roles assigned yet${NC}\n"
fi

# Final summary
echo -e "${GREEN}========================================${NC}"
echo -e "${GREEN}✓ All prerequisites met!${NC}"
echo -e "${GREEN}========================================${NC}\n"

echo -e "${BLUE}Ready to apply migration 033:${NC}"
echo -e "  ${YELLOW}docker exec -i \$(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs/migrations/033_assign_initial_super_admin_up.sql${NC}\n"

exit 0
