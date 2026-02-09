# Fix 404 Error on Oracle Server Deployment

## Problem Summary
The `/api/v1/auth/register` endpoint returns **404 Not Found** on your Oracle server deployment, even though the route is properly implemented in the codebase. The root cause is that the Docker container is running **old code** without the authentication routes.

## Solution: Update Docker Container via Portainer

### Quick Fix (Recommended)

#### Option 1: Using Portainer Stack (If using Stack deployment)
1. Open **Portainer** web interface
2. Navigate to **Stacks** → Find your URLshortener stack
3. Click **"Editor"** tab
4. Click **"Pull and redeploy"** or **"Update the stack"** button
5. This will:
   - Pull latest code from Git
   - Rebuild the container image
   - Restart with updated code

#### Option 2: Manual Rebuild (If using individual containers)
1. **Stop the container in Portainer:**
   - Navigate to **Containers**
   - Find your URLshortener container
   - Click the checkbox → Click **"Stop"**

2. **SSH to your Oracle server** and run:
   ```bash
   # Navigate to project directory
   cd /path/to/URLshorter

   # Pull latest code
   git pull origin main

   # Rebuild the container (force no cache)
   docker-compose -f docker/docker-compose.oracle.yml build --no-cache

   # Start the container
   docker-compose -f docker/docker-compose.oracle.yml up -d
   ```

3. **Or use Portainer Console** (if you prefer GUI):
   - In Portainer, go to **Containers** → Your container
   - Click **"Console"** → Connect with `/bin/bash`
   - Run:
     ```bash
     cd /app  # or wherever your code is mounted
     git status  # Check current version
     git pull origin main  # Update code
     ```
   - Then restart the container from Portainer UI

### Step 1: Verify Container is Running Updated Code

After rebuild, check the logs in Portainer:

1. **Portainer** → **Containers** → Your container → **"Logs"** tab
2. Look for these success messages:
   ```
   Server starting on port 8080
   ✅ Webhook retry worker started
   ✅ Product sync worker started
   [GIN-debug] POST   /api/v1/auth/register
   [GIN-debug] POST   /api/v1/auth/login
   ```

If you see `[GIN-debug] POST /api/v1/auth/register`, the routes are registered correctly! ✅

### Step 2: Test the Endpoint

#### Test 1: Health Check (Verify server is responding)
Using Portainer Console or SSH:
```bash
# Test from inside the container
curl http://localhost:8080/health

# Expected response:
{"status":"ok"}
```

From your browser (replace with your server IP):
```bash
curl http://YOUR_SERVER_IP:8080/health
```

#### Test 2: Registration Endpoint (Should return validation error, NOT 404)
Using Portainer Console or SSH:
```bash
# Test with missing data (should return 400 Bad Request, not 404)
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"email":"test@example.com"}'

# Test with valid data
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test User",
    "email": "test@example.com",
    "password": "SecurePass123!",
    "phone": "+1234567890"
  }'
```

**Expected responses:**
- Missing fields: `400 Bad Request` with validation errors (NOT 404!)
- Valid data: `201 Created` with user object and JWT tokens

**If you still get 404:** The container didn't rebuild with new code. Check:
1. Git pull was successful: `git log --oneline -1` (should show commit `52a7883`)
2. Docker build used the new code: `docker-compose build --no-cache`
3. Container was restarted after rebuild

### Step 3: Verify Database Tables

The previous issue was foreign key constraint errors. Let's verify tables exist:

1. In Portainer, find your **PostgreSQL container**
2. Click **"Console"** → Connect with `/bin/bash`
3. Run:
   ```bash
   psql -U postgres -d urlshortener

   # Check users table exists
   \dt users

   # Check url_mappings has user_id column
   \d url_mappings

   # Verify users table structure
   \d users

   # Exit
   \q
   ```

**Expected output:**
- `users` table should exist
- `url_mappings` should have `user_id` column with foreign key to `users(id)`

### Step 4: Fix Frontend Ad Blocker Issue

Once the 404 is resolved, if you still see `ERR_BLOCKED_BY_CLIENT`:

**This is a browser/extension issue, not a server issue.**

1. **Disable ad blocker** for your server's domain/IP:
   - uBlock Origin: Click icon → "Power" button to whitelist
   - AdBlock Plus: Click icon → "Enabled on this site" to disable

2. **Or test in incognito mode** (most extensions disabled by default)

3. **Or test with curl** (no ad blocker):
   ```bash
   curl -X POST http://YOUR_SERVER_IP:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "name": "Test User",
       "email": "test@example.com",
       "password": "SecurePass123!",
       "phone": "+1234567890"
     }'
   ```

