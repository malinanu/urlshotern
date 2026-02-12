#!/bin/bash
echo "Starting Docker services..."
docker-compose -f docker/docker-compose.local.yml up -d postgres redis

echo "Waiting for PostgreSQL to be ready..."
sleep 5

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

-- Record in migration_history
INSERT INTO migration_history (version, name, applied_at, success)
VALUES (24, 'grant_admin_privileges', NOW(), true)
ON CONFLICT (version) DO NOTHING;

SELECT 'Migration 024 applied successfully!' as status;
EOF

echo "Done! Check the output above to verify is_admin = t (true)"
