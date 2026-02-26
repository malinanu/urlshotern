# Database Update Summary

## 🎯 Problem Solved

Your production database had **24 tables** but needed **40+ tables** for full functionality including:
- ❌ Missing RBAC tables (roles, permissions) → **Can't login as admin**
- ❌ Missing team management tables
- ❌ Missing custom domain tables
- ❌ Missing collaboration features

## ✅ Solution Provided

I've created a complete solution to safely add all missing tables to your production database **without losing any existing data**.

## 📁 Files Created

### 1. SQL Scripts (Core Files)

**`scripts/add-missing-tables.sql`** (400+ lines)
- Adds all 15+ missing tables
- Adds new columns to existing tables
- Creates indexes for performance
- Populates default roles and permissions
- **100% SAFE**: Uses `IF NOT EXISTS` checks everywhere

**`scripts/verify-tables.sql`**
- Checks which tables exist
- Shows table counts
- Verifies RBAC tables are populated
- Shows all columns in url_mappings

### 2. Automation Scripts

**`scripts/apply-missing-tables.ps1`** (PowerShell - Windows)
- Interactive script with safety prompts
- Dry-run mode: `.\apply-missing-tables.ps1 -DryRun`
- Apply: `.\apply-missing-tables.ps1`
- Verify: `.\apply-missing-tables.ps1 -Verify`

**`scripts/apply-missing-tables.sh`** (Bash - Linux/Mac)
- Same features as PowerShell version
- Dry-run: `./apply-missing-tables.sh --dry-run`
- Apply: `./apply-missing-tables.sh`
- Verify: `./apply-missing-tables.sh --verify`

### 3. Documentation

**`scripts/ADD-MISSING-TABLES-GUIDE.md`**
- Complete step-by-step instructions
- Detailed explanations
- Troubleshooting section
- Safety notes and best practices

**`scripts/QUICK-START.md`**
- Quick reference guide
- All 4 methods to apply changes
- Common issues and solutions

**`DATABASE-UPDATE-SUMMARY.md`** (This file)
- Overview of the solution
- Next steps

## 🚀 How to Apply (Choose ONE Method)

### Method 1: Adminer (Recommended - Easiest)

Since you're already using Adminer on your production server:

1. Open Adminer: http://your-server:8081
2. Login to `urlshortener_db`
3. Click **"SQL command"**
4. Open `scripts/add-missing-tables.sql` on your local machine
5. Copy the **ENTIRE contents** (all 400+ lines)
6. Paste into Adminer
7. Click **"Execute"**
8. Should see: `All missing tables and columns have been added successfully!`
9. Run `scripts/verify-tables.sql` to confirm

⏱️ **Time**: 2-3 minutes
✅ **Risk**: Zero (safe operations only)

### Method 2: PowerShell Script (Windows - Automated)

```powershell
cd D:\github\URLshorter\scripts

# Preview changes first (safe)
.\apply-missing-tables.ps1 -DryRun

# Apply changes
.\apply-missing-tables.ps1

# Verify
.\apply-missing-tables.ps1 -Verify
```

### Method 3: Bash Script (Linux/Mac - Automated)

```bash
cd /path/to/URLshorter/scripts

# Make executable (first time only)
chmod +x apply-missing-tables.sh

# Preview changes first (safe)
./apply-missing-tables.sh --dry-run

# Apply changes
./apply-missing-tables.sh

# Verify
./apply-missing-tables.sh --verify
```

### Method 4: Manual Docker Command

```bash
# Copy SQL to container
docker cp scripts/add-missing-tables.sql urlshortener-postgres-1:/tmp/

# Execute
docker exec -it urlshortener-postgres-1 \
  psql -U urlshortener -d urlshortener_db \
  -f /tmp/add-missing-tables.sql
```

## 📊 What Will Be Added

### RBAC System (5 tables)
- ✅ `roles` - System roles (admin, user, premium, team_admin, team_member)
- ✅ `permissions` - Granular permissions (17 default permissions)
- ✅ `role_permissions` - Maps roles to permissions
- ✅ `user_roles` - Assigns roles to users
- ✅ `oauth_providers` - OAuth provider connections

### Team Management (3 tables)
- ✅ `teams` - Organizations/workspaces
- ✅ `team_members` - Team membership with roles
- ✅ `team_invitations` - Pending team invites

### Custom Domains (2 tables)
- ✅ `custom_domains` - User custom domains with verification
- ✅ `domain_analytics` - Per-domain analytics

### Collaboration (8 tables)
- ✅ `url_shares` - Share URLs with users/teams
- ✅ `url_comments` - Comment threads on URLs
- ✅ `url_bookmarks` - User bookmarks
- ✅ `url_favorites` - Favorite URLs
- ✅ `url_collections` - Organize URLs in folders
- ✅ `url_collection_items` - URLs in collections
- ✅ `url_notes` - Private notes on URLs
- ✅ `url_activities` - Activity feed/timeline

