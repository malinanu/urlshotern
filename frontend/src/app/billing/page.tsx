'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { 
  CreditCardIcon,
  CheckCircleIcon,
  XCircleIcon,
  DocumentTextIcon,
  CalendarIcon,
  ExclamationTriangleIcon,
  ArrowPathIcon,
} from '@heroicons/react/24/outline';

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

export default function BillingPage() {
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const [billingData, setBillingData] = useState<BillingData | null>(null);
  const [plans, setPlans] = useState<Plan[]>([]);
  const [loading, setLoading] = useState(true);
  const [upgrading, setUpgrading] = useState<string | null>(null);
  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  const fetchBillingData = async () => {
    if (!isAuthenticated || !user) return;

    const token = getAccessToken();
    if (!token) return;

    try {
      setLoading(true);

      // Fetch subscription and usage data
      const [usageResponse, plansResponse] = await Promise.all([
        fetch('http://localhost:8080/api/v1/billing/usage', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        }),
        fetch('http://localhost:8080/api/v1/billing/plans', {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        })
      ]);

      if (usageResponse.ok) {
        const usageData = await usageResponse.json();
        setBillingData(usageData);
      }

      if (plansResponse.ok) {
        const plansData = await plansResponse.json();
        setPlans(plansData.plans || []);
      }
    } catch (error) {
      console.error('Error fetching billing data:', error);
    } finally {
      setLoading(false);
    }
  };

  useEffect(() => {
    fetchBillingData();
  }, [isAuthenticated, user]);

  const handleUpgrade = async (planId: string) => {
    const token = getAccessToken();
    if (!token) return;

    setUpgrading(planId);
    
    try {
      const response = await fetch('http://localhost:8080/api/v1/billing/checkout', {
        method: 'POST',
        headers: {
          'Authorization': `Bearer ${token}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({
          plan_id: planId,
          success_url: `${window.location.origin}/billing/success`,
          cancel_url: `${window.location.origin}/billing`,
        }),
      });

      if (response.ok) {
        const data = await response.json();
        // In a real implementation, redirect to Stripe Checkout
        window.open(data.checkout_url, '_blank');
      } else {
        console.error('Failed to create checkout session');
      }
    } catch (error) {
      console.error('Error creating checkout session:', error);
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
      const response = await fetch('http://localhost:8080/api/v1/billing/cancel-subscription', {
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
        <div className="min-h-screen flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50 py-8">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8">
          {/* Header */}
          <div className="mb-8">
            <h1 className="text-3xl font-bold text-black">Billing & Subscription</h1>
            <p className="mt-2 text-gray-600">
              Manage your subscription, view usage, and upgrade your plan.
            </p>
          </div>

          {billingData && (
            <>
              {/* Current Plan Card */}
              <div className="bg-white rounded-lg shadow-sm border mb-8">
                <div className="p-6">
                  <div className="flex items-center justify-between mb-4">
                    <div>
                      <h2 className="text-xl font-semibold text-black">Current Plan</h2>
                      <div className="flex items-center mt-1">
                        <span className="text-2xl font-bold text-primary-600 capitalize">
                          {billingData.plan.name}
                        </span>
                        {billingData.plan.price > 0 && (
                          <span className="ml-2 text-gray-600">
                            ${billingData.plan.price}/{billingData.plan.interval}
                          </span>
                        )}
                      </div>
                    </div>
                    <div className="flex items-center">
                      {billingData.subscription.status === 'active' ? (
                        <div className="flex items-center text-green-600">
                          <CheckCircleIcon className="h-5 w-5 mr-1" />
                          <span>Active</span>
                        </div>
                      ) : (
                        <div className="flex items-center text-red-600">
                          <XCircleIcon className="h-5 w-5 mr-1" />
                          <span className="capitalize">{billingData.subscription.status}</span>
                        </div>
                      )}
                    </div>
                  </div>

                  {billingData.subscription.cancel_at_period_end && (
                    <div className="mb-4 p-4 bg-yellow-50 border border-yellow-200 rounded-md">
                      <div className="flex items-start">
                        <ExclamationTriangleIcon className="h-5 w-5 text-yellow-600 mt-0.5 mr-2" />
                        <div>
                          <p className="text-yellow-800 font-medium">Subscription Cancelled</p>
                          <p className="text-yellow-700 text-sm">
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
                        <span className="text-sm font-medium text-gray-700">URLs Used</span>
                        <span className="text-sm text-gray-600">
                          {billingData.usage.current_urls} / {billingData.limits.max_urls === -1 ? '∞' : billingData.limits.max_urls}
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
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
                        <p className="text-sm text-red-600 mt-1">Limit reached! Upgrade to create more URLs.</p>
                      )}
                      {billingData.warnings.url_limit_approaching && !billingData.warnings.url_limit_reached && (
                        <p className="text-sm text-yellow-600 mt-1">Approaching limit. Consider upgrading.</p>
                      )}
                    </div>

                    {/* Clicks Usage */}
                    <div>
                      <div className="flex items-center justify-between mb-2">
                        <span className="text-sm font-medium text-gray-700">Monthly Clicks</span>
                        <span className="text-sm text-gray-600">
                          {billingData.usage.monthly_clicks} / {billingData.limits.max_clicks_month === -1 ? '∞' : billingData.limits.max_clicks_month}
                        </span>
                      </div>
                      <div className="w-full bg-gray-200 rounded-full h-2">
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
                        <p className="text-sm text-red-600 mt-1">Limit reached! Upgrade for more capacity.</p>
                      )}
                      {billingData.warnings.click_limit_approaching && !billingData.warnings.click_limit_reached && (
                        <p className="text-sm text-yellow-600 mt-1">Approaching limit. Consider upgrading.</p>
                      )}
                    </div>
                  </div>

                  {/* Plan Features */}
                  <div className="mt-6">
                    <h3 className="text-sm font-medium text-gray-700 mb-3">Current Plan Features</h3>
                    <div className="grid grid-cols-2 md:grid-cols-3 gap-2">
                      {billingData.plan.features.map((feature, index) => (
                        <div key={index} className="flex items-center text-sm text-gray-600">
                          <CheckCircleIcon className="h-4 w-4 text-green-500 mr-2 flex-shrink-0" />
                          {feature}
                        </div>
                      ))}
                    </div>
                  </div>

                  {/* Action Buttons */}
                  <div className="mt-6 flex space-x-4">
                    {billingData.subscription.plan_type !== 'free' && !billingData.subscription.cancel_at_period_end && (
                      <button
                        onClick={handleCancelSubscription}
                        className="px-4 py-2 text-red-600 hover:text-red-700 border border-red-600 rounded-md hover:bg-red-50 transition-colors"
                      >
                        Cancel Subscription
                      </button>
                    )}
                    
                    {billingData.subscription.current_period_end && (
                      <div className="flex items-center text-sm text-gray-600">
                        <CalendarIcon className="h-4 w-4 mr-1" />
                        Renews on {formatDate(billingData.subscription.current_period_end)}
                      </div>
                    )}
                  </div>
                </div>
              </div>

              {/* Available Plans */}
              <div className="mb-8">
                <h2 className="text-xl font-semibold text-black mb-6">Available Plans</h2>
                <div className="grid grid-cols-1 md:grid-cols-3 gap-6">
                  {plans.map((plan) => (
                    <div
                      key={plan.id}
                      className={`bg-white rounded-lg shadow-sm border-2 p-6 ${
                        plan.id === billingData.subscription.plan_type
                          ? 'border-primary-500 ring-1 ring-primary-500'
                          : 'border-gray-200 hover:border-primary-300'
                      }`}
                    >
                      <div className="text-center mb-6">
                        <h3 className="text-xl font-semibold text-black">{plan.name}</h3>
                        <p className="text-gray-600 mt-1">{plan.description}</p>
                        <div className="mt-4">
                          <span className="text-3xl font-bold text-black">
                            ${plan.price}
                          </span>
                          <span className="text-gray-600">/{plan.interval}</span>
                        </div>
                      </div>

                      <ul className="space-y-3 mb-6">
                        {plan.features.map((feature, index) => (
                          <li key={index} className="flex items-center text-sm">
                            <CheckCircleIcon className="h-4 w-4 text-green-500 mr-2 flex-shrink-0" />
                            {feature}
                          </li>
                        ))}
                      </ul>

                      {plan.id === billingData.subscription.plan_type ? (
                        <button
                          disabled
                          className="w-full py-2 px-4 bg-gray-100 text-gray-500 rounded-md cursor-not-allowed"
                        >
                          Current Plan
                        </button>
                      ) : (
                        <button
                          onClick={() => handleUpgrade(plan.id)}
                          disabled={upgrading === plan.id}
                          className="w-full py-2 px-4 bg-primary-600 text-white rounded-md hover:bg-primary-700 disabled:opacity-50 disabled:cursor-not-allowed flex items-center justify-center"
                        >
                          {upgrading === plan.id ? (
                            <>
                              <ArrowPathIcon className="h-4 w-4 animate-spin mr-2" />
                              Processing...
                            </>
                          ) : plan.price === 0 ? (
                            'Downgrade'
                          ) : (
                            'Upgrade'
                          )}
                        </button>
                      )}
                    </div>
                  ))}
                </div>
              </div>
            </>
          )}
        </div>
      </div>
    </Layout>
  );
}