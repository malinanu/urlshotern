'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { XCircleIcon, ArrowLeftIcon, QuestionMarkCircleIcon } from '@heroicons/react/24/outline';

export default function BillingCancelledPage() {
  const router = useRouter();
  const [countdown, setCountdown] = useState(10);

  useEffect(() => {
    // Countdown timer
    const timer = setInterval(() => {
      setCountdown((prev) => {
        if (prev <= 1) {
          clearInterval(timer);
          router.push('/pricing');
          return 0;
        }
        return prev - 1;
      });
    }, 1000);

    return () => clearInterval(timer);
  }, [router]);

  return (
    <div className="min-h-[80vh] bg-background flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Cancelled Icon */}
        <div className="flex justify-center mb-8">
          <div className="relative">
            <div className="absolute inset-0 bg-muted/50 rounded-full blur-2xl opacity-50"></div>
            <div className="relative bg-card border border-border rounded-full p-6 shadow-sm">
              <XCircleIcon className="h-24 w-24 text-muted-foreground" />
            </div>
          </div>
        </div>

        {/* Cancelled Message */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-foreground mb-4">
            Checkout Cancelled
          </h1>
          <p className="text-xl text-muted-foreground mb-6">
            No worries! You can upgrade anytime you're ready.
          </p>
          <p className="text-muted-foreground">
            Your current plan remains active and you can continue using all your existing features.
          </p>
        </div>

        {/* Why Upgrade Section */}
        <div className="bg-primary/5 rounded-2xl p-8 mb-8 border border-primary/20">
          <h2 className="text-2xl font-bold text-foreground mb-6 text-center">
            Why Upgrade? 🚀
          </h2>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-600 dark:bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-foreground">Unlimited URLs</h3>
                <p className="text-sm text-muted-foreground">Create as many links as you need</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-600 dark:bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-foreground">Advanced Analytics</h3>
                <p className="text-sm text-muted-foreground">Deep insights into your traffic</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-600 dark:bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-foreground">Custom Domains</h3>
                <p className="text-sm text-muted-foreground">Use your own branded domains</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-600 dark:bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-foreground">Priority Support</h3>
                <p className="text-sm text-muted-foreground">Get help when you need it</p>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="bg-card rounded-2xl p-8 shadow-sm border border-border">
          <div className="flex flex-col sm:flex-row gap-4">
            <Link
              href="/pricing"
              className="flex-1 flex items-center justify-center gap-2 bg-primary text-primary-foreground px-6 py-4 rounded-xl font-semibold hover:bg-primary/90 transition-all shadow-sm transform hover:-translate-y-0.5"
            >
              <ArrowLeftIcon className="h-5 w-5" />
              View Plans Again
            </Link>
            <Link
              href="/dashboard"
              className="flex-1 flex items-center justify-center gap-2 bg-secondary text-secondary-foreground px-6 py-4 rounded-xl font-semibold hover:bg-secondary/80 transition-all"
            >
              Back to Dashboard
            </Link>
          </div>

          {/* Auto-redirect notice */}
          <p className="text-center text-sm text-muted-foreground mt-6">
            Redirecting to pricing in <span className="font-bold text-foreground">{countdown}</span> seconds...
          </p>
        </div>

        {/* Help Section */}
        <div className="mt-8 bg-card rounded-xl p-6 shadow-sm border border-amber-500/20">
          <div className="flex items-start gap-4">
            <QuestionMarkCircleIcon className="h-8 w-8 text-amber-500 flex-shrink-0" />
            <div>
              <h3 className="font-semibold text-foreground mb-2">Have questions about pricing?</h3>
              <p className="text-sm text-muted-foreground mb-3">
                We're here to help! Reach out to our team if you have any questions about which plan
                is right for you.
              </p>
              <Link
                href="/contact"
                className="text-primary hover:text-primary/80 font-medium text-sm underline"
              >
                Contact Support
              </Link>
            </div>
          </div>
        </div>
      </div>
    </div>
  );
}
