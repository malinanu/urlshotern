# Database Diagnostic and Fix Script (PowerShell)
# This script checks database state and fixes missing tables/columns

param(
    [string]$DBContainer = "postgresql-central",
    [string]$DBUser = "urlshortener_user",
    [string]$DBName = "urlshortener_db",
    [string]$AppContainer = "urlshortener-app"
)

Write-Host "🔍 Database Diagnostic and Fix Script" -ForegroundColor Cyan
Write-Host "========================================" -ForegroundColor Cyan
Write-Host ""

# Function to run SQL query
function Run-SQL {
    param([string]$Query)
    $result = docker exec -i $DBContainer psql -U $DBUser -d $DBName -t -c "$Query" 2>$null
    return $result.Trim()
}

# Function to check if table exists
function Test-TableExists {
    param([string]$TableName)
    $result = Run-SQL "SELECT EXISTS (SELECT FROM information_schema.tables WHERE table_schema = 'public' AND table_name = '$TableName');"
    return $result -eq "t"
}

# Function to check if column exists
function Test-ColumnExists {
    param([string]$TableName, [string]$ColumnName)
    $result = Run-SQL "SELECT EXISTS (SELECT FROM information_schema.columns WHERE table_schema = 'public' AND table_name = '$TableName' AND column_name = '$ColumnName');"
    return $result -eq "t"
}

Write-Host "📊 Step 1: Checking database connectivity..." -ForegroundColor Yellow
try {
    docker exec -i $DBContainer psql -U $DBUser -d $DBName -c '\q' 2>$null | Out-Null
    Write-Host "✅ Database is accessible" -ForegroundColor Green
} catch {
    Write-Host "❌ Cannot connect to database" -ForegroundColor Red
    Write-Host "   Please check database container is running: docker ps | grep postgresql" -ForegroundColor Yellow
    exit 1
}
Write-Host ""

Write-Host "🔍 Step 2: Checking table existence..." -ForegroundColor Yellow
$TablesOK = $true

# Check api_keys table
if (Test-TableExists "api_keys") {
    Write-Host "✅ api_keys table exists" -ForegroundColor Green
} else {
    Write-Host "❌ api_keys table is MISSING" -ForegroundColor Red
    $TablesOK = $false
}

# Check api_key_usage table
if (Test-TableExists "api_key_usage") {
    Write-Host "✅ api_key_usage table exists" -ForegroundColor Green
} else {
    Write-Host "❌ api_key_usage table is MISSING" -ForegroundColor Red
    $TablesOK = $false
}

# Check url_mappings table
if (Test-TableExists "url_mappings") {
    Write-Host "✅ url_mappings table exists" -ForegroundColor Green

    # Check user_id column
    if (Test-ColumnExists "url_mappings" "user_id") {
        Write-Host "✅ url_mappings.user_id column exists" -ForegroundColor Green
    } else {
        Write-Host "❌ url_mappings.user_id column is MISSING" -ForegroundColor Red
        $TablesOK = $false
    }
} else {
    Write-Host "❌ url_mappings table is MISSING" -ForegroundColor Red
    $TablesOK = $false
}

# Check users table
if (Test-TableExists "users") {
    Write-Host "✅ users table exists" -ForegroundColor Green
} else {
    Write-Host "❌ users table is MISSING" -ForegroundColor Red
    $TablesOK = $false
}

Write-Host ""

if ($TablesOK) {
    Write-Host "✅ All required tables and columns exist!" -ForegroundColor Green
    Write-Host ""
    Write-Host "📊 Database schema is correct. If you're still experiencing issues:" -ForegroundColor Cyan
    Write-Host "   1. Check application logs: docker logs $AppContainer --tail 100"
    Write-Host "   2. Verify environment variables are set correctly"
    Write-Host "   3. Check if JWT_SECRET is configured"
    exit 0
}

Write-Host "⚠️  Database schema issues detected!" -ForegroundColor Yellow
Write-Host ""
Write-Host "🔧 Step 3: Attempting to fix database schema..." -ForegroundColor Yellow
Write-Host ""

# Option 1: Try running migrations
Write-Host "📝 Option 1: Running migrations..." -ForegroundColor Yellow
$migrationOutput = docker exec $AppContainer /app/scripts/run-migrations.sh 2>&1
if ($migrationOutput -match "Migration complete") {
    Write-Host "✅ Migrations completed successfully" -ForegroundColor Green
    Write-Host ""
    Write-Host "🎉 Database should now be fixed! Please verify:" -ForegroundColor Green
    Write-Host "   docker exec -i $DBContainer psql -U $DBUser -d $DBName -c '\dt'"
    exit 0
} else {
    Write-Host "⚠️  Migration script had issues or is not available" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "📝 Option 2: Restarting app container to trigger migrations..." -ForegroundColor Yellow
docker restart $AppContainer
Write-Host "⏳ Waiting for container to restart (10 seconds)..."
Start-Sleep -Seconds 10

# Check logs
Write-Host "📋 Checking startup logs..." -ForegroundColor Yellow
docker logs $AppContainer --tail 50 | Select-String -Pattern "(Migration|🔧|✅|❌)"

Write-Host ""
Write-Host "🔍 Re-checking database schema..." -ForegroundColor Yellow
$Fixed = $true

if (-not (Test-TableExists "api_keys")) {
    $Fixed = $false
}
if (-not (Test-TableExists "api_key_usage")) {
    $Fixed = $false
}
if (-not (Test-ColumnExists "url_mappings" "user_id")) {
    $Fixed = $false
}

if ($Fixed) {
    Write-Host "✅ Database schema is now correct!" -ForegroundColor Green
    exit 0
}

Write-Host ""
Write-Host "⚠️  Automatic fixes didn't resolve all issues" -ForegroundColor Yellow
Write-Host ""
Write-Host "📝 Manual fix required. Please run these commands:" -ForegroundColor Cyan
Write-Host ""
Write-Host "# View current tables:"
Write-Host "docker exec -i $DBContainer psql -U $DBUser -d $DBName -c '\dt'"
Write-Host ""
Write-Host "# Check migration status:"
Write-Host "docker exec -i $DBContainer psql -U $DBUser -d $DBName -c 'SELECT * FROM schema_migrations ORDER BY version;'"
Write-Host ""
Write-Host "# Manual migration (if needed):"
Write-Host "docker exec $AppContainer /app/scripts/run-migrations.sh"
Write-Host ""
Write-Host "# Or restart the app container:"
Write-Host "docker-compose -f docker-compose.oracle.yml restart app"
Write-Host ""
