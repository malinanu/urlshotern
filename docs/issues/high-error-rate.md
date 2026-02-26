# Runbook: High Error Rate

**Alert**: `HighErrorRate` or `ElevatedErrorRate`
**Severity**: P0 (Critical) if error rate > 1%, P1 (Warning) if > 0.5%
**MTTR Target**: 10 minutes

## Symptoms
- Error rate > 1% for 5+ minutes (Critical)
- Error rate > 0.5% for 10+ minutes (Warning)
- Increased 5xx responses
- Users reporting service failures
- Error budget burning rapidly

## Impact
- Service unavailable for affected users
- Data loss or corruption risk
- SLO violation (Availability target: 99.95%)
- Severe error budget consumption
- Revenue impact

## Investigation Steps

### 1. Verify the Alert (< 1 min)
```bash
# Check current error rate
curl -s "http://prometheus:9090/api/v1/query?query=(sum(rate(http_requests_total{status=~\"5..\"}[5m]))/sum(rate(http_requests_total[5m])))*100" | jq

# View in Grafana
open http://grafana:3000/d/service-health
```

**Expected**: Error rate should be < 0.1%
**If error rate > 1%**: This is a P0 incident - escalate immediately

### 2. Identify Error Types and Patterns (< 2 min)
```bash
# Error rate by status code
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total{status=~\"5..\"}[5m]))by(status)" | jq

# Error rate by endpoint
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total{status=~\"5..\"}[5m]))by(endpoint)" | jq

# View error logs
docker logs --tail 200 url-shortener-app | grep "level=error"

# Or query Loki
curl -G -s "http://loki:3100/loki/api/v1/query_range" \
  --data-urlencode 'query={service="url-shortener", level="error"}' \
  --data-urlencode 'limit=100' | jq
```

**Key Questions**:
- Which endpoints are affected?
- What error types (500, 502, 503, 504)?
- Are errors consistent or intermittent?
- When did errors start?

### 3. Check Application Health (< 1 min)
```bash
# Health check endpoint
curl -i http://app:8080/health

# Check if application is running
docker ps | grep url-shortener-app

# Or Kubernetes
kubectl get pods -n url-shortener -l app=url-shortener
```

**Red Flags**:
- Health check returning non-200
- Application container restarting
- Pods in CrashLoopBackOff

### 4. Check Dependencies (< 2 min)
```bash
# Database connectivity
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT 1"

# Redis connectivity
redis-cli -h redis ping

# Check dependency metrics
curl -s "http://prometheus:9090/api/v1/query?query=up{job=~\"postgresql|redis\"}" | jq
```

**Red Flags**:
- Database unreachable
- Redis unreachable
- External services down

### 5. Review Recent Changes (< 2 min)
```bash
# Recent deployments
kubectl rollout history deployment/url-shortener-app -n url-shortener

# Recent git commits
git log --oneline --since="1 hour ago"

# Recent configuration changes
kubectl diff -f k8s/deployment.yaml -n url-shortener
```

**If deployment within last hour**: Strong indicator of deployment-related issue

### 6. Examine Error Logs (< 3 min)
```bash
# Application logs
docker logs --tail 500 --since 10m url-shortener-app | grep -i "error"

# Database logs
docker logs --tail 200 postgres | grep -i "error\|fatal"

# System logs
journalctl -u docker -n 100 --since "10 minutes ago"

# Query structured logs in Loki
{service="url-shortener", level="error"} | json | line_format "{{.message}}"
```

**Look for**:
- Panic/crash messages
- Connection errors
- Timeout errors
- Out of memory errors
- Null pointer exceptions

### 7. Check Resource Exhaustion (< 2 min)
```bash
# CPU, Memory, Disk
curl -s "http://prometheus:9090/api/v1/query?query=100-(avg(irate(node_cpu_seconds_total{mode=\"idle\"}[5m]))*100)" | jq
curl -s "http://prometheus:9090/api/v1/query?query=(1-(node_memory_MemAvailable_bytes/node_memory_MemTotal_bytes))*100" | jq

# Open file descriptors
curl -s "http://prometheus:9090/api/v1/query?query=process_open_fds" | jq

# Goroutines (check for leaks)
curl -s "http://prometheus:9090/api/v1/query?query=goroutines_active" | jq
```

## Common Root Causes & Solutions

### Root Cause 1: Database Connection Failure
**Symptoms**: 500 errors, "connection refused" in logs

**Immediate Mitigation**:
```bash
# Check database status
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT version()"

# If database is down, restart
docker restart postgres
# Or
kubectl rollout restart statefulset/postgres -n url-shortener

# Check connection pool
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT count(*) FROM pg_stat_activity"

# If max connections reached, increase limit temporarily
psql -h postgres -U postgres -c "ALTER SYSTEM SET max_connections = 200"
psql -h postgres -U postgres -c "SELECT pg_reload_conf()"
```

