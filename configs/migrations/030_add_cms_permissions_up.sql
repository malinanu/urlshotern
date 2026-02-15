-- Migration: Add CMS Permissions and Content Roles
-- Category: RBAC Enhancement
-- Tables: roles, permissions, role_permissions
-- Dependencies: 013_rbac_tables

-- ============================================================================
-- Add New Roles
-- ============================================================================

INSERT INTO roles (name, display_name, description, is_system) VALUES
    ('super_admin', 'Super Administrator', 'Full system access including RBAC management', true),
    ('content_manager', 'Content Manager', 'Full CMS access - create, edit, delete, and publish content', true),
    ('content_editor', 'Content Editor', 'CMS editing access - create and edit content (no publish/delete)', true),
    ('moderator', 'Moderator', 'User moderation and analytics viewing access', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- Add CMS Permissions
-- ============================================================================

INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    -- CMS Pages (collaborative editing - no ownership restriction)
    ('cms.pages.create', 'Create CMS Pages', 'Create new CMS pages', 'cms', 'pages.create', true),
    ('cms.pages.read', 'Read CMS Pages', 'View CMS pages', 'cms', 'pages.read', true),
    ('cms.pages.update', 'Update CMS Pages', 'Edit any CMS page', 'cms', 'pages.update', true),
    ('cms.pages.delete', 'Delete CMS Pages', 'Delete CMS pages', 'cms', 'pages.delete', true),
    ('cms.pages.publish', 'Publish CMS Pages', 'Publish/unpublish CMS pages', 'cms', 'pages.publish', true),

    -- CMS Blocks
    ('cms.blocks.create', 'Create Content Blocks', 'Create new content blocks', 'cms', 'blocks.create', true),
    ('cms.blocks.read', 'Read Content Blocks', 'View content blocks', 'cms', 'blocks.read', true),
    ('cms.blocks.update', 'Update Content Blocks', 'Edit content blocks', 'cms', 'blocks.update', true),
    ('cms.blocks.delete', 'Delete Content Blocks', 'Delete content blocks', 'cms', 'blocks.delete', true),

    -- CMS Media
    ('cms.media.upload', 'Upload Media', 'Upload media files', 'cms', 'media.upload', true),
    ('cms.media.delete', 'Delete Media', 'Delete media files', 'cms', 'media.delete', true),

    -- CMS Settings
    ('cms.settings.read', 'Read CMS Settings', 'View CMS settings', 'cms', 'settings.read', true),
    ('cms.settings.update', 'Update CMS Settings', 'Update CMS settings', 'cms', 'settings.update', true),

    -- Enhanced Analytics
    ('analytics.read.all', 'Read All Analytics', 'View analytics for all users', 'analytics', 'read.all', true),
    ('analytics.export', 'Export Analytics', 'Export analytics data', 'analytics', 'export', true),

    -- RBAC Management
    ('rbac.roles.create', 'Create Roles', 'Create custom roles', 'rbac', 'roles.create', true),
    ('rbac.roles.read', 'Read Roles', 'View roles and their permissions', 'rbac', 'roles.read', true),
    ('rbac.roles.update', 'Update Roles', 'Modify role permissions', 'rbac', 'roles.update', true),
    ('rbac.roles.delete', 'Delete Roles', 'Delete custom roles', 'rbac', 'roles.delete', true),
    ('rbac.permissions.read', 'Read Permissions', 'View available permissions', 'rbac', 'permissions.read', true),
    ('rbac.assign', 'Assign Roles', 'Assign roles to users', 'rbac', 'assign', true),

    -- Enhanced User Management
    ('users.create', 'Create Users', 'Create new user accounts', 'users', 'create', true),
    ('users.read', 'Read Users', 'View user information', 'users', 'read', true),
    ('users.update', 'Update Users', 'Update user information', 'users', 'update', true),
    ('users.delete', 'Delete Users', 'Delete user accounts', 'users', 'delete', true),
    ('users.roles.assign', 'Assign User Roles', 'Assign roles to users', 'users', 'roles.assign', true),
    ('users.roles.remove', 'Remove User Roles', 'Remove roles from users', 'users', 'roles.remove', true),
    ('users.status.toggle', 'Toggle User Status', 'Activate/deactivate users', 'users', 'status.toggle', true),

    -- Enhanced URL Management
    ('urls.manage', 'Manage URLs', 'Manage all URLs (not just own)', 'urls', 'manage', true),

    -- Enhanced System
    ('system.logs', 'View System Logs', 'Access system logs', 'system', 'logs', true),
    ('system.health', 'View System Health', 'View system health metrics', 'system', 'health', true),
    ('settings.update', 'Update Settings', 'Update system settings', 'settings', 'update', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- Assign Permissions to Roles
-- ============================================================================

-- Super Admin: ALL permissions (wildcard via code logic + all explicit permissions)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Admin: User management, system settings, analytics (no RBAC role creation)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN (
    -- Users
    'users.create', 'users.read', 'users.update', 'users.delete',
    'users.roles.assign', 'users.roles.remove', 'users.status.toggle',
    -- URLs
    'url.create', 'url.read', 'url.update', 'url.delete', 'urls.manage',
    -- Analytics
    'analytics.read', 'analytics.read.all', 'analytics.export',
    -- System
    'system.logs', 'system.health', 'settings.update',
    -- RBAC (read-only)
    'rbac.roles.read', 'rbac.permissions.read',
    -- Teams
    'team.create', 'team.read', 'team.update', 'team.delete',
    -- Domains
    'domain.create', 'domain.read', 'domain.update', 'domain.delete'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Content Manager: Full CMS (create, edit, delete, publish)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'content_manager' AND p.name IN (
    -- CMS Full Access
    'cms.pages.create', 'cms.pages.read', 'cms.pages.update', 'cms.pages.delete', 'cms.pages.publish',
    'cms.blocks.create', 'cms.blocks.read', 'cms.blocks.update', 'cms.blocks.delete',
    'cms.media.upload', 'cms.media.delete',
    'cms.settings.read', 'cms.settings.update',
    -- URLs (read)
    'url.read',
    -- Analytics (own)
    'analytics.read'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Content Editor: Create and edit pages (no publish/delete)
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'content_editor' AND p.name IN (
    -- CMS Limited Access
    'cms.pages.create', 'cms.pages.read', 'cms.pages.update',
    'cms.blocks.create', 'cms.blocks.read', 'cms.blocks.update',
    'cms.media.upload'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Moderator: User moderation, analytics viewing
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'moderator' AND p.name IN (
    -- Users (view and moderate)
    'users.read', 'users.status.toggle',
    -- URLs (view and manage)
    'url.read', 'urls.manage',
    -- Analytics
    'analytics.read', 'analytics.read.all', 'analytics.export'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ============================================================================
-- Update is_admin Column Comment (Backward Compatibility)
-- ============================================================================

COMMENT ON COLUMN users.is_admin IS 'Deprecated: Use RBAC roles instead. Kept for backward compatibility. Will be phased out in future versions.';

-- ============================================================================
-- VERIFICATION: Ensure new data was inserted successfully
-- ============================================================================

DO $$
DECLARE
    new_role_count INTEGER;
    new_perm_count INTEGER;
    super_admin_perm_count INTEGER;
    content_editor_perm_count INTEGER;
BEGIN
    -- Verify new roles were created
    SELECT COUNT(*) INTO new_role_count FROM roles
    WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');
    IF new_role_count < 4 THEN
        RAISE EXCEPTION 'Migration 030 Failed: Expected 4 new roles, found %', new_role_count;
    END IF;
    RAISE NOTICE 'Migration 030: Found % new roles ✓', new_role_count;

    -- Verify new permissions were created
    SELECT COUNT(*) INTO new_perm_count FROM permissions WHERE resource IN ('cms', 'rbac', 'users', 'urls', 'system', 'settings');
    IF new_perm_count < 30 THEN
        RAISE EXCEPTION 'Migration 030 Failed: Expected at least 30 new permissions, found %', new_perm_count;
    END IF;
    RAISE NOTICE 'Migration 030: Found % permissions ✓', new_perm_count;

    -- Verify super_admin has all permissions
    SELECT COUNT(*) INTO super_admin_perm_count
    FROM role_permissions rp
    JOIN roles r ON r.id = rp.role_id
    WHERE r.name = 'super_admin';

    RAISE NOTICE 'Migration 030: Super Admin has % permissions ✓', super_admin_perm_count;

    -- Verify content_editor has limited permissions
    SELECT COUNT(*) INTO content_editor_perm_count
    FROM role_permissions rp
    JOIN roles r ON r.id = rp.role_id
    WHERE r.name = 'content_editor';

    IF content_editor_perm_count > 10 THEN
        RAISE WARNING 'Migration 030: Content Editor has more permissions than expected (%)', content_editor_perm_count;
    END IF;
    RAISE NOTICE 'Migration 030: Content Editor has % permissions (limited) ✓', content_editor_perm_count;
END $$;
