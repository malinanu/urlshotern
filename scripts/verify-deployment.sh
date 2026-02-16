#!/bin/bash
# Deployment Verification Script for Password Reset Pages
# Run this on the Oracle Ubuntu server after deploying

set -e

echo "========================================="
echo "Password Reset Pages Deployment Verification"
echo "========================================="
echo ""

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
COMPOSE_FILE="docker/docker-compose.oracle.yml"
PRODUCTION_URL="https://trunc.site"

# Function to print status
print_status() {
    if [ $1 -eq 0 ]; then
        echo -e "${GREEN}✓${NC} $2"
    else
        echo -e "${RED}✗${NC} $2"
    fi
}

print_warning() {
    echo -e "${YELLOW}⚠${NC} $1"
}

# Step 1: Check Docker Containers
echo "1. Checking Docker Containers..."
if docker-compose -f $COMPOSE_FILE ps | grep -q "Up"; then
    print_status 0 "Containers are running"
    docker-compose -f $COMPOSE_FILE ps
else
    print_status 1 "Containers are not running properly"
    exit 1
fi
echo ""

# Step 2: Get Container ID
echo "2. Getting Container ID..."
CONTAINER_ID=$(docker-compose -f $COMPOSE_FILE ps -q app 2>/dev/null)
if [ -z "$CONTAINER_ID" ]; then
    print_status 1 "Could not find app container"
    exit 1
fi
print_status 0 "Container ID: $CONTAINER_ID"
echo ""

# Step 3: Check Static Files
echo "3. Checking Static Files in Container..."

if docker exec $CONTAINER_ID test -f /app/static/forgot-password.html; then
    print_status 0 "forgot-password.html exists"
    FILE_SIZE=$(docker exec $CONTAINER_ID stat -f%z /app/static/forgot-password.html 2>/dev/null || docker exec $CONTAINER_ID stat -c%s /app/static/forgot-password.html 2>/dev/null)
    echo "   Size: $FILE_SIZE bytes"
else
    print_status 1 "forgot-password.html NOT FOUND"
fi

if docker exec $CONTAINER_ID test -f /app/static/reset-password.html; then
    print_status 0 "reset-password.html exists"
    FILE_SIZE=$(docker exec $CONTAINER_ID stat -f%z /app/static/reset-password.html 2>/dev/null || docker exec $CONTAINER_ID stat -c%s /app/static/reset-password.html 2>/dev/null)
    echo "   Size: $FILE_SIZE bytes"
else
    print_status 1 "reset-password.html NOT FOUND"
fi
echo ""

# Step 4: Check Next.js Static Assets
echo "4. Checking Next.js Static Assets..."
if docker exec $CONTAINER_ID test -d /app/static/_next/static; then
    print_status 0 "Next.js static assets directory exists"
    ASSET_COUNT=$(docker exec $CONTAINER_ID find /app/static/_next/static -type f | wc -l)
    echo "   Files: $ASSET_COUNT"
else
    print_status 1 "Next.js static assets NOT FOUND"
fi
echo ""

# Step 5: Test HTTP Endpoints
echo "5. Testing HTTP Endpoints..."

# Test forgot-password page
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$PRODUCTION_URL/forgot-password" || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    print_status 0 "GET $PRODUCTION_URL/forgot-password → $HTTP_CODE"
else
    print_status 1 "GET $PRODUCTION_URL/forgot-password → $HTTP_CODE (Expected 200)"
fi

# Test reset-password page
HTTP_CODE=$(curl -s -o /dev/null -w "%{http_code}" "$PRODUCTION_URL/reset-password?token=test" || echo "000")
if [ "$HTTP_CODE" = "200" ]; then
    print_status 0 "GET $PRODUCTION_URL/reset-password → $HTTP_CODE"
else
    print_status 1 "GET $PRODUCTION_URL/reset-password → $HTTP_CODE (Expected 200)"
fi
echo ""

# Step 6: Test API Endpoints
echo "6. Testing API Endpoints..."

# Test forgot-password API
API_RESPONSE=$(curl -s -X POST "$PRODUCTION_URL/api/v1/auth/forgot-password" \
    -H "Content-Type: application/json" \
    -d '{"email":"test@example.com"}' \
    -w "\n%{http_code}" || echo "000")

HTTP_CODE=$(echo "$API_RESPONSE" | tail -n 1)
if [ "$HTTP_CODE" = "200" ] || [ "$HTTP_CODE" = "400" ]; then
    print_status 0 "POST $PRODUCTION_URL/api/v1/auth/forgot-password → $HTTP_CODE"
else
    print_status 1 "POST $PRODUCTION_URL/api/v1/auth/forgot-password → $HTTP_CODE"
fi
echo ""

# Step 7: Check Application Logs
echo "7. Checking Recent Application Logs..."
echo "Last 20 log entries:"
docker-compose -f $COMPOSE_FILE logs --tail=20 app
echo ""

# Step 8: Check for Errors
echo "8. Checking for Errors in Logs..."
ERROR_COUNT=$(docker-compose -f $COMPOSE_FILE logs --tail=100 | grep -i "error" | wc -l)
if [ "$ERROR_COUNT" -eq 0 ]; then
    print_status 0 "No errors found in recent logs"
else
    print_warning "Found $ERROR_COUNT error entries in logs (review above)"
fi

NOT_FOUND_COUNT=$(docker-compose -f $COMPOSE_FILE logs --tail=100 | grep -i "404" | wc -l)
if [ "$NOT_FOUND_COUNT" -eq 0 ]; then
    print_status 0 "No 404 errors found in recent logs"
else
    print_warning "Found $NOT_FOUND_COUNT 404 entries in logs"
fi
echo ""

# Step 9: Environment Variables Check
echo "9. Checking Critical Environment Variables..."
ENV_VARS=("SMTP_HOST" "SMTP_PORT" "SMTP_FROM" "BASE_URL")

for var in "${ENV_VARS[@]}"; do
    VALUE=$(docker-compose -f $COMPOSE_FILE exec -T app env | grep "^$var=" || echo "NOT_SET")
    if [ "$VALUE" != "NOT_SET" ]; then
        print_status 0 "$var is set"
    else
        print_status 1 "$var is NOT set"
    fi
done
echo ""

# Summary
echo "========================================="
echo "Verification Complete!"
echo "========================================="
echo ""
echo "Next Steps:"
echo "1. Test forgot-password page in browser: $PRODUCTION_URL/forgot-password"
echo "2. Test reset-password page in browser: $PRODUCTION_URL/reset-password?token=test"
echo "3. Perform end-to-end password reset flow with real email"
echo "4. Monitor logs: docker-compose -f $COMPOSE_FILE logs -f"
echo ""
