-- ============================================================================
-- ROLLBACK: Add API Keys Support (Migration 001)
-- ============================================================================
-- Description: Removes api_keys and api_key_usage tables
-- Data Loss: ALL API keys and usage history will be deleted
-- Dependencies: None
-- Estimated Time: <1 second
-- Reversible: Yes (re-apply migration 001_add_api_keys_up.sql)
--
-- Pre-rollback Checklist:
-- [ ] Database backup created (pg_dump)
-- [ ] No critical API keys in production
-- [ ] Users notified about API key regeneration requirement
-- [ ] Application code updated to handle missing API keys gracefully
--
-- Safety: Wrapped in transaction, auto-rollback on error
-- ============================================================================

BEGIN;

-- Safety check: Verify migration is applied before rolling back
DO $$
BEGIN
    IF NOT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations') THEN
        RAISE NOTICE 'schema_migrations table does not exist, skipping version check';
    ELSIF NOT EXISTS (SELECT 1 FROM schema_migrations WHERE version = '001') THEN
        RAISE EXCEPTION 'Migration 001 is not applied, cannot rollback';
    END IF;
END $$;

-- Display data loss warning
DO $$
DECLARE
    api_key_count INTEGER;
    usage_count INTEGER;
BEGIN
    IF EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'api_keys') THEN
        SELECT COUNT(*) INTO api_key_count FROM api_keys;
        SELECT COUNT(*) INTO usage_count FROM api_key_usage;
        RAISE WARNING 'About to delete % API keys and % usage records', api_key_count, usage_count;
    END IF;
END $$;

-- Drop dependent table first (api_key_usage references api_keys)
DROP TABLE IF EXISTS api_key_usage CASCADE;

-- Drop main table
DROP TABLE IF EXISTS api_keys CASCADE;

-- Note: All indexes on these tables are automatically dropped with the tables

-- Remove migration from tracking table
DELETE FROM schema_migrations WHERE version = '001';

-- Log successful rollback
RAISE NOTICE 'Migration 001 rolled back successfully';

COMMIT;
