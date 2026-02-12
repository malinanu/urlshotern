-- Migration 023 Down: Remove verification migration record
-- Note: Does NOT remove actual data (handled by 013 down migration)
DELETE FROM schema_migrations WHERE version = '023';
