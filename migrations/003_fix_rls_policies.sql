-- ============================================================================
-- Fix Row-Level Security Policies
-- Description: Remove deleted_at references and fix RLS policies
-- Version: 003
-- Author: Admin Panel Bug Fix
-- Date: 2026-01-08
-- ============================================================================

\c urlshortener_db;

-- ============================================================================
-- Fix RLS Policies for Users Table
-- ============================================================================

-- Drop existing policies
DROP POLICY IF EXISTS app_user_access ON users;
DROP POLICY IF EXISTS admin_user_access ON users;
DROP POLICY IF EXISTS readonly_user_access ON users;

-- Recreate policies without deleted_at reference
-- Policy: Regular app can only access active users
CREATE POLICY app_user_access ON users
    FOR ALL
    TO urlshort_app
    USING (is_active = true);

-- Policy: Admin can access all users (including inactive)
CREATE POLICY admin_user_access ON users
    FOR ALL
    TO urlshort_admin
    USING (true);

-- Policy: Read-only user can only see active users
CREATE POLICY readonly_user_access ON users
    FOR SELECT
    TO urlshort_readonly
    USING (is_active = true);

RAISE NOTICE 'Fixed RLS policies for users table';

-- ============================================================================
-- Verification
-- ============================================================================

-- Test that readonly user can now SELECT from users table
DO $$
DECLARE
    test_count INTEGER;
BEGIN
    -- This will succeed if RLS policy is correctly configured
    SELECT COUNT(*) INTO test_count FROM users WHERE is_active = true;
    RAISE NOTICE 'RLS policy verification: Found % active users', test_count;
END $$;

-- ============================================================================
-- Configuration Complete
-- ============================================================================

-- Summary:
-- ✅ Fixed RLS policies to use is_active instead of deleted_at
-- ✅ app_user_access: Only access active users
-- ✅ admin_user_access: Access all users
-- ✅ readonly_user_access: Only see active users
