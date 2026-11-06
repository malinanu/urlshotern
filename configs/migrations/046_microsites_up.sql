-- Migration 046: Microsites — micro-landing pages & digital vCards (UP)
--
-- Purely additive: three new tables, no changes to existing tables.
-- Pages are served at /m/<slug>; analytics live in microsite_events and are
-- independent of click_events / url_mappings.

-- Pages
CREATE TABLE IF NOT EXISTS microsites (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    team_id      BIGINT,                         -- reserved for team-owned pages (unused in v1)
    slug         VARCHAR(32) NOT NULL,           -- stored lowercase; public URL /m/<slug>
    title        VARCHAR(120) NOT NULL DEFAULT '',
    template     VARCHAR(40)  NOT NULL DEFAULT 'blank',
    status       VARCHAR(20)  NOT NULL DEFAULT 'draft',  -- draft | published
    draft        JSONB NOT NULL DEFAULT '{}'::jsonb,     -- theme + blocks + settings being edited
    published    JSONB,                                  -- snapshot actually served publicly
    version      INTEGER NOT NULL DEFAULT 1,             -- optimistic concurrency for draft saves
    short_code   VARCHAR(50),                            -- optional linked short link
    published_at TIMESTAMP WITH TIME ZONE,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    updated_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    deleted_at   TIMESTAMP WITH TIME ZONE,
    CONSTRAINT chk_microsites_status CHECK (status IN ('draft', 'published'))
);

-- Slugs are unique among live pages only, so deleting a page frees its slug.
CREATE UNIQUE INDEX IF NOT EXISTS uq_microsites_slug_live
    ON microsites (lower(slug)) WHERE deleted_at IS NULL;
CREATE INDEX IF NOT EXISTS idx_microsites_user
    ON microsites (user_id) WHERE deleted_at IS NULL;

-- Uploaded images (avatars, logos, backgrounds)
CREATE TABLE IF NOT EXISTS microsite_assets (
    id           BIGSERIAL PRIMARY KEY,
    user_id      BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    microsite_id BIGINT REFERENCES microsites(id) ON DELETE SET NULL,
    path         VARCHAR(255) NOT NULL,          -- relative to MICROSITE_UPLOAD_PATH
    mime_type    VARCHAR(50)  NOT NULL,
    size_bytes   BIGINT       NOT NULL,
    width        INTEGER      NOT NULL,
    height       INTEGER      NOT NULL,
    sha256       CHAR(64)     NOT NULL,
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW()
);
CREATE INDEX IF NOT EXISTS idx_microsite_assets_user ON microsite_assets (user_id);
CREATE INDEX IF NOT EXISTS idx_microsite_assets_microsite ON microsite_assets (microsite_id);

-- First-party page telemetry. No raw IP addresses are stored.
CREATE TABLE IF NOT EXISTS microsite_events (
    id                BIGSERIAL PRIMARY KEY,
    microsite_id      BIGINT NOT NULL REFERENCES microsites(id) ON DELETE CASCADE,
    event_type        VARCHAR(20) NOT NULL,      -- view | click | vcard_download | call | sms | email | whatsapp
    block_id          VARCHAR(40),
    item_id           VARCHAR(40),
    occurred_at       TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    visitor_hash      CHAR(64),                  -- keyed HMAC of ip|ua|day
    is_unique         BOOLEAN NOT NULL DEFAULT FALSE,
    is_bot            BOOLEAN NOT NULL DEFAULT FALSE,
    country           CHAR(2),
    device_type       VARCHAR(20),
    os                VARCHAR(50),
    browser           VARCHAR(50),
    referrer_domain   VARCHAR(255),
    referrer_category VARCHAR(30),
    utm_source        VARCHAR(100),
    utm_medium        VARCHAR(100),
    utm_campaign      VARCHAR(100)
);
CREATE INDEX IF NOT EXISTS idx_microsite_events_site_time
    ON microsite_events (microsite_id, occurred_at);
CREATE INDEX IF NOT EXISTS idx_microsite_events_site_type_time
    ON microsite_events (microsite_id, event_type, occurred_at);
