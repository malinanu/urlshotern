-- Migration 029: Add Missing CMS Pages
-- Created: 2026-02-15
-- Purpose: Add features, pricing, privacy, and terms pages to CMS

-- ============================================================================
-- FEATURES PAGE CONTENT
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
    'features',
    'Powerful Features for Smart URL Management',
    'Everything you need to create, track, and optimize your short links. From basic URL shortening to advanced analytics and enterprise features.',
    '{
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
        "comparison": {
            "badge": "Free vs Pro",
            "heading": "Compare Plans",
            "tagline": "See what''s included in each plan",
            "free_features": [
                "100 links per month",
                "Basic analytics",
                "Standard support",
                "SSL encryption",
                "API access"
            ],
            "pro_features": [
                "Unlimited links",
                "Advanced analytics",
                "Priority support",
                "Custom domains",
                "Team collaboration",
                "White-label options",
                "SSO integration"
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
    'Features | Powerful URL Management Tools | Trunc',
    'Discover Trunc''s powerful features: lightning-fast redirects, advanced analytics, enterprise security, and more. Everything you need for smart URL management.',
    'features, url shortener, analytics, tracking, custom domains, API',
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
-- PRICING PAGE CONTENT
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
    'pricing',
    'Simple, Transparent Pricing',
    'Choose the plan that fits your needs. All plans include core features with the option to upgrade for advanced analytics and higher limits.',
    '{
        "hero": {
            "badge": "Pricing",
            "title": "Simple, Transparent Pricing",
            "title_highlight": "Transparent Pricing",
            "tagline": "Choose the plan that fits your needs. Upgrade anytime as your business grows."
        },
        "pricing_intro": {
            "heading": "Plans for Every Need",
            "description": "All plans include core URL shortening features. Upgrade to Pro for advanced analytics, custom domains, and enterprise features."
        },
        "feature_comparison": {
            "badge": "Compare Features",
            "heading": "What''s Included",
            "categories": [
                {
                    "name": "Core Features",
                    "features": [
                        {
                            "name": "URL Shortening",
                            "free": true,
                            "pro": true,
                            "enterprise": true
                        },
                        {
                            "name": "SSL Encryption",
                            "free": true,
                            "pro": true,
                            "enterprise": true
                        },
                        {
                            "name": "Basic Analytics",
                            "free": true,
                            "pro": true,
                            "enterprise": true
                        },
                        {
                            "name": "API Access",
                            "free": "Limited",
                            "pro": true,
                            "enterprise": true
                        }
                    ]
                },
                {
                    "name": "Advanced Features",
                    "features": [
                        {
                            "name": "Custom Domains",
                            "free": false,
                            "pro": true,
                            "enterprise": true
                        },
                        {
                            "name": "Advanced Analytics",
                            "free": false,
                            "pro": true,
                            "enterprise": true
                        },
                        {
                            "name": "Team Collaboration",
                            "free": false,
                            "pro": "Up to 5",
                            "enterprise": "Unlimited"
                        },
                        {
                            "name": "White-label",
                            "free": false,
                            "pro": false,
                            "enterprise": true
                        },
                        {
                            "name": "SSO Integration",
                            "free": false,
                            "pro": false,
                            "enterprise": true
                        }
                    ]
                },
                {
                    "name": "Support & SLA",
                    "features": [
                        {
                            "name": "Support Response Time",
                            "free": "48 hours",
                            "pro": "24 hours",
                            "enterprise": "4 hours"
                        },
                        {
                            "name": "Uptime SLA",
                            "free": "99.5%",
                            "pro": "99.9%",
                            "enterprise": "99.99%"
                        },
                        {
                            "name": "Dedicated Account Manager",
                            "free": false,
                            "pro": false,
                            "enterprise": true
                        }
                    ]
                }
            ]
        },
        "faq": {
            "badge": "FAQ",
            "heading": "Frequently Asked Questions",
            "items": [
                {
                    "question": "Can I change plans later?",
                    "answer": "Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and we''ll prorate any payments."
                },
                {
                    "question": "What payment methods do you accept?",
                    "answer": "We accept all major credit cards (Visa, MasterCard, American Express) and PayPal for monthly subscriptions."
                },
                {
                    "question": "Is there a free trial?",
                    "answer": "Yes! Our Free plan is available forever with no credit card required. Upgrade to Pro when you need advanced features."
                },
                {
                    "question": "What happens if I exceed my limits?",
                    "answer": "We''ll send you a notification when you''re approaching your limit. You can upgrade at any time to increase your capacity."
                },
                {
                    "question": "Do you offer refunds?",
                    "answer": "We offer a 30-day money-back guarantee for all paid plans. If you''re not satisfied, contact us for a full refund."
                },
                {
                    "question": "Can I get a custom enterprise plan?",
                    "answer": "Yes! Contact our sales team to discuss custom pricing and features tailored to your organization''s needs."
                }
            ]
        },
        "cta_bottom": {
            "badge": "🚀 Ready to Get Started?",
            "heading": "Start Shortening URLs Today",
            "tagline": "Join thousands of businesses using Trunc. No credit card required for the free plan.",
            "cta_primary": {
                "text": "Start Free",
                "url": "/register"
            },
            "cta_secondary": {
                "text": "Contact Sales",
                "url": "/contact"
            }
        }
    }'::jsonb,
    'Pricing | Simple & Transparent Plans | Trunc',
    'Choose the right plan for your needs. Free forever plan available. Upgrade for advanced analytics, custom domains, and enterprise features.',
    'pricing, plans, subscription, url shortener pricing, custom domains',
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
-- PRIVACY POLICY PAGE CONTENT
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
    'privacy',
    'Privacy Policy',
    'Your privacy is important to us. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our URL shortening service.',
    '{
        "hero": {
            "badge": "Legal",
            "title": "Privacy Policy",
            "tagline": "Last Updated: February 15, 2026",
            "description": "Your privacy is important to us. This Privacy Policy explains how we collect, use, disclose, and safeguard your information when you use our URL shortening service."
        },
        "table_of_contents": {
            "heading": "Table of Contents",
            "sections": [
                {"id": "information-collection", "title": "1. Information We Collect"},
                {"id": "information-use", "title": "2. How We Use Your Information"},
                {"id": "information-sharing", "title": "3. Sharing Your Information"},
                {"id": "data-security", "title": "4. Data Security"},
                {"id": "cookies", "title": "5. Cookies and Tracking"},
                {"id": "your-rights", "title": "6. Your Privacy Rights"},
                {"id": "data-retention", "title": "7. Data Retention"},
                {"id": "children", "title": "8. Children''s Privacy"},
                {"id": "changes", "title": "9. Changes to This Policy"},
                {"id": "contact", "title": "10. Contact Us"}
            ]
        },
        "sections": [
            {
                "id": "information-collection",
                "heading": "1. Information We Collect",
                "content": "We collect information that you provide directly to us, information we obtain automatically when you use our services, and information from third-party sources.",
                "subsections": [
                    {
                        "title": "Information You Provide",
                        "items": [
                            "Account information (name, email, password)",
                            "URLs you submit for shortening",
                            "Custom short codes you create",
                            "Payment information (processed securely by our payment providers)",
                            "Communications with our support team"
                        ]
                    },
                    {
                        "title": "Information Collected Automatically",
                        "items": [
                            "IP address and location data",
                            "Device information (browser, operating system)",
                            "Click data (timestamp, referrer, user agent)",
                            "Usage patterns and analytics data",
                            "Cookies and similar tracking technologies"
                        ]
                    }
                ]
            },
            {
                "id": "information-use",
                "heading": "2. How We Use Your Information",
                "content": "We use the information we collect to provide, maintain, and improve our services, as well as to protect our users and comply with legal obligations.",
                "purposes": [
                    "Provide and operate our URL shortening service",
                    "Process your transactions and send related information",
                    "Send administrative information, updates, and security alerts",
                    "Provide customer support and respond to your requests",
                    "Monitor and analyze trends, usage, and activities",
                    "Detect, prevent, and address technical issues and fraud",
                    "Improve and personalize our services",
                    "Comply with legal obligations and enforce our terms"
                ]
            },
            {
                "id": "information-sharing",
                "heading": "3. Sharing Your Information",
                "content": "We do not sell your personal information. We may share your information only in the following circumstances:",
                "circumstances": [
                    {
                        "title": "Service Providers",
                        "description": "We share data with third-party vendors who perform services on our behalf, such as hosting, analytics, and payment processing."
                    },
                    {
                        "title": "Legal Requirements",
                        "description": "We may disclose information if required by law, court order, or government request, or to protect rights and safety."
                    },
                    {
                        "title": "Business Transfers",
                        "description": "In connection with a merger, acquisition, or sale of assets, your information may be transferred to the acquiring entity."
                    },
                    {
                        "title": "With Your Consent",
                        "description": "We may share information with third parties when you explicitly consent to such sharing."
                    }
                ]
            },
            {
                "id": "data-security",
                "heading": "4. Data Security",
                "content": "We implement industry-standard security measures to protect your information from unauthorized access, alteration, disclosure, or destruction.",
                "measures": [
                    "SSL/TLS encryption for data in transit",
                    "Encryption of sensitive data at rest",
                    "Regular security audits and penetration testing",
                    "Access controls and authentication requirements",
                    "Employee training on data protection practices",
                    "Incident response procedures"
                ],
                "disclaimer": "However, no method of transmission over the Internet or electronic storage is 100% secure. We cannot guarantee absolute security."
            },
            {
                "id": "cookies",
                "heading": "5. Cookies and Tracking Technologies",
                "content": "We use cookies and similar tracking technologies to collect information about your browsing activities and to provide personalized experiences.",
                "types": [
                    {
                        "name": "Essential Cookies",
                        "description": "Required for the service to function properly (authentication, security)."
                    },
                    {
                        "name": "Analytics Cookies",
                        "description": "Help us understand how you use our service to improve user experience."
                    },
                    {
                        "name": "Preference Cookies",
                        "description": "Remember your settings and preferences."
                    }
                ],
                "control": "You can control cookies through your browser settings. Note that disabling certain cookies may limit functionality."
            },
            {
                "id": "your-rights",
                "heading": "6. Your Privacy Rights",
                "content": "Depending on your location, you may have certain rights regarding your personal information:",
                "rights": [
                    {
                        "title": "Access",
                        "description": "Request a copy of the personal information we hold about you."
                    },
                    {
                        "title": "Correction",
                        "description": "Request correction of inaccurate or incomplete information."
                    },
                    {
                        "title": "Deletion",
                        "description": "Request deletion of your personal information, subject to legal requirements."
                    },
                    {
                        "title": "Portability",
                        "description": "Request a copy of your data in a machine-readable format."
                    },
                    {
                        "title": "Opt-out",
                        "description": "Opt out of marketing communications at any time."
                    },
                    {
                        "title": "Restriction",
                        "description": "Request restriction of processing in certain circumstances."
                    }
                ],
                "exercise": "To exercise these rights, please contact us at privacy@urlshorter.com"
            },
            {
                "id": "data-retention",
                "heading": "7. Data Retention",
                "content": "We retain your information for as long as necessary to provide our services and comply with legal obligations.",
                "policies": [
                    "Account data: Retained while your account is active and for a reasonable period after deletion",
                    "Analytics data: Aggregated data may be retained indefinitely",
                    "Legal compliance: Some data may be retained longer to comply with legal requirements",
                    "Security logs: Retained for security and fraud prevention purposes"
                ]
            },
            {
                "id": "children",
                "heading": "8. Children''s Privacy",
                "content": "Our service is not intended for users under the age of 13. We do not knowingly collect personal information from children under 13. If we become aware that we have collected personal information from a child under 13, we will take steps to delete such information."
            },
            {
                "id": "changes",
                "heading": "9. Changes to This Privacy Policy",
                "content": "We may update this Privacy Policy from time to time. We will notify you of any material changes by posting the new policy on this page and updating the \"Last Updated\" date. Your continued use of our service after changes constitutes acceptance of the updated policy."
            },
            {
                "id": "contact",
                "heading": "10. Contact Us",
                "content": "If you have questions or concerns about this Privacy Policy or our data practices, please contact us:",
                "contact_info": {
                    "email": "privacy@urlshorter.com",
                    "address": "Trunc URL Shortener\nAttn: Privacy Officer\n[Your Company Address]",
                    "response_time": "We will respond to your inquiry within 30 days."
                }
            }
        ]
    }'::jsonb,
    'Privacy Policy | Trunc URL Shortener',
    'Learn how Trunc collects, uses, and protects your personal information. Our Privacy Policy explains your rights and our data practices.',
    'privacy policy, data protection, privacy rights, GDPR, data security',
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
-- TERMS OF SERVICE PAGE CONTENT
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
    'terms',
    'Terms of Service',
    'Please read these Terms of Service carefully before using our URL shortening service. By accessing or using our service, you agree to be bound by these terms.',
    '{
        "hero": {
            "badge": "Legal",
            "title": "Terms of Service",
            "tagline": "Last Updated: February 15, 2026",
            "description": "Please read these Terms of Service carefully before using our URL shortening service. By accessing or using our service, you agree to be bound by these terms."
        },
        "table_of_contents": {
            "heading": "Table of Contents",
            "sections": [
                {"id": "acceptance", "title": "1. Acceptance of Terms"},
                {"id": "service-description", "title": "2. Description of Service"},
                {"id": "user-accounts", "title": "3. User Accounts and Registration"},
                {"id": "acceptable-use", "title": "4. Acceptable Use Policy"},
                {"id": "prohibited-uses", "title": "5. Prohibited Uses"},
                {"id": "intellectual-property", "title": "6. Intellectual Property Rights"},
                {"id": "user-content", "title": "7. User-Generated Content"},
                {"id": "payment-terms", "title": "8. Payment and Billing"},
                {"id": "termination", "title": "9. Termination and Suspension"},
                {"id": "disclaimers", "title": "10. Disclaimers and Limitations"},
                {"id": "indemnification", "title": "11. Indemnification"},
                {"id": "governing-law", "title": "12. Governing Law"},
                {"id": "contact", "title": "13. Contact Information"}
            ]
        },
        "sections": [
            {
                "id": "acceptance",
                "heading": "1. Acceptance of Terms",
                "content": "By accessing or using the Trunc URL shortening service, you agree to be bound by these Terms of Service and all applicable laws and regulations. If you do not agree with any of these terms, you are prohibited from using or accessing this service.",
                "key_points": [
                    "These terms constitute a legally binding agreement between you and Trunc",
                    "You must be at least 13 years old to use this service",
                    "By creating an account, you confirm your acceptance of these terms",
                    "Continued use of the service constitutes ongoing acceptance"
                ]
            },
            {
                "id": "service-description",
                "heading": "2. Description of Service",
                "content": "Trunc provides a URL shortening service that allows users to create shortened versions of long URLs for easier sharing and tracking.",
                "features": [
                    "URL shortening and custom short code creation",
                    "Click analytics and tracking",
                    "Link management dashboard",
                    "API access for automated URL creation",
                    "QR code generation",
                    "Custom domain support (Pro and Enterprise plans)"
                ],
                "disclaimer": "We reserve the right to modify, suspend, or discontinue any aspect of the service at any time without notice."
            },
            {
                "id": "user-accounts",
                "heading": "3. User Accounts and Registration",
                "content": "To access certain features of our service, you must create an account and provide accurate, complete information.",
                "requirements": [
                    {
                        "title": "Account Information",
                        "description": "You must provide accurate and complete registration information and keep it up to date."
                    },
                    {
                        "title": "Account Security",
                        "description": "You are responsible for maintaining the confidentiality of your password and account credentials."
                    },
                    {
                        "title": "Account Responsibility",
                        "description": "You are responsible for all activities that occur under your account."
                    },
                    {
                        "title": "One Account Per User",
                        "description": "You may not create multiple accounts to circumvent our limits or restrictions."
                    }
                ],
                "notice": "You must notify us immediately of any unauthorized use of your account."
            },
            {
                "id": "acceptable-use",
                "heading": "4. Acceptable Use Policy",
                "content": "You agree to use our service only for lawful purposes and in accordance with these Terms of Service.",
                "guidelines": [
                    "Comply with all applicable local, state, national, and international laws",
                    "Respect the intellectual property rights of others",
                    "Use the service only for legitimate URL shortening purposes",
                    "Provide accurate information when creating shortened URLs",
                    "Respect rate limits and usage quotas for your plan"
                ]
            },
            {
                "id": "prohibited-uses",
                "heading": "5. Prohibited Uses",
                "content": "You may not use our service for any of the following purposes:",
                "prohibited": [
                    {
                        "category": "Illegal Content",
                        "examples": [
                            "Links to illegal content, pirated materials, or copyright infringement",
                            "Child sexual abuse material or exploitation",
                            "Terrorist content or recruitment materials",
                            "Drug trafficking or illegal substances"
                        ]
                    },
                    {
                        "category": "Malicious Activity",
                        "examples": [
                            "Malware, viruses, or malicious code distribution",
                            "Phishing schemes or social engineering attacks",
                            "Spam, unsolicited commercial messages, or chain letters",
                            "DDoS attacks or attempts to disrupt service availability"
                        ]
                    },
                    {
                        "category": "Harmful Content",
                        "examples": [
                            "Harassment, threats, or incitement to violence",
                            "Hate speech or content promoting discrimination",
                            "Non-consensual intimate images or revenge porn",
                            "Content that violates privacy rights"
                        ]
                    },
                    {
                        "category": "Fraudulent Activity",
                        "examples": [
                            "Scams, fraud, or deceptive practices",
                            "Impersonation or identity theft",
                            "Financial fraud or investment scams",
                            "Manipulation of metrics or analytics"
                        ]
                    }
                ],
                "enforcement": "We reserve the right to remove any shortened URL that violates these terms and to suspend or terminate accounts that engage in prohibited activities."
            },
            {
                "id": "intellectual-property",
                "heading": "6. Intellectual Property Rights",
                "content": "The service and its original content, features, and functionality are owned by Trunc and are protected by international copyright, trademark, patent, trade secret, and other intellectual property laws.",
                "rights": [
                    "All trademarks, logos, and service marks displayed on the service are our property or the property of their respective owners",
                    "You may not use our trademarks without prior written consent",
                    "The software, design, text, graphics, and other content are protected by copyright",
                    "You may not copy, modify, distribute, sell, or lease any part of our service without permission"
                ]
            },
            {
                "id": "user-content",
                "heading": "7. User-Generated Content",
                "content": "You retain all rights to the URLs and destination links you submit to our service. However, you grant us certain licenses to operate the service.",
                "license": {
                    "grant": "By submitting URLs to our service, you grant us a worldwide, non-exclusive, royalty-free license to use, reproduce, and display the URLs for the purpose of providing the service.",
                    "scope": "This license is limited to what is necessary to operate and improve our service.",
                    "termination": "The license terminates when you delete the shortened URL from our service."
                },
                "responsibility": "You are solely responsible for the content of any URLs you submit and any consequences arising from their use."
            },
            {
                "id": "payment-terms",
                "heading": "8. Payment and Billing",
                "content": "Certain features of our service require payment of fees. By subscribing to a paid plan, you agree to the following terms:",
                "terms": [
                    {
                        "title": "Subscription Fees",
                        "description": "Fees are billed in advance on a monthly or annual basis, depending on your chosen plan."
                    },
                    {
                        "title": "Auto-Renewal",
                        "description": "Subscriptions automatically renew unless you cancel before the renewal date."
                    },
                    {
                        "title": "Price Changes",
                        "description": "We may change our fees with 30 days'' notice. Continued use after a price change constitutes acceptance."
                    },
                    {
                        "title": "Refund Policy",
                        "description": "We offer a 30-day money-back guarantee for new subscriptions. Refunds after 30 days are at our discretion."
                    },
                    {
                        "title": "Cancellation",
                        "description": "You may cancel your subscription at any time. Cancellations take effect at the end of the current billing period."
                    }
                ],
                "payment_methods": "We accept major credit cards and PayPal. All payments are processed securely through our payment partners."
            },
            {
                "id": "termination",
                "heading": "9. Termination and Suspension",
                "content": "We reserve the right to terminate or suspend your account and access to the service at our sole discretion, without notice, for conduct that we believe violates these Terms of Service or is harmful to other users, us, or third parties, or for any other reason.",
                "grounds": [
                    "Violation of these Terms of Service or our Acceptable Use Policy",
                    "Fraudulent, illegal, or harmful activity",
                    "Non-payment of fees for paid plans",
                    "Inactivity for an extended period",
                    "At your request"
                ],
                "consequences": [
                    "Upon termination, your right to use the service will immediately cease",
                    "We may delete your account data, shortened URLs, and analytics",
                    "Shortened URLs may stop working immediately",
                    "No refunds will be provided for paid subscriptions upon termination for cause"
                ],
                "data_retention": "We may retain certain data for legal compliance, dispute resolution, and service improvement purposes."
            },
            {
                "id": "disclaimers",
                "heading": "10. Disclaimers and Limitations of Liability",
                "content": "Our service is provided on an \"as is\" and \"as available\" basis without warranties of any kind.",
                "disclaimers": [
                    "We do not warrant that the service will be uninterrupted, secure, or error-free",
                    "We do not guarantee the accuracy, completeness, or reliability of any content or analytics",
                    "We are not responsible for the content of destination URLs or any consequences of using shortened links",
                    "We do not warrant that our service will meet your specific requirements"
                ],
                "limitation": "To the maximum extent permitted by law, Trunc and its affiliates shall not be liable for any indirect, incidental, special, consequential, or punitive damages, or any loss of profits or revenues, whether incurred directly or indirectly, or any loss of data, use, goodwill, or other intangible losses resulting from your use of the service.",
                "cap": "Our total liability to you for all claims arising from your use of the service shall not exceed the amount you paid us in the 12 months prior to the event giving rise to the liability."
            },
            {
                "id": "indemnification",
                "heading": "11. Indemnification",
                "content": "You agree to indemnify, defend, and hold harmless Trunc, its officers, directors, employees, and agents from and against any claims, liabilities, damages, losses, and expenses, including reasonable legal fees, arising out of or in any way connected with:",
                "obligations": [
                    "Your access to or use of the service",
                    "Your violation of these Terms of Service",
                    "Your violation of any third-party rights, including intellectual property rights",
                    "Any URLs you submit to the service and their destination content",
                    "Any harm caused by your shortened URLs"
                ]
            },
            {
                "id": "governing-law",
                "heading": "12. Governing Law and Dispute Resolution",
                "content": "These Terms of Service shall be governed by and construed in accordance with the laws of [Your Jurisdiction], without regard to its conflict of law provisions.",
                "dispute_resolution": {
                    "informal": "We encourage you to contact us first to seek informal resolution of any disputes.",
                    "arbitration": "Any disputes that cannot be resolved informally shall be resolved through binding arbitration in accordance with the rules of [Arbitration Organization].",
                    "exceptions": "Either party may seek injunctive relief in court to prevent irreparable harm.",
                    "class_action_waiver": "You agree to resolve disputes on an individual basis and waive the right to participate in class actions."
                }
            },
            {
                "id": "contact",
                "heading": "13. Contact Information",
                "content": "If you have any questions about these Terms of Service, please contact us:",
                "contact_info": {
                    "email": "legal@urlshorter.com",
                    "support_email": "support@urlshorter.com",
                    "address": "Trunc URL Shortener\nAttn: Legal Department\n[Your Company Address]"
                }
            }
        ],
        "miscellaneous": {
            "heading": "Additional Provisions",
            "provisions": [
                {
                    "title": "Entire Agreement",
                    "description": "These Terms of Service constitute the entire agreement between you and Trunc regarding the service."
                },
                {
                    "title": "Severability",
                    "description": "If any provision of these terms is found to be invalid, the remaining provisions will remain in effect."
                },
                {
                    "title": "Waiver",
                    "description": "Our failure to enforce any right or provision of these terms will not be considered a waiver."
                },
                {
                    "title": "Assignment",
                    "description": "You may not assign or transfer these terms without our consent. We may assign these terms at any time."
                },
                {
                    "title": "Changes to Terms",
                    "description": "We reserve the right to modify these terms at any time. Material changes will be notified via email or service notification."
                }
            ]
        }
    }'::jsonb,
    'Terms of Service | Trunc URL Shortener',
    'Read our Terms of Service to understand the rules and guidelines for using Trunc''s URL shortening service.',
    'terms of service, terms and conditions, user agreement, legal terms',
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
-- VERIFICATION
-- ============================================================================

-- Verify pages were created/updated
DO $$
DECLARE
    page_count INT;
    new_pages INT;
    total_pages INT;
BEGIN
    SELECT COUNT(*) INTO new_pages FROM static_pages
    WHERE slug IN ('features', 'pricing', 'privacy', 'terms');

    SELECT COUNT(*) INTO total_pages FROM static_pages;

    RAISE NOTICE 'Migration 029 Complete:';
    RAISE NOTICE '  - New pages created/updated: % (expected: 4)', new_pages;
    RAISE NOTICE '  - Total pages in database: % (expected: 7)', total_pages;

    IF new_pages < 4 THEN
        RAISE WARNING 'Expected 4 new pages, got %', new_pages;
    END IF;

    IF total_pages < 7 THEN
        RAISE WARNING 'Expected at least 7 total pages, got %', total_pages;
    END IF;
END $$;
