-- Migration 048: Privacy compliance (DOWN)

DROP TABLE IF EXISTS email_suppressions;

ALTER TABLE users DROP COLUMN IF EXISTS age_verified_at;
ALTER TABLE users DROP COLUMN IF EXISTS is_minor;
ALTER TABLE users DROP COLUMN IF EXISTS date_of_birth;
