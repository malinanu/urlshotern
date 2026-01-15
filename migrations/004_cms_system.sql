-- Migration: CMS System Tables
-- Description: Creates all tables needed for the Content Management System
-- Version: 004
-- Date: 2026-01-13

-- =============================================================================
-- STATIC PAGES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS static_pages (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) NOT NULL UNIQUE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    content_blocks TEXT, -- JSON for structured content
    meta_description TEXT,
    meta_keywords TEXT,
    meta_image TEXT,
    meta_title VARCHAR(200),
    canonical_url TEXT,
    schema_markup TEXT, -- JSON-LD schema markup
    is_published BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0,
    template VARCHAR(50) NOT NULL DEFAULT 'default',
    author_id BIGINT, -- References users table
    published_at TIMESTAMP,
    scheduled_at TIMESTAMP,
    expires_at TIMESTAMP,
    featured_image_id BIGINT, -- References media_files table
    category VARCHAR(50),
    tags TEXT, -- JSON array
    custom_fields TEXT, -- JSON for additional fields
    view_count BIGINT NOT NULL DEFAULT 0,
    is_sticky BOOLEAN NOT NULL DEFAULT false,
    allow_comments BOOLEAN NOT NULL DEFAULT false,
    language VARCHAR(10) NOT NULL DEFAULT 'en',
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_static_pages_slug ON static_pages(slug);
CREATE INDEX idx_static_pages_published ON static_pages(is_published, published_at);
CREATE INDEX idx_static_pages_category ON static_pages(category);
CREATE INDEX idx_static_pages_author ON static_pages(author_id);

