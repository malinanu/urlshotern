/**
 * Load Test: Analytics Performance
 *
 * Tests analytics endpoint performance under various query patterns.
 *
 * Expected Performance Targets:
 * - P95 latency: < 300ms
 * - P99 latency: < 1000ms
 * - Throughput: 100-300 RPS
 * - Error rate: < 1%
 */

import http from 'k6/http';
import { check, sleep } from 'k6';
import { Counter, Rate, Trend } from 'k6/metrics';

// Custom metrics
const analyticsSuccessRate = new Rate('analytics_success_rate');
const analyticsLatency = new Trend('analytics_latency_ms');
const cacheableRequests = new Rate('cacheable_requests');

export const options = {
  scenarios: {
    // Normal analytics load
    normal_load: {
      executor: 'constant-arrival-rate',
      rate: 100,
      timeUnit: '1s',
      duration: '10m',
      preAllocatedVUs: 50,
      maxVUs: 100,
      exec: 'analyticsTest',
    },

    // Peak analytics load
    peak_load: {
      executor: 'ramping-arrival-rate',
      startRate: 100,
      timeUnit: '1s',
      preAllocatedVUs: 100,
      maxVUs: 200,
      stages: [
        { duration: '2m', target: 300 },
        { duration: '8m', target: 300 },
        { duration: '2m', target: 100 },
      ],
      startTime: '10m',
      exec: 'analyticsTest',
    },

    // Complex analytics queries
    complex_queries: {
      executor: 'constant-vus',
      vus: 20,
      duration: '5m',
      startTime: '22m',
      exec: 'advancedAnalyticsTest',
    },

    // Dashboard statistics
    dashboard_load: {
      executor: 'constant-arrival-rate',
      rate: 50,
      timeUnit: '1s',
      duration: '5m',
      preAllocatedVUs: 25,
      maxVUs: 50,
      startTime: '27m',
      exec: 'dashboardTest',
    },
  },

  thresholds: {
    'http_req_duration{name:analytics}': [
      'p(95)<300',
      'p(99)<1000',
    ],
    'analytics_success_rate': ['rate>0.99'],
  },
};

const BASE_URL = __ENV.BASE_URL || 'http://localhost:8080';

// Test short codes for analytics
const SHORT_CODES = [
  'abc123',
  'xyz789',
  'test01',
  'demo99',
  'sample',
];

export function analyticsTest() {
  const shortCode = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];

  // Random time period (1, 7, 30, 90 days)
  const days = [1, 7, 30, 90][Math.floor(Math.random() * 4)];

  const url = `${BASE_URL}/api/v1/analytics/${shortCode}?days=${days}`;

  const startTime = new Date().getTime();
  const response = http.get(url, { tags: { name: 'analytics' } });
  const duration = new Date().getTime() - startTime;

  analyticsLatency.add(duration);

  const success = check(response, {
    'status is 200': (r) => r.status === 200,
    'has total_clicks': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total_clicks !== undefined;
      } catch (e) {
        return false;
      }
    },
    'has click_trends': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.click_trends !== undefined;
      } catch (e) {
        return false;
      }
    },
    'latency < 1s': () => duration < 1000,
  });

  analyticsSuccessRate.add(success);

  // Mark frequently accessed analytics as cacheable
  if (days === 7 || days === 30) {
    cacheableRequests.add(1);
  } else {
    cacheableRequests.add(0);
  }

  sleep(0.5);
}

export function advancedAnalyticsTest() {
  const shortCode = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
  const days = 30;

  const url = `${BASE_URL}/api/v1/analytics/${shortCode}/advanced?days=${days}`;

  const response = http.get(url, { tags: { name: 'advanced_analytics' } });

  check(response, {
    'status is 200': (r) => r.status === 200,
    'has geo_distribution': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.geo_distribution !== undefined;
      } catch (e) {
        return false;
      }
    },
    'has device_breakdown': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.device_breakdown !== undefined;
      } catch (e) {
        return false;
      }
    },
    'has referrer_breakdown': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.referrer_breakdown !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(1);
}

export function dashboardTest() {
  const url = `${BASE_URL}/api/v1/analytics/dashboard`;

  const response = http.get(url, { tags: { name: 'dashboard' } });

  check(response, {
    'status is 200': (r) => r.status === 200,
    'has total_urls': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total_urls !== undefined;
      } catch (e) {
        return false;
      }
    },
    'has total_clicks': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.total_clicks !== undefined;
      } catch (e) {
        return false;
      }
    },
  });

  sleep(1);
}

// Test click trends endpoint
export function clickTrendsTest() {
  const shortCode = SHORT_CODES[Math.floor(Math.random() * SHORT_CODES.length)];
  const periods = ['hour', 'day', 'week', 'month'];
  const period = periods[Math.floor(Math.random() * periods.length)];

  const url = `${BASE_URL}/api/v1/analytics/${shortCode}/trends?period=${period}`;

  const response = http.get(url, { tags: { name: 'click_trends' } });

  check(response, {
    'status is 200': (r) => r.status === 200,
    'has trends data': (r) => {
      try {
        const body = JSON.parse(r.body);
        return body.trends !== undefined && Array.isArray(body.trends);
      } catch (e) {
        return false;
      }
    },
  });

  sleep(0.5);
}

export function setup() {
  console.log('Starting analytics load test...');
  console.log(`Base URL: ${BASE_URL}`);

  return { timestamp: new Date().toISOString() };
}

export function teardown(data) {
  console.log(`Test completed at: ${new Date().toISOString()}`);
}
