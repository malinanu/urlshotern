'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import {
  Check,
  X,
  Sparkles,
  ShieldCheck,
  Zap,
  Globe2,
  ChevronDown,
  ArrowRight,
  HelpCircle,
  Lock,
} from 'lucide-react';
import { useAuth } from '@/contexts/AuthContext';
import { getProducts, type PolarProduct, formatPrice } from '@/lib/polar';
import { PLAN_NAMES } from '@/lib/plan-utils';
import { cn } from '@/lib/utils';

export default function PricingPage() {
  const { isAuthenticated, getAccessToken } = useAuth();
  const router = useRouter();
  const [billingInterval, setBillingInterval] = useState<'month' | 'year'>('month');
  const [polarProducts, setPolarProducts] = useState<PolarProduct[]>([]);
  const [productsLoading, setProductsLoading] = useState(true);
  const [productsFetchError, setProductsFetchError] = useState<{
    type: 'network' | 'api' | null;
    message: string;
  } | null>(null);
  const [openFaq, setOpenFaq] = useState<number | null>(null);

  // Fetch Polar products on mount
  useEffect(() => {
    const fetchProducts = async () => {
      try {
        setProductsLoading(true);
        setProductsFetchError(null);
        const products = await getProducts();
        setPolarProducts(products);
      } catch (error) {
        console.error('[Pricing] Failed to load products:', error);
        setProductsFetchError({
          type: 'network',
          message: 'Unable to connect to the pricing catalog. Displaying standard tiers.',
        });
        setPolarProducts([]);
      } finally {
        setProductsLoading(false);
      }
    };

    fetchProducts();
  }, []);

  const handlePlanClick = (planName: string, planType: string) => {
    if (planType === 'free') {
      router.push('/');
      return;
    }

    if (!isAuthenticated) {
      const returnUrl = encodeURIComponent(`/pricing?plan=${planType.toLowerCase()}`);
      router.push(`/login?returnUrl=${returnUrl}`);
      return;
    }

    const billingUrl = `/billing?plan=${planType.toLowerCase()}&interval=${billingInterval}`;
    router.push(billingUrl);
  };

  const getPriceDisplay = (planType: string) => {
    if (planType === 'free') {
      return { price: '$0', period: '/month' };
    }

    let product = polarProducts.find(
      (p: any) => p.plan_type === planType && p.is_recurring && (!p.name.includes('Yearly') && !p.name.includes('yearly'))
    );

    if (billingInterval === 'year') {
      const yearlyMatch = polarProducts.find(
        (p: any) => p.plan_type === planType && (p.name.includes('Yearly') || p.name.includes('yearly')) && p.is_recurring
      );
      if (yearlyMatch) product = yearlyMatch;
    }

    if (!product && polarProducts.length > 0) {
      product = polarProducts.find((p: any) => p.plan_type === planType && p.is_recurring);
    }

    if (!product || productsLoading || !product.prices || product.prices.length === 0) {
      if (planType === 'premium') {
        return billingInterval === 'year'
          ? { price: '$9', period: '/month (billed annually)' }
          : { price: '$12', period: '/month' };
      }
      if (planType === 'enterprise') {
        return billingInterval === 'year'
          ? { price: '$39', period: '/month (billed annually)' }
          : { price: '$49', period: '/month' };
      }
      return { price: '—', period: '' };
    }

    const activePrice = product.prices.find(
      (p: any) => !p.is_archived && p.recurring_interval === billingInterval
    );

    if (activePrice) {
      return {
        price: formatPrice(activePrice.price_amount, activePrice.price_currency),
        period: billingInterval === 'year' ? '/year' : '/month',
      };
    }

    const fallbackPrice = product.prices.find((p: any) => !p.is_archived);
    if (fallbackPrice) {
      return {
        price: formatPrice(fallbackPrice.price_amount, fallbackPrice.price_currency),
        period: fallbackPrice.recurring_interval === 'year' ? '/year' : '/month',
      };
    }

    return { price: '$12', period: '/month' };
  };

  const tiers = [
    {
      name: PLAN_NAMES.free,
      planType: 'free',
      ...getPriceDisplay('free'),
      badge: 'Starter',
      description: 'Ideal for individuals, test projects, and occasional link shortening.',
      features: [
        '10 Short URLs per month',
        'Standard QR code generation',
        '7-day click analytics history',
        '1 Micro-Landing Page',
        'Sub-15ms edge redirects',
        'Community support',
      ],
      notIncluded: [
        'Custom vanity domains',
        'Password protected links',
        'Link expiration scheduling',
        'Team collaboration seats',
      ],
      cta: 'Get Started Free',
      popular: false,
    },
    {
      name: PLAN_NAMES.premium,
      planType: 'premium',
      ...getPriceDisplay('premium'),
      badge: 'Most Popular',
      description: 'For growing creators, marketers, and businesses wanting brand authority.',
      features: [
        'Unlimited Short URLs',
        '1 Custom Vanity Domain (links.brand.com)',
        'Dynamic Vector QR Codes (SVG/PNG)',
        '365-day click telemetry history',
        '5 Digital vCards with 10 Themes',
        'Password protection on any link',
        'Link expiration by date or clicks',
        'Priority email support',
      ],
      notIncluded: [
        'Multi-domain pooling (5+)',
        'Team workspaces & seats',
      ],
      cta: `Upgrade to ${PLAN_NAMES.premium}`,
      popular: true,
    },
    {
      name: PLAN_NAMES.enterprise,
      planType: 'enterprise',
      ...getPriceDisplay('enterprise'),
      badge: 'Power Scale',
      description: 'For teams and organizations requiring multi-domain fleets and API scale.',
      features: [
        'Unlimited Short URLs',
        '5 Custom Branded Domains included',
        'White-label branding on micro-pages',
        'Lifetime unlimited telemetry history',
        'All 10 handcrafted vCard templates',
        '5 Team Member seats with RBAC',
        'Full REST API & Webhook triggers',
        'Dedicated SLA & 24/7 priority support',
      ],
      notIncluded: [],
      cta: `Start with ${PLAN_NAMES.enterprise}`,
      popular: false,
    },
  ];

  const faqs = [
    {
      q: 'Can I use my own custom domain for short links?',
      a: 'Yes! On the Premium plan you can connect 1 custom domain (e.g., links.yourbrand.com). On the Enterprise plan you get 5 custom domains. All domains receive automatic SSL/TLS certificates and sub-15ms Anycast routing.',
    },
    {
      q: 'What is a Dynamic QR Code versus a static QR code?',
      a: 'A static QR code permanently encodes a specific URL and can never be modified. A Trunc Dynamic QR Code routes through your short link, allowing you to change the destination URL anytime without reprinting your flyers, packaging, or business cards.',
    },
    {
      q: 'How does the digital vCard contact download work?',
      a: 'When someone views your Trunc micro-landing page and taps "Save Contact", an industry-standard .vcf file is generated and immediately saved into their iPhone Contacts or Android Google Contacts app with your name, phone, email, and photo.',
    },
    {
      q: 'Can I change or cancel my plan anytime?',
      a: 'Yes. You can upgrade, downgrade, or cancel your subscription at any time from your billing dashboard. When upgrading, changes apply immediately with prorated billing. No long-term lock-in.',
    },
    {
      q: 'What happens if I exceed my link limits on the Free plan?',
      a: 'Existing links continue to redirect and function forever with 100% uptime. To create additional links beyond your quota, you simply upgrade to the Premium plan for unlimited creation.',
    },
  ];

  return (
    <Layout>
      <div className="bg-background text-foreground min-h-screen">
        
        {/* Header Section */}
        <section className="relative overflow-hidden pt-16 pb-12 sm:pt-24 sm:pb-16 border-b border-border bg-gradient-to-b from-primary/10 via-card/30 to-background">
          <div className="pointer-events-none absolute inset-0 bg-grid opacity-40" aria-hidden />
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8 text-center relative z-10">
            <div className="inline-flex items-center gap-2 px-3.5 py-1 rounded-full border border-primary/20 bg-primary/10 text-primary text-xs font-semibold uppercase tracking-wider mb-6">
              <Sparkles className="h-3.5 w-3.5" />
              Transparent, Scalable Pricing
            </div>

            <h1 className="text-4xl sm:text-6xl font-extrabold tracking-tight text-foreground max-w-4xl mx-auto leading-tight">
              Invest in your brand’s{' '}
              <span className="bg-gradient-to-r from-primary via-sky-400 to-indigo-400 bg-clip-text text-transparent">
                connections & trust
              </span>
            </h1>

            <p className="mt-5 text-base sm:text-xl text-muted-foreground max-w-2xl mx-auto leading-relaxed">
              Start free forever. Upgrade when you need custom domains, team seats, dynamic QR vectors, and deep telemetry.
            </p>

            {/* Monthly / Annual Billing Toggle */}
            <div className="mt-10 flex items-center justify-center gap-4">
              <span className={cn('text-sm font-semibold transition-colors', billingInterval === 'month' ? 'text-foreground' : 'text-muted-foreground')}>
                Monthly Billing
              </span>
              
              <button
                type="button"
                onClick={() => setBillingInterval(billingInterval === 'month' ? 'year' : 'month')}
                className="relative inline-flex h-7 w-14 items-center rounded-full bg-muted p-1 transition-colors border border-border focus:outline-none focus:ring-2 focus:ring-primary focus:ring-offset-2"
              >
                <span
                  className={cn(
                    'inline-block h-5 w-5 transform rounded-full bg-primary shadow-md transition-transform',
                    billingInterval === 'year' ? 'translate-x-7' : 'translate-x-0'
                  )}
                />
              </button>

              <div className="flex items-center gap-2">
                <span className={cn('text-sm font-semibold transition-colors', billingInterval === 'year' ? 'text-foreground' : 'text-muted-foreground')}>
                  Annual Billing
                </span>
                <span className="px-2.5 py-0.5 rounded-full text-xs font-bold bg-emerald-500/10 text-emerald-400 border border-emerald-500/20">
                  Save 20%
                </span>
              </div>
            </div>
          </div>
        </section>

        {/* Pricing Cards Grid */}
        <section className="py-16 sm:py-24 border-b border-border">
          <div className="mx-auto max-w-7xl px-4 sm:px-6 lg:px-8">
            <div className="grid grid-cols-1 lg:grid-cols-3 gap-8 items-stretch">
              {tiers.map((tier) => (
                <div
                  key={tier.name}
                  className={cn(
                    'rounded-3xl p-8 flex flex-col justify-between transition-all duration-300 relative',
                    tier.popular
                      ? 'border-2 border-primary bg-card shadow-2xl shadow-primary/10'
                      : 'border border-border bg-card/60 hover:border-border/80'
                  )}
                >
                  {tier.popular && (
                    <div className="absolute -top-3.5 left-1/2 -translate-x-1/2 px-4 py-1 rounded-full text-xs font-extrabold bg-gradient-to-r from-primary to-sky-400 text-primary-foreground shadow-md uppercase tracking-wider">
                      {tier.badge}
                    </div>
                  )}

                  <div>
                    <div className="flex items-center justify-between">
                      <h3 className="text-2xl font-black text-foreground">{tier.name}</h3>
                      {!tier.popular && (
                        <span className="text-xs font-semibold px-2.5 py-0.5 rounded-full bg-muted text-muted-foreground border border-border">
                          {tier.badge}
                        </span>
                      )}
                    </div>

                    <p className="mt-3 text-xs sm:text-sm text-muted-foreground leading-relaxed">
                      {tier.description}
                    </p>

                    {/* Price Tag */}
                    <div className="mt-6 flex items-baseline gap-1.5 pb-6 border-b border-border">
                      <span className="text-4xl sm:text-5xl font-black text-foreground tabular-nums">
                        {tier.price}
                      </span>
                      <span className="text-xs text-muted-foreground font-medium">
                        {tier.period}
                      </span>
                    </div>

                    {/* Features List */}
                    <ul className="mt-6 space-y-3 text-sm">
                      {tier.features.map((feature, idx) => (
                        <li key={idx} className="flex items-start gap-2.5 text-foreground">
                          <Check className="h-4 w-4 text-primary shrink-0 mt-0.5" />
                          <span>{feature}</span>
                        </li>
                      ))}
                      {tier.notIncluded.map((feature, idx) => (
                        <li key={idx} className="flex items-start gap-2.5 text-muted-foreground/60 line-through">
                          <X className="h-4 w-4 text-muted-foreground/40 shrink-0 mt-0.5" />
                          <span>{feature}</span>
                        </li>
                      ))}
                    </ul>
                  </div>

                  <div className="mt-8 pt-6 border-t border-border">
                    <button
                      onClick={() => handlePlanClick(tier.name, tier.planType)}
                      className={cn(
                        'w-full py-3.5 rounded-xl text-sm font-semibold transition-all shadow-md',
                        tier.popular
                          ? 'bg-primary text-primary-foreground hover:bg-primary/90 shadow-primary/20'
                          : 'bg-muted hover:bg-muted/80 text-foreground border border-border'
                      )}
                    >
                      {tier.cta}
                    </button>
                  </div>
                </div>
              ))}
            </div>

            {/* Custom Enterprise Banner */}
            <div className="mt-12 rounded-3xl border border-border bg-gradient-to-r from-card via-card/80 to-muted/40 p-8 sm:p-10 flex flex-col sm:flex-row items-center justify-between gap-6 shadow-md">
              <div>
                <div className="inline-flex items-center gap-1.5 px-2.5 py-0.5 rounded-full text-xs font-semibold bg-purple-500/10 text-purple-400 border border-purple-500/20 mb-2">
                  Need High-Volume Dedicated Infrastructure?
                </div>
                <h3 className="text-2xl font-bold text-foreground">Custom Enterprise & Agency Plans</h3>
                <p className="text-sm text-muted-foreground mt-1 max-w-2xl">
                  Dedicated database partitions, custom domain SSL pools, custom SSO/SAML integrations, 99.99% SLA guarantees, and invoice billing.
                </p>
              </div>
              <Link
                href="mailto:contact@trunc.site?subject=Enterprise%20Inquiry"
                className="px-6 py-3 rounded-xl text-sm font-semibold bg-foreground text-background hover:bg-foreground/90 transition-all shrink-0 shadow-md"
              >
                Contact Sales
              </Link>
            </div>
          </div>
        </section>

        {/* FAQ Section */}
        <section className="py-16 sm:py-24 border-b border-border bg-card/20">
          <div className="mx-auto max-w-4xl px-4 sm:px-6 lg:px-8">
            <div className="text-center mb-12">
              <h2 className="text-3xl font-extrabold tracking-tight text-foreground">
                Frequently Asked Questions
              </h2>
              <p className="text-muted-foreground text-sm sm:text-base mt-2">
                Everything you need to know about plans, domains, and billing.
              </p>
            </div>

            <div className="space-y-4">
              {faqs.map((faq, idx) => {
                const isOpen = openFaq === idx;
                return (
                  <div
                    key={idx}
                    className="rounded-2xl border border-border bg-card overflow-hidden transition-colors"
                  >
                    <button
                      type="button"
                      onClick={() => setOpenFaq(isOpen ? null : idx)}
                      className="w-full p-5 text-left flex items-center justify-between gap-4 text-foreground font-semibold text-sm sm:text-base"
                    >
                      <span>{faq.q}</span>
                      <ChevronDown
                        className={cn(
                          'h-4 w-4 text-muted-foreground transition-transform duration-200 shrink-0',
                          isOpen ? 'rotate-180 text-primary' : ''
                        )}
                      />
                    </button>
                    {isOpen && (
                      <div className="px-5 pb-5 pt-1 text-sm text-muted-foreground leading-relaxed border-t border-border/50">
                        {faq.a}
                      </div>
                    )}
                  </div>
                );
              })}
            </div>
          </div>
        </section>

        {/* Trust Badges Strip */}
        <section className="py-12 bg-background text-center text-xs text-muted-foreground">
          <div className="mx-auto max-w-7xl px-4 flex flex-wrap items-center justify-center gap-8">
            <div className="flex items-center gap-2">
              <ShieldCheck className="h-4 w-4 text-emerald-400" />
              <span>Bank-level SSL / TLS 1.3 Security</span>
            </div>
            <div className="flex items-center gap-2">
              <Zap className="h-4 w-4 text-primary" />
              <span>99.99% Uptime Guarantee</span>
            </div>
            <div className="flex items-center gap-2">
              <Lock className="h-4 w-4 text-purple-400" />
              <span>Zero Vendor Lock-in & Full Data Export</span>
            </div>
          </div>
        </section>

      </div>
    </Layout>
  );
}
