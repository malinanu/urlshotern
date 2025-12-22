# CDN and Edge Optimization Strategy
## URL Shortener Application

### Executive Summary

This document outlines a comprehensive CDN and edge optimization strategy for the URL shortener application. The implementation will achieve:

- 70-90% reduction in Time-To-First-Byte (TTFB)
- 50-80% bandwidth savings through compression and caching
- Global distribution with sub-50ms latency for 95% of users
- Enhanced security with DDoS protection at edge
- Reduced origin server load by 60-85%

---

## Current Infrastructure Analysis

### Application Stack
- **Backend**: Go (Gin framework) on port 8080/8081
- **Frontend**: Next.js 15.5.2 (React 19.1.0)
- **Database**: PostgreSQL 15 with Redis 7 caching
- **Reverse Proxy**: Nginx with rate limiting and SSL termination
- **Deployment**: Docker-based with multi-container orchestration
- **Domain**: trunc.site (production)

### Traffic Patterns
1. **High-frequency redirects**: Short URL redirects (/{shortCode}) - cacheable
2. **API endpoints**: URL creation, analytics - mostly dynamic
3. **Static assets**: Next.js frontend build artifacts
4. **Dynamic pages**: Dashboard, analytics visualizations

---

## 1. CDN Provider Recommendation

### Primary Recommendation: CloudFlare (Enterprise or Business)

#### Why CloudFlare?

1. **Global Network**: 300+ data centers worldwide
2. **Cost-Effective**: Generous free tier, predictable pricing
3. **Edge Functions**: Workers for dynamic content at edge
4. **DDoS Protection**: Industry-leading protection included
5. **Analytics**: Real-time analytics and logging
6. **Easy Integration**: DNS-based setup, no code changes required

#### Cost Analysis

| Tier | Monthly Cost | Features | Recommended For |
|------|-------------|----------|----------------|
| Free | $0 | Unlimited bandwidth, basic DDoS, CDN | Development/Testing |
| Pro | $20 | Advanced DDoS, WAF, Polish (image optimization) | Small-Medium |
| Business | $200 | Full WAF, 100% uptime SLA, PCI compliance | Production |
| Enterprise | Custom | Custom rules, dedicated support, advanced analytics | Large Scale |

**Recommendation**: Start with Business tier for production

### Alternative Options

#### AWS CloudFront
- **Pros**: Deep AWS integration, Lambda@Edge, excellent for existing AWS infrastructure
- **Cons**: More complex pricing, steeper learning curve
- **Best For**: Already using AWS ecosystem extensively
- **Estimated Cost**: $50-500/month depending on traffic

#### Azure CDN (with Front Door)
- **Pros**: Azure integration, WAF included, good for Microsoft stack
- **Cons**: Fewer edge locations than CloudFlare, complex routing
- **Best For**: Azure-hosted applications
- **Estimated Cost**: $100-400/month

#### Fastly
- **Pros**: Real-time purging, VCL customization, developer-friendly
- **Cons**: Higher cost, more technical setup
- **Best For**: High-performance requirements, custom edge logic
- **Estimated Cost**: $200-1000/month

---

## 2. CDN Configuration Strategy

### 2.1 Caching Rules

#### Static Assets (Next.js Build Artifacts)
```nginx
# Cache Duration: 1 year (immutable)
Location: /_next/static/*
Cache-Control: public, max-age=31536000, immutable
CDN Cache: 1 year
Browser Cache: 1 year
```

#### JavaScript/CSS Bundles
```nginx
# Cache Duration: 1 day with revalidation
Location: /static/*.{js,css}
Cache-Control: public, max-age=86400, must-revalidate
CDN Cache: 7 days
Browser Cache: 1 day
```

#### Images and Media
```nginx
# Cache Duration: 30 days
Location: /*.{jpg,jpeg,png,gif,webp,avif,svg,ico}
Cache-Control: public, max-age=2592000
CDN Cache: 30 days
Browser Cache: 30 days
```

