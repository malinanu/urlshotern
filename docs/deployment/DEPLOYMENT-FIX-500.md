# Fix 500 Error on Assignable Roles Endpoint - Deployment Guide

## Issue Summary

**Problem:** `/api/v1/admin/rbac/assignable-roles` endpoint returns 500 Internal Server Error

**Root Cause:** Most likely the latest backend code (commit `e05f81a`) is not deployed to production yet

**Solution:** Deploy the latest code with defensive error handling and improved logging

---

## Pre-Deployment: Code Changes Summary

The following defensive improvements were added to prevent the 500 error:

### 1. Enhanced Logging
- Added entry log when handler is called
- Added logs for authentication failures
- Added log when user has no roles
- Added log when user has no assignable roles
- Added success log after fetching assignable roles

### 2. Defensive Checks
- Check if user has no roles → return empty array instead of proceeding
- Check if user has no assignable roles based on hierarchy → return empty array
- Ensure `assignableRoles` is never `nil` → convert to empty array if needed

### 3. Better Error Messages
- More descriptive error logging for debugging
- Clearer context for each failure point

---

## Deployment Steps (Oracle Server)

### Step 1: Verify Current State

SSH into the Oracle server and check current deployment:

```bash
# Navigate to project directory
cd /path/to/URLshorter

# Check current commit
git log -1 --oneline
# Expected output: e05f81a fix: Resolve 401 error on assignable-roles endpoint
# If different, proceed to Step 2

# Check running containers
docker ps | grep backend

# Check container start time
docker inspect <backend-container-name> | grep StartedAt
```

**Alternative: Use Diagnostic Script**

```bash
# Run the diagnostic script
./scripts/diagnose-500-error.sh

# This will check:
# - Current git commit
# - Docker container status
# - Recent logs
# - Health endpoint
# - And provide specific recommendations
```

---

### Step 2: Pull Latest Code

If commit `e05f81a` is not present:

```bash
# Pull latest code
git pull origin main

# Verify the commit
git log -1 --oneline
# Should show: e05f81a fix: Resolve 401 error on assignable-roles endpoint
```

---

### Step 3: Rebuild Backend Container

```bash
# Option A: Using docker-compose (recommended)
docker-compose -f docker/docker-compose.oracle.yml build backend

# Option B: Using Portainer
# 1. Go to Portainer UI
# 2. Navigate to Stacks → URLshorter
# 3. Click "Editor"
# 4. Scroll down and click "Update the stack"
# 5. Check "Re-pull image and redeploy"
# 6. Click "Update"
```

---

### Step 4: Restart Backend Service

```bash
# Option A: Using docker-compose
docker-compose -f docker/docker-compose.oracle.yml up -d backend

# Option B: Using Portainer
# 1. Go to Containers
# 2. Find backend container
# 3. Click "Restart"

# Wait for container to be healthy
docker ps | grep backend
# Status should show "Up" (not "Restarting")
```

---

### Step 5: Verify Deployment

#### Check Logs for Startup

```bash
# View real-time logs
docker logs <backend-container-name> --tail=50 --follow

# Look for:
# - "Server starting on port 8080" (or your configured port)
# - No panic or error messages
# - Database connection success
```

#### Test Health Endpoint

```bash
# Test basic health
curl http://localhost:8080/health

# Expected response:
# {"status":"healthy"}
```

#### Test Assignable Roles Endpoint

You need a valid authentication token for this test.

**Option A: Test via Frontend**

1. Open browser DevTools (F12) → Network tab
2. Navigate to `https://trunc.site/admin`
3. Click "Manage Roles" on any user
4. Check Network tab for `/api/v1/admin/rbac/assignable-roles` request

**Expected Results:**
- ✅ Status: 200 OK
- ✅ Response body: Array of roles `[{id: 1, name: "admin", ...}, ...]`
- ✅ "Add Role" section appears with dropdown

**Option B: Test with curl**

```bash
# Get your auth token from browser cookies or localStorage
TOKEN="your_auth_token_here"

# Test the endpoint
curl -v -X GET 'http://localhost:8080/api/v1/admin/rbac/assignable-roles' \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json"

# Expected response:
# HTTP/1.1 200 OK
# [{"id":1,"name":"admin","description":"Administrator role"}, ...]
```

---

### Step 6: Monitor Logs for New Behavior

After deployment, test the endpoint and watch the logs:

```bash
docker logs <backend-container-name> --tail=20 --follow
```

**Expected Log Flow (Success):**

