# RBAC System Deployment Guide

## Production Environment: Oracle Ubuntu Server with Portainer

This guide will help you deploy the RBAC system to production at https://trunc.site/admin

---

## Prerequisites Checklist

- [ ] SSH access to Oracle Ubuntu server
- [ ] Admin access to Portainer
- [ ] Database backup completed (CRITICAL - do this first!)
- [ ] Admin JWT token for testing

---

## Step 1: Backup Database (CRITICAL)

**IMPORTANT**: Always backup before running migrations!

```bash
# SSH into your server
ssh user@your-oracle-server

# Create backup
docker exec -it urlshortener-db pg_dump -U urlshortener_user urlshortener_db > /tmp/db_backup_$(date +%Y%m%d_%H%M%S).sql

# Verify backup was created
ls -lh /tmp/db_backup_*
```

---

## Step 2: Pull Latest Code

```bash
# Navigate to your application directory
cd /path/to/urlshortener

# Pull latest code from GitHub
git pull origin main

# Verify you're on the latest commit (93d20ba)
git log --oneline -1
# Should show: 93d20ba feat: Complete RBAC system integration into main application
```

---

## Step 3: Run Migrations

### Check Current Migration Status

```bash
# Find your app container name
docker ps | grep urlshortener

# Check current migration version
docker exec -it <app-container-name> ./scripts/migrate.sh status
```

### Apply Migration 030: CMS Permissions and New Roles

```bash
docker exec -it <app-container-name> ./scripts/migrate.sh up 030
```

**Expected Output:**
```
✓ Migration 030 applied successfully
- Added 4 new roles (super_admin, content_manager, content_editor, moderator)
- Added 30+ permissions for CMS, RBAC, users, analytics
- Created role-permission mappings
```

**Verification Query:**
```bash
docker exec -it <db-container-name> psql -U urlshortener_user -d urlshortener_db -c \
  "SELECT name, display_name FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');"
```

**Expected Result:**
```
      name       |      display_name
-----------------+------------------------
 super_admin     | Super Administrator
 content_manager | Content Manager
 content_editor  | Content Editor
 moderator       | Moderator
(4 rows)
```

### Apply Migration 031: Migrate Existing Admins

```bash
docker exec -it <app-container-name> ./scripts/migrate.sh up 031
```

**Expected Output:**
```
✓ Migration 031 applied successfully
- Migrated existing is_admin users to RBAC
- Assigned 'admin' role to current admin users
- Assigned 'user' role to users without roles
```

**Verification Query:**
```bash
docker exec -it <db-container-name> psql -U urlshortener_user -d urlshortener_db -c \
  "SELECT u.email, r.name as role
   FROM users u
   JOIN user_roles ur ON u.id = ur.user_id
   JOIN roles r ON ur.role_id = r.id
   WHERE u.is_admin = true;"
```

**Expected Result:**
```
       email        |  role
--------------------+--------
 your-admin@email   | admin
(X rows)
```

### Apply Migration 032: RBAC Audit Log

```bash
docker exec -it <app-container-name> ./scripts/migrate.sh up 032
```

**Expected Output:**
```
✓ Migration 032 applied successfully
- Created rbac_audit_log table
- Added indexes for performance
```

**Verification Query:**
```bash
docker exec -it <db-container-name> psql -U urlshortener_user -d urlshortener_db -c \
  "\d rbac_audit_log"
```

**Expected Result:**
```
Table "public.rbac_audit_log"
     Column      |           Type
-----------------+--------------------------
 id              | bigint
 user_id         | bigint
 target_user_id  | bigint
 action          | character varying(50)
 role_id         | bigint
 permission_id   | bigint
 assigned_by     | bigint
 details         | jsonb
 ip_address      | inet
 user_agent      | text
 created_at      | timestamp with time zone
```

---

## Step 4: Rebuild and Deploy Application

### Option A: Docker Compose (Recommended)

```bash
# Navigate to docker directory
cd /path/to/urlshortener/docker

# Rebuild the application
docker-compose -f docker-compose.yml build app

# Restart the application (zero-downtime if configured)
docker-compose -f docker-compose.yml up -d app

# Check logs for any errors
docker-compose -f docker-compose.yml logs -f app
```

