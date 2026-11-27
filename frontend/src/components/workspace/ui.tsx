'use client';

/**
 * Small building blocks for workspace screens (sidebar layout). The class
 * strings mirror the dashboard (app/dashboard/page.tsx) so screens built from
 * these match it: rounded-2xl panels on surface-container-lowest / zinc-950,
 * outline-variant hairlines, Hanken headings, rounded-xl controls.
 */

import type { ButtonHTMLAttributes, ReactNode } from 'react';
import { forwardRef } from 'react';
import { PAGE_STATE_LABEL, type PageState } from '@/lib/microsites';

export function cx(...parts: (string | false | null | undefined)[]): string {
  return parts.filter(Boolean).join(' ');
}

/** Visible focus ring for every custom control. */
export const focusRing =
  'focus-visible:outline-none focus-visible:ring-2 focus-visible:ring-primary/60 focus-visible:ring-offset-2 focus-visible:ring-offset-surface-container-lowest dark:focus-visible:ring-offset-[#07090e]';

export const panelClass =
  'bg-surface-container-lowest dark:bg-zinc-950 border border-outline-variant/30 rounded-2xl shadow-sm';

export const inputClass =
  'w-full px-3 py-2 rounded-xl text-sm bg-surface-container dark:bg-zinc-900 border border-outline-variant/30 text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 focus:border-primary transition-colors';

export function Panel({ className, children }: { className?: string; children: ReactNode }) {
  return <div className={cx(panelClass, className)}>{children}</div>;
}

type Variant = 'primary' | 'secondary' | 'ghost' | 'danger';
type Size = 'sm' | 'md';

const VARIANTS: Record<Variant, string> = {
  primary: 'bg-primary hover:bg-primary-600 text-white shadow-sm disabled:bg-primary/50',
  secondary:
    'border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-900 hover:bg-surface-container dark:hover:bg-zinc-800 text-foreground',
  ghost: 'text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900',
  danger: 'bg-destructive hover:bg-destructive/90 text-white shadow-sm',
};
const SIZES: Record<Size, string> = {
  sm: 'h-8 px-3 text-xs gap-1.5',
  md: 'h-9 px-3.5 text-sm gap-2',
};

export function buttonClass(variant: Variant = 'secondary', size: Size = 'md', className?: string): string {
  return cx(
    'inline-flex items-center justify-center rounded-xl font-semibold whitespace-nowrap transition-colors active:scale-[0.98] disabled:opacity-50 disabled:pointer-events-none',
    focusRing,
    VARIANTS[variant],
    SIZES[size],
    className
  );
}

export const Button = forwardRef<
  HTMLButtonElement,
  ButtonHTMLAttributes<HTMLButtonElement> & { variant?: Variant; size?: Size }
>(function Button({ variant = 'secondary', size = 'md', className, type = 'button', ...rest }, ref) {
  return <button ref={ref} type={type} className={buttonClass(variant, size, className)} {...rest} />;
});

/** Square icon-only button; always pass a label. */
export const IconButton = forwardRef<
  HTMLButtonElement,
  ButtonHTMLAttributes<HTMLButtonElement> & { label: string; size?: 'sm' | 'md' }
>(function IconButton({ label, size = 'md', className, type = 'button', children, ...rest }, ref) {
  return (
    <button
      ref={ref}
      type={type}
      aria-label={label}
      title={label}
      className={cx(
        'inline-flex items-center justify-center rounded-xl text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 transition-colors disabled:opacity-40',
        focusRing,
        size === 'sm' ? 'h-8 w-8' : 'h-9 w-9',
        className
      )}
      {...rest}
    >
      {children}
    </button>
  );
});

