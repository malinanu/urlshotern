# Admin Panel Setup & Testing Guide

This guide walks through enabling and testing the admin panel after implementing environment-controlled admin routes.

## 🚀 Quick Start

### 1. Start Docker Environment

```bash
# Make sure Docker Desktop is running first!
docker-compose -f docker/docker-compose.local.yml up --build
```

Wait for all services to start:
- ✅ Backend API: http://localhost:8090
- ✅ Frontend: http://localhost:3000
- ✅ PostgreSQL: localhost:5432
- ✅ Redis: localhost:6379
- ✅ Adminer: http://localhost:8081

### 2. Grant Admin Privileges

**Option A: Using PowerShell Script (Recommended)**
```powershell
.\scripts\grant-admin.ps1
```

**Option B: Manual SQL Command**
```bash
docker exec -i urlshorter-postgres-1 psql -U urlshortener -d urlshortener_db
```
Then run:
```sql
UPDATE users
SET is_admin = true,
    email_verified = true,
    is_active = true,
    updated_at = NOW()
WHERE email = 'mjjayaweera@gmail.com';

-- Verify
SELECT email, is_admin, email_verified, is_active
FROM users
WHERE email = 'mjjayaweera@gmail.com';
```

Expected output: `is_admin` should be `t` (true)

### 3. Get New JWT Token (CRITICAL!)

**Why?** JWT tokens include the user's role at token generation time. Old tokens won't have the admin role.

**Steps:**
1. Open browser to http://localhost:3000 (or https://trunc.site)
2. **Logout** (click logout button)
3. **Login** again with mjjayaweera@gmail.com
4. A new JWT token will be generated with `"role": "admin"`

**Verify JWT Token:**
- Open browser DevTools (F12)
- Go to: Application → Local Storage → http://localhost:3000
- Copy the `accessToken` value
- Paste into https://jwt.io
- Check the payload contains:
  ```json
  {
    "user_id": 123,
    "email": "mjjayaweera@gmail.com",
    "role": "admin",  // ← Must be present!
    "account_type": "free",
    "token_type": "access"
  }
  ```

### 4. Test Admin API Endpoints

**Test 1: Admin Stats (501 stub)**
```bash
# Replace <YOUR_JWT_TOKEN> with the token from DevTools
curl -X GET http://localhost:8090/api/v1/admin/stats \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json"
```
Expected: `{"error":"Admin stats endpoint not implemented"}` (HTTP 501)

**Test 2: Get User by ID**
```bash
# Replace <USER_ID> and <YOUR_JWT_TOKEN>
curl -X GET http://localhost:8090/api/v1/admin/users/<USER_ID> \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json"
```
Expected: User details (HTTP 200)

**Test 3: Admin User List (501 stub)**
```bash
curl -X GET http://localhost:8090/api/v1/admin/users \
  -H "Authorization: Bearer <YOUR_JWT_TOKEN>" \
  -H "Content-Type: application/json"
```
Expected: `{"error":"User list endpoint not implemented"}` (HTTP 501)

### 5. Test Admin Panel in Browser

1. Navigate to: http://localhost:3000/admin
2. Admin panel should load (no 403 Forbidden error)
3. Open DevTools → Network tab
4. Watch for API calls to `/api/v1/admin/*`
5. Verify responses are 200 OK or 501 (not 401/403)

## 🔍 Troubleshooting

### Issue: "403 Forbidden" on admin routes

**Possible Causes:**
1. **Old JWT token** - Logout and login again
2. **is_admin not set** - Run grant-admin script
3. **ENABLE_ADMIN_ROUTES=false** - Check .env file

**Solution:**
```bash
# 1. Check .env has admin routes enabled
grep ENABLE_ADMIN_ROUTES .env
# Should show: ENABLE_ADMIN_ROUTES=true

# 2. Verify user is admin in database
docker exec -i urlshorter-postgres-1 psql -U urlshortener -d urlshortener_db -c "SELECT email, is_admin FROM users WHERE email = 'mjjayaweera@gmail.com';"

# 3. Logout and login again to get new JWT token
```

### Issue: "404 Not Found" on admin routes

**Cause:** Admin routes not enabled in environment