### Option B: Direct Docker Command

```bash
# Find your app container
docker ps | grep urlshortener-app

# Rebuild image
docker build -t urlshortener:latest .

# Restart container
docker restart <app-container-name>

# Check logs
docker logs -f <app-container-name>
```

### Option C: Via Portainer

1. Open Portainer at https://your-server:9000
2. Navigate to Containers
3. Find `urlshortener-app` container
4. Click "Recreate" button
5. Check "Pull latest image" if using registry
6. Click "Deploy the container"
7. Monitor logs in Portainer console

---

## Step 5: Verify RBAC Functionality

### Get Admin JWT Token

First, you need an admin JWT token. Login to your admin panel:

```bash
# Login and get token
curl -X POST https://trunc.site/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "your-admin@email.com",
    "password": "your-password"
  }'

# Extract the token from response
export TOKEN="<your-jwt-token-here>"
```

### Test 1: List All Roles

```bash
curl https://trunc.site/api/v1/admin/rbac/roles \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected Response:**
```json
{
  "roles": [
    {
      "id": 1,
      "name": "super_admin",
      "display_name": "Super Administrator",
      "description": "Full system access including RBAC management",
      "is_system": true
    },
    {
      "id": 2,
      "name": "admin",
      "display_name": "Administrator",
      "description": "Administrative access to manage users and content"
    },
    {
      "id": 3,
      "name": "content_manager",
      "display_name": "Content Manager",
      "description": "Full CMS access - create, edit, delete, and publish content"
    },
    {
      "id": 4,
      "name": "content_editor",
      "display_name": "Content Editor",
      "description": "CMS editing access - create and edit content (no publish/delete)"
    },
    {
      "id": 5,
      "name": "moderator",
      "display_name": "Moderator",
      "description": "User moderation and analytics viewing access"
    }
  ]
}
```

### Test 2: List All Permissions

```bash
curl https://trunc.site/api/v1/admin/rbac/permissions \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected**: 40+ permissions including cms.pages.*, rbac.*, users.*, etc.

### Test 3: Get Your User's Roles and Permissions

```bash
# Replace USER_ID with your user ID
curl https://trunc.site/api/v1/admin/users/<USER_ID>/roles \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected Response:**
```json
{
  "roles": [
    {
      "id": 2,
      "name": "admin",
      "display_name": "Administrator"
    }
  ],
  "permissions": [
    {
      "id": 1,
      "name": "users.create",
      "resource": "users",
      "action": "create"
    },
    // ... more permissions
  ]
}
```

### Test 4: Create a Content Editor User

```bash
curl -X POST https://trunc.site/api/v1/admin/users/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Content Editor",
    "email": "editor@example.com",
    "account_type": "free",
    "role_ids": [4],
    "email_verified": true,
    "is_active": true,
    "force_password_reset": true
  }' | jq
```

**Expected Response:**
```json
{
  "user": {
    "id": 123456789,
    "name": "Test Content Editor",
    "email": "editor@example.com",
    "is_active": true,
    "is_email_verified": true,
    "account_type": "free"
  },
  "roles": [
    {
      "id": 4,
      "name": "content_editor",
      "display_name": "Content Editor"
    }
  ],
  "temp_password": "RandomlyGenerated123!"
}
```

### Test 5: View Audit Log

```bash
curl "https://trunc.site/api/v1/admin/rbac/audit?limit=10" \
  -H "Authorization: Bearer $TOKEN" | jq
```

**Expected**: List of recent RBAC actions including the user creation from Test 4

---

## Step 6: Test Permission Enforcement

### Test CMS Publish Permission (Should Fail for Content Editor)

First, login as the content editor user you just created:

```bash
# Login as content editor
curl -X POST https://trunc.site/api/v1/auth/login \
  -H "Content-Type: application/json" \
  -d '{
    "email": "editor@example.com",
    "password": "<temp-password-from-test-4>"
  }'

# Save the editor token
export EDITOR_TOKEN="<editor-jwt-token>"

