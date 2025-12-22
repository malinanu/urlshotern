# Distributed System Performance Optimization - Implementation Guide

## Quick Start

This guide provides step-by-step instructions for implementing the distributed system performance optimizations in the URL shortener application.

---

## Prerequisites

- Go 1.23+
- Docker & Docker Compose
- Kubernetes cluster (for production deployment)
- k6 (for load testing)
- Apache Bench (ab) - optional

---

## Phase 1: Local Development Setup (Week 1)

### Step 1: Update Go Dependencies

Add required packages to `go.mod`:

```bash
go get github.com/go-redis/redis/v8
go get github.com/vmihailenco/msgpack/v5
go get github.com/lib/pq
```

### Step 2: Test Optimized Storage Layers

The optimized storage implementations are backward compatible:

1. **PostgreSQL Optimized Storage**: Replace `storage.NewPostgresStorage()` with `storage.NewPostgresOptimizedStorage()`
2. **Redis Cluster Storage**: For development, the cluster storage works with single-node Redis

To test:

```bash
# Start services
docker-compose -f docker/docker-compose.local.yml up -d

# Run application with optimized storage
go run cmd/server/main.go
```

### Step 3: Verify Connection Pooling

Check connection pool statistics:

```go
// In your main.go or monitoring endpoint
dbStats := db.Stats()
log.Printf("DB Pool - Open: %d, Idle: %d, InUse: %d",
    dbStats.OpenConnections,
    dbStats.Idle,
    dbStats.InUse)

redisStats := redis.GetPoolStats()
for addr, stats := range redisStats {
    log.Printf("Redis %s - Hits: %d, Misses: %d, Timeouts: %d",
        addr, stats.Hits, stats.Misses, stats.Timeouts)
}
```

---

## Phase 2: Redis Cluster Setup (Week 2)

### Option A: Local Development with Docker Compose

```bash
# Start Redis cluster
docker-compose -f docker/docker-compose.redis-cluster.yml up -d

# Wait for cluster initialization
sleep 30

# Verify cluster status
docker exec redis-node-1 redis-cli cluster info
```

Update your application configuration:

```go
// In configs/config.go
RedisClusterMode: true
RedisClusterAddrs: []string{
    "redis-node-1:6379",
    "redis-node-2:6380",
    "redis-node-3:6381",
}
```

### Option B: Production Kubernetes Deployment

```bash
# Apply Redis cluster configuration
kubectl apply -f k8s/redis-cluster.yaml

# Wait for all pods to be ready
kubectl wait --for=condition=ready pod -l app=redis-cluster -n urlshortener --timeout=300s

# Initialize cluster (one-time operation)
kubectl exec -it redis-cluster-0 -n urlshortener -- redis-cli --cluster create \
    redis-cluster-0.redis-cluster-headless:6379 \
    redis-cluster-1.redis-cluster-headless:6379 \
    redis-cluster-2.redis-cluster-headless:6379 \
    redis-cluster-3.redis-cluster-headless:6379 \
    redis-cluster-4.redis-cluster-headless:6379 \
    redis-cluster-5.redis-cluster-headless:6379 \
    --cluster-replicas 1 --cluster-yes
```

---

## Phase 3: Circuit Breaker & Bulkhead Implementation (Week 3)

### Step 1: Add Resilience Patterns to Services

Update `internal/services/shortener.go`:

