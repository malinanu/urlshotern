-- Enhanced Analytics Schema for URL Shortener
-- This extends the existing database with advanced analytics tables

-- Geographic Analytics Table
CREATE TABLE IF NOT EXISTS geographic_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    country_name VARCHAR(100) NOT NULL,
    region VARCHAR(100),
    city VARCHAR(100),
    latitude DECIMAL(10, 8),
    longitude DECIMAL(11, 8),
    clicks BIGINT DEFAULT 0,
    unique_ips BIGINT DEFAULT 0,
    last_click TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, country_code, region, city)
);

-- Indexes for geographic analytics
CREATE INDEX IF NOT EXISTS idx_geographic_analytics_short_code ON geographic_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_geographic_analytics_country ON geographic_analytics(country_code);
CREATE INDEX IF NOT EXISTS idx_geographic_analytics_last_click ON geographic_analytics(last_click);

-- Time-based Analytics Table
CREATE TABLE IF NOT EXISTS time_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    click_date DATE NOT NULL,
    hour_of_day INTEGER NOT NULL CHECK (hour_of_day >= 0 AND hour_of_day <= 23),
    day_of_week INTEGER NOT NULL CHECK (day_of_week >= 0 AND day_of_week <= 6),
    day_of_month INTEGER NOT NULL CHECK (day_of_month >= 1 AND day_of_month <= 31),
    month INTEGER NOT NULL CHECK (month >= 1 AND month <= 12),
    year INTEGER NOT NULL,
    clicks BIGINT DEFAULT 0,
    unique_ips BIGINT DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, click_date, hour_of_day)
);

-- Indexes for time analytics
CREATE INDEX IF NOT EXISTS idx_time_analytics_short_code ON time_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_time_analytics_date ON time_analytics(click_date);
CREATE INDEX IF NOT EXISTS idx_time_analytics_hour ON time_analytics(hour_of_day);

-- Device Analytics Table
CREATE TABLE IF NOT EXISTS device_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    device_type VARCHAR(50) NOT NULL, -- mobile, tablet, desktop
    device_brand VARCHAR(100),
    device_model VARCHAR(200),
    os_name VARCHAR(100),
    os_version VARCHAR(50),
    browser_name VARCHAR(100),
    browser_version VARCHAR(50),
    screen_resolution VARCHAR(20),
    user_agent_hash VARCHAR(64) NOT NULL, -- SHA-256 hash for uniqueness
    clicks BIGINT DEFAULT 0,
    unique_ips BIGINT DEFAULT 0,
    last_click TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, user_agent_hash)
);

-- Indexes for device analytics
CREATE INDEX IF NOT EXISTS idx_device_analytics_short_code ON device_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_device_analytics_type ON device_analytics(device_type);
CREATE INDEX IF NOT EXISTS idx_device_analytics_browser ON device_analytics(browser_name);
CREATE INDEX IF NOT EXISTS idx_device_analytics_os ON device_analytics(os_name);

-- Enhanced Referrer Analytics Table
CREATE TABLE IF NOT EXISTS referrer_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    referrer_domain VARCHAR(255),
    referrer_url TEXT,
    referrer_type VARCHAR(50) NOT NULL, -- direct, search, social, email, campaign, referral
    campaign_source VARCHAR(255),
    campaign_medium VARCHAR(255),
    campaign_name VARCHAR(255),
    campaign_term VARCHAR(255),
    campaign_content VARCHAR(255),
    clicks BIGINT DEFAULT 0,
    unique_clicks BIGINT DEFAULT 0,
    conversions BIGINT DEFAULT 0,
    last_click TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, referrer_domain, campaign_source, campaign_medium)
);

-- Indexes for referrer analytics
CREATE INDEX IF NOT EXISTS idx_referrer_analytics_short_code ON referrer_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_referrer_analytics_type ON referrer_analytics(referrer_type);
CREATE INDEX IF NOT EXISTS idx_referrer_analytics_domain ON referrer_analytics(referrer_domain);
CREATE INDEX IF NOT EXISTS idx_referrer_analytics_campaign ON referrer_analytics(campaign_source, campaign_medium);

