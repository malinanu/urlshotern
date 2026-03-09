export const PLAN_NAMES = {
  free: 'Free',
  premium: 'Premium',
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
