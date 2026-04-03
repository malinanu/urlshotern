import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { JsonLd } from '@/components/JsonLd';

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: 'Malin Jayaweera — Author at Trunc Blog',
  description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  alternates: { canonical: `${SITE_URL}/blog/authors/malin-jayaweera` },
  openGraph: {
    type: 'profile',
    url: `${SITE_URL}/blog/authors/malin-jayaweera`,
    siteName: 'Trunc',
    title: 'Malin Jayaweera — Author at Trunc Blog',
    description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  },
  twitter: {
    card: 'summary',
    title: 'Malin Jayaweera — Author at Trunc Blog',
    description: 'Malin Jayaweera is the founder of Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  },
};

const personSchema = {
  '@context': 'https://schema.org',
  '@type': 'Person',
  name: 'Malin Jayaweera',
  url: `${SITE_URL}/blog/authors/malin-jayaweera`,
  jobTitle: 'Founder & CEO',
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
  ],
};

export default function AuthorLayout({ children }: { children: ReactNode }) {
  return (
    <>
      <JsonLd data={personSchema} />
      {children}
    </>
  );
}
