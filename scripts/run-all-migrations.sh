#!/bin/bash
# Unified migration runner for Docker environments
# Orchestrates execution of both migration directories in correct order
set -e

echo "🔄 Running all database migrations..."

# Run core migrations first (configs/migrations/)
echo "📦 Running core migrations from configs/migrations/..."
export MIGRATIONS_DIR=/app/configs/migrations
if [ -f "/app/scripts/run-migrations.sh" ]; then
    /app/scripts/run-migrations.sh
else
    echo "⚠️  Core migration script not found at /app/scripts/run-migrations.sh"
fi

# Run admin/feature migrations (migrations/)
echo "📦 Running admin/feature migrations from migrations/..."
export MIGRATIONS_DIR=/app/migrations
if [ -f "/app/scripts/run-migrations-monolithic.sh" ]; then
    /app/scripts/run-migrations-monolithic.sh
else
    echo "⚠️  Monolithic migration script not found at /app/scripts/run-migrations-monolithic.sh"
fi

echo "✅ All migrations complete!"
