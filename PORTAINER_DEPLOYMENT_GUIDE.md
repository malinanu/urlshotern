# Portainer Deployment Guide - Fix 404 Authentication Error

## Problem
The `/api/v1/auth/register` endpoint returns **404 Not Found** because your Docker container on the Oracle server is running **old code** without authentication routes.

## Solution Overview
You need to rebuild the Docker container with the latest code from Git. This guide provides step-by-step instructions using **Portainer's UI** exclusively.

---

## Method 1: Using Portainer Stacks (Recommended)

If you deployed using a **Stack** in Portainer:

### Step 1: Navigate to Stack
1. Open **Portainer** web interface
2. Click **"Stacks"** in the left sidebar
3. Find your **"urlshortener"** stack (or whatever you named it)

### Step 2: Pull Latest Code and Rebuild
1. Click on your stack name
2. Click **"Editor"** tab at the top
3. Scroll down to find **"Pull and redeploy"** button
4. Click **"Pull and redeploy"**
5. Wait for the process to complete (watch the status spinner)

**What this does:**
- Pulls latest code from your Git repository
- Rebuilds the Docker image with new code
- Restarts all containers with the updated image

### Step 3: Verify Deployment
1. After completion, click **"Containers"** in the left sidebar
2. Find **"urlshortener-app"** container
3. Click on it → **"Logs"** tab
4. Look for these success messages:
   ```
   [GIN-debug] POST   /api/v1/auth/register
   [GIN-debug] POST   /api/v1/auth/login
   Server starting on port 8080
   ✅ Webhook retry worker started
   ```

If you see `[GIN-debug] POST /api/v1/auth/register`, the fix worked! ✅

---

## Method 2: Manual Container Rebuild via Portainer

If you're **NOT** using Stacks, or Method 1 didn't work:

### Step 1: Stop the Container
1. Click **"Containers"** in Portainer sidebar
2. Find **"urlshortener-app"** container
3. Check the checkbox next to it
4. Click **"Stop"** button at the top
5. Wait for status to change to "Stopped" (gray icon)

### Step 2: Use Portainer Console to Pull Code
1. While still in **Containers** list, click on **"urlshortener-app"** name
2. Click **"Console"** tab
3. In the "Connect" section, select:
   - **Command:** `/bin/sh` (or `/bin/bash` if available)
4. Click **"Connect"** button

You now have a terminal inside your container!

### Step 3: Pull Latest Code (Inside Console)
In the console that just opened, type:

```bash
# Navigate to app directory
cd /app

# Check current code version
git log --oneline -1

# Pull latest code from GitHub
git pull origin main

# Verify you got the latest commit (should show: 52a7883)
git log --oneline -1
```

Expected output:
```
52a7883 fix: Reorder database table initialization to resolve foreign key constraint errors
```

### Step 4: Disconnect and Rebuild Image
1. Close the console window
2. Go back to **"Containers"** → **"urlshortener-app"**
3. Click **"Recreate"** button
4. Check **"Pull latest image"** checkbox
5. Check **"Always pull image"** checkbox
6. Click **"Recreate"** to confirm

### Step 5: Verify Logs
1. After container restarts, click **"Logs"** tab
2. Look for authentication routes in logs:
   ```
   [GIN-debug] POST   /api/v1/auth/register
   [GIN-debug] POST   /api/v1/auth/login
   [GIN-debug] POST   /api/v1/auth/refresh
   [GIN-debug] GET    /api/v1/auth/profile
   ```

---

## Method 3: Complete Rebuild (If Above Methods Fail)

If pulling code inside the container didn't work, you need a full rebuild:

### Option A: Using Portainer Stack (Full Rebuild)

1. Go to **"Stacks"** → Your stack
2. Click **"Stop"** button (top right)
3. Wait for all containers to stop
4. Click **"Remove"** button
5. Confirm removal
6. Go to **"Stacks"** → **"Add Stack"** → **"Git Repository"**
7. Enter your Git repository URL
8. Enter Git credentials if needed
9. Specify compose file: `docker-compose.oracle.yml`
10. Add environment variables (copy from old stack if needed)
11. Click **"Deploy the stack"**

