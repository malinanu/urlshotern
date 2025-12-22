-- Analytics aggregation tables for improved performance
-- Stores pre-aggregated analytics data for faster dashboard queries
-- Reduces need to scan millions of click_events rows

-- ============================================================================
-- HOURLY ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Fast access to hourly click data for real-time dashboards
-- Aggregation: Hourly click counts, unique visitors, top sources
-- Update frequency: Every hour by background job
-- Data retention: Last 12 months (rolling window)

CREATE TABLE IF NOT EXISTS hourly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    hour TIMESTAMP WITH TIME ZONE NOT NULL,
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    top_country CHAR(2),
    top_referrer TEXT,
    top_device_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, hour),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

-- Create indexes for hourly analytics queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_hourly_analytics_short_hour
    ON hourly_analytics(short_code, hour DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_hourly_analytics_created
    ON hourly_analytics(created_at DESC);

-- ============================================================================
-- DAILY ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Fast access to daily click data for trend analysis
-- Aggregation: Daily click counts, unique visitors, growth metrics, top sources
-- Update frequency: Once per day (at midnight UTC)
-- Data retention: Unlimited (suitable for long-term trend analysis)

CREATE TABLE IF NOT EXISTS daily_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    date DATE NOT NULL,
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    click_growth DECIMAL(5,2),  -- Percentage change from previous day (-99.99 to +99.99)
    top_country CHAR(2),
    top_country_clicks BIGINT,
    top_referrer TEXT,
    top_referrer_clicks BIGINT,
    top_device_type VARCHAR(50),
    top_device_type_clicks BIGINT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, date),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

-- Create indexes for daily analytics queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_daily_analytics_short_date
    ON daily_analytics(short_code, date DESC);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_daily_analytics_date
    ON daily_analytics(date DESC);

-- ============================================================================
-- WEEKLY ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Quick access to weekly aggregations for trend analysis
-- Aggregation: Weekly click counts, unique visitors, top sources
-- Update frequency: Once per week (Mondays)
-- Data retention: Unlimited

CREATE TABLE IF NOT EXISTS weekly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    week_start DATE NOT NULL,
    week_end DATE NOT NULL,
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    week_growth DECIMAL(5,2),  -- Percentage change from previous week
    top_country CHAR(2),
    top_referrer TEXT,
    top_device_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, week_start),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_weekly_analytics_short_week
    ON weekly_analytics(short_code, week_start DESC);

-- ============================================================================
-- MONTHLY ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Long-term trend analysis and reporting
-- Aggregation: Monthly click counts, unique visitors
-- Update frequency: Once per month
-- Data retention: Unlimited

CREATE TABLE IF NOT EXISTS monthly_analytics (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    year_month DATE NOT NULL,  -- First day of month
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    month_growth DECIMAL(5,2),  -- Percentage change from previous month
    top_country CHAR(2),
    top_referrer TEXT,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, year_month),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_monthly_analytics_short_month
    ON monthly_analytics(short_code, year_month DESC);

-- ============================================================================
-- COUNTRY ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Fast access to geographic distribution data
-- Stores: Click counts per country per short_code
-- Update frequency: Real-time (or batched hourly)
-- Data retention: Last 12 months

CREATE TABLE IF NOT EXISTS country_analytics_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    country_code CHAR(2) NOT NULL,
    country_name VARCHAR(100),
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, country_code),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_country_analytics_short
    ON country_analytics_summary(short_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_country_analytics_country
    ON country_analytics_summary(country_code);

-- ============================================================================
-- DEVICE TYPE ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Track device distribution over time
-- Stores: Click counts per device type per short_code
-- Update frequency: Real-time (or batched hourly)

CREATE TABLE IF NOT EXISTS device_type_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    device_type VARCHAR(50) NOT NULL,  -- mobile, desktop, tablet
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, device_type),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_device_type_summary_short
    ON device_type_summary(short_code);

-- ============================================================================
-- REFERRER ANALYTICS SUMMARY TABLE
-- ============================================================================
-- Purpose: Track top referrers over time
-- Stores: Click counts per referrer domain per short_code
-- Update frequency: Real-time (or batched hourly)
-- Note: Limited to top 100 referrers per short_code

