/**
 * Network Detection and Adaptive Loading Utilities
 * Optimizes resource loading based on network conditions
 */

export type ConnectionType = '4g' | '3g' | '2g' | 'slow-2g' | 'unknown';
export type NetworkQuality = 'high' | 'medium' | 'low' | 'offline';

interface NetworkInformation extends EventTarget {
  downlink?: number;
  effectiveType?: ConnectionType;
  rtt?: number;
  saveData?: boolean;
  onchange?: EventListener;
}

interface NavigatorWithConnection extends Navigator {
  connection?: NetworkInformation;
  mozConnection?: NetworkInformation;
  webkitConnection?: NetworkInformation;
}

/**
 * Get the current network connection information
 */
export function getNetworkInformation(): NetworkInformation | null {
  const nav = navigator as NavigatorWithConnection;
  return nav.connection || nav.mozConnection || nav.webkitConnection || null;
}

/**
 * Determine network quality based on connection information
 */
export function getNetworkQuality(): NetworkQuality {
  if (!navigator.onLine) {
    return 'offline';
  }

  const connection = getNetworkInformation();

  if (!connection) {
    // If Network Information API is not supported, assume medium quality
    return 'medium';
  }

  const { effectiveType, saveData, downlink, rtt } = connection;

  // User has explicitly enabled data saver mode
  if (saveData) {
    return 'low';
  }

  // Check effective connection type
  if (effectiveType) {
    switch (effectiveType) {
      case '4g':
        return 'high';
      case '3g':
        return 'medium';
      case '2g':
      case 'slow-2g':
        return 'low';
      default:
        return 'medium';
    }
  }

  // Fallback to bandwidth and latency checks
  if (downlink !== undefined && rtt !== undefined) {
    if (downlink >= 5 && rtt <= 100) {
      return 'high';
    } else if (downlink >= 1.5 && rtt <= 300) {
      return 'medium';
    } else {
      return 'low';
    }
  }

  return 'medium';
}

/**
 * Check if the user has enabled data saver mode
 */
export function isDataSaverEnabled(): boolean {
  const connection = getNetworkInformation();
  return connection?.saveData || false;
}

/**
 * Get the effective connection type
 */
export function getEffectiveConnectionType(): ConnectionType {
  const connection = getNetworkInformation();
  return connection?.effectiveType || 'unknown';
}

/**
 * Get the download speed in Mbps
 */
export function getDownlinkSpeed(): number | null {
  const connection = getNetworkInformation();
  return connection?.downlink || null;
}

/**
 * Get the round-trip time in milliseconds
 */
export function getRoundTripTime(): number | null {
  const connection = getNetworkInformation();
  return connection?.rtt || null;
}

/**
 * Check if the device is online
 */
export function isOnline(): boolean {
  return navigator.onLine;
}

/**
 * Add listener for network status changes
 */
export function addNetworkChangeListener(
  callback: (quality: NetworkQuality) => void
): () => void {
  const handleOnlineChange = () => {
    callback(getNetworkQuality());
  };

  const connection = getNetworkInformation();

  // Listen to online/offline events
  window.addEventListener('online', handleOnlineChange);
  window.addEventListener('offline', handleOnlineChange);

  // Listen to connection changes if supported
  if (connection) {
    connection.addEventListener('change', handleOnlineChange);
  }

  // Return cleanup function
  return () => {
    window.removeEventListener('online', handleOnlineChange);
    window.removeEventListener('offline', handleOnlineChange);

    if (connection) {
      connection.removeEventListener('change', handleOnlineChange);
    }
  };
}

/**
 * Adaptive loading configuration based on network quality
 */
export interface AdaptiveConfig {
  enableImages: boolean;
  enableAnimations: boolean;
  enableVideoAutoplay: boolean;
  prefetchEnabled: boolean;
  maxConcurrentRequests: number;
  imageQuality: 'high' | 'medium' | 'low';
  enableLazyLoading: boolean;
  chunkSize: number;
}

/**
 * Get adaptive loading configuration based on network quality
 */
export function getAdaptiveConfig(quality?: NetworkQuality): AdaptiveConfig {
  const networkQuality = quality || getNetworkQuality();

  switch (networkQuality) {
    case 'high':
      return {
        enableImages: true,
        enableAnimations: true,
        enableVideoAutoplay: true,
        prefetchEnabled: true,
        maxConcurrentRequests: 6,
        imageQuality: 'high',
        enableLazyLoading: false,
        chunkSize: 50,
      };

    case 'medium':
      return {
        enableImages: true,
        enableAnimations: true,
        enableVideoAutoplay: false,
        prefetchEnabled: true,
        maxConcurrentRequests: 4,
        imageQuality: 'medium',
        enableLazyLoading: true,
        chunkSize: 30,
      };

    case 'low':
      return {
        enableImages: true,
        enableAnimations: false,
        enableVideoAutoplay: false,
        prefetchEnabled: false,
        maxConcurrentRequests: 2,
        imageQuality: 'low',
        enableLazyLoading: true,
        chunkSize: 20,
      };

    case 'offline':
      return {
        enableImages: false,
        enableAnimations: false,
        enableVideoAutoplay: false,
        prefetchEnabled: false,
        maxConcurrentRequests: 0,
        imageQuality: 'low',
        enableLazyLoading: true,
        chunkSize: 10,
      };
  }
}

/**
 * Throttle API requests based on network quality
 */
export function getRequestDelay(quality?: NetworkQuality): number {
  const networkQuality = quality || getNetworkQuality();

  switch (networkQuality) {
    case 'high':
      return 0;
    case 'medium':
      return 100;
    case 'low':
      return 500;
    case 'offline':
      return 0; // No delay, will be queued
  }
}

/**
 * Check if preloading should be enabled
 */
export function shouldPreload(): boolean {
  const quality = getNetworkQuality();
  return quality === 'high' || quality === 'medium';
}

/**
 * Check if images should be loaded
 */
export function shouldLoadImages(): boolean {
  const quality = getNetworkQuality();
  return quality !== 'offline' && !isDataSaverEnabled();
}

/**
 * Get optimal image quality based on network
 */
export function getOptimalImageQuality(): number {
  const quality = getNetworkQuality();

  switch (quality) {
    case 'high':
      return 0.9;
    case 'medium':
      return 0.7;
    case 'low':
      return 0.5;
    case 'offline':
      return 0.3;
  }
}

/**
 * Network monitoring hook data
 */
export interface NetworkStatus {
  isOnline: boolean;
  quality: NetworkQuality;
  effectiveType: ConnectionType;
  downlink: number | null;
  rtt: number | null;
  saveData: boolean;
}

/**
 * Get current network status
 */
export function getNetworkStatus(): NetworkStatus {
  return {
    isOnline: isOnline(),
    quality: getNetworkQuality(),
    effectiveType: getEffectiveConnectionType(),
    downlink: getDownlinkSpeed(),
    rtt: getRoundTripTime(),
    saveData: isDataSaverEnabled(),
  };
}
