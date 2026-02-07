#!/usr/bin/env pwsh
# Test Production Environment Locally
# This script runs the production-like configuration locally for testing

param(
    [Parameter(Position=0)]
    [ValidateSet('up', 'down', 'logs', 'build', 'rebuild', 'status', 'verify')]
    [string]$Action = 'up'
)

$ErrorActionPreference = "Stop"

$COMPOSE_FILE = "docker-compose.prod-local.yml"
$BASE_URL = "http://localhost:8083"

function Write-Header {
    param([string]$Message)
    Write-Host "`n================================================" -ForegroundColor Cyan
    Write-Host "  $Message" -ForegroundColor Cyan
    Write-Host "================================================`n" -ForegroundColor Cyan
}

function Write-Success {
    param([string]$Message)
    Write-Host "✅ $Message" -ForegroundColor Green
}

function Write-Error-Message {
    param([string]$Message)
    Write-Host "❌ $Message" -ForegroundColor Red
}

function Write-Info {
    param([string]$Message)
    Write-Host "ℹ️  $Message" -ForegroundColor Blue
}

function Test-Container {
    param([string]$ContainerName)
    $status = docker ps --filter "name=$ContainerName" --format "{{.Status}}"
    return $null -ne $status -and $status -ne ""
}

function Verify-Deployment {
    Write-Header "Verifying Production Deployment"

    # Check if container is running
    Write-Info "Checking container status..."
    if (Test-Container "urlshortener-app-prod") {
        Write-Success "Container urlshortener-app-prod is running"
    } else {
        Write-Error-Message "Container urlshortener-app-prod is not running"
        return
    }

    # Check if static files exist
    Write-Info "Checking for frontend static files..."
    $staticCheck = docker exec urlshortener-app-prod sh -c "ls -la /app/static/ 2>/dev/null || echo 'NOT_FOUND'"
    if ($staticCheck -match "index.html") {
        Write-Success "Frontend static files found in /app/static/"
    } else {
        Write-Error-Message "Frontend static files not found in /app/static/"
        Write-Info "Static files check output:"
        Write-Host $staticCheck
    }

    # Test health endpoint
    Write-Info "Testing API health endpoint..."
    Start-Sleep -Seconds 2
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/api/v1/health" -UseBasicParsing -TimeoutSec 10
        if ($response.StatusCode -eq 200) {
            Write-Success "API health endpoint responding: $BASE_URL/api/v1/health"
            Write-Info "Response: $($response.Content)"
        }
    } catch {
        Write-Error-Message "API health endpoint failed: $_"
    }

    # Test frontend homepage
    Write-Info "Testing frontend homepage..."
    try {
        $response = Invoke-WebRequest -Uri "$BASE_URL/" -UseBasicParsing -TimeoutSec 10
        if ($response.StatusCode -eq 200 -and $response.Content -match "<!DOCTYPE html>") {
            Write-Success "Frontend homepage responding: $BASE_URL/"
        } else {
            Write-Error-Message "Frontend homepage returned unexpected content"
        }
    } catch {
        Write-Error-Message "Frontend homepage failed: $_"
    }

    # Check logs for frontend build message
    Write-Info "Checking logs for frontend build confirmation..."
    $logs = docker logs urlshortener-app-prod 2>&1 | Select-String -Pattern "frontend" -Context 0,2
    if ($logs) {
        Write-Success "Frontend build logs found:"
        Write-Host $logs
    } else {
        Write-Info "No frontend-specific logs found yet"
    }

    Write-Header "Access Points"
    Write-Host "Frontend:  $BASE_URL/" -ForegroundColor Yellow
    Write-Host "API:       $BASE_URL/api/v1/health" -ForegroundColor Yellow
    Write-Host "Adminer:   http://localhost:8084" -ForegroundColor Yellow
    Write-Host "`nDatabase Connection:" -ForegroundColor Yellow
    Write-Host "  Host:     localhost:5433" -ForegroundColor Gray
    Write-Host "  Database: urlshortener_db" -ForegroundColor Gray
    Write-Host "  User:     urlshortener_user" -ForegroundColor Gray
}

switch ($Action) {
    'up' {
        Write-Header "Starting Production Environment (Local)"
        Write-Info "Using: $COMPOSE_FILE"
        docker-compose -f $COMPOSE_FILE up -d
        Write-Success "Services started"
        Start-Sleep -Seconds 5
        Verify-Deployment
    }

    'down' {
        Write-Header "Stopping Production Environment (Local)"
        docker-compose -f $COMPOSE_FILE down
        Write-Success "Services stopped"
    }

    'logs' {
        Write-Header "Following Logs (Ctrl+C to exit)"
        docker-compose -f $COMPOSE_FILE logs -f
    }

    'build' {
        Write-Header "Building Production Image"
        docker-compose -f $COMPOSE_FILE build
        Write-Success "Build complete"
    }

    'rebuild' {
        Write-Header "Rebuilding and Restarting"
        docker-compose -f $COMPOSE_FILE down
        docker-compose -f $COMPOSE_FILE build --no-cache
        docker-compose -f $COMPOSE_FILE up -d
        Write-Success "Rebuild complete"
        Start-Sleep -Seconds 5
        Verify-Deployment
    }

    'status' {
        Write-Header "Service Status"
        docker-compose -f $COMPOSE_FILE ps
    }

    'verify' {
        Verify-Deployment
    }

    default {
        Write-Host "Usage: .\test-prod-local.ps1 [up|down|logs|build|rebuild|status|verify]"
        Write-Host ""
        Write-Host "Commands:"
        Write-Host "  up       - Start production environment"
        Write-Host "  down     - Stop production environment"
        Write-Host "  logs     - Follow logs"
        Write-Host "  build    - Build production image"
        Write-Host "  rebuild  - Rebuild from scratch and restart"
        Write-Host "  status   - Show service status"
        Write-Host "  verify   - Verify deployment"
    }
}
