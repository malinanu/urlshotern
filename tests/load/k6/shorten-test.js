/**
 * Load Test: URL Shortening Performance
 *
 * Tests URL creation endpoint which involves:
 * - ID generation (Snowflake)
 * - Base62 encoding
 * - Database writes
 * - Cache writes
 *
 * Expected Performance Targets:
 * - P95 latency: < 200ms
 * - P99 latency: < 500ms
 * - Throughput: 500-1,000 RPS
 * - Error rate: < 0.5%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';
import { randomString } from 'https://jslib.k6.io/k6-utils/1.2.0/index.js';

// Custom metrics
const shortenSuccessRate = new Rate('shorten_success_rate');
const shortenLatency = new Trend('shorten_latency_ms');
const duplicateErrors = new Counter('duplicate_errors');
const validationErrors = new Counter('validation_errors');

export const options = {
  scenarios: {
    // Warm-up
    warmup: {
      executor: 'ramping-vus',
      startVUs: 0,
      stages: [
        { duration: '1m', target: 50 },
      ],
      gracefulRampDown: '30s',
      startTime: '0s',
      exec: 'shortenTest',
    },

    // Normal load: baseline URL creation traffic
    normal_load: {
      executor: 'constant-arrival-rate',
      rate: 500,        // 500 URLs per second
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 100,
      maxVUs: 200,
      startTime: '1m30s',
      exec: 'shortenTest',
    },

    // Peak load: high URL creation rate
    peak_load: {
      executor: 'ramping-arrival-rate',
      startRate: 500,
      timeUnit: '1s',
      preAllocatedVUs: 200,
      maxVUs: 500,
      stages: [
        { duration: '2m', target: 1000 },   // Ramp to 1k/s
        { duration: '8m', target: 1000 },   // Sustain
        { duration: '2m', target: 500 },    // Ramp down
      ],
      startTime: '11m30s',
      exec: 'shortenTest',
    },

    // Stress test: push beyond expected capacity
    stress_test: {
      executor: 'ramping-arrival-rate',
      startRate: 1000,
      timeUnit: '1s',
      preAllocatedVUs: 500,
      maxVUs: 1000,
      stages: [
        { duration: '2m', target: 2000 },   // 2k/s
        { duration: '5m', target: 2000 },
        { duration: '2m', target: 3000 },   // 3k/s
        { duration: '3m', target: 3000 },
      ],
      startTime: '23m30s',
      exec: 'shortenTest',
    },

    // Batch operations test
    batch_test: {
      executor: 'constant-vus',
      vus: 50,
      duration: '5m',
      startTime: '35m30s',
      exec: 'batchShortenTest',
    },
  },

  thresholds: {
    'http_req_duration{name:shorten}': [
      'p(50)<100',
      'p(95)<200',
      'p(99)<500',
    ],
    'shorten_success_rate': ['rate>0.995'],
    'http_req_failed{name:shorten}': ['rate<0.005'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';
const API_URL = `${BASE_URL}/api/v1/shorten`;

// Test URL sources for more realistic data
const TEST_DOMAINS = [
  'https://example.com',
  'https://github.com',
  'https://stackoverflow.com',
  'https://medium.com',
  'https://dev.to',
  'https://youtube.com',
  'https://twitter.com',
  'https://reddit.com',
];

export function shortenTest() {
  // Generate unique URL to shorten
  const domain = TEST_DOMAINS[Math.floor(Math.random() * TEST_DOMAINS.length)];
  const path = randomString(10);
  const url = `${domain}/${path}`;

  const payload = JSON.stringify({
    url: url,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'shorten' },
  };

  const startTime = new Date().getTime();
  const response = http.post(API_URL, payload, params);
  const duration = new Date().getTime() - startTime;

  shortenLatency.add(duration);

  const success = check(response, {
    'status is 201': (r) => r.status === 201,
    'has short_code': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.short_code !== undefined;
      } catch (e) {
        return false;
      }
    },
    'has short_url': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.short_url !== undefined;
      } catch (e) {
        return false;
      }
    },
    'latency < 500ms': () => duration < 500,
  });

  shortenSuccessRate.add(success);

  if (!success) {
    if (response.status === 409) {
      duplicateErrors.add(1);
    } else if (response.status === 400) {
      validationErrors.add(1);
    }
    console.error(`Shorten failed: ${response.status} - ${response.body}`);
  }

  sleep(0.1);
}

// Test custom short codes
export function customCodeTest() {
  const customCode = `custom_${randomString(6)}`;
  const domain = TEST_DOMAINS[Math.floor(Math.random() * TEST_DOMAINS.length)];
  const url = `${domain}/${randomString(10)}`;

  const payload = JSON.stringify({
    url: url,
    custom_code: customCode,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'shorten_custom' },
  };

  const response = http.post(API_URL, payload, params);

  check(response, {
    'status is 201': (r) => r.status === 201,
    'uses custom code': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.short_code === customCode;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(0.1);
}

// Test batch URL shortening
export function batchShortenTest() {
  const batchSize = 10;
  const urls = [];

  for (let i = 0; i < batchSize; i++) {
    const domain = TEST_DOMAINS[Math.floor(Math.random() * TEST_DOMAINS.length)];
    const path = randomString(10);
    urls.push({
      url: `${domain}/${path}`,
    });
  }

  const payload = JSON.stringify(urls);
  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'batch_shorten' },
  };

  const response = http.post(`${BASE_URL}/api/v1/batch/shorten`, payload, params);

  check(response, {
    'batch status is 200': (r) => r.status === 200,
    'batch has results': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.results !== undefined && body.results.length > 0;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(1);
}

// Test URL shortening with expiration
export function expirationTest() {
  const domain = TEST_DOMAINS[Math.floor(Math.random() * TEST_DOMAINS.length)];
  const url = `${domain}/${randomString(10)}`;
  const expiresAt = new Date(Date.now() + 3600000).toISOString(); // 1 hour from now

  const payload = JSON.stringify({
    url: url,
    expires_at: expiresAt,
  });

  const params = {
    headers: {
      'Content-Type': 'application/json',
    },
    tags: { name: 'shorten_expiration' },
  };

  const response = http.post(API_URL, payload, params);

  check(response, {
    'status is 201': (r) => r.status === 201,
    'has expiration': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.expires_at !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(0.1);
}

export function setup() {
  console.log('Starting URL shortening load test...');
  console.log(`API URL: ${API_URL}`);

  // Test that the API is accessible
  const testPayload = JSON.stringify({
    url: 'https://example.com/test',
  });

  const response = http.post(API_URL, testPayload, {
    headers: { 'Content-Type': 'application/json' },
  });

  if (response.status !== 201) {
    console.error('WARNING: Shorten API may not be accessible!');
    console.error(`Expected 201, got ${response.status}`);
  }

  return { timestamp: new Date().toISOString() };
}

export function teardown(data) {
  console.log(`Test completed at: ${new Date().toISOString()}`);
  console.log(`Test started at: ${data.timestamp}`);
}
