-- Create database (this may already exist from Docker environment variables)
-- CREATE DATABASE urlshortener_db;

-- Create user (this may already exist from Docker environment variables)
-- CREATE USER urlshortener WITH PASSWORD 'password';
-- GRANT ALL PRIVILEGES ON DATABASE urlshortener_db TO urlshortener;

-- Connect to database
\c urlshortener_db;

-- Create tables
CREATE TABLE IF NOT EXISTS url_mappings (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(10) UNIQUE NOT NULL,
    original_url TEXT NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    expires_at TIMESTAMP WITH TIME ZONE,
    click_count BIGINT DEFAULT 0,
    is_active BOOLEAN DEFAULT TRUE,
    created_by_ip INET
);

-- Create indexes for url_mappings
CREATE INDEX IF NOT EXISTS idx_short_code ON url_mappings(short_code);
CREATE INDEX IF NOT EXISTS idx_created_at ON url_mappings(created_at);
CREATE INDEX IF NOT EXISTS idx_active ON url_mappings(is_active);
CREATE INDEX IF NOT EXISTS idx_expires_at ON url_mappings(expires_at);

-- Create click events table
CREATE TABLE IF NOT EXISTS click_events (
    id BIGINT PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    clicked_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country_code CHAR(2)
);

-- Create indexes for click_events
CREATE INDEX IF NOT EXISTS idx_short_code_time ON click_events(short_code, clicked_at);
CREATE INDEX IF NOT EXISTS idx_clicked_at ON click_events(clicked_at);
CREATE INDEX IF NOT EXISTS idx_country_code ON click_events(country_code);

-- Add foreign key constraint
ALTER TABLE click_events 
ADD CONSTRAINT fk_click_events_short_code 
FOREIGN KEY (short_code) 
REFERENCES url_mappings(short_code) 
ON DELETE CASCADE;