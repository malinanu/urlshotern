#!/bin/bash

# URL Shortener Performance Profiling Script
# This script runs comprehensive profiling on the application

set -e

echo "=========================================="
echo "URL Shortener Performance Profiling Suite"
echo "=========================================="
echo ""

# Configuration
SERVER_PORT=${SERVER_PORT:-8080}
PROFILE_DURATION=${PROFILE_DURATION:-30}
LOAD_DURATION=${LOAD_DURATION:-60}
LOAD_CONNECTIONS=${LOAD_CONNECTIONS:-100}
OUTPUT_DIR="profiling_results_$(date +%Y%m%d_%H%M%S)"

mkdir -p "$OUTPUT_DIR"

echo "Output directory: $OUTPUT_DIR"
echo "Server port: $SERVER_PORT"
echo "Profile duration: ${PROFILE_DURATION}s"
echo ""

# Check if server is running
check_server() {
    if ! curl -s "http://localhost:$SERVER_PORT/health" > /dev/null 2>&1; then
        echo "ERROR: Server is not running on port $SERVER_PORT"
        echo "Start the profiling server with: go run cmd/profiling/profile_server.go"
        exit 1
    fi
    echo "✓ Server is running"
}

# Function to run CPU profiling
profile_cpu() {
    echo ""
    echo "1. CPU Profiling"
    echo "   Duration: ${PROFILE_DURATION}s"
    echo "   Starting CPU profile collection..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/profile?seconds=$PROFILE_DURATION" \
        -o "$OUTPUT_DIR/cpu.prof" &

    PROFILE_PID=$!

    # Generate load during profiling
    echo "   Generating load during profiling..."
    wrk -t4 -c50 -d${PROFILE_DURATION}s --latency \
        "http://localhost:$SERVER_PORT/health" \
        > "$OUTPUT_DIR/cpu_load_results.txt" 2>&1 || true

    wait $PROFILE_PID
    echo "   ✓ CPU profile saved to $OUTPUT_DIR/cpu.prof"
    echo "   Generating flame graph..."
    go tool pprof -png "$OUTPUT_DIR/cpu.prof" > "$OUTPUT_DIR/cpu_flamegraph.png" 2>/dev/null || true
    echo "   ✓ Flame graph saved to $OUTPUT_DIR/cpu_flamegraph.png"
}

# Function to run heap profiling
profile_heap() {
    echo ""
    echo "2. Heap Memory Profiling"
    echo "   Collecting heap snapshot..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/heap" \
        -o "$OUTPUT_DIR/heap.prof"

    echo "   ✓ Heap profile saved to $OUTPUT_DIR/heap.prof"
    echo "   Generating visualization..."
    go tool pprof -png "$OUTPUT_DIR/heap.prof" > "$OUTPUT_DIR/heap_visualization.png" 2>/dev/null || true
    go tool pprof -top "$OUTPUT_DIR/heap.prof" > "$OUTPUT_DIR/heap_top.txt" 2>/dev/null || true
    echo "   ✓ Heap analysis saved"
}

# Function to run goroutine profiling
profile_goroutines() {
    echo ""
    echo "3. Goroutine Profiling"
    echo "   Collecting goroutine snapshot..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/goroutine" \
        -o "$OUTPUT_DIR/goroutine.prof"

    echo "   ✓ Goroutine profile saved to $OUTPUT_DIR/goroutine.prof"
    echo "   Analyzing goroutines..."
    go tool pprof -top "$OUTPUT_DIR/goroutine.prof" > "$OUTPUT_DIR/goroutine_analysis.txt" 2>/dev/null || true
    echo "   ✓ Goroutine analysis saved"
}

# Function to collect allocation profiling
profile_allocs() {
    echo ""
    echo "4. Allocation Profiling"
    echo "   Collecting allocation data..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/allocs" \
        -o "$OUTPUT_DIR/allocs.prof"

    echo "   ✓ Allocation profile saved to $OUTPUT_DIR/allocs.prof"
    go tool pprof -top "$OUTPUT_DIR/allocs.prof" > "$OUTPUT_DIR/allocs_top.txt" 2>/dev/null || true
    echo "   ✓ Allocation analysis saved"
}

# Function to collect block profiling
profile_blocking() {
    echo ""
    echo "5. Blocking Operations Profiling"
    echo "   Collecting blocking profile..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/block" \
        -o "$OUTPUT_DIR/block.prof"

    echo "   ✓ Block profile saved to $OUTPUT_DIR/block.prof"
    go tool pprof -top "$OUTPUT_DIR/block.prof" > "$OUTPUT_DIR/block_top.txt" 2>/dev/null || true
    echo "   ✓ Block analysis saved"
}

