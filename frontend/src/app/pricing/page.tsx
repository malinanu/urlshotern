'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { CheckIcon, XMarkIcon, StarIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import { redirectToCheckout, getProducts, type PolarProduct } from '@/lib/polar';

export default function PricingPage() {
  const { isAuthenticated, getAccessToken } = useAuth();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
  const [polarProducts, setPolarProducts] = useState<PolarProduct[]>([]);
  const [productsLoading, setProductsLoading] = useState(true);
  const [productsFetchError, setProductsFetchError] = useState<{
    type: 'network' | 'api' | null;
    message: string;
  } | null>(null);

  // Fetch Polar products on mount
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setProductsLoading(true);
        setProductsFetchError(null);

        console.log('[Pricing] Fetching products from API...');
        const products = await getProducts();

        setPolarProducts(products);
        console.log('[Pricing] Successfully loaded products:', {
          count: products.length,
          plan_types: products.map(p => p.plan_type)
        });

      } catch (error) {
        console.error('[Pricing] Failed to load products:', error);

        const errorMessage = error instanceof Error ? error.message : 'Unknown error';

        if (errorMessage.includes('fetch') ||
            errorMessage.includes('network') ||
            errorMessage.includes('Failed to fetch')) {
          // Network/CORS error
          setProductsFetchError({
            type: 'network',
            message: 'Unable to connect to the server. Please check your internet connection.'
          });
          console.error('[Pricing] Network error detected. Check:');
          console.error('  1. Backend server is running (http://localhost:8080)');
          console.error('  2. CORS is configured (check browser DevTools Network tab)');
          console.error('  3. No ad blockers or extensions blocking requests');
        } else {
          // API error
          setProductsFetchError({
            type: 'api',
            message: `API error: ${errorMessage}`
          });
          console.error('[Pricing] API error:', errorMessage);
          console.error('  - Backend may be experiencing issues');
        }

        setPolarProducts([]);
      } finally {
        setProductsLoading(false);
      }
    };

    // Always fetch products, regardless of authentication status
    fetchProducts();
  }, []);

  // Handler for plan selection
  const handlePlanClick = async (planName: string, productId: string | null, planType: string) => {
    // Free plan - redirect to home
    if (planType === 'free') {
      router.push('/');
      return;
    }

    // Paid plans - check authentication
    if (!isAuthenticated) {
      // Redirect to login with return URL
      const returnUrl = encodeURIComponent(`/pricing?plan=${planName.toLowerCase()}`);
      router.push(`/login?returnUrl=${returnUrl}`);
      return;
    }

    // Check if we have a product ID
    if (!productId) {
      console.error('[Pricing] No product ID available for plan:', planName);
      alert('This plan is currently unavailable. Please contact support or check back later.');
      return;
    }

    // Get access token
    const accessToken = getAccessToken();
    if (!accessToken) {
      console.error('[Pricing] User is authenticated but no access token found');
      alert('Session expired. Please log in again.');
      router.push(`/login?returnUrl=${encodeURIComponent('/pricing')}`);
      return;
    }

    // User is authenticated - redirect to Polar checkout
    try {
      setIsLoading(true);
      setLoadingPlan(planName);
      console.log(`[Pricing] Creating checkout for product: ${productId}`);
      await redirectToCheckout(productId, accessToken);
    } catch (error) {
      console.error('[Pricing] Failed to create checkout session:', error);

      const errorMessage = error instanceof Error ? error.message : 'Failed to start checkout';

      if (errorMessage.includes('session has expired') || errorMessage.includes('log in again')) {
        // Token expired - redirect to login
        router.push(`/login?returnUrl=${encodeURIComponent('/pricing')}`);
      } else {
        // Other error - show user-friendly message
        alert(errorMessage + '\n\nPlease try again or contact support if the problem persists.');
      }

      setIsLoading(false);
      setLoadingPlan(null);
    }
  };

  // Helper to find Polar product by plan type
  const getPolarProductId = (planType: string): string | null => {
    const product = polarProducts.find(p => p.plan_type === planType);

    if (!product) {
      // Context-aware error messages
      if (productsFetchError?.type === 'network') {
        console.error(`[Pricing] Cannot load ${planType} plan - network error occurred`);
        console.error('[Pricing] Check browser DevTools → Network tab for failed requests');
      } else if (polarProducts.length === 0) {
        console.error(`[Pricing] No products returned from API`);
        console.error('[Pricing] Admin action: POST /api/v1/billing/sync-products');
      } else {
        console.error(`[Pricing] Product not found for plan type: ${planType}`);
        console.error('[Pricing] Available plan types:', polarProducts.map(p => p.plan_type));
        console.error('[Pricing] Admin may need to re-sync from Polar');
      }
    }

    return product?.id || null;
  };

  const plans = [
    {
      name: 'Free',
      price: '$0',
      period: '/month',
      description: 'Perfect for getting started',
      features: [
        { name: '10 short links', included: true },
        { name: 'Basic analytics', included: true },
        { name: 'Standard support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: false },
        { name: 'Team collaboration', included: false },
        { name: 'Advanced analytics', included: false },
        { name: 'API access', included: false },
        { name: 'Priority support', included: false }
      ],
      cta: 'Get Started Free',
      popular: false,
      productId: null, // Free plan - no product ID
      planType: 'free'
    },
    {
      name: 'Pro',
      price: '$5',
      period: '/month',
      description: 'For growing businesses',
      features: [
        { name: '1,000 short links', included: true },
        { name: 'Advanced analytics', included: true },
        { name: 'Priority support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: true },
        { name: 'Custom domains', included: true },
        { name: 'Team collaboration (5 users)', included: true },
        { name: 'API access', included: true },
        { name: 'Password protection', included: true }
      ],
      cta: 'Start Pro Plan',
      popular: true,
      productId: getPolarProductId('premium'), // Get from Polar products
      planType: 'premium'
    },
    {
      name: 'Ultra',
      price: '$29.99',
      period: '/month',
      description: 'For enterprises & power users',
      features: [
        { name: 'Unlimited short links', included: true },
        { name: 'Advanced analytics', included: true },
        { name: '24/7 premium support', included: true },
        { name: 'Link expiration', included: true },
        { name: 'QR codes', included: true },
        { name: 'Custom short codes', included: true },
        { name: 'Custom domains', included: true },
        { name: 'Unlimited team members', included: true },
        { name: 'Unlimited API access', included: true },
        { name: 'Password protection', included: true }
      ],
      cta: 'Choose Ultra',
      popular: false,
      productId: getPolarProductId('enterprise'), // Get from Polar products
      planType: 'enterprise'
    }
  ];

  const faqs = [
    {
      question: 'Can I change plans at any time?',
      answer: 'Yes, you can upgrade or downgrade your plan at any time. Changes take effect immediately, and billing is prorated.'
    },
    {
      question: 'What happens if I exceed my link limit?',
      answer: 'On the Free plan, you\'ll need to upgrade. On paid plans, we\'ll notify you and provide options to upgrade or purchase additional capacity.'
    },
    {
      question: 'Do you offer annual billing?',
      answer: 'Yes! Annual billing comes with a 20% discount. Contact us for annual pricing details.'
    },
    {
      question: 'Is there a free trial for paid plans?',
      answer: 'Yes, we offer a 14-day free trial for all paid plans. No credit card required to start.'
    },
    {
      question: 'Can I use my own domain?',
      answer: 'Custom domains are available on Pro and Enterprise plans. You can use your own branded short domain.'
    },
    {
      question: 'What kind of support do you provide?',
      answer: 'Free users get community support. Pro users get priority email support. Enterprise users get 24/7 phone and email support.'
    }
  ];

  return (
    <Layout>
      {/* Error Banner */}
      {productsFetchError && (
        <div className="bg-yellow-50 border-l-4 border-yellow-400 p-4">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-yellow-400" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-yellow-700">
                  {productsFetchError.type === 'network' ? (
                    <>
                      <strong>Connection Issue:</strong> Unable to load pricing information.
                      Please check your internet connection and refresh the page.
                    </>
                  ) : (
                    <>
                      <strong>Service Issue:</strong> Our pricing service is temporarily unavailable.
                      Please try again in a few moments.
                    </>
                  )}
                </p>
              </div>
            </div>
          </div>
        </div>
      )}

      {/* Hero Section */}
      <section className="bg-gradient-to-br from-primary-50 to-white py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-black mb-6">
              Simple, Transparent{' '}
              <span className="text-primary-600">Pricing</span>
            </h1>
            <p className="text-xl text-black mb-12 max-w-3xl mx-auto">
              Choose the perfect plan for your needs. All plans include our core features 
              with no hidden fees or surprise charges.
            </p>
          </div>
        </div>
      </section>

      {/* Pricing Cards */}
      <section className="py-20 bg-white">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="grid lg:grid-cols-3 gap-8">
            {plans.map((plan) => (
              <div 
                key={plan.name} 
                className={`relative bg-white rounded-2xl border-2 p-8 ${
                  plan.popular 
                    ? 'border-primary-600 shadow-xl transform scale-105' 
                    : 'border-gray-200 shadow-lg'
                }`}
              >
                {plan.popular && (
                  <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                    <div className="bg-primary-600 text-white px-4 py-2 rounded-full text-sm font-medium flex items-center space-x-1">
                      <StarIcon className="h-4 w-4" />
                      <span>Most Popular</span>
                    </div>
                  </div>
                )}

                <div className="text-center">
                  <h3 className="text-2xl font-bold text-black mb-2">{plan.name}</h3>
                  <p className="text-gray-600 mb-6">{plan.description}</p>
                  
                  <div className="mb-6">
                    <span className="text-5xl font-bold text-black">{plan.price}</span>
                    <span className="text-xl text-gray-600">{plan.period}</span>
                  </div>

                  <button
                    onClick={() => handlePlanClick(plan.name, plan.productId, plan.planType)}
                    disabled={isLoading || (plan.planType !== 'free' && productsLoading)}
                    className={`w-full py-3 px-6 rounded-lg font-semibold transition-colors ${
                      plan.popular
                        ? 'bg-primary-600 text-white hover:bg-primary-700 disabled:bg-primary-400'
                        : 'bg-gray-100 text-black hover:bg-gray-200 disabled:bg-gray-300'
                    } disabled:cursor-not-allowed`}
                  >
                    {loadingPlan === plan.name ? 'Loading...' : productsLoading && plan.planType !== 'free' ? 'Loading products...' : plan.cta}
                  </button>
                </div>

                <div className="mt-8">
                  <h4 className="font-semibold text-black mb-4">What's included:</h4>
                  <ul className="space-y-3">
                    {plan.features.map((feature) => (
                      <li key={feature.name} className="flex items-center space-x-3">
                        {feature.included ? (
                          <CheckIcon className="h-5 w-5 text-green-600 flex-shrink-0" />
                        ) : (
                          <XMarkIcon className="h-5 w-5 text-gray-400 flex-shrink-0" />
                        )}
                        <span className={`text-sm ${feature.included ? 'text-black' : 'text-gray-400'}`}>
                          {feature.name}
                        </span>
                      </li>
                    ))}
                  </ul>
                </div>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Feature Comparison Table */}
      <section className="py-20 bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-black mb-4">
              Compare Plans
            </h2>
            <p className="text-xl text-gray-600 max-w-2xl mx-auto">
              See exactly what's included in each plan to make the right choice for your needs.
            </p>
          </div>

          <div className="bg-white rounded-xl shadow-sm overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-gray-50">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-black">Feature</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Free</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Pro</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-black">Ultra</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-gray-200">
                  <tr>
                    <td className="px-6 py-4 text-sm text-black">Short Links</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">10</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">1,000</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr className="bg-gray-50">
                    <td className="px-6 py-4 text-sm text-black">Team Members</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">1</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">5</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr>
                    <td className="px-6 py-4 text-sm text-black">API Access</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">-</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">✓</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Unlimited</td>
                  </tr>
                  <tr className="bg-gray-50">
                    <td className="px-6 py-4 text-sm text-black">Support</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Standard</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">Priority</td>
                    <td className="px-6 py-4 text-center text-sm text-gray-600">24/7 Premium</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="py-20 bg-white">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-black mb-4">
              Frequently Asked Questions
            </h2>
            <p className="text-xl text-gray-600">
              Get answers to common questions about our pricing and plans.
            </p>
          </div>

          <div className="space-y-8">
            {faqs.map((faq, index) => (
              <div key={index} className="bg-gray-50 rounded-lg p-6">
                <h3 className="text-lg font-semibold text-black mb-3">
                  {faq.question}
                </h3>
                <p className="text-gray-600 leading-relaxed">
                  {faq.answer}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section className="py-20 bg-primary-600">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-white mb-4">
            Ready to Get Started?
          </h2>
          <p className="text-xl text-primary-100 mb-8 max-w-2xl mx-auto">
            Join thousands of users who trust Trunc for their link management needs.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href="/"
              className="bg-white text-primary-600 px-8 py-3 rounded-lg font-semibold hover:bg-gray-50 transition-colors"
            >
              Start Free Trial
            </a>
            <a
              href="/contact"
              className="border border-primary-400 text-white px-8 py-3 rounded-lg font-semibold hover:bg-primary-700 transition-colors"
            >
              Contact Sales
            </a>
          </div>
        </div>
      </section>
    </Layout>
  );
}