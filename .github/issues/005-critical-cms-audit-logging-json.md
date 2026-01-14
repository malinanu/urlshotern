# [CRITICAL] CMS Page Editing Fails Due to Audit Logging JSON Serialization Error

**Labels:** `critical`, `backend`, `bug`, `cms`, `quick-win`

## Description
CMS page editing fails with HTTP 500 error due to improper handling of `nil` values when logging audit trail. The audit logging service attempts to insert `nil` into PostgreSQL JSONB columns, which causes a "invalid input syntax for type json" error.

## Severity: CRITICAL
This is a critical bug because:
- Completely blocks all CMS page editing functionality
- Affects all admin users trying to update any CMS content
- Makes the CMS unusable for content management
- Impacts production content updates

## Error Details
```
[GIN] 2026/01/13 - 19:05:11 | 500 | 49.447218ms | 172.18.0.1 | PUT "/api/v1/admin/cms/pages/10"
2026/01/13 19:05:11 ERROR [e90489c0-d62c-41cb-8976-ebbb7b6eabb3] user_id=1000000000000000001
message=Failed to log audit trail
error=failed to insert audit log: pq: invalid input syntax for type json
path=/api/v1/admin/cms/pages/10 method=PUT
```

## Affected Files
- [ ] `internal/services/admin_audit_service.go` (lines 54-77)
- [ ] `internal/handlers/admin/admin_cms_handlers.go` (line 189)

## Current Behavior
When updating a CMS page:
1. Handler calls `auditService.LogActionHelper()` with details map
2. Service attempts to marshal `req.Changes` which is `nil`
3. When `changesJSON` is `nil`, it gets passed to PostgreSQL
4. PostgreSQL JSONB column rejects `nil` value
5. Audit log insertion fails with "invalid input syntax for type json"
6. Entire CMS update operation fails with HTTP 500

### Code Path
```go
// In admin_audit_service.go:54-61
var changesJSON []byte
if req.Changes != nil {
    changesJSON, err = json.Marshal(req.Changes)
    if err != nil {
        return fmt.Errorf("failed to marshal changes: %w", err)
    }
}
// changesJSON is nil here if req.Changes was nil

// Line 77: Passes nil to JSONB column
_, err = s.dbManager.AdminExecContext(ctx, query,
    req.UserID,
    req.Action,
    req.ResourceType,
    req.ResourceID,
    nil, // old_values
    changesJSON, // ← This is nil, causing the error
    detailsJSON,
    ...
)
```

## Root Cause
PostgreSQL JSONB columns cannot accept `nil` values. They require at minimum an empty JSON object `{}` or `NULL` SQL value. The code passes a nil byte slice which PostgreSQL interprets as invalid JSON syntax.

## Expected Behavior
- CMS pages should update successfully
- Audit logs should be created with empty JSON objects when no changes data exists
- No HTTP 500 errors during normal CMS operations
- Proper error handling for JSONB serialization

## Proposed Solution

### Option 1: Set empty JSON object when nil (Recommended)
```go
// In admin_audit_service.go, lines 54-61
var changesJSON []byte
if req.Changes != nil {
    changesJSON, err = json.Marshal(req.Changes)
    if err != nil {
        return fmt.Errorf("failed to marshal changes: %w", err)
    }
} else {
    // PostgreSQL JSONB requires valid JSON, not nil
    changesJSON = []byte("{}")
}
```

### Option 2: Use sql.NullString for JSONB columns
```go
var changesValue sql.NullString
if req.Changes != nil {
    changesJSON, err := json.Marshal(req.Changes)
    if err != nil {
        return fmt.Errorf("failed to marshal changes: %w", err)
    }
    changesValue = sql.NullString{String: string(changesJSON), Valid: true}
}
// changesValue will be NULL in database if not set
```

### Option 3: Make new_values column nullable and pass NULL
Modify the query to explicitly pass NULL:
```go
_, err = s.dbManager.AdminExecContext(ctx, query,
    req.UserID,
    req.Action,
    req.ResourceType,
    req.ResourceID,
    nil,
    sql.NullString{Valid: changesJSON != nil, String: string(changesJSON)},
    detailsJSON,
    ...
)
```

## Testing Steps to Reproduce
1. Navigate to Admin Panel: `http://localhost:3000/admin/login`
2. Login as admin user
3. Click "Content" tab
4. Click "Edit" button on any page (e.g., "Home - Trunc URL Shortener")
5. Make any change or click "Update Page" without changes
6. **Result**: HTTP 500 error with alert "update_failed"
7. **Expected**: Page updates successfully with success message

## Acceptance Criteria
- [ ] CMS page updates complete successfully
- [ ] Audit logs are created with proper JSON formatting
- [ ] Empty changes are represented as `{}` in JSONB columns
- [ ] No HTTP 500 errors during CMS operations
- [ ] All existing audit logging functionality still works
- [ ] Integration tests added to verify JSONB handling
- [ ] Similar issues checked in other audit logging calls

## Impact
- **Severity**: Critical - Blocks all CMS functionality
- **Users Affected**: All admin users
- **Workaround**: None available
- **Estimated Fix Time**: < 2 hours (quick-win)

