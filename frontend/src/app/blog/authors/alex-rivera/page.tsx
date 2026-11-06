import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { posts } from '@/app/blog/posts';

const authorPosts = posts.filter((p) => p.authorSlug === 'alex-rivera');

export default function AlexRiveraAuthorPage() {
  return (
    <Layout>
      <div className="bg-background min-h-screen">
        {/* Hero */}
        <div className="bg-gradient-to-br from-primary/5 via-background to-primary/5 border-b border-border py-14 lg:py-20">
          <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8">
            <Link href="/blog" className="text-sm text-muted-foreground hover:text-foreground transition-colors mb-6 inline-block">
              ← Blog
            </Link>
            <div className="flex items-center gap-6">
              <div className="flex-shrink-0 w-20 h-20 rounded-full bg-primary/10 border-2 border-primary/20 flex items-center justify-center text-3xl font-bold text-primary select-none">
                AR
              </div>
              <div>
                <h1 className="text-3xl sm:text-4xl font-bold text-foreground tracking-tight">
                  Alex Rivera
                </h1>
                <p className="text-muted-foreground mt-1 text-lg">Lead Infrastructure Architect at Trunc</p>
              </div>
            </div>
          </div>
        </div>

        {/* Bio */}
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
          <div className="prose prose-neutral dark:prose-invert max-w-none">
            <p className="text-foreground/80 leading-relaxed text-base">
              Alex Rivera is a lead infrastructure architect at Trunc, specializing in global edge redirection,
              click analytics telemetry, and digital business card platforms. They have spent over a decade
              building high-throughput web infrastructure and developer tooling for scaling teams.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              Before joining Trunc, Alex led infrastructure at several high-growth startups, focusing on
              sub-millisecond latency, DNS routing, and distributed Redis architectures.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              Alex writes about URL management, link analytics, QR code vector workflows, and evidence-based digital marketing
              architecture.
            </p>
          </div>

          {/* Expertise tags */}
          <div className="flex flex-wrap gap-2 mt-6">
            {['URL shortening', 'Link analytics', 'Digital marketing', 'Growth strategy', 'Web analytics', 'SaaS'].map((tag) => (
              <span
                key={tag}
                className="text-xs font-medium text-primary bg-primary/10 px-3 py-1 rounded-full"
              >
                {tag}
              </span>
            ))}
          </div>
        </div>

        {/* Posts by this author */}
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 pb-16">
          <div className="border-t border-border pt-10">
            <h2 className="text-xl font-bold text-foreground mb-6">
              Articles by Alex Rivera
            </h2>
            <div className="space-y-6">
              {authorPosts.map((post) => (
                <Link
                  key={post.slug}
                  href={`/blog/${post.slug}`}
                  className="group block rounded-xl border border-border p-6 hover:border-primary/50 hover:bg-primary/5 transition-colors"
                >
                  <div className="flex items-center gap-2 mb-2">
                    <span className="text-xs font-semibold uppercase tracking-wide text-primary">
                      {post.category}
                    </span>
                    <span className="text-muted-foreground text-xs">·</span>
                    <span className="text-xs text-muted-foreground">{post.readTime} min read</span>
                  </div>
                  <h3 className="text-base font-semibold text-foreground group-hover:text-primary transition-colors leading-snug">
                    {post.title}
                  </h3>
                  <p className="text-sm text-muted-foreground mt-2 leading-relaxed line-clamp-2">
                    {post.description}
                  </p>
                  <time
                    dateTime={post.publishedAt}
                    className="text-xs text-muted-foreground mt-3 block"
                  >
                    {new Date(post.publishedAt).toLocaleDateString('en-US', {
                      year: 'numeric',
                      month: 'long',
                      day: 'numeric',
                    })}
                  </time>
                </Link>
              ))}
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}
