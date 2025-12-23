-- Migration Rollback: Remove added columns from user_sessions table
-- Description: Removes device_info, refresh_token_hash, expires_at, created_at columns
-- Author: Claude Code
-- Date: 2025-12-23

-- Drop indexes
DROP INDEX IF EXISTS idx_user_sessions_refresh_token_hash;
DROP INDEX IF EXISTS idx_user_sessions_expires_at;

-- Drop columns
ALTER TABLE user_sessions DROP COLUMN IF EXISTS created_at;
ALTER TABLE user_sessions DROP COLUMN IF EXISTS expires_at;
ALTER TABLE user_sessions DROP COLUMN IF EXISTS refresh_token_hash;
ALTER TABLE user_sessions DROP COLUMN IF EXISTS device_info;
