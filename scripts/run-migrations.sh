#!/bin/sh
# Database Migration Runner
# Automatically applies pending migrations in order

set -e

MIGRATIONS_DIR="${MIGRATIONS_DIR:-/app/configs/migrations}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_USER="${DB_USER:-urlshortener}"
DB_NAME="${DB_NAME:-urlshortener_db}"

echo "🔄 Starting database migrations..."
echo "📁 Migrations directory: $MIGRATIONS_DIR"
echo "🗄️  Database: $DB_NAME on $DB_HOST:$DB_PORT"

# Wait for database to be ready
echo "⏳ Waiting for database to be ready..."
until PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c '\q' 2>/dev/null; do
  echo "   Database is unavailable - sleeping"
  sleep 2
done

echo "✅ Database is ready!"

# Create schema_migrations table if it doesn't exist
echo "📋 Ensuring schema_migrations table exists..."
PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" <<-EOSQL
	CREATE TABLE IF NOT EXISTS schema_migrations (
	    version VARCHAR(14) PRIMARY KEY,
	    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
	);
EOSQL

# Get list of applied migrations
APPLIED_MIGRATIONS=$(PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c "SELECT version FROM schema_migrations ORDER BY version;")

# Find all *_up.sql migration files
MIGRATION_FILES=$(find "$MIGRATIONS_DIR" -name '*_up.sql' | sort)

MIGRATIONS_APPLIED=0
MIGRATIONS_SKIPPED=0

for migration_file in $MIGRATION_FILES; do
    # Extract version number from filename (e.g., 002_polar_integration_up.sql -> 002)
    filename=$(basename "$migration_file")
    version=$(echo "$filename" | grep -oE '^[0-9]+')

    # Skip if version is empty or not numeric
    if [ -z "$version" ]; then
        continue
    fi

    # Check if migration has already been applied
    if echo "$APPLIED_MIGRATIONS" | grep -q "^[[:space:]]*$version[[:space:]]*$"; then
        echo "⏭️  Migration $version already applied, skipping..."
        MIGRATIONS_SKIPPED=$((MIGRATIONS_SKIPPED + 1))
        continue
    fi

    echo "🔧 Applying migration $version: $filename"

    # Apply the migration
    if PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -f "$migration_file"; then
        # Record the migration
        PGPASSWORD=$DB_PASSWORD psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "INSERT INTO schema_migrations (version) VALUES ('$version') ON CONFLICT DO NOTHING;"
        echo "   ✅ Migration $version applied successfully"
        MIGRATIONS_APPLIED=$((MIGRATIONS_APPLIED + 1))
    else
        echo "   ❌ Migration $version failed!"
        exit 1
    fi
done

echo ""
echo "🎉 Migration complete!"
echo "   Applied: $MIGRATIONS_APPLIED"
echo "   Skipped: $MIGRATIONS_SKIPPED"
echo ""
