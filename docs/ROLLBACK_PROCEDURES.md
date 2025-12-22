# Database Rollback Procedures

This runbook provides detailed step-by-step procedures for rolling back database migrations safely.

## Table of Contents

- [General Rollback Guidelines](#general-rollback-guidelines)
- [Pre-Rollback Checklist](#pre-rollback-checklist)
- [Migration-Specific Rollback Procedures](#migration-specific-rollback-procedures)
  - [Migration 001: API Keys](#migration-001-api-keys)
  - [Migration 002: Polar Integration](#migration-002-polar-integration)
  - [Migration 003: Performance Optimization](#migration-003-performance-optimization)
  - [Migration 004: Analytics Tables](#migration-004-analytics-tables)
  - [Migration 005: Webhook Retry Fields](#migration-005-webhook-retry-fields)
- [Emergency Rollback Procedures](#emergency-rollback-procedures)
- [Recovery Procedures](#recovery-procedures)

## General Rollback Guidelines

### When to Rollback

Rollback a migration when:
- Migration caused unexpected application errors
- Performance degradation after migration
- Data corruption detected
- Deployment needs to be reverted
- Migration was applied to wrong environment

### When NOT to Rollback

Avoid rollback when:
- Production system has been running for days with the migration
- Users have created significant data using new features
- Dependent systems have been updated to use new schema
- Alternative fixes are available (hotfix, data migration)

## Pre-Rollback Checklist

**CRITICAL:** Complete ALL items before proceeding with rollback

### 1. Create Database Backup

```bash
# Create timestamped backup
TIMESTAMP=$(date +%Y%m%d_%H%M%S)
pg_dump -U $DB_USER -h $DB_HOST -d $DB_NAME > "backup_before_rollback_${TIMESTAMP}.sql"

# Verify backup file exists and is not empty
ls -lh backup_before_rollback_${TIMESTAMP}.sql
```

### 2. Export Critical Data (If Applicable)

For migrations that delete data, export to CSV/JSON:

```bash
# Example: Export subscription data before rolling back migration 002
psql -d urlshortener_db -c "\COPY polar_subscriptions TO 'polar_subscriptions_backup.csv' CSV HEADER"
psql -d urlshortener_db -c "\COPY polar_webhook_events TO 'polar_webhook_events_backup.csv' CSV HEADER"
```

### 3. Stop Application Services

```bash
# Using Docker Compose
docker-compose down

# Or using systemd
sudo systemctl stop urlshortener
```

### 4. Verify Database Connection

```bash
# Test connection
psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c "SELECT version();"
```

### 5. Check Migration Status

```bash
# Verify which migrations are applied
make migrate-status
```

### 6. Review Rollback Script

```bash
# View the rollback SQL that will be executed
cat configs/migrations/00X_migration_name_down.sql
```

### 7. Notify Team (Production Only)

- Alert team members about impending rollback
- Post in team chat/Slack
- Create incident ticket if applicable
- Schedule maintenance window if possible

## Migration-Specific Rollback Procedures

### Migration 001: API Keys

**Data Loss:** ALL API keys and usage history

**Impact:**
- Users will need to regenerate API keys
- API authentication will break until keys regenerated
- Historical usage data will be lost

**Dependencies:** None

**Procedure:**

1. **Export API key data (if needed for records)**
   ```sql
   \COPY api_keys TO 'api_keys_backup.csv' CSV HEADER
   \COPY api_key_usage TO 'api_key_usage_backup.csv' CSV HEADER
   ```

2. **Notify users about API key regeneration**
   - Send email notification
   - Post announcement on dashboard
   - Update API documentation

3. **Update application config**
   - Disable API key authentication in code
   - Or ensure graceful fallback if keys missing

4. **Execute rollback**
   ```bash
   make migrate-down VERSION=001
   ```

5. **Verify rollback**
   ```sql
   -- Should return no rows
   SELECT * FROM information_schema.tables WHERE table_name IN ('api_keys', 'api_key_usage');
   ```

6. **Restart application**
   ```bash
   docker-compose up -d
   # Or
   sudo systemctl start urlshortener
   ```

7. **Test application**
   - Verify app starts without errors
   - Test URL creation without API keys
   - Check logs for errors

**Rollback Time:** < 1 second

**Recovery:** Re-apply migration 001, users regenerate API keys

---

### Migration 002: Polar Integration

**Data Loss:** ALL subscription data, webhook history, product cache

**Impact:**
- Customer billing information deleted
- Webhook event history lost (affects idempotency)
- Product/pricing cache cleared
- Payment processing will break

**Dependencies:** Migration 005 (must rollback first if applied)

**⚠️ CRITICAL:** This is the highest-risk rollback. Consider alternatives before proceeding.

**Procedure:**

1. **Check for dependent migrations**
   ```bash
   # If migration 005 is applied, must rollback it first
   make migrate-status

   # If 005 is applied:
   make migrate-down VERSION=005
   ```

2. **Export ALL Polar data**
   ```sql
   \COPY polar_subscriptions TO 'polar_subscriptions_backup.csv' CSV HEADER
   \COPY polar_webhook_events TO 'polar_webhook_events_backup.csv' CSV HEADER
   \COPY polar_products TO 'polar_products_backup.csv' CSV HEADER

   -- Also export to JSON for easier recovery
   \COPY (SELECT row_to_json(polar_subscriptions) FROM polar_subscriptions) TO 'polar_subscriptions_backup.json'
   ```

3. **Disable Polar integration in application**
   - Update config to disable payment processing
   - Redirect billing pages to maintenance mode
   - Stop webhook processing

4. **Notify finance/accounting team**
   - Inform about subscription data deletion
   - Provide exported data files
   - Plan for data restoration or re-sync from Polar

5. **Execute rollback**
   ```bash
   # Production requires confirmation
   ENVIRONMENT=production make migrate-down VERSION=002
   # Type: YES I HAVE A BACKUP
   ```

6. **Verify rollback**
   ```sql
   -- Should return no rows
   SELECT * FROM information_schema.tables
   WHERE table_name IN ('polar_subscriptions', 'polar_webhook_events', 'polar_products');

   -- Check if trigger function was removed
   SELECT routine_name FROM information_schema.routines
   WHERE routine_name = 'update_updated_at_column';
   ```

7. **Restart application**
   ```bash
   docker-compose up -d
   ```

8. **Verify application health**
   - Check application logs
   - Test non-billing functionality
   - Verify billing pages show maintenance mode

**Rollback Time:** 1-2 seconds

**Recovery Options:**
1. **Re-apply migration and restore data:**
   ```bash
   make migrate-up VERSION=002
   psql -d urlshortener_db < polar_subscriptions_backup.sql
   ```

2. **Re-sync from Polar API:**
   - Fetch all subscriptions from Polar
   - Rebuild webhook event log from Polar
   - Re-cache products

---

### Migration 003: Performance Optimization

**Data Loss:** None (indexes only)

**Impact:**
- Query performance will degrade significantly
- Dashboard may load slowly
- Analytics queries may timeout
- No data is deleted

**Dependencies:** None

**⚠️ NOTE:** This rollback is low-risk but will hurt performance. Only rollback if indexes cause issues.

**Procedure:**

1. **Monitor current query performance (for comparison)**
   ```sql
   -- Enable query timing
   \timing

   -- Test key queries
   SELECT COUNT(*) FROM click_events WHERE short_code = 'abc123';
   SELECT * FROM url_mappings WHERE user_id = 1 ORDER BY created_at DESC LIMIT 10;
   ```

2. **Execute rollback**
   ```bash
   make migrate-down VERSION=003
   ```

3. **Verify indexes removed**
   ```sql
   -- Check that indexes are gone
   SELECT indexname FROM pg_indexes
   WHERE indexname LIKE 'idx_click_events%'
      OR indexname LIKE 'idx_url_mappings%'
      OR indexname LIKE 'idx_geographic_analytics%'
      OR indexname LIKE 'idx_device_analytics%'
      OR indexname LIKE 'idx_time_analytics%'
      OR indexname LIKE 'idx_referrer_analytics%';
   ```

4. **Monitor query performance**
   ```sql
   -- Re-test queries (will be slower)
   \timing
   SELECT COUNT(*) FROM click_events WHERE short_code = 'abc123';
   ```

5. **Consider alternatives if performance is too slow:**
   - Re-apply migration 003
   - Create subset of indexes (only the critical ones)
   - Optimize queries instead

**Rollback Time:** 2-5 seconds

**Recovery:** Re-apply migration 003 (indexes recreate instantly with CONCURRENTLY)

---

### Migration 004: Analytics Tables

**Data Loss:** Months of pre-aggregated analytics data

**Impact:**
- Dashboard shows incomplete historical data until re-aggregated
- Analytics reports missing aggregated summaries
- Raw click_events data remains intact
- Slow analytics queries until re-aggregation

**Dependencies:** None

**Procedure:**

1. **Export aggregated data**
   ```sql
   \COPY hourly_analytics TO 'hourly_analytics_backup.csv' CSV HEADER
   \COPY daily_analytics TO 'daily_analytics_backup.csv' CSV HEADER
   \COPY weekly_analytics TO 'weekly_analytics_backup.csv' CSV HEADER
   \COPY monthly_analytics TO 'monthly_analytics_backup.csv' CSV HEADER
   \COPY country_analytics_summary TO 'country_analytics_backup.csv' CSV HEADER
   \COPY device_type_summary TO 'device_analytics_backup.csv' CSV HEADER
   \COPY referrer_summary TO 'referrer_analytics_backup.csv' CSV HEADER
   ```

2. **Stop analytics aggregation jobs**
   ```bash
   # Stop cron job or background worker
   # (implementation-specific)
   ```

3. **Execute rollback**
   ```bash
   make migrate-down VERSION=004
   ```

4. **Verify rollback**
   ```sql
   -- Check tables are gone
   SELECT table_name FROM information_schema.tables
   WHERE table_name LIKE '%_analytics' OR table_name LIKE '%_summary';

   -- Verify raw data intact
   SELECT COUNT(*) FROM click_events;
   ```

5. **Update application to handle missing aggregations**
   - Disable dashboard features that use aggregated data
   - Show message: "Analytics data being regenerated"
   - Or fall back to querying raw click_events (slower)

6. **Plan re-aggregation (if needed)**
   - Schedule background job to rebuild aggregations
   - Or restore from backup CSV files

**Rollback Time:** 1-3 seconds

**Recovery:**
1. Re-apply migration 004
2. Restore from CSV backups
3. Or run aggregation jobs to rebuild from click_events

---

### Migration 005: Webhook Retry Fields

**Data Loss:** Retry timestamps and dead letter queue reasons

**Impact:**
- Webhook retry worker will not function
- Historical retry information lost
- Dead letter queue reasons deleted
- retry_count column remains (was pre-existing)

**Dependencies:** Requires migration 002 (polar_webhook_events table)

**Procedure:**

1. **Stop webhook retry worker**
   ```bash
   # Stop the retry worker service/process
   # (implementation-specific)
   ```

2. **Export retry data (optional)**
   ```sql
   \COPY (SELECT * FROM polar_webhook_events WHERE last_retry_at IS NOT NULL)
         TO 'webhook_retry_data_backup.csv' CSV HEADER;

   \COPY (SELECT * FROM polar_webhook_events WHERE dead_letter_reason IS NOT NULL)
         TO 'dead_letter_queue_backup.csv' CSV HEADER;
   ```

3. **Execute rollback**
   ```bash
   make migrate-down VERSION=005
   ```

4. **Verify rollback**
   ```sql
   -- Check columns are removed
   SELECT column_name FROM information_schema.columns
   WHERE table_name = 'polar_webhook_events'
     AND column_name IN ('last_retry_at', 'dead_letter_reason');

   -- Should return 0 rows

   -- Check original index was recreated
   SELECT indexname FROM pg_indexes
   WHERE indexname = 'idx_polar_webhook_events_processing_status';

   -- Should return 1 row
   ```

5. **Update application configuration**
   - Disable webhook retry worker
   - Or update worker to not use retry fields

6. **Restart application**
   ```bash
   docker-compose up -d
   ```

**Rollback Time:** < 1 second

**Recovery:** Re-apply migration 005, webhook retry worker resumes

---

## Emergency Rollback Procedures

### When Migration Fails Mid-Execution

If a migration fails partway through:

1. **Check if transaction rolled back automatically**
   ```sql
   -- Check migration status
   SELECT * FROM schema_migrations WHERE version = 'XXX';

   -- If present, migration succeeded despite error message
   -- If absent, transaction rolled back automatically
   ```

2. **If in inconsistent state (partial changes applied)**
   ```bash
   # Restore from backup
   psql -d urlshortener_db < backup_before_migration.sql

   # Or manually fix:
   # 1. Identify what was changed
   # 2. Manually reverse changes
   # 3. Delete from schema_migrations if needed
   ```

### When Rollback Fails

If rollback fails:

1. **Check error message for cause**
2. **Common issues:**
   - Dependent migration still applied → Rollback dependencies first
   - Table doesn't exist → Already rolled back, check schema_migrations table
   - Permission denied → Check database user permissions

3. **Manual rollback:**
   ```sql
   -- Manually execute rollback SQL line by line
   BEGIN;
   DROP TABLE IF EXISTS problematic_table CASCADE;
   DELETE FROM schema_migrations WHERE version = 'XXX';
   COMMIT;
   ```

### Data Corruption After Rollback

If data appears corrupt after rollback:

1. **Stop application immediately**
2. **Restore from backup**
   ```bash
   # Drop database
   dropdb urlshortener_db

   # Recreate
   createdb urlshortener_db

   # Restore
   psql -d urlshortener_db < backup_before_rollback.sql
   ```

3. **Investigate root cause before retrying**

## Recovery Procedures

### Restore from Backup

Complete database restore:

```bash
# Stop application
docker-compose down

# Drop and recreate database
dropdb -h $DB_HOST -U $DB_USER urlshortener_db
createdb -h $DB_HOST -U $DB_USER urlshortener_db

# Restore from backup
psql -h $DB_HOST -U $DB_USER -d urlshortener_db < backup_file.sql

# Restart application
docker-compose up -d
```

### Restore Specific Tables

If only specific tables need restoration:

```bash
# Extract specific tables from backup
pg_restore -d urlshortener_db -t polar_subscriptions backup_file.sql

# Or using plain SQL backup:
psql -d urlshortener_db -c "DROP TABLE polar_subscriptions CASCADE;"
psql -d urlshortener_db < backup_file.sql
```

### Verify Data Integrity After Recovery

```sql
-- Check row counts match expectations
SELECT 'users' AS table_name, COUNT(*) FROM users
UNION ALL
SELECT 'url_mappings', COUNT(*) FROM url_mappings
UNION ALL
SELECT 'click_events', COUNT(*) FROM click_events;

-- Verify foreign key integrity
SELECT COUNT(*) FROM click_events ce
LEFT JOIN url_mappings um ON ce.short_code = um.short_code
WHERE um.short_code IS NULL;
-- Should return 0

-- Check for null values in required fields
SELECT COUNT(*) FROM users WHERE email IS NULL;
-- Should return 0
```

## Contact Information

**For Rollback Assistance:**
- Database Admin: [contact info]
- DevOps Team: [contact info]
- On-Call Engineer: [contact info]

**Escalation:**
- P1 (Production Down): [escalation procedure]
- P2 (Degraded Performance): [escalation procedure]

---

**Last Updated:** 2024-01-15
**Document Owner:** DevOps Team
**Review Cycle:** Quarterly
