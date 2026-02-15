-- Rollback Migration 032: Remove RBAC Audit Log

-- Drop indexes
DROP INDEX IF EXISTS idx_rbac_audit_user;
DROP INDEX IF EXISTS idx_rbac_audit_target;
DROP INDEX IF EXISTS idx_rbac_audit_action;
DROP INDEX IF EXISTS idx_rbac_audit_created;
DROP INDEX IF EXISTS idx_rbac_audit_assigned_by;
DROP INDEX IF EXISTS idx_rbac_audit_role;

-- Drop table
DROP TABLE IF EXISTS rbac_audit_log;

RAISE NOTICE 'Migration 032 rolled back successfully';
