# ============================================================================
# Database Migration Script (PowerShell)
# ============================================================================
# Description: Manages database migrations for the URL Shortener project
#
# Usage:
#   .\scripts\migrate.ps1 status                    # Show migration status
#   .\scripts\migrate.ps1 up 001                    # Apply migration 001
#   .\scripts\migrate.ps1 down 005                  # Rollback migration 005
#   .\scripts\migrate.ps1 rollback-to 003           # Rollback to version 003
#
# Environment Variables:
#   $env:DB_HOST     - Database host (default: localhost)
#   $env:DB_PORT     - Database port (default: 5432)
#   $env:DB_NAME     - Database name (default: urlshortener_db)
#   $env:DB_USER     - Database user (default: urlshortener)
#   $env:DB_PASSWORD - Database password
#   $env:ENVIRONMENT - Environment name (production, staging, development)
# ============================================================================

param(
    [Parameter(Mandatory=$true, Position=0)]
    [string]$Command,

    [Parameter(Position=1)]
    [string]$Version,

    [Parameter(Position=2)]
    [string]$Direction
)

# Configuration
$DB_HOST = if ($env:DB_HOST) { $env:DB_HOST } else { "localhost" }
$DB_PORT = if ($env:DB_PORT) { $env:DB_PORT } else { "5432" }
$DB_NAME = if ($env:DB_NAME) { $env:DB_NAME } else { "urlshortener_db" }
$DB_USER = if ($env:DB_USER) { $env:DB_USER } else { "urlshortener" }
$DB_PASSWORD = $env:DB_PASSWORD
$ENVIRONMENT = if ($env:ENVIRONMENT) { $env:ENVIRONMENT } else { "development" }

$MIGRATIONS_DIR = "configs\migrations"
$SCRIPT_DIR = Split-Path -Parent $MyInvocation.MyCommand.Path
$PROJECT_ROOT = Split-Path -Parent $SCRIPT_DIR

# Color output functions
function Write-Info {
    param([string]$Message)
    Write-Host "ℹ $Message" -ForegroundColor Blue
}

function Write-Success {
    param([string]$Message)
    Write-Host "✓ $Message" -ForegroundColor Green
}

function Write-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

function Write-Error-Custom {
    param([string]$Message)
    Write-Host "✗ $Message" -ForegroundColor Red
}

function Test-Prerequisites {
    # Check if psql is available
    try {
        $null = Get-Command psql -ErrorAction Stop
    } catch {
        Write-Error-Custom "psql command not found. Please install PostgreSQL client."
        exit 1
    }

    # Test database connection
    $env:PGPASSWORD = $DB_PASSWORD
    $testQuery = "SELECT 1"
    $result = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $testQuery 2>$null

    if ($LASTEXITCODE -ne 0) {
        Write-Error-Custom "Cannot connect to database. Please check connection settings."
        Write-Info "Connection: $DB_USER@${DB_HOST}:${DB_PORT}/$DB_NAME"
        exit 1
    }

    # Check if migrations directory exists
    if (-not (Test-Path "$PROJECT_ROOT\$MIGRATIONS_DIR")) {
        Write-Error-Custom "Migrations directory not found: $PROJECT_ROOT\$MIGRATIONS_DIR"
        exit 1
    }
}

function Initialize-SchemaMigrationsTable {
    $env:PGPASSWORD = $DB_PASSWORD
    $checkQuery = "SELECT EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'schema_migrations');"
    $exists = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $checkQuery

    if ($exists.Trim() -eq 'f') {
        Write-Warning "schema_migrations table does not exist. Creating it..."
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f "$PROJECT_ROOT\$MIGRATIONS_DIR\000_schema_migrations.sql"
        Write-Success "schema_migrations table created"
    }
}

function Get-MigrationStatus {
    Initialize-SchemaMigrationsTable

    Write-Host ""
    Write-Host "Migration Status:" -ForegroundColor Cyan
    Write-Host "==================" -ForegroundColor Cyan
    Write-Host ""

    $env:PGPASSWORD = $DB_PASSWORD

    # List all migration files and their status
    Get-ChildItem "$PROJECT_ROOT\$MIGRATIONS_DIR\*_up.sql" | ForEach-Object {
        $filename = $_.Name
        $version = $filename -replace '(\d+)_.*', '$1'

        # Check if migration is applied
        $query = "SELECT applied_at FROM schema_migrations WHERE version = '$version';"
        $applied = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $query

        if ($applied.Trim()) {
            Write-Host "✓ $version - Applied at $($applied.Trim())" -ForegroundColor Green
        } else {
            Write-Host "○ $version - Not applied" -ForegroundColor Yellow
        }
    }
    Write-Host ""
}

