-- Migration 034: Assign Super Administrator role to user 112125335771807745
-- Purpose: Bootstrap super_admin for the currently logged-in user
-- Context: User 745 needs super_admin to assign roles to other users

-- Assign super_admin role to user ID 112125335771807745
INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
SELECT
    112125335771807745,                    -- User ID
    r.id,                                  -- Super Admin role ID
    NULL,                                  -- assigned_by NULL indicates bootstrap assignment
    NOW()
FROM roles r
WHERE r.name = 'super_admin'
  AND EXISTS (SELECT 1 FROM users WHERE id = 112125335771807745)
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Verify the assignment
DO $$
DECLARE
    role_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO role_count
    FROM user_roles ur
    JOIN roles r ON ur.role_id = r.id
    WHERE ur.user_id = 112125335771807745
      AND r.name = 'super_admin';

    IF role_count = 0 THEN
        RAISE WARNING 'Super admin role not assigned. Check if user 112125335771807745 exists.';
    ELSE
        RAISE NOTICE 'Successfully assigned super_admin role to user 112125335771807745';
    END IF;
END $$;
