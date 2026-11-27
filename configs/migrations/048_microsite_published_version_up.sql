-- Migration 048: Microsite published version (UP)
--
-- Purely additive. Records which draft version was last published, so the
-- Pages list and editor can show "Unpublished changes" when the draft has been
-- edited since. Pages that are published today are treated as up to date.

ALTER TABLE microsites ADD COLUMN IF NOT EXISTS published_version INTEGER;

UPDATE microsites
   SET published_version = version
 WHERE status = 'published' AND published_version IS NULL;
