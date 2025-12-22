# Performance Tests

This directory contains comprehensive automated performance regression tests for the URL Shortener application.

## Contents

- `api-performance.js` - K6 API load testing suite
- `database_bench_test.go` - Go database performance benchmarks
- `self-healing-utils.js` - Self-healing test utilities with retry logic
- `performance-dashboard.py` - Performance dashboard generator
- `run-benchmarks.sh` - Database benchmark runner script

## Quick Start

```bash
# Run all performance tests
make perf-test

# Or run individual test types
make perf-test-api           # API load tests
make perf-test-db            # Database benchmarks
make perf-test-frontend      # Lighthouse CI (from root)
```

## Test Files

### api-performance.js

Comprehensive API load testing with K6 including:
- Baseline load testing (10 VUs, 2 minutes)
- Spike testing (0 → 200 VUs)
- Soak testing (20 VUs, 5 minutes)
- Stress testing (gradual ramp to 200 VUs)

**Endpoints tested:**
- `POST /api/v1/shorten` - URL shortening
- `GET /{shortCode}` - Redirection
- `GET /api/v1/analytics/{shortCode}` - Analytics
- `GET /health` - Health check

**Thresholds:**
- P95 < 1s
- P99 < 2s
- Error rate < 1%
- Cache hit rate > 90%

### database_bench_test.go

Go benchmarks for database operations:
- `BenchmarkURLInsert` - URL insertion performance
- `BenchmarkURLLookup` - URL lookup by short code
- `BenchmarkURLLookupWithExpiry` - Lookup with expiry check
- `BenchmarkClickInsert` - Click tracking insertion
- `BenchmarkAnalyticsQuery` - Analytics aggregation
- `BenchmarkUserURLsList` - User URLs listing with pagination
- `BenchmarkConcurrentLookups` - Concurrent read performance
- `BenchmarkConcurrentInserts` - Concurrent write performance

**Performance targets:**
- URL lookup: < 100ms (P95)
- URL insert: < 50ms (P95)
- Analytics query: < 100ms (P95)

### self-healing-utils.js

Utilities for stable, reliable performance testing:
- Automatic retry with exponential backoff
- Flaky test detection
- Environment normalization
- Circuit breaker pattern
- Adaptive performance thresholds

**Features:**
- Handles transient failures automatically
- Warms up system before testing
- Normalizes test environment
- Detects and reports flaky tests

### performance-dashboard.py

Python script for generating performance dashboards:
- HTML performance report generation
- Historical trend tracking
- Regression detection
- Statistical analysis

**Usage:**
```bash
python3 performance-dashboard.py
```

**Output:**
- `performance-trends/performance-report.html` - Interactive dashboard
- `performance-trends/history.jsonl` - Historical data

### run-benchmarks.sh

Automated database benchmark runner:
- Runs all Go benchmarks
- Generates summary reports
- Checks performance thresholds
- Exports results for CI/CD

**Usage:**
```bash
cd tests/performance
./run-benchmarks.sh
```

## Configuration

### Environment Variables

```bash
# API load testing
export BASE_URL=http://localhost:8080

# Database benchmarks
export DATABASE_URL=postgres://user:pass@localhost:5432/urlshortener

# Benchmark parameters
export BENCHMARK_TIME=10s
export BENCHMARK_ITERATIONS=1000
```

### Performance Budgets

Performance budgets are defined in `../../performance-budgets.json`

Key budgets:
- API P95: < 1s
- API P99: < 2s
- DB lookup: < 100ms
- Cache hit rate: > 90%

## CI/CD Integration

Tests run automatically in GitHub Actions:
- On every push to `main` or `develop`
- On every pull request
- Daily scheduled runs at 2 AM UTC

See `.github/workflows/performance-testing.yml` for details.

## Results

### Viewing Results

```bash
# K6 results
cat performance-results.json

# Database benchmark results
cat benchmark-results/summary.md

# Performance dashboard
open performance-trends/performance-report.html
```

### Interpreting Results

**Good performance indicators:**
- All thresholds passed
- P95 response times < 1s
- Error rate < 1%
- Consistent performance across runs

**Warning signs:**
- Thresholds approaching limits
- Increasing latency trends
- Error rate 1-5%
- High variance in results

**Critical issues:**
- Threshold failures
- P95 > 1s or P99 > 2s
- Error rate > 5%
- Performance degradation vs baseline

## Self-Healing Features

### Automatic Retry

Transient failures are automatically retried:
- Network errors
- 5xx server errors
- Timeout errors

```javascript
import { resilientRequest } from './self-healing-utils.js';

// Automatically retries on failure
const response = resilientRequest('http://api.example.com/endpoint');
```

### Flaky Test Detection

Tests that fail intermittently are automatically detected:

```javascript
import { FlakyTestDetector } from './self-healing-utils.js';

const detector = new FlakyTestDetector();
detector.recordResult('test-name', passed);

if (detector.isFlaky('test-name')) {
  console.warn('Flaky test detected!');
}
```

### Environment Normalization

System is warmed up before testing:

```javascript
import { withEnvironmentNormalization } from './self-healing-utils.js';

export default withEnvironmentNormalization(testFunction, {
  warmupEndpoints: ['/', '/api/v1/shorten'],
});
```

## Troubleshooting

### K6 Tests Failing

**Check server is running:**
```bash
curl http://localhost:8080/health
```

**Verify environment variables:**
```bash
echo $BASE_URL
```

**Enable verbose output:**
```bash
k6 run --verbose tests/performance/api-performance.js
```

### Database Benchmarks Failing

**Check database connection:**
```bash
psql $DATABASE_URL -c "SELECT 1"
```

**Verify schema:**
```bash
psql $DATABASE_URL -c "\d urls"
```

**Enable Go test verbose mode:**
```bash
go test -v -bench=. ./tests/performance/
```

### High Latency Results

**Possible causes:**
- Cold start (run warmup)
- Resource constraints
- Network latency
- Database not optimized

**Solutions:**
- Increase warmup period
- Check system resources
- Verify indexes exist
- Run on dedicated hardware

## Best Practices

### Before Committing

1. Run quick performance check:
   ```bash
   make perf-test-quick
   ```

2. Review results for regressions

3. Fix any performance issues

### Writing New Tests

1. Use self-healing utilities
2. Set appropriate thresholds
3. Include warmup period
4. Test realistic scenarios
5. Document expected performance

### Analyzing Results

1. Compare against baseline
2. Look for trends over time
3. Investigate anomalies
4. Document findings
5. Update budgets if needed

## Additional Resources

- [Full Documentation](../../docs/PERFORMANCE_TESTING.md)
- [Quick Start Guide](../../docs/PERFORMANCE_QUICKSTART.md)
- [Performance Budgets](../../performance-budgets.json)
- [CI/CD Workflow](../../.github/workflows/performance-testing.yml)

## Contributing

When adding new performance tests:

1. Follow existing patterns
2. Use self-healing utilities
3. Set appropriate thresholds
4. Add to CI/CD workflow
5. Update documentation
6. Include test in Makefile

## Support

For questions or issues:
- Create a GitHub issue with label `performance`
- Review existing documentation
- Check CI/CD logs for details
