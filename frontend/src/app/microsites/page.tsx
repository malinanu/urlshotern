'use client';

import { useCallback, useEffect, useState } from 'react';
import Link from 'next/link';
import { useRouter } from 'next/navigation';
import {
  DevicePhoneMobileIcon,
  PlusIcon,
  TrashIcon,
  ArrowTopRightOnSquareIcon,
  SparklesIcon,
  EyeIcon,
  XMarkIcon,
} from '@heroicons/react/24/outline';
import Layout from '@/components/layout/Layout';
import { useAuth } from '@/contexts/AuthContext';
import { useRequireAuth } from '@/hooks/useRequireAuth';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PLAN_NAMES } from '@/lib/plan-utils';
import {
  apiFetch,
  browserTimeZone,
  isApiError,
  type Microsite,
  type MicrositeEntitlements,
  type MicrositeSummary,
  type MicrositeTemplateInfo,
} from '@/lib/microsites';

interface ListResponse {
  microsites: MicrositeSummary[];
  entitlements: MicrositeEntitlements;
  base_url: string;
}

function slugify(s: string): string {
  return s
    .toLowerCase()
    .normalize('NFKD')
    .replace(/[^a-z0-9]+/g, '-')
    .replace(/^-+|-+$/g, '')
    .slice(0, 32);
}