```
INFO: GetAssignableRoles called
INFO: Successfully retrieved assignable roles
```

**Expected Log Flow (User has no roles):**

```
INFO: GetAssignableRoles called
INFO: User has no roles assigned
```

**Expected Log Flow (User has no assignable roles):**

```
INFO: GetAssignableRoles called
INFO: User has no assignable roles based on hierarchy
```

**Error Scenarios:**

```
ERROR: Failed to fetch user roles: <database error details>
→ Check database connection and user_roles table

ERROR: Failed to fetch assignable roles: <database error details>
→ Check database connection and roles table

ERROR: user_id not found in context
→ Check authentication middleware configuration
```

---

## Troubleshooting

### Issue: Still Getting 500 Error After Deployment

1. **Check if code is actually deployed:**
   ```bash
   docker exec <backend-container-name> git log -1 --oneline
   # Should show: e05f81a
   ```

2. **Check logs for specific error:**
   ```bash
   docker logs <backend-container-name> --tail=100 | grep -i error
   ```

3. **Verify database connectivity:**
   ```bash
   # Check if backend can connect to database
   docker logs <backend-container-name> | grep -i "database"
   ```

### Issue: Container Won't Start

1. **Check for port conflicts:**
   ```bash
   netstat -tulpn | grep 8080
   ```

2. **Check Docker logs:**
   ```bash
   docker logs <backend-container-name>
   ```

3. **Verify environment variables:**
   ```bash
   docker exec <backend-container-name> env | grep -E "DB_|REDIS_"
   ```

### Issue: User Has Roles But Gets Empty Array

This is actually **expected behavior** if:
- User only has roles that can't assign other roles (e.g., "user" role)
- User's roles are not in `RoleAssignmentHierarchy`

**To verify:**

1. Check user's roles:
   ```sql
   SELECT r.name
   FROM user_roles ur
   JOIN roles r ON ur.role_id = r.id
   WHERE ur.user_id = <user_id>
     AND (ur.expires_at IS NULL OR ur.expires_at > NOW());
   ```

2. Check `RoleAssignmentHierarchy` in code:
   ```bash
   grep -A 20 "RoleAssignmentHierarchy" internal/models/rbac.go
   ```

**Expected hierarchy:**
- `super_admin` → can assign all roles
- `admin` → can assign `content_manager`, `moderator`, `user`
- `content_manager`, `moderator`, `user` → can assign nothing

---

## Files Changed in This Fix

### Backend
- `internal/handlers/admin_rbac_handlers.go` (lines 311-382)
  - Added entry logging
  - Added check for empty user roles
  - Added check for no assignable roles
  - Added nil safety check for assignableRoles
  - Added success logging

### Diagnostic Tools
- `scripts/diagnose-500-error.sh` (NEW)
- `scripts/diagnose-500-error.ps1` (NEW)

### Documentation
- `DEPLOYMENT-FIX-500.md` (this file)

---

## Quick Checklist

- [ ] SSH into Oracle server
- [ ] Run diagnostic script: `./scripts/diagnose-500-error.sh`
- [ ] If commit `e05f81a` not present: `git pull origin main`
- [ ] Rebuild backend: `docker-compose -f docker/docker-compose.oracle.yml build backend`
- [ ] Restart backend: `docker-compose -f docker/docker-compose.oracle.yml up -d backend`
- [ ] Check logs: `docker logs <backend-container-name> --tail=50`
- [ ] Test health: `curl http://localhost:8080/health`
- [ ] Test endpoint via frontend: `https://trunc.site/admin` → Click "Manage Roles"
- [ ] Verify 200 response and role list appears
- [ ] Monitor logs for any new errors

---

## Success Criteria

✅ Endpoint returns 200 OK
✅ Response contains array of assignable roles (may be empty)
✅ No 500 errors in logs
✅ "Add Role" section appears in UI
✅ Dropdown is populated with available roles

---

## Next Steps After Successful Deployment

1. **Monitor for 24 hours**
   - Watch for any new errors in logs
   - Verify endpoint works for all admin users

2. **Test edge cases**
   - Test with user who has no roles
   - Test with user who has only "user" role
   - Test with user who has multiple roles

3. **Consider future improvements**
   - Add unit tests for edge cases
   - Add integration tests for role hierarchy
   - Consider caching assignable roles per user role

---

## Contact

If issues persist after following this guide, provide:
- Docker logs: `docker logs <backend-container-name> --tail=100`
- Git commit: `git log -1 --oneline`
- Network request details from browser DevTools
- Database query results for affected user
