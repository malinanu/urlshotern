-- Migration Rollback: Remove Admin Privileges
-- Description: Revoke admin privileges from mjjayaweera@gmail.com

UPDATE users
SET
    is_admin = false,
    updated_at = NOW()
WHERE email = 'mjjayaweera@gmail.com';

-- Verify rollback
DO $$
DECLARE
    updated_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO updated_count
    FROM users
    WHERE email = 'mjjayaweera@gmail.com' AND is_admin = false;

    IF updated_count = 0 THEN
        RAISE WARNING 'User mjjayaweera@gmail.com not found or rollback failed';
    ELSE
        RAISE NOTICE 'Admin privileges revoked successfully from mjjayaweera@gmail.com';
    END IF;
END $$;
