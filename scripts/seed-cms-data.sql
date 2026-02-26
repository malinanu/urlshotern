-- Seed Default CMS Data (Templates, Blocks, Pages, Menus, Settings)
-- Run this in Adminer or psql to populate your production database with the default CMS builder templates.

-- 1. Insert default global settings
INSERT INTO global_settings (key, value, type, category, display_name, description, is_public, sort_order) VALUES
    ('site_name', 'Trunc', 'text', 'general', 'Site Name', 'The name of your website', true, 1),
    ('site_tagline', 'Smart URL Management Platform', 'text', 'general', 'Site Tagline', 'A short description of your site', true, 2),
    ('company_email', 'contact@3logiq.com', 'email', 'contact', 'Company Email', 'Main contact email address', true, 3),
    ('company_phone', '+1-555-0123', 'text', 'contact', 'Company Phone', 'Main contact phone number', true, 4),
    ('company_address', '123 Innovation Drive, Tech City, TC 12345', 'textarea', 'contact', 'Company Address', 'Physical business address', true, 5),
    ('social_twitter', 'https://twitter.com/trunc', 'url', 'social', 'Twitter URL', 'Twitter profile URL', true, 6),
    ('social_linkedin', 'https://linkedin.com/company/trunc', 'url', 'social', 'LinkedIn URL', 'LinkedIn company URL', true, 7),
    ('social_github', 'https://github.com/trunc', 'url', 'social', 'GitHub URL', 'GitHub organization URL', true, 8),
    ('google_analytics', '', 'text', 'tracking', 'Google Analytics ID', 'Google Analytics tracking ID (e.g., GA_MEASUREMENT_ID)', false, 9),
    ('meta_description', 'Professional URL shortening service with advanced analytics, team collaboration, and custom domains.', 'textarea', 'seo', 'Default Meta Description', 'Default meta description for pages without custom descriptions', true, 10),
    ('footer_copyright', '© 2024 3logiq Technologies, Inc. All rights reserved.', 'text', 'footer', 'Copyright Text', 'Copyright notice in footer', true, 11),
    ('header_announcement', '', 'html', 'header', 'Announcement Bar', 'HTML content for announcement bar (leave empty to hide)', true, 12),
    ('maintenance_mode', 'false', 'boolean', 'system', 'Maintenance Mode', 'Enable to show maintenance page to visitors', false, 13)
ON CONFLICT (key) DO NOTHING;

