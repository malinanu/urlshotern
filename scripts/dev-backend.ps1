# Development script to run Go backend server
# Usage: .\scripts\dev-backend.ps1

Write-Host "🚀 Starting Go backend server..." -ForegroundColor Cyan
go run cmd/server/main.go
