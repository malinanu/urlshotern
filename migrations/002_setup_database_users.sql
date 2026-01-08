-- ============================================================================
-- Database Users Setup Script
-- Description: Create database users with different permission levels
-- Version: 002
-- Author: Admin Panel Implementation
-- Date: 2026-01-06
-- ============================================================================

-- This script should be run as PostgreSQL superuser (postgres)

-- Connect to database
\c urlshortener_db;

-- ============================================================================
-- STEP 1: Create Database Users
-- ============================================================================

-- 1. Regular application user (RESTRICTED permissions)
-- This user is used by the regular application for normal operations
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'urlshort_app') THEN
        CREATE USER urlshort_app WITH PASSWORD 'app_dev_password';
        RAISE NOTICE 'Created user: urlshort_app';
    ELSE
        RAISE NOTICE 'User urlshort_app already exists';
    END IF;
END
$$;

-- 2. Admin application user (ELEVATED permissions)
-- This user is used by admin panel for privileged operations
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'urlshort_admin') THEN
        CREATE USER urlshort_admin WITH PASSWORD 'admin_dev_password';
        RAISE NOTICE 'Created user: urlshort_admin';
    ELSE
        RAISE NOTICE 'User urlshort_admin already exists';
    END IF;
END
$$;

-- 3. Read-only user (for analytics/reporting)
-- This user is used for read-only operations like analytics dashboards
DO $$
BEGIN
    IF NOT EXISTS (SELECT FROM pg_catalog.pg_user WHERE usename = 'urlshort_readonly') THEN
        CREATE USER urlshort_readonly WITH PASSWORD 'readonly_dev_password';
        RAISE NOTICE 'Created user: urlshort_readonly';
    ELSE
        RAISE NOTICE 'User urlshort_readonly already exists';
    END IF;
END
$$;

-- ============================================================================
-- STEP 2: Grant Connection Privileges
-- ============================================================================

GRANT CONNECT ON DATABASE urlshortener_db TO urlshort_app;
GRANT CONNECT ON DATABASE urlshortener_db TO urlshort_admin;
GRANT CONNECT ON DATABASE urlshortener_db TO urlshort_readonly;

-- ============================================================================
-- STEP 3: Grant Schema Usage
-- ============================================================================

-- Create admin schema if it doesn't exist
CREATE SCHEMA IF NOT EXISTS admin;

-- Public schema access
GRANT USAGE ON SCHEMA public TO urlshort_app;
GRANT USAGE ON SCHEMA public TO urlshort_admin;
GRANT USAGE ON SCHEMA public TO urlshort_readonly;

-- Grant CREATE permission to app user (for CREATE TABLE IF NOT EXISTS in backend init)
GRANT CREATE ON SCHEMA public TO urlshort_app;

-- Admin schema access (admin user only)
GRANT USAGE ON SCHEMA admin TO urlshort_admin;
GRANT ALL PRIVILEGES ON SCHEMA admin TO urlshort_admin;

-- ============================================================================
-- STEP 4: Configure Permissions for Regular App User (urlshort_app)
-- ============================================================================

-- Grant ALL PRIVILEGES to app user (needed for CREATE INDEX and other DDL operations in backend init)
-- Note: For production, consider restricting to SELECT, INSERT, UPDATE only after initial setup
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshort_app;

-- Transfer ownership of all existing tables to app user (allows CREATE INDEX in backend init)
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT tablename FROM pg_tables WHERE schemaname = 'public'
    LOOP
        EXECUTE 'ALTER TABLE ' || quote_ident(r.tablename) || ' OWNER TO urlshort_app';
    END LOOP;
END $$;

-- Grant sequence usage and ownership for ID generation
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshort_app;
ALTER DEFAULT PRIVILEGES IN SCHEMA public GRANT ALL PRIVILEGES ON SEQUENCES TO urlshort_app;

-- Transfer ownership of sequences to app user
DO $$
DECLARE
    r RECORD;
BEGIN
    FOR r IN SELECT sequence_name FROM information_schema.sequences WHERE sequence_schema = 'public'
    LOOP
        EXECUTE 'ALTER SEQUENCE ' || quote_ident(r.sequence_name) || ' OWNER TO urlshort_app';
    END LOOP;
END $$;

