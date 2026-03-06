import type { Metadata } from "next";

export const metadata: Metadata = {
  title: "Terms of Service - Trunc",
  description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our URL shortening platform.",
  openGraph: {
    title: "Terms of Service - Trunc",
    description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our platform.",
  },
  twitter: {
    card: "summary",
    title: "Terms of Service - Trunc",
    description: "Read the Trunc Terms of Service to understand the rules and guidelines for using our platform.",
  },
};

export default function TermsLayout({ children }: { children: React.ReactNode }) {
  return children;
}
