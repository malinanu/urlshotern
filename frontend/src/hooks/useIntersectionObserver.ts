/**
 * Intersection Observer Hook for Lazy Loading
 * Optimized for mobile performance
 */

import { useEffect, RefObject } from 'react';

interface UseIntersectionObserverOptions {
  target: RefObject<Element>;
  onIntersect: () => void;
  threshold?: number;
  rootMargin?: string;
  enabled?: boolean;
}

export function useIntersectionObserver({
  target,
  onIntersect,
  threshold = 0.1,
  rootMargin = '100px',
  enabled = true,
}: UseIntersectionObserverOptions) {
  useEffect(() => {
    if (!enabled || !target.current) {
      return;
    }

    const observer = new IntersectionObserver(
      (entries) => {
        entries.forEach((entry) => {
          if (entry.isIntersecting) {
            onIntersect();
          }
        });
      },
      {
        threshold,
        rootMargin,
      }
    );

    const element = target.current;
    observer.observe(element);

    return () => {
      observer.unobserve(element);
      observer.disconnect();
    };
  }, [target, onIntersect, threshold, rootMargin, enabled]);
}
