'use client';

import { motion } from 'framer-motion';
import {
  Area,
  AreaChart,
  ResponsiveContainer,
  Tooltip,
  XAxis,
} from 'recharts';
import { ArrowUpRight, MousePointerClick } from 'lucide-react';

// Sample data for the marketing preview only. Clearly labelled as a sample so
// it is never mistaken for a real account's metrics.
const sample = [
  { day: 'Mon', clicks: 240 },
  { day: 'Tue', clicks: 312 },
  { day: 'Wed', clicks: 287 },
  { day: 'Thu', clicks: 410 },
  { day: 'Fri', clicks: 523 },
  { day: 'Sat', clicks: 468 },
  { day: 'Sun', clicks: 612 },
];

/**
 * A real recharts preview used as the hero's product visual. Uses the locked
 * sky-blue accent and adapts to light/dark via theme tokens.
 */
export default function HeroAnalyticsPreview() {
  // Reduced motion is handled globally by MotionProvider (reducedMotion="user"),
  // which keeps the opacity fade but drops the transform for those users.
  return (
    <motion.div
      initial={{ opacity: 0, y: 24, rotate: -1 }}
      animate={{ opacity: 1, y: 0, rotate: 0 }}
      transition={{ type: 'spring', stiffness: 90, damping: 18, delay: 0.15 }}
      className="relative rounded-[calc(var(--radius))] border border-border bg-card/80 p-4 shadow-xl backdrop-blur-sm sm:p-5"
    >
      <div className="mb-4 flex items-center justify-between">
        <div>
          <p className="text-xs font-medium uppercase tracking-wide text-muted-foreground">
            Clicks this week
            <span className="ml-2 rounded bg-muted px-1.5 py-0.5 text-[10px] font-semibold text-muted-foreground">
              Sample
            </span>
          </p>
          <p className="mt-1 flex items-baseline gap-2 font-mono text-2xl font-bold text-foreground">
            2,852
            <span className="inline-flex items-center text-sm font-semibold text-success">
              <ArrowUpRight className="h-4 w-4" />
              18.2%
            </span>
          </p>
        </div>
        <div className="flex h-10 w-10 items-center justify-center rounded-full bg-primary/10 text-primary">
          <MousePointerClick className="h-5 w-5" />
        </div>
      </div>

      <div className="h-32 w-full sm:h-40">
        <ResponsiveContainer width="100%" height="100%">
          <AreaChart data={sample} margin={{ top: 4, right: 4, left: 4, bottom: 0 }}>
            <defs>
              <linearGradient id="heroClicks" x1="0" y1="0" x2="0" y2="1">
                <stop offset="0%" stopColor="hsl(var(--primary))" stopOpacity={0.35} />
                <stop offset="100%" stopColor="hsl(var(--primary))" stopOpacity={0} />
              </linearGradient>
            </defs>
            <XAxis
              dataKey="day"
              tickLine={false}
              axisLine={false}
              tick={{ fontSize: 11, fill: 'hsl(var(--muted-foreground))' }}
            />
            <Tooltip
              cursor={{ stroke: 'hsl(var(--primary))', strokeWidth: 1, strokeOpacity: 0.3 }}
              contentStyle={{
                background: 'hsl(var(--popover))',
                border: '1px solid hsl(var(--border))',
                borderRadius: '0.5rem',
                color: 'hsl(var(--popover-foreground))',
                fontSize: '12px',
              }}
              labelStyle={{ color: 'hsl(var(--muted-foreground))' }}
            />
            <Area
              type="monotone"
              dataKey="clicks"
              stroke="hsl(var(--primary))"
              strokeWidth={2}
              fill="url(#heroClicks)"
            />
          </AreaChart>
        </ResponsiveContainer>
      </div>

      <div className="mt-4 grid grid-cols-3 gap-3 border-t border-border pt-4">
        {[
          { label: 'Top country', value: 'US' },
          { label: 'Top referrer', value: 'X.com' },
          { label: 'Devices', value: 'Mobile' },
        ].map((s) => (
          <div key={s.label}>
            <p className="text-[11px] text-muted-foreground">{s.label}</p>
            <p className="text-sm font-semibold text-foreground">{s.value}</p>
          </div>
        ))}
      </div>
    </motion.div>
  );
}
