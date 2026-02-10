-- Rollback: Advanced Analytics Tables

-- Drop functions
DROP FUNCTION IF EXISTS cleanup_old_analytics_data(INTEGER);
DROP FUNCTION IF EXISTS refresh_daily_analytics_summary();

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS daily_analytics_summary CASCADE;

-- Drop tables
DROP TABLE IF EXISTS analytics_aggregation_state CASCADE;
DROP TABLE IF EXISTS referrer_summary CASCADE;
DROP TABLE IF EXISTS device_type_summary CASCADE;
DROP TABLE IF EXISTS country_analytics_summary CASCADE;
DROP TABLE IF EXISTS monthly_analytics CASCADE;
DROP TABLE IF EXISTS weekly_analytics CASCADE;
DROP TABLE IF EXISTS daily_analytics CASCADE;
DROP TABLE IF EXISTS hourly_analytics CASCADE;
DROP TABLE IF EXISTS realtime_subscriptions CASCADE;
DROP TABLE IF EXISTS attribution_model_performance CASCADE;
DROP TABLE IF EXISTS touchpoint_attributions CASCADE;
