-- Grant cms.pages.* and cms.blocks.* permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin'
  AND p.name IN (
    'cms.pages.create', 'cms.pages.read', 'cms.pages.update',
    'cms.pages.delete', 'cms.pages.publish',
    'cms.blocks.create', 'cms.blocks.read', 'cms.blocks.update', 'cms.blocks.delete',
    'cms.media.upload', 'cms.media.delete',
    'cms.settings.read', 'cms.settings.update'
  )
ON CONFLICT (role_id, permission_id) DO NOTHING;
