-- Migration: Add content blocks for all pages and update pages to use block placeholders
-- Created: 2026-01-16
-- Description: Creates reusable content blocks for Home, Features, Pricing, API Docs pages
--              and updates Privacy/Terms pages to use their existing blocks

-- ============================================================================
-- HOME PAGE CONTENT BLOCKS
-- ============================================================================

INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order) VALUES
(
  'Home - Hero Section',
  'home-hero',
  '<div class="text-center space-y-6">
    <h1 class="text-5xl md:text-6xl font-bold text-gray-900">
      Shorten URLs. <span class="text-primary-600">Track Everything.</span>
    </h1>
    <p class="text-xl text-gray-600 max-w-3xl mx-auto">
      Transform your long URLs into short, memorable links. Track analytics, manage custom domains, and boost your link marketing strategy.
    </p>
  </div>',
  'hero',
  true,
  10
),
(
  'Home - Key Features List',
  'home-features-list',
  '<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-6">
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">🚀</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Fast and Reliable</h3>
      <p class="text-gray-600">Lightning-fast URL shortening powered by Redis caching</p>
    </div>
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">📊</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Analytics Tracking</h3>
      <p class="text-gray-600">Comprehensive analytics with geographic and device insights</p>
    </div>
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">🎯</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Custom Branded Links</h3>
      <p class="text-gray-600">Create memorable short links with custom aliases</p>
    </div>
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">🔒</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Secure & Private</h3>
      <p class="text-gray-600">Enterprise-grade security with SSL encryption</p>
    </div>
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">⚡</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Real-time Tracking</h3>
      <p class="text-gray-600">Monitor clicks in real-time with instant updates</p>
    </div>
    <div class="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
      <div class="text-4xl mb-3">🌍</div>
      <h3 class="text-lg font-semibold text-gray-900 mb-2">Global Analytics</h3>
      <p class="text-gray-600">Track clicks from anywhere in the world</p>
    </div>
  </div>',
  'list',
  true,
  20
),
(
  'Home - CTA Section',
  'home-cta',
  '<div class="bg-primary-600 text-white rounded-xl p-12 text-center">
    <h2 class="text-3xl font-bold mb-4">Ready to Get Started?</h2>
    <p class="text-xl text-primary-100 mb-8 max-w-2xl mx-auto">
      Join thousands of users who trust Trunc for their URL shortening needs
    </p>
    <div class="flex flex-col sm:flex-row gap-4 justify-center">
      <a href="/signup" class="inline-flex items-center justify-center px-8 py-3 bg-white text-primary-600 font-semibold rounded-lg hover:bg-gray-100 transition-colors">
        Start Free Trial
      </a>
      <a href="/pricing" class="inline-flex items-center justify-center px-8 py-3 bg-primary-700 text-white font-semibold rounded-lg hover:bg-primary-800 border border-primary-500 transition-colors">
        View Pricing
      </a>
    </div>
  </div>',
  'cta',
  true,
  30
);

-- ============================================================================
-- FEATURES PAGE CONTENT BLOCKS
-- ============================================================================

INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order) VALUES
(
  'Features - Hero Section',
  'features-hero',
  '<div class="text-center space-y-4 mb-12">
    <h1 class="text-4xl md:text-5xl font-bold text-gray-900">
      Powerful Features for <span class="text-primary-600">Modern Link Management</span>
    </h1>
    <p class="text-xl text-gray-600 max-w-3xl mx-auto">
      Everything you need to shorten, track, and optimize your links
    </p>
  </div>',
  'hero',
  true,
  40
),
(
  'Features - Core Features Grid',
  'features-core-grid',
  '<div class="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
    <div class="bg-white p-8 rounded-xl shadow-sm border border-gray-200">
      <div class="w-12 h-12 bg-blue-100 rounded-lg flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-blue-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M13.828 10.172a4 4 0 00-5.656 0l-4 4a4 4 0 105.656 5.656l1.102-1.101m-.758-4.899a4 4 0 005.656 0l4-4a4 4 0 00-5.656-5.656l-1.1 1.1"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-3">Lightning Fast Shortening</h3>
      <p class="text-gray-600 mb-4">Generate short URLs instantly with sub-second response times</p>
      <ul class="space-y-2 text-sm text-gray-600">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Global CDN distribution</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Auto-scaling infrastructure</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Bulk URL processing</li>
      </ul>
    </div>

    <div class="bg-white p-8 rounded-xl shadow-sm border border-gray-200">
      <div class="w-12 h-12 bg-purple-100 rounded-lg flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-purple-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 19v-6a2 2 0 00-2-2H5a2 2 0 00-2 2v6a2 2 0 002 2h2a2 2 0 002-2zm0 0V9a2 2 0 012-2h2a2 2 0 012 2v10m-6 0a2 2 0 002 2h2a2 2 0 002-2m0 0V5a2 2 0 012-2h2a2 2 0 012 2v14a2 2 0 01-2 2h-2a2 2 0 01-2-2z"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-3">Advanced Analytics</h3>
      <p class="text-gray-600 mb-4">Track every click with detailed insights and geographic data</p>
      <ul class="space-y-2 text-sm text-gray-600">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Real-time click tracking</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Geographic click mapping</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Device and browser analytics</li>
      </ul>
    </div>

    <div class="bg-white p-8 rounded-xl shadow-sm border border-gray-200">
      <div class="w-12 h-12 bg-green-100 rounded-lg flex items-center justify-center mb-4">
        <svg class="w-6 h-6 text-green-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
          <path stroke-linecap="round" stroke-linejoin="round" stroke-width="2" d="M9 12l2 2 4-4m5.618-4.016A11.955 11.955 0 0112 2.944a11.955 11.955 0 01-8.618 3.04A12.02 12.02 0 003 9c0 5.591 3.824 10.29 9 11.622 5.176-1.332 9-6.03 9-11.622 0-1.042-.133-2.052-.382-3.016z"></path>
        </svg>
      </div>
      <h3 class="text-xl font-semibold text-gray-900 mb-3">Enterprise Security</h3>
      <p class="text-gray-600 mb-4">Bank-level security with SSL encryption and spam protection</p>
      <ul class="space-y-2 text-sm text-gray-600">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> SSL/TLS encryption</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> Malware URL scanning</li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span> 99.9% uptime SLA</li>
      </ul>
    </div>
  </div>',
  'list',
  true,
  50
);

-- ============================================================================
-- PRICING PAGE CONTENT BLOCKS
-- ============================================================================

INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order) VALUES
(
  'Pricing - Hero Section',
  'pricing-hero',
  '<div class="text-center space-y-4 mb-12">
    <h1 class="text-4xl md:text-5xl font-bold text-gray-900">
      Simple, <span class="text-primary-600">Transparent Pricing</span>
    </h1>
    <p class="text-xl text-gray-600 max-w-2xl mx-auto">
      Choose the perfect plan for your needs. All plans include core features.
    </p>
  </div>',
  'hero',
  true,
  60
),
(
  'Pricing - Plans Grid',
  'pricing-plans',
  '<div class="grid grid-cols-1 md:grid-cols-3 gap-8 max-w-6xl mx-auto">
    <!-- Free Plan -->
    <div class="bg-white rounded-xl shadow-sm border-2 border-gray-200 p-8">
      <div class="mb-6">
        <h3 class="text-2xl font-bold text-gray-900 mb-2">Free</h3>
        <p class="text-gray-600">Perfect for trying out</p>
      </div>
      <div class="mb-6">
        <span class="text-4xl font-bold text-gray-900">$0</span>
        <span class="text-gray-600">/month</span>
      </div>
      <ul class="space-y-3 mb-8">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">100 links per month</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Basic analytics</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Standard support</span></li>
      </ul>
      <a href="/signup" class="block w-full text-center py-3 px-6 bg-gray-100 text-gray-900 font-semibold rounded-lg hover:bg-gray-200 transition-colors">
        Get Started
      </a>
    </div>

    <!-- Pro Plan -->
    <div class="bg-white rounded-xl shadow-lg border-2 border-primary-600 p-8 relative">
      <div class="absolute top-0 left-1/2 transform -translate-x-1/2 -translate-y-1/2">
        <span class="bg-primary-600 text-white text-sm font-semibold px-4 py-1 rounded-full">Most Popular</span>
      </div>
      <div class="mb-6">
        <h3 class="text-2xl font-bold text-gray-900 mb-2">Pro</h3>
        <p class="text-gray-600">For professionals</p>
      </div>
      <div class="mb-6">
        <span class="text-4xl font-bold text-gray-900">$19</span>
        <span class="text-gray-600">/month</span>
      </div>
      <ul class="space-y-3 mb-8">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">10,000 links per month</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Advanced analytics</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Custom domains</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Priority support</span></li>
      </ul>
      <a href="/signup?plan=pro" class="block w-full text-center py-3 px-6 bg-primary-600 text-white font-semibold rounded-lg hover:bg-primary-700 transition-colors">
        Start Free Trial
      </a>
    </div>

    <!-- Enterprise Plan -->
    <div class="bg-white rounded-xl shadow-sm border-2 border-gray-200 p-8">
      <div class="mb-6">
        <h3 class="text-2xl font-bold text-gray-900 mb-2">Enterprise</h3>
        <p class="text-gray-600">For large teams</p>
      </div>
      <div class="mb-6">
        <span class="text-4xl font-bold text-gray-900">Custom</span>
      </div>
      <ul class="space-y-3 mb-8">
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Unlimited links</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Full analytics suite</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Multiple custom domains</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">Dedicated support</span></li>
        <li class="flex items-start"><span class="text-green-500 mr-2">✓</span><span class="text-gray-700">SLA guarantee</span></li>
      </ul>
      <a href="/contact?type=enterprise" class="block w-full text-center py-3 px-6 bg-gray-100 text-gray-900 font-semibold rounded-lg hover:bg-gray-200 transition-colors">
        Contact Sales
      </a>
    </div>
  </div>',
  'pricing',
  true,
  70
);

