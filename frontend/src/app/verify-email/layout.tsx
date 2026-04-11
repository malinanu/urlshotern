import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Verify Email - Trunc",
  description: "Verify your Trunc account email address.",
  robots: { index: false, follow: false },
};

export default function VerifyEmailLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
