-- Migration 039: Seed comprehensive content blocks and page templates

-- ============================================================
-- CONTENT BLOCKS
-- ============================================================

-- Hero Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Hero - Features Page', 'hero_features', 'hero',
  '{"badge":"Features","title":"Powerful Features for Smart URL Management","title_highlight":"Smart URL Management","tagline":"Everything you need to create, track, and optimize your short links. From basic URL shortening to advanced analytics and enterprise features."}',
  true, 10
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Hero - Pricing Page', 'hero_pricing', 'hero',
  '{"badge":"Pricing","title":"Simple, Transparent Pricing","title_highlight":"Pricing","tagline":"Choose the perfect plan for your needs. All plans include our core features with no hidden fees or surprise charges."}',
  true, 11
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Hero - Blog Listing', 'hero_blog', 'hero',
  '{"badge":"Blog","title":"Insights & Updates","title_highlight":"Insights","tagline":"Tips, tutorials, and news from the Trunc team to help you get the most from your links."}',
  true, 12
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Hero - Careers Page', 'hero_careers', 'hero',
  '{"badge":"Careers","title":"Join the Trunc Team","title_highlight":"Trunc Team","tagline":"Help us build the future of link management. We are a remote-first team passionate about performance, analytics, and developer experience."}',
  true, 13
) ON CONFLICT (identifier) DO NOTHING;

-- Stats Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'About Page Statistics', 'about_stats', 'stats',
  '{"items":[{"value":"10M+","label":"Links Shortened","icon":"LinkIcon","color":"text-primary"},{"value":"99.9%","label":"Uptime","icon":"ShieldCheckIcon","color":"text-green-500"},{"value":"150+","label":"Countries","icon":"GlobeAltIcon","color":"text-primary"},{"value":"500K+","label":"Active Users","icon":"UsersIcon","color":"text-primary"}]}',
  true, 20
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Trust Badges', 'trust_badges', 'stats',
  '{"heading":"Enterprise-grade security you can trust","items":[{"label":"SSL/TLS Encrypted","icon":"ShieldCheckIcon"},{"label":"SOC 2 Type II","icon":"CheckCircleIcon"},{"label":"GDPR Compliant","icon":"LockClosedIcon"},{"label":"99.9% SLA","icon":"ClockIcon"}]}',
  true, 21
) ON CONFLICT (identifier) DO NOTHING;