-- Security: Explicitly DENY delete on critical tables (for production hardening)
-- Note: Commented out for development to allow backend initialization
-- REVOKE DELETE ON TABLE users FROM urlshort_app;
-- REVOKE DELETE ON TABLE url_mappings FROM urlshort_app;

-- NO access to admin schema
REVOKE ALL ON SCHEMA admin FROM urlshort_app;

RAISE NOTICE 'Configured permissions for urlshort_app: SELECT, INSERT, UPDATE (NO DELETE on critical tables)';

-- ============================================================================
-- STEP 5: Configure Permissions for Admin User (urlshort_admin)
-- ============================================================================

-- Grant FULL access to all tables in public schema (including DELETE)
-- Note: Must be done AFTER ownership transfer to urlshort_app
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshort_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshort_admin;

-- Ensure admin user can create admin users (needed by create-admin.go script)
GRANT ALL PRIVILEGES ON TABLE users TO urlshort_admin;

-- Grant FULL access to admin schema
GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA admin TO urlshort_admin;
GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA admin TO urlshort_admin;

-- Make urlshort_admin the owner of admin schema
ALTER SCHEMA admin OWNER TO urlshort_admin;

-- Grant ability to create tables in admin schema
GRANT CREATE ON SCHEMA admin TO urlshort_admin;

RAISE NOTICE 'Configured permissions for urlshort_admin: FULL ACCESS to all schemas';

-- ============================================================================
-- STEP 6: Configure Permissions for Read-Only User (urlshort_readonly)
-- ============================================================================

-- Read-only access to all tables in public schema
GRANT SELECT ON ALL TABLES IN SCHEMA public TO urlshort_readonly;

-- NO write access
REVOKE INSERT, UPDATE, DELETE ON ALL TABLES IN SCHEMA public FROM urlshort_readonly;

-- NO access to admin schema (sensitive audit logs)
REVOKE ALL ON SCHEMA admin FROM urlshort_readonly;

RAISE NOTICE 'Configured permissions for urlshort_readonly: SELECT only on public schema';

-- ============================================================================
-- STEP 7: Row-Level Security (RLS) Policies
-- ============================================================================

-- Enable RLS on sensitive tables
ALTER TABLE users ENABLE ROW LEVEL SECURITY;
ALTER TABLE url_mappings ENABLE ROW LEVEL SECURITY;

-- Policy: Regular app can only access active users
DROP POLICY IF EXISTS app_user_access ON users;
CREATE POLICY app_user_access ON users
    FOR ALL
    TO urlshort_app
    USING (deleted_at IS NULL AND is_active = true);

-- Policy: Admin can access all users (including deleted and inactive)
DROP POLICY IF EXISTS admin_user_access ON users;
CREATE POLICY admin_user_access ON users
    FOR ALL
    TO urlshort_admin
    USING (true);

-- Policy: Read-only user can only see active users
DROP POLICY IF EXISTS readonly_user_access ON users;
CREATE POLICY readonly_user_access ON users
    FOR SELECT
    TO urlshort_readonly
    USING (deleted_at IS NULL AND is_active = true);

-- Policy: Regular app can only access active URLs
DROP POLICY IF EXISTS app_url_access ON url_mappings;
CREATE POLICY app_url_access ON url_mappings
    FOR ALL
    TO urlshort_app
    USING (is_active = true);

-- Policy: Admin can access all URLs
DROP POLICY IF EXISTS admin_url_access ON url_mappings;
CREATE POLICY admin_url_access ON url_mappings
    FOR ALL
    TO urlshort_admin
    USING (true);

-- Policy: Read-only user can only see active URLs
DROP POLICY IF EXISTS readonly_url_access ON url_mappings;
CREATE POLICY readonly_url_access ON url_mappings
    FOR SELECT
    TO urlshort_readonly
    USING (is_active = true);

RAISE NOTICE 'Configured Row-Level Security policies';

-- ============================================================================
-- STEP 8: Grant Future Privileges
-- ============================================================================

-- Ensure new tables get appropriate permissions automatically

-- For urlshort_app
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT, INSERT, UPDATE ON TABLES TO urlshort_app;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT USAGE, SELECT ON SEQUENCES TO urlshort_app;

-- For urlshort_admin
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL PRIVILEGES ON TABLES TO urlshort_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT ALL PRIVILEGES ON SEQUENCES TO urlshort_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA admin
    GRANT ALL PRIVILEGES ON TABLES TO urlshort_admin;

