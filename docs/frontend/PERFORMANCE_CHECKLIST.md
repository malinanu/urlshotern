# Performance Optimization Checklist

Use this checklist before deploying to production or when investigating performance issues.

## Pre-Deployment Checklist

### 1. Build Analysis
- [ ] Run `npm run analyze` to generate bundle report
- [ ] Verify total bundle size < 700 KB
- [ ] Check that recharts is in separate chunk (async)
- [ ] Confirm no duplicate dependencies
- [ ] Verify tree-shaking is working (no unused exports)

### 2. Lighthouse Audit (Desktop)
- [ ] Performance score > 90
- [ ] LCP < 2.5s
- [ ] FID < 100ms
- [ ] CLS < 0.1
- [ ] TTI < 3.8s
- [ ] No render-blocking resources

### 3. Lighthouse Audit (Mobile)
- [ ] Performance score > 85
- [ ] LCP < 2.5s
- [ ] FID < 100ms
- [ ] CLS < 0.1
- [ ] TTI < 5s
- [ ] Network throttled to Slow 4G

### 4. Image Optimization
- [ ] All images use Next.js Image component
- [ ] Width and height attributes set
- [ ] Alt text provided for accessibility
- [ ] Images compressed (< 200 KB each)
- [ ] Modern formats enabled (AVIF/WebP)
- [ ] Lazy loading for below-fold images

### 5. Font Optimization
- [ ] Font preconnect in place
- [ ] Font preload for critical fonts
- [ ] font-display: swap or optional
- [ ] Fonts served from CDN
- [ ] Variable fonts used where possible

### 6. JavaScript Optimization
- [ ] Console logs removed from production
- [ ] Source maps disabled
- [ ] Heavy libraries lazy-loaded
- [ ] Code splitting implemented
- [ ] No unused dependencies in package.json

### 7. CSS Optimization
- [ ] Tailwind purge/content configured
- [ ] Critical CSS inlined
- [ ] No large CSS frameworks (unless necessary)
- [ ] CSS-in-JS optimized (if used)
- [ ] No unused CSS classes

### 8. Caching Strategy
- [ ] Static assets cached for 1 year
- [ ] Immutable header set for versioned assets
- [ ] Service worker configured
- [ ] Cache-Control headers verified
- [ ] CDN caching enabled

### 9. Network Optimization
- [ ] Gzip/Brotli compression enabled
- [ ] HTTP/2 enabled on server
- [ ] DNS prefetch configured
- [ ] Preconnect to external resources
- [ ] No redirect chains

### 10. Mobile Optimization
- [ ] Tested on real mobile device
- [ ] Tested on slow network (3G/4G)
- [ ] Touch targets > 48x48px
- [ ] Viewport meta tag configured
- [ ] Mobile-specific optimizations applied

## Testing Procedures

### Performance Testing Script

```bash
#!/bin/bash

echo "=== Frontend Performance Test Suite ==="
echo ""

# 1. Bundle Analysis
echo "1. Analyzing bundle size..."
ANALYZE=true npm run build
echo "✓ Bundle analysis complete - check browser for report"
echo ""

# 2. Build size check
echo "2. Checking build output..."
npm run build | grep -A 20 "Route (app)"
echo ""

# 3. Lighthouse CI (requires @lhci/cli)
echo "3. Running Lighthouse CI..."
npx lhci autorun
echo ""

echo "=== Performance Test Complete ==="
```

### Manual Testing Checklist

#### Home Page (`/`)
- [ ] Load time < 2s on desktop
- [ ] Load time < 3s on mobile
- [ ] No layout shift on load
- [ ] Hero section loads immediately
- [ ] URL shortener form interactive within 1s

#### Dashboard (`/dashboard`)
- [ ] Stats cards load quickly
- [ ] URL list renders smoothly
- [ ] Virtual scrolling works with 100+ URLs
- [ ] No janky animations
- [ ] Copy to clipboard instant feedback

#### Analytics Page (`/analytics`)
- [ ] URL selector loads immediately
- [ ] Charts lazy load with skeleton
- [ ] Charts render smoothly
- [ ] Tab switching is instant
- [ ] Data fetching shows loading state

## Performance Metrics Targets

### Critical Metrics
| Metric | Target | Current | Status |
|--------|--------|---------|--------|
| Bundle Size (Total) | < 700 KB | _____ KB | ⬜ |
| Bundle Size (JS) | < 500 KB | _____ KB | ⬜ |
| LCP (Desktop) | < 2.5s | _____ s | ⬜ |
| LCP (Mobile) | < 2.5s | _____ s | ⬜ |
| FID | < 100ms | _____ ms | ⬜ |
| CLS | < 0.1 | _____ | ⬜ |
| TTI (Desktop) | < 3.8s | _____ s | ⬜ |
| TTI (Mobile) | < 5s | _____ s | ⬜ |

### Page-Specific Targets
| Page | Desktop TTI | Mobile TTI | Current Desktop | Current Mobile |
|------|-------------|------------|-----------------|----------------|
| Home | < 2s | < 3s | _____ s | _____ s |
| Dashboard | < 3s | < 4s | _____ s | _____ s |
| Analytics | < 4s | < 5s | _____ s | _____ s |

## Common Performance Issues & Solutions

### Issue: High LCP
**Symptoms**: Largest element takes > 2.5s to render
**Solutions**:
- [ ] Optimize hero images
- [ ] Preload critical resources
- [ ] Reduce server response time
- [ ] Remove render-blocking resources
- [ ] Use Next.js Image component

