'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { CheckIcon, ExclamationTriangleIcon, PuzzlePieceIcon } from '@heroicons/react/24/outline';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';

const API_BASE_URL = process.env.NEXT_PUBLIC_API_URL || 'http://localhost:8080';

interface ClientInfo {
  client_id: string;
  name: string;
  host: string;
  any_host: boolean;
}

const PARAMS = ['client_id', 'redirect_uri', 'response_type', 'state', 'scope', 'code_challenge', 'code_challenge_method'] as const;

/**
 * "Sign in with Trunc" consent page. A plugin (Chrome extension, WordPress,
 * Shopify, Canva) sends the user here; Allow returns them to the plugin with a
 * one-time code that it exchanges for its own API key.
 */
function ConnectContent() {
  const params = useSearchParams();
  const router = useRouter();
  const { user, isAuthenticated, isLoading, getAccessToken } = useAuth();
  const [info, setInfo] = useState<ClientInfo | null>(null);
  const [problem, setProblem] = useState<string | null>(null);
  const [busy, setBusy] = useState<'allow' | 'deny' | null>(null);

  const request = useMemo(() => {
    const r: Record<string, string> = {};
    PARAMS.forEach((k) => (r[k] = params.get(k) || ''));
    return r;
  }, [params]);

  useEffect(() => {
    if (!isLoading && !isAuthenticated) {
      const back = `/connect?${params.toString()}`;
      router.replace(`/login?returnTo=${encodeURIComponent(back)}`);
    }
  }, [isLoading, isAuthenticated, params, router]);

  useEffect(() => {
    if (!request.client_id || !request.redirect_uri) {
      setProblem('This connection link is incomplete. Start again from the app you are connecting.');
      return;
    }
    const q = new URLSearchParams({ client_id: request.client_id, redirect_uri: request.redirect_uri });
    fetch(`${API_BASE_URL}/api/v1/oauth/client?${q}`)
      .then(async (r) => {
        const body = await r.json().catch(() => ({}));
        if (!r.ok) throw new Error(body.error_description || 'This app could not be verified.');
        setInfo(body as ClientInfo);
      })
      .catch((e: Error) => setProblem(e.message));
  }, [request.client_id, request.redirect_uri]);

  const decide = async (decision: 'allow' | 'deny') => {
    setBusy(decision);
    try {
      const res = await fetch(`${API_BASE_URL}/api/v1/oauth/authorize`, {
        method: 'POST',
        headers: { 'Content-Type': 'application/json', Authorization: `Bearer ${getAccessToken() || ''}` },
        body: JSON.stringify({ ...request, decision }),
      });
      const body = await res.json().catch(() => ({}));
      if (!res.ok || !body.redirect) throw new Error(body.error_description || 'Could not complete the connection.');
      window.location.assign(body.redirect);
    } catch (e) {
      setProblem((e as Error).message);
      setBusy(null);
    }
  };

  if (isLoading || !isAuthenticated) {
    return (
      <div className="flex min-h-[60vh] items-center justify-center">
        <div className="h-10 w-10 animate-spin rounded-full border-b-2 border-primary" />
      </div>
    );
  }

  return (
    <div className="mx-auto flex min-h-[70vh] max-w-md items-center px-4 py-12">
      <div className="w-full rounded-2xl border border-border bg-card p-6 shadow-sm sm:p-8">
        {problem ? (
          <div role="alert">
            <ExclamationTriangleIcon className="h-8 w-8 text-amber-500" aria-hidden />
            <h1 className="mt-3 text-xl font-bold text-foreground">Can&apos;t connect this app</h1>
            <p className="mt-2 text-sm text-muted-foreground">{problem}</p>
            <Link href="/dashboard" className="mt-6 inline-block text-sm font-semibold text-primary hover:underline">
              Go to your dashboard
            </Link>
          </div>
        ) : !info ? (
          <div className="flex justify-center py-10">
            <div className="h-8 w-8 animate-spin rounded-full border-b-2 border-primary" />
          </div>
        ) : (
          <>
            <span className="grid h-12 w-12 place-items-center rounded-xl bg-primary/10 text-primary">
              <PuzzlePieceIcon className="h-6 w-6" aria-hidden />
            </span>
            <h1 className="mt-4 text-xl font-bold text-foreground">Connect {info.name}</h1>
            <p className="mt-2 text-sm text-muted-foreground">
              <strong className="text-foreground">{info.name}</strong> wants to use your Trunc account
              {user?.email ? (
                <>
                  {' '}
                  (<span className="text-foreground">{user.email}</span>)
                </>
              ) : null}
              .
            </p>

            <ul className="mt-5 space-y-2.5 text-sm text-foreground">
              {['Create short links', 'See your links and their click counts', 'Edit and delete your links'].map((t) => (
                <li key={t} className="flex items-start gap-2">
                  <CheckIcon className="mt-0.5 h-4 w-4 shrink-0 text-emerald-600 dark:text-emerald-400" aria-hidden />
                  {t}
                </li>
              ))}
            </ul>

            <div
              className={
                info.any_host
                  ? 'mt-5 rounded-xl border border-amber-500/30 bg-amber-500/10 p-3 text-sm text-foreground'
                  : 'mt-5 rounded-xl border border-border bg-muted/40 p-3 text-sm text-muted-foreground'
              }
            >
              You will go back to <strong className="break-all text-foreground">{info.host}</strong>.
              {info.any_host ? ' Only continue if this is your site.' : ''}
            </div>

            <p className="mt-4 text-xs text-muted-foreground">
              The app gets its own API key. You can disconnect it any time on{' '}
              <Link href="/api-keys" className="font-semibold text-primary hover:underline">
                API keys
              </Link>
              .
            </p>

            <div className="mt-6 flex flex-col-reverse gap-3 sm:flex-row">
              <button
                type="button"
                disabled={busy !== null}
                onClick={() => decide('deny')}
                className="h-11 flex-1 rounded-xl border border-border bg-background text-sm font-semibold text-foreground hover:bg-muted disabled:opacity-60"
              >
                {busy === 'deny' ? 'Cancelling…' : 'Cancel'}
              </button>
              <button
                type="button"
                disabled={busy !== null}
                onClick={() => decide('allow')}
                className="h-11 flex-1 rounded-xl bg-primary text-sm font-semibold text-primary-foreground hover:bg-primary/90 disabled:opacity-60"
              >
                {busy === 'allow' ? 'Connecting…' : 'Allow'}
              </button>
            </div>
          </>
        )}
      </div>
    </div>
  );
}

export default function ConnectPage() {
  return (
    <Layout>
      <Suspense fallback={null}>
        <ConnectContent />
      </Suspense>
    </Layout>
  );
}
