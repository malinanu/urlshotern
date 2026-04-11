'use client';

import Layout from '@/components/layout/Layout';
import { Button } from '@/components/ui/button';
import { Badge } from '@/components/ui/badge';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import {
  LinkIcon,
  ShieldCheckIcon,
  GlobeAltIcon,
  UsersIcon,
  ChartBarIcon,
  ClockIcon,
  ArrowRightIcon,
  BuildingOfficeIcon,
  HeartIcon,
  StarIcon
} from '@heroicons/react/24/outline';
import { useCMSPage, getContentBlock } from '@/hooks/useCMSPage';

// Fallback data (used if CMS fetch fails)
const fallbackStats = [
  { label: 'Links Shortened', value: '10M+', icon: 'LinkIcon', color: 'text-primary' },
  { label: 'Uptime', value: '99.9%', icon: 'ShieldCheckIcon', color: 'text-green-500' },
  { label: 'Countries', value: '150+', icon: 'GlobeAltIcon', color: 'text-primary' },
  { label: 'Active Users', value: '500K+', icon: 'UsersIcon', color: 'text-primary' }
];

const fallbackValues = [
  {
    title: 'Innovation First',
    description: 'We constantly push the boundaries of what\'s possible in URL management and analytics.',
    icon: 'LinkIcon'
  },
  {
    title: 'Security & Privacy',
    description: 'Your data and links are protected with enterprise-grade security and privacy measures.',
    icon: 'ShieldCheckIcon'
  },
  {
    title: 'Performance',
    description: 'Lightning-fast redirects and real-time analytics ensure optimal user experience.',
    icon: 'ChartBarIcon'
  },
  {
    title: 'Reliability',
    description: '99.9% uptime guarantee ensures your links work when your audience needs them most.',
    icon: 'ClockIcon'
  }
];

const fallbackAchievements = [
  { year: '2024', title: 'Reached 10M Links Shortened', description: 'Milestone achievement in our growth journey' },
  { year: '2023', title: 'SOC 2 Type II Compliance', description: 'Enterprise-grade security certification' },
  { year: '2023', title: '99.9% Uptime Achievement', description: 'Consistent reliability and performance' },
  { year: '2022', title: 'Global Expansion', description: 'Serving users in over 150 countries worldwide' }
];

const iconMap: Record<string, any> = {
  LinkIcon,
  ShieldCheckIcon,
  GlobeAltIcon,
  UsersIcon,
  ChartBarIcon,
  ClockIcon,
  BuildingOfficeIcon,
};

