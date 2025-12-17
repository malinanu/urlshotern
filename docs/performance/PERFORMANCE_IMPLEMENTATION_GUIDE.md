# Performance Optimization Implementation Guide
## Practical Code Examples for Immediate Improvements

This guide provides ready-to-implement code for the highest-impact performance optimizations identified in the UX Performance Metrics Analysis.

---

## Quick Start: Priority 1-4 Implementations (1-2 Days)

### 1. Core Web Vitals Monitoring (4 hours)

#### Step 1: Install web-vitals package

```bash
npm install web-vitals --save
```

#### Step 2: Create Web Vitals Reporter

Create `frontend/src/lib/web-vitals-reporter.ts`:

```typescript
import { onCLS, onFID, onLCP, onFCP, onTTFB, onINP, Metric } from 'web-vitals';

interface WebVitalsMetric extends Metric {
  // Extend with custom fields
  page?: string;
  userId?: string;
  userAgent?: string;
  connectionType?: string;
}

export function reportWebVitals(metric: Metric) {
  const body: WebVitalsMetric = {
    ...metric,
    page: window.location.pathname,
    userAgent: navigator.userAgent,
    connectionType: (navigator as any).connection?.effectiveType || 'unknown',
  };

  // Add user ID if authenticated
  const userId = localStorage.getItem('userId');
  if (userId) {
    body.userId = userId;
  }

  const url = `${process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'}/api/v1/analytics/web-vitals`;

  // Use sendBeacon if available for reliability
  if (navigator.sendBeacon) {
    navigator.sendBeacon(url, JSON.stringify(body));
  } else {
    // Fallback to fetch with keepalive
    fetch(url, {
      method: 'POST',
      headers: { 'Content-Type': 'application/json' },
      body: JSON.stringify(body),
      keepalive: true,
    }).catch(console.error);
  }

  // Also log to console in development
  if (process.env.NODE_ENV === 'development') {
    console.log('📊 Web Vital:', metric.name, metric.value, metric);
  }
}

// Initialize all Core Web Vitals listeners
export function initWebVitals() {
  onCLS(reportWebVitals);
  onFID(reportWebVitals);
  onLCP(reportWebVitals);
  onFCP(reportWebVitals);
  onTTFB(reportWebVitals);

  // New INP metric (replacing FID)
  if (typeof onINP !== 'undefined') {
    onINP(reportWebVitals);
  }
}
```

#### Step 3: Initialize in Root Layout

Update `frontend/src/app/layout.tsx`:

```typescript
import type { Metadata, Viewport } from "next";
import { Geist, Geist_Mono } from "next/font/google";
import { AuthProvider } from "@/contexts/AuthContext";
import { ThemeProvider } from "@/contexts/ThemeContext";
import { PWAInstallPrompt } from "@/components/PWAInstallPrompt";
import "./globals.css";

// Add this import
import { initWebVitals } from "@/lib/web-vitals-reporter";
import { useEffect } from "react";

const geistSans = Geist({
  variable: "--font-geist-sans",
  subsets: ["latin"],
  display: 'swap',           // Add this for better font loading
  fallback: ['system-ui'],   // Add fallback fonts
  preload: true,             // Preload fonts
});

const geistMono = Geist_Mono({
  variable: "--font-geist-mono",
  subsets: ["latin"],
  display: 'swap',
  fallback: ['monospace'],
  preload: true,
});

// ... existing metadata ...

// Add this component to initialize web vitals
function WebVitalsInit() {
  useEffect(() => {
    initWebVitals();
  }, []);

  return null;
}

