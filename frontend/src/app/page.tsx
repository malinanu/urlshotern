'use client';

import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';
import { Button } from '@/components/ui/button';
import { Card, CardContent, CardDescription, CardHeader, CardTitle } from '@/components/ui/card';
import { Badge } from '@/components/ui/badge';

export default function HomePage() {
  const features = [
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

  return (
      <Layout>
        {/* Hero Section - Redesigned with shadcn components */}
        <section className="relative bg-gradient-to-br from-primary-50 via-white to-primary-50/30 py-20 lg:py-28">
          <div className="absolute inset-0 bg-grid-gray-100 opacity-30" />
          <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center space-y-8">
              <div className="space-y-4">
                <Badge variant="secondary" className="text-primary-700 bg-primary-100 hover:bg-primary-200 transition-colors">
                  ✨ New: Advanced Analytics Dashboard
                </Badge>
                <h1 className="text-4xl sm:text-5xl md:text-6xl lg:text-7xl font-bold tracking-tight text-gray-900">
                  Shorten URLs.{' '}
                  <span className="text-primary-600 bg-gradient-to-r from-primary-600 to-primary-700 bg-clip-text text-transparent">
                    Track Everything.
                  </span>
                </h1>
                <p className="text-xl sm:text-2xl text-gray-600 max-w-4xl mx-auto leading-relaxed">
                  Transform your long URLs into short, trackable links with powerful
                  analytics. Perfect for marketing campaigns, social media, and more.
                </p>
              </div>
            </div>

            {/* Trunc Component */}
            <div className="mt-16">
              <URLShortener />
            </div>

            {/* Social Proof */}
            <div className="mt-16 text-center space-y-6">
              <p className="text-sm font-medium text-gray-500 uppercase tracking-wide">
                Trusted by thousands of businesses worldwide
              </p>
              <div className="flex flex-col sm:flex-row justify-center items-center gap-8 sm:gap-12">
                <div className="flex items-center gap-3">
                  <div className="text-3xl font-bold text-gray-800">500K+</div>
                  <div className="text-sm text-gray-500 font-medium">URLs shortened</div>
                </div>
                <div className="hidden sm:block w-px h-8 bg-gray-300" />
                <div className="flex items-center gap-3">
                  <div className="text-3xl font-bold text-gray-800">50M+</div>
                  <div className="text-sm text-gray-500 font-medium">Clicks tracked</div>
                </div>
                <div className="hidden sm:block w-px h-8 bg-gray-300" />
                <div className="flex items-center gap-3">
                  <div className="text-3xl font-bold text-gray-800">99.9%</div>
                  <div className="text-sm text-gray-500 font-medium">Uptime</div>
                </div>
              </div>
            </div>
          </div>
        </section>

        {/* Features Section - Redesigned with shadcn Cards */}
        <section className="py-20 lg:py-28 bg-white">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="text-center space-y-4 mb-16">
              <Badge variant="outline" className="text-primary-600 border-primary-200">
                Features
              </Badge>
              <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-gray-900">
                Why Choose Trunc?
              </h2>
              <p className="text-xl text-gray-600 max-w-3xl mx-auto">
                More than just a URL shortener. Get powerful features that help you
                understand and optimize your links.
              </p>
            </div>

            <div className="grid sm:grid-cols-2 lg:grid-cols-3 gap-8">
              {features.map((feature) => (
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

        {/* CTA Section - Redesigned with shadcn Buttons */}
        <section className="relative py-20 lg:py-28 bg-gradient-to-br from-primary-600 via-primary-700 to-primary-800 overflow-hidden">
          <div className="absolute inset-0 bg-gradient-to-r from-primary-600/20 to-primary-800/20" />
          <div className="absolute top-0 right-0 -mt-40 -mr-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />
          <div className="absolute bottom-0 left-0 -mb-40 -ml-40 w-80 h-80 bg-white/10 rounded-full blur-3xl" />

          <div className="relative max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
            <div className="space-y-8">
              <div className="space-y-4">
                <Badge className="bg-white/20 text-white hover:bg-white/30 border-0">
                  🚀 Get Started Today
                </Badge>
                <h2 className="text-3xl sm:text-4xl lg:text-5xl font-bold text-white">
                  Ready to get started?
                </h2>
                <p className="text-xl text-primary-100 max-w-3xl mx-auto leading-relaxed">
                  Join thousands of businesses using Trunc to track and optimize
                  their links. Start for free today with no credit card required.
                </p>
              </div>

              <div className="flex flex-col sm:flex-row gap-4 justify-center items-center pt-4">
                <Button
                  size="lg"
                  className="bg-white text-primary-700 hover:bg-gray-50 hover:text-primary-800 font-semibold px-8 py-6 text-lg shadow-xl hover:shadow-2xl transition-all duration-300"
                  asChild
                >
                  <a href="/register">Get Started Free</a>
                </Button>
                <Button
                  variant="outline"
                  size="lg"
                  className="border-2 border-white/30 text-white hover:bg-white/10 hover:border-white/50 font-semibold px-8 py-6 text-lg backdrop-blur-sm transition-all duration-300"
                  asChild
                >
                  <a href="/pricing">View Pricing</a>
                </Button>
              </div>

              <div className="pt-8 flex flex-col sm:flex-row justify-center items-center gap-6 text-primary-100">
                <div className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-sm">Free forever plan</span>
                </div>
                <div className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-sm">No credit card required</span>
                </div>
                <div className="flex items-center gap-2">
                  <svg className="w-5 h-5 text-green-400" fill="currentColor" viewBox="0 0 20 20">
                    <path fillRule="evenodd" d="M16.707 5.293a1 1 0 010 1.414l-8 8a1 1 0 01-1.414 0l-4-4a1 1 0 011.414-1.414L8 12.586l7.293-7.293a1 1 0 011.414 0z" clipRule="evenodd" />
                  </svg>
                  <span className="text-sm">Setup in 2 minutes</span>
                </div>
              </div>
            </div>
          </div>
        </section>
      </Layout>
    );
}