'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { CheckCircleIcon, ArrowRightIcon, SparklesIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';

export default function BillingSuccessPage() {
  const router = useRouter();
  const [countdown, setCountdown] = useState(5);
  const [activating, setActivating] = useState(true);
  const { refreshToken, getAccessToken } = useAuth();

  useEffect(() => {
    const activatePlan = async () => {
      const token = getAccessToken();
      if (token) {
        // Force-sync subscription from Polar API
        await fetch('/api/v1/billing/refresh-subscription', {
          method: 'POST',
          headers: { Authorization: `Bearer ${token}` },
        }).catch(() => {});

        // Refresh JWT so account_type reflects the new plan immediately
        await refreshToken().catch(() => {});
      }
      setActivating(false);
    };
    activatePlan();
  }, []); // eslint-disable-line react-hooks/exhaustive-deps

  useEffect(() => {
    if (activating) return;
    // Countdown timer — starts only after plan is activated
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          router.push('/dashboard');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [activating, router]);

  if (activating) {
    return (
      <div className="min-h-[80vh] bg-background flex items-center justify-center px-4">
        <div className="text-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary mx-auto mb-4"></div>
          <p className="text-muted-foreground">Activating your plan...</p>
        </div>
      </div>
    );
  }

  return (
    <div className="min-h-[80vh] bg-background flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Success Icon */}
        <div className="flex justify-center mb-8">
          <div className="relative">
            <div className="absolute inset-0 bg-green-500/20 rounded-full blur-2xl opacity-30 animate-pulse"></div>
            <div className="relative bg-card rounded-full p-6 shadow-xl border border-border">
              <CheckCircleIcon className="h-24 w-24 text-green-600 dark:text-green-500" />
            </div>
          </div>
        </div>

        {/* Success Message */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-foreground mb-4 flex items-center justify-center gap-3">
            <SparklesIcon className="h-10 w-10 text-yellow-500" />
            Payment Successful!
            <SparklesIcon className="h-10 w-10 text-yellow-500" />
          </h1>
          <p className="text-xl text-muted-foreground mb-6">
            Welcome to your upgraded plan! Your subscription is now active.
          </p>
          <p className="text-muted-foreground">
            You now have access to all premium features including advanced analytics,
            custom domains, unlimited URLs, and priority support.
          </p>
        </div>

        {/* Feature Cards */}
        <div className="grid md:grid-cols-3 gap-4 mb-8">
          <div className="bg-card rounded-xl p-6 shadow-sm border border-border">
            <div className="text-center">
              <div className="text-3xl mb-2">📊</div>
              <h3 className="font-semibold text-foreground mb-1">Advanced Analytics</h3>
              <p className="text-sm text-muted-foreground">Detailed insights & reports</p>
            </div>
          </div>
          <div className="bg-card rounded-xl p-6 shadow-sm border border-border">
            <div className="text-center">
              <div className="text-3xl mb-2">🌐</div>
              <h3 className="font-semibold text-foreground mb-1">Custom Domains</h3>
              <p className="text-sm text-muted-foreground">Use your own brand</p>
            </div>
          </div>
          <div className="bg-card rounded-xl p-6 shadow-sm border border-border">
            <div className="text-center">
              <div className="text-3xl mb-2">🔑</div>
              <h3 className="font-semibold text-foreground mb-1">API Access</h3>
              <p className="text-sm text-muted-foreground">Integrate anywhere</p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="bg-card rounded-2xl p-8 shadow-sm border border-border">
          <div className="flex flex-col sm:flex-row gap-4">
            <Link
              href="/dashboard"
              className="flex-1 flex items-center justify-center gap-2 bg-primary text-primary-foreground px-6 py-4 rounded-xl font-semibold hover:bg-primary/90 transition-all shadow-sm hover:shadow-md transform hover:-translate-y-0.5"
            >
              Go to Dashboard
              <ArrowRightIcon className="h-5 w-5" />
            </Link>
            <Link
              href="/billing"
              className="flex-1 flex items-center justify-center gap-2 bg-secondary text-secondary-foreground px-6 py-4 rounded-xl font-semibold hover:bg-secondary/80 transition-all"
            >
              Manage Subscription
            </Link>
          </div>

          {/* Auto-redirect notice */}
          <p className="text-center text-sm text-muted-foreground mt-6">
            Redirecting to dashboard in <span className="font-bold text-primary">{countdown}</span> seconds...
          </p>
        </div>

        {/* Help Section */}
        <div className="mt-8 text-center">
          <p className="text-sm text-muted-foreground">
            Need help getting started?{' '}
            <Link href="/docs" className="text-primary hover:text-primary/80 font-medium underline">
              Check out our documentation
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
