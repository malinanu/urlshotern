#!/bin/bash
# Migration runner for monolithic SQL files (migrations/ directory)
# Handles SQL files without up/down split
set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/migrations}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-urlshortener}"
DB_NAME="${DB_NAME:-urlshortener_db}"

echo "🔄 Running monolithic migrations from $MIGRATIONS_DIR..."

# Wait for database to be ready
echo "⏳ Waiting for database at $DB_HOST:$DB_PORT..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "   Database is unavailable - sleeping"
  sleep 2
done
echo "✅ Database is ready!"

# Get list of applied migrations with 'M' prefix
APPLIED_MIGRATIONS=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version FROM schema_migrations WHERE version LIKE 'M%' ORDER BY version;" 2>/dev/null || echo "")

# Find all numbered migration files (exclude *_down.sql)
if [ ! -d "$MIGRATIONS_DIR" ]; then
    echo "⚠️  Migration directory $MIGRATIONS_DIR does not exist, skipping monolithic migrations"
    exit 0
fi

MIGRATION_FILES=$(find "$MIGRATIONS_DIR" -maxdepth 1 -name '[0-9]*_*.sql' ! -name '*_down.sql' -type f 2>/dev/null | sort)

if [ -z "$MIGRATION_FILES" ]; then
    echo "ℹ️  No monolithic migration files found in $MIGRATIONS_DIR"
    exit 0
fi

MIGRATIONS_APPLIED=0
MIGRATIONS_SKIPPED=0

for migration_file in $MIGRATION_FILES; do
    filename=$(basename "$migration_file")
    # Extract version number and prefix with 'M' for monolithic
    version_num=$(echo "$filename" | grep -oE '^[0-9]+')

    if [ -z "$version_num" ]; then
        echo "⚠️  Could not extract version from $filename, skipping..."
        continue
    fi

    version="M$(printf '%03d' $version_num)"

    # Check if already applied
    if echo "$APPLIED_MIGRATIONS" | grep -q "^[[:space:]]*$version[[:space:]]*$"; then
        echo "⏭️  Migration $version already applied, skipping..."
        ((MIGRATIONS_SKIPPED++))
        continue
    fi

    echo "🔧 Applying migration $version: $filename"

    # Apply the migration
    if PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file" 2>&1; then
        # Record migration in schema_migrations table
        PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "INSERT INTO schema_migrations (version) VALUES ('$version') ON CONFLICT DO NOTHING;" >/dev/null
        echo "   ✅ Migration $version applied successfully"
        ((MIGRATIONS_APPLIED++))
    else
        echo "   ❌ Migration $version failed!"
        exit 1
    fi
done

echo "🎉 Monolithic migrations complete! Applied: $MIGRATIONS_APPLIED, Skipped: $MIGRATIONS_SKIPPED"
