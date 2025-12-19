/**
 * Self-Healing Performance Test Utilities
 *
 * Provides automatic retry logic, flaky test detection, and environment normalization
 * to ensure stable and reliable performance testing in CI/CD pipelines.
 */

import http from 'k6/http';
import { check, sleep } from 'k6';

// Configuration
const RETRY_CONFIG = {
  maxRetries: 3,
  baseDelay: 1000, // ms
  maxDelay: 10000, // ms
  backoffMultiplier: 2,
  jitter: true,
};

const FLAKY_TEST_CONFIG = {
  detectionWindow: 10, // Number of recent runs to analyze
  flakyThreshold: 0.3, // 30% failure rate = flaky
  consecutiveFailuresThreshold: 3,
};

const ENVIRONMENT_NORMALIZATION = {
  warmupRequests: 10,
  warmupDelay: 100, // ms
  stabilizationPeriod: 5000, // ms
  healthCheckRetries: 5,
};

/**
 * Retry a function with exponential backoff
 */
export function withRetry(fn, options = {}) {
  const config = { ...RETRY_CONFIG, ...options };
  let lastError;

  for (let attempt = 0; attempt <= config.maxRetries; attempt++) {
    try {
      return fn();
    } catch (error) {
      lastError = error;

      if (attempt < config.maxRetries) {
        const delay = calculateBackoffDelay(attempt, config);
        console.log(
          `Attempt ${attempt + 1} failed: ${error.message}. Retrying in ${delay}ms...`
        );
        sleep(delay / 1000);
      }
    }
  }

  throw new Error(
    `Failed after ${config.maxRetries + 1} attempts: ${lastError.message}`
  );
}

/**
 * Calculate exponential backoff delay with optional jitter
 */
function calculateBackoffDelay(attempt, config) {
  const exponentialDelay = Math.min(
    config.baseDelay * Math.pow(config.backoffMultiplier, attempt),
    config.maxDelay
  );

  if (config.jitter) {
    // Add random jitter: ±25% of calculated delay
    const jitterRange = exponentialDelay * 0.25;
    const jitter = Math.random() * 2 * jitterRange - jitterRange;
    return Math.max(0, exponentialDelay + jitter);
  }

  return exponentialDelay;
}

/**
 * Resilient HTTP request with automatic retry
 */
export function resilientRequest(url, params = {}, options = {}) {
  return withRetry(() => {
    const response = http.get(url, params);

    // Check for transient errors that should be retried
    if (isTransientError(response)) {
      throw new Error(
        `Transient error: ${response.status} - ${response.error || 'Unknown error'}`
      );
    }

    return response;
  }, options);
}

/**
 * Check if an error is transient and should be retried
 */
function isTransientError(response) {
  // Retry on network errors
  if (response.error) {
    return true;
  }

  // Retry on specific HTTP status codes
  const retryableStatuses = [408, 429, 500, 502, 503, 504];
  return retryableStatuses.includes(response.status);
}

/**
 * Wait for service to be healthy before running tests
 */
export function waitForHealthy(healthUrl, timeout = 30000) {
  const startTime = Date.now();
  const retries = ENVIRONMENT_NORMALIZATION.healthCheckRetries;

  for (let i = 0; i < retries; i++) {
    const response = http.get(healthUrl, { timeout: '5s' });

    if (response.status === 200) {
      console.log(`Service is healthy after ${Date.now() - startTime}ms`);
      return true;
    }

    if (Date.now() - startTime > timeout) {
      throw new Error(`Service health check timeout after ${timeout}ms`);
    }

    console.log(
      `Health check attempt ${i + 1}/${retries} failed. Retrying...`
    );
    sleep(2);
  }

  throw new Error(`Service failed to become healthy after ${retries} attempts`);
}

/**
 * Warm up the system before performance testing
 */
export function warmupSystem(baseUrl, endpoints = ['/']) {
  console.log('Warming up system...');

  endpoints.forEach((endpoint) => {
    for (let i = 0; i < ENVIRONMENT_NORMALIZATION.warmupRequests; i++) {
      http.get(`${baseUrl}${endpoint}`);
      sleep(ENVIRONMENT_NORMALIZATION.warmupDelay / 1000);
    }
  });

  // Stabilization period
  console.log(
    `Stabilizing for ${ENVIRONMENT_NORMALIZATION.stabilizationPeriod}ms...`
  );
  sleep(ENVIRONMENT_NORMALIZATION.stabilizationPeriod / 1000);

  console.log('Warmup complete');
}

/**
 * Flaky test detector
 */
export class FlakyTestDetector {
  constructor() {
    this.testResults = new Map();
  }

  recordResult(testName, passed) {
    if (!this.testResults.has(testName)) {
      this.testResults.set(testName, []);
    }

    const results = this.testResults.get(testName);
    results.push({ passed, timestamp: Date.now() });

    // Keep only recent results
    if (results.length > FLAKY_TEST_CONFIG.detectionWindow) {
      results.shift();
    }
  }

  isFlaky(testName) {
    const results = this.testResults.get(testName);
    if (!results || results.length < FLAKY_TEST_CONFIG.detectionWindow) {
      return false;
    }

    const failures = results.filter((r) => !r.passed).length;
    const failureRate = failures / results.length;

    return failureRate >= FLAKY_TEST_CONFIG.flakyThreshold;
  }

