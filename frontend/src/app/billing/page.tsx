'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS, buildApiUrl } from '@/lib/api-config';
import {
  CreditCardIcon,
  CheckCircleIcon,
  XCircleIcon,
  DocumentTextIcon,
  CalendarIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';
import { getProducts, PolarProduct, formatPrice } from '@/lib/polar';
import { getPlanDisplayName, PLAN_NAMES } from '@/lib/plan-utils';
import { getCached, setCached } from '@/lib/prefetch-cache';

interface Plan {
  id: string;
  name: string;
  description: string;
  price: number;
  currency: string;
  interval: string;
  features: string[];
  limits: {
    max_urls: number;
    max_clicks_month: number;
    analytics_days: number;
    custom_domain: boolean;
    api_access: boolean;
    bulk_import: boolean;
    advanced_features: boolean;
  };
}

interface Subscription {
  id: number;
  user_id: number;
  plan_type: string;
  status: string;
  current_period_end?: string;
  cancel_at_period_end: boolean;
  billing_cycle: string;
  created_at: string;
}

interface Usage {
  current_urls: number;
  monthly_clicks: number;
  period_start: string;
  period_end: string;
}

interface BillingData {
  subscription: Subscription;
  plan: Plan;
  usage: Usage;
  limits: Plan['limits'];
  usage_percentages: {
    urls: number;
    clicks: number;
  };
  warnings: {
    url_limit_approaching: boolean;
    click_limit_approaching: boolean;
    url_limit_reached: boolean;
    click_limit_reached: boolean;
  };
}

// Tier order used to decide whether a plan card is an upgrade or a downgrade
const PLAN_RANK: Record<string, number> = { free: 0, premium: 1, enterprise: 2 };

// Strip the interval suffix so 'enterprise_yearly' and 'enterprise' compare as the same tier
const baseTier = (planId: string) => planId.replace('_yearly', '');

// Version of the auto-renewal disclosure below; sent with checkout as a consent record
const ARL_TERMS_VERSION = 'arl-2026-10';

interface AutoRenewConfirmProps {
  planName: string;
  price: string; // formatted, '—' when unknown
  interval: 'month' | 'year';
  discountCode: string | null;
  agreed: boolean;
  onAgreeChange: (agreed: boolean) => void;
  busy: boolean;
  onCancel: () => void;
  onConfirm: (terms: string) => void;
}

/**
 * Automatic-renewal disclosure and consent shown right before checkout
 * (California Bus. & Prof. Code §17602): the recurring price and frequency,
 * how to cancel, and an unticked box the customer must check to continue.
 */
function AutoRenewConfirm({
  planName, price, interval, discountCode, agreed, onAgreeChange, busy, onCancel, onConfirm,
}: AutoRenewConfirmProps) {
  const priceKnown = price !== '—';
  const period = interval === 'year' ? 'year' : 'month';
  const terms = `${price} per ${period} until canceled`;

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/60 p-4" role="dialog" aria-modal="true" aria-labelledby="arl-title">
      <div className="w-full max-w-lg rounded-lg border border-border bg-card p-6 shadow-xl">
        <h2 id="arl-title" className="text-xl font-semibold text-foreground">
          Confirm your {planName} subscription
        </h2>

        {priceKnown ? (
          <p className="mt-4 text-2xl font-bold text-foreground">
            {price} per {period}
            <span className="block text-sm font-medium text-muted-foreground">
              charged automatically until you cancel
            </span>
          </p>
        ) : (
          <p className="mt-4 text-sm text-destructive">
            We couldn&apos;t load the price for this plan. Please refresh the page and try again.
          </p>
        )}

        <ul className="mt-4 space-y-2 text-sm text-muted-foreground list-disc pl-5">
          <li>
            Your subscription starts today and <strong className="text-foreground">renews automatically every {period}</strong> at{' '}
            {price} (plus any applicable taxes) until you cancel.
          </li>
          <li>
            <strong className="text-foreground">Cancel anytime</strong> on this page (Billing &amp; Subscription → Cancel
            Subscription). Cancelling stops all future charges, and you keep access until the end of the period you&apos;ve
            already paid for.
          </li>
          {discountCode && (
            <li>
              Your discount code <span className="font-mono">{discountCode}</span> is applied at checkout, which shows how
              many payments it covers. Renewals after that are charged at {price} per {period}.
            </li>
          )}
        </ul>

        <label className="mt-5 flex items-start gap-3 rounded-md border border-border bg-muted/30 p-3 text-sm text-foreground">
          <input
            type="checkbox"
            className="mt-0.5 h-4 w-4"
            checked={agreed}
            onChange={(e) => onAgreeChange(e.target.checked)}
            disabled={!priceKnown}
          />
          <span>
            I agree that my {planName} subscription will renew automatically at {price} per {period} until I cancel, and I
            have read the cancellation terms above.
          </span>
        </label>

        <div className="mt-6 flex flex-col-reverse gap-2 sm:flex-row sm:justify-end">
          <button
            type="button"
            onClick={onCancel}
            className="rounded-md border border-border px-4 py-2 text-sm font-medium text-foreground hover:bg-muted"
          >
            Cancel
          </button>
          <button
            type="button"
            onClick={() => onConfirm(terms)}
            disabled={!agreed || !priceKnown || busy}
            className="flex items-center justify-center rounded-md bg-primary px-4 py-2 text-sm font-medium text-primary-foreground hover:bg-primary/90 disabled:cursor-not-allowed disabled:opacity-50"
          >
            {busy ? (
              <>
                <ArrowPathIcon className="mr-2 h-4 w-4 animate-spin" />
                Processing...
              </>
            ) : (
              'Continue to payment'
            )}
          </button>
        </div>
      </div>
    </div>
  );
}

