-- ============================================================================
-- ROLLBACK: Analytics Aggregation Tables (Migration 004)
-- ============================================================================
-- Description: Removes pre-aggregated analytics tables for hourly/daily/weekly summaries
-- Data Loss: ALL pre-aggregated analytics data (weeks/months of summarized data)
-- Impact: Dashboard performance degradation until data is re-aggregated
-- Dependencies: None
-- Estimated Time: 1-3 seconds
-- Reversible: Yes (re-apply migration 004_analytics_tables_up.sql)
--
-- Pre-rollback Checklist:
-- [ ] Database backup created
-- [ ] Export aggregated data to JSON/CSV for historical records
-- [ ] Raw click_events data is intact (aggregations can be regenerated)
-- [ ] Analytics dashboard can handle missing aggregated data
-- [ ] Aggregation jobs are stopped/paused
--
-- IMPORTANT:
-- This rollback deletes aggregated analytics data that may represent
-- weeks or months of pre-calculated summaries. While raw click_events
-- remain intact, regenerating aggregations is time-consuming.
--
-- Consider exporting data first:
-- COPY hourly_analytics TO '/tmp/hourly_analytics_backup.csv' CSV HEADER;
-- COPY daily_analytics TO '/tmp/daily_analytics_backup.csv' CSV HEADER;
-- COPY weekly_analytics TO '/tmp/weekly_analytics_backup.csv' CSV HEADER;
-- COPY monthly_analytics TO '/tmp/monthly_analytics_backup.csv' CSV HEADER;
--
-- Safety: Wrapped in transaction, auto-rollback on error
-- ============================================================================

BEGIN;

-- Safety check: Verify migration is applied
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        RAISE NOTICE 'schema_migrations table does not exist, skipping version check';
    ELSIF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '004') THEN
        RAISE EXCEPTION 'Migration 004 is not applied, cannot rollback';
    END IF;
END $$;

-- Display data loss warning with row counts
DO $$
DECLARE
    hourly_count INTEGER := 0;
    daily_count INTEGER := 0;
    weekly_count INTEGER := 0;
    monthly_count INTEGER := 0;
    country_count INTEGER := 0;
    device_count INTEGER := 0;
    referrer_count INTEGER := 0;
    state_count INTEGER := 0;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'hourly_analytics') THEN
        SELECT COUNT(*) INTO hourly_count FROM hourly_analytics;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'daily_analytics') THEN
        SELECT COUNT(*) INTO daily_count FROM daily_analytics;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'weekly_analytics') THEN
        SELECT COUNT(*) INTO weekly_count FROM weekly_analytics;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'monthly_analytics') THEN
        SELECT COUNT(*) INTO monthly_count FROM monthly_analytics;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'country_analytics_summary') THEN
        SELECT COUNT(*) INTO country_count FROM country_analytics_summary;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'device_type_summary') THEN
        SELECT COUNT(*) INTO device_count FROM device_type_summary;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'referrer_summary') THEN
        SELECT COUNT(*) INTO referrer_count FROM referrer_summary;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'analytics_aggregation_state') THEN
        SELECT COUNT(*) INTO state_count FROM analytics_aggregation_state;
    END IF;

    RAISE WARNING 'About to delete aggregated analytics data:';
    RAISE WARNING '  - % hourly records', hourly_count;
    RAISE WARNING '  - % daily records', daily_count;
    RAISE WARNING '  - % weekly records', weekly_count;
    RAISE WARNING '  - % monthly records', monthly_count;
    RAISE WARNING '  - % country summaries', country_count;
    RAISE WARNING '  - % device summaries', device_count;
    RAISE WARNING '  - % referrer summaries', referrer_count;
    RAISE WARNING '  - % aggregation state records', state_count;
END $$;

-- Drop objects in reverse dependency order

-- Step 1: Drop view first (depends on tables)
DROP VIEW IF EXISTS analytics_storage_stats;

-- Step 2: Drop state tracking table
DROP TABLE IF EXISTS analytics_aggregation_state CASCADE;

-- Step 3: Drop summary tables (no inter-dependencies)
DROP TABLE IF EXISTS referrer_summary CASCADE;
DROP TABLE IF EXISTS device_type_summary CASCADE;
DROP TABLE IF EXISTS country_analytics_summary CASCADE;

-- Step 4: Drop time-based analytics tables (in reverse chronological order)
DROP TABLE IF EXISTS monthly_analytics CASCADE;
DROP TABLE IF EXISTS weekly_analytics CASCADE;
DROP TABLE IF EXISTS daily_analytics CASCADE;
DROP TABLE IF EXISTS hourly_analytics CASCADE;

-- Remove migration from tracking table
DELETE FROM schema_migrations WHERE version = '004';

-- Log successful rollback
RAISE NOTICE 'Migration 004 rolled back successfully';
RAISE NOTICE 'Dropped 8 analytics aggregation tables';
RAISE NOTICE 'REMINDER: Raw click_events data is intact, aggregations can be regenerated';

COMMIT;
