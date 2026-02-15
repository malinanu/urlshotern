-- Migration 034 Rollback: Remove Super Administrator role from user 112125335771807745

-- Remove super_admin role from user ID 112125335771807745
DELETE FROM user_roles ur
USING roles r
WHERE ur.role_id = r.id
  AND ur.user_id = 112125335771807745
  AND r.name = 'super_admin'
  AND ur.assigned_by IS NULL;  -- Only remove bootstrap assignments

-- Verify the removal
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
        RAISE NOTICE 'Successfully removed super_admin role from user 112125335771807745';
    ELSE
        RAISE WARNING 'Super admin role still assigned (may have been manually reassigned)';
    END IF;
END $$;
