-- Migration: Add name column to users table
-- Description: Adds a consolidated 'name' column to users table for display purposes
-- This addresses the compatibility issue where the application expects a 'name' field
-- Author: Claude Code
-- Date: 2025-12-23

-- Add the name column (nullable initially)
ALTER TABLE users ADD COLUMN IF NOT EXISTS name VARCHAR(100);

-- Populate name from first_name and last_name
UPDATE users
SET name = CONCAT(COALESCE(first_name, ''), ' ', COALESCE(last_name, ''))
WHERE name IS NULL AND (first_name IS NOT NULL OR last_name IS NOT NULL);

-- Set name to email username for users without first/last name
UPDATE users
SET name = SPLIT_PART(email, '@', 1)
WHERE name IS NULL OR TRIM(name) = '';

-- Make the column NOT NULL after populating data
ALTER TABLE users ALTER COLUMN name SET NOT NULL;

-- Add index for name column for performance
CREATE INDEX IF NOT EXISTS idx_users_name ON users(name);

-- Add comment for documentation
COMMENT ON COLUMN users.name IS 'User display name (consolidated from first_name and last_name or email)';
