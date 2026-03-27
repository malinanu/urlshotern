import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
  description: "Trunc URL shortener features: real-time click analytics, custom branded domains, QR code generation, A/B link testing, AI insights, and team collaboration. Everything you need to manage your links.",
  alternates: {
    canonical: 'https://trunc.site/features',
  },
  openGraph: {
    title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
    description: "Trunc URL shortener features: real-time click analytics, custom branded domains, QR code generation, A/B link testing, AI insights, and team collaboration.",
  },
  twitter: {
    card: "summary_large_image",
    title: "URL Shortener Features — Analytics, Custom Domains & QR Codes | Trunc",
    description: "Trunc URL shortener features: real-time click analytics, custom branded domains, QR code generation, A/B link testing, AI insights, and team collaboration.",
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