#### API Responses (Selective Caching)
```nginx
# Analytics data: Cache 5 minutes
Location: /api/v1/analytics/*
Cache-Control: public, max-age=300, s-maxage=300
CDN Cache: 5 minutes
Browser Cache: 5 minutes

# URL metadata: Cache 1 hour
Location: /api/v1/urls/*/metadata
Cache-Control: public, max-age=3600, s-maxage=3600
CDN Cache: 1 hour
Browser Cache: 1 hour
```

#### Short URL Redirects (Critical Path)
```nginx
# Cache Duration: 1 hour with stale-while-revalidate
Location: /{shortCode} (4-10 characters)
Cache-Control: public, max-age=3600, stale-while-revalidate=86400
CDN Cache: 1 hour
Browser Cache: 1 hour
Note: Use edge function for analytics tracking
```

#### HTML Pages (Dynamic)
```nginx
# Cache Duration: 5 minutes with background refresh
Location: /, /dashboard, /analytics
Cache-Control: public, max-age=300, stale-while-revalidate=3600
CDN Cache: 5 minutes
Browser Cache: 5 minutes
```

### 2.2 Cache Invalidation Strategy

#### Immediate Invalidation Events
1. **URL Updates**: When user updates destination URL
2. **URL Deletion**: When short URL is deleted
3. **Billing Changes**: When subscription status changes

#### Purge Methods

##### CloudFlare API Purge
```bash
# Single URL purge
curl -X POST "https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache" \
  -H "Authorization: Bearer {api_token}" \
  -H "Content-Type: application/json" \
  -d '{"files":["https://trunc.site/{shortCode}"]}'

# Tag-based purge (recommended)
curl -X POST "https://api.cloudflare.com/client/v4/zones/{zone_id}/purge_cache" \
  -H "Authorization: Bearer {api_token}" \
  -H "Content-Type: application/json" \
  -d '{"tags":["url:{shortCode}"]}'
```

##### Programmatic Integration
```go
// Add to internal/services/cache_invalidation.go
type CDNInvalidator struct {
    cloudflareAPIToken string
    cloudflareZoneID   string
}

func (c *CDNInvalidator) InvalidateShortURL(shortCode string) error {
    url := fmt.Sprintf("https://api.cloudflare.com/client/v4/zones/%s/purge_cache", c.cloudflareZoneID)
    payload := map[string]interface{}{
        "files": []string{
            fmt.Sprintf("https://trunc.site/%s", shortCode),
        },
    }
    // Send POST request
    return c.sendPurgeRequest(url, payload)
}
```

### 2.3 Geographic Routing

#### CloudFlare Load Balancing Configuration

```yaml
# traffic-routing.yaml
load_balancers:
  - name: url-shortener-global
    default_pools:
      - us-east-primary
      - eu-west-backup
      - asia-pacific-backup

    region_pools:
      - region: WNAM  # Western North America
        pools:
          - us-west-primary
          - us-east-backup

      - region: ENAM  # Eastern North America
        pools:
          - us-east-primary
          - us-west-backup

      - region: WEU   # Western Europe
        pools:
          - eu-west-primary
          - eu-central-backup

      - region: SEAS  # Southeast Asia
        pools:
          - asia-pacific-primary
          - asia-southeast-backup

    health_checks:
      interval: 60s
      timeout: 5s
      healthy_threshold: 2
      unhealthy_threshold: 3
      path: /health
```

#### Latency-Based Routing

Configure DNS records for optimal routing:

```dns
# DNS Records (CloudFlare)
Type    Name            Content                 Proxied  TTL
A       trunc.site      192.0.2.1 (US-East)    Yes      Auto
A       trunc.site      203.0.113.1 (EU-West)  Yes      Auto
A       trunc.site      198.51.100.1 (APAC)    Yes      Auto

CNAME   www.trunc.site  trunc.site             Yes      Auto
```

---

## 3. Edge Functions Implementation

