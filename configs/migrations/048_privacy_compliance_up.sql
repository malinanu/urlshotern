-- Migration 048: Privacy compliance (UP)
-- COPPA age gate: date of birth and age flags on users. Existing users keep NULLs.
-- CAN-SPAM: global email suppression list honoured by marketing/campaign sends.

ALTER TABLE users ADD COLUMN IF NOT EXISTS date_of_birth   DATE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS is_minor        BOOLEAN NOT NULL DEFAULT FALSE;
ALTER TABLE users ADD COLUMN IF NOT EXISTS age_verified_at TIMESTAMP WITH TIME ZONE;

CREATE TABLE IF NOT EXISTS email_suppressions (
    email      VARCHAR(320) PRIMARY KEY,  -- stored lower-cased
    reason     VARCHAR(50)  NOT NULL DEFAULT 'unsubscribe',
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
