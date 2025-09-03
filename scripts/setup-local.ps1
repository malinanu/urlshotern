# PowerShell script for Windows 11 setup
Write-Host "Setting up URL Shortener local development environment..." -ForegroundColor Green

# Check if Docker is running
Write-Host "Checking Docker..." -ForegroundColor Blue
try {
    docker --version | Out-Null
    Write-Host "✓ Docker is available" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not available. Please install Docker Desktop." -ForegroundColor Red
    exit 1
}

try {
    docker info | Out-Null
    Write-Host "✓ Docker is running" -ForegroundColor Green
} catch {
    Write-Host "❌ Docker is not running. Please start Docker Desktop." -ForegroundColor Red
    exit 1
}

# Copy environment file
if (!(Test-Path ".env")) {
    Copy-Item ".env.example" ".env"
    Write-Host "✓ Created .env file from example" -ForegroundColor Green
} else {
    Write-Host "✓ .env file already exists" -ForegroundColor Yellow
}

# Start Docker services
Write-Host "Starting Docker services..." -ForegroundColor Blue
try {
    docker-compose -f docker/docker-compose.local.yml up -d
    Write-Host "✓ Docker services started successfully" -ForegroundColor Green
} catch {
    Write-Host "❌ Failed to start Docker services" -ForegroundColor Red
    Write-Host "Error details: $_" -ForegroundColor Red
    exit 1
}

# Wait for services to be ready
Write-Host "Waiting for services to be ready..." -ForegroundColor Blue
Start-Sleep -Seconds 15

# Simple health check
Write-Host "Checking service health..." -ForegroundColor Blue
$postgresRunning = docker-compose -f docker/docker-compose.local.yml ps -q postgres
$redisRunning = docker-compose -f docker/docker-compose.local.yml ps -q redis

if ($postgresRunning) {
    Write-Host "✓ PostgreSQL container is running" -ForegroundColor Green
} else {
    Write-Host "⚠️  PostgreSQL container may not be running" -ForegroundColor Yellow
}

if ($redisRunning) {
    Write-Host "✓ Redis container is running" -ForegroundColor Green
} else {
    Write-Host "⚠️  Redis container may not be running" -ForegroundColor Yellow
}

# Check if Go is installed
Write-Host "Checking Go installation..." -ForegroundColor Blue
try {
    $goVersion = go version
    Write-Host "✓ Go is available: $goVersion" -ForegroundColor Green
    
    # Install Go dependencies if go.mod exists
    if (Test-Path "go.mod") {
        Write-Host "Installing Go dependencies..." -ForegroundColor Blue
        go mod tidy
        go mod download
        Write-Host "✓ Go dependencies installed" -ForegroundColor Green
    }
} catch {
    Write-Host "⚠️  Go is not installed. You'll need Go to run the server locally." -ForegroundColor Yellow
    Write-Host "   Install Go from: https://golang.org/dl/" -ForegroundColor Yellow
}

Write-Host ""
Write-Host "✅ Local setup complete!" -ForegroundColor Green
Write-Host ""
Write-Host "📊 Services available:" -ForegroundColor Cyan
Write-Host "   PostgreSQL: localhost:5432" -ForegroundColor White
Write-Host "   Redis: localhost:6379" -ForegroundColor White
Write-Host "   Adminer (DB Admin): http://localhost:8081" -ForegroundColor White
Write-Host ""
Write-Host "🚀 To start the server:" -ForegroundColor Cyan
Write-Host "   go run cmd/server/main.go" -ForegroundColor White
Write-Host ""
Write-Host "🧪 To test the API (PowerShell):" -ForegroundColor Cyan
Write-Host "   # Shorten URL" -ForegroundColor White
Write-Host '   Invoke-RestMethod -Uri "http://localhost:8080/api/v1/shorten" -Method Post -ContentType "application/json" -Body ''{"url":"https://www.google.com"}''' -ForegroundColor Gray
Write-Host ""
Write-Host "   # Health check" -ForegroundColor White
Write-Host '   Invoke-RestMethod -Uri "http://localhost:8080/health"' -ForegroundColor Gray
Write-Host ""
Write-Host "📝 Adminer login details:" -ForegroundColor Cyan
Write-Host "   System: PostgreSQL" -ForegroundColor White
Write-Host "   Server: postgres" -ForegroundColor White
Write-Host "   Username: urlshortener" -ForegroundColor White
Write-Host "   Password: password" -ForegroundColor White
Write-Host "   Database: urlshortener_db" -ForegroundColor White
Write-Host ""
Write-Host "🛑 To stop services:" -ForegroundColor Cyan
Write-Host "   docker-compose -f docker/docker-compose.local.yml down" -ForegroundColor White