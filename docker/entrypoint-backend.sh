#!/bin/sh
# Docker entrypoint script for backend container
# Runs database migrations before starting the application

# Removed 'set -e' to allow app to start even if migrations fail
# set -e

echo "🚀 Starting URL Shortener Backend..."

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

echo "✅ Backend initialization complete!"
echo "📡 Starting Go backend with air (hot reload enabled)..."
echo ""

# Execute the main command (air for development)
exec "$@"
