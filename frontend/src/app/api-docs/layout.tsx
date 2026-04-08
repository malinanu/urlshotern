import type { Metadata } from "next";
import type { ReactNode } from "react";

export const metadata: Metadata = {
  title: "REST API Documentation — URL Shortener API | Trunc",
  description: "Complete Trunc URL shortener API documentation. Shorten URLs, retrieve analytics, manage links, and generate QR codes programmatically. Free API key included.",
  keywords: "url shortener api, link shortener api, url shortener api free, rest api url shortener, shorten url api",
  alternates: {
    canonical: "https://trunc.site/api-docs",
  },
  openGraph: {
    type: "website",
    url: "https://trunc.site/api-docs",
    siteName: "Trunc",
    title: "REST API Documentation — URL Shortener API | Trunc",
    description: "Complete Trunc URL shortener API documentation. Shorten URLs, retrieve analytics, manage links, and generate QR codes programmatically. Free API key included.",
  },
  twitter: {
    card: "summary_large_image",
    title: "REST API Documentation — URL Shortener API | Trunc",
    description: "Complete Trunc URL shortener API documentation. Shorten URLs, retrieve analytics, manage links, and generate QR codes programmatically.",
  },
};

export default function ApiDocsLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
