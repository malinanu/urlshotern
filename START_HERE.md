# 🚀 START HERE - Fix 404 Auth Error

## What Happened?

Your Oracle server deployment is returning **404 Not Found** on `/api/v1/auth/register` because the Docker container is running **old code** without authentication routes.

## ✅ Good News

The authentication code **IS** properly implemented in your codebase! The issue is just that your Docker container needs to be rebuilt with the latest code.

---

## 📋 Quick Solution (Choose One)

### Option 1: Using Portainer (Easiest) ⭐

**If you deployed with a Stack:**
1. Open **Portainer** web interface
2. Go to **Stacks** → Your URLshortener stack
3. Click **"Pull and redeploy"** button
4. Wait for completion
5. ✅ Done! Skip to **Verify the Fix** section below

**If you deployed with Containers:**
1. Open **Portainer** → **Containers** → **urlshortener-app**
2. Click **"Console"** → Connect with `/bin/sh`
3. Run: `cd /app && git pull origin main`
4. Exit console
5. Click **"Recreate"** → Check "Pull latest image" → Confirm
6. ✅ Done! Skip to **Verify the Fix** section below

**Need detailed Portainer steps?**
→ Read `PORTAINER_DEPLOYMENT_GUIDE.md`

---

### Option 2: Using SSH (Complete Rebuild)

```bash
# SSH to your Oracle server
ssh user@your-server

# Navigate to project directory
cd /path/to/URLshorter

# Pull latest code
git pull origin main

# Rebuild container (force no cache)
docker-compose -f docker-compose.oracle.yml down
docker-compose -f docker-compose.oracle.yml build --no-cache
docker-compose -f docker-compose.oracle.yml up -d

# ✅ Done! Continue to verification below
```

---

## 🧪 Verify the Fix

### Step 1: Check Container Logs

**Via Portainer:**
1. **Containers** → **urlshortener-app** → **Logs**
2. Look for these lines:
   ```
   [GIN-debug] POST   /api/v1/auth/register
   [GIN-debug] POST   /api/v1/auth/login
   Server starting on port 8080
   ```

**Via SSH:**
```bash
docker logs urlshortener-app --tail 50 | grep register
```

### Step 2: Test the Endpoint

**Option A: Using Test Script (Easiest)**
```bash
# From your local machine (Windows)
.\test-auth-endpoints.ps1 "http://YOUR_SERVER_IP:8082"

# From your local machine (Linux/macOS)
./test-auth-endpoints.sh http://YOUR_SERVER_IP:8082

# From server (SSH)
./test-auth-endpoints.sh http://localhost:8082
```

**Option B: Using curl**
```bash
# Test health check
curl http://YOUR_SERVER_IP:8082/health

# Test registration (should return validation error, NOT 404)
curl -X POST http://YOUR_SERVER_IP:8082/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
```

### Step 3: Test from Browser

1. **Disable ad blocker** for your server (important!)
2. Open browser: `http://YOUR_SERVER_IP:8082`
3. Go to registration page
4. Fill in the form
5. Submit

**Expected:** User is created and logged in automatically ✅

**If you see ERR_BLOCKED_BY_CLIENT:**
- Your ad blocker is blocking the request
- Disable ad blocker or whitelist your server
- Test in incognito mode

---

## ✅ Success Indicators

You'll know it's working when:

1. **Container logs show:**
   ```
   [GIN-debug] POST   /api/v1/auth/register
   [GIN-debug] POST   /api/v1/auth/login
   ```

2. **Health check responds:**
   ```bash
   $ curl http://YOUR_SERVER_IP:8082/health
   {"status":"ok"}
   ```

3. **Registration returns 400 or 201 (NOT 404):**
   ```bash
   $ curl -X POST http://YOUR_SERVER_IP:8082/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{"email":"test@example.com"}'

   # Returns: {"error": "validation error..."}
   # (This is good! NOT 404!)
   ```

4. **Valid registration succeeds:**
   ```json
   {
     "user": {
       "id": 1,
       "name": "Test User",
       "email": "test@example.com"
     },
     "access_token": "eyJhbGc...",
     "refresh_token": "eyJhbGc...",
     "expires_in": 86400
   }
   ```

