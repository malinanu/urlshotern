#!/bin/bash

# Diagnostic script for 500 error on assignable-roles endpoint
# This script helps verify deployment status and diagnose the issue

set -e

echo "================================================"
echo "500 Error Diagnostic Script"
echo "================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Step 1: Check if we're on the Oracle server
echo -e "${BLUE}Step 1: Environment Check${NC}"
echo "Current directory: $(pwd)"
echo "Current user: $(whoami)"
echo ""

# Step 2: Check if latest commit is present
echo -e "${BLUE}Step 2: Git Commit Status${NC}"
EXPECTED_COMMIT="e05f81a"
CURRENT_COMMIT=$(git log -1 --oneline | cut -d' ' -f1)

if [[ "$CURRENT_COMMIT" == "$EXPECTED_COMMIT"* ]]; then
    echo -e "${GREEN}✓${NC} Latest commit found: $CURRENT_COMMIT"
else
    echo -e "${YELLOW}⚠${NC} Different commit found: $CURRENT_COMMIT"
    echo "  Expected: $EXPECTED_COMMIT"
    echo ""
    echo "Latest commits:"
    git log -3 --oneline
fi
echo ""

# Step 3: Check Docker containers
echo -e "${BLUE}Step 3: Docker Container Status${NC}"
if command -v docker &> /dev/null; then
    echo "Running containers:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Image}}" | grep -E "backend|urlshorter" || echo "No backend containers found"
    echo ""

    # Check if backend container exists
    BACKEND_CONTAINER=$(docker ps --format "{{.Names}}" | grep -E "backend|urlshorter" | head -1)

    if [ -n "$BACKEND_CONTAINER" ]; then
        echo -e "${GREEN}✓${NC} Found backend container: $BACKEND_CONTAINER"

        # Check when container was started
        echo ""
        echo "Container start time:"
        docker inspect "$BACKEND_CONTAINER" | grep StartedAt

        # Check git commit in container (if possible)
        echo ""
        echo "Checking commit in running container..."
        docker exec "$BACKEND_CONTAINER" git log -1 --oneline 2>/dev/null || echo "Unable to check git commit in container"
    else
        echo -e "${RED}✗${NC} No backend container found"
    fi
else
    echo -e "${YELLOW}⚠${NC} Docker not available on this machine"
fi
echo ""

# Step 4: Check recent backend logs
echo -e "${BLUE}Step 4: Recent Backend Logs${NC}"
if [ -n "$BACKEND_CONTAINER" ]; then
    echo "Last 30 lines of backend logs:"
    echo "---"
    docker logs "$BACKEND_CONTAINER" --tail=30
    echo "---"
    echo ""

    # Check for specific error patterns
    echo "Checking for error patterns..."
    if docker logs "$BACKEND_CONTAINER" --tail=100 | grep -i "panic"; then
        echo -e "${RED}✗${NC} Found panic messages in logs"
    fi

    if docker logs "$BACKEND_CONTAINER" --tail=100 | grep -i "Failed to fetch user roles"; then
        echo -e "${RED}✗${NC} Found 'Failed to fetch user roles' errors"
    fi

    if docker logs "$BACKEND_CONTAINER" --tail=100 | grep -i "Failed to fetch assignable roles"; then
        echo -e "${RED}✗${NC} Found 'Failed to fetch assignable roles' errors"
    fi
else
    echo -e "${YELLOW}⚠${NC} No backend container to check logs"
fi
echo ""

# Step 5: Test the endpoint (if curl is available)
echo -e "${BLUE}Step 5: Endpoint Test${NC}"
if command -v curl &> /dev/null; then
    echo "Testing health endpoint..."
    HEALTH_RESPONSE=$(curl -s -w "\nHTTP_STATUS:%{http_code}" http://localhost:8080/health 2>/dev/null || echo "FAILED")

    if echo "$HEALTH_RESPONSE" | grep -q "HTTP_STATUS:200"; then
        echo -e "${GREEN}✓${NC} Backend health check passed"
    else
        echo -e "${RED}✗${NC} Backend health check failed"
        echo "$HEALTH_RESPONSE"
    fi
else
    echo -e "${YELLOW}⚠${NC} curl not available for endpoint testing"
fi
echo ""

# Step 6: Recommendations
echo -e "${BLUE}Step 6: Recommendations${NC}"
echo ""

if [[ "$CURRENT_COMMIT" != "$EXPECTED_COMMIT"* ]]; then
    echo -e "${YELLOW}Action Required:${NC}"
    echo "1. Pull latest code: git pull origin main"
    echo "2. Rebuild backend: docker-compose -f docker/docker-compose.oracle.yml build backend"
    echo "3. Restart backend: docker-compose -f docker/docker-compose.oracle.yml up -d backend"
    echo "4. Check logs: docker logs <container-name> --tail=50 --follow"
elif [ -z "$BACKEND_CONTAINER" ]; then
    echo -e "${YELLOW}Action Required:${NC}"
    echo "1. Start backend container"
    echo "2. Check docker-compose configuration"
else
    echo -e "${GREEN}Deployment appears up to date${NC}"
    echo ""
    echo "If 500 error persists, check the following:"
    echo "1. View real-time logs: docker logs $BACKEND_CONTAINER --tail=50 --follow"
    echo "2. Test endpoint: curl -v http://localhost:8080/api/v1/admin/rbac/assignable-roles"
    echo "3. Check database connection: psql -h localhost -U <username> -d <database>"
    echo "4. Verify user has roles: SELECT * FROM user_roles WHERE user_id = <your_user_id>;"
fi
echo ""

echo "================================================"
echo "Diagnostic Complete"
echo "================================================"
