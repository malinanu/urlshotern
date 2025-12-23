-- Migration Rollback: Remove added columns from users table
-- Description: Removes provider, provider_id, account_type, and is_admin columns
-- Author: Claude Code
-- Date: 2025-12-23

-- Drop indexes
DROP INDEX IF EXISTS idx_users_is_admin;
DROP INDEX IF EXISTS idx_users_account_type;
DROP INDEX IF EXISTS idx_users_provider;

-- Drop columns
ALTER TABLE users DROP COLUMN IF EXISTS is_admin;
ALTER TABLE users DROP COLUMN IF EXISTS account_type;
ALTER TABLE users DROP COLUMN IF EXISTS provider_id;
ALTER TABLE users DROP COLUMN IF EXISTS provider;
