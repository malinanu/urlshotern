# 🚀 Quick Fix Card - 404 Auth Error

## Problem
`POST /api/v1/auth/register` returns **404 Not Found**

## Root Cause
Docker container running **old code** without auth routes

---

## ✅ Quick Fix (Portainer)

### If Using Stack:
1. **Stacks** → Your stack
2. Click **"Pull and redeploy"**
3. Wait for completion
4. Verify logs show: `[GIN-debug] POST /api/v1/auth/register`

### If Using Containers:
1. **Containers** → **urlshortener-app**
2. **Console** → Connect with `/bin/sh`
3. Run:
   ```bash
   cd /app
   git pull origin main
   git log --oneline -1
   # Should show: 52a7883
   ```
4. Exit console
5. Click **"Recreate"** → Check "Pull latest image" → Confirm

---

## 🧪 Test Commands (Portainer Console)

### Health Check:
```bash
wget -qO- http://localhost:8080/health
# Expected: {"status":"ok"}
```

### Test Registration:
```bash
wget -qO- --post-data='{"name":"Test","email":"test@example.com","password":"Pass123!","phone":"+1234567890"}' \
  --header='Content-Type: application/json' \
  http://localhost:8080/api/v1/auth/register
# Expected: User object with tokens (NOT 404)
```

---

## 📋 Verification Checklist

- [ ] Logs show: `[GIN-debug] POST /api/v1/auth/register`
- [ ] Health check returns: `{"status":"ok"}`
- [ ] Registration returns **201** or **400** (NOT 404)
- [ ] Git commit in container is: `52a7883`

---

## 🔧 If Still Broken

### Check Git Version in Container:
```bash
# Portainer Console
cd /app && git log --oneline -1
# Must show: 52a7883 fix: Reorder database...
```

### Force Rebuild:
```bash
# SSH to server
cd /path/to/URLshorter
git pull origin main
docker-compose -f docker-compose.oracle.yml down
docker-compose -f docker-compose.oracle.yml build --no-cache
docker-compose -f docker-compose.oracle.yml up -d
```

---

## 🌐 Browser Ad Blocker Fix

If seeing `ERR_BLOCKED_BY_CLIENT`:
1. Disable ad blocker for your server
2. Test in incognito mode
3. Or use curl (no ad blocker)

---

## 📞 Quick Diagnostics

### View Logs:
**Portainer** → **Containers** → **urlshortener-app** → **Logs**

### Check Environment:
**Portainer** → **Containers** → **urlshortener-app** → **Inspect** → **Config** → **Env**

### Required Env Vars:
- `JWT_SECRET` (MUST be set)
- `DB_HOST=postgresql-central`
- `SERVER_PORT=8080`

---

## 🎯 Success Indicators

### Good Logs:
```
[GIN-debug] POST   /api/v1/auth/register
[GIN-debug] POST   /api/v1/auth/login
Server starting on port 8080
```

### Good Response:
```json
{
  "user": { "id": 1, "name": "Test", ... },
  "access_token": "eyJhbGc...",
  "refresh_token": "eyJhbGc...",
  "expires_in": 86400
}
```

---

## 🔗 External Access

Frontend: `http://YOUR_SERVER_IP:8082`
Backend API: `http://YOUR_SERVER_IP:8082/api/v1`
Health: `http://YOUR_SERVER_IP:8082/health`

---

## 📚 Full Guides

- `PORTAINER_DEPLOYMENT_GUIDE.md` - Step-by-step Portainer instructions
- `DEPLOYMENT_FIX.md` - General deployment troubleshooting
- `ORACLE_QUICK_FIX.md` - Oracle-specific fixes

---

**Key Files Modified:**
- `internal/routes/routes.go:117` - Auth routes registered
- `internal/handlers/auth.go` - Auth handlers implemented
- `internal/storage/postgres.go` - DB init fixed (commit 52a7883)

**Last Updated:** 2026-02-10
