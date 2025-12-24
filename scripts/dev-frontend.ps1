# Development script to run Next.js frontend dev server
# Usage: .\scripts\dev-frontend.ps1

Write-Host "🎨 Starting Next.js development server..." -ForegroundColor Cyan
Set-Location frontend
npm run dev
