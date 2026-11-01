import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Unsubscribe - Trunc",
  description: "Stop receiving marketing emails from Trunc.",
  robots: { index: false, follow: false },
};

export default function UnsubscribeLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
