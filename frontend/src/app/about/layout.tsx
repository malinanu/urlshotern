import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "About Trunc - Our Story & Mission",
  description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses. Discover our mission and team.",
  alternates: {
    canonical: 'https://trunc.site/about',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/about',
    siteName: 'Trunc',
    title: "About Trunc - Our Story & Mission",
    description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses.",
  },
  twitter: {
    card: "summary_large_image",
    title: "About Trunc - Our Story & Mission",
    description: "Learn about Trunc, the professional URL shortening platform built for marketers, developers, and businesses.",
  },
};

export default function AboutLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
