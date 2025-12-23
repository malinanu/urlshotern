-- Migration: Add custom_alias column to url_mappings
-- Description: Adds custom_alias column for user-defined short codes
-- Author: Claude Code
-- Date: 2025-12-23

-- Add custom_alias column
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS custom_alias VARCHAR(50);

-- Create index for custom_alias
CREATE INDEX IF NOT EXISTS idx_url_mappings_custom_alias ON url_mappings(custom_alias);

-- Add comment
COMMENT ON COLUMN url_mappings.custom_alias IS 'User-defined custom short code alias';
