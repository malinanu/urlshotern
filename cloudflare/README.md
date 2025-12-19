# CloudFlare CDN Integration
## Workers and Edge Computing

This directory contains CloudFlare Workers for edge computing and CDN optimization.

---

## Directory Structure

```
cloudflare/
├── workers/
│   ├── redirect-analytics.js    # Main redirect worker with analytics
│   ├── ab-testing.js            # A/B testing worker
│   ├── wrangler.toml            # Worker configuration
│   └── package.json             # Dependencies
├── configs/
│   └── cache-rules.json         # Cache configuration
└── scripts/
    ├── deploy.sh                # Deployment script
    └── test.sh                  # Testing script
```

---

## Workers Overview

### 1. Redirect Analytics Worker (`redirect-analytics.js`)

**Purpose**: Handles short URL redirects at the edge with sub-5ms latency and tracks analytics.

**Features**:
- KV-based URL caching for instant redirects
- Non-blocking analytics tracking
- Bot detection and filtering
- Automatic fallback to origin
- CloudFlare Analytics Engine integration

**Performance**:
- Response time: <5ms (cached)
- Cache hit ratio: 85-95%
- Global availability: 300+ data centers

**Usage**:
```bash
# Deploy
wrangler publish redirect-analytics.js --env production

# Test
curl -I https://trunc.site/abc123

# View logs
wrangler tail redirect-analytics --env production
```

---

### 2. A/B Testing Worker (`ab-testing.js`)

**Purpose**: Implements A/B testing at the edge without origin server load.

**Features**:
- Multiple experiment support
- Weighted traffic splitting
- Cookie-based variant persistence
- Path rewriting for different variants
- Non-intrusive implementation

**Configuration**:
```javascript
const AB_TESTS = {
  'landing-page-v2': {
    enabled: true,
    variants: {
      'control': { weight: 50, path: '/' },
      'variant-a': { weight: 25, path: '/variants/landing-v2a' },
      'variant-b': { weight: 25, path: '/variants/landing-v2b' },
    },
    applyTo: ['/'],
    cookieName: 'ab_landing_page',
    cookieMaxAge: 2592000,  // 30 days
  },
};
```

**Usage**:
```bash
# Deploy
wrangler publish ab-testing.js --env production

# Test
curl -I https://trunc.site/
# Check Set-Cookie header for variant assignment
```

---

## Setup Instructions

### Prerequisites

1. CloudFlare account with Business plan
2. Node.js 18+ installed
3. Wrangler CLI installed
4. Domain added to CloudFlare

### Installation

```bash
# Install Wrangler CLI globally
npm install -g @cloudflare/wrangler

# Login to CloudFlare
wrangler login

# Navigate to workers directory
cd cloudflare/workers

# Install dependencies (if any)
npm install
```

### Configuration

1. **Update wrangler.toml**:
   ```toml
   account_id = "your_account_id"
   zone_id = "your_zone_id"
   ```

2. **Create KV Namespace**:
   ```bash
   wrangler kv:namespace create "URL_MAPPINGS"
   # Copy the ID and update wrangler.toml
   ```

3. **Set Environment Variables**:
   ```bash
   cp ../../.env.cloudflare.example ../../.env.cloudflare
   # Edit .env.cloudflare with your values
   ```

### Deployment

```bash
# Deploy to production
wrangler publish redirect-analytics.js --env production

# Deploy to staging
wrangler publish redirect-analytics.js --env staging

# Deploy all workers
./scripts/deploy-all.sh
```

---

## Worker Architecture

### Request Flow

```
User Request
    ↓
CloudFlare Edge (300+ locations)
    ↓
Worker Execution
    ├─→ Check KV Cache (sub-1ms)
    │   ├─→ Cache HIT → Redirect (total: <5ms)
    │   └─→ Cache MISS ↓
    ↓
Fetch from Origin
    ↓
Store in KV Cache
    ↓
Return Response to User
    ↓
Track Analytics (async, non-blocking)
```

### Data Flow

