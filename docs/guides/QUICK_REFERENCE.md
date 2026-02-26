# Observability Quick Reference Card

**Print this page and keep it handy during on-call shifts!**

---

## 🚨 Emergency Contacts

| Role | Contact | When to Escalate |
|------|---------|------------------|
| **Primary On-Call** | PagerDuty | All P0/P1 incidents |
| **Secondary On-Call** | PagerDuty | If primary doesn't respond in 5 min |
| **Engineering Lead** | @eng-lead | P0 incidents, architectural decisions |
| **Database Admin** | @dba-team | Database corruption, data loss |
| **DevOps Lead** | @devops | Infrastructure issues, AWS outages |

---

## 🔗 Quick Links

### Dashboards
- [Service Health](http://localhost:3000/d/service-health) - Start here!
- [Grafana Home](http://localhost:3000)
- [Prometheus](http://localhost:9090)
- [Alertmanager](http://localhost:9093)
- [Jaeger Tracing](http://localhost:16686)

### Communication
- Incidents: `#incidents-url-shortener`
- Alerts: `#alerts-url-shortener`
- On-Call Schedule: [PagerDuty](https://pagerduty.com)

---

## 📊 Key Metrics

### Service Health (Check First!)
```promql
# Error Rate (Target: < 0.1%)
(sum(rate(http_requests_total{status=~"5.."}[5m])) / sum(rate(http_requests_total[5m]))) * 100

# P95 Latency (Target: < 1s)
histogram_quantile(0.95, sum(rate(http_request_duration_seconds_bucket[5m])) by (le))

# Request Rate
sum(rate(http_requests_total[5m]))

# SLO Compliance (Target: 99.95%)
sli:availability:ratio_rate5m * 100
```

### Database Health
```bash
# Check connections
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT count(*) FROM pg_stat_activity"

# Check slow queries
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC LIMIT 5"
```

### Cache Health
```bash
# Redis ping
redis-cli -h redis ping

# Redis memory
redis-cli -h redis INFO memory | grep used_memory_human

# Cache hit rate (Target: > 85%)
sum(rate(cache_hits_total[5m])) / (sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m])))
```

---

## 🔥 Common Incidents

### 1. High Error Rate (P0)
```bash
# Check current error rate
curl -s "http://localhost:9090/api/v1/query?query=(sum(rate(http_requests_total{status=~\"5..\"}[5m]))/sum(rate(http_requests_total[5m])))*100" | jq

# View error logs
docker logs url-shortener-app | grep "level=error" | tail -50

# If recent deployment: ROLLBACK
kubectl rollout undo deployment/url-shortener-app
```
**Runbook**: `observability/runbooks/high-error-rate.md`

### 2. High Latency (P1)
```bash
# Check P95 latency
curl -s "http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le))" | jq

# Identify slow endpoints
curl -s "http://localhost:9090/api/v1/query?query=topk(5,histogram_quantile(0.95,sum(rate(http_request_duration_seconds_bucket[5m]))by(le,endpoint)))" | jq

# Check database query latency
curl -s "http://localhost:9090/api/v1/query?query=histogram_quantile(0.95,sum(rate(db_query_duration_seconds_bucket[5m]))by(le))" | jq

# Check cache hit rate
curl -s "http://localhost:9090/api/v1/query?query=sum(rate(cache_hits_total[5m]))/(sum(rate(cache_hits_total[5m]))+sum(rate(cache_misses_total[5m])))" | jq
```
**Runbook**: `observability/runbooks/high-latency.md`

### 3. Database Down (P0)
```bash
# Check database connectivity
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT 1"

# Restart database (if safe)
docker restart postgres

# Check application recovery
curl http://localhost:8080/health
```

### 4. Redis Down (P2)
```bash
# Check Redis status
redis-cli -h redis ping

# Restart Redis
docker restart redis

# Monitor recovery
watch -n 5 'curl http://localhost:9090/api/v1/query?query=up{job="redis"}'
```

### 5. Application Down (P0)
```bash
# Check if app is running
docker ps | grep url-shortener-app

# Check health endpoint
curl http://localhost:8080/health

# View logs
docker logs --tail 100 url-shortener-app

# Restart application
docker restart url-shortener-app
```

---

## 🛠️ Useful Commands

### Application
```bash
# Health check
curl -i http://localhost:8080/health

# Metrics
curl http://localhost:8080/metrics | grep http_requests_total

# Logs (last 100 lines)
docker logs --tail 100 url-shortener-app

# Logs (follow)
docker logs -f url-shortener-app

# Restart
docker restart url-shortener-app
```

### Database
```bash
# Connect
psql -h postgres -U urlshortener -d urlshortener_db

# Active connections
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT count(*) FROM pg_stat_activity"

# Slow queries
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT pid, now() - query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' AND now() - query_start > interval '1 second' ORDER BY duration DESC"

# Kill query
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT pg_terminate_backend(<pid>)"
```

### Cache
```bash
# Ping Redis
redis-cli -h redis ping

# Redis info
redis-cli -h redis INFO

# Memory usage
redis-cli -h redis INFO memory

# Connected clients
redis-cli -h redis INFO clients

# Flush cache (DANGER!)
# redis-cli -h redis FLUSHALL
```

### Prometheus
```bash
# Check targets
curl http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Query metric
curl "http://localhost:9090/api/v1/query?query=up" | jq

# Reload config
curl -X POST http://localhost:9090/-/reload
```

### Grafana
```bash
# Health check
curl http://localhost:3000/api/health

# List dashboards
curl -u admin:${GRAFANA_PASSWORD} http://localhost:3000/api/search

# Data sources
curl -u admin:${GRAFANA_PASSWORD} http://localhost:3000/api/datasources
```

### Alertmanager
```bash
# Check status
curl http://localhost:9093/api/v2/status | jq

# Active alerts
curl http://localhost:9093/api/v2/alerts | jq

# Silence alert (2 hours)
curl -X POST http://localhost:9093/api/v2/silences \
  -H 'Content-Type: application/json' \
  -d '{
    "matchers": [{"name": "alertname", "value": "HighLatency", "isRegex": false}],
    "startsAt": "'$(date -u +%Y-%m-%dT%H:%M:%SZ)'",
    "endsAt": "'$(date -u -d '+2 hours' +%Y-%m-%dT%H:%M:%SZ)'",
    "createdBy": "oncall",
    "comment": "Investigating incident"
  }'
```

---

## 📋 Incident Response Checklist

### Initial Response (< 1 min)
- [ ] Acknowledge alert in PagerDuty
- [ ] Post in `#incidents-url-shortener`
- [ ] Open Service Health dashboard
- [ ] Set Slack status to 🚨 On Incident

### Triage (< 2 min)
- [ ] Assess severity (P0/P1/P2/P3)
- [ ] Check error rate, latency, request rate
- [ ] Identify affected endpoints
- [ ] Check recent deployments

### Investigation (< 5 min)
- [ ] Follow runbook for alert type
- [ ] Check application logs
- [ ] Check database/cache status
- [ ] Review traces in Jaeger
- [ ] Check resource usage (CPU, memory)

### Mitigation (< 10 min for P1)
- [ ] Apply immediate fix
- [ ] Rollback if deployment-related
- [ ] Scale up if resource-constrained
- [ ] Restart services if needed
- [ ] Monitor metrics for improvement

### Resolution
- [ ] Verify metrics returned to normal
- [ ] Monitor for 15 minutes
- [ ] Update incident channel
- [ ] Resolve alert
- [ ] Schedule postmortem

---

## 🗣️ Communication Templates

### Initial Alert
```
🚨 **P1 Incident: [Type]**

Status: Investigating
Impact: [Brief description]
Started: [HH:MM UTC]
Responder: @oncall
Dashboard: [link]
Updates: Every 10 minutes
```

### Status Update
```
⏱️ **Update [HH:MM]**

Status: [Investigating/Mitigating/Resolving]
Root Cause: [What we know]
Action: [What we're doing]
ETA: [Expected resolution]
Next Update: [HH:MM]
```

### Resolution
```
✅ **Resolved [HH:MM]**

Duration: [X minutes]
Root Cause: [Brief description]
Resolution: [What fixed it]
SLO Impact: [Error budget consumed]
Postmortem: [Date/time]
```

---

## 📖 Runbooks Available

| Incident Type | Runbook | Severity |
|---------------|---------|----------|
| High Latency | `high-latency.md` | P1 |
| High Error Rate | `high-error-rate.md` | P0/P1 |
| Database Down | TBD | P0 |
| Redis Down | TBD | P2 |
| Application Down | TBD | P0 |
| Error Budget Burn | TBD | P1 |

**Location**: `observability/runbooks/`

---

## 🎯 SLO Targets

| Metric | Target | Critical | Warning |
|--------|--------|----------|---------|
| Availability | 99.95% | < 99.9% | < 99.95% |
| Latency (P95) | < 1s | > 2s | > 1s |
| Latency (P99) | < 2s | > 5s | > 2s |
| Error Rate | < 0.1% | > 1% | > 0.5% |
| Cache Hit Rate | > 85% | < 70% | < 85% |

---

## ⚡ Emergency Procedures

### Rollback Deployment
```bash
# Kubernetes
kubectl rollout undo deployment/url-shortener-app

# Docker Compose
docker-compose -f docker/docker-compose.observability.yml down app
docker-compose -f docker/docker-compose.observability.yml up -d app
```

### Scale Application
```bash
# Kubernetes
kubectl scale deployment url-shortener-app --replicas=5

# Docker Compose (restart with more resources)
docker-compose -f docker/docker-compose.observability.yml stop app
docker-compose -f docker/docker-compose.observability.yml up -d app
```

### Restart All Services
```bash
docker-compose -f docker/docker-compose.observability.yml restart
```

### Nuclear Option (Full Stack Restart)
```bash
docker-compose -f docker/docker-compose.observability.yml down
docker-compose -f docker/docker-compose.observability.yml up -d
```

---

## 🔍 Troubleshooting Tips

1. **Start with Service Health dashboard** - High-level overview
2. **Check recent changes** - Deployments, config changes
3. **Follow the runbook** - Don't skip steps
4. **Use traces** - Jaeger for performance issues
5. **Check logs** - Loki for error details
6. **Ask for help** - Escalate if stuck > 15 min

---

## 📞 When to Escalate

| Situation | Escalate To | When |
|-----------|-------------|------|
| Can't resolve in 15 min | Senior Engineer | Immediately |
| P0 incident | Engineering Lead | Immediately |
| Data loss risk | Database Admin | Immediately |
| Infrastructure issue | DevOps Lead | If can't resolve |
| Customer-facing outage | Product Manager | For communication |

---

## 🧠 Remember

- **Stay Calm**: Panic doesn't help
- **Follow Runbooks**: They're tested procedures
- **Communicate**: Update incident channel regularly
- **Document**: Take notes for postmortem
- **Ask for Help**: Don't be a hero

---

**Print Date**: 2025-01-07
**Version**: 1.0
**Keep this card updated after each incident!**
