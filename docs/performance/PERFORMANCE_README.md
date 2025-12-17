# Continuous Performance Optimization Framework

## Overview

This directory contains a comprehensive framework for continuously optimizing the performance of the URL Shortener application. The framework enables sustainable 5-10% monthly performance improvements through systematic measurement, optimization, and validation.

## Quick Links

- **Main Framework:** [PERFORMANCE_OPTIMIZATION_FRAMEWORK.md](./PERFORMANCE_OPTIMIZATION_FRAMEWORK.md)
- **Implementation Guide:** [PERFORMANCE_FRAMEWORK_IMPLEMENTATION_GUIDE.md](./PERFORMANCE_FRAMEWORK_IMPLEMENTATION_GUIDE.md)
- **SLA Documentation:** [PERFORMANCE_SLA.md](./PERFORMANCE_SLA.md)
- **Templates:** [templates/](./templates/)
- **Tools:** [../tools/performance/](../tools/performance/)
- **Scripts:** [../scripts/performance/](../scripts/performance/)

---

## What's Included

### Documentation

1. **Performance Optimization Framework** - Complete framework covering:
   - Performance tracking and baselines
   - Optimization backlog management
   - A/B testing infrastructure
   - Capacity planning
   - Team processes and workflows

2. **Implementation Guide** - 8-week step-by-step guide to:
   - Set up monitoring infrastructure
   - Establish performance processes
   - Automate testing and reporting
   - Execute first optimizations

3. **Performance SLA** - Service level agreements defining:
   - Performance targets by service
   - Measurement methodologies
   - Alert severity levels
   - Incident response procedures

### Templates

Located in [templates/](./templates/):

- **Weekly Performance Review** - Template for weekly team meetings
- **Monthly Optimization Plan** - Template for monthly planning sessions
- **Optimization Item** - Detailed template for tracking optimizations
- **Performance A/B Test** - Template for performance experiments

### Tools

Located in [../tools/performance/](../tools/performance/):

1. **Baseline Collector** (`baseline_collector.go`)
   - Collects performance metrics from Prometheus
   - Generates baseline snapshots for comparison
   - Supports historical trend analysis

2. **Capacity Planner** (`capacity_planner.go`)
   - Forecasts future capacity needs
   - Models different growth scenarios
   - Generates scaling recommendations

### Scripts

Located in [../scripts/performance/](../scripts/performance/):

1. **Weekly Report Generator** (`weekly_report.sh`)
   - Automated weekly performance report
   - Sends via email and Slack
   - Tracks week-over-week trends

---

## Quick Start

### 1. Prerequisites

```bash
# Required infrastructure
- Prometheus (metrics collection)
- Grafana (dashboards)
- PostgreSQL (database)
- Redis (cache)

# Required tools
- Go 1.21+
- curl
- jq (optional, for JSON parsing)
```

### 2. Collect First Baseline

```bash
# Collect current performance baseline
make collect-baseline

# Or run manually
cd tools/performance
go run baseline_collector.go \
  --prometheus-url=http://localhost:9090 \
  --lookback=168 \
  --output=../../baselines/baseline_$(date +%Y%m%d).json
```

### 3. Generate Capacity Plan

```bash
# Generate 90-day capacity forecast
make capacity-plan

# Or run manually with specific scenario
cd tools/performance
go run capacity_planner.go \
  --forecast=90 \
  --scenario=expected \
  --output=../../reports/capacity_plan.json
```

### 4. Set Up Weekly Reviews

```bash
# Schedule weekly performance review
# Copy template to your docs
cp docs/templates/WEEKLY_PERFORMANCE_REVIEW_TEMPLATE.md \
   docs/reviews/review_$(date +%Y%m%d).md

# Conduct 45-minute meeting every week
# Fill in metrics and discussion points
```

---

## Daily Usage

### Performance Monitoring

```bash
# Check current metrics
curl http://localhost:9090/api/v1/query?query=http_requests_total

# View Grafana dashboards
open http://localhost:3000/dashboards

# Check alerts
curl http://localhost:9090/api/v1/alerts
```

### Running Performance Tests

```bash
# Quick performance check
make perf-test-quick

# Full performance test suite
make perf-test

# Benchmarks only
make bench
```

