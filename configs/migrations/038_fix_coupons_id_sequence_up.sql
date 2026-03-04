-- Migration: Fix coupons and coupon_redemptions id columns to use auto-increment sequences
-- Problem: Tables were created with `id BIGINT PRIMARY KEY` (no default).
-- GORM's autoIncrement tag omits id from INSERT, expecting the DB to provide it — causing NOT NULL violation.

-- Fix coupons.id
CREATE SEQUENCE IF NOT EXISTS coupons_id_seq;
SELECT setval('coupons_id_seq', COALESCE((SELECT MAX(id) FROM coupons), 0) + 1, false);
ALTER TABLE coupons ALTER COLUMN id SET DEFAULT nextval('coupons_id_seq');

-- Fix coupon_redemptions.id (same schema pattern, same bug)
CREATE SEQUENCE IF NOT EXISTS coupon_redemptions_id_seq;
SELECT setval('coupon_redemptions_id_seq', COALESCE((SELECT MAX(id) FROM coupon_redemptions), 0) + 1, false);
ALTER TABLE coupon_redemptions ALTER COLUMN id SET DEFAULT nextval('coupon_redemptions_id_seq');
