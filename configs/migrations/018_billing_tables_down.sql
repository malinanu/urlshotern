-- Rollback: Billing Tables

-- Drop trigger
DROP TRIGGER IF EXISTS trigger_create_default_subscription ON users;

-- Drop functions
DROP FUNCTION IF EXISTS create_default_subscription();
DROP FUNCTION IF EXISTS refresh_billing_analytics();
DROP FUNCTION IF EXISTS get_current_month_usage(BIGINT);
DROP FUNCTION IF EXISTS update_usage_tracking(BIGINT, DATE, INTEGER, INTEGER, INTEGER, INTEGER, INTEGER);

-- Drop materialized view
DROP MATERIALIZED VIEW IF EXISTS billing_analytics CASCADE;

-- Drop tables
DROP TABLE IF EXISTS coupon_redemptions CASCADE;
DROP TABLE IF EXISTS coupons CASCADE;
DROP TABLE IF EXISTS payment_methods CASCADE;
DROP TABLE IF EXISTS usage_tracking CASCADE;
DROP TABLE IF EXISTS billing_events CASCADE;
DROP TABLE IF EXISTS subscriptions CASCADE;

-- Drop sequences
DROP SEQUENCE IF EXISTS coupon_redemptions_id_seq;
DROP SEQUENCE IF EXISTS coupons_id_seq;
DROP SEQUENCE IF EXISTS payment_methods_id_seq;
DROP SEQUENCE IF EXISTS usage_tracking_id_seq;
DROP SEQUENCE IF EXISTS billing_events_id_seq;
DROP SEQUENCE IF EXISTS subscriptions_id_seq;

-- Remove billing columns from users table
DO $$
BEGIN
    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'stripe_customer_id'
    ) THEN
        ALTER TABLE users DROP COLUMN stripe_customer_id;
    END IF;

    IF EXISTS (
        SELECT 1 FROM information_schema.columns
        WHERE table_name = 'users' AND column_name = 'preferred_plan'
    ) THEN
        ALTER TABLE users DROP COLUMN preferred_plan;
    END IF;
END $$;