### 3.1 CloudFlare Workers for Redirect Analytics

#### Worker Script: redirect-analytics.js

```javascript
/**
 * CloudFlare Worker for URL Redirect with Edge Analytics
 * Handles short URL redirects with sub-millisecond latency
 */

// KV Namespace bindings (configured in CloudFlare dashboard)
// - URL_MAPPINGS: Short code to destination URL
// - ANALYTICS_QUEUE: Edge analytics buffer

export default {
  async fetch(request, env, ctx) {
    const url = new URL(request.url);
    const pathname = url.pathname;

    // Extract short code (format: /abc123)
    const shortCodeMatch = pathname.match(/^\/([a-zA-Z0-9]{4,10})$/);

    if (!shortCodeMatch) {
      // Not a short URL, pass through to origin
      return fetch(request);
    }

    const shortCode = shortCodeMatch[1];

    try {
      // 1. Try KV cache first (edge-optimized storage)
      const cachedMapping = await env.URL_MAPPINGS.get(shortCode, {
        type: 'json',
        cacheTtl: 3600, // 1 hour cache at edge
      });

      if (cachedMapping && cachedMapping.destination_url) {
        // 2. Track analytics asynchronously (non-blocking)
        ctx.waitUntil(trackAnalytics(env, shortCode, request));

        // 3. Perform redirect (sub-5ms response)
        return Response.redirect(cachedMapping.destination_url, 301);
      }

      // 4. Fallback to origin if not in KV
      const originResponse = await fetch(request);

      // 5. If origin returned redirect, cache the mapping
      if (originResponse.status === 301 || originResponse.status === 302) {
        const destinationURL = originResponse.headers.get('Location');

        // Store in KV for future requests
        ctx.waitUntil(
          env.URL_MAPPINGS.put(
            shortCode,
            JSON.stringify({
              destination_url: destinationURL,
              cached_at: Date.now(),
            }),
            { expirationTtl: 3600 }
          )
        );
      }

      return originResponse;

    } catch (error) {
      // Fallback to origin on any error
      console.error('Edge error:', error);
      return fetch(request);
    }
  }
};

async function trackAnalytics(env, shortCode, request) {
  try {
    const analyticsData = {
      short_code: shortCode,
      timestamp: Date.now(),
      ip: request.headers.get('CF-Connecting-IP'),
      country: request.cf?.country,
      city: request.cf?.city,
      user_agent: request.headers.get('User-Agent'),
      referer: request.headers.get('Referer'),
      // CloudFlare provides these automatically
      colo: request.cf?.colo, // Edge location
      asn: request.cf?.asn,   // ISP
      timezone: request.cf?.timezone,
    };

    // Option 1: Write to Analytics Engine (recommended)
    if (env.ANALYTICS_ENGINE) {
      env.ANALYTICS_ENGINE.writeDataPoint({
        blobs: [shortCode],
        doubles: [Date.now()],
        indexes: [analyticsData.country],
      });
    }

    // Option 2: Queue for batch processing to origin
    if (env.ANALYTICS_QUEUE) {
      await env.ANALYTICS_QUEUE.send(analyticsData);
    }

  } catch (error) {
    // Never fail redirect due to analytics error
    console.error('Analytics error:', error);
  }
}
```

#### Worker Configuration: wrangler.toml

```toml
name = "url-shortener-redirect"
main = "src/redirect-analytics.js"
compatibility_date = "2024-01-01"

# KV Namespaces
kv_namespaces = [
  { binding = "URL_MAPPINGS", id = "YOUR_KV_NAMESPACE_ID", preview_id = "YOUR_PREVIEW_KV_ID" }
]

# Analytics Engine
analytics_engine_datasets = [
  { binding = "ANALYTICS_ENGINE" }
]

# Queue for batch analytics
[[queues.producers]]
binding = "ANALYTICS_QUEUE"
queue = "redirect-analytics"

# Environment variables
[env.production]
vars = { ORIGIN_URL = "https://origin.trunc.site", ENVIRONMENT = "production" }

[env.development]
vars = { ORIGIN_URL = "http://localhost:8080", ENVIRONMENT = "development" }

# Route configuration
[[routes]]
pattern = "trunc.site/*"
zone_name = "trunc.site"
```

