-- Rollback: CMS Tables

-- Remove foreign key first
ALTER TABLE static_pages DROP CONSTRAINT IF EXISTS fk_static_pages_featured_image;

-- Drop tables
DROP TABLE IF EXISTS navigation_menus CASCADE;
DROP TABLE IF EXISTS page_templates CASCADE;
DROP TABLE IF EXISTS content_blocks CASCADE;
DROP TABLE IF EXISTS global_settings CASCADE;
DROP TABLE IF EXISTS media_files CASCADE;
DROP TABLE IF EXISTS page_visits CASCADE;
DROP TABLE IF EXISTS page_revisions CASCADE;
DROP TABLE IF EXISTS static_pages CASCADE;

-- Remove CMS permissions
DELETE FROM role_permissions WHERE permission_id IN (SELECT id FROM permissions WHERE name LIKE 'cms.%');
DELETE FROM permissions WHERE name LIKE 'cms.%';
