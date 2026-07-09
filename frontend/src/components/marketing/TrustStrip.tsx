'use client';

import { useEffect, useState } from 'react';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';
import { formatNumber } from '@/utils/formatters';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface PublicStats {
  total_urls: number;
  total_clicks: number;
  total_users: number;
}

const trust = ['Free forever plan', 'No credit card required', 'Set up in 2 minutes'];

/**
 * Trust strip under the hero. Shows real site-wide counters fetched live from
 * the public stats endpoint (works in dev SSR and the static-export prod build).
 * Until the fetch resolves it renders a neutral placeholder; on error it keeps
 * the placeholder rather than showing fabricated numbers.
 */
export default function TrustStrip() {
  const [stats, setStats] = useState<PublicStats | null>(null);

  useEffect(() => {
    let cancelled = false;
    fetch(`${API_BASE_URL}/api/v1/public/stats`)
      .then((r) => (r.ok ? r.json() : null))
      .then((data: PublicStats | null) => {
        if (!cancelled && data) setStats(data);
      })
      .catch(() => {
        /* keep placeholder on failure */
      });
    return () => {
      cancelled = true;
    };
  }, []);

  const items = [
    { value: stats ? formatNumber(stats.total_urls) : '—', label: 'URLs shortened' },
    { value: stats ? formatNumber(stats.total_clicks) : '—', label: 'Clicks tracked' },
    { value: stats ? formatNumber(stats.total_users) : '—', label: 'Users' },
  ];

  return (
    <section className="border-y border-border bg-card/40">
      <div className="mx-auto max-w-7xl px-4 py-6 sm:px-6 sm:py-10 lg:px-8">
        <Stagger className="grid grid-cols-3 gap-4 sm:gap-8">
          {items.map((s) => (
            <StaggerItem
              key={s.label}
              className="flex flex-col items-center text-center sm:flex-row sm:items-baseline sm:justify-center sm:gap-3"
            >
              <span className="font-mono text-2xl font-bold tracking-tight text-foreground tabular-nums sm:text-4xl">
                {s.value}
              </span>
              <span className="mt-1 text-xs font-medium text-muted-foreground sm:mt-0 sm:text-sm">
                {s.label}
              </span>
            </StaggerItem>
          ))}
        </Stagger>

        <div className="mt-6 flex flex-row flex-wrap items-center justify-center gap-x-4 gap-y-2 border-t border-border pt-5 text-xs text-muted-foreground sm:mt-8 sm:gap-8 sm:pt-6 sm:text-sm">
          {trust.map((t) => (
            <span key={t} className="inline-flex items-center gap-2">
              <span className="h-1.5 w-1.5 rounded-full bg-success" aria-hidden />
              {t}
            </span>
          ))}
        </div>
      </div>
    </section>
  );
}
