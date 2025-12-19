# URL Shortener Observability Stack Startup Script (Windows)
# This script starts the complete observability stack with health checks

$ErrorActionPreference = "Stop"

Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  URL Shortener - Observability Stack Deployment" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""

function Write-Success {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Green
}

function Write-Error-Message {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Red
}

function Write-Warning-Message {
    param([string]$Message)
    Write-Host $Message -ForegroundColor Yellow
}

# Check if Docker is running
Write-Host "Checking Docker... " -NoNewline
try {
    docker info | Out-Null
    Write-Success "OK"
} catch {
    Write-Error-Message "FAILED"
    Write-Error-Message "Error: Docker is not running. Please start Docker Desktop and try again."
    exit 1
}

# Check if Docker Compose is available
Write-Host "Checking Docker Compose... " -NoNewline
try {
    docker-compose --version | Out-Null
    Write-Success "OK"
} catch {
    Write-Error-Message "FAILED"
    Write-Error-Message "Error: Docker Compose is not installed."
    exit 1
}

# Check if .env file exists
Write-Host "Checking environment file... " -NoNewline
if (-not (Test-Path ".env")) {
    Write-Warning-Message "WARNING"
    Write-Warning-Message "Warning: .env file not found."
    if (Test-Path ".env.example") {
        Copy-Item ".env.example" ".env"
        Write-Host "Created .env from .env.example"
        Write-Host "Please edit .env and set required variables (SLACK_WEBHOOK_URL, PAGERDUTY_SERVICE_KEY, etc.)"
        Write-Host "Then run this script again."
        exit 1
    } else {
        Write-Error-Message "Error: .env.example not found. Please create .env manually."
        exit 1
    }
} else {
    Write-Success "OK"
}

# Create required directories
Write-Host "Creating directories..."
New-Item -ItemType Directory -Force -Path "observability\grafana\dashboards" | Out-Null
New-Item -ItemType Directory -Force -Path "observability\prometheus\data" | Out-Null
New-Item -ItemType Directory -Force -Path "observability\loki\data" | Out-Null
Write-Success "Directories created"

Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Starting Observability Stack" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""

# Stop any existing containers
Write-Host "Stopping existing containers..."
docker-compose -f docker/docker-compose.observability.yml down 2>$null

# Start the stack
Write-Host "Starting services..."
docker-compose -f docker/docker-compose.observability.yml up -d

Write-Host ""
Write-Host "Waiting for services to start..."
Start-Sleep -Seconds 10

# Health check function
function Test-Service {
    param(
        [string]$ServiceName,
        [string]$Url,
        [int]$MaxAttempts = 30
    )

    Write-Host "Checking $ServiceName... " -NoNewline

    for ($attempt = 1; $attempt -le $MaxAttempts; $attempt++) {
        try {
            $response = Invoke-WebRequest -Uri $Url -UseBasicParsing -TimeoutSec 2 -ErrorAction SilentlyContinue
            if ($response.StatusCode -eq 200) {
                Write-Success "OK"
                return $true
            }
        } catch {
            # Continue trying
        }
        Start-Sleep -Seconds 2
    }

    Write-Error-Message "FAILED"
    return $false
}

Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Service Health Checks" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""

# Check each service
Test-Service -ServiceName "Application" -Url "http://localhost:8080/health"
Test-Service -ServiceName "Metrics Endpoint" -Url "http://localhost:8080/metrics"
Test-Service -ServiceName "Prometheus" -Url "http://localhost:9090/-/healthy"
Test-Service -ServiceName "Grafana" -Url "http://localhost:3000/api/health"
Test-Service -ServiceName "Alertmanager" -Url "http://localhost:9093/-/healthy"
Test-Service -ServiceName "Jaeger" -Url "http://localhost:16686/"
Test-Service -ServiceName "Loki" -Url "http://localhost:3100/ready"

Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Deployment Summary" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""

# Get container status
Write-Host "Container Status:"
docker-compose -f docker/docker-compose.observability.yml ps

Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Access URLs" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Application:        http://localhost:8080" -ForegroundColor White
Write-Host "  Metrics Endpoint:   http://localhost:8080/metrics" -ForegroundColor White
Write-Host ""
Write-Host "  Grafana:            http://localhost:3000" -ForegroundColor Yellow
Write-Host "  Prometheus:         http://localhost:9090" -ForegroundColor Yellow
Write-Host "  Alertmanager:       http://localhost:9093" -ForegroundColor Yellow
Write-Host "  Jaeger:             http://localhost:16686" -ForegroundColor Yellow
Write-Host "  Loki:               http://localhost:3100" -ForegroundColor Yellow
Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Credentials" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  Grafana:" -ForegroundColor White
Write-Host "    Username: admin" -ForegroundColor White
Write-Host "    Password: (check .env file - GRAFANA_PASSWORD)" -ForegroundColor White
Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Next Steps" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  1. Open Grafana: http://localhost:3000" -ForegroundColor White
Write-Host "  2. Navigate to Dashboards -> Browse" -ForegroundColor White
Write-Host "  3. Open 'Service Health Overview' dashboard" -ForegroundColor White
Write-Host "  4. Generate some test traffic:" -ForegroundColor White
Write-Host ""
Write-Host "     for (\$i = 1; \$i -le 100; \$i++) {" -ForegroundColor Gray
Write-Host "       Invoke-RestMethod -Method Post -Uri 'http://localhost:8080/api/v1/shorten' `" -ForegroundColor Gray
Write-Host "         -ContentType 'application/json' `" -ForegroundColor Gray
Write-Host "         -Body (ConvertTo-Json @{url=\""https://example.com/test-\$i\"})" -ForegroundColor Gray
Write-Host "     }" -ForegroundColor Gray
Write-Host ""
Write-Host "  5. View logs: docker-compose -f docker/docker-compose.observability.yml logs -f" -ForegroundColor White
Write-Host "  6. Stop stack: docker-compose -f docker/docker-compose.observability.yml down" -ForegroundColor White
Write-Host ""
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host "  Documentation" -ForegroundColor Cyan
Write-Host "======================================================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "  - Quick Reference:  observability\QUICK_REFERENCE.md" -ForegroundColor White
Write-Host "  - Full Guide:       observability\README.md" -ForegroundColor White
Write-Host "  - Integration:      OBSERVABILITY_INTEGRATION.md" -ForegroundColor White
Write-Host "  - Runbooks:         observability\runbooks\" -ForegroundColor White
Write-Host ""
Write-Success "Observability stack is ready!"
Write-Host ""

# Open Grafana in browser
$openBrowser = Read-Host "Open Grafana in browser? (Y/N)"
if ($openBrowser -eq 'Y' -or $openBrowser -eq 'y') {
    Start-Process "http://localhost:3000"
}
