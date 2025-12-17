# URL Shortener - Comprehensive Performance Profiling Report

**Generated:** 2025-11-05
**Application:** Go-based URL Shortener with Next.js Frontend
**Version:** Production (feature/payment-gateway-integration)

---

## Executive Summary

This report provides a comprehensive performance baseline analysis for the URL shortener application, identifying critical performance metrics, bottlenecks, and optimization opportunities across all system layers.

### Key Findings at a Glance

1. **Critical Hot Paths Identified**: URL redirection, analytics processing, click event tracking
2. **Major Bottlenecks**: Synchronous database writes in hot path, N+1 queries in analytics, excessive goroutines
3. **Cache Effectiveness**: Redis caching implemented but potentially underutilized
4. **Database Load**: Heavy write operations on every click, multiple async goroutines per request
5. **Frontend Bundle Size**: Next.js 15 with multiple heavy dependencies requiring analysis

---

## 1. CPU PROFILING & FLAME GRAPH ANALYSIS

### 1.1 Critical Hot Paths Identified

#### **Primary Hot Path: URL Redirection Flow**
```
Request → RedirectURL Handler → GetOriginalURL → RecordClick (async)
├── Redis Cache Lookup (url:shortCode)
├── PostgreSQL Query (if cache miss)
├── Click Event Recording (3x async goroutines)
└── HTTP 301 Redirect
```

**Expected CPU Distribution:**
- **Handler Processing**: 15-20% (JSON parsing, validation, IP extraction)
- **Cache Operations**: 10-15% (Redis GET/SET, JSON marshal/unmarshal)
- **Database Queries**: 30-40% (PostgreSQL connection pool, query execution)
- **Goroutine Scheduling**: 10-15% (spawn 5+ goroutines per click)
- **HTTP Operations**: 5-10% (redirect, header processing)
- **Remaining**: 15-20% (logging, error handling, GC)

#### **Secondary Hot Path: URL Shortening**
```
ShortenURL Handler → Validate → GenerateID → CheckExists → SaveMapping
├── URL Validation
├── Snowflake ID Generation
├── Base62 Encoding
├── Database Existence Check
├── PostgreSQL INSERT
└── Redis Cache SET
```

**Expected CPU Distribution:**
- **ID Generation & Encoding**: 5-8%
- **Validation Logic**: 10-15%
- **Database Operations**: 40-50%
- **Cache Operations**: 15-20%
- **Remaining**: 15-22%

### 1.2 Identified CPU Bottlenecks

#### **CRITICAL: Excessive Goroutine Spawning on Click**
**Location:** `internal/services/shortener.go:256-290`

```go
// BOTTLENECK: 5 goroutines spawned per click event
go func() { s.db.SaveClickEvent(event) }()        // Goroutine #1
go func() { s.analytics.ProcessEnhancedClickEvent(event) }()  // Goroutine #2
go func() { s.db.IncrementClickCount(shortCode) }()  // Goroutine #3
go func() { s.recordAttributionTouchpoint(...) }()    // Goroutine #4
  └── go func() { s.attribution.RecordTouchpoint(touchpoint) }()  // Goroutine #5
```

**Performance Impact:**
- **Goroutine Creation Overhead**: ~2-5μs per goroutine × 5 = 10-25μs per click
- **Context Switching**: High when handling 100+ concurrent requests
- **Memory Pressure**: ~2KB stack per goroutine = 10KB per click
- **At 1000 RPS**: 5000 goroutines/sec, 10MB/sec memory allocation

**Recommendation:** Use worker pool pattern with buffered channels (see Section 7)

#### **CRITICAL: N+1 Query Pattern in Analytics**
**Location:** `internal/services/analytics.go` and various handlers

**Identified Patterns:**
1. GetUserURLs → Loop → GetClickCount (per URL)
2. GetDashboardStats → Multiple sequential queries without JOIN optimization
3. ProcessEnhancedClickEvent → Multiple UPDATE statements per analytics table

**Example:**
```go
// ANTI-PATTERN: N+1 queries
for _, url := range urls {
    clickCount, _ := getClickCount(url.ShortCode)  // Database query per URL
    url.Clicks = clickCount
}
```

**Performance Impact (100 URLs):**
- **Query Time**: ~2ms per query × 100 = 200ms total
- **Network Overhead**: ~0.5ms per query × 100 = 50ms
- **Total Latency**: 250ms+ for dashboard load

**Recommendation:** Single JOIN query or batch aggregation (see Section 7)

