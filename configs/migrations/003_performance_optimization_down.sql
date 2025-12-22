-- ============================================================================
-- ROLLBACK: Performance Optimization Indexes (Migration 003)
-- ============================================================================
-- Description: Removes performance-optimized indexes created for high-traffic queries
-- Data Loss: None (indexes only, no data deleted)
-- Performance Impact: Significant query performance degradation expected
-- Dependencies: None
-- Estimated Time: 2-5 seconds (depending on table sizes)
-- Reversible: Yes (re-apply migration 003_performance_optimization_up.sql)
--
-- Pre-rollback Checklist:
-- [ ] Database backup created (optional - no data loss risk)
-- [ ] Application can handle slower query performance
-- [ ] Consider scheduling during low-traffic period
-- [ ] Monitor query performance after rollback
--
-- Note: Index drops do not require CONCURRENTLY flag (unlike creation)
-- Note: This rollback is relatively safe - only affects query performance
--
-- Safety: Wrapped in transaction, auto-rollback on error
-- ============================================================================

BEGIN;

-- Safety check: Verify migration is applied
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        RAISE NOTICE 'schema_migrations table does not exist, skipping version check';
    ELSIF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '003') THEN
        RAISE EXCEPTION 'Migration 003 is not applied, cannot rollback';
    END IF;
END $$;

-- Performance warning
RAISE WARNING 'Dropping performance indexes - query performance will degrade significantly';

-- Drop click_events indexes
DROP INDEX IF EXISTS idx_click_events_short_country_time;
DROP INDEX IF EXISTS idx_click_events_ip;
DROP INDEX IF EXISTS idx_click_events_recent;
DROP INDEX IF EXISTS idx_click_events_referrer;
DROP INDEX IF EXISTS idx_click_events_user_agent;

-- Drop url_mappings indexes
DROP INDEX IF EXISTS idx_url_mappings_user_created;
DROP INDEX IF EXISTS idx_url_mappings_domain_active;
DROP INDEX IF EXISTS idx_url_mappings_team_active;

-- Drop geographic_analytics indexes
DROP INDEX IF EXISTS idx_geographic_analytics_lookup;

-- Drop device_analytics indexes
DROP INDEX IF EXISTS idx_device_analytics_device_type;

-- Drop time_analytics indexes
DROP INDEX IF EXISTS idx_time_analytics_heatmap;

-- Drop referrer_analytics indexes
DROP INDEX IF EXISTS idx_referrer_analytics_lookup;

-- Note: The ANALYZE statements from the up migration are one-time operations
-- The database statistics will remain updated, no rollback needed

-- Remove migration from tracking table
DELETE FROM schema_migrations WHERE version = '003';

-- Log successful rollback
RAISE NOTICE 'Migration 003 rolled back successfully';
RAISE NOTICE 'Dropped 12 performance indexes - monitor query performance';

COMMIT;
