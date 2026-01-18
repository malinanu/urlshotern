# Database Migrations Integration Guide

## Overview

This document describes the unified database migration system integrated into both development and production Docker environments. The system automatically runs database migrations before starting the application, ensuring schema consistency across all environments.

## Architecture

### Migration Directories

The system manages two migration directories with different patterns:

1. **`configs/migrations/`** (27 files) - Core schema migrations
   - Pattern: Paired up/down files (`001_name_up.sql`, `001_name_down.sql`)
   - Contains foundational schema, API keys, analytics, etc.
   - Tracked with numeric versions: `001`, `002`, `003`, etc.

2. **`migrations/`** (14 files) - Admin and feature migrations
   - Pattern: Monolithic SQL files (`001_feature.sql`)
   - Contains admin panel, SEO features, content management
   - Tracked with 'M' prefix: `M001`, `M002`, `M003`, etc.

### Migration Scripts

#### Core Scripts

- **`scripts/run-all-migrations.sh`** - Unified orchestrator
  - Executes both migration directories in correct sequence
  - Entry point for Docker containers

- **`scripts/run-migrations.sh`** - Up/down migration runner
  - Handles paired up/down SQL files
  - Manages `configs/migrations/` directory

- **`scripts/run-migrations-monolithic.sh`** - Monolithic migration runner
  - Handles single-file migrations
  - Manages `migrations/` directory
  - NEW: Created for admin/feature migrations

### Version Tracking

All migrations are tracked in the `schema_migrations` table:

```sql
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
```

**Version Format:**
- Core migrations: `001`, `002`, `003`, ... `027`
- Admin/feature migrations: `M001`, `M002`, `M003`, ... `M014`

## Execution Flow

### Development Environment

1. Container starts (`docker-compose -f docker/docker-compose.local.yml up`)
2. `docker/entrypoint-backend.sh` executes
3. Calls `scripts/run-all-migrations.sh`
4. Runs `scripts/run-migrations.sh` (core migrations)
5. Runs `scripts/run-migrations-monolithic.sh` (admin migrations)
6. Starts application with air (hot reload)

### Production Environment

1. Container starts (`docker-compose -f docker/docker-compose.yml up`)
2. `docker/entrypoint.sh` executes
3. Calls `scripts/run-all-migrations.sh`
4. Runs `scripts/run-migrations.sh` (core migrations)
5. Runs `scripts/run-migrations-monolithic.sh` (admin migrations)
6. Starts application binary

## Configuration Changes

### Development Dockerfile (`Dockerfile.backend.dev`)

```dockerfile
# Copy migration scripts and files
COPY scripts/ /app/scripts/
COPY configs/migrations/ /app/configs/migrations/
COPY migrations/ /app/migrations/  # ← NEW: Admin/feature migrations
RUN find /app/scripts -name "*.sh" -exec sed -i 's/\r$//' {} \; && chmod +x /app/scripts/*.sh
```

### Production Dockerfile (`docker/Dockerfile`)

Complete rewrite to support migrations:

- Added PostgreSQL client and bash to runtime image
- Copies both migration directories
- Copies all migration scripts
- Fixes line endings and permissions
- Uses entrypoint for pre-startup migrations
- Maintains non-root user (appuser) security

## Testing

### Fresh Database Test (Development)

```bash
# Tear down and recreate everything
docker-compose -f docker/docker-compose.local.yml down -v
docker-compose -f docker/docker-compose.local.yml up --build

# Monitor logs for:
# - "Running all database migrations..."
# - "Running core migrations from configs/migrations/..."
# - "Running admin/feature migrations from migrations/..."
# - Migration counts and success messages
```

### Verify Migrations in Database

```bash
# Connect to PostgreSQL container
docker exec -it <postgres-container-name> psql -U urlshortener -d urlshortener_db

# Check applied migrations
SELECT version, applied_at
FROM schema_migrations
ORDER BY version;

# Expected output:
#  version | applied_at
# ---------+-------------------
#  001     | 2026-01-19 ...
#  002     | 2026-01-19 ...
#  ...
#  027     | 2026-01-19 ...
#  M001    | 2026-01-19 ...
#  M002    | 2026-01-19 ...
#  ...
#  M014    | 2026-01-19 ...
```

### Production Build Test

```bash
# Build production image
docker build -f docker/Dockerfile -t urlshortener:test .

# Verify migration files are included
docker run --rm urlshortener:test ls -la /home/appuser/scripts
docker run --rm urlshortener:test ls -la /home/appuser/configs/migrations
docker run --rm urlshortener:test ls -la /home/appuser/migrations

# Test production stack
docker-compose -f docker/docker-compose.yml up --build
```

### Idempotency Test

```bash
# Restart containers (migrations should be skipped)
docker-compose -f docker/docker-compose.local.yml restart backend

# Check logs - should see:
# "Migration 001 already applied, skipping..."
# "Migration M001 already applied, skipping..."
```

## Migration Script Details

### run-all-migrations.sh

Orchestrates the complete migration process:

```bash
#!/bin/bash
set -e

echo "🔄 Running all database migrations..."

# Run core migrations first
export MIGRATIONS_DIR=/app/configs/migrations
/app/scripts/run-migrations.sh

# Run admin/feature migrations
export MIGRATIONS_DIR=/app/migrations
/app/scripts/run-migrations-monolithic.sh

echo "✅ All migrations complete!"
```

### run-migrations-monolithic.sh

Key features:
- Extracts version numbers from filenames (e.g., `001_feature.sql` → `M001`)
- Checks `schema_migrations` table for applied migrations
- Skips already-applied migrations
- Inserts version record after successful application
- Uses `ON CONFLICT DO NOTHING` for safety

