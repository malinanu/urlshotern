import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "URL Shortener Features — Analytics, Custom Domains, QR Codes & More | Trunc",
  description: "Explore Trunc's free URL shortener features: real-time click analytics, custom domains, QR code generator, A/B link testing, AI insights, and team collaboration. Better than Bitly, free forever.",
  keywords: "url shortener with analytics dashboard, url shortener with qr code, custom url shortener, free url shortener with analytics, url shortener features",
  alternates: {
    canonical: 'https://trunc.site/features',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/features',
    siteName: 'Trunc',
    title: "URL Shortener Features — Analytics, Custom Domains, QR Codes & More | Trunc",
    description: "Explore Trunc's free URL shortener features: real-time click analytics, custom domains, QR code generator, A/B link testing, AI insights, and team collaboration.",
  },
  twitter: {
    card: "summary_large_image",
    title: "URL Shortener Features — Analytics, Custom Domains, QR Codes & More | Trunc",
    description: "Explore Trunc's free URL shortener features: real-time click analytics, custom domains, QR code generator, A/B link testing, and AI insights.",
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
