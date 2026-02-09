#!/bin/bash

# Authentication Endpoints Testing Script
# Usage: ./test-auth-endpoints.sh [SERVER_URL]
# Example: ./test-auth-endpoints.sh http://localhost:8082

SERVER_URL="${1:-http://localhost:8082}"
API_BASE="${SERVER_URL}/api/v1"

# Colors for output
GREEN='\033[0;32m'
RED='\033[0;31m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

echo "=========================================="
echo "Authentication Endpoints Test"
echo "Server: ${SERVER_URL}"
echo "=========================================="
echo

# Function to test endpoint
test_endpoint() {
    local method=$1
    local endpoint=$2
    local data=$3
    local description=$4

    echo -e "${YELLOW}Testing:${NC} ${method} ${endpoint}"
    echo "Description: ${description}"

    if [ -z "$data" ]; then
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X ${method} "${API_BASE}${endpoint}")
    else
        response=$(curl -s -w "\nHTTP_CODE:%{http_code}" -X ${method} \
            -H "Content-Type: application/json" \
            -d "${data}" \
            "${API_BASE}${endpoint}")
    fi

    http_code=$(echo "$response" | grep "HTTP_CODE:" | cut -d: -f2)
    body=$(echo "$response" | grep -v "HTTP_CODE:")

    if [ "$http_code" == "404" ]; then
        echo -e "${RED}❌ FAILED - 404 Not Found${NC}"
        echo "Response: $body"
    elif [ "$http_code" == "200" ] || [ "$http_code" == "201" ]; then
        echo -e "${GREEN}✅ SUCCESS - ${http_code}${NC}"
        echo "Response: $(echo $body | jq '.' 2>/dev/null || echo $body)"
    else
        echo -e "${YELLOW}⚠ ${http_code}${NC}"
        echo "Response: $(echo $body | jq '.' 2>/dev/null || echo $body)"
    fi
    echo
}

# 1. Health Check
echo "=========================================="
echo "1. Health Check"
echo "=========================================="
test_endpoint "GET" "/health" "" "Basic health check"

# 2. Registration (Bad Request - Missing Fields)
echo "=========================================="
echo "2. Registration - Validation Test"
echo "=========================================="
test_endpoint "POST" "/auth/register" \
    '{"email":"test@example.com"}' \
    "Should return 400 validation error (NOT 404)"

# 3. Registration (Valid)
echo "=========================================="
echo "3. Registration - Valid User"
echo "=========================================="
RANDOM_EMAIL="test$(date +%s)@example.com"
test_endpoint "POST" "/auth/register" \
    "{\"name\":\"Test User\",\"email\":\"${RANDOM_EMAIL}\",\"password\":\"SecurePass123!\",\"phone\":\"+1234567890\"}" \
    "Should create user and return tokens"

# 4. Login (Invalid - User doesn't exist)
echo "=========================================="
echo "4. Login - Invalid Credentials"
echo "=========================================="
test_endpoint "POST" "/auth/login" \
    '{"email":"nonexistent@example.com","password":"WrongPass123!"}' \
    "Should return 401 unauthorized"

# 5. Refresh Token (No token provided)
echo "=========================================="
echo "5. Refresh Token - No Token"
echo "=========================================="
test_endpoint "POST" "/auth/refresh" \
    '{}' \
    "Should return 400 or 401 (NOT 404)"

# 6. OTP Send (Invalid)
echo "=========================================="
echo "6. OTP - Send (No Phone)"
echo "=========================================="
test_endpoint "POST" "/auth/otp/send" \
    '{}' \
    "Should return 400 validation error"

# 7. Email Verification (Invalid Token)
echo "=========================================="
echo "7. Email Verification - Invalid Token"
echo "=========================================="
test_endpoint "GET" "/auth/verify-email?token=invalid" \
    "" \
    "Should return error (NOT 404)"

# 8. Protected Route - Profile (No Auth)
echo "=========================================="
echo "8. Protected Route - Profile (No Token)"
echo "=========================================="
test_endpoint "GET" "/auth/profile" \
    "" \
    "Should return 401 unauthorized (NOT 404)"

echo "=========================================="
echo "Test Summary"
echo "=========================================="
echo -e "${GREEN}Key Check:${NC} If all endpoints return codes OTHER than 404,"
echo "then the authentication routes are properly registered!"
echo
echo -e "${RED}If ANY endpoint returns 404:${NC} The container is running old code."
echo "Follow PORTAINER_DEPLOYMENT_GUIDE.md to rebuild."
echo
echo "=========================================="
echo "Next Steps"
echo "=========================================="
echo "1. ✅ Verify no 404 errors above"
echo "2. Register a real user from frontend"
echo "3. Test login flow"
echo "4. Test protected endpoints with valid token"
echo "=========================================="