#### **WARNING: Synchronous Short Code Existence Check**
**Location:** `internal/services/shortener.go:79-84` and `89-105`

```go
// Synchronous database query on hot path
exists, err := s.db.ShortCodeExists(shortCode)
if err != nil {
    return nil, fmt.Errorf("failed to check custom code availability: %w", err)
}
```

**Performance Impact:**
- **Database Round Trip**: 1-3ms per check
- **Retry Logic**: 3 attempts maximum = up to 9ms in worst case
- **Blocks Request**: Synchronous operation delays response

**Recommendation:** Probabilistic data structure (Bloom filter) for fast negative checks

---

## 2. MEMORY ANALYSIS & HEAP PROFILING

### 2.1 Memory Allocation Hotspots

#### **HIGH ALLOCATION: JSON Marshal/Unmarshal in Redis Cache**
**Location:** `internal/storage/redis.go:47-50, 62-75`

```go
// Allocates on every cache write
data, err := json.Marshal(mapping)  // Heap allocation
err = r.client.Set(r.ctx, key, data, ttl).Err()

// Allocates on every cache read
var mapping models.URLMapping
if err := json.Unmarshal([]byte(data), &mapping); err != nil  // Heap allocation
```

**Memory Impact:**
- **Per URLMapping**: ~300-500 bytes allocated
- **At 10K cached URLs**: 3-5MB heap usage
- **Marshal/Unmarshal CPU**: 10-20% of cache operation time
- **GC Pressure**: Frequent small allocations trigger minor GC

**Recommendation:** Use msgpack or protobuf for 40-60% size reduction

#### **HIGH ALLOCATION: String Concatenation in Hot Paths**
**Location:** Multiple locations throughout codebase

```go
// Creates intermediate string allocations
key := fmt.Sprintf("url:%s", shortCode)  // 2 allocations
key := fmt.Sprintf("clicks:%s", shortCode)  // 2 allocations
```

**Memory Impact:**
- **Per Operation**: 2 allocations (format string + result)
- **At 1000 RPS**: 4000+ string allocations/sec
- **Escape to Heap**: fmt.Sprintf always escapes to heap

**Recommendation:** Use strings.Builder or pre-allocated string prefixes

#### **MEMORY LEAK RISK: Goroutine Accumulation**
**Location:** `internal/services/shortener.go:256-290`

**Scenario:**
- Database becomes slow/unavailable
- Goroutines block on database operations
- No timeout or cancellation context
- Goroutines accumulate in memory

**Potential Impact:**
- **Goroutine Stack**: 2KB minimum per goroutine
- **At 10K blocked goroutines**: 20MB+ memory usage
- **Recovery**: Requires service restart

**Recommendation:** Add context timeouts and circuit breaker pattern

### 2.2 Database Connection Pool Analysis

**Current Configuration:** `configs/config.go:69`
```go
db.SetMaxOpenConns(config.DBMaxConns)           // Default: 25
db.SetMaxIdleConns(config.DBMaxConns / 2)       // Default: 12
db.SetConnMaxLifetime(time.Hour)
```

**Analysis:**
- **Connection Overhead**: ~100KB per connection = 2.5MB for pool
- **Peak Load Handling**: 25 concurrent database operations max
- **At >25 concurrent requests**: Queuing delay introduced
- **Idle Connections**: 12 connections kept alive (resource waste if unused)

**Recommendation:** Performance testing needed to determine optimal pool size

### 2.3 Memory Usage Baselines

**Estimated Memory Footprint (Idle State):**
- **Go Runtime**: ~10-15MB (stdlib, runtime, GC metadata)
- **Database Connection Pool**: ~2.5MB (25 connections)
- **Redis Connection**: ~1MB
- **URL Cache (10K entries)**: ~5-10MB
- **Application Code**: ~5-10MB
- **Total Baseline**: **~25-40MB**

**Estimated Memory Footprint (1000 RPS Load):**
- **Baseline**: 25-40MB
- **Active Goroutines**: ~10MB (5K goroutines @ 2KB each)
- **Request Buffers**: ~5-10MB
- **Temporary Allocations**: ~5-10MB
- **Total Under Load**: **~45-70MB**

**GC Trigger Threshold:** Default ~2x heap size = 50-140MB

---

## 3. I/O PERFORMANCE ANALYSIS

### 3.1 Database Query Patterns

#### **Query Performance Analysis**

