-- Migration: Add polar_discount_id to coupons

ALTER TABLE coupons ADD COLUMN IF NOT EXISTS polar_discount_id VARCHAR(255);
CREATE INDEX IF NOT EXISTS idx_coupons_polar_discount_id ON coupons(polar_discount_id);
