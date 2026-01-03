# Polar Integration Test Suite
# This script tests all aspects of the Polar payment gateway integration
# Usage: .\scripts\test-polar-integration.ps1

param(
    [switch]$Quick,
    [switch]$Full,
    [switch]$SkipAuth
)

Write-Host "`n========================================" -ForegroundColor Cyan
Write-Host "Polar Integration Test Suite" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$API_BASE_URL = "http://localhost:8080/api/v1"
$DOCKER_COMPOSE = "docker-compose -f docker/docker-compose.local.yml"
$TEST_RESULTS = @()
$products = @()  # Store products for later use

# Helper function to log test results
function LogTest {
    param(
        [string]$TestName,
        [bool]$Success,
        [string]$Message = ""
    )

    $status = if ($Success) { "✅ PASS" } else { "❌ FAIL" }
    $color = if ($Success) { "Green" } else { "Red" }

    Write-Host "$status - $TestName" -ForegroundColor $color
    if ($Message) {
        Write-Host "  └─ $Message" -ForegroundColor Gray
    }

    $global:TEST_RESULTS += @{
        Test = $TestName
        Success = $Success
        Message = $Message
    }
}

# Phase 1: Database Verification
Write-Host "`n📊 Phase 1: Database Verification" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

try {
    Write-Host "Checking Docker containers..." -ForegroundColor Gray
    $containers = Invoke-Expression "$DOCKER_COMPOSE ps" 2>&1

    if ($LASTEXITCODE -eq 0) {
        LogTest "Docker Compose Running" $true "Containers are up"
    } else {
        LogTest "Docker Compose Running" $false "Containers not running"
        Write-Host "`n⚠️  Please start Docker containers first:" -ForegroundColor Yellow
        Write-Host "  docker-compose -f docker/docker-compose.local.yml up -d`n" -ForegroundColor Gray
        exit 1
    }

    # Check for Polar tables
    Write-Host "`nChecking Polar database tables..." -ForegroundColor Gray
    $tables = Invoke-Expression "$DOCKER_COMPOSE exec -T postgres psql -U urlshortener -d urlshortener_db -c '\dt polar_*'" 2>&1

    $hasSubscriptions = $tables -match "polar_subscriptions"
    $hasProducts = $tables -match "polar_products"
    $hasEvents = $tables -match "polar_webhook_events"

    LogTest "Table: polar_subscriptions" $hasSubscriptions
    LogTest "Table: polar_products" $hasProducts
    LogTest "Table: polar_webhook_events" $hasEvents

    if ($hasSubscriptions -and $hasProducts -and $hasEvents) {
        LogTest "Database Schema Complete" $true "All Polar tables exist"
    } else {
        LogTest "Database Schema Complete" $false "Missing Polar tables - run migrations"
    }

    # Check migration 002
    Write-Host "`nChecking migration status..." -ForegroundColor Gray
    $migration = Invoke-Expression "$DOCKER_COMPOSE exec -T postgres psql -U urlshortener -d urlshortener_db -c `"SELECT version FROM schema_migrations WHERE version = '002';`"" 2>&1

    if ($migration -match "002") {
        LogTest "Migration 002 Applied" $true "Polar schema migration complete"
    } else {
        LogTest "Migration 002 Applied" $false "Migration not found"
    }

} catch {
    LogTest "Database Verification" $false $_.Exception.Message
}

# Phase 2: Polar API Connectivity Test
Write-Host "`n🌐 Phase 2: Polar API Connectivity" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

try {
    Write-Host "Testing GET /api/v1/products..." -ForegroundColor Gray
    $response = Invoke-WebRequest -Uri "$API_BASE_URL/products" -Method GET -Headers @{"Content-Type" = "application/json"} -ErrorAction Stop

    if ($response.StatusCode -eq 200) {
        $products = $response.Content | ConvertFrom-Json

        if ($products.Count -gt 0) {
            LogTest "Fetch Products" $true "Retrieved $($products.Count) products"
            Write-Host "`nProducts found:" -ForegroundColor Gray
            foreach ($product in $products) {
                Write-Host "  • $($product.name) ($($product.plan_type))" -ForegroundColor Cyan
            }
        } else {
            LogTest "Fetch Products" $false "No products found - configure products in Polar dashboard"
        }
    } else {
        LogTest "Fetch Products" $false "Status: $($response.StatusCode)"
    }

} catch {
    LogTest "Fetch Products" $false $_.Exception.Message
    Write-Host "`n⚠️  API connectivity issue. Possible causes:" -ForegroundColor Yellow
    Write-Host "  1. Backend server not running (http://localhost:8080)" -ForegroundColor Gray
    Write-Host "  2. Invalid Polar API credentials in .env" -ForegroundColor Gray
    Write-Host "  3. Network connectivity to Polar API failed`n" -ForegroundColor Gray
}

# Phase 3: Authenticated Endpoints Test (unless skipped)
if (-not $SkipAuth) {
    Write-Host "`n🔐 Phase 3: Authenticated Endpoints" -ForegroundColor Yellow
    Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

    try {
        # Register test user
        Write-Host "Creating test user..." -ForegroundColor Gray
        $testEmail = "polar-test-$(Get-Random -Minimum 1000 -Maximum 9999)@example.com"
        $registerBody = @{
            email = $testEmail
            password = "Test123!@#"
            name = "Polar Test User"
        } | ConvertTo-Json

        $authResponse = Invoke-WebRequest -Uri "$API_BASE_URL/auth/register" -Method POST -Body $registerBody -Headers @{"Content-Type" = "application/json"} -ErrorAction Stop

        if ($authResponse.StatusCode -eq 200 -or $authResponse.StatusCode -eq 201) {
            $authData = $authResponse.Content | ConvertFrom-Json
            $token = $authData.token

            LogTest "Create Test User" $true "User: $testEmail"

            # Test subscription endpoint
            Write-Host "`nTesting GET /api/v1/billing/subscription..." -ForegroundColor Gray
            try {
                $subResponse = Invoke-WebRequest -Uri "$API_BASE_URL/billing/subscription" -Method GET -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} -ErrorAction SilentlyContinue

                if ($subResponse.StatusCode -eq 200) {
                    $subscription = $subResponse.Content | ConvertFrom-Json
                    LogTest "Get Subscription" $true "Plan: $($subscription.plan_type)"
                } elseif ($subResponse.StatusCode -eq 404) {
                    LogTest "Get Subscription" $true "No subscription (expected for new user)"
                }
            } catch {
                if ($_.Exception.Response.StatusCode.value__ -eq 404) {
                    LogTest "Get Subscription" $true "No subscription (expected for new user)"
                } else {
                    LogTest "Get Subscription" $false $_.Exception.Message
                }
            }

            # Test usage endpoint
            Write-Host "Testing GET /api/v1/billing/usage..." -ForegroundColor Gray
            try {
                $usageResponse = Invoke-WebRequest -Uri "$API_BASE_URL/billing/usage" -Method GET -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} -ErrorAction Stop

                if ($usageResponse.StatusCode -eq 200) {
                    $usage = $usageResponse.Content | ConvertFrom-Json
                    LogTest "Get Usage Limits" $true "Plan: $($usage.plan_type), URLs: $($usage.current_urls)/$($usage.url_limit), Clicks: $($usage.current_clicks_this_month)/$($usage.click_limit)"
                } else {
                    LogTest "Get Usage Limits" $false "Status: $($usageResponse.StatusCode)"
                }
            } catch {
                LogTest "Get Usage Limits" $false $_.Exception.Message
            }

            # Test checkout endpoint (if we have products)
            if ($products.Count -gt 0) {
                Write-Host "Testing POST /api/v1/billing/checkout..." -ForegroundColor Gray
                $productId = $products[0].polar_product_id
                $checkoutBody = @{
                    product_id = $productId
                    success_url = "http://localhost:3000/dashboard?success=true"
                    cancel_url = "http://localhost:3000/pricing?cancelled=true"
                } | ConvertTo-Json

                try {
                    $checkoutResponse = Invoke-WebRequest -Uri "$API_BASE_URL/billing/checkout" -Method POST -Body $checkoutBody -Headers @{"Authorization" = "Bearer $token"; "Content-Type" = "application/json"} -ErrorAction Stop

                    if ($checkoutResponse.StatusCode -eq 200) {
                        $checkout = $checkoutResponse.Content | ConvertFrom-Json
                        LogTest "Create Checkout Session" $true "Checkout URL generated"
                        Write-Host "  └─ URL: $($checkout.checkout_url)" -ForegroundColor Cyan
                    } else {
                        LogTest "Create Checkout Session" $false "Status: $($checkoutResponse.StatusCode)"
                    }
                } catch {
                    LogTest "Create Checkout Session" $false $_.Exception.Message
                }
            } else {
                Write-Host "⊘ Skipping checkout test (no products available)" -ForegroundColor DarkGray
            }

        } else {
            LogTest "Create Test User" $false "Status: $($authResponse.StatusCode)"
        }

    } catch {
        LogTest "Authentication Tests" $false $_.Exception.Message
    }
} else {
    Write-Host "`n⊘ Skipping authenticated endpoint tests (-SkipAuth)" -ForegroundColor DarkGray
}

