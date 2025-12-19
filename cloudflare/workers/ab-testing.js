/**
 * CloudFlare Worker: A/B Testing at Edge
 *
 * Implements A/B testing for landing pages and features without origin server overhead.
 * Supports multiple experiments with weighted traffic splitting.
 */

// Configuration for A/B tests
const AB_TESTS = {
  'landing-page-v2': {
    enabled: true,
    variants: {
      'control': { weight: 50, path: '/' },
      'variant-a': { weight: 25, path: '/variants/landing-v2a' },
      'variant-b': { weight: 25, path: '/variants/landing-v2b' },
    },
    applyTo: ['/'],  // URLs to apply this test to
    cookieName: 'ab_landing_page',
    cookieMaxAge: 2592000,  // 30 days
  },
  'cta-button-color': {
    enabled: true,
    variants: {
      'blue': { weight: 50 },
      'green': { weight: 50 },
    },
    applyTo: ['/'],
    cookieName: 'ab_cta_color',
    cookieMaxAge: 2592000,
  },
  'pricing-layout': {
    enabled: false,  // Disabled test example
    variants: {
      'horizontal': { weight: 50 },
      'vertical': { weight: 50 },
    },
    applyTo: ['/pricing'],
    cookieName: 'ab_pricing_layout',
    cookieMaxAge: 2592000,
  },
};

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);

    // Find applicable A/B tests for this URL
    const applicableTests = Object.entries(AB_TESTS).filter(([_, config]) => {
      return config.enabled && config.applyTo.includes(url.pathname);
    });

    if (applicableTests.length === 0) {
      // No A/B tests for this URL, pass through
      return fetch(request);
    }

    // Get or assign variants for all applicable tests
    const cookies = parseCookies(request.headers.get('Cookie') || '');
    const assignments = {};
    const newCookies = [];

    for (const [testName, config] of applicableTests) {
      const cookieValue = cookies[config.cookieName];

      if (cookieValue && config.variants[cookieValue]) {
        // User already has a variant assigned
        assignments[testName] = cookieValue;
      } else {
        // Assign new variant based on weights
        const variant = assignVariant(config.variants);
        assignments[testName] = variant;

        // Prepare Set-Cookie header
        newCookies.push(
          `${config.cookieName}=${variant}; Path=/; Max-Age=${config.cookieMaxAge}; SameSite=Lax; Secure`
        );
      }
    }

    // Modify request with variant information
    const modifiedRequest = new Request(request);

    // Add headers for backend to read
    for (const [testName, variant] of Object.entries(assignments)) {
      modifiedRequest.headers.set(`X-AB-${testName}`, variant);
    }

    // Check if any test requires path rewriting
    for (const [testName, config] of applicableTests) {
      const variant = assignments[testName];
      const variantConfig = config.variants[variant];

      if (variantConfig.path && variantConfig.path !== url.pathname) {
        // Rewrite path for this variant
        url.pathname = variantConfig.path;
        break;  // Only apply one path rewrite
      }
    }

    // Fetch from origin (or rewritten path)
    const response = await fetch(new Request(url.toString(), modifiedRequest));

    // Add variant cookies to response
    if (newCookies.length > 0) {
      const modifiedResponse = new Response(response.body, response);

      for (const cookie of newCookies) {
        modifiedResponse.headers.append('Set-Cookie', cookie);
      }

      // Add header indicating active experiments
      modifiedResponse.headers.set(
        'X-AB-Tests',
        Object.entries(assignments).map(([test, variant]) => `${test}=${variant}`).join(';')
      );

      return modifiedResponse;
    }

    return response;
  }
};

/**
 * Assign a variant based on weighted distribution
 */
function assignVariant(variants) {
  const totalWeight = Object.values(variants).reduce((sum, v) => sum + v.weight, 0);
  let random = Math.random() * totalWeight;

  for (const [name, config] of Object.entries(variants)) {
    random -= config.weight;
    if (random <= 0) {
      return name;
    }
  }

  // Fallback to first variant (should never reach here)
  return Object.keys(variants)[0];
}

/**
 * Parse cookies from Cookie header string
 */
function parseCookies(cookieHeader) {
  const cookies = {};

  cookieHeader.split(';').forEach(cookie => {
    const [name, ...rest] = cookie.split('=');
    if (name && rest.length > 0) {
      cookies[name.trim()] = rest.join('=').trim();
    }
  });

  return cookies;
}

/**
 * Queue consumer to process A/B test results
 * This runs separately to aggregate A/B test metrics
 */
export async function queue(batch, env) {
  for (const message of batch.messages) {
    const { test_name, variant, event_type, user_id, timestamp } = message.body;

    // Store in KV or send to analytics service
    const key = `ab_results:${test_name}:${variant}:${event_type}`;

    // Increment counter
    const currentCount = parseInt(await env.AB_RESULTS_KV.get(key)) || 0;
    await env.AB_RESULTS_KV.put(key, (currentCount + 1).toString());

    // Acknowledge message
    message.ack();
  }
}