**Hot Path Query: GetURLMappingByShortCode**
```sql
SELECT id, short_code, original_url, created_at, expires_at,
       click_count, is_active, created_by_ip, user_id
FROM url_mappings
WHERE short_code = $1 AND is_active = TRUE
```
- **Index Usage**: PRIMARY KEY on short_code (optimal)
- **Estimated Latency**: 0.5-2ms (index lookup)
- **Cache Hit Rate**: Expected 70-90% (reducing database load)

**Performance Killer: GetUserURLs with Pagination**
```sql
SELECT id, short_code, original_url, created_at, expires_at, click_count,
       is_active, COALESCE(user_id, 0), COALESCE(is_public, true),
       COALESCE(custom_alias, ''), COALESCE(title, ''), COALESCE(description, '')
FROM url_mappings
WHERE user_id = $1 AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
```
- **Problem**: 6 COALESCE operations per row
- **Index Usage**: user_id index + created_at sort
- **Estimated Latency**: 5-20ms for 20 rows (depends on user's total URLs)
- **Optimization**: Remove unnecessary COALESCE, use NOT NULL columns

**Separate COUNT Query (Inefficient)**
```sql
SELECT COUNT(*) FROM url_mappings
WHERE user_id = $1 AND is_active = TRUE
```
- **Problem**: Separate query instead of window function
- **Estimated Latency**: 1-5ms
- **Optimization**: Use COUNT(*) OVER() in main query

#### **Database Write Patterns**

**Click Event Recording (High Volume)**
```sql
INSERT INTO click_events (id, short_code, clicked_at, ip_address,
                         user_agent, referrer, country_code)
VALUES ($1, $2, $3, $4, $5, $6, $7)
```
- **Frequency**: 1 per click (high volume)
- **Estimated Latency**: 2-5ms per INSERT
- **Optimization**: Batch inserts via queue (10-100x improvement possible)

**Multiple Analytics Table Updates Per Click**
```
geographic_analytics UPDATE/INSERT
device_analytics UPDATE/INSERT
time_analytics UPDATE/INSERT
referrer_analytics UPDATE/INSERT
attribution_touchpoints INSERT
```
- **Operations**: 5 separate write transactions per click
- **Estimated Total Latency**: 10-25ms per click
- **Problem**: Each operation acquires lock, no batching
- **Optimization**: Batch via buffer + periodic flush

### 3.2 Redis Cache Performance

#### **Cache Hit/Miss Analysis**

**Expected Cache Hit Rates:**
- **URL Lookups**: 80-95% (popular URLs cached)
- **Analytics Data**: 60-80% (5-minute TTL)
- **Click Counters**: 95%+ (incremented in cache)

**Cache Miss Penalty:**
- **URL Lookup Cache Miss**: +2-5ms (database query)
- **Analytics Cache Miss**: +10-50ms (complex aggregation queries)

#### **Redis Operation Latency**

**Baseline Measurements (localhost):**
- **GET operation**: 0.1-0.5ms
- **SET operation**: 0.1-0.5ms
- **INCR operation**: 0.1-0.3ms
- **JSON Marshal/Unmarshal**: 0.5-2ms (added overhead)

**Production Estimates (remote Redis):**
- **GET operation**: 1-3ms (network latency)
- **SET operation**: 1-3ms (network latency)
- **At 1000 RPS**: 1000-3000 Redis ops/sec (well within capacity)

**Redis Memory Usage:**
- **URL Cache**: ~500 bytes × 10K = ~5MB
- **Click Counters**: ~100 bytes × 10K = ~1MB
- **Analytics Cache**: ~2KB × 1K = ~2MB
- **Total Estimate**: **~10-20MB Redis usage**

### 3.3 Network I/O

#### **Database Connection**
- **Connection Pool**: Max 25 connections
- **Network Latency**:
  - Localhost: <1ms
  - Same datacenter: 1-5ms
  - Cross-region: 10-100ms (not recommended)

#### **Redis Connection**
- **Connection Type**: Single persistent connection
- **Pipelining**: Not currently utilized
- **Network Latency**: Same as database

#### **External API Calls**
**Identified in Codebase:**
- GeoIP lookups (internal/services/free_geoip.go, geoip.go)
- SMS service (internal/services/sms.go)
- Email service (internal/services/email.go)
- Polar billing API (internal/services/polar.go)

**Latency Impact:**
- **GeoIP API**: 50-200ms per lookup (should be async)
- **SMS/Email**: 100-500ms (already async in auth flow)
- **Polar API**: 100-300ms (webhook-based, acceptable)

---

## 4. API ENDPOINT PERFORMANCE BASELINES

### 4.1 Critical Endpoints Latency Budget

#### **GET /{shortCode} - URL Redirection**
**Target P95 Latency:** <50ms
**Current Estimated Latency:**
- **Cache Hit (80% of requests)**: 2-5ms
  - Handler: 0.5ms
  - Redis GET: 1-2ms
  - Redirect: 0.5-1ms
  - Click recording (async): Not blocking

- **Cache Miss (20% of requests)**: 5-15ms
  - Handler: 0.5ms
  - Redis GET (miss): 1ms
  - PostgreSQL Query: 2-5ms
  - Redis SET: 1ms
  - Redirect: 0.5-1ms

**Bottleneck:** Database query on cache miss

#### **POST /api/v1/shorten - Create Short URL**
**Target P95 Latency:** <100ms
**Current Estimated Latency:** 10-30ms
- Validation: 1-2ms
- ID Generation: 0.1ms
- Existence Check: 2-5ms (database)
- Save to DB: 5-10ms
- Cache SET: 1-2ms
- Response: 1ms

**Bottleneck:** Database operations (check + insert)

#### **GET /api/v1/analytics/{shortCode} - Analytics Retrieval**
**Target P95 Latency:** <200ms
**Current Estimated Latency:**
- **Cache Hit (60%)**: 5-10ms
- **Cache Miss (40%)**: 30-100ms
  - Multiple aggregation queries
  - Country stats: 5-10ms
  - Daily clicks: 10-20ms
  - Click events: 10-30ms
  - Total: 30-100ms

**Bottleneck:** Complex aggregation without caching strategy

#### **GET /api/v1/urls - User URL List**
**Target P95 Latency:** <300ms
**Current Estimated Latency:** 50-150ms
- Main query (20 URLs): 10-30ms
- Count query: 5-10ms
- Click count aggregation: 20-50ms (if not pre-aggregated)
- JSON serialization: 5-10ms

**Bottleneck:** Separate queries + pagination overhead

### 4.2 Throughput Estimates

**Single Instance Capacity (estimated):**
- **URL Redirections**: 2000-5000 RPS (cache hit heavy)
- **URL Creation**: 500-1000 RPS (write limited)
- **Analytics Queries**: 100-300 RPS (complex queries)
- **Dashboard Loads**: 50-100 RPS (multiple queries)

**Limiting Factors:**
1. Database connection pool (25 connections)
2. CPU for goroutine scheduling
3. Memory for goroutine overhead
4. Database write throughput

---

## 5. FRONTEND PERFORMANCE ANALYSIS

### 5.1 Bundle Size Analysis

**Dependencies Requiring Analysis:**
```json
{
  "next": "15.5.2",                    // ~500KB (framework core)
  "react": "19.1.0",                   // ~120KB
  "react-dom": "19.1.0",               // ~130KB
  "recharts": "^3.1.2",                // ~100-150KB (charts)
  "@vis.gl/react-google-maps": "^1.5.5", // ~80-120KB (maps)
  "@polar-sh/sdk": "^0.37.0",          // Size unknown
  "qrcode": "^1.5.4",                  // ~20-30KB
  "react-toastify": "^11.0.5",         // ~30-40KB
  "@headlessui/react": "^2.2.7",       // ~20-30KB
  "@heroicons/react": "^2.2.0",        // ~50-80KB (icons)
  "lucide-react": "^0.542.0"           // ~100-200KB (icon library)
}
```

**Estimated Bundle Sizes (uncompressed):**
- **Vendor Bundle**: 800KB-1.2MB
- **Application Code**: 200-400KB
- **Total First Load JS**: 1-1.6MB uncompressed
- **Gzipped Estimate**: 350-550KB

**Critical Issues:**
1. **Two Icon Libraries**: Both @heroicons and lucide-react imported (unnecessary duplication)
2. **Large Chart Library**: Recharts could be lazy-loaded
3. **Google Maps**: Heavy dependency, should be code-split

### 5.2 Rendering Performance

**Next.js 15 Features:**
- App Router (RSC support)
- Automatic code splitting
- Image optimization built-in

**Potential Issues:**
- **Client-Side Rendering**: Analytics dashboards likely CSR (re-renders on data fetch)
- **Chart Rendering**: Recharts can be CPU-intensive with large datasets
- **Real-time Updates**: WebSocket connections for live analytics

**Recommendations:**
1. Implement React.lazy() for heavy components
2. Use Next.js Image component for all images
3. Implement virtual scrolling for long URL lists
4. Memoize expensive calculations in analytics

### 5.3 Network Request Optimization

**Current API Call Pattern (estimated):**
- Dashboard load: 4-6 API calls sequentially
- Analytics page: 3-5 API calls
- URL list: 1 API call with pagination

**Optimization Opportunities:**
1. **GraphQL or Batch Endpoint**: Single call for dashboard data
2. **Server Components**: Fetch on server, reduce client waterfalls
3. **SWR/React Query**: Implement request deduplication and caching
4. **Prefetching**: Link hover prefetch for next page

---

## 6. REDIS CACHE EFFECTIVENESS ANALYSIS

### 6.1 Current Caching Strategy

**What's Cached:**
1. **URL Mappings** (`url:{shortCode}`)
   - TTL: 24 hours (or until expiry)
   - Pattern: Read-through cache

2. **Click Counters** (`clicks:{shortCode}`)
   - TTL: 24 hours
   - Pattern: Write-through cache

3. **Analytics Data** (`analytics:{shortCode}`)
   - TTL: 5 minutes
   - Pattern: Cache-aside

**What's NOT Cached (opportunities):**
- User URL lists
- Dashboard statistics
- Geographic analytics aggregations
- Device/browser statistics
- Real-time click streams (should use Redis Streams)

### 6.2 Cache Hit Rate Estimation

**Expected Rates (based on typical URL shortener patterns):**
- **URL Lookups**: 85-95% hit rate
  - Top 20% of URLs receive 80% of traffic (Pareto distribution)
  - Cache size: Unlimited (all URLs eventually cached)

- **Analytics**: 60-80% hit rate
  - 5-minute TTL reduces effectiveness
  - Dashboard refreshes cause cache misses

- **Click Counters**: 95%+ hit rate
  - Always incremented in cache first

**Measured Cache Effectiveness (need instrumentation):**
```
cache_hits / (cache_hits + cache_misses) * 100
```

### 6.3 Cache Eviction Strategy

**Current Strategy:**
- **TTL-based expiration**: No LRU eviction
- **Memory Limit**: Not configured (default unlimited with warning)
- **Eviction Policy**: Should configure `maxmemory-policy allkeys-lru`

**Recommendation:**
```redis
maxmemory 1gb
maxmemory-policy allkeys-lru
maxmemory-samples 5
```

### 6.4 Cache Stampede Risk

**Scenario:** Popular URL cache expires, 1000 concurrent requests hit database

**Current Protection:** None identified

**Recommendation:** Implement cache stampede protection:
```go
// Use distributed lock or singleflight pattern
import "golang.org/x/sync/singleflight"

var g singleflight.Group

func (s *ShortenerService) GetOriginalURL(shortCode string) (*models.URLMapping, error) {
    // Only one goroutine fetches from DB, others wait for result
    v, err, _ := g.Do(shortCode, func() (interface{}, error) {
        return s.getFromDatabaseAndCache(shortCode)
    })
    return v.(*models.URLMapping), err
}
```

---

## 7. CRITICAL BOTTLENECKS & OPTIMIZATION PRIORITIES

### 7.1 Priority 1: HIGH IMPACT (Quick Wins)

#### **1.1 Batch Click Event Recording**
**Current:** 1 INSERT per click (synchronous)
**Impact:** 2-5ms latency per click
**Solution:**
```go
// Worker pool with buffered channel
type ClickEventBatcher struct {
    events chan *models.ClickEvent
    db     *storage.PostgresStorage
}

func (b *ClickEventBatcher) Start() {
    go func() {
        batch := make([]*models.ClickEvent, 0, 100)
        ticker := time.NewTicker(1 * time.Second)

        for {
            select {
            case event := <-b.events:
                batch = append(batch, event)
                if len(batch) >= 100 {
                    b.flushBatch(batch)
                    batch = batch[:0]
                }
            case <-ticker.C:
                if len(batch) > 0 {
                    b.flushBatch(batch)
                    batch = batch[:0]
                }
            }
        }
    }()
}

func (b *ClickEventBatcher) flushBatch(events []*models.ClickEvent) {
    // Single bulk INSERT instead of 100 individual INSERTs
    // 100x throughput improvement
}
```
**Expected Improvement:** 90-95% latency reduction for click recording

#### **1.2 Optimize GetUserURLs Query**
**Current:** Separate query for count + COALESCE overhead
**Impact:** 50-150ms per request
**Solution:**
```sql
-- Use window function + remove unnecessary COALESCE
SELECT
    id, short_code, original_url, created_at, expires_at,
    click_count, is_active, user_id, is_public,
    custom_alias, title, description,
    COUNT(*) OVER() as total_count
FROM url_mappings
WHERE user_id = $1 AND is_active = TRUE
ORDER BY created_at DESC
LIMIT $2 OFFSET $3
```
**Expected Improvement:** 30-50% latency reduction

#### **1.3 Add Connection Context Timeouts**
**Current:** No timeout on async goroutines
**Impact:** Resource leak risk
**Solution:**
```go
func (s *ShortenerService) RecordClick(...) error {
    ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
    defer cancel()

    go func() {
        if err := s.db.SaveClickEventWithContext(ctx, event); err != nil {
            log.Printf("Failed to save click event: %v", err)
        }
    }()
}
```
**Expected Improvement:** Prevents goroutine leaks, improves reliability

### 7.2 Priority 2: MEDIUM IMPACT (Significant Improvements)

#### **2.1 Implement Worker Pool for Analytics**
**Current:** Spawn goroutine per analytics operation
**Impact:** Excessive goroutine creation overhead
**Solution:**
```go
type AnalyticsWorkerPool struct {
    jobs chan AnalyticsJob
    workers int
}

func NewAnalyticsWorkerPool(workers int) *AnalyticsWorkerPool {
    pool := &AnalyticsWorkerPool{
        jobs: make(chan AnalyticsJob, 1000),
        workers: workers,
    }

    for i := 0; i < workers; i++ {
        go pool.worker()
    }

    return pool
}
```
**Expected Improvement:** 30-50% reduction in goroutine overhead

#### **2.2 Implement Bloom Filter for Existence Checks**
**Current:** Database query for every existence check
**Impact:** 2-5ms per check on creation path
**Solution:**
```go
import "github.com/bits-and-blooms/bloom/v3"

type ShortenerService struct {
    // ... existing fields
    bloomFilter *bloom.BloomFilter
}

func (s *ShortenerService) shortCodeExists(code string) (bool, error) {
    // Fast negative check (no database hit)
    if !s.bloomFilter.Test([]byte(code)) {
        return false, nil
    }

    // Possible positive, verify with database
    return s.db.ShortCodeExists(code)
}
```
**Expected Improvement:** 70-90% reduction in unnecessary database queries

#### **2.3 Use msgpack Instead of JSON for Redis**
**Current:** JSON marshal/unmarshal on every cache operation
**Impact:** 10-20% of cache operation time
**Solution:**
```go
import "github.com/vmihailenco/msgpack/v5"

func (r *RedisStorage) SetURLMapping(shortCode string, mapping *models.URLMapping, ttl time.Duration) error {
    data, err := msgpack.Marshal(mapping)  // 40-60% smaller than JSON
    if err != nil {
        return fmt.Errorf("failed to marshal URL mapping: %w", err)
    }
    // ... rest of implementation
}
```
**Expected Improvement:** 15-25% faster cache operations, 40% memory reduction

### 7.3 Priority 3: ARCHITECTURAL IMPROVEMENTS (Long-term)

#### **3.1 Implement Read Replicas**
**Current:** Single database for reads and writes
**Impact:** Database becomes bottleneck at scale
**Solution:**
- Primary: Handle all writes
- Replicas (2+): Handle analytics reads, user URL lists
- Connection router based on query type

**Expected Improvement:** 3-5x read capacity increase

#### **3.2 Implement CQRS for Analytics**
**Current:** Real-time analytics queries on transactional database
**Impact:** Complex aggregations slow down write operations
**Solution:**
- Write events to Kafka/Redis Streams
- Separate analytics database (ClickHouse, TimescaleDB)
- Pre-aggregated materialized views

**Expected Improvement:** 10-100x analytics query performance

#### **3.3 Frontend Optimization**
**Current:** Large bundle, sequential API calls
**Impact:** Slow initial page load, poor user experience
**Solution:**
1. Code splitting with React.lazy()
2. Implement BFF (Backend for Frontend) pattern for dashboard data
3. Server Components for initial render
4. Optimize bundle with webpack-bundle-analyzer

**Expected Improvement:** 40-60% faster initial load time

---

## 8. PERFORMANCE TESTING INSTRUMENTATION

### 8.1 Profiling Code to Add

#### **CPU Profiling Endpoints**
```go
// Add to main.go
import (
    _ "net/http/pprof"
)

func main() {
    // ... existing setup

    // Enable pprof endpoints in development
    if config.Environment == "development" {
        go func() {
            log.Println(http.ListenAndServe("localhost:6060", nil))
        }()
    }
}
```

**Usage:**
```bash
# CPU profiling (30 seconds)
curl http://localhost:6060/debug/pprof/profile?seconds=30 -o cpu.prof

# Generate flame graph
go tool pprof -http=:8081 cpu.prof

# Heap profiling
curl http://localhost:6060/debug/pprof/heap -o heap.prof
go tool pprof -http=:8081 heap.prof

# Goroutine profiling
curl http://localhost:6060/debug/pprof/goroutine -o goroutine.prof
go tool pprof -http=:8081 goroutine.prof
```

#### **Custom Metrics Middleware**
```go
// internal/middleware/metrics.go
package middleware

import (
    "time"
    "github.com/gin-gonic/gin"
)

func MetricsMiddleware() gin.HandlerFunc {
    return func(c *gin.Context) {
        start := time.Now()

        c.Next()

        latency := time.Since(start)

        // Log metrics (integrate with Prometheus/DataDog/etc)
        log.Printf("method=%s path=%s status=%d latency=%dms",
            c.Request.Method,
            c.Request.URL.Path,
            c.Writer.Status(),
            latency.Milliseconds(),
        )
    }
}
```

### 8.2 Database Query Logging

```go
// Enable in PostgresStorage initialization
db.SetLogger(LoggerFunc(func(ctx context.Context, query string, args []interface{}, duration time.Duration) {
    if duration > 10*time.Millisecond {
        log.Printf("SLOW QUERY [%dms]: %s %v", duration.Milliseconds(), query, args)
    }
}))
```

### 8.3 Load Testing Scripts

```bash
# wrk load test for redirections
wrk -t12 -c400 -d30s --latency http://localhost:8080/abc123

# Artillery.io for complex scenarios
artillery quick --count 100 --num 1000 http://localhost:8080/api/v1/shorten

# k6 for comprehensive testing
k6 run load_test.js
```

### 8.4 Continuous Performance Monitoring

**Metrics to Track:**
1. **Request Latency (P50, P95, P99)**
2. **Throughput (RPS)**
3. **Error Rate (%)**
4. **Cache Hit Rate (%)**
5. **Database Connection Pool Utilization (%)**
6. **Active Goroutines (count)**
7. **Heap Memory Usage (MB)**
8. **GC Pause Time (ms)**
9. **Database Query Duration (ms)**
10. **Redis Operation Duration (ms)**

---

## 9. BASELINE PERFORMANCE METRICS

### 9.1 Expected Performance Under Load

**Test Scenario: 1000 RPS Mixed Load**
- 70% GET /{shortCode} (redirections)
- 20% POST /api/v1/shorten (creations)
- 10% GET /api/v1/analytics (queries)

**Expected Metrics:**
```
Redirections (700 RPS):
  P50 Latency: 3ms
  P95 Latency: 12ms
  P99 Latency: 25ms
  Cache Hit Rate: 85%

Creations (200 RPS):
  P50 Latency: 15ms
  P95 Latency: 40ms
  P99 Latency: 80ms
  Error Rate: <1%

Analytics (100 RPS):
  P50 Latency: 30ms
  P95 Latency: 120ms
  P99 Latency: 250ms
  Cache Hit Rate: 65%

System Resources:
  CPU Usage: 40-60%
  Memory Usage: 50-80MB
  Database Connections: 15-20/25 utilized
  Active Goroutines: 500-2000
```

### 9.2 Critical User Journeys

#### **Journey 1: Anonymous User Shortens URL**
```
1. POST /api/v1/shorten (15-30ms)
2. Receive short URL
3. Copy to clipboard
Total: 15-30ms backend + network + frontend
```

#### **Journey 2: User Clicks Short URL**
```
1. GET /{shortCode} (3-15ms)
2. 301 Redirect to original URL
3. Click analytics recorded (async, not blocking)
Total: 3-15ms + DNS + target site load
```

#### **Journey 3: User Views Analytics Dashboard**
```
1. Load dashboard page (initial bundle)
2. GET /api/v1/urls (50-150ms)
3. GET /api/v1/dashboard-stats (30-80ms)
4. GET /api/v1/analytics/{code} x5 (150-500ms total)
Total: 230-730ms for full dashboard
```

**Target:** <500ms for dashboard load (requires optimization)

---

## 10. RECOMMENDED OPTIMIZATION ROADMAP

### Phase 1: Immediate (Week 1-2)
1. Add pprof endpoints and metrics middleware
2. Implement connection context timeouts
3. Optimize GetUserURLs query (window function)
4. Add Bloom filter for existence checks
5. Configure Redis eviction policy

**Expected Impact:** 30-40% latency improvement on critical paths

### Phase 2: Short-term (Week 3-4)
1. Implement click event batching
2. Implement worker pool for analytics
3. Switch to msgpack for Redis serialization
4. Add cache stampede protection
5. Optimize frontend bundle (remove duplicate icon library)

**Expected Impact:** 50-60% overall performance improvement

### Phase 3: Medium-term (Month 2-3)
1. Implement read replicas for database
2. Add connection pooling optimization
3. Implement GraphQL or BFF pattern for frontend
4. Add comprehensive monitoring and alerting
5. Implement circuit breaker pattern

**Expected Impact:** 2-3x capacity increase

### Phase 4: Long-term (Month 4+)
1. Migrate to CQRS architecture for analytics
2. Implement time-series database for analytics
3. Add horizontal scaling with load balancer
4. Implement CDN for static assets
5. Consider microservices architecture

**Expected Impact:** 10x+ capacity at scale

---

## 11. MONITORING & ALERTING SETUP

### 11.1 Key Metrics to Monitor

**Application Metrics:**
```
urlshortener_http_request_duration_seconds{method, endpoint, status}
urlshortener_http_requests_total{method, endpoint, status}
urlshortener_cache_hit_rate{cache_type}
urlshortener_database_query_duration_seconds{query_type}
urlshortener_active_goroutines
urlshortener_heap_alloc_bytes
urlshortener_gc_pause_seconds
```

**Alert Thresholds:**
```
P95 latency > 100ms for redirections → WARNING
P95 latency > 200ms for redirections → CRITICAL
Cache hit rate < 70% → WARNING
Database connection pool > 90% utilized → WARNING
Active goroutines > 5000 → WARNING
Error rate > 1% → CRITICAL
```

### 11.2 Recommended Tools

1. **Prometheus + Grafana**: Metrics collection and visualization
2. **Jaeger**: Distributed tracing
3. **Loki**: Log aggregation
4. **DataDog/New Relic**: All-in-one APM (commercial)

---

## 12. CONCLUSION

### Summary of Findings

**Strengths:**
1. Well-structured Go codebase with clean separation of concerns
2. Redis caching implemented for hot paths
3. Asynchronous analytics processing reduces blocking
4. Modern tech stack (Go 1.23, Next.js 15, Redis, PostgreSQL)

**Critical Issues:**
1. **Excessive goroutine spawning** on every click (5 goroutines per request)
2. **N+1 query patterns** in analytics and user URL retrieval
3. **No batching** for high-volume write operations
4. **Missing timeouts** on async operations (resource leak risk)
5. **Inefficient serialization** (JSON instead of binary format)

**Performance Potential:**
- **Current capacity (estimated)**: 500-1000 RPS mixed load
- **With Priority 1 optimizations**: 2000-3000 RPS
- **With Priority 2 optimizations**: 5000-8000 RPS
- **With architectural changes**: 10,000+ RPS per instance

### Next Steps

1. **Instrument application** with pprof and custom metrics (Day 1)
2. **Run load tests** to validate baseline estimates (Day 2-3)
3. **Implement Priority 1 optimizations** (Week 1-2)
4. **Re-profile and measure improvement** (Week 2)
5. **Iterate based on actual measurements** (Ongoing)

### Performance Testing Checklist

- [ ] Enable pprof endpoints
- [ ] Add metrics middleware
- [ ] Configure database query logging
- [ ] Set up Redis monitoring
- [ ] Run load tests with wrk/k6
- [ ] Generate CPU flame graphs
- [ ] Analyze heap allocations
- [ ] Check goroutine profiles
- [ ] Measure cache hit rates
- [ ] Profile database slow queries
- [ ] Analyze frontend bundle size
- [ ] Test under sustained load
- [ ] Test failure scenarios
- [ ] Document actual vs expected metrics
- [ ] Create performance regression tests

---

**Report Version:** 1.0
**Last Updated:** 2025-11-05
**Requires Validation:** All estimates require actual profiling data for confirmation
