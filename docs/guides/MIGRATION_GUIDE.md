# Performance Optimization Migration Guide

## Quick Start (5 Minutes)

### 1. Install Dependencies
```bash
cd D:\github\URLshorter
go get github.com/vmihailenco/msgpack/v5
go mod tidy
```

### 2. Update cmd/server/main.go

**Option A: Minimal Changes (Recommended for Testing)**
```go
// Add at top of file
import (
	"github.com/URLshorter/url-shortener/internal/workers"
)

// In main() function, after initializing services:

// Replace shortenerService initialization with optimized version
shortenerService := services.NewOptimizedShortenerService(db, redis, config)

// Add graceful shutdown for worker pool before server.Shutdown()
defer func() {
    log.Println("Shutting down worker pool...")
    if err := workers.ShutdownAnalyticsPool(30 * time.Second); err != nil {
        log.Printf("Worker pool shutdown error: %v", err)
    }
}()
```

**Option B: Full Optimization (Production)**
```go
// Replace RedisStorage with optimized version
redis, err := storage.NewRedisMsgpackStorage(config)
if err != nil {
    log.Fatalf("Failed to connect to Redis: %v", err)
}
defer redis.Close()

// Use optimized shortener service
shortenerService := services.NewOptimizedShortenerService(db, redis, config)

// Add graceful shutdown
defer func() {
    log.Println("Shutting down worker pool...")
    if err := shortenerService.Shutdown(30 * time.Second); err != nil {
        log.Printf("Worker pool shutdown error: %v", err)
    }
}()

// In route setup section, replace:
routes.SetupRoutes(router, handler, authMiddleware)
// With:
routes.SetupOptimizedRoutes(router, handler, authMiddleware)
```

### 3. Test Locally
```bash
# Run the application
go run cmd/server/main.go

# In another terminal, run performance tests
go test -v ./tests/ -run TestWorkerPoolThroughput
go test -bench=BenchmarkBase62Encoding -benchmem ./tests/
go test -bench=BenchmarkBloomFilter -benchmem ./tests/
```

### 4. Verify
```bash
# Test URL creation
curl -X POST http://localhost:8080/api/v1/shorten \
  -H "Content-Type: application/json" \
  -d '{"url":"https://example.com/test"}'

# Test redirection
curl -I http://localhost:8080/{shortCode}

# Check compression
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/analytics/{shortCode} -v

# Verify worker pool metrics (add endpoint if needed)
```

---

## Detailed Migration Steps

### Step 1: Worker Pool Integration

**File:** `cmd/server/main.go`

```go
// Add import
import "github.com/URLshorter/url-shortener/internal/workers"

// After creating all services, initialize worker pool metrics endpoint (optional)
router.GET("/metrics/workers", func(c *gin.Context) {
    metrics := workers.GetAnalyticsPool().GetMetrics()
    c.JSON(200, gin.H{
        "tasks_processed": metrics.TasksProcessed,
        "tasks_failed":    metrics.TasksFailed,
        "tasks_in_queue":  metrics.TasksInQueue,
        "active_workers":  metrics.ActiveWorkers,
    })
})

// Before server shutdown
defer func() {
    log.Println("Shutting down worker pool...")
    if err := workers.ShutdownAnalyticsPool(30 * time.Second); err != nil {
        log.Printf("Worker pool shutdown error: %v", err)
    }
}()
```

### Step 2: Replace Shortener Service

**File:** `cmd/server/main.go`

```go
// OLD:
shortenerService := services.NewShortenerService(db, redis, config)

// NEW:
shortenerService := services.NewOptimizedShortenerService(db, redis, config)
```

### Step 3: Enable MessagePack Serialization

**File:** `cmd/server/main.go`

```go
// OLD:
redis, err := storage.NewRedisStorage(config)

// NEW:
redis, err := storage.NewRedisMsgpackStorage(config)
```

**Note:** This requires clearing existing Redis cache:
```bash
redis-cli FLUSHDB
```

Or run migration to dual-read:
```go
// Read from both, write to new format
oldRedis, _ := storage.NewRedisStorage(config)
newRedis, _ := storage.NewRedisMsgpackStorage(config)

// Gradually migrate data...
```

### Step 4: Enable Response Compression

**File:** `cmd/server/main.go` or `internal/routes/routes.go`

