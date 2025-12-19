#!/bin/bash

# URL Shortener Observability Stack Startup Script
# This script starts the complete observability stack with health checks

set -e

echo "======================================================================"
echo "  URL Shortener - Observability Stack Deployment"
echo "======================================================================"
echo ""

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Check if Docker is running
echo -n "Checking Docker... "
if ! docker info > /dev/null 2>&1; then
    echo -e "${RED}FAILED${NC}"
    echo "Error: Docker is not running. Please start Docker and try again."
    exit 1
fi
echo -e "${GREEN}OK${NC}"

# Check if Docker Compose is available
echo -n "Checking Docker Compose... "
if ! docker-compose --version > /dev/null 2>&1; then
    echo -e "${RED}FAILED${NC}"
    echo "Error: Docker Compose is not installed."
    exit 1
fi
echo -e "${GREEN}OK${NC}"

# Check if .env file exists
echo -n "Checking environment file... "
if [ ! -f .env ]; then
    echo -e "${YELLOW}WARNING${NC}"
    echo "Warning: .env file not found. Creating from .env.example..."
    if [ -f .env.example ]; then
        cp .env.example .env
        echo "Please edit .env and set required variables (SLACK_WEBHOOK_URL, PAGERDUTY_SERVICE_KEY, etc.)"
        echo "Then run this script again."
        exit 1
    else
        echo "Error: .env.example not found. Please create .env manually."
        exit 1
    fi
fi
echo -e "${GREEN}OK${NC}"

# Create required directories
echo "Creating directories..."
mkdir -p observability/grafana/dashboards
mkdir -p observability/prometheus/data
mkdir -p observability/loki/data
echo -e "${GREEN}Directories created${NC}"

echo ""
echo "======================================================================"
echo "  Starting Observability Stack"
echo "======================================================================"
echo ""

# Stop any existing containers
echo "Stopping existing containers..."
docker-compose -f docker/docker-compose.observability.yml down 2>/dev/null || true

# Start the stack
echo "Starting services..."
docker-compose -f docker/docker-compose.observability.yml up -d

echo ""
echo "Waiting for services to start..."
sleep 10

# Health check function
check_service() {
    local service=$1
    local url=$2
    local max_attempts=30
    local attempt=1

    echo -n "Checking $service... "

    while [ $attempt -le $max_attempts ]; do
        if curl -s -f "$url" > /dev/null 2>&1; then
            echo -e "${GREEN}OK${NC}"
            return 0
        fi
        sleep 2
        attempt=$((attempt + 1))
    done

    echo -e "${RED}FAILED${NC}"
    return 1
}

echo ""
echo "======================================================================"
echo "  Service Health Checks"
echo "======================================================================"
echo ""

# Check each service
check_service "Application" "http://localhost:8080/health"
check_service "Metrics Endpoint" "http://localhost:8080/metrics"
check_service "Prometheus" "http://localhost:9090/-/healthy"
check_service "Grafana" "http://localhost:3000/api/health"
check_service "Alertmanager" "http://localhost:9093/-/healthy"
check_service "Jaeger" "http://localhost:16686/"
check_service "Loki" "http://localhost:3100/ready"

echo ""
echo "======================================================================"
echo "  Deployment Summary"
echo "======================================================================"
echo ""

# Get container status
echo "Container Status:"
docker-compose -f docker/docker-compose.observability.yml ps

echo ""
echo "======================================================================"
echo "  Access URLs"
echo "======================================================================"
echo ""
echo "  Application:        http://localhost:8080"
echo "  Metrics Endpoint:   http://localhost:8080/metrics"
echo ""
echo "  Grafana:            http://localhost:3000"
echo "  Prometheus:         http://localhost:9090"
echo "  Alertmanager:       http://localhost:9093"
echo "  Jaeger:             http://localhost:16686"
echo "  Loki:               http://localhost:3100"
echo ""
echo "======================================================================"
echo "  Credentials"
echo "======================================================================"
echo ""
echo "  Grafana:"
echo "    Username: admin"
echo "    Password: (check .env file - GRAFANA_PASSWORD)"
echo ""
echo "======================================================================"
echo "  Next Steps"
echo "======================================================================"
echo ""
echo "  1. Open Grafana: http://localhost:3000"
echo "  2. Navigate to Dashboards -> Browse"
echo "  3. Open 'Service Health Overview' dashboard"
echo "  4. Generate some test traffic:"
echo ""
echo "     for i in {1..100}; do"
echo "       curl -X POST http://localhost:8080/api/v1/shorten \\"
echo "         -H 'Content-Type: application/json' \\"
echo "         -d '{\"url\":\"https://example.com/test-\$i\"}'"
echo "     done"
echo ""
echo "  5. View logs: docker-compose -f docker/docker-compose.observability.yml logs -f"
echo "  6. Stop stack: docker-compose -f docker/docker-compose.observability.yml down"
echo ""
echo "======================================================================"
echo "  Documentation"
echo "======================================================================"
echo ""
echo "  - Quick Reference:  observability/QUICK_REFERENCE.md"
echo "  - Full Guide:       observability/README.md"
echo "  - Integration:      OBSERVABILITY_INTEGRATION.md"
echo "  - Runbooks:         observability/runbooks/"
echo ""
echo -e "${GREEN}Observability stack is ready!${NC}"
echo ""
