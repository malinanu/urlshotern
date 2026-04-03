import { notFound } from 'next/navigation';
import Image from 'next/image';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { getPost, posts, type BlogPost, type ContentSection } from '../posts';

export function generateStaticParams() {
  return posts.map((post) => ({ slug: post.slug }));
}

function ContentBlock({ section }: { section: ContentSection }) {
  switch (section.type) {
    case 'h2':
      return <h2 className="text-2xl font-bold text-foreground mt-10 mb-4">{section.text}</h2>;
    case 'h3':
      return <h3 className="text-xl font-semibold text-foreground mt-8 mb-3">{section.text}</h3>;
    case 'p':
      return <p className="text-foreground/80 leading-relaxed mb-5">{section.text}</p>;
    case 'ul':
      return (
        <ul className="list-disc list-outside pl-6 mb-5 space-y-2">
          {section.items.map((item, i) => (
            <li key={i} className="text-foreground/80 leading-relaxed">{item}</li>
          ))}
        </ul>
      );
    case 'ol':
      return (
        <ol className="list-decimal list-outside pl-6 mb-5 space-y-2">
          {section.items.map((item, i) => (
            <li key={i} className="text-foreground/80 leading-relaxed">{item}</li>
          ))}
        </ol>
      );
    case 'table':
      return (
        <div className="overflow-x-auto my-6 rounded-lg border border-border">
          <table className="w-full text-sm">
            <thead className="bg-muted">
              <tr>
                {section.headers.map((h) => (
                  <th key={h} className="px-4 py-3 text-left font-semibold text-foreground">{h}</th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {section.rows.map((row, i) => (
                <tr key={i} className={i % 2 === 0 ? 'bg-background' : 'bg-muted/30'}>
                  {row.map((cell, j) => (
                    <td key={j} className={`px-4 py-3 text-foreground/80 ${j === 0 ? 'font-medium' : ''}`}>{cell}</td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
    case 'callout':
      return (
        <div className="bg-primary/10 border-l-4 border-primary rounded-r-lg px-5 py-4 my-6">
          <p className="text-foreground font-medium leading-relaxed">{section.text}</p>
        </div>
      );
    case 'cta':
      return (
        <div className="bg-gradient-to-r from-primary to-primary/80 rounded-xl p-8 my-10 text-center">
          <p className="text-primary-foreground text-lg font-medium mb-5">{section.text}</p>
          <Link
            href={section.href}
            className="inline-block bg-white text-primary font-semibold px-8 py-3 rounded-lg hover:bg-gray-50 transition-colors shadow"
          >
            {section.label}
          </Link>
        </div>
      );
    default:
      return null;
  }
}

export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = getPost(slug);
  if (!post) notFound();

  return (
    <Layout>
      <div className="bg-background min-h-screen">
        {/* Hero */}
        <div className="bg-gradient-to-br from-primary/5 via-background to-primary/5 py-14 lg:py-20 border-b border-border">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex items-center gap-3 mb-4">
              <Link href="/blog" className="text-sm text-muted-foreground hover:text-foreground transition-colors">
                ← Blog
              </Link>
              <span className="text-xs font-semibold uppercase tracking-wide text-primary bg-primary/10 px-2.5 py-1 rounded-full">
                {post.category}
              </span>
            </div>
            <h1 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground tracking-tight leading-tight">
              {post.title}
            </h1>
            <div className="flex items-center gap-4 mt-5 text-sm text-muted-foreground flex-wrap">
              <Link
                href={`/blog/authors/${post.authorSlug}`}
                className="font-medium text-foreground hover:text-primary transition-colors"
              >
                {post.author}
              </Link>
              <span>·</span>
              <time dateTime={post.publishedAt}>
                {new Date(post.publishedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
              </time>
              {post.updatedAt && post.updatedAt !== post.publishedAt && (
                <>
                  <span>·</span>
                  <span>Updated {new Date(post.updatedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}</span>
                </>
              )}
              <span>·</span>
              <span>{post.readTime} min read</span>
            </div>
          </div>
        </div>

        {/* Hero image */}
        {post.image && (
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 -mb-6 pt-8">
            <Image
              src={post.image}
              alt={post.imageAlt ?? post.title}
              width={1200}
              height={630}
              className="rounded-xl w-full h-auto shadow-md"
              priority
            />
          </div>
        )}

        {/* Content */}
        <article className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-14">
          {post.content.map((section, i) => (
            <ContentBlock key={i} section={section} />
          ))}
        </article>

        {/* About the Author */}
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 pb-10">
          <div className="border-t border-border pt-10">
            <div className="flex items-start gap-5 bg-muted/40 rounded-xl p-6 border border-border">
              <div className="flex-shrink-0 w-14 h-14 rounded-full bg-primary/10 border-2 border-primary/20 flex items-center justify-center text-2xl font-bold text-primary select-none">
                MJ
              </div>
              <div>
                <p className="text-xs font-semibold uppercase tracking-wide text-muted-foreground mb-1">Written by</p>
                <Link
                  href={`/blog/authors/${post.authorSlug}`}
                  className="text-base font-bold text-foreground hover:text-primary transition-colors"
                >
                  {post.author}
                </Link>
                <p className="text-sm text-muted-foreground mt-0.5 mb-2">Founder & CEO at Trunc</p>
                <p className="text-sm text-foreground/80 leading-relaxed">
                  Malin is the founder of Trunc and has spent over a decade building marketing and analytics tools for growth teams. He writes about URL management, link analytics, and digital marketing strategy.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Related articles */}
        {(() => {
          const related = posts.filter((p: BlogPost) => p.slug !== post.slug);
          if (related.length === 0) return null;
          return (
            <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 pb-10">
              <div className="border-t border-border pt-10">
                <h2 className="text-lg font-semibold text-foreground mb-5">Related articles</h2>
                <div className="grid sm:grid-cols-2 gap-4">
                  {related.map((r: BlogPost) => (
                    <Link
                      key={r.slug}
                      href={`/blog/${r.slug}`}
                      className="group block rounded-lg border border-border p-4 hover:border-primary/50 hover:bg-primary/5 transition-colors"
                    >
                      <span className="text-xs font-semibold uppercase tracking-wide text-primary">{r.category}</span>
                      <p className="mt-1 text-sm font-medium text-foreground group-hover:text-primary transition-colors leading-snug">
                        {r.title}
                      </p>
                      <p className="mt-1 text-xs text-muted-foreground">{r.readTime} min read</p>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          );
        })()}

        {/* Back to Blog */}
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 pb-16">
          <div className="border-t border-border pt-8">
            <Link href="/blog" className="text-primary font-medium hover:underline">
              ← Back to all articles
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
}
