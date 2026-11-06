import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { JsonLd } from '@/components/JsonLd';

const SITE_URL = 'https://trunc.site';

export const metadata: Metadata = {
  title: 'Alex Rivera — Author at Trunc Blog',
  description: 'Alex Rivera is a lead infrastructure architect at Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  alternates: { canonical: `${SITE_URL}/blog/authors/alex-rivera` },
  openGraph: {
    type: 'profile',
    url: `${SITE_URL}/blog/authors/alex-rivera`,
    siteName: 'Trunc',
    title: 'Alex Rivera — Author at Trunc Blog',
    description: 'Alex Rivera is a lead infrastructure architect at Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  },
  twitter: {
    card: 'summary',
    title: 'Alex Rivera — Author at Trunc Blog',
    description: 'Alex Rivera is a lead infrastructure architect at Trunc and writes about URL management, link analytics, and digital marketing strategy.',
  },
};

const personSchema = {
  '@context': 'https://schema.org',
  '@type': 'Person',
  name: 'Alex Rivera',
  url: `${SITE_URL}/blog/authors/alex-rivera`,
  jobTitle: 'Lead Infrastructure Architect',
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
