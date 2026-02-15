-- Migration: RBAC Audit Log
-- Category: RBAC Auditing
-- Tables: rbac_audit_log
-- Dependencies: 030_add_cms_permissions

-- ============================================================================
-- Create RBAC Audit Log Table
-- ============================================================================

CREATE TABLE IF NOT EXISTS rbac_audit_log (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    target_user_id BIGINT REFERENCES users(id) ON DELETE SET NULL,
    action VARCHAR(50) NOT NULL, -- role_assigned, role_removed, user_created, user_updated, user_deleted, etc.
    role_id BIGINT REFERENCES roles(id) ON DELETE SET NULL,
    permission_id BIGINT REFERENCES permissions(id) ON DELETE SET NULL,
    assigned_by BIGINT REFERENCES users(id) ON DELETE SET NULL,
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- ============================================================================
-- Create Indexes for Performance
-- ============================================================================

CREATE INDEX IF NOT EXISTS idx_rbac_audit_user ON rbac_audit_log(user_id);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_target ON rbac_audit_log(target_user_id);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_action ON rbac_audit_log(action);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_created ON rbac_audit_log(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_assigned_by ON rbac_audit_log(assigned_by);
CREATE INDEX IF NOT EXISTS idx_rbac_audit_role ON rbac_audit_log(role_id) WHERE role_id IS NOT NULL;

-- ============================================================================
-- Add Comments for Documentation
-- ============================================================================

COMMENT ON TABLE rbac_audit_log IS 'Audit trail for all RBAC-related actions (role assignments, user management, etc.)';
COMMENT ON COLUMN rbac_audit_log.user_id IS 'User who performed the action (may be NULL for system actions)';
COMMENT ON COLUMN rbac_audit_log.target_user_id IS 'User who was affected by the action (for user management operations)';
COMMENT ON COLUMN rbac_audit_log.action IS 'Type of action performed (role_assigned, role_removed, user_created, etc.)';
COMMENT ON COLUMN rbac_audit_log.role_id IS 'Role involved in the action (if applicable)';
COMMENT ON COLUMN rbac_audit_log.permission_id IS 'Permission involved in the action (if applicable)';
COMMENT ON COLUMN rbac_audit_log.assigned_by IS 'Admin user who initiated the action';
COMMENT ON COLUMN rbac_audit_log.details IS 'Additional JSON details about the action';
COMMENT ON COLUMN rbac_audit_log.ip_address IS 'IP address of the user who performed the action';
COMMENT ON COLUMN rbac_audit_log.user_agent IS 'User agent of the client that performed the action';

-- ============================================================================
-- VERIFICATION: Ensure table was created successfully
-- ============================================================================

DO $$
DECLARE
    table_exists BOOLEAN;
    index_count INTEGER;
BEGIN
    -- Verify table exists
    SELECT EXISTS (
        SELECT FROM information_schema.tables
        WHERE table_schema = 'public' AND table_name = 'rbac_audit_log'
    ) INTO table_exists;

    IF NOT table_exists THEN
        RAISE EXCEPTION 'Migration 032 Failed: rbac_audit_log table was not created';
    END IF;
    RAISE NOTICE 'Migration 032: rbac_audit_log table created ✓';

    -- Verify indexes were created
    SELECT COUNT(*) INTO index_count
    FROM pg_indexes
    WHERE tablename = 'rbac_audit_log';

    IF index_count < 6 THEN
        RAISE WARNING 'Migration 032: Expected 6 indexes, found %', index_count;
    ELSE
        RAISE NOTICE 'Migration 032: % indexes created ✓', index_count;
    END IF;

    RAISE NOTICE 'Migration 032 completed successfully ✓';
END $$;
