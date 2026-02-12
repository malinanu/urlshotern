-- Migration: Authentication Extended Tables
-- Category: Authentication
-- Tables: user_preferences, email_verifications, phone_verifications, password_resets
-- Dependencies: users table

-- User preferences table for storing user settings
-- Already defined in auth_schema.sql but not created by initTables()
-- Keeping this migration empty as these tables don't exist in the source schemas
-- The plan document referenced lines 196-246 but those are ALTER TABLE statements for url_mappings

-- Note: After reviewing database.sql, these tables (user_preferences, email_verifications,
-- phone_verifications, password_resets) are NOT actually defined in the source schema.
-- They were listed in the plan but don't exist in configs/database.sql

-- This migration is intentionally empty as those tables are not in the source schema