### Analytics & Audit (2 tables)
- ✅ `user_activity_logs` - Detailed user activity tracking
- ✅ `audit_logs` - Security audit trail

### Extended URL Features
Adds 13 new columns to `url_mappings`:
- `user_id`, `team_id` - Ownership tracking
- `title`, `description`, `tags` - Rich metadata
- `is_public`, `password_hash` - Access control
- `custom_domain` - Custom domain support
- `utm_source`, `utm_medium`, `utm_campaign`, `utm_term`, `utm_content` - Marketing tracking

## 🎉 After Applying: Enable Admin Login

Once tables are added, create your first admin user:

### Step 1: Register User (via API)

```bash
curl -X POST http://your-server:8080/api/v1/auth/register \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePassword123!",
    "name": "Admin User"
  }'
```

### Step 2: Promote to Admin (via Adminer or psql)

```sql
-- Get user ID
SELECT id, email, name FROM users WHERE email = 'admin@example.com';

-- Get admin role ID
SELECT id, name FROM roles WHERE name = 'admin';

-- Assign admin role (replace with actual IDs)
INSERT INTO user_roles (user_id, role_id)
VALUES (1, 1)  -- Replace with your actual IDs
ON CONFLICT (user_id, role_id) DO NOTHING;

-- Verify
SELECT u.email, r.name as role
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.email = 'admin@example.com';
```

### Step 3: Login and Test

```bash
curl -X POST http://your-server:8080/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "admin@example.com",
    "password": "SecurePassword123!"
  }'
```

Should return JWT token with admin permissions!

## 🛡️ Safety Guarantees

This solution is **100% SAFE** because:

✅ **No data deletion**: Zero `DROP`, `DELETE`, or `TRUNCATE` commands
✅ **Idempotent**: Can be run multiple times safely
✅ **Non-destructive**: Only creates new tables and columns
✅ **Existing data preserved**: All URLs, users, clicks remain untouched
✅ **Conflict-safe**: Uses `IF NOT EXISTS` and `ON CONFLICT DO NOTHING`
✅ **Rollback-friendly**: All changes are additive only

**Worst case scenario**: Some tables already exist → Script skips them (safe)

## 📈 Expected Results

### Before
```
Total tables: 24
- url_mappings ✓
- click_events ✓
- users ✓
- user_sessions ✓
- roles ✗ MISSING
- permissions ✗ MISSING
- teams ✗ MISSING
... (15+ more missing)
```

### After
```
Total tables: 40+
- All original 24 tables ✓
- roles ✓ (with 5 default roles)
- permissions ✓ (with 17 default permissions)
- role_permissions ✓ (populated)
- teams ✓
- custom_domains ✓
- url_collections ✓
... (all tables present)
```

## 🔍 Verification Commands

### Quick Check (via Adminer)
```sql
SELECT COUNT(*) FROM information_schema.tables WHERE table_schema = 'public';
-- Should return 40+
```

### Detailed Check (use verify-tables.sql)
```bash
# Via script
.\apply-missing-tables.ps1 -Verify

# Via docker
docker exec -it urlshortener-postgres-1 \
  psql -U urlshortener -d urlshortener_db \
  -f /tmp/verify-tables.sql
```

## 🔧 Troubleshooting

### "relation already exists" error
✅ **This is NORMAL** - means table already exists, script will continue

### "column already exists" error
✅ **This is NORMAL** - means column was already added

### Script times out in Adminer
- Copy-paste in smaller sections
- Or use command line method instead

### Container name not found
```bash
# Find your container name
docker ps -a | grep postgres

# Use correct name
.\apply-missing-tables.ps1 -ContainerName "your-actual-container-name"
```

## 📚 Documentation Reference

- **Quick Start**: `scripts/QUICK-START.md`
- **Full Guide**: `scripts/ADD-MISSING-TABLES-GUIDE.md`
- **SQL Scripts**: `scripts/add-missing-tables.sql`, `scripts/verify-tables.sql`
- **Automation**: `scripts/apply-missing-tables.{ps1,sh}`

## ✨ Next Steps

1. ✅ **Apply the changes** (choose your preferred method above)
2. ✅ **Verify success** (run verify script)
3. ✅ **Create admin user** (register + promote to admin)
4. ✅ **Test admin login** (verify RBAC works)
5. ✅ **Update application code** (if needed to use new features)
6. ✅ **Document for team** (share this guide)

## 🎓 Prevention for Future

To ensure new deployments get all tables automatically:

1. The `configs/database.sql` already contains all table definitions
2. New Docker deployments with fresh volumes will have all tables
3. Existing deployments can use these scripts to catch up
4. Consider implementing proper migration system later (optional)

---

## 📞 Support

If you encounter issues:
1. Check `scripts/ADD-MISSING-TABLES-GUIDE.md` troubleshooting section
2. Run verification script to diagnose
3. Check Docker logs: `docker logs urlshortener-postgres-1`
4. Verify container is running: `docker ps`

---

**Ready to proceed?** Open `scripts/QUICK-START.md` and choose your method! 🚀
