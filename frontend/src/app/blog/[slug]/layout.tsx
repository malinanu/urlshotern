import type { Metadata } from 'next';
import type { ReactNode } from 'react';
import { getPost } from '../posts';
import { JsonLd } from '@/components/JsonLd';

const SITE_URL = 'https://trunc.site';

export async function generateMetadata({ params }: { params: Promise<{ slug: string }> }): Promise<Metadata> {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) return {};

  const ogImage = post.image
    ? `${SITE_URL}${post.image}`
    : `${SITE_URL}/og-image.webp`;

  return {
    title: `${post.seoTitle ?? post.title} | Trunc`,
    description: post.description,
    alternates: { canonical: `${SITE_URL}/blog/${post.slug}` },
    openGraph: {
      title: post.seoTitle ?? post.title,
      description: post.description,
      type: 'article',
      url: `${SITE_URL}/blog/${post.slug}`,
      siteName: 'Trunc',
      publishedTime: post.publishedAt,
      modifiedTime: post.updatedAt ?? post.publishedAt,
      images: [{ url: ogImage, width: 1200, height: 630, alt: post.imageAlt ?? post.title }],
    },
    twitter: {
      card: 'summary_large_image',
      title: post.seoTitle ?? post.title,
      description: post.description,
      images: [ogImage],
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

  const ogImage = post?.image
    ? `${SITE_URL}${post.image}`
    : `${SITE_URL}/og-image.webp`;

  const articleSchema = post
    ? {
        '@context': 'https://schema.org',
        '@type': 'BlogPosting',
        headline: post.title,
        description: post.description,
        datePublished: post.publishedAt,
        dateModified: post.updatedAt ?? post.publishedAt,
        image: ogImage,
        author: {
          '@type': 'Person',
          '@id': `${SITE_URL}/blog/authors/${post.authorSlug}`,
          name: post.author,
          url: `${SITE_URL}/blog/authors/${post.authorSlug}`,
          sameAs: ['https://x.com/truncsite', 'https://linkedin.com/in/malinjayaweera'],
        },
        publisher: {
          '@type': 'Organization',
          name: 'Trunc',
          url: SITE_URL,
          logo: {
            '@type': 'ImageObject',
            url: `${SITE_URL}/icons/icon-512x512.png`,
          },
        },
        mainEntityOfPage: {
          '@type': 'WebPage',
          '@id': `${SITE_URL}/blog/${slug}`,
        },
      }
    : null;

  const howToSchema = post?.howTo
    ? {
        '@context': 'https://schema.org',
        '@type': 'HowTo',
        name: post.howTo.name,
        ...(post.howTo.totalTime ? { totalTime: post.howTo.totalTime } : {}),
        step: post.howTo.steps.map((s, i) => ({
          '@type': 'HowToStep',
          position: i + 1,
          name: s.name,
          text: s.text,
        })),
      }
    : null;

  const faqSchema = post?.faqs
    ? {
        '@context': 'https://schema.org',
        '@type': 'FAQPage',
        mainEntity: post.faqs.map((f) => ({
          '@type': 'Question',
          name: f.question,
          acceptedAnswer: { '@type': 'Answer', text: f.answer },
        })),
      }
    : null;

  const breadcrumbSchema = post
    ? {
        '@context': 'https://schema.org',
        '@type': 'BreadcrumbList',
        itemListElement: [
          {
            '@type': 'ListItem',
            position: 1,
            name: 'Home',
            item: SITE_URL,
          },
          {
            '@type': 'ListItem',
            position: 2,
            name: 'Blog',
            item: `${SITE_URL}/blog`,
          },
          {
            '@type': 'ListItem',
            position: 3,
            name: post.title,
            item: `${SITE_URL}/blog/${slug}`,
          },
        ],
      }
    : null;

  return (
    <>
      {articleSchema && <JsonLd data={articleSchema} />}
      {breadcrumbSchema && <JsonLd data={breadcrumbSchema} />}
      {howToSchema && <JsonLd data={howToSchema} />}
      {faqSchema && <JsonLd data={faqSchema} />}
      {children}
    </>
  );
}