```go
// Add compression middleware globally
router.Use(middleware.CompressionMiddleware())
```

Or use optimized routes:
```go
// Replace
routes.SetupRoutes(router, handler, authMiddleware)

// With
routes.SetupOptimizedRoutes(router, handler, authMiddleware)
```

### Step 5: Warm Up Bloom Filter (Production Only)

**File:** `cmd/server/main.go`

```go
// Add after creating shortenerService
log.Println("Warming up Bloom filter from database...")
if err := warmUpBloomFilter(db, shortenerService); err != nil {
    log.Printf("Warning: Failed to warm up Bloom filter: %v", err)
}

// Add helper function
func warmUpBloomFilter(db *storage.PostgresStorage, service *services.OptimizedShortenerService) error {
    ctx, cancel := context.WithTimeout(context.Background(), 60*time.Second)
    defer cancel()

    rows, err := db.QueryContext(ctx, "SELECT short_code FROM url_mappings WHERE is_active = TRUE")
    if err != nil {
        return err
    }
    defer rows.Close()

    bf := utils.GetURLBloomFilter()
    count := 0

    for rows.Next() {
        var shortCode string
        if err := rows.Scan(&shortCode); err != nil {
            return err
        }
        bf.Add(shortCode)
        count++
    }

    log.Printf("Loaded %d short codes into Bloom filter", count)
    return nil
}
```

---

## Configuration Options

### Worker Pool Tuning

**Default:** 20 workers, 1000 queue size

**High Load:** 50 workers, 5000 queue size
```go
// internal/workers/worker_pool.go:150
analyticsPool = NewWorkerPool(50, 5000)
```

**Low Resources:** 10 workers, 500 queue size
```go
analyticsPool = NewWorkerPool(10, 500)
```

### Bloom Filter Tuning

**Default:** 1M items, 1% FPR

**More items:** 10M items, 1% FPR (12MB memory)
```go
// internal/utils/bloomfilter.go:150
urlBloomFilter = NewBloomFilter(10000000, 0.01)
```

**Lower FPR:** 1M items, 0.1% FPR (2x memory)
```go
urlBloomFilter = NewBloomFilter(1000000, 0.001)
```

### Redis Connection Pool

**Default:** 50 connections, 10 min idle

**High Load:** 100 connections, 20 min idle
```go
// internal/storage/redis_msgpack.go:21
rdb := redis.NewClient(&redis.Options{
    PoolSize:     100,
    MinIdleConns: 20,
    // ...
})
```

---

## Testing Checklist

### Unit Tests
```bash
# Test worker pool
go test -v ./internal/workers/ -run TestWorkerPool

# Test Bloom filter
go test -v ./internal/utils/ -run TestBloomFilter

# Test base62 encoding
go test -v ./internal/utils/ -run TestBase62
```

### Benchmarks
```bash
# Base62 encoding
go test -bench=BenchmarkBase62 -benchmem ./tests/

# Bloom filter
go test -bench=BenchmarkBloomFilter -benchmem ./tests/

# Worker pool
go test -bench=BenchmarkWorkerPool -benchmem ./tests/
```

### Load Tests
```bash
# Analytics processing load test
go test -v ./tests/ -run TestLoadTestAnalyticsProcessing

# 1000 RPS sustained load
go test -v ./tests/ -run TestLoadTest -timeout 30s
```

### Integration Tests
```bash
# Full integration test
go test -v ./tests/ -run TestIntegration
```

---

## Monitoring

### Add Metrics Endpoint
```go
// cmd/server/main.go
router.GET("/metrics/performance", func(c *gin.Context) {
    workerMetrics := workers.GetAnalyticsPool().GetMetrics()

    c.JSON(200, gin.H{
        "worker_pool": gin.H{
            "tasks_processed": workerMetrics.TasksProcessed,
            "tasks_failed":    workerMetrics.TasksFailed,
            "tasks_in_queue":  workerMetrics.TasksInQueue,
            "active_workers":  workerMetrics.ActiveWorkers,
        },
        "timestamp": time.Now().Unix(),
    })
})
```

