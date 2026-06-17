-- Migration 045: Deduplicate phone-only campaign contacts (UP)
-- Email-only/email+phone contacts already dedupe via uq_campaign_contacts_email.
-- Phone-only contacts (email IS NULL) had no equivalent constraint, so re-importing
-- the same phone number repeatedly created duplicate campaign_contacts rows.
-- This merges existing duplicates before adding a partial unique index.

CREATE TEMP TABLE _phone_dupe_map AS
SELECT r.id AS dup_id, c.keep_id
FROM (
  SELECT id, phone, ROW_NUMBER() OVER (PARTITION BY phone ORDER BY id) AS rn
  FROM campaign_contacts
  WHERE email IS NULL AND phone IS NOT NULL
) r
JOIN (
  SELECT phone, MIN(id) AS keep_id
  FROM campaign_contacts
  WHERE email IS NULL AND phone IS NOT NULL
  GROUP BY phone
) c ON c.phone = r.phone
WHERE r.rn > 1;

-- Drop links/sends on the duplicate that would collide with one the canonical
-- contact already has, then repoint the rest to the canonical contact.
DELETE FROM campaign_contact_links l
USING _phone_dupe_map m
WHERE l.contact_id = m.dup_id
  AND EXISTS (
    SELECT 1 FROM campaign_contact_links l2
    WHERE l2.campaign_id = l.campaign_id AND l2.contact_id = m.keep_id
  );

UPDATE campaign_contact_links l
SET contact_id = m.keep_id
FROM _phone_dupe_map m
WHERE l.contact_id = m.dup_id;

DELETE FROM campaign_sends s
USING _phone_dupe_map m
WHERE s.contact_id = m.dup_id
  AND EXISTS (
    SELECT 1 FROM campaign_sends s2
    WHERE s2.campaign_id = s.campaign_id AND s2.contact_id = m.keep_id AND s2.channel = s.channel
  );

UPDATE campaign_sends s
SET contact_id = m.keep_id
FROM _phone_dupe_map m
WHERE s.contact_id = m.dup_id;

DELETE FROM campaign_contacts c
USING _phone_dupe_map m
WHERE c.id = m.dup_id;

DROP TABLE _phone_dupe_map;

CREATE UNIQUE INDEX IF NOT EXISTS uq_campaign_contacts_phone_only
  ON campaign_contacts (phone)
  WHERE email IS NULL;
