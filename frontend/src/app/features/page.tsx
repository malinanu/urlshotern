'use client';

import Layout from '@/components/layout/Layout';
import {
  LinkIcon,
  ChartBarIcon,
  ShieldCheckIcon,
  CursorArrowRaysIcon,
  GlobeAltIcon,
  ClockIcon,
  DevicePhoneMobileIcon,
  UserGroupIcon,
  CheckCircleIcon
} from '@heroicons/react/24/outline';
import { useCMSPage, getContentBlock } from '@/hooks/useCMSPage';

// Icon mapping for CMS string values
const iconMap: Record<string, any> = {
  LinkIcon,
  ChartBarIcon,
  ShieldCheckIcon,
  CursorArrowRaysIcon,
  GlobeAltIcon,
  ClockIcon,
  DevicePhoneMobileIcon,
  UserGroupIcon,
  CheckCircleIcon,
};

export default function FeaturesPage() {
  // Fetch page content from CMS with fallback
  const { content: page, loading, usedCMS } = useCMSPage('features', {
    slug: 'features',
    title: 'Features',
    content: 'Powerful Features for Smart URL Management',
  });

  const fallbackMainFeatures = [
    {
      name: 'Lightning Fast Shortening',
      description: 'Generate short URLs instantly with our optimized infrastructure powered by Redis caching and efficient algorithms.',
      icon: LinkIcon,
      details: [
        'Sub-second response times',
        'Global CDN distribution',
        'Auto-scaling infrastructure',
        'Bulk URL processing'
      ]
    },
    {
      name: 'Advanced Analytics',
      description: 'Track every click with detailed insights including geographic data, device information, and referrer analysis.',
      icon: ChartBarIcon,
      details: [
        'Real-time click tracking',
        'Geographic click mapping',
        'Device and browser analytics',
        'Referrer source tracking'
      ]
    },
    {
      name: 'Enterprise Security',
      description: 'Bank-level security with SSL encryption, spam protection, and enterprise-grade infrastructure.',
      icon: ShieldCheckIcon,
      details: [
        'SSL/TLS encryption',
        'Malware URL scanning',
        'Spam link detection',
        '99.9% uptime SLA'
      ]
    },
    {
      name: 'Smart Click Insights',
      description: 'Understand your audience better with detailed click analytics and user behavior patterns.',
      icon: CursorArrowRaysIcon,
      details: [
        'Click heatmaps',
        'Time-based analytics',
        'Conversion tracking',
        'A/B testing support'
      ]
    },
    {
      name: 'Global Performance',
      description: 'Worldwide infrastructure ensures your short links work fast everywhere, with multi-region redundancy.',
      icon: GlobeAltIcon,
      details: [
        'Multi-region deployment',
        'Edge caching worldwide',
        'Automatic failover',
        'Load balancing'
      ]
    },
    {
      name: 'Real-time Monitoring',
      description: 'Monitor your links in real-time with instant notifications and comprehensive health checks.',
      icon: ClockIcon,
      details: [
        'Live click monitoring',
        'Alert notifications',
        'Health check endpoints',
        'Performance metrics'
      ]
    }
  ];

  const fallbackAdditionalFeatures = [
    { name: 'Custom Short Codes', icon: CheckCircleIcon },
    { name: 'QR Code Generation', icon: CheckCircleIcon },
    { name: 'Link Expiration', icon: CheckCircleIcon },
    { name: 'Password Protection', icon: CheckCircleIcon },
    { name: 'Team Collaboration', icon: CheckCircleIcon },
    { name: 'API Access', icon: CheckCircleIcon },
    { name: 'Webhook Integration', icon: CheckCircleIcon },
    { name: 'White-label Solution', icon: CheckCircleIcon },
    { name: 'Mobile Apps', icon: CheckCircleIcon },
    { name: '24/7 Support', icon: CheckCircleIcon },
    { name: 'Data Export', icon: CheckCircleIcon },
    { name: 'SSO Integration', icon: CheckCircleIcon }
  ];

  // Extract content blocks
  const hero = getContentBlock(page, 'hero', {
    title: 'Powerful Features for Smart URL Management',
    title_highlight: 'Smart URL Management',
    tagline: 'Everything you need to create, track, and optimize your short links. From basic URL shortening to advanced analytics and enterprise features.'
  });

  const mainFeaturesBlock = getContentBlock(page, 'main_features', {
    badge: 'Core Features',
    heading: 'Core Features',
    tagline: 'Built for performance, designed for insights, trusted by thousands.',
    items: fallbackMainFeatures
  });

  const additionalFeaturesBlock = getContentBlock(page, 'additional_features', {
    badge: 'Additional Features',
    heading: 'Additional Features',
    tagline: 'Even more tools and capabilities to enhance your URL management experience.',
    items: fallbackAdditionalFeatures
  });

  const ctaBottom = getContentBlock(page, 'cta_bottom', {
    heading: 'Ready to Experience These Features?',
    tagline: 'Start using Trunc today and unlock the full potential of your links.',
    cta_primary: { text: 'Try It Now', url: '/' },
    cta_secondary: { text: 'View Pricing', url: '/pricing' }
  });

  // Map CMS data to components
  const mainFeatures = (mainFeaturesBlock.items || fallbackMainFeatures).map((feature: any) => ({
    ...feature,
    // Handle icon mapping if it's a string from CMS
    icon: feature.icon && typeof feature.icon === 'string' ? (iconMap[feature.icon] || LinkIcon) : feature.icon || LinkIcon
  }));

  const additionalFeatures = (additionalFeaturesBlock.items || fallbackAdditionalFeatures).map((feature: any) => ({
    ...feature,
    icon: feature.icon && typeof feature.icon === 'string' ? (iconMap[feature.icon] || CheckCircleIcon) : feature.icon || CheckCircleIcon
  }));

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-[400px]">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      {/* Hero Section - Using CMS Content */}
      <section className="bg-gradient-to-br from-primary/10 to-background py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-foreground mb-6">
              {hero.title.replace(hero.title_highlight || '', '')}
              {hero.title_highlight && (
                <span className="text-primary">{hero.title_highlight}</span>
              )}
            </h1>
            <p className="text-xl text-foreground mb-12 max-w-3xl mx-auto">
              {hero.tagline}
            </p>
          </div>
        </div>
      </section>

      {/* Main Features Section */}
      <section className="py-20 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-foreground mb-4">
              {mainFeaturesBlock.heading}
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              {mainFeaturesBlock.tagline}
            </p>
          </div>

          <div className="grid lg:grid-cols-2 gap-12">
            {mainFeatures.map((feature: any) => (
              <div key={feature.name} className="bg-muted/30 rounded-xl p-8 border border-border">
                <div className="flex items-start space-x-4">
                  <div className="flex-shrink-0">
                    <div className="p-3 bg-primary/10 rounded-lg">
                      <feature.icon className="h-8 w-8 text-primary" />
                    </div>
                  </div>
                  <div className="flex-1">
                    <h3 className="text-xl font-semibold text-foreground mb-3">
                      {feature.name}
                    </h3>
                    <p className="text-muted-foreground mb-4 leading-relaxed">
                      {feature.description}
                    </p>
                    <ul className="space-y-2">
                      {feature.details && feature.details.map((detail: string) => (
                        <li key={detail} className="flex items-center space-x-2">
                          <CheckCircleIcon className="h-5 w-5 text-primary flex-shrink-0" />
                          <span className="text-foreground text-sm">{detail}</span>
                        </li>
                      ))}
                    </ul>
                  </div>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Additional Features Section */}
      <section className="py-20 bg-muted/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-foreground mb-4">
              {additionalFeaturesBlock.heading}
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              {additionalFeaturesBlock.tagline}
            </p>
          </div>

          <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
            {additionalFeatures.map((feature: any) => (
              <div key={feature.name} className="bg-card p-6 rounded-lg shadow-sm border border-border">
                <div className="flex items-center space-x-3">
                  <feature.icon className="h-6 w-6 text-primary flex-shrink-0" />
                  <span className="text-foreground font-medium">{feature.name}</span>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-primary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-primary-foreground mb-4">
            {ctaBottom.heading}
          </h2>
          <p className="text-xl text-primary-foreground/80 mb-8 max-w-2xl mx-auto">
            {ctaBottom.tagline}
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href={ctaBottom.cta_primary?.url || '/'}
              className="bg-background text-primary px-8 py-3 rounded-lg font-semibold hover:bg-muted transition-colors"
            >
              {ctaBottom.cta_primary?.text || 'Try It Now'}
            </a>
            <a
              href={ctaBottom.cta_secondary?.url || '/pricing'}
              className="border border-primary-foreground/30 text-primary-foreground px-8 py-3 rounded-lg font-semibold hover:bg-primary/80 transition-colors"
            >
              {ctaBottom.cta_secondary?.text || 'View Pricing'}
            </a>
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