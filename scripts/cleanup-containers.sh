#!/bin/bash
# URLShortener Container Cleanup Script
#
# This script safely removes all URLShortener containers and images
# to prepare for a fresh deployment. Safe to run - PostgreSQL is external.
#
# Usage: ./scripts/cleanup-containers.sh

set -e

echo "🧹 URLShortener Container Cleanup"
echo "=================================="
echo ""

# Stop containers
echo "📦 Stopping containers..."
docker stop urlshortener-app urlshortener-redis urlshortener-adminer 2>/dev/null || true
echo "   ✓ Containers stopped"

# Remove containers
echo "🗑️  Removing containers..."
docker rm -f urlshortener-app urlshortener-redis urlshortener-adminer 2>/dev/null || true
echo "   ✓ Containers removed"

# Remove images
echo "🖼️  Removing images..."
docker images | grep urlshortener-app | awk '{print $3}' | xargs -r docker rmi -f 2>/dev/null || true
echo "   ✓ Images removed"

# Prune system
echo "🧽 Pruning Docker system..."
docker system prune -f >/dev/null 2>&1
echo "   ✓ System pruned"

echo ""
echo "✅ Cleanup complete!"
echo ""

# Verify clean state
echo "📊 Current state:"
docker ps -a | grep urlshortener && echo "   ⚠️  Some containers still exist" || echo "   ✓ No URLShortener containers"
docker images | grep urlshortener && echo "   ⚠️  Some images still exist" || echo "   ✓ No URLShortener images"

echo ""
echo "Ready for fresh deployment!"
