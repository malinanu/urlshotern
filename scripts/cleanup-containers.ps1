# URLShortener Container Cleanup Script (PowerShell)
#
# This script safely removes all URLShortener containers and images
# to prepare for a fresh deployment. Safe to run - PostgreSQL is external.
#
# Usage: .\scripts\cleanup-containers.ps1

Write-Host "🧹 URLShortener Container Cleanup" -ForegroundColor Cyan
Write-Host "==================================" -ForegroundColor Cyan
Write-Host ""

# Stop containers
Write-Host "📦 Stopping containers..." -ForegroundColor Yellow
docker stop urlshortener-app urlshortener-redis urlshortener-adminer 2>$null
Write-Host "   ✓ Containers stopped" -ForegroundColor Green

# Remove containers
Write-Host "🗑️  Removing containers..." -ForegroundColor Yellow
docker rm -f urlshortener-app urlshortener-redis urlshortener-adminer 2>$null
Write-Host "   ✓ Containers removed" -ForegroundColor Green

# Remove images
Write-Host "🖼️  Removing images..." -ForegroundColor Yellow
$images = docker images --format "{{.Repository}}:{{.Tag}} {{.ID}}" | Where-Object { $_ -match "urlshortener-app" }
if ($images) {
    $images | ForEach-Object {
        $id = ($_ -split '\s+')[1]
        docker rmi -f $id 2>$null
    }
}
Write-Host "   ✓ Images removed" -ForegroundColor Green

# Prune system
Write-Host "🧽 Pruning Docker system..." -ForegroundColor Yellow
docker system prune -f | Out-Null
Write-Host "   ✓ System pruned" -ForegroundColor Green

Write-Host ""
Write-Host "✅ Cleanup complete!" -ForegroundColor Green
Write-Host ""

# Verify clean state
Write-Host "📊 Current state:" -ForegroundColor Cyan
$containers = docker ps -a | Select-String "urlshortener"
if ($containers) {
    Write-Host "   ⚠️  Some containers still exist" -ForegroundColor Yellow
} else {
    Write-Host "   ✓ No URLShortener containers" -ForegroundColor Green
}

$images = docker images | Select-String "urlshortener"
if ($images) {
    Write-Host "   ⚠️  Some images still exist" -ForegroundColor Yellow
} else {
    Write-Host "   ✓ No URLShortener images" -ForegroundColor Green
}

Write-Host ""
Write-Host "Ready for fresh deployment!" -ForegroundColor Green
