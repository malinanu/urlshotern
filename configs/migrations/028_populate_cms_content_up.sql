-- Migration 028: Populate CMS with Frontend Content
-- Created: 2026-02-14
-- Purpose: Migrate hardcoded frontend content to CMS for easy editing

-- ============================================================================
-- HOMEPAGE CONTENT
-- ============================================================================
INSERT INTO static_pages (
    slug,
    title,
    content,
    content_blocks,
    meta_title,
    meta_description,
    meta_keywords,
    is_published,
    created_at,
    updated_at
) VALUES (
    'home',
    'Shorten URLs. Track Everything.',
    'Transform your long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more.',
    '{
        "hero": {
            "badge": "✨ New: Advanced Analytics Dashboard",
            "title": "Shorten URLs. Track Everything.",
            "title_highlight": "Track Everything",
            "tagline": "Transform your long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more.",
            "cta_primary": {
                "text": "Get Started Free",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "View Pricing",
                "url": "/pricing"
            }
        },
        "social_proof": {
            "heading": "Trusted by thousands of businesses worldwide",
            "stats": [
                {
                    "value": "500K+",
                    "label": "URLs shortened"
                },
                {
                    "value": "50M+",
                    "label": "Clicks tracked"
                },
                {
                    "value": "99.9%",
                    "label": "Uptime"
                }
            ]
        },
        "features": {
            "badge": "Features",
            "heading": "Why Choose Trunc?",
            "tagline": "More than just a URL shortener. Get powerful features that help you understand and optimize your links.",
            "items": [
                {
                    "name": "Lightning Fast",
                    "description": "Shorten URLs instantly with our optimized infrastructure.",
                    "icon": "LinkIcon"
                },
                {
                    "name": "Advanced Analytics",
                    "description": "Track clicks, locations, and referrers with detailed insights.",
                    "icon": "ChartBarIcon"
                },
                {
                    "name": "Secure & Reliable",
                    "description": "Enterprise-grade security with 99.9% uptime guarantee.",
                    "icon": "ShieldCheckIcon"
                }
            ]
        },
        "cta_bottom": {
            "badge": "🚀 Get Started Today",
            "heading": "Ready to get started?",
            "tagline": "Join thousands of businesses using Trunc to track and optimize their links. Start for free today with no credit card required.",
            "cta_primary": {
                "text": "Get Started Free",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "View Pricing",
                "url": "/pricing"
            },
            "features": [
                "Free forever plan",
                "No credit card required",
                "Setup in 2 minutes"
            ]
        }
    }'::jsonb,
    'Shorten URLs. Track Everything. | Trunc - Fast URL Shortener',
    'Transform your long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more.',
    'url shortener, link shortener, analytics, tracking, short links',
    true,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    content_blocks = EXCLUDED.content_blocks,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    meta_keywords = EXCLUDED.meta_keywords,
    is_published = EXCLUDED.is_published,
    updated_at = NOW();