-- =============================================================================
-- PAGE REVISIONS TABLE (Version Control)
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_revisions (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    meta_description TEXT,
    meta_keywords TEXT,
    revision_number INTEGER NOT NULL,
    author_id BIGINT,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_revisions_page ON page_revisions(page_id, revision_number DESC);

-- =============================================================================
-- PAGE VISITS TABLE (Analytics)
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_visits (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    ip_address VARCHAR(45),
    user_agent TEXT,
    referrer TEXT,
    country_code VARCHAR(2),
    visited_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_visits_page ON page_visits(page_id, visited_at DESC);
CREATE INDEX idx_page_visits_country ON page_visits(country_code);

-- =============================================================================
-- MEDIA FILES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS media_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path TEXT NOT NULL,
    file_url TEXT NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_type VARCHAR(50) NOT NULL, -- image, document, video, etc.
    width INTEGER,
    height INTEGER,
    alt TEXT,
    caption TEXT,
    is_public BOOLEAN NOT NULL DEFAULT true,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_media_files_user ON media_files(user_id);
CREATE INDEX idx_media_files_type ON media_files(file_type);

-- =============================================================================
-- GLOBAL SETTINGS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS global_settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(100) NOT NULL UNIQUE,
    value TEXT NOT NULL,
    type VARCHAR(20) NOT NULL, -- text, html, json, boolean, number
    category VARCHAR(50) NOT NULL,
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN NOT NULL DEFAULT false,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_global_settings_category ON global_settings(category);
CREATE INDEX idx_global_settings_public ON global_settings(is_public);

-- =============================================================================
-- CONTENT BLOCKS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS content_blocks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    identifier VARCHAR(100) NOT NULL UNIQUE,
    content TEXT NOT NULL,
    type VARCHAR(50) NOT NULL, -- hero, testimonial, cta, footer, etc.
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_content_blocks_type ON content_blocks(type);
CREATE INDEX idx_content_blocks_active ON content_blocks(is_active);

-- =============================================================================
-- PAGE TEMPLATES TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS page_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) NOT NULL UNIQUE,
    description TEXT,
    content TEXT NOT NULL, -- Template HTML with placeholders
    fields TEXT, -- JSON defining custom fields
    preview_url TEXT,
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_page_templates_active ON page_templates(is_active);

-- =============================================================================
-- NAVIGATION MENUS TABLE
-- =============================================================================
CREATE TABLE IF NOT EXISTS navigation_menus (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(50) NOT NULL, -- header, footer, sidebar, etc.
    items TEXT NOT NULL, -- JSON array of menu items
    is_active BOOLEAN NOT NULL DEFAULT true,
    sort_order INTEGER NOT NULL DEFAULT 0,
    created_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP,
    updated_at TIMESTAMP NOT NULL DEFAULT CURRENT_TIMESTAMP
);

CREATE INDEX idx_navigation_menus_location ON navigation_menus(location);

-- =============================================================================
-- INSERT DEFAULT DATA
-- =============================================================================

-- Insert existing pages from the frontend
INSERT INTO static_pages (slug, title, content, meta_description, meta_title, is_published, is_sticky, template, language, sort_order) VALUES
('home', 'Home', 'Shorten URLs. Track Everything.

Transform your long URLs into short, memorable links. Track analytics, manage custom domains, and boost your link marketing strategy.

Key Features

🚀 Fast and reliable URL shortening
📊 Comprehensive analytics and tracking
🎯 Custom branded short links
🔒 Secure and private
⚡ Real-time click tracking
🌍 Geographic analytics', 'Create short URLs with advanced analytics. Track clicks, analyze traffic, and optimize your link marketing strategy with our powerful URL shortener.', 'URL Shortener - Fast, Secure & Analytics-Powered', true, true, 'default', 'en', 0),
('features', 'Features', 'Lightning fast URL shortening with advanced analytics and custom domains.', 'Explore powerful features including advanced analytics, custom domains, QR codes, and more.', NULL, true, false, 'default', 'en', 1),
('pricing', 'Pricing', 'Choose the perfect plan for your needs.', 'Flexible pricing plans for individuals and businesses.', NULL, true, false, 'default', 'en', 2),
('api-docs', 'API Documentation', 'Complete API reference for developers.', 'Developer-friendly API documentation with code examples.', NULL, true, false, 'default', 'en', 3),
('analytics', 'Analytics', 'Comprehensive analytics for your shortened URLs.', 'Track clicks, geographic data, devices, and more with detailed analytics.', NULL, true, false, 'default', 'en', 4),
('about', 'About Us', 'Learn more about our mission and team.', 'About our URL shortening service and our commitment to privacy and performance.', NULL, true, false, 'default', 'en', 5),
('contact', 'Contact', 'Get in touch with our team.', 'Contact us for support, partnerships, or general inquiries.', NULL, true, false, 'default', 'en', 6),
('privacy', 'Privacy Policy', '<div class="space-y-8"><div class="text-center border-b border-gray-200 pb-8"><h1 class="text-4xl font-bold text-gray-900 mb-4">Privacy Policy</h1><p class="text-lg text-gray-600 max-w-2xl mx-auto">Your privacy is important to us. This Privacy Policy explains how Trunc collects, uses, and protects your information when you use our URL shortener service.</p><p class="text-sm text-gray-500 mt-4"><strong>Last updated:</strong> January 16, 2026</p></div><div class="space-y-10"><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">1. Information We Collect</h2><h3 class="text-xl font-medium text-gray-800 mb-3">1.1 Information You Provide</h3><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-6"><li>Account information (name, email address, password)</li><li>URLs you submit for shortening</li><li>Custom aliases you create</li><li>Contact form submissions and support communications</li><li>Payment information (processed securely by our payment providers)</li></ul><h3 class="text-xl font-medium text-gray-800 mb-3">1.2 Information We Collect Automatically</h3><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>IP address and geolocation data</li><li>Browser type and version</li><li>Device information and operating system</li><li>Referring website and pages visited</li><li>Click data and analytics for your short links</li><li>Usage patterns and feature interactions</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">2. How We Use Your Information</h2><p class="text-gray-700 mb-4">We use the information we collect to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Provide and maintain our URL shortening service</li><li>Generate analytics and insights for your links</li><li>Prevent fraud and abuse of our service</li><li>Communicate with you about your account and our services</li><li>Improve our service and develop new features</li><li>Comply with legal obligations and enforce our terms</li><li>Process payments and manage billing</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">3. Information Sharing and Disclosure</h2><h3 class="text-xl font-medium text-gray-800 mb-3">We do NOT sell your personal information.</h3><p class="text-gray-700 mb-4">We may share your information in the following circumstances:</p><ul class="list-disc pl-6 space-y-3 text-gray-700"><li><strong>Service Providers:</strong> With trusted third-party service providers who assist us in operating our service (cloud hosting, analytics, payment processing)</li><li><strong>Legal Requirements:</strong> When required by law, court order, or government request</li><li><strong>Safety and Security:</strong> To protect the rights, property, or safety of Trunc, our users, or the public</li><li><strong>Business Transfers:</strong> In connection with a merger, acquisition, or sale of assets</li><li><strong>With Consent:</strong> When you explicitly consent to sharing your information</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">4. Data Security</h2><p class="text-gray-700 mb-4">We implement industry-standard security measures to protect your information:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>SSL/TLS encryption for data transmission</li><li>Encrypted storage of sensitive information</li><li>Regular security audits and monitoring</li><li>Access controls and authentication systems</li><li>Secure backup and disaster recovery procedures</li></ul><div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mt-4"><p class="text-blue-800"><strong>Note:</strong> While we implement strong security measures, no method of transmission over the Internet is 100% secure. We cannot guarantee absolute security of your information.</p></div></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">5. Your Privacy Rights</h2><p class="text-gray-700 mb-4">Depending on your location, you may have the following rights:</p><div class="grid md:grid-cols-2 gap-6"><div><h4 class="font-semibold text-gray-800 mb-2">Access and Portability</h4><p class="text-gray-600 text-sm">Request a copy of your personal information and data export</p></div><div><h4 class="font-semibold text-gray-800 mb-2">Correction</h4><p class="text-gray-600 text-sm">Update or correct inaccurate personal information</p></div><div><h4 class="font-semibold text-gray-800 mb-2">Deletion</h4><p class="text-gray-600 text-sm">Request deletion of your personal information</p></div><div><h4 class="font-semibold text-gray-800 mb-2">Opt-out</h4><p class="text-gray-600 text-sm">Opt out of certain data processing activities</p></div></div><p class="text-gray-700 mt-6">To exercise these rights, please contact us at <a href="mailto:privacy@urlshorter.com" class="text-blue-600 hover:underline">privacy@urlshorter.com</a>. We will respond to your request within 30 days.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">6. Cookies and Tracking</h2><p class="text-gray-700 mb-4">We use cookies and similar technologies to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4"><li>Maintain your login session</li><li>Remember your preferences and settings</li><li>Analyze usage patterns and improve our service</li><li>Provide personalized content and features</li></ul><p class="text-gray-700">You can control cookie settings through your browser preferences. However, disabling cookies may affect the functionality of our service.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">7. Data Retention</h2><p class="text-gray-700 mb-4">We retain your information for as long as necessary to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Provide our services to you</li><li>Comply with legal obligations</li><li>Resolve disputes and enforce our agreements</li><li>Maintain business records and analytics</li></ul><p class="text-gray-700 mt-4">When you delete your account, we will delete or anonymize your personal information within 90 days, except where we are required to retain it for legal purposes.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">8. International Data Transfers</h2><p class="text-gray-700 mb-4">Trunc is based in the United States. If you are accessing our service from outside the U.S., your information may be transferred to and processed in the United States and other countries where our service providers operate.</p><p class="text-gray-700">We ensure appropriate safeguards are in place for international data transfers in compliance with applicable privacy laws.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">9. Children''s Privacy</h2><p class="text-gray-700">Our service is not intended for children under 13 years of age. We do not knowingly collect personal information from children under 13. If we become aware that we have collected personal information from a child under 13, we will take steps to delete such information promptly.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">10. Changes to This Privacy Policy</h2><p class="text-gray-700 mb-4">We may update this Privacy Policy from time to time. We will notify you of any material changes by:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Posting the updated policy on our website</li><li>Sending an email notification to your registered email address</li><li>Displaying a notice in our service</li></ul><p class="text-gray-700 mt-4">Your continued use of our service after any changes indicates your acceptance of the updated Privacy Policy.</p></section><section class="bg-gray-50 p-6 rounded-lg"><h2 class="text-2xl font-semibold text-gray-900 mb-4">11. Contact Us</h2><p class="text-gray-700 mb-4">If you have any questions, concerns, or requests regarding this Privacy Policy or our privacy practices, please contact us:</p><div class="space-y-2 text-gray-700"><p><strong>Email:</strong> <a href="mailto:privacy@urlshorter.com" class="text-blue-600 hover:underline">privacy@urlshorter.com</a></p><p><strong>Address:</strong> Trunc Technologies, 123 Tech Boulevard, San Francisco, CA 94105</p><p><strong>Phone:</strong> <a href="tel:+1-555-0123" class="text-blue-600 hover:underline">+1 (555) 012-3456</a></p></div></section></div></div>', 'Read our privacy policy to learn how we protect and handle your data.', NULL, true, false, 'default', 'en', 7),
('terms', 'Terms of Service', '<div class="space-y-8"><div class="text-center border-b border-gray-200 pb-8"><h1 class="text-4xl font-bold text-gray-900 mb-4">Terms of Service</h1><p class="text-lg text-gray-600 max-w-2xl mx-auto">Please read these Terms of Service carefully before using Trunc. By accessing or using our service, you agree to be bound by these terms.</p><p class="text-sm text-gray-500 mt-4"><strong>Last updated:</strong> January 16, 2026</p></div><div class="space-y-10"><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">1. Acceptance of Terms</h2><p class="text-gray-700 mb-4">By accessing and using Trunc ("Service"), you accept and agree to be bound by the terms and provisions of this agreement. If you do not agree to these Terms of Service, please do not use our Service.</p><p class="text-gray-700">These terms apply to all users of the Service, including without limitation users who are browsers, vendors, customers, merchants, and/or contributors of content.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">2. Description of Service</h2><p class="text-gray-700 mb-4">Trunc provides a URL shortening service that allows users to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Convert long URLs into shorter, more manageable links</li><li>Track analytics and statistics for shortened URLs</li><li>Customize short URLs with custom aliases</li><li>Access API endpoints for programmatic URL shortening</li><li>Manage and organize shortened links through a dashboard</li><li>Monitor click data, geographic information, and referrer sources</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">3. User Accounts and Registration</h2><h3 class="text-xl font-medium text-gray-800 mb-3">3.1 Account Creation</h3><p class="text-gray-700 mb-4">To access certain features of our Service, you must register for an account. When you register, you agree to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-6"><li>Provide accurate, current, and complete information</li><li>Maintain and promptly update your account information</li><li>Maintain the security of your password and account</li><li>Accept responsibility for all activities under your account</li><li>Notify us immediately of any unauthorized use of your account</li></ul><h3 class="text-xl font-medium text-gray-800 mb-3">3.2 Account Eligibility</h3><p class="text-gray-700">You must be at least 13 years old to use our Service. By registering, you represent and warrant that you meet this age requirement and have the legal capacity to enter into these terms.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">4. Acceptable Use Policy</h2><p class="text-gray-700 mb-4">You agree to use Trunc only for lawful purposes and in a way that does not infringe upon the rights of others or restrict their use and enjoyment of the Service.</p><div class="bg-yellow-50 border border-yellow-200 rounded-lg p-4"><p class="text-yellow-900 font-semibold mb-2">You must NOT use our Service to:</p><ul class="list-disc pl-6 space-y-2 text-yellow-900"><li>Shorten URLs that link to illegal content, malware, or phishing sites</li><li>Distribute spam, viruses, or any malicious code</li><li>Violate any applicable laws or regulations</li><li>Infringe upon intellectual property rights of others</li><li>Harass, threaten, or harm others</li><li>Engage in fraudulent or deceptive practices</li><li>Bypass our security measures or abuse our service infrastructure</li></ul></div></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">5. Prohibited Activities</h2><p class="text-gray-700 mb-4">The following activities are strictly prohibited:</p><div class="grid md:grid-cols-2 gap-4"><div class="bg-red-50 border border-red-200 rounded-lg p-4"><h4 class="font-semibold text-red-900 mb-2">Content Violations</h4><ul class="list-disc pl-4 space-y-1 text-red-800 text-sm"><li>Adult content or pornography</li><li>Hate speech or discriminatory content</li><li>Violence or graphic content</li><li>Copyright infringement</li></ul></div><div class="bg-red-50 border border-red-200 rounded-lg p-4"><h4 class="font-semibold text-red-900 mb-2">Security Violations</h4><ul class="list-disc pl-4 space-y-1 text-red-800 text-sm"><li>Phishing or social engineering</li><li>Malware distribution</li><li>System hacking attempts</li><li>Unauthorized access</li></ul></div><div class="bg-red-50 border border-red-200 rounded-lg p-4"><h4 class="font-semibold text-red-900 mb-2">Abuse</h4><ul class="list-disc pl-4 space-y-1 text-red-800 text-sm"><li>Spam or unsolicited messages</li><li>Rate limiting bypass</li><li>Service disruption attempts</li><li>Resource abuse</li></ul></div><div class="bg-red-50 border border-red-200 rounded-lg p-4"><h4 class="font-semibold text-red-900 mb-2">Illegal Activities</h4><ul class="list-disc pl-4 space-y-1 text-red-800 text-sm"><li>Illegal goods or services</li><li>Money laundering</li><li>Identity theft</li><li>Fraud or scams</li></ul></div></div></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">6. Intellectual Property Rights</h2><h3 class="text-xl font-medium text-gray-800 mb-3">6.1 Our Content</h3><p class="text-gray-700 mb-4">All content on Trunc, including but not limited to text, graphics, logos, icons, images, audio clips, digital downloads, and software, is the property of Trunc or its content suppliers and is protected by international copyright laws.</p><h3 class="text-xl font-medium text-gray-800 mb-3">6.2 Trademarks</h3><p class="text-gray-700 mb-4">The Trunc name and logo are trademarks of Trunc. You may not use our trademarks without our prior written consent.</p><h3 class="text-xl font-medium text-gray-800 mb-3">6.3 User License</h3><p class="text-gray-700">We grant you a limited, non-exclusive, non-transferable license to access and use our Service for personal or internal business purposes, subject to these Terms.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">7. User Content and Links</h2><p class="text-gray-700 mb-4">When you submit URLs for shortening, you retain ownership of the original content. However, you grant us:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4"><li>The right to store and process your URLs for service operation</li><li>The right to display analytics and statistics related to your links</li><li>The right to remove links that violate these Terms</li></ul><p class="text-gray-700 mb-4"><strong>You are solely responsible for:</strong></p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>All URLs you submit to our Service</li><li>Ensuring you have the right to share destination URLs</li><li>The content at the destination of your shortened links</li><li>Compliance with applicable laws regarding your links</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">8. Service Availability and Modifications</h2><p class="text-gray-700 mb-4">We strive to provide reliable service, but we cannot guarantee uninterrupted access. We reserve the right to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Modify or discontinue the Service at any time with or without notice</li><li>Impose limits on certain features or restrict access to parts of the Service</li><li>Perform maintenance and updates that may temporarily interrupt service</li><li>Change our pricing, features, or service offerings</li></ul><p class="text-gray-700 mt-4">We will make reasonable efforts to notify users of significant changes or extended maintenance periods.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">9. Payment Terms</h2><h3 class="text-xl font-medium text-gray-800 mb-3">9.1 Subscription Plans</h3><p class="text-gray-700 mb-4">Certain features and services may require payment. By subscribing to a paid plan, you agree to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-6"><li>Pay all fees associated with your selected plan</li><li>Provide accurate and complete billing information</li><li>Authorize recurring charges for subscription renewals</li><li>Pay any applicable taxes</li></ul><h3 class="text-xl font-medium text-gray-800 mb-3">9.2 Billing and Renewals</h3><p class="text-gray-700 mb-4">Subscriptions automatically renew unless cancelled before the renewal date. You can cancel your subscription at any time through your account settings.</p><h3 class="text-xl font-medium text-gray-800 mb-3">9.3 Refunds</h3><p class="text-gray-700">Refunds are handled on a case-by-case basis. Please contact our support team to discuss refund requests. We reserve the right to refuse refunds for services already rendered.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">10. Disclaimers and Limitations of Liability</h2><div class="bg-gray-50 border border-gray-300 rounded-lg p-6"><h3 class="text-xl font-medium text-gray-800 mb-3">10.1 Service "As Is"</h3><p class="text-gray-700 mb-4">THE SERVICE IS PROVIDED "AS IS" AND "AS AVAILABLE" WITHOUT WARRANTIES OF ANY KIND, EITHER EXPRESS OR IMPLIED, INCLUDING BUT NOT LIMITED TO IMPLIED WARRANTIES OF MERCHANTABILITY, FITNESS FOR A PARTICULAR PURPOSE, OR NON-INFRINGEMENT.</p><h3 class="text-xl font-medium text-gray-800 mb-3">10.2 No Warranty</h3><p class="text-gray-700 mb-4">We do not warrant that:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4"><li>The Service will meet your specific requirements</li><li>The Service will be uninterrupted, timely, secure, or error-free</li><li>The results obtained from using the Service will be accurate or reliable</li><li>Any errors in the Service will be corrected</li></ul><h3 class="text-xl font-medium text-gray-800 mb-3">10.3 Limitation of Liability</h3><p class="text-gray-700">TO THE MAXIMUM EXTENT PERMITTED BY LAW, TRUNC SHALL NOT BE LIABLE FOR ANY INDIRECT, INCIDENTAL, SPECIAL, CONSEQUENTIAL, OR PUNITIVE DAMAGES, OR ANY LOSS OF PROFITS OR REVENUES, WHETHER INCURRED DIRECTLY OR INDIRECTLY, OR ANY LOSS OF DATA, USE, GOODWILL, OR OTHER INTANGIBLE LOSSES RESULTING FROM YOUR USE OF THE SERVICE.</p></div></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">11. Indemnification</h2><p class="text-gray-700 mb-4">You agree to indemnify, defend, and hold harmless Trunc and its officers, directors, employees, and agents from and against any claims, liabilities, damages, losses, and expenses, including reasonable legal fees, arising out of or in any way connected with:</p><ul class="list-disc pl-6 space-y-2 text-gray-700"><li>Your access to or use of the Service</li><li>Your violation of these Terms</li><li>Your violation of any third-party rights, including intellectual property rights</li><li>Your URLs or the content they link to</li><li>Any harm caused to third parties through your use of the Service</li></ul></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">12. Termination</h2><h3 class="text-xl font-medium text-gray-800 mb-3">12.1 By You</h3><p class="text-gray-700 mb-4">You may terminate your account at any time by contacting us or using the account deletion feature in your settings.</p><h3 class="text-xl font-medium text-gray-800 mb-3">12.2 By Us</h3><p class="text-gray-700 mb-4">We reserve the right to suspend or terminate your account and access to the Service at any time, with or without notice, for any reason, including but not limited to:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4"><li>Violation of these Terms</li><li>Fraudulent or illegal activity</li><li>Abusive behavior toward other users or staff</li><li>Extended period of inactivity</li><li>Request by law enforcement or government agencies</li></ul><h3 class="text-xl font-medium text-gray-800 mb-3">12.3 Effect of Termination</h3><p class="text-gray-700">Upon termination, your right to use the Service will immediately cease. All shortened URLs associated with your account may be deactivated or deleted. We are not liable for any loss or damage resulting from termination of your account.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">13. Governing Law and Dispute Resolution</h2><h3 class="text-xl font-medium text-gray-800 mb-3">13.1 Governing Law</h3><p class="text-gray-700 mb-4">These Terms shall be governed by and construed in accordance with the laws of the State of California, United States, without regard to its conflict of law provisions.</p><h3 class="text-xl font-medium text-gray-800 mb-3">13.2 Dispute Resolution</h3><p class="text-gray-700 mb-4">Any dispute arising from these Terms or your use of the Service shall be resolved through:</p><ol class="list-decimal pl-6 space-y-2 text-gray-700"><li><strong>Informal Resolution:</strong> Contact us at legal@urlshorter.com to attempt to resolve the dispute informally</li><li><strong>Binding Arbitration:</strong> If informal resolution fails, disputes will be resolved through binding arbitration in accordance with the rules of the American Arbitration Association</li><li><strong>Class Action Waiver:</strong> You agree to resolve disputes on an individual basis and waive any right to participate in a class action lawsuit</li></ol></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">14. Changes to Terms</h2><p class="text-gray-700 mb-4">We reserve the right to modify these Terms at any time. We will notify users of material changes by:</p><ul class="list-disc pl-6 space-y-2 text-gray-700 mb-4"><li>Posting the updated Terms on our website</li><li>Sending email notifications to registered users</li><li>Displaying a prominent notice in the Service</li></ul><p class="text-gray-700">Your continued use of the Service after any changes constitutes acceptance of the new Terms. If you do not agree to the modified Terms, you must stop using the Service.</p></section><section><h2 class="text-2xl font-semibold text-gray-900 mb-4">15. Miscellaneous</h2><h3 class="text-xl font-medium text-gray-800 mb-3">15.1 Entire Agreement</h3><p class="text-gray-700 mb-4">These Terms constitute the entire agreement between you and Trunc regarding the use of the Service.</p><h3 class="text-xl font-medium text-gray-800 mb-3">15.2 Severability</h3><p class="text-gray-700 mb-4">If any provision of these Terms is found to be unenforceable or invalid, that provision will be limited or eliminated to the minimum extent necessary so that these Terms will otherwise remain in full force and effect.</p><h3 class="text-xl font-medium text-gray-800 mb-3">15.3 Waiver</h3><p class="text-gray-700 mb-4">Our failure to enforce any right or provision of these Terms will not be considered a waiver of those rights.</p><h3 class="text-xl font-medium text-gray-800 mb-3">15.4 Assignment</h3><p class="text-gray-700">You may not assign or transfer these Terms or your rights hereunder without our prior written consent. We may assign these Terms without restriction.</p></section><section class="bg-gray-50 p-6 rounded-lg"><h2 class="text-2xl font-semibold text-gray-900 mb-4">16. Contact Us</h2><p class="text-gray-700 mb-4">If you have any questions about these Terms of Service, please contact us:</p><div class="space-y-2 text-gray-700"><p><strong>Email:</strong> <a href="mailto:legal@urlshorter.com" class="text-blue-600 hover:underline">legal@urlshorter.com</a></p><p><strong>Support:</strong> <a href="mailto:support@urlshorter.com" class="text-blue-600 hover:underline">support@urlshorter.com</a></p><p><strong>Address:</strong> Trunc Technologies, 123 Tech Boulevard, San Francisco, CA 94105</p><p><strong>Phone:</strong> <a href="tel:+1-555-0123" class="text-blue-600 hover:underline">+1 (555) 012-3456</a></p></div><div class="bg-blue-50 border border-blue-200 rounded-lg p-4 mt-6"><p class="text-blue-900 text-sm"><strong>Important:</strong> By using Trunc, you acknowledge that you have read, understood, and agree to be bound by these Terms of Service and our Privacy Policy.</p></div></section></div></div>', 'Review our terms of service before using the platform.', NULL, true, false, 'default', 'en', 8),
('conversions', 'Conversions', 'Track conversions and measure campaign success.', 'Monitor conversion events and analyze campaign performance.', NULL, true, false, 'default', 'en', 9)
ON CONFLICT (slug) DO NOTHING;

-- Insert default page templates
INSERT INTO page_templates (name, slug, description, content, is_active) VALUES
('default', 'default', 'Default page template', '<div class="container">{{content}}</div>', true),
('landing', 'landing', 'Landing page template', '<div class="landing-page">{{content}}</div>', true),
('blog', 'blog', 'Blog post template', '<article class="blog-post">{{content}}</article>', true)
ON CONFLICT (slug) DO NOTHING;

-- Insert default navigation menu
INSERT INTO navigation_menus (name, location, items, is_active) VALUES
('main', 'header', '[{"label":"Home","url":"/"},{"label":"Features","url":"/features"},{"label":"Pricing","url":"/pricing"},{"label":"API Docs","url":"/api-docs"},{"label":"About","url":"/about"},{"label":"Contact","url":"/contact"}]', true),
('footer', 'footer', '[{"label":"Privacy Policy","url":"/privacy"},{"label":"Terms of Service","url":"/terms"},{"label":"Contact","url":"/contact"}]', true);

COMMENT ON TABLE static_pages IS 'CMS static pages with SEO and analytics';
COMMENT ON TABLE page_revisions IS 'Version control for page edits';
COMMENT ON TABLE page_visits IS 'Page visit analytics';
COMMENT ON TABLE media_files IS 'Uploaded media files';
COMMENT ON TABLE global_settings IS 'Site-wide configuration settings';
COMMENT ON TABLE content_blocks IS 'Reusable content components';
COMMENT ON TABLE page_templates IS 'Page layout templates';
COMMENT ON TABLE navigation_menus IS 'Dynamic site navigation menus';