export default function MicrositesPage() {
  const router = useRouter();
  const { isLoading: authLoading } = useRequireAuth();
  const { user, getAccessToken } = useAuth();
  const [data, setData] = useState<ListResponse | null>(null);
  const [templates, setTemplates] = useState<MicrositeTemplateInfo[]>([]);
  const [loading, setLoading] = useState(true);
  const [error, setError] = useState('');
  const [showCreate, setShowCreate] = useState(false);

  const load = useCallback(async () => {
    setError('');
    try {
      const token = getAccessToken();
      const [list, tmpl] = await Promise.all([
        apiFetch<ListResponse>(API_ENDPOINTS.microsites.list, token),
        apiFetch<{ templates: MicrositeTemplateInfo[] }>(API_ENDPOINTS.microsites.templates, token),
      ]);
      setData(list);
      setTemplates(tmpl.templates);
    } catch (e) {
      setError(isApiError(e) ? e.message : 'Could not load your pages.');
    } finally {
      setLoading(false);
    }
  }, [getAccessToken]);

  useEffect(() => {
    if (user) load();
  }, [user, load]);

  const handleDelete = async (site: MicrositeSummary) => {
    if (!confirm(`Delete "${site.title}"? Its address (/m/${site.slug}) will stop working.`)) return;
    try {
      await apiFetch(API_ENDPOINTS.microsites.delete(site.id), getAccessToken(), { method: 'DELETE' });
      await load();
    } catch (e) {
      setError(isApiError(e) ? e.message : 'Could not delete the page.');
    }
  };

  if (authLoading || !user) {
    return (
      <Layout>
        <div className="min-h-screen bg-background flex items-center justify-center">
          <p className="text-muted-foreground">Loading…</p>
        </div>
      </Layout>
    );
  }

  const ent = data?.entitlements;
  const limitLabel = ent ? (ent.max_pages === -1 ? 'Unlimited pages' : `${ent.used_pages} of ${ent.max_pages} pages used`) : '';

  return (
    <Layout>
      <div className="min-h-screen bg-background py-8">
        <div className="max-w-5xl mx-auto px-4 sm:px-6 lg:px-8">
          <div className="flex flex-col sm:flex-row sm:items-center sm:justify-between gap-4 mb-8">
            <div className="flex items-center space-x-3">
              <div className="p-2 bg-primary/10 rounded-lg">
                <DevicePhoneMobileIcon className="h-6 w-6 text-primary" />
              </div>
              <div>
                <h1 className="text-2xl font-bold text-foreground">Pages</h1>
                <p className="text-sm text-muted-foreground mt-0.5">
                  Business cards, link-in-bio, shop and event pages{limitLabel && ` · ${limitLabel}`}
                </p>
              </div>
            </div>
            {ent && ent.can_create_more && (
              <button
                onClick={() => setShowCreate(true)}
                className="inline-flex items-center justify-center space-x-2 px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium transition-colors"
              >
                <PlusIcon className="h-4 w-4" />
                <span>Create page</span>
              </button>
            )}
          </div>

          {error && (
            <div role="alert" className="mb-6 rounded-lg border border-red-300 bg-red-50 dark:bg-red-950/30 px-4 py-3 text-sm text-red-700 dark:text-red-300">
              {error}
            </div>
          )}

          {ent && !ent.can_create_more && (
            <div className="mb-6 flex flex-col sm:flex-row sm:items-center gap-3 rounded-xl border border-border bg-card px-5 py-4">
              <SparklesIcon className="h-5 w-5 text-primary shrink-0" />
              <p className="text-sm text-muted-foreground flex-1">
                You&apos;ve used all {ent.max_pages} page{ent.max_pages === 1 ? '' : 's'} on your plan. Upgrade to{' '}
                <strong className="text-foreground">{PLAN_NAMES.premium}</strong> for more pages.
              </p>
              <Link href="/billing" className="text-sm font-medium text-primary hover:underline">
                Upgrade plan
              </Link>
            </div>
          )}

          {loading ? (
            <div className="grid gap-4 sm:grid-cols-2">
              {[0, 1].map((i) => (
                <div key={i} className="h-36 rounded-xl border border-border bg-card animate-pulse" />
              ))}
            </div>
          ) : data && data.microsites.length === 0 ? (
            <div className="bg-card border border-border rounded-xl p-12 text-center">
              <div className="mx-auto w-14 h-14 bg-primary/10 rounded-full flex items-center justify-center mb-4">
                <DevicePhoneMobileIcon className="h-7 w-7 text-primary" />
              </div>
              <h2 className="text-xl font-semibold text-foreground mb-2">Create your first page</h2>
              <p className="text-muted-foreground mb-6 max-w-md mx-auto">
                Share one link that lets people save your contact in a tap, call or message you, and find all your links.
              </p>
              <button
                onClick={() => setShowCreate(true)}
                className="inline-flex items-center space-x-2 px-6 py-2.5 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 font-medium transition-colors"
              >
                <PlusIcon className="h-4 w-4" />
                <span>Create page</span>
              </button>
            </div>
          ) : (
            <ul className="grid gap-4 sm:grid-cols-2">
              {data?.microsites.map((site) => (
                <li key={site.id} className="rounded-xl border border-border bg-card p-5 flex flex-col gap-3">
                  <div className="flex items-start justify-between gap-3">
                    <div className="min-w-0">
                      <h2 className="font-semibold text-foreground truncate">{site.title}</h2>
                      <p className="text-sm text-muted-foreground truncate font-mono">/m/{site.slug}</p>
                    </div>
                    <span
                      className={`shrink-0 rounded-full px-2.5 py-0.5 text-xs font-medium ${
                        site.status === 'published'
                          ? 'bg-green-100 text-green-800 dark:bg-green-900/40 dark:text-green-300'
                          : 'bg-muted text-muted-foreground'
                      }`}
                    >
                      {site.status === 'published' ? 'Published' : 'Draft'}
                    </span>
                  </div>
                  <p className="text-sm text-muted-foreground flex items-center gap-1.5">
                    <EyeIcon className="h-4 w-4" />
                    {site.views_30d.toLocaleString()} views in the last 30 days
                  </p>
                  <div className="flex items-center gap-2 mt-auto pt-2">
                    <Link
                      href={`/microsites/edit?id=${site.id}`}
                      className="px-3 py-1.5 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium"
                    >
                      Edit
                    </Link>
                    <Link
                      href={`/microsites/edit?id=${site.id}&tab=track`}
                      className="px-3 py-1.5 border border-border rounded-lg hover:bg-muted text-sm font-medium text-foreground"
                    >
                      Analytics
                    </Link>
                    {site.status === 'published' && data && (
                      <a
                        href={`${data.base_url}${site.slug}`}
                        target="_blank"
                        rel="noopener noreferrer"
                        className="p-1.5 text-muted-foreground hover:text-foreground"
                        aria-label={`Open /m/${site.slug}`}
                      >
                        <ArrowTopRightOnSquareIcon className="h-5 w-5" />
                      </a>
                    )}
                    <button
                      onClick={() => handleDelete(site)}
                      className="ml-auto p-1.5 text-muted-foreground hover:text-red-600"
                      aria-label={`Delete ${site.title}`}
                    >
                      <TrashIcon className="h-5 w-5" />
                    </button>
                  </div>
                </li>
              ))}
            </ul>
          )}
        </div>
      </div>

      {showCreate && (
        <CreateDialog
          templates={templates}
          baseUrl={data?.base_url ?? '/m/'}
          defaultName={user.name}
          onClose={() => setShowCreate(false)}
          onCreated={(site) => router.push(`/microsites/edit?id=${site.id}`)}
        />
      )}
    </Layout>
  );
}

