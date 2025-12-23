-- Migration: Add missing columns to user_sessions table
-- Description: Adds authentication-related columns (device_info, refresh_token_hash, expires_at, created_at)
-- The table currently has analytics columns but is missing auth session columns
-- Author: Claude Code
-- Date: 2025-12-23

-- Add device_info as JSONB (can store device details)
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS device_info JSONB;

-- Add refresh_token_hash for JWT refresh tokens
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS refresh_token_hash VARCHAR(255);

-- Add expires_at for session expiration
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS expires_at TIMESTAMP WITH TIME ZONE;

-- Add created_at if it doesn't exist
ALTER TABLE user_sessions ADD COLUMN IF NOT EXISTS created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Create indexes for better performance
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_refresh_token_hash ON user_sessions(refresh_token_hash);

-- Add comments for documentation
COMMENT ON COLUMN user_sessions.device_info IS 'JSON object containing device information';
COMMENT ON COLUMN user_sessions.refresh_token_hash IS 'Hashed refresh token for session renewal';
COMMENT ON COLUMN user_sessions.expires_at IS 'Session expiration timestamp';
