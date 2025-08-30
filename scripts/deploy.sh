#!/bin/bash

set -e

echo "🚀 Deploying URL Shortener to production..."

# Build for multiple architectures
echo "Building Docker images..."
docker buildx build --platform linux/amd64,linux/arm64 -t url-shortener:latest .

# Create production environment file if it doesn't exist
if [ ! -f .env.prod ]; then
    echo "Creating production environment file..."
    cp .env.example .env.prod
    echo "⚠️  Please update .env.prod with production values before deploying!"
    exit 1
fi

# Check if production environment variables are set
if [ -z "$PRODUCTION_SERVER" ]; then
    echo "❌ PRODUCTION_SERVER environment variable not set"
    echo "   Set it with: export PRODUCTION_SERVER=user@your-server-ip"
    exit 1
fi

# Copy files to production server
echo "Copying files to production server..."
rsync -avz --exclude 'node_modules' --exclude '.git' --exclude 'logs' \
    ./ $PRODUCTION_SERVER:~/url-shortener/

# Deploy on production server
echo "Deploying on production server..."
ssh $PRODUCTION_SERVER << 'EOF'
cd ~/url-shortener

# Load production environment
export $(cat .env.prod | xargs)

# Stop existing services
docker-compose -f docker/docker-compose.yml down

# Pull latest images and start services
docker-compose -f docker/docker-compose.yml up -d --build

# Wait for services to be healthy
echo "Waiting for services to start..."
sleep 30

# Check service health
if docker-compose -f docker/docker-compose.yml ps app | grep -q "Up"; then
    echo "✅ Application is running"
else
    echo "❌ Application failed to start"
    docker-compose -f docker/docker-compose.yml logs app
    exit 1
fi

# Run database migrations if needed
# docker-compose -f docker/docker-compose.yml exec app ./url-shortener --migrate

EOF

echo "✅ Deployment complete!"
echo "🔗 Your URL shortener should be available at your configured domain"

# Test the deployment
if [ ! -z "$BASE_URL" ]; then
    echo "Testing deployment..."
    curl -f "$BASE_URL/health" && echo "✅ Health check passed" || echo "❌ Health check failed"
fi