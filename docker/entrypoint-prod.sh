#!/bin/sh
# Docker entrypoint script for production container
# Runs database migrations before starting the application

echo "🚀 Starting URL Shortener Production Backend..."

# Install PostgreSQL client if not present (Alpine)
if ! command -v psql > /dev/null 2>&1; then
    echo "📦 Installing PostgreSQL client..."
    apk add --no-cache postgresql-client
fi

# Run database migrations
echo "🔄 Running database migrations..."
if [ -f "./scripts/run-migrations.sh" ]; then
    chmod +x ./scripts/run-migrations.sh
    ./scripts/run-migrations.sh
else
    echo "⚠️  Migration script not found, skipping migrations"
fi

echo "✅ Initialization complete!"
echo "📡 Starting Go backend..."
echo ""

# Execute the main command (starts the Go binary)
exec "$@"