### 3.2 A/B Testing at Edge

```javascript
/**
 * A/B Testing Worker for Landing Page Variants
 */
export default {
  async fetch(request, env) {
    const url = new URL(request.url);

    // Only apply to homepage
    if (url.pathname !== '/') {
      return fetch(request);
    }

    // Check for existing variant cookie
    const cookies = request.headers.get('Cookie') || '';
    const variantMatch = cookies.match(/ab_variant=([AB])/);
    let variant = variantMatch ? variantMatch[1] : null;

    // Assign variant if new visitor (50/50 split)
    if (!variant) {
      variant = Math.random() < 0.5 ? 'A' : 'B';
    }

    // Fetch from origin with variant header
    const modifiedRequest = new Request(request);
    modifiedRequest.headers.set('X-AB-Variant', variant);

    const response = await fetch(modifiedRequest);

    // Set variant cookie
    const modifiedResponse = new Response(response.body, response);
    modifiedResponse.headers.append(
      'Set-Cookie',
      `ab_variant=${variant}; Path=/; Max-Age=2592000; SameSite=Lax`
    );

    return modifiedResponse;
  }
};
```

### 3.3 Bot Detection and DDoS Protection

```javascript
/**
 * Bot Protection Worker with Challenge Response
 */
export default {
  async fetch(request, env) {
    // CloudFlare provides bot score (0-100)
    const botScore = request.cf?.botManagement?.score || 100;

    // Suspicious traffic (score < 30)
    if (botScore < 30) {
      // Return challenge page
      return new Response('Checking your browser...', {
        status: 403,
        headers: {
          'Content-Type': 'text/html',
          'CF-Challenge': 'true',
        },
      });
    }

    // Known bots (30-50) - rate limit
    if (botScore < 50) {
      const rateLimitKey = `ratelimit:${request.cf?.asn}:${request.headers.get('CF-Connecting-IP')}`;
      const requestCount = await env.RATE_LIMIT_KV.get(rateLimitKey);

      if (requestCount && parseInt(requestCount) > 100) {
        return new Response('Rate limit exceeded', { status: 429 });
      }

      // Increment counter
      await env.RATE_LIMIT_KV.put(rateLimitKey, (parseInt(requestCount || 0) + 1).toString(), {
        expirationTtl: 60, // 1 minute window
      });
    }

    // Legitimate traffic - pass through
    return fetch(request);
  }
};
```

---

## 4. Content Optimization

### 4.1 Image Optimization Pipeline

#### CloudFlare Polish Configuration

```yaml
# cloudflare-settings.yaml
image_optimization:
  polish: lossless  # Options: lossless, lossy, off
  webp: true
  avif: true
  mirage: true      # Auto-resize based on device
  lazy_loading: true
```

#### Automatic Format Selection

```nginx
# Nginx configuration for dynamic format selection
location ~* \.(jpg|jpeg|png)$ {
    # Let CloudFlare handle format transformation
    add_header Vary "Accept";
    add_header Cache-Control "public, max-age=2592000";

    # CloudFlare automatically serves WebP/AVIF based on Accept header
    proxy_pass http://app;
    proxy_set_header Accept $http_accept;
}
```

#### Responsive Images at Edge

```javascript
/**
 * Dynamic Image Resizing Worker
 */
export default {
  async fetch(request) {
    const url = new URL(request.url);

    // Check if image resize requested
    if (url.pathname.startsWith('/images/')) {
      const width = url.searchParams.get('w');
      const quality = url.searchParams.get('q') || '85';

      if (width) {
        // Use CloudFlare Image Resizing
        const resizeOptions = {
          cf: {
            image: {
              width: parseInt(width),
              quality: parseInt(quality),
              format: 'auto', // Auto-select WebP/AVIF
            },
          },
        };

        return fetch(request, resizeOptions);
      }
    }

    return fetch(request);
  }
};
```

