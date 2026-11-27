'use client';

import { useEffect, useRef, useState, type ReactNode } from 'react';
import { cachedHtml, loadHtml } from '@/lib/thumbnails';

const RENDER_WIDTH = 390;

/**
 * A rendered page shown at phone width and scaled down, inside a sandboxed
 * srcdoc iframe (no scripts, no same-origin access, no pointer events). The
 * HTML loads when the frame scrolls near the viewport.
 */
export default function PhoneFrame({
  cacheKey,
  url,
  getToken,
  width,
  aspect = 1.72,
  fallback,
  className = '',
  keepPrevious = false,
}: {
  cacheKey: string;
  url: string;
  /** Reads the access token; called only in the browser. */
  getToken: () => string | null;
  /** Frame width in CSS pixels. */
  width: number;
  /** Height divided by width. */
  aspect?: number;
  /** Shown if the page cannot be rendered. */
  fallback: ReactNode;
  className?: string;
  /** Keep showing the previous render while a new key loads (live previews). */
  keepPrevious?: boolean;
}) {
  const [html, setHtml] = useState<string | undefined>(() => (typeof window === 'undefined' ? undefined : cachedHtml(cacheKey)));
  const [failed, setFailed] = useState(false);
  const [near, setNear] = useState(false);
  const ref = useRef<HTMLDivElement>(null);

  useEffect(() => {
    const el = ref.current;
    if (!el) return;
    const io = new IntersectionObserver(
      (entries) => {
        if (entries.some((e) => e.isIntersecting)) {
          setNear(true);
          io.disconnect();
        }
      },
      { rootMargin: '300px' }
    );
    io.observe(el);
    return () => io.disconnect();
  }, []);

  useEffect(() => {
    const hit = cachedHtml(cacheKey);
    if (hit !== undefined) {
      setHtml(hit);
      setFailed(false);
      return;
    }
    if (!keepPrevious) setHtml(undefined);
    if (!near) return;
    let cancelled = false;
    setFailed(false);
    loadHtml(cacheKey, url, getToken())
      .then((doc) => !cancelled && setHtml(doc))
      .catch(() => !cancelled && setFailed(true));
    return () => {
      cancelled = true;
    };
  }, [cacheKey, url, getToken, near, keepPrevious]);

  const height = Math.round(width * aspect);
  const scale = width / RENDER_WIDTH;

  return (
    <div
      ref={ref}
      aria-hidden="true"
      className={`relative overflow-hidden rounded-[18px] border border-black/10 dark:border-white/10 bg-white shadow-[0_10px_28px_-14px_rgba(0,0,0,0.55)] ${className}`}
      style={{ width, height }}
    >
      {html ? (
        <div style={{ width: RENDER_WIDTH, height: height / scale, transform: `scale(${scale})`, transformOrigin: 'top left' }}>
          <iframe
            title=""
            srcDoc={html}
            sandbox=""
            tabIndex={-1}
            scrolling="no"
            className="w-full h-full border-0 pointer-events-none select-none"
          />
        </div>
      ) : failed ? (
        <div className="absolute inset-0 flex flex-col items-center justify-center gap-2 bg-surface-container dark:bg-zinc-900 text-muted-foreground">
          {fallback}
        </div>
      ) : (
        <div className="absolute inset-0 animate-pulse bg-surface-container dark:bg-zinc-900" />
      )}
    </div>
  );
}