CREATE TABLE IF NOT EXISTS referrer_summary (
    id BIGSERIAL PRIMARY KEY,
    short_code VARCHAR(10) NOT NULL,
    referrer_domain VARCHAR(255),
    referrer_type VARCHAR(50),  -- social, search, direct, email, referral, etc.
    click_count BIGINT DEFAULT 0,
    unique_visitors BIGINT DEFAULT 0,
    last_update TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    UNIQUE(short_code, referrer_domain),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_referrer_summary_short
    ON referrer_summary(short_code);

CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_referrer_summary_domain
    ON referrer_summary(referrer_domain);

-- ============================================================================
-- ANALYTICS MATERIALIZATION TRACKING TABLE
-- ============================================================================
-- Purpose: Track last aggregation time for each URL to prevent redundant work
-- Stores: Timestamp of last hourly/daily/weekly/monthly aggregation

CREATE TABLE IF NOT EXISTS analytics_aggregation_state (
    short_code VARCHAR(10) PRIMARY KEY,
    last_hourly_aggregation TIMESTAMP WITH TIME ZONE,
    last_daily_aggregation TIMESTAMP WITH TIME ZONE,
    last_weekly_aggregation TIMESTAMP WITH TIME ZONE,
    last_monthly_aggregation TIMESTAMP WITH TIME ZONE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    FOREIGN KEY (short_code) REFERENCES url_mappings(short_code) ON DELETE CASCADE
);

-- ============================================================================
-- PERFORMANCE STATISTICS AND ANALYSIS
-- ============================================================================

-- View for comparing raw vs aggregated data sizes
CREATE OR REPLACE VIEW analytics_storage_stats AS
SELECT
    'click_events' as table_name,
    count(*) as row_count,
    pg_size_pretty(pg_total_relation_size('click_events'::regclass)) as size
FROM click_events
UNION ALL
SELECT
    'hourly_analytics',
    count(*),
    pg_size_pretty(pg_total_relation_size('hourly_analytics'::regclass))
FROM hourly_analytics
UNION ALL
SELECT
    'daily_analytics',
    count(*),
    pg_size_pretty(pg_total_relation_size('daily_analytics'::regclass))
FROM daily_analytics
UNION ALL
SELECT
    'weekly_analytics',
    count(*),
    pg_size_pretty(pg_total_relation_size('weekly_analytics'::regclass))
FROM weekly_analytics
UNION ALL
SELECT
    'monthly_analytics',
    count(*),
    pg_size_pretty(pg_total_relation_size('monthly_analytics'::regclass))
FROM monthly_analytics;

-- ============================================================================
-- AGGREGATION PROCEDURE TEMPLATE
-- ============================================================================
-- Run these queries periodically (via cron job or background task)
-- to keep aggregated tables up to date:

/*
-- Hourly aggregation (run every hour)
INSERT INTO hourly_analytics (short_code, hour, click_count, unique_visitors, top_country, top_referrer)
SELECT
    ce.short_code,
    DATE_TRUNC('hour', ce.clicked_at) as hour,
    COUNT(*) as click_count,
    COUNT(DISTINCT ce.ip_address) as unique_visitors,
    (array_agg(ce.country_code ORDER BY COUNT(*) DESC LIMIT 1))[1] as top_country,
    (array_agg(ce.referrer ORDER BY COUNT(*) DESC LIMIT 1))[1] as top_referrer
FROM click_events ce
WHERE ce.clicked_at >= DATE_TRUNC('hour', NOW()) - INTERVAL '1 hour'
AND ce.clicked_at < DATE_TRUNC('hour', NOW())
GROUP BY ce.short_code, DATE_TRUNC('hour', ce.clicked_at)
ON CONFLICT (short_code, hour) DO UPDATE
SET
    click_count = EXCLUDED.click_count,
    unique_visitors = EXCLUDED.unique_visitors,
    top_country = EXCLUDED.top_country,
    top_referrer = EXCLUDED.top_referrer,
    updated_at = NOW();

-- Daily aggregation (run once per day at midnight)
INSERT INTO daily_analytics (short_code, date, click_count, unique_visitors, top_country, top_referrer)
SELECT
    ce.short_code,
    DATE(ce.clicked_at) as date,
    COUNT(*) as click_count,
    COUNT(DISTINCT ce.ip_address) as unique_visitors,
    (array_agg(ce.country_code ORDER BY COUNT(*) DESC LIMIT 1))[1] as top_country,
    (array_agg(ce.referrer ORDER BY COUNT(*) DESC LIMIT 1))[1] as top_referrer
FROM click_events ce
WHERE DATE(ce.clicked_at) = CURRENT_DATE - INTERVAL '1 day'
GROUP BY ce.short_code, DATE(ce.clicked_at)
ON CONFLICT (short_code, date) DO UPDATE
SET
    click_count = EXCLUDED.click_count,
    unique_visitors = EXCLUDED.unique_visitors,
    top_country = EXCLUDED.top_country,
    top_referrer = EXCLUDED.top_referrer,
    updated_at = NOW();
*/

-- ============================================================================
-- ANALYZE updated tables
-- ============================================================================
ANALYZE hourly_analytics;
ANALYZE daily_analytics;
ANALYZE weekly_analytics;
ANALYZE monthly_analytics;
ANALYZE country_analytics_summary;
ANALYZE device_type_summary;
ANALYZE referrer_summary;
ANALYZE analytics_aggregation_state;
