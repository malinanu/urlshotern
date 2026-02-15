-- Rollback Migration 030: Remove CMS Permissions and Content Roles

-- Remove role-permission assignments for new roles
DELETE FROM role_permissions
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator'));

-- Remove user-role assignments for new roles
DELETE FROM user_roles
WHERE role_id IN (SELECT id FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator'));

-- Remove new roles
DELETE FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');

-- Remove new permissions
DELETE FROM permissions WHERE resource IN ('cms', 'rbac')
  OR name IN ('users.create', 'users.read', 'users.update', 'users.delete',
              'users.roles.assign', 'users.roles.remove', 'users.status.toggle',
              'urls.manage', 'system.logs', 'system.health', 'settings.update',
              'analytics.read.all', 'analytics.export');

-- Restore is_admin comment
COMMENT ON COLUMN users.is_admin IS NULL;

RAISE NOTICE 'Migration 030 rolled back successfully';
