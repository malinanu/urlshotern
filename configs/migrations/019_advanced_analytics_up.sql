-- Migration: Advanced Analytics Tables
-- Category: Analytics
-- Tables: touchpoint_attributions, attribution_model_performance, realtime_subscriptions, hourly/daily/weekly/monthly aggregations
-- Dependencies: attribution_touchpoints, conversion_goals, conversions tables

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

-- Hourly analytics aggregation table
CREATE TABLE IF NOT EXISTS hourly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    hour_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    unique_countries INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, hour_timestamp)
);

CREATE INDEX IF NOT EXISTS idx_hourly_analytics_short_code ON hourly_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_hourly_analytics_timestamp ON hourly_analytics(hour_timestamp);

-- Daily analytics aggregation table
CREATE TABLE IF NOT EXISTS daily_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    date DATE NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    unique_countries INTEGER DEFAULT 0,
    top_country VARCHAR(2),
    top_referrer VARCHAR(255),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, date)
);

CREATE INDEX IF NOT EXISTS idx_daily_analytics_short_code ON daily_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_daily_analytics_date ON daily_analytics(date);

-- Weekly analytics aggregation table
CREATE TABLE IF NOT EXISTS weekly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    unique_countries INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, week_start)
);

CREATE INDEX IF NOT EXISTS idx_weekly_analytics_short_code ON weekly_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_weekly_analytics_week_start ON weekly_analytics(week_start);

-- Monthly analytics aggregation table
CREATE TABLE IF NOT EXISTS monthly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    month_start DATE NOT NULL,
    month_end DATE NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    unique_countries INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, month_start)
);

CREATE INDEX IF NOT EXISTS idx_monthly_analytics_short_code ON monthly_analytics(short_code);
CREATE INDEX IF NOT EXISTS idx_monthly_analytics_month_start ON monthly_analytics(month_start);

-- Country analytics summary table
CREATE TABLE IF NOT EXISTS country_analytics_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    country_code VARCHAR(2) NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    last_click TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, country_code)
);

CREATE INDEX IF NOT EXISTS idx_country_analytics_summary_short_code ON country_analytics_summary(short_code);
CREATE INDEX IF NOT EXISTS idx_country_analytics_summary_country ON country_analytics_summary(country_code);

-- Device type summary table
CREATE TABLE IF NOT EXISTS device_type_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    device_type VARCHAR(50) NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, device_type)
);

CREATE INDEX IF NOT EXISTS idx_device_type_summary_short_code ON device_type_summary(short_code);
CREATE INDEX IF NOT EXISTS idx_device_type_summary_device_type ON device_type_summary(device_type);

-- Referrer summary table
CREATE TABLE IF NOT EXISTS referrer_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(50) NOT NULL,
    referrer_domain VARCHAR(255) NOT NULL,
    total_clicks INTEGER DEFAULT 0,
    unique_visitors INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, referrer_domain)
);

CREATE INDEX IF NOT EXISTS idx_referrer_summary_short_code ON referrer_summary(short_code);
CREATE INDEX IF NOT EXISTS idx_referrer_summary_referrer ON referrer_summary(referrer_domain);

-- Analytics aggregation state table (tracks aggregation progress)
CREATE TABLE IF NOT EXISTS analytics_aggregation_state (
    id BIGSERIAL PRIMARY KEY,
    aggregation_type VARCHAR(50) NOT NULL,
    last_processed_timestamp TIMESTAMP WITH TIME ZONE NOT NULL,
    status VARCHAR(20) DEFAULT 'idle',
    error_message TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(aggregation_type)
);

-- Create materialized view for daily analytics summary (from enhanced_analytics_schema.sql)
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
    DELETE FROM hourly_analytics WHERE hour_timestamp < NOW() - INTERVAL '1 day' * retention_days;
    DELETE FROM daily_analytics WHERE date < CURRENT_DATE - retention_days;
    DELETE FROM weekly_analytics WHERE week_start < CURRENT_DATE - retention_days;
    DELETE FROM monthly_analytics WHERE month_start < CURRENT_DATE - retention_days;
END;
$$ LANGUAGE plpgsql;
