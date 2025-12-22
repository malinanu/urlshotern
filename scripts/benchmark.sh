#!/bin/bash

# Performance Benchmark Script for URL Shortener
# Tests various optimization scenarios and measures improvements

set -e

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Configuration
BASE_URL="${BASE_URL:-http://localhost:8080}"
API_URL="${BASE_URL}/api/v1"
ITERATIONS="${ITERATIONS:-1000}"
CONCURRENCY="${CONCURRENCY:-50}"

echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}URL Shortener Performance Benchmark${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo -e "Base URL: ${GREEN}${BASE_URL}${NC}"
echo -e "Iterations: ${GREEN}${ITERATIONS}${NC}"
echo -e "Concurrency: ${GREEN}${CONCURRENCY}${NC}"
echo ""

# Check if required tools are installed
command -v curl >/dev/null 2>&1 || { echo -e "${RED}curl is required but not installed.${NC}" >&2; exit 1; }
command -v jq >/dev/null 2>&1 || { echo -e "${RED}jq is required but not installed.${NC}" >&2; exit 1; }
command -v ab >/dev/null 2>&1 || { echo -e "${YELLOW}Apache Bench (ab) not found. Some tests will be skipped.${NC}"; }

# Health check
echo -e "${YELLOW}Performing health check...${NC}"
if curl -sf "${BASE_URL}/health" >/dev/null; then
    echo -e "${GREEN}✓ Server is healthy${NC}"
else
    echo -e "${RED}✗ Health check failed. Is the server running?${NC}"
    exit 1
fi
echo ""

# Test 1: URL Creation Performance
echo -e "${BLUE}Test 1: URL Creation Performance${NC}"
echo "Creating ${ITERATIONS} URLs..."

START_TIME=$(date +%s.%N)
for i in $(seq 1 ${ITERATIONS}); do
    SHORT_CODE=$(curl -sf -X POST "${API_URL}/shorten" \
        -H "Content-Type: application/json" \
        -d "{\"url\":\"https://example.com/test${i}\"}" | jq -r '.short_code')

    if [ -z "$SHORT_CODE" ]; then
        echo -e "${RED}✗ Failed to create URL${NC}"
        break
    fi

    # Store some short codes for later tests
    if [ $i -le 100 ]; then
        echo "$SHORT_CODE" >> /tmp/short_codes.txt
    fi

    # Progress indicator
    if [ $((i % 100)) -eq 0 ]; then
        echo -ne "\rProgress: ${i}/${ITERATIONS}"
    fi
done
END_TIME=$(date +%s.%N)

ELAPSED=$(echo "$END_TIME - $START_TIME" | bc)
RPS=$(echo "scale=2; $ITERATIONS / $ELAPSED" | bc)

echo ""
echo -e "${GREEN}✓ Created ${ITERATIONS} URLs in ${ELAPSED}s${NC}"
echo -e "  Throughput: ${GREEN}${RPS} req/s${NC}"
echo ""

# Test 2: Cache Hit Performance (URL Redirects)
echo -e "${BLUE}Test 2: Cache Hit Performance${NC}"
echo "Testing redirect performance with cached URLs..."

if [ ! -f /tmp/short_codes.txt ]; then
    echo -e "${RED}✗ No short codes available${NC}"
else
    TOTAL_TIME=0
    COUNT=0

    while IFS= read -r SHORT_CODE; do
        START=$(date +%s.%N)
        HTTP_CODE=$(curl -sf -o /dev/null -w "%{http_code}" "${BASE_URL}/${SHORT_CODE}")
        END=$(date +%s.%N)

        DURATION=$(echo "($END - $START) * 1000" | bc)
        TOTAL_TIME=$(echo "$TOTAL_TIME + $DURATION" | bc)
        COUNT=$((COUNT + 1))

        if [ "$HTTP_CODE" != "301" ] && [ "$HTTP_CODE" != "302" ]; then
            echo -e "${RED}✗ Unexpected status code: ${HTTP_CODE}${NC}"
        fi
    done < /tmp/short_codes.txt

    AVG_TIME=$(echo "scale=2; $TOTAL_TIME / $COUNT" | bc)
    echo -e "${GREEN}✓ Average redirect time: ${AVG_TIME}ms${NC}"

    if [ $(echo "$AVG_TIME < 10" | bc) -eq 1 ]; then
        echo -e "  ${GREEN}Excellent! Cache is working effectively.${NC}"
    elif [ $(echo "$AVG_TIME < 50" | bc) -eq 1 ]; then
        echo -e "  ${YELLOW}Good performance, but could be improved.${NC}"
    else
        echo -e "  ${RED}Slow response time. Check cache configuration.${NC}"
    fi
fi
echo ""

# Test 3: Concurrent Request Handling
echo -e "${BLUE}Test 3: Concurrent Request Handling${NC}"
echo "Testing with ${CONCURRENCY} concurrent connections..."

if command -v ab >/dev/null 2>&1; then
    # Create a temporary file with JSON payload
    echo '{"url":"https://example.com/concurrent-test"}' > /tmp/payload.json

    ab -n ${ITERATIONS} -c ${CONCURRENCY} -p /tmp/payload.json \
        -T "application/json" "${API_URL}/shorten" > /tmp/ab_results.txt 2>&1

    REQUESTS_PER_SEC=$(grep "Requests per second" /tmp/ab_results.txt | awk '{print $4}')
    TIME_PER_REQUEST=$(grep "Time per request" /tmp/ab_results.txt | head -1 | awk '{print $4}')
    TRANSFER_RATE=$(grep "Transfer rate" /tmp/ab_results.txt | awk '{print $3}')

    echo -e "${GREEN}✓ Concurrent test completed${NC}"
    echo -e "  Requests/sec: ${GREEN}${REQUESTS_PER_SEC}${NC}"
    echo -e "  Time/request: ${GREEN}${TIME_PER_REQUEST}ms${NC}"
    echo -e "  Transfer rate: ${GREEN}${TRANSFER_RATE} KB/sec${NC}"

    rm -f /tmp/payload.json /tmp/ab_results.txt
else
    echo -e "${YELLOW}Skipping concurrent test (ab not installed)${NC}"
fi
echo ""

# Test 4: Database Connection Pool Efficiency
echo -e "${BLUE}Test 4: Database Connection Pool Test${NC}"
echo "Testing connection pool under load..."

POOL_TEST_COUNT=500
CONCURRENT_CONN=100

START_TIME=$(date +%s.%N)
for i in $(seq 1 ${POOL_TEST_COUNT}); do
    curl -sf "${BASE_URL}/health" >/dev/null &

    # Limit concurrent requests
    if [ $((i % CONCURRENT_CONN)) -eq 0 ]; then
        wait
    fi
done
wait
END_TIME=$(date +%s.%N)

ELAPSED=$(echo "$END_TIME - $START_TIME" | bc)
echo -e "${GREEN}✓ ${POOL_TEST_COUNT} concurrent health checks completed in ${ELAPSED}s${NC}"
echo ""

# Test 5: Cache Performance Comparison
echo -e "${BLUE}Test 5: Cache Hit vs Cache Miss Comparison${NC}"

# Create a new URL (cache miss)
NEW_URL="https://example.com/cache-test-$(date +%s)"
RESPONSE=$(curl -sf -X POST "${API_URL}/shorten" \
    -H "Content-Type: application/json" \
    -d "{\"url\":\"${NEW_URL}\"}")
NEW_SHORT_CODE=$(echo "$RESPONSE" | jq -r '.short_code')

# Test cache miss (first access)
START=$(date +%s.%N)
curl -sf -o /dev/null "${BASE_URL}/${NEW_SHORT_CODE}"
END=$(date +%s.%N)
MISS_TIME=$(echo "($END - $START) * 1000" | bc)

# Test cache hit (subsequent access)
sleep 0.1
START=$(date +%s.%N)
curl -sf -o /dev/null "${BASE_URL}/${NEW_SHORT_CODE}"
END=$(date +%s.%N)
HIT_TIME=$(echo "($END - $START) * 1000" | bc)

IMPROVEMENT=$(echo "scale=2; ($MISS_TIME - $HIT_TIME) / $MISS_TIME * 100" | bc)

echo -e "  Cache miss: ${YELLOW}${MISS_TIME}ms${NC}"
echo -e "  Cache hit:  ${GREEN}${HIT_TIME}ms${NC}"
echo -e "  Improvement: ${GREEN}${IMPROVEMENT}%${NC}"
echo ""

# Test 6: Response Time Distribution
echo -e "${BLUE}Test 6: Response Time Distribution (P50, P95, P99)${NC}"

SAMPLE_SIZE=1000
declare -a RESPONSE_TIMES

echo "Collecting ${SAMPLE_SIZE} samples..."

for i in $(seq 1 ${SAMPLE_SIZE}); do
    if [ -f /tmp/short_codes.txt ]; then
        SHORT_CODE=$(shuf -n 1 /tmp/short_codes.txt)
        START=$(date +%s.%N)
        curl -sf -o /dev/null "${BASE_URL}/${SHORT_CODE}"
        END=$(date +%s.%N)
        DURATION=$(echo "($END - $START) * 1000" | bc)
        RESPONSE_TIMES+=($DURATION)
    fi

    if [ $((i % 100)) -eq 0 ]; then
        echo -ne "\rProgress: ${i}/${SAMPLE_SIZE}"
    fi
done

echo ""

# Sort response times
IFS=$'\n' SORTED_TIMES=($(sort -n <<<"${RESPONSE_TIMES[*]}"))
unset IFS

# Calculate percentiles
P50_INDEX=$((SAMPLE_SIZE * 50 / 100))
P95_INDEX=$((SAMPLE_SIZE * 95 / 100))
P99_INDEX=$((SAMPLE_SIZE * 99 / 100))

P50=${SORTED_TIMES[$P50_INDEX]}
P95=${SORTED_TIMES[$P95_INDEX]}
P99=${SORTED_TIMES[$P99_INDEX]}

echo -e "  P50 (median): ${GREEN}${P50}ms${NC}"
echo -e "  P95: ${GREEN}${P95}ms${NC}"
echo -e "  P99: ${GREEN}${P99}ms${NC}"
echo ""

# Summary
echo -e "${BLUE}=====================================${NC}"
echo -e "${BLUE}Benchmark Summary${NC}"
echo -e "${BLUE}=====================================${NC}"
echo ""
echo -e "URL Creation:"
echo -e "  Throughput: ${GREEN}${RPS} req/s${NC}"
echo ""
echo -e "Cache Performance:"
echo -e "  Hit latency: ${GREEN}${HIT_TIME}ms${NC}"
echo -e "  Miss latency: ${YELLOW}${MISS_TIME}ms${NC}"
echo -e "  Cache benefit: ${GREEN}${IMPROVEMENT}%${NC}"
echo ""
echo -e "Response Time Distribution:"
echo -e "  P50: ${GREEN}${P50}ms${NC}"
echo -e "  P95: ${GREEN}${P95}ms${NC}"
echo -e "  P99: ${GREEN}${P99}ms${NC}"
echo ""

# Performance Rating
if [ $(echo "$P95 < 50" | bc) -eq 1 ] && [ $(echo "$RPS > 100" | bc) -eq 1 ]; then
    echo -e "${GREEN}Overall Rating: EXCELLENT ⭐⭐⭐⭐⭐${NC}"
elif [ $(echo "$P95 < 100" | bc) -eq 1 ] && [ $(echo "$RPS > 50" | bc) -eq 1 ]; then
    echo -e "${GREEN}Overall Rating: GOOD ⭐⭐⭐⭐${NC}"
elif [ $(echo "$P95 < 200" | bc) -eq 1 ] && [ $(echo "$RPS > 20" | bc) -eq 1 ]; then
    echo -e "${YELLOW}Overall Rating: FAIR ⭐⭐⭐${NC}"
else
    echo -e "${RED}Overall Rating: NEEDS IMPROVEMENT ⭐⭐${NC}"
fi
echo ""

# Cleanup
rm -f /tmp/short_codes.txt

echo -e "${GREEN}Benchmark completed!${NC}"