-- ============================================================================
-- API DOCS PAGE CONTENT BLOCKS
-- ============================================================================

INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order) VALUES
(
  'API Docs - Hero Section',
  'api-docs-hero',
  '<div class="text-center space-y-4 mb-12">
    <h1 class="text-4xl md:text-5xl font-bold text-gray-900">
      <span class="text-primary-600">API</span> Documentation
    </h1>
    <p class="text-xl text-gray-600 max-w-3xl mx-auto">
      Complete API reference for developers. Build powerful integrations with our RESTful API.
    </p>
  </div>',
  'hero',
  true,
  80
),
(
  'API Docs - Quick Start',
  'api-docs-quickstart',
  '<div class="bg-white rounded-xl shadow-sm border border-gray-200 p-8 mb-8">
    <h2 class="text-2xl font-bold text-gray-900 mb-4">Quick Start</h2>
    <p class="text-gray-600 mb-6">
      Get started with the Trunc API in minutes. First, obtain your API key from your dashboard.
    </p>
    <div class="bg-gray-900 text-gray-100 rounded-lg p-6 overflow-x-auto">
      <code class="text-sm">
# Get your API key<br/>
curl -X POST https://trunc.link/api/v1/shorten \\<br/>
  -H "Authorization: Bearer YOUR_API_KEY" \\<br/>
  -H "Content-Type: application/json" \\<br/>
  -d ''{"url": "https://example.com/very-long-url"}''
      </code>
    </div>
  </div>',
  'code',
  true,
  90
),
(
  'API Docs - Endpoints Overview',
  'api-docs-endpoints',
  '<div class="space-y-6">
    <h2 class="text-2xl font-bold text-gray-900 mb-6">API Endpoints</h2>

    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div class="flex items-center mb-3">
        <span class="bg-green-100 text-green-800 text-xs font-semibold px-2.5 py-1 rounded mr-3">POST</span>
        <code class="text-sm font-mono text-gray-900">/api/v1/shorten</code>
      </div>
      <p class="text-gray-600 mb-3">Create a new short URL</p>
      <details class="text-sm text-gray-700">
        <summary class="cursor-pointer font-medium hover:text-primary-600">View details</summary>
        <div class="mt-3 pl-4 border-l-2 border-gray-200">
          <p class="mb-2"><strong>Request Body:</strong></p>
          <pre class="bg-gray-50 p-3 rounded text-xs overflow-x-auto"><code>{
  "url": "https://example.com",
  "custom_alias": "optional-alias"
}</code></pre>
        </div>
      </details>
    </div>

    <div class="bg-white rounded-lg shadow-sm border border-gray-200 p-6">
      <div class="flex items-center mb-3">
        <span class="bg-blue-100 text-blue-800 text-xs font-semibold px-2.5 py-1 rounded mr-3">GET</span>
        <code class="text-sm font-mono text-gray-900">/api/v1/analytics/:shortCode</code>
      </div>
      <p class="text-gray-600 mb-3">Get analytics for a short URL</p>
      <details class="text-sm text-gray-700">
        <summary class="cursor-pointer font-medium hover:text-primary-600">View details</summary>
        <div class="mt-3 pl-4 border-l-2 border-gray-200">
          <p class="mb-2"><strong>Response:</strong></p>
          <pre class="bg-gray-50 p-3 rounded text-xs overflow-x-auto"><code>{
  "clicks": 1234,
  "countries": {...},
  "devices": {...}
}</code></pre>
        </div>
      </details>
    </div>
  </div>',
  'info',
  true,
  100
);

