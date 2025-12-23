# Database Migrations Guide

This document explains how database migrations work in this project and how to use them.

## Overview

The URL Shortener uses a migration-based approach to manage database schema changes. Migrations are SQL files that are applied in order to keep the database schema up to date.

## Automatic Migration System

**✅ Migrations run automatically when the Docker container starts!**

When you start the application with Docker, the following happens:

1. Container starts
2. Waits for database to be ready
3. Creates `schema_migrations` table if it doesn't exist
4. Checks which migrations have been applied
5. Applies any pending migrations in order
6. Starts the application

This means **on any new PC, all migrations will run automatically** when you start the containers!

## Migration Files Location

```
configs/migrations/
├── 000_schema_migrations.sql         # Creates migration tracking table
├── 002_polar_integration_up.sql      # Polar payment tables
├── 002_polar_integration_down.sql    # Rollback for polar tables
├── 006_add_name_column_to_users_up.sql
├── 007_fix_users_column_names_up.sql
├── 008_add_missing_users_columns_up.sql
├── 009_add_verification_tables_up.sql
├── 010_fix_user_sessions_up.sql
├── 011_add_custom_alias_to_url_mappings_up.sql
└── ... (more migrations)
```

## Migration Naming Convention

Migrations follow this naming pattern:
```
<version>_<description>_<direction>.sql
```

Examples:
- `002_polar_integration_up.sql` - Migration version 002, going "up" (applying)
- `002_polar_integration_down.sql` - Migration version 002, going "down" (rolling back)

## Creating a New Migration

### 1. Choose the Next Version Number

Check the highest existing version:
```bash
ls configs/migrations/*_up.sql | sort -V | tail -1
```

Use the next sequential number (e.g., if last is `011`, use `012`).

### 2. Create the Migration Files

Create both "up" and "down" versions:

**Up migration** (`012_your_feature_up.sql`):
```sql
-- Migration: Your Feature Description
-- Description: Detailed explanation of what this migration does
-- Author: Your Name
-- Date: 2025-12-23

-- Your SQL changes here
ALTER TABLE users ADD COLUMN new_field VARCHAR(100);

-- Add indexes if needed
CREATE INDEX idx_users_new_field ON users(new_field);

-- Add comments
COMMENT ON COLUMN users.new_field IS 'Description of the field';
```

**Down migration** (`012_your_feature_down.sql`):
```sql
-- Migration Rollback: Your Feature Description
-- Author: Your Name
-- Date: 2025-12-23

-- Reverse the changes
DROP INDEX IF EXISTS idx_users_new_field;
ALTER TABLE users DROP COLUMN IF EXISTS new_field;
```

### 3. Test the Migration

#### Option A: Manual Testing (Current Environment)
```bash
# Apply migration
cat configs/migrations/012_your_feature_up.sql | docker exec -i docker-postgres-1 psql -U urlshortener -d urlshortener_db

# Record it
docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "INSERT INTO schema_migrations (version) VALUES ('012');"

# Test rollback
cat configs/migrations/012_your_feature_down.sql | docker exec -i docker-postgres-1 psql -U urlshortener -d urlshortener_db

# Remove migration record
docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "DELETE FROM schema_migrations WHERE version = '012';"
```

#### Option B: Rebuild Containers (Clean Test)
```bash
# Stop and remove containers
docker-compose -f docker/docker-compose.yml down -v

# Rebuild and start (migrations will run automatically)
docker-compose -f docker/docker-compose.yml build
docker-compose -f docker/docker-compose.yml up -d

# Check logs to see migrations running
docker-compose -f docker/docker-compose.yml logs app | grep migration
```

## Running Migrations Manually

### Using the Migration Script

**On Linux/macOS:**
```bash
chmod +x scripts/run-migrations.sh
./scripts/run-migrations.sh
```

**On Windows (PowerShell):**
```powershell
.\scripts\run-migrations.ps1
```

**With custom parameters:**
```powershell
.\scripts\run-migrations.ps1 `
  -DbHost localhost `
  -DbPort 5432 `
  -DbUser urlshortener `
  -DbPassword securepassword `
  -DbName urlshortener_db
