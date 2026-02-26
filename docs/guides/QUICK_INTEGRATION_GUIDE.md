# Quick Integration Guide

This guide helps you integrate the RBAC implementation into your existing codebase.

## ⚡ Quick Start (5 Steps)

### Step 1: Update `cmd/server/main.go`

Find where you initialize services and add RBAC:

```go
// After initializing postgresStorage and redisStorage, add:

// Initialize RBAC Service (NEW)
rbacService := services.NewRBACServiceImpl(postgresStorage, redisStorage)

// Update AdminService to include RBAC (MODIFIED)
adminService := services.NewAdminService(postgresStorage, rbacService)

// Update AuthMiddleware to use RBAC (MODIFIED)
authMiddleware := middleware.NewAuthMiddleware(
    jwtService,
    userService,
    rbacService,  // Changed from old RBACService to RBACServiceImpl
    redisStorage,
)

// Create AdminRBACHandlers (NEW)
adminRBACHandlers := handlers.NewAdminRBACHandlers(adminService, rbacService)

// When creating the main handler, add adminRBACHandlers
// You'll need to update handlers.NewHandler to accept this parameter
```

### Step 2: Update `internal/handlers/handlers.go`

Modify the NewHandler function signature:

```go
func NewHandler(
    shortenerService *services.ShortenerService,
    analyticsService *services.AnalyticsService,
    advancedAnalyticsService *services.AdvancedAnalyticsService,
    conversionService *services.ConversionTrackingService,
    abTestService *services.ABTestingService,
    realtimeService *services.RealtimeAnalyticsService,
    attributionService *services.AttributionService,
    authHandlers *AuthHandlers,
    analyticsHandlers *AnalyticsHandlers,
    db *storage.PostgresStorage,
    gormDB *gorm.DB,
    config *configs.Config,
    adminRBACHandlers *AdminRBACHandlers, // ADD THIS PARAMETER
) *Handler {
    // ... existing code ...

    return &Handler{
        // ... existing fields ...
        AdminRBACHandlers: adminRBACHandlers, // ADD THIS FIELD
    }
}
```

### Step 3: Run Migrations

**Production (Oracle Ubuntu Server):**
```bash
ssh user@your-server
cd /path/to/urlshortener

# Apply migrations via Docker
docker exec -it urlshortener-app ./scripts/migrate.sh up 030
docker exec -it urlshortener-app ./scripts/migrate.sh up 031
docker exec -it urlshortener-app ./scripts/migrate.sh up 032

# Verify
docker exec -it urlshortener-app psql -U urlshortener_user -d urlshortener_db -c "SELECT name FROM roles WHERE name IN ('super_admin', 'content_manager', 'content_editor', 'moderator');"
```

**Local Development:**
```bash
# Start services
docker-compose -f docker/docker-compose.local.yml up -d

# Wait for DB to be ready
sleep 5

# Apply migrations
docker-compose -f docker/docker-compose.local.yml exec app ./scripts/migrate.sh up 030
docker-compose -f docker/docker-compose.local.yml exec app ./scripts/migrate.sh up 031
docker-compose -f docker/docker-compose.local.yml exec app ./scripts/migrate.sh up 032
```

### Step 4: Build and Deploy

```bash
# Build the application
go build -o bin/server cmd/server/main.go

# Or with Docker
docker-compose -f docker/docker-compose.yml build

# Deploy (your deployment process)
# Example: Push to git, pull on server, restart containers
```

### Step 5: Verify

Test the new endpoints:

```bash
# Get your admin JWT token (replace with your actual token)
TOKEN="your-jwt-token-here"

# Test listing roles
curl https://trunc.site/api/v1/admin/rbac/roles \
  -H "Authorization: Bearer $TOKEN"

# Expected: List of roles including super_admin, admin, content_manager, etc.

# Test creating a user (as super admin)
curl -X POST https://trunc.site/api/v1/admin/users/create \
  -H "Authorization: Bearer $TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Editor",
    "email": "test@example.com",
    "account_type": "free",
    "role_ids": [4],
    "email_verified": true,
    "is_active": true
  }'

# Expected: User created with content_editor role
```

---

## 🔍 Compilation Fixes

If you encounter compilation errors, here's how to fix them:

### Error: "RBACService undefined"
**Fix:** The middleware now uses `RBACServiceImpl` instead of `RBACService`. Update your initialization code in `main.go`.

### Error: "AdminService constructor mismatch"
**Fix:** AdminService now requires rbacService parameter:
```go
adminService := services.NewAdminService(postgresStorage, rbacService)
```

### Error: "undefined: models.PermissionCMSPagesRead"
**Fix:** Run `go mod tidy` and rebuild. The new permission constants are in `internal/models/rbac.go`.

