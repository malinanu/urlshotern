-- CMS Tables for static page management and API keys

-- Static pages table for CMS
CREATE TABLE IF NOT EXISTS static_pages (
    id BIGSERIAL PRIMARY KEY,
    slug VARCHAR(100) UNIQUE NOT NULL,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    content_blocks JSONB,
    meta_description TEXT,
    meta_keywords TEXT,
    meta_image VARCHAR(255),
    meta_title VARCHAR(200),
    canonical_url VARCHAR(255),
    schema_markup JSONB,
    is_published BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    template VARCHAR(50) DEFAULT 'default',
    author_id BIGINT REFERENCES users(id),
    published_at TIMESTAMP WITH TIME ZONE,
    scheduled_at TIMESTAMP WITH TIME ZONE,
    expires_at TIMESTAMP WITH TIME ZONE,
    featured_image_id BIGINT,
    category VARCHAR(50),
    tags JSONB,
    custom_fields JSONB,
    view_count BIGINT DEFAULT 0,
    is_sticky BOOLEAN DEFAULT FALSE,
    allow_comments BOOLEAN DEFAULT TRUE,
    language VARCHAR(10) DEFAULT 'en',
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API Keys table for programmatic access
CREATE TABLE IF NOT EXISTS api_keys (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    key_hash VARCHAR(255) UNIQUE NOT NULL,
    key_prefix VARCHAR(20) NOT NULL,
    name VARCHAR(100) NOT NULL,
    description TEXT,
    permissions JSONB DEFAULT '[]',
    rate_limit INTEGER DEFAULT 1000,
    last_used_at TIMESTAMP WITH TIME ZONE,
    last_used_ip INET,
    expires_at TIMESTAMP WITH TIME ZONE,
    is_active BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page revisions for version control
CREATE TABLE IF NOT EXISTS page_revisions (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    title VARCHAR(200) NOT NULL,
    content TEXT NOT NULL,
    meta_description TEXT,
    meta_keywords TEXT,
    revision_number INTEGER NOT NULL,
    author_id BIGINT REFERENCES users(id),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page visits tracking
CREATE TABLE IF NOT EXISTS page_visits (
    id BIGSERIAL PRIMARY KEY,
    page_id BIGINT NOT NULL REFERENCES static_pages(id) ON DELETE CASCADE,
    ip_address INET,
    user_agent TEXT,
    referrer TEXT,
    country_code CHAR(2),
    visited_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- API key usage tracking
CREATE TABLE IF NOT EXISTS api_key_usage (
    id BIGSERIAL PRIMARY KEY,
    api_key_id BIGINT NOT NULL REFERENCES api_keys(id) ON DELETE CASCADE,
    endpoint VARCHAR(100) NOT NULL,
    method VARCHAR(10) NOT NULL,
    status_code INTEGER,
    response_time_ms INTEGER,
    ip_address INET,
    user_agent TEXT,
    request_size INTEGER,
    response_size INTEGER,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes
CREATE INDEX IF NOT EXISTS idx_static_pages_slug ON static_pages(slug);
CREATE INDEX IF NOT EXISTS idx_static_pages_published ON static_pages(is_published);
CREATE INDEX IF NOT EXISTS idx_static_pages_sort_order ON static_pages(sort_order);
CREATE INDEX IF NOT EXISTS idx_static_pages_author_id ON static_pages(author_id);

CREATE INDEX IF NOT EXISTS idx_api_keys_user_id ON api_keys(user_id);
CREATE INDEX IF NOT EXISTS idx_api_keys_key_hash ON api_keys(key_hash);
CREATE INDEX IF NOT EXISTS idx_api_keys_active ON api_keys(is_active);
CREATE INDEX IF NOT EXISTS idx_api_keys_expires_at ON api_keys(expires_at);

CREATE INDEX IF NOT EXISTS idx_page_revisions_page_id ON page_revisions(page_id);
CREATE INDEX IF NOT EXISTS idx_page_visits_page_id ON page_visits(page_id);
CREATE INDEX IF NOT EXISTS idx_page_visits_visited_at ON page_visits(visited_at);

CREATE INDEX IF NOT EXISTS idx_api_key_usage_api_key_id ON api_key_usage(api_key_id);
CREATE INDEX IF NOT EXISTS idx_api_key_usage_created_at ON api_key_usage(created_at);

-- Media files table for file management
CREATE TABLE IF NOT EXISTS media_files (
    id BIGSERIAL PRIMARY KEY,
    user_id BIGINT NOT NULL REFERENCES users(id) ON DELETE CASCADE,
    filename VARCHAR(255) NOT NULL,
    original_name VARCHAR(255) NOT NULL,
    file_path VARCHAR(500) NOT NULL,
    file_url VARCHAR(500) NOT NULL,
    file_size BIGINT NOT NULL,
    mime_type VARCHAR(100) NOT NULL,
    file_type VARCHAR(50) NOT NULL,
    width INTEGER,
    height INTEGER,
    alt TEXT,
    caption TEXT,
    is_public BOOLEAN DEFAULT TRUE,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Global site settings table
CREATE TABLE IF NOT EXISTS global_settings (
    id BIGSERIAL PRIMARY KEY,
    key VARCHAR(100) UNIQUE NOT NULL,
    value TEXT NOT NULL,
    type VARCHAR(20) DEFAULT 'text',
    category VARCHAR(50) DEFAULT 'general',
    display_name VARCHAR(100) NOT NULL,
    description TEXT,
    is_public BOOLEAN DEFAULT FALSE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Content blocks table for reusable components
CREATE TABLE IF NOT EXISTS content_blocks (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    identifier VARCHAR(100) UNIQUE NOT NULL,
    content TEXT NOT NULL,
    type VARCHAR(50) DEFAULT 'generic',
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Page templates table
CREATE TABLE IF NOT EXISTS page_templates (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    slug VARCHAR(100) UNIQUE NOT NULL,
    description TEXT,
    content TEXT NOT NULL,
    fields JSONB,
    preview_url VARCHAR(255),
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Navigation menus table
CREATE TABLE IF NOT EXISTS navigation_menus (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(100) NOT NULL,
    location VARCHAR(50) NOT NULL,
    items JSONB NOT NULL,
    is_active BOOLEAN DEFAULT TRUE,
    sort_order INTEGER DEFAULT 0,
    created_at TIMESTAMP WITH TIME ZONE DEFAULT NOW(),
    updated_at TIMESTAMP WITH TIME ZONE DEFAULT NOW()
);

-- Create indexes for new tables
CREATE INDEX IF NOT EXISTS idx_media_files_user_id ON media_files(user_id);
CREATE INDEX IF NOT EXISTS idx_media_files_file_type ON media_files(file_type);
CREATE INDEX IF NOT EXISTS idx_media_files_is_public ON media_files(is_public);
CREATE INDEX IF NOT EXISTS idx_media_files_created_at ON media_files(created_at);

CREATE INDEX IF NOT EXISTS idx_global_settings_key ON global_settings(key);
CREATE INDEX IF NOT EXISTS idx_global_settings_category ON global_settings(category);
CREATE INDEX IF NOT EXISTS idx_global_settings_is_public ON global_settings(is_public);

CREATE INDEX IF NOT EXISTS idx_content_blocks_identifier ON content_blocks(identifier);
CREATE INDEX IF NOT EXISTS idx_content_blocks_type ON content_blocks(type);
CREATE INDEX IF NOT EXISTS idx_content_blocks_is_active ON content_blocks(is_active);

CREATE INDEX IF NOT EXISTS idx_page_templates_slug ON page_templates(slug);
CREATE INDEX IF NOT EXISTS idx_page_templates_is_active ON page_templates(is_active);

CREATE INDEX IF NOT EXISTS idx_navigation_menus_location ON navigation_menus(location);
CREATE INDEX IF NOT EXISTS idx_navigation_menus_is_active ON navigation_menus(is_active);

-- Add foreign key for featured image
ALTER TABLE static_pages ADD CONSTRAINT fk_static_pages_featured_image 
    FOREIGN KEY (featured_image_id) REFERENCES media_files(id) ON DELETE SET NULL;

-- Insert default static pages
INSERT INTO static_pages (slug, title, content, meta_description, is_published, sort_order, published_at) VALUES
    ('home', 'Trunc - Professional URL Shortener', '<div class="home-page"><section class="hero bg-gradient-to-br from-primary-50 to-white py-20"><div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8"><div class="text-center"><h1 class="text-4xl md:text-6xl font-bold text-gray-900 mb-6">Shorten URLs<br><span class="text-primary-600">Amplify Results</span></h1><p class="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">Transform long URLs into powerful short links with advanced analytics, custom branding, and enterprise-grade reliability.</p><div class="flex flex-col sm:flex-row gap-4 justify-center"><a href="/register" class="btn btn-primary btn-lg">Start Free</a><a href="/features" class="btn btn-secondary btn-lg">View Features</a></div></div></div></section><section class="features py-20"><div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8"><div class="grid md:grid-cols-3 gap-8"><div class="text-center"><div class="w-16 h-16 bg-primary-100 rounded-lg flex items-center justify-center mx-auto mb-4"><svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path></svg></div><h3 class="text-xl font-semibold text-gray-900 mb-2">Lightning Fast</h3><p class="text-gray-600">Shorten URLs instantly with our optimized infrastructure.</p></div><div class="text-center"><div class="w-16 h-16 bg-primary-100 rounded-lg flex items-center justify-center mx-auto mb-4"><svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg></div><h3 class="text-xl font-semibold text-gray-900 mb-2">Advanced Analytics</h3><p class="text-gray-600">Track clicks, locations, and referrers with detailed insights.</p></div><div class="text-center"><div class="w-16 h-16 bg-primary-100 rounded-lg flex items-center justify-center mx-auto mb-4"><svg class="w-8 h-8 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.031 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path></svg></div><h3 class="text-xl font-semibold text-gray-900 mb-2">Secure & Reliable</h3><p class="text-gray-600">Enterprise-grade security with 99.9% uptime guarantee.</p></div></div></div></section></div>', 'Trunc - Professional URL shortener with advanced analytics, custom branding, and enterprise reliability. Start shortening URLs today.', true, 0, NOW()),

    ('features', 'Features - Trunc URL Shortener', '<div class="features-page"><section class="hero py-20 bg-white"><div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center"><h1 class="text-4xl md:text-5xl font-bold text-gray-900 mb-6">Powerful Features for<br><span class="text-primary-600">Modern URL Management</span></h1><p class="text-xl text-gray-600 mb-8 max-w-3xl mx-auto">Discover all the tools and capabilities that make Trunc the best URL shortening platform for individuals and businesses.</p></div></section><section class="main-features py-20 bg-gray-50"><div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8"><div class="grid lg:grid-cols-2 gap-16"><div class="feature-block"><div class="flex items-center mb-6"><div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mr-4"><svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path></svg></div><h3 class="text-2xl font-bold text-gray-900">Lightning Fast Shortening</h3></div><p class="text-gray-600 mb-4">Generate short URLs instantly with our optimized infrastructure powered by Redis caching and efficient algorithms.</p><ul class="space-y-2 text-gray-600"><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Sub-second response times</li><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Global CDN distribution</li><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Auto-scaling infrastructure</li></ul></div><div class="feature-block"><div class="flex items-center mb-6"><div class="w-12 h-12 bg-primary-100 rounded-lg flex items-center justify-center mr-4"><svg class="w-6 h-6 text-primary-600" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path></svg></div><h3 class="text-2xl font-bold text-gray-900">Advanced Analytics</h3></div><p class="text-gray-600 mb-4">Track every click with detailed insights including geographic data, device information, and referrer analysis.</p><ul class="space-y-2 text-gray-600"><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Real-time click tracking</li><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Geographic insights</li><li class="flex items-center"><svg class="w-5 h-5 text-green-500 mr-2" fill="none" stroke="currentColor" viewBox="0 0 24 24"><path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M5 13l4 4L19 7"></path></svg>Device and browser data</li></ul></div></div></div></section></div>', 'Discover Trunc''s powerful features including lightning-fast URL shortening, advanced analytics, custom domains, and enterprise security.', true, 1, NOW()),

    ('about', 'About Us', '<h1>About Trunc</h1><p>Trunc is a powerful URL shortening service that helps you create short, memorable links for your long URLs. Track clicks, analyze performance, and manage your links with advanced analytics.</p><h2>Our Mission</h2><p>To provide the most reliable and feature-rich URL shortening service for individuals and businesses.</p><h2>Features</h2><ul><li>Custom short codes</li><li>Advanced analytics</li><li>Team collaboration</li><li>API access</li><li>Custom domains</li></ul>', 'Learn about Trunc - the powerful URL shortening service with advanced analytics and team collaboration features.', true, 2, NOW()),
    
    ('contact', 'Contact Us', '<h1>Contact Us</h1><p>Have questions or need support? We''re here to help!</p><h2>Get in Touch</h2><div class="contact-methods"><div class="contact-item"><h3>Email Support</h3><p>For general inquiries and support:</p><p><a href="mailto:support@3logiq.com">support@3logiq.com</a></p></div><div class="contact-item"><h3>Sales</h3><p>For enterprise solutions:</p><p><a href="mailto:sales@3logiq.com">sales@3logiq.com</a></p></div><div class="contact-item"><h3>Technical Support</h3><p>For API and integration help:</p><p><a href="mailto:technical@3logiq.com">technical@3logiq.com</a></p></div></div><h2>Office Hours</h2><p>Monday - Friday: 9:00 AM - 6:00 PM (UTC)<br>Weekend support available for premium customers</p>', 'Contact Trunc support team for help with your account, API integration, or any questions about our service.', true, 2, NOW()),
    
    ('privacy', 'Privacy Policy', '<h1>Privacy Policy</h1><p><strong>Last updated:</strong> ' || TO_CHAR(NOW(), 'Month DD, YYYY') || '</p><h2>Information We Collect</h2><p>We collect information you provide directly to us, such as when you create an account, create shortened URLs, or contact us for support.</p><h3>Account Information</h3><ul><li>Email address</li><li>Name</li><li>Password (encrypted)</li></ul><h3>Usage Information</h3><ul><li>URLs you shorten</li><li>Click analytics data</li><li>Device and browser information</li><li>IP addresses</li></ul><h2>How We Use Your Information</h2><p>We use the information we collect to:</p><ul><li>Provide and maintain our service</li><li>Process transactions</li><li>Send you technical notices and support messages</li><li>Analyze usage patterns to improve our service</li></ul><h2>Information Sharing</h2><p>We do not sell, trade, or otherwise transfer your personal information to third parties without your consent, except as described in this policy.</p><h2>Data Security</h2><p>We implement appropriate security measures to protect your personal information against unauthorized access, alteration, disclosure, or destruction.</p><h2>Contact Us</h2><p>If you have questions about this Privacy Policy, please contact us at <a href="mailto:privacy@3logiq.com">privacy@3logiq.com</a>.</p>', 'Trunc Privacy Policy - Learn how we collect, use, and protect your personal information and data.', true, 3, NOW()),
    
    ('terms', 'Terms of Service', '<h1>Terms of Service</h1><p><strong>Last updated:</strong> ' || TO_CHAR(NOW(), 'Month DD, YYYY') || '</p><h2>Acceptance of Terms</h2><p>By accessing and using Trunc, you accept and agree to be bound by the terms and provision of this agreement.</p><h2>Description of Service</h2><p>Trunc is a URL shortening service that allows users to create shortened versions of long URLs and track their performance.</p><h2>User Accounts</h2><p>To access certain features, you must create an account. You are responsible for:</p><ul><li>Maintaining the confidentiality of your account</li><li>All activities that occur under your account</li><li>Providing accurate and complete information</li></ul><h2>Acceptable Use</h2><p>You agree not to use Trunc for:</p><ul><li>Illegal activities</li><li>Spam or unsolicited communications</li><li>Malware distribution</li><li>Phishing or fraud</li><li>Content that violates intellectual property rights</li></ul><h2>Service Availability</h2><p>We strive to maintain high availability but do not guarantee uninterrupted service. We may suspend service for maintenance or other operational reasons.</p><h2>Limitation of Liability</h2><p>Trunc shall not be liable for any indirect, incidental, special, or consequential damages resulting from the use or inability to use our service.</p><h2>Termination</h2><p>We reserve the right to terminate or suspend accounts that violate these terms of service.</p><h2>Contact Information</h2><p>Questions about the Terms of Service should be sent to <a href="mailto:legal@3logiq.com">legal@3logiq.com</a>.</p>', 'Trunc Terms of Service - Rules and guidelines for using our URL shortening service.', true, 4, NOW()),
    
    ('help', 'Help Center', '<h1>Help Center</h1><p>Welcome to the Trunc Help Center. Find answers to common questions and learn how to make the most of our service.</p><h2>Getting Started</h2><h3>How to create a short URL</h3><ol><li>Paste your long URL into the shortening box</li><li>Click "Shorten URL" or press Enter</li><li>Copy your new short URL</li><li>Share it anywhere you like!</li></ol><h3>Creating an Account</h3><p>While you can create short URLs without an account, registering gives you access to:</p><ul><li>Link management dashboard</li><li>Click analytics</li><li>Custom short codes</li><li>Bulk URL shortening</li><li>API access</li></ul><h2>Features</h2><h3>Analytics</h3><p>Track your link performance with detailed analytics including:</p><ul><li>Click counts and trends</li><li>Geographic data</li><li>Device and browser information</li><li>Referrer sources</li></ul><h3>Custom Domains</h3><p>Premium users can use their own custom domains for branded short links.</p><h3>API Access</h3><p>Integrate Trunc into your applications with our REST API.</p><h2>Frequently Asked Questions</h2><h3>How long do short URLs last?</h3><p>Short URLs created with Trunc never expire by default. Premium users can set custom expiration dates.</p><h3>Can I edit the destination of a short URL?</h3><p>Yes, registered users can update the destination URL from their dashboard.</p><h3>Is there a limit to how many URLs I can shorten?</h3><p>Free users can shorten up to 1,000 URLs per month. Premium plans offer higher limits.</p><h2>Need More Help?</h2><p>If you can''t find what you''re looking for, <a href="/contact">contact our support team</a>.</p>', 'Trunc Help Center - Find answers to common questions and learn how to use our URL shortening service effectively.', true, 5, NOW()),
    
    ('status', 'Service Status', '<h1>Service Status</h1><p>Current status of Trunc services and infrastructure.</p><div class="status-dashboard"><div class="status-item operational"><h3>URL Shortening Service</h3><p class="status">Operational</p><p class="uptime">99.9% uptime (last 30 days)</p></div><div class="status-item operational"><h3>Analytics Dashboard</h3><p class="status">Operational</p><p class="uptime">99.8% uptime (last 30 days)</p></div><div class="status-item operational"><h3>API Service</h3><p class="status">Operational</p><p class="uptime">99.9% uptime (last 30 days)</p></div><div class="status-item operational"><h3>Database</h3><p class="status">Operational</p><p class="uptime">100% uptime (last 30 days)</p></div></div><h2>Recent Incidents</h2><p>No incidents reported in the last 30 days.</p><h2>Scheduled Maintenance</h2><p>No scheduled maintenance at this time.</p><p><em>Last updated: ' || TO_CHAR(NOW(), 'Month DD, YYYY at HH24:MI UTC') || '</em></p>', 'Check the current operational status of Trunc services including uptime statistics and incident reports.', true, 6, NOW()),
    
    ('blog', 'Blog', '<h1>Trunc Blog</h1><p>Stay updated with the latest news, features, and tips from Trunc.</p><div class="blog-posts"><article class="blog-post"><h2><a href="/blog/welcome-to-trunc">Welcome to Trunc</a></h2><p class="post-meta">Published on ' || TO_CHAR(NOW() - INTERVAL ''7 days'', 'Month DD, YYYY') || '</p><p>We''re excited to launch Trunc, a powerful and reliable URL shortening service designed for modern web users and businesses...</p><a href="/blog/welcome-to-trunc" class="read-more">Read more</a></article><article class="blog-post"><h2><a href="/blog/introducing-analytics">Introducing Advanced Analytics</a></h2><p class="post-meta">Published on ' || TO_CHAR(NOW() - INTERVAL ''14 days'', 'Month DD, YYYY') || '</p><p>Track your link performance like never before with our new advanced analytics dashboard featuring geographic insights, device analytics, and more...</p><a href="/blog/introducing-analytics" class="read-more">Read more</a></article><article class="blog-post"><h2><a href="/blog/api-launch">API Now Available</a></h2><p class="post-meta">Published on ' || TO_CHAR(NOW() - INTERVAL ''21 days'', 'Month DD, YYYY') || '</p><p>Integrate Trunc into your applications with our comprehensive REST API. Generate API keys and start building today...</p><a href="/blog/api-launch" class="read-more">Read more</a></article></div><p><a href="/blog/archive">View all blog posts →</a></p>', 'Trunc blog - Latest news, feature announcements, and tips for getting the most out of our URL shortening service.', true, 7, NOW()),
    
    ('community', 'Community', '<h1>Trunc Community</h1><p>Connect with other Trunc users, share tips, and get help from the community.</p><h2>Join Our Community</h2><div class="community-links"><div class="community-item"><h3>Discord Server</h3><p>Chat with other users and get real-time help.</p><p><a href="#" class="btn btn-primary">Join Discord</a></p></div><div class="community-item"><h3>GitHub</h3><p>Contribute to our open-source projects and report issues.</p><p><a href="#" class="btn btn-primary">View on GitHub</a></p></div><div class="community-item"><h3>Reddit Community</h3><p>Share tips, ask questions, and discuss URL shortening best practices.</p><p><a href="#" class="btn btn-primary">Join r/Trunc</a></p></div></div><h2>Community Guidelines</h2><ul><li>Be respectful and helpful to other community members</li><li>Stay on topic and keep discussions relevant to Trunc</li><li>No spam, self-promotion, or advertising</li><li>Search before posting to avoid duplicate discussions</li><li>Follow platform-specific rules and guidelines</li></ul><h2>Community Resources</h2><ul><li><a href="/help">Help Center</a> - Comprehensive documentation and guides</li><li><a href="/api-docs">API Documentation</a> - Technical reference for developers</li><li><a href="/blog">Blog</a> - Latest news and feature updates</li><li><a href="/contact">Contact Support</a> - Direct help from our team</li></ul><h2>Featured Community Contributions</h2><p>Check out some amazing projects and integrations created by our community members:</p><ul><li>Trunc WordPress Plugin</li><li>Browser Extension for Quick Shortening</li><li>Mobile App for iOS and Android</li><li>Zapier Integration</li></ul>', 'Join the Trunc community to connect with other users, share tips, contribute to projects, and get community support.', true, 8, NOW())
ON CONFLICT (slug) DO NOTHING;

-- Add CMS permissions
INSERT INTO permissions (name, display_name, description, resource, action, is_system) VALUES
    ('cms.read', 'Read CMS Content', 'View CMS pages and content', 'cms', 'read', true),
    ('cms.create', 'Create CMS Content', 'Create new CMS pages', 'cms', 'create', true),
    ('cms.update', 'Update CMS Content', 'Edit existing CMS pages', 'cms', 'update', true),
    ('cms.delete', 'Delete CMS Content', 'Delete CMS pages', 'cms', 'delete', true),
    ('cms.publish', 'Publish CMS Content', 'Publish and unpublish pages', 'cms', 'publish', true),
    ('api_key.read', 'Read API Keys', 'View API keys', 'api_key', 'read', true),
    ('api_key.create', 'Create API Keys', 'Generate new API keys', 'api_key', 'create', true),
    ('api_key.delete', 'Delete API Keys', 'Revoke API keys', 'api_key', 'delete', true)
ON CONFLICT (name) DO NOTHING;

-- Assign CMS permissions to admin role
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name = 'admin' AND p.name IN ('cms.read', 'cms.create', 'cms.update', 'cms.delete', 'cms.publish')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Assign API key permissions to users and premium users
INSERT INTO role_permissions (role_id, permission_id)
SELECT r.id, p.id FROM roles r, permissions p
WHERE r.name IN ('user', 'premium') AND p.name IN ('api_key.read', 'api_key.create', 'api_key.delete')
ON CONFLICT (role_id, permission_id) DO NOTHING;

-- Insert default global settings
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

-- Insert default content blocks
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
    
    ('Hero Section', 'hero_home',
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

    ('CTA Section', 'cta_bottom',
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
    </div>', 'cta', true, 3)
ON CONFLICT (identifier) DO NOTHING;

-- Insert default page templates
INSERT INTO page_templates (name, slug, description, content, fields, is_active, sort_order) VALUES
    ('Default Page', 'default', 'Standard page layout with header, content, and footer', 
    '<div class="min-h-screen bg-white">
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            <div class="prose prose-lg max-w-none">
                {{content}}
            </div>
        </div>
    </div>', '{}', true, 1),
    
    ('Landing Page', 'landing', 'Marketing-focused landing page with hero section and CTA',
    '<div class="min-h-screen">
        {{block:hero_home}}
        <div class="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
            <div class="prose prose-lg max-w-none">
                {{content}}
            </div>
        </div>
        {{block:cta_bottom}}
    </div>', '{"hero_title": "Hero Title", "hero_subtitle": "Hero Subtitle", "cta_text": "Call to Action"}', true, 2),
    
    ('Blog Post', 'blog', 'Blog post layout with featured image and author info',
    '<article class="min-h-screen bg-white">
        <div class="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
            {{#if featured_image}}
            <div class="mb-8">
                <img src="{{featured_image}}" alt="{{title}}" class="w-full h-64 object-cover rounded-xl">
            </div>
            {{/if}}
            <header class="mb-8">
                <h1 class="text-4xl font-bold text-gray-900 mb-4">{{title}}</h1>
                <div class="flex items-center text-gray-600 text-sm">
                    <span>By {{author_name}}</span>
                    <span class="mx-2">•</span>
                    <time>{{published_at}}</time>
                    {{#if category}}
                    <span class="mx-2">•</span>
                    <span class="bg-blue-100 text-blue-800 px-2 py-1 rounded">{{category}}</span>
                    {{/if}}
                </div>
            </header>
            <div class="prose prose-lg max-w-none">
                {{content}}
            </div>
        </div>
    </article>', '{"show_author": true, "show_date": true, "show_category": true}', true, 3)
ON CONFLICT (slug) DO NOTHING;

-- Insert default navigation menus
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