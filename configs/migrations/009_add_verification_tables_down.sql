-- Migration Rollback: Remove verification and password reset tables
-- Description: Drops email_verifications, phone_verifications, and password_resets tables
-- Author: Claude Code
-- Date: 2025-12-23

-- Drop indexes
DROP INDEX IF EXISTS idx_password_resets_token;
DROP INDEX IF EXISTS idx_password_resets_user_id;
DROP INDEX IF EXISTS idx_phone_verifications_user_id;
DROP INDEX IF EXISTS idx_email_verifications_token;
DROP INDEX IF EXISTS idx_email_verifications_user_id;

-- Drop tables
DROP TABLE IF EXISTS password_resets;
DROP TABLE IF EXISTS phone_verifications;
DROP TABLE IF EXISTS email_verifications;