-- Features Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Core Features Grid', 'core_features', 'features',
  '{"badge":"Core Features","heading":"Core Features","tagline":"Built for performance, designed for insights, trusted by thousands.","items":[{"name":"Lightning Fast Shortening","icon":"LinkIcon","description":"Generate short URLs instantly with our optimized infrastructure powered by Redis caching.","details":["Sub-second response times","Global CDN distribution","Auto-scaling infrastructure","Bulk URL processing"]},{"name":"Advanced Analytics","icon":"ChartBarIcon","description":"Track every click with detailed insights including geographic data, device information, and referrer analysis.","details":["Real-time click tracking","Geographic click mapping","Device and browser analytics","Referrer source tracking"]},{"name":"Enterprise Security","icon":"ShieldCheckIcon","description":"Bank-level security with SSL encryption, spam protection, and enterprise-grade infrastructure.","details":["SSL/TLS encryption","Malware URL scanning","Spam link detection","99.9% uptime SLA"]},{"name":"Smart Click Insights","icon":"CursorArrowRaysIcon","description":"Understand your audience better with detailed click analytics and user behavior patterns.","details":["Click heatmaps","Time-based analytics","Conversion tracking","A/B testing support"]},{"name":"Global Performance","icon":"GlobeAltIcon","description":"Worldwide infrastructure ensures your short links work fast everywhere.","details":["Multi-region deployment","Edge caching worldwide","Automatic failover","Load balancing"]},{"name":"Real-time Monitoring","icon":"ClockIcon","description":"Monitor your links in real-time with instant notifications and comprehensive health checks.","details":["Live click monitoring","Alert notifications","Health check endpoints","Performance metrics"]}]}',
  true, 30
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Additional Features Checklist', 'additional_features', 'features',
  '{"badge":"Additional Features","heading":"Additional Features","tagline":"Even more tools and capabilities to enhance your URL management experience.","items":["Custom Short Codes","QR Code Generation","Link Expiration","Password Protection","Team Collaboration","API Access","Webhook Integration","White-label Solution","Mobile Apps","24/7 Support","Data Export","SSO Integration"]}',
  true, 31
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'How It Works', 'how_it_works', 'features',
  '{"badge":"How It Works","heading":"Up and running in minutes","tagline":"Three simple steps to start shortening and tracking your links.","steps":[{"number":"1","title":"Paste Your URL","description":"Enter any long URL into Trunc and we will generate a short link instantly.","icon":"LinkIcon"},{"number":"2","title":"Customize & Share","description":"Set a custom alias, expiry date, or password — then share your link anywhere.","icon":"CursorArrowRaysIcon"},{"number":"3","title":"Track Everything","description":"Watch clicks, locations, devices, and referrers roll in on your real-time dashboard.","icon":"ChartBarIcon"}]}',
  true, 32
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'About - Core Values', 'about_values', 'features',
  '{"badge":"Our Values","heading":"What Drives Us Forward","tagline":"Our core values shape every decision we make and define how we build products that truly serve our customers.","items":[{"title":"Innovation First","icon":"LinkIcon","badge":"Core Principle","description":"We constantly push the boundaries of what''s possible in URL management and analytics."},{"title":"Security & Privacy","icon":"ShieldCheckIcon","badge":"Core Principle","description":"Your data and links are protected with enterprise-grade security and privacy measures."},{"title":"Performance","icon":"ChartBarIcon","badge":"Core Principle","description":"Lightning-fast redirects and real-time analytics ensure optimal user experience."},{"title":"Reliability","icon":"ClockIcon","badge":"Core Principle","description":"99.9% uptime guarantee ensures your links work when your audience needs them most."}]}',
  true, 33
) ON CONFLICT (identifier) DO NOTHING;

-- CTA Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'CTA - Features Page', 'cta_features', 'cta',
  '{"heading":"Ready to Experience These Features?","tagline":"Start using Trunc today and unlock the full potential of your links.","cta_primary":{"text":"Try It Now","url":"/"},"cta_secondary":{"text":"View Pricing","url":"/pricing"}}',
  true, 40
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'CTA - Pricing Page', 'cta_pricing', 'cta',
  '{"heading":"Ready to Get Started?","tagline":"Join thousands of users who trust Trunc for their link management needs.","cta_primary":{"text":"Start Free Trial","url":"/"},"cta_secondary":{"text":"Contact Sales","url":"/contact"}}',
  true, 41
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'CTA - About Page', 'cta_about', 'cta',
  '{"badge":"🚀 Ready to Get Started?","heading":"Join thousands of businesses using Trunc","tagline":"Start shortening URLs and tracking analytics with our powerful platform. No credit card required.","cta_primary":{"text":"Start Free Trial","url":"/register"},"cta_secondary":{"text":"Contact Sales","url":"/contact"},"features":["Free forever plan","No credit card required","Setup in 2 minutes"]}',
  true, 42
) ON CONFLICT (identifier) DO NOTHING;

-- Pricing Block
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Pricing Plans', 'pricing_plans', 'pricing',
  '{"badge":"Pricing","heading":"Simple, Transparent Pricing","tagline":"Choose the perfect plan for your needs.","plans":[{"name":"Free","price_monthly":0,"description":"Perfect for getting started","cta":"Get Started Free","cta_url":"/register","features_included":["10 short links","Basic analytics","Standard support","Link expiration","QR codes"],"features_excluded":["Custom short codes","Team collaboration","Advanced analytics","API access","Priority support"]},{"name":"Pro","badge":"⭐ Most Popular","description":"For growing businesses","cta":"Start Pro Plan","cta_url":"/register","features_included":["1,000 short links","Advanced analytics","Priority support","Link expiration","QR codes","Custom short codes","Custom domains","Team collaboration (5 users)","API access","Password protection"]},{"name":"Ultra","description":"For enterprises & power users","cta":"Choose Ultra","cta_url":"/register","features_included":["Unlimited short links","Advanced analytics","24/7 premium support","Link expiration","QR codes","Custom short codes","Custom domains","Unlimited team members","Unlimited API access","Password protection"]}]}',
  true, 50
) ON CONFLICT (identifier) DO NOTHING;

