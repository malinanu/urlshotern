import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Get Started Free - Trunc URL Shortener",
  description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes. No credit card required.",
  alternates: {
    canonical: 'https://trunc.site/register',
  },
  openGraph: {
    title: "Get Started Free - Trunc URL Shortener",
    description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Get Started Free - Trunc URL Shortener",
    description: "Create your free Trunc account. Start shortening URLs, tracking clicks, and gaining insights in minutes.",
  },
};

export default function RegisterLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
