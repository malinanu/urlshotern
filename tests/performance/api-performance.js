import http from 'k6/http';
import { check, group, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';
import { randomString, randomItem } from 'https://jslib.k6.io/k6-utils/1.4.0/index.js';

// Custom metrics
const errorRate = new Rate('error_rate');
const apiLatency = new Trend('api_latency');
const shortenLatency = new Trend('shorten_latency');
const redirectLatency = new Trend('redirect_latency');
const analyticsLatency = new Trend('analytics_latency');
const cacheHitRate = new Rate('cache_hit_rate');
const failedRequests = new Counter('failed_requests');

// Test configuration
export const options = {
  scenarios: {
    // Baseline load test
    baseline: {
      executor: 'constant-vus',
      vus: 10,
      duration: '2m',
      gracefulStop: '30s',
      tags: { test_type: 'baseline' },
    },

    // Spike test
    spike: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '30s', target: 50 },
        { duration: '1m', target: 50 },
        { duration: '30s', target: 200 },
        { duration: '1m', target: 200 },
        { duration: '30s', target: 0 },
      ],
      gracefulStop: '30s',
      tags: { test_type: 'spike' },
      startTime: '2m30s',
    },

    // Soak test
    soak: {
      executor: 'constant-vus',
      vus: 20,
      duration: '5m',
      gracefulStop: '30s',
      tags: { test_type: 'soak' },
      startTime: '6m',
    },

    // Stress test
    stress: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '2m', target: 50 },
        { duration: '5m', target: 50 },
        { duration: '2m', target: 100 },
        { duration: '5m', target: 100 },
        { duration: '2m', target: 200 },
        { duration: '5m', target: 200 },
        { duration: '2m', target: 0 },
      ],
      gracefulStop: '30s',
      tags: { test_type: 'stress' },
      startTime: '11m30s',
    },
  },

  // Performance thresholds
  thresholds: {
    // Overall HTTP metrics
    http_req_duration: [
      'p(95)<1000', // 95% of requests should be below 1s
      'p(99)<2000', // 99% of requests should be below 2s
      'avg<500',    // Average should be below 500ms
    ],
    http_req_failed: ['rate<0.01'], // Error rate should be below 1%

    // Custom metrics
    error_rate: ['rate<0.01'],
    'shorten_latency': ['p(95)<800', 'p(99)<1500'],
    'redirect_latency': ['p(95)<200', 'p(99)<400'],
    'analytics_latency': ['p(95)<500', 'p(99)<1000'],
    'cache_hit_rate': ['rate>0.9'], // 90% cache hit rate
    'failed_requests': ['count<50'],

    // HTTP specific
    'http_req_connecting': ['p(95)<100'],
    'http_req_tls_handshaking': ['p(95)<200'],
    'http_req_sending': ['p(95)<10'],
    'http_req_waiting': ['p(95)<800'],
    'http_req_receiving': ['p(95)<10'],
  },

  // Graceful ramp-down period
  teardownTimeout: '60s',
};

// Base URL from environment or default
const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_BASE = `${BASE_URL}/api/v1`;

// Store created short codes for reuse
let createdShortCodes = [];

// Setup function - runs once per VU
export function setup() {
  // Create some initial URLs for testing
  const initialUrls = [];

  for (let i = 0; i < 100; i++) {
    const payload = JSON.stringify({
      url: `https://example.com/page-${randomString(10)}`,
    });

    const res = http.post(`${API_BASE}/shorten`, payload, {
      headers: { 'Content-Type': 'application/json' },
    });

    if (res.status === 201) {
      const data = JSON.parse(res.body);
      initialUrls.push(data.short_code);
    }
  }

  return { initialUrls };
}

// Main test function
export default function (data) {
  // Add initial URLs to shared array
  if (data.initialUrls && data.initialUrls.length > 0) {
    createdShortCodes = data.initialUrls;
  }

  group('URL Shortening API', function () {
    testShortenURL();
    sleep(0.5);
  });

  group('URL Redirection', function () {
    testRedirection();
    sleep(0.3);
  });

  group('Analytics API', function () {
    testAnalytics();
    sleep(0.5);
  });

  group('Health Check', function () {
    testHealthCheck();
    sleep(0.2);
  });

  // Random sleep between iterations
  sleep(Math.random() * 2 + 1);
}

