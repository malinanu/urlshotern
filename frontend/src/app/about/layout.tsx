import type { Metadata } from "next";
import type { ReactNode } from "react";

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: "About Trunc — URL Shortener for Marketers & Developers",
  description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights. Built for marketers, developers, and businesses of all sizes.",
  keywords: "about trunc, url shortener company, free url shortener, custom url shortener, url shortener with analytics",
  alternates: {
    canonical: `${SITE_URL}/about`,
  },
  openGraph: {
    type: 'website',
    url: `${SITE_URL}/about`,
    siteName: 'Trunc',
    title: "About Trunc — URL Shortener for Marketers & Developers",
    description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights. Built for marketers, developers, and businesses of all sizes.",
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: "summary_large_image",
    title: "About Trunc — URL Shortener for Marketers & Developers",
    description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights.",
    images: [`${SITE_URL}/og-image.webp`],
  },
};

export default function AboutLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
