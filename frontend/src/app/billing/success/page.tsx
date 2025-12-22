'use client';

import { useEffect, useState } from 'react';
import { useRouter } from 'next/navigation';
import Link from 'next/link';
import { CheckCircleIcon, ArrowRightIcon, SparklesIcon } from '@heroicons/react/24/outline';

export default function BillingSuccessPage() {
  const router = useRouter();
  const [countdown, setCountdown] = useState(5);

  useEffect(() => {
    // Countdown timer
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
  }, [router]);

  return (
    <div className="min-h-screen bg-gradient-to-br from-green-50 via-white to-blue-50 flex items-center justify-center px-4">
      <div className="max-w-2xl w-full">
        {/* Success Icon */}
        <div className="flex justify-center mb-8">
          <div className="relative">
            <div className="absolute inset-0 bg-green-400 rounded-full blur-2xl opacity-30 animate-pulse"></div>
            <div className="relative bg-white rounded-full p-6 shadow-xl">
              <CheckCircleIcon className="h-24 w-24 text-green-500" />
            </div>
          </div>
        </div>

        {/* Success Message */}
        <div className="text-center mb-8">
          <h1 className="text-4xl font-bold text-gray-900 mb-4 flex items-center justify-center gap-3">
            <SparklesIcon className="h-10 w-10 text-yellow-500" />
            Payment Successful!
            <SparklesIcon className="h-10 w-10 text-yellow-500" />
          </h1>
          <p className="text-xl text-gray-600 mb-6">
            Welcome to your upgraded plan! Your subscription is now active.
          </p>
          <p className="text-gray-500">
            You now have access to all premium features including advanced analytics,
            custom domains, unlimited URLs, and priority support.
          </p>
        </div>

        {/* Feature Cards */}
        <div className="grid md:grid-cols-3 gap-4 mb-8">
          <div className="bg-white rounded-xl p-6 shadow-lg border-2 border-green-100">
            <div className="text-center">
              <div className="text-3xl mb-2">📊</div>
              <h3 className="font-semibold text-gray-900 mb-1">Advanced Analytics</h3>
              <p className="text-sm text-gray-600">Detailed insights & reports</p>
            </div>
          </div>
          <div className="bg-white rounded-xl p-6 shadow-lg border-2 border-blue-100">
            <div className="text-center">
              <div className="text-3xl mb-2">🌐</div>
              <h3 className="font-semibold text-gray-900 mb-1">Custom Domains</h3>
              <p className="text-sm text-gray-600">Use your own brand</p>
            </div>
          </div>
          <div className="bg-white rounded-xl p-6 shadow-lg border-2 border-purple-100">
            <div className="text-center">
              <div className="text-3xl mb-2">🔑</div>
              <h3 className="font-semibold text-gray-900 mb-1">API Access</h3>
              <p className="text-sm text-gray-600">Integrate anywhere</p>
            </div>
          </div>
        </div>

        {/* Action Buttons */}
        <div className="bg-white rounded-2xl p-8 shadow-xl border-2 border-gray-100">
          <div className="flex flex-col sm:flex-row gap-4">
            <Link
              href="/dashboard"
              className="flex-1 flex items-center justify-center gap-2 bg-gradient-to-r from-primary-600 to-blue-600 text-white px-6 py-4 rounded-xl font-semibold hover:from-primary-700 hover:to-blue-700 transition-all shadow-lg hover:shadow-xl transform hover:scale-105"
            >
              Go to Dashboard
              <ArrowRightIcon className="h-5 w-5" />
            </Link>
            <Link
              href="/billing"
              className="flex-1 flex items-center justify-center gap-2 bg-gray-100 text-gray-700 px-6 py-4 rounded-xl font-semibold hover:bg-gray-200 transition-all"
            >
              Manage Subscription
            </Link>
          </div>

          {/* Auto-redirect notice */}
          <p className="text-center text-sm text-gray-500 mt-6">
            Redirecting to dashboard in <span className="font-bold text-primary-600">{countdown}</span> seconds...
          </p>
        </div>

        {/* Help Section */}
        <div className="mt-8 text-center">
          <p className="text-sm text-gray-600">
            Need help getting started?{' '}
            <Link href="/docs" className="text-primary-600 hover:text-primary-700 font-medium underline">
              Check out our documentation
            </Link>
          </p>
        </div>
      </div>
    </div>
  );
}