```
Short URL Request (trunc.site/abc123)
    ↓
Worker checks KV store
    ├─→ Found: Return cached destination URL
    └─→ Not found: Query origin server
         ↓
    Store in KV for future requests
         ↓
    Track analytics data:
    - IP address, country, city
    - User agent, referer
    - Timestamp, cache status
    - Edge location (colo)
         ↓
    Send to Analytics Engine (batched)
         ↓
    Backend processes analytics later
```

---

## KV Storage Schema

### URL Mappings

**Key**: Short code (e.g., `abc123`)

**Value** (JSON):
```json
{
  "destination_url": "https://example.com/very/long/url",
  "cached_at": 1704672000000,
  "short_code": "abc123",
  "expires_at": 1704758400000
}
```

**TTL**: 1 hour (3600 seconds)

### Metadata

```json
{
  "version": "1.0",
  "last_updated": "2025-01-07T00:00:00Z"
}
```

---

## Analytics Data Structure

### Analytics Engine Schema

```javascript
{
  blobs: [
    shortCode,          // URL identifier
    country,            // Two-letter country code
    city,               // City name
    colo,              // CloudFlare data center code
  ],
  doubles: [
    timestamp,          // Unix timestamp (ms)
    cacheHit,          // 1 = cached, 0 = origin
    botScore,          // 0-100 (lower = more bot-like)
  ],
  indexes: [
    shortCode,         // Enable filtering by short code
  ]
}
```

### Processing

Analytics data is:
1. Collected at edge (non-blocking)
2. Batched and sent to Analytics Engine
3. Processed by queue consumer worker
4. Aggregated in PostgreSQL database
5. Displayed in dashboard

---

## Performance Optimization

### Cache Strategy

1. **Static Assets**: Cache for 1 year
   ```nginx
   Cache-Control: public, max-age=31536000, immutable
   ```

2. **Short URL Redirects**: Cache for 1 hour
   ```nginx
   Cache-Control: public, max-age=3600, stale-while-revalidate=86400
   ```

3. **API Responses**: Cache for 5 minutes or bypass
   ```nginx
   Cache-Control: public, max-age=300
   ```

### Worker Performance

- **CPU Time**: <5ms per request (target: <10ms)
- **Memory Usage**: <10MB (target: <50MB)
- **KV Read Latency**: <1ms (global average)
- **Origin Requests**: 5-15% of total (cache miss rate)

### Monitoring

```bash
# Real-time worker metrics
wrangler tail redirect-analytics --env production

# CPU usage
wrangler metrics redirect-analytics --env production

# Error rate
wrangler logs redirect-analytics --env production --status error
```

---

## Testing

### Local Testing

```bash
# Install Miniflare for local testing
npm install -g miniflare

# Run worker locally
miniflare redirect-analytics.js --kv URL_MAPPINGS

# Test with curl
curl http://localhost:8787/abc123
```

### Integration Testing

```bash
# Deploy to staging environment
wrangler publish redirect-analytics.js --env staging

# Run test suite
npm test

# Load testing
k6 run tests/load-test.js
```

### Performance Testing

```bash
# Test cache hit rate
for i in {1..100}; do
  curl -I https://trunc.site/abc123 2>&1 | grep cf-cache-status
done | sort | uniq -c

# Expected output:
# 1 cf-cache-status: MISS
# 99 cf-cache-status: HIT
```

---

## Troubleshooting

### Worker Not Executing

**Symptoms**: No X-Edge headers in response

**Solution**:
```bash
# Check worker route
wrangler routes list

# Verify worker is deployed
wrangler deployments list redirect-analytics

# Check worker logs
wrangler tail redirect-analytics --env production
```

### KV Store Not Working

**Symptoms**: All requests hit origin, no caching

**Solution**:
```bash
# Verify KV binding
cat wrangler.toml | grep -A 3 kv_namespaces

# Test KV directly
wrangler kv:key get --binding=URL_MAPPINGS "abc123"

# List all keys
wrangler kv:key list --binding=URL_MAPPINGS
```

### High CPU Usage

