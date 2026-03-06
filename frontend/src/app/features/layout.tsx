import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Features - Trunc URL Shortener",
  description: "Explore Trunc's powerful features: advanced analytics, custom domains, QR codes, A/B testing, AI insights, and more. Everything you need to manage your links.",
  openGraph: {
    title: "Features - Trunc URL Shortener",
    description: "Explore Trunc's powerful features: advanced analytics, custom domains, QR codes, A/B testing, AI insights, and more.",
  },
  twitter: {
    card: "summary",
    title: "Features - Trunc URL Shortener",
    description: "Explore Trunc's powerful features: advanced analytics, custom domains, QR codes, A/B testing, AI insights, and more.",
  },
};

export default function FeaturesLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
