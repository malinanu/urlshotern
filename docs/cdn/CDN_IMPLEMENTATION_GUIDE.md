# CDN Implementation Guide
## Step-by-Step Implementation for URL Shortener

This guide provides detailed implementation steps for deploying the CDN and edge optimization infrastructure.

---

## Prerequisites

### Required Accounts
- [ ] CloudFlare account (Business tier recommended)
- [ ] Domain registered and ready to transfer
- [ ] Git repository access
- [ ] Production server access

### Required Tools
```bash
# Install required tools
npm install -g @cloudflare/wrangler  # CloudFlare Workers CLI
npm install -g lighthouse            # Performance testing
curl -O https://dl.google.com/go/go1.21.0.linux-amd64.tar.gz  # Go (if not installed)
```

### Environment Variables
Create `.env.cloudflare` file:
```bash
CLOUDFLARE_API_TOKEN=your_api_token_here
CLOUDFLARE_ZONE_ID=your_zone_id_here
CLOUDFLARE_ACCOUNT_ID=your_account_id_here
CLOUDFLARE_KV_NAMESPACE_ID=your_kv_namespace_id
ORIGIN_IP=your_server_ip
BASE_URL=https://trunc.site
```

---

## Phase 1: CloudFlare Account Setup (Day 1)

### Step 1.1: Create CloudFlare Account
1. Visit https://dash.cloudflare.com/sign-up
2. Choose Business plan ($200/month) for production features
3. Verify email and complete account setup

### Step 1.2: Add Domain to CloudFlare
1. Click "Add a Site" in CloudFlare dashboard
2. Enter domain: `trunc.site`
3. Select Business plan
4. CloudFlare will scan existing DNS records

### Step 1.3: Update Nameservers
1. CloudFlare provides nameservers (e.g., `anna.ns.cloudflare.com`)
2. Update nameservers at your domain registrar:
   ```
   Old: your-registrar-ns1.com
   New: anna.ns.cloudflare.com
        bob.ns.cloudflare.com
   ```
3. Wait 24-48 hours for propagation
4. CloudFlare will email when active

### Step 1.4: Generate API Token
1. Go to Profile → API Tokens
2. Click "Create Token"
3. Use "Edit zone DNS" template
4. Permissions:
   - Zone → DNS → Edit
   - Zone → Cache Purge → Purge
   - Zone → Worker Routes → Edit
   - Account → Workers KV Storage → Edit
5. Copy token to `.env.cloudflare`

---

## Phase 2: SSL/TLS Configuration (Day 2)

### Step 2.1: Enable Full (Strict) SSL
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/ssl" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"full_strict"}'
```

### Step 2.2: Configure Origin Certificate
1. Go to SSL/TLS → Origin Server
2. Click "Create Certificate"
3. Select:
   - Private key type: RSA (2048)
   - Hostnames: `trunc.site`, `*.trunc.site`
   - Certificate validity: 15 years
4. Download certificate and key
5. Install on origin server:
   ```bash
   sudo cp origin-cert.pem /etc/letsencrypt/live/trunc.site/fullchain.pem
   sudo cp origin-key.pem /etc/letsencrypt/live/trunc.site/privkey.pem
   sudo chmod 600 /etc/letsencrypt/live/trunc.site/privkey.pem
   ```

### Step 2.3: Enable Security Features
```bash
# Always Use HTTPS
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/always_use_https" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'

# Automatic HTTPS Rewrites
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/automatic_https_rewrites" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'

# TLS 1.3
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/tls_1_3" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

---

## Phase 3: DNS Configuration (Day 2)

### Step 3.1: Configure DNS Records
```bash
# Apex domain
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "type": "A",
    "name": "trunc.site",
    "content": "'${ORIGIN_IP}'",
    "ttl": 1,
    "proxied": true
  }'

# WWW subdomain
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "type": "CNAME",
    "name": "www",
    "content": "trunc.site",
    "ttl": 1,
    "proxied": true
  }'
```

### Step 3.2: Verify DNS
```bash
# Check DNS propagation
dig trunc.site +short
dig www.trunc.site +short

# Should return CloudFlare IPs (e.g., 104.21.x.x)
```

---

## Phase 4: Performance Optimization (Day 3)