## Troubleshooting

### Migration Fails to Apply

**Symptom:** Error message in Docker logs during migration

**Solution:**
1. Check migration SQL syntax
2. Verify database connection parameters
3. Check migration file permissions
4. Review logs for specific error message

```bash
# View detailed logs
docker-compose -f docker/docker-compose.local.yml logs backend
```

### Migration Script Not Found

**Symptom:** "Migration script not found, skipping migrations"

**Solution:**
1. Verify Dockerfile copies script files
2. Rebuild Docker image: `docker-compose up --build`
3. Check script permissions in container

```bash
# Inspect container
docker exec -it <container-name> ls -la /app/scripts
```

### Line Ending Issues

**Symptom:** Scripts fail with `/bin/bash^M: bad interpreter`

**Solution:**
Scripts are automatically fixed:
- `.gitattributes` ensures LF endings in Git
- Dockerfiles run `sed -i 's/\r$//'` to strip CRLF
- Manual fix: `dos2unix scripts/*.sh`

### Duplicate Migrations

**Symptom:** Same migration appears to run twice

**Solution:**
- System uses `ON CONFLICT DO NOTHING` - safe by design
- Verify `schema_migrations` table shows correct versions
- Check that version extraction logic works correctly

```sql
-- Check for duplicates
SELECT version, COUNT(*)
FROM schema_migrations
GROUP BY version
HAVING COUNT(*) > 1;
```

### Permission Denied

**Symptom:** "Permission denied" when executing scripts

**Solution:**
```bash
# Fix permissions in Dockerfile (already included)
RUN chmod +x ./scripts/*.sh

# Or manually in container
docker exec -it <container-name> chmod +x /app/scripts/*.sh
```

## Adding New Migrations

### Core Schema Migration (configs/migrations/)

1. Create paired files:
   ```bash
   configs/migrations/028_new_feature_up.sql
   configs/migrations/028_new_feature_down.sql
   ```

2. Will be tracked as version `028`

3. Automatically runs on next container start

### Admin/Feature Migration (migrations/)

1. Create monolithic file:
   ```bash
   migrations/015_new_admin_feature.sql
   ```

2. Will be tracked as version `M015`

3. Automatically runs on next container start

## File Structure

```
URLshorter/
├── configs/
│   └── migrations/          # Core schema migrations (up/down pattern)
│       ├── 000_schema_migrations.sql
│       ├── 001_add_api_keys_up.sql
│       ├── 001_add_api_keys_down.sql
│       └── ... (27 files total)
├── migrations/              # Admin/feature migrations (monolithic)
│   ├── 001_phase1_admin_foundation.sql
│   ├── 002_admin_authentication.sql
│   └── ... (14 files total)
├── scripts/
│   ├── run-all-migrations.sh           # NEW: Unified orchestrator
│   ├── run-migrations.sh               # Existing: Up/down runner
│   └── run-migrations-monolithic.sh    # NEW: Monolithic runner
├── docker/
│   ├── entrypoint.sh                   # Production entrypoint (updated)
│   ├── entrypoint-backend.sh           # Dev entrypoint (updated)
│   └── Dockerfile                      # Production Dockerfile (rewritten)
├── Dockerfile.backend.dev              # Dev Dockerfile (updated)
└── .gitattributes                      # NEW: Line ending configuration
```

## Security Considerations

### Non-Root User

Production container runs as `appuser` (non-root):
- Migrations run before switching users
- Files owned by `appuser:appuser`
- Enhances security posture

### Database Credentials

- Use environment variables for sensitive data
- Never hardcode passwords
- Consider using Docker secrets in production

```yaml
# docker-compose.yml example
environment:
  DB_PASSWORD: ${DB_PASSWORD}  # From .env file
```

### Migration Safety

- Always test migrations in development first
- Use transactions where possible
- Keep down migrations for rollback capability
- Back up database before major migrations

## Performance

### Migration Speed

- First run: Applies all migrations (~41 total)
- Subsequent runs: Skips applied migrations (fast)
- Database connection pooling reduces overhead

### Startup Time Impact

- Fresh database: +10-15 seconds
- Existing database: +1-2 seconds
- Parallel execution not implemented (sequential is safer)

## Environment Variables

Required variables for migration scripts:

```bash
DB_HOST=localhost        # Database hostname
DB_PORT=5432            # Database port
DB_USER=urlshortener    # Database user
DB_PASSWORD=***         # Database password
DB_NAME=urlshortener_db # Database name
```

Set in `docker-compose.yml`:

```yaml
environment:
  DB_HOST: postgres
  DB_PORT: 5432
  DB_USER: urlshortener
  DB_PASSWORD: ${DB_PASSWORD}
  DB_NAME: urlshortener_db
```

## Best Practices

1. **Always use migrations for schema changes** - Never modify database directly
2. **Test locally first** - Verify migrations work before pushing
3. **Keep migrations small** - One logical change per migration
4. **Write down migrations** - Always provide rollback capability
5. **Version control** - All migrations tracked in Git
6. **Sequential numbering** - Avoid gaps in version numbers
7. **Descriptive names** - Use clear, descriptive migration filenames

## References

- [PostgreSQL Documentation](https://www.postgresql.org/docs/)
- [Docker Multi-Stage Builds](https://docs.docker.com/build/building/multi-stage/)
- [Database Migration Best Practices](https://www.liquibase.org/get-started/best-practices)

## Support

For issues or questions:
1. Check troubleshooting section above
2. Review Docker logs: `docker-compose logs backend`
3. Verify database state: `SELECT * FROM schema_migrations;`
4. Create GitHub issue with logs and error messages

---

**Last Updated:** 2026-01-19
**Version:** 1.0.0
**Author:** Development Team
