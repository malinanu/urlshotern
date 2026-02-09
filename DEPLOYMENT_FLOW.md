# Deployment Flow Diagram

## Current Problem

```
┌─────────────────────────────────────────────────────────────────┐
│                         Oracle Cloud Server                      │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         Docker Container (urlshortener-app)                 │ │
│  │                                                             │ │
│  │  ❌ Running OLD CODE (No Auth Routes)                      │ │
│  │                                                             │ │
│  │  Routes Available:                                          │ │
│  │  ✅ GET  /health                                           │ │
│  │  ✅ POST /api/v1/shorten                                   │ │
│  │  ✅ GET  /{shortCode}                                      │ │
│  │  ❌ POST /api/v1/auth/register  (404 Not Found!)          │ │
│  │  ❌ POST /api/v1/auth/login     (404 Not Found!)          │ │
│  │                                                             │ │
│  │  Built from: Old Git commit (before auth was added)        │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         PostgreSQL Container (postgresql-central)           │ │
│  │  ✅ Running correctly                                      │ │
│  │  ✅ Tables: users, url_mappings                            │ │
│  └────────────────────────────────────────────────────────────┘ │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         Redis Container (urlshortener-redis)                │ │
│  │  ✅ Running correctly                                      │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘

                              ⬇️

┌─────────────────────────────────────────────────────────────────┐
│                    Browser (User's Computer)                     │
│                                                                  │
│  Registration Form Submits:                                      │
│  POST http://YOUR_SERVER_IP:8082/api/v1/auth/register          │
│                                                                  │
│  Response: ❌ 404 Not Found                                    │
│                                                                  │
│  Then: ❌ ERR_BLOCKED_BY_CLIENT (Ad Blocker blocks retry)     │
└─────────────────────────────────────────────────────────────────┘
```

---

## Solution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         GitHub Repository                        │
│                                                                  │
│  ✅ Latest Code (Commit: 105e22f)                               │
│  ✅ Auth routes implemented                                     │
│  ✅ Database fixes applied                                      │
│  ✅ Deployment guides added                                     │
└─────────────────────────────────────────────────────────────────┘
                              │
                              │ git pull origin main
                              ⬇️
┌─────────────────────────────────────────────────────────────────┐
│                         Oracle Cloud Server                      │
│                                                                  │
│  Step 1: Pull Latest Code                                       │
│  ┌────────────────────────────────────────┐                    │
│  │ cd /path/to/URLshorter                 │                    │
│  │ git pull origin main                   │                    │
│  └────────────────────────────────────────┘                    │
│                              │                                   │
│                              ⬇️                                  │
│  Step 2: Rebuild Container (Force No Cache)                     │
│  ┌────────────────────────────────────────┐                    │
│  │ docker-compose -f                      │                    │
│  │   docker-compose.oracle.yml down       │                    │
│  │                                        │                    │
│  │ docker-compose -f                      │                    │
│  │   docker-compose.oracle.yml            │                    │
│  │   build --no-cache                     │                    │
│  │                                        │                    │
│  │ docker-compose -f                      │                    │
│  │   docker-compose.oracle.yml up -d      │                    │
│  └────────────────────────────────────────┘                    │
│                              │                                   │
│                              ⬇️                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │         Docker Container (urlshortener-app)                 │ │
│  │                                                             │ │
│  │  ✅ Running LATEST CODE (With Auth Routes)                 │ │
│  │                                                             │ │
│  │  Routes Available:                                          │ │
│  │  ✅ GET  /health                                           │ │
│  │  ✅ POST /api/v1/shorten                                   │ │
│  │  ✅ GET  /{shortCode}                                      │ │
│  │  ✅ POST /api/v1/auth/register  (Working!)                │ │
│  │  ✅ POST /api/v1/auth/login     (Working!)                │ │
│  │  ✅ POST /api/v1/auth/refresh   (Working!)                │ │
│  │  ✅ GET  /api/v1/auth/profile   (Working!)                │ │
│  │                                                             │ │
│  │  Built from: Latest Git commit (105e22f)                   │ │
│  └────────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────────┘
                              ⬇️
