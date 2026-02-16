-- ============================================================================
-- Assign Super Admin Role to mjjayaweera@gmail.com
-- ============================================================================

-- Step 1: Check existing users
SELECT id, email, username, created_at
FROM users
ORDER BY created_at DESC;

-- Step 2: Check existing roles
SELECT id, name, display_name, is_system
FROM roles
ORDER BY id;

-- Step 3: Find the user ID for mjjayaweera@gmail.com
SELECT id, email, username
FROM users
WHERE email = 'mjjayaweera@gmail.com';

-- Step 4: Find the super_admin role ID
SELECT id, name
FROM roles
WHERE name = 'super_admin';

-- Step 5: Check if user already has super_admin role
SELECT ur.*, r.name as role_name, u.email as user_email
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
JOIN users u ON ur.user_id = u.id
WHERE u.email = 'mjjayaweera@gmail.com';

-- Step 6: Assign super_admin role to mjjayaweera@gmail.com
-- NOTE: Replace <user_id> and <role_id> with actual IDs from Step 3 and 4
-- Also replace <your_admin_id> with your current admin user ID

-- Get your admin ID (logged-in user)
SELECT id, email, username
FROM users
WHERE email = 'your-admin-email@example.com';  -- Replace with your email

-- Assign the role (uncomment and run after getting the IDs)
-- INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at, expires_at)
-- VALUES (
--     (SELECT id FROM users WHERE email = 'mjjayaweera@gmail.com'),
--     (SELECT id FROM roles WHERE name = 'super_admin'),
--     (SELECT id FROM users WHERE email = 'your-admin-email@example.com'),  -- Your admin ID
--     NOW(),
--     NULL  -- No expiration
-- )
-- ON CONFLICT (user_id, role_id) DO NOTHING;

-- Step 7: Verify the assignment
SELECT
    u.id as user_id,
    u.email,
    u.username,
    r.id as role_id,
    r.name as role_name,
    r.display_name,
    ur.assigned_at,
    ur.expires_at,
    assigner.email as assigned_by_email
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
LEFT JOIN users assigner ON ur.assigned_by = assigner.id
WHERE u.email = 'mjjayaweera@gmail.com'
ORDER BY ur.assigned_at DESC;
