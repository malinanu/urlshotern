-- ============================================================================
-- Create Super Admin User Script
-- Description: Creates the first super admin user for the system
-- Usage: psql -U postgres -d urlshortener_db -f create_super_admin.sql
-- ============================================================================

-- Connect to database
\c urlshortener_db;

-- ============================================================================
-- Configuration - CHANGE THESE VALUES
-- ============================================================================
-- Super Admin Details
-- Email: admin@example.com
-- Password: Admin123!@# (CHANGE THIS IMMEDIATELY AFTER FIRST LOGIN!)
-- Snowflake ID: 1 (for simplicity, or generate a proper one)

-- ============================================================================
-- Step 1: Create Super Admin User
-- ============================================================================

DO $$
DECLARE
    admin_user_id BIGINT := 1; -- Simple ID for super admin
    admin_email VARCHAR(255) := 'admin@example.com';
    admin_name VARCHAR(100) := 'Super Administrator';
    -- Password hash for 'Admin123!@#' (use bcrypt in production!)
    -- This is a placeholder - generate with: bcrypt.GenerateFromPassword([]byte("Admin123!@#"), bcrypt.DefaultCost)
    admin_password_hash VARCHAR(255) := '$2a$10$placeholder.hash.change.this.immediately';
    super_admin_role_id BIGINT;
BEGIN
    -- Check if super admin already exists
    IF EXISTS (SELECT 1 FROM users WHERE email = admin_email) THEN
        RAISE NOTICE 'Super admin user already exists with email: %', admin_email;
    ELSE
        -- Insert super admin user
        INSERT INTO users (
            id,
            email,
            name,
            password_hash,
            email_verified,
            phone_verified,
            account_type,
            is_active,
            is_admin,
            created_at,
            updated_at
        ) VALUES (
            admin_user_id,
            admin_email,
            admin_name,
            admin_password_hash,
            TRUE,  -- Email verified by default
            FALSE, -- Phone not verified
            'super_admin', -- Account type
            TRUE,  -- Active
            TRUE,  -- Is admin (legacy flag)
            NOW(),
            NOW()
        );

        RAISE NOTICE 'Created super admin user with ID: % and email: %', admin_user_id, admin_email;

        -- Assign super_admin role to user
        SELECT id INTO super_admin_role_id FROM roles WHERE name = 'super_admin';

        IF super_admin_role_id IS NOT NULL THEN
            INSERT INTO user_roles (
                user_id,
                role_id,
                assigned_by,
                assigned_at
            ) VALUES (
                admin_user_id,
                super_admin_role_id,
                admin_user_id, -- Self-assigned
                NOW()
            );

            RAISE NOTICE 'Assigned super_admin role to user';
        ELSE
            RAISE NOTICE 'Warning: super_admin role not found. Run migration 001 first.';
        END IF;

        -- Log the admin creation in audit logs (if admin schema exists)
        IF EXISTS (SELECT 1 FROM pg_tables WHERE schemaname = 'admin' AND tablename = 'audit_logs') THEN
            INSERT INTO admin.audit_logs (
                id,
                user_id,
                username,
                action,
                resource_type,
                resource_id,
                details,
                ip_address,
                status,
                severity,
                created_at
            ) VALUES (
                1000, -- Simple ID for initial log
                admin_user_id,
                admin_email,
                'user.create',
                'user',
                admin_user_id::VARCHAR,
                '{"method": "sql_script", "type": "super_admin_creation"}'::JSONB,
                '127.0.0.1'::INET,
                'success',
                'critical',
                NOW()
            );
        END IF;
    END IF;
END $$;

-- ============================================================================
-- Step 2: Verify Super Admin Creation
-- ============================================================================

DO $$
DECLARE
    admin_count INTEGER;
    admin_roles TEXT;
BEGIN
    -- Check super admin exists
    SELECT COUNT(*) INTO admin_count
    FROM users
    WHERE account_type = 'super_admin' OR is_admin = TRUE;

    RAISE NOTICE 'Total admin users: %', admin_count;

    -- Show admin roles
    SELECT string_agg(r.name, ', ') INTO admin_roles
    FROM users u
    JOIN user_roles ur ON u.id = ur.user_id
    JOIN roles r ON ur.role_id = r.id
    WHERE u.account_type = 'super_admin' OR u.is_admin = TRUE;

    RAISE NOTICE 'Admin roles: %', COALESCE(admin_roles, 'None');
END $$;

-- ============================================================================
-- Step 3: Display Super Admin Credentials
-- ============================================================================

SELECT
    id,
    email,
    name,
    account_type,
    is_admin,
    email_verified,
    is_active,
    created_at
FROM users
WHERE account_type = 'super_admin' OR is_admin = TRUE;

-- ============================================================================
-- IMPORTANT SECURITY NOTES
-- ============================================================================

/*
⚠️ CRITICAL SECURITY WARNINGS:

1. CHANGE THE PASSWORD IMMEDIATELY after first login!
   - Default email: admin@example.com
   - Default password: Admin123!@#

2. The password hash in this script is a PLACEHOLDER.
   You MUST generate a proper bcrypt hash before using this script.

3. To generate a proper password hash in Go:

   import "golang.org/x/crypto/bcrypt"

   hash, err := bcrypt.GenerateFromPassword([]byte("YourSecurePassword"), bcrypt.DefaultCost)
   if err != nil {
       log.Fatal(err)
   }
   fmt.Println(string(hash))

4. Replace the 'admin_password_hash' variable above with your generated hash.

5. After creating the super admin:
   - Log in immediately
   - Change the password via the UI or API
   - Enable MFA if available
   - Review audit logs regularly

6. Never commit the actual password or hash to version control.

7. For production deployments, use environment variables or secrets management.

============================================================================
*/

-- ============================================================================
-- Alternative: Update Existing User to Super Admin
-- ============================================================================

/*
-- If you want to promote an existing user to super admin, use this instead:

DO $$
DECLARE
    existing_user_id BIGINT;
    super_admin_role_id BIGINT;
BEGIN
    -- Find user by email
    SELECT id INTO existing_user_id FROM users WHERE email = 'user@example.com';

    IF existing_user_id IS NULL THEN
        RAISE EXCEPTION 'User not found with email: user@example.com';
    END IF;

    -- Update to super admin
    UPDATE users
    SET
        account_type = 'super_admin',
        is_admin = TRUE,
        updated_at = NOW()
    WHERE id = existing_user_id;

    -- Assign super_admin role
    SELECT id INTO super_admin_role_id FROM roles WHERE name = 'super_admin';

    IF super_admin_role_id IS NOT NULL THEN
        INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
        VALUES (existing_user_id, super_admin_role_id, existing_user_id, NOW())
        ON CONFLICT (user_id, role_id) DO NOTHING;
    END IF;

    RAISE NOTICE 'Promoted user % to super admin', existing_user_id;
END $$;
*/
