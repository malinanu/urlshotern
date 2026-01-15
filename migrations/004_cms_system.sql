-- Migration: CMS System Tables
-- Description: Creates all tables needed for the Content Management System
-- Version: 004
-- Date: 2026-01-13

-- =============================================================================
-- STATIC PAGES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS static_pages (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    content_blocks TEXT, -- JSON for structured content
    meta_description TEXT,
    meta_keywords TEXT,
    meta_image TEXT,
    meta_title VARCHAR(200),
    canonical_url TEXT,
    schema_markup TEXT, -- JSON-LD schema markup
    is_published BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0,
    template VARCHAR(50) NOT NULL DEFAULT 'default',
    author_id BIGINT, -- References users table
    published_at TIMESTAMP,
    scheduled_at TIMESTAMP,
    expires_at TIMESTAMP,
    featured_image_id BIGINT, -- References media_files table
    category VARCHAR(50),
    tags TEXT, -- JSON array
    custom_fields TEXT, -- JSON for additional fields
    view_count BIGINT NOT NULL DEFAULT 0,
    is_sticky BOOLEAN NOT NULL DEFAULT false,
    allow_comments BOOLEAN NOT NULL DEFAULT false,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_static_pages_slug ON static_pages(slug);
CREATE INDEX idx_static_pages_published ON static_pages(is_published, published_at);
CREATE INDEX idx_static_pages_category ON static_pages(category);
CREATE INDEX idx_static_pages_author ON static_pages(author_id);

-- =============================================================================
-- PAGE REVISIONS TABLE (Version Control)
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_revisions (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    meta_description TEXT,
    meta_keywords TEXT,
    revision_number INTEGER NOT NULL,
    author_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_revisions_page ON page_revisions(page_id, revision_number DESC);

-- =============================================================================
-- PAGE VISITS TABLE (Analytics)
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_visits (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    country_code VARCHAR(2),
    visited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_visits_page ON page_visits(page_id, visited_at DESC);
CREATE INDEX idx_page_visits_country ON page_visits(country_code);

-- =============================================================================
-- MEDIA FILES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS media_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_type VARCHAR(50) NOT NULL, -- image, document, video, etc.
    width INTEGER,
    height INTEGER,
    alt TEXT,
    caption TEXT,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_files_user ON media_files(user_id);
CREATE INDEX idx_media_files_type ON media_files(file_type);

-- =============================================================================
-- GLOBAL SETTINGS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS global_settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    type VARCHAR(20) NOT NULL, -- text, html, json, boolean, number
    category VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_global_settings_category ON global_settings(category);
CREATE INDEX idx_global_settings_public ON global_settings(is_public);

-- =============================================================================
-- CONTENT BLOCKS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS content_blocks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- hero, testimonial, cta, footer, etc.
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_content_blocks_type ON content_blocks(type);
CREATE INDEX idx_content_blocks_active ON content_blocks(is_active);

-- =============================================================================
-- PAGE TEMPLATES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    content TEXT NOT NULL, -- Template HTML with placeholders
    fields TEXT, -- JSON defining custom fields
    preview_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_templates_active ON page_templates(is_active);

-- =============================================================================
-- NAVIGATION MENUS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS navigation_menus (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(50) NOT NULL, -- header, footer, sidebar, etc.
    items TEXT NOT NULL, -- JSON array of menu items
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_navigation_menus_location ON navigation_menus(location);

-- =============================================================================
-- INSERT DEFAULT DATA
-- =============================================================================

-- Insert existing pages from the frontend
INSERT INTO static_pages (slug, title, content, meta_description, meta_title, is_published, is_sticky, template, language, sort_order) VALUES
('home', 'Home', '<h1>Welcome to URL Shortener</h1><p>Transform your long URLs into short, memorable links. Track analytics, manage custom domains, and boost your link marketing strategy.</p><h2>Key Features</h2><ul><li>🚀 Fast and reliable URL shortening</li><li>📊 Comprehensive analytics and tracking</li><li>🎯 Custom branded short links</li><li>🔒 Secure and private</li><li>⚡ Real-time click tracking</li><li>🌍 Geographic analytics</li></ul><h2>Get Started</h2><p>Create your first short URL in seconds. No credit card required.</p>', 'Create short URLs with advanced analytics. Track clicks, analyze traffic, and optimize your link marketing strategy with our powerful URL shortener.', 'URL Shortener - Fast, Secure & Analytics-Powered', true, true, 'default', 'en', 0),
('features', 'Features', 'Lightning fast URL shortening with advanced analytics and custom domains.', 'Explore powerful features including advanced analytics, custom domains, QR codes, and more.', NULL, true, false, 'default', 'en', 1),
('pricing', 'Pricing', 'Choose the perfect plan for your needs.', 'Flexible pricing plans for individuals and businesses.', NULL, true, false, 'default', 'en', 2),
('api-docs', 'API Documentation', 'Complete API reference for developers.', 'Developer-friendly API documentation with code examples.', NULL, true, false, 'default', 'en', 3),
('analytics', 'Analytics', 'Comprehensive analytics for your shortened URLs.', 'Track clicks, geographic data, devices, and more with detailed analytics.', NULL, true, false, 'default', 'en', 4),
('about', 'About Us', 'Learn more about our mission and team.', 'About our URL shortening service and our commitment to privacy and performance.', NULL, true, false, 'default', 'en', 5),
('contact', 'Contact', 'Get in touch with our team.', 'Contact us for support, partnerships, or general inquiries.', NULL, true, false, 'default', 'en', 6),
('privacy', 'Privacy Policy', 'Our commitment to protecting your privacy.', 'Read our privacy policy to learn how we protect and handle your data.', NULL, true, false, 'default', 'en', 7),
('terms', 'Terms of Service', 'Terms and conditions for using our service.', 'Review our terms of service before using the platform.', NULL, true, false, 'default', 'en', 8),
('conversions', 'Conversions', 'Track conversions and measure campaign success.', 'Monitor conversion events and analyze campaign performance.', NULL, true, false, 'default', 'en', 9)
ON CONFLICT (slug) DO NOTHING;

-- Insert default page templates
INSERT INTO page_templates (name, slug, description, content, is_active) VALUES
('default', 'default', 'Default page template', '<div class="container">{{content}}</div>', true),
('landing', 'landing', 'Landing page template', '<div class="landing-page">{{content}}</div>', true),
('blog', 'blog', 'Blog post template', '<article class="blog-post">{{content}}</article>', true)
ON CONFLICT (slug) DO NOTHING;

-- Insert default navigation menu
INSERT INTO navigation_menus (name, location, items, is_active) VALUES
('main', 'header', '[{"label":"Home","url":"/"},{"label":"Features","url":"/features"},{"label":"Pricing","url":"/pricing"},{"label":"API Docs","url":"/api-docs"},{"label":"About","url":"/about"},{"label":"Contact","url":"/contact"}]', true),
('footer', 'footer', '[{"label":"Privacy Policy","url":"/privacy"},{"label":"Terms of Service","url":"/terms"},{"label":"Contact","url":"/contact"}]', true);

COMMENT ON TABLE static_pages IS 'CMS static pages with SEO and analytics';
COMMENT ON TABLE page_revisions IS 'Version control for page edits';
COMMENT ON TABLE page_visits IS 'Page visit analytics';
COMMENT ON TABLE media_files IS 'Uploaded media files';
COMMENT ON TABLE global_settings IS 'Site-wide configuration settings';
COMMENT ON TABLE content_blocks IS 'Reusable content components';
COMMENT ON TABLE page_templates IS 'Page layout templates';
COMMENT ON TABLE navigation_menus IS 'Dynamic site navigation menus';
