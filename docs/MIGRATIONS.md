# Database Migrations Guide

This guide explains how to manage database schema changes using the URL Shortener migration system.

## Table of Contents

- [Overview](#overview)
- [Quick Start](#quick-start)
- [Migration Commands](#migration-commands)
- [Creating New Migrations](#creating-new-migrations)
- [Best Practices](#best-practices)
- [Troubleshooting](#troubleshooting)
- [Migration System Architecture](#migration-system-architecture)

## Overview

The URL Shortener project uses a custom migration system to manage database schema changes safely and reliably. The system provides:

- **Version Tracking**: Track which migrations have been applied
- **Rollback Support**: Safely rollback migrations when needed
- **Safety Checks**: Prevent accidental data loss in production
- **Cross-Platform**: Works on Linux, macOS, and Windows

### Migration File Structure

```
configs/migrations/
├── 000_schema_migrations.sql        # Migration tracking table
├── 001_add_api_keys_up.sql          # Migration 001 (forward)
├── 001_add_api_keys_down.sql        # Migration 001 (rollback)
├── 002_polar_integration_up.sql     # Migration 002 (forward)
├── 002_polar_integration_down.sql   # Migration 002 (rollback)
...
```

Each migration has two files:
- `*_up.sql`: Applies the migration (adds tables, columns, indexes)
- `*_down.sql`: Rolls back the migration (removes changes)

## Quick Start

### Check Migration Status

View which migrations have been applied:

```bash
# Using make (recommended)
make migrate-status

# Or directly
./scripts/migrate.sh status
```

### Apply a Migration

Apply a specific migration (e.g., migration 001):

```bash
# Using make
make migrate-up VERSION=001

# Or directly
./scripts/migrate.sh up 001
```

### Rollback a Migration

Rollback a specific migration:

```bash
# Using make
make migrate-down VERSION=005

# Or directly
./scripts/migrate.sh down 005
```

### Rollback to a Specific Version

Rollback all migrations above a certain version:

```bash
# Rollback to version 003 (rolls back 005, 004 in order)
make migrate-rollback-to VERSION=003

# Or directly
./scripts/migrate.sh rollback-to 003
```

## Migration Commands

### Linux / macOS

All commands use the `scripts/migrate.sh` bash script:

```bash
# Check status
./scripts/migrate.sh status

# Apply migration
./scripts/migrate.sh up 001

# Rollback migration
./scripts/migrate.sh down 005

# Rollback to version
./scripts/migrate.sh rollback-to 003

# Dry run (preview without executing)
./scripts/migrate.sh dry-run down 005
```

### Windows PowerShell

Use the PowerShell script:

```powershell
# Check status
.\scripts\migrate.ps1 status

# Apply migration
.\scripts\migrate.ps1 up 001

# Rollback migration
.\scripts\migrate.ps1 down 005

# Rollback to version
.\scripts\migrate.ps1 rollback-to 003
```

### Using Makefile (Cross-Platform)

```bash
# Check status
make migrate-status

# Apply migration
make migrate-up VERSION=001

# Rollback migration
make migrate-down VERSION=005

# Rollback to version
make migrate-rollback-to VERSION=003

# Dry run
make migrate-dry-run DIRECTION=down VERSION=005
```

## Creating New Migrations

### Step 1: Choose a Version Number

Migrations are numbered sequentially (001, 002, 003, etc.). The next available number is the highest existing number + 1.

### Step 2: Create Up Migration

Create the "up" migration file:

```bash
# Create file: configs/migrations/006_add_feature_name_up.sql
```

Example content:

```sql
-- ============================================================================
-- Migration 006: Add Feature Name
-- ============================================================================
-- Description: Adds support for [feature description]
-- Author: Your Name
-- Date: 2024-01-15
-- ============================================================================

BEGIN;

-- Create table
CREATE TABLE IF NOT EXISTS feature_table (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    feature_data JSONB NOT NULL DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Add indexes
CREATE INDEX IF NOT EXISTS idx_feature_table_user_id
ON feature_table(user_id);

-- Add comments
COMMENT ON TABLE feature_table IS 'Stores feature data';

-- Record migration
INSERT INTO schema_migrations (version, description, checksum)
VALUES ('006', 'add_feature_name', 'CALCULATED_BY_SCRIPT');

COMMIT;
```

### Step 3: Create Down Migration

Create the "down" migration file:

```bash
# Create file: configs/migrations/006_add_feature_name_down.sql
```

Example content:

```sql
-- ============================================================================
-- ROLLBACK: Add Feature Name (Migration 006)
-- ============================================================================
-- Data Loss: [Describe what data will be deleted]
-- Dependencies: [List dependent migrations]
-- ============================================================================

BEGIN;

-- Safety check
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '006') THEN
        RAISE EXCEPTION 'Migration 006 is not applied, cannot rollback';
    END IF;
END $$;

-- Drop table
DROP TABLE IF EXISTS feature_table CASCADE;

-- Remove from tracking
DELETE FROM schema_migrations WHERE version = '006';

COMMIT;
```

### Step 4: Test the Migration

1. **Test applying the migration:**
   ```bash
   make migrate-up VERSION=006
   ```

2. **Verify the changes:**
   ```bash
   psql -d urlshortener_db -c "\dt"  # List tables
   psql -d urlshortener_db -c "\d feature_table"  # Describe table
   ```

3. **Test rollback:**
   ```bash
   make migrate-down VERSION=006
   ```

4. **Verify rollback:**
   ```bash
   psql -d urlshortener_db -c "\dt"  # Verify table is gone
   ```

5. **Re-apply to verify idempotency:**
   ```bash
   make migrate-up VERSION=006
   ```

## Best Practices

### 1. Always Use Transactions

Wrap migrations in `BEGIN;` and `COMMIT;`:

```sql
BEGIN;
-- Your migration code here
COMMIT;
```

This ensures automatic rollback if any command fails.

### 2. Use IF NOT EXISTS / IF EXISTS

Make migrations idempotent:

```sql
CREATE TABLE IF NOT EXISTS my_table (...);
CREATE INDEX IF NOT EXISTS idx_name ON my_table(column);
DROP TABLE IF EXISTS my_table;
```

### 3. Add Safety Checks in Rollbacks

Always verify migration is applied before rolling back:

```sql
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '006') THEN
        RAISE EXCEPTION 'Migration 006 is not applied, cannot rollback';
    END IF;
END $$;
```

### 4. Document Data Loss

Clearly document what data will be lost in rollback:

```sql
-- Data Loss: ALL feature data will be deleted
-- Impact: Users will need to reconfigure feature settings
```

### 5. Check Dependencies

If migration B depends on migration A, check in rollback:

```sql
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM schema_migrations WHERE version = '007') THEN
        RAISE EXCEPTION 'Migration 007 depends on 006. Rollback 007 first.';
    END IF;
END $$;
```

### 6. Use CONCURRENTLY for Indexes (Up Migrations Only)

For large tables, create indexes without locking:

```sql
CREATE INDEX CONCURRENTLY IF NOT EXISTS idx_name
ON large_table(column);
```

**Note:** `CONCURRENTLY` cannot be used inside a transaction. For these cases, remove `BEGIN;`/`COMMIT;` and handle errors manually.

### 7. Test Before Production

1. Test in development environment
2. Test in staging environment with production-like data
3. Create database backup before applying in production
4. Apply during low-traffic maintenance window

### 8. Version Control

Commit both `*_up.sql` and `*_down.sql` files together:

```bash
git add configs/migrations/006_*
git commit -m "Add migration 006: [description]"
```

## Troubleshooting

### Migration Already Applied

**Problem:** "Migration 001 is already applied"

**Solution:** This is normal. The migration system tracks applied migrations. If you need to re-apply, first rollback then apply again:

```bash
make migrate-down VERSION=001
make migrate-up VERSION=001
```

### Migration Not Found

**Problem:** "Migration file not found: 006_*_up.sql"

**Solution:** Ensure the migration file exists and follows the naming convention:
- File must be in `configs/migrations/`
- Must match pattern: `NNN_description_up.sql`

### Cannot Connect to Database

**Problem:** "Cannot connect to database"

**Solution:** Set environment variables:

```bash
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=urlshortener_db
export DB_USER=urlshortener
export DB_PASSWORD=your_password

make migrate-status
```

### Schema Migrations Table Missing

**Problem:** "schema_migrations table does not exist"

**Solution:** The system will automatically create it, or manually apply:

```bash
psql -d urlshortener_db -f configs/migrations/000_schema_migrations.sql
```

### Rollback Failed - Dependent Migration

**Problem:** "Migration 007 depends on 006. Rollback 007 first."

**Solution:** Rollback migrations in reverse order:

```bash
make migrate-down VERSION=007
make migrate-down VERSION=006
```

Or use rollback-to to handle automatically:

```bash
make migrate-rollback-to VERSION=005
```

### Production Safety Check

**Problem:** Prompted "Type 'YES I HAVE A BACKUP' to continue"

**Solution:** This is intentional for production safety. Create a backup first:

```bash
pg_dump -U urlshortener -h localhost -d urlshortener_db > backup_$(date +%Y%m%d_%H%M%S).sql
```

Then retry the rollback and confirm with the exact phrase.

## Migration System Architecture

### Components

1. **Migration Files** (`configs/migrations/*.sql`)
   - SQL files containing schema changes
   - Separate up/down files for each migration

2. **Migration Script** (`scripts/migrate.sh` or `migrate.ps1`)
   - Orchestrates migration execution
   - Handles safety checks and transaction management
   - Tracks timing and checksums

3. **Tracking Table** (`schema_migrations`)
   - Records which migrations have been applied
   - Stores execution metadata (time, user, checksum)

4. **Makefile Targets** (`Makefile`)
   - Convenient wrappers around migration scripts
   - Cross-platform support

### Migration Flow

**Applying a Migration:**

```
1. Check prerequisites (psql available, database accessible)
2. Ensure schema_migrations table exists
3. Check if migration is already applied
4. Calculate file checksum
5. Execute migration SQL file
6. Record in schema_migrations table
7. Report success/failure
```

**Rolling Back a Migration:**

```
1. Check prerequisites
2. Verify migration is applied
3. Check for dependent migrations
4. Production safety check (if applicable)
5. Execute rollback SQL file
6. Remove from schema_migrations table
7. Report success/failure
```

### Environment Variables

| Variable | Description | Default |
|----------|-------------|---------|
| `DB_HOST` | Database host | localhost |
| `DB_PORT` | Database port | 5432 |
| `DB_NAME` | Database name | urlshortener_db |
| `DB_USER` | Database user | urlshortener |
| `DB_PASSWORD` | Database password | (required) |
| `ENVIRONMENT` | Environment name | development |

Set `ENVIRONMENT=production` to enable safety prompts for rollbacks.

## Related Documentation

- [Rollback Procedures](ROLLBACK_PROCEDURES.md) - Detailed rollback runbook
- [Configuration Guide](CONFIGURATION.md) - Database configuration reference
- [Troubleshooting Guide](TROUBLESHOOTING.md) - General troubleshooting tips

## Support

For migration issues or questions:
1. Check this guide and the [Rollback Procedures](ROLLBACK_PROCEDURES.md)
2. Review migration file comments for specific guidance
3. Check migration script output for detailed error messages
4. Create a backup before attempting manual fixes
