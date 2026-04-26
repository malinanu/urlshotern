-- Migration 043: Promotional Campaign System (DOWN)

DROP TABLE IF EXISTS campaign_sends          CASCADE;
DROP TABLE IF EXISTS campaign_contact_links  CASCADE;
DROP TABLE IF EXISTS campaign_contacts       CASCADE;
DROP TABLE IF EXISTS campaigns              CASCADE;