export default function AboutPage() {
  // Fetch page content from CMS with fallback
  const { content: page, loading, usedCMS } = useCMSPage('about', {
    slug: 'about',
    title: 'We Are Trunc',
    content: 'Pioneering the next generation of URL management with cutting-edge technology.',
  });

  // Extract content blocks from CMS or use fallback
  const hero = getContentBlock(page, 'hero', {
    badge: 'About Trunc',
    badge_icon: 'BuildingOfficeIcon',
    title: 'We Are Trunc',
    title_highlight: 'Trunc',
    tagline: 'Pioneering the next generation of URL management with cutting-edge technology, unparalleled performance, and a vision that connects the digital world.',
  });

  const statsBlock = getContentBlock(page, 'stats', fallbackStats);
  const mission = getContentBlock(page, 'mission', {
    badge: 'Our Mission',
    heading: 'Connecting Every Digital Experience',
    description: 'We envision a world where every URL is intelligent, every click tells a story, and every connection drives meaningful outcomes. Trunc isn\'t just a tool—it\'s the bridge between intention and impact in the digital ecosystem.',
    cta_primary: { text: 'Get Started', url: '/register' },
    cta_secondary: { text: 'Explore Trunc Features', url: '/features' },
  });

  const valuesBlock = getContentBlock(page, 'values', {
    badge: 'Our Values',
    heading: 'What Drives Us Forward',
    tagline: 'Our core values shape every decision we make and define how we build products that truly serve our customers.',
    items: fallbackValues,
  });

  const builtWith = getContentBlock(page, 'built_with', {
    badge: 'How We Build',
    heading: 'Built for Speed, Built to Last',
    description: 'Trunc is built with a focus on reliability, performance, and simplicity. Every feature is designed to solve a real problem — no bloat, no noise. Just fast short links and the analytics you actually need.',
    principles: [
      { title: '50ms Redirects', description: 'Links resolve in under 50 milliseconds globally, so your audience never notices the redirect.' },
      { title: '99.9% Uptime', description: 'Redundant infrastructure means your links keep working even when traffic spikes unexpectedly.' },
      { title: 'Privacy First', description: 'We collect only the data needed to power your analytics. No selling data. No third-party tracking.' },
      { title: 'Open API', description: 'Everything you can do in the dashboard you can automate via the REST API. No features locked behind the UI.' },
    ],
  });

  const achievementsBlock = getContentBlock(page, 'achievements', {
    badge: 'Achievements',
    heading: 'Key Milestones',
    items: fallbackAchievements,
  });

  const ctaBottom = getContentBlock(page, 'cta_bottom', {
    badge: '🚀 Ready to Get Started?',
    heading: 'Join thousands of businesses using Trunc',
    tagline: 'Start shortening URLs and tracking analytics with our powerful platform. No credit card required to get started.',
    cta_primary: { text: 'Start Free Trial', url: '/register' },
    cta_secondary: { text: 'Contact Sales', url: '/contact' },
    features: ['Free forever plan', 'No credit card required', 'Setup in 2 minutes'],
  });

  // Map icon names to actual components
  const stats = (statsBlock || fallbackStats).map((stat: any) => ({
    ...stat,
    icon: stat.icon && typeof stat.icon === 'string' ? (iconMap[stat.icon] || LinkIcon) : stat.icon || LinkIcon,
  }));

  const values = (valuesBlock.items || fallbackValues).map((value: any) => ({
    ...value,
    icon: value.icon && typeof value.icon === 'string' ? (iconMap[value.icon] || LinkIcon) : value.icon || LinkIcon,
  }));

  const achievements = achievementsBlock.items || fallbackAchievements;

  return (
    <Layout>
      {/* Hero Section */}
      <section className="relative bg-gradient-to-br from-primary/5 via-background to-primary/5 py-20 lg:py-28">
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30 dark:opacity-5" />
        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-8">
            <div className="space-y-4">
              <Badge variant="secondary" className="text-primary bg-primary/10 hover:bg-primary/20 transition-colors">
                {hero.badge_icon && iconMap[hero.badge_icon] && (
                  (() => {
                    const IconComponent = iconMap[hero.badge_icon];
                    return <IconComponent className="h-4 w-4 mr-2" />;
                  })()
                )}
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

          {/* Stats Grid */}
          <div className="mt-16 grid grid-cols-2 lg:grid-cols-4 gap-8">
            {(stats || []).map((stat, index) => (
              <div key={index} className="text-center space-y-3">
                <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary/10 to-primary/20 rounded-2xl flex items-center justify-center">
                  <stat.icon className={`h-8 w-8 ${stat.color}`} />
                </div>
                <div className="space-y-1">
                  <div className="text-3xl font-bold text-foreground">{stat.value}</div>
                  <div className="text-sm font-medium text-muted-foreground">{stat.label}</div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Mission Section */}
      <section className="py-20 lg:py-28 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-2 gap-16 items-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <Badge variant="outline" className="text-primary border-primary/20">
                  {mission.badge}
                </Badge>
                <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-foreground">
                  {mission.heading}
                </h2>
                <p className="text-xl text-muted-foreground leading-relaxed">
                  {mission.description}
                </p>
              </div>

              <div className="flex flex-col sm:flex-row gap-4">
                <Button size="lg" className="flex items-center gap-2 bg-primary hover:bg-primary/90 text-primary-foreground" asChild>
                  <a href={mission.cta_primary.url}>
                    {mission.cta_primary.text}
                    <ArrowRightIcon className="h-5 w-5" />
                  </a>
                </Button>
                <Button size="lg" variant="outline" className="border-primary/20 text-primary hover:bg-primary/10" asChild>
                  <a href={mission.cta_secondary.url}>{mission.cta_secondary.text}</a>
                </Button>
              </div>
            </div>

            <div className="lg:order-first">
              <Card className="bg-gradient-to-br from-primary/5 to-background border-primary/10">
                <CardContent className="p-8">
                  <div className="grid grid-cols-2 gap-4 mb-6">
                    <div className="space-y-2">
                      <div className="h-3 bg-primary/20 rounded"></div>
                      <div className="h-2 bg-primary/10 rounded w-3/4"></div>
                      <div className="h-2 bg-primary/10 rounded w-1/2"></div>
                    </div>
                    <div className="space-y-2">
                      <div className="h-3 bg-primary/20 rounded opacity-75"></div>
                      <div className="h-2 bg-primary/10 rounded w-2/3"></div>
                      <div className="h-2 bg-primary/10 rounded w-3/4"></div>
                    </div>
                  </div>
                  <div className="text-center text-sm text-muted-foreground font-medium">
                    Intelligent Analytics Dashboard
                  </div>
                </CardContent>
              </Card>
            </div>
          </div>
        </div>
      </section>

      {/* Values Section */}
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              {valuesBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              {valuesBlock.heading}
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto">
              {valuesBlock.tagline}
            </p>
          </div>

          <div className="grid md:grid-cols-2 gap-8">
            {(values || []).map((value, index) => (
              <Card
                key={index}
                className="relative group hover:shadow-xl transition-all duration-300 border-0 shadow-md bg-white"
              >
                <CardHeader className="space-y-4">
                  <div className="flex items-center gap-4">
                    <div className="w-12 h-12 bg-gradient-to-br from-primary-100 to-primary-200 rounded-xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                      <value.icon className="h-6 w-6 text-primary-600" />
                    </div>
                    <CardTitle className="text-xl font-bold text-gray-900 group-hover:text-primary-600 transition-colors">
                      {value.title}
                    </CardTitle>
                  </div>
                  <CardDescription className="text-gray-600 leading-relaxed">
                    {value.description}
                  </CardDescription>
                </CardHeader>
                <CardContent>
                  <div className="flex items-center gap-2 text-primary-600">
                    <StarIcon className="h-4 w-4" />
                    <span className="text-sm font-medium">Core Principle</span>
                  </div>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Built With Purpose Section */}
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              {builtWith.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              {builtWith.heading}
            </h2>
            <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
              {builtWith.description}
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-4 gap-8">
            {(builtWith.principles || []).map((principle: any, index: number) => (
              <Card key={index} className="bg-gradient-to-br from-primary/5 to-background border-primary/10 text-center hover:shadow-lg transition-shadow duration-300">
                <CardHeader className="space-y-3">
                  <div className="mx-auto w-12 h-12 bg-gradient-to-br from-primary/10 to-primary/20 rounded-xl flex items-center justify-center">
                    <ShieldCheckIcon className="h-6 w-6 text-primary" />
                  </div>
                  <CardTitle className="text-lg font-bold text-gray-900">{principle.title}</CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-gray-600 leading-relaxed">{principle.description}</CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Achievements Section */}
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-4 mb-16">
            <Badge variant="outline" className="text-primary-600 border-primary-200">
              {achievementsBlock.badge}
            </Badge>
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              {achievementsBlock.heading}
            </h2>
          </div>

          <div className="grid gap-6">
            {(achievements || []).map((achievement, index) => (
              <Card key={index} className="border-l-4 border-l-primary-600 bg-white hover:shadow-lg transition-shadow duration-300">
                <CardHeader>
                  <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4">
                    <div className="space-y-2">
                      <CardTitle className="text-xl font-bold text-gray-900">{achievement.title}</CardTitle>
                      <CardDescription className="text-gray-600">{achievement.description}</CardDescription>
                    </div>
                    <Badge className="bg-primary-100 text-primary-700 hover:bg-primary-200 w-fit">
                      {achievement.year}
                    </Badge>
                  </div>
                </CardHeader>
              </Card>
            ))}
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
                <a href={ctaBottom.cta_primary.url}>
                  {ctaBottom.cta_primary.text}
                  <ArrowRightIcon className="h-5 w-5 ml-2" />
                </a>
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