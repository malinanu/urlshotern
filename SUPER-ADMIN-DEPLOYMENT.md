# Super Administrator Assignment - Deployment Guide

## Overview

This guide walks you through assigning the Super Administrator role to user ID `112125335771807740` using database migration 033.

**Why a migration?**
- Version controlled and documented
- Repeatable and idempotent
- Can be rolled back if needed
- Best practice for bootstrap/seed data

---

## Prerequisites

1. User ID `112125335771807740` must exist in the database
2. The `super_admin` role must exist (created in migration 030)
3. Access to your production Oracle Ubuntu server
4. Docker running with PostgreSQL container

---

## Deployment Steps

### Step 1: Commit and Push Migration Files

First, let's commit the new migration files to git:

```bash
# Add migration files
git add configs/migrations/033_assign_initial_super_admin_up.sql
git add configs/migrations/033_assign_initial_super_admin_down.sql

# Commit
git commit -m "feat: Add migration to assign initial super_admin role

- Migration 033 assigns super_admin to user 112125335771807740
- Idempotent with ON CONFLICT DO NOTHING
- Includes verification and rollback migration
- Required to bootstrap first super admin (chicken-and-egg problem)"

# Push to repository
git push origin main
```

### Step 2: Deploy to Production Server

SSH into your Oracle Ubuntu server and pull the latest changes:

```bash
# SSH into production
ssh user@your-oracle-server

# Navigate to project directory
cd /path/to/URLshorter

# Pull latest changes
git pull origin main

# Verify migration files are present
ls -la configs/migrations/033_*
```

Expected output:
```
-rw-r--r-- 1 user user  XXX Feb 15 XX:XX 033_assign_initial_super_admin_down.sql
-rw-r--r-- 1 user user  XXX Feb 15 XX:XX 033_assign_initial_super_admin_up.sql
```

### Step 3: Check Migration Status

Before applying, check the current migration status:

```bash
# Using Docker (recommended for production)
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT version, applied_at, description
FROM schema_migrations
ORDER BY version DESC
LIMIT 5;
EOF
```

Expected: Latest version should be 032 or earlier.

### Step 4: Apply Migration 033

Run the migration:

```bash
# Option A: Using migration script (if available in container)
docker exec -i urlshorter-app /app/scripts/migrate.sh up 033

# Option B: Direct SQL execution via Docker
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs/migrations/033_assign_initial_super_admin_up.sql
```

Expected output:
```
INSERT 0 1
NOTICE:  Successfully assigned super_admin role to user 112125335771807740
```

**Possible outputs:**
- `INSERT 0 1` - Role successfully assigned
- `INSERT 0 0` - Role already assigned (migration is idempotent)
- `WARNING: Super admin role not assigned...` - User or role doesn't exist (see troubleshooting)

### Step 5: Verify Migration Applied

Check that migration 033 is recorded:

```bash
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT version, applied_at, description
FROM schema_migrations
WHERE version = '033'
ORDER BY version DESC;
EOF
```

Expected:
```
 version |        applied_at         |             description
---------+---------------------------+-------------------------------------
 033     | 2026-02-15 XX:XX:XX+00    | assign initial super admin
```

### Step 6: Verify Role Assignment

Confirm the user now has super_admin role:

```bash
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT
    u.id,
    u.name,
    u.email,
    r.name as role_name,
    r.display_name,
    ur.assigned_at,
    ur.assigned_by
FROM user_roles ur
JOIN users u ON ur.user_id = u.id
JOIN roles r ON ur.role_id = r.id
WHERE u.id = 112125335771807740;
EOF
```

Expected output:
```
        id         |   name    |      email       | role_name   |     display_name     |        assigned_at         | assigned_by
-------------------+-----------+------------------+-------------+---------------------+---------------------------+-------------
 112125335771807740| Your Name | your@email.com   | super_admin | Super Administrator | 2026-02-15 XX:XX:XX+00   | NULL
```

