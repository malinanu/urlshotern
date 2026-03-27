import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { getPost } from '../posts';
import { JsonLd } from '@/components/JsonLd';

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) return {};
  return {
    title: `${post.title} | Trunc Blog`,
    description: post.description,
    alternates: { canonical: `https://trunc.site/blog/${post.slug}` },
    openGraph: {
      title: post.title,
      description: post.description,
      type: 'article',
      publishedTime: post.publishedAt,
    },
    twitter: {
      card: 'summary_large_image',
      title: post.title,
      description: post.description,
    },
  };
}

export default async function BlogPostLayout({
  children,
  params,
}: {
  children: ReactNode;
  params: Promise<{ slug: string }>;
}) {
  const { slug } = await params;
  const post = getPost(slug);

  const articleSchema = post
    ? {
        '@context': 'https://schema.org',
        '@type': 'Article',
        headline: post.title,
        description: post.description,
        datePublished: post.publishedAt,
        publisher: {
          '@type': 'Organization',
          name: 'Trunc',
          url: 'https://trunc.site',
        },
        mainEntityOfPage: {
          '@type': 'WebPage',
          '@id': `https://trunc.site/blog/${slug}`,
        },
      }
    : null;

  return (
    <>
      {articleSchema && <JsonLd data={articleSchema} />}
      {children}
    </>
  );
}
