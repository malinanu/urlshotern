import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { posts } from './posts';

export default function BlogPage() {
  return (
    <Layout>
      <div className="bg-background min-h-screen">
        {/* Header */}
        <div className="bg-gradient-to-br from-primary/5 via-background to-primary/5 py-16 lg:py-20 border-b border-border">
          <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <p className="text-sm font-semibold text-primary uppercase tracking-widest mb-3">Blog</p>
            <h1 className="text-4xl sm:text-5xl font-bold text-foreground tracking-tight">
              URL Shortener Tips, Guides & News
            </h1>
            <p className="mt-4 text-xl text-muted-foreground max-w-2xl mx-auto">
              Learn how to shorten URLs, track clicks, run better campaigns, and get more from your links.
            </p>
          </div>
        </div>

        {/* Posts */}
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 py-16">
          <div className="divide-y divide-border">
            {posts.map((post) => (
              <article key={post.slug} className="py-10 first:pt-0">
                <div className="flex items-center gap-3 mb-3">
                  <span className="text-xs font-semibold uppercase tracking-wide text-primary bg-primary/10 px-2.5 py-1 rounded-full">
                    {post.category}
                  </span>
                  <time className="text-sm text-muted-foreground">
                    {new Date(post.publishedAt).toLocaleDateString('en-US', { year: 'numeric', month: 'long', day: 'numeric' })}
                  </time>
                  <span className="text-sm text-muted-foreground">· {post.readTime} min read</span>
                </div>
                <h2 className="text-2xl font-bold text-foreground hover:text-primary transition-colors mb-3">
                  <Link href={`/blog/${post.slug}`}>{post.title}</Link>
                </h2>
                <p className="text-muted-foreground leading-relaxed mb-4">{post.description}</p>
                <Link
                  href={`/blog/${post.slug}`}
                  className="text-primary font-medium hover:underline text-sm"
                >
                  Read article →
                </Link>
              </article>
            ))}
          </div>
        </div>
      </div>
    </Layout>
  );
}