  hasConsecutiveFailures(testName) {
    const results = this.testResults.get(testName);
    if (!results) {
      return false;
    }

    let consecutiveFailures = 0;
    for (let i = results.length - 1; i >= 0; i--) {
      if (results[i].passed) {
        break;
      }
      consecutiveFailures++;
    }

    return (
      consecutiveFailures >= FLAKY_TEST_CONFIG.consecutiveFailuresThreshold
    );
  }

  generateReport() {
    const report = [];

    this.testResults.forEach((results, testName) => {
      const failures = results.filter((r) => !r.passed).length;
      const failureRate = failures / results.length;

      report.push({
        testName,
        totalRuns: results.length,
        failures,
        failureRate,
        isFlaky: this.isFlaky(testName),
        hasConsecutiveFailures: this.hasConsecutiveFailures(testName),
      });
    });

    return report;
  }
}

/**
 * Environment normalization wrapper
 */
export function withEnvironmentNormalization(testFn, options = {}) {
  return function normalizedTest(data) {
    const baseUrl = options.baseUrl || __ENV.BASE_URL || 'http://localhost:8080';

    // Wait for service to be healthy
    waitForHealthy(`${baseUrl}/health`);

    // Warm up the system
    warmupSystem(baseUrl, options.warmupEndpoints);

    // Run the actual test
    return testFn(data);
  };
}

/**
 * Adaptive performance threshold
 * Adjusts thresholds based on environment conditions
 */
export class AdaptiveThreshold {
  constructor(baseThreshold, tolerance = 0.2) {
    this.baseThreshold = baseThreshold;
    this.tolerance = tolerance;
    this.measurements = [];
  }

  record(value) {
    this.measurements.push(value);
    if (this.measurements.length > 100) {
      this.measurements.shift();
    }
  }

  getThreshold() {
    if (this.measurements.length < 10) {
      return this.baseThreshold;
    }

    // Calculate median of recent measurements
    const sorted = [...this.measurements].sort((a, b) => a - b);
    const median = sorted[Math.floor(sorted.length / 2)];

    // Adjust threshold based on recent performance
    const adaptedThreshold = this.baseThreshold * (1 + this.tolerance);

    // Use the higher of base or adapted threshold
    return Math.max(this.baseThreshold, Math.min(adaptedThreshold, median * 1.5));
  }

  check(value) {
    this.record(value);
    return value <= this.getThreshold();
  }
}

/**
 * Intelligent retry with circuit breaker pattern
 */
export class CircuitBreaker {
  constructor(options = {}) {
    this.failureThreshold = options.failureThreshold || 5;
    this.resetTimeout = options.resetTimeout || 60000; // 60s
    this.state = 'CLOSED'; // CLOSED, OPEN, HALF_OPEN
    this.failures = 0;
    this.lastFailureTime = null;
  }

  execute(fn) {
    if (this.state === 'OPEN') {
      if (Date.now() - this.lastFailureTime > this.resetTimeout) {
        this.state = 'HALF_OPEN';
      } else {
        throw new Error('Circuit breaker is OPEN');
      }
    }

    try {
      const result = fn();
      this.onSuccess();
      return result;
    } catch (error) {
      this.onFailure();
      throw error;
    }
  }

  onSuccess() {
    this.failures = 0;
    this.state = 'CLOSED';
  }

  onFailure() {
    this.failures++;
    this.lastFailureTime = Date.now();

    if (this.failures >= this.failureThreshold) {
      this.state = 'OPEN';
      console.log('Circuit breaker opened due to excessive failures');
    }
  }

  getState() {
    return this.state;
  }
}

/**
 * Smart assertion with auto-healing
 */
export function smartCheck(response, checks, options = {}) {
  const detector = options.detector || new FlakyTestDetector();
  const testName = options.testName || 'unnamed-test';

  const result = check(response, checks);
  detector.recordResult(testName, result);

  if (!result && detector.isFlaky(testName)) {
    console.warn(`Flaky test detected: ${testName}`);
    // Could implement auto-skip or different handling here
  }

  return result;
}

/**
 * Performance test wrapper with full self-healing capabilities
 */
export function selfHealingTest(testConfig) {
  const {
    name,
    testFn,
    warmupEndpoints = ['/'],
    retryConfig = {},
    adaptiveThresholds = {},
  } = testConfig;

  const detector = new FlakyTestDetector();
  const circuitBreaker = new CircuitBreaker();
  const thresholds = {};

  // Initialize adaptive thresholds
  Object.entries(adaptiveThresholds).forEach(([key, baseValue]) => {
    thresholds[key] = new AdaptiveThreshold(baseValue);
  });

  return withEnvironmentNormalization(
    (data) => {
      return circuitBreaker.execute(() => {
        return withRetry(
          () => {
            return testFn(data, { detector, thresholds });
          },
          retryConfig
        );
      });
    },
    { warmupEndpoints }
  );
}

export default {
  withRetry,
  resilientRequest,
  waitForHealthy,
  warmupSystem,
  FlakyTestDetector,
  withEnvironmentNormalization,
  AdaptiveThreshold,
  CircuitBreaker,
  smartCheck,
  selfHealingTest,
};
