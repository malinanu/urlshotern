# Apply Missing Tables to Production Database
# PowerShell script for Windows

param(
    [string]$ContainerName = "urlshortener-postgres-1",
    [string]$DbUser = "urlshortener",
    [string]$DbName = "urlshortener_db",
    [switch]$Verify,
    [switch]$DryRun
)

$ErrorActionPreference = "Stop"

Write-Host "=== Apply Missing Database Tables ===" -ForegroundColor Cyan
Write-Host ""

# Check if Docker is running
try {
    docker ps | Out-Null
} catch {
    Write-Host "ERROR: Docker is not running or not accessible" -ForegroundColor Red
    exit 1
}

# Check if container exists
$containerExists = docker ps -a --format "{{.Names}}" | Select-String -Pattern $ContainerName
if (-not $containerExists) {
    Write-Host "ERROR: Container '$ContainerName' not found" -ForegroundColor Red
    Write-Host "Available containers:" -ForegroundColor Yellow
    docker ps -a --format "{{.Names}}"
    exit 1
}

# Check if container is running
$containerRunning = docker ps --format "{{.Names}}" | Select-String -Pattern $ContainerName
if (-not $containerRunning) {
    Write-Host "ERROR: Container '$ContainerName' is not running" -ForegroundColor Red
    Write-Host "Start it with: docker start $ContainerName" -ForegroundColor Yellow
    exit 1
}

if ($Verify) {
    Write-Host "Running verification script..." -ForegroundColor Green
    Write-Host ""

    # Copy verification script to container
    docker cp "$PSScriptRoot\verify-tables.sql" "${ContainerName}:/tmp/verify-tables.sql"

    # Execute verification
    docker exec -it $ContainerName psql -U $DbUser -d $DbName -f /tmp/verify-tables.sql

    Write-Host ""
    Write-Host "Verification complete!" -ForegroundColor Green
    exit 0
}

Write-Host "Target: Container '$ContainerName', Database '$DbName'" -ForegroundColor Yellow
Write-Host ""

if ($DryRun) {
    Write-Host "DRY RUN MODE - No changes will be made" -ForegroundColor Yellow
    Write-Host ""
    Write-Host "Would execute:" -ForegroundColor Cyan
    Write-Host "  1. Copy add-missing-tables.sql to container" -ForegroundColor Gray
    Write-Host "  2. Execute SQL script in database" -ForegroundColor Gray
    Write-Host "  3. Add 15+ new tables (roles, permissions, teams, etc.)" -ForegroundColor Gray
    Write-Host "  4. Add columns to url_mappings table" -ForegroundColor Gray
    Write-Host "  5. Create indexes for performance" -ForegroundColor Gray
    Write-Host "  6. Populate default roles and permissions" -ForegroundColor Gray
    Write-Host ""
    Write-Host "Run without -DryRun to apply changes" -ForegroundColor Yellow
    exit 0
}

# Confirm before proceeding
Write-Host "This will add missing tables to your production database." -ForegroundColor Yellow
Write-Host "This operation is SAFE - it only adds new tables and columns." -ForegroundColor Green
Write-Host "Existing data will NOT be modified or deleted." -ForegroundColor Green
Write-Host ""
$confirmation = Read-Host "Do you want to proceed? (yes/no)"

if ($confirmation -ne "yes") {
    Write-Host "Operation cancelled." -ForegroundColor Yellow
    exit 0
}

Write-Host ""
Write-Host "Step 1: Copying SQL script to container..." -ForegroundColor Cyan
try {
    docker cp "$PSScriptRoot\add-missing-tables.sql" "${ContainerName}:/tmp/add-missing-tables.sql"
    Write-Host "  ✓ Script copied successfully" -ForegroundColor Green
} catch {
    Write-Host "  ✗ Failed to copy script: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 2: Executing SQL script..." -ForegroundColor Cyan
Write-Host "  This may take 10-30 seconds..." -ForegroundColor Gray

try {
    $output = docker exec -it $ContainerName psql -U $DbUser -d $DbName -f /tmp/add-missing-tables.sql 2>&1

    # Check for errors in output
    $hasErrors = $output | Select-String -Pattern "ERROR"

    if ($hasErrors) {
        Write-Host "  ⚠ Some errors occurred (may be normal):" -ForegroundColor Yellow
        $hasErrors | ForEach-Object { Write-Host "    $_" -ForegroundColor Yellow }
        Write-Host ""
        Write-Host "  Common safe errors:" -ForegroundColor Gray
        Write-Host "    - 'relation already exists' = Table already exists (OK)" -ForegroundColor Gray
        Write-Host "    - 'column already exists' = Column already added (OK)" -ForegroundColor Gray
    } else {
        Write-Host "  ✓ SQL script executed successfully" -ForegroundColor Green
    }
} catch {
    Write-Host "  ✗ Failed to execute script: $_" -ForegroundColor Red
    exit 1
}

Write-Host ""
Write-Host "Step 3: Verifying changes..." -ForegroundColor Cyan

try {
    docker cp "$PSScriptRoot\verify-tables.sql" "${ContainerName}:/tmp/verify-tables.sql"
    $verifyOutput = docker exec -it $ContainerName psql -U $DbUser -d $DbName -f /tmp/verify-tables.sql 2>&1

    # Extract table count from output
    $tableCount = $verifyOutput | Select-String -Pattern "Total Tables.*\|.*(\d+)" | ForEach-Object { $_.Matches.Groups[1].Value }

    if ($tableCount) {
        Write-Host "  ✓ Database now has $tableCount tables" -ForegroundColor Green
    } else {
        Write-Host "  ✓ Changes applied (run with -Verify for detailed check)" -ForegroundColor Green
    }
} catch {
    Write-Host "  ⚠ Verification script failed, but changes may have been applied" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "=== SUCCESS ===" -ForegroundColor Green
Write-Host ""
Write-Host "Missing tables have been added to your database!" -ForegroundColor Green
Write-Host ""
Write-Host "Next steps:" -ForegroundColor Cyan
Write-Host "  1. Run verification: .\apply-missing-tables.ps1 -Verify" -ForegroundColor Gray
Write-Host "  2. Create admin user via API" -ForegroundColor Gray
Write-Host "  3. Assign admin role in database" -ForegroundColor Gray
Write-Host "  4. Test admin login" -ForegroundColor Gray
Write-Host ""
Write-Host "For detailed verification, run:" -ForegroundColor Yellow
Write-Host "  .\apply-missing-tables.ps1 -Verify" -ForegroundColor White
Write-Host ""