**Key points:**
- `role_name` should be `super_admin`
- `assigned_by` should be `NULL` (indicates bootstrap assignment)

### Step 7: Refresh User Session

**CRITICAL:** Log out and log back in to get a fresh JWT token:

1. Go to https://trunc.site/admin
2. Click **Log out**
3. **Log back in** with the same credentials
4. Your JWT token will now include the super_admin role

### Step 8: Test Super Admin Permissions

Verify you can now assign roles:

1. Navigate to **Users** page in admin panel
2. Select any user
3. Try assigning the `super_admin` role
4. Should succeed without "insufficient privileges" error ✅

---

## Rollback Procedure (If Needed)

If you need to remove the super_admin assignment:

```bash
# Rollback migration 033
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs/migrations/033_assign_initial_super_admin_down.sql
```

This will:
- Remove the super_admin role from user 112125335771807740
- Only removes assignments where `assigned_by IS NULL` (bootstrap assignments)
- Won't affect manually re-assigned roles

---

## Troubleshooting

### Issue: "Super admin role not assigned" Warning

**Cause:** User ID or super_admin role doesn't exist

**Solution:** Verify prerequisites:

```bash
# Check if user exists
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT id, name, email FROM users WHERE id = 112125335771807740;
EOF

# Check if super_admin role exists
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db << 'EOF'
SELECT id, name, display_name FROM roles WHERE name = 'super_admin';
EOF
```

**If super_admin role missing:**
```bash
# Apply migration 030 first
docker exec -i $(docker ps -q -f name=postgres) psql -U urlshortener -d urlshortener_db < configs/migrations/030_add_cms_permissions_up.sql
```

**If user missing:**
- The user needs to be created first through normal registration
- Or check if you have the correct user ID

### Issue: Still Getting "Insufficient Privileges" Error

**Cause:** JWT token not refreshed

**Solution:**
1. Log out completely from admin panel
2. Clear browser cache/cookies for trunc.site
3. Log back in to get fresh JWT with new role

### Issue: Migration Already Applied

**Symptom:** `INSERT 0 0` instead of `INSERT 0 1`

**This is normal!** The migration is idempotent - it won't fail if already applied.

---

## Post-Deployment Checklist

After successful deployment:

- [ ] Migration 033 shows in `schema_migrations` table
- [ ] User 112125335771807740 has super_admin role in `user_roles` table
- [ ] User can log out and log back in successfully
- [ ] User can access all admin panel pages
- [ ] User can assign super_admin role to other users via UI
- [ ] User can manage all RBAC permissions
- [ ] Document who has super_admin access in team docs

---

## Security Notes

1. **Audit Trail:** The `assigned_by = NULL` indicates this was a system/bootstrap assignment
2. **Limited Scope:** Only assigns to one specific user ID
3. **Reversible:** Can be rolled back using the down migration
4. **Future Assignments:** After this, all super_admin assignments should go through the API (proper audit trail)

---

## Alternative: Manual Assignment (Not Recommended)

If migrations aren't working, you can assign manually via SQL:

```sql
INSERT INTO user_roles (user_id, role_id, assigned_by, assigned_at)
VALUES (112125335771807740, 1, NULL, NOW())
ON CONFLICT (user_id, role_id) DO NOTHING;
```

**However, this is NOT recommended because:**
- Not version controlled
- Not documented in migration history
- Harder to track and audit

---

## Summary

**Time Required:** ~5 minutes

**Risk Level:** Low (idempotent, reversible)

**Recommended Method:**
1. Git commit and push migration files ✅
2. SSH to production server ✅
3. Pull latest code ✅
4. Apply migration 033 via Docker ✅
5. Verify assignment ✅
6. Log out and back in ✅
7. Test super admin permissions ✅

**Result:** User 112125335771807740 becomes the first Super Administrator, breaking the chicken-and-egg problem!