```

### Checking Migration Status

**See which migrations have been applied:**
```bash
docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "SELECT version, applied_at FROM schema_migrations ORDER BY version;"
```

**See pending migrations:**
```bash
# List all migration files
ls configs/migrations/*_up.sql

# Compare with applied migrations
docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "SELECT version FROM schema_migrations;"
```

## Migration Best Practices

### ✅ DO:

1. **Always create both up and down migrations** - Even if you never plan to rollback
2. **Use IF EXISTS / IF NOT EXISTS** - Makes migrations idempotent
3. **Add comments** - Explain what the migration does and why
4. **Test rollbacks** - Ensure the down migration actually reverses the up migration
5. **Use transactions** - Wrap complex migrations in BEGIN/COMMIT
6. **Add indexes** - Don't forget to index new columns that will be queried
7. **Provide default values** - When adding NOT NULL columns to existing tables
8. **Document dependencies** - Note if a migration depends on application code changes

### ❌ DON'T:

1. **Don't modify existing migrations** - Once applied in production, never change them
2. **Don't skip version numbers** - Use sequential numbering
3. **Don't mix DDL and data changes** - Keep schema and data migrations separate
4. **Don't use DROP TABLE without IF EXISTS** - Can cause issues on rollback
5. **Don't forget foreign keys** - Maintain referential integrity
6. **Don't hardcode data** - Use appropriate default values instead

## Example Migration Template

```sql
-- Migration: <Feature Name>
-- Description: <Detailed description of what this does>
-- Dependencies: <Any app code or other migrations this depends on>
-- Author: <Your name>
-- Date: <YYYY-MM-DD>

-- Start transaction for atomic operation
BEGIN;

-- Your changes here
ALTER TABLE table_name ADD COLUMN new_column VARCHAR(100);

-- Add indexes for performance
CREATE INDEX IF NOT EXISTS idx_table_new_column ON table_name(new_column);

-- Add constraints
ALTER TABLE table_name ADD CONSTRAINT constraint_name CHECK (new_column IS NOT NULL);

-- Add documentation
COMMENT ON COLUMN table_name.new_column IS 'Description of what this column stores';

-- Commit transaction
COMMIT;
```

## Troubleshooting

### Migration fails to apply

1. Check the error message in logs:
   ```bash
   docker-compose -f docker/docker-compose.yml logs app
   ```

2. Connect to database and check table structure:
   ```bash
   docker exec -it docker-postgres-1 psql -U urlshortener -d urlshortener_db
   \d table_name
   ```

3. Check if migration was partially applied:
   ```bash
   docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "SELECT version FROM schema_migrations WHERE version = 'XXX';"
   ```

4. If needed, manually fix and retry:
   ```bash
   # Fix the database state manually
   # Delete the failed migration record
   docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "DELETE FROM schema_migrations WHERE version = 'XXX';"

   # Fix the migration file
   # Restart the container to retry
   docker-compose -f docker/docker-compose.yml restart app
   ```

### Migration already applied but not in database

If `schema_migrations` table is out of sync:

```bash
# Manually record the migration
docker exec docker-postgres-1 psql -U urlshortener -d urlshortener_db -c "INSERT INTO schema_migrations (version) VALUES ('XXX') ON CONFLICT DO NOTHING;"
```

### Starting fresh

To completely reset the database:

```bash
# Stop containers and remove volumes
docker-compose -f docker/docker-compose.yml down -v

# Start again (this will recreate everything)
docker-compose -f docker/docker-compose.yml up -d
```

## Migration History

| Version | Description | Date |
|---------|-------------|------|
| 000 | Initial schema migrations table | 2025-12-23 |
| 002 | Polar payment integration | 2025-12-23 |
| 006 | Add name column to users | 2025-12-23 |
| 007 | Fix users column names (email_verified) | 2025-12-23 |
| 008 | Add missing users columns (provider, account_type) | 2025-12-23 |
| 009 | Add verification tables | 2025-12-23 |
| 010 | Fix user_sessions table | 2025-12-23 |
| 011 | Add custom_alias to url_mappings | 2025-12-23 |

## See Also

- [Database Schema Documentation](../configs/database.sql)
- [Docker Setup Guide](../docker/README.md)
- [Development Guide](../README.md)
