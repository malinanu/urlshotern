-- ============================================================================
-- Migration Tracking Table
-- ============================================================================
-- Description: Creates the schema_migrations table to track applied migrations
-- This table enables:
--   - Idempotent migration execution (avoid duplicate runs)
--   - Migration version tracking
--   - Rollback dependency validation
--   - Audit trail of schema changes
--
-- Usage: This should be the first migration executed
-- Run manually: psql -U urlshortener -d urlshortener_db -f configs/migrations/000_schema_migrations.sql
-- ============================================================================

CREATE TABLE IF NOT EXISTS schema_migrations (
    version VARCHAR(50) PRIMARY KEY,
    applied_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    applied_by VARCHAR(100) DEFAULT CURRENT_USER,
    execution_time_ms INTEGER,
    description TEXT,
    checksum VARCHAR(64)  -- SHA256 hash of migration file content for integrity validation
);

-- Index for quick chronological lookups
CREATE INDEX IF NOT EXISTS idx_schema_migrations_applied_at
ON schema_migrations(applied_at DESC);

COMMENT ON TABLE schema_migrations IS 'Tracks which database migrations have been applied';
COMMENT ON COLUMN schema_migrations.version IS 'Migration version number (e.g., 001, 002, 003)';
COMMENT ON COLUMN schema_migrations.applied_at IS 'Timestamp when migration was applied';
COMMENT ON COLUMN schema_migrations.applied_by IS 'Database user who applied the migration';
COMMENT ON COLUMN schema_migrations.execution_time_ms IS 'Migration execution time in milliseconds';
COMMENT ON COLUMN schema_migrations.description IS 'Human-readable description of the migration';
COMMENT ON COLUMN schema_migrations.checksum IS 'SHA256 checksum of migration file for integrity verification';
