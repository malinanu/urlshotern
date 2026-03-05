-- Rollback migration 039: Remove seeded content blocks and page templates

-- Remove page templates
DELETE FROM page_templates WHERE slug IN ('default', 'landing', 'blog', 'minimal', 'docs', 'full-width');

-- Remove content blocks added in this migration
DELETE FROM content_blocks WHERE identifier IN (
  'hero_features',
  'hero_pricing',
  'hero_blog',
  'hero_careers',
  'about_stats',
  'trust_badges',
  'core_features',
  'additional_features',
  'how_it_works',
  'about_values',
  'cta_features',
  'cta_pricing',
  'cta_about',
  'pricing_plans',
  'faq_pricing',
  'about_mission',
  'about_team',
  'about_achievements',
  'footer_links_product',
  'footer_links_company',
  'social_links',
  'nav_links',
  'footer_description',
  'announcement_bar',
  'cookie_notice'
);
