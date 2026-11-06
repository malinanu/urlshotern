-- Migration 047: Microsite form submissions (UP)
--
-- Purely additive. Newsletter and contact form submissions from published
-- microsite pages (Pro and Ultra plans). No raw IP addresses are stored.

CREATE TABLE IF NOT EXISTS microsite_submissions (
    id           BIGSERIAL PRIMARY KEY,
    microsite_id BIGINT NOT NULL REFERENCES microsites(id) ON DELETE CASCADE,
    block_id     VARCHAR(40) NOT NULL,
    kind         VARCHAR(20) NOT NULL,          -- newsletter | contact
    email        VARCHAR(254) NOT NULL,
    name         VARCHAR(100),
    message      TEXT,
    visitor_hash CHAR(64),
    created_at   TIMESTAMP WITH TIME ZONE NOT NULL DEFAULT NOW(),
    CONSTRAINT chk_microsite_submissions_kind CHECK (kind IN ('newsletter', 'contact'))
);
CREATE INDEX IF NOT EXISTS idx_microsite_submissions_site_time
    ON microsite_submissions (microsite_id, created_at DESC);
