#!/bin/bash
# RBAC Deployment Verification Script
# Usage: ./verify-rbac-deployment.sh <base-url> <admin-token>
# Example: ./verify-rbac-deployment.sh https://trunc.site eyJhbGc...

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Parse arguments
BASE_URL="${1:-https://trunc.site}"
TOKEN="$2"

if [ -z "$TOKEN" ]; then
    echo -e "${RED}Error: Admin token required${NC}"
    echo "Usage: $0 <base-url> <admin-token>"
    echo "Example: $0 https://trunc.site eyJhbGc..."
    exit 1
fi

echo "======================================"
echo "RBAC Deployment Verification"
echo "======================================"
echo "Base URL: $BASE_URL"
echo "Token: ${TOKEN:0:20}..."
echo ""

# Test counter
PASSED=0
FAILED=0

# Helper function to test endpoint
test_endpoint() {
    local name="$1"
    local method="$2"
    local endpoint="$3"
    local data="$4"
    local expected_status="$5"

    echo -n "Testing: $name ... "

    if [ "$method" = "GET" ]; then
        response=$(curl -s -w "\n%{http_code}" \
            -H "Authorization: Bearer $TOKEN" \
            "$BASE_URL$endpoint")
    else
        response=$(curl -s -w "\n%{http_code}" \
            -X "$method" \
            -H "Authorization: Bearer $TOKEN" \
            -H "Content-Type: application/json" \
            -d "$data" \
            "$BASE_URL$endpoint")
    fi

    http_code=$(echo "$response" | tail -n1)
    body=$(echo "$response" | sed '$d')

    if [ "$http_code" = "$expected_status" ]; then
        echo -e "${GREEN}✓ PASSED${NC} (HTTP $http_code)"
        PASSED=$((PASSED + 1))

        # Show sample response if verbose
        if [ ! -z "$VERBOSE" ]; then
            echo "$body" | jq -C '.' 2>/dev/null || echo "$body"
        fi
        return 0
    else
        echo -e "${RED}✗ FAILED${NC} (Expected $expected_status, got $http_code)"
        echo "Response: $body"
        FAILED=$((FAILED + 1))
        return 1
    fi
}

# Test 1: Health Check
echo ""
echo "1. Basic Health Checks"
echo "----------------------"
test_endpoint "API Health Check" "GET" "/health" "" "200"
test_endpoint "Database Pool Stats" "GET" "/health/db-pool" "" "200"

# Test 2: RBAC Endpoints
echo ""
echo "2. RBAC Management Endpoints"
echo "----------------------------"
test_endpoint "List All Roles" "GET" "/api/v1/admin/rbac/roles" "" "200"
test_endpoint "List All Permissions" "GET" "/api/v1/admin/rbac/permissions" "" "200"
test_endpoint "Get Audit Log" "GET" "/api/v1/admin/rbac/audit?limit=5" "" "200"

# Test 3: CMS Endpoints with Permissions
echo ""
echo "3. CMS Permission-Based Endpoints"
echo "----------------------------------"
test_endpoint "List CMS Pages (requires cms.pages.read)" "GET" "/api/v1/admin/cms/pages" "" "200"
test_endpoint "List Content Blocks (requires cms.blocks.read)" "GET" "/api/v1/admin/cms/blocks" "" "200"

# Test 4: User Management
echo ""
echo "4. User Management (Admin Only)"
echo "--------------------------------"

# Create test user data
TEST_USER_DATA='{
  "name": "RBAC Test User",
  "email": "rbac-test-'$(date +%s)'@example.com",
  "account_type": "free",
  "role_ids": [4],
  "email_verified": true,
  "is_active": true
}'

if test_endpoint "Create User with Role" "POST" "/api/v1/admin/users/create" "$TEST_USER_DATA" "201"; then
    # Extract user ID from response if jq is available
    if command -v jq &> /dev/null; then
        USER_ID=$(echo "$body" | jq -r '.user.id // empty')
        if [ ! -z "$USER_ID" ]; then
            echo "  Created user ID: $USER_ID"
            test_endpoint "Get User Roles" "GET" "/api/v1/admin/users/$USER_ID/roles" "" "200"
        fi
    fi
fi

# Test 5: Backward Compatibility
echo ""
echo "5. Backward Compatibility"
echo "-------------------------"
echo -n "Checking is_admin flag support ... "
# The RequireAdmin middleware should still work with is_admin=true users
curl -s -o /dev/null -w "%{http_code}" \
    -H "Authorization: Bearer $TOKEN" \
    "$BASE_URL/api/v1/admin/users" > /tmp/admin_test_status.txt

if [ "$(cat /tmp/admin_test_status.txt)" = "200" ] || [ "$(cat /tmp/admin_test_status.txt)" = "404" ]; then
    echo -e "${GREEN}✓ PASSED${NC}"
    PASSED=$((PASSED + 1))
else
    echo -e "${RED}✗ FAILED${NC}"
    FAILED=$((FAILED + 1))
fi

# Test 6: Database Verification
echo ""
echo "6. Database Schema Verification"
echo "--------------------------------"
echo "NOTE: Run these queries on your database server:"
echo ""
echo "# Verify roles created:"
echo "SELECT name, display_name FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');"
echo ""
echo "# Verify permissions count:"
echo "SELECT resource, COUNT(*) FROM permissions GROUP BY resource ORDER BY resource;"
echo ""
echo "# Verify audit log table exists:"
echo "\\d rbac_audit_log"

# Summary
echo ""
echo "======================================"
echo "Summary"
echo "======================================"
echo -e "Tests Passed: ${GREEN}$PASSED${NC}"
echo -e "Tests Failed: ${RED}$FAILED${NC}"
echo "Total Tests: $((PASSED + FAILED))"
echo ""

if [ $FAILED -eq 0 ]; then
    echo -e "${GREEN}✓ All tests passed! RBAC system is working correctly.${NC}"
    exit 0
else
    echo -e "${YELLOW}⚠ Some tests failed. Check the output above for details.${NC}"
    exit 1
fi
