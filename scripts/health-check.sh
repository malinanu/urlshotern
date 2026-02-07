#!/bin/bash
# URLShortener Health Check Script
#
# Checks the health of all URLShortener components
#
# Usage: ./scripts/health-check.sh

echo "🏥 URLShortener Health Check"
echo "============================"
echo ""

# Container status
echo "📦 Container Status:"
echo "-------------------"
if docker ps | grep -q urlshortener; then
    docker ps | grep urlshortener | awk '{print "  ✓ " $NF " - " $7}'
else
    echo "  ❌ No URLShortener containers running"
fi
echo ""

# App health
echo "🔍 App Health:"
echo "-------------"
if curl -sf http://localhost:8090/health >/dev/null 2>&1; then
    echo "  ✅ App is healthy"
    curl -s http://localhost:8090/health
else
    echo "  ❌ App is unhealthy or not responding"
fi
echo ""

# Redis check
echo "📝 Redis Status:"
echo "---------------"
if docker exec urlshortener-redis redis-cli ping >/dev/null 2>&1; then
    echo "  ✅ Redis responding: $(docker exec urlshortener-redis redis-cli ping)"
else
    echo "  ❌ Redis not responding"
fi
echo ""

# Database connectivity
echo "🗄️  Database Connection:"
echo "----------------------"
if curl -sf http://localhost:8090/health/db-pool >/dev/null 2>&1; then
    echo "  ✅ Database connected"
    curl -s http://localhost:8090/health/db-pool | head -5
else
    echo "  ❌ Database connection issue"
fi
echo ""

# Goroutines monitoring
echo "⚙️  Goroutines:"
echo "-------------"
if curl -sf http://localhost:8090/health/goroutines >/dev/null 2>&1; then
    echo "  ✅ Goroutines monitoring active"
    curl -s http://localhost:8090/health/goroutines | head -5
else
    echo "  ⚠️  Goroutines endpoint not available"
fi
echo ""

# Resource usage
echo "💾 Resource Usage:"
echo "-----------------"
docker stats --no-stream --format "table {{.Name}}\t{{.CPUPerc}}\t{{.MemUsage}}" | grep urlshortener || echo "  No stats available"
echo ""

# Recent errors in logs
echo "📋 Recent Logs (last 10 lines):"
echo "-------------------------------"
if docker ps | grep -q urlshortener-app; then
    docker logs urlshortener-app --tail 10 2>&1 | sed 's/^/  /'
else
    echo "  App container not running"
fi
echo ""

# Summary
echo "📊 Summary:"
echo "----------"
APP_HEALTHY=$(curl -sf http://localhost:8090/health >/dev/null 2>&1 && echo "✅" || echo "❌")
REDIS_HEALTHY=$(docker exec urlshortener-redis redis-cli ping >/dev/null 2>&1 && echo "✅" || echo "❌")
DB_HEALTHY=$(curl -sf http://localhost:8090/health/db-pool >/dev/null 2>&1 && echo "✅" || echo "❌")

echo "  App:      $APP_HEALTHY"
echo "  Redis:    $REDIS_HEALTHY"
echo "  Database: $DB_HEALTHY"
echo ""

if [ "$APP_HEALTHY" = "✅" ] && [ "$REDIS_HEALTHY" = "✅" ] && [ "$DB_HEALTHY" = "✅" ]; then
    echo "✅ All systems operational!"
    exit 0
else
    echo "⚠️  Some systems need attention"
    exit 1
fi
