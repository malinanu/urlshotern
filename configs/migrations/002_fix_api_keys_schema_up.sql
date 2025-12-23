-- Migration: Fix API Keys Schema
-- This migration updates the api_keys table to add missing columns and creates the api_key_usage table
-- Run this if you have an existing database with the old schema

-- Add missing columns to api_keys table
ALTER TABLE api_keys
  ADD COLUMN IF NOT EXISTS key_prefix VARCHAR(100),
  ADD COLUMN IF NOT EXISTS description TEXT,
  ADD COLUMN IF NOT EXISTS rate_limit INTEGER DEFAULT 1000,
  ADD COLUMN IF NOT EXISTS last_used_ip INET,
  ADD COLUMN IF NOT EXISTS updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW();

-- Update permissions default value for existing rows (convert {} to [])
UPDATE api_keys
SET permissions = '[]'::jsonb
WHERE permissions = '{}'::jsonb;

-- Alter column default for future insertions
ALTER TABLE api_keys ALTER COLUMN permissions SET DEFAULT '[]';

-- Add UNIQUE constraint to key_hash if it doesn't exist
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM pg_constraint
        WHERE conname = 'api_keys_key_hash_key'
    ) THEN
        ALTER TABLE api_keys ADD CONSTRAINT api_keys_key_hash_key UNIQUE (key_hash);
    END IF;
END $$;

-- Update existing rows to have non-null key_prefix if any exist
-- Generate a random prefix for existing keys
UPDATE api_keys
SET key_prefix = 'sk_' || substring(md5(random()::text) from 1 for 12)
WHERE key_prefix IS NULL;

-- Make key_prefix NOT NULL after populating existing rows
ALTER TABLE api_keys ALTER COLUMN key_prefix SET NOT NULL;

-- Create API Key Usage table if it doesn't exist
CREATE TABLE IF NOT EXISTS api_key_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    request_size INTEGER,
    response_size INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- Create additional indexes for performance
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_created_at ON api_keys(created_at);

CREATE INDEX IF NOT EXISTS idx_api_key_usage_api_key_id ON api_key_usage(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_created_at ON api_key_usage(created_at);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_endpoint ON api_key_usage(endpoint);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_status_code ON api_key_usage(status_code);