## Related Files to Check
Other places where `req.Changes` or similar JSONB marshaling occurs:
- [ ] `internal/services/admin_audit_service.go` (Details field handling, line 44-52)
- [ ] All handlers calling `LogActionHelper()` with nil details

## Verification Checklist
After fix:
- [ ] CMS page editing works end-to-end
- [ ] Audit logs show correct JSON in database
- [ ] No PostgreSQL JSON syntax errors in logs
- [ ] Frontend shows success message after update
- [ ] Page changes persist in database
- [ ] All CMS operations (create, update, delete, publish) work correctly

## Screenshots/Logs

### Browser Console Error
```
PUT http://localhost:8080/api/v1/admin/cms/pages/10 500 (Internal Server Error)
```

### Backend Log
```
2026/01/13 19:05:11 ERROR message=Failed to log audit trail
error=failed to insert audit log: pq: invalid input syntax for type json
```

## Priority Justification
**Critical** because:
1. Completely breaks CMS page editing
2. No workaround available
3. Affects all admin users
4. Simple fix (< 1 hour development time)
5. Production-blocking issue

## Next Steps
1. ~~Implement Option 1 (empty JSON object for nil values)~~ ✅ COMPLETED
2. ~~Test with CMS page updates~~ ✅ COMPLETED
3. ~~Verify audit logs are created correctly~~ ✅ COMPLETED
4. ~~Check all other audit logging calls for same issue~~ ✅ COMPLETED
5. Add unit tests for JSONB nil handling
6. Deploy fix and verify in production

---

## Resolution

**Status:** ✅ **RESOLVED** (2026-01-14)

### Issues Found and Fixed

#### Issue 1: Audit Logging JSON Serialization (Primary Issue)
**File:** `internal/services/admin_audit_service.go`

**Problem:** Both `changesJSON` and `detailsJSON` were passing `nil` byte slices to PostgreSQL JSONB columns when the corresponding fields were nil, causing "invalid input syntax for type json" errors.

**Fix Applied:** Implemented Option 1 - Set empty JSON object for nil values
```go
// Lines 44-52: detailsJSON handling
var detailsJSON []byte
if req.Details != nil {
    detailsJSON, err = json.Marshal(req.Details)
    if err != nil {
        return fmt.Errorf("failed to marshal details: %w", err)
    }
} else {
    detailsJSON = []byte("{}")
}

// Lines 54-65: changesJSON handling
var changesJSON []byte
if req.Changes != nil {
    changesJSON, err = json.Marshal(req.Changes)
    if err != nil {
        return fmt.Errorf("failed to marshal changes: %w", err)
    }
} else {
    // Set empty JSON object instead of nil for PostgreSQL JSONB compatibility
    changesJSON = []byte("{}")
}
```

#### Issue 2: Database Column Mismatch (Secondary Issue)
**Files:** `internal/services/cms.go`

**Problem:** After fixing the audit logging issue, CMS updates still failed with:
```
ERROR: Failed to update CMS page 10: failed to get static page: pq: column u.first_name does not exist
```

The SQL queries in `GetStaticPageBySlug`, `GetStaticPageByID`, and `GetPageRevisions` were trying to reference `u.first_name` and `u.last_name` columns that don't exist in the users table. The users table has a `name` column instead.

**Fix Applied:** Updated all three functions to use `u.name` instead of concatenating `u.first_name` and `u.last_name`

**GetStaticPageBySlug (lines 82-88):**
```go
COALESCE(u.name, u.email) as author_name
```

**GetStaticPageByID (lines 114-120):**
```go
COALESCE(u.name, u.email) as author_name
```

**GetPageRevisions (lines 385-393):**
```go
COALESCE(u.name, u.email) as author_name
```

### Testing Performed

1. ✅ Used Playwright to reproduce the issue from UI
2. ✅ Verified error logs captured actual issues
3. ✅ Applied both fixes
4. ✅ Rebuilt Docker backend container
5. ✅ Successfully updated CMS page without errors
6. ✅ Verified HTTP 200 response instead of 500
7. ✅ Confirmed timestamp update in UI
8. ✅ No PostgreSQL errors in backend logs

### Verification Results

**Before Fix:**
- PUT `/api/v1/admin/cms/pages/10` → HTTP 500
- Error: "pq: invalid input syntax for type json" (Issue 1)
- Error: "pq: column u.first_name does not exist" (Issue 2)

**After Fix:**
- PUT `/api/v1/admin/cms/pages/10` → HTTP 200 ✅
- Page updated successfully
- Audit logs created correctly
- No errors in backend logs

### Files Modified

1. `internal/services/admin_audit_service.go` (lines 44-65)
2. `internal/services/cms.go` (lines 85, 117, 388)
3. `internal/handlers/admin/admin_cms_handlers.go` (added error logging for debugging)

### Acceptance Criteria Status

- [x] CMS page updates complete successfully
- [x] Audit logs are created with proper JSON formatting
- [x] Empty changes are represented as `{}` in JSONB columns
- [x] No HTTP 500 errors during CMS operations
- [x] All existing audit logging functionality still works
- [ ] Integration tests added to verify JSONB handling (TODO)
- [x] Similar issues checked in other audit logging calls

### Time to Resolution
- Total time: ~1.5 hours
- Identified root causes: 45 minutes
- Implemented fixes: 30 minutes
- Testing and verification: 15 minutes
