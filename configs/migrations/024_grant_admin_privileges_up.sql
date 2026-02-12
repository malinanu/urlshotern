-- Migration: Grant Admin Privileges to Specific User
-- Description: Set is_admin=true for mjjayaweera@gmail.com
-- Category: User Management
-- Dependencies: users table must exist

-- Update user to grant admin privileges
UPDATE users
SET
    is_admin = true,
    email_verified = true,  -- Ensure email is verified for admin access
    is_active = true,       -- Ensure account is active
    updated_at = NOW()
WHERE email = 'mjjayaweera@gmail.com';

-- Verify the update was successful
DO $$
DECLARE
    updated_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO updated_count
    FROM users
    WHERE email = 'mjjayaweera@gmail.com' AND is_admin = true;

    IF updated_count = 0 THEN
        RAISE WARNING 'User mjjayaweera@gmail.com not found or admin privileges not granted';
    ELSE
        RAISE NOTICE 'Admin privileges granted successfully to mjjayaweera@gmail.com';
    END IF;
END $$;
