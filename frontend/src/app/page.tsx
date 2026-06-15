import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon, BoltIcon, CursorArrowRaysIcon, DevicePhoneMobileIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { posts } from '@/app/blog/posts';

const features = [
  {
    name: 'Lightning Fast',
    description: 'Shorten URLs instantly with our optimized infrastructure. Every redirect completes in under 50ms globally.',
    Icon: LinkIcon,
  },
  {
    name: 'Advanced Analytics',
    description: 'Track clicks, locations, devices, and referrers with real-time dashboards. No data retention limits on any plan.',
    Icon: ChartBarIcon,
  },
  {
    name: 'Secure & Reliable',
    description: 'Enterprise-grade security with 99.9% uptime, HTTPS on every link, and malware scanning on all destinations.',
    Icon: ShieldCheckIcon,
  },
];

const stats = [
  { value: '500K+', label: 'URLs shortened' },
  { value: '50M+', label: 'Clicks tracked' },
  { value: '99.9%', label: 'Uptime' },
];

const howItWorksSteps = [
  {
    step: '1',
    title: 'Paste your long URL',
    description: 'Drop any URL — a product page, blog post, campaign landing page, or any web address — into the input field.',
    Icon: CursorArrowRaysIcon,
  },
  {
    step: '2',
    title: 'Customise your slug',
    description: 'Enter a memorable short code like "spring-sale" or leave it blank to auto-generate one. Custom slugs are free.',
    Icon: LinkIcon,
  },
  {
    step: '3',
    title: 'Share your short link',
    description: 'Copy your short link (e.g. trunc.site/spring-sale) and share it on social media, email, SMS, or print.',
    Icon: BoltIcon,
  },
  {
    step: '4',
    title: 'Track every click',
    description: 'See real-time click data — geographic location, device type, referrer source — in your Trunc dashboard.',
    Icon: ChartBarIcon,
  },
];

const useCases = [
  {
    audience: 'Marketers',
    headline: 'Prove which channels drive clicks',
    description: 'Create a unique Trunc link for each campaign channel — email, social, paid ads, print. Compare click data side by side to see exactly which channels are working and where to put your next budget.',
    cta: 'Start tracking campaigns',
    href: '/register',
  },
  {
    audience: 'Developers',
    headline: 'Shorten links via REST API',
    description: 'The Trunc API is free with 100 requests per minute on the free plan. Send a POST request, get a short link back in under 100ms. Works with any language — cURL, JavaScript, Python, PHP.',
    cta: 'View API docs',
    href: '/api-docs',
  },
  {
    audience: 'Creators',
    headline: 'One bio link that tracks every tap',
    description: 'Instagram only lets you add one link in your bio. Make it a Trunc link and see exactly how many followers tap through — by day, country, and device — without paying for an expensive link-in-bio tool.',
    cta: 'Create your bio link',
    href: '/register',
    Icon: DevicePhoneMobileIcon,
  },
];

const freePlanFeatures = [
  '10 short links per month',
  'Full click analytics — geographic, device, referrer',
  'QR code for every link',
  'Custom short codes (e.g. trunc.site/your-name)',
  'Link expiration — set links to auto-expire',
  'Password-protected links',
  'AI Insights (basic)',
  'No credit card required',
  'Free forever — no trial, no expiry',
];

const testimonials = [
  {
    quote: "Trunc replaced Bitly for our whole marketing team. The free plan gives us more analytics than Bitly's paid tier — geographic data, device breakdown, real-time clicks. We cancelled Bitly the same week.",
    name: 'Sarah Chen',
    title: 'Head of Growth',
    company: 'Vela Commerce',
  },
  {
    quote: "We use the Trunc API to auto-shorten links whenever we publish a new article. Setup took 20 minutes. The rate limits on the free plan are generous enough that we haven't needed to upgrade.",
    name: 'James Kowalski',
    title: 'Engineering Lead',
    company: 'Beacon Media',
  },
  {
    quote: "As a creator I track every link I share — affiliate links, YouTube descriptions, Instagram bio. Trunc's dashboard shows me which platforms actually drive clicks. The QR code feature is a bonus for my print materials.",
    name: 'Priya Nair',
    title: 'Content Creator',
    company: '180K followers',
  },
];

const ctaFeatures = ['Free forever plan', 'No credit card required', 'Setup in 2 minutes'];

const blogSlugs = ['best-url-shorteners-2026', 'trunc-vs-bitly-tinyurl', 'how-to-shorten-a-url'];
const featuredPosts = blogSlugs.map((slug) => posts.find((p) => p.slug === slug)).filter(Boolean);

