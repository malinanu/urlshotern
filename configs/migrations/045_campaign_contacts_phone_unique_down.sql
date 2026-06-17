-- Migration 045: Deduplicate phone-only campaign contacts (DOWN)
-- The merge of duplicate rows is not reversible; this only drops the constraint.

DROP INDEX IF EXISTS uq_campaign_contacts_phone_only;