function testShortenURL() {
  const payload = JSON.stringify({
    url: `https://example.com/test-${randomString(10)}`,
    custom_code: Math.random() > 0.8 ? randomString(6) : undefined,
  });

  const startTime = new Date().getTime();
  const res = http.post(`${API_BASE}/shorten`, payload, {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'ShortenURL' },
  });
  const duration = new Date().getTime() - startTime;

  shortenLatency.add(duration);
  apiLatency.add(duration);

  const success = check(res, {
    'status is 201': (r) => r.status === 201,
    'has short_code': (r) => {
      try {
        const data = JSON.parse(r.body);
        return data.short_code !== undefined;
      } catch {
        return false;
      }
    },
    'has short_url': (r) => {
      try {
        const data = JSON.parse(r.body);
        return data.short_url !== undefined;
      } catch {
        return false;
      }
    },
    'response time < 1s': (r) => r.timings.duration < 1000,
  });

  if (!success) {
    errorRate.add(1);
    failedRequests.add(1);
  } else {
    errorRate.add(0);
    // Store the short code for later use
    try {
      const data = JSON.parse(res.body);
      createdShortCodes.push(data.short_code);
    } catch (e) {
      // Ignore parsing errors
    }
  }
}

function testRedirection() {
  if (createdShortCodes.length === 0) {
    return;
  }

  const shortCode = randomItem(createdShortCodes);

  const startTime = new Date().getTime();
  const res = http.get(`${BASE_URL}/${shortCode}`, {
    tags: { name: 'Redirect' },
    redirects: 0, // Don't follow redirects
  });
  const duration = new Date().getTime() - startTime;

  redirectLatency.add(duration);
  apiLatency.add(duration);

  // Check if response came from cache (custom header)
  const fromCache = res.headers['X-Cache-Hit'] === 'true';
  cacheHitRate.add(fromCache ? 1 : 0);

  const success = check(res, {
    'status is 301 or 302': (r) => r.status === 301 || r.status === 302,
    'has Location header': (r) => r.headers['Location'] !== undefined,
    'response time < 200ms': (r) => r.timings.duration < 200,
  });

  if (!success) {
    errorRate.add(1);
    failedRequests.add(1);
  } else {
    errorRate.add(0);
  }
}

function testAnalytics() {
  if (createdShortCodes.length === 0) {
    return;
  }

  const shortCode = randomItem(createdShortCodes);

  const startTime = new Date().getTime();
  const res = http.get(`${API_BASE}/analytics/${shortCode}`, {
    tags: { name: 'Analytics' },
  });
  const duration = new Date().getTime() - startTime;

  analyticsLatency.add(duration);
  apiLatency.add(duration);

  const success = check(res, {
    'status is 200': (r) => r.status === 200,
    'has analytics data': (r) => {
      try {
        const data = JSON.parse(r.body);
        return data.total_clicks !== undefined;
      } catch {
        return false;
      }
    },
    'response time < 500ms': (r) => r.timings.duration < 500,
  });

  if (!success) {
    errorRate.add(1);
    failedRequests.add(1);
  } else {
    errorRate.add(0);
  }
}

function testHealthCheck() {
  const res = http.get(`${BASE_URL}/health`, {
    tags: { name: 'HealthCheck' },
  });

  check(res, {
    'status is 200': (r) => r.status === 200,
    'response time < 100ms': (r) => r.timings.duration < 100,
  });
}

// Teardown function - runs once after all VUs complete
export function teardown(data) {
  // Optional: cleanup test data
  console.log('Test completed successfully');
}

// Handle summary data
export function handleSummary(data) {
  return {
    'performance-report.json': JSON.stringify(data, null, 2),
    stdout: textSummary(data, { indent: ' ', enableColors: true }),
  };
}

function textSummary(data, options) {
  const indent = options.indent || '';
  const enableColors = options.enableColors || false;

  let summary = '\n';
  summary += `${indent}========================================\n`;
  summary += `${indent}  Performance Test Summary\n`;
  summary += `${indent}========================================\n\n`;

  // Add test metrics
  if (data.metrics) {
    summary += `${indent}Metrics:\n`;
    summary += `${indent}  HTTP Requests: ${data.metrics.http_reqs?.values?.count || 0}\n`;
    summary += `${indent}  Failed Requests: ${data.metrics.http_req_failed?.values?.rate ? (data.metrics.http_req_failed.values.rate * 100).toFixed(2) : 0}%\n`;
    summary += `${indent}  Avg Duration: ${data.metrics.http_req_duration?.values?.avg?.toFixed(2) || 0}ms\n`;
    summary += `${indent}  P95 Duration: ${data.metrics.http_req_duration?.values['p(95)']?.toFixed(2) || 0}ms\n`;
    summary += `${indent}  P99 Duration: ${data.metrics.http_req_duration?.values['p(99)']?.toFixed(2) || 0}ms\n`;
  }

  summary += '\n';
  return summary;
}
