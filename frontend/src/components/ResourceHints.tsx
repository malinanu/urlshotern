/**
 * Resource Hints Component
 * Implements preconnect, dns-prefetch, and preload for critical resources
 */

'use client';

import { useEffect } from 'react';
import { useNetworkQuality } from '@/hooks/useNetworkStatus';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

export function ResourceHints() {
  const networkQuality = useNetworkQuality();

  useEffect(() => {
    // Only preconnect on good network conditions
    if (networkQuality === 'high' || networkQuality === 'medium') {
      // API preconnect (fonts are self-hosted, so no third-party font hosts)
      addPreconnect(API_BASE_URL);
    }
  }, [networkQuality]);

  return null;
}

/**
 * Add preconnect link
 */
function addPreconnect(url: string) {
  if (typeof window === 'undefined') return;

  const existing = document.querySelector(`link[rel="preconnect"][href="${url}"]`);
  if (existing) return;

  const link = document.createElement('link');
  link.rel = 'preconnect';
  link.href = url;
  link.crossOrigin = 'anonymous';
  document.head.appendChild(link);
}

/**
 * Preload critical resources
 */
export function preloadCriticalResources() {
  if (typeof window === 'undefined') return;

  // Preload critical API endpoints
  const criticalEndpoints = [
    '/api/v1/user/profile',
    '/api/v1/my-urls',
  ];

  criticalEndpoints.forEach((endpoint) => {
    const link = document.createElement('link');
    link.rel = 'preload';
    link.as = 'fetch';
    link.href = `${API_BASE_URL}${endpoint}`;
    link.crossOrigin = 'anonymous';
    document.head.appendChild(link);
  });
}

/**
 * Prefetch next page resources
 */
export function prefetchPage(href: string) {
  if (typeof window === 'undefined') return;

  const existing = document.querySelector(`link[rel="prefetch"][href="${href}"]`);
  if (existing) return;

  const link = document.createElement('link');
  link.rel = 'prefetch';
  link.href = href;
  document.head.appendChild(link);
}

/**
 * Hook to prefetch on hover
 */
export function usePrefetchOnHover() {
  const networkQuality = useNetworkQuality();

  const prefetchOnHover = (href: string) => {
    // Only prefetch on good network
    if (networkQuality === 'high' || networkQuality === 'medium') {
      prefetchPage(href);
    }
  };

  return prefetchOnHover;
}
