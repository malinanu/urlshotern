import type { Metadata } from 'next';
import PublicBuilderPageClient from './_client';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

async function fetchPageMeta(slug: string) {
  try {
    const res = await fetch(`${API_BASE_URL}/api/v1/pages/${slug}`, {
      next: { revalidate: 300 },
    });
    if (!res.ok) return null;
    const data = await res.json();
    return data.page ?? data ?? null;
  } catch {
    return null;
  }
}

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const page = await fetchPageMeta(slug);
  if (!page) return { title: 'Page Not Found - Trunc' };

  const title = page.meta_title || page.title;
  const description = page.meta_description || '';

  return {
    title: `${title} - Trunc`,
    description,
    openGraph: { title: `${title} - Trunc`, description },
    twitter: { card: 'summary', title: `${title} - Trunc`, description },
  };
}

export function generateStaticParams() {
  return [{ slug: '_placeholder' }];
}

export default async function Page({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  return <PublicBuilderPageClient slug={slug} />;
}
