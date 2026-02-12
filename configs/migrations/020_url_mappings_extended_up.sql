-- Migration: URL Mappings Extended Columns
-- Category: URL Management
-- Tables: url_mappings (ALTER TABLE to add columns)
-- Dependencies: users, teams tables

-- Update URL mappings table to support user ownership and additional features
-- Add columns first (PostgreSQL doesn't allow inline REFERENCES in ALTER TABLE ADD COLUMN)
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS user_id BIGINT;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS team_id BIGINT;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS title VARCHAR(255);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS description TEXT;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS tags JSONB DEFAULT '[]';
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT TRUE;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS password_hash VARCHAR(255);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS custom_domain VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_source VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_medium VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_campaign VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_term VARCHAR(100);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS utm_content VARCHAR(100);

-- Create indexes for new columns
CREATE INDEX IF NOT EXISTS idx_url_mappings_user_id ON url_mappings(user_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_team_id ON url_mappings(team_id);
CREATE INDEX IF NOT EXISTS idx_url_mappings_is_public ON url_mappings(is_public);

-- Add foreign key constraints separately (with existence checks for idempotency)
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_url_mappings_user_id'
    ) THEN
        ALTER TABLE url_mappings
        ADD CONSTRAINT fk_url_mappings_user_id
        FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE;
    END IF;

    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints
        WHERE constraint_name = 'fk_url_mappings_team_id'
    ) THEN
        ALTER TABLE url_mappings
        ADD CONSTRAINT fk_url_mappings_team_id
        FOREIGN KEY (team_id) REFERENCES teams(id) ON DELETE CASCADE;
    END IF;
END $$;

-- Create an alias for the url_mappings table as 'urls' for consistency
CREATE OR REPLACE VIEW urls AS SELECT * FROM url_mappings;