### Error: "undefined: models.CreateUserAdminRequest"
**Fix:** These models are in `internal/models/admin.go`. Make sure to rebuild after pulling the changes.

---

## 📋 Integration Checklist

- [ ] Updated `cmd/server/main.go` to initialize rbacService
- [ ] Updated AdminService initialization to include rbacService
- [ ] Updated AuthMiddleware to use RBACServiceImpl
- [ ] Created AdminRBACHandlers in main.go
- [ ] Updated handlers.NewHandler signature to accept adminRBACHandlers
- [ ] Updated Handler struct initialization to include AdminRBACHandlers
- [ ] Ran `go mod tidy`
- [ ] Built application successfully (`go build`)
- [ ] Applied migrations 030, 031, 032
- [ ] Verified new roles exist in database
- [ ] Tested API endpoints
- [ ] Checked audit log functionality
- [ ] Verified backward compatibility (is_admin still works)

---

## 🚨 Common Issues & Solutions

### Issue: "no such table: rbac_audit_log"
**Solution:** Migration 032 didn't run. Apply it:
```bash
docker exec -it <container> ./scripts/migrate.sh up 032
```

### Issue: "permission denied" when creating users
**Solution:** Your user needs the `users.create` permission. Check:
```sql
SELECT p.name FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN user_roles ur ON rp.role_id = ur.role_id
WHERE ur.user_id = YOUR_USER_ID;
```

### Issue: Content Editor can publish pages
**Solution:** Check role-permission mappings:
```sql
SELECT p.name FROM permissions p
JOIN role_permissions rp ON p.id = rp.permission_id
JOIN roles r ON rp.role_id = r.id
WHERE r.name = 'content_editor' AND p.name LIKE 'cms.pages%';
```
Should NOT include `cms.pages.publish`.

### Issue: Redis caching not working
**Solution:** Verify Redis connection in main.go and check logs:
```bash
docker logs <app-container> 2>&1 | grep -i redis
```

---

## 🎯 Testing Your Integration

### Test 1: Backward Compatibility
```bash
# Users with is_admin=true should still have admin access
SELECT u.id, u.email, u.is_admin FROM users WHERE is_admin = true;

# They should also have the admin role assigned
SELECT u.email, r.name FROM users u
JOIN user_roles ur ON u.id = ur.user_id
JOIN roles r ON ur.role_id = r.id
WHERE u.is_admin = true;
```

### Test 2: Permission Enforcement
```bash
# Login as Content Editor
EDITOR_TOKEN="content-editor-jwt-token"

# Try to publish (should fail with 403)
curl -X PUT https://trunc.site/api/v1/admin/cms/pages/1/publish \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"is_published": true}'

# Expected: {"error":"permission_denied","message":"Required permission: cms.pages.publish"}

# Try to edit (should succeed)
curl -X PUT https://trunc.site/api/v1/admin/cms/pages/1 \
  -H "Authorization: Bearer $EDITOR_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{"title":"Updated Title"}'

# Expected: 200 OK
```

### Test 3: Role Hierarchy
```bash
# Admin tries to create super_admin (should fail)
ADMIN_TOKEN="admin-jwt-token"

curl -X POST https://trunc.site/api/v1/admin/users/create \
  -H "Authorization: Bearer $ADMIN_TOKEN" \
  -H "Content-Type: application/json" \
  -d '{
    "name": "Test Super",
    "email": "super@example.com",
    "account_type": "free",
    "role_ids": [1],
    "is_active": true
  }'

# Expected: {"error":"insufficient privileges to assign role 'super_admin'"}
```

### Test 4: Audit Logging
```bash
# Create a user and check audit log
curl -X POST https://trunc.site/api/v1/admin/users/create \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN" \
  -d '{ ... }'

# Check audit log
curl "https://trunc.site/api/v1/admin/rbac/audit?limit=10&action=user_created" \
  -H "Authorization: Bearer $SUPER_ADMIN_TOKEN"

# Expected: Audit entry with action=user_created
```

---

## 📞 Need Help?

1. **Check logs**: `docker logs <container-name> --tail 100`
2. **Verify database**: Connect to PostgreSQL and run verification queries from `MIGRATION_GUIDE.md`
3. **Check Redis**: `redis-cli ping` should return PONG
4. **Review code**: All new files are documented with comments

---

## 🎉 You're Done!

If all checks pass, your RBAC system is fully integrated and ready to use!

Next steps:
- Create your first content editor via the API
- Test role-based access control
- Review audit logs
- Build frontend UI for user/role management

For detailed API documentation, see `RBAC_IMPLEMENTATION_SUMMARY.md`.