### Collecting Metrics

```bash
# Daily baseline collection (automated via cron recommended)
make collect-baseline

# Manual collection with specific timeframe
cd tools/performance
go run baseline_collector.go --lookback=24 --output=baseline_daily.json
```

---

## Weekly Workflow

### Monday: Weekly Review Meeting

1. **Prepare (30 minutes before meeting):**
   ```bash
   # Collect latest metrics
   make collect-baseline

   # Copy review template
   cp docs/templates/WEEKLY_PERFORMANCE_REVIEW_TEMPLATE.md \
      docs/reviews/review_$(date +%Y%m%d).md

   # Fill in metrics from Grafana
   ```

2. **Conduct Meeting (45 minutes):**
   - Review key metrics (10 min)
   - Discuss alerts/incidents (15 min)
   - Review in-flight optimizations (10 min)
   - Prioritize backlog (10 min)

3. **Follow-up (15 minutes after meeting):**
   - Update GitHub project board
   - Send meeting notes
   - Create action items

### Throughout the Week

- **Daily:** Monitor dashboards and alerts
- **Mid-week:** Check optimization progress
- **Friday:** Prepare next week's priorities

### Friday: Weekly Report

```bash
# Generate and send weekly report
make weekly-report

# Or manually
./scripts/performance/weekly_report.sh
```

---

## Monthly Workflow

### First Week: Planning

1. **Review Previous Month:**
   - Collect month's metrics
   - Validate completed optimizations
   - Calculate ROI accuracy

2. **Plan Next Month:**
   ```bash
   # Copy monthly plan template
   cp docs/templates/MONTHLY_OPTIMIZATION_PLAN_TEMPLATE.md \
      docs/plans/plan_$(date +%Y%m).md

   # Fill in:
   # - Performance goals
   # - Prioritized optimizations
   # - Capacity planning
   # - Team allocation
   ```

3. **Conduct Planning Meeting (2 hours):**
   - Review previous month (30 min)
   - Trend analysis (30 min)
   - Capacity planning (20 min)
   - Optimization roadmap (30 min)
   - Next steps (10 min)

### Throughout the Month

- **Weekly:** Execute planned optimizations
- **Mid-month:** Progress check and adjustments
- **End of month:** Validation and documentation

---

## Key Processes

### 1. Adding New Optimizations

```bash
# 1. Create optimization item from template
cp docs/templates/OPTIMIZATION_ITEM_TEMPLATE.md \
   optimizations/OPT-2024-001-cache-ttl-optimization.md

# 2. Fill in all sections:
# - Impact assessment
# - Effort estimate
# - Technical details
# - Success criteria
# - Testing plan

# 3. Calculate ROI score
ROI = (Impact Score × User Impact %) / Effort Score

# 4. Prioritize and add to backlog
# Add to GitHub Projects board

# 5. Assign owner and target date
```

### 2. Running A/B Tests

```bash
# 1. Create test from template
cp docs/templates/PERFORMANCE_AB_TEST_TEMPLATE.md \
   ab-tests/ABT-PERF-2024-001-cache-ttl.md

# 2. Configure test using existing A/B testing service
# See internal/services/ab_testing.go

# 3. Monitor test metrics
# Use Grafana dashboard

# 4. Analyze results
# Use statistical significance calculator

# 5. Make deployment decision
# Deploy winner or continue testing
```

### 3. Handling Performance Incidents

```bash
# 1. Incident detected via alert
# PagerDuty/Slack notification

# 2. Acknowledge and investigate
# Check dashboards, recent deployments, logs

# 3. Mitigate
# Rollback, scale up, or apply hot fix

# 4. Document
# Create incident post-mortem

# 5. Prevent
# Add monitoring, update runbooks, implement fixes
```

---

## Performance Budgets

### Current Budgets

| Endpoint | p95 Latency | Error Rate | Cache Hit Rate |
|----------|-------------|------------|----------------|
| GET /{shortCode} | < 50ms | < 0.1% | > 95% |
| POST /api/v1/shorten | < 200ms | < 0.5% | > 90% |
| GET /api/v1/analytics | < 500ms | < 1% | > 85% |