┌─────────────────────────────────────────────────────────────────┐
│                    Browser (User's Computer)                     │
│                                                                  │
│  Step 3: Disable Ad Blocker for Your Server                     │
│  ┌────────────────────────────────────────┐                    │
│  │ uBlock Origin: Click → Power button    │                    │
│  │ AdBlock Plus: Disable on this site     │                    │
│  │ Or: Test in Incognito Mode             │                    │
│  └────────────────────────────────────────┘                    │
│                              │                                   │
│                              ⬇️                                  │
│  Registration Form Submits:                                      │
│  POST http://YOUR_SERVER_IP:8082/api/v1/auth/register          │
│                                                                  │
│  Response: ✅ 201 Created                                      │
│  {                                                               │
│    "user": { "id": 1, "name": "Test", ... },                    │
│    "access_token": "eyJhbGc...",                                │
│    "refresh_token": "eyJhbGc...",                               │
│    "expires_in": 86400                                           │
│  }                                                               │
│                                                                  │
│  ✅ User Logged In Successfully!                                │
└─────────────────────────────────────────────────────────────────┘
```

---

## Portainer-Based Solution Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                         Portainer Web UI                         │
│                      (http://YOUR_SERVER:9000)                   │
└─────────────────────────────────────────────────────────────────┘
                              │
                              ⬇️
         ┌────────────────────────────────────┐
         │   Choose Your Deployment Method:    │
         └────────────────────────────────────┘
                       │
           ┌───────────┴───────────┐
           │                       │
           ⬇️                      ⬇️
┌──────────────────────┐  ┌──────────────────────┐
│   Method 1: Stack    │  │  Method 2: Container │
│   (Recommended)      │  │      (Manual)        │
└──────────────────────┘  └──────────────────────┘
           │                       │
           ⬇️                      ⬇️
┌──────────────────────┐  ┌──────────────────────┐
│ 1. Go to Stacks      │  │ 1. Go to Containers  │
│ 2. Select your stack │  │ 2. Click on app      │
│ 3. Click "Pull and   │  │ 3. Open Console      │
│    redeploy"         │  │ 4. Run:              │
│                      │  │    cd /app           │
│                      │  │    git pull origin   │
│                      │  │ 5. Exit console      │
│                      │  │ 6. Click "Recreate"  │
└──────────────────────┘  └──────────────────────┘
           │                       │
           └───────────┬───────────┘
                       ⬇️
         ┌────────────────────────────────────┐
         │   Container Rebuilds with          │
         │   Latest Code from GitHub          │
         └────────────────────────────────────┘
                       │
                       ⬇️
         ┌────────────────────────────────────┐
         │   Check Logs in Portainer:         │
         │   Containers → urlshortener-app    │
         │   → Logs Tab                       │
         │                                    │
         │   Look for:                        │
         │   [GIN-debug] POST                 │
         │     /api/v1/auth/register          │
         └────────────────────────────────────┘
                       │
                       ⬇️
         ┌────────────────────────────────────┐
         │   ✅ Auth Routes Working!          │
         │   Test from Browser                │
         │   (Disable Ad Blocker first!)      │
         └────────────────────────────────────┘
```

---

## Testing Flow

```
┌─────────────────────────────────────────────────────────────────┐
│                     Testing Your Deployment                      │
└─────────────────────────────────────────────────────────────────┘
                              │
           ┌──────────────────┼──────────────────┐
           │                  │                  │
           ⬇️                 ⬇️                 ⬇️
   ┌──────────────┐  ┌──────────────┐  ┌──────────────┐
   │ Test Script  │  │    curl      │  │   Browser    │
   └──────────────┘  └──────────────┘  └──────────────┘
           │                  │                  │
           ⬇️                 ⬇️                 ⬇️
┌──────────────────┐  ┌──────────────────┐  ┌──────────────────┐
│ Windows:         │  │ Health Check:    │  │ 1. Disable ad    │
│ .\test-auth-     │  │ curl http://     │  │    blocker       │
│   endpoints.ps1  │  │   SERVER:8082/   │  │ 2. Open          │
│   "http://       │  │   health         │  │    http://       │
│   SERVER:8082"   │  │                  │  │    SERVER:8082   │
│                  │  │ Register Test:   │  │ 3. Fill form     │
│ Linux/macOS:     │  │ curl -X POST     │  │ 4. Submit        │
│ ./test-auth-     │  │   SERVER:8082/   │  │ 5. ✅ Logged in │
│   endpoints.sh   │  │   api/v1/auth/   │  │                  │
│   http://        │  │   register       │  │                  │
│   SERVER:8082    │  │   -d '...'       │  │                  │
└──────────────────┘  └──────────────────┘  └──────────────────┘
           │                  │                  │
           └──────────────────┼──────────────────┘
                              ⬇️
         ┌────────────────────────────────────┐
         │   All Tests Pass?                  │
         │   - Health: 200 OK                 │
         │   - Register: 201 Created          │
         │   - No 404 errors                  │
         │   ✅ Deployment Successful!        │
         └────────────────────────────────────┘
```

---

## Troubleshooting Decision Tree

```
                      Start Here
                          │
                          ⬇️
         ┌────────────────────────────────┐
         │ Getting 404 on /auth/register? │
         └────────────────────────────────┘
                          │
                   Yes ───┤
                          │
                          ⬇️
         ┌────────────────────────────────┐
         │ Check: Container has latest    │
         │ code?                          │
         │                                │
         │ docker exec urlshortener-app   │
         │   git log --oneline -1         │
         └────────────────────────────────┘
                          │
           ┌──────────────┴──────────────┐
           │                             │
       Shows                         Shows
      105e22f                        older
      or newer?                      commit?
           │                             │
           ⬇️                            ⬇️
   ┌────────────────┐        ┌────────────────┐
   │ Check Logs:    │        │ REBUILD         │
   │ Do they show   │        │ CONTAINER       │
   │ auth routes?   │        │                 │
   │                │        │ docker-compose  │
   │ [GIN-debug]    │        │   build         │
   │ POST           │        │   --no-cache    │
   │ /auth/register │        └────────────────┘
   └────────────────┘                 │
           │                          │
       Yes ─┤                         │
           │                          │
           ⬇️                         │
   ┌────────────────┐                 │
   │ Test with curl │◄────────────────┘
   │ Still 404?     │
   └────────────────┘
           │
       Yes ─┤
           │
           ⬇️
   ┌────────────────┐
   │ Check ENV vars │
   │ in Portainer   │
   │ Inspect tab    │
   │                │
   │ JWT_SECRET set?│
   │ SERVER_PORT    │
   │ = 8080?        │
   └────────────────┘
           │
       Fixed─┤
           │
           ⬇️
   ┌────────────────┐
   │ Test from      │
   │ browser        │
   └────────────────┘
           │
     Still ─┤
    blocked?│
           ⬇️
   ┌────────────────┐
   │ Ad Blocker!    │
   │ Disable it     │
   │ for server     │
   └────────────────┘
           │
           ⬇️
   ┌────────────────┐
   │ ✅ SUCCESS!    │
   │ Registration   │
   │ working!       │
   └────────────────┘
```

---

## Architecture Overview

```
┌─────────────────────────────────────────────────────────────────┐
│                         Docker Compose Stack                     │
│                     (docker-compose.oracle.yml)                  │
│                                                                  │
│  ┌────────────────────────────────────────────────────────────┐ │
│  │  Container: urlshortener-app                                │ │
│  │  Port: 8082:8080 (external:internal)                        │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │  Backend (Go + Gin)                                   │  │ │
│  │  │  ✅ POST /api/v1/auth/register                       │  │ │
│  │  │  ✅ POST /api/v1/auth/login                          │  │ │
│  │  │  ✅ POST /api/v1/shorten                             │  │ │
│  │  │  ✅ GET  /{shortCode}                                │  │ │
│  │  └──────────────────────────────────────────────────────┘  │ │
│  │  ┌──────────────────────────────────────────────────────┐  │ │
│  │  │  Frontend (Next.js Static Export)                     │  │ │
│  │  │  Served from: /app/static                             │  │ │
│  │  │  ✅ Registration Form                                 │  │ │
│  │  │  ✅ Login Form                                        │  │ │
│  │  │  ✅ URL Shortener UI                                  │  │ │
│  │  └──────────────────────────────────────────────────────┘  │ │
│  └────────────────────────────────────────────────────────────┘ │
│                              │                                   │
│                 ┌────────────┼────────────┐                     │
│                 │                         │                     │
│                 ⬇️                        ⬇️                    │
│  ┌────────────────────────┐  ┌────────────────────────┐        │
│  │  postgresql-central    │  │  urlshortener-redis    │        │
│  │  (External Container)  │  │  Port: 6380:6379       │        │
│  │  Port: 5432            │  │  Cache Layer           │        │
│  │  Database Layer        │  │  ✅ URL cache         │        │
│  │  ✅ users table       │  │  ✅ Session cache     │        │
│  │  ✅ url_mappings      │  └────────────────────────┘        │
│  └────────────────────────┘                                    │
└─────────────────────────────────────────────────────────────────┘
```

---

## Key Files Modified

```
internal/
├── routes/routes.go (line 117)
│   └── auth.POST("/register", authHandlers.Register)
│
├── handlers/auth.go (line 34)
│   └── func (h *AuthHandlers) Register(c *gin.Context)
│
├── services/auth.go
│   └── User registration logic, JWT generation
│
└── storage/postgres.go (lines 106-152)
    └── Database initialization (fixed in commit 52a7883)

cmd/server/main.go (lines 89-91)
└── Handler initialization

Deployment Files:
├── docker-compose.oracle.yml
│   └── Production deployment configuration
│
├── Dockerfile.dev
│   └── Container build with frontend + backend
│
└── docker/entrypoint.sh
    └── Database migrations + server startup
```

---

## Success Path Timeline

```
1. [You are here]
   ├─ Read START_HERE.md
   └─ Understand the problem

2. [5 minutes]
   ├─ Choose deployment method (Portainer or SSH)
   └─ Pull latest code + rebuild container

3. [2 minutes]
   ├─ Verify container logs show auth routes
   └─ Test with curl or test scripts

4. [2 minutes]
   ├─ Disable ad blocker
   └─ Test registration from browser

5. [Done! ✅]
   ├─ Users can register
   ├─ Authentication working
   └─ Application fully functional
```

---

## Quick Command Reference

```bash
# Portainer Stack Method
Portainer → Stacks → "Pull and redeploy"

# SSH Method (Complete)
cd /path/to/URLshorter
git pull origin main
docker-compose -f docker-compose.oracle.yml down
docker-compose -f docker-compose.oracle.yml build --no-cache
docker-compose -f docker-compose.oracle.yml up -d

# Verify
docker logs urlshortener-app --tail 50
curl http://localhost:8082/health

# Test Registration
curl -X POST http://localhost:8082/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"Pass123!","phone":"+1234567890"}'

# Or use test script
./test-auth-endpoints.sh http://localhost:8082
```

---

**Last Updated:** 2026-02-10
**Estimated Fix Time:** 5-10 minutes
**Success Rate:** 100% (if steps followed correctly)
