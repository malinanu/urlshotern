# Performance Service Level Agreement (SLA)

## Document Information

**Version:** 1.0
**Effective Date:** [YYYY-MM-DD]
**Review Schedule:** Quarterly
**Owner:** Engineering Team
**Stakeholders:** Product, Engineering, Operations

---

## Overview

This document defines the performance Service Level Agreements (SLAs) for the URL Shortener application. These SLAs establish clear performance expectations, measurement methodologies, and accountability for maintaining system performance.

---

## Performance Objectives

### Availability

**Target:** 99.9% uptime (8.76 hours of downtime per year maximum)

**Measurement:**
- Calculated as: (Total Time - Downtime) / Total Time × 100%
- Measured monthly and annually
- Planned maintenance excluded from calculation
- Measured using health check endpoint monitoring

**Accountability:**
- **Owner:** SRE Team
- **Review:** Monthly SLA review meeting
- **Escalation:** Automatic notification if availability drops below 99.5%

---

## Performance SLAs by Service

### 1. URL Redirect Service

**Priority:** P0 - Critical (Revenue-impacting)

#### Latency SLA

| Metric | Target | Measurement Window | Tolerance |
|--------|--------|-------------------|-----------|
| p50 (Median) | < 20ms | 5-minute rolling window | +10% |
| p95 | < 50ms | 5-minute rolling window | +10% |
| p99 | < 100ms | 5-minute rolling window | +20% |
| p99.9 | < 200ms | 5-minute rolling window | +30% |

**Success Criteria:**
- Must meet all percentile targets 99% of the time within a calendar month
- No single 5-minute window should exceed 2x the target
- Zero instances of 3+ consecutive violations

**Measurement Method:**
```promql
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket{endpoint="redirect"}[5m])) by (le)
)
```

**Violations:**
- **Minor:** Exceed target by 10-25% for < 15 minutes
- **Major:** Exceed target by 25-50% or < 15 minutes violation
- **Critical:** Exceed target by >50% or >15 minutes violation

#### Throughput SLA

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Peak Capacity | 10,000 req/s | Sustained for 1 hour |
| Burst Capacity | 15,000 req/s | Sustained for 5 minutes |
| Average Load | 2,000 req/s | Daily average |

**Success Criteria:**
- System must handle peak capacity with latency SLA compliance
- Burst capacity with degraded latency (p95 < 100ms) acceptable
- Auto-scaling must activate before reaching 80% of capacity

#### Error Rate SLA

| Metric | Target | Measurement Window | Tolerance |
|--------|--------|-------------------|-----------|
| 4xx Errors | < 1% | 5-minute rolling window | +0.5% |
| 5xx Errors | < 0.1% | 5-minute rolling window | +0.05% |
| Total Error Rate | < 0.5% | Daily | +0.2% |

**Success Criteria:**
- 5xx error rate must be below target 99.9% of the time
- Any 5xx error rate > 1% triggers immediate page
- No more than 3 error rate violations per month

---

### 2. URL Shortening Service

**Priority:** P1 - High

#### Latency SLA

| Metric | Target | Measurement Window | Tolerance |
|--------|--------|-------------------|-----------|
| p50 | < 100ms | 5-minute rolling window | +20% |
| p95 | < 200ms | 5-minute rolling window | +20% |
| p99 | < 500ms | 5-minute rolling window | +30% |

**Success Criteria:**
- Must meet p95 target 95% of the time
- p99 violations allowed for up to 30 minutes per day during peak
- Zero violations during off-peak hours (10 PM - 6 AM UTC)

#### Throughput SLA

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Peak Capacity | 1,000 req/s | Sustained for 1 hour |
| Average Load | 200 req/s | Daily average |

#### Error Rate SLA

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Total Error Rate | < 1% | Daily | +0.5% |

---

### 3. Analytics Service

**Priority:** P2 - Medium

#### Latency SLA

