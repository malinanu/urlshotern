'use client';

import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { useCMSPage, getContentBlock } from '@/hooks/useCMSPage';

// Fallback features data (used if CMS fetch fails)
const fallbackFeatures = [
  {
    name: 'Lightning Fast',
    description: 'Shorten URLs instantly with our optimized infrastructure.',
    icon: LinkIcon,
  },
  {
    name: 'Advanced Analytics',
    description: 'Track clicks, locations, and referrers with detailed insights.',
    icon: ChartBarIcon,
  },
  {
    name: 'Secure & Reliable',
    description: 'Enterprise-grade security with 99.9% uptime guarantee.',
    icon: ShieldCheckIcon,
  },
];

const iconMap: Record<string, any> = {
  LinkIcon,
  ChartBarIcon,
  ShieldCheckIcon,
};

export default function HomePage() {
  // Fetch page content from CMS with fallback
  const { content: page, loading, usedCMS } = useCMSPage('home', {
    slug: 'home',
    title: 'Free URL Shortener — Shorten Links & Track Clicks',
    content: 'Transform long URLs into short, trackable links with powerful analytics.',
  });

  // Extract content blocks from CMS or use fallback
  const hero = getContentBlock(page, 'hero', {
    badge: '✨ New: Advanced Analytics Dashboard',
    title: 'Free URL Shortener — Shorten Links & Track Clicks',
    title_highlight: 'Shorten Links & Track Clicks',
    tagline: 'Transform long URLs into short, trackable links with powerful analytics. Perfect for marketing campaigns, social media, and more. No credit card required.',
  });

  const socialProof = getContentBlock(page, 'social_proof', {
    heading: 'Trusted by thousands of businesses worldwide',
    stats: [
      { value: '500K+', label: 'URLs shortened' },
      { value: '50M+', label: 'Clicks tracked' },
      { value: '99.9%', label: 'Uptime' },
    ],
  });

  const featuresBlock = getContentBlock(page, 'features', {
    badge: 'Features',
    heading: 'Why Choose Trunc?',
    tagline: 'More than just a URL shortener. Get powerful features that help you understand and optimize your links.',
    items: fallbackFeatures,
  });

  const ctaBottom = getContentBlock(page, 'cta_bottom', {
    badge: '🚀 Get Started Today',
    heading: 'Ready to get started?',
    tagline: 'Join thousands of businesses using Trunc to track and optimize their links. Start for free today with no credit card required.',
    cta_primary: { text: 'Get Started Free', url: '/register' },
    cta_secondary: { text: 'View Pricing', url: '/pricing' },
    features: ['Free forever plan', 'No credit card required', 'Setup in 2 minutes'],
  });

  // Map icon names to actual components
  const features = (featuresBlock.items || fallbackFeatures).map((feature: any) => ({
    ...feature,
    icon: feature.icon && typeof feature.icon === 'string' ? (iconMap[feature.icon] || LinkIcon) : feature.icon || LinkIcon,
  }));

  return (
    <Layout>
      {/* Hero Section - Redesigned with shadcn components */}
      <section className="relative bg-gradient-to-br from-primary/5 via-background to-primary/5 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30 dark:opacity-5" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary bg-primary/10 hover:bg-primary/20 transition-colors">
                {hero.badge}
              </Badge>
              <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-foreground">
                {hero.title.replace(hero.title_highlight, '')}{' '}
                <span className="text-primary bg-gradient-to-r from-primary to-primary/80 bg-clip-text text-transparent">
                  {hero.title_highlight}
                </span>
              </h1>
              <p className="text-xl sm:text-2xl text-muted-foreground max-w-4xl mx-auto leading-relaxed">
                {hero.tagline}
              </p>
            </div>
          </div>

          {/* Trunc Component */}
          <div className="mt-16">
            <URLShortener />
          </div>

          {/* Social Proof */}
          <div className="mt-16 text-center space-y-6">
            <p className="text-sm font-medium text-muted-foreground uppercase tracking-wide">
              {socialProof.heading}
            </p>
            <div className="flex flex-col sm:flex-row justify-center items-center gap-8 sm:gap-12">
              {(socialProof?.stats || []).map((stat: any, index: number) => (
                <>
                  <div key={stat.label} className="flex items-center gap-3">
                    <div className="text-3xl font-bold text-foreground">{stat.value}</div>
                    <div className="text-sm text-muted-foreground font-medium">{stat.label}</div>
                  </div>
                  {index < (socialProof?.stats || []).length - 1 && (
                    <div key={`divider-${index}`} className="hidden sm:block w-px h-8 bg-border" />
                  )}
                </>
              ))}
            </div>
          </div>
        </div>
      </section>

      {/* Features Section - Redesigned with shadcn Cards */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary border-primary/20">
              {featuresBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
              {featuresBlock.heading}
            </h2>
            <p className="text-xl text-muted-foreground max-w-3xl mx-auto">
              {featuresBlock.tagline}
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
            {(features || []).map((feature) => (
              <Card
                key={feature.name}
                className="relative group hover:shadow-xl transition-all duration-300 bg-card border-border shadow-sm"
              >
                <CardHeader className="text-center space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <feature.icon className="h-8 w-8 text-primary" />
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
        </div>
      </section>

      {/* CTA Section - Redesigned with shadcn Buttons */}
      <section className="relative py-20 lg:py-28 bg-primary overflow-hidden">
        <div className="absolute inset-0 bg-gradient-to-br from-black/10 to-transparent" />
        <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <div className="space-y-8">
            <div className="space-y-4">
              <Badge className="bg-white/20 text-primary-foreground hover:bg-white/30 border-0">
                {ctaBottom.badge}
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-primary-foreground">
                {ctaBottom.heading}
              </h2>
              <p className="text-xl text-primary-foreground/80 max-w-3xl mx-auto leading-relaxed">
                {ctaBottom.tagline}
              </p>
            </div>

            <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
              <Button
                size="lg"
                className="bg-background text-primary hover:bg-muted font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                asChild
              >
                <a href={ctaBottom.cta_primary.url}>{ctaBottom.cta_primary.text}</a>
              </Button>
              <Button
                variant="outline"
                size="lg"
                className="bg-transparent border-2 border-primary-foreground/30 text-primary-foreground hover:bg-primary-foreground/10 hover:border-primary-foreground/50 hover:text-primary-foreground font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                asChild
              >
                <a href={ctaBottom.cta_secondary.url}>{ctaBottom.cta_secondary.text}</a>
              </Button>
            </div>

            <div className="pt-8 flex flex-col sm:flex-row justify-center items-center gap-6 text-primary-foreground/80">
              {(ctaBottom?.features || []).map((feature: string, index: number) => (
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

      {/* Development Indicator */}
      {process.env.NODE_ENV === 'development' && usedCMS && (
        <div className="fixed bottom-4 right-4 bg-green-500 text-white px-3 py-1 rounded-full text-xs font-medium shadow-lg">
          ✓ Using CMS
        </div>
      )}
    </Layout>
  );
}