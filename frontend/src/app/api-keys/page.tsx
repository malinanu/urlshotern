'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import APIKeyManager from '@/components/api-keys/APIKeyManager';
import { useAuth } from '@/contexts/AuthContext';
import { KeyIcon } from '@heroicons/react/24/outline';

export default function APIKeysPage() {
  const { isAuthenticated, isLoading } = useAuth();
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
        <div className="flex justify-center items-center min-h-[60vh]">
          <div className="animate-spin rounded-full h-12 w-12 border-b-2 border-primary-600"></div>
        </div>
      </Layout>
    );
  }

  if (!isAuthenticated) {
    return null;
  }

  return (
    <Layout>
      <div className="max-w-7xl mx-auto px-4 sm:px-6 lg:px-8 py-8">
        {/* Page Header */}
        <div className="mb-8">
          <div className="flex items-center gap-3 mb-2">
            <KeyIcon className="h-8 w-8 text-primary-600" />
            <h1 className="text-3xl font-bold text-gray-900">API Keys</h1>
          </div>
          <p className="text-gray-600 mt-2">
            Create and manage API keys to integrate Trunc URL shortener into your applications.
            Use these keys to programmatically shorten URLs, track analytics, and more.
          </p>
        </div>

        {/* Integration Guide */}
        <div className="bg-blue-50 border border-blue-200 rounded-lg p-4 mb-6">
          <h3 className="text-sm font-semibold text-blue-900 mb-2">
            How to use API Keys
          </h3>
          <ol className="text-sm text-blue-800 space-y-1 list-decimal list-inside">
            <li>Click "Create New API Key" below to generate a new key</li>
            <li>Copy the key (it will only be shown once!)</li>
            <li>Use it in your application's Authorization header: <code className="bg-blue-100 px-1 rounded">Bearer YOUR_API_KEY</code></li>
            <li>Check our <a href="/api-docs" className="underline font-medium">API Documentation</a> for integration examples</li>
          </ol>
        </div>

        {/* API Key Manager Component */}
        <APIKeyManager />
      </div>
    </Layout>
  );
}
