-- ============================================================================
-- Migration: Phase 1 - Admin Panel Foundation & Security
-- Version: 001
-- Description: Create admin schema, admin-specific tables, and security setup
-- Author: Admin Panel Implementation
-- Date: 2026-01-06
-- ============================================================================

-- Connect to database
\c urlshortener_db;

-- ============================================================================
-- STEP 1: Create Admin Schema
-- ============================================================================

CREATE SCHEMA IF NOT EXISTS admin;

COMMENT ON SCHEMA admin IS 'Schema for admin-only tables and sensitive data';

-- ============================================================================
-- STEP 2: Update Users Table for Enhanced Admin Support
-- ============================================================================

-- Add admin-specific fields if they don't exist
ALTER TABLE users ADD COLUMN IF NOT EXISTS admin_notes TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS banned_at TIMESTAMP WITH TIME ZONE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS banned_by BIGINT REFERENCES users(id);
ALTER TABLE users ADD COLUMN IF NOT EXISTS ban_reason TEXT;
ALTER TABLE users ADD COLUMN IF NOT EXISTS deleted_at TIMESTAMP WITH TIME ZONE;

-- Update account_type to support admin roles
-- Existing: 'free', 'premium', 'enterprise'
-- New: 'user', 'support_admin', 'billing_admin', 'content_admin', 'system_admin', 'super_admin'
COMMENT ON COLUMN users.account_type IS 'Account type: user, free, premium, enterprise, support_admin, billing_admin, content_admin, system_admin, super_admin';

-- Add index for admin queries
CREATE INDEX IF NOT EXISTS idx_users_account_type ON users(account_type);
CREATE INDEX IF NOT EXISTS idx_users_banned_at ON users(banned_at);
CREATE INDEX IF NOT EXISTS idx_users_deleted_at ON users(deleted_at);

-- ============================================================================
-- STEP 3: Create Admin Audit Logs Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT,  -- Reference to public.users (no FK for isolation)
    username VARCHAR(100),  -- Cached username for easier querying
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    details JSONB DEFAULT '{}',
    changes JSONB,  -- Before/after values: {"before": {...}, "after": {...}}
    ip_address INET,
    user_agent TEXT,
    session_id VARCHAR(100),
    status VARCHAR(20) DEFAULT 'success',  -- success, failed, unauthorized
    error_message TEXT,
    severity VARCHAR(20) DEFAULT 'info',  -- info, warning, critical
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for admin audit logs
CREATE INDEX IF NOT EXISTS idx_admin_audit_user_id ON admin.audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_action ON admin.audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_admin_audit_resource ON admin.audit_logs(resource_type, resource_id);
CREATE INDEX IF NOT EXISTS idx_admin_audit_status ON admin.audit_logs(status);
CREATE INDEX IF NOT EXISTS idx_admin_audit_severity ON admin.audit_logs(severity);
CREATE INDEX IF NOT EXISTS idx_admin_audit_created_at ON admin.audit_logs(created_at DESC);

COMMENT ON TABLE admin.audit_logs IS 'Admin action audit trail - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 4: Create System Settings Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.system_settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value JSONB NOT NULL,
    category VARCHAR(50),  -- 'security', 'email', 'billing', 'features', etc.
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,  -- Can non-admins read this setting?
    is_encrypted BOOLEAN DEFAULT FALSE,  -- Is the value encrypted?
    data_type VARCHAR(20) DEFAULT 'string',  -- string, number, boolean, json
    validation_rules JSONB,  -- JSON schema or validation rules
    updated_by BIGINT,  -- Reference to public.users
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_settings_category ON admin.system_settings(category);
CREATE INDEX IF NOT EXISTS idx_system_settings_is_public ON admin.system_settings(is_public);

COMMENT ON TABLE admin.system_settings IS 'Global system configuration - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 5: Create Blocked Domains Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.blocked_domains (
    id BIGSERIAL PRIMARY KEY,
    domain VARCHAR(255) UNIQUE NOT NULL,
    reason TEXT,
    blocked_by BIGINT,  -- Reference to public.users
    blocked_by_username VARCHAR(100),
    is_active BOOLEAN DEFAULT TRUE,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_blocked_domains_domain ON admin.blocked_domains(domain);
CREATE INDEX IF NOT EXISTS idx_blocked_domains_is_active ON admin.blocked_domains(is_active);
CREATE INDEX IF NOT EXISTS idx_blocked_domains_created_at ON admin.blocked_domains(created_at DESC);

