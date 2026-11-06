'use client';

import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { ArrowDownTrayIcon, InboxIcon, LockClosedIcon, TrashIcon } from '@heroicons/react/24/outline';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import { apiFetch, isApiError, type MicrositeBlock, type MicrositeSubmission } from '@/lib/microsites';

const PAGE = 50;

interface Props {
  micrositeId: number;
  blocks: MicrositeBlock[];
  canUseForms: boolean;
}

/** Newsletter signups and contact messages sent from the page's forms. */
export default function InboxPanel({ micrositeId, blocks, canUseForms }: Props) {
  const { getAccessToken } = useAuth();
  const [items, setItems] = useState<MicrositeSubmission[]>([]);
  const [total, setTotal] = useState(0);
  const [offset, setOffset] = useState(0);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [exporting, setExporting] = useState(false);

  const formNames = Object.fromEntries(
    blocks.filter((b) => b.type === 'form').map((b) => [b.id, b.data?.title || (b.data?.kind === 'contact' ? 'Contact form' : 'Newsletter')])
  );
  const hasForms = blocks.some((b) => b.type === 'form');

  const load = useCallback(
    async (from: number) => {
      setLoading(true);
      setError('');
      try {
        const res = await apiFetch<{ submissions: MicrositeSubmission[]; total: number }>(
          `${API_ENDPOINTS.microsites.submissions(micrositeId)}?limit=${PAGE}&offset=${from}`,
          getAccessToken()
        );
        setItems(res.submissions);
        setTotal(res.total);
        setOffset(from);
      } catch (e) {
        setError(isApiError(e) ? e.message : 'Could not load submissions.');
      } finally {
        setLoading(false);
      }
    },
    [micrositeId, getAccessToken]
  );

  useEffect(() => {
    load(0);
  }, [load]);

  const remove = async (s: MicrositeSubmission) => {
    if (!confirm(`Delete the submission from ${s.email}? This can't be undone.`)) return;
    try {
      await apiFetch(API_ENDPOINTS.microsites.deleteSubmission(micrositeId, s.id), getAccessToken(), { method: 'DELETE' });
      const from = items.length === 1 && offset > 0 ? offset - PAGE : offset;
      await load(from);
    } catch (e) {
      setError(isApiError(e) ? e.message : 'Could not delete the submission.');
    }
  };

  const exportCSV = async () => {
    setExporting(true);
    setError('');
    try {
      const token = getAccessToken();
      const res = await fetch(API_ENDPOINTS.microsites.exportSubmissions(micrositeId), {
        headers: token ? { Authorization: `Bearer ${token}` } : {},
      });
      if (!res.ok) throw new Error(`Export failed (${res.status})`);
      const url = URL.createObjectURL(await res.blob());
      const a = document.createElement('a');
      a.href = url;
      a.download = `page-${micrositeId}-submissions.csv`;
      document.body.appendChild(a);
      a.click();
      a.remove();
      URL.revokeObjectURL(url);
    } catch (e) {
      setError(e instanceof Error ? e.message : 'Could not export.');
    } finally {
      setExporting(false);
    }
  };

  return (
    <div className="max-w-4xl grid gap-4">
      <div className="flex flex-wrap items-center gap-3">
        <div className="flex-1 min-w-0">
          <h2 className="text-lg font-semibold text-foreground">Inbox</h2>
          <p className="text-sm text-muted-foreground">
            {total === 0 ? 'Signups and messages from your forms appear here.' : `${total.toLocaleString()} submission${total === 1 ? '' : 's'}`}
          </p>
        </div>
        {total > 0 && (
          <button
            type="button"
            onClick={exportCSV}
            disabled={exporting}
            className="inline-flex items-center gap-1.5 rounded-lg border border-border px-3 py-2 text-sm font-medium text-foreground hover:bg-muted disabled:opacity-50"
          >
            <ArrowDownTrayIcon className="h-4 w-4" />
            {exporting ? 'Exporting…' : 'Export CSV'}
          </button>
        )}
      </div>

      {!canUseForms && (
        <div className="flex flex-col sm:flex-row sm:items-center gap-3 rounded-xl border border-border bg-card px-5 py-4">
          <LockClosedIcon className="h-5 w-5 text-primary shrink-0" />
          <p className="text-sm text-muted-foreground flex-1">
            Newsletter and contact forms are part of the Pro and Ultra plans.
            {total > 0 && ' Your earlier submissions stay here to read and export.'}
          </p>
          <Link href="/billing" className="text-sm font-medium text-primary hover:underline">
            Upgrade plan
          </Link>
        </div>
      )}

      {error && (
        <p role="alert" className="rounded-lg border border-red-300 bg-red-50 dark:bg-red-950/30 px-4 py-2 text-sm text-red-700 dark:text-red-300">
          {error}
        </p>
      )}

      {loading && items.length === 0 ? (
        <div className="grid gap-2">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-20 rounded-lg border border-border bg-card animate-pulse" />
          ))}
        </div>
      ) : items.length === 0 ? (
        <div className="rounded-xl border border-border bg-card p-10 text-center">
          <InboxIcon className="mx-auto h-8 w-8 text-muted-foreground" />
          <p className="mt-3 text-sm text-foreground font-medium">No submissions yet</p>
          <p className="mt-1 text-sm text-muted-foreground">
            {hasForms
              ? 'Share your page; new signups and messages land here and in your email.'
              : canUseForms
                ? 'Add a Newsletter or contact form block in Build to start collecting.'
                : 'Upgrade to add a form to your page.'}
          </p>
        </div>
      ) : (
        <ul className="grid gap-2" aria-label="Submissions">
          {items.map((s) => (
            <li key={s.id} className="rounded-lg border border-border bg-card px-4 py-3">
              <div className="flex flex-wrap items-baseline gap-x-3 gap-y-1">
                <a href={`mailto:${s.email}`} className="text-sm font-medium text-foreground hover:underline break-all">
                  {s.email}
                </a>
                {s.name && <span className="text-sm text-muted-foreground">{s.name}</span>}
                <span className="ml-auto text-xs text-muted-foreground">
                  <time dateTime={s.created_at}>{new Date(s.created_at).toLocaleString()}</time>
                </span>
                <button
                  type="button"
                  onClick={() => remove(s)}
                  className="p-1 text-muted-foreground hover:text-red-600"
                  aria-label={`Delete submission from ${s.email}`}
                >
                  <TrashIcon className="h-4 w-4" />
                </button>
              </div>
              {s.message && <p className="mt-2 text-sm text-foreground whitespace-pre-line break-words">{s.message}</p>}
              <p className="mt-1 text-xs text-muted-foreground">
                {s.kind === 'contact' ? 'Message' : 'Newsletter signup'} · {formNames[s.block_id] ?? 'Removed form'}
              </p>
            </li>
          ))}
        </ul>
      )}

      {total > PAGE && (
        <div className="flex items-center justify-between text-sm">
          <button
            type="button"
            onClick={() => load(Math.max(0, offset - PAGE))}
            disabled={offset === 0 || loading}
            className="rounded-lg border border-border px-3 py-1.5 text-foreground hover:bg-muted disabled:opacity-40"
          >
            Newer
          </button>
          <span className="text-muted-foreground">
            {offset + 1}–{Math.min(offset + PAGE, total)} of {total.toLocaleString()}
          </span>
          <button
            type="button"
            onClick={() => load(offset + PAGE)}
            disabled={offset + PAGE >= total || loading}
            className="rounded-lg border border-border px-3 py-1.5 text-foreground hover:bg-muted disabled:opacity-40"
          >
            Older
          </button>
        </div>
      )}
    </div>
  );
}
