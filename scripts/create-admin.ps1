#!/usr/bin/env pwsh
# PowerShell script to create admin users

param(
    [Parameter(Mandatory=$false)]
    [string]$Name,

    [Parameter(Mandatory=$false)]
    [string]$Email,

    [Parameter(Mandatory=$false)]
    [string]$Password,

    [Parameter(Mandatory=$false)]
    [ValidateSet("super_admin", "system_admin", "support_admin", "billing_admin", "content_admin")]
    [string]$Type = "super_admin"
)

# Colors for output
$Green = [System.ConsoleColor]::Green
$Red = [System.ConsoleColor]::Red
$Yellow = [System.ConsoleColor]::Yellow
$Cyan = [System.ConsoleColor]::Cyan

function Write-ColorOutput {
    param([string]$Message, [System.ConsoleColor]$Color = [System.ConsoleColor]::White)
    $originalColor = [Console]::ForegroundColor
    [Console]::ForegroundColor = $Color
    Write-Host $Message
    [Console]::ForegroundColor = $originalColor
}

Write-ColorOutput "`n🔧 Admin User Creation Script" $Cyan
Write-ColorOutput "================================`n" $Cyan

# Interactive mode if parameters not provided
if (-not $Name) {
    $Name = Read-Host "Enter admin name"
}

if (-not $Email) {
    $Email = Read-Host "Enter admin email"
}

if (-not $Password) {
    $securePassword = Read-Host "Enter admin password" -AsSecureString
    $BSTR = [System.Runtime.InteropServices.Marshal]::SecureStringToBSTR($securePassword)
    $Password = [System.Runtime.InteropServices.Marshal]::PtrToStringAuto($BSTR)
}

if ($Name -eq "" -or $Email -eq "" -or $Password -eq "") {
    Write-ColorOutput "❌ Error: Name, email, and password are required!" $Red
    exit 1
}

Write-Host "`nCreating admin user with:"
Write-Host "  Name: $Name"
Write-Host "  Email: $Email"
Write-Host "  Type: $Type"
Write-Host ""

# Run the Go script
$scriptDir = Split-Path -Parent $MyInvocation.MyCommand.Path
$projectRoot = Split-Path -Parent $scriptDir

Push-Location $projectRoot
try {
    go run scripts/create-admin.go -name="$Name" -email="$Email" -password="$Password" -type="$Type"

    if ($LASTEXITCODE -eq 0) {
        Write-ColorOutput "`n✅ Success! Admin user created." $Green
        Write-ColorOutput "`nLogin at: http://localhost:3000/admin/login" $Yellow
        Write-ColorOutput "Email: $Email" $Yellow
    } else {
        Write-ColorOutput "`n❌ Failed to create admin user. Check the error above." $Red
    }
} finally {
    Pop-Location
}
