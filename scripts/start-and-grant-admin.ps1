# Complete setup script: Start Docker and grant admin privileges

Write-Host "`n=== URL Shortener - Grant Admin Privileges ===" -ForegroundColor Cyan
Write-Host "This script will:" -ForegroundColor Yellow
Write-Host "  1. Start PostgreSQL and Redis containers" -ForegroundColor Yellow
Write-Host "  2. Grant admin privileges to mjjayaweera@gmail.com" -ForegroundColor Yellow
Write-Host "  3. Verify the changes`n" -ForegroundColor Yellow

# Check if Docker is running
Write-Host "Checking Docker status..." -ForegroundColor Cyan
$dockerRunning = docker ps 2>&1
if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Docker is not running!" -ForegroundColor Red
    Write-Host "Please start Docker Desktop and try again." -ForegroundColor Yellow
    exit 1
}

# Start services
Write-Host "`nStarting Docker services..." -ForegroundColor Cyan
docker-compose -f docker/docker-compose.local.yml up -d postgres redis

if ($LASTEXITCODE -ne 0) {
    Write-Host "❌ Failed to start Docker services" -ForegroundColor Red
    exit 1
}

# Wait for PostgreSQL to be ready
Write-Host "Waiting for PostgreSQL to be ready..." -ForegroundColor Cyan
Start-Sleep -Seconds 5

# Check if containers are running
$pgContainer = docker ps --filter "name=postgres" --format "{{.Names}}" 2>$null
if (-not $pgContainer) {
    Write-Host "❌ PostgreSQL container not found. Checking all containers..." -ForegroundColor Red
    docker ps
    exit 1
}

Write-Host "Found PostgreSQL container: $pgContainer" -ForegroundColor Green

# Apply migration
Write-Host "`nApplying admin privileges migration..." -ForegroundColor Cyan

$sql = @"
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
    to_char(updated_at, 'YYYY-MM-DD HH24:MI:SS') as updated_at
FROM users
WHERE email = 'mjjayaweera@gmail.com';

-- Record in migration_history
INSERT INTO migration_history (version, name, applied_at, success)
VALUES (24, 'grant_admin_privileges', NOW(), true)
ON CONFLICT (version) DO NOTHING;

SELECT 'Migration 024 applied successfully!' as status;
"@

$sql | docker exec -i $pgContainer psql -U urlshortener -d urlshortener_db

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ SUCCESS! Admin privileges granted!" -ForegroundColor Green
    Write-Host "`nVerify the output above:" -ForegroundColor Yellow
    Write-Host "  - is_admin should be 't' (true)" -ForegroundColor Yellow
    Write-Host "  - email_verified should be 't' (true)" -ForegroundColor Yellow
    Write-Host "  - is_active should be 't' (true)" -ForegroundColor Yellow
} else {
    Write-Host "`n❌ Migration failed!" -ForegroundColor Red
    Write-Host "Check the error messages above." -ForegroundColor Yellow
}

Write-Host "`nTo start the full application, run:" -ForegroundColor Cyan
Write-Host "  docker-compose -f docker/docker-compose.local.yml up" -ForegroundColor White
