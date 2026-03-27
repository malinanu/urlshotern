import { notFound } from 'next/navigation';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { getPost, posts, type ContentSection } from '../posts';

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
            <div className="flex items-center gap-4 mt-5 text-sm text-muted-foreground">
              <time dateTime={post.publishedAt}>
                {new Date(post.publishedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
              </time>
              <span>·</span>
              <span>{post.readTime} min read</span>
            </div>
          </div>
        </div>

        {/* Content */}
        <article className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-14">
          {post.content.map((section, i) => (
            <ContentBlock key={i} section={section} />
          ))}
        </article>

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