-- ============================================================================
-- ABOUT PAGE CONTENT
-- ============================================================================
INSERT INTO static_pages (
    slug,
    title,
    content,
    content_blocks,
    meta_title,
    meta_description,
    meta_keywords,
    is_published,
    created_at,
    updated_at
) VALUES (
    'about',
    'We Are Trunc',
    'Pioneering the next generation of URL management with cutting-edge technology, unparalleled performance, and a vision that connects the digital world.',
    '{
        "hero": {
            "badge": "About Trunc",
            "badge_icon": "BuildingOfficeIcon",
            "title": "We Are Trunc",
            "title_highlight": "Trunc",
            "tagline": "Pioneering the next generation of URL management with cutting-edge technology, unparalleled performance, and a vision that connects the digital world."
        },
        "stats": [
            {
                "label": "Links Shortened",
                "value": "10M+",
                "icon": "LinkIcon",
                "color": "text-primary-600"
            },
            {
                "label": "Uptime",
                "value": "99.9%",
                "icon": "ShieldCheckIcon",
                "color": "text-green-500"
            },
            {
                "label": "Countries",
                "value": "150+",
                "icon": "GlobeAltIcon",
                "color": "text-primary-600"
            },
            {
                "label": "Active Users",
                "value": "500K+",
                "icon": "UsersIcon",
                "color": "text-primary-600"
            }
        ],
        "mission": {
            "badge": "Our Mission",
            "heading": "Connecting Every Digital Experience",
            "description": "We envision a world where every URL is intelligent, every click tells a story, and every connection drives meaningful outcomes. Trunc isn''t just a tool—it''s the bridge between intention and impact in the digital ecosystem.",
            "cta_primary": {
                "text": "Get Started",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "Learn More",
                "url": "/features"
            }
        },
        "values": {
            "badge": "Our Values",
            "heading": "What Drives Us Forward",
            "tagline": "Our core values shape every decision we make and define how we build products that truly serve our customers.",
            "items": [
                {
                    "title": "Innovation First",
                    "description": "We constantly push the boundaries of what''s possible in URL management and analytics.",
                    "icon": "LinkIcon"
                },
                {
                    "title": "Security & Privacy",
                    "description": "Your data and links are protected with enterprise-grade security and privacy measures.",
                    "icon": "ShieldCheckIcon"
                },
                {
                    "title": "Performance",
                    "description": "Lightning-fast redirects and real-time analytics ensure optimal user experience.",
                    "icon": "ChartBarIcon"
                },
                {
                    "title": "Reliability",
                    "description": "99.9% uptime guarantee ensures your links work when your audience needs them most.",
                    "icon": "ClockIcon"
                }
            ]
        },
        "team": {
            "badge": "Our Team",
            "heading": "Global Remote Team",
            "description": "Our diverse team of engineers, designers, and product experts work from around the world to bring you the best URL shortening experience.",
            "cta": {
                "text": "View Open Positions",
                "url": "/careers"
            }
        },
        "achievements": {
            "badge": "Achievements",
            "heading": "Key Milestones",
            "items": [
                {
                    "year": "2024",
                    "title": "Reached 10M Links Shortened",
                    "description": "Milestone achievement in our growth journey"
                },
                {
                    "year": "2023",
                    "title": "SOC 2 Type II Compliance",
                    "description": "Enterprise-grade security certification"
                },
                {
                    "year": "2023",
                    "title": "99.9% Uptime Achievement",
                    "description": "Consistent reliability and performance"
                },
                {
                    "year": "2022",
                    "title": "Global Expansion",
                    "description": "Serving users in over 150 countries worldwide"
                }
            ]
        },
        "cta_bottom": {
            "badge": "🚀 Ready to Get Started?",
            "heading": "Join thousands of businesses using Trunc",
            "tagline": "Start shortening URLs and tracking analytics with our powerful platform. No credit card required to get started.",
            "cta_primary": {
                "text": "Start Free Trial",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "Contact Sales",
                "url": "/contact"
            },
            "features": [
                "Free forever plan",
                "No credit card required",
                "Setup in 2 minutes"
            ]
        }
    }'::jsonb,
    'About Trunc | We Are Trunc',
    'Pioneering the next generation of URL management with cutting-edge technology, unparalleled performance, and a vision that connects the digital world.',
    'about trunc, url shortener company, link management',
    true,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    content_blocks = EXCLUDED.content_blocks,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    meta_keywords = EXCLUDED.meta_keywords,
    is_published = EXCLUDED.is_published,
    updated_at = NOW();

