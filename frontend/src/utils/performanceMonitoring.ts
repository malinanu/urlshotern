/**
 * Mobile Performance Monitoring Utilities
 * Track and optimize mobile app performance
 */

export interface PerformanceMetrics {
  fcp: number | null; // First Contentful Paint
  lcp: number | null; // Largest Contentful Paint
  fid: number | null; // First Input Delay
  cls: number | null; // Cumulative Layout Shift
  ttfb: number | null; // Time to First Byte
  tti: number | null; // Time to Interactive
  tbt: number | null; // Total Blocking Time
}

export interface ResourceTiming {
  name: string;
  duration: number;
  size: number;
  type: string;
}

/**
 * Get Core Web Vitals metrics
 */
export function getCoreWebVitals(): Promise<PerformanceMetrics> {
  return new Promise((resolve) => {
    const metrics: PerformanceMetrics = {
      fcp: null,
      lcp: null,
      fid: null,
      cls: null,
      ttfb: null,
      tti: null,
      tbt: null,
    };

    // First Contentful Paint
    const paintEntries = performance.getEntriesByType('paint');
    const fcpEntry = paintEntries.find((entry) => entry.name === 'first-contentful-paint');
    if (fcpEntry) {
      metrics.fcp = fcpEntry.startTime;
    }

    // Time to First Byte
    const navigationEntries = performance.getEntriesByType('navigation') as PerformanceNavigationTiming[];
    if (navigationEntries.length > 0) {
      metrics.ttfb = navigationEntries[0].responseStart - navigationEntries[0].requestStart;
    }

    // Largest Contentful Paint
    if ('PerformanceObserver' in window) {
      try {
        const lcpObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries();
          const lastEntry = entries[entries.length - 1] as any;
          metrics.lcp = lastEntry.renderTime || lastEntry.loadTime;
        });
        lcpObserver.observe({ type: 'largest-contentful-paint', buffered: true });

        // First Input Delay
        const fidObserver = new PerformanceObserver((list) => {
          const entries = list.getEntries();
          entries.forEach((entry: any) => {
            metrics.fid = entry.processingStart - entry.startTime;
          });
        });
        fidObserver.observe({ type: 'first-input', buffered: true });

        // Cumulative Layout Shift
        let clsValue = 0;
        const clsObserver = new PerformanceObserver((list) => {
          for (const entry of list.getEntries()) {
            if (!(entry as any).hadRecentInput) {
              clsValue += (entry as any).value;
              metrics.cls = clsValue;
            }
          }
        });
        clsObserver.observe({ type: 'layout-shift', buffered: true });

        // Resolve after a delay to collect metrics
        setTimeout(() => {
          resolve(metrics);
        }, 3000);
      } catch (error) {
        console.error('Performance observation error:', error);
        resolve(metrics);
      }
    } else {
      resolve(metrics);
    }
  });
}

/**
 * Get resource loading performance
 */
export function getResourceTimings(): ResourceTiming[] {
  const resources = performance.getEntriesByType('resource') as PerformanceResourceTiming[];

  return resources.map((resource) => ({
    name: resource.name,
    duration: resource.duration,
    size: resource.transferSize || 0,
    type: getResourceType(resource.name),
  }));
}

/**
 * Determine resource type from URL
 */
function getResourceType(url: string): string {
  if (url.match(/\.(js|mjs)$/)) return 'script';
  if (url.match(/\.css$/)) return 'stylesheet';
  if (url.match(/\.(png|jpg|jpeg|gif|svg|webp)$/)) return 'image';
  if (url.match(/\.(woff|woff2|ttf|eot)$/)) return 'font';
  if (url.includes('/api/')) return 'api';
  return 'other';
}

/**
 * Calculate Total Blocking Time (TBT)
 */
export function getTotalBlockingTime(): number {
  let tbt = 0;
  const longTasks = performance.getEntriesByType('longtask') as any[];

  longTasks.forEach((task) => {
    if (task.duration > 50) {
      tbt += task.duration - 50;
    }
  });

  return tbt;
}

/**
 * Monitor JavaScript execution time
 */