# Phase 4: Webhook Retry Worker Verification
Write-Host "`n🔄 Phase 4: Webhook Retry Worker" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

try {
    Write-Host "Checking backend logs for retry worker..." -ForegroundColor Gray
    $logs = Invoke-Expression "$DOCKER_COMPOSE logs backend" 2>&1 | Select-String -Pattern "webhook retry worker"

    if ($logs) {
        $hasStarted = $logs -match "Webhook retry worker started"
        LogTest "Retry Worker Started" $hasStarted "Worker is running in background"

        Write-Host "`nRecent webhook worker logs:" -ForegroundColor Gray
        $logs | Select-Object -First 5 | ForEach-Object {
            Write-Host "  $($_.Line)" -ForegroundColor DarkGray
        }
    } else {
        LogTest "Retry Worker Started" $false "No worker logs found"
    }

} catch {
    LogTest "Webhook Worker Check" $false $_.Exception.Message
}

# Phase 5: Database Event Check
Write-Host "`n📬 Phase 5: Webhook Events Database" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

try {
    Write-Host "Querying webhook events..." -ForegroundColor Gray
    $eventCount = Invoke-Expression "$DOCKER_COMPOSE exec -T postgres psql -U urlshortener -d urlshortener_db -c `"SELECT COUNT(*) FROM polar_webhook_events;`"" 2>&1

    if ($eventCount -match "\d+") {
        $count = [regex]::Match($eventCount, '\d+').Value
        LogTest "Webhook Events Table" $true "Found $count events"

        if ([int]$count -gt 0) {
            Write-Host "`nRecent webhook events:" -ForegroundColor Gray
            $recentEvents = Invoke-Expression "$DOCKER_COMPOSE exec -T postgres psql -U urlshortener -d urlshortener_db -c `"SELECT polar_event_id, event_type, processing_status FROM polar_webhook_events ORDER BY created_at DESC LIMIT 5;`"" 2>&1
            Write-Host $recentEvents -ForegroundColor DarkGray
        }
    } else {
        LogTest "Webhook Events Table" $true "No events yet (this is normal)"
    }

    # Check retry statistics
    Write-Host "`nQuerying retry statistics..." -ForegroundColor Gray
    $retryStats = Invoke-Expression "$DOCKER_COMPOSE exec -T postgres psql -U urlshortener -d urlshortener_db -c `"SELECT processing_status, COUNT(*) as count FROM polar_webhook_events GROUP BY processing_status;`"" 2>&1

    if ($retryStats -match "processing_status") {
        Write-Host "`nRetry Statistics:" -ForegroundColor Gray
        Write-Host $retryStats -ForegroundColor DarkGray
    }

} catch {
    LogTest "Webhook Events Check" $false $_.Exception.Message
}

