'use client';

import React from 'react';
import Link from 'next/link';

type TeaserVariant = 'unauthenticated' | 'upgrade-pro' | 'upgrade-ultra';

interface AnalyticsTeaserProps {
  variant: TeaserVariant;
  children: React.ReactNode;
  className?: string;
}

const teaserContent: Record<TeaserVariant, { heading: string; body: string; cta: string; href: string }> = {
  'unauthenticated': {
    heading: 'See who is clicking your links',
    body: 'Geographic breakdowns, device stats, referrer sources and more.',
    cta: 'Create free account',
    href: '/register',
  },
  'upgrade-pro': {
    heading: 'Unlock advanced analytics',
    body: 'City-level geo, heatmaps, UTM campaigns, real-time feed and more with Pro.',
    cta: 'Upgrade to Pro',
    href: '/billing',
  },
  'upgrade-ultra': {
    heading: 'Unlock AI-powered insights',
    body: 'Performance summaries, anomaly detection, recommendations and audience intelligence with Ultra.',
    cta: 'Upgrade to Ultra',
    href: '/billing',
  },
};

/**
 * AnalyticsTeaser wraps any analytics component with a blur overlay and
 * a contextual upgrade/sign-up call to action.
 *
 * Usage:
 *   <AnalyticsTeaser variant="upgrade-pro">
 *     <HeatmapComponent mockData={MOCK_HEATMAP} />
 *   </AnalyticsTeaser>
 */
export function AnalyticsTeaser({ variant, children, className = '' }: AnalyticsTeaserProps) {
  const { heading, body, cta, href } = teaserContent[variant];

  return (
    <div className={`relative overflow-hidden rounded-xl ${className}`}>
      {/* Blurred content underneath */}
      <div
        className="pointer-events-none select-none"
        style={{ filter: 'blur(6px)', opacity: 0.6 }}
        aria-hidden="true"
      >
        {children}
      </div>

      {/* Overlay */}
      <div className="absolute inset-0 flex items-center justify-center bg-white/40 dark:bg-gray-900/40 backdrop-blur-sm">
        <div className="mx-auto max-w-sm rounded-2xl border border-gray-200 dark:border-gray-700 bg-white dark:bg-gray-800 p-8 text-center shadow-2xl">
          <div className="mb-2 text-3xl">📊</div>
          <h3 className="mb-2 text-lg font-semibold text-gray-900 dark:text-white">
            {heading}
          </h3>
          <p className="mb-6 text-sm text-gray-500 dark:text-gray-400">{body}</p>
          <Link
            href={href}
            className="inline-block rounded-lg bg-indigo-600 px-6 py-2.5 text-sm font-semibold text-white shadow hover:bg-indigo-700 transition-colors"
          >
            {cta}
          </Link>
        </div>
      </div>
    </div>
  );
}

export default AnalyticsTeaser;