function BillingContent() {
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const searchParams = useSearchParams();
  const [billingData, setBillingData] = useState<BillingData | null>(() => getCached<BillingData>('billing:sub'));
  const [plans, setPlans] = useState<Plan[]>(() => {
    const cached = getCached<{ plans: Plan[] }>('billing:products');
    return cached?.plans ?? [];
  });
  const [loading, setLoading] = useState(getCached('billing:sub') === null);
  const [upgrading, setUpgrading] = useState<string | null>(null);
  const [syncing, setSyncing] = useState(false);
  const [syncMessage, setSyncMessage] = useState<{ text: string; ok: boolean } | null>(null);

  const [billingInterval, setBillingInterval] = useState<'month' | 'year'>(
    (searchParams?.get('interval') as 'month' | 'year') || 'month'
  );
  const [discountCode, setDiscountCode] = useState(searchParams?.get('discount') || '');
  const [discountData, setDiscountData] = useState<any>(null);
  const [validatingDiscount, setValidatingDiscount] = useState(false);
  const [discountError, setDiscountError] = useState<string | null>(null);

  // California ARL: renewal terms are shown and agreed to before checkout starts
  const [confirmPlan, setConfirmPlan] = useState<Plan | null>(null);
  const [autoRenewAgreed, setAutoRenewAgreed] = useState(false);

  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Sync billing interval tab to the user's actual billing cycle
  useEffect(() => {
    if (billingData?.subscription?.billing_cycle && !searchParams?.get('interval')) {
      setBillingInterval(billingData.subscription.billing_cycle as 'month' | 'year');
    }
  }, [billingData]);

  // Handle plan auto-scroll from query parameter mapping
  useEffect(() => {
    const planParam = searchParams?.get('plan');
    if (planParam && !isLoading && !loading) {
      // Find the tier card with an ID or smoothly scroll down to the plans grid
      const plansSection = document.getElementById('pricing-plans');
      if (plansSection) {
        setTimeout(() => {
          plansSection.scrollIntoView({ behavior: 'smooth', block: 'start' });
        }, 300); // slight delay to ensure render
      }
    }
  }, [searchParams, isLoading, loading]);

  const [polarProducts, setPolarProducts] = useState<PolarProduct[]>([]);

  // Get live price from Polar for a given plan type + interval
  const getPolarPrice = (planType: string, interval: 'month' | 'year'): string => {
    if (planType === 'free') return '$0';
    const basePlan = planType.replace('_yearly', '');
    let product = interval === 'year'
      ? polarProducts.find(p => p.plan_type === basePlan && (p.name.includes('Yearly') || p.name.includes('yearly')) && p.is_recurring)
      : polarProducts.find(p => p.plan_type === basePlan && p.is_recurring && !p.name.includes('Yearly') && !p.name.includes('yearly'));
    if (!product) product = polarProducts.find(p => p.plan_type === basePlan && p.is_recurring);
    if (!product?.prices?.length) return '—';
    const price = product.prices.find(p => !p.is_archived && p.recurring_interval === interval)
      ?? product.prices.find(p => !p.is_archived);
    if (!price) return '—';
    return formatPrice(price.price_amount, price.price_currency);
  };

  const fetchBillingData = async () => {
    if (!isAuthenticated || !user) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      if (getCached('billing:sub') === null) setLoading(true);

      // Await sync so the billing data fetch reflects the latest plan from Polar
      try {
        await fetch(API_ENDPOINTS.billing.refreshSubscription, {
          method: 'POST',
          headers: { 'Authorization': `Bearer ${token}` },
        });
      } catch (_) {}

      // Fetch subscription, usage data, plans, and polar products concurrently
      const [usageResponse, plansResponse, products] = await Promise.all([
        fetch(API_ENDPOINTS.billing.usage, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }),
        fetch(API_ENDPOINTS.billing.plans, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }),
        getProducts()
      ]);

      if (usageResponse.ok) {
        const usageData = await usageResponse.json();
        setBillingData(usageData);
        setCached('billing:sub', usageData);
      }

      if (plansResponse.ok) {
        const plansData = await plansResponse.json();
        setPlans(plansData.plans || []);
        setCached('billing:products', plansData);
      }

      setPolarProducts(products);
    } catch (error) {
      console.error('Error fetching billing data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBillingData();
  }, [isAuthenticated, user]);

  const handleSyncWithPolar = async () => {
    const token = getAccessToken();
    if (!token || syncing) return;
    setSyncing(true);
    setSyncMessage(null);
    try {
      const res = await fetch(API_ENDPOINTS.billing.refreshSubscription, {
        method: 'POST',
        headers: { 'Authorization': `Bearer ${token}` },
      });
      if (res.ok) {
        await fetchBillingData();
        setSyncMessage({ text: 'Plan synced successfully.', ok: true });
      } else {
        setSyncMessage({ text: 'Sync failed. Try again later.', ok: false });
      }
    } catch (_) {
      setSyncMessage({ text: 'Sync failed. Try again later.', ok: false });
    } finally {
      setSyncing(false);
      setTimeout(() => setSyncMessage(null), 4000);
    }
  };

  const handleValidateDiscount = async () => {
    if (!discountCode) {
      setDiscountData(null);
      setDiscountError(null);
      return;
    }
    setValidatingDiscount(true);
    setDiscountError(null);
    try {
      // Find the currently selected plan to validate against
      const selectedPlan = plans.find((p: any) => p.interval === billingInterval && p.id !== 'free');
      const planId = selectedPlan ? selectedPlan.id : '';

      const response = await fetch(`${API_ENDPOINTS.billing.validateCoupon}?code=${discountCode}&plan_id=${planId}`, {
        headers: {
          'Authorization': `Bearer ${getAccessToken()}`,
        }
      });
      const data = await response.json();
      if (response.ok && data.valid) {
        setDiscountData(data);
      } else {
        setDiscountData(null);
        setDiscountError(data.error || 'Invalid discount code');
      }
    } catch (e) {
      setDiscountData(null);
      setDiscountError('Error validating discount code');
    } finally {
      setValidatingDiscount(false);
    }
  };

  // If discountCode from param exists on mount, validate it
  useEffect(() => {
    if (searchParams?.get('discount') && plans.length > 0) {
      handleValidateDiscount();
    }
  }, [searchParams, plans.length]);

  const handleUpgrade = async (planId: string, autoRenewTerms?: string) => {
    const token = getAccessToken();
    if (!token) return;

    setUpgrading(planId);

    // Downgrading to Free = cancelling the active subscription, not a checkout
    if (planId === 'free' || planId.startsWith('free')) {
      setUpgrading(null);
      await handleCancelSubscription();
      return;
    }

    try {
      const basePlanType = planId.replace('_yearly', '');

      const isYearly = planId.endsWith('_yearly') || billingInterval === 'year';

      // Find the correct active Polar product based on the interval and tier
      let product = polarProducts.find((p: any) => p.plan_type === basePlanType && p.is_recurring && (!p.name.includes("Yearly") && !p.name.includes("yearly")));
      if (isYearly && basePlanType !== 'free') {
        product = polarProducts.find((p: any) => p.plan_type === basePlanType && (p.name.includes("Yearly") || p.name.includes("yearly")) && p.is_recurring);
      }

      if (!product) {
        product = polarProducts.find((p: any) => p.plan_type === basePlanType && p.is_recurring);
      }

      if (!product) {
        console.error('Polar product not found for plan:', basePlanType);
        alert('Plan configuration error. Please ensure prices are configured in Polar.');
        setUpgrading(null);
        return;
      }

      const finalProductId = product.id;
      let finalPriceId: string | undefined;

      if (!product.prices || product.prices.length === 0) {
        alert('No active prices configured for this product in Polar.');
        setUpgrading(null);
        return;
      }

      if (isYearly) {
        const yearlyPrice = product.prices.find((p: any) => !p.is_archived && p.recurring_interval === 'year');
        if (yearlyPrice) {
          finalPriceId = yearlyPrice.id;
        } else {
          alert('A Yearly price has not been configured in Polar yet for this plan.');
          setUpgrading(null);
          return;
        }
      } else {
        const monthlyPrice = product.prices.find((p: any) => !p.is_archived && p.recurring_interval === 'month');
        if (monthlyPrice) {
          finalPriceId = monthlyPrice.id;
        } else {
          // Fall back to any non-archived price if specific monthly is not found 
          const fallbackPrice = product.prices.find((p: any) => !p.is_archived);
          if (fallbackPrice) {
            finalPriceId = fallbackPrice.id;
          } else {
            alert('No valid monthly price found in Polar for this plan.');
            setUpgrading(null);
            return;
          }
        }
      }

      const response = await fetch(API_ENDPOINTS.billing.checkout, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          product_id: finalProductId,
          price_id: finalPriceId,
          discount_code: discountData ? discountCode : undefined,
          success_url: `${window.location.origin}/billing/success`,
          cancel_url: `${window.location.origin}/billing`,
          auto_renew_consent: Boolean(autoRenewTerms),
          consent_text_version: ARL_TERMS_VERSION,
          auto_renew_terms: autoRenewTerms,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        if (data.checkout_url) {
          window.location.href = data.checkout_url;
        } else {
          console.error('Failed to create checkout session, no URL returned');
          alert('Failed to obtain a checkout session. Please try again later.');
        }
      } else {
        const errorData = await response.json();
        console.error('Failed to create checkout session:', errorData);
        alert(errorData.error || 'Failed to create checkout session');
      }
    } catch (error) {
      console.error('Error creating checkout session:', error);
      alert('An error occurred. Please try again.');
    } finally {
      setUpgrading(null);
    }
  };

  const handleCancelSubscription = async () => {
    if (!confirm('Are you sure you want to cancel your subscription? It will remain active until the end of the current billing period.')) {
      return;
    }

    const token = getAccessToken();
    if (!token) return;

    try {
      const response = await fetch(API_ENDPOINTS.billing.cancel, {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
      });

      if (response.ok) {
        await fetchBillingData(); // Refresh data
        alert('Subscription cancelled successfully. You will retain access until the end of your billing period.');
      } else {
        console.error('Failed to cancel subscription');
        alert('Failed to cancel subscription. Please try again.');
      }
    } catch (error) {
      console.error('Error cancelling subscription:', error);
      alert('Error cancelling subscription. Please try again.');
    }
  };

  const formatDate = (dateString: string) => {
    return new Date(dateString).toLocaleDateString('en-US', {
      year: 'numeric',
      month: 'long',
      day: 'numeric',
    });
  };

  const getUsageBarColor = (percentage: number, warning: boolean, reached: boolean) => {
    if (reached) return 'bg-red-500';
    if (warning) return 'bg-yellow-500';
    return 'bg-blue-500';
  };

  if (isLoading || loading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-[80vh] flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-[80vh] bg-background py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-foreground">Billing & Subscription</h1>
            <p className="mt-2 text-muted-foreground">
              Manage your subscription, view usage, and upgrade your plan.
            </p>
          </div>

          {billingData && (
            <>
              {/* Current Plan Card */}
              <div className="bg-card rounded-lg shadow-sm border border-border mb-8">
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div>
                      <h2 className="text-xl font-semibold text-foreground">Current Plan</h2>
                      <div className="flex items-center mt-1">
                        <span className="text-2xl font-bold text-primary">
                          {getPlanDisplayName(billingData.subscription.plan_type, polarProducts)}
                        </span>
                        {billingData.subscription.plan_type !== 'free' && (
                          <span className="ml-2 text-muted-foreground">
                            {getPolarPrice(billingData.subscription.plan_type, billingData.subscription.billing_cycle as 'month' | 'year' || 'month')}/{billingData.subscription.billing_cycle || 'month'}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center">
                      {billingData.subscription.status === 'active' ? (
                        <div className="flex items-center text-green-600 dark:text-green-500">
                          <CheckCircleIcon className="h-5 w-5 mr-1" />
                          <span>Active</span>
                        </div>
                      ) : (
                        <div className="flex items-center text-destructive">
                          <XCircleIcon className="h-5 w-5 mr-1" />
                          <span className="capitalize">{billingData.subscription.status}</span>
                        </div>
                      )}
                    </div>
                  </div>

                  {billingData.subscription.cancel_at_period_end && (
                    <div className="mb-4 p-4 bg-amber-500/10 border border-amber-500/20 rounded-md">
                      <div className="flex items-start">
                        <ExclamationTriangleIcon className="h-5 w-5 text-amber-600 dark:text-amber-500 mt-0.5 mr-2" />
                        <div>
                          <p className="text-amber-600 dark:text-amber-500 font-medium">Subscription Cancelled</p>
                          <p className="text-amber-600/80 dark:text-amber-500/80 text-sm">
                            Your subscription will end on{' '}
                            {billingData.subscription.current_period_end
                              ? formatDate(billingData.subscription.current_period_end)
                              : 'the end of the current period'
                            }.
                          </p>
                        </div>
                      </div>
                    </div>
                  )}

                  {/* Usage Statistics */}
                  <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
                    {/* URLs Usage */}
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-muted-foreground">URLs Used</span>
                        <span className="text-sm text-muted-foreground">
                          {billingData.usage.current_urls} / {billingData.limits.max_urls === -1 ? '∞' : billingData.limits.max_urls}
                        </span>
                      </div>
                      <div className="w-full bg-secondary rounded-full h-2">
                        <div
                          className={`h-2 rounded-full transition-all duration-300 ${getUsageBarColor(
                            billingData.usage_percentages.urls,
                            billingData.warnings.url_limit_approaching,
                            billingData.warnings.url_limit_reached
                          )}`}
                          style={{
                            width: `${Math.min(billingData.usage_percentages.urls, 100)}%`
                          }}
                        />
                      </div>
                      {billingData.warnings.url_limit_reached && (
                        <p className="text-sm text-destructive mt-1">Limit reached! Upgrade to create more URLs.</p>
                      )}
                      {billingData.warnings.url_limit_approaching && !billingData.warnings.url_limit_reached && (
                        <p className="text-sm text-amber-600 dark:text-amber-500 mt-1">Approaching limit. Consider upgrading.</p>
                      )}
                    </div>

                    {/* Clicks Usage */}
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-muted-foreground">Monthly Clicks</span>
                        <span className="text-sm text-muted-foreground">
                          {billingData.usage.monthly_clicks} / {billingData.limits.max_clicks_month === -1 ? '∞' : billingData.limits.max_clicks_month}
                        </span>
                      </div>
                      <div className="w-full bg-secondary rounded-full h-2">
                        <div
                          className={`h-2 rounded-full transition-all duration-300 ${getUsageBarColor(
                            billingData.usage_percentages.clicks,
                            billingData.warnings.click_limit_approaching,
                            billingData.warnings.click_limit_reached
                          )}`}
                          style={{
                            width: `${Math.min(billingData.usage_percentages.clicks, 100)}%`
                          }}
                        />
                      </div>
                      {billingData.warnings.click_limit_reached && (
                        <p className="text-sm text-destructive mt-1">Limit reached! Upgrade for more capacity.</p>
                      )}
                      {billingData.warnings.click_limit_approaching && !billingData.warnings.click_limit_reached && (
                        <p className="text-sm text-amber-600 dark:text-amber-500 mt-1">Approaching limit. Consider upgrading.</p>
                      )}
                    </div>
                  </div>

                  {/* Plan Features */}
                  <div className="mt-6">
                    <h3 className="text-sm font-medium text-muted-foreground mb-3">Current Plan Features</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                      {billingData.plan.features.map((feature: any, index: number) => (
                        <div key={index} className="flex items-center text-sm text-muted-foreground">
                          <CheckCircleIcon className="h-4 w-4 text-green-600 dark:text-green-500 mr-2 flex-shrink-0" />
                          {feature}
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="mt-6 flex flex-wrap items-center gap-3">
                    {billingData.subscription.plan_type !== 'free' && !billingData.subscription.cancel_at_period_end && (
                      <button
                        onClick={handleCancelSubscription}
                        className="px-4 py-2 text-destructive hover:text-destructive/80 border border-destructive rounded-md hover:bg-destructive/10 transition-colors"
                      >
                        Cancel Subscription
                      </button>
                    )}

                    <button
                      onClick={handleSyncWithPolar}
                      disabled={syncing}
                      className="flex items-center gap-1.5 px-4 py-2 text-sm border border-border rounded-md text-muted-foreground hover:text-foreground hover:bg-muted transition-colors disabled:opacity-50"
                    >
                      <ArrowPathIcon className={`h-4 w-4 ${syncing ? 'animate-spin' : ''}`} />
                      {syncing ? 'Syncing…' : 'Sync with Polar'}
                    </button>

                    {syncMessage && (
                      <span className={`text-sm ${syncMessage.ok ? 'text-green-600 dark:text-green-500' : 'text-destructive'}`}>
                        {syncMessage.text}
                      </span>
                    )}

                    {billingData.subscription.current_period_end && (
                      <div className="flex items-center text-sm text-muted-foreground">
                        <CalendarIcon className="h-4 w-4 mr-1" />
                        Renews on {formatDate(billingData.subscription.current_period_end)}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Available Plans */}
              <div id="pricing-plans" className="mb-8">
                <div className="flex flex-col md:flex-row md:items-center justify-between mb-6 gap-4">
                  <h2 className="text-xl font-semibold text-foreground">Available Plans</h2>
                  <div className="bg-muted p-1 rounded-lg inline-flex self-start md:self-auto">
                    <button
                      onClick={() => setBillingInterval('month')}
                      className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${billingInterval === 'month' ? 'bg-background shadow text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                    >
                      Monthly
                    </button>
                    <button
                      onClick={() => setBillingInterval('year')}
                      className={`px-4 py-2 rounded-md text-sm font-medium transition-colors ${billingInterval === 'year' ? 'bg-background shadow text-foreground' : 'text-muted-foreground hover:text-foreground'}`}
                    >
                      Yearly <span className="ml-1 text-[10px] uppercase text-green-500 font-bold">2 Months Free</span>
                    </button>
                  </div>
                </div>

                {/* Discount Code Section */}
                <div className="mb-8 bg-card rounded-lg p-6 border border-border">
                  <label className="block text-sm font-medium text-foreground mb-2">Have a discount code?</label>
                  <div className="flex flex-col sm:flex-row gap-2 max-w-md">
                    <input
                      type="text"
                      value={discountCode}
                      onChange={(e) => {
                        setDiscountCode(e.target.value.toUpperCase());
                        setDiscountError(null);
                        setDiscountData(null);
                      }}
                      placeholder="Enter code"
                      data-cs-mask
                      className="flex-1 bg-background border border-input rounded-md px-3 py-2 text-sm text-foreground focus:outline-none focus:ring-2 focus:ring-primary uppercase"
                    />
                    <button
                      onClick={handleValidateDiscount}
                      disabled={validatingDiscount || !discountCode}
                      className="bg-secondary text-secondary-foreground hover:bg-secondary/80 px-4 py-2 rounded-md text-sm font-medium disabled:opacity-50"
                    >
                      {validatingDiscount ? "Validating..." : "Apply"}
                    </button>
                  </div>
                  {discountError && <p className="text-destructive text-sm mt-2">{discountError}</p>}
                  {discountData && (
                    <p className="text-green-500 text-sm mt-2 flex items-center">
                      <CheckCircleIcon className="w-4 h-4 mr-1" />
                      Discount applied: {discountData.description || `${discountData.discount_value}${discountData.discount_type === 'percentage' ? '%' : ' months'} off`}
                    </p>
                  )}
                </div>

                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {plans.filter(p => p.interval === billingInterval || p.id === 'free').map((plan) => {
                    const currentTier = baseTier(billingData.subscription.plan_type);
                    const planTier = baseTier(plan.id);
                    const isCurrent = planTier === currentTier;
                    const isDowngrade = !isCurrent && (PLAN_RANK[planTier] ?? 0) < (PLAN_RANK[currentTier] ?? 0);
                    // No plan-change endpoint exists: a checkout for a lower paid tier would open a second subscription
                    const cancelScheduled = billingData.subscription.cancel_at_period_end;
                    const downgradeBlockedReason = !isDowngrade
                      ? null
                      : planTier !== 'free'
                        ? cancelScheduled
                          ? 'Available once your current plan ends.'
                          : 'Cancel your current subscription first, then choose this plan.'
                        : cancelScheduled
                          ? 'Cancellation already scheduled.'
                          : null;

                    return (
                    <div
                      key={plan.id}
                      className={`bg-card rounded-lg shadow-sm border-2 p-6 ${
                        isCurrent
                          ? 'border-primary ring-1 ring-primary'
                          : 'border-border hover:border-primary/50'
                        }`}
                    >
                      <div className="text-center mb-6">
                        <h3 className="text-xl font-semibold text-foreground">
                          {PLAN_NAMES[plan.id.replace('_yearly', '') as keyof typeof PLAN_NAMES] ?? plan.name}
                        </h3>
                        <p className="text-muted-foreground mt-1">{plan.description}</p>
                        <div className="mt-4">
                          <span className="text-3xl font-bold text-foreground">
                            {plan.id === 'free' ? '$0' : getPolarPrice(plan.id, billingInterval)}
                          </span>
                          <span className="text-muted-foreground">/{billingInterval}</span>
                        </div>
                      </div>

                      <ul className="space-y-3 mb-6">
                        {plan.features.map((feature: any, index: number) => (
                          <li key={index} className="flex items-center text-sm text-muted-foreground">
                            <CheckCircleIcon className="h-4 w-4 text-green-600 dark:text-green-500 mr-2 flex-shrink-0" />
                            {feature}
                          </li>
                        ))}
                      </ul>

                      {isCurrent ? (
                        <button
                          disabled
                          className="w-full py-2 px-4 bg-muted text-muted-foreground rounded-md cursor-not-allowed"
                        >
                          Current Plan
                        </button>
                      ) : downgradeBlockedReason ? (
                        <>
                          <button
                            disabled
                            className="w-full py-2 px-4 bg-muted text-muted-foreground rounded-md cursor-not-allowed"
                          >
                            Downgrade
                          </button>
                          <p className="mt-2 text-xs text-center text-muted-foreground">{downgradeBlockedReason}</p>
                        </>
                      ) : (
                        <button
                          onClick={() => {
                            if (isDowngrade) {
                              handleUpgrade(plan.id);
                              return;
                            }
                            // Paid plan: show the auto-renewal terms and get consent first
                            setAutoRenewAgreed(false);
                            setConfirmPlan(plan);
                          }}
                          disabled={upgrading === plan.id}
                          className="w-full py-2 px-4 bg-primary text-primary-foreground rounded-md hover:bg-primary/90 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                        >
                          {upgrading === plan.id ? (
                            <>
                              <ArrowPathIcon className="h-4 w-4 animate-spin mr-2" />
                              Processing...
                            </>
                          ) : isDowngrade ? (
                            'Downgrade'
                          ) : (
                            'Upgrade'
                          )}
                        </button>
                      )}
                    </div>
                    );
                  })}
                </div>
              </div>

              {confirmPlan && (
                <AutoRenewConfirm
                  planName={PLAN_NAMES[baseTier(confirmPlan.id) as keyof typeof PLAN_NAMES] ?? confirmPlan.name}
                  price={getPolarPrice(confirmPlan.id, billingInterval)}
                  interval={billingInterval}
                  discountCode={discountData ? discountCode : null}
                  agreed={autoRenewAgreed}
                  onAgreeChange={setAutoRenewAgreed}
                  busy={upgrading === confirmPlan.id}
                  onCancel={() => setConfirmPlan(null)}
                  onConfirm={async (terms) => {
                    await handleUpgrade(confirmPlan.id, terms);
                    setConfirmPlan(null);
                  }}
                />
              )}
            </>
          )}
        </div>
      </div>
    </Layout>
  );
}

export default function BillingPage() {
  return (
    <Suspense fallback={
      <Layout>
        <div className="min-h-[80vh] flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        </div>
      </Layout>
    }>
      <BillingContent />
    </Suspense>
  );
}