```go
import (
    "github.com/URLshorter/url-shortener/internal/resilience"
)

type ShortenerService struct {
    db          *storage.PostgresStorage
    redis       *storage.RedisStorage
    config      *configs.Config

    // Resilience patterns
    dbCircuitBreaker    *resilience.CircuitBreaker
    redisCircuitBreaker *resilience.CircuitBreaker
    dbBulkhead          *resilience.Bulkhead
    redisBulkhead       *resilience.Bulkhead
}

func NewShortenerService(db *storage.PostgresStorage, redis *storage.RedisStorage, config *configs.Config) *ShortenerService {
    service := &ShortenerService{
        db:     db,
        redis:  redis,
        config: config,
    }

    // Initialize circuit breakers
    service.dbCircuitBreaker = resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
        Name:                   "database",
        MaxRequests:            1,
        Interval:               60 * time.Second,
        Timeout:                30 * time.Second,
        MaxConsecutiveFailures: 5,
        MaxFailureRatio:        0.5,
    })

    service.redisCircuitBreaker = resilience.NewCircuitBreaker(resilience.CircuitBreakerConfig{
        Name:                   "redis",
        MaxRequests:            1,
        Interval:               60 * time.Second,
        Timeout:                30 * time.Second,
        MaxConsecutiveFailures: 5,
        MaxFailureRatio:        0.5,
    })

    // Initialize bulkheads
    service.dbBulkhead = resilience.NewBulkhead(resilience.BulkheadConfig{
        Name:          "database",
        MaxConcurrent: 25,
        MaxQueueDepth: 50,
    })

    service.redisBulkhead = resilience.NewBulkhead(resilience.BulkheadConfig{
        Name:          "redis",
        MaxConcurrent: 50,
        MaxQueueDepth: 100,
    })

    return service
}

// Example usage in GetOriginalURL
func (s *ShortenerService) GetOriginalURL(shortCode string) (*models.URLMapping, error) {
    // Try Redis with circuit breaker and bulkhead
    var mapping *models.URLMapping

    err := s.redisBulkhead.Call(func() error {
        return s.redisCircuitBreaker.Call(func() error {
            var err error
            mapping, err = s.redis.GetURLMapping(shortCode)
            return err
        })
    })

    if err != nil && err != storage.ErrCacheKeyNotFound {
        // Log cache error but continue with database lookup
        log.Printf("Redis error (circuit breaker): %v", err)
    }

    // If not in cache, check database with circuit breaker and bulkhead
    if mapping == nil {
        err = s.dbBulkhead.Call(func() error {
            return s.dbCircuitBreaker.Call(func() error {
                var err error
                mapping, err = s.db.GetURLMappingByShortCode(shortCode)
                return err
            })
        })

        if err != nil {
            return nil, err
        }

        // Cache the result
        cacheTTL := 24 * time.Hour
        if mapping.ExpiresAt != nil {
            timeUntilExpiry := time.Until(*mapping.ExpiresAt)
            if timeUntilExpiry < cacheTTL {
                cacheTTL = timeUntilExpiry
            }
        }
        s.redis.SetURLMapping(shortCode, mapping, cacheTTL)
    }

    return mapping, nil
}
```

### Step 2: Add Health Check Endpoints

Create `internal/handlers/health.go`:

```go
package handlers

import (
    "github.com/gin-gonic/gin"
    "net/http"
)

func (h *Handler) GetHealthStatus(c *gin.Context) {
    health := map[string]interface{}{
        "status": "healthy",
        "services": map[string]interface{}{
            "database": map[string]interface{}{
                "healthy": h.db.IsHealthy(),
                "circuit_breaker": h.shortener.dbCircuitBreaker.State().String(),
                "bulkhead": h.shortener.dbBulkhead.Stats(),
            },
            "redis": map[string]interface{}{
                "healthy": h.redis.IsHealthy(),
                "circuit_breaker": h.shortener.redisCircuitBreaker.State().String(),
                "bulkhead": h.shortener.redisBulkhead.Stats(),
            },
        },
    }

    c.JSON(http.StatusOK, health)
}
```

---

## Phase 4: Kubernetes Deployment (Week 4)

### Step 1: Create Namespace and Secrets

```bash
# Create namespace
kubectl create namespace urlshortener

# Create PostgreSQL credentials
kubectl create secret generic postgres-credentials \
    --from-literal=username=urlshortener \
    --from-literal=password=<YOUR_SECURE_PASSWORD> \
    -n urlshortener

# Create Redis credentials
kubectl create secret generic redis-credentials \
    --from-literal=password=<YOUR_SECURE_PASSWORD> \
    -n urlshortener

# Create JWT secret
kubectl create secret generic jwt-secret \
    --from-literal=secret=<YOUR_JWT_SECRET> \
    -n urlshortener

# Create Grafana credentials
kubectl create secret generic grafana-credentials \
    --from-literal=password=<YOUR_GRAFANA_PASSWORD> \
    -n monitoring
```

### Step 2: Deploy Infrastructure

```bash
# Deploy PostgreSQL with read replicas
kubectl apply -f k8s/postgres.yaml

# Wait for PostgreSQL to be ready
kubectl wait --for=condition=ready pod -l app=postgres -n urlshortener --timeout=300s

# Deploy Redis cluster
kubectl apply -f k8s/redis-cluster.yaml

# Wait for Redis to be ready
kubectl wait --for=condition=ready pod -l app=redis-cluster -n urlshortener --timeout=300s
```

### Step 3: Deploy Application

