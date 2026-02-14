#!/bin/bash

# Production Diagnostic Script for 502 Bad Gateway Errors
# Run this on your production server to diagnose the issue

echo "=========================================="
echo "URLshortener Production Diagnostics"
echo "=========================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Step 1: Check Backend Process
echo -e "${YELLOW}[1/5] Checking Backend Process...${NC}"
if ps aux | grep -E "server|urlshortener" | grep -v grep > /dev/null; then
    echo -e "${GREEN}✓ Backend process is running${NC}"
    ps aux | grep -E "server|urlshortener" | grep -v grep
else
    echo -e "${RED}✗ Backend process NOT found${NC}"
    echo "The backend server is not running. This is likely the cause of 502 errors."
fi
echo ""

# Step 2: Check Backend Logs (last 20 lines)
echo -e "${YELLOW}[2/5] Checking Backend Logs...${NC}"
if [ -f "/var/log/urlshortener/error.log" ]; then
    echo "Last 20 lines of error log:"
    tail -20 /var/log/urlshortener/error.log
elif command -v journalctl &> /dev/null; then
    echo "Checking systemd logs:"
    journalctl -u urlshortener -n 20 --no-pager
elif command -v docker &> /dev/null; then
    echo "Checking Docker logs:"
    docker logs --tail=20 $(docker ps -q -f name=backend) 2>/dev/null || echo "No backend container found"
else
    echo -e "${YELLOW}Could not find logs. Check manually.${NC}"
fi
echo ""

# Step 3: Check Database Migration Status
echo -e "${YELLOW}[3/5] Checking Database Migrations...${NC}"
echo "Checking which migrations have been applied..."

# Try to query schema_migrations table
PGPASSWORD=${DB_PASSWORD:-urlshortener} psql -h ${DB_HOST:-localhost} -U ${DB_USER:-urlshortener} -d ${DB_NAME:-urlshortener_db} -t -c "SELECT version FROM schema_migrations ORDER BY version;" 2>/dev/null | grep -v "^$" || echo "Could not query schema_migrations"

echo ""
echo "Checking if required tables exist..."
PGPASSWORD=${DB_PASSWORD:-urlshortener} psql -h ${DB_HOST:-localhost} -U ${DB_USER:-urlshortener} -d ${DB_NAME:-urlshortener_db} -t -c "
SELECT
    CASE WHEN table_name IS NOT NULL THEN '✓ ' || table_name ELSE '✗ ' || missing_table END as status
FROM (
    VALUES
        ('static_pages'),
        ('content_blocks'),
        ('page_templates'),
        ('activity_logs'),
        ('meta_tags'),
        ('seo_analyses'),
        ('page_seo'),
        ('url_redirects'),
        ('robots_txt')
) AS required(missing_table)
LEFT JOIN information_schema.tables ON table_name = missing_table AND table_schema = 'public';
" 2>/dev/null || echo "Could not query database tables"
echo ""

# Step 4: Check Nginx/Proxy Status
echo -e "${YELLOW}[4/5] Checking Reverse Proxy...${NC}"
if [ -f "/var/log/nginx/error.log" ]; then
    echo "Recent Nginx errors related to 502:"
    tail -50 /var/log/nginx/error.log | grep "502\|upstream" | tail -10
elif command -v journalctl &> /dev/null; then
    echo "Checking Caddy logs:"
    journalctl -u caddy -n 20 --no-pager | grep "502\|error" || echo "No errors found"
else
    echo -e "${YELLOW}No reverse proxy logs found${NC}"
fi
echo ""

# Step 5: Check Recent Deployment
echo -e "${YELLOW}[5/5] Checking Deployment Status...${NC}"
if [ -d "/opt/URLshorter" ] || [ -d "/var/www/URLshorter" ] || [ -d "$HOME/URLshorter" ]; then
    APP_DIR=$(find / -maxdepth 3 -name "URLshorter" -type d 2>/dev/null | head -1)
    if [ -n "$APP_DIR" ]; then
        echo "Found application at: $APP_DIR"
        cd "$APP_DIR" || exit
        echo "Current git commit:"
        git log -1 --oneline 2>/dev/null || echo "Not a git repository"
        echo ""
        echo "Checking for migration 027:"
        if [ -f "configs/migrations/027_seo_tables_up.sql" ]; then
            echo -e "${GREEN}✓ Migration 027 exists${NC}"
        else
            echo -e "${RED}✗ Migration 027 NOT found${NC}"
        fi
    fi
else
    echo -e "${YELLOW}Could not locate application directory${NC}"
fi

echo ""
echo "=========================================="
echo "Diagnostic Summary"
echo "=========================================="
echo ""
echo "Next steps:"
echo "1. If backend is NOT running → restart backend service"
echo "2. If tables are missing → apply migrations 017, 025, 027"
echo "3. If seeing database errors in logs → apply migrations first, then restart backend"
echo ""
