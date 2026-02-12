-- Migration: Custom Domains Tables
-- Category: Custom Domains
-- Tables: custom_domains, domain_analytics
-- Dependencies: users, teams tables

-- Custom domains table
CREATE TABLE IF NOT EXISTS custom_domains (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id BIGINT REFERENCES teams(id) ON DELETE CASCADE,
    domain VARCHAR(255) UNIQUE NOT NULL,
    verification_method VARCHAR(20) DEFAULT 'dns',
    verification_token VARCHAR(255) UNIQUE NOT NULL,
    is_verified BOOLEAN DEFAULT FALSE,
    verified_at TIMESTAMP WITH TIME ZONE,
    ssl_enabled BOOLEAN DEFAULT FALSE,
    ssl_cert_path TEXT,
    ssl_key_path TEXT,
    ssl_expires_at TIMESTAMP WITH TIME ZONE,
    health_check_url TEXT,
    last_health_check TIMESTAMP WITH TIME ZONE,
    health_status VARCHAR(20) DEFAULT 'unknown',
    dns_records JSONB DEFAULT '{}',
    settings JSONB DEFAULT '{}',
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Domain analytics table
CREATE TABLE IF NOT EXISTS domain_analytics (
    id BIGSERIAL PRIMARY KEY,
    domain_id BIGINT NOT NULL REFERENCES custom_domains(id) ON DELETE CASCADE,
    total_urls INTEGER DEFAULT 0,
    total_clicks BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    date DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(domain_id, date)
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_custom_domains_user_id ON custom_domains(user_id);
CREATE INDEX IF NOT EXISTS idx_custom_domains_domain ON custom_domains(domain);
CREATE INDEX IF NOT EXISTS idx_custom_domains_verified ON custom_domains(is_verified);
