#!/bin/sh
# Docker entrypoint script
# Runs database migrations before starting the application

# Removed 'set -e' to allow app to start even if migrations fail
# set -e

echo "🚀 Starting URL Shortener Application..."

# Install PostgreSQL client if not present
if ! command -v psql &> /dev/null; then
    echo "📦 Installing PostgreSQL client..."
    apk add --no-cache postgresql-client
fi

# Run database migrations
echo "🔄 Running database migrations..."
if [ -f "/app/scripts/run-migrations.sh" ]; then
    chmod +x /app/scripts/run-migrations.sh
    /app/scripts/run-migrations.sh
else
    echo "⚠️  Migration script not found, skipping migrations"
fi

echo "✅ Application initialization complete!"
echo ""

# Check if frontend is built, if not, build it now
if [ ! -f "/app/static/index.html" ]; then
    echo "🎨 Frontend not found, building now..."
    if [ -d "/app/frontend" ]; then
        cd /app/frontend && npm run build
        echo "📦 Copying frontend build to /app/static..."
        mkdir -p /app/static
        cp -r /app/frontend/out/* /app/static/
        echo "✅ Frontend build complete!"
    else
        echo "⚠️  Frontend directory not found, skipping frontend build"
    fi
    cd /app
fi

# Execute the main command (air for development, or the Go binary for production)
exec "$@"