COMMENT ON TABLE admin.blocked_domains IS 'Blocked domains for URL shortening - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 6: Create Flagged URLs Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.flagged_urls (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(20),  -- Reference to public.url_mappings
    url_id BIGINT,  -- Reference to public.url_mappings(id)
    original_url TEXT,  -- Cached for reference
    reason VARCHAR(100),
    details TEXT,
    flagged_by BIGINT,  -- Reference to public.users
    flagged_by_username VARCHAR(100),
    status VARCHAR(20) DEFAULT 'pending',  -- pending, reviewed, dismissed, removed
    priority VARCHAR(20) DEFAULT 'medium',  -- low, medium, high, critical
    reviewed_by BIGINT,  -- Reference to public.users
    reviewed_by_username VARCHAR(100),
    reviewed_at TIMESTAMP WITH TIME ZONE,
    review_notes TEXT,
    action_taken VARCHAR(50),  -- url_disabled, url_deleted, user_warned, user_banned, dismissed
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_flagged_urls_short_code ON admin.flagged_urls(short_code);
CREATE INDEX IF NOT EXISTS idx_flagged_urls_status ON admin.flagged_urls(status);
CREATE INDEX IF NOT EXISTS idx_flagged_urls_priority ON admin.flagged_urls(priority);
CREATE INDEX IF NOT EXISTS idx_flagged_urls_created_at ON admin.flagged_urls(created_at DESC);
CREATE INDEX IF NOT EXISTS idx_flagged_urls_reviewed_at ON admin.flagged_urls(reviewed_at);

COMMENT ON TABLE admin.flagged_urls IS 'Flagged URLs for moderation - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 7: Create System Alerts Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.system_alerts (
    id BIGSERIAL PRIMARY KEY,
    type VARCHAR(50) NOT NULL,  -- error_rate, response_time, disk_space, memory, db_connections, etc.
    severity VARCHAR(20) NOT NULL,  -- info, warning, critical
    title VARCHAR(200) NOT NULL,
    message TEXT NOT NULL,
    details JSONB DEFAULT '{}',
    threshold_value NUMERIC,
    current_value NUMERIC,
    status VARCHAR(20) DEFAULT 'active',  -- active, acknowledged, resolved, dismissed
    acknowledged_by BIGINT,  -- Reference to public.users
    acknowledged_by_username VARCHAR(100),
    acknowledged_at TIMESTAMP WITH TIME ZONE,
    acknowledgment_notes TEXT,
    resolved_by BIGINT,
    resolved_by_username VARCHAR(100),
    resolved_at TIMESTAMP WITH TIME ZONE,
    resolution_notes TEXT,
    notified_users BIGINT[],  -- Array of user IDs notified
    notification_sent_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_system_alerts_type ON admin.system_alerts(type);
CREATE INDEX IF NOT EXISTS idx_system_alerts_status ON admin.system_alerts(status);
CREATE INDEX IF NOT EXISTS idx_system_alerts_severity ON admin.system_alerts(severity);
CREATE INDEX IF NOT EXISTS idx_system_alerts_created_at ON admin.system_alerts(created_at DESC);

COMMENT ON TABLE admin.system_alerts IS 'System health and monitoring alerts - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 8: Create Scheduled Reports Table (in admin schema)
-- ============================================================================

