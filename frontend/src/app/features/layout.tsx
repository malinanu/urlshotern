import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
  description: "Trunc features: real-time click analytics, custom domains, QR codes, A/B link testing, AI insights, and team collaboration — all in one URL shortener.",
  alternates: {
    canonical: 'https://trunc.site/features',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/features',
    siteName: 'Trunc',
    title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
    description: "Trunc features: real-time click analytics, custom domains, QR codes, A/B link testing, AI insights, and team collaboration — all in one URL shortener.",
  },
  twitter: {
    card: "summary_large_image",
    title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
    description: "Trunc features: real-time click analytics, custom domains, QR codes, A/B link testing, AI insights, and team collaboration — all in one URL shortener.",
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
