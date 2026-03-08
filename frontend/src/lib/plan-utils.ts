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
  const fallback: Record<string, string> = { premium: 'Pro', enterprise: 'Ultra', free: 'Free' };
  return fallback[planType] ?? planType;
}
