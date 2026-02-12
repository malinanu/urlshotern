#!/bin/bash
# Quick script to grant admin privileges to mjjayaweera@gmail.com

echo "Applying admin privileges migration..."

docker exec -i urlshorter-postgres-1 psql -U urlshortener -d urlshortener_db << 'EOF'
-- Grant admin privileges
UPDATE users
SET
    is_admin = true,
    email_verified = true,
    is_active = true,
    updated_at = NOW()
WHERE email = 'mjjayaweera@gmail.com';

-- Verify the update
SELECT
    email,
    is_admin,
    email_verified,
    is_active,
    updated_at
FROM users
WHERE email = 'mjjayaweera@gmail.com';

-- Record in migration_history if table exists
INSERT INTO migration_history (version, name, applied_at, success)
VALUES (24, 'grant_admin_privileges', NOW(), true)
ON CONFLICT (version) DO NOTHING;

EOF

echo "✅ Admin privileges granted successfully!"
echo "Verify by checking the output above - is_admin should be 't' (true)"
