-- Create database (this may already exist from Docker environment variables)
-- CREATE DATABASE urlshortener_db;

-- Create user (this may already exist from Docker environment variables)
-- CREATE USER urlshortener WITH PASSWORD 'password';
-- GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener;

-- Connect to database
\c urlshortener_db;

-- Create tables
CREATE TABLE IF NOT EXISTS url_mappings (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_by_ip INET
);

-- Create indexes for url_mappings
CREATE INDEX IF NOT EXISTS idx_short_code ON url_mappings(short_code);
CREATE INDEX IF NOT EXISTS idx_created_at ON url_mappings(created_at);
CREATE INDEX IF NOT EXISTS idx_active ON url_mappings(is_active);
CREATE INDEX IF NOT EXISTS idx_expires_at ON url_mappings(expires_at);

-- Create click events table
CREATE TABLE IF NOT EXISTS click_events (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country_code CHAR(2)
);

-- Create indexes for click_events
CREATE INDEX IF NOT EXISTS idx_short_code_time ON click_events(short_code, clicked_at);
CREATE INDEX IF NOT EXISTS idx_clicked_at ON click_events(clicked_at);
CREATE INDEX IF NOT EXISTS idx_country_code ON click_events(country_code);

-- Add foreign key constraint
ALTER TABLE click_events 
ADD CONSTRAINT fk_click_events_short_code 
FOREIGN KEY (short_code) 
REFERENCES url_mappings(short_code) 
ON DELETE CASCADE;

-- User Management Tables

-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGSERIAL PRIMARY KEY,
    email VARCHAR(255) UNIQUE NOT NULL,
    username VARCHAR(50) UNIQUE,
    password_hash VARCHAR(255) NOT NULL,
    first_name VARCHAR(100),
    last_name VARCHAR(100),
    phone VARCHAR(20),
    avatar_url TEXT,
    bio TEXT,
    timezone VARCHAR(50) DEFAULT 'UTC',
    language VARCHAR(10) DEFAULT 'en',
    is_email_verified BOOLEAN DEFAULT FALSE,
    is_phone_verified BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    email_verification_token VARCHAR(255),
    email_verification_expires TIMESTAMP WITH TIME ZONE,
    phone_verification_code VARCHAR(6),
    phone_verification_expires TIMESTAMP WITH TIME ZONE,
    password_reset_token VARCHAR(255),
    password_reset_expires TIMESTAMP WITH TIME ZONE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    last_login_ip INET,
    failed_login_attempts INTEGER DEFAULT 0,
    locked_until TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User preferences table
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id BIGINT PRIMARY KEY REFERENCES users(id) ON DELETE CASCADE,
    theme VARCHAR(20) DEFAULT 'light',
    notifications_email BOOLEAN DEFAULT TRUE,
    notifications_push BOOLEAN DEFAULT TRUE,
    notifications_sms BOOLEAN DEFAULT FALSE,
    analytics_enabled BOOLEAN DEFAULT TRUE,
    public_profile BOOLEAN DEFAULT FALSE,
    custom_domain_limit INTEGER DEFAULT 0,
    url_limit INTEGER DEFAULT 1000,
    click_tracking BOOLEAN DEFAULT TRUE,
    password_expires_at TIMESTAMP WITH TIME ZONE,
    two_factor_enabled BOOLEAN DEFAULT FALSE,
    two_factor_secret VARCHAR(32),
    backup_codes JSONB,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- OAuth providers table
CREATE TABLE IF NOT EXISTS oauth_providers (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    provider VARCHAR(50) NOT NULL,
    provider_user_id VARCHAR(255) NOT NULL,
    provider_email VARCHAR(255),
    access_token TEXT,
    refresh_token TEXT,
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(provider, provider_user_id)
);

