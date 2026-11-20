export const PLAN_NAMES = {
  free: 'Free',
  premium: 'Premium',
  growth: 'Growth',
  enterprise: 'Enterprise',
} as const;

export function getPlanDisplayName(planType: string, polarProducts?: any[]): string {
  if (polarProducts?.length) {
    const match = polarProducts.find(
      (p: any) =>
        p.plan_type === planType &&
        p.is_recurring &&
        !p.name.toLowerCase().includes('yearly')
    );
    if (match) return match.name;
  }
  const fallback: Record<string, string> = PLAN_NAMES;
  return fallback[planType] ?? planType;
}

export type PlanType = keyof typeof PLAN_NAMES;

// Plans from lowest to highest. Matches PredefinedPlans in internal/services/billing.go.
const PLAN_ORDER: PlanType[] = ['free', 'premium', 'growth', 'enterprise'];

/**
 * A plan id as stored on the user (`account_type`) or the subscription (`plan_type`).
 * Unknown or missing values count as Free; the legacy "pro" id is Premium.
 */
export function normalizePlan(value: string | null | undefined): PlanType {
  const v = (value || '').toLowerCase().replace(/_yearly$/, '');
  if (v === 'pro') return 'premium';
  return (PLAN_ORDER as string[]).includes(v) ? (v as PlanType) : 'free';
}

/** True when `plan` is `min` or a higher plan, e.g. planAtLeast('growth', 'premium'). */
export function planAtLeast(plan: string | null | undefined, min: PlanType): boolean {
  return PLAN_ORDER.indexOf(normalizePlan(plan)) >= PLAN_ORDER.indexOf(min);
}

/** The highest plan. Accounts on it should never be shown an upgrade prompt. */
export function isTopPlan(plan: string | null | undefined): boolean {
  return normalizePlan(plan) === 'enterprise';
}
