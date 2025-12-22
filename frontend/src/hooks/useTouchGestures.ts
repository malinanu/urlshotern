/**
 * Touch Gesture Hooks
 * Provides swipe, long-press, and pinch gesture support
 */

'use client';

import { useGesture } from '@use-gesture/react';
import { useRef, RefObject } from 'react';

export interface SwipeGestureOptions {
  onSwipeLeft?: () => void;
  onSwipeRight?: () => void;
  onSwipeUp?: () => void;
  onSwipeDown?: () => void;
  threshold?: number;
}

/**
 * Hook for handling swipe gestures
 */
export function useSwipeGesture(options: SwipeGestureOptions) {
  const {
    onSwipeLeft,
    onSwipeRight,
    onSwipeUp,
    onSwipeDown,
    threshold = 50,
  } = options;

  const bind = useGesture({
    onDrag: ({ movement: [mx, my], swipe: [swipeX, swipeY], event }) => {
      event.preventDefault();

      // Horizontal swipe
      if (Math.abs(mx) > Math.abs(my)) {
        if (swipeX === -1 && onSwipeLeft) {
          onSwipeLeft();
        } else if (swipeX === 1 && onSwipeRight) {
          onSwipeRight();
        }
      }
      // Vertical swipe
      else {
        if (swipeY === -1 && onSwipeUp) {
          onSwipeUp();
        } else if (swipeY === 1 && onSwipeDown) {
          onSwipeDown();
        }
      }
    },
  }, {
    drag: {
      threshold,
      filterTaps: true,
      axis: undefined,
    },
  });

  return bind;
}

export interface LongPressOptions {
  onLongPress: () => void;
  delay?: number;
}

/**
 * Hook for handling long-press gestures
 */
export function useLongPress(options: LongPressOptions) {
  const { onLongPress, delay = 500 } = options;
  const timerRef = useRef<NodeJS.Timeout | null>(null);

  const bind = useGesture({
    onPointerDown: () => {
      timerRef.current = setTimeout(() => {
        onLongPress();
      }, delay);
    },
    onPointerUp: () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    },
    onPointerLeave: () => {
      if (timerRef.current) {
        clearTimeout(timerRef.current);
        timerRef.current = null;
      }
    },
  });

  return bind;
}

export interface PinchZoomOptions {
  onPinchZoom: (scale: number) => void;
  minScale?: number;
  maxScale?: number;
}

/**
 * Hook for handling pinch-to-zoom gestures
 */
export function usePinchZoom(options: PinchZoomOptions) {
  const { onPinchZoom, minScale = 0.5, maxScale = 3 } = options;

  const bind = useGesture({
    onPinch: ({ offset: [scale], event }) => {
      event.preventDefault();
      const clampedScale = Math.min(Math.max(scale, minScale), maxScale);
      onPinchZoom(clampedScale);
    },
  }, {
    pinch: {
      from: [1, 0],
      scaleBounds: { min: minScale, max: maxScale },
    },
  });

  return bind;
}

/**
 * Hook for pull-to-refresh gesture
 */
export interface PullToRefreshOptions {
  onRefresh: () => Promise<void>;
  threshold?: number;
  enabled?: boolean;
}

export function usePullToRefresh(
  containerRef: RefObject<HTMLElement>,
  options: PullToRefreshOptions
) {
  const { onRefresh, threshold = 80, enabled = true } = options;
  const startY = useRef(0);
  const isRefreshing = useRef(false);

  const bind = useGesture({
    onDrag: async ({ movement: [, my], first, last, event }) => {
      if (!enabled || isRefreshing.current) return;

      const container = containerRef.current;
      if (!container) return;

      // Only trigger if scrolled to top
      if (container.scrollTop > 0) return;

      if (first) {
        startY.current = my;
      }

      if (my > threshold && last) {
        isRefreshing.current = true;
        event.preventDefault();

        try {
          await onRefresh();
        } finally {
          isRefreshing.current = false;
        }
      }
    },
  }, {
    drag: {
      axis: 'y',
      filterTaps: true,
    },
  });

  return bind;
}

/**
 * Debounce utility for touch events
 */
export function useTouchDebounce<T extends (...args: any[]) => any>(
  callback: T,
  delay: number = 300
): T {
  const timeoutRef = useRef<NodeJS.Timeout | null>(null);

  return ((...args: Parameters<T>) => {
    if (timeoutRef.current) {
      clearTimeout(timeoutRef.current);
    }

    timeoutRef.current = setTimeout(() => {
      callback(...args);
    }, delay);
  }) as T;
}

/**
 * Throttle utility for touch events
 */
export function useTouchThrottle<T extends (...args: any[]) => any>(
  callback: T,
  delay: number = 100
): T {
  const lastRunRef = useRef(0);

  return ((...args: Parameters<T>) => {
    const now = Date.now();

    if (now - lastRunRef.current >= delay) {
      lastRunRef.current = now;
      callback(...args);
    }
  }) as T;
}