---

## 📚 Documentation Files

Choose the guide that fits your situation:

1. **PORTAINER_DEPLOYMENT_GUIDE.md** ⭐ Start here if using Portainer
   - Step-by-step Portainer UI instructions
   - Screenshots of where to click
   - Three deployment methods

2. **QUICK_FIX_CARD.md** ⚡ One-page quick reference
   - Essential commands only
   - Print and keep handy

3. **DEPLOYMENT_FIX.md** 🔧 General troubleshooting
   - Works for any deployment method
   - Covers Docker, SSH, Portainer

4. **IMPLEMENTATION_SUMMARY.md** 📖 Complete analysis
   - Full problem breakdown
   - Architecture details
   - Code references

5. **test-auth-endpoints.sh** / **.ps1** 🧪 Testing scripts
   - Automated endpoint testing
   - Color-coded output

---

## 🆘 Still Having Issues?

### Issue: Still getting 404 after rebuild

**Solution:**
```bash
# Verify container has latest code
docker exec urlshortener-app git log --oneline -1

# Should show:
# b789d53 docs: Add comprehensive deployment guides...
# or
# 52a7883 fix: Reorder database table initialization...

# If it shows an older commit, force rebuild:
docker-compose -f docker-compose.oracle.yml build --no-cache
```

### Issue: ERR_BLOCKED_BY_CLIENT in browser

**Solution:** This is your ad blocker!
1. Disable ad blocker for your server
2. Test in incognito mode
3. Whitelist your server IP in ad blocker settings

### Issue: Database errors

**Solution:** Check database tables exist
```bash
# Via Portainer Console (postgresql-central container)
psql -U postgres -d urlshortener_db

# Check tables
\dt users
\dt url_mappings

# Exit
\q
```

### Issue: Container won't start

**Solution:** Check environment variables
1. **Portainer** → **Containers** → **urlshortener-app** → **Inspect**
2. Look in **Config** → **Env** section
3. Verify **JWT_SECRET** is set (required!)

```bash
# Generate JWT_SECRET if missing
openssl rand -base64 32
```

---

## 🎯 Next Steps After Fix

1. ✅ Verify authentication works
2. Test user registration from frontend
3. Test login flow
4. Test protected endpoints (profile, etc.)
5. Configure email service (optional, for verification emails)
6. Configure SMS service (optional, for OTP)
7. Set up monitoring/alerting

---

## 📞 Need More Help?

**Choose your guide based on your deployment method:**

- Using **Portainer**? → Read `PORTAINER_DEPLOYMENT_GUIDE.md`
- Want a **quick fix**? → Read `QUICK_FIX_CARD.md`
- Need **troubleshooting**? → Read `DEPLOYMENT_FIX.md`
- Want **full details**? → Read `IMPLEMENTATION_SUMMARY.md`

**All guides include:**
- ✅ Step-by-step instructions
- ✅ Verification checklists
- ✅ Troubleshooting tips
- ✅ Success indicators

---

## 🔑 Key Points

1. **Your code is fine!** The auth routes ARE implemented
2. **Docker container has old code** - needs rebuild
3. **Easy fix** - Pull and redeploy in Portainer (or rebuild via SSH)
4. **Ad blocker issue** is separate - disable for your server
5. **Test scripts** will help verify everything works

---

## ⚡ TL;DR (Too Long; Didn't Read)

```bash
# On Oracle server (SSH)
cd /path/to/URLshorter
git pull origin main
docker-compose -f docker-compose.oracle.yml down
docker-compose -f docker-compose.oracle.yml build --no-cache
docker-compose -f docker-compose.oracle.yml up -d

# Test
curl http://localhost:8082/health
# Should return: {"status":"ok"}

# Test registration (should NOT return 404)
curl -X POST http://localhost:8082/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'
# Should return: validation error (NOT 404!)
```

**Or via Portainer:** Stacks → "Pull and redeploy" ✅

---

**Last Updated:** 2026-02-10
**Status:** Ready to deploy
**Estimated Time:** 5-10 minutes
