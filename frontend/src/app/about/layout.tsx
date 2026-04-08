import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "About Trunc — Free URL Shortener Built for Marketers & Developers",
  description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights. Built for marketers, developers, and businesses of all sizes.",
  keywords: "about trunc, url shortener company, free url shortener, custom url shortener, url shortener with analytics",
  alternates: {
    canonical: 'https://trunc.site/about',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/about',
    siteName: 'Trunc',
    title: "About Trunc — Free URL Shortener Built for Marketers & Developers",
    description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights. Built for marketers, developers, and businesses of all sizes.",
  },
  twitter: {
    card: "summary_large_image",
    title: "About Trunc — Free URL Shortener Built for Marketers & Developers",
    description: "Trunc is a free URL shortener with real-time analytics, custom domains, QR codes, and AI insights.",
  },
};

export default function AboutLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
