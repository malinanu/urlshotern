import type { Metadata } from "next";
import type { ReactNode } from "react";
import { PLAN_NAMES } from '@/lib/plan-utils';

export const metadata: Metadata = {
  title: "Pricing - Trunc URL Shortener",
  description: `Simple, transparent pricing for every team. Start free, upgrade when you need more. Compare Trunc plans: ${PLAN_NAMES.free}, ${PLAN_NAMES.premium}, and ${PLAN_NAMES.enterprise}.`,
  openGraph: {
    title: "Pricing - Trunc URL Shortener",
    description: "Simple, transparent pricing for every team. Start free, upgrade when you need more.",
  },
  twitter: {
    card: "summary",
    title: "Pricing - Trunc URL Shortener",
    description: "Simple, transparent pricing for every team. Start free, upgrade when you need more.",
  },
};

export default function PricingLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
