# Deployment Verification Script for Password Reset Pages
# Run this on the Oracle Ubuntu server after deploying (via SSH with PowerShell remoting)

param(
    [string]$ComposeFile = "docker/docker-compose.oracle.yml",
    [string]$ProductionUrl = "https://trunc.site"
)

Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Password Reset Pages Deployment Verification" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""

function Print-Status {
    param([bool]$Success, [string]$Message)
    if ($Success) {
        Write-Host "✓ $Message" -ForegroundColor Green
    } else {
        Write-Host "✗ $Message" -ForegroundColor Red
    }
}

function Print-Warning {
    param([string]$Message)
    Write-Host "⚠ $Message" -ForegroundColor Yellow
}

# Step 1: Check Docker Containers
Write-Host "1. Checking Docker Containers..." -ForegroundColor White
try {
    $containerStatus = docker-compose -f $ComposeFile ps 2>&1
    if ($containerStatus -match "Up") {
        Print-Status $true "Containers are running"
        docker-compose -f $ComposeFile ps
    } else {
        Print-Status $false "Containers are not running properly"
        exit 1
    }
} catch {
    Print-Status $false "Error checking container status: $_"
    exit 1
}
Write-Host ""

# Step 2: Get Container ID
Write-Host "2. Getting Container ID..." -ForegroundColor White
try {
    $containerId = (docker-compose -f $ComposeFile ps -q app).Trim()
    if ([string]::IsNullOrEmpty($containerId)) {
        Print-Status $false "Could not find app container"
        exit 1
    }
    Print-Status $true "Container ID: $containerId"
} catch {
    Print-Status $false "Error getting container ID: $_"
    exit 1
}
Write-Host ""

# Step 3: Check Static Files
Write-Host "3. Checking Static Files in Container..." -ForegroundColor White

try {
    $forgotPasswordExists = docker exec $containerId test -f /app/static/forgot-password.html 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Status $true "forgot-password.html exists"
        $fileSize = docker exec $containerId stat -c%s /app/static/forgot-password.html 2>&1
        Write-Host "   Size: $fileSize bytes" -ForegroundColor Gray
    } else {
        Print-Status $false "forgot-password.html NOT FOUND"
    }

    $resetPasswordExists = docker exec $containerId test -f /app/static/reset-password.html 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Status $true "reset-password.html exists"
        $fileSize = docker exec $containerId stat -c%s /app/static/reset-password.html 2>&1
        Write-Host "   Size: $fileSize bytes" -ForegroundColor Gray
    } else {
        Print-Status $false "reset-password.html NOT FOUND"
    }
} catch {
    Print-Status $false "Error checking static files: $_"
}
Write-Host ""

# Step 4: Check Next.js Static Assets
Write-Host "4. Checking Next.js Static Assets..." -ForegroundColor White
try {
    $assetDirExists = docker exec $containerId test -d /app/static/_next/static 2>&1
    if ($LASTEXITCODE -eq 0) {
        Print-Status $true "Next.js static assets directory exists"
        $assetCount = docker exec $containerId find /app/static/_next/static -type f | Measure-Object | Select-Object -ExpandProperty Count
        Write-Host "   Files: $assetCount" -ForegroundColor Gray
    } else {
        Print-Status $false "Next.js static assets NOT FOUND"
    }
} catch {
    Print-Status $false "Error checking Next.js assets: $_"
}
Write-Host ""

# Step 5: Test HTTP Endpoints
Write-Host "5. Testing HTTP Endpoints..." -ForegroundColor White

try {
    # Test forgot-password page
    $response = Invoke-WebRequest -Uri "$ProductionUrl/forgot-password" -Method Get -UseBasicParsing -ErrorAction SilentlyContinue
    if ($response.StatusCode -eq 200) {
        Print-Status $true "GET $ProductionUrl/forgot-password → $($response.StatusCode)"
    } else {
        Print-Status $false "GET $ProductionUrl/forgot-password → $($response.StatusCode) (Expected 200)"
    }
} catch {
    Print-Status $false "GET $ProductionUrl/forgot-password → Error: $($_.Exception.Message)"
}

