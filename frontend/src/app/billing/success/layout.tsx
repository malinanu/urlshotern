import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Payment Successful - Trunc",
  robots: { index: false, follow: false },
};

export default function BillingSuccessLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
