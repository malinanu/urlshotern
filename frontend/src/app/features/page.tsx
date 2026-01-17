'use client';

import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';
import { Button } from '@/components/ui/button';
import Layout from '@/components/layout/Layout';
import {
  LinkIcon,
  ChartBarIcon,
  ShieldCheckIcon,
  CursorArrowRaysIcon,
  GlobeAltIcon,
  ClockIcon,
  CheckCircleIcon
} from '@heroicons/react/24/outline';

// Core features with descriptions
const coreFeatures = [
  {
    name: 'Lightning Fast',
    description: 'Generate short URLs in milliseconds with our high-performance infrastructure powered by Redis caching and optimized database queries.',
    icon: LinkIcon
  },
  {
    name: 'Advanced Analytics',
    description: 'Track clicks, geographic data, referrers, and device information. Get comprehensive insights into your link performance with detailed reports.',
    icon: ChartBarIcon
  },
  {
    name: 'Enterprise Security',
    description: 'HTTPS encryption, rate limiting, DDoS protection, and secure API authentication. Your data and links are protected with enterprise-grade security.',
    icon: ShieldCheckIcon
  },
  {
    name: 'Click Insights',
    description: 'Real-time click tracking with detailed visitor analytics. Understand your audience with geographic, device, and browser insights.',
    icon: CursorArrowRaysIcon
  },
  {
    name: 'Global Performance',
    description: 'Distributed infrastructure ensures fast redirects worldwide. Our globally distributed network provides low-latency access for users everywhere.',
    icon: GlobeAltIcon
  },
  {
    name: '24/7 Monitoring',
    description: 'Continuous health checks and automated failover ensure maximum uptime. Our monitoring systems detect and resolve issues before they impact you.',
    icon: ClockIcon
  }
];

const additionalFeatures = [
  'Custom Short Codes',
  'QR Code Generation',
  'Link Expiration',
  'Password Protection',
  'Team Collaboration',
  'API Access',
  'Webhook Integration',
  'White-label Solution',
  'Mobile Apps',
  '24/7 Support',
  'Data Export',
  'SSO Integration'
];

export default function FeaturesPage() {
  return (
    <Layout>
      {/* Hero Section - Matching Home Page Design */}
      <section className="relative bg-gradient-to-br from-primary-50 via-white to-primary-50/30 py-20 lg:py-28">
        {/* Grid pattern overlay like home page */}
        <div className="absolute inset-0 bg-grid-gray-100 opacity-30" />

        <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center space-y-6">
            <Badge variant="secondary" className="text-primary-700 bg-primary-100 hover:bg-primary-200 transition-colors">
              Features
            </Badge>

            <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold text-gray-900 tracking-tight">
              Powerful Features for{' '}
              <span className="text-primary-600">Smart URL Management</span>
            </h1>

            <p className="text-xl text-gray-600 max-w-3xl mx-auto leading-relaxed">
              Everything you need to create, track, and optimize your short links. From lightning-fast shortening to comprehensive analytics.
            </p>
          </div>
        </div>
      </section>

      {/* Core Features Section - Using Card Components Like Home Page */}
      <section className="py-20 lg:py-28 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 space-y-4">
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Core Features
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              Built for performance, designed for insights, trusted by thousands.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
            {coreFeatures.map((feature) => (
              <Card
                key={feature.name}
                className="relative group hover:shadow-xl transition-all duration-300 border-0 shadow-md bg-gradient-to-br from-white to-gray-50/50"
              >
                <CardHeader className="text-center space-y-4">
                  <div className="mx-auto w-16 h-16 bg-gradient-to-br from-primary-100 to-primary-200 rounded-2xl flex items-center justify-center group-hover:scale-110 transition-transform duration-300">
                    <feature.icon className="h-8 w-8 text-primary-600" />
                  </div>
                  <CardTitle className="text-xl font-bold text-gray-900 group-hover:text-primary-600 transition-colors">
                    {feature.name}
                  </CardTitle>
                </CardHeader>
                <CardContent>
                  <CardDescription className="text-gray-600 leading-relaxed text-center">
                    {feature.description}
                  </CardDescription>
                </CardContent>
              </Card>
            ))}
          </div>
        </div>
      </section>

      {/* Additional Features Section */}
      <section className="py-20 lg:py-28 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16 space-y-4">
            <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
              Additional Features
            </h2>
            <p className="text-xl text-gray-600">
              Even more tools to enhance your URL management experience.
            </p>
          </div>

          <div className="grid sm:grid-cols-2 md:grid-cols-3 lg:grid-cols-4 gap-6">
            {additionalFeatures.map((feature) => (
              <div key={feature} className="flex items-center space-x-3 bg-white rounded-lg p-4 shadow-sm hover:shadow-md transition-shadow">
                <CheckCircleIcon className="h-6 w-6 text-primary-600 flex-shrink-0" />
                <span className="text-gray-900 font-medium">{feature}</span>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section - Matching Home Page Style */}
      <section className="relative py-20 lg:py-28 bg-gradient-to-br from-primary-600 via-primary-700 to-primary-800 overflow-hidden">
        {/* Decorative elements like home page */}
        <div className="absolute inset-0 bg-gradient-to-r from-primary-600/20 to-primary-800/20" />
        <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
        <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

        <div className="relative max-w-4xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <Badge className="mb-6 bg-white/20 text-white hover:bg-white/30 transition-colors border-white/30">
            🚀 Get Started Today
          </Badge>

          <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white mb-6">
            Ready to Experience These Features?
          </h2>

          <p className="text-xl text-primary-100 mb-10 max-w-2xl mx-auto">
            Start using Trunc today and unlock the full potential of your links.
          </p>

          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <Button
              size="lg"
              className="bg-white text-primary-700 hover:bg-gray-50 hover:text-primary-800 font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
            >
              Try It Now
            </Button>
            <Button
              size="lg"
              variant="outline"
              className="border-2 border-white text-white hover:bg-white/10 font-semibold px-8 py-6 text-lg transition-all duration-300"
            >
              View Pricing
            </Button>
          </div>
        </div>
      </section>
    </Layout>
  );
}
