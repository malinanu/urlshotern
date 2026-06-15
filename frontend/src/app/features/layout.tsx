import type { Metadata } from "next";
import type { ReactNode } from "react";

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: "URL Shortener Features: Analytics, QR Codes & Custom Domains | Trunc",
  description: "Explore Trunc's free URL shortener features: real-time click analytics, custom branded domains, QR code generator, A/B link testing, and AI insights.",
  keywords: "url shortener with analytics dashboard, url shortener with qr code, custom url shortener, free url shortener with analytics, url shortener features",
  alternates: {
    canonical: `${SITE_URL}/features`,
  },
  openGraph: {
    type: 'website',
    url: `${SITE_URL}/features`,
    siteName: 'Trunc',
    title: "URL Shortener Features: Analytics, QR Codes & Custom Domains | Trunc",
    description: "Explore Trunc's free URL shortener features: real-time click analytics, custom branded domains, QR code generator, A/B link testing, and AI insights.",
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "URL Shortener Features: Analytics, QR Codes & Custom Domains | Trunc",
    description: "Explore Trunc's free URL shortener features: real-time click analytics, custom branded domains, QR code generator, A/B link testing, and AI insights.",
    images: [`${SITE_URL}/og-image.webp`],
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
