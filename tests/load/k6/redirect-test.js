/**
 * Load Test: URL Redirection Performance
 *
 * This test validates the performance of URL redirections,
 * which is the most critical path in the URL shortener.
 *
 * Expected Performance Targets:
 * - P95 latency: < 50ms (with Redis cache)
 * - P99 latency: < 100ms
 * - Throughput: 10,000+ RPS
 * - Error rate: < 0.1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const cacheHitRate = new Rate('cache_hit_rate');
const redirectSuccessRate = new Rate('redirect_success_rate');
const redirectLatency = new Trend('redirect_latency_ms');
const errorCounter = new Counter('errors');

// Test configuration
export const options = {
  scenarios: {
    // Warm-up phase: gradual ramp-up to prepare caches
    warmup: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 100 },  // Ramp up to 100 VUs
      ],
      gracefulRampDown: '30s',
      startTime: '0s',
      exec: 'redirectTest',
    },

    // Normal load: sustained baseline traffic
    normal_load: {
      executor: 'constant-vus',
      vus: 200,
      duration: '10m',
      startTime: '2m30s',  // Start after warmup
      exec: 'redirectTest',
    },

    // Peak load: simulate peak traffic hours
    peak_load: {
      executor: 'ramping-vus',
      startVUs: 200,
      stages: [
        { duration: '2m', target: 1000 },   // Ramp to peak
        { duration: '10m', target: 1000 },  // Sustain peak
        { duration: '2m', target: 200 },    // Ramp down
      ],
      startTime: '12m30s',
      exec: 'redirectTest',
    },

    // Stress test: push to breaking point
    stress_test: {
      executor: 'ramping-arrival-rate',
      startRate: 1000,
      timeUnit: '1s',
      preAllocatedVUs: 2000,
      maxVUs: 5000,
      stages: [
        { duration: '2m', target: 5000 },   // Ramp to 5k RPS
        { duration: '5m', target: 5000 },   // Sustain
        { duration: '2m', target: 10000 },  // Push to 10k RPS
        { duration: '5m', target: 10000 },  // Sustain if possible
        { duration: '2m', target: 15000 },  // Push beyond target
        { duration: '3m', target: 15000 },  // Breaking point discovery
      ],
      startTime: '24m30s',
      exec: 'redirectTest',
    },

    // Spike test: sudden traffic spikes
    spike_test: {
      executor: 'ramping-vus',
      startVUs: 200,
      stages: [
        { duration: '30s', target: 200 },
        { duration: '30s', target: 2000 },  // Sudden spike
        { duration: '2m', target: 2000 },
        { duration: '30s', target: 200 },   // Back to normal
        { duration: '2m', target: 200 },
        { duration: '30s', target: 3000 },  // Larger spike
        { duration: '2m', target: 3000 },
        { duration: '30s', target: 200 },
      ],
      startTime: '43m30s',
      exec: 'redirectTest',
    },
  },

  thresholds: {
    // Response time thresholds
    'http_req_duration': [
      'p(50)<30',   // 50% of requests should be below 30ms
      'p(95)<50',   // 95% of requests should be below 50ms
      'p(99)<100',  // 99% of requests should be below 100ms
    ],

    // Throughput thresholds
    'http_reqs': ['rate>1000'], // Minimum 1000 RPS

    // Error thresholds
    'redirect_success_rate': ['rate>0.999'], // 99.9% success rate
    'http_req_failed': ['rate<0.001'],       // Less than 0.1% errors

    // Cache performance
    'cache_hit_rate': ['rate>0.90'], // 90%+ cache hit rate expected
  },
};

// Load test data - pre-created short codes
const SHORT_CODES = [
  'abc123',
  'xyz789',
  'test01',
  'demo99',
  'sample',
  // Add more from your test data
];

// Base URL from environment or default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

export function redirectTest() {
  // Select random short code (simulates cache hits and misses)
  const shortCode = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
  const url = `${BASE_URL}/${shortCode}`;

  const startTime = new Date().getTime();

  // Make redirect request with follow redirects disabled to measure redirect time only
  const response = http.get(url, {
    redirects: 0,  // Don't follow redirects to measure redirect latency
    tags: { name: 'redirect' },
  });

  const endTime = new Date().getTime();
  const duration = endTime - startTime;

  // Record metrics
  redirectLatency.add(duration);

  // Check response
  const success = check(response, {
    'status is 301': (r) => r.status === 301,
    'has location header': (r) => r.headers['Location'] !== undefined,
    'redirect latency < 100ms': (r) => duration < 100,
  });

  redirectSuccessRate.add(success);

  if (!success) {
    errorCounter.add(1);
    console.error(`Redirect failed for ${shortCode}: ${response.status} - ${response.body}`);
  }

  // Check for cache indicator (custom header if implemented)
  if (response.headers['X-Cache-Hit'] === 'true') {
    cacheHitRate.add(1);
  } else {
    cacheHitRate.add(0);
  }

  // Small sleep to simulate real user behavior (optional)
  sleep(0.1);
}

// Setup function - runs once before test
export function setup() {
  console.log('Starting redirect load test...');
  console.log(`Base URL: ${BASE_URL}`);
  console.log(`Test short codes: ${SHORT_CODES.length}`);

  // Verify test data exists
  const testUrl = `${BASE_URL}/${SHORT_CODES[0]}`;
  const response = http.get(testUrl, { redirects: 0 });

  if (response.status !== 301) {
    console.error('WARNING: Test data may not be set up correctly!');
    console.error(`Expected 301, got ${response.status} for ${testUrl}`);
  }

  return { timestamp: new Date().toISOString() };
}

// Teardown function - runs once after test
export function teardown(data) {
  console.log(`Test completed at: ${new Date().toISOString()}`);
  console.log(`Test started at: ${data.timestamp}`);
}
