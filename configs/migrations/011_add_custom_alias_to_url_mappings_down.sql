-- Migration Rollback: Remove custom_alias column
-- Author: Claude Code
-- Date: 2025-12-23

-- Drop index
DROP INDEX IF EXISTS idx_url_mappings_custom_alias;

-- Drop column
ALTER TABLE url_mappings DROP COLUMN IF EXISTS custom_alias;