-- 2. Insert default content blocks
INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order) VALUES
    ('Main Footer', 'main_footer', 
    '<div class="bg-gray-900 text-white">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-12">
            <div class="grid grid-cols-1 md:grid-cols-4 gap-8">
                <div class="col-span-1 md:col-span-2">
                    <h3 class="text-lg font-semibold mb-4">{{site_name}}</h3>
                    <p class="text-gray-300 mb-6">{{site_tagline}}</p>
                    <div class="flex space-x-4">
                        <a href="{{social_twitter}}" class="text-gray-400 hover:text-white">Twitter</a>
                        <a href="{{social_linkedin}}" class="text-gray-400 hover:text-white">LinkedIn</a>
                        <a href="{{social_github}}" class="text-gray-400 hover:text-white">GitHub</a>
                    </div>
                </div>
                <div>
                    <h4 class="font-medium mb-4">Product</h4>
                    <ul class="space-y-2 text-sm text-gray-300">
                        <li><a href="/features" class="hover:text-white">Features</a></li>
                        <li><a href="/pricing" class="hover:text-white">Pricing</a></li>
                        <li><a href="/api-docs" class="hover:text-white">API</a></li>
                        <li><a href="/integrations" class="hover:text-white">Integrations</a></li>
                    </ul>
                </div>
                <div>
                    <h4 class="font-medium mb-4">Support</h4>
                    <ul class="space-y-2 text-sm text-gray-300">
                        <li><a href="/help" class="hover:text-white">Help Center</a></li>
                        <li><a href="/contact" class="hover:text-white">Contact</a></li>
                        <li><a href="/status" class="hover:text-white">Status</a></li>
                        <li><a href="/community" class="hover:text-white">Community</a></li>
                    </ul>
                </div>
            </div>
            <div class="border-t border-gray-800 mt-8 pt-8 text-center text-sm text-gray-400">
                {{footer_copyright}}
            </div>
        </div>
    </div>', 'footer', true, 1),
    
    ('Homepage Hero Section', 'hero_home',
    '<div class="relative overflow-hidden bg-white">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-24">
            <div class="text-center">
                <h1 class="text-4xl sm:text-6xl font-bold text-gray-900 mb-8">
                    Shorten URLs, 
                    <span class="bg-gradient-to-r from-blue-600 to-purple-600 bg-clip-text text-transparent">
                        Amplify Results
                    </span>
                </h1>
                <p class="text-xl text-gray-600 mb-10 max-w-3xl mx-auto">
                    Create powerful short links with advanced analytics, team collaboration, and enterprise-grade security. 
                    Perfect for marketing campaigns, social media, and business communications.
                </p>
                <div class="flex flex-col sm:flex-row gap-4 justify-center">
                    <a href="/register" class="bg-blue-600 text-white px-8 py-4 rounded-xl font-semibold hover:bg-blue-700 transition-colors">
                        Start Free Trial
                    </a>
                    <a href="/demo" class="border-2 border-gray-300 text-gray-700 px-8 py-4 rounded-xl font-semibold hover:bg-gray-50 transition-colors">
                        Watch Demo
                    </a>
                </div>
            </div>
        </div>
    </div>', 'hero', true, 2),

    ('Features Grid', 'features_grid',
    '<div class="py-16 bg-gray-50">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="grid grid-cols-1 md:grid-cols-3 gap-8">
                <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <div class="w-12 h-12 bg-blue-100 text-blue-600 rounded-lg flex items-center justify-center mb-4">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13 10V3L4 14h7v7l9-11h-7z"></path></svg>
                    </div>
                    <h3 class="text-xl font-bold mb-2">Lightning Fast</h3>
                    <p class="text-gray-600">Built on modern infrastructure to ensure your links redirect instantly, anywhere in the world.</p>
                </div>
                <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <div class="w-12 h-12 bg-purple-100 text-purple-600 rounded-lg flex items-center justify-center mb-4">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg>
                    </div>
                    <h3 class="text-xl font-bold mb-2">Advanced Analytics</h3>
                    <p class="text-gray-600">Track clicks, locations, devices, and referrers to truly understand your audience.</p>
                </div>
                <div class="bg-white p-6 rounded-xl shadow-sm border border-gray-100">
                    <div class="w-12 h-12 bg-green-100 text-green-600 rounded-lg flex items-center justify-center mb-4">
                        <svg class="w-6 h-6" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M12 15v2m-6 4h12a2 2 0 002-2v-6a2 2 0 00-2-2H6a2 2 0 00-2 2v6a2 2 0 002 2zm10-10V7a4 4 0 00-8 0v4h8z"></path></svg>
                    </div>
                    <h3 class="text-xl font-bold mb-2">Enterprise Security</h3>
                    <p class="text-gray-600">SSL encryption, SOC2 compliance, and single sign-on keep your team''s data safe.</p>
                </div>
            </div>
        </div>
    </div>', 'features', true, 3),

    ('Social Proof Stats', 'social_proof_stats',
    '<div class="bg-white py-12 border-y border-gray-100">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div class="grid grid-cols-2 md:grid-cols-4 gap-8 text-center">
                <div>
                    <div class="text-4xl font-extrabold text-blue-600">10M+</div>
                    <div class="text-sm font-medium text-gray-500 uppercase tracking-wide mt-1">Links Created</div>
                </div>
                <div>
                    <div class="text-4xl font-extrabold text-blue-600">5B+</div>
                    <div class="text-sm font-medium text-gray-500 uppercase tracking-wide mt-1">Clicks Tracked</div>
                </div>
                <div>
                    <div class="text-4xl font-extrabold text-blue-600">50K+</div>
                    <div class="text-sm font-medium text-gray-500 uppercase tracking-wide mt-1">Active Users</div>
                </div>
                <div>
                    <div class="text-4xl font-extrabold text-blue-600">99.9%</div>
                    <div class="text-sm font-medium text-gray-500 uppercase tracking-wide mt-1">Uptime</div>
                </div>
            </div>
        </div>
    </div>', 'stats', true, 4),

    ('CTA Bottom', 'cta_bottom',
    '<div class="bg-gradient-to-r from-blue-600 to-purple-700 text-white py-16">
        <div class="max-w-4xl mx-auto text-center px-4 sm:px-6 lg:px-8">
            <h2 class="text-3xl md:text-4xl font-bold mb-6">Ready to Get Started?</h2>
            <p class="text-xl mb-8 opacity-90">
                Join thousands of businesses that trust {{site_name}} for their link management needs.
            </p>
            <div class="flex flex-col sm:flex-row gap-4 justify-center">
                <a href="/register" class="bg-white text-blue-600 px-8 py-4 rounded-xl font-semibold hover:bg-gray-50 transition-colors">
                    Create Account
                </a>
                <a href="/contact" class="border-2 border-white text-white px-8 py-4 rounded-xl font-semibold hover:bg-white hover:text-blue-600 transition-colors">
                    Contact Sales
                </a>
            </div>
        </div>
    </div>', 'cta', true, 5)
ON CONFLICT (identifier) DO NOTHING;

-- 3. Insert default page templates
INSERT INTO page_templates (name, slug, description, content, fields, is_active, sort_order) VALUES
    ('Blank Canvas', 'blank', 'An empty canvas ready for you to build upon', 
    '[]', '{}', true, 1),
    
    ('Default Layout', 'default', 'Standard layout with header and footer', 
    '[
      {
        "id": "tpl-header",
        "type": "content-block",
        "content": {"blockId": "main_header"},
        "settings": {"padding": "md"}
      },
      {
        "id": "tpl-content",
        "type": "text",
        "content": {"html": "<div class=\"max-w-7xl mx-auto px-4 py-8\"><h1 class=\"text-3xl font-bold\">Page Title</h1><p class=\"mt-4\">Page content goes here...</p></div>"},
        "settings": {"padding": "lg"}
      },
      {
        "id": "tpl-footer",
        "type": "content-block",
        "content": {"blockId": "main_footer"},
        "settings": {"padding": "md"}
      }
    ]', '{}', true, 2),
    
    ('Marketing Landing Page', 'marketing_landing', 'Complete landing page layout with Hero, Features, Stats, and CTA',
    '[
      {
        "id": "tpl-hero",
        "type": "content-block",
        "content": {"blockId": "hero_home"},
        "settings": {"padding": "none"}
      },
      {
        "id": "tpl-stats",
        "type": "content-block",
        "content": {"blockId": "social_proof_stats"},
        "settings": {"padding": "none"}
      },
      {
        "id": "tpl-features",
        "type": "content-block",
        "content": {"blockId": "features_grid"},
        "settings": {"padding": "none"}
      },
      {
        "id": "tpl-cta",
        "type": "content-block",
        "content": {"blockId": "cta_bottom"},
        "settings": {"padding": "none"}
      }
    ]', '{"hero_title": "Hero Title", "cta_text": "Call to Action"}', true, 3),

    ('About Us Page', 'about', 'Standard structural About Us page setup',
    '[
      {
        "id": "tpl-about-header",
        "type": "text",
        "content": {"html": "<div class=\"bg-blue-600 text-white text-center py-20\"><h1 class=\"text-5xl font-bold mb-4\">About Us</h1><p class=\"text-xl opacity-90\">Learn about our mission and team.</p></div>"},
        "settings": {"padding": "none"}
      },
      {
        "id": "tpl-about-content",
        "type": "columns",
        "content": {
          "columns": [
            {"content": "<h2 class=\"text-2xl font-bold mb-4\">Our Mission</h2><p>Lorem ipsum dolor sit amet, consectetur adipiscing elit. Nullam in odio eget purus commodo tempor.</p>"},
            {"content": "<img src=\"https://images.unsplash.com/photo-1522071820081-009f0129c71c?ixlib=rb-1.2.1&auto=format&fit=crop&w=800&q=80\" alt=\"Team\" class=\"rounded-lg shadow-md\" />"}
          ]
        },
        "settings": {"padding": "xl"}
      },
      {
        "id": "tpl-about-cta",
        "type": "content-block",
        "content": {"blockId": "cta_bottom"},
        "settings": {"padding": "none"}
      }
    ]', '{}', true, 4)
ON CONFLICT (slug) DO NOTHING;

-- 4. Insert default static pages
INSERT INTO static_pages (slug, title, content, content_blocks, meta_description, is_published, sort_order, published_at) VALUES
    ('home', 'Trunc - Professional URL Shortener', '<div class="home-page"><p>Welcome to Trunc. This page uses the visual builder.</p></div>', 
    '[
      {"id":"block-1","type":"content-block","content":{"blockId":2},"settings":{}},
      {"id":"block-2","type":"content-block","content":{"blockId":4},"settings":{}},
      {"id":"block-3","type":"content-block","content":{"blockId":3},"settings":{}},
      {"id":"block-4","type":"content-block","content":{"blockId":5},"settings":{}},
      {"id":"block-5","type":"content-block","content":{"blockId":1},"settings":{}}
    ]', 
    'Trunc - Professional URL shortener with advanced analytics, custom branding, and enterprise reliability.', true, 0, NOW()),

    ('features', 'Features - Trunc URL Shortener', '<div class="features-page"><p>Discover all the tools and capabilities.</p></div>',
    '[
      {"id":"feat-1","type":"text","content":{"html":"<div class=\"py-20 text-center bg-gray-50\"><h1 class=\"text-4xl font-bold mb-4\">Features</h1></div>"},"settings":{}},
      {"id":"feat-2","type":"content-block","content":{"blockId":3},"settings":{}},
      {"id":"feat-3","type":"content-block","content":{"blockId":1},"settings":{}}
    ]', 
    'Discover Trunc''s powerful features including lightning-fast URL shortening, advanced analytics, custom domains, and enterprise security.', true, 1, NOW())
ON CONFLICT (slug) DO NOTHING;

-- 5. Insert default navigation menus
INSERT INTO navigation_menus (name, location, items, is_active, sort_order) VALUES
    ('Main Navigation', 'header', 
    '[
        {"label": "Home", "url": "/", "type": "internal"},
        {"label": "Features", "url": "/features", "type": "internal"},
        {"label": "Pricing", "url": "/pricing", "type": "internal"},
        {"label": "About", "url": "/about", "type": "internal"},
        {"label": "Help", "url": "/help", "type": "internal"},
        {"label": "Contact", "url": "/contact", "type": "internal"}
    ]', true, 1),
    
    ('Footer Links', 'footer',
    '[
        {"label": "Privacy Policy", "url": "/privacy", "type": "internal"},
        {"label": "Terms of Service", "url": "/terms", "type": "internal"},
        {"label": "Status", "url": "/status", "type": "internal"},
        {"label": "Blog", "url": "/blog", "type": "internal"}
    ]', true, 2)
ON CONFLICT (name, location) DO NOTHING;

SELECT 'Successfully seeded default CMS Page Templates, Content Blocks, Settings, and Menus.' AS status;
