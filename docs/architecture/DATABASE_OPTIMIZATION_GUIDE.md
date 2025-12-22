# Database Optimization Implementation Guide

## Quick Start

This guide walks through implementing the database optimizations step-by-step with zero downtime.

### Timeline
- **Phase 1 (Preparation)**: 2-3 hours
- **Phase 2 (Index Creation)**: 3-4 hours (off-peak)
- **Phase 3 (Application Updates)**: 1-2 hours (rolling deployment)
- **Phase 4 (Validation)**: Ongoing
- **Total**: ~8-10 hours

---

## Phase 1: Preparation (No Downtime)

### 1.1 Backup Current State

```bash
# Create full database backup
pg_dump -U urlshortener -h localhost urlshortener_db > backup_pre_optimization.sql

# Verify backup
psql -U urlshortener -h localhost urlshortener_db < backup_pre_optimization.sql --dry-run
```

### 1.2 Generate Baseline Metrics

Run these queries to establish performance baselines:

```sql
-- Current query performance
SELECT
    query,
    calls,
    mean_exec_time,
    max_exec_time,
    total_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 50
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Current index usage
SELECT
    schemaname,
    tablename,
    indexname,
    idx_scan,
    idx_tup_read,
    idx_tup_fetch
FROM pg_stat_user_indexes
ORDER BY idx_scan DESC;

-- Cache hit ratio (goal: >99%)
SELECT
    sum(heap_blks_read) as disk_read,
    sum(heap_blks_hit) as cache_hit,
    sum(heap_blks_hit) / (sum(heap_blks_hit) + sum(heap_blks_read)) as ratio
FROM pg_statio_user_tables;

-- Table bloat
SELECT
    current_database(),
    schemaname,
    tablename,
    pg_size_pretty(pg_total_relation_size(schemaname||'.'||tablename)) as size
FROM pg_tables
WHERE schemaname = 'public'
ORDER BY pg_total_relation_size(schemaname||'.'||tablename) DESC;
```

Save this output for comparison after optimization.

### 1.3 Deploy Code Changes (No Database Changes Yet)

Deploy these files with feature flags disabled:

```bash
# Copy new files to project
cp internal/services/click_batch_processor.go <your-project>/internal/services/
cp internal/storage/postgres_optimizations.go <your-project>/internal/storage/

# Update configuration
# See section 1.4 below for environment variables
```

### 1.4 Update Configuration

Add these environment variables to `.env` or configuration system:

```env
# Feature Flags (disabled during Phase 1)
ENABLE_BATCH_CLICKS=false
ENABLE_OPTIMIZED_QUERIES=false
ENABLE_QUERY_CACHING=false
ENABLE_ANALYTICS_AGGREGATION=false

# Batch Processor Settings (for Phase 3)
CLICK_BATCH_SIZE=100
CLICK_BATCH_TIMEOUT_MS=5000
CLICK_QUEUE_SIZE=1000

# Database Connection Pool (current settings)
DB_MAX_CONNECTIONS=25
DB_IDLE_CONNECTIONS=12
DB_MAX_CONN_LIFETIME=3600
DB_CONN_TIMEOUT=30

# Cache Settings
REDIS_ANALYTICS_TTL_SECONDS=300
REDIS_URL_TTL_SECONDS=86400
```

---

## Phase 2: Index Creation (Optional Maintenance Window)

### 2.1 Schedule Index Creation

Choose low-traffic window (typically 2-4 AM for most applications):

```bash
# SSH to database server
ssh postgres@db.example.com

# Run migration with time tracking
cd /var/lib/postgresql
time psql -U urlshortener -d urlshortener_db -f configs/migrations/003_performance_optimization.sql
```

### 2.2 Monitor Index Creation Progress

```bash
# In another terminal, monitor progress
psql -U urlshortener -d urlshortener_db

-- Check index creation status
SELECT
    schemaname,
    tablename,
    indexname,
    idx_blks_read,
    idx_blks_hit
FROM pg_stat_user_indexes
ORDER BY indexname;

-- Check for blocking queries
SELECT
    pid,
    query,
    state,
    wait_event_type,
    wait_event
FROM pg_stat_activity
WHERE state != 'idle';
```