### Enforcement

```bash
# Check performance budget in CI/CD
make performance-budget-check

# Block deployment if budget exceeded by >10%
# Require manual approval for >5% degradation
```

---

## Dashboards

### Executive Dashboard
- **URL:** http://localhost:3000/d/performance-executive
- **Update Frequency:** 5 minutes
- **Metrics:** High-level KPIs, system health, cost trends

### Operations Dashboard
- **URL:** http://localhost:3000/d/performance-operations
- **Update Frequency:** 30 seconds
- **Metrics:** Real-time latency, errors, resource utilization

### Engineering Dashboard
- **URL:** http://localhost:3000/d/performance-engineering
- **Update Frequency:** 1 minute
- **Metrics:** Detailed breakdowns, cache stats, database queries

---

## Alerts

### Critical (P0)
- 5xx error rate > 1%
- Latency p95 > 2x budget for > 5 minutes
- System availability < 99%

### High (P1)
- 5xx error rate > 0.5% for > 15 minutes
- Latency p95 > 1.5x budget for > 15 minutes
- Resource utilization > 90%

### Medium (P2)
- Latency p95 > 1.2x budget for > 30 minutes
- Cache hit rate < 90%
- Slow query count > 50/hour

---

## Success Metrics

### After 8 Weeks
- [ ] 10-20% improvement in at least one key metric
- [ ] Performance budgets enforced in CI/CD
- [ ] Zero performance regressions
- [ ] Automated monitoring operational

### After 6 Months
- [ ] Sustained 5-10% monthly improvements
- [ ] 90%+ ROI accuracy
- [ ] Performance culture established
- [ ] Cost per million requests reduced by 20%

---

## Common Commands

```bash
# Collect baseline
make collect-baseline

# Generate capacity plan
make capacity-plan

# Generate weekly report
make weekly-report

# Run performance tests
make perf-test

# Check performance budget
make performance-budget-check

# Compare benchmarks
make benchmark-compare

# Run quick load test
make perf-test-quick
```

---

## Troubleshooting

### Issue: Baseline collector fails

```bash
# Check Prometheus connectivity
curl http://localhost:9090/api/v1/query?query=up

# Verify metrics exist
curl http://localhost:9090/api/v1/query?query=http_requests_total

# Check logs
cd tools/performance
go run baseline_collector.go --prometheus-url=http://localhost:9090 2>&1
```

### Issue: Dashboards show no data

```bash
# Verify Prometheus data source in Grafana
# Check Configuration > Data Sources

# Test query in Prometheus directly
curl http://localhost:9090/api/v1/query?query=http_requests_total

# Check time range in Grafana
# Ensure it matches available data
```

### Issue: Performance budget check fails in CI

```bash
# Review latest metrics
make collect-baseline

# Compare with budget
cat baselines/baseline_latest.json | jq '.endpoints'

# Identify source of regression
# Check recent commits and deployments
```

---

## Support

### Internal Resources
- **Slack:** #performance
- **Wiki:** [Performance Optimization](https://wiki.example.com/performance)
- **Team:** performance-team@example.com

### External Resources
- [Google SRE Book](https://sre.google/sre-book/)
- [Prometheus Documentation](https://prometheus.io/docs/)
- [Performance Optimization Guide](https://web.dev/performance/)

---

## Contributing

### Adding New Tools

1. Create tool in `tools/performance/`
2. Add Go module dependencies
3. Document usage in tool file
4. Add Makefile target
5. Update this README

### Improving Processes

1. Propose changes in #performance
2. Discuss in weekly review
3. Trial for 1 month
4. Document if successful
5. Update framework docs

### Sharing Learnings

1. Document in optimization item
2. Present in team meeting
3. Add to playbooks
4. Share in tech talk

---

## Changelog

### Version 1.0 (2024-XX-XX)
- Initial framework release
- Baseline collector tool
- Capacity planner tool
- Weekly report generator
- Performance SLA documentation
- Implementation guide

---

## License

Internal use only. Not for distribution outside the organization.

---

**Maintained by:** Performance Engineering Team
**Last Updated:** 2024-XX-XX
**Version:** 1.0
