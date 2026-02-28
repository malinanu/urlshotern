-- Migration: Remove polar_discount_id from coupons

DROP INDEX IF EXISTS idx_coupons_polar_discount_id;
ALTER TABLE coupons DROP COLUMN IF EXISTS polar_discount_id;
