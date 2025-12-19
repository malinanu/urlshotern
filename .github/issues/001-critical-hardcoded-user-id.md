# [CRITICAL] Remove Hardcoded User ID Fallback

**Labels:** `critical`, `security`, `backend`, `bug`, `quick-win`

## Description
The application uses a hardcoded fallback to user ID 1 when authentication fails, creating a critical security vulnerability that allows unauthorized access.

## Severity: CRITICAL
This is a critical security vulnerability because:
- Any request without valid authentication defaults to user ID 1
- Attackers can access user 1's data without authentication
- Complete bypass of authentication system
- Immediate production security risk

## Affected Files
- [ ] `internal/handlers/handlers.go` (lines 358-359)
- [ ] `internal/handlers/handlers.go` (lines 395-397)
- [ ] `internal/handlers/handlers.go` (lines 477-480)

## Current Behavior
```go
userID, exists := middleware.GetUserID(c)
if !exists {
    // Temporary: Use default user ID 1 for development
    userID = 1
}
```
When authentication fails or user ID is not found in context, the system defaults to user ID 1 instead of returning an error.

## Expected Behavior
- Authentication should be required for all protected endpoints
- Failed authentication should return 401 Unauthorized
- No default user ID fallback
- Proper error messages to the client

## Proposed Solution
1. Remove the hardcoded `userID = 1` fallback
2. Return appropriate HTTP 401 error when user ID is missing:
   ```go
   userID, exists := middleware.GetUserID(c)
   if !exists {
       c.JSON(401, gin.H{"error": "Authentication required"})
       return
   }
   ```
3. Ensure authentication middleware is applied to all protected routes
4. Add integration tests to verify auth requirements

## Acceptance Criteria
- [ ] Hardcoded user ID 1 fallback removed from all handlers
- [ ] Endpoints return 401 when user not authenticated
- [ ] No unauthorized access possible
- [ ] Integration tests added to verify auth enforcement
- [ ] All existing authenticated endpoints still work correctly
- [ ] Documentation updated with authentication requirements

## Security Impact
- **Risk Level**: Critical
- **Attack Vector**: Bypassing authentication
- **Data at Risk**: All user data for user ID 1
- **Recommendation**: Fix immediately before any production deployment

## Related Issues
- #2 (Usage Limit Enforcement) - depends on proper authentication
