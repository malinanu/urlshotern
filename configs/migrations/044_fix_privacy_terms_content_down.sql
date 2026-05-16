-- Migration 044: Fix Privacy and Terms page content (Down)
-- Revert content to block JSON placeholders

UPDATE static_pages
SET content = '[{"id":"block-privacy","type":"text","content":{"html":"<h1>Privacy Policy</h1><p>Your privacy policy goes here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]',
    updated_at = NOW()
WHERE slug = 'privacy';

UPDATE static_pages
SET content = '[{"id":"block-terms","type":"text","content":{"html":"<h1>Terms of Service</h1><p>Your terms go here.</p>"},"styles":{"padding":"32px","margin":"0px","backgroundColor":"transparent","textColor":"#000000","borderRadius":"0px","border":"none","textAlign":"left","fontSize":"16px","fontWeight":"normal"},"settings":{"visible":true}}]',
    updated_at = NOW()
WHERE slug = 'terms';
