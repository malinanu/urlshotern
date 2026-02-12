-- Migration: Billing and Subscription Management Tables
-- Category: Billing
-- Tables: subscriptions, billing_events, usage_tracking, payment_methods, coupons, coupon_redemptions
-- Dependencies: users table

-- Note: This migration uses generate_snowflake_id() function which should exist
-- If it doesn't exist, tables will use BIGSERIAL instead

-- Subscriptions table
CREATE TABLE IF NOT EXISTS subscriptions (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('subscriptions_id_seq'::regclass) END),
    user_id BIGINT NOT NULL,
    plan_type VARCHAR(50) NOT NULL DEFAULT 'free',
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    stripe_customer_id VARCHAR(255),
    stripe_subscription_id VARCHAR(255),
    current_period_start TIMESTAMP WITH TIME ZONE,
    current_period_end TIMESTAMP WITH TIME ZONE,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    trial_end TIMESTAMP WITH TIME ZONE,
    billing_cycle VARCHAR(20) DEFAULT 'monthly',
    metadata JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

-- Create sequence for subscriptions if snowflake function doesn't exist
CREATE SEQUENCE IF NOT EXISTS subscriptions_id_seq;

-- Create indexes for subscriptions
CREATE INDEX IF NOT EXISTS idx_subscriptions_user_id ON subscriptions(user_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_stripe_customer_id ON subscriptions(stripe_customer_id);
CREATE INDEX IF NOT EXISTS idx_subscriptions_status ON subscriptions(status);
CREATE INDEX IF NOT EXISTS idx_subscriptions_plan_type ON subscriptions(plan_type);

-- Billing events table for tracking billing-related events
CREATE TABLE IF NOT EXISTS billing_events (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('billing_events_id_seq'::regclass) END),
    user_id BIGINT NOT NULL,
    subscription_id BIGINT,
    event_type VARCHAR(100) NOT NULL,
    stripe_event_id VARCHAR(255),
    amount INTEGER,
    currency VARCHAR(3) DEFAULT 'usd',
    status VARCHAR(50),
    metadata JSONB DEFAULT '{}',
    processed_at TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL
);

CREATE SEQUENCE IF NOT EXISTS billing_events_id_seq;

-- Create indexes for billing events
CREATE INDEX IF NOT EXISTS idx_billing_events_user_id ON billing_events(user_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_subscription_id ON billing_events(subscription_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_stripe_event_id ON billing_events(stripe_event_id);
CREATE INDEX IF NOT EXISTS idx_billing_events_event_type ON billing_events(event_type);
CREATE INDEX IF NOT EXISTS idx_billing_events_created_at ON billing_events(created_at);

-- Usage tracking table for monitoring user usage patterns
CREATE TABLE IF NOT EXISTS usage_tracking (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('usage_tracking_id_seq'::regclass) END),
    user_id BIGINT NOT NULL,
    tracking_period DATE NOT NULL,
    urls_created INTEGER DEFAULT 0,
    urls_active INTEGER DEFAULT 0,
    clicks_count INTEGER DEFAULT 0,
    api_calls_count INTEGER DEFAULT 0,
    custom_domains_count INTEGER DEFAULT 0,
    last_updated TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    UNIQUE(user_id, tracking_period)
);

CREATE SEQUENCE IF NOT EXISTS usage_tracking_id_seq;

-- Create indexes for usage tracking
CREATE INDEX IF NOT EXISTS idx_usage_tracking_user_id ON usage_tracking(user_id);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_period ON usage_tracking(tracking_period);
CREATE INDEX IF NOT EXISTS idx_usage_tracking_user_period ON usage_tracking(user_id, tracking_period);

-- Payment methods table
CREATE TABLE IF NOT EXISTS payment_methods (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('payment_methods_id_seq'::regclass) END),
    user_id BIGINT NOT NULL,
    stripe_payment_method_id VARCHAR(255) NOT NULL,
    type VARCHAR(50) NOT NULL,
    is_default BOOLEAN DEFAULT FALSE,
    card_brand VARCHAR(50),
    card_last4 VARCHAR(4),
    card_exp_month INTEGER,
    card_exp_year INTEGER,
    billing_address JSONB DEFAULT '{}',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE
);

CREATE SEQUENCE IF NOT EXISTS payment_methods_id_seq;

-- Create indexes for payment methods
CREATE INDEX IF NOT EXISTS idx_payment_methods_user_id ON payment_methods(user_id);
CREATE INDEX IF NOT EXISTS idx_payment_methods_stripe_pm_id ON payment_methods(stripe_payment_method_id);

-- Coupons and discounts table
CREATE TABLE IF NOT EXISTS coupons (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('coupons_id_seq'::regclass) END),
    code VARCHAR(50) UNIQUE NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    discount_type VARCHAR(20) NOT NULL,
    discount_value INTEGER NOT NULL,
    currency VARCHAR(3) DEFAULT 'usd',
    max_redemptions INTEGER,
    current_redemptions INTEGER DEFAULT 0,
    valid_from TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    valid_until TIMESTAMP WITH TIME ZONE,
    applicable_plans TEXT[],
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE SEQUENCE IF NOT EXISTS coupons_id_seq;

-- Create indexes for coupons
CREATE INDEX IF NOT EXISTS idx_coupons_code ON coupons(code);
CREATE INDEX IF NOT EXISTS idx_coupons_is_active ON coupons(is_active);
CREATE INDEX IF NOT EXISTS idx_coupons_valid_until ON coupons(valid_until);

