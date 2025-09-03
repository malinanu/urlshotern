import Layout from '@/components/layout/Layout';
import URLShortener from '@/components/URLShortener';
import { LinkIcon, ChartBarIcon, ShieldCheckIcon } from '@heroicons/react/24/outline';

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
      {/* Hero Section */}
      <section className="bg-gradient-to-br from-primary-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-gray-900 mb-6">
              Shorten URLs.{' '}
              <span className="text-primary-600">Track Everything.</span>
            </h1>
            <p className="text-xl text-gray-600 mb-12 max-w-3xl mx-auto">
              Transform your long URLs into short, trackable links with powerful
              analytics. Perfect for marketing campaigns, social media, and more.
            </p>
          </div>

          {/* URL Shortener Component */}
          <URLShortener />

          {/* Trust indicators */}
          <div className="mt-12 text-center">
            <p className="text-sm text-gray-500">
              Trusted by thousands of businesses worldwide
            </p>
            <div className="mt-4 flex justify-center items-center space-x-8 opacity-60">
              <div className="text-2xl font-bold text-gray-400">500K+</div>
              <div className="text-sm text-gray-400">URLs shortened</div>
              <div className="w-px h-6 bg-gray-300"></div>
              <div className="text-2xl font-bold text-gray-400">50M+</div>
              <div className="text-sm text-gray-400">Clicks tracked</div>
            </div>
          </div>
        </div>
      </section>

      {/* Features Section */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h2 className="text-3xl font-bold text-gray-900 mb-4">
              Why Choose URLShorter?
            </h2>
            <p className="text-xl text-gray-600 mb-16 max-w-2xl mx-auto">
              More than just a URL shortener. Get powerful features that help you
              understand and optimize your links.
            </p>
          </div>

          <div className="grid md:grid-cols-3 gap-12">
            {features.map((feature) => (
              <div key={feature.name} className="text-center">
                <div className="flex justify-center mb-4">
                  <div className="p-3 bg-primary-100 rounded-full">
                    <feature.icon className="h-8 w-8 text-primary-600" />
                  </div>
                </div>
                <h3 className="text-xl font-semibold text-gray-900 mb-3">
                  {feature.name}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {feature.description}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* CTA Section */}
      <section className="py-20 bg-primary-600">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-white mb-4">
            Ready to get started?
          </h2>
          <p className="text-xl text-primary-100 mb-8 max-w-2xl mx-auto">
            Join thousands of businesses using URLShorter to track and optimize
            their links. Start for free today.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href="/register"
              className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-50 transition-colors"
            >
              Get Started Free
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