export function measureJavaScriptExecution(name: string): () => void {
  const startMark = `${name}-start`;
  const endMark = `${name}-end`;

  performance.mark(startMark);

  return () => {
    performance.mark(endMark);
    performance.measure(name, startMark, endMark);

    const measure = performance.getEntriesByName(name)[0];
    console.log(`${name}: ${measure.duration.toFixed(2)}ms`);

    // Clean up marks
    performance.clearMarks(startMark);
    performance.clearMarks(endMark);
    performance.clearMeasures(name);

    return measure.duration;
  };
}

/**
 * Monitor memory usage (if available)
 */
export function getMemoryInfo(): any {
  if ('memory' in performance) {
    return (performance as any).memory;
  }
  return null;
}

/**
 * Get network information for performance analysis
 */
export function getNetworkInfo(): any {
  const nav = navigator as any;
  const connection = nav.connection || nav.mozConnection || nav.webkitConnection;

  if (connection) {
    return {
      effectiveType: connection.effectiveType,
      downlink: connection.downlink,
      rtt: connection.rtt,
      saveData: connection.saveData,
    };
  }

  return null;
}

/**
 * Performance budget checker
 */
export interface PerformanceBudget {
  fcp: number;
  lcp: number;
  fid: number;
  cls: number;
  totalSize: number;
  totalRequests: number;
}

export function checkPerformanceBudget(
  budget: PerformanceBudget
): Promise<{ passed: boolean; violations: string[] }> {
  return new Promise(async (resolve) => {
    const violations: string[] = [];
    const metrics = await getCoreWebVitals();
    const resources = getResourceTimings();

    // Check FCP
    if (metrics.fcp && metrics.fcp > budget.fcp) {
      violations.push(`FCP: ${metrics.fcp.toFixed(0)}ms exceeds budget of ${budget.fcp}ms`);
    }

    // Check LCP
    if (metrics.lcp && metrics.lcp > budget.lcp) {
      violations.push(`LCP: ${metrics.lcp.toFixed(0)}ms exceeds budget of ${budget.lcp}ms`);
    }

    // Check FID
    if (metrics.fid && metrics.fid > budget.fid) {
      violations.push(`FID: ${metrics.fid.toFixed(0)}ms exceeds budget of ${budget.fid}ms`);
    }

    // Check CLS
    if (metrics.cls && metrics.cls > budget.cls) {
      violations.push(`CLS: ${metrics.cls.toFixed(3)} exceeds budget of ${budget.cls}`);
    }

    // Check total size
    const totalSize = resources.reduce((sum, r) => sum + r.size, 0);
    if (totalSize > budget.totalSize) {
      violations.push(
        `Total size: ${(totalSize / 1024).toFixed(0)}KB exceeds budget of ${(budget.totalSize / 1024).toFixed(0)}KB`
      );
    }

    // Check total requests
    if (resources.length > budget.totalRequests) {
      violations.push(
        `Total requests: ${resources.length} exceeds budget of ${budget.totalRequests}`
      );
    }

    resolve({
      passed: violations.length === 0,
      violations,
    });
  });
}

/**
 * Log performance metrics to console
 */
export async function logPerformanceMetrics() {
  const metrics = await getCoreWebVitals();
  const resources = getResourceTimings();
  const networkInfo = getNetworkInfo();
  const memoryInfo = getMemoryInfo();

  console.group('📊 Performance Metrics');

  console.group('Core Web Vitals');
  if (metrics.fcp) console.log(`FCP: ${metrics.fcp.toFixed(0)}ms`);
  if (metrics.lcp) console.log(`LCP: ${metrics.lcp.toFixed(0)}ms`);
  if (metrics.fid) console.log(`FID: ${metrics.fid.toFixed(0)}ms`);
  if (metrics.cls) console.log(`CLS: ${metrics.cls.toFixed(3)}`);
  if (metrics.ttfb) console.log(`TTFB: ${metrics.ttfb.toFixed(0)}ms`);
  console.groupEnd();

  console.group('Resources');
  const totalSize = resources.reduce((sum, r) => sum + r.size, 0);
  console.log(`Total requests: ${resources.length}`);
  console.log(`Total size: ${(totalSize / 1024).toFixed(2)}KB`);

  const byType = resources.reduce((acc, r) => {
    acc[r.type] = (acc[r.type] || 0) + 1;
    return acc;
  }, {} as Record<string, number>);
  console.table(byType);
  console.groupEnd();

  if (networkInfo) {
    console.group('Network');
    console.log(`Effective type: ${networkInfo.effectiveType}`);
    console.log(`Downlink: ${networkInfo.downlink}Mbps`);
    console.log(`RTT: ${networkInfo.rtt}ms`);
    console.log(`Save data: ${networkInfo.saveData}`);
    console.groupEnd();
  }

  if (memoryInfo) {
    console.group('Memory');
    console.log(`JS Heap Size: ${(memoryInfo.usedJSHeapSize / 1048576).toFixed(2)}MB`);
    console.log(`JS Heap Limit: ${(memoryInfo.jsHeapSizeLimit / 1048576).toFixed(2)}MB`);
    console.groupEnd();
  }

  console.groupEnd();
}