function Invoke-ApplyMigration {
    param([string]$Version)

    $migrationFiles = Get-ChildItem "$PROJECT_ROOT\$MIGRATIONS_DIR\${Version}_*_up.sql"

    if ($migrationFiles.Count -eq 0) {
        Write-Error-Custom "Migration file not found: ${Version}_*_up.sql"
        exit 1
    }

    $actualFile = $migrationFiles[0].FullName

    # Check if already applied
    $env:PGPASSWORD = $DB_PASSWORD
    $query = "SELECT version FROM schema_migrations WHERE version = '$Version';"
    $isApplied = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $query

    if ($isApplied.Trim()) {
        Write-Warning "Migration $Version is already applied"
        return
    }

    Write-Info "Applying migration $Version..."

    # Calculate checksum
    $checksum = (Get-FileHash $actualFile -Algorithm SHA256).Hash
    $description = [System.IO.Path]::GetFileNameWithoutExtension($actualFile) -replace '^[0-9]+_', '' -replace '_up$', ''

    # Execute migration and track timing
    $startTime = Get-Date

    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $actualFile

    if ($LASTEXITCODE -eq 0) {
        $endTime = Get-Date
        $executionTime = [int](($endTime - $startTime).TotalMilliseconds)

        # Record in schema_migrations table
        $insertQuery = "INSERT INTO schema_migrations (version, execution_time_ms, description, checksum) VALUES ('$Version', $executionTime, '$description', '$checksum');"
        psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -c $insertQuery

        Write-Success "Migration $Version applied successfully (${executionTime}ms)"
    } else {
        Write-Error-Custom "Migration $Version failed"
        exit 1
    }
}

function Invoke-RollbackMigration {
    param([string]$Version)

    $migrationFiles = Get-ChildItem "$PROJECT_ROOT\$MIGRATIONS_DIR\${Version}_*_down.sql"

    if ($migrationFiles.Count -eq 0) {
        Write-Error-Custom "Rollback file not found: ${Version}_*_down.sql"
        exit 1
    }

    $actualFile = $migrationFiles[0].FullName

    # Check if migration is applied
    $env:PGPASSWORD = $DB_PASSWORD
    $query = "SELECT version FROM schema_migrations WHERE version = '$Version';"
    $isApplied = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $query

    if (-not $isApplied.Trim()) {
        Write-Warning "Migration $Version is not applied, cannot rollback"
        return
    }

    # Production safety check
    if ($ENVIRONMENT -eq "production") {
        Write-Warning "You are about to ROLLBACK a migration in PRODUCTION!"
        Write-Warning "This may result in DATA LOSS. Ensure you have a recent backup."
        $timestamp = Get-Date -Format "yyyyMMdd_HHmmss"
        Write-Info "Backup command: pg_dump -U $DB_USER -h $DB_HOST -d $DB_NAME > backup_$timestamp.sql"
        Write-Host ""
        $confirm = Read-Host "Type 'YES I HAVE A BACKUP' to continue"
        if ($confirm -ne "YES I HAVE A BACKUP") {
            Write-Info "Rollback cancelled"
            exit 0
        }
    }

    Write-Info "Rolling back migration $Version..."

    psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -f $actualFile

    if ($LASTEXITCODE -eq 0) {
        Write-Success "Migration $Version rolled back successfully"
    } else {
        Write-Error-Custom "Rollback of migration $Version failed"
        exit 1
    }
}

function Invoke-RollbackToVersion {
    param([string]$TargetVersion)

    # Get all applied migrations
    $env:PGPASSWORD = $DB_PASSWORD
    $query = "SELECT version FROM schema_migrations ORDER BY version DESC;"
    $appliedMigrations = psql -h $DB_HOST -p $DB_PORT -U $DB_USER -d $DB_NAME -t -c $query

    $migrationsToRollback = @()

    foreach ($line in $appliedMigrations -split "`n") {
        $version = $line.Trim()
        if ($version -and ([int]$version -gt [int]$TargetVersion)) {
            $migrationsToRollback += $version
        }
    }

    if ($migrationsToRollback.Count -eq 0) {
        Write-Info "Already at version $TargetVersion or lower"
        return
    }

    Write-Info "Will rollback $($migrationsToRollback.Count) migration(s): $($migrationsToRollback -join ', ')"

    foreach ($version in $migrationsToRollback) {
        Invoke-RollbackMigration $version
    }

    Write-Success "Rolled back to version $TargetVersion"
}

# Main script execution

switch ($Command.ToLower()) {
    "status" {
        Test-Prerequisites
        Get-MigrationStatus
    }

    "up" {
        if (-not $Version) {
            Write-Error-Custom "Missing version number"
            Write-Host "Usage: .\scripts\migrate.ps1 up <version>"
            exit 1
        }
        Test-Prerequisites
        Initialize-SchemaMigrationsTable
        Invoke-ApplyMigration $Version
    }

    "down" {
        if (-not $Version) {
            Write-Error-Custom "Missing version number"
            Write-Host "Usage: .\scripts\migrate.ps1 down <version>"
            exit 1
        }
        Test-Prerequisites
        Initialize-SchemaMigrationsTable
        Invoke-RollbackMigration $Version
    }

    "rollback-to" {
        if (-not $Version) {
            Write-Error-Custom "Missing version number"
            Write-Host "Usage: .\scripts\migrate.ps1 rollback-to <version>"
            exit 1
        }
        Test-Prerequisites
        Initialize-SchemaMigrationsTable
        Invoke-RollbackToVersion $Version
    }

    default {
        Write-Error-Custom "Unknown command: $Command"
        Write-Host ""
        Write-Host "Usage: .\scripts\migrate.ps1 <command> [version]" -ForegroundColor Cyan
        Write-Host ""
        Write-Host "Commands:" -ForegroundColor Cyan
        Write-Host "  status                    Show which migrations have been applied"
        Write-Host "  up <version>              Apply a specific migration"
        Write-Host "  down <version>            Rollback a specific migration"
        Write-Host "  rollback-to <version>     Rollback to a specific version"
        exit 1
    }
}
