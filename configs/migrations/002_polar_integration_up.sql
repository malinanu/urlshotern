-- Migration: Polar Payment Gateway Integration
-- Description: Creates tables for Polar subscriptions, products, and webhook events
-- Author: Claude Code
-- Date: 2025-01-19

-- ============================================
-- Polar Subscriptions Table
-- ============================================
CREATE TABLE IF NOT EXISTS polar_subscriptions (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    polar_customer_id VARCHAR(255) NOT NULL,
    polar_subscription_id VARCHAR(255),
    polar_product_id VARCHAR(255) NOT NULL,
    polar_price_id VARCHAR(255),
    polar_checkout_id VARCHAR(255),
    status VARCHAR(50) NOT NULL DEFAULT 'active',
    plan_type VARCHAR(50) NOT NULL DEFAULT 'free',
    billing_cycle VARCHAR(20),
    current_period_start TIMESTAMP,
    current_period_end TIMESTAMP,
    cancel_at TIMESTAMP,
    cancel_at_period_end BOOLEAN DEFAULT FALSE,
    cancelled_at TIMESTAMP,
    ended_at TIMESTAMP,
    trial_start TIMESTAMP,
    trial_end TIMESTAMP,
    benefits JSONB DEFAULT '[]'::jsonb,
    metadata JSONB DEFAULT '{}'::jsonb,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for polar_subscriptions
CREATE INDEX idx_polar_subscriptions_user_id ON polar_subscriptions(user_id);
CREATE INDEX idx_polar_subscriptions_status ON polar_subscriptions(status);
CREATE INDEX idx_polar_subscriptions_customer_id ON polar_subscriptions(polar_customer_id);
CREATE INDEX idx_polar_subscriptions_subscription_id ON polar_subscriptions(polar_subscription_id);
CREATE UNIQUE INDEX idx_user_polar_sub ON polar_subscriptions(user_id, polar_subscription_id) WHERE polar_subscription_id IS NOT NULL;

-- ============================================
-- Polar Products Table (Cached products from Polar)
-- ============================================
CREATE TABLE IF NOT EXISTS polar_products (
    id VARCHAR(255) PRIMARY KEY,
    polar_product_id VARCHAR(255) UNIQUE NOT NULL,
    name VARCHAR(255) NOT NULL,
    description TEXT,
    plan_type VARCHAR(50) NOT NULL,
    is_recurring BOOLEAN DEFAULT TRUE,
    prices JSONB DEFAULT '[]'::jsonb,
    features JSONB DEFAULT '[]'::jsonb,
    benefits JSONB DEFAULT '[]'::jsonb,
    organization_id VARCHAR(255) NOT NULL,
    is_archived BOOLEAN DEFAULT FALSE,
    created_at TIMESTAMP NOT NULL,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    last_synced_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for polar_products
CREATE INDEX idx_polar_products_plan_type ON polar_products(plan_type);
CREATE INDEX idx_polar_products_org_id ON polar_products(organization_id);
CREATE INDEX idx_polar_products_is_archived ON polar_products(is_archived);

-- ============================================
-- Polar Webhook Events Table
-- ============================================
CREATE TABLE IF NOT EXISTS polar_webhook_events (
    id BIGSERIAL PRIMARY KEY,
    polar_event_id VARCHAR(255) UNIQUE NOT NULL,
    event_type VARCHAR(100) NOT NULL,
    payload JSONB NOT NULL,
    processed_at TIMESTAMP,
    processing_status VARCHAR(50) DEFAULT 'pending',
    error_message TEXT,
    retry_count INTEGER DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

-- Indexes for polar_webhook_events
CREATE INDEX idx_polar_webhook_events_event_type ON polar_webhook_events(event_type);
CREATE INDEX idx_polar_webhook_events_status ON polar_webhook_events(processing_status);
CREATE INDEX idx_polar_webhook_events_processed_at ON polar_webhook_events(processed_at);
CREATE UNIQUE INDEX idx_polar_webhook_events_polar_event_id ON polar_webhook_events(polar_event_id);

-- ============================================
-- Trigger to auto-update updated_at timestamp
-- ============================================
CREATE OR REPLACE FUNCTION update_updated_at_column()
RETURNS TRIGGER AS $$
BEGIN
    NEW.updated_at = CURRENT_TIMESTAMP;
    RETURN NEW;
END;
$$ LANGUAGE plpgsql;

-- Apply trigger to polar_subscriptions
CREATE TRIGGER update_polar_subscriptions_updated_at
    BEFORE UPDATE ON polar_subscriptions
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to polar_products
CREATE TRIGGER update_polar_products_updated_at
    BEFORE UPDATE ON polar_products
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- Apply trigger to polar_webhook_events
CREATE TRIGGER update_polar_webhook_events_updated_at
    BEFORE UPDATE ON polar_webhook_events
    FOR EACH ROW
    EXECUTE FUNCTION update_updated_at_column();

-- ============================================
-- Comments for Documentation
-- ============================================
COMMENT ON TABLE polar_subscriptions IS 'Stores user subscriptions managed through Polar payment gateway';
COMMENT ON COLUMN polar_subscriptions.user_id IS 'Foreign key reference to users table';
COMMENT ON COLUMN polar_subscriptions.polar_customer_id IS 'Polar customer UUID';
COMMENT ON COLUMN polar_subscriptions.polar_subscription_id IS 'Polar subscription UUID (null for free tier)';
COMMENT ON COLUMN polar_subscriptions.status IS 'Subscription status: active, cancelled, incomplete, trialing, past_due';
COMMENT ON COLUMN polar_subscriptions.plan_type IS 'Plan tier: free, premium, enterprise';
COMMENT ON COLUMN polar_subscriptions.benefits IS 'JSON array of benefits granted to this subscription';

COMMENT ON TABLE polar_products IS 'Cached products and pricing from Polar for fast local lookups';
COMMENT ON COLUMN polar_products.plan_type IS 'Maps to internal plan types (free, premium, enterprise)';
COMMENT ON COLUMN polar_products.is_recurring IS 'Whether this product has recurring billing';

COMMENT ON TABLE polar_webhook_events IS 'Log of all webhook events received from Polar for idempotency and debugging';
COMMENT ON COLUMN polar_webhook_events.polar_event_id IS 'Unique event ID from Polar for idempotency';
COMMENT ON COLUMN polar_webhook_events.processing_status IS 'Status: pending, processed, failed';

-- ============================================
-- Sample Data (Optional - for development)
-- ============================================
-- Insert a sample free product
INSERT INTO polar_products (id, polar_product_id, name, description, plan_type, is_recurring, prices, benefits, organization_id, created_at)
VALUES (
    'free_plan',
    'free_plan',
    'Free Plan',
    'Perfect for personal use with basic features',
    'free',
    FALSE,
    '[]'::jsonb,
    '[{"id":"free_urls","type":"custom","description":"Up to 100 URLs"},{"id":"free_clicks","type":"custom","description":"5,000 clicks per month"}]'::jsonb,
    'default',
    CURRENT_TIMESTAMP
) ON CONFLICT (id) DO NOTHING;

-- ============================================
-- Migration Rollback (for reference)
-- ============================================
-- To rollback this migration, run:
/*
DROP TRIGGER IF EXISTS update_polar_webhook_events_updated_at ON polar_webhook_events;
DROP TRIGGER IF EXISTS update_polar_products_updated_at ON polar_products;
DROP TRIGGER IF EXISTS update_polar_subscriptions_updated_at ON polar_subscriptions;
DROP TABLE IF EXISTS polar_webhook_events;
DROP TABLE IF EXISTS polar_products;
DROP TABLE IF EXISTS polar_subscriptions;
*/
