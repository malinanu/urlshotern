-- Migration: CMS Tables
-- Category: CMS
-- Tables: static_pages, page_revisions, page_visits, media_files, global_settings, content_blocks, page_templates, navigation_menus
-- Dependencies: users table

-- Static pages table for CMS
CREATE TABLE IF NOT EXISTS static_pages (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    content_blocks JSONB,
    meta_description TEXT,
    meta_keywords TEXT,
    meta_image VARCHAR(255),
    meta_title VARCHAR(200),
    canonical_url VARCHAR(255),
    schema_markup JSONB,
    is_published BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    template VARCHAR(50) DEFAULT 'default',
    author_id BIGINT REFERENCES users(id),
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    featured_image_id BIGINT,
    category VARCHAR(50),
    tags JSONB,
    custom_fields JSONB,
    view_count BIGINT DEFAULT 0,
    is_sticky BOOLEAN DEFAULT FALSE,
    allow_comments BOOLEAN DEFAULT TRUE,
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page revisions for version control
CREATE TABLE IF NOT EXISTS page_revisions (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    meta_description TEXT,
    meta_keywords TEXT,
    revision_number INTEGER NOT NULL,
    author_id BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page visits tracking
CREATE TABLE IF NOT EXISTS page_visits (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country_code CHAR(2),
    visited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Media files table for file management
CREATE TABLE IF NOT EXISTS media_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    width INTEGER,
    height INTEGER,
    alt TEXT,
    caption TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Global site settings table
CREATE TABLE IF NOT EXISTS global_settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'text',
    category VARCHAR(50) DEFAULT 'general',
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content blocks table for reusable components
CREATE TABLE IF NOT EXISTS content_blocks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    identifier VARCHAR(100) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'generic',
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page templates table
CREATE TABLE IF NOT EXISTS page_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    fields JSONB,
    preview_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Navigation menus table
CREATE TABLE IF NOT EXISTS navigation_menus (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(50) NOT NULL,
    items JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_static_pages_slug ON static_pages(slug);
CREATE INDEX IF NOT EXISTS idx_static_pages_published ON static_pages(is_published);
CREATE INDEX IF NOT EXISTS idx_static_pages_sort_order ON static_pages(sort_order);
CREATE INDEX IF NOT EXISTS idx_static_pages_author_id ON static_pages(author_id);

CREATE INDEX IF NOT EXISTS idx_page_revisions_page_id ON page_revisions(page_id);
CREATE INDEX IF NOT EXISTS idx_page_visits_page_id ON page_visits(page_id);
CREATE INDEX IF NOT EXISTS idx_page_visits_visited_at ON page_visits(visited_at);

CREATE INDEX IF NOT EXISTS idx_media_files_user_id ON media_files(user_id);
CREATE INDEX IF NOT EXISTS idx_media_files_file_type ON media_files(file_type);
CREATE INDEX IF NOT EXISTS idx_media_files_is_public ON media_files(is_public);
CREATE INDEX IF NOT EXISTS idx_media_files_created_at ON media_files(created_at);

CREATE INDEX IF NOT EXISTS idx_global_settings_key ON global_settings(key);
CREATE INDEX IF NOT EXISTS idx_global_settings_category ON global_settings(category);
CREATE INDEX IF NOT EXISTS idx_global_settings_is_public ON global_settings(is_public);

CREATE INDEX IF NOT EXISTS idx_content_blocks_identifier ON content_blocks(identifier);
CREATE INDEX IF NOT EXISTS idx_content_blocks_type ON content_blocks(type);
CREATE INDEX IF NOT EXISTS idx_content_blocks_is_active ON content_blocks(is_active);

CREATE INDEX IF NOT EXISTS idx_page_templates_slug ON page_templates(slug);
CREATE INDEX IF NOT EXISTS idx_page_templates_is_active ON page_templates(is_active);

CREATE INDEX IF NOT EXISTS idx_navigation_menus_location ON navigation_menus(location);
CREATE INDEX IF NOT EXISTS idx_navigation_menus_is_active ON navigation_menus(is_active);

-- Add foreign key for featured image
ALTER TABLE static_pages ADD CONSTRAINT fk_static_pages_featured_image
    FOREIGN KEY (featured_image_id) REFERENCES media_files(id) ON DELETE SET NULL;

-- Add CMS permissions
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    ('cms.read', 'Read CMS Content', 'View CMS pages and content', 'cms', 'read', true),
    ('cms.create', 'Create CMS Content', 'Create new CMS pages', 'cms', 'create', true),
    ('cms.update', 'Update CMS Content', 'Edit existing CMS pages', 'cms', 'update', true),
    ('cms.delete', 'Delete CMS Content', 'Delete CMS pages', 'cms', 'delete', true),
    ('cms.publish', 'Publish CMS Content', 'Publish and unpublish pages', 'cms', 'publish', true)
ON CONFLICT (name) DO NOTHING;

-- Assign CMS permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('cms.read', 'cms.create', 'cms.update', 'cms.delete', 'cms.publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;
