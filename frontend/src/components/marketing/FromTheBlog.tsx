import Link from 'next/link';
import { ArrowRight } from 'lucide-react';
import { posts } from '@/app/blog/posts';
import { Stagger, StaggerItem } from '@/components/motion/Stagger';

// Latest posts, newest first. Rendered as direct 1-hop internal links from the
// homepage so crawlers reach the blog articles from the site's top page.
const latest = [...posts]
  .sort((a, b) => new Date(b.publishedAt).getTime() - new Date(a.publishedAt).getTime())
  .slice(0, 3);

export default function FromTheBlog() {
  return (
    <section className="border-t border-border py-14 sm:py-20 lg:py-28">
      <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
        <div className="flex items-end justify-between gap-4">
          <div className="max-w-2xl">
            <h2 className="text-2xl font-bold tracking-tight text-foreground sm:text-3xl md:text-4xl">
              From the blog
            </h2>
            <p className="mt-3 text-base text-muted-foreground sm:mt-4 sm:text-lg">
              Practical guides on shortening links, tracking clicks, and growing your reach.
            </p>
          </div>
          <Link
            href="/blog"
            className="hidden shrink-0 items-center gap-1 text-sm font-medium text-primary hover:underline sm:inline-flex"
          >
            All articles
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>

        <Stagger className="mt-8 grid grid-cols-1 gap-3 sm:mt-12 sm:gap-4 md:grid-cols-3">
          {latest.map((post) => (
            <StaggerItem key={post.slug} className="h-full">
              <Link
                href={`/blog/${post.slug}`}
                className="group flex h-full flex-col rounded-[calc(var(--radius))] border border-border bg-card p-4 transition-all duration-300 hover:-translate-y-1 hover:border-primary/30 sm:p-6"
              >
                <span className="text-xs font-medium uppercase tracking-wide text-primary">
                  {post.category}
                </span>
                <h3 className="mt-3 text-lg font-semibold text-foreground group-hover:text-primary">
                  {post.title}
                </h3>
                <p className="mt-2 flex-1 text-sm leading-relaxed text-muted-foreground line-clamp-2 sm:line-clamp-none">
                  {post.description}
                </p>
                <span className="mt-4 text-xs text-muted-foreground">
                  {post.readTime} min read
                </span>
              </Link>
            </StaggerItem>
          ))}
        </Stagger>

        <div className="mt-8 sm:hidden">
          <Link
            href="/blog"
            className="inline-flex items-center gap-1 text-sm font-medium text-primary hover:underline"
          >
            All articles
            <ArrowRight className="h-4 w-4" />
          </Link>
        </div>
      </div>
    </section>
  );
}
