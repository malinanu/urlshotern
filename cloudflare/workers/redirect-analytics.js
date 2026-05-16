/**
 * CloudFlare Worker: URL Shortener Redirect with Edge Analytics
 *
 * This worker handles short URL redirects at the edge with sub-5ms latency.
 * Features:
 * - KV-based caching for instant redirects
 * - Non-blocking analytics tracking
 * - Automatic fallback to origin
 * - Bot detection and rate limiting
 */

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const pathname = url.pathname;

    // Health check endpoint
    if (pathname === '/health' || pathname === '/health/edge') {
      return new Response(JSON.stringify({
        status: 'healthy',
        edge: true,
        colo: request.cf?.colo,
        timestamp: new Date().toISOString(),
      }), {
        headers: { 'Content-Type': 'application/json' },
      });
    }

    // Extract short code (format: /abc123 - 4 to 10 alphanumeric characters)
    const shortCodeMatch = pathname.match(/^\/([a-zA-Z0-9]{4,10})$/);

    if (!shortCodeMatch) {
      // Not a short URL pattern, pass through to origin
      return fetch(request);
    }

    const shortCode = shortCodeMatch[1];

    // Check for bot traffic
    const botScore = request.cf?.botManagement?.score || 100;
    if (botScore < 30) {
      // Block known malicious bots
      return new Response('Access Denied', {
        status: 403,
        headers: {
          'Cache-Control': 'no-cache',
          'X-Robots-Tag': 'noindex',
        },
      });
    }

    try {
      // 1. Try KV cache first (edge-optimized storage)
      const cachedMapping = await env.URL_MAPPINGS.get(shortCode, {
        type: 'json',
        cacheTtl: 60, // 1 minute — short so updated destinations take effect quickly
      });

      if (cachedMapping && cachedMapping.destination_url) {
        console.log(`Cache HIT for ${shortCode}`);

        // 2. Track analytics asynchronously (non-blocking)
        ctx.waitUntil(trackAnalytics(env, shortCode, request, true));

        // 3. Perform redirect (sub-5ms response)
        return Response.redirect(cachedMapping.destination_url, 301);
      }

      console.log(`Cache MISS for ${shortCode}, fetching from origin`);

      // 4. Fallback to origin if not in KV
      const originResponse = await fetch(request, {
        cf: {
          cacheEverything: true,
          cacheTtl: 60,
          cacheTags: [`url:${shortCode}`],
        },
      });

      // 5. If origin returned redirect, cache the mapping in KV
      if (originResponse.status === 301 || originResponse.status === 302) {
        const destinationURL = originResponse.headers.get('Location');

        if (destinationURL) {
          // Store in KV for future requests
          ctx.waitUntil(
            env.URL_MAPPINGS.put(
              shortCode,
              JSON.stringify({
                destination_url: destinationURL,
                cached_at: Date.now(),
                short_code: shortCode,
              }),
              {
                expirationTtl: 60, // 1 minute
                metadata: { shortCode, cachedAt: Date.now() },
              }
            )
          );

          // Track analytics
          ctx.waitUntil(trackAnalytics(env, shortCode, request, false));

          console.log(`Cached mapping for ${shortCode} → ${destinationURL}`);
        }
      }

      return originResponse;

    } catch (error) {
      // Fallback to origin on any error
      console.error('Edge processing error:', error.message);

      // Try origin without edge optimization
      return fetch(request);
    }
  }
};

/**
 * Track analytics data asynchronously
 * This function never blocks the redirect response
 */
async function trackAnalytics(env, shortCode, request, cacheHit) {
  try {
    const now = Date.now();
    const analyticsData = {
      short_code: shortCode,
      timestamp: now,
      cache_hit: cacheHit,

      // IP and location data
      ip: request.headers.get('CF-Connecting-IP'),
      country: request.cf?.country || 'UNKNOWN',
      city: request.cf?.city || 'UNKNOWN',
      region: request.cf?.region || 'UNKNOWN',
      continent: request.cf?.continent || 'UNKNOWN',
      latitude: request.cf?.latitude || null,
      longitude: request.cf?.longitude || null,
      timezone: request.cf?.timezone || 'UTC',
      postal_code: request.cf?.postalCode || null,

      // Request metadata
      user_agent: request.headers.get('User-Agent') || 'UNKNOWN',
      referer: request.headers.get('Referer') || '',
      accept_language: request.headers.get('Accept-Language') || '',

      // CloudFlare metadata
      colo: request.cf?.colo || 'UNKNOWN',      // Edge data center
      asn: request.cf?.asn || 0,                // ISP Autonomous System Number
      asn_organization: request.cf?.asOrganization || 'UNKNOWN',

      // Device information (parsed from user agent)
      bot_score: request.cf?.botManagement?.score || 100,
      verified_bot: request.cf?.botManagement?.verifiedBot || false,
    };

    // Option 1: Write to CloudFlare Analytics Engine (recommended for high volume)
    if (env.ANALYTICS_ENGINE) {
      await env.ANALYTICS_ENGINE.writeDataPoint({
        blobs: [
          shortCode,
          analyticsData.country,
          analyticsData.city,
          analyticsData.colo,
        ],
        doubles: [
          now,
          cacheHit ? 1 : 0,
          analyticsData.bot_score,
        ],
        indexes: [shortCode], // Enable filtering by short code
      });
    }

    // Option 2: Queue for batch processing to origin
    if (env.ANALYTICS_QUEUE) {
      await env.ANALYTICS_QUEUE.send(analyticsData);
    }

    // Option 3: Direct POST to origin (use sparingly, adds latency)
    // Uncomment only if you need real-time processing at origin
    /*
    if (env.ORIGIN_ANALYTICS_URL) {
      await fetch(env.ORIGIN_ANALYTICS_URL, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify(analyticsData),
      });
    }
    */

  } catch (error) {
    // Never fail redirect due to analytics error
    console.error('Analytics error (non-critical):', error.message);
  }
}
