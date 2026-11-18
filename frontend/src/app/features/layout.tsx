import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Features — Digital Business Cards, Short Links, QR Codes & Analytics | Trunc",
  description: "Make a digital business card from 12 designs, plus short links, QR codes and mobile pages, with analytics for every share. Try a card with your name and create a free account.",
  keywords: "digital business card, digital visiting card, qr code business card, url shortener with analytics, url shortener with qr code, link in bio page, url shortener features",
  alternates: {
    canonical: 'https://trunc.site/features',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/features',
    siteName: 'Trunc',
    title: "Features — Digital Business Cards, Short Links, QR Codes & Analytics | Trunc",
    description: "12 digital business card designs, short links, QR codes and mobile pages, with analytics for every share. Free plan, no credit card.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Features — Digital Business Cards, Short Links, QR Codes & Analytics | Trunc",
    description: "12 digital business card designs, short links, QR codes and mobile pages, with analytics for every share. Free plan, no credit card.",
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