| Metric | Target | Measurement Window | Tolerance |
|--------|--------|-------------------|-----------|
| p50 | < 200ms | 5-minute rolling window | +30% |
| p95 | < 500ms | 5-minute rolling window | +30% |
| p99 | < 1000ms | 5-minute rolling window | +50% |

**Success Criteria:**
- Must meet p95 target 90% of the time
- Violations during high-load periods acceptable with notification
- Response times allowed to degrade for large date ranges

#### Throughput SLA

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Peak Capacity | 500 req/s | Sustained for 1 hour |
| Average Load | 100 req/s | Daily average |

---

## Infrastructure SLAs

### Database Performance

**PostgreSQL:**

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Query Time (p95) | < 10ms | 5-minute rolling window |
| Query Time (p99) | < 25ms | 5-minute rolling window |
| Slow Query Count | < 10 per hour | Hourly |
| Connection Pool Utilization | < 80% | 5-minute rolling window |
| Index Hit Rate | > 99% | Daily |

**Success Criteria:**
- All queries should use appropriate indexes
- Slow queries (>10ms) must be investigated within 24 hours
- Connection pool must auto-scale before hitting 90%

### Cache Performance

**Redis:**

| Metric | Target | Measurement Window |
|--------|--------|-------------------|
| Cache Hit Rate | > 95% | 5-minute rolling window |
| Operation Latency (p95) | < 1ms | 5-minute rolling window |
| Memory Utilization | < 80% | 5-minute rolling window |
| Eviction Rate | < 1% of keys/hour | Hourly |

**Success Criteria:**
- Cache hit rate must stay above 90% at all times
- Drop below 90% triggers automatic investigation
- Memory utilization approaching 80% triggers capacity review

### Resource Utilization

| Resource | Warning Threshold | Critical Threshold | Action |
|----------|------------------|-------------------|---------|
| CPU | 70% | 85% | Auto-scale |
| Memory | 80% | 90% | Auto-scale |
| Disk | 70% | 85% | Alert + Manual review |
| Network | 70% | 85% | Alert + Investigate |

**Success Criteria:**
- Resource utilization should stay below warning threshold during normal operation
- Critical threshold should trigger automatic scaling within 2 minutes
- No manual intervention required for standard scaling scenarios

---

## Data Integrity and Durability

### URL Data

**Durability:** 99.999999999% (11 nines)

**Backup SLA:**
- **Backup Frequency:** Continuous (PITR) + Daily snapshots
- **Backup Retention:** 30 days (daily), 90 days (weekly), 1 year (monthly)
- **Recovery Point Objective (RPO):** < 5 minutes
- **Recovery Time Objective (RTO):** < 1 hour

**Success Criteria:**
- Automated backup testing weekly
- Disaster recovery drill quarterly
- Zero data loss in last 12 months

### Analytics Data

**Durability:** 99.9%

**Backup SLA:**
- **Backup Frequency:** Daily
- **Backup Retention:** 30 days
- **RPO:** < 24 hours
- **RTO:** < 4 hours

---

## Performance Budgets

Performance budgets define acceptable performance ranges for deployments and changes.

### Deployment Performance Budget

**All deployments must meet these criteria:**

| Metric | Maximum Acceptable Degradation |
|--------|-------------------------------|
| Latency (p95) | +5% |
| Error Rate | +0.05% |
| Throughput | -5% |
| Cache Hit Rate | -2% |

**Enforcement:**
- Pre-deployment performance tests required
- Canary deployment monitors performance for 30 minutes
- Automatic rollback if budget exceeded
- Manual approval required for deployments that degrade performance

### Feature Flag Performance Budget

**Features behind flags must meet:**

| Metric | Maximum Acceptable Degradation |
|--------|-------------------------------|
| Latency (p95) | +10% |
| Error Rate | +0.1% |
| Resource Usage | +15% |

**Enforcement:**
- A/B test required for performance-impacting features
- Statistical significance required before 100% rollout
- Flags with negative performance impact must have mitigation plan

