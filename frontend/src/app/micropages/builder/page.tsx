'use client';

import React, { Suspense } from 'react';
import { useSearchParams } from 'next/navigation';
import MicroPageBuilderClient from '../[id]/builder/_client';

function BuilderContent() {
  const searchParams = useSearchParams();
  const id = searchParams.get('id');

  if (!id) {
    return (
      <div className="min-h-screen bg-slate-950 flex flex-col items-center justify-center text-white px-4">
        <h2 className="text-xl font-bold mb-2">Missing Micro-Page ID</h2>
        <p className="text-slate-400 text-sm mb-4">Please select a micro-page from the dashboard to edit.</p>
        <a
          href="/micropages"
          className="px-4 py-2 rounded-xl bg-blue-600 hover:bg-blue-500 text-white font-medium text-sm transition-colors"
        >
          Return to Micro-Pages
        </a>
      </div>
    );
  }

  return <MicroPageBuilderClient id={id} />;
}

export default function MicroPageBuilderPage() {
  return (
    <Suspense
      fallback={
        <div className="min-h-screen bg-slate-950 flex items-center justify-center">
          <div className="w-8 h-8 border-2 border-blue-500 border-t-transparent rounded-full animate-spin" />
        </div>
      }
    >
      <BuilderContent />
    </Suspense>
  );
}
