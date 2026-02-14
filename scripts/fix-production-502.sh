#!/bin/bash

# Production Fix Script for 502 Bad Gateway Errors
# This script applies the missing migrations and restarts the backend

set -e  # Exit on error

echo "=========================================="
echo "URLshortener 502 Fix Script"
echo "=========================================="
echo ""

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m'

# Configuration - adjust these if needed
DB_HOST=${DB_HOST:-localhost}
DB_USER=${DB_USER:-urlshortener}
DB_NAME=${DB_NAME:-urlshortener_db}
DB_PASSWORD=${DB_PASSWORD:-urlshortener}

# Function to run SQL file
run_migration() {
    local migration_file=$1
    local migration_version=$2

    echo -e "${YELLOW}Applying migration $migration_version...${NC}"

    if [ ! -f "$migration_file" ]; then
        echo -e "${RED}Error: Migration file not found: $migration_file${NC}"
        return 1
    fi

    # Apply migration
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -f "$migration_file"

    # Record in schema_migrations
    PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
        INSERT INTO schema_migrations (version, executed_at)
        VALUES ('$migration_version', NOW())
        ON CONFLICT (version) DO NOTHING;
    "

    echo -e "${GREEN}✓ Migration $migration_version applied successfully${NC}"
    echo ""
}

# Step 1: Check if we're in the right directory
if [ ! -d "configs/migrations" ]; then
    echo -e "${RED}Error: Not in URLshorter directory${NC}"
    echo "Please cd to the URLshorter directory first"
    exit 1
fi

echo -e "${YELLOW}Step 1: Checking current migration status...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "SELECT version, executed_at FROM schema_migrations ORDER BY version;" 2>/dev/null || {
    echo -e "${RED}Error: Cannot connect to database${NC}"
    exit 1
}
echo ""

# Step 2: Check which migrations need to be applied
echo -e "${YELLOW}Step 2: Checking which migrations to apply...${NC}"

NEEDS_017=false
NEEDS_025=false
NEEDS_027=false

# Check if migrations are already applied
if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "SELECT version FROM schema_migrations WHERE version = '017';" 2>/dev/null | grep -q 017; then
    NEEDS_017=true
    echo "  → Migration 017 (CMS tables) needs to be applied"
fi

if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "SELECT version FROM schema_migrations WHERE version = '025';" 2>/dev/null | grep -q 025; then
    NEEDS_025=true
    echo "  → Migration 025 (Activity logs) needs to be applied"
fi

if ! PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -t -c "SELECT version FROM schema_migrations WHERE version = '027';" 2>/dev/null | grep -q 027; then
    NEEDS_027=true
    echo "  → Migration 027 (SEO tables) needs to be applied"
fi

if [ "$NEEDS_017" = false ] && [ "$NEEDS_025" = false ] && [ "$NEEDS_027" = false ]; then
    echo -e "${GREEN}All migrations already applied!${NC}"
    echo "The 502 error might be caused by something else."
    echo "Check backend logs for more details."
    exit 0
fi
echo ""

# Step 3: Apply migrations
echo -e "${YELLOW}Step 3: Applying migrations...${NC}"

if [ "$NEEDS_017" = true ]; then
    run_migration "configs/migrations/017_cms_tables_up.sql" "017"
fi

if [ "$NEEDS_025" = true ]; then
    run_migration "configs/migrations/025_create_activity_logs_up.sql" "025"
fi

if [ "$NEEDS_027" = true ]; then
    run_migration "configs/migrations/027_seo_tables_up.sql" "027"
fi

# Step 4: Verify tables were created
echo -e "${YELLOW}Step 4: Verifying tables were created...${NC}"
PGPASSWORD=$DB_PASSWORD psql -h $DB_HOST -U $DB_USER -d $DB_NAME -c "
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN (
    'static_pages', 'content_blocks', 'page_templates',
    'meta_tags', 'seo_analyses', 'activity_logs',
    'global_settings', 'robots_txt', 'url_redirects'
)
ORDER BY table_name;
"
echo ""

# Step 5: Restart backend
echo -e "${YELLOW}Step 5: Restarting backend service...${NC}"

# Try different restart methods
if systemctl is-active --quiet urlshortener 2>/dev/null; then
    echo "Restarting systemd service..."
    sudo systemctl restart urlshortener
    sleep 2
    systemctl status urlshortener --no-pager
    echo -e "${GREEN}✓ Backend restarted via systemd${NC}"
elif docker ps | grep -q backend; then
    echo "Restarting Docker container..."
    docker restart $(docker ps -q -f name=backend)
    sleep 2
    docker ps | grep backend
    echo -e "${GREEN}✓ Backend restarted via Docker${NC}"
elif pgrep -f "server" > /dev/null; then
    echo "Restarting process manually..."
    pkill -f "server"
    sleep 1
    cd /opt/URLshorter || cd /var/www/URLshorter || cd ~/URLshorter || exit
    nohup ./server > /var/log/urlshortener/app.log 2>&1 &
    echo -e "${GREEN}✓ Backend restarted manually${NC}"
else
    echo -e "${YELLOW}Could not detect running backend. Please restart manually.${NC}"
fi
echo ""

# Step 6: Verification
echo -e "${YELLOW}Step 6: Verifying fix...${NC}"
sleep 3
echo "Testing health endpoint..."
curl -s http://localhost:8080/health || echo "Could not reach backend"
echo ""

echo ""
echo "=========================================="
echo -e "${GREEN}Fix Complete!${NC}"
echo "=========================================="
echo ""
echo "Please test the admin endpoints:"
echo "  1. https://trunc.site/api/v1/admin/seo/analyses"
echo "  2. https://trunc.site/api/v1/admin/cms/pages"
echo "  3. https://trunc.site/api/v1/admin/content-blocks"
echo ""
echo "If you still see 502 errors, check backend logs:"
echo "  journalctl -u urlshortener -n 50 --no-pager"
echo "  OR"
echo "  docker logs backend-container-name"
echo ""
