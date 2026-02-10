-- Migration: RBAC (Role-Based Access Control) Tables
-- Category: RBAC
-- Tables: oauth_providers, roles, permissions, role_permissions, user_roles
-- Dependencies: users table

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
