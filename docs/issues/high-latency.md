# Runbook: High Latency

**Alert**: `HighP95Latency` or `HighP99Latency`
**Severity**: P1 (Critical) if P99 > 2s, P2 (Warning) if P95 > 1s
**MTTR Target**: 15 minutes

## Symptoms
- P95 latency > 1 second for 5+ minutes
- P99 latency > 2 seconds for 5+ minutes
- Users reporting slow page loads
- Timeouts in client applications

## Impact
- Poor user experience
- Potential SLO violation
- Error budget consumption
- Risk of cascading failures

## Investigation Steps

### 1. Verify the Alert (< 1 min)
```bash
# Check current latency percentiles
curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,endpoint))" | jq

# View in Grafana
open http://grafana:3000/d/app-performance?viewPanel=2
```

**Expected**: Latency should be < 1s for P95, < 2s for P99
**If confirmed**: Proceed to step 2

### 2. Identify Affected Endpoints (< 2 min)
```bash
# Find slowest endpoints
curl -s "http://prometheus:9090/api/v1/query?query=topk(5,histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,endpoint)))" | jq

# Check in Grafana dashboard
open http://grafana:3000/d/app-performance?viewPanel=5
```

**Questions**:
- Is latency affecting all endpoints or specific ones?
- When did the latency spike start?
- Were there any deployments around that time?

### 3. Check System Resources (< 2 min)
```bash
# CPU usage
curl -s "http://prometheus:9090/api/v1/query?query=100-(avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m]))*100)" | jq

# Memory usage
curl -s "http://prometheus:9090/api/v1/query?query=(1-(node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes))*100" | jq

# Check Grafana infrastructure dashboard
open http://grafana:3000/d/infrastructure
```

**Red Flags**:
- CPU > 80%: Resource exhaustion
- Memory > 85%: Memory pressure
- Disk I/O high: Storage bottleneck

### 4. Check Database Performance (< 3 min)
```bash
# Database query latency
curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le,operation))" | jq

# Check for slow queries
psql -h postgres -U urlshortener -d urlshortener_db -c "
SELECT pid, now() - query_start AS duration, query, state
FROM pg_stat_activity
WHERE state = 'active' AND now() - query_start > interval '1 second'
ORDER BY duration DESC
LIMIT 10;"

# Check database connections
curl -s "http://prometheus:9090/api/v1/query?query=db_connections_active" | jq

# View database dashboard
open http://grafana:3000/d/db-performance
```

**Red Flags**:
- Query latency > 500ms
- Slow queries running
- Connection pool > 80% utilized
- Long-running transactions

### 5. Check Cache Performance (< 2 min)
```bash
# Cache hit ratio
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(cache_hits_total[5m]))/(sum(rate(cache_hits_total[5m]))+sum(rate(cache_misses_total[5m])))" | jq

# Redis status
redis-cli -h redis INFO stats

# Cache operation latency
curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(cache_operation_duration_seconds_bucket[5m]))by(le))" | jq
```

**Red Flags**:
- Cache hit ratio < 85%
- Redis memory usage > 85%
- High cache operation latency

### 6. Check for External Dependencies (< 2 min)
```bash
# Check distributed traces for slow external calls
open http://jaeger:16686

# Look for:
# - Slow database queries
# - Slow external API calls
# - Slow cache operations
# - High GC pauses
```

### 7. Review Recent Changes (< 2 min)
```bash
# Check recent deployments
kubectl rollout history deployment/url-shortener-app -n url-shortener

# Check recent git commits
git log --oneline --since="2 hours ago"

# Review recent configuration changes
kubectl get configmap url-shortener-config -o yaml -n url-shortener
```

## Common Root Causes & Solutions

### Root Cause 1: Database Slow Queries
**Symptoms**: High database query latency, queries > 1s

**Immediate Mitigation**:
```bash
# Identify and kill slow queries
psql -h postgres -U urlshortener -d urlshortener_db -c "
SELECT pg_terminate_backend(pid)
FROM pg_stat_activity
WHERE state = 'active'
  AND now() - query_start > interval '30 seconds'
  AND query NOT LIKE '%pg_stat_activity%';"

# Increase connection pool temporarily
kubectl set env deployment/url-shortener-app DB_MAX_CONNECTIONS=50 -n url-shortener
```

**Long-term Fix**:
- Add missing indexes
- Optimize query execution plans
- Implement query result caching
- Consider read replicas for read-heavy queries

### Root Cause 2: Cache Miss Storm
**Symptoms**: Low cache hit rate (< 70%), high database load

**Immediate Mitigation**:
```bash
# Check if Redis is down
redis-cli -h redis ping

# If Redis is down, restart it
kubectl rollout restart deployment/redis -n url-shortener

# Pre-warm cache with popular URLs
curl -X POST http://app:8080/admin/cache/warm
```

**Long-term Fix**:
- Implement cache warming on startup
- Increase cache TTL for stable data
- Add cache fallback logic
- Implement cache stampede prevention

### Root Cause 3: High CPU Usage
**Symptoms**: CPU > 80%, increased response times

**Immediate Mitigation**:
```bash
# Scale up application instances
kubectl scale deployment url-shortener-app --replicas=5 -n url-shortener

# Or increase CPU limits
kubectl set resources deployment url-shortener-app --limits=cpu=2000m -n url-shortener

# Check for CPU-intensive operations
docker stats url-shortener-app
```

