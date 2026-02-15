# ============================================================================
# Super Admin Migration Prerequisites Verification Script (PowerShell)
# ============================================================================
# Purpose: Verify that all prerequisites are met before applying migration 033
# Usage: .\scripts\verify-super-admin-prereqs.ps1

# Configuration
$USER_ID = "112125335771807740"
$ROLE_NAME = "super_admin"

# Color output functions
function Write-Header {
    param([string]$Message)
    Write-Host "`n========================================" -ForegroundColor Blue
    Write-Host $Message -ForegroundColor Blue
    Write-Host "========================================`n" -ForegroundColor Blue
}

function Write-Step {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Error {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Info {
    param([string]$Message)
    Write-Host "ⓘ $Message" -ForegroundColor Cyan
}

# Function to run SQL query
function Invoke-DatabaseQuery {
    param([string]$Query)

    $containerId = docker ps -q -f name=postgres | Select-Object -First 1
    if (-not $containerId) {
        throw "PostgreSQL container not found"
    }

    $result = docker exec -i $containerId psql -U urlshortener -d urlshortener_db -t -c $Query 2>&1
    return $result
}

Write-Header "Super Admin Migration Prerequisites Check"

# Check 1: PostgreSQL container is running
Write-Step "[1/5] Checking PostgreSQL container..."
try {
    $postgresContainer = docker ps -q -f name=postgres | Select-Object -First 1
    if ($postgresContainer) {
        Write-Success "PostgreSQL container is running"
    } else {
        Write-Error "PostgreSQL container not found"
        Write-Warning "Please start your Docker containers first"
        exit 1
    }
} catch {
    Write-Error "Cannot check Docker containers: $_"
    exit 1
}

# Check 2: Database connection
Write-Step "`n[2/5] Testing database connection..."
try {
    $testQuery = Invoke-DatabaseQuery "SELECT 1;"
    if ($testQuery -match "1") {
        Write-Success "Database connection successful"
    } else {
        Write-Error "Cannot connect to database"
        exit 1
    }
} catch {
    Write-Error "Database connection failed: $_"
    exit 1
}

# Check 3: User exists
Write-Step "`n[3/5] Checking if user $USER_ID exists..."
try {
    $userQuery = "SELECT id, name, email FROM users WHERE id = $USER_ID;"
    $userInfo = Invoke-DatabaseQuery $userQuery
    $userInfoClean = ($userInfo | Out-String).Trim()

    if ($userInfoClean -and $userInfoClean -match "\d+") {
        Write-Success "User found:"
        Write-Host "  $userInfoClean" -ForegroundColor Gray
    } else {
        Write-Error "User $USER_ID not found in database"
        Write-Warning "Please verify the user ID or create the user first"
        exit 1
    }
} catch {
    Write-Error "Failed to check user: $_"
    exit 1
}

# Check 4: Super admin role exists
Write-Step "`n[4/5] Checking if super_admin role exists..."
try {
    $roleQuery = "SELECT id, name, display_name FROM roles WHERE name = '$ROLE_NAME';"
    $roleInfo = Invoke-DatabaseQuery $roleQuery
    $roleInfoClean = ($roleInfo | Out-String).Trim()

    if ($roleInfoClean -and $roleInfoClean -match "super_admin") {
        Write-Success "Role found:"
        Write-Host "  $roleInfoClean" -ForegroundColor Gray
    } else {
        Write-Error "Super admin role not found"
        Write-Warning "Please run migration 030 first:"
        Write-Host "  docker exec -i `$(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs\migrations\030_add_cms_permissions_up.sql" -ForegroundColor Yellow
        exit 1
    }
} catch {
    Write-Error "Failed to check role: $_"
    exit 1
}

# Check 5: Current role assignment status
Write-Step "`n[5/5] Checking current role assignments for user..."
try {
    $rolesQuery = "SELECT r.name FROM user_roles ur JOIN roles r ON ur.role_id = r.id WHERE ur.user_id = $USER_ID;"
    $currentRoles = Invoke-DatabaseQuery $rolesQuery
    $currentRolesClean = ($currentRoles | Out-String).Trim()

    if ($currentRolesClean) {
        Write-Info "User currently has roles:"
        Write-Host "  $currentRolesClean" -ForegroundColor Gray

        if ($currentRolesClean -match "super_admin") {
            Write-Warning "User already has super_admin role"
            Write-Warning "Migration 033 will be skipped (idempotent)"
        } else {
            Write-Success "User does not have super_admin role yet"
        }
    } else {
        Write-Info "User has no roles assigned yet"
    }
} catch {
    Write-Warning "Could not check current roles: $_"
}

# Final summary
Write-Host "`n========================================" -ForegroundColor Green
Write-Success "All prerequisites met!"
Write-Host "========================================`n" -ForegroundColor Green

Write-Host "Ready to apply migration 033:" -ForegroundColor Blue
Write-Host "  docker exec -i `$(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs\migrations\033_assign_initial_super_admin_up.sql`n" -ForegroundColor Yellow

exit 0