function CreateDialog({
  templates,
  baseUrl,
  defaultName,
  onClose,
  onCreated,
}: {
  templates: MicrositeTemplateInfo[];
  baseUrl: string;
  defaultName?: string;
  onClose: () => void;
  onCreated: (site: Microsite) => void;
}) {
  const { getAccessToken } = useAuth();
  const [template, setTemplate] = useState(templates[0]?.id ?? 'executive_card');
  const [title, setTitle] = useState(defaultName ?? '');
  const [slug, setSlug] = useState(slugify(defaultName ?? ''));
  const [slugTouched, setSlugTouched] = useState(false);
  const [slugStatus, setSlugStatus] = useState<{ available: boolean; reason?: string } | null>(null);
  const [error, setError] = useState('');
  const [submitting, setSubmitting] = useState(false);

  useEffect(() => {
    if (!slugTouched) setSlug(slugify(title));
  }, [title, slugTouched]);

  useEffect(() => {
    if (slug.length < 3) {
      setSlugStatus(null);
      return;
    }
    const t = setTimeout(async () => {
      try {
        const res = await apiFetch<{ available: boolean; reason?: string }>(
          `${API_ENDPOINTS.microsites.slugAvailable}?slug=${encodeURIComponent(slug)}`,
          getAccessToken()
        );
        setSlugStatus(res);
      } catch {
        setSlugStatus(null);
      }
    }, 300);
    return () => clearTimeout(t);
  }, [slug, getAccessToken]);

  const submit = async (e: React.FormEvent) => {
    e.preventDefault();
    setError('');
    setSubmitting(true);
    try {
      const site = await apiFetch<Microsite>(API_ENDPOINTS.microsites.create, getAccessToken(), {
        method: 'POST',
        body: JSON.stringify({ title: title.trim(), slug, template, timezone: browserTimeZone() }),
      });
      onCreated(site);
    } catch (err) {
      setError(isApiError(err) ? err.message : 'Could not create the page.');
      setSubmitting(false);
    }
  };

  return (
    <div className="fixed inset-0 z-50 flex items-center justify-center bg-black/50 p-4" role="dialog" aria-modal="true" aria-labelledby="create-page-title">
      <form
        onSubmit={submit}
        className="w-full max-w-lg md:max-w-4xl max-h-[calc(100vh-2rem)] overflow-y-auto rounded-xl bg-card border border-border shadow-xl p-6 grid gap-5 md:grid-cols-[1fr_300px] md:gap-6"
      >
        <div className="space-y-5 min-w-0">
        <div className="flex items-center justify-between">
          <h2 id="create-page-title" className="text-lg font-semibold text-foreground">Create a page</h2>
          <button type="button" onClick={onClose} className="p-1 text-muted-foreground hover:text-foreground" aria-label="Close">
            <XMarkIcon className="h-5 w-5" />
          </button>
        </div>

        <fieldset>
          <legend className="text-sm font-medium text-foreground mb-2">Start from</legend>
          <div className="grid gap-2">
            {templates.map((t) => (
              <label
                key={t.id}
                className={`flex gap-3 rounded-lg border p-3 cursor-pointer ${
                  template === t.id ? 'border-primary bg-primary/5' : 'border-border hover:bg-muted'
                }`}
              >
                <input type="radio" name="template" value={t.id} checked={template === t.id} onChange={() => setTemplate(t.id)} className="mt-1" />
                <span>
                  <span className="block text-sm font-medium text-foreground">{t.name}</span>
                  <span className="block text-xs text-muted-foreground">{t.description}</span>
                </span>
              </label>
            ))}
          </div>
        </fieldset>

        <div>
          <label htmlFor="page-title" className="block text-sm font-medium text-foreground mb-1">Page name</label>
          <input
            id="page-title"
            value={title}
            onChange={(e) => setTitle(e.target.value)}
            maxLength={120}
            required
            className="w-full rounded-lg border border-border bg-background px-3 py-2 text-sm text-foreground"
          />
        </div>

        <div>
          <label htmlFor="page-slug" className="block text-sm font-medium text-foreground mb-1">Address</label>
          <div className="flex items-center rounded-lg border border-border bg-background overflow-hidden">
            <span className="px-3 py-2 text-sm text-muted-foreground bg-muted border-r border-border whitespace-nowrap">{baseUrl.replace(/^https?:\/\//, '')}</span>
            <input
              id="page-slug"
              value={slug}
              onChange={(e) => {
                setSlugTouched(true);
                setSlug(e.target.value.toLowerCase().replace(/[^a-z0-9-]/g, '').slice(0, 32));
              }}
              required
              minLength={3}
              className="flex-1 min-w-0 px-3 py-2 text-sm text-foreground bg-transparent outline-none font-mono"
              aria-describedby="slug-status"
            />
          </div>
          <p id="slug-status" className="mt-1 text-xs min-h-[1rem]" aria-live="polite">
            {slugStatus &&
              (slugStatus.available ? (
                <span className="text-green-600">Available</span>
              ) : (
                <span className="text-red-600">{slugStatus.reason ?? 'Already taken'}</span>
              ))}
          </p>
        </div>

        {error && <p role="alert" className="text-sm text-red-600">{error}</p>}

        <div className="flex justify-end gap-2">
          <button type="button" onClick={onClose} className="px-4 py-2 rounded-lg border border-border text-sm font-medium text-foreground hover:bg-muted">
            Cancel
          </button>
          <button
            type="submit"
            disabled={submitting || slugStatus?.available === false || slug.length < 3}
            className="px-4 py-2 rounded-lg bg-primary text-primary-foreground text-sm font-medium hover:bg-primary/90 disabled:opacity-50"
          >
            {submitting ? 'Creating…' : 'Create page'}
          </button>
        </div>
        </div>
        <TemplatePreview template={template} name={title} />
      </form>
    </div>
  );
}

/** Server-rendered sample of a template, shown in a sandboxed, scaled iframe. */
function TemplatePreview({ template, name }: { template: string; name: string }) {
  const { getAccessToken } = useAuth();
  const [html, setHtml] = useState<Record<string, string>>({});
  const [debouncedName, setDebouncedName] = useState(name);

  useEffect(() => {
    const t = setTimeout(() => setDebouncedName(name.trim()), 400);
    return () => clearTimeout(t);
  }, [name]);

  const key = `${template}|${debouncedName}`;
  useEffect(() => {
    if (html[key] !== undefined) return;
    let cancelled = false;
    const q = new URLSearchParams({ name: debouncedName, tz: browserTimeZone() });
    apiFetch<{ html: string }>(`${API_ENDPOINTS.microsites.templatePreview(template)}?${q}`, getAccessToken())
      .then((res) => {
        if (!cancelled) setHtml((h) => ({ ...h, [key]: res.html }));
      })
      .catch(() => {
        if (!cancelled) setHtml((h) => ({ ...h, [key]: '' }));
      });
    return () => {
      cancelled = true;
    };
  }, [key, template, debouncedName, html, getAccessToken]);

  const doc = html[key];
  const width = 390;
  const scale = 300 / width;
  return (
    <div className="hidden md:block" aria-hidden="true">
      <p className="text-sm font-medium text-foreground mb-2">Preview</p>
      <div className="rounded-[1.75rem] border-[8px] border-gray-900 overflow-hidden bg-white shadow-lg" style={{ width: 300 + 16, height: 600 }}>
        <div style={{ width, height: 584 / scale, transform: `scale(${scale})`, transformOrigin: 'top left' }}>
          {doc ? (
            <iframe title="Template preview" srcDoc={doc} sandbox="" tabIndex={-1} className="w-full h-full border-0" />
          ) : (
            <div className="w-full h-full animate-pulse bg-muted" />
          )}
        </div>
      </div>
      <p className="mt-2 text-xs text-muted-foreground">Sample content. You can change every word, color and block.</p>
    </div>
  );
}
