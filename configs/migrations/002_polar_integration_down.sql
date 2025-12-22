-- ============================================================================
-- ROLLBACK: Polar Payment Gateway Integration (Migration 002)
-- ============================================================================
-- Description: Removes Polar subscription tables, webhook tracking, and triggers
-- Data Loss: ALL subscription data, webhook history, and product cache
-- Dependencies: Migration 005 adds columns to polar_webhook_events
--              Must rollback 005 first before rolling back 002
-- Estimated Time: 1-2 seconds
-- Reversible: Yes (re-apply migration 002_polar_integration_up.sql)
--
-- Pre-rollback Checklist:
-- [ ] Database backup created (pg_dump)
-- [ ] Migration 005 has been rolled back (if applied)
-- [ ] Export subscription data to JSON/CSV for recovery
-- [ ] Polar integration disabled in application config
-- [ ] Payment processing suspended
-- [ ] Customer billing records exported
--
-- CRITICAL WARNING:
-- This rollback deletes ALL customer subscription data including:
-- - Active and historical subscriptions
-- - Webhook event history (for idempotency)
-- - Cached product/pricing information
-- This data CANNOT be recovered without a database backup!
--
-- Safety: Wrapped in transaction, auto-rollback on error
-- ============================================================================

BEGIN;

-- Safety check: Verify migration is applied
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        RAISE NOTICE 'schema_migrations table does not exist, skipping version check';
    ELSIF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '002') THEN
        RAISE EXCEPTION 'Migration 002 is not applied, cannot rollback';
    END IF;
END $$;

-- Dependency check: Ensure migration 005 is rolled back first
DO $$
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        IF EXISTS (SELECT 1 FROM schema_migrations WHERE version = '005') THEN
            RAISE EXCEPTION 'Migration 005 depends on 002. Must rollback 005 first.';
        END IF;
    END IF;
END $$;

-- Display data loss warning with counts
DO $$
DECLARE
    subscription_count INTEGER := 0;
    webhook_count INTEGER := 0;
    product_count INTEGER := 0;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'polar_subscriptions') THEN
        SELECT COUNT(*) INTO subscription_count FROM polar_subscriptions;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'polar_webhook_events') THEN
        SELECT COUNT(*) INTO webhook_count FROM polar_webhook_events;
    END IF;
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'polar_products') THEN
        SELECT COUNT(*) INTO product_count FROM polar_products;
    END IF;

    RAISE WARNING 'About to delete:';
    RAISE WARNING '  - % subscription records', subscription_count;
    RAISE WARNING '  - % webhook events', webhook_count;
    RAISE WARNING '  - % cached products', product_count;
END $$;

-- Step 1: Drop triggers first (must drop before dropping function)
DROP TRIGGER IF EXISTS update_polar_webhook_events_updated_at ON polar_webhook_events;
DROP TRIGGER IF EXISTS update_polar_products_updated_at ON polar_products;
DROP TRIGGER IF EXISTS update_polar_subscriptions_updated_at ON polar_subscriptions;

-- Step 2: Drop the trigger function
-- Note: Only drop if it's not used by other tables
-- Check for usage: SELECT event_object_table FROM information_schema.triggers
--                  WHERE action_statement LIKE '%update_updated_at_column%'
DO $$
DECLARE
    function_usage_count INTEGER;
BEGIN
    -- Count how many triggers still use this function
    SELECT COUNT(*) INTO function_usage_count
    FROM information_schema.triggers
    WHERE action_statement LIKE '%update_updated_at_column%';

    IF function_usage_count = 0 THEN
        -- Safe to drop the function
        DROP FUNCTION IF EXISTS update_updated_at_column();
        RAISE NOTICE 'Dropped function update_updated_at_column()';
    ELSE
        RAISE WARNING 'Function update_updated_at_column() is still used by % triggers, not dropping', function_usage_count;
    END IF;
END $$;

-- Step 3: Drop tables (order doesn't matter as there are no foreign keys between them)
DROP TABLE IF EXISTS polar_webhook_events CASCADE;
DROP TABLE IF EXISTS polar_products CASCADE;
DROP TABLE IF EXISTS polar_subscriptions CASCADE;

-- Remove migration from tracking table
DELETE FROM schema_migrations WHERE version = '002';

-- Log successful rollback
RAISE NOTICE 'Migration 002 rolled back successfully';
RAISE NOTICE 'REMINDER: Customer subscription data has been deleted!';

COMMIT;
