# Fix 500 Error - Quick Start (5 Minutes)

## The Problem
`/api/v1/admin/rbac/assignable-roles` returns 500 error instead of role list.

## Most Likely Cause
**Backend code not deployed to production yet.** Code was pushed to GitHub but Oracle server needs to rebuild and restart.

---

## Quick Fix (On Oracle Server)

### 1. SSH into server
```bash
ssh user@oracle-server-ip
cd /path/to/URLshorter
```

### 2. Run diagnostic script
```bash
./scripts/diagnose-500-error.sh
```

The script will tell you exactly what to do.

---

## If Diagnostic Says "Not Deployed"

### Pull, Rebuild, Restart
```bash
# Pull latest code
git pull origin main

# Rebuild backend
docker-compose -f docker/docker-compose.oracle.yml build backend

# Restart backend
docker-compose -f docker/docker-compose.oracle.yml up -d backend

# Watch logs
docker logs <backend-container-name> --tail=50 --follow
```

---

## If Diagnostic Says "Already Deployed"

### Check Logs for Error
```bash
docker logs <backend-container-name> --tail=100 | grep -i error
```

Common errors and fixes:
- **"Failed to fetch user roles"** → Database connection issue
- **"Failed to fetch assignable roles"** → Database query error
- **"user_id not found in context"** → Authentication middleware issue

---

## Using Portainer Instead

1. Go to **Portainer UI**
2. Navigate to **Stacks → URLshorter**
3. Click **Editor**
4. Click **Update the stack**
5. Check **"Re-pull image and redeploy"**
6. Click **Update**
7. Wait for deployment to complete
8. Check **Containers** → Backend should show "Up" status

---

## Verify Fix

### Test in Browser
1. Go to `https://trunc.site/admin`
2. Open DevTools (F12) → Network tab
3. Click "Manage Roles" on any user
4. Check `/api/v1/admin/rbac/assignable-roles` request

**Success:**
- ✅ Status: 200 OK
- ✅ Response: Array of roles
- ✅ "Add Role" dropdown appears

---

## Code Changes (Already Made)

✅ Added logging to track request flow
✅ Added check for users with no roles
✅ Added check for users with no assignable roles
✅ Added nil safety checks
✅ Better error messages

**No UI changes needed** - Frontend already handles the response correctly.

---

## What Changed

**File:** `internal/handlers/admin_rbac_handlers.go`

**Before:**
```go
func (h *AdminRBACHandlers) GetAssignableRoles(c *gin.Context) {
    // ... get user ID ...
    userRolesData, err := h.rbacService.GetUserRoles(userID)
    // ... process roles ...
    c.JSON(http.StatusOK, assignableRoles) // Could be nil!
}
```

**After:**
```go
func (h *AdminRBACHandlers) GetAssignableRoles(c *gin.Context) {
    middleware.LogInfo(c, "GetAssignableRoles called")
    // ... get user ID with better error logging ...

    userRolesData, err := h.rbacService.GetUserRoles(userID)
    if err != nil { /* handle error */ }

    // NEW: Handle empty roles
    if len(userRolesData) == 0 {
        c.JSON(http.StatusOK, []*models.Role{})
        return
    }

    // ... process assignable roles ...

    // NEW: Handle no assignable roles
    if len(assignableRoleNames) == 0 {
        c.JSON(http.StatusOK, []*models.Role{})
        return
    }

    // ... fetch from database ...

    // NEW: Ensure never nil
    if assignableRoles == nil {
        assignableRoles = []*models.Role{}
    }

    middleware.LogInfo(c, "Successfully retrieved assignable roles")
    c.JSON(http.StatusOK, assignableRoles)
}
```

---

## Need More Help?

📖 **Full deployment guide:** `DEPLOYMENT-FIX-500.md`

💻 **Diagnostic scripts:**
- Linux/macOS: `./scripts/diagnose-500-error.sh`
- Windows: `.\scripts\diagnose-500-error.ps1`

🔍 **Check logs:**
```bash
docker logs <backend-container-name> --tail=100 --follow
```

---

## Expected Log Output (After Fix)

**Success case:**
```
INFO: GetAssignableRoles called
INFO: Successfully retrieved assignable roles
```

**User has no roles:**
```
INFO: GetAssignableRoles called
INFO: User has no roles assigned
```

**User can't assign roles:**
```
INFO: GetAssignableRoles called
INFO: User has no assignable roles based on hierarchy
```

---

## Still Getting 500?

Provide these for debugging:
1. Output of `./scripts/diagnose-500-error.sh`
2. Last 100 lines of backend logs
3. Network request details from browser DevTools
4. Your user's roles from database:
   ```sql
   SELECT r.name FROM user_roles ur
   JOIN roles r ON ur.role_id = r.id
   WHERE ur.user_id = <your_id>;
   ```