# Function to collect mutex profiling
profile_mutex() {
    echo ""
    echo "6. Mutex Contention Profiling"
    echo "   Collecting mutex profile..."

    curl -s "http://localhost:$SERVER_PORT/debug/pprof/mutex" \
        -o "$OUTPUT_DIR/mutex.prof"

    echo "   ✓ Mutex profile saved to $OUTPUT_DIR/mutex.prof"
    go tool pprof -top "$OUTPUT_DIR/mutex.prof" > "$OUTPUT_DIR/mutex_top.txt" 2>/dev/null || true
    echo "   ✓ Mutex analysis saved"
}

# Function to collect application metrics
collect_metrics() {
    echo ""
    echo "7. Application Metrics"
    echo "   Collecting custom metrics..."

    # Baseline metrics
    curl -s "http://localhost:$SERVER_PORT/debug/metrics" \
        > "$OUTPUT_DIR/metrics_baseline.json"
    echo "   ✓ Baseline metrics collected"

    # Run load test
    echo "   Running load test..."
    wrk -t8 -c$LOAD_CONNECTIONS -d${LOAD_DURATION}s --latency \
        -s scripts/load_test.lua \
        "http://localhost:$SERVER_PORT" \
        > "$OUTPUT_DIR/load_test_results.txt" 2>&1 || true

    # Post-load metrics
    sleep 2
    curl -s "http://localhost:$SERVER_PORT/debug/metrics" \
        > "$OUTPUT_DIR/metrics_under_load.json"
    echo "   ✓ Under-load metrics collected"
}

# Function to generate summary report
generate_report() {
    echo ""
    echo "8. Generating Summary Report"

    cat > "$OUTPUT_DIR/PROFILING_SUMMARY.txt" << EOF
URL Shortener - Profiling Summary
Generated: $(date)
=====================================

Configuration:
- Server Port: $SERVER_PORT
- Profile Duration: ${PROFILE_DURATION}s
- Load Test Duration: ${LOAD_DURATION}s
- Concurrent Connections: $LOAD_CONNECTIONS

Files Generated:
- cpu.prof              : CPU profile (view with: go tool pprof -http=:8081 cpu.prof)
- cpu_flamegraph.png    : CPU flame graph visualization
- heap.prof             : Heap memory profile
- heap_visualization.png: Heap usage visualization
- heap_top.txt          : Top memory allocators
- goroutine.prof        : Goroutine profile
- goroutine_analysis.txt: Goroutine stack analysis
- allocs.prof           : Allocation profile
- allocs_top.txt        : Top allocators
- block.prof            : Blocking operations profile
- mutex.prof            : Mutex contention profile
- metrics_baseline.json : Metrics before load
- metrics_under_load.json: Metrics during load
- load_test_results.txt : Load test output

Analysis Commands:
==================

1. Interactive CPU Profile:
   go tool pprof -http=:8081 $OUTPUT_DIR/cpu.prof

2. Interactive Heap Profile:
   go tool pprof -http=:8081 $OUTPUT_DIR/heap.prof

3. View Top CPU Consumers:
   go tool pprof -top $OUTPUT_DIR/cpu.prof

4. View Top Memory Allocators:
   go tool pprof -top $OUTPUT_DIR/heap.prof

5. View Goroutine Stacks:
   go tool pprof -text $OUTPUT_DIR/goroutine.prof

6. Compare Profiles:
   go tool pprof -base=baseline.prof -http=:8081 current.prof

Performance Recommendations:
============================
See PERFORMANCE_PROFILING_REPORT.md for detailed analysis and optimization strategies.

Key Areas to Investigate:
1. CPU hot paths (look for high sample count in cpu.prof)
2. Memory allocations (check heap.prof and allocs.prof)
3. Goroutine leaks (analyze goroutine.prof)
4. Lock contention (review mutex.prof)
5. Blocking operations (check block.prof)

Next Steps:
===========
1. Review generated profiles using 'go tool pprof'
2. Compare metrics_baseline.json vs metrics_under_load.json
3. Identify performance bottlenecks
4. Implement optimizations from PERFORMANCE_PROFILING_REPORT.md
5. Re-run profiling to measure improvements
EOF

    echo "   ✓ Summary report generated"
}

# Main execution
main() {
    check_server

    echo ""
    echo "Starting comprehensive profiling..."
    echo "This will take approximately $((PROFILE_DURATION + LOAD_DURATION + 60)) seconds"
    echo ""

    profile_cpu
    profile_heap
    profile_goroutines
    profile_allocs
    profile_blocking
    profile_mutex
    collect_metrics
    generate_report

    echo ""
    echo "=========================================="
    echo "Profiling Complete!"
    echo "=========================================="
    echo ""
    echo "Results saved to: $OUTPUT_DIR/"
    echo ""
    echo "Quick Start:"
    echo "  View CPU profile:  go tool pprof -http=:8081 $OUTPUT_DIR/cpu.prof"
    echo "  View heap profile: go tool pprof -http=:8081 $OUTPUT_DIR/heap.prof"
    echo "  Read summary:      cat $OUTPUT_DIR/PROFILING_SUMMARY.txt"
    echo ""
}

# Run main function
main
