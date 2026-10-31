'use client';

import { Suspense } from 'react';
import Layout from '@/components/layout/Layout';
import MicrositeEditor from '@/components/microsites/MicrositeEditor';
import { useRequireAuth } from '@/hooks/useRequireAuth';

// The page ID comes from the query string (?id=) rather than a dynamic path
// segment, so this route exports as a single static HTML file.
export default function MicrositeEditPage() {
  const { isAuthenticated, isLoading } = useRequireAuth();

  return (
    <Layout>
      {isLoading || !isAuthenticated ? (
        <div className="min-h-[60vh] flex items-center justify-center text-muted-foreground">Loading…</div>
      ) : (
        <Suspense fallback={<div className="min-h-[60vh] flex items-center justify-center text-muted-foreground">Loading…</div>}>
          <MicrositeEditor />
        </Suspense>
      )}
    </Layout>
  );
}