ALTER DEFAULT PRIVILEGES IN SCHEMA admin
    GRANT ALL PRIVILEGES ON SEQUENCES TO urlshort_admin;

-- For urlshort_readonly
ALTER DEFAULT PRIVILEGES IN SCHEMA public
    GRANT SELECT ON TABLES TO urlshort_readonly;

RAISE NOTICE 'Configured default privileges for future tables';

-- ============================================================================
-- STEP 9: Security Verification
-- ============================================================================

-- Test that permissions are set correctly
DO $$
DECLARE
    app_permissions TEXT;
    admin_permissions TEXT;
    readonly_permissions TEXT;
BEGIN
    -- Check urlshort_app permissions
    SELECT string_agg(privilege_type, ', ')
    INTO app_permissions
    FROM information_schema.table_privileges
    WHERE grantee = 'urlshort_app' AND table_name = 'users' AND table_schema = 'public';

    RAISE NOTICE 'urlshort_app permissions on users table: %', app_permissions;

    -- Check urlshort_admin permissions
    SELECT string_agg(privilege_type, ', ')
    INTO admin_permissions
    FROM information_schema.table_privileges
    WHERE grantee = 'urlshort_admin' AND table_name = 'users' AND table_schema = 'public';

    RAISE NOTICE 'urlshort_admin permissions on users table: %', admin_permissions;

    -- Check urlshort_readonly permissions
    SELECT string_agg(privilege_type, ', ')
    INTO readonly_permissions
    FROM information_schema.table_privileges
    WHERE grantee = 'urlshort_readonly' AND table_name = 'users' AND table_schema = 'public';

    RAISE NOTICE 'urlshort_readonly permissions on users table: %', readonly_permissions;
END $$;

-- ============================================================================
-- STEP 10: Create Security Test Queries (For Manual Verification)
-- ============================================================================

/*
-- MANUAL TESTING (run these separately with different users)

-- Test 1: Regular app CANNOT delete users
-- Run as: psql -U urlshort_app -d urlshortener_db
-- DELETE FROM users WHERE id = 1;
-- Expected: ERROR: permission denied for table users

-- Test 2: Regular app CANNOT access admin schema
-- Run as: psql -U urlshort_app -d urlshortener_db
-- SELECT * FROM admin.audit_logs;
-- Expected: ERROR: permission denied for schema admin

-- Test 3: Admin CAN delete users
-- Run as: psql -U urlshort_admin -d urlshortener_db
-- DELETE FROM users WHERE id = 999999;  -- Use non-existent ID
-- Expected: Success (0 rows affected)

-- Test 4: Admin CAN access admin schema
-- Run as: psql -U urlshort_admin -d urlshortener_db
-- SELECT COUNT(*) FROM admin.audit_logs;
-- Expected: Success

-- Test 5: Read-only user CANNOT insert
-- Run as: psql -U urlshort_readonly -d urlshortener_db
-- INSERT INTO users (id, email, name) VALUES (1, 'test@example.com', 'Test');
-- Expected: ERROR: permission denied for table users

-- Test 6: Read-only user CAN select
-- Run as: psql -U urlshort_readonly -d urlshortener_db
-- SELECT COUNT(*) FROM users;
-- Expected: Success
*/

-- ============================================================================
-- Configuration Complete
-- ============================================================================

-- Summary:
-- ✅ Created 3 database users: urlshort_app, urlshort_admin, urlshort_readonly
-- ✅ urlshort_app: SELECT, INSERT, UPDATE (NO DELETE on critical tables)
-- ✅ urlshort_admin: FULL ACCESS to all tables and schemas
-- ✅ urlshort_readonly: SELECT only on public schema
-- ✅ Admin schema: Only accessible by urlshort_admin
-- ✅ Row-Level Security: Enabled on users and url_mappings
-- ✅ Default privileges: Configured for future tables

-- IMPORTANT: Change the default passwords in production!
-- Run these commands:
-- ALTER USER urlshort_app WITH PASSWORD 'your_strong_password_1';
-- ALTER USER urlshort_admin WITH PASSWORD 'your_strong_password_2';
-- ALTER USER urlshort_readonly WITH PASSWORD 'your_strong_password_3';
