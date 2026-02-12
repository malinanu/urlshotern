-- Migration: Add deleted_at column to users table
-- Description: Adds soft delete support for users
-- Author: Claude Code
-- Date: 2026-02-13

-- Add deleted_at column for soft deletes
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Create index for better query performance on non-deleted users
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at) WHERE deleted_at IS NULL;

-- Add comment for documentation
COMMENT ON COLUMN users.deleted_at IS 'Soft delete timestamp - NULL means user is not deleted';
