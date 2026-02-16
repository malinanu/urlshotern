#!/usr/bin/env pwsh
# ============================================================================
# Assign Super Admin Role to mjjayaweera@gmail.com
# ============================================================================

param(
    [string]$Email = "mjjayaweera@gmail.com",
    [string]$DbHost = "localhost",
    [string]$DbPort = "5432",
    [string]$DbName = "urlshortener",
    [string]$DbUser = "urlshortener_user",
    [string]$DbPassword = "urlshortener_password"
)

Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host "Assign Super Admin Role Script" -ForegroundColor Cyan
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host ""

# Check if psql is installed
$psqlPath = Get-Command psql -ErrorAction SilentlyContinue
if (-not $psqlPath) {
    Write-Host "ERROR: psql command not found. Please install PostgreSQL client tools." -ForegroundColor Red
    Write-Host ""
    Write-Host "Alternative: Run the SQL commands manually in Adminer or pgAdmin" -ForegroundColor Yellow
    Write-Host "  1. Open http://localhost:8081 (Adminer)" -ForegroundColor Yellow
    Write-Host "  2. Login with database credentials" -ForegroundColor Yellow
    Write-Host "  3. Execute the SQL from: scripts/assign-super-admin.sql" -ForegroundColor Yellow
    exit 1
}

Write-Host "Step 1: Checking existing users..." -ForegroundColor Green
Write-Host ""

$env:PGPASSWORD = $DbPassword
$checkUsersQuery = "SELECT id, email, username FROM users ORDER BY created_at DESC;"
psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $checkUsersQuery

Write-Host ""
Write-Host "Step 2: Finding user: $Email" -ForegroundColor Green
Write-Host ""

$findUserQuery = "SELECT id, email, username FROM users WHERE email = '$Email';"
$userResult = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $findUserQuery

if ([string]::IsNullOrWhiteSpace($userResult)) {
    Write-Host "ERROR: User with email '$Email' not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "Please check the email address or create the user first." -ForegroundColor Yellow
    exit 1
}

Write-Host "Found user: $userResult" -ForegroundColor Cyan
Write-Host ""

Write-Host "Step 3: Finding super_admin role..." -ForegroundColor Green
Write-Host ""

$findRoleQuery = "SELECT id, name, display_name FROM roles WHERE name = 'super_admin';"
$roleResult = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $findRoleQuery

if ([string]::IsNullOrWhiteSpace($roleResult)) {
    Write-Host "ERROR: super_admin role not found!" -ForegroundColor Red
    Write-Host ""
    Write-Host "The role might not exist in the database. Check roles table." -ForegroundColor Yellow
    exit 1
}

Write-Host "Found role: $roleResult" -ForegroundColor Cyan
Write-Host ""

Write-Host "Step 4: Checking existing role assignments for $Email..." -ForegroundColor Green
Write-Host ""

$checkRolesQuery = @"
SELECT r.name as role, ur.assigned_at
FROM user_roles ur
JOIN roles r ON ur.role_id = r.id
JOIN users u ON ur.user_id = u.id
WHERE u.email = '$Email';
"@

psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $checkRolesQuery

Write-Host ""
Write-Host "Step 5: Assigning super_admin role to $Email..." -ForegroundColor Green
Write-Host ""

# Get the first user as the assigner (or use a specific admin user)
$getAssignerQuery = "SELECT id FROM users ORDER BY id LIMIT 1;"
$assignerId = (psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c $getAssignerQuery).Trim()

$assignRoleQuery = @"
INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at, expires_at)
VALUES (
    (SELECT id FROM users WHERE email = '$Email'),
    (SELECT id FROM roles WHERE name = 'super_admin'),
    $assignerId,
    NOW(),
    NULL
)
ON CONFLICT (user_id, role_id) DO NOTHING
RETURNING user_id, role_id, assigned_at;
"@

$assignResult = psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $assignRoleQuery

if ($LASTEXITCODE -eq 0) {
    Write-Host "SUCCESS: super_admin role assigned to $Email" -ForegroundColor Green
} else {
    Write-Host "ERROR: Failed to assign role. Check the error message above." -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 6: Verifying role assignment..." -ForegroundColor Green
Write-Host ""

$verifyQuery = @"
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
WHERE u.email = '$Email'
ORDER BY ur.assigned_at DESC;
"@

psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c $verifyQuery

Write-Host ""
Write-Host "==============================================================================" -ForegroundColor Cyan
Write-Host "Done! User $Email should now have super_admin role." -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Yellow
Write-Host "  1. Log out of the admin panel" -ForegroundColor Yellow
Write-Host "  2. Log in as $Email" -ForegroundColor Yellow
Write-Host "  3. You should now have full super admin access" -ForegroundColor Yellow
Write-Host "==============================================================================" -ForegroundColor Cyan
