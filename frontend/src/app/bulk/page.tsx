'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';
import Layout from '@/components/layout/Layout';
import BulkURLShortener from '@/components/BulkURLShortener';
import { useAuth } from '@/contexts/AuthContext';
import { DocumentDuplicateIcon } from '@heroicons/react/24/outline';

export default function BulkPage() {
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
              <DocumentDuplicateIcon className="h-8 w-8 text-primary" />
              <h1 className="text-3xl font-bold text-foreground">Bulk Operations</h1>
            </div>
            <p className="text-muted-foreground">
              Create multiple short URLs at once by importing from a file or entering them manually.
            </p>
          </div>

          <BulkURLShortener />

          {/* Help Section */}
          <div className="mt-8 bg-card rounded-lg shadow-sm border border-border p-6">
            <h2 className="text-lg font-semibold text-foreground mb-4">How to Use Bulk Operations</h2>

            <div className="grid grid-cols-1 md:grid-cols-2 gap-6">
              <div>
                <h3 className="font-medium text-foreground mb-2">Import Methods</h3>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• <strong>File Import:</strong> Upload a .txt or .csv file with URLs (one per line)</li>
                  <li>• <strong>Paste URLs:</strong> Copy and paste multiple URLs at once</li>
                  <li>• <strong>Manual Entry:</strong> Add URLs one by one using the form</li>
                </ul>
              </div>

              <div>
                <h3 className="font-medium text-foreground mb-2">Features</h3>
                <ul className="text-sm text-muted-foreground space-y-1">
                  <li>• Custom short codes for each URL</li>
                  <li>• Export results as CSV</li>
                  <li>• Copy all short URLs at once</li>
                  <li>• Real-time validation and error reporting</li>
                </ul>
              </div>
            </div>

            <div className="mt-6 p-4 bg-primary/5 rounded-lg border border-primary/20">
              <h4 className="font-medium text-primary mb-2">Pro Tip</h4>
              <p className="text-sm text-foreground/80">
                For best results, ensure your URLs include the protocol (http:// or https://).
                The system can process up to 100 URLs per batch.
              </p>
            </div>
          </div>
        </div>
      </div>
    </Layout>
  );
}