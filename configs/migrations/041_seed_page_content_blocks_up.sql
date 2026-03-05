-- Features page
INSERT INTO static_pages (slug, title, content, content_blocks, template, is_published, sort_order)
VALUES (
  'features', 'Features', '',
  '["hero_features","core_features","additional_features","how_it_works","trust_badges","cta_features"]',
  'landing', true, 2
)
ON CONFLICT (slug) DO UPDATE SET
  content_blocks = EXCLUDED.content_blocks,
  template       = EXCLUDED.template,
  updated_at     = NOW();

-- Pricing page
INSERT INTO static_pages (slug, title, content, content_blocks, template, is_published, sort_order)
VALUES (
  'pricing', 'Pricing', '',
  '["hero_pricing","pricing_plans","faq_pricing","cta_pricing"]',
  'landing', true, 3
)
ON CONFLICT (slug) DO UPDATE SET
  content_blocks = EXCLUDED.content_blocks,
  template       = EXCLUDED.template,
  updated_at     = NOW();

-- About page
INSERT INTO static_pages (slug, title, content, content_blocks, template, is_published, sort_order)
VALUES (
  'about', 'About', '',
  '["about_stats","about_mission","about_values","about_team","about_achievements","cta_about"]',
  'default', true, 4
)
ON CONFLICT (slug) DO UPDATE SET
  content_blocks = EXCLUDED.content_blocks,
  template       = EXCLUDED.template,
  updated_at     = NOW();