### Option B: Using SSH to Server (Complete Rebuild)

If Portainer methods aren't working, SSH to your server:

```bash
# Navigate to project directory
cd /path/to/URLshorter

# Pull latest code
git pull origin main

# Stop containers
docker-compose -f docker-compose.oracle.yml down

# Remove old images (force fresh build)
docker images | grep urlshortener | awk '{print $3}' | xargs docker rmi -f

# Rebuild with no cache (ensures latest code)
docker-compose -f docker-compose.oracle.yml build --no-cache

# Start containers
docker-compose -f docker-compose.oracle.yml up -d

# Check logs
docker logs urlshortener-app --tail 50
```

---

## Testing the Fix

### Test 1: Health Check (Using Portainer Console)

1. Go to **"Containers"** → **"urlshortener-app"** → **"Console"**
2. Connect with `/bin/sh`
3. Run:
   ```bash
   # Test health endpoint
   wget -qO- http://localhost:8080/health

   # Should return: {"status":"ok"}
   ```

### Test 2: Registration Endpoint (Using Portainer Console)

Still in the console:
```bash
# Test registration (should return 400 validation error, NOT 404)
wget -qO- --post-data='{"email":"test@example.com"}' \
  --header='Content-Type: application/json' \
  http://localhost:8080/api/v1/auth/register

# Test with valid data
wget -qO- --post-data='{"name":"Test User","email":"test@example.com","password":"SecurePass123!","phone":"+1234567890"}' \
  --header='Content-Type: application/json' \
  http://localhost:8080/api/v1/auth/register
```

**Expected:**
- Missing fields: Returns validation error (NOT 404!)
- Valid data: Returns user object with JWT tokens

### Test 3: From Your Browser

From your local machine, test the public endpoint:

```bash
# Replace YOUR_SERVER_IP with your Oracle server IP
curl -X POST http://YOUR_SERVER_IP:8082/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "SecurePass123!",
    "phone": "+1234567890"
  }'
```

**Note:** Port **8082** is exposed externally (see `docker-compose.oracle.yml` line 14)

---

## Fixing ERR_BLOCKED_BY_CLIENT (Ad Blocker Issue)

After fixing the 404 error, if you see `ERR_BLOCKED_BY_CLIENT` in browser console:

**This is a browser ad blocker issue, not a server problem.**

### Solutions:

1. **Disable Ad Blocker** for your server's domain/IP:
   - **uBlock Origin:** Click icon → Click power button
   - **AdBlock Plus:** Click icon → Disable on this site
   - **Brave Browser:** Click shield icon → Disable shields

2. **Test in Incognito/Private Mode:**
   - Most extensions are disabled by default
   - Open incognito window
   - Navigate to your frontend
   - Try registration

3. **Whitelist Your Server:**
   - Add server IP to ad blocker whitelist
   - Usually in extension settings → Whitelist

4. **Test with curl** (no ad blocker):
   - Use curl commands above to verify server works
   - If curl works but browser doesn't = ad blocker issue

---

## Verification Checklist

After deployment, verify:

- [ ] Container logs show: `[GIN-debug] POST /api/v1/auth/register`
- [ ] Health check returns: `{"status":"ok"}`
- [ ] Registration endpoint returns 400 or 201 (NOT 404)
- [ ] Frontend can connect (ad blocker disabled)
- [ ] Database has `users` table: `\dt users` in PostgreSQL
- [ ] Environment variables are set (check in Portainer Inspect tab)

---

## Common Issues & Solutions

### Issue: Container Won't Start After Rebuild

**Check logs in Portainer:**
1. **Containers** → **urlshortener-app** → **Logs**
2. Look for error messages

**Common causes:**
- Database connection failed (check `DB_HOST`, `DB_PASSWORD`)
- Missing JWT_SECRET environment variable
- Port 8080 already in use

