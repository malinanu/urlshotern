'use client';

import { useEffect, useState } from 'react';
import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

interface CMSPage {
  title: string;
  content: string;
  meta_description: string;
  meta_keywords: string;
  slug: string;
}

interface ParsedContent {
  heroTitle: string;
  heroSubtitle?: string;
  heroDescription: string;
  featuresTitle?: string;
  featuresDescription?: string;
}

export default function HomePage() {
  const [cmsContent, setCmsContent] = useState<CMSPage | null>(null);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState<string | null>(null);

  useEffect(() => {
    // Fetch CMS content for homepage
    const apiUrl = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

    fetch(`${apiUrl}/api/v1/pages/home`)
      .then(res => {
        if (!res.ok) {
          throw new Error(`HTTP ${res.status}: ${res.statusText}`);
        }
        return res.json();
      })
      .then(data => {
        if (data.page) {
          setCmsContent(data.page);

          // Update document metadata dynamically
          if (data.page.title) {
            document.title = data.page.title;
          }

          // Update meta description
          if (data.page.meta_description) {
            const metaDesc = document.querySelector('meta[name="description"]');
            if (metaDesc) {
              metaDesc.setAttribute('content', data.page.meta_description);
            } else {
              const meta = document.createElement('meta');
              meta.name = 'description';
              meta.content = data.page.meta_description;
              document.head.appendChild(meta);
            }
          }

          // Update meta keywords
          if (data.page.meta_keywords) {
            const metaKeywords = document.querySelector('meta[name="keywords"]');
            if (metaKeywords) {
              metaKeywords.setAttribute('content', data.page.meta_keywords);
            } else {
              const meta = document.createElement('meta');
              meta.name = 'keywords';
              meta.content = data.page.meta_keywords;
              document.head.appendChild(meta);
            }
          }
        }
        setLoading(false);
      })
      .catch(err => {
        console.error('Failed to load CMS content:', err);
        setError(err.message);
        setLoading(false);
      });
  }, []);

  // If CMS content is loaded, use it directly as HTML
  const shouldRenderCMSContent = cmsContent?.content && !loading;

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
            dangerouslySetInnerHTML={{ __html: cmsContent?.content || '' }}
          />
        ) : (
          /* Fallback content when CMS is unavailable */
          <>
            {/* Hero Section */}
            <section className="relative bg-gradient-to-br from-primary-50 via-white to-primary-50/30 py-20 lg:py-28">
              <div className="absolute inset-0 bg-grid-gray-100 opacity-30" />
              <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
                <div className="text-center space-y-8">
                  <div className="space-y-4">
                    <Badge variant="secondary" className="text-primary-700 bg-primary-100 hover:bg-primary-200 transition-colors">
                      ✨ New: Advanced Analytics Dashboard
                    </Badge>
                    <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-gray-900">
                      Shorten URLs. <br />
                      <span className="text-primary-600">Track Everything.</span>
                    </h1>
                    <p className="text-xl sm:text-2xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
                      Transform your long URLs into short, trackable links with powerful analytics.
                    </p>
                  </div>
                </div>

                <div className="mt-16">
                  <URLShortener />
                </div>
              </div>
            </section>
          </>
        )}
      </Layout>
    );
}