-- ============================================================================
-- CONTACT PAGE CONTENT
-- ============================================================================
INSERT INTO static_pages (
    slug,
    title,
    content,
    content_blocks,
    custom_fields,
    meta_title,
    meta_description,
    meta_keywords,
    is_published,
    created_at,
    updated_at
) VALUES (
    'contact',
    'Get in Touch',
    'Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc.',
    '{
        "hero": {
            "badge": "Contact Us",
            "badge_icon": "ChatBubbleLeftRightIcon",
            "title": "Get in Touch",
            "title_highlight": "Touch",
            "tagline": "Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc."
        },
        "contact_methods": {
            "badge": "Contact Options",
            "heading": "Choose How to Connect",
            "tagline": "Multiple ways to get the help you need, when you need it.",
            "methods": [
                {
                    "icon": "EnvelopeIcon",
                    "title": "Email Support",
                    "description": "Get detailed help via email",
                    "contact": "hello@urlshorter.com",
                    "response_time": "Within 24 hours",
                    "color": "text-primary-600"
                },
                {
                    "icon": "ChatBubbleLeftRightIcon",
                    "title": "Live Chat",
                    "description": "Instant help when you need it",
                    "contact": "Start Chat",
                    "response_time": "Immediate",
                    "color": "text-green-500"
                }
            ]
        },
        "contact_form": {
            "badge": "Contact Form",
            "heading": "Send us a Message",
            "tagline": "Fill out the form below and we''ll get back to you as soon as possible.",
            "form_title": "Let''s Talk",
            "form_description": "We''re here to help you succeed with Trunc."
        },
        "support_resources": {
            "badge": "Additional Resources",
            "heading": "Alternative Support Channels",
            "items": [
                {
                    "icon": "UserIcon",
                    "title": "Documentation",
                    "description": "Find answers in our comprehensive guides",
                    "cta": {
                        "text": "Browse Documentation",
                        "url": "/docs"
                    }
                }
            ]
        },
        "cta_bottom": {
            "badge": "🚀 Ready to Get Started?",
            "heading": "Start Your Journey with Trunc",
            "tagline": "Join thousands of businesses using Trunc to power their links. No credit card required to get started.",
            "cta_primary": {
                "text": "Start Free Trial",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "View Pricing",
                "url": "/pricing"
            }
        }
    }'::jsonb,
    '{
        "inquiry_types": [
            {"value": "general", "label": "General Inquiry"},
            {"value": "technical", "label": "Technical Support"},
            {"value": "billing", "label": "Billing Question"},
            {"value": "sales", "label": "Sales Inquiry"},
            {"value": "partnership", "label": "Partnership"},
            {"value": "feedback", "label": "Feedback"}
        ],
        "priority_levels": [
            {"value": "low", "label": "Low Priority"},
            {"value": "medium", "label": "Medium Priority"},
            {"value": "high", "label": "High Priority"},
            {"value": "urgent", "label": "Urgent"}
        ],
        "form_fields": [
            {
                "name": "name",
                "label": "Name",
                "type": "text",
                "required": true,
                "placeholder": "Your full name"
            },
            {
                "name": "email",
                "label": "Email",
                "type": "email",
                "required": true,
                "placeholder": "your@email.com"
            },
            {
                "name": "company",
                "label": "Company",
                "type": "text",
                "required": false,
                "placeholder": "Your company name (optional)"
            },
            {
                "name": "inquiryType",
                "label": "Inquiry Type",
                "type": "select",
                "required": true,
                "options": "inquiry_types"
            },
            {
                "name": "priority",
                "label": "Priority",
                "type": "select",
                "required": true,
                "options": "priority_levels"
            },
            {
                "name": "subject",
                "label": "Subject",
                "type": "text",
                "required": true,
                "placeholder": "Brief description of your inquiry"
            },
            {
                "name": "message",
                "label": "Message",
                "type": "textarea",
                "required": true,
                "placeholder": "Please provide details about your inquiry...",
                "rows": 5
            }
        ]
    }'::jsonb,
    'Contact Us | Get in Touch | Trunc',
    'Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc.',
    'contact trunc, support, help, get in touch',
    true,
    NOW(),
    NOW()
) ON CONFLICT (slug) DO UPDATE SET
    title = EXCLUDED.title,
    content = EXCLUDED.content,
    content_blocks = EXCLUDED.content_blocks,
    custom_fields = EXCLUDED.custom_fields,
    meta_title = EXCLUDED.meta_title,
    meta_description = EXCLUDED.meta_description,
    meta_keywords = EXCLUDED.meta_keywords,
    is_published = EXCLUDED.is_published,
    updated_at = NOW();

-- ============================================================================
-- REUSABLE CONTENT BLOCKS
-- ============================================================================

