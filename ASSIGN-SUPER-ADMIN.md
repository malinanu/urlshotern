# Quick Start: Assign Super Administrator Role

## Problem

You need to assign the `super_admin` role to user ID `112125335771807740`, but the API returns:

```
400 Bad Request: insufficient privileges to assign role 'super_admin'
```

**Root Cause:** Only super_admins can assign super_admin roles (chicken-and-egg problem).

**Solution:** Use database migration 033 to bootstrap the first super administrator.

---

## Quick Start (5 Minutes)

### 1. Verify Prerequisites

**Windows:**
```powershell
.\scripts\verify-super-admin-prereqs.ps1
```

**Linux/macOS:**
```bash
chmod +x scripts/verify-super-admin-prereqs.sh
./scripts/verify-super-admin-prereqs.sh
```

This checks:
- ✓ PostgreSQL container is running
- ✓ Database connection works
- ✓ User 112125335771807740 exists
- ✓ super_admin role exists
- ✓ User doesn't already have the role

### 2. Commit & Push Migration (Development)

```bash
git add configs/migrations/033_*.sql
git commit -m "feat: Add migration to assign initial super_admin role"
git push origin main
```

### 3. Deploy to Production

SSH into your Oracle Ubuntu server:

```bash
# SSH to server
ssh user@your-oracle-server

# Pull latest code
cd /path/to/URLshorter
git pull origin main

# Verify migration files
ls -la configs/migrations/033_*

# Apply migration
docker exec -i $(docker ps -q -f name=postgres) \
  psql -U urlshortener -d urlshortener_db \
  < configs/migrations/033_assign_initial_super_admin_up.sql
```

Expected output:
```
INSERT 0 1
NOTICE: Successfully assigned super_admin role to user 112125335771807740
```

### 4. Verify Assignment

```bash
docker exec -i $(docker ps -q -f name=postgres) \
  psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT u.name, u.email, r.name as role
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE u.id = 112125335771807740 AND r.name = 'super_admin';
EOF
```

Should show the super_admin role assigned.

### 5. Refresh Your Session

**CRITICAL:** Log out and log back in to get a fresh JWT token:

1. Go to https://trunc.site/admin
2. Click **Log out**
3. **Log back in**
4. Your new super_admin permissions are now active ✅

### 6. Test

Navigate to Users page → Try assigning super_admin role → Should succeed!

---

## Files Created

- `configs/migrations/033_assign_initial_super_admin_up.sql` - Migration to assign role
- `configs/migrations/033_assign_initial_super_admin_down.sql` - Rollback migration
- `scripts/verify-super-admin-prereqs.sh` - Prerequisites check (Linux/macOS)
- `scripts/verify-super-admin-prereqs.ps1` - Prerequisites check (Windows)
- `SUPER-ADMIN-DEPLOYMENT.md` - Comprehensive deployment guide

---

## Rollback (If Needed)

```bash
docker exec -i $(docker ps -q -f name=postgres) \
  psql -U urlshortener -d urlshortener_db \
  < configs/migrations/033_assign_initial_super_admin_down.sql
```

---

## Need More Details?

See `SUPER-ADMIN-DEPLOYMENT.md` for:
- Complete troubleshooting guide
- Alternative deployment methods
- Security considerations
- Post-deployment checklist

---

## Why This Works

- **Problem:** RBAC validation prevents admins from assigning super_admin roles
- **Solution:** Migration bypasses application layer, assigns directly in database
- **Safety:** Idempotent (won't fail if already assigned), reversible, version controlled
- **Best Practice:** Bootstrap data via migrations, not manual SQL

---

**Time Required:** 5 minutes
**Risk Level:** Low (idempotent, reversible)
**Recommended:** Yes, this is the proper way to bootstrap super admins
