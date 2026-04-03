import type { Metadata } from 'next';
import type { ReactNode } from 'react';

export const metadata: Metadata = {
  title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
  description: 'Learn how to shorten URLs, track link clicks, boost marketing performance, and compare the best URL shorteners. Tips and guides from the Trunc team.',
  alternates: {
    canonical: 'https://trunc.site/blog',
  },
  openGraph: {
    type: 'website',
    url: 'https://trunc.site/blog',
    siteName: 'Trunc',
    title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
    description: 'Learn how to shorten URLs, track link clicks, and boost marketing performance. Tips and guides from the Trunc team.',
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Blog — URL Shortener Tips, Guides & News | Trunc',
    description: 'Learn how to shorten URLs, track link clicks, and boost marketing performance.',
  },
};

export default function BlogLayout({ children }: { children: ReactNode }) {
  return <>{children}</>;
}
