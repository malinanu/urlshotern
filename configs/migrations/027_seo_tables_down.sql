-- Rollback SEO tables migration

-- Drop tables in reverse order (considering dependencies)
DROP TABLE IF EXISTS robots_txt CASCADE;
DROP TABLE IF EXISTS url_redirects CASCADE;
DROP TABLE IF EXISTS seo_analyses CASCADE;
DROP TABLE IF EXISTS page_seo CASCADE;
DROP TABLE IF EXISTS meta_tags CASCADE;

-- Remove SEO permissions
DELETE FROM role_permissions WHERE permission_id IN (
    SELECT id FROM permissions WHERE resource = 'seo'
);
DELETE FROM permissions WHERE resource = 'seo';

-- Remove SEO-related global_settings
DELETE FROM global_settings WHERE category = 'seo' OR key IN (
    'site_title', 'site_description', 'site_keywords',
    'google_analytics_id', 'google_search_console_verification',
    'og_image', 'twitter_site', 'robots_txt_content', 'sitemap_xml'
);

-- Verification
DO $$
DECLARE
    seo_table_count INTEGER;
    seo_perm_count INTEGER;
BEGIN
    -- Verify SEO tables are dropped
    SELECT COUNT(*) INTO seo_table_count
    FROM information_schema.tables
    WHERE table_name IN ('meta_tags', 'page_seo', 'seo_analyses', 'url_redirects', 'robots_txt');

    -- Verify SEO permissions are removed
    SELECT COUNT(*) INTO seo_perm_count
    FROM permissions
    WHERE resource = 'seo';

    IF seo_table_count > 0 THEN
        RAISE WARNING 'SEO Rollback Warning: % SEO tables still exist', seo_table_count;
    END IF;

    IF seo_perm_count > 0 THEN
        RAISE WARNING 'SEO Rollback Warning: % SEO permissions still exist', seo_perm_count;
    END IF;

    RAISE NOTICE 'SEO Rollback Complete: % tables remaining, % permissions remaining', seo_table_count, seo_perm_count;
END $$;
