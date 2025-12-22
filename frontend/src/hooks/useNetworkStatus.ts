/**
 * React Hook for Network Status Monitoring
 * Provides real-time network quality information
 */

import { useState, useEffect } from 'react';
import {
  getNetworkStatus,
  addNetworkChangeListener,
  NetworkStatus,
  NetworkQuality,
  getAdaptiveConfig,
  AdaptiveConfig,
} from '@/utils/networkDetection';

/**
 * Hook to monitor network status and quality
 */
export function useNetworkStatus() {
  const [networkStatus, setNetworkStatus] = useState<NetworkStatus>(
    getNetworkStatus()
  );

  useEffect(() => {
    // Update network status when it changes
    const cleanup = addNetworkChangeListener(() => {
      setNetworkStatus(getNetworkStatus());
    });

    return cleanup;
  }, []);

  return networkStatus;
}

/**
 * Hook to get adaptive configuration based on network quality
 */
export function useAdaptiveLoading(): AdaptiveConfig {
  const { quality } = useNetworkStatus();
  const [config, setConfig] = useState<AdaptiveConfig>(
    getAdaptiveConfig(quality)
  );

  useEffect(() => {
    setConfig(getAdaptiveConfig(quality));
  }, [quality]);

  return config;
}

/**
 * Hook to check if device is online
 */
export function useOnlineStatus(): boolean {
  const { isOnline } = useNetworkStatus();
  return isOnline;
}

/**
 * Hook to get network quality
 */
export function useNetworkQuality(): NetworkQuality {
  const { quality } = useNetworkStatus();
  return quality;
}

/**
 * Hook for conditional rendering based on network quality
 */
export function useNetworkAwareRender() {
  const { quality } = useNetworkStatus();

  return {
    shouldRenderImages: quality !== 'offline',
    shouldRenderAnimations: quality === 'high' || quality === 'medium',
    shouldEnablePrefetch: quality === 'high',
    shouldEnableLazyLoad: quality === 'low' || quality === 'medium',
  };
}
