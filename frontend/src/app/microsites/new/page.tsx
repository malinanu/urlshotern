'use client';

import { Suspense, useEffect, useMemo, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { CheckCircleIcon, ChevronRightIcon, ExclamationCircleIcon, SparklesIcon } from '@heroicons/react/24/outline';
import WorkspaceLayout from '@/components/layout/WorkspaceLayout';
import TemplateGallery from '@/components/micropages/TemplateGallery';
import PhoneFrame from '@/components/microsites/PhoneFrame';
import { KindIcon } from '@/components/microsites/list/kind';
import { Button, Skeleton, buttonClass, cx, focusRing, inputClass, panelClass } from '@/components/workspace/ui';
import { useAuth } from '@/contexts/AuthContext';
import { useRequireAuth } from '@/hooks/useRequireAuth';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PLAN_NAMES } from '@/lib/plan-utils';
import {
  KIND_TEMPLATE,
  PAGE_KINDS,
  apiFetch,
  browserTimeZone,
  isApiError,
  pageKind,
  type Microsite,
  type MicrositeEntitlements,
  type PageKind,
} from '@/lib/microsites';

function slugify(s: string): string {
  return s
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 32)
    .replace(/-+$/g, '');
}

export default function NewPagePage() {
  return (
    <Suspense fallback={null}>
      <NewPage />
    </Suspense>
  );
}