### Step 4.1: Enable Auto Minify
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/minify" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":{"css":"on","html":"on","js":"on"}}'
```

### Step 4.2: Enable Brotli Compression
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/brotli" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

### Step 4.3: Enable HTTP/3
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/http3" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

### Step 4.4: Enable Early Hints
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/early_hints" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

### Step 4.5: Configure Polish (Image Optimization)
1. Go to Speed → Optimization
2. Enable Polish: Lossless
3. Enable WebP
4. Enable AVIF

---

## Phase 5: Caching Rules (Day 3-4)

### Step 5.1: Create Page Rules

#### Static Assets (Priority 1)
```bash
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/pagerules" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "targets": [
      {
        "target": "url",
        "constraint": {
          "operator": "matches",
          "value": "*trunc.site/_next/static/*"
        }
      }
    ],
    "actions": [
      {
        "id": "cache_level",
        "value": "cache_everything"
      },
      {
        "id": "edge_cache_ttl",
        "value": 31536000
      },
      {
        "id": "browser_cache_ttl",
        "value": 31536000
      }
    ],
    "priority": 1,
    "status": "active"
  }'
```

#### Short URL Redirects (Priority 2)
```bash
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/pagerules" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "targets": [
      {
        "target": "url",
        "constraint": {
          "operator": "matches",
          "value": "*trunc.site/[a-zA-Z0-9]*"
        }
      }
    ],
    "actions": [
      {
        "id": "cache_level",
        "value": "cache_everything"
      },
      {
        "id": "edge_cache_ttl",
        "value": 3600
      }
    ],
    "priority": 2,
    "status": "active"
  }'
```

#### API Bypass (Priority 3)
```bash
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/pagerules" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "targets": [
      {
        "target": "url",
        "constraint": {
          "operator": "matches",
          "value": "*trunc.site/api/v1/shorten*"
        }
      }
    ],
    "actions": [
      {
        "id": "cache_level",
        "value": "bypass"
      }
    ],
    "priority": 3,
    "status": "active"
  }'
```

### Step 5.2: Configure Cache Headers in Nginx
Update your Nginx configuration with the new `nginx-cdn.conf`:
```bash
cd /path/to/urlshortener
sudo cp docker/nginx-cdn.conf /etc/nginx/nginx.conf
sudo nginx -t  # Test configuration
sudo systemctl reload nginx
```

---

## Phase 6: CloudFlare Workers Deployment (Day 4-5)

### Step 6.1: Install Wrangler CLI
```bash
npm install -g @cloudflare/wrangler
wrangler login
```

### Step 6.2: Create KV Namespace
```bash
cd cloudflare/workers
wrangler kv:namespace create "URL_MAPPINGS"

# Output will show:
# Add the following to wrangler.toml:
# { binding = "URL_MAPPINGS", id = "xxxxxxxxxxxxx" }

# Update wrangler.toml with the generated ID
```

### Step 6.3: Configure Wrangler
Update `cloudflare/workers/wrangler.toml` with your account details:
```toml
account_id = "your_account_id"
zone_id = "your_zone_id"

[[kv_namespaces]]
binding = "URL_MAPPINGS"
id = "your_kv_namespace_id"
```

### Step 6.4: Deploy Redirect Analytics Worker
```bash
cd cloudflare/workers
wrangler publish redirect-analytics.js --env production

# Verify deployment
curl https://trunc.site/health/edge
```

### Step 6.5: Deploy A/B Testing Worker (Optional)
```bash
wrangler publish ab-testing.js --env production
```

### Step 6.6: Configure Worker Routes
```bash
# Route for redirect analytics
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/workers/routes" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "pattern": "trunc.site/*",
    "script": "url-shortener-redirect-prod"
  }'
```

---

## Phase 7: Backend Integration (Day 5-6)

### Step 7.1: Add CDN Invalidation to URL Service

Update `internal/services/shortener.go`:

```go
import (
    "github.com/URLshorter/url-shortener/internal/services"
)

// Add to ShortenerService struct
type ShortenerService struct {
    // ... existing fields
    cdnInvalidator *services.CDNInvalidator
}

// Initialize in NewShortenerService
func NewShortenerService(/* ... existing params */, cdnInvalidator *services.CDNInvalidator) *ShortenerService {
    return &ShortenerService{
        // ... existing fields
        cdnInvalidator: cdnInvalidator,
    }
}

