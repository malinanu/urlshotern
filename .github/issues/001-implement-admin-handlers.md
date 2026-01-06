# Implement Admin Handler Functions (Phase 2)

## Priority
🔴 High

## Labels
`enhancement`, `phase-2`, `admin-panel`, `backend`

## Description
Phase 1 created 97 admin routes, but the handler functions are not yet implemented. All routes are properly wired with middleware protection but will panic when called.

## Problem
Routes in `internal/routes/admin_routes.go` reference handler methods that don't exist:
- User management: 15 handlers
- URL management: 10 handlers
- Domain moderation: 3 handlers
- Analytics: 9 handlers
- Billing: 12 handlers
- RBAC: 12 handlers
- Settings: 6 handlers
- CMS: 8 handlers
- Audit logs: 8 handlers
- Monitoring: 7 handlers
- API keys: 7 handlers

## Expected Behavior
All 97 admin routes should have working handler implementations that:
- Accept and validate requests
- Enforce business logic
- Return appropriate responses
- Handle errors gracefully
- Log actions via audit service

## Proposed Solution

### File Structure
Create handler files in `internal/handlers/`:
```
internal/handlers/
├── admin_users.go        # 15 user management handlers
├── admin_urls.go         # 10 URL management handlers
├── admin_domains.go      # 3 domain moderation handlers
├── admin_analytics.go    # 9 analytics handlers
├── admin_billing.go      # 12 billing handlers
├── admin_rbac.go         # 12 RBAC handlers
├── admin_settings.go     # 6 settings handlers
├── admin_cms.go          # 8 CMS handlers
├── admin_audit.go        # 8 audit log handlers
├── admin_monitoring.go   # 7 monitoring handlers
└── admin_api_keys.go     # 7 API key handlers
```

### Implementation Order (Recommended)
1. **User Management** (highest priority)
   - `AdminUserList` - List users with pagination and filters
   - `AdminGetUser` - Get user details
   - `AdminUpdateUser` - Update user information
   - `AdminBanUser` - Ban user account
   - `AdminUnbanUser` - Unban user account
   - ... (10 more)

2. **URL Management**
3. **Audit Logs** (for testing/debugging)
4. **Analytics**
5. **RBAC**
6. **Remaining handlers**

### Handler Template
```go
// AdminUserList lists all users with pagination and filtering
func (h *Handler) AdminUserList(c *gin.Context) {
    // 1. Parse query parameters
    // 2. Validate input
    // 3. Call service layer
    // 4. Return response
    // Note: Audit logging handled by middleware
}
```

## Affected Files
- `internal/routes/admin_routes.go` (route definitions exist)
- `internal/handlers/admin_*.go` (NEW - to be created)

## Definition of Done
- [ ] All 97 handler functions implemented
- [ ] Request/response DTOs defined
- [ ] Input validation added
- [ ] Error handling implemented
- [ ] All handlers tested manually
- [ ] Integration tests written (optional)
- [ ] Documentation updated

## References
- Phase 1 Implementation: `PHASE1_COMPLETE.md`
- Admin Routes: `internal/routes/admin_routes.go:1`
- Admin Panel Plan: `ADMIN_PANEL_PLAN.md`

## Related Issues
- #2 Update cmd/server/main.go to use DatabaseManager
- #3 Test SQL Migrations

## Notes
This is the main Phase 2 implementation task. Consider breaking it down into smaller issues per handler group (e.g., "Implement User Management Handlers", "Implement URL Management Handlers", etc.).
