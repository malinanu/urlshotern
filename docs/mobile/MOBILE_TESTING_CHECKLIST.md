# Mobile Performance Testing Checklist

## Pre-Testing Setup

- [ ] Build production bundle: `npm run build`
- [ ] Start production server: `npm start`
- [ ] Clear browser cache and service workers
- [ ] Disable browser extensions
- [ ] Use incognito/private mode

## Performance Testing

### Lighthouse Audit

1. **Desktop Test**
   - [ ] Open Chrome DevTools
   - [ ] Run Lighthouse audit
   - [ ] Performance score > 90
   - [ ] Accessibility score > 90
   - [ ] Best Practices score > 90
   - [ ] SEO score > 90
   - [ ] PWA score > 90

2. **Mobile Test**
   - [ ] Switch to mobile emulation
   - [ ] Run Lighthouse audit (mobile)
   - [ ] Performance score > 85
   - [ ] Check Core Web Vitals
   - [ ] FCP < 1.8s
   - [ ] LCP < 2.5s
   - [ ] FID < 100ms
   - [ ] CLS < 0.1

### Network Conditions

Test on different network speeds:

- [ ] **Fast 4G** (4 Mbps down, 3 Mbps up, 20ms RTT)
  - [ ] Page loads < 3s
  - [ ] Animations smooth
  - [ ] Images load quickly

- [ ] **Slow 3G** (400 Kbps down, 400 Kbps up, 400ms RTT)
  - [ ] Page loads < 10s
  - [ ] Reduced quality images load
  - [ ] Critical content prioritized
  - [ ] Loading states visible

- [ ] **Offline**
  - [ ] Cached pages load
  - [ ] Offline indicator shows
  - [ ] Queued actions saved
  - [ ] Graceful error messages

### Device Testing

Test on actual devices:

- [ ] **iOS Devices**
  - [ ] iPhone SE (small screen)
  - [ ] iPhone 14/15 (standard)
  - [ ] iPhone 14/15 Pro Max (large)
  - [ ] iPad (tablet)

- [ ] **Android Devices**
  - [ ] Low-end (< 2GB RAM)
  - [ ] Mid-range (2-4GB RAM)
  - [ ] High-end (> 4GB RAM)
  - [ ] Tablet

## PWA Functionality

### Installation

- [ ] Install prompt appears
- [ ] Installation successful
- [ ] App launches in standalone mode
- [ ] App icon appears correctly
- [ ] Splash screen displays
- [ ] No browser chrome visible

### Offline Support

- [ ] Service worker registers
- [ ] Pages cached on first visit
- [ ] Offline pages load from cache
- [ ] API responses cached
- [ ] Offline indicator visible
- [ ] Queued actions sync when online

### Push Notifications

- [ ] Notification permission prompt
- [ ] Notifications received
- [ ] Notification click opens app
- [ ] Notifications work when app closed

### Share Target

- [ ] App appears in share menu
- [ ] Shared URLs open in app
- [ ] Share data processed correctly

## Feature Testing

### URL Shortening

- [ ] Create URL (online)
- [ ] Create URL (offline - queued)
- [ ] Copy short URL
- [ ] QR code generates
- [ ] Custom alias works
- [ ] Expiration date works

### Analytics Page

- [ ] Loads within 4s on mobile
- [ ] Charts lazy load
- [ ] Scroll performance smooth
- [ ] Touch interactions responsive
- [ ] Data caches for offline
- [ ] Network indicator shows

### Dashboard

- [ ] URL list loads
- [ ] Virtual scrolling works
- [ ] Infinite scroll works
- [ ] Pull to refresh works
- [ ] Swipe actions work
- [ ] Search/filter responsive

### Touch Gestures

- [ ] Swipe left to delete
- [ ] Swipe right to edit
- [ ] Long press for options
- [ ] Pull to refresh
- [ ] Pinch to zoom (if applicable)
- [ ] Double tap to select

## Performance Metrics

### Target Metrics

Record actual metrics and compare to targets:

| Metric | Target | Mobile 4G | Mobile 3G | Desktop |
|--------|--------|-----------|-----------|---------|
| FCP | < 1.8s | _____ | _____ | _____ |
| LCP | < 2.5s | _____ | _____ | _____ |
| FID | < 100ms | _____ | _____ | _____ |
| CLS | < 0.1 | _____ | _____ | _____ |
| TTI | < 3.8s | _____ | _____ | _____ |
| TBT | < 200ms | _____ | _____ | _____ |
| Speed Index | < 3.4s | _____ | _____ | _____ |

### Resource Analysis

- [ ] Total page size < 500KB (initial load)
- [ ] Total requests < 50 (initial load)
- [ ] JavaScript size < 200KB
- [ ] CSS size < 50KB
- [ ] Images optimized
- [ ] Fonts optimized