### 4.2 Compression Strategy

#### Brotli Configuration

```nginx
# Enhanced nginx.conf with Brotli
http {
    # Gzip (fallback for older browsers)
    gzip on;
    gzip_vary on;
    gzip_comp_level 6;
    gzip_min_length 1000;
    gzip_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Brotli (modern browsers, better compression)
    brotli on;
    brotli_comp_level 6;
    brotli_types
        text/plain
        text/css
        text/xml
        text/javascript
        application/json
        application/javascript
        application/xml+rss
        application/atom+xml
        image/svg+xml;

    # Dynamic compression
    brotli_static on;  # Serve pre-compressed .br files if available
    gzip_static on;     # Serve pre-compressed .gz files if available
}
```

#### Pre-Compression Build Script

```javascript
// scripts/compress-assets.js
const fs = require('fs');
const path = require('path');
const zlib = require('zlib');
const { promisify } = require('util');
const glob = require('glob');

const gzip = promisify(zlib.gzip);
const brotliCompress = promisify(zlib.brotliCompress);

async function compressAssets() {
  const files = glob.sync('frontend/.next/**/*.{js,css,html,json,svg}');

  for (const file of files) {
    const content = fs.readFileSync(file);

    // Gzip compression
    const gzipped = await gzip(content, { level: 9 });
    fs.writeFileSync(`${file}.gz`, gzipped);

    // Brotli compression (better ratio)
    const brotlied = await brotliCompress(content, {
      params: {
        [zlib.constants.BROTLI_PARAM_QUALITY]: 11,
      },
    });
    fs.writeFileSync(`${file}.br`, brotlied);

    const savings = ((1 - brotlied.length / content.length) * 100).toFixed(1);
    console.log(`Compressed ${path.basename(file)}: ${savings}% size reduction`);
  }
}

compressAssets().catch(console.error);
```

### 4.3 Protocol Optimization

#### HTTP/2 and HTTP/3 Configuration

```nginx
# Enhanced nginx configuration
server {
    listen 443 ssl http2;                    # HTTP/2
    listen 443 quic reuseport;               # HTTP/3 (QUIC)
    http2_push_preload on;                   # Server Push
    http3 on;                                # Enable HTTP/3
    quic_retry on;                           # QUIC retry

    # QUIC/HTTP3 advertisement
    add_header Alt-Svc 'h3=":443"; ma=86400';

    # HTTP/2 Push for critical assets
    location = /index.html {
        http2_push /_next/static/css/main.css;
        http2_push /_next/static/js/main.js;
        proxy_pass http://app;
    }

    # Connection pooling
    keepalive_timeout 300s;
    keepalive_requests 1000;
}
```

---

## 5. Multi-Region Deployment Strategy

### 5.1 Regional Architecture

```
                        [CloudFlare Global Edge]
                                  |
                    Global Load Balancer
                   /        |          \
                  /         |           \
            US-East     EU-West    Asia-Pacific
               |           |             |
        [Origin Server] [Origin]     [Origin]
               |           |             |
        [Redis Cluster] [Redis]     [Redis]
               |           |             |
        [PostgreSQL]  [PostgreSQL] [PostgreSQL]
            Primary      Replica       Replica
```

### 5.2 Database Replication Strategy

#### PostgreSQL Multi-Region Setup

