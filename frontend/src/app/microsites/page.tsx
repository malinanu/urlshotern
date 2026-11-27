'use client';

import { Suspense, useCallback, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { ArrowPathIcon, MagnifyingGlassIcon, PlusIcon, SparklesIcon, XMarkIcon } from '@heroicons/react/24/outline';
import WorkspaceLayout from '@/components/layout/WorkspaceLayout';
import { QRCodeModal } from '@/components/QRCode';
import ConfirmDialog from '@/components/workspace/ConfirmDialog';
import { Button, Segmented, Skeleton, buttonClass, cx, focusRing, panelClass } from '@/components/workspace/ui';
import PageCard from '@/components/microsites/list/PageCard';
import PagesStarter from '@/components/microsites/list/PagesStarter';
import { useAuth } from '@/contexts/AuthContext';
import { useRequireAuth } from '@/hooks/useRequireAuth';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  PAGE_KINDS,
  apiFetch,
  isApiError,
  pageKind,
  type Microsite,
  type MicrositeEntitlements,
  type MicrositeSummary,
  type PageKind,
} from '@/lib/microsites';

interface ListResponse {
  microsites: MicrositeSummary[];
  entitlements: MicrositeEntitlements;
  base_url: string;
}

type KindFilter = 'all' | PageKind;
type StatusFilter = 'all' | 'published' | 'draft';
type SortKey = 'edited' | 'views' | 'name';

const SORTS: { value: SortKey; label: string }[] = [
  { value: 'edited', label: 'Last edited' },
  { value: 'views', label: 'Most viewed' },
  { value: 'name', label: 'Name' },
];

export default function MicrositesPage() {
  return (
    <Suspense fallback={null}>
      <PagesList />
    </Suspense>
  );
}

