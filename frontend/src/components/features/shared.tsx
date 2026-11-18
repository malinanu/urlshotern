import type { ReactNode } from 'react';
import { Reveal } from '@/components/motion/Reveal';
import { cn } from '@/lib/utils';
import { SAMPLE_CARD } from '@/components/micropages/cards/cardData';
import type { CardData } from '@/components/micropages/cards/types';

/** Every sign-up button on the page goes here. */
export const SIGNUP_HREF = '/register';

/**
 * The QR codes on the demo cards open the sign-up page, so a visitor who scans one
 * with their phone lands somewhere useful instead of on a sample person's page.
 */
export const DEMO_QR_VALUE = 'https://trunc.site/register';

/** The sample card used across the page. Contact details are clearly fictional. */
export const DEMO_CARD: CardData = { ...SAMPLE_CARD, qrValue: DEMO_QR_VALUE };

interface SectionHeadingProps {
  eyebrow?: string;
  title: ReactNode;
  lead?: ReactNode;
  align?: 'left' | 'center';
  id?: string;
  className?: string;
}

export function SectionHeading({ eyebrow, title, lead, align = 'left', id, className }: SectionHeadingProps) {
  return (
    <Reveal className={cn('max-w-2xl', align === 'center' && 'mx-auto text-center', className)}>
      {eyebrow && <p className="text-sm font-semibold uppercase tracking-wider text-primary">{eyebrow}</p>}
      <h2
        id={id}
        className="mt-2 text-balance text-[1.75rem] font-bold leading-tight tracking-tight text-foreground sm:text-4xl"
      >
        {title}
      </h2>
      {lead && <p className="mt-3 text-pretty text-base leading-relaxed text-muted-foreground sm:mt-4 sm:text-lg">{lead}</p>}
    </Reveal>
  );
}

export type PlanTier = 'all' | 'premium' | 'growth' | 'enterprise';

const TIER_LABEL: Record<PlanTier, string> = {
  all: 'All plans',
  premium: 'Premium+',
  growth: 'Growth+',
  enterprise: 'Enterprise',
};

const TIER_CLASS: Record<PlanTier, string> = {
  all: 'border-emerald-500/30 bg-emerald-500/10 text-emerald-700 dark:text-emerald-300',
  premium: 'border-primary/30 bg-primary/10 text-primary',
  growth: 'border-indigo-500/30 bg-indigo-500/10 text-indigo-700 dark:text-indigo-300',
  enterprise: 'border-foreground/20 bg-foreground/[0.06] text-foreground',
};

/** A small label naming the plans a feature is on. */
export function PlanTag({ tier, className }: { tier: PlanTier; className?: string }) {
  return (
    <span
      className={cn(
        'inline-flex shrink-0 items-center whitespace-nowrap rounded-full border px-2 py-0.5 text-[11px] font-semibold leading-4',
        TIER_CLASS[tier],
        className
      )}
    >
      {TIER_LABEL[tier]}
    </span>
  );
}
