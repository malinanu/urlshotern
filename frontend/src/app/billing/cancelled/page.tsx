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
    <div className="min-h-screen bg-gradient-to-br from-gray-50 via-white to-gray-100 flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Cancelled Icon */}
        <div className="flex justify-center mb-8">
          <div className="relative">
            <div className="absolute inset-0 bg-gray-300 rounded-full blur-2xl opacity-20"></div>
            <div className="relative bg-white rounded-full p-6 shadow-xl">
              <XCircleIcon className="h-24 w-24 text-gray-400" />
            </div>
          </div>
        </div>

        {/* Cancelled Message */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4">
            Checkout Cancelled
          </h1>
          <p className="text-xl text-gray-600 mb-6">
            No worries! You can upgrade anytime you're ready.
          </p>
          <p className="text-gray-500">
            Your current plan remains active and you can continue using all your existing features.
          </p>
        </div>

        {/* Why Upgrade Section */}
        <div className="bg-gradient-to-br from-primary-50 to-blue-50 rounded-2xl p-8 mb-8 border-2 border-primary-100">
          <h2 className="text-2xl font-bold text-gray-900 mb-6 text-center">
            Why Upgrade? 🚀
          </h2>
          <div className="grid md:grid-cols-2 gap-4">
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-gray-900">Unlimited URLs</h3>
                <p className="text-sm text-gray-600">Create as many links as you need</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-gray-900">Advanced Analytics</h3>
                <p className="text-sm text-gray-600">Deep insights into your traffic</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-gray-900">Custom Domains</h3>
                <p className="text-sm text-gray-600">Use your own branded domains</p>
              </div>
            </div>
            <div className="flex items-start gap-3">
              <div className="flex-shrink-0 w-8 h-8 bg-green-500 rounded-full flex items-center justify-center text-white font-bold">
                ✓
              </div>
              <div>
                <h3 className="font-semibold text-gray-900">Priority Support</h3>
                <p className="text-sm text-gray-600">Get help when you need it</p>
              </div>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
          <div className="flex flex-col sm:flex-row gap-4">
            <Link
              href="/pricing"
              className="flex-1 flex items-center justify-center gap-2 bg-gradient-to-r from-primary-600 to-blue-600 text-white px-6 py-4 rounded-xl font-semibold hover:from-primary-700 hover:to-blue-700 transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
            >
              <ArrowLeftIcon className="h-5 w-5" />
              View Plans Again
            </Link>
            <Link
              href="/dashboard"
              className="flex-1 flex items-center justify-center gap-2 bg-gray-100 text-gray-700 px-6 py-4 rounded-xl font-semibold hover:bg-gray-200 transition-all"
            >
              Back to Dashboard
            </Link>
          </div>

          {/* Auto-redirect notice */}
          <p className="text-center text-sm text-gray-500 mt-6">
            Redirecting to pricing in <span className="font-bold text-gray-700">{countdown}</span> seconds...
          </p>
        </div>

        {/* Help Section */}
        <div className="mt-8 bg-white rounded-xl p-6 shadow-lg border-2 border-yellow-100">
          <div className="flex items-start gap-4">
            <QuestionMarkCircleIcon className="h-8 w-8 text-yellow-500 flex-shrink-0" />
            <div>
              <h3 className="font-semibold text-gray-900 mb-2">Have questions about pricing?</h3>
              <p className="text-sm text-gray-600 mb-3">
                We're here to help! Reach out to our team if you have any questions about which plan
                is right for you.
              </p>
              <Link
                href="/contact"
                className="text-primary-600 hover:text-primary-700 font-medium text-sm underline"
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
