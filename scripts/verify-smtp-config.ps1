# SMTP Configuration Verification Script for Production
# Usage: .\verify-smtp-config.ps1

param(
    [string]$ContainerName = "app"
)

Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "SMTP Configuration Verification" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

# Find container ID
Write-Host "Step 1: Finding container..." -ForegroundColor White

try {
    $containerId = docker ps --filter "name=$ContainerName" --format "{{.ID}}" | Select-Object -First 1

    if (-not $containerId) {
        Write-Host "✗ Container not found" -ForegroundColor Red
        Write-Host "Make sure your app container is running" -ForegroundColor Yellow
        exit 1
    }

    Write-Host "✓ Container found: $containerId" -ForegroundColor Green
    Write-Host ""
} catch {
    Write-Host "✗ Error finding container: $_" -ForegroundColor Red
    exit 1
}

# Function to check environment variable
function Check-EnvVar {
    param(
        [string]$VarName,
        [bool]$Required = $false
    )

    $value = docker exec $containerId env | Select-String "^${VarName}=" | ForEach-Object { $_ -replace "^${VarName}=", "" }

    if (-not $value) {
        if ($Required) {
            Write-Host "✗ ${VarName}: NOT SET (Required)" -ForegroundColor Red
            return $false
        } else {
            Write-Host "⚠ ${VarName}: NOT SET (Optional)" -ForegroundColor Yellow
            return $true
        }
    } else {
        # Mask password
        if ($VarName -eq "SMTP_PASSWORD") {
            Write-Host "✓ ${VarName}: **************** (Set)" -ForegroundColor Green
        } else {
            Write-Host "✓ ${VarName}: $value" -ForegroundColor Green
        }
        return $true
    }
}

# Check SMTP environment variables
Write-Host "Step 2: Checking SMTP environment variables..." -ForegroundColor White
Write-Host ""

$allOk = $true
$allOk = (Check-EnvVar -VarName "ENVIRONMENT" -Required $true) -and $allOk
$allOk = (Check-EnvVar -VarName "SMTP_HOST" -Required $true) -and $allOk
$allOk = (Check-EnvVar -VarName "SMTP_PORT" -Required $false) -and $allOk
$allOk = (Check-EnvVar -VarName "SMTP_USERNAME" -Required $true) -and $allOk
$allOk = (Check-EnvVar -VarName "SMTP_PASSWORD" -Required $true) -and $allOk
$allOk = (Check-EnvVar -VarName "FROM_EMAIL" -Required $false) -and $allOk
$allOk = (Check-EnvVar -VarName "FROM_NAME" -Required $false) -and $allOk

Write-Host ""

# Check for SMTP errors in logs
Write-Host "Step 3: Checking for SMTP errors in logs..." -ForegroundColor White
Write-Host ""

$errorLog = docker logs $containerId 2>&1 | Select-String "failed to create email provider" | Select-Object -Last 5

if ($errorLog) {
    Write-Host "✗ SMTP provider creation errors found:" -ForegroundColor Red
    $errorLog | ForEach-Object { Write-Host $_ -ForegroundColor Red }
    Write-Host ""
    $allOk = $false
} else {
    Write-Host "✓ No SMTP provider creation errors" -ForegroundColor Green
}

# Check if emails are going to console
Write-Host ""
Write-Host "Step 4: Checking if emails are being logged to console..." -ForegroundColor White
Write-Host ""

$consoleLog = docker logs $containerId 2>&1 | Select-String "EMAIL \(CONSOLE\)" | Select-Object -Last 1

if ($consoleLog) {
    Write-Host "✗ Emails are being logged to console (SMTP not working)" -ForegroundColor Red
    Write-Host "Last console email log:" -ForegroundColor Yellow
    Write-Host $consoleLog -ForegroundColor Yellow
    Write-Host ""
    $allOk = $false
} else {
    Write-Host "✓ No console email logs found (SMTP should be working)" -ForegroundColor Green
}

# Test SMTP connectivity
Write-Host ""
Write-Host "Step 5: Testing SMTP connectivity..." -ForegroundColor White
Write-Host ""

$smtpHost = docker exec $containerId env | Select-String "^SMTP_HOST=" | ForEach-Object { $_ -replace "^SMTP_HOST=", "" }
$smtpPort = docker exec $containerId env | Select-String "^SMTP_PORT=" | ForEach-Object { $_ -replace "^SMTP_PORT=", "" }

if (-not $smtpPort) {
    $smtpPort = 587
}

if ($smtpHost) {
    Write-Host "Testing connection to ${smtpHost}:${smtpPort}..." -ForegroundColor White

    try {
        $tcpClient = New-Object System.Net.Sockets.TcpClient
        $connectTask = $tcpClient.ConnectAsync($smtpHost, $smtpPort)

        if ($connectTask.Wait(5000)) {
            Write-Host "✓ SMTP server is reachable" -ForegroundColor Green
            $tcpClient.Close()
        } else {
            throw "Connection timeout"
        }
    } catch {
        Write-Host "✗ Cannot connect to SMTP server" -ForegroundColor Red
        Write-Host "Possible causes:" -ForegroundColor Yellow
        Write-Host "  - SMTP server is down" -ForegroundColor Yellow
        Write-Host "  - Port $smtpPort is blocked by firewall" -ForegroundColor Yellow
        Write-Host "  - Network connectivity issues" -ForegroundColor Yellow
        $allOk = $false
    }
} else {
    Write-Host "⚠ SMTP_HOST not set, skipping connectivity test" -ForegroundColor Yellow
}

# Summary
Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host "Summary" -ForegroundColor Cyan
Write-Host "==================================================" -ForegroundColor Cyan
Write-Host ""

if ($allOk) {
    Write-Host "✓ SMTP configuration looks good!" -ForegroundColor Green
    Write-Host ""
    Write-Host "Next steps:" -ForegroundColor White
    Write-Host "1. Test password reset at: https://trunc.site/forgot-password" -ForegroundColor White
    Write-Host "2. Check your email inbox (and spam folder)" -ForegroundColor White
    Write-Host "3. Verify reset link works correctly" -ForegroundColor White
} else {
    Write-Host "✗ SMTP configuration has issues" -ForegroundColor Red
    Write-Host ""
    Write-Host "Actions needed:" -ForegroundColor Yellow
    Write-Host "1. Add missing SMTP environment variables in Portainer" -ForegroundColor White
    Write-Host "2. Verify SMTP credentials are correct" -ForegroundColor White
    Write-Host "3. Restart the container after making changes" -ForegroundColor White
    Write-Host "4. Run this script again to verify" -ForegroundColor White
    Write-Host ""
    Write-Host "See SMTP-CONFIGURATION-GUIDE.md for detailed instructions" -ForegroundColor Cyan
}

Write-Host ""
Write-Host "==================================================" -ForegroundColor Cyan
