import type { Metadata } from 'next';
import PublicMicroPageClient from './_client';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchMicroPageData(slug: string) {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/micropages/public/${slug}`, {
      next: { revalidate: 60 },
    });
    if (!res.ok) return null;
    return await res.json();
  } catch {
    return null;
  }
}

export async function generateMetadata({
  params,
}: {
  params: Promise<{ slug: string }>;
}): Promise<Metadata> {
  const { slug } = await params;
  const page = await fetchMicroPageData(slug);

  if (!page) {
    return {
      title: 'Micro-Page Not Found | Trunc',
      description: 'The requested micro-page does not exist.',
    };
  }

  const title = `${page.title} | Digital vCard`;
  const description = page.bio || `Connect with ${page.title} and download their digital vCard.`;

  return {
    title,
    description,
    openGraph: {
      title,
      description,
      images: page.avatar_url ? [page.avatar_url] : [],
      type: 'profile',
    },
    twitter: {
      card: 'summary',
      title,
      description,
      images: page.avatar_url ? [page.avatar_url] : [],
    },
  };
}

export function generateStaticParams() {
  return [{ slug: '_placeholder' }];
}

export default async function Page({
  params,
}: {
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const initialData = await fetchMicroPageData(slug);

  return <PublicMicroPageClient slug={slug} initialData={initialData} />;
}