// Add invalidation when URL is updated
func (s *ShortenerService) UpdateUserURL(userID int64, shortCode string, req *models.UpdateURLRequest) (*models.URLMapping, error) {
    // ... existing update logic

    // Invalidate CDN cache
    if s.cdnInvalidator != nil {
        ctx := context.Background()
        if err := s.cdnInvalidator.InvalidateShortURL(ctx, shortCode, s.config.BaseURL); err != nil {
            // Log error but don't fail the update
            log.Printf("Failed to invalidate CDN cache: %v", err)
        }
    }

    return updatedURL, nil
}
```

### Step 7.2: Update Main Server Initialization

Update `cmd/server/main.go`:

```go
// Add CDN invalidator initialization
cdnInvalidator := services.NewCDNInvalidator(
    services.CDNProviderCloudFlare,
    cfg.CloudFlareAPIToken,
    cfg.CloudFlareZoneID,
)

// Pass to service
shortenerService := services.NewShortenerService(
    storage,
    redis,
    cfg,
    cdnInvalidator,
)
```

### Step 7.3: Add Configuration
Update `configs/config.go`:

```go
type Config struct {
    // ... existing fields

    // CloudFlare Configuration
    CloudFlareAPIToken string
    CloudFlareZoneID   string
    CloudFlareAccountID string
}

func LoadConfig() (*Config, error) {
    config := &Config{
        // ... existing fields
        CloudFlareAPIToken:  getEnv("CLOUDFLARE_API_TOKEN", ""),
        CloudFlareZoneID:    getEnv("CLOUDFLARE_ZONE_ID", ""),
        CloudFlareAccountID: getEnv("CLOUDFLARE_ACCOUNT_ID", ""),
    }
    return config, nil
}
```

### Step 7.4: Update Environment Variables
Add to `.env`:
```bash
CLOUDFLARE_API_TOKEN=your_api_token
CLOUDFLARE_ZONE_ID=your_zone_id
CLOUDFLARE_ACCOUNT_ID=your_account_id
```

### Step 7.5: Rebuild and Deploy
```bash
# Build new version
docker-compose build app

# Deploy with zero downtime
docker-compose up -d app

# Verify
curl https://trunc.site/health
```

---

## Phase 8: Monitoring & Verification (Day 6-7)

### Step 8.1: Test CDN Caching
```bash
# Test static assets
curl -I https://trunc.site/_next/static/css/main.css
# Should see: cf-cache-status: HIT

# Test short URL redirect
curl -I https://trunc.site/abc123
# Should see: cf-cache-status: HIT (after first request)

# Test API endpoint
curl -I https://trunc.site/api/v1/health
# Should see: cf-cache-status: DYNAMIC or BYPASS
```

### Step 8.2: Performance Testing
```bash
# Run Lighthouse audit
lighthouse https://trunc.site --view

# Expected scores:
# Performance: 90+
# SEO: 95+
# Best Practices: 95+
```

### Step 8.3: Load Testing
```bash
# Install k6 (load testing tool)
brew install k6  # macOS
# or
wget https://github.com/grafana/k6/releases/download/v0.47.0/k6-v0.47.0-linux-amd64.tar.gz

# Run load test
k6 run - <<EOF
import http from 'k6/http';
import { check } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 100 },
    { duration: '1m', target: 1000 },
    { duration: '30s', target: 0 },
  ],
};

export default function () {
  let res = http.get('https://trunc.site/abc123');
  check(res, {
    'status is 301': (r) => r.status === 301,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });
}
EOF
```

### Step 8.4: Monitor CloudFlare Analytics
1. Go to CloudFlare dashboard
2. Select your zone (trunc.site)
3. Navigate to Analytics & Logs
4. Check:
   - Cache hit ratio (target: >80%)
   - Total requests
   - Bandwidth saved
   - Threats stopped

---

## Phase 9: Security Configuration (Day 7)

### Step 9.1: Enable WAF
```bash
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/waf" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

### Step 9.2: Configure Rate Limiting
```bash
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/rate_limits" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{
    "match": {
      "request": {
        "url": "*/api/v1/shorten"
      }
    },
    "threshold": 100,
    "period": 600,
    "action": {
      "mode": "challenge",
      "timeout": 86400
    },
    "description": "API rate limit"
  }'
```

