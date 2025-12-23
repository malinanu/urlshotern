-- Users table
CREATE TABLE IF NOT EXISTS users (
    id BIGINT PRIMARY KEY DEFAULT generate_snowflake_id(),
    name VARCHAR(100) NOT NULL,
    email VARCHAR(255) UNIQUE NOT NULL,
    password_hash VARCHAR(255), -- NULL for OAuth-only users
    phone VARCHAR(20),
    phone_verified BOOLEAN DEFAULT FALSE,
    email_verified BOOLEAN DEFAULT FALSE,
    provider VARCHAR(50) DEFAULT 'email', -- 'email', 'google'
    provider_id VARCHAR(255), -- Google ID if OAuth
    avatar_url TEXT,
    account_type VARCHAR(20) DEFAULT 'free', -- 'free', 'premium', 'enterprise'
    is_active BOOLEAN DEFAULT TRUE,
    is_admin BOOLEAN DEFAULT FALSE,
    last_login_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- User sessions
CREATE TABLE IF NOT EXISTS user_sessions (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    device_info JSONB,
    ip_address INET,
    user_agent TEXT,
    refresh_token_hash VARCHAR(255),
    expires_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Email verification tokens
CREATE TABLE IF NOT EXISTS email_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Phone verification (OTP)
CREATE TABLE IF NOT EXISTS phone_verifications (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    phone VARCHAR(20) NOT NULL,
    otp_code VARCHAR(6) NOT NULL,
    attempts INTEGER DEFAULT 0,
    expires_at TIMESTAMP WITH TIME ZONE,
    verified_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Password reset tokens
CREATE TABLE IF NOT EXISTS password_resets (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    token VARCHAR(255) UNIQUE NOT NULL,
    expires_at TIMESTAMP WITH TIME ZONE,
    used_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Update URL mappings to include user ownership
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS user_id BIGINT;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS is_public BOOLEAN DEFAULT TRUE;
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS custom_alias VARCHAR(50);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS title VARCHAR(200);
ALTER TABLE url_mappings ADD COLUMN IF NOT EXISTS description TEXT;

-- Add foreign key constraint if not exists
DO $$
BEGIN
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.table_constraints 
        WHERE constraint_name = 'url_mappings_user_id_fkey'
    ) THEN
        ALTER TABLE url_mappings ADD CONSTRAINT url_mappings_user_id_fkey 
        FOREIGN KEY (user_id) REFERENCES users(id);
    END IF;
END $$;

-- API keys for programmatic access
CREATE TABLE IF NOT EXISTS api_keys (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    user_id BIGINT NOT NULL,
    key_hash VARCHAR(255) NOT NULL UNIQUE,
    key_prefix VARCHAR(100) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    rate_limit INTEGER DEFAULT 1000,
    last_used_at TIMESTAMP WITH TIME ZONE,
    last_used_ip INET,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- API Key Usage table for tracking and analytics
CREATE TABLE IF NOT EXISTS api_key_usage (
    id UUID PRIMARY KEY DEFAULT gen_random_uuid(),
    api_key_id UUID NOT NULL,
    endpoint VARCHAR(255) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER NOT NULL,
    response_time_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    request_size INTEGER,
    response_size INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (api_key_id) REFERENCES api_keys(id) ON DELETE CASCADE
);

-- User preferences
CREATE TABLE IF NOT EXISTS user_preferences (
    user_id BIGINT PRIMARY KEY,
    default_expiration INTEGER, -- days
    analytics_public BOOLEAN DEFAULT FALSE,
    email_notifications BOOLEAN DEFAULT TRUE,
    marketing_emails BOOLEAN DEFAULT FALSE,
    timezone VARCHAR(50) DEFAULT 'UTC',
    theme VARCHAR(20) DEFAULT 'light',
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Audit logs
CREATE TABLE IF NOT EXISTS audit_logs (
    id BIGINT PRIMARY KEY DEFAULT generate_snowflake_id(),
    user_id BIGINT,
    action VARCHAR(100) NOT NULL,
    resource_type VARCHAR(50),
    resource_id VARCHAR(100),
    details JSONB,
    ip_address INET,
    user_agent TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE SET NULL
);

-- Indexes for performance
CREATE INDEX IF NOT EXISTS idx_users_email ON users(email);
CREATE INDEX IF NOT EXISTS idx_users_provider ON users(provider, provider_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_user_id ON user_sessions(user_id);
CREATE INDEX IF NOT EXISTS idx_user_sessions_expires_at ON user_sessions(expires_at);
CREATE INDEX IF NOT EXISTS idx_email_verifications_token ON email_verifications(token);
CREATE INDEX IF NOT EXISTS idx_email_verifications_user_id ON email_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_phone_verifications_user_id ON phone_verifications(user_id);
CREATE INDEX IF NOT EXISTS idx_password_resets_token ON password_resets(token);
CREATE INDEX IF NOT EXISTS idx_url_mappings_user_id ON url_mappings(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_is_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);
CREATE INDEX IF NOT EXISTS idx_api_keys_created_at ON api_keys(created_at);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_api_key_id ON api_key_usage(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_created_at ON api_key_usage(created_at);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_endpoint ON api_key_usage(endpoint);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_status_code ON api_key_usage(status_code);
CREATE INDEX IF NOT EXISTS idx_audit_logs_user_id ON audit_logs(user_id);
CREATE INDEX IF NOT EXISTS idx_audit_logs_created_at ON audit_logs(created_at);