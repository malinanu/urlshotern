-- Migration 046: Microsites (DOWN)
-- Drops only the tables created by 046_microsites_up.sql.

DROP TABLE IF EXISTS microsite_events CASCADE;
DROP TABLE IF EXISTS microsite_assets CASCADE;
DROP TABLE IF EXISTS microsites CASCADE;
