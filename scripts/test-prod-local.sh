#!/usr/bin/env bash
# Test Production Environment Locally
# This script runs the production-like configuration locally for testing

set -euo pipefail

COMPOSE_FILE="docker-compose.prod-local.yml"
BASE_URL="http://localhost:8083"

# Colors
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
CYAN='\033[0;36m'
NC='\033[0m' # No Color

print_header() {
    echo -e "\n${CYAN}================================================${NC}"
    echo -e "${CYAN}  $1${NC}"
    echo -e "${CYAN}================================================${NC}\n"
}

print_success() {
    echo -e "${GREEN}✅ $1${NC}"
}

print_error() {
    echo -e "${RED}❌ $1${NC}"
}

print_info() {
    echo -e "${BLUE}ℹ️  $1${NC}"
}

check_container() {
    local container_name="$1"
    docker ps --filter "name=${container_name}" --format "{{.Status}}" | grep -q "Up"
}

verify_deployment() {
    print_header "Verifying Production Deployment"

    # Check if container is running
    print_info "Checking container status..."
    if check_container "urlshortener-app-prod"; then
        print_success "Container urlshortener-app-prod is running"
    else
        print_error "Container urlshortener-app-prod is not running"
        return 1
    fi

    # Check if static files exist
    print_info "Checking for frontend static files..."
    if docker exec urlshortener-app-prod sh -c "ls -la /app/static/ 2>/dev/null" | grep -q "index.html"; then
        print_success "Frontend static files found in /app/static/"
    else
        print_error "Frontend static files not found in /app/static/"
        docker exec urlshortener-app-prod sh -c "ls -la /app/static/ 2>/dev/null || echo 'Directory not found'"
    fi

    # Test health endpoint
    print_info "Testing API health endpoint..."
    sleep 2
    if curl -sf "${BASE_URL}/api/v1/health" > /dev/null; then
        print_success "API health endpoint responding: ${BASE_URL}/api/v1/health"
        curl -s "${BASE_URL}/api/v1/health" | jq '.' 2>/dev/null || curl -s "${BASE_URL}/api/v1/health"
    else
        print_error "API health endpoint failed"
    fi

    # Test frontend homepage
    print_info "Testing frontend homepage..."
    if curl -sf "${BASE_URL}/" | grep -q "<!DOCTYPE html>"; then
        print_success "Frontend homepage responding: ${BASE_URL}/"
    else
        print_error "Frontend homepage failed or returned unexpected content"
    fi

    # Check logs for frontend build message
    print_info "Checking logs for frontend build confirmation..."
    if docker logs urlshortener-app-prod 2>&1 | grep -i frontend | head -5; then
        print_success "Frontend build logs found"
    else
        print_info "No frontend-specific logs found yet"
    fi

    print_header "Access Points"
    echo -e "${YELLOW}Frontend:  ${BASE_URL}/${NC}"
    echo -e "${YELLOW}API:       ${BASE_URL}/api/v1/health${NC}"
    echo -e "${YELLOW}Adminer:   http://localhost:8084${NC}"
    echo -e "\n${YELLOW}Database Connection:${NC}"
    echo -e "  Host:     localhost:5433"
    echo -e "  Database: urlshortener_db"
    echo -e "  User:     urlshortener_user"
}

case "${1:-up}" in
    up)
        print_header "Starting Production Environment (Local)"
        print_info "Using: ${COMPOSE_FILE}"
        docker-compose -f "${COMPOSE_FILE}" up -d
        print_success "Services started"
        sleep 5
        verify_deployment
        ;;

    down)
        print_header "Stopping Production Environment (Local)"
        docker-compose -f "${COMPOSE_FILE}" down
        print_success "Services stopped"
        ;;

    logs)
        print_header "Following Logs (Ctrl+C to exit)"
        docker-compose -f "${COMPOSE_FILE}" logs -f
        ;;

    build)
        print_header "Building Production Image"
        docker-compose -f "${COMPOSE_FILE}" build
        print_success "Build complete"
        ;;

    rebuild)
        print_header "Rebuilding and Restarting"
        docker-compose -f "${COMPOSE_FILE}" down
        docker-compose -f "${COMPOSE_FILE}" build --no-cache
        docker-compose -f "${COMPOSE_FILE}" up -d
        print_success "Rebuild complete"
        sleep 5
        verify_deployment
        ;;

    status)
        print_header "Service Status"
        docker-compose -f "${COMPOSE_FILE}" ps
        ;;

    verify)
        verify_deployment
        ;;

    *)
        echo "Usage: $0 [up|down|logs|build|rebuild|status|verify]"
        echo ""
        echo "Commands:"
        echo "  up       - Start production environment"
        echo "  down     - Stop production environment"
        echo "  logs     - Follow logs"
        echo "  build    - Build production image"
        echo "  rebuild  - Rebuild from scratch and restart"
        echo "  status   - Show service status"
        echo "  verify   - Verify deployment"
        exit 1
        ;;
esac
