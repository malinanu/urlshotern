import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { posts } from '@/app/blog/posts';

const authorPosts = posts.filter((p) => p.authorSlug === 'malin-jayaweera');

export default function MalinJayaweeraAuthorPage() {
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
                MJ
              </div>
              <div>
                <h1 className="text-3xl sm:text-4xl font-bold text-foreground tracking-tight">
                  Malin Jayaweera
                </h1>
                <p className="text-muted-foreground mt-1 text-lg">Founder &amp; CEO at Trunc</p>
              </div>
            </div>
          </div>
        </div>

        {/* Bio */}
        <div className="max-w-3xl mx-auto px-4 sm:px-6 lg:px-8 py-10">
          <div className="prose prose-neutral dark:prose-invert max-w-none">
            <p className="text-foreground/80 leading-relaxed text-base">
              Malin Jayaweera is the founder and CEO of Trunc, a free URL shortener with real-time
              click analytics, custom branded domains, QR code generation, and AI-powered insights.
              He has spent over a decade building marketing infrastructure and analytics tools for
              growth teams across e-commerce, SaaS, and media.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              Before starting Trunc, Malin worked as a software engineer and growth lead at several
              early-stage startups, where he repeatedly found that marketers were paying for expensive
              link management tools that offered less transparency than they deserved. Trunc was built
              to fix that — giving every team, from solo creators to enterprise marketing departments,
              full analytics from day one without the enterprise price tag.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              Malin writes about URL management, link analytics, and evidence-based digital marketing
              strategy. His writing draws on real data from the Trunc platform and first-hand experience
              running campaigns across paid search, email, and social.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              His articles cover the full spectrum of link management: from practical how-to guides on
              shortening URLs and tracking affiliate links, to in-depth comparisons of URL shortener
              platforms, to developer tutorials on integrating link shortening via REST API. Every piece
              is grounded in real platform data and tested workflows rather than generic marketing advice.
            </p>
            <p className="text-foreground/80 leading-relaxed text-base mt-4">
              Outside of writing, Malin focuses on building features that close the gap between
              enterprise-grade analytics and free-tier accessibility — including Trunc's AI Insights
              engine, which analyses click patterns to surface actionable recommendations for marketers
              who do not have a data team behind them.
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
              Articles by Malin Jayaweera
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
