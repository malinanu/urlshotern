#!/bin/bash
# URLShortener Oracle Server Deployment Script
#
# This script automates deployment to Oracle Cloud Server with
# centralized PostgreSQL architecture.
#
# Prerequisites:
#   - postgresql-central container running
#   - shared-db-network exists
#   - .env file configured
#
# Usage: ./scripts/deploy-oracle.sh

set -e

PROJECT_DIR="$(cd "$(dirname "${BASH_SOURCE[0]}")/.." && pwd)"

echo "🚀 URLShortener Oracle Server Deployment"
echo "=========================================="
echo ""

# Check prerequisites
echo "🔍 Checking prerequisites..."

# Check PostgreSQL
if ! docker ps | grep -q postgresql-central; then
    echo "❌ ERROR: postgresql-central container not running!"
    echo ""
    echo "Please start PostgreSQL first:"
    echo "  docker run -d --name postgresql-central --network shared-db-network -p 5432:5432 -e POSTGRES_PASSWORD=... postgres:15-alpine"
    exit 1
fi
echo "   ✓ PostgreSQL running"

# Check network
if ! docker network inspect shared-db-network >/dev/null 2>&1; then
    echo "⚠️  Creating shared-db-network..."
    docker network create --driver bridge shared-db-network
    echo "   ✓ Network created"
else
    echo "   ✓ shared-db-network exists"
fi

# Check .env file
if [ ! -f "$PROJECT_DIR/.env" ]; then
    echo "❌ ERROR: .env file missing!"
    echo ""
    echo "Please create .env file with required variables:"
    echo "  DB_PASSWORD=..."
    echo "  BASE_URL=..."
    exit 1
fi
echo "   ✓ .env file exists"

# Load environment
source "$PROJECT_DIR/.env"

echo ""
echo "🧹 Cleaning up old deployment..."

# Stop old containers
docker stop urlshortener-app urlshortener-redis urlshortener-adminer 2>/dev/null || true

# Remove old containers
docker rm -f urlshortener-app urlshortener-redis urlshortener-adminer 2>/dev/null || true

# Remove old images (force clean rebuild)
docker images | grep urlshortener-app | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true

echo "   ✓ Cleanup complete"

echo ""
echo "🔨 Building and starting containers..."

cd "$PROJECT_DIR"

# Try docker-compose first
if docker-compose -f docker-compose.oracle.yml up -d --build --force-recreate 2>/dev/null; then
    echo "   ✓ Deployed via docker-compose"
else
    echo "   ⚠️  docker-compose failed, using manual deployment..."

    # Build app image
    echo "   Building app image..."
    docker build -f Dockerfile.backend.dev -t urlshortener-app:latest . >/dev/null

    # Start Redis
    echo "   Starting Redis..."
    docker run -d \
      --name urlshortener-redis \
      --network urlshortener-network \
      -p 127.0.0.1:6380:6379 \
      -v urlshortener_redis-data:/data \
      --restart unless-stopped \
      redis:7-alpine redis-server --appendonly yes >/dev/null

    # Start Adminer
    echo "   Starting Adminer..."
    docker run -d \
      --name urlshortener-adminer \
      --network shared-db-network \
      -p 8081:8080 \
      -e ADMINER_DEFAULT_SERVER=postgresql-central \
      --restart unless-stopped \
      adminer:latest >/dev/null

    # Start App
    echo "   Starting app..."
    docker run -d \
      --name urlshortener-app \
      --network shared-db-network \
      -p 8090:8080 \
      -e SERVER_PORT=8080 \
      -e DB_HOST=postgresql-central \
      -e DB_PORT=5432 \
      -e DB_USER=urlshortener_user \
      -e DB_PASSWORD="${DB_PASSWORD}" \
      -e DB_NAME=urlshortener_db \
      -e REDIS_HOST=urlshortener-redis \
      -e REDIS_PORT=6379 \
      -e BASE_URL="${BASE_URL}" \
      -e ENVIRONMENT=production \
      -e NODE_ID=1 \
      --restart unless-stopped \
      urlshortener-app:latest >/dev/null

    # Connect app to Redis network
    docker network connect urlshortener-network urlshortener-app 2>/dev/null || true

    echo "   ✓ Manual deployment complete"
fi

echo ""
echo "⏳ Waiting for services to start..."
sleep 10

echo ""
echo "📊 Deployment Status:"
docker ps | grep urlshortener || echo "   ⚠️  No containers running!"

echo ""
echo "🏥 Health Check:"
if curl -sf http://localhost:8090/health >/dev/null 2>&1; then
    echo "   ✓ App is healthy"
    curl -s http://localhost:8090/health | head -1
else
    echo "   ❌ App health check failed"
    echo "   Check logs: docker logs urlshortener-app"
fi

echo ""
echo "✅ Deployment complete!"
echo ""
echo "Access points:"
echo "  • App:     http://localhost:8090/health"
echo "  • Adminer: http://localhost:8081"
echo ""
echo "Next steps:"
echo "  • Check logs: docker logs urlshortener-app"
echo "  • Run health check: ./scripts/health-check.sh"
echo "  • Test API: curl http://localhost:8090/health"