try {
    # Test reset-password page
    $response = Invoke-WebRequest -Uri "$ProductionUrl/reset-password?token=test" -Method Get -UseBasicParsing -ErrorAction SilentlyContinue
    if ($response.StatusCode -eq 200) {
        Print-Status $true "GET $ProductionUrl/reset-password → $($response.StatusCode)"
    } else {
        Print-Status $false "GET $ProductionUrl/reset-password → $($response.StatusCode) (Expected 200)"
    }
} catch {
    Print-Status $false "GET $ProductionUrl/reset-password → Error: $($_.Exception.Message)"
}
Write-Host ""

# Step 6: Test API Endpoints
Write-Host "6. Testing API Endpoints..." -ForegroundColor White

try {
    $body = @{ email = "test@example.com" } | ConvertTo-Json
    $response = Invoke-WebRequest -Uri "$ProductionUrl/api/v1/auth/forgot-password" `
        -Method Post `
        -ContentType "application/json" `
        -Body $body `
        -UseBasicParsing `
        -ErrorAction SilentlyContinue

    if ($response.StatusCode -eq 200 -or $response.StatusCode -eq 400) {
        Print-Status $true "POST $ProductionUrl/api/v1/auth/forgot-password → $($response.StatusCode)"
    } else {
        Print-Status $false "POST $ProductionUrl/api/v1/auth/forgot-password → $($response.StatusCode)"
    }
} catch {
    Print-Status $false "POST $ProductionUrl/api/v1/auth/forgot-password → Error: $($_.Exception.Message)"
}
Write-Host ""

# Step 7: Check Application Logs
Write-Host "7. Checking Recent Application Logs..." -ForegroundColor White
Write-Host "Last 20 log entries:" -ForegroundColor Gray
docker-compose -f $ComposeFile logs --tail=20 app
Write-Host ""

# Step 8: Check for Errors
Write-Host "8. Checking for Errors in Logs..." -ForegroundColor White
try {
    $logs = docker-compose -f $ComposeFile logs --tail=100 | Out-String
    $errorCount = ($logs | Select-String -Pattern "error" -AllMatches -CaseSensitive:$false).Matches.Count

    if ($errorCount -eq 0) {
        Print-Status $true "No errors found in recent logs"
    } else {
        Print-Warning "Found $errorCount error entries in logs (review above)"
    }

    $notFoundCount = ($logs | Select-String -Pattern "404" -AllMatches).Matches.Count
    if ($notFoundCount -eq 0) {
        Print-Status $true "No 404 errors found in recent logs"
    } else {
        Print-Warning "Found $notFoundCount 404 entries in logs"
    }
} catch {
    Print-Warning "Error checking logs: $_"
}
Write-Host ""

# Step 9: Environment Variables Check
Write-Host "9. Checking Critical Environment Variables..." -ForegroundColor White
$envVars = @("SMTP_HOST", "SMTP_PORT", "SMTP_FROM", "BASE_URL")

foreach ($var in $envVars) {
    try {
        $value = docker-compose -f $ComposeFile exec -T app env | Select-String "^$var="
        if ($value) {
            Print-Status $true "$var is set"
        } else {
            Print-Status $false "$var is NOT set"
        }
    } catch {
        Print-Status $false "$var check failed: $_"
    }
}
Write-Host ""

# Summary
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host "Verification Complete!" -ForegroundColor Cyan
Write-Host "=========================================" -ForegroundColor Cyan
Write-Host ""
Write-Host "Next Steps:" -ForegroundColor Yellow
Write-Host "1. Test forgot-password page in browser: $ProductionUrl/forgot-password"
Write-Host "2. Test reset-password page in browser: $ProductionUrl/reset-password?token=test"
Write-Host "3. Perform end-to-end password reset flow with real email"
Write-Host "4. Monitor logs: docker-compose -f $ComposeFile logs -f"
Write-Host ""
