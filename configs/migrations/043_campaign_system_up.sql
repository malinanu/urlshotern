-- Migration 043: Promotional Campaign System (UP)


-- Campaigns table
CREATE TABLE IF NOT EXISTS campaigns (
    id          BIGSERIAL PRIMARY KEY,
    name        VARCHAR(255) NOT NULL,
    description TEXT,
    ai_context  TEXT,           -- Hint fed to the AI when composing emails
    channels    TEXT[]  NOT NULL DEFAULT '{}',  -- ['email','sms','whatsapp']
    status      VARCHAR(50) NOT NULL DEFAULT 'draft',  -- draft | sending | done
    created_by  BIGINT REFERENCES users(id) ON DELETE SET NULL,
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);

-- Contacts table (shared across campaigns)
CREATE TABLE IF NOT EXISTS campaign_contacts (
    id         BIGSERIAL PRIMARY KEY,
    name       VARCHAR(255),
    email      VARCHAR(320),
    phone      VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_campaign_contacts_email UNIQUE (email),
    CONSTRAINT chk_campaign_contacts_has_contact CHECK (email IS NOT NULL OR phone IS NOT NULL)
);

-- Per-send tracking: one row per contact × campaign × channel
CREATE TABLE IF NOT EXISTS campaign_sends (
    id          BIGSERIAL PRIMARY KEY,
    campaign_id BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    contact_id  BIGINT NOT NULL REFERENCES campaign_contacts(id) ON DELETE CASCADE,
    channel     VARCHAR(30) NOT NULL,  -- email | sms | whatsapp
    status      VARCHAR(30) NOT NULL DEFAULT 'pending',  -- pending | sent | failed
    error       TEXT,
    sent_at     TIMESTAMP WITH TIME ZONE,
    joined_at   TIMESTAMP WITH TIME ZONE,  -- set when the contact registers
    created_at  TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT uq_campaign_sends UNIQUE (campaign_id, contact_id, channel)
);

-- Junction table: which contacts belong to which campaign
CREATE TABLE IF NOT EXISTS campaign_contact_links (
    campaign_id BIGINT NOT NULL REFERENCES campaigns(id) ON DELETE CASCADE,
    contact_id  BIGINT NOT NULL REFERENCES campaign_contacts(id) ON DELETE CASCADE,
    PRIMARY KEY (campaign_id, contact_id)
);

-- Indexes
CREATE INDEX IF NOT EXISTS idx_campaign_sends_campaign_id ON campaign_sends(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_sends_contact_id  ON campaign_sends(contact_id);
CREATE INDEX IF NOT EXISTS idx_campaign_contact_links_campaign ON campaign_contact_links(campaign_id);
CREATE INDEX IF NOT EXISTS idx_campaign_contacts_email   ON campaign_contacts(email);
CREATE INDEX IF NOT EXISTS idx_campaign_contacts_phone   ON campaign_contacts(phone);
