-- Migration Rollback: Fix API Keys Schema
-- This migration reverts the api_keys schema changes

-- Drop the api_key_usage table
DROP TABLE IF EXISTS api_key_usage;

-- Remove added indexes
DROP INDEX IF EXISTS idx_api_keys_key_hash;
DROP INDEX IF EXISTS idx_api_keys_is_active;
DROP INDEX IF EXISTS idx_api_keys_expires_at;
DROP INDEX IF EXISTS idx_api_keys_created_at;

-- Remove UNIQUE constraint from key_hash
ALTER TABLE api_keys DROP CONSTRAINT IF EXISTS api_keys_key_hash_key;

-- Remove added columns
ALTER TABLE api_keys
  DROP COLUMN IF EXISTS key_prefix,
  DROP COLUMN IF EXISTS description,
  DROP COLUMN IF EXISTS rate_limit,
  DROP COLUMN IF EXISTS last_used_ip,
  DROP COLUMN IF EXISTS updated_at;

-- Revert permissions default (though this won't affect existing data)
ALTER TABLE api_keys ALTER COLUMN permissions SET DEFAULT '{}';
