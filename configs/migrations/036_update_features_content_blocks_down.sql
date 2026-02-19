-- Migration 036 Rollback: Remove added content block sections from features page
-- Preserves the hero section which was already working before this migration.

UPDATE static_pages
SET content_blocks = content_blocks - 'main_features' - 'additional_features' - 'cta_bottom',
    updated_at = NOW()
WHERE slug = 'features';