# Try to publish a page (should fail with 403)
curl -X PUT https://trunc.site/api/v1/admin/cms/pages/1/publish \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"is_published": true}'
```

**Expected Response (403 Forbidden):**
```json
{
  "error": "permission_denied",
  "message": "Required permission: cms.pages.publish"
}
```

### Test CMS Edit Permission (Should Succeed for Content Editor)

```bash
curl -X PUT https://trunc.site/api/v1/admin/cms/pages/1 \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "title": "Updated by Content Editor"
  }'
```

**Expected Response (200 OK):**
```json
{
  "page": {
    "id": 1,
    "title": "Updated by Content Editor",
    // ... other page fields
  }
}
```

---

## Rollback Plan (If Something Goes Wrong)

### Restore Database Backup

```bash
# Stop application
docker stop <app-container-name>

# Restore database
docker exec -i <db-container-name> psql -U urlshortener_user urlshortener_db < /tmp/db_backup_YYYYMMDD_HHMMSS.sql

# Rollback migrations (in reverse order)
docker exec -it <app-container-name> ./scripts/migrate.sh down 032
docker exec -it <app-container-name> ./scripts/migrate.sh down 031
docker exec -it <app-container-name> ./scripts/migrate.sh down 030

# Restart application with previous code
git checkout 8f483fc  # Previous commit before integration
docker-compose -f docker-compose.yml up -d app
```

---

## Troubleshooting

### Issue: Migration fails with "relation already exists"

**Solution**: The table may already exist from a previous attempt. Check:
```bash
docker exec -it <db-container-name> psql -U urlshortener_user -d urlshortener_db -c "\dt rbac_audit_log"
```

If table exists, you can skip that migration or manually verify schema matches.

### Issue: "permission denied" when calling RBAC endpoints

**Solution**:
1. Verify your user has admin role:
   ```sql
   SELECT r.name FROM user_roles ur
   JOIN roles r ON ur.role_id = r.id
   WHERE ur.user_id = YOUR_USER_ID;
   ```
2. Check JWT token is valid and not expired
3. Verify middleware is properly initialized in main.go

### Issue: Redis connection errors

**Solution**:
1. Check Redis is running:
   ```bash
   docker ps | grep redis
   ```
2. Verify Redis connection in application logs:
   ```bash
   docker logs <app-container-name> 2>&1 | grep -i redis
   ```
3. Check Redis configuration in environment variables

### Issue: Application won't start after deployment

**Solution**:
1. Check logs for errors:
   ```bash
   docker logs <app-container-name> --tail 100
   ```
2. Common issues:
   - Missing environment variables
   - Database connection failed
   - Port already in use
   - Invalid migration state

---

## Post-Deployment Checklist

- [ ] All 3 migrations applied successfully
- [ ] Database verification queries return expected results
- [ ] Application started without errors
- [ ] Admin can list roles (Test 1)
- [ ] Admin can list permissions (Test 2)
- [ ] Admin can view their roles (Test 3)
- [ ] Admin can create users with roles (Test 4)
- [ ] Audit log captures actions (Test 5)
- [ ] Content Editor CANNOT publish pages (Test 6)
- [ ] Content Editor CAN edit pages (Test 6)
- [ ] Backward compatibility: is_admin flag still works
- [ ] Production URL accessible: https://trunc.site/admin

---

## Next Steps After Successful Deployment

1. **Create Content Team Users**
   - Create content managers for publishing access
   - Create content editors for editing access
   - Send invitation emails with temporary passwords

2. **Frontend Integration**
   - Build admin UI for user/role management
   - Add role-based menu rendering
   - Implement permission-based button visibility

3. **Monitoring**
   - Set up alerts for RBAC audit log
   - Monitor permission check performance (Redis cache hit rate)
   - Track role assignment patterns

4. **Documentation**
   - Document your team's role assignment process
   - Create user guides for different roles
   - Update onboarding documentation

---

## Support

If you encounter issues during deployment:
1. Check application logs: `docker logs <app-container-name>`
2. Check database logs: `docker logs <db-container-name>`
3. Review MIGRATION_GUIDE.md for detailed troubleshooting
4. Check RBAC_IMPLEMENTATION_SUMMARY.md for API documentation

---

**Deployment Script Generated**: 2026-02-15
**RBAC Implementation Version**: 1.0.0
**Commit**: 93d20ba
