# Performance Testing - Quick Start Guide

Get started with automated performance testing in 5 minutes.

## Prerequisites

Install the required tools:

```bash
# Node.js and npm (for Lighthouse CI)
# Already installed with the project

# K6 for API load testing
# Windows (using Chocolatey)
choco install k6

# macOS
brew install k6

# Linux
sudo gpg -k
sudo gpg --no-default-keyring --keyring /usr/share/keyrings/k6-archive-keyring.gpg \
  --keyserver hkp://keyserver.ubuntu.com:80 --recv-keys C5AD17C747E3415A3642D57D77C6C491D6AC1D69
echo "deb [signed-by=/usr/share/keyrings/k6-archive-keyring.gpg] https://dl.k6.io/deb stable main" | \
  sudo tee /etc/apt/sources.list.d/k6.list
sudo apt-get update
sudo apt-get install k6

# Python 3 for dashboard generation
# Already installed on most systems
pip install pandas matplotlib numpy
```

## Running Tests

### Quick Performance Check (5 minutes)

```bash
# Start your development environment
make dev

# In another terminal, run the server
go run cmd/server/main.go

# Run a quick performance check
make perf-test-quick
```

This runs a lightweight 30-second load test to verify basic performance.

### Full Performance Test Suite (15 minutes)

```bash
# Run all performance tests
make perf-test
```

This includes:
- Frontend Lighthouse CI tests
- API K6 load tests
- Database benchmarks

### Individual Test Types

```bash
# Frontend only
make perf-test-frontend

# API only
make perf-test-api

# Database only
make perf-test-db
```

## Understanding Results

### Lighthouse CI Results

Look for:
- **Performance Score**: Should be ≥ 85
- **LCP (Largest Contentful Paint)**: Should be < 2.5s (green)
- **CLS (Cumulative Layout Shift)**: Should be < 0.1
- **Bundle Size**: JavaScript should be < 700 KB

Example output:
```
✓ Performance score: 92
✓ LCP: 2.1s
✓ CLS: 0.05
✓ JavaScript bundle: 650 KB
```

### K6 API Results

Look for:
- **P95 Response Time**: < 1000ms
- **P99 Response Time**: < 2000ms
- **Error Rate**: < 1%
- **All thresholds**: PASSED

Example output:
```
✓ http_req_duration..........: avg=245ms min=120ms med=230ms max=890ms p(95)=450ms p(99)=780ms
✓ http_req_failed............: 0.12%
✓ http_reqs..................: 12450
```

### Database Benchmarks

Look for:
- **URL Lookup**: < 100ms per operation
- **URL Insert**: < 50ms per operation
- **ns/op**: Lower is better
- **B/op**: Memory allocations

Example output:
```
BenchmarkURLLookup-8        15000    75432 ns/op    1024 B/op    12 allocs/op
BenchmarkURLInsert-8        20000    45123 ns/op     512 B/op     8 allocs/op
```

## Viewing the Dashboard

Generate and view the performance dashboard:

```bash
# Generate dashboard
make perf-dashboard

# Open the HTML report
# Windows
start performance-trends/performance-report.html

# macOS
open performance-trends/performance-report.html

# Linux
xdg-open performance-trends/performance-report.html
```

## Common Issues

### Issue: "k6 not found"

**Solution:**
```bash
# Install k6 (see Prerequisites above)
# Verify installation
k6 version
```

### Issue: "Server not responding"

**Solution:**
```bash
# Ensure services are running
docker-compose -f docker/docker-compose.local.yml ps

# Check server health
curl http://localhost:8080/health
```

### Issue: "Database connection failed"

**Solution:**
```bash
# Check PostgreSQL is running
docker-compose -f docker/docker-compose.local.yml ps postgres

# Verify connection
psql -h localhost -U postgres -d urlshortener
```

## CI/CD Integration

Performance tests run automatically on:
- Every push to `main` or `develop`
- Every pull request
- Daily at 2 AM UTC

View results:
1. Go to GitHub Actions
2. Select the "Performance Testing" workflow
3. Click on the latest run
4. Review job outputs and artifacts

## Performance Budgets

Current budgets are defined in `performance-budgets.json`:

| Category | Metric | Budget |
|----------|--------|--------|
| Frontend | LCP | < 2.5s |
| Frontend | JS Bundle | < 700 KB |
| API | P95 Response | < 1s |
| API | Error Rate | < 1% |
| Database | Lookup | < 100ms |
| Cache | Hit Rate | > 90% |

Tests will **fail** if budgets are exceeded.

## Next Steps

1. **Review Full Documentation**: See `docs/PERFORMANCE_TESTING.md`
2. **Customize Budgets**: Edit `performance-budgets.json` for your needs
3. **Set Up Alerts**: Configure Slack webhooks for notifications
4. **Track Trends**: Review weekly performance dashboards

## Getting Help

- Documentation: `docs/PERFORMANCE_TESTING.md`
- GitHub Issues: Tag with `performance`
- Team Channel: #performance-testing (if configured)

## Cheat Sheet

```bash
# Quick commands
make perf-test-quick          # 30-second quick check
make perf-test               # Full test suite
make perf-test-frontend      # Frontend only
make perf-test-api           # API only
make perf-test-db            # Database only
make perf-dashboard          # Generate dashboard
make perf-clean              # Clean artifacts

# Manual runs
npx lhci autorun             # Lighthouse CI
k6 run tests/performance/api-performance.js  # K6
go test -bench=. ./tests/performance/        # Go benchmarks

# View results
cat .lighthouseci/manifest.json              # Lighthouse
cat performance-results.json                 # K6
cat tests/performance/benchmark-results/summary.md  # DB
```