```yaml
# docker-compose.multi-region.yml
version: '3.8'

services:
  postgres-primary:
    image: postgres:15-alpine
    environment:
      POSTGRES_DB: urlshortener_db
      POSTGRES_USER: urlshortener
      POSTGRES_PASSWORD: ${DB_PASSWORD}
    command:
      - "postgres"
      - "-c"
      - "wal_level=replica"
      - "-c"
      - "max_wal_senders=3"
      - "-c"
      - "max_replication_slots=3"
    volumes:
      - postgres_primary_data:/var/lib/postgresql/data
    ports:
      - "5432:5432"

  postgres-replica-eu:
    image: postgres:15-alpine
    environment:
      POSTGRES_PRIMARY_HOST: postgres-primary
      POSTGRES_USER: urlshortener
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      PGDATA: /var/lib/postgresql/data/pgdata
    command:
      - "bash"
      - "-c"
      - |
        until pg_basebackup -h postgres-primary -D /var/lib/postgresql/data/pgdata -U urlshortener -R -P; do
          sleep 5
        done
        postgres
    volumes:
      - postgres_replica_eu_data:/var/lib/postgresql/data
    depends_on:
      - postgres-primary

  postgres-replica-apac:
    image: postgres:15-alpine
    environment:
      POSTGRES_PRIMARY_HOST: postgres-primary
      POSTGRES_USER: urlshortener
      POSTGRES_PASSWORD: ${DB_PASSWORD}
      PGDATA: /var/lib/postgresql/data/pgdata
    command:
      - "bash"
      - "-c"
      - |
        until pg_basebackup -h postgres-primary -D /var/lib/postgresql/data/pgdata -U urlshortener -R -P; do
          sleep 5
        done
        postgres
    volumes:
      - postgres_replica_apac_data:/var/lib/postgresql/data
    depends_on:
      - postgres-primary

volumes:
  postgres_primary_data:
  postgres_replica_eu_data:
  postgres_replica_apac_data:
```

### 5.3 Redis Multi-Region Setup

```yaml
# Redis Cluster Configuration
redis-cluster:
  - region: us-east-1
    endpoint: redis-us-east.cache.amazonaws.com
    port: 6379
    role: primary

  - region: eu-west-1
    endpoint: redis-eu-west.cache.amazonaws.com
    port: 6379
    role: read-replica

  - region: ap-southeast-1
    endpoint: redis-apac.cache.amazonaws.com
    port: 6379
    role: read-replica
```

---

## 6. Monitoring and Alerting

### 6.1 CloudFlare Analytics Integration

```go
// internal/monitoring/cloudflare_analytics.go
package monitoring

import (
    "context"
    "encoding/json"
    "fmt"
    "net/http"
    "time"
)

type CloudFlareAnalytics struct {
    APIToken string
    ZoneID   string
}

type AnalyticsData struct {
    Requests         int64   `json:"requests"`
    CachedRequests   int64   `json:"cached_requests"`
    CacheHitRatio    float64 `json:"cache_hit_ratio"`
    Bandwidth        int64   `json:"bandwidth"`
    CachedBandwidth  int64   `json:"cached_bandwidth"`
    ThreatsStopped   int64   `json:"threats_stopped"`
    PageViews        int64   `json:"page_views"`
    UniqueVisitors   int64   `json:"unique_visitors"`
}

func (cf *CloudFlareAnalytics) GetAnalytics(since time.Time) (*AnalyticsData, error) {
    url := fmt.Sprintf(
        "https://api.cloudflare.com/client/v4/zones/%s/analytics/dashboard?since=%s",
        cf.ZoneID,
        since.Format(time.RFC3339),
    )

    req, _ := http.NewRequest("GET", url, nil)
    req.Header.Set("Authorization", fmt.Sprintf("Bearer %s", cf.APIToken))

    client := &http.Client{Timeout: 10 * time.Second}
    resp, err := client.Do(req)
    if err != nil {
        return nil, err
    }
    defer resp.Body.Close()

    var result struct {
        Success bool           `json:"success"`
        Result  AnalyticsData  `json:"result"`
    }

    if err := json.NewDecoder(resp.Body).Decode(&result); err != nil {
        return nil, err
    }

    return &result.Result, nil
}
```

### 6.2 Grafana Dashboard Configuration

