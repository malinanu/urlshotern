# Performance Testing Quick Reference

## Quick Commands

```bash
# Quick 30-second performance check
make perf-test-quick

# Run all performance tests
make perf-test

# Individual test types
make perf-test-frontend    # Lighthouse CI
make perf-test-api         # K6 load tests
make perf-test-db          # Database benchmarks

# Generate performance dashboard
make perf-dashboard

# Clean performance artifacts
make perf-clean
```

## Performance Budgets at a Glance

| Category | Metric | Budget | Status Check |
|----------|--------|--------|--------------|
| **Frontend** | | | |
| | LCP | < 2.5s | ✅ < 2.5s, ⚠️ 2.5-4s, ❌ > 4s |
| | Performance Score | ≥ 85 | ✅ ≥ 85, ⚠️ 50-84, ❌ < 50 |
| | JS Bundle | < 700 KB | ✅ < 700 KB, ❌ ≥ 700 KB |
| | Total Size | < 2 MB | ✅ < 2 MB, ❌ ≥ 2 MB |
| **API** | | | |
| | P95 Response | < 1s | ✅ < 1s, ⚠️ 1-2s, ❌ > 2s |
| | P99 Response | < 2s | ✅ < 2s, ⚠️ 2-5s, ❌ > 5s |
| | Error Rate | < 1% | ✅ < 1%, ⚠️ 1-5%, ❌ > 5% |
| | Cache Hit | > 90% | ✅ > 90%, ⚠️ 80-90%, ❌ < 80% |
| **Database** | | | |
| | Lookup | < 100ms | ✅ < 100ms, ⚠️ 100-200ms, ❌ > 200ms |
| | Insert | < 50ms | ✅ < 50ms, ⚠️ 50-100ms, ❌ > 100ms |
| | Analytics | < 100ms | ✅ < 100ms, ⚠️ 100-200ms, ❌ > 200ms |

## CI/CD Triggers

| Trigger | When | Purpose |
|---------|------|---------|
| Push to `main` | Every commit | Production performance tracking |
| Push to `develop` | Every commit | Development performance tracking |
| Pull Request | On creation/update | Regression detection |
| Schedule | Daily @ 2 AM UTC | Trend monitoring |
| Manual | On demand | Ad-hoc testing |

## Understanding Test Results

### Lighthouse CI

**Good:**
- Performance Score: 90-100 (green)
- LCP: < 2.5s
- All Core Web Vitals green

**Needs Improvement:**
- Performance Score: 50-89 (orange)
- LCP: 2.5-4s
- Some metrics in orange

**Poor:**
- Performance Score: < 50 (red)
- LCP: > 4s
- Multiple metrics in red

### K6 API Tests

**Healthy:**
```
✓ http_req_duration: avg=245ms p(95)=450ms p(99)=780ms
✓ http_req_failed: 0.12%
✓ All thresholds passed
```

**Warning:**
```
⚠ http_req_duration: p(95)=950ms
⚠ http_req_failed: 2.5%
⚠ Some thresholds approaching limits
```

**Critical:**
```
✗ http_req_duration: p(95)=1450ms
✗ http_req_failed: 7.3%
✗ Multiple threshold failures
```

### Database Benchmarks

**Good:**
```
BenchmarkURLLookup-8    15000    75432 ns/op   (< 100ms ✓)
BenchmarkURLInsert-8    20000    45123 ns/op   (< 50ms ✓)
```

**Poor:**
```
BenchmarkURLLookup-8    5000     185432 ns/op  (> 100ms ✗)
BenchmarkURLInsert-8    8000     95123 ns/op   (> 50ms ✗)
```

## Troubleshooting

| Issue | Quick Fix |
|-------|-----------|
| k6 not found | `brew install k6` (macOS) or `choco install k6` (Windows) |
| Server not responding | `curl http://localhost:8080/health` |
| DB connection failed | Check PostgreSQL: `docker-compose ps postgres` |
| Tests timing out | Increase timeout in config files |
| High variance | Run multiple iterations, use warmup |
| False positives | Enable self-healing utilities |

