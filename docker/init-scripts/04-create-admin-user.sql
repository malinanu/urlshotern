-- Docker Development Default Users
-- This script creates default admin and test users for local development
--
-- SECURITY WARNING: These credentials are for DEVELOPMENT ONLY!
-- DO NOT use in production environments
--
-- Default Credentials:
-- Admin User:  admin@example.com / admin123
-- Test User:   test@example.com / admin123

-- Insert default admin user for Docker development
-- Password: admin123 (bcrypt hash with cost 12)
INSERT INTO users (
    id,
    name,
    email,
    password_hash,
    email_verified,
    phone_verified,
    provider,
    account_type,
    is_admin,
    is_active,
    created_at,
    updated_at
) VALUES (
    1000000000000000001,  -- Snowflake ID for consistent testing
    'Admin User',
    'admin@example.com',
    '$2a$12$Aakb.tFxs.Fu3J8o5TDqqer53E39B2VJ3EAkpLlvrxXpbSDuMRK1W',  -- bcrypt hash of "admin123"
    true,  -- email_verified
    false, -- phone_verified
    'email',
    'super_admin',
    true,  -- is_admin
    true,  -- is_active
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Insert a regular test user for testing normal login
INSERT INTO users (
    id,
    name,
    email,
    password_hash,
    email_verified,
    phone_verified,
    provider,
    account_type,
    is_admin,
    is_active,
    created_at,
    updated_at
) VALUES (
    1000000000000000002,  -- Snowflake ID for consistent testing
    'Test User',
    'test@example.com',
    '$2a$12$Aakb.tFxs.Fu3J8o5TDqqer53E39B2VJ3EAkpLlvrxXpbSDuMRK1W',  -- bcrypt hash of "admin123"
    true,  -- email_verified
    false, -- phone_verified
    'email',
    'free',
    false, -- is_admin
    true,  -- is_active
    NOW(),
    NOW()
) ON CONFLICT (email) DO NOTHING;

-- Create user preferences for admin user
INSERT INTO user_preferences (
    user_id,
    theme,
    timezone,
    email_notifications,
    analytics_public,
    marketing_emails
) VALUES (
    1000000000000000001,
    'light',
    'UTC',
    true,
    false,
    false
) ON CONFLICT (user_id) DO NOTHING;

-- Create user preferences for test user
INSERT INTO user_preferences (
    user_id,
    theme,
    timezone,
    email_notifications,
    analytics_public,
    marketing_emails
) VALUES (
    1000000000000000002,
    'light',
    'UTC',
    true,
    false,
    false
) ON CONFLICT (user_id) DO NOTHING;

-- Log the creation
DO $$
BEGIN
    RAISE NOTICE '';
    RAISE NOTICE '==================================================';
    RAISE NOTICE 'DOCKER DEVELOPMENT USERS CREATED';
    RAISE NOTICE '==================================================';
    RAISE NOTICE 'Admin User:  admin@example.com / admin123';
    RAISE NOTICE 'Test User:   test@example.com / admin123';
    RAISE NOTICE '';
    RAISE NOTICE 'WARNING: Change these credentials immediately!';
    RAISE NOTICE 'These are for DEVELOPMENT ONLY';
    RAISE NOTICE '==================================================';
    RAISE NOTICE '';
END $$;
