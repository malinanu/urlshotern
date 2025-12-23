-- Migration: Fix users table column names for consistency
-- Description: Renames columns to match application expectations
-- The application code expects 'email_verified' and 'phone_verified'
-- but the database has 'is_email_verified' and 'is_phone_verified'
-- Author: Claude Code
-- Date: 2025-12-23

-- Rename is_email_verified to email_verified
ALTER TABLE users RENAME COLUMN is_email_verified TO email_verified;

-- Rename is_phone_verified to phone_verified
ALTER TABLE users RENAME COLUMN is_phone_verified TO phone_verified;

-- Add comments for documentation
COMMENT ON COLUMN users.email_verified IS 'Whether the user email address has been verified';
COMMENT ON COLUMN users.phone_verified IS 'Whether the user phone number has been verified';