### Step 9.3: Configure Bot Management
1. Go to Security → Bots
2. Configure Bot Fight Mode
3. Allow verified bots (Google, Bing, etc.)
4. Challenge suspicious bots

---

## Phase 10: Optimization & Fine-Tuning (Ongoing)

### Step 10.1: Cache Warming
Run cache warming script after deployment:
```bash
# Create list of popular URLs
cat > popular_urls.txt <<EOF
https://trunc.site/
https://trunc.site/pricing
https://trunc.site/features
https://trunc.site/abc123
https://trunc.site/xyz789
EOF

# Warm cache
while read url; do
  curl -I "$url"
  sleep 0.1
done < popular_urls.txt
```

### Step 10.2: Monitor and Adjust TTLs
Based on analytics, adjust cache TTLs:
- High traffic, rarely updated: Increase TTL
- Frequently updated: Decrease TTL
- Real-time data: Bypass cache

### Step 10.3: A/B Testing
Enable A/B testing worker and monitor conversion rates:
```bash
# Enable A/B test in worker configuration
# Update cloudflare/workers/ab-testing.js
```

---

## Rollback Procedure

If issues occur, rollback steps:

### Step 1: Disable Workers
```bash
# Delete worker routes
curl -X DELETE "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/workers/routes/${ROUTE_ID}" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}"
```

### Step 2: Bypass CDN
```bash
# Set DNS to non-proxied (direct to origin)
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/dns_records/${RECORD_ID}" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"proxied":false}'
```

### Step 3: Revert Nginx Configuration
```bash
sudo cp /etc/nginx/nginx.conf.backup /etc/nginx/nginx.conf
sudo systemctl reload nginx
```

---

## Cost Tracking

### Month 1 Costs
- CloudFlare Business: $200
- Development time: 40 hours
- Testing: 8 hours
- Total: $200 + labor

### Monthly Ongoing Costs
- CloudFlare Business: $200/month
- Workers (Bundled plan): Included
- KV storage: ~$0.50/month (first 1GB free)
- Bandwidth: Unlimited on Business plan

### Expected Savings
- Bandwidth: $400/month saved
- Compute: $180/month saved
- **Net savings**: $380/month ($4,560/year)

---

## Success Metrics

Track these KPIs after implementation:

| Metric | Before CDN | After CDN | Target |
|--------|-----------|-----------|--------|
| TTFB (US) | 450ms | 50ms | <100ms |
| TTFB (EU) | 850ms | 80ms | <150ms |
| TTFB (APAC) | 1200ms | 120ms | <200ms |
| Cache Hit Ratio | 0% | 85% | >80% |
| Origin Load | 100% | 15% | <20% |
| Page Load Time | 2.8s | 0.9s | <1.5s |
| Lighthouse Score | 72 | 95 | >90 |

---

## Troubleshooting

### Issue: Cache Not Working
**Symptoms**: cf-cache-status always shows MISS or BYPASS

**Solution**:
1. Check page rules priority
2. Verify Cache-Control headers from origin
3. Check if CloudFlare development mode is enabled (disable it)
4. Verify CloudFlare is proxying traffic (orange cloud enabled)

### Issue: Workers Not Executing
**Symptoms**: X-Edge headers not present

**Solution**:
1. Verify worker route pattern matches request
2. Check worker logs: `wrangler tail redirect-analytics`
3. Verify KV namespace is bound correctly
4. Check worker quota limits

### Issue: High Origin Load
**Symptoms**: Origin still receiving most traffic

**Solution**:
1. Increase cache TTLs
2. Check if cookies are bypassing cache
3. Verify "Cache Everything" page rules
4. Review query string handling

---

## Next Steps

After successful implementation:

1. **Week 2**: Monitor analytics and adjust TTLs
2. **Week 3**: Implement multi-region deployment
3. **Week 4**: Enable advanced A/B testing
4. **Month 2**: Optimize based on real traffic patterns
5. **Month 3**: Implement edge computing for analytics

---

## Support Resources

- CloudFlare Documentation: https://developers.cloudflare.com/
- Workers Documentation: https://developers.cloudflare.com/workers/
- Community Forum: https://community.cloudflare.com/
- Support: support@cloudflare.com (Business plan includes 24/7 support)
