-- Restore identifier for the Homepage Hero block that was wiped to empty string
UPDATE content_blocks
SET identifier = 'hero_home'
WHERE id = 1
  AND identifier = ''
  AND name = 'Homepage Hero Section';
