'use client';

import { useState } from 'react';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { posts } from './posts';
import {
  Sparkles,
  BookOpen,
  Clock,
  Calendar,
  ArrowRight,
  TrendingUp,
  Tag,
  Mail,
  CheckCircle2,
} from 'lucide-react';
import { cn } from '@/lib/utils';

export default function BlogPage() {
  const [activeCategory, setActiveCategory] = useState<string>('All');
  const [subscribed, setSubscribed] = useState(false);
  const [emailInput, setEmailInput] = useState('');

  const categories = ['All', ...Array.from(new Set(posts.map((p) => p.category)))];

  const filteredPosts =
    activeCategory === 'All'
      ? posts
      : posts.filter((p) => p.category.toLowerCase() === activeCategory.toLowerCase());

  const featuredPost = posts[0];

  const handleSubscribe = (e: React.FormEvent) => {
    e.preventDefault();
    if (emailInput.trim()) {
      setSubscribed(true);
      setEmailInput('');
    }
  };

  return (
    <Layout>
      <div className="bg-background text-foreground min-h-screen">
        
        {/* Header Hero Section */}
        <section className="relative overflow-hidden pt-16 pb-12 sm:pt-24 sm:pb-16 border-b border-border bg-gradient-to-b from-primary/10 via-card/30 to-background">
          <div className="pointer-events-none absolute inset-0 bg-grid opacity-40" aria-hidden />
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-3.5 py-1 rounded-full border border-primary/20 bg-primary/10 text-primary text-xs font-semibold uppercase tracking-wider mb-6">
              <Sparkles className="h-3.5 w-3.5" />
              The Trunc Publication
            </div>

            <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight text-foreground max-w-4xl mx-auto leading-tight">
              Insights on Links, QR Codes &{' '}
              <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
                Audience Growth
              </span>
            </h1>

            <p className="mt-5 text-base sm:text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
              Actionable guides, infrastructure benchmarks, and marketing playbooks to help you convert every digital touchpoint.
            </p>
          </div>
        </section>

        {/* Featured Post Hero Card */}
        {featuredPost && (
          <section className="py-12 border-b border-border bg-card/20">
            <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
              <div className="rounded-3xl border border-border bg-card overflow-hidden shadow-xl hover:border-primary/40 transition-all group">
                <div className="grid grid-cols-1 lg:grid-cols-12 gap-8 items-center p-6 sm:p-10">
                  
                  {/* Left Metadata & Excerpt */}
                  <div className="lg:col-span-7 space-y-4">
                    <div className="flex items-center gap-3">
                      <span className="px-3 py-1 rounded-full text-xs font-bold uppercase tracking-wider bg-primary/10 text-primary border border-primary/20">
                        Featured Article
                      </span>
                      <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                        <Calendar className="h-3.5 w-3.5" />
                        {new Date(featuredPost.publishedAt).toLocaleDateString('en-US', {
                          month: 'short',
                          day: 'numeric',
                          year: 'numeric',
                        })}
                      </span>
                      <span className="text-xs text-muted-foreground flex items-center gap-1.5">
                        <Clock className="h-3.5 w-3.5" />
                        {featuredPost.readTime} min read
                      </span>
                    </div>

                    <h2 className="text-2xl sm:text-4xl font-extrabold tracking-tight text-foreground group-hover:text-primary transition-colors">
                      <Link href={`/blog/${featuredPost.slug}`}>
                        {featuredPost.title}
                      </Link>
                    </h2>

                    <p className="text-sm sm:text-base text-muted-foreground leading-relaxed line-clamp-3">
                      {featuredPost.description}
                    </p>

                    <div className="pt-2 flex items-center justify-between">
                      <div className="flex items-center gap-3">
                        <div className="w-9 h-9 rounded-full bg-gradient-to-r from-blue-600 to-indigo-600 p-0.5 shadow-sm">
                          <div className="w-full h-full bg-card rounded-full flex items-center justify-center font-bold text-xs text-foreground">
                            AR
                          </div>
                        </div>
                        <div>
                          <div className="text-xs font-bold text-foreground">{featuredPost.author}</div>
                          <div className="text-[11px] text-muted-foreground">Lead Infrastructure Architect</div>
                        </div>
                      </div>

                      <Link
                        href={`/blog/${featuredPost.slug}`}
                        className="inline-flex items-center gap-1.5 text-xs sm:text-sm font-semibold text-primary group-hover:translate-x-1 transition-transform"
                      >
                        <span>Read Full Guide</span>
                        <ArrowRight className="h-4 w-4" />
                      </Link>
                    </div>
                  </div>

                  {/* Right Stylized Graphic Banner */}
                  <div className="lg:col-span-5">
                    <div className="rounded-2xl bg-gradient-to-br from-primary/20 via-sky-500/10 to-indigo-500/20 border border-primary/20 p-8 text-center flex flex-col items-center justify-center min-h-[220px]">
                      <div className="w-16 h-16 rounded-2xl bg-primary/20 text-primary border border-primary/30 flex items-center justify-center text-2xl font-bold mb-3 shadow-lg">
                        🔗
                      </div>
                      <div className="font-bold text-foreground text-sm uppercase tracking-widest">
                        URL Shortening Mastery
                      </div>
                      <div className="text-xs text-muted-foreground mt-1">
                        Sub-15ms Edge Routing · Analytics · Custom Vanity
                      </div>
                    </div>
                  </div>

                </div>
              </div>
            </div>
          </section>
        )}

        {/* Category Tabs & Article Grid */}
        <section className="py-16 sm:py-20 border-b border-border">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            
            {/* Category Filter Pills */}
            <div className="flex items-center gap-2 overflow-x-auto pb-4 mb-10">
              {categories.map((cat) => {
                const isActive = activeCategory.toLowerCase() === cat.toLowerCase();
                return (
                  <button
                    key={cat}
                    onClick={() => setActiveCategory(cat)}
                    className={cn(
                      'px-4 py-2 rounded-xl text-xs sm:text-sm font-semibold transition-all whitespace-nowrap',
                      isActive
                        ? 'bg-primary text-primary-foreground shadow-sm'
                        : 'bg-card border border-border text-muted-foreground hover:text-foreground hover:bg-muted'
                    )}
                  >
                    {cat}
                  </button>
                );
              })}
            </div>

            {/* Articles Grid */}
            <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-3 gap-8">
              {filteredPosts.map((post) => (
                <article
                  key={post.slug}
                  className="rounded-3xl border border-border bg-card p-6 flex flex-col justify-between hover:border-primary/40 hover:shadow-lg transition-all group"
                >
                  <div>
                    {/* Top Tag & Time */}
                    <div className="flex items-center justify-between text-xs text-muted-foreground mb-4">
                      <span className="px-2.5 py-0.5 rounded-full font-semibold bg-muted border border-border text-foreground text-[11px]">
                        {post.category}
                      </span>
                      <span className="flex items-center gap-1 font-mono text-[11px]">
                        <Clock className="h-3 w-3" />
                        {post.readTime} min
                      </span>
                    </div>

                    <h3 className="text-xl font-bold tracking-tight text-foreground group-hover:text-primary transition-colors line-clamp-2">
                      <Link href={`/blog/${post.slug}`}>
                        {post.title}
                      </Link>
                    </h3>

                    <p className="mt-3 text-xs sm:text-sm text-muted-foreground leading-relaxed line-clamp-3">
                      {post.description}
                    </p>
                  </div>

                  <div className="mt-6 pt-4 border-t border-border flex items-center justify-between text-xs">
                    <span className="text-muted-foreground">
                      {new Date(post.publishedAt).toLocaleDateString('en-US', {
                        month: 'short',
                        day: 'numeric',
                      })}
                    </span>
                    <Link
                      href={`/blog/${post.slug}`}
                      className="font-semibold text-primary group-hover:translate-x-1 transition-transform inline-flex items-center gap-1"
                    >
                      <span>Read article</span>
                      <ArrowRight className="h-3.5 w-3.5" />
                    </Link>
                  </div>
                </article>
              ))}
            </div>

          </div>
        </section>

        {/* Newsletter Capture Section */}
        <section className="py-20 bg-gradient-to-br from-card via-background to-card text-center">
          <div className="mx-auto max-w-3xl px-4 sm:px-6 lg:px-8">
            <div className="w-12 h-12 rounded-2xl bg-primary/10 border border-primary/20 text-primary mx-auto flex items-center justify-center mb-4">
              <Mail className="h-6 w-6" />
            </div>
            
            <h2 className="text-3xl font-extrabold tracking-tight text-foreground sm:text-4xl">
              Stay ahead of digital link strategy
            </h2>
            <p className="mt-3 text-sm sm:text-base text-muted-foreground max-w-xl mx-auto">
              Join 15,000+ creators, marketers, and developers getting our bi-weekly teardown of top growth hacks, link metrics, and platform updates.
            </p>

            {subscribed ? (
              <div className="mt-6 inline-flex items-center gap-2 px-4 py-2.5 rounded-xl bg-emerald-500/10 text-emerald-400 border border-emerald-500/20 text-sm font-semibold">
                <CheckCircle2 className="h-4 w-4" /> You are subscribed! Welcome aboard.
              </div>
            ) : (
              <form onSubmit={handleSubscribe} className="mt-8 flex flex-col sm:flex-row gap-3 max-w-md mx-auto">
                <input
                  type="email"
                  required
                  value={emailInput}
                  onChange={(e) => setEmailInput(e.target.value)}
                  placeholder="Enter your work email"
                  className="flex-1 px-4 py-3 rounded-xl bg-card border border-border text-foreground text-sm focus:outline-none focus:border-primary focus:ring-2 focus:ring-primary/20"
                />
                <button
                  type="submit"
                  className="px-6 py-3 rounded-xl bg-primary text-primary-foreground font-semibold text-sm hover:bg-primary/90 shadow-md transition-all shrink-0"
                >
                  Subscribe Free
                </button>
              </form>
            )}
          </div>
        </section>

      </div>
    </Layout>
  );
}
