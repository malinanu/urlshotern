# RBAC Migration Guide

This guide explains how to apply migrations 030, 031, and 032 which implement the complete RBAC system.

## Migrations Overview

- **030_add_cms_permissions** - Adds CMS permissions, new roles (super_admin, content_manager, content_editor, moderator), and role-permission mappings
- **031_migrate_admins_to_rbac** - Auto-migrates existing `is_admin=true` users to RBAC admin role
- **032_rbac_audit_log** - Creates audit log table for tracking all RBAC actions

## Option 1: Production Server (Oracle Ubuntu with Portainer)

### Step 1: Access the Database Container

```bash
# SSH into your Oracle Ubuntu server
ssh user@your-oracle-server

# Find the database container name
docker ps | grep postgres

# Access the database container
docker exec -it <postgres-container-name> psql -U urlshortener_user -d urlshortener_db
```

### Step 2: Apply Migrations Manually

```sql
-- Check current migration status
SELECT version, applied_at FROM schema_migrations ORDER BY version DESC LIMIT 10;

-- Apply Migration 030 (CMS Permissions and Roles)
\i /path/to/migrations/030_add_cms_permissions_up.sql

-- Apply Migration 031 (Migrate Admins to RBAC)
\i /path/to/migrations/031_migrate_admins_to_rbac_up.sql

-- Apply Migration 032 (Audit Log)
\i /path/to/migrations/032_rbac_audit_log_up.sql

-- Verify migrations were successful
SELECT version, applied_at FROM schema_migrations ORDER BY version DESC LIMIT 5;

-- Verify new tables and data
SELECT name, display_name FROM roles;
SELECT COUNT(*) FROM permissions WHERE resource IN ('cms', 'rbac', 'users');
\d rbac_audit_log
```

### Step 3: Using the Migration Script (Recommended)

```bash
# Copy migration files to server
scp configs/migrations/030_*.sql user@server:/path/to/urlshortener/configs/migrations/
scp configs/migrations/031_*.sql user@server:/path/to/urlshortener/configs/migrations/
scp configs/migrations/032_*.sql user@server:/path/to/urlshortener/configs/migrations/

# SSH into server
ssh user@your-oracle-server
cd /path/to/urlshortener

# Set environment variables (adjust as needed)
export DB_HOST=localhost
export DB_PORT=5432
export DB_NAME=urlshortener_db
export DB_USER=urlshortener_user
export DB_PASSWORD=your_password

# Check migration status
./scripts/migrate.sh status

# Apply migrations one by one
./scripts/migrate.sh up 030
./scripts/migrate.sh up 031
./scripts/migrate.sh up 032

# Or via Docker exec
docker exec -it urlshortener-app /app/scripts/migrate.sh up 030
docker exec -it urlshortener-app /app/scripts/migrate.sh up 031
docker exec -it urlshortener-app /app/scripts/migrate.sh up 032
```

## Option 2: Local Development (Docker Compose)

### Step 1: Start Local Environment

```bash
# Start Docker services
docker-compose -f docker/docker-compose.local.yml up -d

# Wait for database to be ready
sleep 5
```

### Step 2: Apply Migrations

```bash
# Using Docker exec
docker-compose -f docker/docker-compose.local.yml exec db psql -U urlshortener -d urlshortener_db -f /migrations/030_add_cms_permissions_up.sql
docker-compose -f docker/docker-compose.local.yml exec db psql -U urlshortener -d urlshortener_db -f /migrations/031_migrate_admins_to_rbac_up.sql
docker-compose -f docker/docker-compose.local.yml exec db psql -U urlshortener -d urlshortener_db -f /migrations/032_rbac_audit_log_up.sql

# Or connect to database and run manually
docker-compose -f docker/docker-compose.local.yml exec db psql -U urlshortener -d urlshortener_db
```

## Verification Steps

After applying migrations, verify the changes:

```sql
-- 1. Check roles were created
SELECT id, name, display_name, is_system
FROM roles
WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');

-- Expected: 4 rows

-- 2. Check CMS permissions were created
SELECT COUNT(*) FROM permissions WHERE resource = 'cms';
-- Expected: ~13 permissions

-- 3. Check RBAC permissions were created
SELECT COUNT(*) FROM permissions WHERE resource = 'rbac';
-- Expected: ~6 permissions

-- 4. Check existing admins were migrated
SELECT u.email, r.name as role
FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.is_admin = true;

-- Expected: All admins should have 'admin' role

-- 5. Check audit log table exists
\d rbac_audit_log
-- Expected: Table with columns id, user_id, target_user_id, action, role_id, etc.

-- 6. Check role-permission assignments
SELECT r.name, COUNT(rp.permission_id) as permission_count
FROM roles r
LEFT JOIN role_permissions rp ON r.id = rp.role_id
WHERE r.name IN ('super_admin', 'admin', 'content_manager', 'content_editor', 'moderator')
GROUP BY r.name
ORDER BY r.name;

-- Expected counts:
-- super_admin: ALL permissions (40+)
-- admin: ~30 permissions
-- content_manager: ~10 permissions
-- content_editor: ~6 permissions
-- moderator: ~5 permissions
```

## Rollback (If Needed)

If you need to rollback the migrations:

```bash
# Rollback in reverse order
./scripts/migrate.sh down 032
./scripts/migrate.sh down 031
./scripts/migrate.sh down 030
```

Or manually:

```sql
\i configs/migrations/032_rbac_audit_log_down.sql
\i configs/migrations/031_migrate_admins_to_rbac_down.sql
\i configs/migrations/030_add_cms_permissions_down.sql
```

## Troubleshooting

### Migration fails with "permission denied"
- Ensure the database user has CREATE TABLE and INSERT privileges
- Run as superuser: `ALTER USER urlshortener_user WITH SUPERUSER;`

### "role already exists" error
- The migrations use `ON CONFLICT DO NOTHING`, so this shouldn't happen
- If it does, check if migrations were partially applied: `SELECT * FROM schema_migrations;`

### Existing admins not migrated
- Check `is_admin` column: `SELECT id, email, is_admin FROM users WHERE is_admin = true;`
- Manually run migration 031 again (it's idempotent)

### Audit log not working
- Verify table exists: `\dt rbac_audit_log`
- Check indexes: `\di rbac_audit_log*`
- Test insert: `INSERT INTO rbac_audit_log (action, created_at) VALUES ('test', NOW());`

## Post-Migration Checklist

After successful migration:

- [ ] All 3 migrations applied successfully
- [ ] 4 new roles created (super_admin, content_manager, content_editor, moderator)
- [ ] CMS permissions created (13+)
- [ ] RBAC permissions created (6+)
- [ ] Existing admin users have 'admin' role assigned
- [ ] All regular users have 'user' role assigned
- [ ] Audit log table created with indexes
- [ ] Role-permission mappings exist for all roles
- [ ] Test permission checking: `SELECT * FROM role_permissions WHERE role_id = 1 LIMIT 5;`

## Next Steps

After migrations are applied:

1. Deploy the new backend code with RBAC service implementation
2. Test admin user creation via API: `POST /api/v1/admin/users/create`
3. Test role assignment: `POST /api/v1/admin/users/:id/roles/assign`
4. Verify audit log: `GET /api/v1/admin/rbac/audit`
5. Update frontend to use new RBAC endpoints

## Support

If you encounter issues:
1. Check migration logs: `SELECT * FROM schema_migrations ORDER BY applied_at DESC;`
2. Review database logs: `docker logs <postgres-container> --tail 100`
3. Verify database connection: `docker exec <postgres-container> pg_isready`
4. Check application logs after deployment