/** Small rounded label in the dashboard's mono badge style (plan names, counts, types). */
export function Badge({ children, tone = 'primary', className }: { children: ReactNode; tone?: 'primary' | 'muted' | 'indigo'; className?: string }) {
  const tones = {
    primary: 'bg-primary/10 text-primary',
    muted: 'bg-surface-container dark:bg-zinc-900 text-muted-foreground border border-outline-variant/30',
    indigo: 'bg-indigo-500/10 text-indigo-600 dark:text-indigo-400 border border-indigo-500/20',
  };
  return (
    <span className={cx('inline-flex items-center gap-1 text-[10px] font-bold px-2 py-0.5 rounded-full font-mono whitespace-nowrap', tones[tone], className)}>
      {children}
    </span>
  );
}

const DOT: Record<PageState, string> = {
  draft: 'bg-zinc-400 dark:bg-zinc-500',
  published: 'bg-emerald-500',
  changes: 'bg-amber-500',
};

/** A coloured dot with a sentence-case label; colour is never the only signal. */
export function StatusDot({ state, className }: { state: PageState; className?: string }) {
  return (
    <span className={cx('inline-flex items-center gap-1.5 text-xs font-medium text-foreground', className)}>
      <span className={cx('h-2 w-2 rounded-full shrink-0', DOT[state])} aria-hidden="true" />
      {PAGE_STATE_LABEL[state]}
    </span>
  );
}

/** Segmented control for a small set of mutually exclusive views or filters. */
export function Segmented<T extends string>({
  options,
  value,
  onChange,
  label,
  className,
}: {
  options: { value: T; label: ReactNode; count?: number }[];
  value: T;
  onChange: (v: T) => void;
  label: string;
  className?: string;
}) {
  return (
    <div
      role="tablist"
      aria-label={label}
      className={cx('inline-flex items-center gap-0.5 p-0.5 rounded-xl bg-surface-container dark:bg-zinc-900 border border-outline-variant/30', className)}
    >
      {options.map((o) => {
        const active = o.value === value;
        return (
          <button
            key={o.value}
            type="button"
            role="tab"
            aria-selected={active}
            onClick={() => onChange(o.value)}
            className={cx(
              'inline-flex items-center gap-1.5 h-7 px-3 rounded-[10px] text-xs font-semibold whitespace-nowrap transition-colors',
              focusRing,
              active
                ? 'bg-surface-container-lowest dark:bg-zinc-800 text-foreground shadow-sm'
                : 'text-muted-foreground hover:text-foreground'
            )}
          >
            {o.label}
            {o.count !== undefined && (
              <span className={cx('tabular-nums text-[11px]', active ? 'text-muted-foreground' : 'text-muted-foreground/70')}>{o.count}</span>
            )}
          </button>
        );
      })}
    </div>
  );
}

/**
 * A tiny line of real daily values. The label says what it shows, for screen
 * readers; the chart itself is decorative.
 */
export function Sparkline({ values, label, className }: { values: number[]; label: string; className?: string }) {
  const w = 72;
  const h = 22;
  const max = Math.max(1, ...values);
  const step = values.length > 1 ? w / (values.length - 1) : w;
  const pts = values.map((v, i) => [i * step, h - 2 - (v / max) * (h - 4)] as const);
  const line = pts.map(([x, y], i) => `${i ? 'L' : 'M'}${x.toFixed(1)},${y.toFixed(1)}`).join(' ');
  const area = `${line} L${w},${h} L0,${h} Z`;
  const flat = values.every((v) => v === 0);
  return (
    <svg viewBox={`0 0 ${w} ${h}`} width={w} height={h} className={cx('overflow-visible', className)} role="img" aria-label={label}>
      {!flat && <path d={area} className="fill-primary/10" />}
      <path
        d={line}
        fill="none"
        strokeWidth={1.5}
        strokeLinejoin="round"
        strokeLinecap="round"
        className={flat ? 'stroke-outline-variant/60' : 'stroke-primary'}
      />
    </svg>
  );
}

export function Skeleton({ className }: { className?: string }) {
  return <div className={cx('animate-pulse rounded-lg bg-surface-container dark:bg-zinc-900', className)} />;
}
