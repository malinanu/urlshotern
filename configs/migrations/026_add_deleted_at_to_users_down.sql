-- Rollback: Remove deleted_at column from users table
-- Description: Removes soft delete support for users
-- Author: Claude Code
-- Date: 2026-02-13

-- Drop index
DROP INDEX IF EXISTS idx_users_deleted_at;

-- Remove deleted_at column
ALTER TABLE users DROP COLUMN IF EXISTS deleted_at;
