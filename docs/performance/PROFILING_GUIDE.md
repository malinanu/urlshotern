# Performance Profiling Guide

This guide provides step-by-step instructions for profiling the URL shortener application to identify performance bottlenecks and establish performance baselines.

## Quick Start

### 1. Start the Profiling Server

```bash
# Build and run the profiling server
go run cmd/profiling/profile_server.go
```

The server starts with:
- Main application on port 8080 (or configured port)
- pprof endpoints at `/debug/pprof/`
- Custom metrics at `/debug/metrics`
- Automatic goroutine and memory logging every 5-10 seconds

### 2. Run Comprehensive Profiling

```bash
# Make the script executable
chmod +x scripts/run_profiling.sh

# Run the full profiling suite
./scripts/run_profiling.sh
```

This script will:
1. Collect CPU profile (30 seconds with load)
2. Collect heap memory snapshot
3. Collect goroutine profile
4. Collect allocation profile
5. Collect blocking operations profile
6. Collect mutex contention profile
7. Run load test and collect metrics
8. Generate comprehensive summary report

**Expected duration:** ~2-3 minutes

### 3. Analyze Results

```bash
# Navigate to results directory
cd profiling_results_YYYYMMDD_HHMMSS/

# View CPU flame graph
open cpu_flamegraph.png

# Interactive CPU profile
go tool pprof -http=:8081 cpu.prof

# Interactive heap profile
go tool pprof -http=:8081 heap.prof

# View metrics analysis
python3 ../scripts/analyze_profiles.py metrics_baseline.json metrics_under_load.json analysis_report.txt
```

---

## Manual Profiling Commands

### CPU Profiling

Capture CPU profile for 30 seconds:
```bash
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof
```

Analyze interactively:
```bash
go tool pprof -http=:8081 cpu.prof
```

View top CPU consumers:
```bash
go tool pprof -top cpu.prof
```

Generate flame graph:
```bash
go tool pprof -png cpu.prof > cpu_flamegraph.png
```

### Memory Profiling

Capture heap snapshot:
```bash
curl http://localhost:8080/debug/pprof/heap -o heap.prof
```

Analyze heap allocations:
```bash
go tool pprof -http=:8081 heap.prof
```

View top allocators:
```bash
go tool pprof -top heap.prof
```

Compare heap over time:
```bash
# Take baseline
curl http://localhost:8080/debug/pprof/heap -o heap_baseline.prof

# Run some operations...

# Take snapshot after load
curl http://localhost:8080/debug/pprof/heap -o heap_load.prof

# Compare
go tool pprof -base=heap_baseline.prof -http=:8081 heap_load.prof
```

### Goroutine Profiling

Capture goroutine snapshot:
```bash
curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof
```

View goroutine stacks:
```bash
go tool pprof -text goroutine.prof
```

Analyze interactively:
```bash
go tool pprof -http=:8081 goroutine.prof
```

### Allocation Profiling

Capture allocation profile:
```bash
curl http://localhost:8080/debug/pprof/allocs -o allocs.prof
```

View top allocators:
```bash
go tool pprof -top allocs.prof
```

### Blocking Operations

Capture blocking profile:
```bash
curl http://localhost:8080/debug/pprof/block -o block.prof
```

Analyze:
```bash
go tool pprof -http=:8081 block.prof
```

### Mutex Contention

Capture mutex profile:
```bash
curl http://localhost:8080/debug/pprof/mutex -o mutex.prof
```

Analyze:
```bash
go tool pprof -http=:8081 mutex.prof
```

---

## Load Testing

### Using wrk (HTTP Benchmarking Tool)

**Install wrk:**
```bash
# macOS
brew install wrk

# Ubuntu/Debian
sudo apt-get install wrk

# Windows
# Download from https://github.com/wg/wrk/releases
```

**Simple load test:**
```bash
# 12 threads, 400 connections, 30 seconds
wrk -t12 -c400 -d30s --latency http://localhost:8080/health
```

**Realistic mixed workload:**
```bash
wrk -t8 -c100 -d60s --latency -s scripts/load_test.lua http://localhost:8080
```

### Using Apache Bench (ab)

```bash
# 10000 requests, 100 concurrent
ab -n 10000 -c 100 http://localhost:8080/health
```

### Using hey (Go-based load generator)

```bash
# Install
go install github.com/rakyll/hey@latest

# Run test
hey -n 10000 -c 100 -m GET http://localhost:8080/health
```

### Using k6 (Modern load testing)

```bash
# Install
brew install k6

# Create test script (k6_test.js)
cat > k6_test.js << 'EOF'
import http from 'k6/http';
import { check, sleep } from 'k6';

export let options = {
  stages: [
    { duration: '30s', target: 50 },
    { duration: '1m', target: 100 },
    { duration: '30s', target: 0 },
  ],
};

export default function() {
  let res = http.get('http://localhost:8080/health');
  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });
  sleep(1);
}
EOF

# Run test
k6 run k6_test.js
```

---

