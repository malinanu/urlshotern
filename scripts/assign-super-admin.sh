#!/usr/bin/env bash
# ============================================================================
# Assign Super Admin Role to mjjayaweera@gmail.com
# ============================================================================

set -e  # Exit on error

# Configuration
EMAIL="${1:-mjjayaweera@gmail.com}"
DB_HOST="${DB_HOST:-localhost}"
DB_PORT="${DB_PORT:-5432}"
DB_NAME="${DB_NAME:-urlshortener}"
DB_USER="${DB_USER:-urlshortener_user}"
DB_PASSWORD="${DB_PASSWORD:-urlshortener_password}"

echo "=============================================================================="
echo "Assign Super Admin Role Script"
echo "=============================================================================="
echo ""

# Check if psql is installed
if ! command -v psql &> /dev/null; then
    echo "ERROR: psql command not found. Please install PostgreSQL client tools."
    echo ""
    echo "Alternative: Run the SQL commands manually in Adminer or pgAdmin"
    echo "  1. Open http://localhost:8081 (Adminer)"
    echo "  2. Login with database credentials"
    echo "  3. Execute the SQL from: scripts/assign-super-admin.sql"
    exit 1
fi

export PGPASSWORD="$DB_PASSWORD"

echo "Step 1: Checking existing users..."
echo ""
psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c \
    "SELECT id, email, username FROM users ORDER BY created_at DESC;"

echo ""
echo "Step 2: Finding user: $EMAIL"
echo ""

USER_RESULT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT id, email, username FROM users WHERE email = '$EMAIL';")

if [ -z "$USER_RESULT" ]; then
    echo "ERROR: User with email '$EMAIL' not found!"
    echo ""
    echo "Please check the email address or create the user first."
    exit 1
fi

echo "Found user: $USER_RESULT"
echo ""

echo "Step 3: Finding super_admin role..."
echo ""

ROLE_RESULT=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT id, name, display_name FROM roles WHERE name = 'super_admin';")

if [ -z "$ROLE_RESULT" ]; then
    echo "ERROR: super_admin role not found!"
    echo ""
    echo "The role might not exist in the database. Check roles table."
    exit 1
fi

echo "Found role: $ROLE_RESULT"
echo ""

echo "Step 4: Checking existing role assignments for $EMAIL..."
echo ""

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT r.name as role, ur.assigned_at
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
JOIN users u ON ur.user_id = u.id
WHERE u.email = '$EMAIL';
"

echo ""
echo "Step 5: Assigning super_admin role to $EMAIL..."
echo ""

# Get the first user as the assigner
ASSIGNER_ID=$(psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -t -c \
    "SELECT id FROM users ORDER BY id LIMIT 1;" | tr -d ' ')

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at, expires_at)
VALUES (
    (SELECT id FROM users WHERE email = '$EMAIL'),
    (SELECT id FROM roles WHERE name = 'super_admin'),
    $ASSIGNER_ID,
    NOW(),
    NULL
)
ON CONFLICT (user_id, role_id) DO NOTHING
RETURNING user_id, role_id, assigned_at;
"

if [ $? -eq 0 ]; then
    echo "SUCCESS: super_admin role assigned to $EMAIL"
else
    echo "ERROR: Failed to assign role. Check the error message above."
    exit 1
fi

echo ""
echo "Step 6: Verifying role assignment..."
echo ""

psql -h "$DB_HOST" -p "$DB_PORT" -U "$DB_USER" -d "$DB_NAME" -c "
SELECT
    u.email,
    u.username,
    r.name as role,
    r.display_name as role_display_name,
    ur.assigned_at,
    assigner.email as assigned_by
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
LEFT JOIN users assigner ON ur.assigned_by = assigner.id
WHERE u.email = '$EMAIL'
ORDER BY ur.assigned_at DESC;
"

echo ""
echo "=============================================================================="
echo "Done! User $EMAIL should now have super_admin role."
echo ""
echo "Next steps:"
echo "  1. Log out of the admin panel"
echo "  2. Log in as $EMAIL"
echo "  3. You should now have full super admin access"
echo "=============================================================================="
