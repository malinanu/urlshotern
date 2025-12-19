// K6 Load Testing Script for URL Shortener
// Install k6: https://k6.io/docs/getting-started/installation/
// Run: k6 run scripts/load-test.js

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Rate, Trend, Counter } from 'k6/metrics';

// Custom metrics
const errorRate = new Rate('errors');
const redirectTime = new Trend('redirect_time');
const createTime = new Trend('create_time');
const createdURLs = new Counter('created_urls');

// Test configuration
export const options = {
  stages: [
    // Ramp-up: 0 to 100 users over 2 minutes
    { duration: '2m', target: 100 },
    // Steady state: 100 users for 5 minutes
    { duration: '5m', target: 100 },
    // Peak load: 100 to 500 users over 2 minutes
    { duration: '2m', target: 500 },
    // Sustained peak: 500 users for 5 minutes
    { duration: '5m', target: 500 },
    // Spike test: 500 to 1000 users over 1 minute
    { duration: '1m', target: 1000 },
    // Sustained spike: 1000 users for 3 minutes
    { duration: '3m', target: 1000 },
    // Ramp-down: 1000 to 0 users over 2 minutes
    { duration: '2m', target: 0 },
  ],
  thresholds: {
    // 95% of requests must complete below 500ms
    'http_req_duration': ['p(95)<500'],
    // Error rate must be below 5%
    'errors': ['rate<0.05'],
    // 95% of redirects must complete below 100ms (cache hit)
    'redirect_time': ['p(95)<100'],
    // 95% of URL creations must complete below 200ms
    'create_time': ['p(95)<200'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1`;

// Test data
const testURLs = [
  'https://example.com',
  'https://github.com',
  'https://stackoverflow.com',
  'https://reddit.com',
  'https://twitter.com',
];

// Shared array to store created short codes
let shortCodes = [];

export default function () {
  const scenario = Math.random();

  if (scenario < 0.7) {
    // 70% of traffic: URL redirects (read-heavy workload)
    testURLRedirect();
  } else if (scenario < 0.9) {
    // 20% of traffic: Create new URLs
    testCreateURL();
  } else {
    // 10% of traffic: Analytics queries
    testAnalytics();
  }

  sleep(0.1); // Think time between requests
}

function testURLRedirect() {
  if (shortCodes.length === 0) {
    // Create initial URLs if none exist
    const code = createURL();
    if (code) shortCodes.push(code);
    return;
  }

  // Randomly select a short code
  const shortCode = shortCodes[Math.floor(Math.random() * shortCodes.length)];

  const startTime = new Date().getTime();
  const res = http.get(`${BASE_URL}/${shortCode}`, {
    redirects: 0, // Don't follow redirects
    tags: { name: 'URLRedirect' },
  });
  const duration = new Date().getTime() - startTime;

  const success = check(res, {
    'redirect status is 301 or 302': (r) => r.status === 301 || r.status === 302,
    'has location header': (r) => r.headers['Location'] !== undefined,
    'response time < 500ms': (r) => duration < 500,
  });

  errorRate.add(!success);
  redirectTime.add(duration);
}

function testCreateURL() {
  const code = createURL();
  if (code) shortCodes.push(code);

  // Limit array size to prevent memory issues
  if (shortCodes.length > 1000) {
    shortCodes = shortCodes.slice(-1000);
  }
}

function createURL() {
  const url = testURLs[Math.floor(Math.random() * testURLs.length)];
  const payload = JSON.stringify({
    url: `${url}?test=${Date.now()}&rand=${Math.random()}`,
  });

  const params = {
    headers: { 'Content-Type': 'application/json' },
    tags: { name: 'CreateURL' },
  };

  const startTime = new Date().getTime();
  const res = http.post(`${API_URL}/shorten`, payload, params);
  const duration = new Date().getTime() - startTime;

  const success = check(res, {
    'create status is 200': (r) => r.status === 200,
    'has short_code': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.short_code !== undefined;
      } catch (e) {
        return false;
      }
    },
    'response time < 1000ms': (r) => duration < 1000,
  });

  errorRate.add(!success);
  createTime.add(duration);

  if (success) {
    try {
      const body = JSON.parse(res.body);
      createdURLs.add(1);
      return body.short_code;
    } catch (e) {
      return null;
    }
  }

  return null;
}

function testAnalytics() {
  if (shortCodes.length === 0) return;

  const shortCode = shortCodes[Math.floor(Math.random() * shortCodes.length)];

  const res = http.get(`${API_URL}/analytics/${shortCode}`, {
    tags: { name: 'Analytics' },
  });

  const success = check(res, {
    'analytics status is 200': (r) => r.status === 200,
    'has total_clicks': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total_clicks !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  errorRate.add(!success);
}

// Smoke test: Verify basic functionality with minimal load
export function smokeTest() {
  const res = http.get(`${BASE_URL}/health`);
  check(res, {
    'health check status is 200': (r) => r.status === 200,
  });
}

// Stress test: Push system to breaking point
export const stressOptions = {
  stages: [
    { duration: '2m', target: 1000 },
    { duration: '5m', target: 2000 },
    { duration: '2m', target: 3000 },
    { duration: '5m', target: 3000 },
    { duration: '2m', target: 0 },
  ],
};

// Soak test: Sustained load over extended period
export const soakOptions = {
  stages: [
    { duration: '5m', target: 500 },
    { duration: '1h', target: 500 },
    { duration: '5m', target: 0 },
  ],
};

// Setup function: Run once before test
export function setup() {
  console.log('Starting load test...');
  console.log(`Base URL: ${BASE_URL}`);

  // Health check
  const res = http.get(`${BASE_URL}/health`);
  if (res.status !== 200) {
    throw new Error('Health check failed! Server is not ready.');
  }

  console.log('Health check passed. Starting test...');
  return { timestamp: Date.now() };
}

// Teardown function: Run once after test
export function teardown(data) {
  const duration = (Date.now() - data.timestamp) / 1000;
  console.log(`Load test completed. Duration: ${duration}s`);
}
