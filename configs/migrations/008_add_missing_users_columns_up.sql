-- Migration: Add missing columns to users table
-- Description: Adds OAuth provider columns, account_type, and is_admin flag
-- These columns are required by the application but were missing from the initial schema
-- Author: Claude Code
-- Date: 2025-12-23

-- Add provider column (for OAuth providers like 'email', 'google', etc.)
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider VARCHAR(50) DEFAULT 'email';

-- Add provider_id column (OAuth provider ID)
ALTER TABLE users ADD COLUMN IF NOT EXISTS provider_id VARCHAR(255);

-- Add account_type column (free, premium, enterprise)
ALTER TABLE users ADD COLUMN IF NOT EXISTS account_type VARCHAR(20) DEFAULT 'free';

-- Add is_admin flag
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_admin BOOLEAN DEFAULT FALSE;

-- Create indexes for better query performance
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider);
CREATE INDEX IF NOT EXISTS idx_users_account_type ON users(account_type);
CREATE INDEX IF NOT EXISTS idx_users_is_admin ON users(is_admin);

-- Add comments for documentation
COMMENT ON COLUMN users.provider IS 'Authentication provider: email, google, github, etc.';
COMMENT ON COLUMN users.provider_id IS 'Unique ID from OAuth provider';
COMMENT ON COLUMN users.account_type IS 'User account tier: free, premium, enterprise';
COMMENT ON COLUMN users.is_admin IS 'Whether the user has administrator privileges';
