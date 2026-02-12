-- Migration 023: Verify and Ensure Initial Data
-- Purpose: Idempotent verification that all required RBAC data exists
-- Safe to run multiple times

-- Ensure roles exist (upsert pattern)
INSERT INTO roles (name, display_name, description, is_system, is_active) VALUES
    ('admin', 'Administrator', 'Full system access', true, true),
    ('user', 'User', 'Standard user access', true, true),
    ('premium', 'Premium User', 'Premium user with additional features', true, true),
    ('team_admin', 'Team Administrator', 'Team administration access', true, true),
    ('team_member', 'Team Member', 'Team member access', true, true)
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description,
    is_active = EXCLUDED.is_active;

-- Ensure permissions exist (upsert pattern)
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    ('user.create', 'Create Users', 'Create new users', 'user', 'create', true),
    ('user.read', 'Read Users', 'View user information', 'user', 'read', true),
    ('user.update', 'Update Users', 'Update user information', 'user', 'update', true),
    ('user.delete', 'Delete Users', 'Delete users', 'user', 'delete', true),
    ('url.create', 'Create URLs', 'Create new URLs', 'url', 'create', true),
    ('url.read', 'Read URLs', 'View URLs', 'url', 'read', true),
    ('url.update', 'Update URLs', 'Update URLs', 'url', 'update', true),
    ('url.delete', 'Delete URLs', 'Delete URLs', 'url', 'delete', true),
    ('analytics.read', 'Read Analytics', 'View analytics data', 'analytics', 'read', true),
    ('team.create', 'Create Teams', 'Create new teams', 'team', 'create', true),
    ('team.read', 'Read Teams', 'View team information', 'team', 'read', true),
    ('team.update', 'Update Teams', 'Update team information', 'team', 'update', true),
    ('team.delete', 'Delete Teams', 'Delete teams', 'team', 'delete', true),
    ('domain.create', 'Create Custom Domains', 'Add custom domains', 'domain', 'create', true),
    ('domain.read', 'Read Custom Domains', 'View custom domains', 'domain', 'read', true),
    ('domain.update', 'Update Custom Domains', 'Update custom domains', 'domain', 'update', true),
    ('domain.delete', 'Delete Custom Domains', 'Delete custom domains', 'domain', 'delete', true),
    ('cms.read', 'Read CMS Content', 'View CMS pages and content', 'cms', 'read', true),
    ('cms.create', 'Create CMS Content', 'Create new CMS pages', 'cms', 'create', true),
    ('cms.update', 'Update CMS Content', 'Edit existing CMS pages', 'cms', 'update', true),
    ('cms.delete', 'Delete CMS Content', 'Delete CMS pages', 'cms', 'delete', true),
    ('cms.publish', 'Publish CMS Content', 'Publish and unpublish pages', 'cms', 'publish', true),
    ('api_key.read', 'Read API Keys', 'View API keys', 'api_key', 'read', true),
    ('api_key.create', 'Create API Keys', 'Generate new API keys', 'api_key', 'create', true),
    ('api_key.delete', 'Delete API Keys', 'Revoke API keys', 'api_key', 'delete', true)
ON CONFLICT (name) DO UPDATE SET
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description;

-- Ensure role_permissions assignments
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.is_system = true
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read', 'user.read', 'user.update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'premium' AND p.name IN ('url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read', 'user.read', 'user.update', 'domain.create', 'domain.read', 'domain.update', 'domain.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- CMS permissions for admin
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('cms.read', 'cms.create', 'cms.update', 'cms.delete', 'cms.publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- API key permissions for users and premium
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name IN ('user', 'premium') AND p.name IN ('api_key.read', 'api_key.create', 'api_key.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Ensure global settings exist
INSERT INTO global_settings (key, value, type, category, display_name, description, is_public, sort_order) VALUES
    ('site_name', 'Trunc', 'text', 'general', 'Site Name', 'The name of your website', true, 1),
    ('site_tagline', 'Smart URL Management Platform', 'text', 'general', 'Site Tagline', 'A short description of your site', true, 2),
    ('company_email', 'contact@3logiq.com', 'email', 'contact', 'Company Email', 'Main contact email address', true, 3),
    ('company_phone', '+1-555-0123', 'text', 'contact', 'Company Phone', 'Main contact phone number', true, 4),
    ('company_address', '123 Innovation Drive, Tech City, TC 12345', 'textarea', 'contact', 'Company Address', 'Physical business address', true, 5),
    ('social_twitter', 'https://twitter.com/trunc', 'url', 'social', 'Twitter URL', 'Twitter profile URL', true, 6),
    ('social_linkedin', 'https://linkedin.com/company/trunc', 'url', 'social', 'LinkedIn URL', 'LinkedIn company URL', true, 7),
    ('social_github', 'https://github.com/trunc', 'url', 'social', 'GitHub URL', 'GitHub organization URL', true, 8),
    ('google_analytics', '', 'text', 'tracking', 'Google Analytics ID', 'Google Analytics tracking ID', false, 9),
    ('meta_description', 'Professional URL shortening service with advanced analytics, team collaboration, and custom domains.', 'textarea', 'seo', 'Default Meta Description', 'Default meta description for pages', true, 10),
    ('footer_copyright', '© 2024 3logiq Technologies, Inc. All rights reserved.', 'text', 'footer', 'Copyright Text', 'Copyright notice in footer', true, 11),
    ('header_announcement', '', 'html', 'header', 'Announcement Bar', 'HTML content for announcement bar', true, 12),
    ('maintenance_mode', 'false', 'boolean', 'system', 'Maintenance Mode', 'Enable to show maintenance page', false, 13)
ON CONFLICT (key) DO UPDATE SET
    value = EXCLUDED.value,
    display_name = EXCLUDED.display_name,
    description = EXCLUDED.description;

-- Comprehensive Verification (RBAC + CMS)
DO $$
DECLARE
    role_count INTEGER;
    perm_count INTEGER;
    admin_perm_count INTEGER;
    settings_count INTEGER;
BEGIN
    -- RBAC verification
    SELECT COUNT(*) INTO role_count FROM roles WHERE is_system = true;
    SELECT COUNT(*) INTO perm_count FROM permissions WHERE is_system = true;
    SELECT COUNT(*) INTO admin_perm_count
    FROM role_permissions rp JOIN roles r ON r.id = rp.role_id WHERE r.name = 'admin';

    -- CMS verification
    SELECT COUNT(*) INTO settings_count FROM global_settings;

    IF role_count < 5 THEN
        RAISE EXCEPTION 'Verification Failed: Expected 5 roles, found %', role_count;
    END IF;

    IF perm_count < 25 THEN
        RAISE EXCEPTION 'Verification Failed: Expected 25 permissions (17 RBAC + 8 CMS), found %', perm_count;
    END IF;

    IF admin_perm_count < 22 THEN
        RAISE EXCEPTION 'Verification Failed: Admin should have 22+ permissions, found %', admin_perm_count;
    END IF;

    IF settings_count < 13 THEN
        RAISE EXCEPTION 'Verification Failed: Expected 13 global settings, found %', settings_count;
    END IF;

    RAISE NOTICE '✅ Complete Verification Passed:';
    RAISE NOTICE '  - Roles: %', role_count;
    RAISE NOTICE '  - Permissions: %', perm_count;
    RAISE NOTICE '  - Admin permissions: %', admin_perm_count;
    RAISE NOTICE '  - Global settings: %', settings_count;
END $$;
