'use client';

import { Suspense } from 'react';
import { useRouter } from 'next/navigation';
import WorkspaceLayout from '@/components/layout/WorkspaceLayout';
import MicrositeEditor from '@/components/microsites/MicrositeEditor';
import { useRequireAuth } from '@/hooks/useRequireAuth';

// The page ID comes from the query string (?id=) rather than a dynamic path
// segment, so this route exports as a single static HTML file.
export default function MicrositeEditPage() {
  const router = useRouter();
  const { isAuthenticated, isLoading } = useRequireAuth();
  const loading = <div className="min-h-[60vh] flex items-center justify-center text-sm text-muted-foreground">Loading…</div>;

  return (
    <WorkspaceLayout
      activeNav="pages"
      wide
      collapsedNav
      onCreateClick={() => router.push('/microsites/new')}
      onSearchSubmit={(q) => router.push(q ? `/microsites?q=${encodeURIComponent(q)}` : '/microsites')}
      searchPlaceholder="Search pages… (⌘K)"
    >
      {isLoading || !isAuthenticated ? loading : <Suspense fallback={loading}>{<MicrositeEditor />}</Suspense>}
    </WorkspaceLayout>
  );
}
