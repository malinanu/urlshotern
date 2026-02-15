-- Migration 033: Assign initial Super Administrator role
-- Purpose: Bootstrap the first super_admin user to break the chicken-and-egg problem
-- Context: Only super_admins can assign super_admin roles, so we need to create the first one via migration

-- Assign super_admin role to user ID 112125335771807740
-- This is idempotent - won't fail if already assigned
INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
SELECT
    112125335771807740,                    -- User ID
    r.id,                                  -- Super Admin role ID (from roles table)
    NULL,                                  -- assigned_by NULL indicates bootstrap/system assignment
    NOW()                                  -- Current timestamp
FROM roles r
WHERE r.name = 'super_admin'
  AND EXISTS (SELECT 1 FROM users WHERE id = 112125335771807740)  -- Only insert if user exists
ON CONFLICT (user_id, role_id) DO NOTHING;  -- Idempotent: skip if already assigned

-- Verify the assignment was successful
DO $$
DECLARE
    role_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO role_count
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = 112125335771807740
      AND r.name = 'super_admin';

    IF role_count = 0 THEN
        RAISE WARNING 'Super admin role not assigned. Check if user 112125335771807740 exists and super_admin role is created.';
    ELSE
        RAISE NOTICE 'Successfully assigned super_admin role to user 112125335771807740';
    END IF;
END $$;