function NewPage() {
  const router = useRouter();
  const params = useSearchParams();
  const { isLoading: authLoading } = useRequireAuth();
  const { user, getAccessToken } = useAuth();

  const initialKind = (PAGE_KINDS.some((k) => k.id === params.get('kind')) ? params.get('kind') : 'card') as PageKind;
  const [kind, setKind] = useState<PageKind>(initialKind);
  const [template, setTemplate] = useState(KIND_TEMPLATE[initialKind]);
  const [ent, setEnt] = useState<MicrositeEntitlements | null>(null);
  const [baseUrl, setBaseUrl] = useState('');
  const [title, setTitle] = useState('');
  const [titleTouched, setTitleTouched] = useState(false);
  const [slug, setSlug] = useState('');
  const [slugTouched, setSlugTouched] = useState(false);
  const [slugStatus, setSlugStatus] = useState<{ slug: string; available: boolean; reason?: string } | null>(null);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);
  const [previewName, setPreviewName] = useState('');

  const defaultName = user?.name?.trim() || '';

  useEffect(() => {
    if (!user) return;
    apiFetch<{ entitlements: MicrositeEntitlements; base_url: string }>(API_ENDPOINTS.microsites.list, getAccessToken())
      .then((res) => {
        setEnt(res.entitlements);
        setBaseUrl(res.base_url);
      })
      .catch((e) => setError(isApiError(e) ? e.message : 'Could not load your plan.'));
  }, [user, getAccessToken]);

  // Suggest the account name for cards, and a name per type otherwise, until the user types.
  const suggestedTitle = kind === 'card' || kind === 'links' ? defaultName : '';
  const effectiveTitle = titleTouched ? title : suggestedTitle;
  const effectiveSlug = slugTouched ? slug : slugify(effectiveTitle);

  useEffect(() => {
    const t = setTimeout(() => setPreviewName(effectiveTitle.trim()), 400);
    return () => clearTimeout(t);
  }, [effectiveTitle]);

  useEffect(() => {
    if (effectiveSlug.length < 3) return;
    const t = setTimeout(async () => {
      try {
        const res = await apiFetch<{ available: boolean; reason?: string }>(
          `${API_ENDPOINTS.microsites.slugAvailable}?slug=${encodeURIComponent(effectiveSlug)}`,
          getAccessToken()
        );
        setSlugStatus({ slug: effectiveSlug, ...res });
      } catch {
        setSlugStatus(null);
      }
    }, 300);
    return () => clearTimeout(t);
  }, [effectiveSlug, getAccessToken]);

  const pickKind = (k: PageKind) => {
    setKind(k);
    setTemplate(KIND_TEMPLATE[k]);
  };

  const status = slugStatus && slugStatus.slug === effectiveSlug ? slugStatus : null;
  const slugOk = effectiveSlug.length >= 3 && status?.available !== false;
  const canSubmit = !!ent?.can_create_more && effectiveTitle.trim() !== '' && slugOk && !submitting;

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    if (!canSubmit) return;
    setError('');
    setSubmitting(true);
    try {
      const site = await apiFetch<Microsite>(API_ENDPOINTS.microsites.create, getAccessToken(), {
        method: 'POST',
        body: JSON.stringify({ title: effectiveTitle.trim(), slug: effectiveSlug, template, timezone: browserTimeZone() }),
      });
      router.push(`/microsites/edit?id=${site.id}`);
    } catch (err) {
      setError(isApiError(err) ? err.message : 'Could not create the page.');
      setSubmitting(false);
    }
  };

  const previewQuery = useMemo(() => new URLSearchParams({ name: previewName, tz: browserTimeZone() }).toString(), [previewName]);
  const shownBase = (baseUrl || 'trunc.site/m/').replace(/^https?:\/\//, '');
  const info = PAGE_KINDS.find((k) => k.id === kind)!;

  return (
    <WorkspaceLayout
      activeNav="pages"
      onCreateClick={() => router.push('/microsites/new')}
      onSearchSubmit={(q) => router.push(q ? `/microsites?q=${encodeURIComponent(q)}` : '/microsites')}
      searchPlaceholder="Search pages… (⌘K)"
    >
      <nav aria-label="Breadcrumb" className="mb-2 flex items-center gap-1 text-xs text-muted-foreground">
        <Link href="/microsites" className={cx('rounded hover:text-foreground', focusRing)}>
          Pages
        </Link>
        <ChevronRightIcon className="h-3 w-3" />
        <span className="text-foreground">New page</span>
      </nav>
      <h1 className="font-hanken text-2xl font-bold tracking-tight text-foreground">New page</h1>
      <p className="mt-1 text-sm text-muted-foreground">Pick what it is for and a starting design. You can change every word, colour and block afterwards.</p>

      {ent && !ent.can_create_more ? (
        <div className={cx(panelClass, 'mt-6 flex flex-col items-start gap-3 p-6')}>
          <p className="font-hanken text-base font-bold text-foreground">You&apos;ve used all {ent.max_pages} {ent.max_pages === 1 ? 'page' : 'pages'} on your plan</p>
          <p className="text-sm text-muted-foreground">
            Upgrade to {PLAN_NAMES.premium} for up to 10 pages, or to {PLAN_NAMES.growth} for unlimited pages. You can also delete a page you no longer use.
          </p>
          <div className="flex flex-wrap gap-2">
            <Link href="/billing" className={buttonClass('primary', 'md')}>
              <SparklesIcon className="h-4 w-4" />
              See plans
            </Link>
            <Link href="/microsites" className={buttonClass('secondary', 'md')}>
              Back to pages
            </Link>
          </div>
        </div>
      ) : (
        <form onSubmit={submit} className="mt-6 grid gap-8 lg:grid-cols-[minmax(0,1fr)_320px] xl:grid-cols-[minmax(0,1fr)_360px]">
          <div className="min-w-0 space-y-8">
            <section aria-labelledby="kind-heading">
              <h2 id="kind-heading" className="mb-3 text-sm font-semibold text-foreground">
                What is it for?
              </h2>
              <div role="radiogroup" aria-labelledby="kind-heading" className="grid grid-cols-2 gap-2 sm:grid-cols-3 xl:grid-cols-5">
                {PAGE_KINDS.map((k) => {
                  const on = kind === k.id;
                  return (
                    <button
                      key={k.id}
                      type="button"
                      role="radio"
                      aria-checked={on}
                      onClick={() => pickKind(k.id)}
                      className={cx(
                        'flex flex-col items-start gap-1.5 rounded-xl border p-3 text-left transition-colors',
                        focusRing,
                        on
                          ? 'border-primary bg-primary/5 ring-1 ring-primary'
                          : 'border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 hover:bg-surface-container dark:hover:bg-zinc-900'
                      )}
                    >
                      <KindIcon kind={k.id} className={cx('h-5 w-5', on ? 'text-primary' : 'text-muted-foreground')} />
                      <span className="text-sm font-semibold text-foreground">{k.name}</span>
                      <span className="text-xs leading-snug text-muted-foreground">{k.description}</span>
                    </button>
                  );
                })}
              </div>
            </section>

            {kind === 'card' && (
              <section aria-labelledby="design-heading">
                <div className="mb-3 flex flex-wrap items-baseline justify-between gap-2">
                  <h2 id="design-heading" className="text-sm font-semibold text-foreground">
                    Card design
                  </h2>
                  <button
                    type="button"
                    onClick={() => setTemplate('executive_card')}
                    aria-pressed={template === 'executive_card'}
                    className={cx(
                      'rounded-lg px-2 py-1 text-xs font-semibold',
                      focusRing,
                      template === 'executive_card' ? 'bg-primary/10 text-primary' : 'text-muted-foreground hover:text-foreground'
                    )}
                  >
                    {template === 'executive_card' ? 'Letterpress card selected' : 'Or use the Letterpress card'}
                  </button>
                </div>
                <TemplateGallery
                  value={template}
                  onSelect={(id) => setTemplate(id)}
                  idPrefix="card_"
                  showClassic={false}
                  showPalettes={false}
                  scrollClassName=""
                  tone="workspace"
                  startOnCollection
                />
              </section>
            )}

            <section aria-labelledby="details-heading" className={cx(panelClass, 'space-y-5 p-5')}>
              <h2 id="details-heading" className="text-sm font-semibold text-foreground">
                Name and address
              </h2>
              <div>
                <label htmlFor="page-title" className="mb-1.5 block text-xs font-semibold text-foreground">
                  Page name
                </label>
                <input
                  id="page-title"
                  value={effectiveTitle}
                  onChange={(e) => {
                    setTitleTouched(true);
                    setTitle(e.target.value);
                  }}
                  maxLength={120}
                  required
                  placeholder={kind === 'event' ? 'Spring Product Meetup' : kind === 'shop' ? 'Corner Bakery' : 'Your name'}
                  className={inputClass}
                />
                <p className="mt-1 text-xs text-muted-foreground">Shown at the top of the page and in browser tabs.</p>
              </div>

              <div>
                <label htmlFor="page-slug" className="mb-1.5 block text-xs font-semibold text-foreground">
                  Address
                </label>
                <div className="flex items-center overflow-hidden rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 focus-within:border-primary focus-within:ring-2 focus-within:ring-primary/40">
                  <span className="shrink-0 border-r border-outline-variant/30 px-3 py-2 font-mono text-xs text-muted-foreground">{shownBase}</span>
                  <input
                    id="page-slug"
                    value={effectiveSlug}
                    onChange={(e) => {
                      setSlugTouched(true);
                      setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '').slice(0, 32));
                    }}
                    required
                    minLength={3}
                    aria-describedby="slug-status"
                    className="min-w-0 flex-1 bg-transparent px-3 py-2 font-mono text-sm text-foreground outline-none"
                  />
                </div>
                <p id="slug-status" className="mt-1 flex min-h-[1.25rem] items-center gap-1 text-xs" aria-live="polite">
                  {effectiveSlug.length > 0 && effectiveSlug.length < 3 ? (
                    <span className="text-muted-foreground">At least 3 letters, numbers or hyphens.</span>
                  ) : status ? (
                    status.available ? (
                      <span className="inline-flex items-center gap-1 text-emerald-600 dark:text-emerald-400">
                        <CheckCircleIcon className="h-3.5 w-3.5" /> Available
                      </span>
                    ) : (
                      <span className="inline-flex items-center gap-1 text-destructive">
                        <ExclamationCircleIcon className="h-3.5 w-3.5" /> {status.reason ?? 'Already taken. Try another address.'}
                      </span>
                    )
                  ) : (
                    <span className="text-muted-foreground">This is the link people open, and what the QR code points to.</span>
                  )}
                </p>
              </div>

              {error && (
                <p role="alert" className="rounded-xl border border-destructive/20 bg-destructive/10 px-3 py-2 text-xs font-medium text-destructive">
                  {error}
                </p>
              )}

              <div className="flex flex-wrap items-center justify-end gap-2 border-t border-outline-variant/20 pt-4">
                <Link href="/microsites" className={buttonClass('ghost', 'md')}>
                  Cancel
                </Link>
                <Button type="submit" variant="primary" disabled={!canSubmit}>
                  {submitting ? 'Creating…' : `Create ${info.name.toLowerCase()}`}
                </Button>
              </div>
            </section>
          </div>

          <aside className="hidden lg:block" aria-label="Preview">
            <div className="sticky top-24">
              <div className="mb-2 flex items-center justify-between">
                <span className="text-sm font-semibold text-foreground">Preview</span>
                <span className="inline-flex items-center gap-1 text-xs text-muted-foreground">
                  <KindIcon kind={pageKind(template)} className="h-3.5 w-3.5" />
                  {PAGE_KINDS.find((k) => k.id === pageKind(template))?.name}
                </span>
              </div>
              <div className="flex justify-center rounded-2xl border border-outline-variant/30 bg-surface-container/60 dark:bg-zinc-900/50 py-6 [background-image:radial-gradient(hsl(var(--outline-variant)/0.45)_1px,transparent_1px)] [background-size:14px_14px]">
                {authLoading || !user ? (
                  <Skeleton className="h-[520px] w-[262px] rounded-[18px]" />
                ) : (
                  <PhoneFrame
                    key={template}
                    cacheKey={`trunc:tpl:${template}:${previewName}`}
                    url={`${API_ENDPOINTS.microsites.templatePreview(template)}?${previewQuery}`}
                    getToken={getAccessToken}
                    width={262}
                    aspect={2}
                    keepPrevious
                    fallback={<KindIcon kind={pageKind(template)} className="h-6 w-6" />}
                  />
                )}
              </div>
              <p className="mt-2 text-xs text-muted-foreground">Sample content until you edit it.</p>
            </div>
          </aside>
        </form>
      )}
    </WorkspaceLayout>
  );
}