```bash
# Build and push Docker image
docker build -t <YOUR_REGISTRY>/urlshortener:latest -f docker/Dockerfile .
docker push <YOUR_REGISTRY>/urlshortener:latest

# Update image in deployment.yaml, then deploy
kubectl apply -f k8s/deployment.yaml

# Deploy HPA and PDB
kubectl apply -f k8s/hpa.yaml

# Deploy network policies
kubectl apply -f k8s/network-policies.yaml

# Deploy ingress
kubectl apply -f k8s/ingress.yaml
```

### Step 4: Verify Deployment

```bash
# Check pod status
kubectl get pods -n urlshortener

# Check HPA status
kubectl get hpa -n urlshortener

# Check ingress
kubectl get ingress -n urlshortener

# View logs
kubectl logs -f deployment/urlshortener-app -n urlshortener

# Port forward for testing
kubectl port-forward service/urlshortener-service 8080:80 -n urlshortener
```

---

## Phase 5: Monitoring & Observability (Week 5)

### Step 1: Deploy Monitoring Stack

```bash
# Create monitoring namespace
kubectl create namespace monitoring

# Deploy Prometheus and Grafana
kubectl apply -f k8s/monitoring.yaml

# Wait for monitoring stack to be ready
kubectl wait --for=condition=ready pod -l app=prometheus -n monitoring --timeout=300s
kubectl wait --for=condition=ready pod -l app=grafana -n monitoring --timeout=300s
```

### Step 2: Access Dashboards

```bash
# Port forward Grafana
kubectl port-forward service/grafana 3000:3000 -n monitoring

# Port forward Prometheus
kubectl port-forward service/prometheus 9090:9090 -n monitoring
```

Access:
- Grafana: http://localhost:3000 (admin / password from secret)
- Prometheus: http://localhost:9090

### Step 3: Configure Alerts

Alerts are automatically configured via the Prometheus ConfigMap. View active alerts at:
http://localhost:9090/alerts

---

## Phase 6: Performance Testing & Validation (Week 6)

### Step 1: Run Benchmark Tests

```bash
# Make script executable
chmod +x scripts/benchmark.sh

# Run benchmark
./scripts/benchmark.sh

# With custom parameters
ITERATIONS=5000 CONCURRENCY=100 BASE_URL=http://localhost:8080 ./scripts/benchmark.sh
```

### Step 2: Run Load Tests with k6

```bash
# Install k6
# macOS: brew install k6
# Windows: choco install k6
# Linux: See https://k6.io/docs/getting-started/installation/

# Run load test
k6 run scripts/load-test.js

# Run with custom environment
k6 run --env BASE_URL=http://localhost:8080 scripts/load-test.js

# Run stress test
k6 run --config stressOptions scripts/load-test.js

# Run soak test (1 hour)
k6 run --config soakOptions scripts/load-test.js
```

### Step 3: Analyze Results

Expected improvements:

| Metric | Before | After | Target |
|--------|--------|-------|--------|
| URL Creation | 500/s | 2,000/s | 2,000/s |
| URL Redirect | 1,000/s | 10,000/s | 10,000/s |
| Cache Hit Latency | 10ms | 1ms | <2ms |
| P95 Response Time | 500ms | 100ms | <200ms |
| Error Rate | 1% | <0.1% | <0.5% |

---

## Troubleshooting

### Redis Cluster Issues

```bash
# Check cluster status
kubectl exec -it redis-cluster-0 -n urlshortener -- redis-cli cluster info

# Check cluster nodes
kubectl exec -it redis-cluster-0 -n urlshortener -- redis-cli cluster nodes

# Reset cluster (WARNING: deletes all data)
kubectl delete -f k8s/redis-cluster.yaml
kubectl delete pvc -l app=redis-cluster -n urlshortener
kubectl apply -f k8s/redis-cluster.yaml
```

### Database Connection Pool Exhaustion

```bash
# Check active connections
kubectl exec -it postgres-primary-0 -n urlshortener -- \
    psql -U urlshortener -d urlshortener_db \
    -c "SELECT count(*) FROM pg_stat_activity;"

# Check pool settings
kubectl exec -it postgres-primary-0 -n urlshortener -- \
    psql -U urlshortener -d urlshortener_db \
    -c "SHOW max_connections;"
```

### Circuit Breaker Stuck Open

Check the health endpoint to see circuit breaker status:

```bash
curl http://localhost:8080/health | jq '.services'
```

If stuck, the circuit breaker will automatically transition to half-open after the timeout period (default: 30s). You can also restart the pods to reset:

```bash
kubectl rollout restart deployment/urlshortener-app -n urlshortener
```

