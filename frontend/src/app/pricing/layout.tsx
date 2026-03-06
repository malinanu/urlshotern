import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Pricing - Trunc URL Shortener",
  description: "Simple, transparent pricing for every team. Start free, upgrade when you need more. Compare Trunc plans: Free, Pro, and Ultra.",
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