**Long-term Fix**:
- Implement connection pooling properly
- Set appropriate connection limits
- Implement circuit breakers
- Add database high availability (read replicas, failover)

### Root Cause 2: Redis Cache Failure
**Symptoms**: 500 errors, high database load, slow responses

**Immediate Mitigation**:
```bash
# Check Redis status
redis-cli -h redis INFO server

# If Redis is down, restart
docker restart redis
# Or
kubectl rollout restart deployment/redis -n url-shortener

# Application should handle Redis failures gracefully
# If not, consider temporary failover:
kubectl set env deployment/url-shortener-app CACHE_ENABLED=false -n url-shortener
```

**Long-term Fix**:
- Implement Redis high availability (Sentinel/Cluster)
- Handle cache failures gracefully
- Implement fallback to database
- Add cache circuit breaker

### Root Cause 3: Recent Deployment Introduced Bugs
**Symptoms**: Error rate spiked immediately after deployment

**Immediate Mitigation**:
```bash
# ROLLBACK IMMEDIATELY
kubectl rollout undo deployment/url-shortener-app -n url-shortener

# Verify errors decreasing
watch -n 5 'curl -s "http://prometheus:9090/api/v1/query?query=(sum(rate(http_requests_total{status=~\"5..\"}[5m]))/sum(rate(http_requests_total[5m])))*100" | jq'

# Monitor for 5 minutes to ensure stability
```

**Long-term Fix**:
- Review code changes
- Add missing error handling
- Improve test coverage
- Implement canary deployments
- Add synthetic monitoring pre-production

### Root Cause 4: Memory/Resource Exhaustion
**Symptoms**: Out of memory errors, slow performance before errors

**Immediate Mitigation**:
```bash
# Restart application to clear memory
kubectl rollout restart deployment/url-shortener-app -n url-shortener

# Scale up resources temporarily
kubectl set resources deployment url-shortener-app \
  --limits=cpu=2000m,memory=2Gi \
  --requests=cpu=1000m,memory=1Gi \
  -n url-shortener

# Scale horizontally
kubectl scale deployment url-shortener-app --replicas=5 -n url-shortener
```

**Long-term Fix**:
- Profile application for memory leaks
- Optimize memory usage
- Implement resource limits properly
- Add memory monitoring and alerts

### Root Cause 5: External Service Degradation
**Symptoms**: Timeouts, 503 errors, traces show slow external calls

**Immediate Mitigation**:
```bash
# Identify failing external service from traces
open http://jaeger:16686

# If non-critical service, disable feature flag
kubectl set env deployment/url-shortener-app FEATURE_EXTERNAL_SERVICE=false -n url-shortener

# If critical service, implement circuit breaker
# (This should be in code, but can be done via config)
kubectl set env deployment/url-shortener-app CIRCUIT_BREAKER_ENABLED=true -n url-shortener
```

**Long-term Fix**:
- Implement circuit breakers for all external calls
- Add retry logic with exponential backoff
- Implement fallback mechanisms
- Add external service monitoring
- Set appropriate timeouts

### Root Cause 6: Traffic Spike / DDoS
**Symptoms**: Sudden increase in traffic, resource exhaustion

**Immediate Mitigation**:
```bash
# Check traffic rate
curl -s "http://prometheus:9090/api/v1/query?query=sum(rate(http_requests_total[5m]))" | jq

# Enable rate limiting
kubectl apply -f k8s/rate-limit-config.yaml -n url-shortener

# Scale up to handle traffic
kubectl scale deployment url-shortener-app --replicas=10 -n url-shortener

# If DDoS suspected, block malicious IPs
# (Add to nginx config or WAF)
```

**Long-term Fix**:
- Implement rate limiting
- Add CDN/WAF protection
- Implement auto-scaling
- Add DDoS protection (Cloudflare, AWS Shield)

### Root Cause 7: Database Lock Contention
**Symptoms**: Timeouts, slow queries, high active connections

**Immediate Mitigation**:
```bash
# Check for locks
psql -h postgres -U urlshortener -d urlshortener_db -c "
SELECT blocked_locks.pid AS blocked_pid,
       blocked_activity.usename AS blocked_user,
       blocking_locks.pid AS blocking_pid,
       blocking_activity.usename AS blocking_user,
       blocked_activity.query AS blocked_statement,
       blocking_activity.query AS blocking_statement
FROM pg_catalog.pg_locks blocked_locks
JOIN pg_catalog.pg_stat_activity blocked_activity ON blocked_activity.pid = blocked_locks.pid
JOIN pg_catalog.pg_locks blocking_locks
    ON blocking_locks.locktype = blocked_locks.locktype
    AND blocking_locks.database IS NOT DISTINCT FROM blocked_locks.database
    AND blocking_locks.relation IS NOT DISTINCT FROM blocked_locks.relation
    AND blocking_locks.pid != blocked_locks.pid
JOIN pg_catalog.pg_stat_activity blocking_activity ON blocking_activity.pid = blocking_locks.pid
WHERE NOT blocked_locks.granted;"

# Kill blocking queries if safe
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT pg_terminate_backend(<blocking_pid>)"
```

