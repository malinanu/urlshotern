# Production Fix Script for 502 Bad Gateway Errors (Windows PowerShell)
# This script applies the missing migrations and restarts the backend

param(
    [string]$DBHost = "localhost",
    [string]$DBUser = "urlshortener",
    [string]$DBName = "urlshortener_db",
    [string]$DBPassword = "urlshortener"
)

Write-Host "==========================================" -ForegroundColor Yellow
Write-Host "URLshortener 502 Fix Script (Windows)" -ForegroundColor Yellow
Write-Host "==========================================" -ForegroundColor Yellow
Write-Host ""

# Function to run SQL file
function Run-Migration {
    param(
        [string]$MigrationFile,
        [string]$MigrationVersion
    )

    Write-Host "Applying migration $MigrationVersion..." -ForegroundColor Yellow

    if (-not (Test-Path $MigrationFile)) {
        Write-Host "Error: Migration file not found: $MigrationFile" -ForegroundColor Red
        return $false
    }

    # Set password environment variable
    $env:PGPASSWORD = $DBPassword

    # Apply migration
    & psql -h $DBHost -U $DBUser -d $DBName -f $MigrationFile
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error applying migration $MigrationVersion" -ForegroundColor Red
        return $false
    }

    # Record in schema_migrations
    $sql = @"
INSERT INTO schema_migrations (version, executed_at)
VALUES ('$MigrationVersion', NOW())
ON CONFLICT (version) DO NOTHING;
"@

    $sql | & psql -h $DBHost -U $DBUser -d $DBName
    if ($LASTEXITCODE -ne 0) {
        Write-Host "Error recording migration $MigrationVersion" -ForegroundColor Red
        return $false
    }

    Write-Host "✓ Migration $MigrationVersion applied successfully" -ForegroundColor Green
    Write-Host ""
    return $true
}

# Step 1: Check if we're in the right directory
if (-not (Test-Path "configs\migrations")) {
    Write-Host "Error: Not in URLshorter directory" -ForegroundColor Red
    Write-Host "Please cd to the URLshorter directory first" -ForegroundColor Yellow
    exit 1
}

Write-Host "Step 1: Checking current migration status..." -ForegroundColor Yellow
$env:PGPASSWORD = $DBPassword
& psql -h $DBHost -U $DBUser -d $DBName -c "SELECT version, executed_at FROM schema_migrations ORDER BY version;"
if ($LASTEXITCODE -ne 0) {
    Write-Host "Error: Cannot connect to database" -ForegroundColor Red
    exit 1
}
Write-Host ""

# Step 2: Check which migrations need to be applied
Write-Host "Step 2: Checking which migrations to apply..." -ForegroundColor Yellow

$needs017 = $false
$needs025 = $false
$needs027 = $false

# Check if migrations are already applied
$env:PGPASSWORD = $DBPassword

$check017 = & psql -h $DBHost -U $DBUser -d $DBName -t -c "SELECT version FROM schema_migrations WHERE version = '017';" 2>$null
if ($check017 -notmatch "017") {
    $needs017 = $true
    Write-Host "  → Migration 017 (CMS tables) needs to be applied" -ForegroundColor Cyan
}

$check025 = & psql -h $DBHost -U $DBUser -d $DBName -t -c "SELECT version FROM schema_migrations WHERE version = '025';" 2>$null
if ($check025 -notmatch "025") {
    $needs025 = $true
    Write-Host "  → Migration 025 (Activity logs) needs to be applied" -ForegroundColor Cyan
}

$check027 = & psql -h $DBHost -U $DBUser -d $DBName -t -c "SELECT version FROM schema_migrations WHERE version = '027';" 2>$null
if ($check027 -notmatch "027") {
    $needs027 = $true
    Write-Host "  → Migration 027 (SEO tables) needs to be applied" -ForegroundColor Cyan
}

if (-not $needs017 -and -not $needs025 -and -not $needs027) {
    Write-Host "All migrations already applied!" -ForegroundColor Green
    Write-Host "The 502 error might be caused by something else." -ForegroundColor Yellow
    Write-Host "Check backend logs for more details." -ForegroundColor Yellow
    exit 0
}
Write-Host ""

# Step 3: Apply migrations
Write-Host "Step 3: Applying migrations..." -ForegroundColor Yellow

if ($needs017) {
    Run-Migration -MigrationFile "configs\migrations\017_cms_tables_up.sql" -MigrationVersion "017"
}

if ($needs025) {
    Run-Migration -MigrationFile "configs\migrations\025_create_activity_logs_up.sql" -MigrationVersion "025"
}

if ($needs027) {
    Run-Migration -MigrationFile "configs\migrations\027_seo_tables_up.sql" -MigrationVersion "027"
}

# Step 4: Verify tables were created
Write-Host "Step 4: Verifying tables were created..." -ForegroundColor Yellow
$env:PGPASSWORD = $DBPassword
& psql -h $DBHost -U $DBUser -d $DBName -c @"
SELECT table_name
FROM information_schema.tables
WHERE table_schema = 'public'
AND table_name IN (
    'static_pages', 'content_blocks', 'page_templates',
    'meta_tags', 'seo_analyses', 'activity_logs',
    'global_settings', 'robots_txt', 'url_redirects'
)
ORDER BY table_name;
"@
Write-Host ""

# Step 5: Restart backend (Windows-specific)
Write-Host "Step 5: Restarting backend service..." -ForegroundColor Yellow

# Try to find and restart the Go backend process
$serverProcess = Get-Process | Where-Object { $_.ProcessName -like "*server*" -or $_.CommandLine -like "*URLshorter*" }

if ($serverProcess) {
    Write-Host "Stopping existing backend process..." -ForegroundColor Cyan
    Stop-Process -Name "server" -Force -ErrorAction SilentlyContinue
    Start-Sleep -Seconds 2

    Write-Host "Please restart the backend manually:" -ForegroundColor Yellow
    Write-Host "  cd D:\github\URLshorter" -ForegroundColor Cyan
    Write-Host "  go run cmd\server\main.go" -ForegroundColor Cyan
} else {
    Write-Host "No backend process found. Please start it manually:" -ForegroundColor Yellow
    Write-Host "  cd D:\github\URLshorter" -ForegroundColor Cyan
    Write-Host "  go run cmd\server\main.go" -ForegroundColor Cyan
}
Write-Host ""

# Step 6: Verification instructions
Write-Host "Step 6: Verification" -ForegroundColor Yellow
Write-Host ""
Write-Host "After restarting the backend, test the admin endpoints:" -ForegroundColor Cyan
Write-Host "  1. https://trunc.site/api/v1/admin/seo/analyses" -ForegroundColor White
Write-Host "  2. https://trunc.site/api/v1/admin/cms/pages" -ForegroundColor White
Write-Host "  3. https://trunc.site/api/v1/admin/content-blocks" -ForegroundColor White
Write-Host ""

Write-Host "==========================================" -ForegroundColor Green
Write-Host "Migration Fix Complete!" -ForegroundColor Green
Write-Host "==========================================" -ForegroundColor Green
Write-Host ""