-- ============================================================================
-- UPDATE PRIVACY PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-12">
  {{block:privacy-data-collection-summary}}

  {{block:privacy-data-usage-summary}}

  {{block:privacy-quick-actions}}

  <section class="prose prose-lg max-w-none">
    <h2>Full Privacy Policy</h2>
    <p>For complete details about how we handle your data, please review the sections below.</p>
  </section>

  {{block:privacy-terms-contact-cta}}
</div>',
updated_at = NOW()
WHERE slug = 'privacy';

-- ============================================================================
-- UPDATE TERMS PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-12">
  {{block:terms-service-overview}}

  {{block:terms-prohibited-content}}

  {{block:terms-link-management}}

  {{block:terms-api-usage}}

  {{block:terms-sla}}

  {{block:privacy-terms-contact-cta}}
</div>',
updated_at = NOW()
WHERE slug = 'terms';

-- ============================================================================
-- UPDATE HOME PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-16">
  {{block:home-hero}}

  <section>
    <h2 class="text-3xl font-bold text-gray-900 text-center mb-10">Key Features</h2>
    {{block:home-features-list}}
  </section>

  {{block:home-cta}}
</div>',
updated_at = NOW()
WHERE slug = 'home';

-- ============================================================================
-- UPDATE FEATURES PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-16">
  {{block:features-hero}}

  {{block:features-core-grid}}

  <section class="bg-gray-50 rounded-2xl p-12 text-center">
    <h2 class="text-2xl font-bold text-gray-900 mb-4">Ready to experience these features?</h2>
    <p class="text-gray-600 mb-8">Start your free trial today and see the difference</p>
    <a href="/signup" class="inline-flex items-center justify-center px-8 py-3 bg-primary-600 text-white font-semibold rounded-lg hover:bg-primary-700 transition-colors">
      Get Started Free
    </a>
  </section>
</div>',
updated_at = NOW()
WHERE slug = 'features';

-- ============================================================================
-- UPDATE PRICING PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-16">
  {{block:pricing-hero}}

  {{block:pricing-plans}}

  <section class="bg-primary-50 rounded-2xl p-12 text-center">
    <h2 class="text-2xl font-bold text-gray-900 mb-4">All plans include</h2>
    <div class="grid grid-cols-1 md:grid-cols-3 gap-6 max-w-3xl mx-auto mt-8">
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>Custom aliases</div>
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>Link expiration</div>
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>QR code generation</div>
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>API access</div>
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>Password protection</div>
      <div class="text-gray-700"><span class="text-green-500 mr-2">✓</span>Link rotation</div>
    </div>
  </section>
</div>',
updated_at = NOW()
WHERE slug = 'pricing';

-- ============================================================================
-- UPDATE API DOCS PAGE TO USE CONTENT BLOCKS
-- ============================================================================

UPDATE static_pages
SET content = '<div class="space-y-12">
  {{block:api-docs-hero}}

  {{block:api-docs-quickstart}}

  {{block:api-docs-endpoints}}

  <section class="bg-gray-50 rounded-xl p-8">
    <h2 class="text-2xl font-bold text-gray-900 mb-4">Need Help?</h2>
    <p class="text-gray-600 mb-6">
      Our support team is here to help you integrate the API. Contact us for assistance or check out our detailed guides.
    </p>
    <div class="flex gap-4">
      <a href="/contact" class="inline-flex items-center justify-center px-6 py-2 bg-primary-600 text-white font-semibold rounded-lg hover:bg-primary-700 transition-colors">
        Contact Support
      </a>
      <a href="/docs" class="inline-flex items-center justify-center px-6 py-2 bg-white text-gray-900 font-semibold rounded-lg border border-gray-300 hover:bg-gray-50 transition-colors">
        View Guides
      </a>
    </div>
  </section>
</div>',
updated_at = NOW()
WHERE slug = 'api-docs';
