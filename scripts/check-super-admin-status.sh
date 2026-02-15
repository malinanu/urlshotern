#!/bin/bash
# Check super admin status on production

echo "=========================================="
echo "Super Admin Status Check"
echo "=========================================="
echo ""

# Check which users have super_admin role
echo "Users with super_admin role:"
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT
    u.id,
    u.name,
    u.email,
    r.name as role_name,
    ur.assigned_at,
    ur.assigned_by
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE r.name = 'super_admin'
ORDER BY ur.assigned_at;
EOF

echo ""
echo "=========================================="
echo ""

# Check both users
echo "User 112125335771807740 roles:"
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT u.id, u.name, u.email, r.name as role_name
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE u.id = 112125335771807740;
EOF

echo ""
echo "User 112125335771807745 roles:"
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT u.id, u.name, u.email, r.name as role_name
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE u.id = 112125335771807745;
EOF

echo ""
echo "=========================================="
echo ""

# Check migration status
echo "Migration 033 status:"
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT version, applied_at, description
FROM schema_migrations
WHERE version = '033';
EOF