/**
 * Setup automatic performance monitoring
 */
export function setupPerformanceMonitoring(
  reportCallback?: (metrics: PerformanceMetrics) => void
) {
  // Monitor on page load
  if (document.readyState === 'complete') {
    monitorPerformance();
  } else {
    window.addEventListener('load', monitorPerformance);
  }

  async function monitorPerformance() {
    const metrics = await getCoreWebVitals();

    // Log to console in development
    if (process.env.NODE_ENV === 'development') {
      await logPerformanceMetrics();
    }

    // Report to callback
    if (reportCallback) {
      reportCallback(metrics);
    }

    // Report to analytics (implement your analytics service here)
    // sendToAnalytics(metrics);
  }
}

/**
 * Battery-aware performance throttling
 */
export async function shouldThrottlePerformance(): Promise<boolean> {
  if ('getBattery' in navigator) {
    try {
      const battery = await (navigator as any).getBattery();
      // Throttle if battery is low and not charging
      return battery.level < 0.2 && !battery.charging;
    } catch (error) {
      return false;
    }
  }
  return false;
}

/**
 * Device capability detection
 */
export interface DeviceCapabilities {
  isMobile: boolean;
  hasTouch: boolean;
  deviceMemory: number | null;
  hardwareConcurrency: number;
  maxTouchPoints: number;
  screenSize: { width: number; height: number };
  devicePixelRatio: number;
}

export function getDeviceCapabilities(): DeviceCapabilities {
  return {
    isMobile: /Android|webOS|iPhone|iPad|iPod|BlackBerry|IEMobile|Opera Mini/i.test(
      navigator.userAgent
    ),
    hasTouch: 'ontouchstart' in window || navigator.maxTouchPoints > 0,
    deviceMemory: (navigator as any).deviceMemory || null,
    hardwareConcurrency: navigator.hardwareConcurrency || 1,
    maxTouchPoints: navigator.maxTouchPoints || 0,
    screenSize: {
      width: window.screen.width,
      height: window.screen.height,
    },
    devicePixelRatio: window.devicePixelRatio || 1,
  };
}

/**
 * Adaptive performance configuration
 */
export function getAdaptivePerformanceConfig() {
  const capabilities = getDeviceCapabilities();
  const memory = capabilities.deviceMemory;
  const cores = capabilities.hardwareConcurrency;

  // Low-end device (< 2GB RAM, < 2 cores)
  if ((memory && memory < 2) || cores < 2) {
    return {
      enableAnimations: false,
      enableTransitions: false,
      enableShadows: false,
      imageQuality: 0.6,
      maxConcurrentRequests: 2,
      virtualScrollOverscan: 2,
      debounceDelay: 500,
    };
  }

  // Mid-range device (2-4GB RAM, 2-4 cores)
  if ((memory && memory < 4) || cores < 4) {
    return {
      enableAnimations: true,
      enableTransitions: true,
      enableShadows: false,
      imageQuality: 0.75,
      maxConcurrentRequests: 4,
      virtualScrollOverscan: 5,
      debounceDelay: 300,
    };
  }

  // High-end device
  return {
    enableAnimations: true,
    enableTransitions: true,
    enableShadows: true,
    imageQuality: 0.9,
    maxConcurrentRequests: 6,
    virtualScrollOverscan: 10,
    debounceDelay: 100,
  };
}
