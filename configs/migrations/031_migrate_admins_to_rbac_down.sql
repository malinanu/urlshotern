-- Rollback Migration 031: Remove RBAC role assignments created during migration

-- Remove admin role assignments that were auto-assigned (where assigned_by is NULL)
DELETE FROM user_roles
WHERE assigned_by IS NULL
  AND role_id IN (SELECT id FROM roles WHERE name = 'admin');

-- Remove user role assignments that were auto-assigned (where assigned_by is NULL)
DELETE FROM user_roles
WHERE assigned_by IS NULL
  AND role_id IN (SELECT id FROM roles WHERE name = 'user');

RAISE NOTICE 'Migration 031 rolled back successfully';