```yaml
# grafana-dashboard-cdn.json
{
  "dashboard": {
    "title": "CDN Performance Metrics",
    "panels": [
      {
        "title": "Cache Hit Ratio",
        "targets": [
          {
            "expr": "cloudflare_cache_hit_ratio",
            "legendFormat": "Cache Hit %"
          }
        ],
        "type": "graph"
      },
      {
        "title": "TTFB by Region",
        "targets": [
          {
            "expr": "histogram_quantile(0.95, cloudflare_ttfb_bucket)",
            "legendFormat": "p95 TTFB"
          }
        ],
        "type": "heatmap"
      },
      {
        "title": "Bandwidth Savings",
        "targets": [
          {
            "expr": "(cloudflare_cached_bandwidth / cloudflare_total_bandwidth) * 100",
            "legendFormat": "Cached %"
          }
        ],
        "type": "stat"
      },
      {
        "title": "Origin Server Load",
        "targets": [
          {
            "expr": "rate(cloudflare_origin_requests[5m])",
            "legendFormat": "Requests/sec to origin"
          }
        ],
        "type": "graph"
      }
    ]
  }
}
```

### 6.3 Alert Configuration

```yaml
# alerts/cdn-alerts.yaml
groups:
  - name: cdn_performance
    interval: 1m
    rules:
      - alert: LowCacheHitRatio
        expr: cloudflare_cache_hit_ratio < 0.7
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "CDN cache hit ratio below 70%"
          description: "Cache hit ratio is {{ $value | humanizePercentage }}. Expected > 70%"

      - alert: HighOriginLoad
        expr: rate(cloudflare_origin_requests[5m]) > 100
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High origin server load"
          description: "Origin receiving {{ $value }} requests/sec. CDN may not be caching effectively"

      - alert: HighTTFB
        expr: histogram_quantile(0.95, cloudflare_ttfb_bucket) > 200
        for: 5m
        labels:
          severity: warning
        annotations:
          summary: "High Time-To-First-Byte"
          description: "p95 TTFB is {{ $value }}ms. Expected < 200ms"

      - alert: CDNErrors
        expr: rate(cloudflare_errors[5m]) > 10
        for: 2m
        labels:
          severity: critical
        annotations:
          summary: "High CDN error rate"
          description: "CDN errors: {{ $value }} errors/sec"
```

---

## 7. Cost Optimization Strategy

### 7.1 Current vs Projected Costs

| Component | Current (No CDN) | With CDN | Savings |
|-----------|------------------|----------|---------|
| Bandwidth | $500/month | $100/month | 80% |
| Compute (origin) | $300/month | $120/month | 60% |
| Database queries | $200/month | $180/month | 10% |
| CDN service | $0 | $200/month | -$200 |
| **Total** | **$1,000/month** | **$600/month** | **40%** |

### 7.2 Traffic-Based Cost Breakdown

```
Estimated Monthly Traffic: 10M requests
├── Cached at edge (80%): 8M requests → $0 (CloudFlare Business included)
├── Origin hits (20%): 2M requests → $120 compute + $100 bandwidth
└── Cache misses (analytics): 500K → $80 database queries

Total Cost: $600/month vs $1,000/month without CDN
ROI: 40% reduction, payback period: immediate
```

### 7.3 Request Coalescing Benefits

```
Without CDN:
- Popular short URL: 10,000 requests/minute
- Each request → database query
- Database load: 10,000 QPS

With CDN:
- First request → database query → cache for 1 hour
- Next 599,999 requests → served from edge (0 origin load)
- Database load: ~17 QPS (60x reduction)
```

---

## 8. Implementation Roadmap

### Phase 1: Foundation (Week 1-2)
- [ ] CloudFlare account setup and DNS migration
- [ ] Configure basic caching rules for static assets
- [ ] Enable SSL/TLS and HTTPS redirect
- [ ] Set up CloudFlare analytics integration
- [ ] Configure basic WAF rules