**Symptoms**: Worker exceeds CPU time limit

**Solution**:
1. Optimize KV reads (use cacheTtl)
2. Reduce JSON parsing operations
3. Move heavy computation to queue consumer
4. Use Analytics Engine instead of direct writes

---

## Security

### Bot Protection

```javascript
// In redirect-analytics.js
const botScore = request.cf?.botManagement?.score || 100;

if (botScore < 30) {
  // Block known malicious bots
  return new Response('Access Denied', { status: 403 });
}
```

### Rate Limiting

Configured in CloudFlare dashboard:
- API endpoints: 100 requests/10 minutes
- Redirects: 1000 requests/minute
- Login: 5 requests/minute

### DDoS Protection

CloudFlare automatically handles:
- Layer 3/4 attacks (SYN floods, UDP amplification)
- Layer 7 attacks (HTTP floods, slowloris)
- Distributed attacks across multiple IPs

---

## Cost Management

### Worker Costs

**Bundled Plan** ($5/month, included in Business plan):
- 10 million requests/month included
- 400,000 GB-seconds CPU time
- 10GB KV storage

**Additional Costs**:
- Extra requests: $0.50 per million
- Extra KV storage: $0.50 per GB
- Analytics Engine: $5 per million writes

### Cost Optimization

1. **Increase Cache TTL**: Reduce origin requests
2. **Batch Analytics**: Use queue consumer
3. **Optimize KV Reads**: Use cacheTtl parameter
4. **Monitor Usage**: Set up cost alerts

---

## Monitoring and Alerts

### CloudFlare Dashboard

1. Go to Workers & Pages → redirect-analytics
2. View metrics:
   - Requests per second
   - Error rate
   - CPU time (p50, p95, p99)
   - Success rate

### Custom Alerts

```yaml
alerts:
  - name: worker_error_rate
    condition: error_rate > 5%
    action: email_team

  - name: worker_cpu_time
    condition: p99_cpu_time > 50ms
    action: slack_alert

  - name: kv_quota
    condition: kv_usage > 80%
    action: email_admin
```

---

## Migration Guide

### From Origin-Only to CDN+Workers

1. **Deploy workers alongside origin**:
   ```bash
   wrangler publish redirect-analytics.js --env staging
   ```

2. **Test with staging domain**:
   ```bash
   curl https://staging.trunc.site/abc123
   ```

3. **Gradually shift traffic**:
   - Start with 10% traffic
   - Monitor error rates
   - Increase to 50%, then 100%

4. **Complete migration**:
   ```bash
   wrangler publish redirect-analytics.js --env production
   ```

---

## Best Practices

1. **Always Test Locally First**:
   ```bash
   miniflare redirect-analytics.js
   ```

2. **Use Staging Environment**:
   Deploy to staging before production

3. **Monitor Error Rates**:
   Set up alerts for >1% error rate

4. **Cache Aggressively**:
   Maximize cache hit ratio (target: >85%)

5. **Optimize KV Reads**:
   Use cacheTtl to reduce KV bandwidth

6. **Handle Errors Gracefully**:
   Always fall back to origin on worker errors

7. **Log Important Events**:
   Use console.log() for debugging (visible in wrangler tail)

8. **Version Your Workers**:
   Use git tags to track deployed versions

---

## Additional Resources

- [CloudFlare Workers Documentation](https://developers.cloudflare.com/workers/)
- [Wrangler CLI Reference](https://developers.cloudflare.com/workers/wrangler/)
- [Workers KV Documentation](https://developers.cloudflare.com/workers/runtime-apis/kv/)
- [Analytics Engine Documentation](https://developers.cloudflare.com/analytics/analytics-engine/)
- [Workers Examples](https://developers.cloudflare.com/workers/examples/)

---

## Support

For issues or questions:
1. Check CloudFlare Workers documentation
2. Search CloudFlare Community forum
3. Contact CloudFlare Business support (24/7)
4. Review application logs: `wrangler tail`

---

**Last Updated**: 2025-01-07
**Maintainer**: DevOps Team
**Version**: 1.0
