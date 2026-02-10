# Quick Start: Add Missing Database Tables

## 🎯 Goal
Add 15+ missing tables (RBAC, teams, domains, etc.) to your production database that currently has only 24 tables.

## ⚡ Fastest Method: Use Adminer

### Option 1: Adminer (Recommended - Visual & Safe)

1. **Open Adminer**: http://your-server:8081
2. **Login** with your database credentials
3. **Click "SQL command"**
4. **Copy & paste** entire `add-missing-tables.sql` file
5. **Execute**
6. **Verify**: Run `verify-tables.sql`

✅ Takes 2 minutes
✅ Visual feedback
✅ Can see results immediately

### Option 2: Automated Script (Windows)

```powershell
# Navigate to scripts directory
cd D:\github\URLshorter\scripts

# Preview changes (safe)
.\apply-missing-tables.ps1 -DryRun

# Apply changes
.\apply-missing-tables.ps1

# Verify success
.\apply-missing-tables.ps1 -Verify
```

### Option 3: Automated Script (Linux/Mac)

```bash
# Make script executable
chmod +x scripts/apply-missing-tables.sh

# Preview changes (safe)
./scripts/apply-missing-tables.sh --dry-run

# Apply changes
./scripts/apply-missing-tables.sh

# Verify success
./scripts/apply-missing-tables.sh --verify
```

### Option 4: Manual Docker Command

```bash
# Copy SQL file to container
docker cp scripts/add-missing-tables.sql urlshortener-postgres-1:/tmp/

# Execute SQL
docker exec -it urlshortener-postgres-1 \
  psql -U urlshortener -d urlshortener_db \
  -f /tmp/add-missing-tables.sql

# Verify
docker exec -it urlshortener-postgres-1 \
  psql -U urlshortener -d urlshortener_db \
  -f /tmp/verify-tables.sql
```

## 📊 Before & After

**Current State**: 24 tables
**After Update**: 40+ tables

**New Tables Added**:
- 🔐 **RBAC**: `roles`, `permissions`, `role_permissions`, `user_roles`, `oauth_providers`
- 👥 **Teams**: `teams`, `team_members`, `team_invitations`
- 🌐 **Domains**: `custom_domains`, `domain_analytics`
- 🤝 **Collaboration**: `url_shares`, `url_comments`, `url_bookmarks`, `url_favorites`, `url_collections`, `url_collection_items`, `url_notes`, `url_activities`
- 📊 **Analytics**: `user_activity_logs`, `audit_logs`

## ✅ Verify Success

Run this SQL in Adminer or psql:

```sql
-- Should return 40+
SELECT COUNT(*) FROM information_schema.tables
WHERE table_schema = 'public';

-- Should show "✓ EXISTS" for all
SELECT
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'roles')
         THEN '✓ EXISTS' ELSE '✗ MISSING' END AS roles_table,
    CASE WHEN EXISTS (SELECT 1 FROM information_schema.tables WHERE table_name = 'permissions')
         THEN '✓ EXISTS' ELSE '✗ MISSING' END AS permissions_table;
```

## 🎉 Enable Admin Login

Once tables are added, create your first admin:

```sql
-- 1. Register user first via API, then get user ID
SELECT id, email FROM users WHERE email = 'your-email@example.com';

-- 2. Get admin role ID
SELECT id FROM roles WHERE name = 'admin';

-- 3. Assign admin role (replace IDs)
INSERT INTO user_roles (user_id, role_id)
VALUES (YOUR_USER_ID, ADMIN_ROLE_ID);
```

## 🛡️ Safety Features

✅ Uses `CREATE TABLE IF NOT EXISTS` - won't overwrite existing tables
✅ Uses conditional column adds - won't duplicate columns
✅ Uses `ON CONFLICT DO NOTHING` - won't create duplicate data
✅ No `DROP` commands - existing data is safe
✅ Can be run multiple times safely

## 📚 Full Documentation

For detailed instructions, troubleshooting, and explanations:
- **Step-by-step guide**: `scripts/ADD-MISSING-TABLES-GUIDE.md`
- **SQL scripts**: `scripts/add-missing-tables.sql` & `scripts/verify-tables.sql`
- **Automation scripts**: `scripts/apply-missing-tables.{ps1,sh}`

## ❓ Common Issues

### "Container not found"
```bash
# List containers
docker ps -a

# Use correct container name
.\apply-missing-tables.ps1 -ContainerName "your-container-name"
```

### "Table already exists" errors
- ✅ This is NORMAL and SAFE - means table already exists
- Script will skip and continue

### Adminer won't load script
- Make sure to copy the **ENTIRE** file (400+ lines)
- Try pasting in sections if timeout occurs
- Or use command line method instead

## 🚀 Pro Tips

1. **Backup first**: Export database via Adminer before changes
2. **Use DryRun**: Test with `--dry-run` flag first
3. **Verify twice**: Run verification script before AND after
4. **Check logs**: Monitor Docker logs during execution

---

**Need Help?** Check the full guide: `scripts/ADD-MISSING-TABLES-GUIDE.md`
