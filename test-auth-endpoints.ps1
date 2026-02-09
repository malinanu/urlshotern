# Authentication Endpoints Testing Script (PowerShell)
# Usage: .\test-auth-endpoints.ps1 [ServerUrl]
# Example: .\test-auth-endpoints.ps1 "http://localhost:8082"

param(
    [string]$ServerUrl = "http://localhost:8082"
)

$ApiBase = "$ServerUrl/api/v1"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Authentication Endpoints Test" -ForegroundColor Cyan
Write-Host "Server: $ServerUrl" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host

# Function to test endpoint
function Test-Endpoint {
    param(
        [string]$Method,
        [string]$Endpoint,
        [string]$Data,
        [string]$Description
    )

    Write-Host "Testing: $Method $Endpoint" -ForegroundColor Yellow
    Write-Host "Description: $Description"

    $url = "$ApiBase$Endpoint"

    try {
        if ($Data) {
            $response = Invoke-WebRequest -Uri $url -Method $Method `
                -ContentType "application/json" `
                -Body $Data `
                -UseBasicParsing `
                -ErrorAction SilentlyContinue
        } else {
            $response = Invoke-WebRequest -Uri $url -Method $Method `
                -UseBasicParsing `
                -ErrorAction SilentlyContinue
        }

        $statusCode = $response.StatusCode
        $body = $response.Content

        if ($statusCode -eq 404) {
            Write-Host "❌ FAILED - 404 Not Found" -ForegroundColor Red
            Write-Host "Response: $body"
        } elseif ($statusCode -eq 200 -or $statusCode -eq 201) {
            Write-Host "✅ SUCCESS - $statusCode" -ForegroundColor Green
            try {
                $json = $body | ConvertFrom-Json | ConvertTo-Json -Depth 10
                Write-Host "Response: $json"
            } catch {
                Write-Host "Response: $body"
            }
        } else {
            Write-Host "⚠ $statusCode" -ForegroundColor Yellow
            Write-Host "Response: $body"
        }
    } catch {
        $statusCode = $_.Exception.Response.StatusCode.Value__
        $errorBody = $_.ErrorDetails.Message

        if ($statusCode -eq 404) {
            Write-Host "❌ FAILED - 404 Not Found" -ForegroundColor Red
            Write-Host "Response: $errorBody"
        } elseif ($statusCode -eq 200 -or $statusCode -eq 201) {
            Write-Host "✅ SUCCESS - $statusCode" -ForegroundColor Green
            Write-Host "Response: $errorBody"
        } else {
            Write-Host "⚠ $statusCode" -ForegroundColor Yellow
            Write-Host "Response: $errorBody"
        }
    }

    Write-Host
}

# 1. Health Check
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "1. Health Check" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "GET" -Endpoint "/health" -Data "" -Description "Basic health check"

# 2. Registration (Bad Request - Missing Fields)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "2. Registration - Validation Test" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "POST" -Endpoint "/auth/register" `
    -Data '{"email":"test@example.com"}' `
    -Description "Should return 400 validation error (NOT 404)"

# 3. Registration (Valid)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "3. Registration - Valid User" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
$randomEmail = "test$(Get-Date -Format 'yyyyMMddHHmmss')@example.com"
$registerData = @{
    name = "Test User"
    email = $randomEmail
    password = "SecurePass123!"
    phone = "+1234567890"
} | ConvertTo-Json
Test-Endpoint -Method "POST" -Endpoint "/auth/register" `
    -Data $registerData `
    -Description "Should create user and return tokens"

# 4. Login (Invalid - User doesn't exist)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "4. Login - Invalid Credentials" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "POST" -Endpoint "/auth/login" `
    -Data '{"email":"nonexistent@example.com","password":"WrongPass123!"}' `
    -Description "Should return 401 unauthorized"

# 5. Refresh Token (No token provided)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "5. Refresh Token - No Token" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "POST" -Endpoint "/auth/refresh" `
    -Data '{}' `
    -Description "Should return 400 or 401 (NOT 404)"

# 6. OTP Send (Invalid)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "6. OTP - Send (No Phone)" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "POST" -Endpoint "/auth/otp/send" `
    -Data '{}' `
    -Description "Should return 400 validation error"

# 7. Email Verification (Invalid Token)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "7. Email Verification - Invalid Token" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "GET" -Endpoint "/auth/verify-email?token=invalid" `
    -Data "" `
    -Description "Should return error (NOT 404)"

# 8. Protected Route - Profile (No Auth)
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "8. Protected Route - Profile (No Token)" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Test-Endpoint -Method "GET" -Endpoint "/auth/profile" `
    -Data "" `
    -Description "Should return 401 unauthorized (NOT 404)"

Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Test Summary" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Key Check: " -ForegroundColor Green -NoNewline
Write-Host "If all endpoints return codes OTHER than 404,"
Write-Host "then the authentication routes are properly registered!"
Write-Host
Write-Host "If ANY endpoint returns 404: " -ForegroundColor Red -NoNewline
Write-Host "The container is running old code."
Write-Host "Follow PORTAINER_DEPLOYMENT_GUIDE.md to rebuild."
Write-Host
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "Next Steps" -ForegroundColor Cyan
Write-Host "==========================================" -ForegroundColor Cyan
Write-Host "1. ✅ Verify no 404 errors above"
Write-Host "2. Register a real user from frontend"
Write-Host "3. Test login flow"
Write-Host "4. Test protected endpoints with valid token"
Write-Host "==========================================" -ForegroundColor Cyan