**Solution:**
1. Go to **"Containers"** → **"urlshortener-app"** → **"Inspect"** tab
2. Check **"Config"** → **"Env"** section
3. Verify all required variables are set

### Issue: Still Getting 404 After Rebuild

**Cause:** Container is still using old code

**Verification:**
1. Open console: **Containers** → **urlshortener-app** → **Console**
2. Run: `git log --oneline -1`
3. Should show: `52a7883 fix: Reorder database table initialization...`

**If it shows an older commit:**
- The container wasn't rebuilt
- Try Method 3 (Complete Rebuild) above

### Issue: Database Connection Errors

**Check PostgreSQL container:**
1. Go to **"Containers"** → **"postgresql-central"**
2. Check it's running (green status)
3. Click **"Logs"** to verify no errors

**Check network connection:**
1. **Containers** → **"urlshortener-app"** → **"Console"**
2. Run: `ping postgresql-central`
3. Should resolve and respond

**Verify credentials:**
1. **Containers** → **"urlshortener-app"** → **"Inspect"**
2. Check **Config** → **Env**:
   - `DB_HOST=postgresql-central`
   - `DB_USER=urlshortener_user`
   - `DB_PASSWORD=<correct-password>`
   - `DB_NAME=urlshortener_db`

---

## Environment Variables Reference

Required environment variables (check in Portainer Inspect tab):

```bash
# Server
SERVER_PORT=8080
ENVIRONMENT=production
BASE_URL=https://trunc.site

# Database (Centralized PostgreSQL)
DB_HOST=postgresql-central
DB_PORT=5432
DB_USER=urlshortener_user
DB_PASSWORD=<your-password>
DB_NAME=urlshortener_db
DB_SSLMODE=disable

# Redis
REDIS_HOST=urlshortener-redis
REDIS_PORT=6379

# JWT (REQUIRED for auth)
JWT_SECRET=<generate-random-secret>
JWT_ISSUER=urlshortener
```

**Generate JWT_SECRET:**
```bash
# In Portainer console or SSH
openssl rand -base64 32
```

---

## Quick Reference: Portainer Navigation

- **View Containers:** Sidebar → Containers
- **View Logs:** Containers → Click container → Logs tab
- **Open Console:** Containers → Click container → Console tab
- **Restart Container:** Containers → Select container → Restart button
- **View Environment:** Containers → Click container → Inspect tab → Config → Env
- **Stacks:** Sidebar → Stacks

---

## Success Indicators

When everything is working correctly:

### Container Logs (Portainer → Containers → urlshortener-app → Logs)
```
[GIN-debug] POST   /api/v1/auth/register
[GIN-debug] POST   /api/v1/auth/login
[GIN-debug] GET    /api/v1/auth/profile
Server starting on port 8080
✅ Webhook retry worker started
```

### Health Check (Portainer Console)
```bash
$ wget -qO- http://localhost:8080/health
{"status":"ok"}
```

### Registration Response (Portainer Console)
```bash
$ wget -qO- --post-data='{"name":"Test","email":"test@example.com","password":"Pass123!","phone":"+1234567890"}' \
  --header='Content-Type: application/json' \
  http://localhost:8080/api/v1/auth/register

{
  "user": {
    "id": 1,
    "name": "Test",
    "email": "test@example.com",
    "email_verified": false,
    "phone_verified": false
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

---

## Need More Help?

If issues persist:

1. **Export container logs:**
   - Portainer → Containers → urlshortener-app → Logs
   - Copy last 100 lines

2. **Export environment variables:**
   - Portainer → Containers → urlshortener-app → Inspect
   - Copy Config → Env section

3. **Check database status:**
   - Portainer → Containers → postgresql-central → Logs

4. **Test database connection:**
   - Open console in app container
   - Run: `wget -qO- http://localhost:8080/health/db-pool`

Provide these details when asking for help.

---

**Last Updated:** 2026-02-10
**Related Files:**
- `docker-compose.oracle.yml` - Oracle server deployment config
- `Dockerfile.dev` - Container build configuration
- `DEPLOYMENT_FIX.md` - General deployment troubleshooting
