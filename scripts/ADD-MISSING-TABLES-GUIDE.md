# Adding Missing Tables to Production Database

## Overview
Your production database currently has 24 tables, but your application needs 40+ tables for full functionality including RBAC (roles/permissions), teams, custom domains, and collaboration features.

This guide will help you safely add the missing tables **without losing any existing data**.

## Prerequisites
- Access to Adminer (http://your-server:8081)
- Production database credentials
- 5-10 minutes

## Step-by-Step Instructions

### Step 1: Verify Current State

1. Open Adminer in your browser
2. Login to your production database:
   - **Server**: `postgres` (or `localhost`)
   - **Username**: `urlshortener`
   - **Password**: (your DB_PASSWORD)
   - **Database**: `urlshortener_db`

3. Click on **SQL command** in the left sidebar

4. Copy and paste the entire contents of `scripts/verify-tables.sql`

5. Click **Execute** to see:
   - Total number of tables (should show 24 currently)
   - Which critical tables are missing (marked with ✗ MISSING)
   - Current columns in url_mappings table

### Step 2: Backup (Recommended)

Before making changes, create a backup:

```bash
# Via Docker (if running locally)
docker exec -t urlshortener-postgres-1 pg_dump -U urlshortener urlshortener_db > backup_$(date +%Y%m%d_%H%M%S).sql

# Or via Adminer
# Click "Export" in the left sidebar, select all tables, and download
```

### Step 3: Apply Missing Tables

1. In Adminer, click on **SQL command** again

2. **Important**: Copy the **ENTIRE contents** of `scripts/add-missing-tables.sql`
   - This file is ~400 lines - make sure you copy everything
   - Do NOT run it line by line - paste the whole file at once

3. Click **Execute**

4. You should see: `All missing tables and columns have been added successfully!`

5. If you see any errors:
   - Don't panic! The script uses `CREATE TABLE IF NOT EXISTS`, so it's safe
   - Most errors are about tables that already exist (which is fine)
   - Only worry about foreign key errors or syntax errors

### Step 4: Verify Success

1. Run the `verify-tables.sql` script again

2. Confirm:
   - Total tables should now be **40+** (up from 24)
   - All critical tables should show **✓ EXISTS**
   - Roles count should show **5** (admin, user, premium, team_admin, team_member)
   - Permissions count should show **17+**

3. Click on **Tables** in left sidebar to see the full list:
   - ✅ roles
   - ✅ permissions
   - ✅ role_permissions
   - ✅ user_roles
   - ✅ oauth_providers
   - ✅ teams
   - ✅ team_members
   - ✅ team_invitations
   - ✅ custom_domains
   - ✅ domain_analytics
   - ✅ url_shares
   - ✅ url_comments
   - ✅ url_bookmarks
   - ✅ url_favorites
   - ✅ url_collections
   - ✅ url_collection_items
   - ✅ url_notes
   - ✅ url_activities
   - ✅ user_activity_logs
   - ✅ audit_logs

### Step 5: Test Admin Login

Now that RBAC tables exist, test admin functionality:

1. **Register a new user** (or use existing user):
   ```bash
   curl -X POST http://your-server:8080/api/v1/auth/register \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@example.com",
       "password": "securepassword",
       "name": "Admin User"
     }'
   ```

2. **Promote user to admin** via Adminer SQL command:
   ```sql
   -- Get the user ID
   SELECT id, email, name FROM users WHERE email = 'admin@example.com';

   -- Get the admin role ID
   SELECT id, name FROM roles WHERE name = 'admin';

   -- Assign admin role to user (replace USER_ID and ROLE_ID with actual values)
   INSERT INTO user_roles (user_id, role_id)
   VALUES (USER_ID, ROLE_ID)
   ON CONFLICT (user_id, role_id) DO NOTHING;
   ```

3. **Login and verify admin access**:
   ```bash
   curl -X POST http://your-server:8080/api/v1/auth/login \
     -H "Content-Type: application/json" \
     -d '{
       "email": "admin@example.com",
       "password": "securepassword"
     }'
   ```

## What Was Added

### RBAC System (5 tables)
- `roles` - User roles (admin, user, premium, etc.)
- `permissions` - Granular permissions (user.create, url.read, etc.)
- `role_permissions` - Links roles to permissions
- `user_roles` - Assigns roles to users
- `oauth_providers` - OAuth authentication support

### Team Management (3 tables)
- `teams` - Organizations/teams
- `team_members` - Team membership
- `team_invitations` - Pending invitations

### Custom Domains (2 tables)
- `custom_domains` - User's custom domains
- `domain_analytics` - Domain-level analytics

### Collaboration (8 tables)
- `url_shares` - Share URLs with users/teams
- `url_comments` - Comment on URLs
- `url_bookmarks` - Bookmark URLs
- `url_favorites` - Favorite URLs
- `url_collections` - Organize URLs in collections
- `url_collection_items` - URLs in collections
- `url_notes` - Private notes on URLs
- `url_activities` - Activity feed

### Analytics & Audit (2 tables)
- `user_activity_logs` - User activity tracking
- `audit_logs` - Security audit trail

### Extended URL Features
Added columns to `url_mappings`:
- `user_id`, `team_id` - Ownership
- `title`, `description`, `tags` - Metadata
- `is_public`, `password_hash` - Access control
- `custom_domain` - Custom domain support
- `utm_*` - UTM parameter tracking

## Safety Notes

✅ **Safe Operations** - This script:
- Uses `CREATE TABLE IF NOT EXISTS` - won't overwrite existing tables
- Uses `IF NOT EXISTS` checks for columns - won't duplicate columns
- Uses `ON CONFLICT DO NOTHING` for inserts - won't create duplicates
- Preserves all existing data
- Can be run multiple times safely

✅ **No Data Loss** - This script:
- Does NOT use `DROP TABLE`
- Does NOT use `DELETE` or `TRUNCATE`
- Only adds new tables and columns
- Existing URLs, users, and click events remain untouched

## Troubleshooting

### Error: "relation already exists"
- **Solution**: Ignore this - it means the table already exists, which is fine

### Error: "column already exists"
- **Solution**: Ignore this - it means the column was already added

### Error: "foreign key constraint"
- **Cause**: Referenced table doesn't exist yet
- **Solution**: Make sure you pasted the ENTIRE script, not just parts

### Error: "permission denied"
- **Cause**: Database user lacks permissions
- **Solution**: Run as superuser or grant permissions:
  ```sql
  GRANT ALL PRIVILEGES ON ALL TABLES IN SCHEMA public TO urlshortener;
  GRANT ALL PRIVILEGES ON ALL SEQUENCES IN SCHEMA public TO urlshortener;
  ```

### Tables still missing after running script
1. Check Adminer's SQL output for error messages
2. Verify you pasted the entire script (should be ~400 lines)
3. Try running sections separately if needed
4. Check PostgreSQL logs: `docker logs urlshortener-postgres-1`

## Alternative: Docker Command Method

If Adminer doesn't work, you can apply via command line:

```bash
# Copy script to container
docker cp scripts/add-missing-tables.sql urlshortener-postgres-1:/tmp/

# Execute script
docker exec -it urlshortener-postgres-1 psql -U urlshortener -d urlshortener_db -f /tmp/add-missing-tables.sql

# Verify
docker exec -it urlshortener-postgres-1 psql -U urlshortener -d urlshortener_db -f /tmp/verify-tables.sql
```

## Next Steps

After successfully adding tables:

1. ✅ Verify admin login works
2. ✅ Test API endpoints that require RBAC
3. ✅ Update application code to use new features
4. ✅ Document the new capabilities
5. ✅ Consider updating `configs/database.sql` to match production

## Preventing This in the Future

To ensure new databases get all tables automatically:

1. **Update `configs/database.sql`** to match what you just added
2. **For new deployments**: Fresh Docker volumes will get all tables
3. **For existing deployments**: Use these scripts to add missing tables
4. **Consider migration system**: Add proper migration tracking

## Questions?

- Check if tables exist: Run `verify-tables.sql`
- See table structure: Click table name in Adminer
- Count records: `SELECT COUNT(*) FROM table_name;`
- Rollback (if needed): Restore from backup

---

**Status Check**: After completing these steps, your production database will have all 40+ tables needed for full application functionality including RBAC-based admin access! 🎉
