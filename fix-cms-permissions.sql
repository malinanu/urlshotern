-- Fix CMS Permissions for Admin Role
-- Run this on your production database

-- Step 1: Check current state
-- Check which roles the admin user has
SELECT
    u.id as user_id,
    u.email,
    u.name,
    r.id as role_id,
    r.name as role_name,
    r.display_name
FROM users u
LEFT JOIN user_roles ur ON u.id = ur.user_id
LEFT JOIN roles r ON ur.role_id = r.id
WHERE u.is_admin = true;

-- Step 2: Check if cms.pages.read permission exists
SELECT id, name, resource, action, description
FROM permissions
WHERE name = 'cms.pages.read';

-- Step 3: Check current permissions for admin role
SELECT
    r.name as role_name,
    p.name as permission_name,
    p.resource,
    p.action
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
LEFT JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin'
ORDER BY p.resource, p.action;

-- Step 4: Assign cms.pages.read permission to admin role
-- First, let's insert the permission if it doesn't exist
INSERT INTO permissions (name, resource, action, description, created_at, updated_at)
VALUES (
    'cms.pages.read',
    'cms.pages',
    'read',
    'Read CMS pages',
    NOW(),
    NOW()
)
ON CONFLICT (name) DO NOTHING;

-- Now assign the permission to the admin role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    r.id,
    p.id,
    NOW()
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.name = 'cms.pages.read'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Step 5: Verify the assignment
SELECT
    r.name as role_name,
    p.name as permission_name,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin' AND p.name = 'cms.pages.read';

-- Step 6: Also add all CMS permissions to admin role
INSERT INTO role_permissions (role_id, permission_id, created_at)
SELECT
    r.id,
    p.id,
    NOW()
FROM roles r
CROSS JOIN permissions p
WHERE r.name = 'admin'
  AND p.resource = 'cms.pages'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Final verification: Show all admin permissions
SELECT
    p.name,
    p.resource,
    p.action,
    p.description
FROM roles r
JOIN role_permissions rp ON r.id = rp.role_id
JOIN permissions p ON rp.permission_id = p.id
WHERE r.name = 'admin'
ORDER BY p.resource, p.action;