**Long-term Fix**:
- Profile application for CPU hotspots
- Optimize expensive operations
- Implement rate limiting
- Add horizontal auto-scaling

### Root Cause 4: Memory Pressure / GC Pauses
**Symptoms**: Memory > 85%, high GC duration

**Immediate Mitigation**:
```bash
# Increase memory limits
kubectl set resources deployment url-shortener-app --limits=memory=2Gi -n url-shortener

# Restart application to clear memory
kubectl rollout restart deployment/url-shortener-app -n url-shortener
```

**Long-term Fix**:
- Profile memory usage
- Fix memory leaks
- Optimize data structures
- Tune garbage collector

### Root Cause 5: External Service Degradation
**Symptoms**: Traces show slow external calls

**Immediate Mitigation**:
```bash
# Implement circuit breaker
# Enable fallback logic
# Increase timeout settings temporarily

# If using a degraded third-party service:
# - Switch to backup provider
# - Disable non-critical features
# - Return cached/stale data
```

**Long-term Fix**:
- Implement circuit breakers
- Add retry logic with exponential backoff
- Implement fallback mechanisms
- Set up monitoring for external dependencies

### Root Cause 6: Network Issues
**Symptoms**: High network latency, packet loss

**Immediate Mitigation**:
```bash
# Check network metrics
curl -s "http://prometheus:9090/api/v1/query?query=rate(node_network_transmit_drop_total[5m])" | jq

# Check for DNS issues
nslookup postgres
nslookup redis

# Restart networking (if using Kubernetes)
kubectl rollout restart deployment/coredns -n kube-system
```

**Long-term Fix**:
- Investigate network infrastructure
- Optimize service mesh configuration
- Consider service locality

### Root Cause 7: Recent Deployment
**Symptoms**: Latency increased after recent deployment

**Immediate Mitigation**:
```bash
# Rollback to previous version
kubectl rollout undo deployment/url-shortener-app -n url-shortener

# Monitor metrics for improvement
watch -n 5 'curl -s "http://prometheus:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))" | jq'
```

**Long-term Fix**:
- Review code changes
- Add performance tests to CI/CD
- Implement canary deployments
- Add synthetic monitoring

## Mitigation Checklist

- [ ] Alert acknowledged in PagerDuty
- [ ] Incident channel created: `#incident-YYYYMMDD-latency`
- [ ] Severity assessed and communicated
- [ ] Root cause identified
- [ ] Mitigation applied
- [ ] Metrics improving (monitor for 15 minutes)
- [ ] Alert resolved
- [ ] Postmortem scheduled
- [ ] Runbook updated with learnings

## Communication Templates

### Initial Alert (Post in #incidents-url-shortener)
```
🚨 **P1 Incident: High Latency**

**Status**: Investigating
**Impact**: P95 latency at Xs (target: <1s)
**Affected**: All users
**Responder**: @oncall
**Dashboard**: http://grafana:3000/d/app-performance
**Updates**: Every 10 minutes or on significant change
```

### Status Update
```
⏱️ **Update at HH:MM**

**Status**: Mitigating
**Root Cause**: [Brief description]
**Action**: [What we're doing]
**ETA**: [Expected resolution time]
**Next Update**: HH:MM
```

### Resolution
```
✅ **Incident Resolved at HH:MM**

**Duration**: X minutes
**Root Cause**: [Brief description]
**Resolution**: [What fixed it]
**SLO Impact**: X% error budget consumed
**Postmortem**: Will be completed within 48 hours
```

## Prevention

1. **Implement Performance Testing**
   - Add load tests to CI/CD pipeline
   - Set performance budgets
   - Test with production-like data volumes

2. **Optimize Database Queries**
   - Review slow query logs weekly
   - Add missing indexes
   - Implement query result caching

3. **Enhance Caching Strategy**
   - Increase cache TTL for stable data
   - Implement cache warming
   - Add cache stampede prevention

4. **Set Up Synthetic Monitoring**
   - Create synthetic user journeys
   - Monitor from multiple regions
   - Alert on performance degradation

5. **Implement Auto-Scaling**
   - Configure HPA based on latency metrics
   - Set appropriate scaling thresholds
   - Test scaling behavior under load

## Related Runbooks
- [High Error Rate](high-error-rate.md)
- [Database Down](database-down.md)
- [High CPU Usage](high-cpu.md)
- [Error Budget Burn](error-budget-burn.md)

## Metrics to Monitor
- `http_request_duration_seconds_bucket` (P50, P95, P99)
- `db_query_duration_seconds_bucket`
- `cache_operation_duration_seconds_bucket`
- `cache_hit_ratio`
- `db_connections_active`
- `node_cpu_seconds_total`
- `node_memory_MemAvailable_bytes`

## Useful Queries

### Prometheus Queries
```promql
# P95 latency by endpoint
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, endpoint))

# Slowest endpoints (top 5)
topk(5, histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le, endpoint)))

# Database query latency P95
histogram_quantile(0.95, sum(rate(db_query_duration_seconds_bucket[5m])) by (le, operation))

# Cache hit ratio
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))
```

### Loki Queries
```logql
# Find slow requests in logs
{service="url-shortener"} |= "latency_ms" | json | latency_ms > 1000

# Find database slow queries
{service="url-shortener"} |= "Database query" | json | duration > 1s

# Find errors contributing to latency
{service="url-shortener", level="error"} | json
```

---

**Last Updated**: 2025-01-07
**Next Review**: 2025-04-07
**Runbook Version**: 1.0
