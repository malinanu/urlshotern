# Database Migration Runner (PowerShell)
# Automatically applies pending migrations in order

param(
    [string]$MigrationsDir = "configs\migrations",
    [string]$DbHost = "localhost",
    [string]$DbPort = "5432",
    [string]$DbUser = "urlshortener",
    [string]$DbName = "urlshortener_db",
    [string]$DbPassword = "securepassword"
)

Write-Host "🔄 Starting database migrations..." -ForegroundColor Cyan
Write-Host "📁 Migrations directory: $MigrationsDir"
Write-Host "🗄️  Database: $DbName on ${DbHost}:${DbPort}"

# Set environment variable for PostgreSQL password
$env:PGPASSWORD = $DbPassword

# Wait for database to be ready
Write-Host "⏳ Waiting for database to be ready..." -ForegroundColor Yellow
$maxAttempts = 30
$attempt = 0

while ($attempt -lt $maxAttempts) {
    try {
        $result = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -c '\q' 2>&1
        if ($LASTEXITCODE -eq 0) {
            Write-Host "✅ Database is ready!" -ForegroundColor Green
            break
        }
    }
    catch {
        # Database not ready
    }

    $attempt++
    Write-Host "   Database is unavailable - sleeping (attempt $attempt/$maxAttempts)"
    Start-Sleep -Seconds 2
}

if ($attempt -eq $maxAttempts) {
    Write-Host "❌ Database failed to become ready!" -ForegroundColor Red
    exit 1
}

# Create schema_migrations table if it doesn't exist
Write-Host "📋 Ensuring schema_migrations table exists..."
$createTableSql = @"
CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(14) PRIMARY KEY,
    applied_at TIMESTAMP DEFAULT CURRENT_TIMESTAMP
);
"@

$createTableSql | & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName

# Get list of applied migrations
$appliedMigrations = & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName -t -c "SELECT version FROM schema_migrations ORDER BY version;" | ForEach-Object { $_.Trim() }

# Find all *_up.sql migration files
$migrationFiles = Get-ChildItem -Path $MigrationsDir -Filter "*_up.sql" | Sort-Object Name

$migrationsApplied = 0
$migrationsSkipped = 0

foreach ($file in $migrationFiles) {
    # Extract version number from filename
    if ($file.Name -match '^(\d+)_') {
        $version = $matches[1]
    }
    else {
        continue
    }

    # Check if migration has already been applied
    if ($appliedMigrations -contains $version) {
        Write-Host "⏭️  Migration $version already applied, skipping..." -ForegroundColor Gray
        $migrationsSkipped++
        continue
    }

    Write-Host "🔧 Applying migration $version: $($file.Name)" -ForegroundColor Yellow

    # Apply the migration
    $migrationContent = Get-Content $file.FullName -Raw
    $migrationContent | & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName

    if ($LASTEXITCODE -eq 0) {
        # Record the migration
        "INSERT INTO schema_migrations (version) VALUES ('$version') ON CONFLICT DO NOTHING;" | & psql -h $DbHost -p $DbPort -U $DbUser -d $DbName
        Write-Host "   ✅ Migration $version applied successfully" -ForegroundColor Green
        $migrationsApplied++
    }
    else {
        Write-Host "   ❌ Migration $version failed!" -ForegroundColor Red
        exit 1
    }
}

Write-Host ""
Write-Host "🎉 Migration complete!" -ForegroundColor Green
Write-Host "   Applied: $migrationsApplied"
Write-Host "   Skipped: $migrationsSkipped"
Write-Host ""

# Clear password
Remove-Item Env:\PGPASSWORD