export default function RootLayout({
  children,
}: Readonly<{
  children: React.NodeNode;
}>) {
  return (
    <html lang="en">
      <head>
        {/* Add resource hints for performance */}
        <link rel="dns-prefetch" href={process.env.NEXT_PUBLIC_API_URL} />
        <link rel="preconnect" href={process.env.NEXT_PUBLIC_API_URL} />

        <script dangerouslySetInnerHTML={{
          __html: `
            if ('serviceWorker' in navigator) {
              window.addEventListener('load', () => {
                navigator.serviceWorker.register('/sw.js')
                  .then((registration) => {
                    console.log('SW registered: ', registration);
                  })
                  .catch((registrationError) => {
                    console.log('SW registration failed: ', registrationError);
                  });
              });
            }
          `
        }} />
      </head>
      <body
        className={`${geistSans.variable} ${geistMono.variable} antialiased`}
        suppressHydrationWarning
      >
        <ThemeProvider>
          <AuthProvider>
            <WebVitalsInit />
            {children}
            <PWAInstallPrompt />
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
```

#### Step 4: Add Backend Endpoint

Create `internal/handlers/web_vitals.go`:

```go
package handlers

import (
	"net/http"
	"time"

	"github.com/gin-gonic/gin"
)

type WebVitalMetric struct {
	Name           string    `json:"name"`
	Value          float64   `json:"value"`
	Rating         string    `json:"rating"`
	Delta          float64   `json:"delta"`
	ID             string    `json:"id"`
	NavigationType string    `json:"navigationType"`
	Page           string    `json:"page"`
	UserID         string    `json:"userId,omitempty"`
	UserAgent      string    `json:"userAgent"`
	ConnectionType string    `json:"connectionType"`
	Timestamp      time.Time `json:"timestamp"`
}

// RecordWebVital records Core Web Vitals metrics from the frontend
func (h *Handler) RecordWebVital(c *gin.Context) {
	var metric WebVitalMetric

	if err := c.ShouldBindJSON(&metric); err != nil {
		c.JSON(http.StatusBadRequest, gin.H{"error": err.Error()})
		return
	}

	metric.Timestamp = time.Now()

	// TODO: Store in database or analytics service
	// For now, log to console for monitoring
	// In production, send to DataDog, New Relic, or custom analytics

	// Log metric
	go func() {
		// Log structured metric
		// You can integrate with your analytics service here
		// Example: Send to DataDog, Prometheus, InfluxDB, etc.
	}()

	c.JSON(http.StatusOK, gin.H{"status": "recorded"})
}
```

Add route in `internal/routes/routes.go`:

```go
// Add this line in SetupRoutes function
api.POST("/analytics/web-vitals", handler.RecordWebVital)
```

---

### 2. Lazy Load Heavy Components (1 day)

#### Update Analytics Page

Update `frontend/src/app/analytics/page.tsx`:

```typescript
'use client';

import { Suspense, lazy } from 'react';
import dynamic from 'next/dynamic';
import Layout from '@/components/layout/Layout';
import { useSearchParams } from 'next/navigation';

// Lazy load heavy components
const AdvancedAnalytics = dynamic(
  () => import('@/components/analytics/AdvancedAnalytics'),
  {
    loading: () => <AnalyticsSkeleton />,
    ssr: false, // Don't render on server
  }
);

const GeographicMap = dynamic(
  () => import('@/components/analytics/GeographicMap'),
  {
    loading: () => <MapSkeleton />,
    ssr: false,
  }
);

const DeviceAnalytics = dynamic(
  () => import('@/components/analytics/DeviceAnalytics'),
  {
    loading: () => <DeviceSkeleton />,
    ssr: false,
  }
);

const ReferrerAnalytics = dynamic(
  () => import('@/components/analytics/ReferrerAnalytics'),
  {
    loading: () => <ReferrerSkeleton />,
    ssr: false,
  }
);

// Skeleton components for better UX
function AnalyticsSkeleton() {
  return (
    <div className="animate-pulse space-y-4">
      <div className="h-64 bg-gray-200 rounded-lg"></div>
      <div className="h-48 bg-gray-200 rounded-lg"></div>
    </div>
  );
}

function MapSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-96 bg-gray-200 rounded-lg flex items-center justify-center">
        <p className="text-gray-500">Loading map...</p>
      </div>
    </div>
  );
}

function DeviceSkeleton() {
  return (
    <div className="animate-pulse space-y-4">
      <div className="h-32 bg-gray-200 rounded-lg"></div>
      <div className="h-32 bg-gray-200 rounded-lg"></div>
    </div>
  );
}

