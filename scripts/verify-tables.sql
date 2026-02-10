-- Verify Database Tables
-- Run this to check which tables exist in your database

-- Count total tables
SELECT
    'Total Tables' AS metric,
    COUNT(*)::text AS value
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE';

-- List all tables
SELECT
    table_name,
    (SELECT COUNT(*) FROM information_schema.columns WHERE columns.table_name = tables.table_name) AS column_count
FROM information_schema.tables
WHERE table_schema = 'public'
  AND table_type = 'BASE TABLE'
ORDER BY table_name;

-- Check for critical RBAC tables
SELECT
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS roles_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS permissions_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'user_roles') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS user_roles_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'role_permissions') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS role_permissions_table;

-- Check for team tables
SELECT
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'teams') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS teams_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'team_members') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS team_members_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'team_invitations') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS team_invitations_table;

-- Check for custom domain tables
SELECT
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'custom_domains') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS custom_domains_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'domain_analytics') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS domain_analytics_table;

-- Check for collaboration tables
SELECT
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'url_shares') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS url_shares_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'url_comments') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS url_comments_table,
    CASE
        WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'url_collections') THEN '✓ EXISTS'
        ELSE '✗ MISSING'
    END AS url_collections_table;

-- Check url_mappings for new columns
SELECT
    column_name,
    data_type,
    is_nullable,
    column_default
FROM information_schema.columns
WHERE table_name = 'url_mappings'
ORDER BY ordinal_position;

-- Check if roles have been populated
SELECT
    'Roles Count' AS metric,
    COUNT(*)::text AS value
FROM roles
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles');

-- Check if permissions have been populated
SELECT
    'Permissions Count' AS metric,
    COUNT(*)::text AS value
FROM permissions
WHERE EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions');