## Custom Metrics Endpoint

The profiling server exposes custom metrics at `/debug/metrics`:

```bash
curl http://localhost:8080/debug/metrics | jq
```

**Response structure:**
```json
{
  "endpoints": {
    "/{shortCode}:GET": {
      "count": 1000,
      "errors": 5,
      "avg_ms": 12,
      "p50_ms": 10,
      "p95_ms": 25,
      "p99_ms": 45,
      "error_rate": 0.5
    }
  },
  "cache_hit_rate": 85.5,
  "cache_hits": 850,
  "cache_misses": 150,
  "database_queries": 200,
  "slow_queries": 15,
  "runtime": {
    "goroutines": 45,
    "heap_alloc_mb": 32.5,
    "heap_sys_mb": 48.0,
    "heap_objects": 125000,
    "gc_runs": 12,
    "last_gc": "2025-11-05T10:30:00Z",
    "gc_pause_ns": 250000
  }
}
```

---

## Interpreting pprof Results

### CPU Profile

**In pprof web UI:**
1. **Flame graph**: Shows call stack hierarchy and CPU time
2. **Top view**: Lists functions consuming most CPU
3. **Source view**: Shows hot lines in source code
4. **Graph view**: Visual representation of call relationships

**Key metrics to look for:**
- `flat`: Time spent in function itself
- `cum`: Cumulative time (including called functions)
- High `flat` values indicate CPU hotspots
- High `cum` with low `flat` indicates bottleneck in called functions

**Example interpretation:**
```
flat    cum      Function
50ms    500ms    RecordClick
 5ms    450ms    ├─ SaveClickEvent (database write)
 2ms    200ms    ├─ ProcessEnhancedClickEvent
 3ms    150ms    └─ RecordTouchpoint
```
→ `RecordClick` spends 50ms in its own code, but 500ms total including database operations

### Heap Profile

**Key views:**
1. **alloc_space**: Total allocated memory (shows allocation rate)
2. **inuse_space**: Currently held memory (shows memory usage)
3. **alloc_objects**: Number of allocations
4. **inuse_objects**: Objects currently in memory

**What to look for:**
- High `alloc_space` but low `inuse_space` → High allocation churn (GC pressure)
- High `inuse_space` → Memory leak or legitimate high usage
- Many small allocations → Consider object pooling

### Goroutine Profile

**Indicators of issues:**
- Thousands of goroutines → Potential leak
- Many blocked goroutines → Contention or slow I/O
- Goroutines waiting on channels → Possible deadlock

**Example healthy profile:**
```
10 goroutines in HTTP handlers (active requests)
5 goroutines in worker pools (expected)
3 background goroutines (timers, cleanups)
---
Total: 18 goroutines
```

**Example problematic profile:**
```
5000 goroutines blocked on database operations
2000 goroutines waiting on channels
---
Total: 7000+ goroutines → PROBLEM
```

---

## Performance Benchmarking

### Establish Baseline

1. **Clean state:**
```bash
# Restart services
docker-compose restart postgres redis

# Start profiling server
go run cmd/profiling/profile_server.go
```

2. **Collect baseline:**
```bash
curl http://localhost:8080/debug/metrics > baseline.json
```

3. **Run consistent load:**
```bash
wrk -t4 -c50 -d60s --latency http://localhost:8080/health
```

4. **Collect under-load metrics:**
```bash
curl http://localhost:8080/debug/metrics > underload.json
```

5. **Compare:**
```bash
python3 scripts/analyze_profiles.py baseline.json underload.json report.txt
```

### Measure Optimization Impact

**Before optimization:**
```bash
./scripts/run_profiling.sh
mv profiling_results_* profiling_before/
```

**After optimization:**
```bash
./scripts/run_profiling.sh
mv profiling_results_* profiling_after/
```

**Compare:**
```bash
# CPU comparison
go tool pprof -base=profiling_before/cpu.prof -http=:8081 profiling_after/cpu.prof

# Heap comparison
go tool pprof -base=profiling_before/heap.prof -http=:8081 profiling_after/heap.prof

# Metrics comparison
python3 scripts/analyze_profiles.py \
    profiling_before/metrics_under_load.json \
    profiling_after/metrics_under_load.json \
    improvement_report.txt
```

---

## Continuous Performance Testing

### Pre-Commit Performance Test

Add to `.git/hooks/pre-commit`:
```bash
#!/bin/bash
# Quick performance regression check

echo "Running performance regression test..."

# Build application
go build -o /tmp/urlshortener cmd/server/main.go

# Start server in background
/tmp/urlshortener &
SERVER_PID=$!
sleep 2

# Quick load test
wrk -t2 -c10 -d5s http://localhost:8080/health > /tmp/perf_result.txt

# Check P95 latency
P95=$(grep "99.000%" /tmp/perf_result.txt | awk '{print $2}')
echo "P95 Latency: $P95"

# Kill server
kill $SERVER_PID

# Fail if P95 > 100ms
if [ $(echo "$P95 > 100" | bc) -eq 1 ]; then
    echo "FAILED: Performance regression detected (P95 > 100ms)"
    exit 1
fi

echo "Performance test passed"
```