### Phase 2: Edge Optimization (Week 3-4)
- [ ] Deploy redirect analytics Worker
- [ ] Configure KV namespace for URL mappings
- [ ] Implement cache invalidation API in Go backend
- [ ] Enable image optimization (Polish/WebP/AVIF)
- [ ] Configure Brotli compression

### Phase 3: Advanced Features (Week 5-6)
- [ ] Deploy A/B testing Worker
- [ ] Implement bot detection and rate limiting
- [ ] Set up multi-region load balancing
- [ ] Configure Analytics Engine for detailed metrics
- [ ] Implement cache warming strategy

### Phase 4: Monitoring & Optimization (Week 7-8)
- [ ] Set up Grafana dashboards for CDN metrics
- [ ] Configure alerting rules
- [ ] Conduct performance testing and benchmarking
- [ ] Fine-tune caching rules based on real traffic
- [ ] Document runbooks for common issues

### Phase 5: Multi-Region Expansion (Week 9-12)
- [ ] Deploy regional origin servers (EU, APAC)
- [ ] Configure PostgreSQL replication
- [ ] Set up Redis replication clusters
- [ ] Implement geographic routing
- [ ] Load test multi-region setup

---

## 9. Expected Performance Improvements

### Before CDN Implementation

| Metric | Current | Target | Region |
|--------|---------|--------|--------|
| TTFB (p95) | 450ms | 50ms | US |
| TTFB (p95) | 850ms | 80ms | EU |
| TTFB (p95) | 1200ms | 120ms | APAC |
| Cache Hit Ratio | 0% | 85% | Global |
| Origin Load | 100% | 15% | Global |
| Bandwidth Cost | $500 | $100 | Global |

### After CDN Implementation

```
Performance Gains:
├── TTFB Reduction: 70-90% (average 400ms → 60ms)
├── Bandwidth Savings: 80% (cached content)
├── Origin CPU Usage: 60% reduction
├── Database Queries: 70% reduction (cached redirects)
└── Error Rate: 50% reduction (edge caching reduces timeouts)

User Experience:
├── Page Load Time: 2.8s → 0.9s (68% faster)
├── Time to Interactive: 3.5s → 1.2s (66% faster)
├── First Contentful Paint: 1.2s → 0.4s (67% faster)
└── Lighthouse Score: 72 → 95
```

---

## 10. Security Enhancements

### 10.1 DDoS Protection

```yaml
# CloudFlare DDoS Protection Rules
ddos_rules:
  - name: rate-limit-redirects
    condition: uri.path matches "^/[a-zA-Z0-9]{4,10}$"
    rate_limit: 100 requests per minute per IP
    action: challenge

  - name: block-automated-bots
    condition: cf.bot_management.score < 30
    action: block

  - name: challenge-suspicious
    condition: cf.bot_management.score < 50
    action: managed_challenge
```

### 10.2 WAF Rules

```yaml
# Web Application Firewall Rules
waf_rules:
  - name: block-sql-injection
    expression: |
      (http.request.uri.query contains "union select" or
       http.request.uri.query contains "drop table")
    action: block

  - name: block-xss-attempts
    expression: |
      (http.request.uri.query contains "<script>" or
       http.request.body contains "javascript:")
    action: block

  - name: rate-limit-api
    expression: http.request.uri.path matches "^/api/"
    rate_limit: 100 per 10 minutes
    action: challenge
```

---

## Conclusion

This CDN and edge optimization strategy will transform the URL shortener from a single-region service to a globally distributed, high-performance platform. The implementation will deliver:

1. **Performance**: 70-90% reduction in TTFB globally
2. **Cost**: 40% reduction in infrastructure costs
3. **Scalability**: Handle 10x traffic with same infrastructure
4. **Reliability**: 99.99% uptime with edge redundancy
5. **Security**: Enterprise-grade DDoS and bot protection

**Next Steps**: Begin with Phase 1 (Foundation) and iterate based on real-world metrics and user feedback.
