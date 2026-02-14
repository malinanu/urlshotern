-- Rollback Migration 029: Remove Missing Pages
-- Created: 2026-02-15
-- Purpose: Rollback the addition of features, pricing, privacy, and terms pages

-- ============================================================================
-- REMOVE PAGES ADDED IN MIGRATION 029
-- ============================================================================

DELETE FROM static_pages
WHERE slug IN ('features', 'pricing', 'privacy', 'terms');

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Verify pages were removed
DO $$
DECLARE
    removed_count INT;
    remaining_count INT;
BEGIN
    SELECT COUNT(*) INTO removed_count FROM static_pages
    WHERE slug IN ('features', 'pricing', 'privacy', 'terms');

    SELECT COUNT(*) INTO remaining_count FROM static_pages;

    RAISE NOTICE 'Migration 029 Rollback Complete:';
    RAISE NOTICE '  - Pages removed: 4';
    RAISE NOTICE '  - Remaining pages that should have been removed: % (expected: 0)', removed_count;
    RAISE NOTICE '  - Total pages remaining in database: % (expected: 3)', remaining_count;

    IF removed_count > 0 THEN
        RAISE WARNING 'Some pages were not removed: %', removed_count;
    END IF;
END $$;
