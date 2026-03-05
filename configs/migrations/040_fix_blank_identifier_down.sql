-- Rollback: wipe identifier again (noop in practice)
UPDATE content_blocks SET identifier = '' WHERE id = 1 AND identifier = 'hero_home';
