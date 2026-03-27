import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "Privacy Policy - Trunc",
  description: "Read Trunc's privacy policy to understand how we collect, use, and protect your data when you use our URL shortening service.",
  alternates: {
    canonical: 'https://trunc.site/privacy',
  },
  openGraph: {
    title: "Privacy Policy - Trunc",
    description: "Read Trunc's privacy policy to understand how we collect, use, and protect your data.",
  },
  twitter: {
    card: "summary_large_image",
    title: "Privacy Policy - Trunc",
    description: "Read Trunc's privacy policy to understand how we collect, use, and protect your data.",
  },
};

export default function PrivacyLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
