# Quick script to grant admin privileges to mjjayaweera@gmail.com

Write-Host "Applying admin privileges migration..." -ForegroundColor Cyan

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
    updated_at
FROM users
WHERE email = 'mjjayaweera@gmail.com';

-- Record in migration_history if table exists
INSERT INTO migration_history (version, name, applied_at, success)
VALUES (24, 'grant_admin_privileges', NOW(), true)
ON CONFLICT (version) DO NOTHING;
"@

$sql | docker exec -i urlshorter-postgres-1 psql -U urlshortener -d urlshortener_db

if ($LASTEXITCODE -eq 0) {
    Write-Host "`n✅ Admin privileges granted successfully!" -ForegroundColor Green
    Write-Host "Verify by checking the output above - is_admin should be 't' (true)" -ForegroundColor Yellow
} else {
    Write-Host "`n❌ Migration failed. Make sure Docker containers are running." -ForegroundColor Red
    Write-Host "Start Docker with: docker-compose -f docker/docker-compose.local.yml up" -ForegroundColor Yellow
}
