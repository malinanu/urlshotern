-- SEO-related database tables

-- Meta tags table
CREATE TABLE IF NOT EXISTS meta_tags (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT REFERENCES static_pages(id) ON DELETE CASCADE,
    name VARCHAR(100) NOT NULL,
    content TEXT NOT NULL,
    property VARCHAR(100), -- For Open Graph and Twitter properties
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

-- URL redirects for SEO
CREATE TABLE IF NOT EXISTS url_redirects (
    id BIGSERIAL PRIMARY KEY,
    source_url VARCHAR(500) NOT NULL,
    target_url VARCHAR(500) NOT NULL,
    redirect_type INTEGER NOT NULL DEFAULT 301, -- 301, 302, etc.
    is_active BOOLEAN DEFAULT TRUE,
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW(),
    updated_by BIGINT REFERENCES users(id),
    updated_at TIMESTAMP NOT NULL DEFAULT NOW(),
    UNIQUE(source_url)
);

-- SEO analysis results
CREATE TABLE IF NOT EXISTS seo_analyses (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT REFERENCES static_pages(id) ON DELETE SET NULL,
    url VARCHAR(500) NOT NULL,
    title VARCHAR(255),
    meta_description TEXT,
    h1_tags JSONB, -- Array of H1 tags
    h2_tags JSONB, -- Array of H2 tags
    image_alt_count INTEGER DEFAULT 0,
    word_count INTEGER DEFAULT 0,
    internal_links INTEGER DEFAULT 0,
    external_links INTEGER DEFAULT 0,
    mobile_friendly BOOLEAN DEFAULT FALSE,
    loading_speed DECIMAL(5,2) DEFAULT 0.0, -- In seconds
    ssl_enabled BOOLEAN DEFAULT FALSE,
    schema_markup JSONB, -- Structured data
    score INTEGER DEFAULT 0, -- SEO score out of 100
    issues JSONB, -- Array of SEO issues
    recommendations JSONB, -- Array of recommendations
    created_by BIGINT NOT NULL REFERENCES users(id),
    created_at TIMESTAMP NOT NULL DEFAULT NOW()
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

-- Create indexes for better performance
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
INSERT INTO global_settings (key, value, type, category, display_name, description, is_public, sort_order, created_by, created_at, updated_at)
VALUES 
    ('site_title', '3logiq', 'text', 'seo', 'Site Title', 'Default site title for SEO', true, 1, 1, NOW(), NOW()),
    ('site_description', 'Professional URL shortener service', 'text', 'seo', 'Site Description', 'Default site description for SEO', true, 2, 1, NOW(), NOW()),
    ('site_keywords', 'url shortener, link shortener, custom links', 'text', 'seo', 'Site Keywords', 'Default keywords for SEO', true, 3, 1, NOW(), NOW()),
    ('google_analytics_id', '', 'text', 'seo', 'Google Analytics ID', 'Google Analytics tracking ID', false, 4, 1, NOW(), NOW()),
    ('google_search_console_verification', '', 'text', 'seo', 'Google Search Console', 'Google Search Console verification code', false, 5, 1, NOW(), NOW()),
    ('og_image', '', 'text', 'seo', 'Default OG Image', 'Default Open Graph image URL', true, 6, 1, NOW(), NOW()),
    ('twitter_site', '@3logiq', 'text', 'seo', 'Twitter Site Handle', 'Twitter site handle for cards', true, 7, 1, NOW(), NOW()),
    ('robots_txt', 'User-agent: *\nDisallow:\nSitemap: /sitemap.xml', 'textarea', 'seo', 'Robots.txt Content', 'Content for robots.txt file', true, 8, 1, NOW(), NOW()),
    ('sitemap_xml', '', 'textarea', 'seo', 'Sitemap XML', 'Generated XML sitemap content', false, 9, 1, NOW(), NOW())
ON CONFLICT (key) DO NOTHING;

-- Insert default robots.txt
INSERT INTO robots_txt (content, is_active, created_by, created_at, updated_by, updated_at)
VALUES (
    E'User-agent: *\nDisallow: /admin\nDisallow: /api\nAllow: /\n\nSitemap: /sitemap.xml',
    true,
    1,
    NOW(),
    1,
    NOW()
) ON CONFLICT DO NOTHING;