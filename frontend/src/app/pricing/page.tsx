'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { CheckIcon, XMarkIcon, StarIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import { getProducts, type PolarProduct, formatPrice } from '@/lib/polar';
import { PLAN_NAMES } from '@/lib/plan-utils';

export default function PricingPage() {
  const { isAuthenticated, getAccessToken } = useAuth();
  const router = useRouter();
  const [isLoading, setIsLoading] = useState(false);
  const [loadingPlan, setLoadingPlan] = useState<string | null>(null);
  const [billingInterval, setBillingInterval] = useState<'month' | 'year'>('month');
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
  const handlePlanClick = async (planName: string, productId: string | null, planType: string, priceId: string | null = null) => {
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

    // Get access token to ensure session is genuinely active
    const accessToken = getAccessToken();
    if (!accessToken) {
      console.error('[Pricing] User is authenticated but no access token found');
      alert('Session expired. Please log in again.');
      router.push(`/login?returnUrl=${encodeURIComponent('/pricing')}`);
      return;
    }

    // Instead of completing checkout directly on /pricing, route centrally to /billing
    // The query explicitly selects the tab and the plan
    const billingUrl = `/billing?plan=${planType.toLowerCase()}&interval=${billingInterval}`;
    router.push(billingUrl);
  };



  /**
   * Extract the primary price for display from a Polar product
   * Returns formatted price string, billing period, the specific Price ID, and Product ID
   */
  const getPriceDisplay = (planType: string): { price: string; period: string; priceId: string | null; productId: string | null } => {
    // Free plan avoids checkout by just redirecting to /dashboard natively or bypassing to 0
    if (planType === 'free') {
      return { price: '$0', period: '/month', priceId: 'd8f317f0-a783-4309-8644-7ab98c68c3dd', productId: 'd8f317f0-a783-4309-8644-7ab98c68c3dd' };
    }

    // Because the production database mapped to Polar is returning Yearly as entirely separate Products (e.g "Ultra Yearly")
    // we need to dynamically look for the appropriate product match based on the toggle interval requested.
    let productMatchStr = planType;
    if (billingInterval === 'year' && planType !== 'free') {
      productMatchStr = `${planType}`; // Polar plan_type currently holds both name bases but we might need to filter by name. Let's look closer at JSON.
    }

    // According to JSON, both share "plan_type": "enterprise" or "premium".
    // "name": "Ultra Yearly" vs "Ultra".
    let product = polarProducts.find((p: any) => p.plan_type === planType && p.is_recurring && (!p.name.includes("Yearly") && !p.name.includes("yearly")));
    if (billingInterval === 'year' && planType !== 'free') {
      product = polarProducts.find((p: any) => p.plan_type === planType && (p.name.includes("Yearly") || p.name.includes("yearly")) && p.is_recurring);
    }

    // If the precise year/month variant product is not found, fallback to the base
    if (!product) {
      product = polarProducts.find((p: any) => p.plan_type === planType && p.is_recurring);
    }

    if (!product || productsLoading || !product.prices || product.prices.length === 0) {
      return { price: '—', period: '', priceId: null, productId: null };
    }

    // Attempt to find the price matching the current user-selected billing interval normally
    const activePrice = product.prices.find(
      (p: any) => !p.is_archived && p.recurring_interval === billingInterval
    );

    if (activePrice) {
      return {
        price: formatPrice(activePrice.price_amount, activePrice.price_currency),
        period: billingInterval === 'year' ? '/year' : '/month',
        priceId: activePrice.id,
        productId: product.id
      };
    }

    // Fall back to monthly price native sync if requested interval isn't explicitly found natively
    const fallbackPrice = product.prices.find((p: any) => !p.is_archived && p.recurring_interval === 'month') || product.prices.find((p: any) => !p.is_archived);
    if (fallbackPrice) {
      const interval = fallbackPrice.recurring_interval || 'month';
      return {
        price: formatPrice(fallbackPrice.price_amount, fallbackPrice.price_currency),
        period: interval === 'year' ? '/year' : '/month',
        priceId: fallbackPrice.id,
        productId: product.id
      };
    }

    // All prices archived
    return { price: 'Unavailable', period: '', priceId: null, productId: null };
  };

  const plans = [
    {
      name: PLAN_NAMES.free,
      ...getPriceDisplay('free'),
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
      planType: 'free'
    },
    {
      name: PLAN_NAMES.premium,
      ...getPriceDisplay('premium'),
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
      cta: `Start ${PLAN_NAMES.premium} Plan`,
      popular: true,
      planType: 'premium'
    },
    {
      name: PLAN_NAMES.enterprise,
      ...getPriceDisplay('enterprise'),
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
      cta: `Choose ${PLAN_NAMES.enterprise}`,
      popular: false,
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
      answer: `Custom domains are available on ${PLAN_NAMES.premium} and ${PLAN_NAMES.enterprise} plans. You can use your own branded short domain.`
    },
    {
      question: 'What kind of support do you provide?',
      answer: `${PLAN_NAMES.free} users get community support. ${PLAN_NAMES.premium} users get priority email support. ${PLAN_NAMES.enterprise} users get 24/7 phone and email support.`
    }
  ];

  return (
    <Layout>
      {/* Error Banner */}
      {productsFetchError && (
        <div className="bg-amber-500/10 border-l-4 border-amber-500 p-4">
          <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
            <div className="flex">
              <div className="flex-shrink-0">
                <svg className="h-5 w-5 text-amber-500" viewBox="0 0 20 20" fill="currentColor">
                  <path fillRule="evenodd" d="M8.257 3.099c.765-1.36 2.722-1.36 3.486 0l5.58 9.92c.75 1.334-.213 2.98-1.742 2.98H4.42c-1.53 0-2.493-1.646-1.743-2.98l5.58-9.92zM11 13a1 1 0 11-2 0 1 1 0 012 0zm-1-8a1 1 0 00-1 1v3a1 1 0 002 0V6a1 1 0 00-1-1z" clipRule="evenodd" />
                </svg>
              </div>
              <div className="ml-3">
                <p className="text-sm text-amber-600 dark:text-amber-500">
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
      <section className="bg-gradient-to-br from-primary/10 to-background py-20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center">
            <h1 className="text-4xl md:text-6xl font-bold text-foreground mb-6">
              Simple, Transparent{' '}
              <span className="text-primary">Pricing</span>
            </h1>
            <p className="text-xl text-foreground mb-12 max-w-3xl mx-auto">
              Choose the perfect plan for your needs. All plans include our core features
              with no hidden fees or surprise charges.
            </p>
          </div>
        </div>
      </section>

      {/* Pricing Cards */}
      <section className="py-20 bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">

          {/* Monthly/Yearly Toggle */}
          <div className="flex justify-center items-center mb-12 space-x-4">
            <span className={`text-lg font-medium ${billingInterval === 'month' ? 'text-foreground' : 'text-muted-foreground'}`}>
              Monthly
            </span>
            <button
              onClick={() => setBillingInterval(prev => prev === 'month' ? 'year' : 'month')}
              className="relative inline-flex h-8 w-16 items-center rounded-full bg-primary transition-colors focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2"
            >
              <span
                className={`inline-block h-6 w-6 transform rounded-full bg-white transition-transform ${billingInterval === 'year' ? 'translate-x-9' : 'translate-x-1'
                  }`}
              />
            </button>
            <span className={`text-lg font-medium flex items-center ${billingInterval === 'year' ? 'text-foreground' : 'text-muted-foreground'}`}>
              Yearly
              <span className="ml-2 inline-flex items-center rounded-md bg-green-500/10 px-2 py-1 text-xs font-medium text-green-500 ring-1 ring-inset ring-green-500/20">
                Save 20%
              </span>
            </span>
          </div>

          <div className="grid lg:grid-cols-3 gap-8">
            {plans.map((plan) => (
              <div
                key={plan.name}
                className={`relative bg-card rounded-2xl border-2 p-8 transition-transform ${plan.popular
                  ? 'border-primary shadow-xl scale-105'
                  : 'border-border shadow-md hover:border-primary/50'
                  }`}
              >
                {plan.popular && (
                  <div className="absolute -top-4 left-1/2 transform -translate-x-1/2">
                    <div className="bg-primary text-primary-foreground px-4 py-2 rounded-full text-sm font-medium flex items-center space-x-1">
                      <StarIcon className="h-4 w-4" />
                      <span>Most Popular</span>
                    </div>
                  </div>
                )}

                <div className="text-center">
                  <h3 className="text-2xl font-bold text-foreground mb-2">{plan.name}</h3>
                  <p className="text-muted-foreground mb-6">{plan.description}</p>

                  <div className="mb-6">
                    <span className="text-5xl font-bold text-foreground">{plan.price}</span>
                    <span className="text-xl text-muted-foreground">{plan.period}</span>
                  </div>

                  <button
                    onClick={() => handlePlanClick(plan.name, plan.productId, plan.planType, plan.priceId)}
                    disabled={isLoading || (plan.planType !== 'free' && productsLoading)}
                    className={`w-full py-3 px-6 rounded-lg font-semibold transition-colors ${plan.popular
                      ? 'bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50'
                      : 'bg-muted text-foreground hover:bg-muted/80 disabled:opacity-50'
                      } disabled:cursor-not-allowed`}
                  >
                    {loadingPlan === plan.name ? 'Loading...' : productsLoading && plan.planType !== 'free' ? 'Loading products...' : plan.cta}
                  </button>
                </div>

                <div className="mt-8">
                  <h4 className="font-semibold text-foreground mb-4">What's included:</h4>
                  <ul className="space-y-3">
                    {plan.features.map((feature: any, index: number) => (
                      <li key={feature.name} className="flex items-center space-x-3">
                        {feature.included ? (
                          <CheckIcon className="h-5 w-5 text-primary flex-shrink-0" />
                        ) : (
                          <XMarkIcon className="h-5 w-5 text-muted-foreground opacity-50 flex-shrink-0" />
                        )}
                        <span className={`text-sm ${feature.included ? 'text-foreground' : 'text-muted-foreground opacity-50'}`}>
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
      <section className="py-20 bg-muted/20">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-foreground mb-4">
              Compare Plans
            </h2>
            <p className="text-xl text-muted-foreground max-w-2xl mx-auto">
              See exactly what's included in each plan to make the right choice for your needs.
            </p>
          </div>

          <div className="bg-card rounded-xl shadow-sm border border-border overflow-hidden">
            <div className="overflow-x-auto">
              <table className="w-full">
                <thead className="bg-muted/50">
                  <tr>
                    <th className="px-6 py-4 text-left text-sm font-semibold text-foreground">Feature</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-foreground">{PLAN_NAMES.free}</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-foreground">{PLAN_NAMES.premium}</th>
                    <th className="px-6 py-4 text-center text-sm font-semibold text-foreground">{PLAN_NAMES.enterprise}</th>
                  </tr>
                </thead>
                <tbody className="divide-y divide-border">
                  <tr>
                    <td className="px-6 py-4 text-sm text-foreground">Short Links</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">10</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">1,000</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">Unlimited</td>
                  </tr>
                  <tr className="bg-muted/20">
                    <td className="px-6 py-4 text-sm text-foreground">Team Members</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">1</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">5</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">Unlimited</td>
                  </tr>
                  <tr>
                    <td className="px-6 py-4 text-sm text-foreground">API Access</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">-</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">✓</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">Unlimited</td>
                  </tr>
                  <tr className="bg-muted/20">
                    <td className="px-6 py-4 text-sm text-foreground">Support</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">Standard</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">Priority</td>
                    <td className="px-6 py-4 text-center text-sm text-muted-foreground">24/7 Premium</td>
                  </tr>
                </tbody>
              </table>
            </div>
          </div>
        </div>
      </section>

      {/* FAQ Section */}
      <section className="py-20 bg-background">
        <div className="max-w-4xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="text-center mb-16">
            <h2 className="text-3xl font-bold text-foreground mb-4">
              Frequently Asked Questions
            </h2>
            <p className="text-xl text-muted-foreground">
              Get answers to common questions about our pricing and plans.
            </p>
          </div>

          <div className="space-y-8">
            {faqs.map((faq, index) => (
              <div key={index} className="bg-muted/30 rounded-lg p-6 border border-border">
                <h3 className="text-lg font-semibold text-foreground mb-3">
                  {faq.question}
                </h3>
                <p className="text-muted-foreground leading-relaxed">
                  {faq.answer}
                </p>
              </div>
            ))}
          </div>
        </div>
      </section>

      {/* Final CTA */}
      <section className="py-20 bg-primary">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 text-center">
          <h2 className="text-3xl font-bold text-primary-foreground mb-4">
            Ready to Get Started?
          </h2>
          <p className="text-xl text-primary-foreground/80 mb-8 max-w-2xl mx-auto">
            Join thousands of users who trust Trunc for their link management needs.
          </p>
          <div className="flex flex-col sm:flex-row gap-4 justify-center">
            <a
              href="/"
              className="bg-background text-primary px-8 py-3 rounded-lg font-semibold hover:bg-muted transition-colors"
            >
              Start Free Trial
            </a>
            <a
              href="/contact"
              className="border border-primary-foreground/30 text-primary-foreground px-8 py-3 rounded-lg font-semibold hover:bg-primary/80 transition-colors"
            >
              Contact Sales
            </a>
          </div>
        </div>
      </section>
    </Layout>
  );
}