### Issue: High FID
**Symptoms**: Page unresponsive after load
**Solutions**:
- [ ] Reduce JavaScript execution time
- [ ] Implement code splitting
- [ ] Use web workers for heavy tasks
- [ ] Optimize third-party scripts
- [ ] Defer non-critical JavaScript

### Issue: High CLS
**Symptoms**: Content jumps during load
**Solutions**:
- [ ] Set dimensions for images
- [ ] Reserve space for dynamic content
- [ ] Use font-display: optional
- [ ] Avoid inserting content above fold
- [ ] Use skeleton loaders

### Issue: Large Bundle Size
**Symptoms**: Bundle > 1 MB
**Solutions**:
- [ ] Run bundle analyzer
- [ ] Remove unused dependencies
- [ ] Implement dynamic imports
- [ ] Use lighter library alternatives
- [ ] Enable tree-shaking

### Issue: Slow Mobile Performance
**Symptoms**: Mobile 3x slower than desktop
**Solutions**:
- [ ] Reduce JavaScript execution
- [ ] Optimize images for mobile
- [ ] Implement progressive enhancement
- [ ] Use adaptive loading
- [ ] Test on real devices

## DevTools Performance Checklist

### Chrome DevTools - Performance Tab
1. **Record page load**:
   - [ ] Record 5-10 seconds of page load
   - [ ] Check for long tasks (> 50ms)
   - [ ] Identify render-blocking resources
   - [ ] Check for forced reflows

2. **Analyze timeline**:
   - [ ] FCP happens early (< 1.8s)
   - [ ] LCP happens early (< 2.5s)
   - [ ] No long tasks block main thread
   - [ ] Smooth 60 FPS animations

### Chrome DevTools - Network Tab
1. **Analyze requests**:
   - [ ] < 50 requests on initial load
   - [ ] Total transfer size < 1 MB
   - [ ] Compression enabled (gzip/brotli)
   - [ ] No failed requests
   - [ ] Resources load in parallel

2. **Check resource priority**:
   - [ ] Critical resources have high priority
   - [ ] Images are lazy loaded
   - [ ] Third-party scripts deferred
   - [ ] Fonts preloaded

### Chrome DevTools - Lighthouse Tab
1. **Run audit**:
   - [ ] Performance: > 90 (desktop), > 85 (mobile)
   - [ ] Accessibility: > 95
   - [ ] Best Practices: > 95
   - [ ] SEO: > 90

2. **Review opportunities**:
   - [ ] Implement all "Opportunities" suggestions
   - [ ] Fix all "Diagnostics" warnings
   - [ ] Address all "Passed Audits" that can improve

## Continuous Monitoring Setup

### 1. GitHub Actions Integration
```yaml
# .github/workflows/performance.yml
name: Performance Audit

on: [pull_request]

jobs:
  lighthouse:
    runs-on: ubuntu-latest
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-node@v3
      - run: npm ci
      - run: npm run build
      - uses: treosh/lighthouse-ci-action@v9
        with:
          urls: |
            http://localhost:3000
            http://localhost:3000/dashboard
          budgetPath: ./lighthouse-budget.json
```

### 2. Budget Configuration
Create `lighthouse-budget.json`:
```json
{
  "bundles": [
    {
      "resourceSizes": [
        {
          "resourceType": "script",
          "budget": 500
        },
        {
          "resourceType": "stylesheet",
          "budget": 50
        },
        {
          "resourceType": "document",
          "budget": 18
        },
        {
          "resourceType": "total",
          "budget": 700
        }
      ]
    }
  ],
  "timings": [
    {
      "metric": "interactive",
      "budget": 5000
    },
    {
      "metric": "first-contentful-paint",
      "budget": 1800
    },
    {
      "metric": "largest-contentful-paint",
      "budget": 2500
    },
    {
      "metric": "cumulative-layout-shift",
      "budget": 0.1
    }
  ]
}
```

### 3. Real User Monitoring
```typescript
// lib/web-vitals.ts
import { getCLS, getFID, getFCP, getLCP, getTTFB, Metric } from 'web-vitals';

function sendToAnalytics(metric: Metric) {
  // Send to your analytics provider
  console.log(metric);

  // Example: Google Analytics 4
  if (typeof window !== 'undefined' && (window as any).gtag) {
    (window as any).gtag('event', metric.name, {
      value: Math.round(metric.name === 'CLS' ? metric.delta * 1000 : metric.delta),
      metric_id: metric.id,
      metric_value: metric.value,
      metric_delta: metric.delta,
    });
  }
}

export function reportWebVitals() {
  getCLS(sendToAnalytics);
  getFID(sendToAnalytics);
  getFCP(sendToAnalytics);
  getLCP(sendToAnalytics);
  getTTFB(sendToAnalytics);
}
```

## Sign-Off Checklist

Before marking performance work as complete:

### Development
- [ ] All optimizations implemented
- [ ] Bundle analysis shows improvements
- [ ] Local testing shows improved metrics
- [ ] No regression in functionality
- [ ] Documentation updated

### Testing
- [ ] Desktop performance tested
- [ ] Mobile performance tested
- [ ] Slow network tested (3G)
- [ ] Multiple browsers tested
- [ ] Real device testing completed

### Deployment
- [ ] Performance budget configured
- [ ] Lighthouse CI set up
- [ ] Monitoring in place
- [ ] Rollback plan prepared
- [ ] Team notified of changes

### Post-Deployment
- [ ] Production metrics verified
- [ ] RUM data collected
- [ ] No performance regressions
- [ ] User feedback positive
- [ ] Documentation finalized

---

**Performance Champion**: _______________
**Review Date**: _______________
**Next Review**: _______________
