-- Migration: SEO Tables
-- Category: SEO
-- Tables: meta_tags, page_seo, seo_analyses, url_redirects, robots_txt
-- Dependencies: static_pages (from migration 017), users, global_settings (from migration 017)

-- Meta tags table
CREATE TABLE IF NOT EXISTS meta_tags (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT REFERENCES static_pages(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    property VARCHAR(100),
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(page_id, name, property)
);

-- Page-specific SEO data
CREATE TABLE IF NOT EXISTS page_seo (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    meta_title VARCHAR(255),
    meta_description TEXT,
    meta_keywords VARCHAR(500),
    canonical_url VARCHAR(500),
    og_title VARCHAR(255),
    og_description TEXT,
    og_image VARCHAR(500),
    twitter_card VARCHAR(50),
    twitter_title VARCHAR(255),
    twitter_description TEXT,
    twitter_image VARCHAR(500),
    schema_markup JSONB,
    no_index BOOLEAN DEFAULT FALSE,
    no_follow BOOLEAN DEFAULT FALSE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(page_id)
);

-- SEO analysis results
CREATE TABLE IF NOT EXISTS seo_analyses (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT REFERENCES static_pages(id) ON DELETE SET NULL,
    url VARCHAR(500) NOT NULL,
    title VARCHAR(255),
    meta_description TEXT,
    h1_tags JSONB,
    h2_tags JSONB,
    image_alt_count INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    internal_links INTEGER DEFAULT 0,
    external_links INTEGER DEFAULT 0,
    mobile_friendly BOOLEAN DEFAULT FALSE,
    loading_speed DECIMAL(5,2) DEFAULT 0.0,
    ssl_enabled BOOLEAN DEFAULT FALSE,
    schema_markup JSONB,
    score INTEGER DEFAULT 0,
    issues JSONB,
    recommendations JSONB,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- URL redirects
CREATE TABLE IF NOT EXISTS url_redirects (
    id BIGSERIAL PRIMARY KEY,
    source_url VARCHAR(500) NOT NULL,
    target_url VARCHAR(500) NOT NULL,
    redirect_type INTEGER NOT NULL DEFAULT 301,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(source_url)
);

-- Robots.txt configuration
CREATE TABLE IF NOT EXISTS robots_txt (
    id BIGSERIAL PRIMARY KEY,
    content TEXT NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT NOT NULL REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_meta_tags_page_id ON meta_tags(page_id);
CREATE INDEX IF NOT EXISTS idx_meta_tags_name ON meta_tags(name);
CREATE INDEX IF NOT EXISTS idx_page_seo_page_id ON page_seo(page_id);
CREATE INDEX IF NOT EXISTS idx_url_redirects_source ON url_redirects(source_url);
CREATE INDEX IF NOT EXISTS idx_url_redirects_active ON url_redirects(is_active);
CREATE INDEX IF NOT EXISTS idx_seo_analyses_page_id ON seo_analyses(page_id);
CREATE INDEX IF NOT EXISTS idx_seo_analyses_url ON seo_analyses(url);
CREATE INDEX IF NOT EXISTS idx_seo_analyses_created_by ON seo_analyses(created_by);
CREATE INDEX IF NOT EXISTS idx_seo_analyses_score ON seo_analyses(score);

-- Insert default SEO settings into global_settings
INSERT INTO global_settings (key, value, type, category, display_name, description, is_public, sort_order)
VALUES
    ('site_title', 'Trunc', 'text', 'seo', 'Site Title', 'Default site title for SEO', true, 20),
    ('site_description', 'Professional URL shortener service', 'text', 'seo', 'Site Description', 'Default site description for SEO', true, 21),
    ('site_keywords', 'url shortener, link shortener, custom links', 'text', 'seo', 'Site Keywords', 'Default keywords for SEO', true, 22),
    ('google_analytics_id', '', 'text', 'seo', 'Google Analytics ID', 'Google Analytics tracking ID', false, 23),
    ('google_search_console_verification', '', 'text', 'seo', 'Google Search Console', 'Google Search Console verification code', false, 24),
    ('og_image', '', 'text', 'seo', 'Default OG Image', 'Default Open Graph image URL', true, 25),
    ('twitter_site', '@trunc', 'text', 'seo', 'Twitter Site Handle', 'Twitter site handle for cards', true, 26),
    ('robots_txt_content', 'User-agent: *\nDisallow:\nSitemap: /sitemap.xml', 'textarea', 'seo', 'Robots.txt Content', 'Content for robots.txt file', true, 27),
    ('sitemap_xml', '', 'textarea', 'seo', 'Sitemap XML', 'Generated XML sitemap content', false, 28)
ON CONFLICT (key) DO NOTHING;

-- Insert default robots.txt
INSERT INTO robots_txt (content, is_active, created_by, created_at, updated_by, updated_at)
SELECT
    E'User-agent: *\nDisallow: /admin\nDisallow: /api\nAllow: /\n\nSitemap: /sitemap.xml',
    true,
    u.id,
    NOW(),
    u.id,
    NOW()
FROM users u
WHERE u.email = 'admin@trunc.site' OR u.id = 1
LIMIT 1;

-- Add SEO permissions
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    ('seo.read', 'Read SEO Data', 'View SEO analyses and settings', 'seo', 'read', true),
    ('seo.create', 'Create SEO Data', 'Create SEO analyses and settings', 'seo', 'create', true),
    ('seo.update', 'Update SEO Data', 'Update SEO settings and meta tags', 'seo', 'update', true),
    ('seo.delete', 'Delete SEO Data', 'Delete SEO analyses', 'seo', 'delete', true)
ON CONFLICT (name) DO NOTHING;

-- Assign SEO permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('seo.read', 'seo.create', 'seo.update', 'seo.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Verification for SEO tables
DO $$
DECLARE
    seo_table_count INTEGER;
    seo_perm_count INTEGER;
    seo_settings_count INTEGER;
BEGIN
    -- Count SEO tables
    SELECT COUNT(*) INTO seo_table_count
    FROM information_schema.tables
    WHERE table_name IN ('meta_tags', 'page_seo', 'seo_analyses', 'url_redirects', 'robots_txt');

    -- Count SEO permissions
    SELECT COUNT(*) INTO seo_perm_count
    FROM permissions
    WHERE resource = 'seo' AND is_system = true;

    -- Count SEO settings
    SELECT COUNT(*) INTO seo_settings_count
    FROM global_settings
    WHERE category = 'seo';

    IF seo_table_count < 5 THEN
        RAISE EXCEPTION 'SEO Migration Failed: Expected 5 SEO tables, found %', seo_table_count;
    END IF;

    IF seo_perm_count < 4 THEN
        RAISE EXCEPTION 'SEO Migration Failed: Expected 4 SEO permissions, found %', seo_perm_count;
    END IF;

    IF seo_settings_count < 9 THEN
        RAISE NOTICE 'SEO Migration Warning: Expected 9 SEO settings, found %', seo_settings_count;
    END IF;

    RAISE NOTICE 'SEO Migration Verification: % tables, % permissions, % settings ✓', seo_table_count, seo_perm_count, seo_settings_count;
END $$;
