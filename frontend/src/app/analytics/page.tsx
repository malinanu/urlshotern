'use client';

import { useState, useEffect, Suspense, startTransition } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import RealTimeAnalytics from '@/components/analytics/RealTimeAnalytics';
import AdvancedAnalytics from '@/components/analytics/AdvancedAnalytics';
import { AIInsightsPanel } from '@/components/analytics/AIInsightsPanel';
import { useAuth } from '@/contexts/AuthContext';
import {
  ChartBarIcon,
  PresentationChartLineIcon,
  ChartPieIcon,
  SparklesIcon,
} from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function AnalyticsContent() {
  const { isAuthenticated, isLoading } = useAuth();
  const [selectedShortCode, setSelectedShortCode] = useState<string>('');
  const [userUrls, setUserUrls] = useState<any[]>([]);
  const [activeView, setActiveView] = useState<'realtime' | 'advanced' | 'ai'>('realtime');
  const router = useRouter();
  const searchParams = useSearchParams();

  // Get shortCode from URL params if provided
  useEffect(() => {
    const shortCode = searchParams.get('shortCode');
    if (shortCode) {
      setSelectedShortCode(shortCode);
    }
  }, [searchParams]);

  // Redirect if not authenticated
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      router.push('/login');
    }
  }, [isAuthenticated, isLoading, router]);

  // Fetch user URLs for selection
  useEffect(() => {
    const fetchUserUrls = async () => {
      if (!isAuthenticated) return;

      try {
        const token = localStorage.getItem('access_token');
        const response = await fetch(`${API_BASE_URL}/api/v1/my-urls`, {
          headers: {
            'Authorization': `Bearer ${token}`,
            'Content-Type': 'application/json',
          },
        });

        if (response.ok) {
          const data = await response.json();
          setUserUrls(data.urls || []);

          // If no shortCode selected and we have URLs, select the first one
          if (!selectedShortCode && data.urls && data.urls.length > 0) {
            setSelectedShortCode(data.urls[0].short_code);
          }
        }
      } catch (error) {
        console.error('Error fetching user URLs:', error);
      }
    };

    fetchUserUrls();
  }, [isAuthenticated, selectedShortCode]);

  if (isLoading) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary"></div>
        </div>
      </Layout>
    );
  }

  if (!isAuthenticated) {
    return null; // Will redirect
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background">
        <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
          <div className="mb-8">
            <div className="flex items-center gap-3 mb-2">
              <ChartBarIcon className="h-8 w-8 text-primary" />
              <h1 className="text-3xl font-bold text-foreground">Analytics Dashboard</h1>
            </div>
            <p className="text-muted-foreground">
              Monitor your URL performance with real-time analytics and insights.
            </p>
          </div>

          {/* URL Selection */}
          <div className="bg-card rounded-lg shadow-sm border border-border p-6 mb-6">
            <label htmlFor="url-select" className="block text-sm font-medium text-foreground mb-2">
              Select URL to analyze
            </label>
            <select
              id="url-select"
              value={selectedShortCode}
              onChange={(e) => setSelectedShortCode(e.target.value)}
              className="w-full max-w-md px-3 py-2 border border-input rounded-md focus:ring-2 focus:ring-primary focus:border-primary bg-background text-foreground"
            >
              <option value="">Select a URL...</option>
              {userUrls.map((url) => (
                <option key={url.short_code} value={url.short_code}>
                  {url.short_url} - {url.original_url.length > 50
                    ? url.original_url.substring(0, 50) + '...'
                    : url.original_url}
                </option>
              ))}
            </select>
          </div>

          {/* Analytics Navigation */}
          {selectedShortCode && (
            <div className="bg-card rounded-lg shadow-sm border border-border p-6 mb-6">
              <div className="border-b border-border">
                <nav className="flex space-x-8">
                  <button
                    onClick={() => startTransition(() => setActiveView('realtime'))}
                    className={`py-2 px-1 border-b-2 font-medium text-sm ${activeView === 'realtime'
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                      }`}
                  >
                    <div className="flex items-center gap-2">
                      <PresentationChartLineIcon className="h-4 w-4" />
                      Real-Time Analytics
                    </div>
                  </button>
                  <button
                    onClick={() => startTransition(() => setActiveView('advanced'))}
                    className={`py-2 px-1 border-b-2 font-medium text-sm ${activeView === 'advanced'
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                      }`}
                  >
                    <div className="flex items-center gap-2">
                      <ChartPieIcon className="h-4 w-4" />
                      Advanced Analytics
                    </div>
                  </button>
                  <button
                    onClick={() => startTransition(() => setActiveView('ai'))}
                    className={`py-2 px-1 border-b-2 font-medium text-sm ${activeView === 'ai'
                      ? 'border-primary text-primary'
                      : 'border-transparent text-muted-foreground hover:text-foreground hover:border-border'
                      }`}
                  >
                    <div className="flex items-center gap-2">
                      <SparklesIcon className="h-4 w-4" />
                      AI Insights
                    </div>
                  </button>
                </nav>
              </div>
            </div>
          )}

          {/* Analytics Component */}
          {selectedShortCode ? (
            <>
              {activeView === 'realtime' && (
                <Suspense fallback={<div className="h-64 bg-card border border-border rounded-lg animate-pulse" />}>
                  <RealTimeAnalytics
                    shortCode={selectedShortCode}
                    refreshInterval={30000}
                  />
                </Suspense>
              )}
              {activeView === 'advanced' && (
                <Suspense fallback={<div className="h-64 bg-card border border-border rounded-lg animate-pulse" />}>
                  <AdvancedAnalytics
                    shortCode={selectedShortCode}
                    days={30}
                  />
                </Suspense>
              )}
              {activeView === 'ai' && (
                <Suspense fallback={<div className="h-64 bg-card border border-border rounded-lg animate-pulse" />}>
                  <AIInsightsPanel shortCode={selectedShortCode} />
                </Suspense>
              )}
            </>
          ) : (
            <div className="bg-card rounded-lg shadow-sm border border-border p-12">
              <div className="text-center">
                <ChartBarIcon className="h-16 w-16 text-muted-foreground mx-auto mb-4" />
                <h3 className="text-xl font-semibold text-foreground mb-2">
                  Select a URL to View Analytics
                </h3>
                <p className="text-muted-foreground">
                  Choose one of your shortened URLs from the dropdown above to see detailed analytics.
                </p>
                {userUrls.length === 0 && (
                  <div className="mt-6">
                    <p className="text-sm text-muted-foreground mb-4">
                      You don't have any URLs yet. Create your first shortened URL to get started.
                    </p>
                    <button
                      onClick={() => router.push('/dashboard')}
                      className="px-6 py-3 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 transition-colors"
                    >
                      Go to Dashboard
                    </button>
                  </div>
                )}
              </div>
            </div>
          )}
        </div>
      </div>
    </Layout>
  );
}

export default function AnalyticsPage() {
  return (
    <Suspense fallback={<div>Loading...</div>}>
      <AnalyticsContent />
    </Suspense>
  );
}