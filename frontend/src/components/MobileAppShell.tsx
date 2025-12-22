/**
 * Mobile App Shell Component
 * Provides consistent mobile-optimized layout with performance features
 */

'use client';

import { useEffect, useState } from 'react';
import { ResourceHints } from './ResourceHints';
import { useNetworkStatus } from '@/hooks/useNetworkStatus';
import { setupPerformanceMonitoring } from '@/utils/performanceMonitoring';
import { syncOfflineData } from '@/utils/offlineSync';

interface MobileAppShellProps {
  children: React.ReactNode;
}

export function MobileAppShell({ children }: MobileAppShellProps) {
  const { isOnline, quality } = useNetworkStatus();
  const [isSyncing, setIsSyncing] = useState(false);
  const [showOfflineBanner, setShowOfflineBanner] = useState(false);

  // Setup performance monitoring
  useEffect(() => {
    if (typeof window !== 'undefined') {
      setupPerformanceMonitoring((metrics) => {
        // Log metrics in development
        if (process.env.NODE_ENV === 'development') {
          console.log('📊 Performance Metrics:', {
            FCP: metrics.fcp?.toFixed(0) + 'ms',
            LCP: metrics.lcp?.toFixed(0) + 'ms',
            FID: metrics.fid?.toFixed(0) + 'ms',
            CLS: metrics.cls?.toFixed(3),
          });
        }
      });
    }
  }, []);

  // Handle online/offline transitions
  useEffect(() => {
    if (!isOnline) {
      setShowOfflineBanner(true);
    } else {
      setShowOfflineBanner(false);
      // Sync offline data when back online
      syncPendingData();
    }
  }, [isOnline]);

  // Sync pending offline data
  async function syncPendingData() {
    if (isSyncing) return;

    setIsSyncing(true);
    try {
      const results = await syncOfflineData();
      if (results.success > 0) {
        console.log(`✅ Synced ${results.success} pending actions`);
      }
      if (results.failed > 0) {
        console.warn(`⚠️ Failed to sync ${results.failed} actions`);
      }
    } catch (error) {
      console.error('Sync error:', error);
    } finally {
      setIsSyncing(false);
    }
  }

  return (
    <>
      <ResourceHints />

      {/* Offline Banner */}
      {showOfflineBanner && (
        <div className="fixed top-0 left-0 right-0 z-50 bg-yellow-500 text-white px-4 py-2 text-sm font-medium text-center">
          <div className="flex items-center justify-center gap-2">
            <svg className="w-4 h-4" fill="none" stroke="currentColor" viewBox="0 0 24 24">
              <path
                strokeLinecap="round"
                strokeLinejoin="round"
                strokeWidth={2}
                d="M18.364 5.636a9 9 0 010 12.728m0 0l-2.829-2.829m2.829 2.829L21 21M15.536 8.464a5 5 0 010 7.072m0 0l-2.829-2.829m-4.243 2.829a4.978 4.978 0 01-1.414-2.83m-1.414 5.658a9 9 0 01-2.167-9.238m7.824 2.167a1 1 0 111.414 1.414m-1.414-1.414L3 3m8.293 8.293l1.414 1.414"
              />
            </svg>
            You're offline. Some features may be limited.
          </div>
        </div>
      )}

      {/* Sync Indicator */}
      {isSyncing && (
        <div className="fixed top-12 left-0 right-0 z-50 bg-blue-500 text-white px-4 py-2 text-sm font-medium text-center">
          <div className="flex items-center justify-center gap-2">
            <div className="animate-spin rounded-full h-4 w-4 border-2 border-white border-t-transparent"></div>
            Syncing offline changes...
          </div>
        </div>
      )}

      {/* Network Quality Indicator (Development only) */}
      {process.env.NODE_ENV === 'development' && (
        <div className="fixed bottom-4 right-4 z-50 bg-gray-900 text-white px-3 py-2 rounded-lg text-xs font-mono shadow-lg">
          <div className="flex items-center gap-2">
            <div
              className={`w-2 h-2 rounded-full ${
                quality === 'high'
                  ? 'bg-green-400'
                  : quality === 'medium'
                  ? 'bg-yellow-400'
                  : quality === 'low'
                  ? 'bg-orange-400'
                  : 'bg-red-400'
              }`}
            />
            {quality.toUpperCase()}
          </div>
        </div>
      )}

      {/* Main Content */}
      <div className="min-h-screen">{children}</div>
    </>
  );
}