Expected timeline:
- **Small tables** (< 1M rows): 1-2 minutes
- **Medium tables** (1M-10M rows): 5-15 minutes
- **Large tables** (> 10M rows): 30-60 minutes

### 2.3 Verify Index Creation

```sql
-- Verify all indexes were created
SELECT
    tablename,
    indexname,
    indexdef
FROM pg_indexes
WHERE tablename IN (
    'click_events',
    'url_mappings',
    'geographic_analytics',
    'device_analytics',
    'time_analytics'
)
AND indexname LIKE 'idx_%'
ORDER BY tablename, indexname;

-- Expected count: ~15 new indexes

-- Check index sizes
SELECT
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size
FROM pg_stat_user_indexes
WHERE tablename IN ('click_events', 'url_mappings')
AND indexname LIKE 'idx_%'
ORDER BY pg_relation_size(indexrelid) DESC;
```

### 2.4 Create Analytics Tables

```bash
psql -U urlshortener -d urlshortener_db -f configs/migrations/004_analytics_tables.sql
```

---

## Phase 3: Application Updates (Rolling Deployment)

### 3.1 Deploy New Code with Feature Flags

Deploy application with new code but feature flags disabled:

```bash
# Build new version
go build -o url-shortener cmd/server/main.go

# Rolling deployment (1 instance at a time)
# Keep feature flags disabled
ENABLE_BATCH_CLICKS=false
ENABLE_OPTIMIZED_QUERIES=false
```

### 3.2 Initialize Batch Processor in Main Application

Update `cmd/server/main.go` to initialize batch processor:

```go
package main

import (
    // ... existing imports
    "github.com/URLshorter/url-shortener/internal/services"
)

func main() {
    // ... existing setup code

    // Initialize batch click processor
    clickBatchProcessor := services.NewClickBatchProcessor(
        dbStorage,
        redisStorage,
        100,              // Batch size
        5*time.Second,    // Batch timeout
    )

    // Start processor
    ctx := context.Background()
    clickBatchProcessor.Start(ctx)
    defer clickBatchProcessor.Stop()

    // Update shortener service to use batch processor
    // Instead of: shortenerService.RecordClick()
    // Use: clickBatchProcessor.AddClick(event)

    // ... rest of setup
}
```

### 3.3 Gradually Enable Features

Update `ShortenerService.RecordClick()` to use batch processor:

```go
// File: internal/services/shortener.go

func (s *ShortenerService) RecordClick(shortCode, clientIP, userAgent, referrer string) error {
    // Generate ID for click event
    id, err := utils.GenerateID()
    if err != nil {
        return fmt.Errorf("failed to generate click event ID: %w", err)
    }

    // Create click event
    event := &models.ClickEvent{
        ID:          id,
        ShortCode:   shortCode,
        ClickedAt:   time.Now(),
        IPAddress:   clientIP,
        UserAgent:   userAgent,
        Referrer:    referrer,
        CountryCode: s.getCountryFromIP(clientIP),
    }

    // Feature flag: Use batch processor or individual inserts
    if os.Getenv("ENABLE_BATCH_CLICKS") == "true" {
        return s.batchProcessor.AddClick(event)
    }

    // Fallback to individual inserts
    go func() {
        if err := s.db.SaveClickEvent(event); err != nil {
            fmt.Printf("Failed to save click event: %v\n", err)
        }
    }()

    // ... rest of RecordClick
}
```

### 3.4 Update Analytics Queries

Replace `GetAnalytics()` calls with optimized version:

```go
// File: internal/storage/postgres.go

// Update analytics service to use optimized query
func (p *PostgresStorage) GetAnalyticsOptimized(ctx context.Context, shortCode string, days int) (*models.AnalyticsResponse, error) {
    // See postgres_optimizations.go for implementation
}

// Update with caching layer
func (p *PostgresStorage) GetAnalyticsWithCache(
    ctx context.Context,
    shortCode string,
    days int,
    redis *RedisStorage,
) (*models.AnalyticsResponse, error) {
    // See postgres_optimizations.go for implementation
}
```

