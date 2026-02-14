-- Migration 028 Rollback: Remove populated CMS content
-- Created: 2026-02-14

-- ============================================================================
-- REMOVE CONTENT BLOCKS
-- ============================================================================
DELETE FROM content_blocks WHERE identifier IN (
    'hero_home',
    'features_grid',
    'social_proof_stats',
    'cta_bottom',
    'hero_about',
    'hero_contact'
);

-- ============================================================================
-- RESET PAGES TO SEED DATA
-- ============================================================================

-- Reset homepage to minimal seed data
UPDATE static_pages
SET
    title = 'Home',
    content = 'Welcome to our URL shortener',
    content_blocks = NULL,
    meta_title = NULL,
    meta_description = NULL,
    meta_keywords = NULL,
    updated_at = NOW()
WHERE slug = 'home';

-- Reset about page to minimal seed data
UPDATE static_pages
SET
    title = 'About Us',
    content = 'Learn more about our company',
    content_blocks = NULL,
    meta_title = NULL,
    meta_description = NULL,
    meta_keywords = NULL,
    updated_at = NOW()
WHERE slug = 'about';

-- Reset contact page to minimal seed data
UPDATE static_pages
SET
    title = 'Contact',
    content = 'Get in touch with us',
    content_blocks = NULL,
    custom_fields = NULL,
    meta_title = NULL,
    meta_description = NULL,
    meta_keywords = NULL,
    updated_at = NOW()
WHERE slug = 'contact';

-- ============================================================================
-- VERIFICATION
-- ============================================================================
DO $$
DECLARE
    block_count INT;
BEGIN
    SELECT COUNT(*) INTO block_count FROM content_blocks WHERE identifier IN ('hero_home', 'features_grid', 'social_proof_stats', 'cta_bottom', 'hero_about', 'hero_contact');

    RAISE NOTICE 'Migration 028 Rollback Complete:';
    RAISE NOTICE '  - Content blocks removed: % remaining (expected: 0)', block_count;
    RAISE NOTICE '  - Pages reset to seed data: 3 (home, about, contact)';

    IF block_count > 0 THEN
        RAISE WARNING 'Expected 0 content blocks, still have %', block_count;
    END IF;
END $$;
