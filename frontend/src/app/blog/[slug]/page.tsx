import { notFound } from 'next/navigation';
import Image from 'next/image';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { getPost, posts, type BlogPost, type ContentSection } from '../posts';
import {
  Calendar,
  Clock,
  ArrowLeft,
  Share2,
  Bookmark,
  CheckCircle2,
  Sparkles,
} from 'lucide-react';

export function generateStaticParams() {
  return posts.map((post) => ({ slug: post.slug }));
}

function ContentBlock({ section }: { section: ContentSection }) {
  switch (section.type) {
    case 'h2':
      return (
        <h2 className="text-2xl sm:text-3xl font-extrabold text-foreground mt-12 mb-4 tracking-tight border-b border-border/60 pb-2">
          {section.text}
        </h2>
      );
    case 'h3':
      return (
        <h3 className="text-xl sm:text-2xl font-bold text-foreground mt-8 mb-3 tracking-tight">
          {section.text}
        </h3>
      );
    case 'p':
      return (
        <p className="text-foreground/90 leading-relaxed text-base sm:text-lg mb-6">
          {section.text}
        </p>
      );
    case 'ul':
      return (
        <ul className="list-disc list-outside pl-6 mb-6 space-y-2 text-foreground/90 leading-relaxed text-base sm:text-lg">
          {section.items.map((item, i) => (
            <li key={i}>{item}</li>
          ))}
        </ul>
      );
    case 'ol':
      return (
        <ol className="list-decimal list-outside pl-6 mb-6 space-y-2 text-foreground/90 leading-relaxed text-base sm:text-lg">
          {section.items.map((item, i) => (
            <li key={i}>{item}</li>
          ))}
        </ol>
      );
    case 'table':
      return (
        <div className="overflow-x-auto my-8 rounded-2xl border border-border shadow-md">
          <table className="w-full text-sm">
            <thead className="bg-muted/70 border-b border-border">
              <tr>
                {section.headers.map((h) => (
                  <th key={h} className="px-5 py-3.5 text-left font-bold text-foreground">
                    {h}
                  </th>
                ))}
              </tr>
            </thead>
            <tbody className="divide-y divide-border">
              {section.rows.map((row, i) => (
                <tr key={i} className={i % 2 === 0 ? 'bg-card' : 'bg-muted/20'}>
                  {row.map((cell, j) => (
                    <td
                      key={j}
                      className={`px-5 py-3.5 text-foreground/90 ${j === 0 ? 'font-semibold text-foreground' : ''}`}
                    >
                      {cell}
                    </td>
                  ))}
                </tr>
              ))}
            </tbody>
          </table>
        </div>
      );
    case 'callout':
      return (
        <div className="bg-primary/10 border-l-4 border-primary rounded-r-2xl p-5 my-8 flex items-start gap-3">
          <Sparkles className="h-5 w-5 text-primary shrink-0 mt-0.5" />
          <p className="text-foreground text-sm sm:text-base leading-relaxed font-medium">
            {section.text}
          </p>
        </div>
      );
    case 'cta':
      return (
        <div className="rounded-2xl border border-primary/30 bg-gradient-to-br from-primary/10 via-card to-card p-6 sm:p-8 my-10 text-center shadow-lg">
          <p className="text-foreground text-base sm:text-lg font-bold mb-4 max-w-lg mx-auto">
            {section.text}
          </p>
          <a
            href={section.href}
            className="inline-flex items-center gap-2 px-6 py-3 rounded-xl bg-primary text-primary-foreground font-semibold text-sm hover:bg-primary/90 transition-all shadow-md"
          >
            {section.label}
          </a>
        </div>
      );
    default:
      return null;
  }
}

