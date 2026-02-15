-- Migration: Migrate Existing Admins to RBAC
-- Category: RBAC Data Migration
-- Tables: user_roles
-- Dependencies: 030_add_cms_permissions

-- ============================================================================
-- Assign 'admin' role to users where is_admin=true
-- ============================================================================

INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
SELECT u.id, r.id, NULL, NOW()
FROM users u
CROSS JOIN roles r
WHERE u.is_admin = true AND r.name = 'admin'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur2
    WHERE ur2.user_id = u.id AND ur2.role_id = r.id
  )
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================================================
-- Assign 'user' role to users without any role
-- ============================================================================

INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
SELECT u.id, r.id, NULL, NOW()
FROM users u
CROSS JOIN roles r
WHERE r.name = 'user'
  AND NOT EXISTS (
    SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id
  )
ON CONFLICT (user_id, role_id) DO NOTHING;

-- ============================================================================
-- VERIFICATION: Ensure admins were migrated
-- ============================================================================

DO $$
DECLARE
    admin_user_count INTEGER;
    admin_role_count INTEGER;
    users_without_roles INTEGER;
BEGIN
    -- Count users with is_admin=true
    SELECT COUNT(*) INTO admin_user_count FROM users WHERE is_admin = true;

    -- Count users with admin role
    SELECT COUNT(DISTINCT ur.user_id) INTO admin_role_count
    FROM user_roles ur
    JOIN roles r ON r.id = ur.role_id
    WHERE r.name = 'admin';

    -- Count users without any role
    SELECT COUNT(*) INTO users_without_roles
    FROM users u
    WHERE NOT EXISTS (SELECT 1 FROM user_roles ur WHERE ur.user_id = u.id);

    RAISE NOTICE 'Migration 031: % users have is_admin=true', admin_user_count;
    RAISE NOTICE 'Migration 031: % users have admin role', admin_role_count;
    RAISE NOTICE 'Migration 031: % users have no roles', users_without_roles;

    IF admin_user_count > 0 AND admin_role_count = 0 THEN
        RAISE EXCEPTION 'Migration 031 Failed: Users with is_admin=true exist but no admin roles assigned';
    END IF;

    IF users_without_roles > 0 THEN
        RAISE WARNING 'Migration 031: % users still have no roles assigned', users_without_roles;
    END IF;

    RAISE NOTICE 'Migration 031 completed successfully ✓';
END $$;