### Battery Impact

- [ ] No excessive CPU usage
- [ ] Animations GPU accelerated
- [ ] Background tasks minimal
- [ ] Battery usage acceptable

## Accessibility

- [ ] Screen reader compatible
- [ ] Keyboard navigation works
- [ ] Focus indicators visible
- [ ] Touch targets ≥ 44x44px
- [ ] Color contrast sufficient
- [ ] ARIA labels present

## Edge Cases

- [ ] Very long URLs display correctly
- [ ] Many URLs (1000+) perform well
- [ ] Network switches (online/offline)
- [ ] App backgrounded/foregrounded
- [ ] Low battery mode
- [ ] Low memory devices
- [ ] Slow processors
- [ ] High DPI screens
- [ ] Landscape/portrait orientation

## Browser Compatibility

- [ ] Chrome (latest)
- [ ] Safari (latest)
- [ ] Firefox (latest)
- [ ] Edge (latest)
- [ ] Samsung Internet
- [ ] UC Browser

## Security & Privacy

- [ ] HTTPS enforced
- [ ] Sensitive data not cached
- [ ] Tokens stored securely
- [ ] CSP headers configured
- [ ] No mixed content
- [ ] Privacy policy linked

## Monitoring Setup

- [ ] Performance monitoring enabled
- [ ] Error tracking configured
- [ ] Analytics tracking works
- [ ] Service worker errors logged
- [ ] Network errors logged

## Production Checklist

- [ ] All tests passed
- [ ] Performance budget met
- [ ] No console errors
- [ ] No console warnings
- [ ] Service worker deployed
- [ ] Cache headers configured
- [ ] Compression enabled (gzip/brotli)
- [ ] CDN configured
- [ ] HTTP/2 enabled
- [ ] SSL certificate valid

## Performance Testing Tools

### Online Tools

- **PageSpeed Insights**: https://pagespeed.web.dev/
- **WebPageTest**: https://www.webpagetest.org/
- **GTmetrix**: https://gtmetrix.com/
- **Pingdom**: https://tools.pingdom.com/

### Chrome DevTools

- **Lighthouse**: Audits → Run audit
- **Performance**: Record and analyze
- **Network**: Throttling and monitoring
- **Coverage**: Unused JavaScript/CSS
- **Memory**: Heap snapshots

### Command Line

```bash
# Lighthouse CLI
lighthouse https://your-domain.com --view --preset=desktop
lighthouse https://your-domain.com --view --preset=mobile

# WebPageTest API
webpagetest test https://your-domain.com --key YOUR_API_KEY

# Chrome DevTools Protocol
chrome --headless --disable-gpu --dump-dom https://your-domain.com
```

## Automated Testing Script

```javascript
// performance-test.js
const lighthouse = require('lighthouse');
const chromeLauncher = require('chrome-launcher');

async function runLighthouse(url) {
  const chrome = await chromeLauncher.launch({chromeFlags: ['--headless']});
  const options = {
    logLevel: 'info',
    output: 'html',
    onlyCategories: ['performance', 'pwa'],
    port: chrome.port
  };

  const runnerResult = await lighthouse(url, options);

  // Check performance budget
  const performance = runnerResult.lhr.categories.performance.score * 100;
  const pwa = runnerResult.lhr.categories.pwa.score * 100;

  console.log('Performance score:', performance);
  console.log('PWA score:', pwa);

  if (performance < 85 || pwa < 90) {
    throw new Error('Performance budget not met');
  }

  await chrome.kill();
}

runLighthouse('https://your-domain.com');
```

## Post-Deployment Monitoring

Monitor these metrics for 7 days after deployment:

- [ ] Average FCP
- [ ] Average LCP
- [ ] Average FID
- [ ] Average CLS
- [ ] Service worker activation rate
- [ ] Cache hit rate
- [ ] Offline usage
- [ ] Install rate
- [ ] Uninstall rate
- [ ] Error rate
- [ ] User engagement

## Optimization Iterations

Track improvements over time:

| Version | FCP | LCP | FID | CLS | TTI | Notes |
|---------|-----|-----|-----|-----|-----|-------|
| Before | 3.2s | 5.8s | 250ms | 0.25 | 15.8s | Baseline |
| v1.0 | _____ | _____ | _____ | _____ | _____ | Initial optimizations |
| v1.1 | _____ | _____ | _____ | _____ | _____ | Further improvements |

## Sign-Off

- [ ] All critical tests passed
- [ ] Performance targets met
- [ ] No blocking issues
- [ ] Documentation updated
- [ ] Team reviewed
- [ ] Ready for production

**Tested by**: _________________
**Date**: _________________
**Approval**: _________________
