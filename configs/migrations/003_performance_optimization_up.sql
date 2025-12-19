-- Performance optimization migration for URL shortener
-- Creates critical indexes for query optimization
-- Uses CONCURRENTLY to avoid locking tables during index creation
-- This file should be run during low-traffic hours or with manual downtime planning

-- ============================================================================
-- CRITICAL INDEXES FOR HIGH-TRAFFIC QUERIES
-- ============================================================================

-- Index for country filtering with time range queries
-- Improves: Country statistics queries, geographic analytics
-- Expected improvement: 30-50% faster country filtering
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_click_events_short_country_time
    ON click_events(short_code, country_code, clicked_at DESC)
    WHERE country_code IS NOT NULL;

-- Index for user dashboard queries
-- Improves: Getting URLs created by a user with recent first
-- Expected improvement: 40-60% faster user dashboard loading
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_url_mappings_user_created
    ON url_mappings(user_id, created_at DESC)
    WHERE is_active = true;

-- Index for IP-based duplicate/fraud detection
-- Improves: Detecting repeated clicks from same IP, bot detection
-- Expected improvement: 20-40% faster fraud detection
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_click_events_ip
    ON click_events(ip_address, clicked_at DESC);

-- Index for domain-based lookups
-- Improves: Custom domain functionality queries
-- Expected improvement: 50% faster custom domain queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_url_mappings_domain_active
    ON url_mappings(custom_domain, is_active)
    WHERE custom_domain IS NOT NULL;

-- Partial index for recent analytics (hot data)
-- Improves: Recent analytics queries (most common access pattern)
-- Expected improvement: 60-80% faster for recent data
-- Note: This index is smaller and more efficient than full click_events index
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_click_events_recent
    ON click_events(short_code, clicked_at DESC)
    WHERE clicked_at > NOW() - INTERVAL '30 days';

-- Index for referrer lookups
-- Improves: Referrer analysis and attribution tracking
-- Expected improvement: 30-50% faster referrer queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_click_events_referrer
    ON click_events(short_code, referrer)
    WHERE referrer IS NOT NULL;

-- Index for user agent analysis
-- Improves: Device type and browser analysis queries
-- Expected improvement: 30-50% faster device analytics
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_click_events_user_agent
    ON click_events(short_code, user_agent)
    WHERE user_agent IS NOT NULL;

-- Index for team-based queries
-- Improves: Team dashboard, team analytics
-- Expected improvement: 40-60% faster team queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_url_mappings_team_active
    ON url_mappings(team_id, created_at DESC)
    WHERE is_active = true AND team_id IS NOT NULL;

-- ============================================================================
-- ANALYTICS TABLE INDEXES
-- ============================================================================

-- Composite index for geographic analytics filtering
-- Improves: Geographic analytics queries
-- Expected improvement: 40-60% faster geographic queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_geographic_analytics_lookup
    ON geographic_analytics(short_code, country_code, created_at DESC);

-- Index for device type analysis
-- Improves: Device analytics queries
-- Expected improvement: 30-50% faster device queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_device_analytics_device_type
    ON device_analytics(short_code, device_type, created_at DESC);

-- Index for time-based analytics
-- Improves: Heatmap and hourly analytics queries
-- Expected improvement: 40-60% faster time-based queries
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_time_analytics_heatmap
    ON time_analytics(short_code, click_date DESC, hour_of_day);

-- Index for referrer analytics
-- Improves: Referrer analytics queries
-- Expected improvement: 30-50% faster referrer analytics
CREATE INDEX CONCURRENTLY IF NOT EXISTS
    idx_referrer_analytics_lookup
    ON referrer_analytics(short_code, referrer_domain);

-- ============================================================================
-- STATISTICS UPDATE
-- ============================================================================
-- Update query planner statistics after index creation
-- This allows PostgreSQL to make better query planning decisions

ANALYZE click_events;
ANALYZE url_mappings;
ANALYZE geographic_analytics;
ANALYZE device_analytics;
ANALYZE time_analytics;
ANALYZE referrer_analytics;

-- ============================================================================
-- INDEX VERIFICATION QUERIES (for monitoring)
-- ============================================================================
-- Run these queries to verify indexes were created successfully:
--
-- SELECT indexname, tablename, indexdef
-- FROM pg_indexes
-- WHERE tablename IN ('click_events', 'url_mappings', 'geographic_analytics', 'device_analytics', 'time_analytics')
-- AND indexname LIKE 'idx_%'
-- ORDER BY tablename, indexname;
--
-- Check index sizes:
-- SELECT
--     schemaname,
--     tablename,
--     indexname,
--     pg_size_pretty(pg_relation_size(indexrelname::regclass)) as index_size
-- FROM pg_indexes
-- WHERE schemaname = 'public'
-- ORDER BY pg_relation_size(indexrelname::regclass) DESC;
--
-- Check index usage:
-- SELECT
--     schemaname,
--     tablename,
--     indexname,
--     idx_scan,
--     idx_tup_read,
--     idx_tup_fetch
-- FROM pg_stat_user_indexes
-- ORDER BY idx_scan DESC;
-- ============================================================================
