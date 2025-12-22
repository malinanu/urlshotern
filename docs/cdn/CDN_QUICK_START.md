# CDN Quick Start Guide
## Get Your URL Shortener on CloudFlare CDN in 1 Hour

This quick start guide will get you from zero to a fully CDN-enabled URL shortener in under 1 hour.

---

## Prerequisites (5 minutes)

- [ ] CloudFlare account (sign up at https://dash.cloudflare.com/sign-up)
- [ ] Domain name (trunc.site)
- [ ] Credit card for CloudFlare Business plan ($200/month)
- [ ] Terminal access to your server
- [ ] Git repository cloned locally

---

## Step 1: CloudFlare Account Setup (10 minutes)

### 1.1 Sign Up and Add Domain
```bash
# Visit CloudFlare dashboard
open https://dash.cloudflare.com/sign-up

# Add your domain: trunc.site
# Select "Business" plan
# Complete payment
```

### 1.2 Update Nameservers
```bash
# CloudFlare will show nameservers like:
# anna.ns.cloudflare.com
# bob.ns.cloudflare.com

# Update at your domain registrar (GoDaddy, Namecheap, etc.)
# This can take 24-48 hours, but we can continue setup
```

### 1.3 Get API Token
```bash
# Go to: Profile → API Tokens → Create Token
# Use template: "Edit zone DNS"
# Copy token to safe location
```

---

## Step 2: Configure Environment (5 minutes)

### 2.1 Create CloudFlare Environment File
```bash
cd /path/to/URLshorter

cat > .env.cloudflare <<EOF
CLOUDFLARE_API_TOKEN=your_api_token_here
CLOUDFLARE_ZONE_ID=your_zone_id_here
CLOUDFLARE_ACCOUNT_ID=your_account_id_here
BASE_URL=https://trunc.site
ORIGIN_IP=your_server_ip
EOF

# Add to .gitignore
echo ".env.cloudflare" >> .gitignore
```

### 2.2 Update Main .env File
```bash
# Add these lines to your existing .env file
cat >> .env <<EOF

# CloudFlare Configuration
CLOUDFLARE_API_TOKEN=${CLOUDFLARE_API_TOKEN}
CLOUDFLARE_ZONE_ID=${CLOUDFLARE_ZONE_ID}
CLOUDFLARE_ACCOUNT_ID=${CLOUDFLARE_ACCOUNT_ID}
EOF
```

---

## Step 3: Deploy CloudFlare Workers (15 minutes)

### 3.1 Install Wrangler CLI
```bash
npm install -g @cloudflare/wrangler
wrangler login
```

### 3.2 Create KV Namespace
```bash
cd cloudflare/workers
wrangler kv:namespace create "URL_MAPPINGS"

# Copy the output ID and update wrangler.toml:
# [[kv_namespaces]]
# binding = "URL_MAPPINGS"
# id = "paste_id_here"
```

### 3.3 Update Worker Configuration
```bash
# Edit cloudflare/workers/wrangler.toml
# Update these fields:
# - account_id
# - zone_id
# - KV namespace id (from previous step)
```

### 3.4 Deploy Workers
```bash
# Deploy redirect analytics worker
wrangler publish redirect-analytics.js --env production

# Verify deployment
curl https://trunc.site/health/edge
```

---

## Step 4: Configure CDN Settings (10 minutes)

### 4.1 Run Automated Setup Script
```bash
cd /path/to/URLshorter

# Make script executable
chmod +x scripts/deploy-cdn.sh

# Run deployment script
./scripts/deploy-cdn.sh
```

**OR** Manual configuration:

### 4.2 Manual Configuration (if script fails)

#### Enable SSL/TLS
```bash
source .env.cloudflare

# Full Strict SSL
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/ssl" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"full_strict"}'

# Always Use HTTPS
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/always_use_https" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'
```

#### Enable Performance Features
```bash
# Brotli compression
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/brotli" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'

# HTTP/3
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/http3" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":"on"}'

# Auto Minify
curl -X PATCH "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/minify" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"value":{"css":"on","html":"on","js":"on"}}'
```

---

## Step 5: Update Backend Integration (10 minutes)

### 5.1 Update Go Configuration
```bash
# The CDN invalidation code is already in:
# internal/services/cdn_cache_invalidation.go

# Update configs/config.go (already done in repository)
# Update cmd/server/main.go to initialize CDN invalidator
```

### 5.2 Rebuild and Deploy
```bash
# Build new Docker image
docker-compose build app

# Deploy (zero downtime)
docker-compose up -d app

# Verify
curl https://trunc.site/health
```

---

## Step 6: Update Nginx Configuration (5 minutes)

### 6.1 Backup Current Configuration
```bash
sudo cp /etc/nginx/nginx.conf /etc/nginx/nginx.conf.backup
```

### 6.2 Install CDN-Optimized Configuration
```bash
sudo cp docker/nginx-cdn.conf /etc/nginx/nginx.conf

# Test configuration
sudo nginx -t

# Reload nginx
sudo systemctl reload nginx
```

---

## Step 7: Verify Deployment (10 minutes)

### 7.1 Test Cache Headers
```bash
# Test static assets (should cache for 1 year)
curl -I https://trunc.site/_next/static/css/main.css
# Look for: cf-cache-status: HIT
# Look for: cache-control: public, max-age=31536000, immutable

# Test short URL redirect (should cache for 1 hour)
curl -I https://trunc.site/abc123
# Look for: cf-cache-status: HIT (after 2nd request)
# Look for: cache-control: public, max-age=3600

# Test API endpoint (should bypass cache)
curl -I https://trunc.site/api/v1/health
# Look for: cf-cache-status: DYNAMIC or BYPASS
```

### 7.2 Check CloudFlare Dashboard
1. Go to https://dash.cloudflare.com
2. Select your domain
3. Navigate to Analytics & Logs
4. Verify:
   - Requests are being proxied
   - Cache hit ratio is increasing
   - SSL is active

### 7.3 Performance Test
```bash
# Install lighthouse
npm install -g lighthouse

# Run performance audit
lighthouse https://trunc.site --view

# Expected scores:
# Performance: 90+
# Best Practices: 95+
```

---

## Step 8: Monitor and Optimize (Ongoing)

### 8.1 Set Up Monitoring Dashboard
```bash
# Access CloudFlare analytics
open https://dash.cloudflare.com

# Key metrics to watch:
# - Cache hit ratio (target: >80%)
# - Bandwidth saved
# - Threats blocked
# - Request rate
```

### 8.2 Check Origin Server Load
```bash
# SSH into your server
ssh your-server

# Check CPU usage (should drop by 60%)
top

# Check bandwidth usage (should drop by 90%)
vnstat -m

# Check application logs
docker-compose logs -f app
```

---

## Troubleshooting

### Issue: Workers Not Working
```bash
# Check worker logs
wrangler tail redirect-analytics --env production

# Verify worker route
curl -I https://trunc.site/test
# Should see X-Edge header if worker is running
```

### Issue: Cache Not Working
```bash
# Verify DNS is proxied (orange cloud)
dig trunc.site +short
# Should return CloudFlare IPs

# Check page rules
curl -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/pagerules" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}"

# Purge cache and retry
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"purge_everything":true}'
```

### Issue: SSL Errors
```bash
# Verify SSL mode is Full (Strict)
curl -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/settings/ssl" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}"

# Install origin certificate on server
# Download from: CloudFlare Dashboard → SSL/TLS → Origin Server
```

---

## Expected Results

After completing this quick start, you should see:

### Performance Improvements
- TTFB reduced from 450ms → 50ms (90% improvement)
- Page load time: 2.8s → 0.9s (68% faster)
- Lighthouse score: 72 → 95

### Cost Savings
- Monthly cost reduced from $830 → $425 (49% savings)
- Bandwidth cost: $500 → $50 (90% reduction)
- Can handle 10x traffic with same infrastructure

### Security Enhancements
- DDoS protection enabled automatically
- WAF filtering malicious requests
- SSL/TLS with perfect forward secrecy
- Bot management and rate limiting

---

## Next Steps

Now that your CDN is deployed:

1. **Week 1**: Monitor analytics and fine-tune cache rules
2. **Week 2**: Implement A/B testing with edge workers
3. **Week 3**: Deploy multi-region servers for redundancy
4. **Week 4**: Optimize based on real traffic patterns

---

## Useful Commands

### Purge Cache
```bash
# Purge everything (use sparingly!)
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"purge_everything":true}'

# Purge specific URL
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"files":["https://trunc.site/abc123"]}'

# Purge by tag
curl -X POST "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/purge_cache" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" \
  -H "Content-Type: application/json" \
  --data '{"tags":["url:abc123"]}'
```

### View Worker Logs
```bash
# Real-time logs
wrangler tail redirect-analytics --env production

# Filter by status code
wrangler tail redirect-analytics --env production --status error
```

### Check Cache Hit Ratio
```bash
# Via API
curl -X GET "https://api.cloudflare.com/client/v4/zones/${CLOUDFLARE_ZONE_ID}/analytics/dashboard?since=-1440" \
  -H "Authorization: Bearer ${CLOUDFLARE_API_TOKEN}" | jq '.result.timeseries[].requests.cached_rate'
```

---

## Support

- CloudFlare Documentation: https://developers.cloudflare.com/
- CloudFlare Community: https://community.cloudflare.com/
- Business Support: Available 24/7 via dashboard
- Emergency: support@cloudflare.com

---

## Checklist

Use this checklist to track your progress:

- [ ] CloudFlare account created
- [ ] Domain added to CloudFlare
- [ ] Nameservers updated (can take 24-48h)
- [ ] API token generated
- [ ] Environment variables configured
- [ ] Wrangler CLI installed
- [ ] KV namespace created
- [ ] Workers deployed
- [ ] CDN settings configured
- [ ] Backend integration updated
- [ ] Nginx configuration updated
- [ ] Cache headers verified
- [ ] Performance tested
- [ ] Monitoring dashboard configured
- [ ] Team trained on new setup

---

**Time to Complete**: 1 hour (excluding DNS propagation)
**Difficulty**: Intermediate
**Cost**: $200/month (CloudFlare Business)
**Expected Savings**: $405/month
**ROI**: Immediate

**You're done! Your URL shortener is now globally distributed with sub-50ms latency worldwide.**