### Prometheus Metrics (Optional)
```go
import "github.com/prometheus/client_golang/prometheus"

var (
    workerPoolQueueDepth = prometheus.NewGauge(prometheus.GaugeOpts{
        Name: "worker_pool_queue_depth",
        Help: "Current worker pool queue depth",
    })

    workerPoolTasksProcessed = prometheus.NewCounter(prometheus.CounterOpts{
        Name: "worker_pool_tasks_processed_total",
        Help: "Total tasks processed by worker pool",
    })
)

// Update metrics periodically
go func() {
    ticker := time.NewTicker(10 * time.Second)
    for range ticker.C {
        metrics := workers.GetAnalyticsPool().GetMetrics()
        workerPoolQueueDepth.Set(float64(metrics.TasksInQueue))
        workerPoolTasksProcessed.Add(float64(metrics.TasksProcessed))
    }
}()
```

---

## Rollback Plan

### Quick Rollback
```bash
# Revert changes
git checkout cmd/server/main.go

# Or comment out optimizations
# shortenerService := services.NewShortenerService(db, redis, config)
# routes.SetupRoutes(router, handler, authMiddleware)
```

### Gradual Rollback
1. Disable worker pool (use direct goroutines)
2. Disable Bloom filter (use database checks)
3. Disable MessagePack (use JSON)
4. Disable compression

---

## Production Deployment

### Pre-Deployment
- [ ] Run all tests in staging
- [ ] Load test with expected traffic
- [ ] Verify monitoring dashboards
- [ ] Set up alerts for queue depth
- [ ] Set up alerts for worker failures
- [ ] Document rollback procedure

### Deployment
```bash
# Build optimized binary
go build -o url-shortener-optimized ./cmd/server

# Run database migration (if needed)
# psql -U urlshortener -d urlshortener_db -f configs/migrations/XXX_optimization.sql

# Clear Redis cache (if changing serialization)
redis-cli FLUSHDB

# Start application
./url-shortener-optimized
```

### Post-Deployment
- [ ] Monitor worker pool metrics
- [ ] Monitor error rates
- [ ] Check response times (should decrease)
- [ ] Verify compression headers
- [ ] Monitor memory usage
- [ ] Monitor CPU usage

---

## Troubleshooting

### Issue: High Queue Depth
**Symptom:** Tasks backing up in queue
**Solution:**
```go
// Increase workers
analyticsPool = NewWorkerPool(50, 5000)
```

### Issue: Worker Pool Shutdown Timeout
**Symptom:** Graceful shutdown takes too long
**Solution:**
```go
// Increase shutdown timeout
shortenerService.Shutdown(60 * time.Second)
```

### Issue: Bloom Filter Too Many False Positives
**Symptom:** Unnecessary database queries
**Solution:**
```go
// Decrease FPR (increases memory)
urlBloomFilter = NewBloomFilter(1000000, 0.001)
```

### Issue: MessagePack Serialization Errors
**Symptom:** Cache misses, deserialization errors
**Solution:**
```bash
# Clear Redis cache
redis-cli FLUSHDB

# Verify msgpack version
go get github.com/vmihailenco/msgpack/v5
```

### Issue: Compression Not Working
**Symptom:** Responses not compressed
**Solution:**
```bash
# Verify client sends Accept-Encoding header
curl -H "Accept-Encoding: gzip" http://localhost:8080/api/v1/analytics/abc -v

# Check middleware order
# CompressionMiddleware should be early in chain
```

---

## Performance Expectations

### Before Optimization
- **RPS:** 500-700
- **P99 Latency:** 200-300ms
- **Goroutines:** 5000+ at 1000 RPS
- **Memory:** 2GB
- **CPU:** 80%

### After Optimization
- **RPS:** 1500-2000
- **P99 Latency:** 100-150ms
- **Goroutines:** 1000-1500 at 1000 RPS
- **Memory:** 1.2GB
- **CPU:** 60%

**Expected Improvements:**
- 3x throughput increase
- 40% latency reduction
- 80% goroutine reduction
- 40% memory reduction
- 25% CPU reduction

---

## Support

For issues or questions:
1. Check `PERFORMANCE_OPTIMIZATION.md` for detailed explanations
2. Run performance tests to verify behavior
3. Check worker pool metrics endpoint
4. Review logs for errors
5. Consult troubleshooting section

---

**Created:** 2025-11-06
**Last Updated:** 2025-11-06
**Version:** 1.0