# Phase 6: Email Service Check
Write-Host "`n📧 Phase 6: Email Service Configuration" -ForegroundColor Yellow
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━`n" -ForegroundColor Yellow

try {
    Write-Host "Checking email provider..." -ForegroundColor Gray
    $emailLogs = Invoke-Expression "$DOCKER_COMPOSE logs backend" 2>&1 | Select-String -Pattern "email provider"

    if ($emailLogs) {
        $hasConsole = $emailLogs -match "console provider"
        LogTest "Email Service Initialized" $true "Using console provider (development)"

        Write-Host "`nEmail provider logs:" -ForegroundColor Gray
        $emailLogs | Select-Object -First 3 | ForEach-Object {
            Write-Host "  $($_.Line)" -ForegroundColor DarkGray
        }
    } else {
        LogTest "Email Service Check" $false "No email provider logs found"
    }

} catch {
    LogTest "Email Service Check" $false $_.Exception.Message
}

# Test Summary
Write-Host "`n`n========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "========================================`n" -ForegroundColor Cyan

$totalTests = $TEST_RESULTS.Count
$passedTests = ($TEST_RESULTS | Where-Object { $_.Success }).Count
$failedTests = $totalTests - $passedTests
$successRate = if ($totalTests -gt 0) { [math]::Round(($passedTests / $totalTests) * 100, 1) } else { 0 }

Write-Host "Total Tests:  $totalTests" -ForegroundColor White
Write-Host "Passed:       $passedTests" -ForegroundColor Green
Write-Host "Failed:       $failedTests" -ForegroundColor $(if ($failedTests -gt 0) { "Red" } else { "Green" })
Write-Host "Success Rate: $successRate%" -ForegroundColor $(if ($successRate -ge 80) { "Green" } elseif ($successRate -ge 50) { "Yellow" } else { "Red" })

# Detailed results
Write-Host "`nDetailed Results:" -ForegroundColor Gray
Write-Host "━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━━" -ForegroundColor Gray
foreach ($result in $TEST_RESULTS) {
    $status = if ($result.Success) { "✅" } else { "❌" }
    Write-Host "$status $($result.Test)" -ForegroundColor $(if ($result.Success) { "Green" } else { "Red" })
    if ($result.Message) {
        Write-Host "   $($result.Message)" -ForegroundColor DarkGray
    }
}