-- Hero Home
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'Homepage Hero Section',
    'hero_home',
    'hero',
    '{
        "badge": "✨ New: Advanced Analytics Dashboard",
        "title": "Shorten URLs. Track Everything.",
        "title_highlight": "Track Everything",
        "tagline": "Transform your long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more.",
        "cta_primary": {
            "text": "Get Started Free",
            "url": "/register"
        },
        "cta_secondary": {
            "text": "View Pricing",
            "url": "/pricing"
        }
    }'::jsonb,
    true,
    1,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Features Grid
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'Features Grid',
    'features_grid',
    'features',
    '{
        "badge": "Features",
        "heading": "Why Choose Trunc?",
        "tagline": "More than just a URL shortener. Get powerful features that help you understand and optimize your links.",
        "items": [
            {
                "name": "Lightning Fast",
                "description": "Shorten URLs instantly with our optimized infrastructure.",
                "icon": "LinkIcon"
            },
            {
                "name": "Advanced Analytics",
                "description": "Track clicks, locations, and referrers with detailed insights.",
                "icon": "ChartBarIcon"
            },
            {
                "name": "Secure & Reliable",
                "description": "Enterprise-grade security with 99.9% uptime guarantee.",
                "icon": "ShieldCheckIcon"
            }
        ]
    }'::jsonb,
    true,
    2,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Social Proof Stats
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'Social Proof Stats',
    'social_proof_stats',
    'stats',
    '{
        "heading": "Trusted by thousands of businesses worldwide",
        "stats": [
            {
                "value": "500K+",
                "label": "URLs shortened"
            },
            {
                "value": "50M+",
                "label": "Clicks tracked"
            },
            {
                "value": "99.9%",
                "label": "Uptime"
            }
        ]
    }'::jsonb,
    true,
    3,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Bottom CTA
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'Bottom CTA Section',
    'cta_bottom',
    'cta',
    '{
        "badge": "🚀 Get Started Today",
        "heading": "Ready to get started?",
        "tagline": "Join thousands of businesses using Trunc to track and optimize their links. Start for free today with no credit card required.",
        "cta_primary": {
            "text": "Get Started Free",
            "url": "/register"
        },
        "cta_secondary": {
            "text": "View Pricing",
            "url": "/pricing"
        },
        "features": [
            "Free forever plan",
            "No credit card required",
            "Setup in 2 minutes"
        ]
    }'::jsonb,
    true,
    4,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Hero About
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'About Page Hero Section',
    'hero_about',
    'hero',
    '{
        "badge": "About Trunc",
        "badge_icon": "BuildingOfficeIcon",
        "title": "We Are Trunc",
        "title_highlight": "Trunc",
        "tagline": "Pioneering the next generation of URL management with cutting-edge technology, unparalleled performance, and a vision that connects the digital world."
    }'::jsonb,
    true,
    5,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- Hero Contact
INSERT INTO content_blocks (
    name,
    identifier,
    type,
    content,
    is_active,
    sort_order,
    created_at,
    updated_at
) VALUES (
    'Contact Page Hero Section',
    'hero_contact',
    'hero',
    '{
        "badge": "Contact Us",
        "badge_icon": "ChatBubbleLeftRightIcon",
        "title": "Get in Touch",
        "title_highlight": "Touch",
        "tagline": "Ready to transform your link management? Our team of experts is here to answer your questions and help you unlock the full potential of Trunc."
    }'::jsonb,
    true,
    6,
    NOW(),
    NOW()
) ON CONFLICT (identifier) DO UPDATE SET
    name = EXCLUDED.name,
    content = EXCLUDED.content,
    is_active = EXCLUDED.is_active,
    updated_at = NOW();

-- ============================================================================
-- VERIFICATION
-- ============================================================================

-- Verify pages were created/updated
DO $$
DECLARE
    page_count INT;
    block_count INT;
BEGIN
    SELECT COUNT(*) INTO page_count FROM static_pages WHERE slug IN ('home', 'about', 'contact');
    SELECT COUNT(*) INTO block_count FROM content_blocks WHERE identifier IN ('hero_home', 'features_grid', 'social_proof_stats', 'cta_bottom', 'hero_about', 'hero_contact');

    RAISE NOTICE 'Migration 028 Complete:';
    RAISE NOTICE '  - Pages populated: % (expected: 3)', page_count;
    RAISE NOTICE '  - Content blocks created: % (expected: 6)', block_count;

    IF page_count < 3 THEN
        RAISE WARNING 'Expected 3 pages, got %', page_count;
    END IF;

    IF block_count < 6 THEN
        RAISE WARNING 'Expected 6 content blocks, got %', block_count;
    END IF;
END $$;
