-- Migration 048: Microsite published version (DOWN)
ALTER TABLE microsites DROP COLUMN IF EXISTS published_version;