### Step 5: Verify Environment Variables

In Portainer, check your container has the required environment variables:

1. **Portainer** → **Containers** → Your container → **"Inspect"** tab
2. Scroll to **"Config"** → **"Env"** section
3. Verify these are set:
   ```
   SERVER_PORT=8080
   JWT_SECRET=<your-secret>
   JWT_ISSUER=<your-issuer>
   DB_HOST=<database-host>
   DB_PORT=5432
   DB_USER=postgres
   DB_PASSWORD=<password>
   DB_NAME=urlshortener
   REDIS_HOST=<redis-host>
   REDIS_PORT=6379
   ```

Missing `JWT_SECRET` or `JWT_ISSUER` will cause authentication to fail (but not 404).

## Verification Checklist

After rebuilding, verify:

- [ ] Container logs show `[GIN-debug] POST /api/v1/auth/register`
- [ ] `curl http://localhost:8080/health` returns `{"status":"ok"}`
- [ ] `POST /api/v1/auth/register` returns 400 or 201 (NOT 404)
- [ ] Database has `users` and `url_mappings` tables
- [ ] Environment variables are set correctly
- [ ] Frontend can reach the backend (ad blocker disabled)

## Troubleshooting

### Problem: Still getting 404 after rebuild

**Cause:** Container wasn't rebuilt or is using cached layers

**Solution:**
```bash
# Force complete rebuild
docker-compose -f docker/docker-compose.oracle.yml down
docker-compose -f docker/docker-compose.oracle.yml build --no-cache
docker-compose -f docker/docker-compose.oracle.yml up -d

# Verify code version inside container
docker exec -it <container-name> git log --oneline -1
# Should show: 52a7883 fix: Reorder database table initialization...
```

### Problem: Database errors during startup

**Cause:** Tables weren't initialized in correct order

**Solution:** This was fixed in commit `52a7883`. If still occurring:
```bash
# Connect to database
psql -U postgres -d urlshortener

# Drop and recreate (WARNING: deletes all data)
DROP TABLE IF EXISTS url_mappings CASCADE;
DROP TABLE IF EXISTS users CASCADE;

# Restart backend container to re-initialize
```

### Problem: ERR_BLOCKED_BY_CLIENT persists

**Cause:** Browser ad blocker blocking `/register` endpoint

**Solution:**
1. Test with `curl` first to verify server works
2. Disable ad blocker for your server's domain
3. Use incognito mode for testing
4. Add server IP to ad blocker whitelist

## Quick Reference Commands

```bash
# View container logs
docker logs <container-name> --tail 100

# Check running containers
docker ps

# Rebuild and restart
docker-compose -f docker/docker-compose.oracle.yml up --build -d

# Enter container shell
docker exec -it <container-name> /bin/bash

# Check Git version in container
docker exec -it <container-name> git log --oneline -5

# Test endpoint from host
curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test","email":"test@example.com","password":"Pass123!"}'
```

## Expected Success Output

When everything is working, you should see:

**Container Logs:**
```
[GIN-debug] POST   /api/v1/auth/register
[GIN-debug] POST   /api/v1/auth/login
Server starting on port 8080
✅ Webhook retry worker started
```

**Health Check:**
```bash
$ curl http://localhost:8080/health
{"status":"ok"}
```

**Registration:**
```bash
$ curl -X POST http://localhost:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{"name":"Test User","email":"test@example.com","password":"SecurePass123!","phone":"+1234567890"}'

{
  "user": {
    "id": 1,
    "name": "Test User",
    "email": "test@example.com",
    "phone": "+1234567890",
    "email_verified": false,
    "phone_verified": false,
    "created_at": "2026-02-10T12:00:00Z"
  },
  "access_token": "eyJhbGciOiJIUzI1NiIs...",
  "refresh_token": "eyJhbGciOiJIUzI1NiIs...",
  "expires_in": 86400
}
```

## Next Steps After Fix

1. Test user registration from frontend
2. Test login flow
3. Test protected endpoints with JWT token
4. Verify email verification flow (if SMTP configured)
5. Test OTP flow (if SMS service configured)

## Need Help?

If you encounter issues not covered here:
1. Check container logs in Portainer
2. Verify database connection: `curl http://localhost:8080/health/db-pool`
3. Check Git commit in container: `docker exec <container> git log --oneline -1`
4. Provide error messages from logs

---

**Last Updated:** 2026-02-10
**Related Commits:**
- `52a7883` - Fix database table initialization order
- `93868a7` - Add HEAD request handlers
- `4222f52` - Fix static file routing
