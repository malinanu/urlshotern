'use client';

import { useEffect, useState } from 'react';
import Link from 'next/link';
import Layout from '@/components/layout/Layout';
import { API_ENDPOINTS } from '@/lib/api-config';
import { CheckCircleIcon, ExclamationCircleIcon } from '@heroicons/react/24/outline';

type Status = 'ready' | 'working' | 'done' | 'error';

/**
 * Landing page for the unsubscribe link in every email footer (CAN-SPAM).
 * Unsubscribing takes one click and no login. It's a button rather than an
 * automatic request so mail-scanning bots that open links don't unsubscribe people.
 */
export default function UnsubscribePage() {
  const [token, setToken] = useState<string | null>(null);
  const [status, setStatus] = useState<Status>('ready');
  const [message, setMessage] = useState('');

  useEffect(() => {
    setToken(new URLSearchParams(window.location.search).get('token'));
  }, []);

  const unsubscribe = async () => {
    if (!token) return;
    setStatus('working');
    try {
      const res = await fetch(API_ENDPOINTS.legal.unsubscribe, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json' },
        body: JSON.stringify({ token }),
      });
      const data = await res.json().catch(() => ({}));
      if (!res.ok) {
        setStatus('error');
        setMessage(data.error || 'We could not process your request. Please try again.');
        return;
      }
      setStatus('done');
      setMessage(data.email ? `${data.email} won't receive Trunc marketing emails anymore.` : '');
    } catch {
      setStatus('error');
      setMessage('We could not reach the server. Please try again in a moment.');
    }
  };

  return (
    <Layout>
      <div className="min-h-[70vh] flex flex-col justify-center py-12 sm:px-6 lg:px-8">
        <div className="sm:mx-auto sm:w-full sm:max-w-md">
          <div className="bg-card py-8 px-4 border border-border sm:rounded-lg sm:px-10 text-center">
            {status === 'done' ? (
              <>
                <CheckCircleIcon className="h-14 w-14 text-green-600 dark:text-green-500 mx-auto mb-4" />
                <h1 className="text-2xl font-bold text-foreground mb-3">You&apos;re unsubscribed</h1>
                <p className="text-muted-foreground mb-2">{message}</p>
                <p className="text-sm text-muted-foreground mb-6">
                  You&apos;ll still get emails you need for your account, such as password resets and security
                  codes.
                </p>
              </>
            ) : !token ? (
              <>
                <ExclamationCircleIcon className="h-14 w-14 text-muted-foreground mx-auto mb-4" />
                <h1 className="text-2xl font-bold text-foreground mb-3">Unsubscribe link incomplete</h1>
                <p className="text-muted-foreground mb-6">
                  Please use the full unsubscribe link from the bottom of one of our emails, or turn off
                  marketing emails in your account settings.
                </p>
              </>
            ) : (
              <>
                <h1 className="text-2xl font-bold text-foreground mb-3">Unsubscribe from marketing emails</h1>
                <p className="text-muted-foreground mb-6">
                  Stop receiving promotional emails from Trunc. Account and security emails will still be sent.
                </p>
                <button
                  type="button"
                  onClick={unsubscribe}
                  disabled={status === 'working'}
                  className="w-full py-2 px-4 rounded-md text-sm font-medium bg-primary text-primary-foreground hover:bg-primary/90 disabled:opacity-50"
                >
                  {status === 'working' ? 'Unsubscribing…' : 'Unsubscribe'}
                </button>
                {status === 'error' && <p className="mt-4 text-sm text-destructive">{message}</p>}
              </>
            )}
            <Link href="/" className="mt-6 inline-block text-sm text-primary hover:text-primary/80">
              Back to Trunc
            </Link>
          </div>
        </div>
      </div>
    </Layout>
  );
}
