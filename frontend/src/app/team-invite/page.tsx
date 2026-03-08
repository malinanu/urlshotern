'use client';

import { useState, useEffect, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { UserGroupIcon, CheckCircleIcon, XCircleIcon } from '@heroicons/react/24/outline';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

function TeamInviteContent() {
  const { isAuthenticated, isLoading, getAccessToken } = useAuth();
  const router = useRouter();
  const searchParams = useSearchParams();
  const token = searchParams.get('token');

  const [status, setStatus] = useState<'idle' | 'loading' | 'accepted' | 'rejected' | 'error'>('idle');
  const [errorMsg, setErrorMsg] = useState('');

  // Redirect to login if not authenticated, preserving the invite URL
  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      const returnTo = encodeURIComponent(`/team-invite?token=${token}`);
      router.push(`/login?returnTo=${returnTo}`);
    }
  }, [isAuthenticated, isLoading, token, router]);

  if (!token) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center py-12 px-4">
          <div className="max-w-md w-full text-center bg-card border border-border rounded-xl p-8">
            <XCircleIcon className="h-12 w-12 text-destructive mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-foreground mb-2">Invalid Invitation</h1>
            <p className="text-muted-foreground mb-6">This invitation link is missing or invalid.</p>
            <Link href="/teams" className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors">
              Go to Teams
            </Link>
          </div>
        </div>
      </Layout>
    );
  }

  if (isLoading || !isAuthenticated) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
        </div>
      </Layout>
    );
  }

  const handleAction = async (action: 'accept' | 'reject') => {
    setStatus('loading');
    setErrorMsg('');
    try {
      const authToken = getAccessToken();
      const res = await fetch(`${API_BASE_URL}/api/v1/teams/invitations/${action}`, {
        method: 'POST',
        headers: {
          Authorization: `Bearer ${authToken}`,
          'Content-Type': 'application/json',
        },
        body: JSON.stringify({ token }),
      });

      const data = await res.json();

      if (!res.ok) {
        setErrorMsg(data.message || `Failed to ${action} invitation`);
        setStatus('error');
        return;
      }

      setStatus(action === 'accept' ? 'accepted' : 'rejected');
    } catch {
      setErrorMsg('Network error. Please try again.');
      setStatus('error');
    }
  };

  if (status === 'accepted') {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center py-12 px-4">
          <div className="max-w-md w-full text-center bg-card border border-border rounded-xl p-8">
            <CheckCircleIcon className="h-12 w-12 text-green-500 mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-foreground mb-2">You joined the team!</h1>
            <p className="text-muted-foreground mb-6">You can now collaborate with your team members.</p>
            <Link href="/teams" className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors">
              View Teams
            </Link>
          </div>
        </div>
      </Layout>
    );
  }

  if (status === 'rejected') {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center py-12 px-4">
          <div className="max-w-md w-full text-center bg-card border border-border rounded-xl p-8">
            <XCircleIcon className="h-12 w-12 text-muted-foreground mx-auto mb-4" />
            <h1 className="text-xl font-semibold text-foreground mb-2">Invitation declined</h1>
            <p className="text-muted-foreground mb-6">You&apos;ve declined the team invitation.</p>
            <Link href="/dashboard" className="px-4 py-2 bg-primary text-primary-foreground rounded-lg text-sm font-medium hover:bg-primary/90 transition-colors">
              Go to Dashboard
            </Link>
          </div>
        </div>
      </Layout>
    );
  }

  return (
    <Layout>
      <div className="min-h-screen bg-background flex items-center justify-center py-12 px-4">
        <div className="max-w-md w-full bg-card border border-border rounded-xl p-8">
          <div className="flex justify-center mb-5">
            <div className="p-3 bg-primary/10 rounded-full">
              <UserGroupIcon className="h-8 w-8 text-primary" />
            </div>
          </div>

          <h1 className="text-xl font-semibold text-foreground text-center mb-2">
            You&apos;ve been invited to a team
          </h1>
          <p className="text-sm text-muted-foreground text-center mb-8">
            Accept the invitation below to start collaborating with your team.
          </p>

          {(status === 'error') && (
            <div className="mb-4 p-3 bg-destructive/10 border border-destructive/20 rounded-lg text-sm text-destructive text-center">
              {errorMsg}
            </div>
          )}

          <div className="flex gap-3">
            <button
              onClick={() => handleAction('reject')}
              disabled={status === 'loading'}
              className="flex-1 px-4 py-2.5 text-sm font-medium text-foreground border border-border rounded-lg hover:bg-muted disabled:opacity-50 transition-colors"
            >
              Decline
            </button>
            <button
              onClick={() => handleAction('accept')}
              disabled={status === 'loading'}
              className="flex-1 px-4 py-2.5 text-sm font-medium text-primary-foreground bg-primary rounded-lg hover:bg-primary/90 disabled:opacity-50 transition-colors"
            >
              {status === 'loading' ? 'Processing…' : 'Accept Invitation'}
            </button>
          </div>

          <p className="text-xs text-muted-foreground text-center mt-4">
            Invitations expire after 7 days.
          </p>
        </div>
      </div>
    </Layout>
  );
}

export default function TeamInvitePage() {
  return (
    <Suspense fallback={
      <div className="min-h-screen bg-background flex items-center justify-center">
        <div className="animate-spin rounded-full h-8 w-8 border-b-2 border-primary" />
      </div>
    }>
      <TeamInviteContent />
    </Suspense>
  );
}
