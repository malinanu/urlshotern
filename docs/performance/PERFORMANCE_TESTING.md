# Performance Testing Guide

## Overview

This document describes the comprehensive automated performance regression testing system for the URL Shortener application. The system provides 100% performance regression detection with zero-friction CI/CD integration.

## Table of Contents

- [Architecture](#architecture)
- [Performance Budgets](#performance-budgets)
- [Testing Types](#testing-types)
- [CI/CD Integration](#cicd-integration)
- [Self-Healing Tests](#self-healing-tests)
- [Monitoring & Alerting](#monitoring--alerting)
- [Running Tests Locally](#running-tests-locally)
- [Troubleshooting](#troubleshooting)

## Architecture

The performance testing system consists of four main components:

```
┌─────────────────────────────────────────────────────────────┐
│                     CI/CD Pipeline                          │
│  ┌──────────────┐  ┌──────────────┐  ┌──────────────┐     │
│  │  Lighthouse  │  │     K6       │  │   Database   │     │
│  │   Frontend   │  │     API      │  │  Benchmarks  │     │
│  └──────┬───────┘  └──────┬───────┘  └──────┬───────┘     │
│         │                  │                  │             │
│         └──────────┬───────┴──────────────────┘             │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │  Performance Budget  │                           │
│         │  Enforcement Engine  │                           │
│         └──────────┬───────────┘                           │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │  Regression Detector │                           │
│         │   & Trend Tracker    │                           │
│         └──────────┬───────────┘                           │
│                    │                                        │
│         ┌──────────▼───────────┐                           │
│         │   Alert & Report     │                           │
│         │     Generator        │                           │
│         └──────────────────────┘                           │
└─────────────────────────────────────────────────────────────┘
```

## Performance Budgets

Performance budgets are defined in `performance-budgets.json` and enforced automatically in CI/CD.

### Frontend Budgets

| Metric | Budget | Description |
|--------|--------|-------------|
| LCP | < 2.5s | Largest Contentful Paint |
| FID | < 100ms | First Input Delay |
| CLS | < 0.1 | Cumulative Layout Shift |
| FCP | < 1.8s | First Contentful Paint |
| TTI | < 3.8s | Time to Interactive |
| TBT | < 300ms | Total Blocking Time |
| Performance Score | ≥ 85 | Lighthouse performance score |
| JavaScript Bundle | < 700 KB | Total JS size (gzipped) |
| CSS Bundle | < 100 KB | Total CSS size (gzipped) |
| Total Page Weight | < 2 MB | All resources combined |

### API Budgets

| Metric | Budget | Description |
|--------|--------|-------------|
| P95 Response Time | < 1s | 95th percentile response time |
| P99 Response Time | < 2s | 99th percentile response time |
| Average Response | < 500ms | Mean response time |
| Error Rate | < 1% | HTTP error rate |
| Shorten Endpoint (P95) | < 800ms | URL shortening performance |
| Redirect Endpoint (P95) | < 200ms | Redirection performance |
| Analytics Endpoint (P95) | < 500ms | Analytics query performance |

### Database Budgets

| Metric | Budget | Description |
|--------|--------|-------------|
| URL Lookup | < 100ms | Query latency (P95) |
| URL Insert | < 50ms | Insert latency (P95) |
| Analytics Query | < 100ms | Aggregation query (P95) |
| Click Insert | < 50ms | Click tracking insert (P95) |
| Cache Hit Rate | > 90% | Redis cache effectiveness |

## Testing Types

### 1. Lighthouse CI - Frontend Performance

**Purpose:** Measure Core Web Vitals and frontend performance metrics.

**Configuration:** `lighthouserc.js`

**Runs:**
- 3 iterations per URL for stable metrics
- Desktop preset
- Tests homepage, pricing, dashboard, and analytics pages

**Key Metrics:**
- Performance, Accessibility, Best Practices, SEO scores
- Core Web Vitals (LCP, FID, CLS)
- Resource budgets and network metrics

### 2. K6 - API Load Testing

**Purpose:** Comprehensive API performance and load testing.

**Configuration:** `tests/performance/api-performance.js`

**Test Scenarios:**
- **Baseline:** 10 VUs for 2 minutes
- **Spike:** Ramp from 0 to 200 VUs
- **Soak:** 20 VUs for 5 minutes
- **Stress:** Gradual ramp to 200 VUs

**Endpoints Tested:**
- `POST /api/v1/shorten` - URL shortening
- `GET /{shortCode}` - Redirection
- `GET /api/v1/analytics/{shortCode}` - Analytics
- `GET /health` - Health check

### 3. Database Benchmarks

**Purpose:** Measure database query performance and identify bottlenecks.

**Configuration:** `tests/performance/database_bench_test.go`

**Benchmarks:**
- URL insert/lookup operations
- Click tracking insertion
- Analytics aggregation queries
- Concurrent read/write performance
- Index effectiveness testing

**Running Manually:**
```bash
cd tests/performance
./run-benchmarks.sh
```

## CI/CD Integration

### GitHub Actions Workflow

Performance tests run automatically on:
- Every push to `main` or `develop`
- Every pull request to `main`
- Daily scheduled runs at 2 AM UTC
- Manual workflow dispatch

**Workflow File:** `.github/workflows/performance-testing.yml`

### Workflow Jobs

1. **lighthouse-ci** - Frontend performance testing
2. **k6-api-performance** - API load testing
3. **database-benchmarks** - Database performance testing
4. **performance-trends** - Trend analysis and tracking
5. **performance-regression-check** - Regression detection
6. **synthetic-monitoring** - Production monitoring

### Performance Gates

Tests will **FAIL** the build if:
- Any performance budget is exceeded
- Regression > 10% detected vs baseline
- Critical regression > 25% detected
- Error rate exceeds thresholds

### PR Comments

The workflow automatically comments on pull requests with:
- Lighthouse performance scores
- K6 API performance metrics
- Database benchmark results
- Regression alerts (if any)

## Self-Healing Tests

### Features

1. **Automatic Retry with Exponential Backoff**
   - Retries transient failures (network errors, 5xx responses)
   - Configurable max retries and backoff strategy
   - Jitter to avoid thundering herd

2. **Flaky Test Detection**
   - Tracks test success/failure rates
   - Identifies tests with >30% failure rate
   - Auto-reports flaky tests in CI

3. **Environment Normalization**
   - Warmup requests before testing
   - Health check verification
   - Stabilization period

4. **Circuit Breaker Pattern**
   - Prevents cascading failures
   - Auto-recovery mechanism
   - Configurable failure thresholds

5. **Adaptive Thresholds**
   - Adjusts thresholds based on recent performance
   - Tolerates environmental variations
   - Reduces false positives

### Usage Example

```javascript
import { selfHealingTest } from './self-healing-utils.js';

export default selfHealingTest({
  name: 'URL Shortening Test',
  testFn: (data, { detector, thresholds }) => {
    // Your test logic here
  },
  warmupEndpoints: ['/', '/api/v1/shorten'],
  retryConfig: {
    maxRetries: 3,
    baseDelay: 1000,
  },
  adaptiveThresholds: {
    responseTime: 500, // ms
  },
});
```

## Monitoring & Alerting

### Performance Trends Dashboard

**Location:** `performance-trends/performance-report.html`

**Features:**
- Historical performance data (30 days)
- Trend visualization
- Regression highlights
- Metric comparisons

**Generation:**
```bash
python3 tests/performance/performance-dashboard.py
```

### Alert Channels

Alerts are sent on:
- Performance regression detected
- Budget exceeded
- Production degradation

**Channels:**
- GitHub PR comments
- Slack notifications (configure webhook)
- Email alerts (configure recipients)

### Regression Detection

**Thresholds:**
- **Warning:** > 10% regression
- **Critical:** > 25% regression

**Baseline:**
- Uses latest `main` branch metrics
- Updates automatically on merge

## Running Tests Locally

### Prerequisites

```bash
# Install dependencies
npm install -g @lhci/cli k6
pip install pandas matplotlib numpy

# Ensure services are running
docker-compose -f docker/docker-compose.local.yml up -d
```

### Frontend Performance Tests

```bash
# Build frontend
cd frontend
npm run build
npm start

# Run Lighthouse CI
npx lhci autorun
```

### API Performance Tests

```bash
# Start backend server
go run cmd/server/main.go

# Run K6 tests
k6 run tests/performance/api-performance.js
```

### Database Benchmarks

```bash
# Set environment variables
export DATABASE_URL="postgres://user:pass@localhost:5432/urlshortener"

# Run benchmarks
cd tests/performance
chmod +x run-benchmarks.sh
./run-benchmarks.sh
```

### Full Test Suite

```bash
# Run all performance tests
make perf-test

# Or manually trigger GitHub Action
gh workflow run performance-testing.yml
```

## Interpreting Results

### Lighthouse Results

**Good:**
- Performance score ≥ 90
- All Core Web Vitals in green
- No accessibility violations

**Needs Improvement:**
- Performance score 50-89
- Some Core Web Vitals in orange
- Bundle size warnings

**Poor:**
- Performance score < 50
- Core Web Vitals in red
- Budget exceeded

### K6 Results

**Success Indicators:**
- P95 < 1s, P99 < 2s
- Error rate < 1%
- All thresholds passed
- Cache hit rate > 90%

**Warning Signs:**
- P95 approaching 1s
- Error rate 1-5%
- Threshold warnings
- Cache hit rate < 90%

**Critical Issues:**
- P95 > 1s or P99 > 2s
- Error rate > 5%
- Multiple threshold failures
- System instability

### Database Benchmarks

**Healthy Performance:**
- Lookups < 100ms (P95)
- Inserts < 50ms (P95)
- Consistent performance across runs
- Indexes utilized effectively

**Performance Concerns:**
- Increasing latency trends
- High variance in results
- Sequential scans instead of index usage
- Connection pool exhaustion

## Troubleshooting

### Common Issues

#### 1. Lighthouse Fails with Timeout

**Symptom:** Lighthouse CI times out waiting for server

**Solution:**
```bash
# Increase timeout in lighthouserc.js
startServerReadyTimeout: 60000

# Verify server starts correctly
npm run build && npm start
curl http://localhost:3000/health
```

#### 2. K6 Tests Fail with Connection Refused

**Symptom:** Cannot connect to API server

**Solution:**
```bash
# Ensure server is running
curl http://localhost:8080/health

# Check BASE_URL environment variable
export BASE_URL=http://localhost:8080
```

#### 3. Database Benchmarks Show High Latency

**Symptom:** All queries slower than expected

**Solution:**
```bash
# Check database health
psql -h localhost -U postgres -d urlshortener -c "SELECT version();"

# Verify indexes exist
psql -h localhost -U postgres -d urlshortener -c "\d urls"

# Check connection pool settings
# Ensure max connections configured correctly
```

#### 4. Flaky Tests in CI

**Symptom:** Tests pass locally but fail in CI

**Solution:**
- Check if using self-healing utilities
- Increase warmup period
- Add retry logic
- Review resource constraints in CI environment

#### 5. Performance Regression False Positives

**Symptom:** Regressions detected but no code changes

**Solution:**
- Use adaptive thresholds
- Increase regression threshold tolerance
- Run multiple iterations for stable baseline
- Check CI environment resource availability

### Debug Mode

Enable verbose logging:

```bash
# Lighthouse
DEBUG=lhci:* npx lhci autorun

# K6
k6 run --verbose tests/performance/api-performance.js

# Database benchmarks
go test -v -bench=. ./tests/performance/
```

## Best Practices

### For Developers

1. **Run tests before pushing:**
   ```bash
   make perf-test-quick
   ```

2. **Review budgets regularly:**
   - Adjust budgets in `performance-budgets.json`
   - Document reasons for changes

3. **Optimize early:**
   - Don't wait for CI failures
   - Use local profiling tools
   - Address warnings proactively

4. **Monitor trends:**
   - Review performance dashboard weekly
   - Track metrics over time
   - Identify gradual degradation

### For DevOps/SRE

1. **Baseline Updates:**
   - Update baseline after major releases
   - Document performance characteristics
   - Track infrastructure changes

2. **Alert Tuning:**
   - Adjust thresholds based on SLOs
   - Reduce noise from false positives
   - Ensure critical alerts are actionable

3. **Capacity Planning:**
   - Use load test results for scaling
   - Monitor resource utilization trends
   - Plan for peak traffic scenarios

## Integration with Other Tools

### Datadog / New Relic

Export metrics for APM integration:

```javascript
// In K6 script
import { DatadogClient } from 'k6/x/datadog';

const datadog = new DatadogClient();
datadog.sendMetric('api.response_time', value);
```

### Grafana

Import performance metrics:

```bash
# Export K6 results to InfluxDB
k6 run --out influxdb=http://localhost:8086/k6 tests/performance/api-performance.js
```

### Sentry

Integrate performance monitoring:

```javascript
Sentry.setMeasurement("lcp", lcpValue, "millisecond");
```

## Maintenance

### Weekly Tasks

- Review performance dashboard
- Check for new regressions
- Update baselines if needed

### Monthly Tasks

- Review and adjust budgets
- Analyze performance trends
- Update documentation

### Quarterly Tasks

- Audit test coverage
- Evaluate new testing tools
- Review alert effectiveness
- Capacity planning review

## Support

For questions or issues:
- Create a GitHub issue with label `performance`
- Contact: platform-team@example.com
- Slack: #performance-testing

## References

- [Lighthouse CI Documentation](https://github.com/GoogleChrome/lighthouse-ci)
- [K6 Documentation](https://k6.io/docs/)
- [Web Vitals](https://web.dev/vitals/)
- [Go Benchmarking](https://golang.org/pkg/testing/#hdr-Benchmarks)
