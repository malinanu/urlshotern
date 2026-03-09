'use client';

import { useEffect, useState, Suspense } from 'react';
import { useRouter, useSearchParams } from 'next/navigation';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { CheckCircleIcon, ExclamationTriangleIcon, EnvelopeIcon } from '@heroicons/react/24/outline';

function VerifyEmailContent() {
  const searchParams = useSearchParams();
  const router = useRouter();
  const { user, getAccessToken, logout } = useAuth();
  const token = searchParams.get('token');

  // Token mode state
  const [verifying, setVerifying] = useState(false);
  const [verifyResult, setVerifyResult] = useState<'success' | 'error' | null>(null);
  const [verifyError, setVerifyError] = useState<string | null>(null);

  // Resend mode state
  const [resendState, setResendState] = useState<'idle' | 'sending' | 'sent' | 'error'>('idle');
  const [resendError, setResendError] = useState<string | null>(null);

  // If a token is in the URL, verify it on mount
  useEffect(() => {
    if (!token) return;

    setVerifying(true);
    fetch(`${API_ENDPOINTS.auth.verifyEmail}?token=${encodeURIComponent(token)}`)
      .then(async (res) => {
        if (res.ok) {
          setVerifyResult('success');
          setTimeout(() => router.push('/'), 2000);
        } else {
          const data = await res.json().catch(() => ({}));
          setVerifyError(data.error || 'This verification link is invalid or has expired.');
          setVerifyResult('error');
        }
      })
      .catch(() => {
        setVerifyError('Network error. Please try again.');
        setVerifyResult('error');
      })
      .finally(() => setVerifying(false));
  }, [token, router]);

  const handleResend = async () => {
    setResendState('sending');
    setResendError(null);
    try {
      const res = await fetch(API_ENDPOINTS.auth.resendEmail, {
        method: 'POST',
        headers: { Authorization: `Bearer ${getAccessToken()}` },
      });
      if (res.ok) {
        setResendState('sent');
      } else {
        const data = await res.json().catch(() => ({}));
        setResendError(data.error || 'Failed to send email. Please try again.');
        setResendState('error');
      }
    } catch {
      setResendError('Network error. Please try again.');
      setResendState('error');
    }
  };

  const handleLogout = () => {
    logout();
    router.push('/login');
  };

  // ── Token mode ──────────────────────────────────────────────────────────────
  if (token) {
    return (
      <Layout>
        <div className="min-h-[80vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
          <div className="sm:mx-auto sm:w-full sm:max-w-md">
            <div className="bg-card py-10 px-6 border border-border sm:rounded-lg text-center space-y-4">
              {verifying && (
                <>
                  <div className="flex justify-center">
                    <div className="h-12 w-12 animate-spin rounded-full border-4 border-primary border-t-transparent" />
                  </div>
                  <p className="text-muted-foreground">Verifying your email address…</p>
                </>
              )}

              {verifyResult === 'success' && (
                <>
                  <CheckCircleIcon className="h-14 w-14 text-green-500 mx-auto" />
                  <h2 className="text-xl font-bold text-foreground">Email verified!</h2>
                  <p className="text-muted-foreground">Redirecting you to the dashboard…</p>
                </>
              )}

              {verifyResult === 'error' && (
                <>
                  <ExclamationTriangleIcon className="h-14 w-14 text-yellow-500 mx-auto" />
                  <h2 className="text-xl font-bold text-foreground">Verification failed</h2>
                  <p className="text-muted-foreground">{verifyError}</p>
                  {user && (
                    <button
                      onClick={handleResend}
                      disabled={resendState === 'sending' || resendState === 'sent'}
                      className="mt-2 w-full py-2 px-4 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 disabled:opacity-50"
                    >
                      {resendState === 'sending' ? 'Sending…' : resendState === 'sent' ? 'Email sent!' : 'Resend verification email'}
                    </button>
                  )}
                  {!user && (
                    <Link href="/login" className="text-primary hover:underline text-sm">
                      Sign in to resend verification email
                    </Link>
                  )}
                </>
              )}
            </div>
          </div>
        </div>
      </Layout>
    );
  }

  // ── No-token mode (check your inbox) ────────────────────────────────────────
  return (
    <Layout>
      <div className="min-h-[80vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-card py-10 px-6 border border-border sm:rounded-lg text-center space-y-4">
            <EnvelopeIcon className="h-14 w-14 text-primary mx-auto" />
            <h2 className="text-2xl font-bold text-foreground">Check your inbox</h2>
            <p className="text-muted-foreground">
              We sent a verification email to{' '}
              <span className="font-medium text-foreground">{user?.email ?? 'your email address'}</span>.
              Click the link in the email to activate your account.
            </p>

            {resendState === 'sent' ? (
              <div className="flex items-center justify-center space-x-2 text-green-600 dark:text-green-400">
                <CheckCircleIcon className="h-5 w-5" />
                <span className="text-sm font-medium">Verification email sent!</span>
              </div>
            ) : (
              <>
                {resendError && (
                  <p className="text-sm text-destructive">{resendError}</p>
                )}
                <button
                  onClick={handleResend}
                  disabled={resendState === 'sending'}
                  className="w-full py-2 px-4 rounded-md bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 disabled:opacity-50"
                >
                  {resendState === 'sending' ? 'Sending…' : 'Resend verification email'}
                </button>
              </>
            )}

            <button
              onClick={handleLogout}
              className="w-full py-2 px-4 rounded-md border border-border text-sm font-medium text-foreground hover:bg-muted"
            >
              Sign out
            </button>
          </div>
        </div>
      </div>
    </Layout>
  );
}

export default function VerifyEmailPage() {
  return (
    <Suspense>
      <VerifyEmailContent />
    </Suspense>
  );
}
