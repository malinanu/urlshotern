DELETE FROM role_permissions
WHERE role_id = (SELECT id FROM roles WHERE name = 'admin')
  AND permission_id IN (
    SELECT id FROM permissions WHERE name IN (
      'cms.pages.create', 'cms.pages.read', 'cms.pages.update',
      'cms.pages.delete', 'cms.pages.publish',
      'cms.blocks.create', 'cms.blocks.read', 'cms.blocks.update', 'cms.blocks.delete',
      'cms.media.upload', 'cms.media.delete',
      'cms.settings.read', 'cms.settings.update'
    )
  );
