# Test SQL Migrations in Test Database

## Priority
🔴 Critical

## Labels
`testing`, `database`, `phase-1`, `migrations`

## Description
SQL migration scripts created in Phase 1 have not been tested in a live database environment. We need to verify they execute successfully and create the expected schema.

## Problem
Two migration files exist but remain untested:
1. `migrations/001_phase1_admin_foundation.sql` (400+ lines)
2. `migrations/002_setup_database_users.sql` (300+ lines)

Without testing, we risk:
- SQL syntax errors
- Missing dependencies
- Incorrect permissions
- Failed constraint creation
- Broken indexes

## Expected Behavior
Migrations should:
1. Execute without errors
2. Create all expected tables and schemas
3. Insert all permissions and roles
4. Set up database users correctly
5. Enforce Row-Level Security policies
6. Grant appropriate permissions

## Test Plan

### Setup Test Environment
```bash
# Create test database
createdb urlshortener_test

# Or use Docker
docker run -d \
  --name postgres-test \
  -e POSTGRES_DB=urlshortener_test \
  -e POSTGRES_USER=postgres \
  -e POSTGRES_PASSWORD=postgres \
  -p 5433:5432 \
  postgres:15
```

### Test Migration 001
```bash
# Run migration
psql -U postgres -d urlshortener_test \
  -f migrations/001_phase1_admin_foundation.sql

# Verify schema created
psql -U postgres -d urlshortener_test \
  -c "\dn admin"

# Verify tables created
psql -U postgres -d urlshortener_test \
  -c "\dt admin.*"

# Count permissions inserted
psql -U postgres -d urlshortener_test \
  -c "SELECT COUNT(*) FROM permissions WHERE name LIKE 'admin.%';"
# Expected: 19

# Count roles inserted
psql -U postgres -d urlshortener_test \
  -c "SELECT COUNT(*) FROM roles WHERE name IN ('super_admin', 'system_admin', 'support_admin', 'billing_admin', 'content_admin');"
# Expected: 5

# Verify table structures
psql -U postgres -d urlshortener_test \
  -c "\d admin.audit_logs"
psql -U postgres -d urlshortener_test \
  -c "\d admin.system_settings"
psql -U postgres -d urlshortener_test \
  -c "\d admin.blocked_domains"
psql -U postgres -d urlshortener_test \
  -c "\d admin.flagged_urls"
psql -U postgres -d urlshortener_test \
  -c "\d admin.system_alerts"
psql -U postgres -d urlshortener_test \
  -c "\d admin.scheduled_reports"
```

### Test Migration 002
```bash
# Run migration
psql -U postgres -d urlshortener_test \
  -f migrations/002_setup_database_users.sql

# Verify users created
psql -U postgres -d urlshortener_test \
  -c "\du urlshort_*"

# Test app user permissions (SHOULD FAIL - no DELETE)
psql -U urlshort_app -d urlshortener_test \
  -c "DELETE FROM users WHERE id = 1;"
# Expected: ERROR: permission denied

# Test app user cannot access admin schema (SHOULD FAIL)
psql -U urlshort_app -d urlshortener_test \
  -c "SELECT * FROM admin.audit_logs;"
# Expected: ERROR: permission denied for schema admin

# Test admin user CAN access admin schema (SHOULD SUCCEED)
psql -U urlshort_admin -d urlshortener_test \
  -c "SELECT COUNT(*) FROM admin.audit_logs;"
# Expected: Success with count 0

# Test admin user CAN delete (SHOULD SUCCEED)
psql -U urlshort_admin -d urlshortener_test \
  -c "DELETE FROM users WHERE id = 999999;"
# Expected: Success (0 rows affected)

# Test readonly user cannot write (SHOULD FAIL)
psql -U urlshort_readonly -d urlshortener_test \
  -c "INSERT INTO users (email, name) VALUES ('test@test.com', 'Test');"
# Expected: ERROR: permission denied

# Test readonly user CAN read (SHOULD SUCCEED)
psql -U urlshort_readonly -d urlshortener_test \
  -c "SELECT COUNT(*) FROM users;"
# Expected: Success
```

### Document Results
Create a test report documenting:
- [ ] Which SQL statements executed successfully
- [ ] Which SQL statements failed (if any)
- [ ] Schema verification results
- [ ] Permission test results
- [ ] Any errors or warnings encountered

## Potential Issues to Watch For

### Common SQL Errors
- Missing tables that migrations depend on
- Incorrect data types
- Missing sequences
- Foreign key constraint failures
- Unique constraint violations
- Permission denied errors

### Database User Issues
- Users already exist (conflict)
- Passwords need escaping
- Permission grants fail
- RLS policies not applying
- Role inheritance problems

## Rollback Plan
If migrations fail, create rollback scripts:

```sql
-- Rollback 002
DROP OWNED BY urlshort_app CASCADE;
DROP OWNED BY urlshort_admin CASCADE;
DROP OWNED BY urlshort_readonly CASCADE;
DROP USER urlshort_app;
DROP USER urlshort_admin;
DROP USER urlshort_readonly;

-- Rollback 001
DROP SCHEMA admin CASCADE;
DELETE FROM permissions WHERE name LIKE 'admin.%';
DELETE FROM roles WHERE name IN ('super_admin', 'system_admin', 'support_admin', 'billing_admin', 'content_admin');
```

## Definition of Done
- [ ] Test database environment set up
- [ ] Migration 001 executed successfully
- [ ] Migration 002 executed successfully
- [ ] All schema elements verified
- [ ] All permissions verified
- [ ] Database user permissions tested
- [ ] Test results documented
- [ ] Any issues found are fixed
- [ ] Rollback scripts created
- [ ] Documentation updated

## Affected Files
- `migrations/001_phase1_admin_foundation.sql`
- `migrations/002_setup_database_users.sql`
- Possibly: New rollback scripts to create

## References
- Phase 1 Complete: `PHASE1_COMPLETE.md`
- Database Security: `ADMIN_DATABASE_SECURITY.md`
- Migration Scripts: `migrations/`

## Related Issues
- #2 Update cmd/server/main.go
- #4 Create Super Admin User

## Notes
Critical to complete before deploying Phase 1 to any environment. Consider automating these tests in CI/CD pipeline.
