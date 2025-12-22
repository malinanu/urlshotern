-- ============================================================================
-- ROLLBACK: Webhook Retry Fields (Migration 005)
-- ============================================================================
-- Description: Removes webhook retry mechanism fields from polar_webhook_events
-- Data Loss: Retry timestamps and dead letter queue reasons
-- Dependencies: Requires migration 002 (polar_webhook_events table)
-- Estimated Time: <1 second
-- Reversible: Yes (re-apply migration 005_webhook_retry_fields_up.sql)
--
-- Pre-rollback Checklist:
-- [ ] Database backup created (optional - low risk)
-- [ ] Webhook retry worker is stopped
-- [ ] No critical webhooks in retry queue
-- [ ] Dead letter queue reviewed and exported
--
-- Impact:
-- - Webhook retry worker will not function without these fields
-- - Historical retry information will be lost
-- - Dead letter queue reasons will be deleted
-- - retry_count column remains (was already present before migration 005)
--
-- Safety: Wrapped in transaction, auto-rollback on error
-- ============================================================================

BEGIN;

-- Safety check: Verify migration is applied
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        RAISE NOTICE 'schema_migrations table does not exist, skipping version check';
    ELSIF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '005') THEN
        RAISE EXCEPTION 'Migration 005 is not applied, cannot rollback';
    END IF;
END $$;

-- Dependency check: Ensure polar_webhook_events table exists
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'polar_webhook_events') THEN
        RAISE EXCEPTION 'Table polar_webhook_events does not exist. Cannot rollback migration 005.';
    END IF;
END $$;

-- Display data loss warning
DO $$
DECLARE
    retry_count INTEGER := 0;
    dead_letter_count INTEGER := 0;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'polar_webhook_events' AND column_name = 'last_retry_at') THEN
        SELECT COUNT(*) INTO retry_count
        FROM polar_webhook_events
        WHERE last_retry_at IS NOT NULL;
    END IF;

    IF EXISTS (SELECT 1 FROM information_schema.columns
               WHERE table_name = 'polar_webhook_events' AND column_name = 'dead_letter_reason') THEN
        SELECT COUNT(*) INTO dead_letter_count
        FROM polar_webhook_events
        WHERE dead_letter_reason IS NOT NULL;
    END IF;

    IF retry_count > 0 OR dead_letter_count > 0 THEN
        RAISE WARNING 'About to delete:';
        RAISE WARNING '  - Retry timestamps for % webhook events', retry_count;
        RAISE WARNING '  - Dead letter reasons for % webhook events', dead_letter_count;
    END IF;
END $$;

-- Step 1: Drop indexes that reference the columns we're about to drop
DROP INDEX IF EXISTS idx_webhook_retry;
DROP INDEX IF EXISTS idx_polar_webhook_events_last_retry_at;

-- Step 2: Recreate the original processing_status index that was dropped in the up migration
-- Note: Migration 005 up script dropped idx_polar_webhook_events_processing_status
-- and replaced it with idx_webhook_retry (composite index)
CREATE INDEX IF NOT EXISTS idx_polar_webhook_events_processing_status
ON polar_webhook_events(processing_status);

-- Step 3: Drop the columns added by migration 005
ALTER TABLE polar_webhook_events
DROP COLUMN IF EXISTS last_retry_at,
DROP COLUMN IF EXISTS dead_letter_reason;

-- Step 4: Update column comments back to original
COMMENT ON COLUMN polar_webhook_events.retry_count IS 'Number of retry attempts';
COMMENT ON COLUMN polar_webhook_events.processing_status IS 'Status: pending, processed, failed';

-- Remove migration from tracking table
DELETE FROM schema_migrations WHERE version = '005';

-- Log successful rollback
RAISE NOTICE 'Migration 005 rolled back successfully';
RAISE NOTICE 'Dropped columns: last_retry_at, dead_letter_reason';
RAISE NOTICE 'Recreated index: idx_polar_webhook_events_processing_status';

COMMIT;
