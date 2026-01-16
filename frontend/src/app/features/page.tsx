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
    // Fetch CMS content
    const fetchPageData = async () => {
      try {
        const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';
        const response = await fetch(`${apiUrl}/api/v1/pages/features`);
        if (response.ok) {
          const data = await response.json();
          if (data.page?.title) {
            document.title = data.page.title;
          }
          if (data.page?.meta_description) {
            const metaDesc = document.querySelector('meta[name="description"]');
            if (metaDesc) {
              metaDesc.setAttribute('content', data.page.meta_description);
            }
          }
          setPageData(data.page);
        } else {
          setError('Failed to load page content');
        }
      } catch (err) {
        console.error('Failed to load CMS content:', err);
        setError(err instanceof Error ? err.message : 'Unknown error');
      } finally {
        setLoading(false);
      }
    };

    fetchPageData();
  }, []);

  // If CMS content is loaded, use it; otherwise show fallback
  const shouldRenderCMSContent = pageData?.content && !loading;

  return (
    <Layout>
      {/* Display error banner if CMS content failed to load */}
      {error && (
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <p className="text-sm text-yellow-700">
              <span className="font-medium">Note:</span> Unable to load dynamic content. Showing default content. ({error})
            </p>
          </div>
        </div>
      )}

      {loading ? (
        <div className="py-20">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="animate-pulse space-y-8">
              <div className="h-16 bg-gray-200 rounded w-3/4 mx-auto"></div>
              <div className="h-8 bg-gray-200 rounded w-2/3 mx-auto"></div>
            </div>
          </div>
        </div>
      ) : shouldRenderCMSContent ? (
        /* Render CMS Content as HTML */
        <div
          className="cms-content"
          dangerouslySetInnerHTML={{ __html: pageData?.content || '' }}
        />
      ) : (
        /* Fallback content when CMS is unavailable */
        <>
          <section className="bg-gradient-to-br from-primary-50 to-white py-20">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="text-center">
                <h1 className="text-4xl md:text-6xl font-bold text-black mb-6">
                  Powerful Features for{' '}
                  <span className="text-primary-600">Smart URL Management</span>
                </h1>
                <p className="text-xl text-black mb-12 max-w-3xl mx-auto">
                  Everything you need to create, track, and optimize your short links.
                </p>
              </div>
            </div>
          </section>

          <section className="py-20 bg-white">
            <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
              <div className="text-center mb-16">
                <h2 className="text-3xl font-bold text-black mb-4">Core Features</h2>
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
                        <h3 className="text-xl font-semibold text-black mb-3">{feature.name}</h3>
                        <p className="text-gray-600 mb-4 leading-relaxed">{feature.description}</p>
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
        </>
      )}
    </Layout>
  );
}