-- Coupon redemptions table
CREATE TABLE IF NOT EXISTS coupon_redemptions (
    id BIGINT PRIMARY KEY DEFAULT (CASE WHEN EXISTS(SELECT 1 FROM pg_proc WHERE proname = 'generate_snowflake_id') THEN generate_snowflake_id() ELSE nextval('coupon_redemptions_id_seq'::regclass) END),
    coupon_id BIGINT NOT NULL,
    user_id BIGINT NOT NULL,
    subscription_id BIGINT,
    redeemed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),

    FOREIGN KEY (coupon_id) REFERENCES coupons(id) ON DELETE CASCADE,
    FOREIGN KEY (user_id) REFERENCES users(id) ON DELETE CASCADE,
    FOREIGN KEY (subscription_id) REFERENCES subscriptions(id) ON DELETE SET NULL,
    UNIQUE(coupon_id, user_id)
);

CREATE SEQUENCE IF NOT EXISTS coupon_redemptions_id_seq;

-- Create indexes for coupon redemptions
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_coupon_id ON coupon_redemptions(coupon_id);
CREATE INDEX IF NOT EXISTS idx_coupon_redemptions_user_id ON coupon_redemptions(user_id);

-- Function to update usage tracking
CREATE OR REPLACE FUNCTION update_usage_tracking(
    p_user_id BIGINT,
    p_period DATE,
    p_urls_created INTEGER DEFAULT 0,
    p_urls_active INTEGER DEFAULT 0,
    p_clicks_count INTEGER DEFAULT 0,
    p_api_calls_count INTEGER DEFAULT 0,
    p_custom_domains_count INTEGER DEFAULT 0
)
RETURNS VOID AS $$
BEGIN
    INSERT INTO usage_tracking (
        user_id,
        tracking_period,
        urls_created,
        urls_active,
        clicks_count,
        api_calls_count,
        custom_domains_count
    ) VALUES (
        p_user_id,
        p_period,
        p_urls_created,
        p_urls_active,
        p_clicks_count,
        p_api_calls_count,
        p_custom_domains_count
    )
    ON CONFLICT (user_id, tracking_period)
    DO UPDATE SET
        urls_created = usage_tracking.urls_created + p_urls_created,
        urls_active = COALESCE(p_urls_active, usage_tracking.urls_active),
        clicks_count = usage_tracking.clicks_count + p_clicks_count,
        api_calls_count = usage_tracking.api_calls_count + p_api_calls_count,
        custom_domains_count = COALESCE(p_custom_domains_count, usage_tracking.custom_domains_count),
        last_updated = NOW();
END;
$$ LANGUAGE plpgsql;

-- Function to get current month usage for a user
CREATE OR REPLACE FUNCTION get_current_month_usage(p_user_id BIGINT)
RETURNS TABLE (
    urls_created INTEGER,
    urls_active INTEGER,
    clicks_count INTEGER,
    api_calls_count INTEGER,
    custom_domains_count INTEGER
) AS $$
BEGIN
    RETURN QUERY
    SELECT
        COALESCE(ut.urls_created, 0),
        COALESCE(ut.urls_active, 0),
        COALESCE(ut.clicks_count, 0),
        COALESCE(ut.api_calls_count, 0),
        COALESCE(ut.custom_domains_count, 0)
    FROM usage_tracking ut
    WHERE ut.user_id = p_user_id
    AND ut.tracking_period = DATE_TRUNC('month', CURRENT_DATE)::DATE;

    IF NOT FOUND THEN
        RETURN QUERY SELECT 0, 0, 0, 0, 0;
    END IF;
END;
$$ LANGUAGE plpgsql;

-- Trigger to automatically create free subscription for new users
CREATE OR REPLACE FUNCTION create_default_subscription()
RETURNS TRIGGER AS $$
BEGIN
    INSERT INTO subscriptions (user_id, plan_type, status, billing_cycle)
    VALUES (NEW.id, 'free', 'active', 'monthly');

    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Create trigger if it doesn't exist
DROP TRIGGER IF EXISTS trigger_create_default_subscription ON users;
CREATE TRIGGER trigger_create_default_subscription
    AFTER INSERT ON users
    FOR EACH ROW
    EXECUTE FUNCTION create_default_subscription();

-- Create materialized view for billing analytics
CREATE MATERIALIZED VIEW IF NOT EXISTS billing_analytics AS
SELECT
    DATE_TRUNC('month', s.created_at) as month,
    s.plan_type,
    COUNT(*) as subscription_count,
    COUNT(CASE WHEN s.status = 'active' THEN 1 END) as active_subscriptions,
    COUNT(CASE WHEN s.status = 'cancelled' THEN 1 END) as cancelled_subscriptions,
    COUNT(CASE WHEN s.trial_end > NOW() THEN 1 END) as trial_subscriptions
FROM subscriptions s
GROUP BY DATE_TRUNC('month', s.created_at), s.plan_type
ORDER BY month DESC, s.plan_type;

-- Create index on the materialized view
CREATE INDEX IF NOT EXISTS idx_billing_analytics_month_plan
ON billing_analytics(month, plan_type);

-- Function to refresh billing analytics
CREATE OR REPLACE FUNCTION refresh_billing_analytics()
RETURNS VOID AS $$
BEGIN
    REFRESH MATERIALIZED VIEW billing_analytics;
END;
$$ LANGUAGE plpgsql;

-- Add billing-related columns to existing users table if not already present
DO $$
BEGIN
    -- Add stripe customer ID if not exists
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'stripe_customer_id'
    ) THEN
        ALTER TABLE users ADD COLUMN stripe_customer_id VARCHAR(255);
        CREATE INDEX IF NOT EXISTS idx_users_stripe_customer_id ON users(stripe_customer_id);
    END IF;

    -- Add subscription tier preference
    IF NOT EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'preferred_plan'
    ) THEN
        ALTER TABLE users ADD COLUMN preferred_plan VARCHAR(50) DEFAULT 'free';
    END IF;
END $$;