### 3.5 Enable Features Incrementally

Start with one feature at a time and monitor metrics:

**Step 1: Enable Batch Clicks (30% traffic)**
```env
ENABLE_BATCH_CLICKS=true
ENABLE_OPTIMIZED_QUERIES=false
ENABLE_QUERY_CACHING=false
```

Wait 2 hours, monitor metrics. If healthy:

**Step 2: Enable Optimized Queries (50% traffic)**
```env
ENABLE_BATCH_CLICKS=true
ENABLE_OPTIMIZED_QUERIES=true
ENABLE_QUERY_CACHING=false
```

Wait 2 hours, monitor metrics. If healthy:

**Step 3: Enable Query Caching (100% traffic)**
```env
ENABLE_BATCH_CLICKS=true
ENABLE_OPTIMIZED_QUERIES=true
ENABLE_QUERY_CACHING=true
```

### 3.6 Rollback Plan (If Issues Occur)

If performance degrades:

```bash
# Quickly disable new features
export ENABLE_BATCH_CLICKS=false
export ENABLE_OPTIMIZED_QUERIES=false
export ENABLE_QUERY_CACHING=false

# Restart application
systemctl restart url-shortener

# Investigate and rollback if needed
```

---

## Phase 4: Validation & Monitoring

### 4.1 Key Metrics to Monitor

```bash
# Application Metrics
- Request latency (p50, p95, p99)
- Analytics query time
- Click processing latency
- Cache hit rate
- Database connection pool utilization

# Database Metrics
- Query execution time
- Database CPU usage
- Database memory usage
- Slow query log
- Index usage
```

### 4.2 Performance Comparison Queries

Run these same queries from Phase 1 to compare:

```sql
-- Query performance after optimization
SELECT
    query,
    calls,
    mean_exec_time,
    max_exec_time,
    total_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 50
ORDER BY mean_exec_time DESC
LIMIT 20;

-- Expected results:
-- - Most queries should be 50-80% faster
-- - Fewer slow queries visible
-- - Click insert queries should be gone/minimal
```

### 4.3 Expected Performance Gains

| Operation | Before | After | Improvement |
|-----------|--------|-------|-------------|
| Analytics request | 200-300ms | 40-50ms | 75-80% |
| Click insertion (100 events) | 500ms | 5-10ms | 95% |
| Repeated analytics (cached) | 200ms | 1-2ms | 99% |
| Database write QPS @ 5000 RPS | 5000+ | 50-100 | 95% |

### 4.4 Optimize Batch Parameters

If metrics show room for improvement, adjust batch settings:

```env
# Current (conservative)
CLICK_BATCH_SIZE=100
CLICK_BATCH_TIMEOUT_MS=5000

# More aggressive (if latency is acceptable)
CLICK_BATCH_SIZE=250
CLICK_BATCH_TIMEOUT_MS=2000

# Less aggressive (if latency increases)
CLICK_BATCH_SIZE=50
CLICK_BATCH_TIMEOUT_MS=10000

# Restart and retest
systemctl restart url-shortener
```

---

## Phase 5: Ongoing Maintenance

### 5.1 Regular Index Maintenance

```sql
-- Weekly: Check for index bloat
SELECT
    schemaname,
    tablename,
    indexname,
    pg_size_pretty(pg_relation_size(indexrelid)) as size,
    idx_scan
FROM pg_stat_user_indexes
WHERE schemaname = 'public'
ORDER BY indexname;

-- Monthly: Rebuild bloated indexes if needed
REINDEX INDEX CONCURRENTLY idx_click_events_short_code_time;
```

### 5.2 Update Statistics

```sql
-- Daily at off-peak time
ANALYZE click_events;
ANALYZE url_mappings;
ANALYZE geographic_analytics;
ANALYZE device_analytics;
```

### 5.3 Monitor Slow Queries

```sql
-- Check pg_stat_statements extension
SELECT
    query,
    calls,
    mean_exec_time,
    max_exec_time
FROM pg_stat_statements
WHERE mean_exec_time > 100
ORDER BY mean_exec_time DESC
LIMIT 20;
```

