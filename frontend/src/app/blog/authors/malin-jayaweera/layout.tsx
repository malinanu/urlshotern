import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { JsonLd } from '@/components/JsonLd';

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: 'Malin Jayaweera — URL Shortener Expert & Founder | Trunc',
  description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  alternates: { canonical: `${SITE_URL}/blog/authors/malin-jayaweera` },
  openGraph: {
    type: 'profile',
    url: `${SITE_URL}/blog/authors/malin-jayaweera`,
    siteName: 'Trunc',
    title: 'Malin Jayaweera — URL Shortener Expert & Founder | Trunc',
    description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
    images: [{ url: `${SITE_URL}/og-image.webp`, width: 1200, height: 630 }],
  },
  twitter: {
    card: 'summary_large_image',
    title: 'Malin Jayaweera — URL Shortener Expert & Founder | Trunc',
    description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
    images: [`${SITE_URL}/og-image.webp`],
  },
};

const profileSchema = {
  '@context': 'https://schema.org',
  '@type': 'ProfilePage',
  url: `${SITE_URL}/blog/authors/malin-jayaweera`,
  mainEntity: {
    '@type': 'Person',
    '@id': `${SITE_URL}/blog/authors/malin-jayaweera`,
    name: 'Malin Jayaweera',
    url: `${SITE_URL}/blog/authors/malin-jayaweera`,
    jobTitle: 'Founder & CEO',
    description: 'Malin Jayaweera is the founder and CEO of Trunc, a URL shortener with advanced analytics and AI insights. He writes about link management, digital marketing, and growth strategies.',
    image: {
      '@type': 'ImageObject',
      url: `${SITE_URL}/og-image.webp`,
    },
    sameAs: [
      'https://x.com/truncsite',
      'https://linkedin.com/in/malinjayaweera',
    ],
    worksFor: {
      '@type': 'Organization',
      name: 'Trunc',
      url: SITE_URL,
    },
    knowsAbout: [
      'URL shortening',
      'Link analytics',
      'Digital marketing',
      'Web analytics',
      'Growth marketing',
      'SaaS',
    ],
  },
};

const breadcrumbSchema = {
  '@context': 'https://schema.org',
  '@type': 'BreadcrumbList',
  itemListElement: [
    { '@type': 'ListItem', position: 1, name: 'Home', item: SITE_URL },
    { '@type': 'ListItem', position: 2, name: 'Blog', item: `${SITE_URL}/blog` },
    { '@type': 'ListItem', position: 3, name: 'Malin Jayaweera', item: `${SITE_URL}/blog/authors/malin-jayaweera` },
  ],
};

export default function AuthorLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <JsonLd data={profileSchema} />
      <JsonLd data={breadcrumbSchema} />
      {children}
    </>
  );
}
