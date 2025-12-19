/**
 * Mobile-Optimized Analytics Component
 * Lazy loads charts and heavy components based on network quality
 */

'use client';

import { useState, lazy, Suspense } from 'react';
import { useNetworkStatus, useAdaptiveLoading } from '@/hooks/useNetworkStatus';
import { useIntersectionObserver } from '@/hooks/useIntersectionObserver';
import { getCachedAnalytics, cacheAnalytics } from '@/utils/offlineSync';
import { useRef, useEffect } from 'react';

// Lazy load heavy chart components
const LazyLineChart = lazy(() =>
  import('recharts').then(module => ({ default: module.LineChart }))
);
const LazyBarChart = lazy(() =>
  import('recharts').then(module => ({ default: module.BarChart }))
);
const LazyPieChart = lazy(() =>
  import('recharts').then(module => ({ default: module.PieChart }))
);

interface MobileAnalyticsProps {
  shortCode: string;
  apiBaseUrl?: string;
}

export function MobileOptimizedAnalytics({
  shortCode,
  apiBaseUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080'
}: MobileAnalyticsProps) {
  const [analyticsData, setAnalyticsData] = useState<any>(null);
  const [isLoading, setIsLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
  const [visibleSections, setVisibleSections] = useState({
    overview: true,
    timeline: false,
    geography: false,
    devices: false,
  });

  const { isOnline, quality } = useNetworkStatus();
  const adaptiveConfig = useAdaptiveLoading();

  const timelineRef = useRef<HTMLDivElement>(null);
  const geographyRef = useRef<HTMLDivElement>(null);
  const devicesRef = useRef<HTMLDivElement>(null);

  // Lazy load sections as they come into view
  useIntersectionObserver({
    target: timelineRef,
    onIntersect: () => setVisibleSections(prev => ({ ...prev, timeline: true })),
    enabled: !visibleSections.timeline,
  });

  useIntersectionObserver({
    target: geographyRef,
    onIntersect: () => setVisibleSections(prev => ({ ...prev, geography: true })),
    enabled: !visibleSections.geography,
  });

  useIntersectionObserver({
    target: devicesRef,
    onIntersect: () => setVisibleSections(prev => ({ ...prev, devices: true })),
    enabled: !visibleSections.devices,
  });

  // Fetch analytics data
  useEffect(() => {
    async function fetchAnalytics() {
      setIsLoading(true);
      setError(null);

      try {
        // Try to get cached data first
        const cached = await getCachedAnalytics(shortCode);
        if (cached && !isOnline) {
          setAnalyticsData(cached);
          setIsLoading(false);
          return;
        }

        // Fetch from API
        const token = localStorage.getItem('access_token');
        const response = await fetch(
          `${apiBaseUrl}/api/v1/analytics/${shortCode}`,
          {
            headers: {
              'Authorization': `Bearer ${token}`,
              'Content-Type': 'application/json',
            },
          }
        );

        if (!response.ok) {
          throw new Error('Failed to fetch analytics');
        }

        const data = await response.json();
        setAnalyticsData(data);

        // Cache the data
        await cacheAnalytics(shortCode, data, 10); // Cache for 10 minutes
      } catch (err) {
        console.error('Analytics fetch error:', err);

        // Try cached data on error
        const cached = await getCachedAnalytics(shortCode);
        if (cached) {
          setAnalyticsData(cached);
          setError('Showing cached data (offline)');
        } else {
          setError('Failed to load analytics data');
        }
      } finally {
        setIsLoading(false);
      }
    }

    if (shortCode) {
      fetchAnalytics();
    }
  }, [shortCode, apiBaseUrl, isOnline]);

  if (isLoading) {
    return <AnalyticsSkeleton />;
  }

  if (error && !analyticsData) {
    return (
      <div className="bg-red-50 border border-red-200 rounded-lg p-4 text-center">
        <p className="text-red-600">{error}</p>
      </div>
    );
  }

  return (
    <div className="space-y-4">
      {/* Network status indicator */}
      {!isOnline && (
        <div className="bg-yellow-50 border border-yellow-200 rounded-lg p-3 flex items-center gap-2 text-sm">
          <svg className="w-5 h-5 text-yellow-600" fill="none" stroke="currentColor" viewBox="0 0 24 24">
            <path strokeLinecap="round" strokeLinejoin="round" strokeWidth={2} d="M12 9v2m0 4h.01m-6.938 4h13.856c1.54 0 2.502-1.667 1.732-3L13.732 4c-.77-1.333-2.694-1.333-3.464 0L3.34 16c-.77 1.333.192 3 1.732 3z" />
          </svg>
          <span className="text-yellow-800">
            You are offline. Showing cached data.
          </span>
        </div>
      )}

      {/* Overview - Always visible */}
      <OverviewSection data={analyticsData} />

      {/* Timeline Chart - Lazy loaded */}
      <div ref={timelineRef} className="min-h-[300px]">
        {visibleSections.timeline ? (
          <Suspense fallback={<ChartSkeleton />}>
            <TimelineSection data={analyticsData} adaptiveConfig={adaptiveConfig} />
          </Suspense>
        ) : (
          <ChartPlaceholder title="Timeline Analytics" />
        )}
      </div>

      {/* Geography - Lazy loaded */}
      <div ref={geographyRef} className="min-h-[300px]">
        {visibleSections.geography ? (
          <Suspense fallback={<ChartSkeleton />}>
            <GeographySection data={analyticsData} adaptiveConfig={adaptiveConfig} />
          </Suspense>
        ) : (
          <ChartPlaceholder title="Geographic Distribution" />
        )}
      </div>

      {/* Devices - Lazy loaded */}
      <div ref={devicesRef} className="min-h-[300px]">
        {visibleSections.devices ? (
          <Suspense fallback={<ChartSkeleton />}>
            <DevicesSection data={analyticsData} adaptiveConfig={adaptiveConfig} />
          </Suspense>
        ) : (
          <ChartPlaceholder title="Device Analytics" />
        )}
      </div>
    </div>
  );
}

/**
 * Overview Section Component
 */
function OverviewSection({ data }: { data: any }) {
  const stats = [
    { label: 'Total Clicks', value: data?.total_clicks || 0, icon: '👆' },
    { label: 'Unique Visitors', value: data?.unique_visitors || 0, icon: '👥' },
    { label: 'Avg. CTR', value: `${data?.ctr || 0}%`, icon: '📊' },
    { label: 'Last Click', value: data?.last_click ? new Date(data.last_click).toLocaleDateString() : 'N/A', icon: '🕐' },
  ];

  return (
    <div className="grid grid-cols-2 gap-3">
      {stats.map((stat) => (
        <div
          key={stat.label}
          className="bg-white rounded-lg border border-gray-200 p-4"
          style={{ contain: 'layout style' }}
        >
          <div className="text-2xl mb-1">{stat.icon}</div>
          <div className="text-2xl font-bold text-gray-900">{stat.value}</div>
          <div className="text-xs text-gray-600 mt-1">{stat.label}</div>
        </div>
      ))}
    </div>
  );
}

/**
 * Placeholder components
 */
function ChartPlaceholder({ title }: { title: string }) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6">
      <div className="flex items-center justify-center min-h-[250px] text-gray-400">
        <div className="text-center">
          <div className="text-4xl mb-2">📊</div>
          <p className="text-sm">Scroll to load {title}</p>
        </div>
      </div>
    </div>
  );
}