-- FAQ Block
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'FAQ - Pricing Page', 'faq_pricing', 'faq',
  '{"badge":"FAQ","heading":"Frequently Asked Questions","items":[{"question":"Can I change plans at any time?","answer":"Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and billing is prorated."},{"question":"What happens if I exceed my link limit?","answer":"On the Free plan, you''ll need to upgrade. On paid plans, we''ll notify you and provide options to upgrade or purchase additional capacity."},{"question":"Do you offer annual billing?","answer":"Yes! Annual billing comes with a 20% discount. Contact us for annual pricing details."},{"question":"Is there a free trial for paid plans?","answer":"Yes, we offer a 14-day free trial for all paid plans. No credit card required to start."},{"question":"Can I use my own domain?","answer":"Custom domains are available on Pro and Enterprise plans. You can use your own branded short domain."},{"question":"What kind of support do you provide?","answer":"Free users get community support. Pro users get priority email support. Enterprise users get 24/7 phone and email support."}]}',
  true, 60
) ON CONFLICT (identifier) DO NOTHING;

-- Text / Mission Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'About - Mission', 'about_mission', 'text',
  '{"badge":"Our Mission","heading":"Connecting Every Digital Experience","description":"We envision a world where every URL is intelligent, every click tells a story, and every connection drives meaningful outcomes. Trunc isn''t just a tool — it''s the bridge between intention and impact in the digital ecosystem.","cta_primary":{"text":"Get Started","url":"/register"},"cta_secondary":{"text":"Learn More","url":"/features"}}',
  true, 70
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'About - Team', 'about_team', 'text',
  '{"badge":"Our Team","heading":"Global Remote Team","description":"Our diverse team of engineers, designers, and product experts work from around the world to bring you the best URL shortening experience.","icon":"UsersIcon","cta":{"text":"View Open Positions","url":"/careers"}}',
  true, 71
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'About - Achievements', 'about_achievements', 'text',
  '{"badge":"Achievements","heading":"Key Milestones","items":[{"year":"2024","title":"Reached 10M Links Shortened","description":"Milestone achievement in our growth journey"},{"year":"2023","title":"SOC 2 Type II Compliance","description":"Enterprise-grade security certification"},{"year":"2023","title":"99.9% Uptime Achievement","description":"Consistent reliability and performance"},{"year":"2022","title":"Global Expansion","description":"Serving users in over 150 countries worldwide"}]}',
  true, 72
) ON CONFLICT (identifier) DO NOTHING;

-- List / Footer Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Footer Links - Product', 'footer_links_product', 'list',
  '{"heading":"Product","items":[{"text":"Features","url":"/features"},{"text":"Pricing","url":"/pricing"},{"text":"API Documentation","url":"/api"},{"text":"Analytics","url":"/analytics"}]}',
  true, 80
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Footer Links - Company', 'footer_links_company', 'list',
  '{"heading":"Company","items":[{"text":"About","url":"/about"},{"text":"Contact","url":"/contact"},{"text":"Privacy Policy","url":"/privacy"},{"text":"Terms of Service","url":"/terms"}]}',
  true, 81
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Social Links', 'social_links', 'list',
  '{"items":[{"platform":"Twitter","url":"https://twitter.com/trunc","icon":"twitter"},{"platform":"LinkedIn","url":"https://linkedin.com/company/trunc","icon":"linkedin"},{"platform":"GitHub","url":"https://github.com/trunc","icon":"github"}]}',
  true, 82
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Navigation Links', 'nav_links', 'list',
  '{"items":[{"text":"Home","url":"/"},{"text":"Features","url":"/features"},{"text":"Pricing","url":"/pricing"},{"text":"API Docs","url":"/api-docs"}]}',
  true, 83
) ON CONFLICT (identifier) DO NOTHING;

