'use client';

import { useEffect, useState } from 'react';
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

interface Page {
  id: number;
  slug: string;
  title: string;
  content: string;
  meta_description: string;
  is_published: boolean;
  created_at: string;
  updated_at: string;
}

export default function FeaturesPage() {
  const [pageData, setPageData] = useState<Page | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);
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

  useEffect(() => {
    const fetchPageData = async () => {
      try {
        const response = await fetch('/api/v1/pages/features');
        if (!response.ok) {
          throw new Error('Failed to fetch page data');
        }
        const data = await response.json();
        setPageData(data);
      } catch (err) {
        setError(err instanceof Error ? err.message : 'Failed to load page');
      } finally {
        setLoading(false);
      }
    };

    fetchPageData();
  }, []);

  if (loading) {
    return (
      <Layout>
        <div className="flex justify-center items-center min-h-[400px]">
          <div className="animate-spin rounded-full h-32 w-32 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  if (error || !pageData) {
    return (
      <Layout>
        {/* Fallback content when CMS is unavailable */}
        <section className="bg-gradient-to-br from-primary-50 to-white py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center">
              <h1 className="text-4xl md:text-6xl font-bold text-black mb-6">
                Powerful Features for{' '}
                <span className="text-primary-600">Smart URL Management</span>
              </h1>
              <p className="text-xl text-black mb-12 max-w-3xl mx-auto">
                Everything you need to create, track, and optimize your short links. 
                From basic URL shortening to advanced analytics and enterprise features.
              </p>
            </div>
          </div>
        </section>

        <section className="py-20 bg-white">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold text-black mb-4">
                Core Features
              </h2>
              <p className="text-xl text-gray-600 max-w-2xl mx-auto">
                Built for performance, designed for insights, trusted by thousands.
              </p>
            </div>

            <div className="grid lg:grid-cols-2 gap-12">
              {fallbackMainFeatures.map((feature) => (
                <div key={feature.name} className="bg-gray-50 rounded-xl p-8">
                  <div className="flex items-start space-x-4">
                    <div className="flex-shrink-0">
                      <div className="p-3 bg-primary-100 rounded-lg">
                        <feature.icon className="h-8 w-8 text-primary-600" />
                      </div>
                    </div>
                    <div className="flex-1">
                      <h3 className="text-xl font-semibold text-black mb-3">
                        {feature.name}
                      </h3>
                      <p className="text-gray-600 mb-4 leading-relaxed">
                        {feature.description}
                      </p>
                      <ul className="space-y-2">
                        {feature.details.map((detail) => (
                          <li key={detail} className="flex items-center space-x-2">
                            <CheckCircleIcon className="h-5 w-5 text-primary-600 flex-shrink-0" />
                            <span className="text-black text-sm">{detail}</span>
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

        <section className="py-20 bg-gray-50">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center mb-16">
              <h2 className="text-3xl font-bold text-black mb-4">
                Additional Features
              </h2>
              <p className="text-xl text-gray-600 max-w-2xl mx-auto">
                Even more tools and capabilities to enhance your URL management experience.
              </p>
            </div>

            <div className="grid md:grid-cols-2 lg:grid-cols-3 xl:grid-cols-4 gap-6">
              {fallbackAdditionalFeatures.map((feature) => (
                <div key={feature.name} className="bg-white p-6 rounded-lg shadow-sm border border-gray-200">
                  <div className="flex items-center space-x-3">
                    <feature.icon className="h-6 w-6 text-primary-600 flex-shrink-0" />
                    <span className="text-black font-medium">{feature.name}</span>
                  </div>
                </div>
              ))}
            </div>
          </div>
        </section>

        <section className="py-20 bg-primary-600">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <h2 className="text-3xl font-bold text-white mb-4">
              Ready to Experience These Features?
            </h2>
            <p className="text-xl text-primary-100 mb-8 max-w-2xl mx-auto">
              Start using Trunc today and unlock the full potential of your links.
            </p>
            <div className="flex flex-col sm:flex-row gap-4 justify-center">
              <a
                href="/"
                className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-50 transition-colors"
              >
                Try It Now
              </a>
              <a
                href="/pricing"
                className="border border-primary-400 text-white px-8 py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors"
              >
                View Pricing
              </a>
            </div>
          </div>
        </section>
      </Layout>
    );
  }

  return (
    <Layout>
      <div dangerouslySetInnerHTML={{ __html: pageData.content }} />
    </Layout>
  );
}