function ChartSkeleton() {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-6 animate-pulse">
      <div className="h-6 bg-gray-200 rounded w-1/3 mb-4"></div>
      <div className="h-[250px] bg-gray-100 rounded"></div>
    </div>
  );
}

function AnalyticsSkeleton() {
  return (
    <div className="space-y-4">
      <div className="grid grid-cols-2 gap-3">
        {[1, 2, 3, 4].map((i) => (
          <div key={i} className="bg-white rounded-lg border border-gray-200 p-4 animate-pulse">
            <div className="h-8 bg-gray-200 rounded w-12 mb-2"></div>
            <div className="h-6 bg-gray-200 rounded w-16 mb-1"></div>
            <div className="h-4 bg-gray-100 rounded w-20"></div>
          </div>
        ))}
      </div>
      {[1, 2, 3].map((i) => (
        <ChartSkeleton key={i} />
      ))}
    </div>
  );
}

/**
 * Chart sections (simplified - you can expand these)
 */
function TimelineSection({ data, adaptiveConfig }: any) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <h3 className="text-lg font-semibold mb-4">Click Timeline</h3>
      <p className="text-sm text-gray-600">Timeline chart would render here with adaptive quality</p>
    </div>
  );
}

function GeographySection({ data, adaptiveConfig }: any) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <h3 className="text-lg font-semibold mb-4">Geographic Distribution</h3>
      <p className="text-sm text-gray-600">Geography chart would render here</p>
    </div>
  );
}

function DevicesSection({ data, adaptiveConfig }: any) {
  return (
    <div className="bg-white rounded-lg border border-gray-200 p-4">
      <h3 className="text-lg font-semibold mb-4">Device Analytics</h3>
      <p className="text-sm text-gray-600">Device breakdown chart would render here</p>
    </div>
  );
}