-- HTML Blocks
INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Footer Description', 'footer_description', 'html',
  '<p>The most reliable URL shortener with advanced analytics, custom domains, and enterprise-grade security. Shorten, track, and optimize your links.</p>',
  true, 90
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Announcement Bar', 'announcement_bar', 'html',
  '<!-- Empty announcement bar — edit to activate -->',
  false, 91
) ON CONFLICT (identifier) DO NOTHING;

INSERT INTO content_blocks (name, identifier, type, content, is_active, sort_order)
VALUES (
  'Cookie Notice', 'cookie_notice', 'html',
  '<div class="cookie-notice">We use cookies to improve your experience. By continuing, you agree to our <a href="/privacy">Privacy Policy</a>.</div>',
  false, 92
) ON CONFLICT (identifier) DO NOTHING;

-- ============================================================
-- PAGE TEMPLATES
-- ============================================================

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Default Page', 'default',
  'Standard page with header content area and footer',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <header>{{block:nav_links}}</header>
  <main class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
    {{content}}
  </main>
  <footer class="bg-gray-900 text-white py-12">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">{{block:footer_description}}</div>
  </footer>
</body>
</html>',
  1, true
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Landing Page', 'landing',
  'Full-width sections optimised for conversion',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <header>{{block:announcement_bar}}{{block:nav_links}}</header>
  <main>
    <section class="bg-gradient-to-br from-blue-600 to-purple-700 text-white py-24">
      <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
        {{content}}
      </div>
    </section>
  </main>
  <footer class="bg-gray-900 text-white py-12">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 grid grid-cols-1 md:grid-cols-3 gap-8">
      <div>{{block:footer_description}}</div>
      <div>{{block:footer_links_product}}</div>
      <div>{{block:footer_links_company}}</div>
    </div>
  </footer>
</body>
</html>',
  2, true
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Blog Post', 'blog',
  'Article layout with author info and reading time',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <header>{{block:nav_links}}</header>
  <main class="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-20">
    <article class="prose prose-lg max-w-none">
      <header class="mb-8 not-prose">
        <div class="text-sm text-gray-500 mb-2">{{author}} · {{reading_time}} min read · {{date}}</div>
        <h1 class="text-4xl font-bold text-gray-900 mb-4">{{title}}</h1>
      </header>
      {{content}}
    </article>
  </main>
  <footer class="bg-gray-900 text-white py-12">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">{{block:footer_description}}</div>
  </footer>
</body>
</html>',
  3, true
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Minimal', 'minimal',
  'Clean content-only layout with minimal chrome',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <main class="max-w-4xl mx-auto px-4 py-16">
    {{content}}
  </main>
</body>
</html>',
  4, true
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Documentation', 'docs',
  'Two-column layout with sidebar table of contents',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <header>{{block:nav_links}}</header>
  <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
    <div class="flex gap-12">
      <aside class="hidden lg:block w-64 flex-shrink-0">
        <nav class="sticky top-8 space-y-1">{{toc}}</nav>
      </aside>
      <main class="flex-1 prose prose-lg max-w-none">
        {{content}}
      </main>
    </div>
  </div>
  <footer class="bg-gray-900 text-white py-12">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">{{block:footer_description}}</div>
  </footer>
</body>
</html>',
  5, true
) ON CONFLICT (slug) DO NOTHING;

INSERT INTO page_templates (name, slug, description, content, sort_order, is_active)
VALUES (
  'Full Width', 'full-width',
  'Edge-to-edge layout with no max-width container',
  '<!DOCTYPE html>
<html lang="en">
<head><meta charset="UTF-8"><meta name="viewport" content="width=device-width, initial-scale=1.0"><title>{{title}}</title></head>
<body class="bg-white text-gray-900">
  <header>{{block:nav_links}}</header>
  <main class="w-full">
    {{content}}
  </main>
  <footer class="bg-gray-900 text-white py-12 w-full">
    <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">{{block:footer_description}}</div>
  </footer>
</body>
</html>',
  6, true
) ON CONFLICT (slug) DO NOTHING;
