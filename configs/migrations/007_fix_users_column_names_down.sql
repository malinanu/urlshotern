-- Migration Rollback: Revert users table column names
-- Description: Reverts column name changes from the up migration
-- Author: Claude Code
-- Date: 2025-12-23

-- Revert email_verified to is_email_verified
ALTER TABLE users RENAME COLUMN email_verified TO is_email_verified;

-- Revert phone_verified to is_phone_verified
ALTER TABLE users RENAME COLUMN phone_verified TO is_phone_verified;