**Solution:**
```bash
# Check .env file
cat .env | grep ENABLE_ADMIN_ROUTES
# Should be: ENABLE_ADMIN_ROUTES=true

# If missing or false, update .env and restart Docker
docker-compose -f docker/docker-compose.local.yml restart backend
```

### Issue: "401 Unauthorized"

**Cause:** Missing or invalid JWT token

**Solution:**
1. Make sure you're logged in
2. Copy JWT token from browser DevTools → Application → Local Storage
3. Include in request: `-H "Authorization: Bearer <token>"`

### Issue: Backend won't start after changes

**Cause:** Compilation error

**Solution:**
```bash
# Test compilation
go build ./...

# Check for errors in Docker logs
docker-compose -f docker/docker-compose.local.yml logs backend
```

## 🔒 Security Notes

### Environment Variable Control

The admin routes are controlled by the `ENABLE_ADMIN_ROUTES` environment variable:

- **Default:** `false` (admin routes completely disabled)
- **Local Development:** `true` (enabled for testing)
- **Production:** Should be `false` by default
  - Set to `true` only when admin access is needed
  - Set back to `false` after admin work is complete

### Why This is More Secure

1. **Attack Surface Reduction**: Admin routes don't exist when disabled (404 responses)
2. **No Code Execution**: No middleware overhead when disabled
3. **Granular Control**: Can enable/disable without code changes
4. **Defense in Depth**: Multiple layers:
   - Environment variable (routes enabled/disabled)
   - Authentication (valid JWT token)
   - Authorization (is_admin flag)
   - Rate limiting
   - Account status (active, email verified)

### Production Best Practices

```bash
# Production .env - Admin routes DISABLED by default
ENABLE_ADMIN_ROUTES=false

# Only when admin access needed (maintenance window)
ENABLE_ADMIN_ROUTES=true
# Restart backend
docker-compose restart backend

# After admin work complete
ENABLE_ADMIN_ROUTES=false
# Restart backend
docker-compose restart backend
```

## 📋 Implementation Checklist

- [x] Add `EnableAdminRoutes` to Config struct
- [x] Add `ENABLE_ADMIN_ROUTES` to .env.example
- [x] Update routes.go with environment check
- [x] Fix setupAdminRoutes function signature
- [x] Remove CMS routes (handler.CMSHandlers doesn't exist)
- [x] Update SetupRoutes callers in main.go and profile_server.go
- [x] Test compilation (go build ./...)
- [ ] Start Docker containers
- [ ] Grant admin privileges to mjjayaweera@gmail.com
- [ ] Logout and login to get new JWT token
- [ ] Test admin API endpoints
- [ ] Test admin panel in browser
- [ ] Commit changes to git
- [ ] Push to GitHub
- [ ] Deploy to production (with ENABLE_ADMIN_ROUTES=false)

## 🎯 Quick Command Reference

```bash
# Start Docker
docker-compose -f docker/docker-compose.local.yml up --build

# Grant admin privileges
.\scripts\grant-admin.ps1

# Check admin status in database
docker exec -i urlshorter-postgres-1 psql -U urlshortener -d urlshortener_db -c "SELECT email, is_admin, email_verified, is_active FROM users WHERE email = 'mjjayaweera@gmail.com';"

# Test admin endpoint
curl -X GET http://localhost:8090/api/v1/admin/stats -H "Authorization: Bearer <TOKEN>"

# View backend logs
docker-compose -f docker/docker-compose.local.yml logs -f backend

# Restart backend after .env changes
docker-compose -f docker/docker-compose.local.yml restart backend
```

## ✅ Success Indicators

You know the admin panel is working when:

1. ✅ `is_admin = true` in database for mjjayaweera@gmail.com
2. ✅ JWT token payload contains `"role": "admin"`
3. ✅ Admin API endpoints return 200/501 (not 401/403)
4. ✅ Admin panel loads in browser without errors
5. ✅ Backend logs show admin routes registered
6. ✅ No compilation errors

## 🚨 Common Mistakes to Avoid

1. **Don't skip logout/login** - Old JWT tokens don't have admin role
2. **Don't forget to restart backend** - After changing .env
3. **Don't use old tokens** - Always check JWT payload at jwt.io
4. **Don't keep admin routes enabled in production** - Security risk
5. **Don't grant admin without email_verified** - User must verify email

---

**Need Help?** Check the troubleshooting section or review backend logs.
