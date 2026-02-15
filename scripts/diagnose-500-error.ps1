# Diagnostic script for 500 error on assignable-roles endpoint
# This script helps verify deployment status and diagnose the issue

$ErrorActionPreference = "Continue"

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "500 Error Diagnostic Script" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
Write-Host ""

# Step 1: Check environment
Write-Host "Step 1: Environment Check" -ForegroundColor Blue
Write-Host "Current directory: $(Get-Location)"
Write-Host "Current user: $env:USERNAME"
Write-Host ""

# Step 2: Check git commit
Write-Host "Step 2: Git Commit Status" -ForegroundColor Blue
$EXPECTED_COMMIT = "e05f81a"
$CURRENT_COMMIT = (git log -1 --oneline).Split(' ')[0]

if ($CURRENT_COMMIT -like "$EXPECTED_COMMIT*") {
    Write-Host "✓ Latest commit found: $CURRENT_COMMIT" -ForegroundColor Green
} else {
    Write-Host "⚠ Different commit found: $CURRENT_COMMIT" -ForegroundColor Yellow
    Write-Host "  Expected: $EXPECTED_COMMIT"
    Write-Host ""
    Write-Host "Latest commits:"
    git log -3 --oneline
}
Write-Host ""

# Step 3: Check Docker containers
Write-Host "Step 3: Docker Container Status" -ForegroundColor Blue
if (Get-Command docker -ErrorAction SilentlyContinue) {
    Write-Host "Running containers:"
    docker ps --format "table {{.Names}}\t{{.Status}}\t{{.Image}}" | Select-String -Pattern "backend|urlshorter"
    Write-Host ""

    # Get backend container
    $BACKEND_CONTAINER = docker ps --format "{{.Names}}" | Select-String -Pattern "backend|urlshorter" | Select-Object -First 1

    if ($BACKEND_CONTAINER) {
        Write-Host "✓ Found backend container: $BACKEND_CONTAINER" -ForegroundColor Green

        # Check when container was started
        Write-Host ""
        Write-Host "Container start time:"
        docker inspect $BACKEND_CONTAINER | Select-String -Pattern "StartedAt"

        # Check git commit in container
        Write-Host ""
        Write-Host "Checking commit in running container..."
        try {
            docker exec $BACKEND_CONTAINER git log -1 --oneline 2>$null
        } catch {
            Write-Host "Unable to check git commit in container"
        }
    } else {
        Write-Host "✗ No backend container found" -ForegroundColor Red
    }
} else {
    Write-Host "⚠ Docker not available on this machine" -ForegroundColor Yellow
}
Write-Host ""

# Step 4: Check recent logs
Write-Host "Step 4: Recent Backend Logs" -ForegroundColor Blue
if ($BACKEND_CONTAINER) {
    Write-Host "Last 30 lines of backend logs:"
    Write-Host "---"
    docker logs $BACKEND_CONTAINER --tail=30
    Write-Host "---"
    Write-Host ""

    # Check for error patterns
    Write-Host "Checking for error patterns..."
    $logs = docker logs $BACKEND_CONTAINER --tail=100

    if ($logs | Select-String -Pattern "panic" -CaseSensitive:$false) {
        Write-Host "✗ Found panic messages in logs" -ForegroundColor Red
    }

    if ($logs | Select-String -Pattern "Failed to fetch user roles") {
        Write-Host "✗ Found 'Failed to fetch user roles' errors" -ForegroundColor Red
    }

    if ($logs | Select-String -Pattern "Failed to fetch assignable roles") {
        Write-Host "✗ Found 'Failed to fetch assignable roles' errors" -ForegroundColor Red
    }
} else {
    Write-Host "⚠ No backend container to check logs" -ForegroundColor Yellow
}
Write-Host ""

# Step 5: Test endpoint
Write-Host "Step 5: Endpoint Test" -ForegroundColor Blue
try {
    $response = Invoke-WebRequest -Uri "http://localhost:8080/health" -UseBasicParsing -ErrorAction Stop
    if ($response.StatusCode -eq 200) {
        Write-Host "✓ Backend health check passed" -ForegroundColor Green
    }
} catch {
    Write-Host "✗ Backend health check failed" -ForegroundColor Red
    Write-Host $_.Exception.Message
}
Write-Host ""

# Step 6: Recommendations
Write-Host "Step 6: Recommendations" -ForegroundColor Blue
Write-Host ""

if ($CURRENT_COMMIT -notlike "$EXPECTED_COMMIT*") {
    Write-Host "Action Required:" -ForegroundColor Yellow
    Write-Host "1. Pull latest code: git pull origin main"
    Write-Host "2. Rebuild backend: docker-compose -f docker/docker-compose.oracle.yml build backend"
    Write-Host "3. Restart backend: docker-compose -f docker/docker-compose.oracle.yml up -d backend"
    Write-Host "4. Check logs: docker logs <container-name> --tail=50 --follow"
} elseif (-not $BACKEND_CONTAINER) {
    Write-Host "Action Required:" -ForegroundColor Yellow
    Write-Host "1. Start backend container"
    Write-Host "2. Check docker-compose configuration"
} else {
    Write-Host "Deployment appears up to date" -ForegroundColor Green
    Write-Host ""
    Write-Host "If 500 error persists, check the following:"
    Write-Host "1. View real-time logs: docker logs $BACKEND_CONTAINER --tail=50 --follow"
    Write-Host "2. Test endpoint with auth: curl -v http://localhost:8080/api/v1/admin/rbac/assignable-roles"
    Write-Host "3. Check database connection"
    Write-Host "4. Verify user has roles in database"
}
Write-Host ""

Write-Host "================================================" -ForegroundColor Cyan
Write-Host "Diagnostic Complete" -ForegroundColor Cyan
Write-Host "================================================" -ForegroundColor Cyan
