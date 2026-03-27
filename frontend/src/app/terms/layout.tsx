import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Terms of Service - Trunc",
  description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our URL shortening platform.",
  alternates: {
    canonical: 'https://trunc.site/terms',
  },
  openGraph: {
    title: "Terms of Service - Trunc",
    description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our platform.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Terms of Service - Trunc",
    description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our platform.",
  },
};

export default function TermsLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