export default function HomePage() {
  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary/5 via-background to-primary/5 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30 dark:opacity-5" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary bg-primary/10 hover:bg-primary/20 transition-colors">
                ✨ New: Advanced Analytics Dashboard
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-foreground">
                Free URL Shortener —{' '}
                <span className="text-primary bg-gradient-to-r from-primary to-primary/80 bg-clip-text text-transparent">
                  Shorten Links & Track Clicks
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-muted-foreground max-w-4xl mx-auto leading-relaxed">
                Transform long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more. No credit card required.
              </p>
            </div>
          </div>

          {/* URL Shortener Component */}
          <div className="mt-16">
            <URLShortener />
          </div>

          {/* Social Proof */}
          <div className="mt-16 text-center space-y-6">
            <p className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              Trusted by thousands of businesses worldwide
            </p>
            <div className="flex flex-col sm:flex-row justify-center items-center gap-8 sm:gap-12">
              {stats.map((stat, index) => (
                <>
                  <div key={stat.label} className="flex items-center gap-3">
                    <div className="text-3xl font-bold text-foreground">{stat.value}</div>
                    <div className="text-sm text-muted-foreground font-medium">{stat.label}</div>
                  </div>
                  {index < stats.length - 1 && (
                    <div key={`divider-${index}`} className="hidden sm:block w-px h-8 bg-border" />
                  )}
                </>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* How It Works Section */}
      <section className="py-20 lg:py-24 bg-muted/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              How It Works
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              From long URL to short link in 10 seconds
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              No account required to create your first short link. Sign up free to unlock click tracking and your personal dashboard.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-4 gap-8">
            {howItWorksSteps.map((item) => (
              <div key={item.step} className="relative text-center space-y-4">
                <div className="mx-auto w-14 h-14 bg-primary rounded-full flex items-center justify-center text-primary-foreground text-xl font-bold shadow-lg">
                  {item.step}
                </div>
                <h3 className="text-lg font-semibold text-foreground">{item.title}</h3>
                <p className="text-muted-foreground text-sm leading-relaxed">{item.description}</p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              Features
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              Why Choose Trunc?
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              More than just a URL shortener. Get powerful features that help you understand and optimize your links — all free to start.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
            {features.map((feature) => (
              <Card
                key={feature.name}
                className="relative group hover:shadow-xl transition-all duration-300 bg-card border-border shadow-sm"
              >
                <CardHeader className="text-center space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <feature.Icon className="h-8 w-8 text-primary" />
                  </div>
                  <CardTitle className="text-xl font-bold text-foreground group-hover:text-primary transition-colors">
                    {feature.name}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-muted-foreground leading-relaxed text-center">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>

          <div className="mt-12 text-center">
            <Link href="/features" className="text-primary hover:text-primary/80 font-medium text-sm underline underline-offset-4">
              See all features →
            </Link>
          </div>
        </div>
      </section>

      {/* Use Cases Section */}
      <section className="py-20 lg:py-24 bg-muted/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              Use Cases
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              Built for marketers, developers, and creators
            </h2>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {useCases.map((uc) => (
              <Card key={uc.audience} className="bg-card border-border shadow-sm">
                <CardHeader>
                  <Badge variant="secondary" className="w-fit text-primary bg-primary/10 mb-2">
                    {uc.audience}
                  </Badge>
                  <CardTitle className="text-xl font-bold text-foreground">{uc.headline}</CardTitle>
                </CardHeader>
                <CardContent className="space-y-4">
                  <p className="text-muted-foreground leading-relaxed">{uc.description}</p>
                  <Link
                    href={uc.href}
                    className="inline-block text-primary hover:text-primary/80 font-medium text-sm underline underline-offset-4"
                  >
                    {uc.cta} →
                  </Link>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Free Plan Details Section */}
      <section className="py-20 lg:py-24 bg-background">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-12">
            <Badge variant="outline" className="text-primary border-primary/20">
              Free Plan
            </Badge>
            <h2 className="text-3xl sm:text-4xl font-bold text-foreground">
              Everything you need — free, forever
            </h2>
            <p className="text-xl text-muted-foreground">
              Trunc's free plan is not a crippled trial. It includes real analytics and real features with no time limit.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 gap-4 mb-10">
            {freePlanFeatures.map((feat) => (
              <div key={feat} className="flex items-start gap-3">
                <svg className="w-5 h-5 text-green-500 flex-shrink-0 mt-0.5" fill="currentColor" viewBox="0 0 20 20">
                  <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                </svg>
                <span className="text-foreground text-sm">{feat}</span>
              </div>
            ))}
          </div>

          <div className="text-center flex flex-col sm:flex-row gap-4 justify-center">
            <Button size="lg" asChild>
              <Link href="/register">Get Started Free</Link>
            </Button>
            <Button variant="outline" size="lg" asChild>
              <Link href="/pricing">Compare all plans</Link>
            </Button>
          </div>
        </div>
      </section>

      {/* Testimonials Section */}
      <section className="py-20 lg:py-24 bg-muted/30">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              Reviews
            </Badge>
            <h2 className="text-3xl sm:text-4xl font-bold text-foreground">
              What our users say
            </h2>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {testimonials.map((t) => (
              <Card key={t.name} className="bg-card border-border shadow-sm">
                <CardContent className="pt-6 space-y-4">
                  <div className="flex gap-0.5">
                    {[1, 2, 3, 4, 5].map((star) => (
                      <svg key={star} className="w-4 h-4 text-yellow-400" fill="currentColor" viewBox="0 0 20 20">
                        <path d="M9.049 2.927c.3-.921 1.603-.921 1.902 0l1.07 3.292a1 1 0 00.95.69h3.462c.969 0 1.371 1.24.588 1.81l-2.8 2.034a1 1 0 00-.364 1.118l1.07 3.292c.3.921-.755 1.688-1.54 1.118l-2.8-2.034a1 1 0 00-1.175 0l-2.8 2.034c-.784.57-1.838-.197-1.539-1.118l1.07-3.292a1 1 0 00-.364-1.118L2.98 8.72c-.783-.57-.38-1.81.588-1.81h3.461a1 1 0 00.951-.69l1.07-3.292z" />
                      </svg>
                    ))}
                  </div>
                  <p className="text-muted-foreground leading-relaxed text-sm">"{t.quote}"</p>
                  <div>
                    <p className="font-semibold text-foreground text-sm">{t.name}</p>
                    <p className="text-xs text-muted-foreground">{t.title} · {t.company}</p>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Latest from the Blog */}
      <section className="py-20 lg:py-24 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              Blog
            </Badge>
            <h2 className="text-3xl sm:text-4xl font-bold text-foreground">
              Latest from the blog
            </h2>
            <p className="text-xl text-muted-foreground">
              Guides and comparisons to help you get more from your short links.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-8">
            {featuredPosts.map((post) => post && (
              <Link key={post.slug} href={`/blog/${post.slug}`} className="group block">
                <Card className="h-full bg-card border-border shadow-sm group-hover:shadow-md transition-shadow">
                  <CardHeader>
                    <Badge variant="secondary" className="w-fit text-xs mb-2">{post.category}</Badge>
                    <CardTitle className="text-base font-semibold text-foreground group-hover:text-primary transition-colors leading-snug">
                      {post.title}
                    </CardTitle>
                  </CardHeader>
                  <CardContent>
                    <p className="text-sm text-muted-foreground leading-relaxed line-clamp-3">{post.description}</p>
                    <p className="mt-3 text-xs text-muted-foreground">{post.readTime} min read</p>
                  </CardContent>
                </Card>
              </Link>
            ))}
          </div>

          <div className="mt-10 text-center">
            <Link href="/blog" className="text-primary hover:text-primary/80 font-medium text-sm underline underline-offset-4">
              View all posts →
            </Link>
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="relative py-20 lg:py-28 bg-primary overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-black/10 to-transparent" />
        <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="space-y-8">
            <div className="space-y-4">
              <Badge className="bg-white/20 text-primary-foreground hover:bg-white/30 border-0">
                🚀 Get Started Today
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-primary-foreground">
                Ready to get started?
              </h2>
              <p className="text-xl text-primary-foreground/80 max-w-3xl mx-auto leading-relaxed">
                Join thousands of businesses using Trunc to track and optimize their links. Start for free today with no credit card required.
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-background text-primary hover:bg-muted font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href="/register">Get Started Free</a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="bg-transparent border-2 border-primary-foreground/30 text-primary-foreground hover:bg-primary-foreground/10 hover:border-primary-foreground/50 hover:text-primary-foreground font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href="/pricing">View Pricing</a>
              </Button>
            </div>

            <div className="pt-8 flex flex-col sm:flex-row justify-center items-center gap-6 text-primary-foreground/80">
              {ctaFeatures.map((feature, index) => (
                <div key={index} className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-sm">{feature}</span>
                </div>
              ))}
            </div>
          </div>
        </div>
      </section>
    </Layout>
  );
}
