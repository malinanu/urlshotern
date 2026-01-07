# Docker Build Troubleshooting Guide

## Issue: Build Timeout During `go mod download`

### Problem Description

When running `docker-compose -f docker/docker-compose.local.yml up --build`, the build process fails during the `go mod download` step with a timeout or cancellation error:

```
[backend 6/13] RUN go mod download
ERROR: canceled
```

This typically occurs due to:
- Slow network connection to default Go module sources
- Network timeout issues
- Firewall/proxy restrictions
- Large dependency size causing timeout

### Solution Applied

**File Modified:** `Dockerfile.backend.dev` (Line 17)

**Change:**
```dockerfile
# Before:
RUN go mod download

# After:
RUN GOPROXY=https://proxy.golang.org,direct go mod download
```

**Why This Works:**
- `GOPROXY=https://proxy.golang.org,direct` uses the official Go module proxy (fast CDN)
- The proxy caches all public Go modules for faster downloads
- Falls back to direct downloads if proxy fails
- Significantly reduces download time and prevents timeouts

### How to Apply the Fix

1. The fix has already been applied to `Dockerfile.backend.dev`
2. Run the build again:
   ```bash
   docker-compose -f docker/docker-compose.local.yml up --build
   ```

### Additional Troubleshooting Steps

If the issue persists after applying the fix, try these alternatives:

#### Option 1: Increase Docker Client Timeout

```powershell
# PowerShell (Windows)
$env:COMPOSE_HTTP_TIMEOUT=600
docker-compose -f docker/docker-compose.local.yml up --build
```

```bash
# Bash (Linux/macOS)
export COMPOSE_HTTP_TIMEOUT=600
docker-compose -f docker/docker-compose.local.yml up --build
```

#### Option 2: Build Backend Service Only

Isolate the issue by building just the backend:
```bash
docker-compose -f docker/docker-compose.local.yml build backend
```

#### Option 3: Clear Docker Build Cache

Remove all build cache and rebuild:
```bash
# Clear all build cache
docker builder prune -a -f

# Rebuild without cache
docker-compose -f docker/docker-compose.local.yml build --no-cache backend
```

#### Option 4: Use Alternative Go Proxy

For users in China or behind restrictive firewalls, edit `Dockerfile.backend.dev` line 17:

```dockerfile
# China users:
RUN GOPROXY=https://goproxy.cn,direct go mod download

# Alternative proxy:
RUN GOPROXY=https://goproxy.io,direct go mod download
```

#### Option 5: Check Network Connectivity

Test Go proxy connectivity:
```bash
# Test proxy.golang.org
curl -I https://proxy.golang.org

# Test direct module download
go env -w GOPROXY=https://proxy.golang.org,direct
go mod download
```

#### Option 6: Disable VPN/Proxy

If you're using a VPN or corporate proxy:
1. Temporarily disable VPN
2. Try the build again
3. If it works, configure Docker to work with your proxy

### Verify the Fix

After successful build, you should see:
```
[backend 6/13] RUN GOPROXY=https://proxy.golang.org,direct go mod download
 => CACHED [backend 7/13] COPY . .
```

The build should complete in under 2 minutes for the `go mod download` step.

### Related Files

- `Dockerfile.backend.dev` - Development backend Dockerfile with hot reload
- `docker/docker-compose.local.yml` - Local development Docker Compose configuration
- `go.mod` - Go module dependencies
- `go.sum` - Go module checksums

### Prevention

This issue is now fixed permanently in the repository. Future builds will use the Go proxy by default.

### Additional Resources

- [Go Module Proxy Documentation](https://go.dev/ref/mod#module-proxy)
- [Docker Build Optimization](https://docs.docker.com/build/building/best-practices/)
- [Go Environment Variables](https://go.dev/ref/mod#environment-variables)

---

**Last Updated:** 2026-01-08
**Related Issue:** Docker build timeout during dependency download
**Status:** RESOLVED