-- Attribution Touchpoints Table
CREATE TABLE IF NOT EXISTS attribution_touchpoints (
    id BIGSERIAL PRIMARY KEY,
    session_id VARCHAR(64) NOT NULL,
    short_code VARCHAR(50) NOT NULL,
    user_ip VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    campaign_source VARCHAR(255),
    campaign_medium VARCHAR(255),
    campaign_name VARCHAR(255),
    touchpoint_order INTEGER NOT NULL,
    touchpoint_time TIMESTAMP WITH TIME ZONE NOT NULL,
    conversion_id VARCHAR(64),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for attribution touchpoints
CREATE INDEX IF NOT EXISTS idx_attribution_touchpoints_session ON attribution_touchpoints(session_id);
CREATE INDEX IF NOT EXISTS idx_attribution_touchpoints_short_code ON attribution_touchpoints(short_code);
CREATE INDEX IF NOT EXISTS idx_attribution_touchpoints_conversion ON attribution_touchpoints(conversion_id);
CREATE INDEX IF NOT EXISTS idx_attribution_touchpoints_time ON attribution_touchpoints(touchpoint_time);

-- Conversion Goals Table (extend existing)
CREATE TABLE IF NOT EXISTS conversion_goals (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    goal_name VARCHAR(100) NOT NULL,
    goal_type VARCHAR(50) NOT NULL, -- url_visit, custom_event, form_submit, purchase
    target_url TEXT,
    custom_event_name VARCHAR(100),
    goal_value DECIMAL(12, 2) DEFAULT 0,
    attribution_window INTEGER DEFAULT 30, -- days
    is_active BOOLEAN DEFAULT true,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for conversion goals
CREATE INDEX IF NOT EXISTS idx_conversion_goals_user_id ON conversion_goals(user_id);
CREATE INDEX IF NOT EXISTS idx_conversion_goals_active ON conversion_goals(is_active);

-- Conversions Table (extend existing)
CREATE TABLE IF NOT EXISTS conversions (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    goal_id BIGINT NOT NULL REFERENCES conversion_goals(id),
    conversion_id VARCHAR(64) NOT NULL UNIQUE,
    conversion_type VARCHAR(50) NOT NULL,
    conversion_value DECIMAL(12, 2) DEFAULT 0,
    user_ip VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    session_id VARCHAR(64),
    click_id BIGINT,
    conversion_time TIMESTAMP WITH TIME ZONE NOT NULL,
    attribution_model VARCHAR(50) DEFAULT 'last_click',
    time_to_conversion INTEGER, -- minutes from first touch
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Indexes for conversions
CREATE INDEX IF NOT EXISTS idx_conversions_short_code ON conversions(short_code);
CREATE INDEX IF NOT EXISTS idx_conversions_goal_id ON conversions(goal_id);
CREATE INDEX IF NOT EXISTS idx_conversions_session ON conversions(session_id);
CREATE INDEX IF NOT EXISTS idx_conversions_time ON conversions(conversion_time);

-- A/B Testing Tables (extend existing)
CREATE TABLE IF NOT EXISTS ab_tests (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    test_name VARCHAR(200) NOT NULL,
    test_type VARCHAR(50) NOT NULL,
    description TEXT,
    traffic_split VARCHAR(500), -- JSON configuration
    start_date TIMESTAMP WITH TIME ZONE,
    end_date TIMESTAMP WITH TIME ZONE,
    sample_size INTEGER DEFAULT 1000,
    confidence DECIMAL(5, 2) DEFAULT 95.0,
    is_active BOOLEAN DEFAULT false,
    status VARCHAR(20) DEFAULT 'draft', -- draft, running, completed, paused
    winner VARCHAR(50),
    confidence_level DECIMAL(5, 2),
    min_sample_size INTEGER DEFAULT 100,
    conversion_goal_id BIGINT REFERENCES conversion_goals(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ab_test_variants (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant_name VARCHAR(100) NOT NULL,
    short_code VARCHAR(50) NOT NULL,
    traffic_allocation INTEGER NOT NULL, -- percentage
    is_control BOOLEAN DEFAULT false,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

CREATE TABLE IF NOT EXISTS ab_test_results (
    id BIGSERIAL PRIMARY KEY,
    test_id BIGINT NOT NULL REFERENCES ab_tests(id) ON DELETE CASCADE,
    variant_id BIGINT NOT NULL REFERENCES ab_test_variants(id) ON DELETE CASCADE,
    session_id VARCHAR(64) NOT NULL,
    event_type VARCHAR(20) NOT NULL, -- assignment, conversion
    conversion_value DECIMAL(12, 2) DEFAULT 0,
    timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    UNIQUE(test_id, session_id, event_type)
);

-- Indexes for A/B testing
CREATE INDEX IF NOT EXISTS idx_ab_tests_user_id ON ab_tests(user_id);
CREATE INDEX IF NOT EXISTS idx_ab_tests_status ON ab_tests(status);
CREATE INDEX IF NOT EXISTS idx_ab_test_variants_test_id ON ab_test_variants(test_id);
CREATE INDEX IF NOT EXISTS idx_ab_test_results_test_id ON ab_test_results(test_id);
CREATE INDEX IF NOT EXISTS idx_ab_test_results_variant_id ON ab_test_results(variant_id);

-- Real-time Analytics Cache Table (for WebSocket subscriptions)
CREATE TABLE IF NOT EXISTS realtime_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    connection_id VARCHAR(100) NOT NULL,
    user_id BIGINT,
    subscribed_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    last_ping TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, connection_id)
);

-- Index for real-time subscriptions
CREATE INDEX IF NOT EXISTS idx_realtime_subscriptions_short_code ON realtime_subscriptions(short_code);
CREATE INDEX IF NOT EXISTS idx_realtime_subscriptions_last_ping ON realtime_subscriptions(last_ping);

-- Update existing click_events table to include session tracking
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS session_id VARCHAR(64);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS device_fingerprint VARCHAR(64);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS utm_source VARCHAR(255);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS utm_medium VARCHAR(255);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS utm_campaign VARCHAR(255);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS utm_term VARCHAR(255);
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS utm_content VARCHAR(255);

-- Add indexes for new columns
CREATE INDEX IF NOT EXISTS idx_click_events_session_id ON click_events(session_id);
CREATE INDEX IF NOT EXISTS idx_click_events_utm_source ON click_events(utm_source);

-- Create materialized views for fast analytics queries
CREATE MATERIALIZED VIEW IF NOT EXISTS daily_analytics_summary AS
SELECT 
    short_code,
    DATE(clicked_at) as click_date,
    COUNT(*) as total_clicks,
    COUNT(DISTINCT ip_address) as unique_visitors,
    COUNT(DISTINCT country_code) as unique_countries
FROM click_events 
WHERE clicked_at >= CURRENT_DATE - INTERVAL '90 days'
GROUP BY short_code, DATE(clicked_at);

-- Index for materialized view
CREATE UNIQUE INDEX IF NOT EXISTS idx_daily_analytics_summary 
ON daily_analytics_summary(short_code, click_date);

-- Refresh function for materialized view
CREATE OR REPLACE FUNCTION refresh_daily_analytics_summary()
RETURNS void AS $$
BEGIN
    REFRESH MATERIALIZED VIEW CONCURRENTLY daily_analytics_summary;
END;
$$ LANGUAGE plpgsql;

-- Touchpoint Attributions Table (for storing calculated attribution values)
CREATE TABLE IF NOT EXISTS touchpoint_attributions (
    id BIGSERIAL PRIMARY KEY,
    touchpoint_id BIGINT NOT NULL REFERENCES attribution_touchpoints(id) ON DELETE CASCADE,
    attribution_model VARCHAR(50) NOT NULL,
    attribution_value DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    weight DECIMAL(8,6) NOT NULL DEFAULT 0.000000,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(touchpoint_id, attribution_model)
);

-- Create index on touchpoint_attributions for performance
CREATE INDEX IF NOT EXISTS idx_touchpoint_attributions_model_touchpoint 
ON touchpoint_attributions(attribution_model, touchpoint_id);

-- Attribution Model Performance Table (for tracking model effectiveness)
CREATE TABLE IF NOT EXISTS attribution_model_performance (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    attribution_model VARCHAR(50) NOT NULL,
    total_conversions INTEGER NOT NULL DEFAULT 0,
    total_attributed_value DECIMAL(12,2) NOT NULL DEFAULT 0.00,
    avg_attribution_accuracy DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    model_confidence DECIMAL(5,4) NOT NULL DEFAULT 0.0000,
    date_range_start DATE NOT NULL,
    date_range_end DATE NOT NULL,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, attribution_model, date_range_start, date_range_end)
);

-- Create index on attribution model performance
CREATE INDEX IF NOT EXISTS idx_attribution_model_performance_short_code_model 
ON attribution_model_performance(short_code, attribution_model);

-- Create a function to clean up old analytics data (data retention)
CREATE OR REPLACE FUNCTION cleanup_old_analytics_data(retention_days INTEGER DEFAULT 365)
RETURNS void AS $$
BEGIN
    DELETE FROM geographic_analytics WHERE last_click < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM time_analytics WHERE click_date < CURRENT_DATE - retention_days;
    DELETE FROM device_analytics WHERE last_click < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM referrer_analytics WHERE last_click < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM attribution_touchpoints WHERE touchpoint_time < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM touchpoint_attributions WHERE created_at < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM realtime_subscriptions WHERE last_ping < NOW() - INTERVAL '1 hour';
END;
$$ LANGUAGE plpgsql;