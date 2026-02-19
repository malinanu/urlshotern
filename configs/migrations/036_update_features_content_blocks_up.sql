-- Migration 036: Update Features Page Content Blocks
-- Created: 2026-02-20
-- Purpose: Ensure all 4 content block sections (hero, main_features, additional_features, cta_bottom)
--          are present in the features page. Migration 029 may not have applied completely.

UPDATE static_pages
SET content_blocks = '{
    "hero": {
        "badge": "Features",
        "title": "Powerful Features for Modern Link Management",
        "title_highlight": "Link Management",
        "tagline": "Everything you need to shorten, track, and optimize your links with enterprise-grade tools and insights."
    },
    "main_features": {
        "badge": "Core Features",
        "heading": "Built for Performance, Designed for Insights",
        "tagline": "Powerful capabilities that set Trunc apart from other URL shorteners.",
        "items": [
            {
                "name": "Lightning Fast Redirects",
                "description": "Sub-50ms redirect times with global CDN distribution and Redis caching for instant link resolution.",
                "icon": "LinkIcon",
                "details": [
                    "Sub-second response times",
                    "Global CDN distribution",
                    "Auto-scaling infrastructure",
                    "Bulk URL processing"
                ]
            },
            {
                "name": "Advanced Analytics",
                "description": "Real-time click tracking with detailed insights including geographic data, device information, and referrer analysis.",
                "icon": "ChartBarIcon",
                "details": [
                    "Real-time click tracking",
                    "Geographic click mapping",
                    "Device and browser analytics",
                    "Referrer source tracking"
                ]
            },
            {
                "name": "Enterprise Security",
                "description": "Bank-level security with SSL encryption, spam protection, and enterprise-grade infrastructure.",
                "icon": "ShieldCheckIcon",
                "details": [
                    "SSL/TLS encryption",
                    "Malware URL scanning",
                    "Spam link detection",
                    "99.9% uptime SLA"
                ]
            },
            {
                "name": "Smart Click Insights",
                "description": "Understand your audience better with detailed click analytics and user behavior patterns.",
                "icon": "CursorArrowRaysIcon",
                "details": [
                    "Click heatmaps",
                    "Time-based analytics",
                    "Conversion tracking",
                    "A/B testing support"
                ]
            },
            {
                "name": "Global Performance",
                "description": "Worldwide infrastructure ensures your short links work fast everywhere with multi-region redundancy.",
                "icon": "GlobeAltIcon",
                "details": [
                    "Multi-region deployment",
                    "Edge caching worldwide",
                    "Automatic failover",
                    "Load balancing"
                ]
            },
            {
                "name": "Real-time Monitoring",
                "description": "Monitor your links in real-time with instant notifications and comprehensive health checks.",
                "icon": "ClockIcon",
                "details": [
                    "Live click monitoring",
                    "Alert notifications",
                    "Health check endpoints",
                    "Performance metrics"
                ]
            }
        ]
    },
    "additional_features": {
        "badge": "Additional Features",
        "heading": "Even More Capabilities",
        "tagline": "Additional tools and capabilities to enhance your URL management experience.",
        "items": [
            {"name": "Custom Short Codes", "icon": "CheckCircleIcon"},
            {"name": "QR Code Generation", "icon": "CheckCircleIcon"},
            {"name": "Link Expiration", "icon": "CheckCircleIcon"},
            {"name": "Password Protection", "icon": "CheckCircleIcon"},
            {"name": "Team Collaboration", "icon": "CheckCircleIcon"},
            {"name": "API Access", "icon": "CheckCircleIcon"},
            {"name": "Webhook Integration", "icon": "CheckCircleIcon"},
            {"name": "White-label Solution", "icon": "CheckCircleIcon"},
            {"name": "Mobile Apps", "icon": "CheckCircleIcon"},
            {"name": "24/7 Support", "icon": "CheckCircleIcon"},
            {"name": "Data Export", "icon": "CheckCircleIcon"},
            {"name": "SSO Integration", "icon": "CheckCircleIcon"}
        ]
    },
    "cta_bottom": {
        "badge": "🚀 Ready to Get Started?",
        "heading": "Experience These Features Today",
        "tagline": "Start using Trunc and unlock the full potential of your links. No credit card required.",
        "cta_primary": {
            "text": "Try It Now",
            "url": "/register"
        },
        "cta_secondary": {
            "text": "View Pricing",
            "url": "/pricing"
        }
    }
}'::jsonb,
    updated_at = NOW()
WHERE slug = 'features';