### High Latency Issues

1. Check cache hit rate:
   ```bash
   curl http://localhost:8080/metrics | grep cache_hit_rate
   ```

2. Check database query performance:
   ```bash
   kubectl exec -it postgres-primary-0 -n urlshortener -- \
       psql -U urlshortener -d urlshortener_db \
       -c "SELECT query, mean_exec_time FROM pg_stat_statements ORDER BY mean_exec_time DESC LIMIT 10;"
   ```

3. Check HPA scaling:
   ```bash
   kubectl describe hpa urlshortener-hpa -n urlshortener
   ```

---

## Performance Tuning Tips

### 1. Database Optimization

- **Analyze tables regularly**: `VACUUM ANALYZE url_mappings;`
- **Monitor index usage**: Check `pg_stat_user_indexes`
- **Adjust connection pool**: Based on workload, increase `DBMaxConns`

### 2. Redis Optimization

- **Monitor memory usage**: Should stay below 80%
- **Check eviction rate**: High evictions = need more memory
- **Tune maxmemory-policy**: `allkeys-lru` for cache-like behavior

### 3. Application Optimization

- **Adjust bulkhead limits**: Based on resource capacity
- **Tune circuit breaker thresholds**: Based on error patterns
- **Enable request compression**: For large payloads

### 4. Kubernetes Optimization

- **Resource limits**: Set realistic limits based on observed usage
- **HPA thresholds**: Lower thresholds for faster scaling
- **PDB**: Ensure minimum availability during rollouts

---

## Rollback Procedures

### Application Rollback

```bash
# View rollout history
kubectl rollout history deployment/urlshortener-app -n urlshortener

# Rollback to previous version
kubectl rollout undo deployment/urlshortener-app -n urlshortener

# Rollback to specific revision
kubectl rollout undo deployment/urlshortener-app --to-revision=2 -n urlshortener
```

### Database Migration Rollback

```bash
# Connect to database
kubectl exec -it postgres-primary-0 -n urlshortener -- \
    psql -U urlshortener -d urlshortener_db

# Run rollback migration
\i /path/to/rollback.sql
```

---

## Maintenance

### Regular Tasks

1. **Weekly**: Review metrics and alerts
2. **Monthly**: Analyze slow queries and optimize indexes
3. **Quarterly**: Review and adjust resource allocations
4. **Annually**: Capacity planning and architecture review

### Database Maintenance

```bash
# Vacuum and analyze (run during low-traffic periods)
kubectl exec -it postgres-primary-0 -n urlshortener -- \
    psql -U urlshortener -d urlshortener_db \
    -c "VACUUM ANALYZE;"

# Reindex (if needed)
kubectl exec -it postgres-primary-0 -n urlshortener -- \
    psql -U urlshortener -d urlshortener_db \
    -c "REINDEX DATABASE urlshortener_db;"
```

### Redis Maintenance

```bash
# Check memory fragmentation
kubectl exec -it redis-cluster-0 -n urlshortener -- \
    redis-cli INFO memory | grep fragmentation

# Trigger defragmentation if > 1.5
kubectl exec -it redis-cluster-0 -n urlshortener -- \
    redis-cli MEMORY PURGE
```

---

## Production Checklist

Before going to production:

- [ ] All secrets are properly configured
- [ ] SSL/TLS certificates are valid
- [ ] Monitoring and alerting are configured
- [ ] Backup strategy is in place
- [ ] Disaster recovery plan is documented
- [ ] Load testing shows acceptable performance
- [ ] Circuit breakers are properly tuned
- [ ] Resource limits are set appropriately
- [ ] Network policies are enforced
- [ ] Logging is configured and centralized
- [ ] On-call rotation is established
- [ ] Runbooks are documented

---

## Support

For issues or questions:
1. Check application logs: `kubectl logs -f deployment/urlshortener-app -n urlshortener`
2. Check metrics in Grafana
3. Review circuit breaker and bulkhead statistics
4. Check GitHub issues
5. Contact the development team

---

## References

- [Redis Cluster Tutorial](https://redis.io/topics/cluster-tutorial)
- [PostgreSQL Performance Tuning](https://wiki.postgresql.org/wiki/Performance_Optimization)
- [Kubernetes Best Practices](https://kubernetes.io/docs/concepts/configuration/overview/)
- [Circuit Breaker Pattern](https://martinfowler.com/bliki/CircuitBreaker.html)
- [Bulkhead Pattern](https://docs.microsoft.com/en-us/azure/architecture/patterns/bulkhead)