## File Locations

```
Performance Testing Files:
├── .github/workflows/performance-testing.yml   # CI/CD
├── lighthouserc.js                             # Lighthouse config
├── performance-budgets.json                    # Budgets
├── tests/performance/
│   ├── api-performance.js                      # K6 tests
│   ├── database_bench_test.go                  # Go benchmarks
│   ├── self-healing-utils.js                   # Utilities
│   ├── performance-dashboard.py                # Dashboard
│   └── run-benchmarks.sh                       # Runner
└── docs/
    ├── PERFORMANCE_TESTING.md                  # Full guide
    └── PERFORMANCE_QUICKSTART.md               # Quick start
```

## Common Scenarios

### Before Pushing Code

```bash
make perf-test-quick
# Review results
# Fix any regressions
git commit && git push
```

### Investigating Regression

```bash
# Check specific component
make perf-test-api

# Generate detailed report
make perf-dashboard

# Review trends
open performance-trends/performance-report.html
```

### Updating Baselines

```bash
# After major feature/optimization
git checkout main
make perf-test
# Results become new baseline
```

### Local Development

```bash
# Start services
make dev

# Run server
go run cmd/server/main.go

# Test in another terminal
make perf-test-quick
```

## Environment Variables

```bash
# API testing
export BASE_URL=http://localhost:8080

# Database benchmarks
export DATABASE_URL=postgres://user:pass@localhost:5432/urlshortener

# Benchmark configuration
export BENCHMARK_TIME=10s
export BENCHMARK_ITERATIONS=1000

# Lighthouse CI
export LHCI_BUILD_CONTEXT__CURRENT_HASH=$GITHUB_SHA
```

## Integration Examples

### With Datadog

```javascript
import { DatadogClient } from 'k6/x/datadog';
const dd = new DatadogClient();
dd.sendMetric('api.response_time', value);
```

### With Grafana

```bash
k6 run --out influxdb=http://localhost:8086/k6 \
  tests/performance/api-performance.js
```

### With Slack

Add to GitHub Secrets:
```
SLACK_WEBHOOK=https://hooks.slack.com/services/YOUR/WEBHOOK/URL
```

## Key Metrics Explained

| Metric | What It Measures | Good Value |
|--------|------------------|------------|
| **LCP** | Largest Contentful Paint - Main content load | < 2.5s |
| **FID** | First Input Delay - Interactivity | < 100ms |
| **CLS** | Cumulative Layout Shift - Visual stability | < 0.1 |
| **P95** | 95th percentile - Most users' experience | < 1s |
| **P99** | 99th percentile - Worst case scenarios | < 2s |
| **RPS** | Requests per second - Throughput | > 100 |
| **Hit Rate** | Cache effectiveness | > 90% |

## Self-Healing Features

| Feature | What It Does | When To Use |
|---------|--------------|-------------|
| Retry Logic | Auto-retries transient failures | Always enabled |
| Flaky Detection | Identifies unreliable tests | Review periodically |
| Warmup | Prepares system before testing | Before all tests |
| Circuit Breaker | Prevents cascade failures | High-load scenarios |
| Adaptive Thresholds | Adjusts to environment | CI/CD environments |

## Resources

- 📖 Full Documentation: `docs/PERFORMANCE_TESTING.md`
- 🚀 Quick Start: `docs/PERFORMANCE_QUICKSTART.md`
- 🧪 Test README: `tests/performance/README.md`
- 📊 Dashboard: `performance-trends/performance-report.html`
- ⚙️ Budgets: `performance-budgets.json`

## Support

- GitHub Issues: Tag with `performance`
- Slack: #performance-testing
- Email: platform-team@example.com

---

**Last Updated:** 2025-01-07
**Version:** 1.0.0
