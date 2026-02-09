#!/bin/bash
# Database Diagnostic and Fix Script
# This script checks database state and fixes missing tables/columns

set -e

echo "🔍 Database Diagnostic and Fix Script"
echo "========================================"
echo ""

# Configuration
DB_CONTAINER="${DB_CONTAINER:-postgresql-central}"
DB_USER="${DB_USER:-urlshortener_user}"
DB_NAME="${DB_NAME:-urlshortener_db}"
APP_CONTAINER="${APP_CONTAINER:-urlshortener-app}"

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Function to run SQL query
run_sql() {
    docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -t -c "$1" 2>/dev/null | xargs
}

# Function to check if table exists
table_exists() {
    local table_name=$1
    local result=$(run_sql "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$table_name');")
    [[ "$result" == "t" ]]
}

# Function to check if column exists
column_exists() {
    local table_name=$1
    local column_name=$2
    local result=$(run_sql "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '$table_name' AND column_name = '$column_name');")
    [[ "$result" == "t" ]]
}

echo "📊 Step 1: Checking database connectivity..."
if docker exec -i "$DB_CONTAINER" psql -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; then
    echo -e "${GREEN}✅ Database is accessible${NC}"
else
    echo -e "${RED}❌ Cannot connect to database${NC}"
    echo "   Please check database container is running: docker ps | grep postgresql"
    exit 1
fi
echo ""

echo "🔍 Step 2: Checking table existence..."
TABLES_OK=true

# Check api_keys table
if table_exists "api_keys"; then
    echo -e "${GREEN}✅ api_keys table exists${NC}"
else
    echo -e "${RED}❌ api_keys table is MISSING${NC}"
    TABLES_OK=false
fi

# Check api_key_usage table
if table_exists "api_key_usage"; then
    echo -e "${GREEN}✅ api_key_usage table exists${NC}"
else
    echo -e "${RED}❌ api_key_usage table is MISSING${NC}"
    TABLES_OK=false
fi

# Check url_mappings table
if table_exists "url_mappings"; then
    echo -e "${GREEN}✅ url_mappings table exists${NC}"

    # Check user_id column
    if column_exists "url_mappings" "user_id"; then
        echo -e "${GREEN}✅ url_mappings.user_id column exists${NC}"
    else
        echo -e "${RED}❌ url_mappings.user_id column is MISSING${NC}"
        TABLES_OK=false
    fi
else
    echo -e "${RED}❌ url_mappings table is MISSING${NC}"
    TABLES_OK=false
fi

# Check users table
if table_exists "users"; then
    echo -e "${GREEN}✅ users table exists${NC}"
else
    echo -e "${RED}❌ users table is MISSING${NC}"
    TABLES_OK=false
fi

echo ""

if [ "$TABLES_OK" = true ]; then
    echo -e "${GREEN}✅ All required tables and columns exist!${NC}"
    echo ""
    echo "📊 Database schema is correct. If you're still experiencing issues:"
    echo "   1. Check application logs: docker logs $APP_CONTAINER --tail 100"
    echo "   2. Verify environment variables are set correctly"
    echo "   3. Check if JWT_SECRET is configured"
    exit 0
fi

echo -e "${YELLOW}⚠️  Database schema issues detected!${NC}"
echo ""
echo "🔧 Step 3: Attempting to fix database schema..."
echo ""

# Option 1: Try running migrations
echo "📝 Option 1: Running migrations..."
if docker exec "$APP_CONTAINER" /app/scripts/run-migrations.sh 2>&1 | grep -q "Migration complete"; then
    echo -e "${GREEN}✅ Migrations completed successfully${NC}"
    echo ""
    echo "🎉 Database should now be fixed! Please verify:"
    echo "   docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c '\dt'"
    exit 0
else
    echo -e "${YELLOW}⚠️  Migration script had issues or is not available${NC}"
fi

echo ""
echo "📝 Option 2: Restarting app container to trigger migrations..."
docker restart "$APP_CONTAINER"
echo "⏳ Waiting for container to restart (10 seconds)..."
sleep 10

# Check logs
echo "📋 Checking startup logs..."
docker logs "$APP_CONTAINER" --tail 50 | grep -E "(Migration|🔧|✅|❌)" || true

echo ""
echo "🔍 Re-checking database schema..."
FIXED=true

if ! table_exists "api_keys"; then
    FIXED=false
fi
if ! table_exists "api_key_usage"; then
    FIXED=false
fi
if ! column_exists "url_mappings" "user_id"; then
    FIXED=false
fi

if [ "$FIXED" = true ]; then
    echo -e "${GREEN}✅ Database schema is now correct!${NC}"
    exit 0
fi

echo ""
echo -e "${YELLOW}⚠️  Automatic fixes didn't resolve all issues${NC}"
echo ""
echo "📝 Manual fix required. Please run these commands:"
echo ""
echo "# View current tables:"
echo "docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c '\dt'"
echo ""
echo "# Check migration status:"
echo "docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME -c 'SELECT * FROM schema_migrations ORDER BY version;'"
echo ""
echo "# Manual migration (if needed):"
echo "docker exec $APP_CONTAINER /app/scripts/run-migrations.sh"
echo ""
echo "# Or apply specific migration:"
echo "docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME < configs/migrations/001_add_api_keys_up.sql"
echo "docker exec -i $DB_CONTAINER psql -U $DB_USER -d $DB_NAME < configs/migrations/003_fix_api_keys_id_type_up.sql"
echo ""
