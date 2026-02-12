-- Rollback: URL Mappings Extended Columns

-- Drop the view
DROP VIEW IF EXISTS urls;

-- Drop indexes
DROP INDEX IF EXISTS idx_url_mappings_is_public;
DROP INDEX IF EXISTS idx_url_mappings_team_id;
DROP INDEX IF EXISTS idx_url_mappings_user_id;

-- Remove columns from url_mappings
ALTER TABLE url_mappings DROP COLUMN IF EXISTS utm_content;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS utm_term;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS utm_campaign;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS utm_medium;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS utm_source;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS custom_domain;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS password_hash;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS is_public;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS tags;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS description;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS title;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS team_id;
ALTER TABLE url_mappings DROP COLUMN IF EXISTS user_id;
