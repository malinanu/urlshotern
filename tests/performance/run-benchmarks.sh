#!/bin/bash

# Performance Benchmark Runner
# Runs database benchmarks and exports results for CI/CD

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
NC='\033[0m' # No Color

# Configuration
BENCHMARK_TIME=${BENCHMARK_TIME:-"10s"}
BENCHMARK_ITERATIONS=${BENCHMARK_ITERATIONS:-"1000"}
# SECURITY: Default connection for local testing only
# For production benchmarks, set DATABASE_URL environment variable with proper credentials
DATABASE_URL=${DATABASE_URL:-"postgres://urlshortener:devpassword@localhost:5432/urlshortener_db?sslmode=disable"}
OUTPUT_DIR=${OUTPUT_DIR:-"./benchmark-results"}
THRESHOLD_FILE=${THRESHOLD_FILE:-"./performance-thresholds.json"}

echo "========================================="
echo "  Database Performance Benchmark Suite"
echo "========================================="
echo ""

# Create output directory
mkdir -p "$OUTPUT_DIR"

# Function to run a benchmark
run_benchmark() {
    local name=$1
    local test_name=$2

    echo -e "${YELLOW}Running: $name${NC}"

    go test -bench="$test_name" \
        -benchtime="$BENCHMARK_TIME" \
        -benchmem \
        -cpuprofile="$OUTPUT_DIR/${test_name}_cpu.prof" \
        -memprofile="$OUTPUT_DIR/${test_name}_mem.prof" \
        -run=^$ \
        . 2>&1 | tee "$OUTPUT_DIR/${test_name}_output.txt"

    echo ""
}

# Export environment variables
export DATABASE_URL

# Run all benchmarks
echo "Starting benchmark suite..."
echo ""

run_benchmark "URL Insert" "BenchmarkURLInsert"
run_benchmark "URL Lookup" "BenchmarkURLLookup"
run_benchmark "URL Lookup with Expiry" "BenchmarkURLLookupWithExpiry"
run_benchmark "Click Insert" "BenchmarkClickInsert"
run_benchmark "Analytics Query" "BenchmarkAnalyticsQuery"
run_benchmark "User URLs List" "BenchmarkUserURLsList"
run_benchmark "Concurrent Lookups" "BenchmarkConcurrentLookups"
run_benchmark "Concurrent Inserts" "BenchmarkConcurrentInserts"
run_benchmark "Indexed vs Sequential" "BenchmarkIndexedVsSequentialScan"

# Generate summary report
echo -e "${GREEN}Generating summary report...${NC}"

cat > "$OUTPUT_DIR/summary.md" << 'EOF'
# Database Performance Benchmark Results

## Test Configuration
- Benchmark Time:
EOF

echo "- Benchmark Time: $BENCHMARK_TIME" >> "$OUTPUT_DIR/summary.md"
echo "- Database: PostgreSQL" >> "$OUTPUT_DIR/summary.md"
echo "- Date: $(date)" >> "$OUTPUT_DIR/summary.md"
echo "" >> "$OUTPUT_DIR/summary.md"

echo "## Results" >> "$OUTPUT_DIR/summary.md"
echo "" >> "$OUTPUT_DIR/summary.md"

# Parse and add results
for file in "$OUTPUT_DIR"/*_output.txt; do
    if [ -f "$file" ]; then
        echo "### $(basename "$file" _output.txt)" >> "$OUTPUT_DIR/summary.md"
        echo '```' >> "$OUTPUT_DIR/summary.md"
        grep -E "^Benchmark|ns/op|B/op|allocs/op" "$file" >> "$OUTPUT_DIR/summary.md" || true
        echo '```' >> "$OUTPUT_DIR/summary.md"
        echo "" >> "$OUTPUT_DIR/summary.md"
    fi
done

# Check against thresholds
echo -e "${YELLOW}Checking performance thresholds...${NC}"

check_threshold() {
    local benchmark=$1
    local threshold=$2
    local result_file="$OUTPUT_DIR/${benchmark}_output.txt"

    if [ -f "$result_file" ]; then
        # Extract ns/op value
        ns_per_op=$(grep "ns/op" "$result_file" | awk '{print $3}' | head -1)

        if [ -n "$ns_per_op" ]; then
            # Remove any non-numeric characters
            ns_per_op=$(echo "$ns_per_op" | tr -dc '0-9.')

            # Compare with threshold
            if (( $(echo "$ns_per_op > $threshold" | bc -l) )); then
                echo -e "${RED}FAIL: $benchmark exceeded threshold ($ns_per_op ns/op > $threshold ns/op)${NC}"
                return 1
            else
                echo -e "${GREEN}PASS: $benchmark within threshold ($ns_per_op ns/op <= $threshold ns/op)${NC}"
                return 0
            fi
        fi
    fi

    echo -e "${YELLOW}SKIP: Could not find results for $benchmark${NC}"
    return 0
}

# Performance thresholds (in nanoseconds)
# These should be adjusted based on your requirements
failed=0

# P95 < 100ms for database queries
check_threshold "BenchmarkURLLookup" 100000000 || ((failed++))
check_threshold "BenchmarkURLLookupWithExpiry" 100000000 || ((failed++))
check_threshold "BenchmarkAnalyticsQuery" 100000000 || ((failed++))

# Inserts should be fast
check_threshold "BenchmarkURLInsert" 50000000 || ((failed++))
check_threshold "BenchmarkClickInsert" 50000000 || ((failed++))

echo ""
echo "========================================="
echo "  Benchmark Suite Complete"
echo "========================================="
echo ""
echo "Results saved to: $OUTPUT_DIR"
echo ""

if [ $failed -gt 0 ]; then
    echo -e "${RED}$failed benchmark(s) exceeded performance thresholds${NC}"
    exit 1
else
    echo -e "${GREEN}All benchmarks passed performance thresholds${NC}"
    exit 0
fi
