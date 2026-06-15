import type { Metadata } from 'next';
import type { ReactNode } from 'react';

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
  description: 'Learn how to shorten URLs, track link clicks, boost marketing performance, and compare the best URL shorteners. Tips and guides from the Trunc team.',
  alternates: {
    canonical: `${SITE_URL}/blog`,
  },
  openGraph: {
    type: 'website',
    url: `${SITE_URL}/blog`,
    siteName: 'Trunc',
    title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
    description: 'Learn how to shorten URLs, track link clicks, and boost marketing performance. Tips and guides from the Trunc team.',
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
    description: 'Learn how to shorten URLs, track link clicks, and boost marketing performance.',
    images: [`${SITE_URL}/og-image.webp`],
  },
};

export default function BlogLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
