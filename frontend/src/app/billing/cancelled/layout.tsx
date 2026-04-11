import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Checkout Cancelled - Trunc",
  robots: { index: false, follow: false },
};

export default function BillingCancelledLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
