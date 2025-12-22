module.exports = {
  ci: {
    collect: {
      // Number of runs for each URL to get stable metrics
      numberOfRuns: 3,

      // URLs to test
      url: [
        'http://localhost:3000/',
        'http://localhost:3000/pricing',
        'http://localhost:3000/dashboard',
        'http://localhost:3000/analytics',
      ],

      // Start a server before collecting
      startServerCommand: 'npm run start',
      startServerReadyPattern: 'ready on',
      startServerReadyTimeout: 30000,

      // Settings for the run
      settings: {
        preset: 'desktop',
        // Throttling settings (simulated fast 3G)
        throttling: {
          rttMs: 40,
          throughputKbps: 10 * 1024,
          cpuSlowdownMultiplier: 1,
        },
        // Skip PWA audit in CI
        skipAudits: ['uses-http2', 'service-worker'],
      },
    },

    assert: {
      preset: 'lighthouse:recommended',
      assertions: {
        // Performance budgets
        'categories:performance': ['error', { minScore: 0.85 }],
        'categories:accessibility': ['error', { minScore: 0.90 }],
        'categories:best-practices': ['error', { minScore: 0.90 }],
        'categories:seo': ['error', { minScore: 0.85 }],

        // Core Web Vitals
        'first-contentful-paint': ['error', { maxNumericValue: 1800 }],
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'total-blocking-time': ['error', { maxNumericValue: 300 }],
        'interactive': ['error', { maxNumericValue: 3800 }],
        'speed-index': ['error', { maxNumericValue: 3400 }],

        // Resource budgets
        'resource-summary:script:size': ['error', { maxNumericValue: 716800 }], // 700 KB
        'resource-summary:stylesheet:size': ['error', { maxNumericValue: 102400 }], // 100 KB
        'resource-summary:image:size': ['error', { maxNumericValue: 512000 }], // 500 KB
        'resource-summary:total:size': ['error', { maxNumericValue: 2097152 }], // 2 MB

        // Network requests
        'resource-summary:third-party:count': ['warn', { maxNumericValue: 10 }],
        'total-byte-weight': ['error', { maxNumericValue: 2097152 }], // 2 MB

        // Rendering performance
        'dom-size': ['warn', { maxNumericValue: 1500 }],
        'bootup-time': ['warn', { maxNumericValue: 3500 }],
        'mainthread-work-breakdown': ['warn', { maxNumericValue: 4000 }],

        // User experience
        'uses-responsive-images': 'warn',
        'offscreen-images': 'warn',
        'unminified-css': 'error',
        'unminified-javascript': 'error',
        'unused-css-rules': 'warn',
        'unused-javascript': 'warn',
        'modern-image-formats': 'warn',
        'uses-optimized-images': 'warn',
        'uses-text-compression': 'error',
        'uses-long-cache-ttl': 'warn',

        // Accessibility
        'color-contrast': 'error',
        'image-alt': 'error',
        'label': 'error',
        'tap-targets': 'warn',
      },
    },

    upload: {
      target: 'temporary-public-storage',
      // For permanent storage, use:
      // target: 'lhci',
      // serverBaseUrl: process.env.LHCI_SERVER_URL,
      // token: process.env.LHCI_TOKEN,
    },

    server: {
      // Configuration for LHCI server (optional)
      // port: 9001,
      // storage: {
      //   storageMethod: 'sql',
      //   sqlDialect: 'postgres',
      //   sqlConnectionUrl: process.env.LHCI_DATABASE_URL,
      // },
    },
  },
};
