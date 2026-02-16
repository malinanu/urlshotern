#!/bin/bash
# SMTP Configuration Verification Script for Production
# Usage: ./verify-smtp-config.sh

set -e

echo "=================================================="
echo "SMTP Configuration Verification"
echo "=================================================="
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Find container ID
echo "Step 1: Finding container..."
CONTAINER_ID=$(docker ps --filter "name=app" --format "{{.ID}}" | head -n 1)

if [ -z "$CONTAINER_ID" ]; then
    echo -e "${RED}✗ Container not found${NC}"
    echo "Make sure your app container is running"
    exit 1
fi

echo -e "${GREEN}✓ Container found: $CONTAINER_ID${NC}"
echo ""

# Check SMTP environment variables
echo "Step 2: Checking SMTP environment variables..."
echo ""

# Function to check env var
check_env_var() {
    local var_name=$1
    local required=$2
    local value=$(docker exec $CONTAINER_ID env | grep "^${var_name}=" | cut -d'=' -f2-)

    if [ -z "$value" ]; then
        if [ "$required" = "true" ]; then
            echo -e "${RED}✗ $var_name: NOT SET (Required)${NC}"
            return 1
        else
            echo -e "${YELLOW}⚠ $var_name: NOT SET (Optional)${NC}"
            return 0
        fi
    else
        # Mask password
        if [ "$var_name" = "SMTP_PASSWORD" ]; then
            echo -e "${GREEN}✓ $var_name: **************** (Set)${NC}"
        else
            echo -e "${GREEN}✓ $var_name: $value${NC}"
        fi
        return 0
    fi
}

# Check each variable
ALL_OK=true
check_env_var "ENVIRONMENT" "true" || ALL_OK=false
check_env_var "SMTP_HOST" "true" || ALL_OK=false
check_env_var "SMTP_PORT" "false"
check_env_var "SMTP_USERNAME" "true" || ALL_OK=false
check_env_var "SMTP_PASSWORD" "true" || ALL_OK=false
check_env_var "FROM_EMAIL" "false"
check_env_var "FROM_NAME" "false"

echo ""

# Check for SMTP errors in logs
echo "Step 3: Checking for SMTP errors in logs..."
echo ""

ERROR_LOG=$(docker logs $CONTAINER_ID 2>&1 | grep -i "failed to create email provider" | tail -n 5)

if [ -n "$ERROR_LOG" ]; then
    echo -e "${RED}✗ SMTP provider creation errors found:${NC}"
    echo "$ERROR_LOG"
    echo ""
    ALL_OK=false
else
    echo -e "${GREEN}✓ No SMTP provider creation errors${NC}"
fi

# Check if emails are going to console
echo ""
echo "Step 4: Checking if emails are being logged to console..."
echo ""

CONSOLE_LOG=$(docker logs $CONTAINER_ID 2>&1 | grep "EMAIL (CONSOLE)" | tail -n 1)

if [ -n "$CONSOLE_LOG" ]; then
    echo -e "${RED}✗ Emails are being logged to console (SMTP not working)${NC}"
    echo "Last console email log:"
    echo "$CONSOLE_LOG"
    echo ""
    ALL_OK=false
else
    echo -e "${GREEN}✓ No console email logs found (SMTP should be working)${NC}"
fi

# Test SMTP connectivity
echo ""
echo "Step 5: Testing SMTP connectivity..."
echo ""

SMTP_HOST=$(docker exec $CONTAINER_ID env | grep "^SMTP_HOST=" | cut -d'=' -f2-)
SMTP_PORT=$(docker exec $CONTAINER_ID env | grep "^SMTP_PORT=" | cut -d'=' -f2-)
SMTP_PORT=${SMTP_PORT:-587}

if [ -n "$SMTP_HOST" ]; then
    echo "Testing connection to $SMTP_HOST:$SMTP_PORT..."

    if timeout 5 bash -c "echo > /dev/tcp/$SMTP_HOST/$SMTP_PORT" 2>/dev/null; then
        echo -e "${GREEN}✓ SMTP server is reachable${NC}"
    else
        echo -e "${RED}✗ Cannot connect to SMTP server${NC}"
        echo "Possible causes:"
        echo "  - SMTP server is down"
        echo "  - Port $SMTP_PORT is blocked by firewall"
        echo "  - Network connectivity issues"
        ALL_OK=false
    fi
else
    echo -e "${YELLOW}⚠ SMTP_HOST not set, skipping connectivity test${NC}"
fi

# Summary
echo ""
echo "=================================================="
echo "Summary"
echo "=================================================="
echo ""

if [ "$ALL_OK" = true ]; then
    echo -e "${GREEN}✓ SMTP configuration looks good!${NC}"
    echo ""
    echo "Next steps:"
    echo "1. Test password reset at: https://trunc.site/forgot-password"
    echo "2. Check your email inbox (and spam folder)"
    echo "3. Verify reset link works correctly"
else
    echo -e "${RED}✗ SMTP configuration has issues${NC}"
    echo ""
    echo "Actions needed:"
    echo "1. Add missing SMTP environment variables in Portainer"
    echo "2. Verify SMTP credentials are correct"
    echo "3. Restart the container after making changes"
    echo "4. Run this script again to verify"
    echo ""
    echo "See SMTP-CONFIGURATION-GUIDE.md for detailed instructions"
fi

echo ""
echo "=================================================="
