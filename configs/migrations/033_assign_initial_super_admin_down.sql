-- Migration 033 Rollback: Remove initial Super Administrator role assignment

-- Remove super_admin role from user ID 112125335771807740
DELETE FROM user_roles ur
USING roles r
WHERE ur.role_id = r.id
  AND ur.user_id = 112125335771807740
  AND r.name = 'super_admin'
  AND ur.assigned_by IS NULL;  -- Only remove bootstrap assignments (assigned_by = NULL)

-- Verify the removal
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
        RAISE NOTICE 'Successfully removed super_admin role from user 112125335771807740';
    ELSE
        RAISE WARNING 'Super admin role still assigned (may have been manually reassigned with assigned_by value)';
    END IF;
END $$;
