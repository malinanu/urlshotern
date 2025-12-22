# URL Shortener Performance Guide

## Performance Architecture Overview

The URL Shortener is designed with multiple layers of optimization for high performance and scalability:

### 1. Frontend Performance
- **Next.js Static Generation**: Pre-generated pages for faster loading
- **CDN Distribution**: CloudFront/CloudFlare for global content delivery
- **Image Optimization**: Automatic WebP conversion and compression
- **Code Splitting**: Lazy loading of components and routes
- **Service Worker**: PWA caching for offline functionality

### 2. Backend Performance
- **Go Concurrency**: Goroutines for handling high concurrent loads
- **Connection Pooling**: Database and Redis connection pools
- **Caching Strategy**: Multi-layer caching (Application → Redis → Database)
- **Horizontal Scaling**: Auto-scaling based on CPU and memory metrics

### 3. Caching Strategy

#### Layer 1: CDN/Edge Caching
- **Static Assets**: 1 year cache (immutable)
- **HTML Pages**: 10 minutes cache
- **API Responses**: No cache (dynamic content)
- **Redirects**: 5 minutes cache

#### Layer 2: Nginx Reverse Proxy
- **URL Redirects**: 5 minutes cache
- **Static Content**: 1 hour cache
- **API Endpoints**: No cache
- **Rate Limiting**: 100 req/s per IP for redirects, 10 req/s for API

#### Layer 3: Application Cache (Redis)
- **URL Mappings**: 1 hour TTL
- **User Sessions**: 24 hours TTL
- **Analytics Data**: 15 minutes TTL
- **Rate Limit Counters**: 1 minute TTL

#### Layer 4: Database Optimizations
- **Read Replicas**: For analytics queries
- **Connection Pooling**: 100 max connections
- **Query Optimization**: Proper indexing and query planning
- **Partitioning**: Time-based partitioning for analytics data

## Performance Metrics & SLAs

### Service Level Objectives (SLOs)

| Metric | Target | Measurement Period |
|--------|--------|-------------------|
| URL Redirect Latency | < 50ms (95th percentile) | 7 days |
| API Response Time | < 200ms (95th percentile) | 7 days |
| Uptime | 99.9% | Monthly |
| Error Rate | < 0.1% | Daily |
| Cache Hit Rate | > 80% | Daily |

### Key Performance Indicators (KPIs)

1. **Throughput**
   - Target: 10,000 redirects/second
   - Target: 1,000 API requests/second

2. **Latency Distribution**
   - P50: < 20ms (redirects), < 100ms (API)
   - P90: < 40ms (redirects), < 150ms (API)
   - P95: < 50ms (redirects), < 200ms (API)
   - P99: < 100ms (redirects), < 500ms (API)

3. **Resource Utilization**
   - CPU: < 70% average
   - Memory: < 80% average
   - Network: < 80% bandwidth

## Load Testing

### Continuous Performance Testing

Load tests run automatically every Monday at 4 AM UTC using k6:

```javascript
// Load test configuration
export let options = {
  stages: [
    { duration: '2m', target: 50 },   // Ramp up
    { duration: '5m', target: 100 },  // Steady state
    { duration: '2m', target: 200 },  // Peak load
    { duration: '5m', target: 200 },  // Sustain peak
    { duration: '2m', target: 0 },    // Ramp down
  ],
  thresholds: {
    http_req_duration: ['p(95)<500'],
    http_req_failed: ['rate<0.01'],
  },
};
```

### Manual Performance Testing

Run comprehensive load tests before major releases:

```bash
# Install k6
brew install k6

# Run basic load test
k6 run tests/performance/basic-load-test.js

# Run stress test
k6 run tests/performance/stress-test.js

# Run spike test
k6 run tests/performance/spike-test.js
```

### Performance Test Scenarios

1. **URL Creation Load Test**
   ```bash
   k6 run --vus 100 --duration 5m tests/performance/url-creation.js
   ```

2. **Redirect Performance Test**
   ```bash
   k6 run --vus 500 --duration 10m tests/performance/url-redirect.js
   ```

3. **Mixed Workload Test**
   ```bash
   k6 run tests/performance/mixed-workload.js
   ```

## Monitoring & Alerting

### Performance Dashboards

1. **Application Performance**
   - Response time trends
   - Error rate monitoring
   - Throughput metrics
   - Cache hit rates

2. **Infrastructure Performance**
   - CPU and memory utilization
   - Network I/O
   - Disk I/O
   - Database performance

3. **Business Metrics**
   - URLs created per minute
   - Unique visitors
   - Geographic distribution
   - Popular domains

### Performance Alerts

```yaml
# High latency alert
- alert: HighLatency
  expr: histogram_quantile(0.95, rate(http_request_duration_seconds_bucket[5m])) > 0.5
  for: 5m
  labels:
    severity: warning
  annotations:
    summary: "High response time detected"

# Low cache hit rate alert
- alert: LowCacheHitRate
  expr: rate(cache_hits_total[5m]) / rate(cache_requests_total[5m]) < 0.8
  for: 10m
  labels:
    severity: warning
  annotations:
    summary: "Cache hit rate below threshold"

# High error rate alert
- alert: HighErrorRate
  expr: rate(http_requests_total{status=~"5.."}[5m]) / rate(http_requests_total[5m]) > 0.01
  for: 2m
  labels:
    severity: critical
  annotations:
    summary: "High error rate detected"
```

## Performance Optimization Techniques

