-- Rollback: Remove auto-increment defaults from coupons and coupon_redemptions
ALTER TABLE coupons ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS coupons_id_seq;

ALTER TABLE coupon_redemptions ALTER COLUMN id DROP DEFAULT;
DROP SEQUENCE IF EXISTS coupon_redemptions_id_seq;
