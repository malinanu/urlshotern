# Incident Response Runbooks - URL Shortener Service

## Quick Start Guide

**MTTD Target**: < 2 minutes (Mean Time To Detect)
**MTTR Target**: < 15 minutes for P1 incidents (Mean Time To Resolve)

### Incident Severity Levels

| Severity | Description | Response Time | Examples |
|----------|-------------|---------------|----------|
| **P0 - Critical** | Complete service outage | Immediate (< 5 min) | Application down, database unavailable |
| **P1 - High** | Major functionality degraded | < 15 min | High error rate, severe latency |
| **P2 - Medium** | Partial functionality degraded | < 1 hour | Cache down, slow queries |
| **P3 - Low** | Minor issues, workarounds available | < 4 hours | Low cache hit rate, elevated warnings |

---

## Available Runbooks

### Application Issues
- [High Latency](#runbook-high-latency)
- [High Error Rate](#runbook-high-error-rate)
- [Application Down](#runbook-application-down)
- [Zero Traffic / No URL Creations](#runbook-zero-traffic)
- [Error Budget Burn](#runbook-error-budget-burn)

### Database Issues
- [Database Down](#runbook-database-down)
- [Connection Pool Exhausted](#runbook-connection-pool-exhausted)
- [Slow Queries](#runbook-slow-queries)
- [High Database Connections](#runbook-high-db-connections)

### Cache Issues
- [Redis Down](#runbook-redis-down)
- [Low Cache Hit Rate](#runbook-low-cache-hit-rate)
- [Redis Memory High](#runbook-redis-memory-high)
- [High Eviction Rate](#runbook-high-eviction-rate)

### Infrastructure Issues
- [High CPU Usage](#runbook-high-cpu)
- [High Memory Usage](#runbook-high-memory)
- [Low Disk Space](#runbook-low-disk-space)

---

## General Incident Response Process

### 1. **Acknowledge Alert** (< 1 min)
- Acknowledge in PagerDuty/Alertmanager
- Post in incident channel: `#incidents-url-shortener`
- Set Slack status: 🚨 Incident Response

### 2. **Assess Severity** (< 2 min)
- Check Service Health dashboard
- Determine impact (% of users affected, revenue impact)
- Assign severity level
- Notify stakeholders if P0/P1

### 3. **Investigate** (< 5 min)
- Follow specific runbook for alert type
- Check recent deployments/changes
- Review error logs in Loki
- Examine distributed traces in Jaeger
- Check related metrics in Grafana

### 4. **Mitigate** (< 10 min)
- Apply immediate mitigation (rollback, scale up, failover)
- Communicate actions in incident channel
- Monitor metrics for improvement

### 5. **Resolve**
- Verify all metrics returned to normal
- Monitor for 15 minutes to ensure stability
- Update incident status
- Resolve alert

### 6. **Post-Incident** (Within 48 hours)
- Write blameless postmortem
- Identify root cause
- Create action items for prevention
- Update runbooks with learnings

---

## Quick Links

### Dashboards
- [Service Health Overview](http://grafana:3000/d/service-health)
- [Application Performance](http://grafana:3000/d/app-performance)
- [Database Performance](http://grafana:3000/d/db-performance)
- [SLO Tracking](http://grafana:3000/d/slo-tracking)

### Tools
- [Grafana](http://grafana:3000)
- [Prometheus](http://prometheus:9090)
- [Alertmanager](http://alertmanager:9093)
- [Jaeger](http://jaeger:16686)
- [Loki](http://loki:3100)

### Communication
- Incident Channel: `#incidents-url-shortener`
- On-Call Schedule: [PagerDuty Schedule](https://pagerduty.com/schedules/url-shortener)
- Status Page: [status.example.com](https://status.example.com)

### Documentation
- [Architecture Diagram](../docs/architecture.md)
- [Deployment Guide](../docs/deployment.md)
- [Database Schema](../configs/database.sql)

---

## Escalation Paths

### Level 1: On-Call Engineer
- Initial response and investigation
- Execute standard runbooks
- Escalate if unable to resolve in 15 minutes

### Level 2: Senior Engineer / Team Lead
- Complex issues requiring deep system knowledge
- Architectural decisions
- Coordination with other teams

### Level 3: Engineering Manager
- P0 incidents requiring executive communication
- Cross-team coordination
- Resource allocation decisions

### Level 4: VP Engineering / CTO
- Company-wide impact
- External communication
- Strategic decisions

---

## Useful Commands

### Check Application Status
```bash
# Check if app is responding
curl -i http://app:8080/health

# Check metrics endpoint
curl http://app:8080/metrics | grep http_requests_total

# View recent logs
docker logs --tail 100 url-shortener-app
```

### Database Checks
```bash
# Check database connectivity
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT 1"

# Check active connections
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT count(*) FROM pg_stat_activity"

# Check for slow queries
psql -h postgres -U urlshortener -d urlshortener_db -c "SELECT pid, now() - pg_stat_activity.query_start AS duration, query FROM pg_stat_activity WHERE state = 'active' ORDER BY duration DESC"
```

### Cache Checks
```bash
# Check Redis connectivity
redis-cli -h redis ping

# Check Redis memory
redis-cli -h redis INFO memory

# Check cache hit rate
redis-cli -h redis INFO stats | grep keyspace
```

### Kubernetes Commands (if applicable)
```bash
# Check pod status
kubectl get pods -n url-shortener

# View pod logs
kubectl logs -f deployment/url-shortener-app -n url-shortener

# Describe pod issues
kubectl describe pod <pod-name> -n url-shortener

# Scale deployment
kubectl scale deployment url-shortener-app --replicas=5 -n url-shortener

# Rollback deployment
kubectl rollout undo deployment/url-shortener-app -n url-shortener
```

---

## Contact Information

### On-Call Rotation
- Primary: [PagerDuty Schedule](https://pagerduty.com)
- Secondary: [PagerDuty Schedule](https://pagerduty.com)

### Key Personnel
- Engineering Lead: @engineering-lead
- Database Admin: @dba-team
- DevOps Lead: @devops-lead
- Product Manager: @product-manager

### External Contacts
- AWS Support: +1-XXX-XXX-XXXX (Premium Support)
- CDN Provider: support@cdn.com
- Database Vendor: support@postgres.com

---

## Postmortem Template

After incident resolution, use this template:

```markdown
# Incident Postmortem: [TITLE]

**Incident ID**: INC-YYYY-MM-DD-XXX
**Date**: YYYY-MM-DD
**Severity**: P0/P1/P2/P3
**Duration**: X hours Y minutes
**Impact**: % of users affected, revenue impact

## Summary
Brief description of the incident.

## Timeline
- HH:MM - Alert triggered
- HH:MM - Engineer acknowledged
- HH:MM - Root cause identified
- HH:MM - Mitigation applied
- HH:MM - Incident resolved

## Root Cause
Detailed explanation of what caused the incident.

## Impact
- Users affected:
- Revenue impact:
- SLO impact:
- Error budget consumed:

## Resolution
What was done to resolve the incident.

## Prevention
Action items to prevent recurrence:
1. [Action item with owner and deadline]
2. [Action item with owner and deadline]

## Lessons Learned
- What went well:
- What could be improved:
- Runbook updates needed:

## Follow-up Items
- [ ] Item 1 - Owner: @name - Due: YYYY-MM-DD
- [ ] Item 2 - Owner: @name - Due: YYYY-MM-DD
```

---

## Runbook Maintenance

Runbooks should be:
- **Tested**: Quarterly disaster recovery drills
- **Updated**: After each incident
- **Reviewed**: Quarterly by the team
- **Versioned**: All changes tracked in git

Last Updated: 2025-01-07
Next Review Date: 2025-04-07