**Long-term Fix**:
- Optimize queries to reduce lock time
- Use appropriate isolation levels
- Implement optimistic locking where possible
- Consider partitioning large tables

## Mitigation Checklist

- [ ] Alert acknowledged (< 1 min)
- [ ] Incident channel created: `#incident-YYYYMMDD-errors`
- [ ] Severity assessed (P0 if > 1%)
- [ ] Stakeholders notified (if P0)
- [ ] Status page updated (if customer-facing)
- [ ] Root cause identified (< 5 min)
- [ ] Mitigation applied (< 10 min)
- [ ] Error rate decreasing
- [ ] Monitor for 15 minutes
- [ ] Alert resolved
- [ ] Postmortem scheduled (within 24h)
- [ ] Runbook updated

## Communication Templates

### P0 Alert (Post immediately in #incidents and page leadership)
```
🚨 **P0 INCIDENT: HIGH ERROR RATE**

**Status**: INVESTIGATING
**Error Rate**: X% (CRITICAL - Target: <0.1%)
**Impact**: SERVICE DEGRADED - X% of requests failing
**Started**: HH:MM UTC
**Responder**: @oncall
**Dashboard**: http://grafana:3000/d/service-health
**War Room**: Zoom link
**Updates**: Every 5 minutes
```

### Status Update (Every 5 min for P0, 10 min for P1)
```
⏱️ **Update HH:MM**

**Status**: [Investigating/Mitigating/Resolving]
**Error Rate**: X% (was Y%)
**Root Cause**: [What we know]
**Action**: [What we're doing now]
**ETA**: [Expected resolution]
**Next Update**: HH:MM
```

### Resolution
```
✅ **INCIDENT RESOLVED - HH:MM UTC**

**Duration**: X minutes
**Peak Error Rate**: Y%
**Root Cause**: [Brief description]
**Resolution**: [What fixed it]
**SLO Impact**: Z% error budget consumed
**User Impact**: Approximately X users affected
**Postmortem**: Scheduled for [date/time]
**Thank you**: @team for rapid response
```

## Prevention

1. **Improve Deployment Safety**
   - Implement canary deployments (10% -> 50% -> 100%)
   - Add deployment smoke tests
   - Implement automatic rollback on error rate increase
   - Require code review for critical paths

2. **Enhance Resilience**
   - Implement circuit breakers for all external dependencies
   - Add retry logic with exponential backoff
   - Implement fallback mechanisms
   - Add graceful degradation for non-critical features

3. **Improve Monitoring**
   - Add synthetic monitoring
   - Implement real-user monitoring (RUM)
   - Add error categorization and tracking
   - Set up error budget alerts

4. **Testing**
   - Add integration tests for error scenarios
   - Implement chaos engineering (failure injection)
   - Load test before major releases
   - Test database failover scenarios

5. **Dependencies**
   - Implement high availability for PostgreSQL
   - Implement Redis Sentinel/Cluster
   - Add health checks for all dependencies
   - Document dependency SLAs

## Related Runbooks
- [High Latency](high-latency.md)
- [Database Down](database-down.md)
- [Application Down](application-down.md)
- [Error Budget Burn](error-budget-burn.md)

## Metrics to Monitor During Incident
- `http_requests_total{status=~"5.."}` (Error count)
- Error rate: `(errors / total_requests) * 100`
- `sli_availability_ratio_rate5m` (SLI compliance)
- `error_budget_remaining` (Error budget)
- `db_queries_total{status="error"}` (Database errors)
- `goroutines_active` (Resource leaks)

## Useful Queries

### Prometheus
```promql
# Current error rate percentage
(sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) * 100

# Error rate by endpoint
sum(rate(http_requests_total{status=~"5.."}[5m])) by (endpoint)

# Error rate by status code
sum(rate(http_requests_total{status=~"5.."}[5m])) by (status)

# SLO violation (availability < 99.95%)
sli:availability:ratio_rate5m < 0.9995
```

### Loki
```logql
# All error logs in last 10 minutes
{service="url-shortener", level="error"} | json

# Errors with stack traces
{service="url-shortener", level="error"} |= "stack" | json

# Database connection errors
{service="url-shortener"} |= "database" |= "error" | json

# Specific error patterns
{service="url-shortener"} |~ "(?i)(panic|fatal|crash)"
```

---

**Last Updated**: 2025-01-07
**Next Review**: 2025-04-07
**Runbook Version**: 1.0