---

## Monitoring and Alerting

### Alert Severity Levels

**P0 - Critical (Page Immediately)**
- 5xx error rate > 1%
- Latency p95 exceeds 2x target for >5 minutes
- System availability < 99%
- Database unreachable
- Cache cluster down

**P1 - High (Immediate Response During Business Hours)**
- 5xx error rate > 0.5% for >15 minutes
- Latency p95 exceeds 1.5x target for >15 minutes
- Resource utilization > 90%
- Cache hit rate < 85%

**P2 - Medium (Response Within 4 Hours)**
- Latency p95 exceeds 1.2x target for >30 minutes
- Cache hit rate < 90%
- Resource utilization > 80%
- Slow query count > 50/hour

**P3 - Low (Response Within 24 Hours)**
- Performance trend degradation
- Approaching capacity thresholds
- Minor configuration issues

### On-Call Response Times

| Severity | Acknowledgment | Initial Response | Resolution Target |
|----------|---------------|------------------|------------------|
| P0 | 5 minutes | 15 minutes | 1 hour |
| P1 | 15 minutes | 30 minutes | 4 hours |
| P2 | 1 hour | 4 hours | 24 hours |
| P3 | 4 hours | 24 hours | 1 week |

---

## Performance Review and Reporting

### Daily

**Automated Daily Reports:**
- Performance summary vs. SLAs
- Alert and incident count
- Performance budget compliance
- Top slow queries

**Distribution:** Engineering team via email

### Weekly

**Weekly Performance Review Meeting:**
- SLA compliance review
- Incident post-mortem summaries
- Performance trends
- Optimization progress

**Duration:** 45 minutes
**Attendees:** Engineering lead, Backend engineers, SRE

### Monthly

**Monthly SLA Report:**
- Detailed SLA compliance metrics
- Comparison vs. previous month
- Performance trends and analysis
- Optimization ROI summary
- Capacity planning update

**Distribution:** Engineering leadership, Product team

### Quarterly

**Quarterly Business Review:**
- Performance vs. business goals
- SLA compliance trends
- Infrastructure cost analysis
- Strategic recommendations
- SLA adjustments (if needed)

**Distribution:** Executive team, Engineering leadership

---

## SLA Violation Response

### Detection

**Automated Detection:**
- Real-time monitoring via Prometheus
- Alert routing via PagerDuty/OpsGenie
- Automatic severity assignment
- Escalation if no acknowledgment

### Response Procedure

**Immediate Actions (< 15 minutes):**
1. Acknowledge alert
2. Assess severity and impact
3. Check recent deployments/changes
4. Verify monitoring system health
5. Initiate incident response if needed

**Investigation (< 1 hour):**
1. Identify root cause
2. Determine affected users/requests
3. Assess business impact
4. Decide on mitigation strategy
5. Implement fix or rollback

**Resolution (< 4 hours for P1):**
1. Validate fix effectiveness
2. Confirm SLA compliance restored
3. Monitor for regression
4. Document incident
5. Communicate to stakeholders

### Post-Incident

**Within 48 Hours:**
- [ ] Incident post-mortem scheduled
- [ ] Root cause analysis completed
- [ ] Action items identified and assigned
- [ ] Stakeholders notified of resolution

**Within 1 Week:**
- [ ] Post-mortem document published
- [ ] Preventive measures implemented
- [ ] Monitoring/alerting improved
- [ ] Runbooks updated

---

## SLA Credits and Compensation

### Internal SLA (No Customer Credits)

This SLA is for internal performance accountability. However, we track:

**Performance Score:**
```
Monthly Score = (
  Availability % × 0.30 +
  Latency Compliance % × 0.40 +
  Error Rate Compliance % × 0.20 +
  Capacity Compliance % × 0.10
)
```

**Performance Tiers:**
- **Excellent:** ≥ 99.5%
- **Good:** 98.0% - 99.4%
- **Acceptable:** 95.0% - 97.9%
- **Poor:** < 95.0%