### 5.4 Batch Processor Monitoring

Monitor batch processor metrics in application logs:

```bash
# Check processor metrics every 5 minutes
curl http://localhost:8080/api/metrics

# Expected metrics:
# - total_clicks: increasing steadily
# - batches_added: every 5 seconds or when queue reaches 100
# - failed_clicks: minimal (< 0.1%)
# - queue_size: < batch_size * 5
```

---

## Configuration Examples

### Development Environment

```env
DB_MAX_CONNECTIONS=10
DB_IDLE_CONNECTIONS=5
ENABLE_BATCH_CLICKS=false
ENABLE_OPTIMIZED_QUERIES=false
ENABLE_QUERY_CACHING=true
CLICK_BATCH_SIZE=50
CLICK_BATCH_TIMEOUT_MS=10000
```

### Staging Environment

```env
DB_MAX_CONNECTIONS=20
DB_IDLE_CONNECTIONS=10
ENABLE_BATCH_CLICKS=true
ENABLE_OPTIMIZED_QUERIES=true
ENABLE_QUERY_CACHING=true
CLICK_BATCH_SIZE=100
CLICK_BATCH_TIMEOUT_MS=5000
```

### Production Environment

```env
DB_MAX_CONNECTIONS=32
DB_IDLE_CONNECTIONS=8
ENABLE_BATCH_CLICKS=true
ENABLE_OPTIMIZED_QUERIES=true
ENABLE_QUERY_CACHING=true
CLICK_BATCH_SIZE=100
CLICK_BATCH_TIMEOUT_MS=5000
REDIS_ANALYTICS_TTL_SECONDS=300
REDIS_URL_TTL_SECONDS=86400
```

---

## Troubleshooting

### Issue: High Latency with Batch Processor

**Symptoms**: Analytics queries taking longer after enabling batch clicks

**Solution**:
1. Increase `REDIS_ANALYTICS_TTL_SECONDS` to 600 (10 minutes)
2. Check `queue_size` metrics - if consistently high, increase `CLICK_BATCH_SIZE`
3. Verify `failed_clicks` metric - high values indicate queue is full

### Issue: Index Creation Timeout

**Symptoms**: Migration script hangs on index creation

**Solution**:
```sql
-- Check blocking queries
SELECT * FROM pg_stat_activity WHERE state = 'active';

-- Kill blocking query if safe
SELECT pg_terminate_backend(pid) FROM pg_stat_activity WHERE query LIKE '%click_events%' AND pid != pg_backend_pid();

-- Retry index creation
CREATE INDEX CONCURRENTLY idx_click_events_short_code_time ON click_events(short_code, country_code, clicked_at DESC);
```

### Issue: Cache Hit Rate Low

**Symptoms**: Query caching not providing expected performance gain

**Possible Causes**:
- Cache TTL too short: Increase `REDIS_ANALYTICS_TTL_SECONDS`
- Distributed queries: Each user getting different date ranges
- Cache keys not normalizing: Check query parameters

**Solution**:
```sql
-- Check Redis cache size
INFO memory

-- Clear cache and restart
FLUSHALL

-- Verify cache key strategy
SELECT * FROM redis_keys PATTERN 'analytics:*'
```

---

## Success Criteria

Optimization is complete when:

- [ ] All new indexes created successfully
- [ ] Query latency reduced 50%+ from baseline
- [ ] Batch processor healthy (< 0.1% failed clicks)
- [ ] Cache hit rate > 70%
- [ ] Database CPU usage down 30-50%
- [ ] No slow queries > 100ms in pg_stat_statements
- [ ] Zero data integrity issues
- [ ] All tests passing

---

## References

- PostgreSQL EXPLAIN: https://www.postgresql.org/docs/current/sql-explain.html
- Index Performance: https://www.postgresql.org/docs/current/indexes.html
- Query Tuning: https://www.postgresql.org/docs/current/using-explain.html
- Connection Pooling: https://wiki.postgresql.org/wiki/Number_Of_Database_Connections

---

**Last Updated**: 2025-11-06
**Status**: Ready for Implementation
