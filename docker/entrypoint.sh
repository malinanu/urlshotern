#!/bin/sh
# Docker entrypoint script for production
# Runs database migrations before starting the application

echo "🚀 Starting URL Shortener Application..."

# Run database migrations
echo "🔄 Running database migrations..."
if [ -f "/home/appuser/scripts/run-all-migrations.sh" ]; then
    cd /home/appuser
    chmod +x ./scripts/run-all-migrations.sh
    chmod +x ./scripts/run-migrations.sh
    chmod +x ./scripts/run-migrations-monolithic.sh
    ./scripts/run-all-migrations.sh
elif [ -f "/home/appuser/scripts/run-migrations.sh" ]; then
    # Fallback to old behavior
    cd /home/appuser
    chmod +x ./scripts/run-migrations.sh
    ./scripts/run-migrations.sh
else
    echo "⚠️  Migration script not found, skipping migrations"
fi

echo "✅ Application initialization complete!"
echo ""

# Execute the main command
exec "$@"