export default async function BlogPostPage({ params }: { params: Promise<{ slug: string }> }) {
  const { slug } = await params;
  const post = getPost(slug);

  if (!post) {
    notFound();
  }

  return (
    <Layout>
      <div className="bg-background text-foreground min-h-screen">
        
        {/* Article Header */}
        <header className="border-b border-border bg-gradient-to-b from-primary/10 via-card/20 to-background py-14 sm:py-20">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
            
            <Link
              href="/blog"
              className="inline-flex items-center gap-1.5 text-xs sm:text-sm font-semibold text-muted-foreground hover:text-primary transition-colors mb-6"
            >
              <ArrowLeft className="h-4 w-4" />
              <span>Back to all guides & news</span>
            </Link>

            <div className="flex flex-wrap items-center gap-3 mb-4">
              <span className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-primary/10 text-primary border border-primary/20">
                {post.category}
              </span>
              <span className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1">
                <Calendar className="h-3.5 w-3.5" />
                {new Date(post.publishedAt).toLocaleDateString('en-US', {
                  month: 'long',
                  day: 'numeric',
                  year: 'numeric',
                })}
              </span>
              <span className="text-xs sm:text-sm text-muted-foreground flex items-center gap-1 font-mono">
                <Clock className="h-3.5 w-3.5" />
                {post.readTime} min read
              </span>
            </div>

            <h1 className="text-3xl sm:text-5xl font-black tracking-tight text-foreground leading-tight">
              {post.title}
            </h1>

            <p className="mt-4 text-base sm:text-xl text-muted-foreground leading-relaxed">
              {post.description}
            </p>

            {/* Author Attribution */}
            <div className="mt-8 pt-6 border-t border-border flex items-center justify-between">
              <div className="flex items-center gap-3">
                <div className="w-10 h-10 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600 p-0.5 shadow-sm">
                  <div className="w-full h-full bg-card rounded-full flex items-center justify-center font-bold text-xs text-foreground">
                    MJ
                  </div>
                </div>
                <div>
                  <div className="text-sm font-bold text-foreground">{post.author}</div>
                  <div className="text-xs text-muted-foreground">Founder & Tech Lead</div>
                </div>
              </div>

              <div className="text-xs text-muted-foreground">
                Updated {post.updatedAt ? new Date(post.updatedAt).toLocaleDateString('en-US', { month: 'short', year: 'numeric' }) : 'recently'}
              </div>
            </div>

          </div>
        </header>

        {/* Hero image if exists */}
        {post.image && (
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 -mb-6 pt-8">
            <Image
              src={post.image}
              alt={post.imageAlt ?? post.title}
              width={1200}
              height={630}
              className="rounded-2xl w-full h-auto shadow-xl border border-border"
              priority
            />
          </div>
        )}

        {/* Main Content Body */}
        <article className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-14">
          {post.content.map((section, i) => (
            <ContentBlock key={i} section={section} />
          ))}
        </article>

        {/* About the Author Card */}
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 pb-14">
          <div className="border-t border-border pt-10">
            <div className="flex flex-col sm:flex-row items-start gap-5 bg-card rounded-2xl p-6 sm:p-8 border border-border shadow-md">
              <div className="w-16 h-16 rounded-2xl bg-gradient-to-br from-primary to-indigo-600 text-primary-foreground flex items-center justify-center text-xl font-bold shrink-0 shadow-md">
                MJ
              </div>
              <div>
                <p className="text-xs font-bold uppercase tracking-wider text-primary mb-1">About the Author</p>
                <div className="text-lg font-bold text-foreground">{post.author}</div>
                <p className="text-xs text-muted-foreground mt-0.5 mb-3">Founder & Lead Engineer at Trunc</p>
                <p className="text-sm text-foreground/80 leading-relaxed">
                  Malin is the founder of Trunc and has spent over a decade building marketing, DNS, and analytics tools for growth teams. He writes about URL management, digital business cards, link telemetry, and developer infrastructure.
                </p>
              </div>
            </div>
          </div>
        </div>

        {/* Related Articles */}
        {(() => {
          const related = posts.filter((p: BlogPost) => p.slug !== post.slug).slice(0, 2);
          if (related.length === 0) return null;
          return (
            <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 pb-16">
              <div className="border-t border-border pt-10">
                <h2 className="text-xl font-bold text-foreground mb-6">Related Articles</h2>
                <div className="grid sm:grid-cols-2 gap-6">
                  {related.map((r: BlogPost) => (
                    <Link
                      key={r.slug}
                      href={`/blog/${r.slug}`}
                      className="group block rounded-2xl border border-border bg-card p-6 hover:border-primary/50 transition-all shadow-sm"
                    >
                      <span className="text-[11px] font-bold uppercase tracking-wider text-primary px-2.5 py-0.5 rounded-full bg-primary/10">
                        {r.category}
                      </span>
                      <h3 className="mt-3 text-base font-bold text-foreground group-hover:text-primary transition-colors leading-snug">
                        {r.title}
                      </h3>
                      <p className="mt-2 text-xs text-muted-foreground line-clamp-2">
                        {r.description}
                      </p>
                      <div className="mt-4 text-xs text-muted-foreground font-mono">
                        {r.readTime} min read
                      </div>
                    </Link>
                  ))}
                </div>
              </div>
            </div>
          );
        })()}

      </div>
    </Layout>
  );
}
