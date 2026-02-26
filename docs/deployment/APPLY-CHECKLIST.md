# Database Update Checklist

Use this checklist to track your progress applying missing tables to production.

## 📋 Pre-Flight Checklist

- [ ] **Backup database** (optional but recommended)
  - Via Adminer: Click "Export" → Download SQL
  - Or via Docker: `docker exec -t urlshortener-postgres-1 pg_dump -U urlshortener urlshortener_db > backup.sql`

- [ ] **Verify current state**
  - Open Adminer: http://your-server:8081
  - Login to database
  - Run `scripts/verify-tables.sql`
  - Confirm you have ~24 tables (not 40+)

- [ ] **Check production access**
  - Can access Adminer ✓
  - Can login to database ✓
  - Can run SQL commands ✓

## 🚀 Application Steps (Choose ONE method)

### Method A: Adminer (Recommended)

- [ ] Open `scripts/add-missing-tables.sql` in text editor
- [ ] Copy **entire file** (all 400+ lines)
- [ ] Open Adminer → SQL command
- [ ] Paste SQL script
- [ ] Click "Execute"
- [ ] Check for success message
- [ ] Note any errors (most are safe)

### Method B: PowerShell Script (Windows)

- [ ] Open PowerShell
- [ ] Navigate to: `cd D:\github\URLshorter\scripts`
- [ ] Preview: `.\apply-missing-tables.ps1 -DryRun`
- [ ] Apply: `.\apply-missing-tables.ps1`
- [ ] Type "yes" to confirm
- [ ] Wait for completion (~30 seconds)

### Method C: Bash Script (Linux/Mac)

- [ ] Open terminal
- [ ] Navigate to: `cd scripts`
- [ ] Make executable: `chmod +x apply-missing-tables.sh`
- [ ] Preview: `./apply-missing-tables.sh --dry-run`
- [ ] Apply: `./apply-missing-tables.sh`
- [ ] Type "yes" to confirm
- [ ] Wait for completion (~30 seconds)

## ✅ Verification Steps

- [ ] Run `scripts/verify-tables.sql` in Adminer
- [ ] Confirm total tables = 40+ (was 24)
- [ ] Check RBAC tables exist:
  - [ ] `roles` table exists
  - [ ] `permissions` table exists
  - [ ] `user_roles` table exists
  - [ ] `role_permissions` table exists

- [ ] Check team tables exist:
  - [ ] `teams` table exists
  - [ ] `team_members` table exists

- [ ] Check collaboration tables exist:
  - [ ] `url_collections` table exists
  - [ ] `url_shares` table exists
  - [ ] `url_comments` table exists

- [ ] Verify data populated:
  - [ ] Roles table has 5 rows (admin, user, premium, team_admin, team_member)
  - [ ] Permissions table has 17+ rows
  - [ ] role_permissions table has rows (roles linked to permissions)

- [ ] Check url_mappings has new columns:
  - [ ] `user_id` column exists
  - [ ] `team_id` column exists
  - [ ] `title` column exists
  - [ ] `tags` column exists

## 🎉 Enable Admin Access

### Step 1: Register User
- [ ] Call register API endpoint
- [ ] Use your admin email
- [ ] Use strong password
- [ ] Save credentials securely

```bash
curl -X POST http://your-server:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "YourSecurePassword123!",
    "name": "Admin User"
  }'
```

### Step 2: Promote to Admin (via Adminer SQL)

- [ ] Get user ID:
```sql
SELECT id, email, name FROM users WHERE email = 'admin@example.com';
-- Note the ID: _____
```

- [ ] Get admin role ID:
```sql
SELECT id, name FROM roles WHERE name = 'admin';
-- Note the ID: _____
```

- [ ] Assign admin role:
```sql
INSERT INTO user_roles (user_id, role_id)
VALUES (YOUR_USER_ID, ADMIN_ROLE_ID);  -- Replace with actual numbers
```

- [ ] Verify assignment:
```sql
SELECT u.email, r.name as role
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@example.com';
-- Should show: admin@example.com | admin
```

### Step 3: Test Admin Login

- [ ] Call login API endpoint
```bash
curl -X POST http://your-server:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "YourSecurePassword123!"
  }'
```

- [ ] Verify response includes JWT token
- [ ] Verify token includes admin role/permissions
- [ ] Test admin-protected endpoint (if available)

## 📊 Final Verification

- [ ] Database has 40+ tables ✓
- [ ] All RBAC tables populated ✓
- [ ] Admin user created ✓
- [ ] Admin can login ✓
- [ ] No existing data lost ✓
- [ ] Application runs without errors ✓

## 🎯 Success Criteria

When all items are checked, you should have:
- ✅ Full database schema (40+ tables)
- ✅ RBAC system operational
- ✅ Admin user with full permissions
- ✅ Team management tables ready
- ✅ Collaboration features available
- ✅ Custom domain support enabled
- ✅ All existing data preserved

## 📝 Notes & Issues

Document any issues or observations here:

```
Date applied: _______________
Method used: _______________
Time taken: _______________
Tables before: _______________
Tables after: _______________

Issues encountered:
-
-

Resolution:
-
-

Final status: [ ] Success  [ ] Partial  [ ] Failed
```

## 🔄 If Something Goes Wrong

### Rollback (if needed)
- [ ] Stop application
- [ ] Restore from backup (if created)
- [ ] Or manually drop added tables (they're all new, safe to remove)
- [ ] Review error logs
- [ ] Try again or seek help

### Get Help
- [ ] Check `scripts/ADD-MISSING-TABLES-GUIDE.md` troubleshooting section
- [ ] Review Docker logs: `docker logs urlshortener-postgres-1`
- [ ] Check PostgreSQL logs in Adminer
- [ ] Verify container is running: `docker ps`

---

**Print this checklist or keep it open while applying changes!** ✓