### 1. Database Optimizations

#### Indexing Strategy
```sql
-- Primary indexes for fast lookups
CREATE INDEX CONCURRENTLY idx_url_mappings_short_code ON url_mappings(short_code);
CREATE INDEX CONCURRENTLY idx_url_mappings_user_id ON url_mappings(user_id);
CREATE INDEX CONCURRENTLY idx_click_events_short_code ON click_events(short_code);

-- Partial indexes for active URLs only
CREATE INDEX CONCURRENTLY idx_url_mappings_active 
ON url_mappings(short_code) WHERE is_active = true;

-- Composite indexes for analytics queries
CREATE INDEX CONCURRENTLY idx_click_events_analytics 
ON click_events(short_code, created_at) WHERE created_at >= NOW() - INTERVAL '30 days';
```

#### Query Optimization
```go
// Use prepared statements
stmt, err := db.Prepare("SELECT original_url FROM url_mappings WHERE short_code = $1 AND is_active = true")

// Use connection pooling
db.SetMaxOpenConns(100)
db.SetMaxIdleConns(10)
db.SetConnMaxLifetime(time.Hour)

// Implement query timeout
ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
defer cancel()
```

### 2. Redis Optimizations

#### Configuration
```redis
# Memory optimization
maxmemory 2gb
maxmemory-policy allkeys-lru

# Persistence optimization
save 900 1
save 300 10
save 60 10000

# Network optimization
tcp-keepalive 300
```

#### Caching Patterns
```go
// Cache-aside pattern
func GetURL(shortCode string) (*URL, error) {
    // Try cache first
    if cached := redis.Get("url:" + shortCode); cached != nil {
        return parseURL(cached), nil
    }
    
    // Fallback to database
    url, err := db.GetURL(shortCode)
    if err != nil {
        return nil, err
    }
    
    // Cache the result
    redis.Set("url:"+shortCode, serializeURL(url), time.Hour)
    return url, nil
}
```

### 3. Application Optimizations

#### Go Performance Tips
```go
// Use sync.Pool for object reuse
var requestPool = sync.Pool{
    New: func() interface{} {
        return &Request{}
    },
}

// Preallocate slices
urls := make([]URL, 0, expectedSize)

// Use buffered channels
results := make(chan Result, 100)

// Implement circuit breaker
if circuitBreaker.IsOpen() {
    return nil, errors.New("service unavailable")
}
```

### 4. CDN Configuration

#### CloudFlare Settings
```yaml
# Cache everything
Cache Level: Cache Everything

# Browser cache TTL
Browser Cache TTL: 1 year (static assets), 4 hours (HTML)

# Edge cache TTL
Edge Cache TTL: 1 month (static), 10 minutes (dynamic)

# Minification
Auto Minify: CSS, JavaScript, HTML

# Compression
Brotli: Enabled
Gzip: Enabled
```

## Capacity Planning

### Traffic Projections

| Timeframe | URLs Created/Day | Redirects/Day | Peak RPS | Storage |
|-----------|------------------|---------------|----------|---------|
| Month 1   | 10K             | 100K         | 50       | 10GB    |
| Month 6   | 50K             | 1M           | 200      | 50GB    |
| Year 1    | 200K            | 5M           | 1000     | 200GB   |
| Year 2    | 500K            | 20M          | 4000     | 1TB     |

### Scaling Recommendations

#### Horizontal Scaling Triggers
- CPU > 70% for 10 minutes → Scale out
- Memory > 80% for 10 minutes → Scale out
- Response time > SLA for 5 minutes → Scale out

#### Vertical Scaling Considerations
- Database: Scale up when connection pool is consistently maxed
- Redis: Scale up when memory usage > 80%
- Application: Scale up individual pods before scaling out

### Infrastructure Scaling

```yaml
# HPA Configuration
apiVersion: autoscaling/v2
kind: HorizontalPodAutoscaler
metadata:
  name: urlshortener-hpa
spec:
  minReplicas: 3
  maxReplicas: 50
  metrics:
  - type: Resource
    resource:
      name: cpu
      target:
        type: Utilization
        averageUtilization: 70
  - type: Resource
    resource:
      name: memory
      target:
        type: Utilization
        averageUtilization: 80
  - type: Pods
    pods:
      metric:
        name: http_requests_per_second
      target:
        type: AverageValue
        averageValue: "100"
```

## Performance Best Practices

### Development Guidelines

1. **Always profile before optimizing**
   ```bash
   go tool pprof http://localhost:8080/debug/pprof/profile
   ```

2. **Measure everything**
   ```go
   timer := prometheus.NewTimer(httpDuration)
   defer timer.ObserveDuration()
   ```

3. **Implement graceful degradation**
   ```go
   if redisDown {
       // Fallback to database only
       return getFromDatabase(key)
   }
   ```

4. **Use context for timeouts**
   ```go
   ctx, cancel := context.WithTimeout(ctx, 5*time.Second)
   defer cancel()
   ```

### Deployment Best Practices

1. **Blue-green deployments** for zero-downtime updates
2. **Canary releases** for gradual rollouts
3. **Performance testing** in staging environment
4. **Monitoring** deployment impact on performance metrics

### Operational Best Practices

1. **Regular performance reviews** (monthly)
2. **Capacity planning** based on growth projections
3. **Performance budget** for new features
4. **Incident response** for performance degradation

This performance guide ensures the URL Shortener maintains optimal performance while scaling to handle millions of requests per day.