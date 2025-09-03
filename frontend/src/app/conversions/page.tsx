'use client';

import { useState, useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import ConversionGoalManager from '@/components/conversion/ConversionGoalManager';
import ConversionAnalytics from '@/components/conversion/ConversionAnalytics';
import { useAuth } from '@/contexts/AuthContext';
import {
  ChartBarIcon,
  Cog6ToothIcon,
} from '@heroicons/react/24/outline';

export default function ConversionsPage() {
  const { isAuthenticated, isLoading } = useAuth();
  const [activeTab, setActiveTab] = useState<'goals' | 'analytics'>('goals');
  const router = useRouter();

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen bg-gray-50 flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  if (!isAuthenticated) {
    return null; // Will redirect
  }

  return (
    <Layout>
      <div className="min-h-screen bg-gray-50">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          {/* Page Header */}
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <ChartBarIcon className="h-8 w-8 text-primary-600" />
              <h1 className="text-3xl font-bold text-gray-900">Conversion Tracking</h1>
            </div>
            <p className="text-gray-600">
              Set up conversion goals and analyze your funnel performance to optimize user journeys.
            </p>
          </div>

          {/* Tab Navigation */}
          <div className="bg-white rounded-lg shadow-sm border border-gray-200 mb-6">
            <div className="border-b border-gray-200">
              <nav className="flex space-x-8 px-6">
                <button
                  onClick={() => setActiveTab('goals')}
                  className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                    activeTab === 'goals'
                      ? 'border-primary-500 text-primary-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <Cog6ToothIcon className="h-4 w-4" />
                    Goal Management
                  </div>
                </button>
                <button
                  onClick={() => setActiveTab('analytics')}
                  className={`py-4 px-1 border-b-2 font-medium text-sm whitespace-nowrap ${
                    activeTab === 'analytics'
                      ? 'border-primary-500 text-primary-600'
                      : 'border-transparent text-gray-500 hover:text-gray-700 hover:border-gray-300'
                  }`}
                >
                  <div className="flex items-center gap-2">
                    <ChartBarIcon className="h-4 w-4" />
                    Analytics
                  </div>
                </button>
              </nav>
            </div>

            {/* Tab Content */}
            <div className="p-6">
              {activeTab === 'goals' && <ConversionGoalManager />}
              {activeTab === 'analytics' && <ConversionAnalytics />}
            </div>
          </div>

          {/* Help Section */}
          <div className="bg-blue-50 border border-blue-200 rounded-lg p-6">
            <h3 className="text-lg font-semibold text-blue-900 mb-2">Getting Started with Conversion Tracking</h3>
            <div className="text-blue-800 space-y-2">
              <p><strong>1. Create Goals:</strong> Define what actions you want users to take (URL visits, form submissions, purchases, custom events).</p>
              <p><strong>2. Track Events:</strong> Use our tracking API or JavaScript snippet to record conversion events.</p>
              <p><strong>3. Analyze Performance:</strong> View conversion rates, attribution models, and optimize your funnel.</p>
              <p><strong>4. Attribution Models:</strong> Compare first-click, last-click, and linear attribution to understand user journeys.</p>
            </div>
          </div>

          {/* Integration Code Example */}
          <div className="bg-gray-50 border border-gray-200 rounded-lg p-6 mt-6">
            <h3 className="text-lg font-semibold text-gray-900 mb-4">JavaScript Tracking Example</h3>
            <pre className="bg-gray-800 text-green-400 p-4 rounded-md overflow-x-auto text-sm">
{`// Track a conversion event
async function trackConversion(goalId, sessionId, value = 0) {
  try {
    const response = await fetch('/api/v1/conversions/track', {
      method: 'POST',
      headers: {
        'Content-Type': 'application/json',
      },
      body: JSON.stringify({
        goal_id: goalId,
        session_id: sessionId,
        conversion_value: value,
        custom_data: {
          page_url: window.location.href,
          user_agent: navigator.userAgent
        }
      })
    });
    
    if (response.ok) {
      console.log('Conversion tracked successfully');
    }
  } catch (error) {
    console.error('Failed to track conversion:', error);
  }
}

// Example usage
trackConversion(123, 'user-session-id', 29.99);`}
            </pre>
          </div>
        </div>
      </div>
    </Layout>
  );
}