CREATE TABLE IF NOT EXISTS admin.scheduled_reports (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    report_type VARCHAR(50) NOT NULL,  -- user_growth, revenue, url_stats, system_health, etc.
    schedule VARCHAR(50) NOT NULL,  -- daily, weekly, monthly, custom_cron
    cron_expression VARCHAR(100),  -- For custom schedules
    recipients TEXT[],  -- Email addresses
    filters JSONB DEFAULT '{}',
    format VARCHAR(20) DEFAULT 'pdf',  -- pdf, csv, excel, json
    is_active BOOLEAN DEFAULT TRUE,
    last_run_at TIMESTAMP WITH TIME ZONE,
    last_run_status VARCHAR(20),  -- success, failed, running
    last_run_error TEXT,
    next_run_at TIMESTAMP WITH TIME ZONE,
    total_runs INTEGER DEFAULT 0,
    successful_runs INTEGER DEFAULT 0,
    failed_runs INTEGER DEFAULT 0,
    created_by BIGINT,  -- Reference to public.users
    created_by_username VARCHAR(100),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE INDEX IF NOT EXISTS idx_scheduled_reports_next_run ON admin.scheduled_reports(next_run_at) WHERE is_active = TRUE;
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_is_active ON admin.scheduled_reports(is_active);
CREATE INDEX IF NOT EXISTS idx_scheduled_reports_report_type ON admin.scheduled_reports(report_type);

COMMENT ON TABLE admin.scheduled_reports IS 'Scheduled report configurations - ADMIN ACCESS ONLY';

-- ============================================================================
-- STEP 9: Create Admin Permissions
-- ============================================================================

-- Insert admin-specific permissions
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    -- Admin User Management
    ('admin.user.view', 'View All Users (Admin)', 'View all user accounts', 'admin_user', 'view', true),
    ('admin.user.manage', 'Manage Users (Admin)', 'Create, update, ban, delete users', 'admin_user', 'manage', true),
    ('admin.user.impersonate', 'Impersonate Users', 'Impersonate users for support', 'admin_user', 'impersonate', true),

    -- Admin URL Management
    ('admin.url.view', 'View All URLs (Admin)', 'View all shortened URLs', 'admin_url', 'view', true),
    ('admin.url.manage', 'Manage URLs (Admin)', 'Moderate and delete URLs', 'admin_url', 'manage', true),
    ('admin.url.flag', 'Flag URLs', 'Flag URLs for review', 'admin_url', 'flag', true),

    -- Admin Analytics
    ('admin.analytics.view', 'View System Analytics', 'View system-wide analytics', 'admin_analytics', 'view', true),
    ('admin.analytics.export', 'Export Analytics', 'Export analytics data', 'admin_analytics', 'export', true),

    -- Admin Billing
    ('admin.billing.view', 'View All Subscriptions', 'View all user subscriptions', 'admin_billing', 'view', true),
    ('admin.billing.manage', 'Manage Subscriptions', 'Manage user subscriptions and plans', 'admin_billing', 'manage', true),

    -- Admin RBAC
    ('admin.rbac.view', 'View Roles & Permissions', 'View roles and permissions', 'admin_rbac', 'view', true),
    ('admin.rbac.manage', 'Manage RBAC', 'Create and manage roles and permissions', 'admin_rbac', 'manage', true),

    -- Admin Settings
    ('admin.settings.view', 'View System Settings', 'View system settings', 'admin_settings', 'view', true),
    ('admin.settings.manage', 'Manage System Settings', 'Modify system settings', 'admin_settings', 'manage', true),

    -- Admin CMS
    ('admin.cms.view', 'View CMS Content', 'View CMS pages and media', 'admin_cms', 'view', true),
    ('admin.cms.manage', 'Manage CMS', 'Create and manage CMS content', 'admin_cms', 'manage', true),

    -- Admin Audit Logs
    ('admin.audit.view', 'View Audit Logs', 'View admin audit logs', 'admin_audit', 'view', true),

    -- Admin Monitoring
    ('admin.monitoring.view', 'View System Health', 'View system health and alerts', 'admin_monitoring', 'view', true),
    ('admin.monitoring.manage', 'Manage Monitoring', 'Manage alerts and health checks', 'admin_monitoring', 'manage', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- STEP 10: Create Admin Roles
-- ============================================================================

-- Insert admin roles
INSERT INTO roles (name, display_name, description, is_system) VALUES
    ('super_admin', 'Super Administrator', 'Full system access with all permissions', true),
    ('system_admin', 'System Administrator', 'User and content management, system monitoring', true),
    ('support_admin', 'Support Administrator', 'View-only with user assistance capabilities', true),
    ('billing_admin', 'Billing Administrator', 'Subscription and billing management', true),
    ('content_admin', 'Content Administrator', 'CMS and SEO management', true)
ON CONFLICT (name) DO NOTHING;

-- ============================================================================
-- STEP 11: Assign Permissions to Admin Roles
-- ============================================================================

-- Super Admin: All permissions
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'super_admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- System Admin: User, URL, Analytics, Settings, Monitoring
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'system_admin' AND p.name IN (
    'admin.user.view', 'admin.user.manage',
    'admin.url.view', 'admin.url.manage', 'admin.url.flag',
    'admin.analytics.view', 'admin.analytics.export',
    'admin.rbac.view',
    'admin.settings.view',
    'admin.audit.view',
    'admin.monitoring.view', 'admin.monitoring.manage'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Support Admin: View-only with user assistance
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'support_admin' AND p.name IN (
    'admin.user.view', 'admin.user.impersonate',
    'admin.url.view',
    'admin.analytics.view',
    'admin.audit.view'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Billing Admin: Billing and subscriptions only
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'billing_admin' AND p.name IN (
    'admin.billing.view', 'admin.billing.manage',
    'admin.analytics.view',
    'admin.audit.view'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Content Admin: CMS and SEO
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'content_admin' AND p.name IN (
    'admin.cms.view', 'admin.cms.manage',
    'admin.settings.view'
)
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- ============================================================================
-- STEP 12: Insert Default System Settings
-- ============================================================================

INSERT INTO admin.system_settings (key, value, category, description, is_public, data_type) VALUES
    -- Security Settings
    ('security.password_min_length', '8', 'security', 'Minimum password length', false, 'number'),
    ('security.password_require_uppercase', 'true', 'security', 'Require uppercase in password', false, 'boolean'),
    ('security.password_require_lowercase', 'true', 'security', 'Require lowercase in password', false, 'boolean'),
    ('security.password_require_numbers', 'true', 'security', 'Require numbers in password', false, 'boolean'),
    ('security.password_require_special', 'true', 'security', 'Require special characters in password', false, 'boolean'),
    ('security.max_login_attempts', '5', 'security', 'Maximum login attempts before lockout', false, 'number'),
    ('security.lockout_duration_minutes', '30', 'security', 'Account lockout duration in minutes', false, 'number'),
    ('security.session_timeout_minutes', '1440', 'security', 'Session timeout in minutes (24 hours)', false, 'number'),
    ('security.admin_session_timeout_minutes', '30', 'security', 'Admin session timeout in minutes', false, 'number'),
    ('security.mfa_required_for_admin', 'false', 'security', 'Require MFA for admin accounts', false, 'boolean'),

    -- Feature Flags
    ('features.user_registration_enabled', 'true', 'features', 'Allow new user registration', true, 'boolean'),
    ('features.custom_domains_enabled', 'true', 'features', 'Enable custom domains', true, 'boolean'),
    ('features.api_keys_enabled', 'true', 'features', 'Enable API key generation', true, 'boolean'),
    ('features.team_collaboration_enabled', 'true', 'features', 'Enable team collaboration features', true, 'boolean'),

    -- Rate Limiting
    ('rate_limit.url_creation_per_hour', '100', 'rate_limit', 'URLs that can be created per hour (free users)', false, 'number'),
    ('rate_limit.url_creation_per_hour_premium', '1000', 'rate_limit', 'URLs that can be created per hour (premium users)', false, 'number'),
    ('rate_limit.api_requests_per_minute', '60', 'rate_limit', 'API requests per minute', false, 'number'),

    -- Email Settings
    ('email.from_address', '"URL Shortener" <noreply@urlshortener.com>', 'email', 'Email from address', false, 'string'),
    ('email.support_email', 'support@urlshortener.com', 'email', 'Support email address', true, 'string'),

    -- System Info
    ('system.name', 'URL Shortener', 'system', 'System name', true, 'string'),
    ('system.version', '1.0.0', 'system', 'System version', true, 'string'),
    ('system.maintenance_mode', 'false', 'system', 'Maintenance mode enabled', false, 'boolean'),
    ('system.maintenance_message', 'System is under maintenance. Please check back later.', 'system', 'Maintenance mode message', true, 'string')
ON CONFLICT (key) DO NOTHING;

-- ============================================================================
-- STEP 13: Create Functions for Audit Logging
-- ============================================================================

-- Function to automatically log admin actions
CREATE OR REPLACE FUNCTION admin.log_audit(
    p_user_id BIGINT,
    p_username VARCHAR(100),
    p_action VARCHAR(100),
    p_resource_type VARCHAR(50),
    p_resource_id VARCHAR(100),
    p_details JSONB,
    p_changes JSONB,
    p_ip_address INET,
    p_user_agent TEXT,
    p_session_id VARCHAR(100),
    p_status VARCHAR(20),
    p_error_message TEXT,
    p_severity VARCHAR(20)
) RETURNS BIGINT AS $$
DECLARE
    new_id BIGINT;
BEGIN
    INSERT INTO admin.audit_logs (
        user_id, username, action, resource_type, resource_id,
        details, changes, ip_address, user_agent, session_id,
        status, error_message, severity
    ) VALUES (
        p_user_id, p_username, p_action, p_resource_type, p_resource_id,
        p_details, p_changes, p_ip_address, p_user_agent, p_session_id,
        COALESCE(p_status, 'success'), p_error_message, COALESCE(p_severity, 'info')
    ) RETURNING id INTO new_id;

    RETURN new_id;
END;
$$ LANGUAGE plpgsql;

COMMENT ON FUNCTION admin.log_audit IS 'Helper function to log admin audit events';

-- ============================================================================
-- STEP 14: Grant Permissions
-- ============================================================================

-- Grant usage on admin schema to urlshort_admin user (will be created later)
-- This will be executed after database users are created

-- ============================================================================
-- STEP 15: Cleanup and Verification
-- ============================================================================

-- Verify tables were created
DO $$
DECLARE
    table_count INTEGER;
BEGIN
    SELECT COUNT(*) INTO table_count
    FROM information_schema.tables
    WHERE table_schema = 'admin';

    RAISE NOTICE 'Created % tables in admin schema', table_count;
END $$;

-- ============================================================================
-- Migration Complete
-- ============================================================================

-- Summary of changes:
-- ✅ Created admin schema
-- ✅ Enhanced users table with admin fields
-- ✅ Created admin.audit_logs table
-- ✅ Created admin.system_settings table
-- ✅ Created admin.blocked_domains table
-- ✅ Created admin.flagged_urls table
-- ✅ Created admin.system_alerts table
-- ✅ Created admin.scheduled_reports table
-- ✅ Added 19 new admin permissions
-- ✅ Created 5 admin roles with permission assignments
-- ✅ Inserted default system settings
-- ✅ Created audit logging helper function