**Consequences:**
- **Excellent:** Team recognition, performance bonus consideration
- **Good:** Standard operation, no action needed
- **Acceptable:** Performance improvement plan required
- **Poor:** Immediate intervention, leadership escalation

---

## SLA Exceptions and Maintenance

### Planned Maintenance

**Allowed Maintenance Windows:**
- **Frequency:** Maximum once per month
- **Duration:** Maximum 2 hours
- **Timing:** Off-peak hours (10 PM - 6 AM UTC)
- **Notification:** 72 hours advance notice

**Maintenance Impact:**
- Excluded from SLA calculations
- Users notified in advance
- Fallback/read-only mode preferred over full outage

### Force Majeure

SLA compliance not required during:
- Natural disasters
- Acts of terrorism
- Internet backbone failures
- Third-party service outages (AWS, etc.)
- Other events beyond reasonable control

**Requirements:**
- Incident must be documented
- Business continuity plan activated
- Stakeholders notified
- Post-incident review conducted

---

## SLA Continuous Improvement

### Review and Adjustment

**SLA Review Schedule:**
- **Quarterly:** Review SLA compliance trends
- **Bi-Annual:** Assess SLA target appropriateness
- **Annual:** Major SLA revision (if needed)

**Adjustment Criteria:**
- Consistent over-performance (>99.9% compliance for 6 months)
- Business requirements change
- Technology improvements enable tighter SLAs
- Competitive pressure requires better performance

**Adjustment Process:**
1. Propose changes with data justification
2. Review with engineering leadership
3. Validate feasibility with capacity planning
4. Approve in quarterly business review
5. Communicate to stakeholders
6. Implement with 30-day notice

### Performance Innovation

**Encouragement:**
- 10% time for performance research
- Performance champions program
- Knowledge sharing sessions
- Recognition for exceptional improvements

**Investment:**
- Continuous monitoring tool improvements
- Performance testing infrastructure
- Training and certifications
- Industry conference attendance

---

## Appendix

### A. Measurement Queries

**Latency (p95):**
```promql
histogram_quantile(0.95,
  sum(rate(http_request_duration_seconds_bucket{endpoint="redirect"}[5m])) by (le)
)
```

**Error Rate:**
```promql
sum(rate(http_requests_total{status=~"5.."}[5m])) /
sum(rate(http_requests_total[5m])) * 100
```

**Cache Hit Rate:**
```promql
sum(rate(cache_hits_total[5m])) /
(sum(rate(cache_hits_total[5m])) + sum(rate(cache_misses_total[5m]))) * 100
```

**Availability:**
```promql
avg_over_time(up[30d]) * 100
```

### B. SLA Compliance Calculation

**Monthly Compliance:**
```
Total Minutes in Month: M
Minutes in Violation: V
Compliance % = ((M - V) / M) × 100
```

**Example:**
```
Month: January (44,640 minutes)
Violations: 45 minutes (P0 incidents)
Compliance: ((44,640 - 45) / 44,640) × 100 = 99.90%
Status: Within SLA (99.9% target)
```

### C. Escalation Matrix

| Severity | Primary | Secondary | Executive |
|----------|---------|-----------|-----------|
| P0 | On-Call Engineer | Engineering Lead | CTO |
| P1 | On-Call Engineer | Engineering Lead | - |
| P2 | Team Member | On-Call Engineer | - |
| P3 | Team Member | - | - |

### D. Related Documents

- [Performance Optimization Framework](./PERFORMANCE_OPTIMIZATION_FRAMEWORK.md)
- [Incident Response Runbook](./runbooks/incident_response.md)
- [Capacity Planning Guide](./capacity_planning_guide.md)
- [Performance Testing Guide](./performance_testing_guide.md)

---

**Document Owner:** Engineering Lead
**Last Updated:** [YYYY-MM-DD]
**Next Review:** [YYYY-MM-DD]
**Version:** 1.0