-- User sessions table
CREATE TABLE IF NOT EXISTS user_sessions (
    id VARCHAR(128) PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    device_type VARCHAR(20),
    browser VARCHAR(50),
    platform VARCHAR(50),
    location VARCHAR(100),
    is_mobile BOOLEAN DEFAULT FALSE,
    screen_resolution VARCHAR(20),
    timezone VARCHAR(50),
    started_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ended_at TIMESTAMP WITH TIME ZONE,
    duration_minutes INTEGER,
    is_active BOOLEAN DEFAULT TRUE
);

-- RBAC Tables

-- Roles table
CREATE TABLE IF NOT EXISTS roles (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) UNIQUE NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_system BOOLEAN DEFAULT FALSE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Permissions table
CREATE TABLE IF NOT EXISTS permissions (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) UNIQUE NOT NULL,
    display_name VARCHAR(150) NOT NULL,
    description TEXT,
    resource VARCHAR(50),
    action VARCHAR(50),
    is_system BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Role permissions junction table
CREATE TABLE IF NOT EXISTS role_permissions (
    id BIGSERIAL PRIMARY KEY,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    permission_id BIGINT NOT NULL REFERENCES permissions(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(role_id, permission_id)
);

-- User roles junction table
CREATE TABLE IF NOT EXISTS user_roles (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role_id BIGINT NOT NULL REFERENCES roles(id) ON DELETE CASCADE,
    assigned_by BIGINT REFERENCES users(id),
    assigned_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    UNIQUE(user_id, role_id)
);

-- Team Management Tables

-- Teams table
CREATE TABLE IF NOT EXISTS teams (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    logo_url TEXT,
    owner_id BIGINT NOT NULL REFERENCES users(id),
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Team members table
CREATE TABLE IF NOT EXISTS team_members (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    role VARCHAR(20) DEFAULT 'member',
    permissions JSONB DEFAULT '[]',
    invited_by BIGINT REFERENCES users(id),
    joined_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(team_id, user_id)
);

-- Team invitations table
CREATE TABLE IF NOT EXISTS team_invitations (
    id BIGSERIAL PRIMARY KEY,
    team_id BIGINT NOT NULL REFERENCES teams(id) ON DELETE CASCADE,
    email VARCHAR(255) NOT NULL,
    role VARCHAR(20) DEFAULT 'member',
    permissions JSONB DEFAULT '[]',
    invited_by BIGINT NOT NULL REFERENCES users(id),
    invitation_token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE NOT NULL,
    accepted_at TIMESTAMP WITH TIME ZONE,
    accepted_by BIGINT REFERENCES users(id),
    declined_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Update URL mappings table to support user ownership
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS user_id BIGINT REFERENCES users(id);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS team_id BIGINT REFERENCES teams(id);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS title VARCHAR(255);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]';
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT TRUE;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS custom_domain VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_source VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_medium VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_campaign VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_term VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_content VARCHAR(100);

-- Create an alias for the url_mappings table as 'urls' for consistency
CREATE VIEW IF NOT EXISTS urls AS SELECT * FROM url_mappings;

-- Custom Domains Tables

-- Custom domains table
CREATE TABLE IF NOT EXISTS custom_domains (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    domain VARCHAR(255) UNIQUE NOT NULL,
    verification_method VARCHAR(20) DEFAULT 'dns',
    verification_token VARCHAR(255) UNIQUE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    ssl_enabled BOOLEAN DEFAULT FALSE,
    ssl_cert_path TEXT,
    ssl_key_path TEXT,
    ssl_expires_at TIMESTAMP WITH TIME ZONE,
    health_check_url TEXT,
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    dns_records JSONB DEFAULT '{}',
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Domain analytics table
CREATE TABLE IF NOT EXISTS domain_analytics (
    id BIGSERIAL PRIMARY KEY,
    domain_id BIGINT NOT NULL REFERENCES custom_domains(id) ON DELETE CASCADE,
    total_urls INTEGER DEFAULT 0,
    total_clicks BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(domain_id, date)
);

-- Collaboration Tables

-- URL shares table
CREATE TABLE IF NOT EXISTS url_shares (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    sharer_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    shared_with_type VARCHAR(20) NOT NULL,
    shared_with_id BIGINT,
    share_type VARCHAR(20) NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- URL comments table
CREATE TABLE IF NOT EXISTS url_comments (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    content TEXT NOT NULL,
    parent_id BIGINT REFERENCES url_comments(id),
    is_edited BOOLEAN DEFAULT FALSE,
    is_deleted BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- URL bookmarks table
CREATE TABLE IF NOT EXISTS url_bookmarks (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    tags TEXT,
    notes TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, url_id)
);

-- URL favorites table
CREATE TABLE IF NOT EXISTS url_favorites (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(user_id, url_id)
);

-- URL collections table
CREATE TABLE IF NOT EXISTS url_collections (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    color VARCHAR(7) DEFAULT '#3b82f6',
    icon VARCHAR(50) DEFAULT 'folder',
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    deleted_at TIMESTAMP WITH TIME ZONE
);

-- URL collection items junction table
CREATE TABLE IF NOT EXISTS url_collection_items (
    id BIGSERIAL PRIMARY KEY,
    collection_id BIGINT NOT NULL REFERENCES url_collections(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    sort_order INTEGER DEFAULT 0,
    added_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(collection_id, url_id)
);

-- URL notes table
CREATE TABLE IF NOT EXISTS url_notes (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    title VARCHAR(200),
    content TEXT NOT NULL,
    is_private BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- URL activity table
CREATE TABLE IF NOT EXISTS url_activities (
    id BIGSERIAL PRIMARY KEY,
    url_id BIGINT NOT NULL REFERENCES url_mappings(id) ON DELETE CASCADE,
    user_id BIGINT REFERENCES users(id),
    activity_type VARCHAR(50) NOT NULL,
    description TEXT NOT NULL,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User Analytics Tables

-- User activity logs table
CREATE TABLE IF NOT EXISTS user_activity_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    session_id VARCHAR(128),
    activity_type VARCHAR(50) NOT NULL,
    description TEXT,
    url_id BIGINT REFERENCES url_mappings(id),
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    device_type VARCHAR(20),
    browser VARCHAR(50),
    platform VARCHAR(50),
    location VARCHAR(100),
    duration_ms BIGINT,
    screen_resolution VARCHAR(20),
    timezone VARCHAR(50),
    is_mobile BOOLEAN DEFAULT FALSE,
    is_bot BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Audit logs table
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT REFERENCES users(id),
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    old_values JSONB,
    new_values JSONB,
    metadata JSONB DEFAULT '{}',
    ip_address INET,
    user_agent TEXT,
    success BOOLEAN DEFAULT TRUE,
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for performance

-- User table indexes
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_username ON users(username);
CREATE INDEX IF NOT EXISTS idx_users_is_active ON users(is_active);
CREATE INDEX IF NOT EXISTS idx_users_created_at ON users(created_at);

-- User sessions indexes
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_started_at ON user_sessions(started_at);
CREATE INDEX IF NOT EXISTS idx_user_sessions_active ON user_sessions(is_active);

-- URL mappings indexes (additional to existing ones)
CREATE INDEX IF NOT EXISTS idx_url_mappings_user_id ON url_mappings(user_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_team_id ON url_mappings(team_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_is_public ON url_mappings(is_public);

-- Team indexes
CREATE INDEX IF NOT EXISTS idx_teams_owner_id ON teams(owner_id);
CREATE INDEX IF NOT EXISTS idx_teams_slug ON teams(slug);
CREATE INDEX IF NOT EXISTS idx_team_members_team_id ON team_members(team_id);
CREATE INDEX IF NOT EXISTS idx_team_members_user_id ON team_members(user_id);

-- Custom domain indexes
CREATE INDEX IF NOT EXISTS idx_custom_domains_user_id ON custom_domains(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX IF NOT EXISTS idx_custom_domains_verified ON custom_domains(is_verified);

-- Collaboration indexes
CREATE INDEX IF NOT EXISTS idx_url_shares_url_id ON url_shares(url_id);
CREATE INDEX IF NOT EXISTS idx_url_shares_sharer_id ON url_shares(sharer_id);
CREATE INDEX IF NOT EXISTS idx_url_comments_url_id ON url_comments(url_id);
CREATE INDEX IF NOT EXISTS idx_url_comments_user_id ON url_comments(user_id);
CREATE INDEX IF NOT EXISTS idx_url_bookmarks_user_id ON url_bookmarks(user_id);
CREATE INDEX IF NOT EXISTS idx_url_favorites_user_id ON url_favorites(user_id);
CREATE INDEX IF NOT EXISTS idx_url_collection_items_collection_id ON url_collection_items(collection_id);

-- Analytics indexes
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_user_id ON user_activity_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_session_id ON user_activity_logs(session_id);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_activity_type ON user_activity_logs(activity_type);
CREATE INDEX IF NOT EXISTS idx_user_activity_logs_created_at ON user_activity_logs(created_at);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_action ON audit_logs(action);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);

-- Insert default roles and permissions
INSERT INTO roles (name, display_name, description, is_system) VALUES
    ('admin', 'Administrator', 'Full system access', true),
    ('user', 'User', 'Standard user access', true),
    ('premium', 'Premium User', 'Premium user with additional features', true),
    ('team_admin', 'Team Administrator', 'Team administration access', true),
    ('team_member', 'Team Member', 'Team member access', true)
ON CONFLICT (name) DO NOTHING;

INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    ('user.create', 'Create Users', 'Create new users', 'user', 'create', true),
    ('user.read', 'Read Users', 'View user information', 'user', 'read', true),
    ('user.update', 'Update Users', 'Update user information', 'user', 'update', true),
    ('user.delete', 'Delete Users', 'Delete users', 'user', 'delete', true),
    ('url.create', 'Create URLs', 'Create new URLs', 'url', 'create', true),
    ('url.read', 'Read URLs', 'View URLs', 'url', 'read', true),
    ('url.update', 'Update URLs', 'Update URLs', 'url', 'update', true),
    ('url.delete', 'Delete URLs', 'Delete URLs', 'url', 'delete', true),
    ('analytics.read', 'Read Analytics', 'View analytics data', 'analytics', 'read', true),
    ('team.create', 'Create Teams', 'Create new teams', 'team', 'create', true),
    ('team.read', 'Read Teams', 'View team information', 'team', 'read', true),
    ('team.update', 'Update Teams', 'Update team information', 'team', 'update', true),
    ('team.delete', 'Delete Teams', 'Delete teams', 'team', 'delete', true),
    ('domain.create', 'Create Custom Domains', 'Add custom domains', 'domain', 'create', true),
    ('domain.read', 'Read Custom Domains', 'View custom domains', 'domain', 'read', true),
    ('domain.update', 'Update Custom Domains', 'Update custom domains', 'domain', 'update', true),
    ('domain.delete', 'Delete Custom Domains', 'Delete custom domains', 'domain', 'delete', true)
ON CONFLICT (name) DO NOTHING;

-- Assign permissions to roles
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin'
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'user' AND p.name IN ('url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read', 'user.read', 'user.update')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'premium' AND p.name IN ('url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read', 'user.read', 'user.update', 'domain.create', 'domain.read', 'domain.update', 'domain.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'team_admin' AND p.name IN ('team.read', 'team.update', 'team.delete', 'url.create', 'url.read', 'url.update', 'url.delete', 'analytics.read')
ON CONFLICT (role_id, permission_id) DO NOTHING;

INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'team_member' AND p.name IN ('team.read', 'url.read', 'analytics.read')
ON CONFLICT (role_id, permission_id) DO NOTHING;