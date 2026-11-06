'use client';

import { useEffect } from 'react';
import { useRouter } from 'next/navigation';

export default function PagesIndexRedirect() {
  const router = useRouter();

  useEffect(() => {
    router.replace('/microsites');
  }, [router]);

  return (
    <div className="min-h-screen flex items-center justify-center bg-surface-container-lowest text-on-surface">
      <div className="flex flex-col items-center gap-3">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary"></div>
        <span className="text-xs font-mono text-muted-foreground">Redirecting to Pages...</span>
      </div>
    </div>
  );
}