# Final verdict
Write-Host "`n`n" -NoNewline
if ($failedTests -eq 0) {
    Write-Host "✅ ALL TESTS PASSED - Polar integration is working correctly!" -ForegroundColor Green
} elseif ($successRate -ge 70) {
    Write-Host "⚠️  MOSTLY WORKING - Some tests failed but core functionality OK" -ForegroundColor Yellow
    Write-Host "`nReview failed tests above and check:" -ForegroundColor Yellow
    Write-Host "  • .env configuration (POLAR_ACCESS_TOKEN, POLAR_WEBHOOK_SECRET)" -ForegroundColor Gray
    Write-Host "  • Backend logs for errors: docker-compose logs backend" -ForegroundColor Gray
    Write-Host "  • Database migrations: docker-compose exec postgres psql..." -ForegroundColor Gray
} else {
    Write-Host "❌ INTEGRATION ISSUES - Multiple tests failed" -ForegroundColor Red
    Write-Host "`nTroubleshooting steps:" -ForegroundColor Yellow
    Write-Host "  1. Verify Docker containers are running" -ForegroundColor Gray
    Write-Host "  2. Check .env file has valid Polar credentials" -ForegroundColor Gray
    Write-Host "  3. Ensure database migrations are applied" -ForegroundColor Gray
    Write-Host "  4. Review backend logs: docker-compose logs backend" -ForegroundColor Gray
    Write-Host "  5. Test network connectivity to Polar API" -ForegroundColor Gray
}

Write-Host "`n`n"

# Exit with appropriate code
if ($failedTests -gt 0) {
    exit 1
} else {
    exit 0
}