function PagesList() {
  const router = useRouter();
  const params = useSearchParams();
  const { isLoading: authLoading } = useRequireAuth();
  const { user, getAccessToken } = useAuth();

  const [data, setData] = useState<ListResponse | null>(null);
  const [loading, setLoading] = useState(true);
  const [loadError, setLoadError] = useState('');
  const [notice, setNotice] = useState<{ tone: 'ok' | 'error'; text: string; href?: string } | null>(null);

  const [query, setQuery] = useState(() => params.get('q') ?? '');
  const [kind, setKind] = useState<KindFilter>('all');
  const [status, setStatus] = useState<StatusFilter>('all');
  const [sort, setSort] = useState<SortKey>('edited');

  const [qrPage, setQrPage] = useState<MicrositeSummary | null>(null);
  const [deleting, setDeleting] = useState<MicrositeSummary | null>(null);
  const [deleteBusy, setDeleteBusy] = useState(false);
  const [deleteError, setDeleteError] = useState('');

  const load = useCallback(async () => {
    setLoadError('');
    try {
      const list = await apiFetch<ListResponse>(API_ENDPOINTS.microsites.list, getAccessToken());
      setData(list);
    } catch (e) {
      setLoadError(isApiError(e) ? e.message : 'Could not load your pages.');
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    if (user) load();
  }, [user, load]);

  useEffect(() => {
    if (!notice) return;
    const t = setTimeout(() => setNotice(null), 6000);
    return () => clearTimeout(t);
  }, [notice]);

  const pages = useMemo(() => data?.microsites ?? [], [data]);
  const ent = data?.entitlements;
  const baseUrl = data?.base_url ?? '/m/';

  const counts = useMemo(() => {
    const c: Record<KindFilter, number> = { all: pages.length, card: 0, links: 0, shop: 0, event: 0, other: 0 };
    for (const p of pages) c[pageKind(p.template)]++;
    return c;
  }, [pages]);

  const totals = useMemo(
    () => ({
      published: pages.filter((p) => p.status === 'published').length,
      views: pages.reduce((n, p) => n + p.views_30d, 0),
      leads: pages.reduce((n, p) => n + (p.submissions_30d ?? 0), 0),
    }),
    [pages]
  );

  const shown = useMemo(() => {
    const q = query.trim().toLowerCase();
    const list = pages.filter(
      (p) =>
        (kind === 'all' || pageKind(p.template) === kind) &&
        (status === 'all' || p.status === status) &&
        (!q || p.title.toLowerCase().includes(q) || p.slug.includes(q))
    );
    const by: Record<SortKey, (a: MicrositeSummary, b: MicrositeSummary) => number> = {
      edited: (a, b) => b.updated_at.localeCompare(a.updated_at),
      views: (a, b) => b.views_30d - a.views_30d || b.updated_at.localeCompare(a.updated_at),
      name: (a, b) => a.title.localeCompare(b.title, undefined, { sensitivity: 'base' }),
    };
    return [...list].sort(by[sort]);
  }, [pages, kind, status, query, sort]);

  const filtered = query.trim() !== '' || kind !== 'all' || status !== 'all';
  const clearFilters = () => {
    setQuery('');
    setKind('all');
    setStatus('all');
  };

  const duplicate = async (page: MicrositeSummary) => {
    try {
      const copy = await apiFetch<Microsite>(API_ENDPOINTS.microsites.duplicate(page.id), getAccessToken(), { method: 'POST' });
      setNotice({ tone: 'ok', text: `Made a copy at /m/${copy.slug}.`, href: `/microsites/edit?id=${copy.id}` });
      await load();
    } catch (e) {
      setNotice({ tone: 'error', text: isApiError(e) ? e.message : 'Could not duplicate the page.' });
    }
  };

  const togglePublish = async (page: MicrositeSummary) => {
    const publish = page.status !== 'published';
    try {
      await apiFetch(
        publish ? API_ENDPOINTS.microsites.publish(page.id) : API_ENDPOINTS.microsites.unpublish(page.id),
        getAccessToken(),
        { method: 'POST' }
      );
      setNotice({ tone: 'ok', text: publish ? `“${page.title}” is live.` : `“${page.title}” is offline. Its address now shows “not found”.` });
      await load();
    } catch (e) {
      setNotice({
        tone: 'error',
        text: isApiError(e) ? `Could not ${publish ? 'publish' : 'unpublish'}: ${e.message}` : 'Something went wrong. Try again.',
        href: publish ? `/microsites/edit?id=${page.id}` : undefined,
      });
    }
  };

  const confirmDelete = async () => {
    if (!deleting) return;
    setDeleteBusy(true);
    setDeleteError('');
    try {
      await apiFetch(API_ENDPOINTS.microsites.delete(deleting.id), getAccessToken(), { method: 'DELETE' });
      setNotice({ tone: 'ok', text: `Deleted “${deleting.title}”.` });
      setDeleting(null);
      await load();
    } catch (e) {
      setDeleteError(isApiError(e) ? e.message : 'Could not delete the page.');
    } finally {
      setDeleteBusy(false);
    }
  };

  const canCreate = !!ent?.can_create_more;
  const firstRun = !loading && !loadError && pages.length === 0;

  return (
    <WorkspaceLayout
      activeNav="pages"
      pageCount={pages.length || undefined}
      onCreateClick={() => router.push(canCreate || !ent ? '/microsites/new' : '/billing')}
      searchQuery={query}
      onSearchChange={setQuery}
      searchPlaceholder="Search pages by name or address… (⌘K)"
    >
      {/* Header */}
      <div className="mb-6 flex flex-col gap-4 sm:flex-row sm:items-end sm:justify-between">
        <div className="min-w-0">
          <h1 className="font-hanken text-2xl font-bold tracking-tight text-foreground">Pages</h1>
          <p className="mt-1 text-sm text-muted-foreground">
            {loading || authLoading
              ? 'Business cards, link-in-bio, shop and event pages.'
              : pages.length === 0
                ? 'Business cards, link-in-bio, shop and event pages, each with its own link and QR code.'
                : `${pages.length} ${pages.length === 1 ? 'page' : 'pages'} · ${totals.published} published · ${totals.views.toLocaleString()} ${
                    totals.views === 1 ? 'view' : 'views'
                  }${totals.leads ? ` and ${totals.leads} ${totals.leads === 1 ? 'lead' : 'leads'}` : ''} in the last 30 days`}
          </p>
        </div>

        <div className="flex flex-wrap items-center gap-3">
          {ent && <UsageMeter used={ent.used_pages} max={ent.max_pages} />}
          {ent && !canCreate ? (
            <Link href="/billing" className={buttonClass('secondary', 'md')}>
              <SparklesIcon className="h-4 w-4 text-primary" />
              Upgrade for more pages
            </Link>
          ) : (
            <Link href="/microsites/new" className={buttonClass('primary', 'md')} aria-disabled={!ent}>
              <PlusIcon className="h-4 w-4 stroke-[2.5]" />
              New page
            </Link>
          )}
        </div>
      </div>

      {notice && (
        <div
          role={notice.tone === 'error' ? 'alert' : 'status'}
          className={cx(
            'mb-4 flex items-center justify-between gap-3 rounded-xl border px-3 py-2.5 text-sm font-medium',
            notice.tone === 'ok'
              ? 'border-emerald-500/20 bg-emerald-500/10 text-emerald-700 dark:text-emerald-400'
              : 'border-destructive/20 bg-destructive/10 text-destructive'
          )}
        >
          <span className="min-w-0">
            {notice.text}{' '}
            {notice.href && (
              <Link href={notice.href} className="underline underline-offset-2">
                {notice.tone === 'ok' ? 'Open it' : 'Open the editor'}
              </Link>
            )}
          </span>
          <button type="button" onClick={() => setNotice(null)} aria-label="Dismiss" className={cx('rounded-md p-1 opacity-70 hover:opacity-100', focusRing)}>
            <XMarkIcon className="h-4 w-4" />
          </button>
        </div>
      )}

      {loadError ? (
        <div className={cx(panelClass, 'flex flex-col items-start gap-3 p-6')} role="alert">
          <p className="text-sm font-medium text-foreground">{loadError}</p>
          <Button onClick={() => { setLoading(true); load(); }}>
            <ArrowPathIcon className="h-4 w-4" />
            Try again
          </Button>
        </div>
      ) : firstRun ? (
        <PagesStarter getToken={getAccessToken} name={user?.name?.trim() || user?.email?.split('@')[0] || ''} />
      ) : (
        <>
          {/* Toolbar */}
          <div className="mb-4 flex flex-col gap-3 lg:flex-row lg:items-center lg:justify-between">
            <div className="-mx-1 overflow-x-auto px-1 pb-1 lg:pb-0">
              <Segmented<KindFilter>
                label="Page type"
                value={kind}
                onChange={setKind}
                options={[
                  { value: 'all', label: 'All', count: counts.all },
                  ...PAGE_KINDS.map((k) => ({ value: k.id as KindFilter, label: k.plural, count: counts[k.id] })),
                ]}
              />
            </div>
            <div className="flex flex-wrap items-center gap-2">
              <Segmented<StatusFilter>
                label="Status"
                value={status}
                onChange={setStatus}
                options={[
                  { value: 'all', label: 'Any status' },
                  { value: 'published', label: 'Published' },
                  { value: 'draft', label: 'Drafts' },
                ]}
              />
              <label className="sr-only" htmlFor="pages-sort">
                Sort pages
              </label>
              <select
                id="pages-sort"
                value={sort}
                onChange={(e) => setSort(e.target.value as SortKey)}
                className={cx(
                  'h-8 rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 pl-3 pr-8 text-xs font-semibold text-foreground',
                  focusRing
                )}
              >
                {SORTS.map((s) => (
                  <option key={s.value} value={s.value}>
                    {s.label}
                  </option>
                ))}
              </select>
            </div>
          </div>

          {query.trim() && (
            <p className="mb-3 flex items-center gap-2 text-xs text-muted-foreground">
              <MagnifyingGlassIcon className="h-3.5 w-3.5" />
              {shown.length} {shown.length === 1 ? 'page matches' : 'pages match'} “{query.trim()}”
              <button type="button" onClick={() => setQuery('')} className={cx('rounded font-semibold text-primary hover:underline', focusRing)}>
                Clear search
              </button>
            </p>
          )}

          {loading || authLoading ? (
            <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4" aria-busy="true" aria-label="Loading pages">
              {Array.from({ length: 4 }, (_, i) => (
                <li key={i} className={cx(panelClass, 'overflow-hidden')}>
                  <div className="flex h-[232px] justify-center bg-surface-container/60 dark:bg-zinc-900/50 pt-5">
                    <Skeleton className="h-[258px] w-[150px] rounded-[18px]" />
                  </div>
                  <div className="space-y-3 p-4">
                    <Skeleton className="h-3 w-24" />
                    <Skeleton className="h-4 w-3/4" />
                    <Skeleton className="h-3 w-1/2" />
                  </div>
                </li>
              ))}
            </ul>
          ) : shown.length === 0 ? (
            <EmptyFilter kind={kind} filtered={filtered} onClear={clearFilters} canCreate={canCreate} />
          ) : (
            <ul className="grid gap-4 sm:grid-cols-2 xl:grid-cols-3 2xl:grid-cols-4">
              {shown.map((page) => (
                <PageCard
                  key={page.id}
                  page={page}
                  baseUrl={baseUrl}
                  getToken={getAccessToken}
                  canDuplicate={canCreate}
                  actions={{
                    onQR: setQrPage,
                    onDuplicate: duplicate,
                    onTogglePublish: togglePublish,
                    onDelete: (p) => {
                      setDeleteError('');
                      setDeleting(p);
                    },
                  }}
                />
              ))}
            </ul>
          )}
        </>
      )}

      <QRCodeModal
        isOpen={!!qrPage}
        onClose={() => setQrPage(null)}
        value={qrPage ? `${baseUrl}${qrPage.slug}?src=qr` : ''}
        title={qrPage ? `QR code for ${qrPage.title}` : 'QR code'}
        description="Scans are counted as “QR code” traffic in the page’s analytics."
      />

      <ConfirmDialog
        open={!!deleting}
        title={`Delete “${deleting?.title ?? ''}”?`}
        confirmLabel="Delete page"
        busy={deleteBusy}
        error={deleteError}
        onConfirm={confirmDelete}
        onClose={() => setDeleting(null)}
      >
        <p>
          <span className="font-mono text-foreground">{baseUrl.replace(/^https?:\/\//, '')}{deleting?.slug}</span> will stop working, and
          anyone who scans its QR code will see “not found”. The address becomes free for anyone to claim.
        </p>
      </ConfirmDialog>
    </WorkspaceLayout>
  );
}

function UsageMeter({ used, max }: { used: number; max: number }) {
  if (max === -1) {
    return <span className="text-xs font-medium text-muted-foreground">Unlimited pages</span>;
  }
  const pct = Math.min(100, Math.round((used / Math.max(1, max)) * 100));
  const full = used >= max;
  return (
    <div className="flex items-center gap-2.5" title={`${used} of ${max} pages on your plan`}>
      <div className="text-right">
        <div className="text-xs font-semibold tabular-nums text-foreground">
          {used} of {max} pages
        </div>
        <div
          className="mt-1 h-1.5 w-28 overflow-hidden rounded-full bg-surface-container-high dark:bg-zinc-800"
          role="progressbar"
          aria-valuemin={0}
          aria-valuemax={max}
          aria-valuenow={used}
          aria-label="Pages used on your plan"
        >
          <div className={cx('h-full rounded-full', full ? 'bg-amber-500' : 'bg-primary')} style={{ width: `${pct}%` }} />
        </div>
      </div>
    </div>
  );
}

function EmptyFilter({
  kind,
  filtered,
  onClear,
  canCreate,
}: {
  kind: KindFilter;
  filtered: boolean;
  onClear: () => void;
  canCreate: boolean;
}) {
  const info = kind === 'all' ? null : PAGE_KINDS.find((k) => k.id === kind);
  const onlyKind = info && filtered;
  return (
    <div className={cx(panelClass, 'flex flex-col items-center gap-3 px-6 py-14 text-center')}>
      {onlyKind ? (
        <>
          <p className="font-hanken text-base font-bold text-foreground">No {info.plural.toLowerCase()} match</p>
          <p className="max-w-sm text-sm text-muted-foreground">{info.description}</p>
          <div className="mt-1 flex flex-wrap justify-center gap-2">
            {canCreate && (
              <Link href={`/microsites/new?kind=${info.id}`} className={buttonClass('primary', 'sm')}>
                <PlusIcon className="h-4 w-4" />
                New {info.name.toLowerCase()}
              </Link>
            )}
            <Button size="sm" onClick={onClear}>
              Clear filters
            </Button>
          </div>
        </>
      ) : (
        <>
          <p className="font-hanken text-base font-bold text-foreground">No pages match these filters</p>
          <Button size="sm" onClick={onClear}>
            Clear filters
          </Button>
        </>
      )}
    </div>
  );
}
