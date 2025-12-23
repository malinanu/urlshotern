-- Migration Rollback: Remove name column from users table
-- Description: Removes the name column added in the up migration
-- Author: Claude Code
-- Date: 2025-12-23

-- Drop the index
DROP INDEX IF EXISTS idx_users_name;

-- Drop the name column
ALTER TABLE users DROP COLUMN IF EXISTS name;
