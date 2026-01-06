# Update cmd/server/main.go to Use DatabaseManager

## Priority
🔴 High

## Labels
`bug`, `phase-1`, `database`, `integration`

## Description
The main application entry point (`cmd/server/main.go`) has not been updated to use the new `DatabaseManager` introduced in Phase 1. This prevents the admin panel from functioning with proper database security.

## Problem
Current state:
- Application likely uses old `PostgresStorage` or single database connection
- New `DatabaseManager` with 3 separate connections not integrated
- `routes.SetupRoutes()` signature changed but main.go not updated
- Admin middleware not initialized

## Expected Behavior
Main application should:
1. Initialize `DatabaseManager` with 3 connections (app, admin, readonly)
2. Initialize `SnowflakeGenerator`
3. Initialize `AdminAuditService`
4. Initialize `AdminMiddleware`
5. Pass `AdminMiddleware` to `routes.SetupRoutes()`

## Current Signature
```go
routes.SetupRoutes(router, handler, authMiddleware)
```

## Required Signature
```go
routes.SetupRoutes(router, handler, authMiddleware, adminMiddleware)
```

## Proposed Solution

Update `cmd/server/main.go` with:

```go
// Initialize database manager (replaces single DB connection)
dbManager, err := storage.NewDatabaseManager(config)
if err != nil {
    log.Fatalf("Failed to initialize database manager: %v", err)
}
defer dbManager.Close()

// Initialize Snowflake ID generator
snowflake, err := utils.NewSnowflakeGenerator(config.NodeID)
if err != nil {
    log.Fatalf("Failed to initialize Snowflake: %v", err)
}

// Initialize audit service
auditService := services.NewAdminAuditService(dbManager, snowflake)

// Initialize RBAC service (if not already initialized)
rbacService := services.NewRBACService(dbManager)

// Initialize admin middleware
adminMiddleware := middleware.NewAdminMiddleware(dbManager, auditService, rbacService)

// Setup routes with admin middleware
routes.SetupRoutes(router, handler, authMiddleware, adminMiddleware)
```

## Environment Variables Required
Ensure `.env` includes:
```bash
# App database user
DB_APP_USER=urlshort_app
DB_APP_PASSWORD=<secure_password>
DB_APP_MAX_CONNECTIONS=50

# Admin database user
DB_ADMIN_USER=urlshort_admin
DB_ADMIN_PASSWORD=<secure_password>
DB_ADMIN_MAX_CONNECTIONS=10

# Read-only database user
DB_READONLY_USER=urlshort_readonly
DB_READONLY_PASSWORD=<secure_password>
DB_READONLY_MAX_CONNECTIONS=20
```

## Affected Files
- `cmd/server/main.go` (primary file to modify)
- Possibly: Existing handler initialization code

## Testing Checklist
After implementation:
- [ ] Application starts without errors
- [ ] Database connections established (3 pools)
- [ ] Admin routes accessible
- [ ] Regular routes still work
- [ ] Middleware chain executes correctly
- [ ] Audit logging works
- [ ] Health check shows all DB pools

## Definition of Done
- [ ] `cmd/server/main.go` updated
- [ ] Application compiles successfully
- [ ] Application runs without errors
- [ ] Database manager properly initialized
- [ ] Admin middleware integrated
- [ ] Routes function correctly
- [ ] Manual testing completed

## References
- Database Manager: `internal/storage/database_manager.go:1`
- Admin Middleware: `internal/middleware/admin.go:1`
- Routes Setup: `internal/routes/routes.go:20`
- Configuration: `configs/config.go:1`

## Related Issues
- #1 Implement Admin Handler Functions
- #3 Test SQL Migrations

## Notes
This is a prerequisite for testing the admin panel functionality. Should be completed before handler implementation.