### CI/CD Integration

**GitHub Actions example:**
```yaml
name: Performance Test

on: [pull_request]

jobs:
  performance:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: password
      redis:
        image: redis:7

    steps:
      - uses: actions/checkout@v2

      - name: Setup Go
        uses: actions/setup-go@v2
        with:
          go-version: 1.23

      - name: Run Performance Tests
        run: |
          go run cmd/profiling/profile_server.go &
          sleep 5
          ./scripts/run_profiling.sh
          python3 scripts/analyze_profiles.py \
            profiling_results_*/metrics_baseline.json \
            profiling_results_*/metrics_under_load.json \
            analysis.txt
          cat analysis.txt

      - name: Upload Results
        uses: actions/upload-artifact@v2
        with:
          name: profiling-results
          path: profiling_results_*
```

---

## Troubleshooting

### pprof Not Working

**Symptoms:** Cannot connect to pprof endpoints

**Solutions:**
1. Ensure server is running: `curl http://localhost:8080/health`
2. Check port: `lsof -i :8080`
3. Check firewall settings
4. Verify pprof import in code

### High Memory Usage

**Symptoms:** Memory grows continuously

**Debug steps:**
```bash
# Take heap snapshot
curl http://localhost:8080/debug/pprof/heap -o heap.prof

# View top allocators
go tool pprof -top heap.prof

# Check for goroutine leaks
curl http://localhost:8080/debug/pprof/goroutine -o goroutine.prof
go tool pprof -text goroutine.prof
```

### Slow Queries

**Symptoms:** High latency on database operations

**Debug steps:**
1. Enable query logging (see `storage/postgres.go`)
2. Check for missing indexes
3. Use EXPLAIN ANALYZE on slow queries
4. Monitor connection pool utilization

### Load Test Failures

**Symptoms:** Errors during load testing

**Common causes:**
1. Database connection pool exhausted
2. File descriptor limit reached
3. Memory exhaustion
4. Network timeout

**Solutions:**
```bash
# Increase file descriptor limit
ulimit -n 10000

# Increase connection pool
export DB_MAX_CONNECTIONS=100

# Monitor during test
watch -n 1 'curl -s http://localhost:8080/debug/metrics | jq .runtime'
```

---

## Best Practices

### 1. Profile Regularly
- Profile every major release
- Profile after significant changes
- Profile before and after optimizations

### 2. Profile Realistic Workloads
- Use production-like data volumes
- Simulate actual user patterns
- Test peak load scenarios

### 3. Profile in Staging
- Match production environment
- Use production database size
- Test with production traffic patterns

### 4. Keep Historical Data
- Store profiling results with git tags
- Track performance over time
- Create performance regression tests

### 5. Act on Findings
- Prioritize high-impact optimizations
- Measure improvement after changes
- Document optimization decisions

---

## Tools Reference

### Go Profiling Tools
- `go tool pprof`: Built-in profiler
- `go-torch`: Flame graph generator (legacy, now in pprof)
- `profile`: High-level profiling package

### Load Testing Tools
- **wrk**: Fast HTTP benchmarking
- **hey**: Go-based load generator
- **k6**: Modern load testing with scripting
- **Apache Bench (ab)**: Classic benchmarking tool
- **Gatling**: Enterprise-grade load testing

### Monitoring Tools
- **Prometheus**: Metrics collection
- **Grafana**: Metrics visualization
- **Jaeger**: Distributed tracing
- **DataDog/New Relic**: Commercial APM

---

## Additional Resources

- [Go pprof Documentation](https://pkg.go.dev/net/http/pprof)
- [Go Blog: Profiling Go Programs](https://go.dev/blog/pprof)
- [Profiling and Optimizing Go](https://github.com/dgryski/go-perfbook)
- [High Performance Go Workshop](https://dave.cheney.net/high-performance-go-workshop/)
- [Performance Profiling Report](./PERFORMANCE_PROFILING_REPORT.md)

---

## Quick Reference Card

```bash
# Start profiling server
go run cmd/profiling/profile_server.go

# Full profiling suite
./scripts/run_profiling.sh

# Manual CPU profile
curl http://localhost:8080/debug/pprof/profile?seconds=30 -o cpu.prof
go tool pprof -http=:8081 cpu.prof

# Manual heap profile
curl http://localhost:8080/debug/pprof/heap -o heap.prof
go tool pprof -http=:8081 heap.prof

# View metrics
curl http://localhost:8080/debug/metrics | jq

# Load test
wrk -t8 -c100 -d60s --latency -s scripts/load_test.lua http://localhost:8080

# Analyze results
python3 scripts/analyze_profiles.py baseline.json load.json report.txt
```

---

**For detailed performance analysis and optimization strategies, see [PERFORMANCE_PROFILING_REPORT.md](./PERFORMANCE_PROFILING_REPORT.md)**