function ReferrerSkeleton() {
  return (
    <div className="animate-pulse">
      <div className="h-48 bg-gray-200 rounded-lg"></div>
    </div>
  );
}

export default function AnalyticsPage() {
  const searchParams = useSearchParams();
  const shortCode = searchParams.get('shortCode');

  if (!shortCode) {
    return (
      <Layout>
        <div className="min-h-screen flex items-center justify-center">
          <p className="text-gray-600">Please provide a short code to view analytics.</p>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <h1 className="text-3xl font-bold text-gray-900 mb-8">
            Analytics for {shortCode}
          </h1>

          {/* Basic stats load immediately */}
          <Suspense fallback={<AnalyticsSkeleton />}>
            <AdvancedAnalytics shortCode={shortCode} />
          </Suspense>

          {/* Map loads when visible */}
          <Suspense fallback={<MapSkeleton />}>
            <GeographicMap shortCode={shortCode} />
          </Suspense>

          {/* Device analytics */}
          <Suspense fallback={<DeviceSkeleton />}>
            <DeviceAnalytics shortCode={shortCode} />
          </Suspense>

          {/* Referrer analytics */}
          <Suspense fallback={<ReferrerSkeleton />}>
            <ReferrerAnalytics shortCode={shortCode} />
          </Suspense>
        </div>
      </div>
    </Layout>
  );
}
```

#### Lazy Load QR Code Modal

Update `frontend/src/components/URLShortener.tsx`:

```typescript
'use client';

import { useState } from 'react';
import dynamic from 'next/dynamic';
// ... other imports ...

// Lazy load QR Code modal
const QRCodeModal = dynamic(
  () => import('@/components/QRCode').then(mod => ({ default: mod.QRCodeModal })),
  { ssr: false }
);

export default function URLShortener({ onUrlCreated }: URLShortenerProps) {
  // ... existing code ...

  const [showQRModal, setShowQRModal] = useState(false);

  return (
    <div className="max-w-2xl mx-auto">
      {/* ... existing form code ... */}

      {/* Only load QR modal when needed */}
      {shortenedURL && showQRModal && (
        <QRCodeModal
          isOpen={showQRModal}
          onClose={() => setShowQRModal(false)}
          value={shortenedURL.short_url}
          title="QR Code for Short URL"
          description={`QR code for ${shortenedURL.short_url}`}
          size={300}
        />
      )}
    </div>
  );
}
```

---

### 3. Parallel Data Fetching (4 hours)

#### Update Dashboard Data Fetching

Update `frontend/src/app/dashboard/page.tsx`:

```typescript
'use client';

import { useEffect, useState } from 'react';
// ... other imports ...

export default function DashboardPage() {
  // ... existing state ...

  // BEFORE: Sequential fetching (slow)
  // const fetchData = async (retryOnTokenRefresh = true) => {
  //   const urlsResponse = await makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/my-urls`);
  //   // ... then stats response
  // };

  // AFTER: Parallel fetching (fast)
  const fetchData = async (retryOnTokenRefresh = true) => {
    if (!isAuthenticated || !user) return;

    const token = getAccessToken();
    if (!token) return;

    const makeAuthenticatedRequest = async (url: string) => {
      const response = await fetch(url, {
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.status === 401 && retryOnTokenRefresh) {
        try {
          await refreshToken();
          return fetchData(false);
        } catch (refreshError) {
          console.error('Token refresh failed:', refreshError);
        }
      }

      return response;
    };

    try {
      setLoadingUrls(true);
      setLoadingStats(true);

      // Fetch BOTH in parallel
      const [urlsResponse, statsResponse] = await Promise.all([
        makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/my-urls`),
        makeAuthenticatedRequest(`${API_BASE_URL}/api/v1/dashboard/stats`)
      ]);

      // Process URLs
      if (urlsResponse && urlsResponse.ok) {
        const urlsData = await urlsResponse.json();
        setUrls(urlsData.urls || []);
      } else if (urlsResponse) {
        console.error('Failed to fetch URLs:', urlsResponse.status);
      }

      // Process stats
      if (statsResponse && statsResponse.ok) {
        const statsData = await statsResponse.json();
        setStats(statsData);
      } else if (statsResponse) {
        console.error('Failed to fetch stats:', statsResponse.status);
      }
    } catch (error) {
      console.error('Error fetching dashboard data:', error);
    } finally {
      setLoadingUrls(false);
      setLoadingStats(false);
    }
  };

  // ... rest of component ...
}
```

---

### 4. Font Optimization (2 hours)

Already shown in Step 3 of Web Vitals implementation above. Additional optimization:

#### Add font-display CSS

Update `frontend/src/app/globals.css`:

```css
@tailwind base;
@tailwind components;
@tailwind utilities;

/* Add font-display for system font fallback */
@layer base {
  body {
    font-family: var(--font-geist-sans), system-ui, -apple-system, BlinkMacSystemFont, "Segoe UI", Roboto, sans-serif;
    font-feature-settings: "kern" 1, "liga" 1;
    -webkit-font-smoothing: antialiased;
    -moz-osx-font-smoothing: grayscale;
  }

  code, pre {
    font-family: var(--font-geist-mono), ui-monospace, "Cascadia Code", "Source Code Pro", Menlo, Consolas, monospace;
  }
}

/* Rest of existing CSS... */
```

---

## Medium-Term Optimizations (1 Week)

### 5. Implement SWR for Request Caching (2 days)

#### Install SWR

```bash
npm install swr
```

#### Create SWR Configuration

Create `frontend/src/lib/swr-config.ts`:

```typescript
import { SWRConfiguration } from 'swr';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

// Default fetcher with authentication
export const fetcher = async (url: string, token?: string) => {
  const headers: HeadersInit = {
    'Content-Type': 'application/json',
  };

  if (token) {
    headers.Authorization = `Bearer ${token}`;
  }

  const response = await fetch(`${API_BASE_URL}${url}`, { headers });

  if (!response.ok) {
    const error = new Error('An error occurred while fetching the data.');
    // Attach extra info to the error object
    (error as any).info = await response.json();
    (error as any).status = response.status;
    throw error;
  }

  return response.json();
};

// SWR global configuration
export const swrConfig: SWRConfiguration = {
  revalidateOnFocus: false,        // Don't revalidate on window focus
  revalidateOnReconnect: true,     // Revalidate on network reconnect
  dedupingInterval: 5000,          // Dedupe requests within 5 seconds
  errorRetryCount: 3,              // Retry 3 times on error
  errorRetryInterval: 1000,        // Wait 1 second between retries
  shouldRetryOnError: true,        // Retry on error
  focusThrottleInterval: 30000,    // Throttle focus revalidation to 30s
};
```

#### Create SWR Provider

Create `frontend/src/contexts/SWRProvider.tsx`:

```typescript
'use client';

import { SWRConfig } from 'swr';
import { swrConfig } from '@/lib/swr-config';

export function SWRProvider({ children }: { children: React.ReactNode }) {
  return <SWRConfig value={swrConfig}>{children}</SWRConfig>;
}
```

#### Update Root Layout

Add SWR Provider to `frontend/src/app/layout.tsx`:

```typescript
import { SWRProvider } from "@/contexts/SWRProvider";

export default function RootLayout({
  children,
}: Readonly<{
  children: React.ReactNode;
}>) {
  return (
    <html lang="en">
      {/* ... head content ... */}
      <body className={`${geistSans.variable} ${geistMono.variable} antialiased`}>
        <ThemeProvider>
          <AuthProvider>
            <SWRProvider>
              <WebVitalsInit />
              {children}
              <PWAInstallPrompt />
            </SWRProvider>
          </AuthProvider>
        </ThemeProvider>
      </body>
    </html>
  );
}
```

#### Use SWR in Dashboard

Update `frontend/src/app/dashboard/page.tsx`:

```typescript
'use client';

import useSWR from 'swr';
import { fetcher } from '@/lib/swr-config';
import { useAuth } from '@/contexts/AuthContext';

export default function DashboardPage() {
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const token = getAccessToken();

  // Fetch URLs with SWR
  const { data: urlsData, error: urlsError, mutate: mutateUrls } = useSWR(
    isAuthenticated && token ? ['/api/v1/my-urls', token] : null,
    ([url, token]) => fetcher(url, token),
    {
      revalidateOnMount: true,
      dedupingInterval: 10000, // Cache for 10 seconds
    }
  );

  // Fetch stats with SWR
  const { data: statsData, error: statsError } = useSWR(
    isAuthenticated && token ? ['/api/v1/dashboard/stats', token] : null,
    ([url, token]) => fetcher(url, token),
    {
      revalidateOnMount: true,
      dedupingInterval: 30000, // Cache for 30 seconds
    }
  );

  const urls = urlsData?.urls || [];
  const stats = statsData || null;
  const loadingUrls = !urlsData && !urlsError;
  const loadingStats = !statsData && !statsError;

  // Callback for when a new URL is created
  const handleUrlCreated = async (newUrl: any) => {
    // Optimistic update - instantly show in UI
    mutateUrls(
      (currentData: any) => ({
        ...currentData,
        urls: [newUrl, ...(currentData?.urls || [])],
      }),
      false // Don't revalidate immediately
    );

    // Revalidate in background after a short delay
    setTimeout(() => {
      mutateUrls();
    }, 1000);
  };

  // ... rest of component ...
}
```

---

### 6. Image Optimization with Next.js Image (1 day)

#### Update Components to Use Next.js Image

Example: If you add hero images or user avatars:

```typescript
import Image from 'next/image';

// Hero section image
<div className="relative w-full h-96">
  <Image
    src="/images/hero-background.jpg"
    alt="URL Shortener Hero"
    fill
    priority  // Load immediately for above-fold images
    sizes="100vw"
    className="object-cover"
    placeholder="blur"
    blurDataURL="data:image/jpeg;base64,/9j/4AAQSkZJRg..." // Generate with plaiceholder
  />
</div>

// User avatar
<div className="relative w-12 h-12 rounded-full overflow-hidden">
  <Image
    src={user.avatarUrl || '/images/default-avatar.png'}
    alt={user.name}
    fill
    sizes="48px"
    className="object-cover"
  />
</div>

// Logo
<Image
  src="/logo.svg"
  alt="Trunc Logo"
  width={120}
  height={40}
  priority
/>
```

#### Configure Image Optimization

Update `frontend/next.config.ts`:

```typescript
import type { NextConfig } from "next";

const nextConfig: NextConfig = {
  eslint: {
    ignoreDuringBuilds: true,
  },
  typescript: {
    ignoreBuildErrors: true,
  },
  images: {
    formats: ['image/avif', 'image/webp'],
    deviceSizes: [640, 750, 828, 1080, 1200, 1920, 2048, 3840],
    imageSizes: [16, 32, 48, 64, 96, 128, 256, 384],
    minimumCacheTTL: 60,
    dangerouslyAllowSVG: true,
    contentDispositionType: 'attachment',
    contentSecurityPolicy: "default-src 'self'; script-src 'none'; sandbox;",
  },
  // PWA and mobile optimization
  async headers() {
    return [
      {
        source: '/manifest.json',
        headers: [
          {
            key: 'Content-Type',
            value: 'application/manifest+json',
          },
        ],
      },
      {
        source: '/sw.js',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=0, must-revalidate',
          },
        ],
      },
      // Cache static assets aggressively
      {
        source: '/:all*(svg|jpg|jpeg|png|gif|ico|webp|avif)',
        headers: [
          {
            key: 'Cache-Control',
            value: 'public, max-age=31536000, immutable',
          },
        ],
      },
    ]
  },
  // Enable static export optimization
  output: 'standalone',
};

export default nextConfig;
```

---

### 7. Bundle Analysis and Optimization (2 days)

#### Install Bundle Analyzer

```bash
npm install @next/bundle-analyzer --save-dev
```

#### Configure Bundle Analyzer

Update `frontend/next.config.ts`:

```typescript
import type { NextConfig } from "next";
import bundleAnalyzer from '@next/bundle-analyzer';

const withBundleAnalyzer = bundleAnalyzer({
  enabled: process.env.ANALYZE === 'true',
});

const nextConfig: NextConfig = {
  // ... existing config ...
};

export default withBundleAnalyzer(nextConfig);
```

#### Add Script to package.json

Update `frontend/package.json`:

```json
{
  "scripts": {
    "dev": "next dev",
    "build": "next build",
    "start": "next start",
    "lint": "eslint",
    "analyze": "ANALYZE=true npm run build"
  }
}
```

#### Run Analysis

```bash
cd frontend
npm run analyze
```

This will open a visual bundle analyzer in your browser.

---

## Advanced Optimizations

### 8. Performance Budgets (2 days)

#### Add Lighthouse CI

Create `.lighthouserc.js`:

```javascript
module.exports = {
  ci: {
    collect: {
      startServerCommand: 'npm run start',
      url: [
        'http://localhost:3000',
        'http://localhost:3000/dashboard',
        'http://localhost:3000/analytics',
        'http://localhost:3000/pricing',
      ],
      numberOfRuns: 3,
    },
    assert: {
      preset: 'lighthouse:recommended',
      assertions: {
        // Core Web Vitals
        'largest-contentful-paint': ['error', { maxNumericValue: 2500 }],
        'cumulative-layout-shift': ['error', { maxNumericValue: 0.1 }],
        'first-contentful-paint': ['warn', { maxNumericValue: 1800 }],
        'total-blocking-time': ['error', { maxNumericValue: 300 }],
        'interactive': ['error', { maxNumericValue: 3800 }],

        // Bundle sizes
        'total-byte-weight': ['error', { maxNumericValue: 1000000 }], // 1MB
        'mainthread-work-breakdown': ['warn', { maxNumericValue: 4000 }],

        // Best practices
        'uses-optimized-images': 'error',
        'uses-webp-images': 'warn',
        'offscreen-images': 'warn',
        'render-blocking-resources': 'warn',
        'unused-javascript': ['warn', { maxNumericValue: 100000 }],

        // Accessibility
        'color-contrast': 'error',
        'image-alt': 'error',
      },
    },
    upload: {
      target: 'temporary-public-storage',
    },
  },
};
```

#### Add to CI/CD

Create `.github/workflows/performance.yml`:

```yaml
name: Performance Check

on:
  pull_request:
    branches: [ main, develop ]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3

      - name: Setup Node.js
        uses: actions/setup-node@v3
        with:
          node-version: '20'
          cache: 'npm'
          cache-dependency-path: frontend/package-lock.json

      - name: Install dependencies
        run: |
          cd frontend
          npm ci

      - name: Build frontend
        run: |
          cd frontend
          npm run build
        env:
          NEXT_PUBLIC_API_URL: http://localhost:8080

      - name: Run Lighthouse CI
        run: |
          npm install -g @lhci/cli
          cd frontend
          lhci autorun
        env:
          LHCI_GITHUB_APP_TOKEN: ${{ secrets.LHCI_GITHUB_APP_TOKEN }}

      - name: Check bundle size
        run: |
          cd frontend
          npm run analyze
```

---

## Testing Performance Improvements

### Create Performance Testing Script

Create `frontend/scripts/test-performance.js`:

```javascript
const { chromium } = require('playwright');

async function testPerformance() {
  const browser = await chromium.launch();
  const context = await browser.newContext();
  const page = await context.newPage();

  // Enable performance measurement
  await page.coverage.startJSCoverage();
  await page.coverage.startCSSCoverage();

  const metrics = {};

  // Test Homepage
  console.log('Testing Homepage...');
  await page.goto('http://localhost:3000');

  const performanceTiming = JSON.parse(
    await page.evaluate(() => JSON.stringify(window.performance.timing))
  );

  metrics.homepage = {
    domContentLoaded: performanceTiming.domContentLoadedEventEnd - performanceTiming.navigationStart,
    loadComplete: performanceTiming.loadEventEnd - performanceTiming.navigationStart,
    firstPaint: JSON.parse(
      await page.evaluate(() =>
        JSON.stringify(performance.getEntriesByType('paint')[0])
      )
    ).startTime,
  };

  // Test Dashboard
  console.log('Testing Dashboard...');
  await page.goto('http://localhost:3000/dashboard');
  await page.waitForLoadState('networkidle');

  // Collect JavaScript and CSS coverage
  const jsCoverage = await page.coverage.stopJSCoverage();
  const cssCoverage = await page.coverage.stopCSSCoverage();

  let totalBytes = 0;
  let usedBytes = 0;

  for (const entry of [...jsCoverage, ...cssCoverage]) {
    totalBytes += entry.text.length;
    for (const range of entry.ranges) {
      usedBytes += range.end - range.start - 1;
    }
  }

  metrics.coverage = {
    total: totalBytes,
    used: usedBytes,
    unused: totalBytes - usedBytes,
    percentageUsed: ((usedBytes / totalBytes) * 100).toFixed(2),
  };

  console.log('\nPerformance Metrics:');
  console.log(JSON.stringify(metrics, null, 2));

  await browser.close();
}

testPerformance();
```

---

## Monitoring Dashboard Setup

### Backend: Store Web Vitals

Create `internal/models/web_vitals.go`:

```go
package models

import "time"

type WebVital struct {
	ID             int64     `json:"id"`
	Name           string    `json:"name"`           // LCP, FID, CLS, FCP, TTFB
	Value          float64   `json:"value"`          // Metric value
	Rating         string    `json:"rating"`         // good, needs-improvement, poor
	Page           string    `json:"page"`           // /dashboard, /analytics, etc.
	UserID         *int64    `json:"userId,omitempty"`
	UserAgent      string    `json:"userAgent"`
	ConnectionType string    `json:"connectionType"` // 4g, 3g, slow-2g, etc.
	DeviceType     string    `json:"deviceType"`     // mobile, tablet, desktop
	CreatedAt      time.Time `json:"createdAt"`
}
```

Create database migration `configs/migrations/003_web_vitals.sql`:

```sql
CREATE TABLE IF NOT EXISTS web_vitals (
    id BIGSERIAL PRIMARY KEY,
    name VARCHAR(50) NOT NULL,
    value DOUBLE PRECISION NOT NULL,
    rating VARCHAR(50),
    page VARCHAR(255),
    user_id BIGINT REFERENCES users(id),
    user_agent TEXT,
    connection_type VARCHAR(50),
    device_type VARCHAR(50),
    created_at TIMESTAMP WITH TIME ZONE DEFAULT CURRENT_TIMESTAMP,

    INDEX idx_web_vitals_name (name),
    INDEX idx_web_vitals_page (page),
    INDEX idx_web_vitals_created_at (created_at)
);
```

---

## Summary Checklist

### Immediate (Week 1)
- [ ] Implement Core Web Vitals monitoring
- [ ] Lazy load Recharts and Google Maps
- [ ] Parallel data fetching on dashboard
- [ ] Font optimization (display: swap)
- [ ] Add resource hints (preconnect, dns-prefetch)

### Short-term (Week 2)
- [ ] Install and configure SWR
- [ ] Implement optimistic updates
- [ ] Add Next.js Image optimization
- [ ] Run bundle analyzer
- [ ] Fix file case inconsistencies

### Medium-term (Week 3-4)
- [ ] Set up Lighthouse CI
- [ ] Implement performance budgets
- [ ] Optimize chart rendering
- [ ] Add static generation for marketing pages
- [ ] Virtual scrolling for large lists

### Ongoing
- [ ] Monitor Core Web Vitals weekly
- [ ] Review bundle size on every PR
- [ ] User testing for perceived performance
- [ ] A/B test performance improvements

---

**Next Step:** Start with Web Vitals monitoring today to establish your baseline, then tackle lazy loading tomorrow for immediate user impact!
