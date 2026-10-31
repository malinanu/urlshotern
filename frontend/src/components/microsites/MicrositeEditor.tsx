'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import {
  ArrowLeftIcon,
  ArrowTopRightOnSquareIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  LinkIcon,
  LockClosedIcon,
  PlusIcon,
  QrCodeIcon,
} from '@heroicons/react/24/outline';
import { QRCodeModal } from '@/components/QRCode';
import { useAuth } from '@/contexts/AuthContext';
import { API_ENDPOINTS } from '@/lib/api-config';
import {
  BLOCK_CATALOG,
  BLOCK_GROUPS,
  apiFetch,
  describeValidation,
  isApiError,
  newId,
  type BlockType,
  type Microsite,
  type MicrositeBlock,
  type MicrositeContent,
  type MicrositeEntitlements,
} from '@/lib/microsites';
import BlockList from './BlockList';
import DesignPanel from './DesignPanel';
import InboxPanel from './InboxPanel';
import PreviewFrame from './PreviewFrame';
import TrackPanel from './TrackPanel';

type Tab = 'build' | 'design' | 'track' | 'inbox';
const TABS: Tab[] = ['build', 'design', 'track', 'inbox'];
type SaveState = 'saved' | 'saving' | 'unsaved' | 'error' | 'conflict';

const SAVE_LABELS: Record<SaveState, string> = {
  saved: 'All changes saved',
  saving: 'Saving…',
  unsaved: 'Unsaved changes',
  error: 'Not saved',
  conflict: 'Changed elsewhere',
};

export default function MicrositeEditor() {
  const router = useRouter();
  const params = useSearchParams();
  const id = Number(params.get('id'));
  const tab = (TABS.includes(params.get('tab') as Tab) ? params.get('tab') : 'build') as Tab;
  const { getAccessToken } = useAuth();

  const [site, setSite] = useState<Microsite | null>(null);
  const [entitlements, setEntitlements] = useState<MicrositeEntitlements | null>(null);
  const [title, setTitle] = useState('');
  const [content, setContent] = useState<MicrositeContent | null>(null);
  const [selected, setSelected] = useState<string | null>(null);
  const [loadError, setLoadError] = useState('');
  const [saveState, setSaveState] = useState<SaveState>('saved');
  const [saveError, setSaveError] = useState('');
  const [publishing, setPublishing] = useState(false);
  const [notice, setNotice] = useState('');
  const [showQR, setShowQR] = useState(false);
  const [copied, setCopied] = useState(false);

  // Refs give the debounced save the latest values without re-creating it.
  const titleRef = useRef(title);
  const contentRef = useRef(content);
  const versionRef = useRef(0);
  const revision = useRef(0);
  const inFlight = useRef<Promise<boolean> | null>(null);
  titleRef.current = title;
  contentRef.current = content;

  const load = useCallback(async () => {
    try {
      const res = await apiFetch<{ microsite: Microsite; entitlements: MicrositeEntitlements }>(
        API_ENDPOINTS.microsites.get(id),
        getAccessToken()
      );
      setSite(res.microsite);
      setEntitlements(res.entitlements);
      setTitle(res.microsite.title);
      setContent(res.microsite.draft);
      versionRef.current = res.microsite.version;
      setSaveState('saved');
      setSaveError('');
    } catch (e) {
      setLoadError(isApiError(e) ? e.message : 'Could not load this page.');
    }
  }, [id, getAccessToken]);

  useEffect(() => {
    if (id > 0) load();
    else setLoadError('No page selected.');
  }, [id, load]);

  const save = useCallback(async (): Promise<boolean> => {
    if (inFlight.current) await inFlight.current;
    if (!contentRef.current) return false;
    const rev = revision.current;
    setSaveState('saving');
    const run = (async () => {
      try {
        const res = await apiFetch<Microsite>(API_ENDPOINTS.microsites.update(id), getAccessToken(), {
          method: 'PUT',
          body: JSON.stringify({ version: versionRef.current, title: titleRef.current, content: contentRef.current }),
        });
        versionRef.current = res.version;
        setSite((s) => (s ? { ...s, version: res.version, title: res.title, updated_at: res.updated_at } : s));
        setSaveError('');
        setSaveState(revision.current === rev ? 'saved' : 'unsaved');
        return true;
      } catch (e) {
        if (isApiError(e) && e.error === 'version_conflict') {
          setSaveState('conflict');
        } else {
          setSaveError(
            isApiError(e) ? describeValidation(e.field, e.message, contentRef.current?.blocks ?? []) : 'Could not save changes.'
          );
          setSaveState('error');
        }
        return false;
      }
    })();
    inFlight.current = run;
    const ok = await run;
    inFlight.current = null;
    return ok;
  }, [id, getAccessToken]);

  // Debounced autosave
  useEffect(() => {
    if (saveState !== 'unsaved') return;
    const t = setTimeout(save, 1000);
    return () => clearTimeout(t);
  }, [title, content, saveState, save]);

  // Warn before leaving with unsaved edits
  useEffect(() => {
    const handler = (e: BeforeUnloadEvent) => {
      if (saveState === 'unsaved' || saveState === 'saving' || saveState === 'error') e.preventDefault();
    };
    window.addEventListener('beforeunload', handler);
    return () => window.removeEventListener('beforeunload', handler);
  }, [saveState]);

  const edit = (next: (c: MicrositeContent) => MicrositeContent) => {
    setContent((c) => (c ? next(c) : c));
    revision.current++;
    setSaveState((s) => (s === 'conflict' ? s : 'unsaved'));
  };

  const editTitle = (t: string) => {
    setTitle(t);
    revision.current++;
    setSaveState((s) => (s === 'conflict' ? s : 'unsaved'));
  };

  const addBlock = (type: BlockType) => {
    const entry = BLOCK_CATALOG.find((b) => b.type === type);
    if (!entry || (entry.requiresForms && !entitlements?.can_use_forms)) return;
    const block: MicrositeBlock = { id: newId('blk'), type, data: entry.create() };
    edit((c) => ({ ...c, blocks: [...c.blocks, block] }));
    setSelected(block.id);
  };

  const setTab = (t: Tab) => router.replace(`/microsites/edit?id=${id}${t === 'build' ? '' : `&tab=${t}`}`);

  const publish = async () => {
    setNotice('');
    setPublishing(true);
    try {
      if (saveState !== 'saved' && !(await save())) {
        setNotice('Fix the problem above before publishing.');
        return;
      }
      const res = await apiFetch<Microsite>(API_ENDPOINTS.microsites.publish(id), getAccessToken(), { method: 'POST' });
      setSite((s) => (s ? { ...s, status: res.status, published_at: res.published_at, published: res.published } : s));
      setNotice('Published! Your page is live.');
    } catch (e) {
      setNotice(
        isApiError(e)
          ? `Can't publish yet — ${describeValidation(e.field, e.message, contentRef.current?.blocks ?? [])}`
          : 'Could not publish.'
      );
    } finally {
      setPublishing(false);
    }
  };

  const unpublish = async () => {
    if (!confirm('Take this page offline? Its link will show “not found” until you publish again.')) return;
    try {
      const res = await apiFetch<Microsite>(API_ENDPOINTS.microsites.unpublish(id), getAccessToken(), { method: 'POST' });
      setSite((s) => (s ? { ...s, status: res.status } : s));
      setNotice('Page is now offline.');
    } catch (e) {
      setNotice(isApiError(e) ? e.message : 'Could not unpublish.');
    }
  };

  const createShortLink = async () => {
    try {
      const res = await apiFetch<{ short_code: string; short_url: string }>(API_ENDPOINTS.microsites.shortLink(id), getAccessToken(), {
        method: 'POST',
      });
      setSite((s) => (s ? { ...s, short_code: res.short_code } : s));
      setNotice(`Short link ready: ${res.short_url}`);
    } catch (e) {
      setNotice(isApiError(e) ? e.message : 'Could not create a short link.');
    }
  };

  const copyLink = () => {
    if (!site) return;
    navigator.clipboard.writeText(site.public_url).then(() => {
      setCopied(true);
      setTimeout(() => setCopied(false), 2000);
    });
  };

  if (loadError) {
    return (
      <div className="min-h-[60vh] flex flex-col items-center justify-center gap-3 px-4 text-center">
        <p className="text-foreground">{loadError}</p>
        <Link href="/microsites" className="text-primary hover:underline text-sm">
          Back to pages
        </Link>
      </div>
    );
  }

  if (!site || !content) {
    return <div className="min-h-[60vh] flex items-center justify-center text-muted-foreground">Loading…</div>;
  }

  const published = site.status === 'published';
  const shortUrl = site.short_code ? site.public_url.replace(/\/m\/.*$/, `/${site.short_code}`) : '';

  return (
    <div className="min-h-screen bg-background">
      {/* Top bar */}
      <div className="sticky top-0 z-20 border-b border-border bg-background/95 backdrop-blur">
        <div className="max-w-[1600px] mx-auto px-4 py-3 flex flex-wrap items-center gap-3">
          <Link href="/microsites" className="inline-flex items-center gap-1 text-sm text-muted-foreground hover:text-foreground">
            <ArrowLeftIcon className="h-4 w-4" />
            Pages
          </Link>
          <input
            value={title}
            onChange={(e) => editTitle(e.target.value)}
            maxLength={120}
            aria-label="Page name"
            className="min-w-0 flex-1 sm:flex-none sm:w-56 rounded-md bg-transparent px-2 py-1 text-base font-semibold text-foreground hover:bg-muted focus:bg-muted focus:outline-none"
          />
          <div className="flex items-center gap-1 rounded-lg border border-border px-2 py-1 text-sm">
            <span className="font-mono text-muted-foreground truncate max-w-[14rem]">{site.public_url.replace(/^https?:\/\//, '')}</span>
            <button type="button" onClick={copyLink} className="p-1 text-muted-foreground hover:text-foreground" aria-label="Copy page link">
              {copied ? <CheckIcon className="h-4 w-4 text-green-600" /> : <ClipboardDocumentIcon className="h-4 w-4" />}
            </button>
            <button type="button" onClick={() => setShowQR(true)} className="p-1 text-muted-foreground hover:text-foreground" aria-label="Show QR code">
              <QrCodeIcon className="h-4 w-4" />
            </button>
            {published && (
              <a href={site.public_url} target="_blank" rel="noopener noreferrer" className="p-1 text-muted-foreground hover:text-foreground" aria-label="Open live page">
                <ArrowTopRightOnSquareIcon className="h-4 w-4" />
              </a>
            )}
          </div>

          <div className="ml-auto flex items-center gap-3">
            <span
              role="status"
              className={`text-xs ${saveState === 'error' || saveState === 'conflict' ? 'text-red-600' : 'text-muted-foreground'}`}
            >
              {SAVE_LABELS[saveState]}
            </span>
            {published && (
              <button type="button" onClick={unpublish} className="text-sm text-muted-foreground hover:text-foreground">
                Unpublish
              </button>
            )}
            <button
              type="button"
              onClick={publish}
              disabled={publishing || saveState === 'conflict'}
              className="px-4 py-2 bg-primary text-primary-foreground rounded-lg hover:bg-primary/90 text-sm font-medium disabled:opacity-50"
            >
              {publishing ? 'Publishing…' : published ? 'Update live page' : 'Publish'}
            </button>
          </div>
        </div>

        <nav className="max-w-[1600px] mx-auto px-4 flex gap-1" aria-label="Editor sections">
          {TABS.map((t) => (
            <button
              key={t}
              type="button"
              onClick={() => setTab(t)}
              aria-current={tab === t ? 'page' : undefined}
              className={`px-4 py-2 text-sm font-medium border-b-2 -mb-px capitalize ${
                tab === t ? 'border-primary text-foreground' : 'border-transparent text-muted-foreground hover:text-foreground'
              }`}
            >
              {t}
            </button>
          ))}
        </nav>
      </div>

      {(saveError || notice || saveState === 'conflict') && (
        <div className="max-w-[1600px] mx-auto px-4 pt-4 grid gap-2">
          {saveState === 'conflict' && (
            <p role="alert" className="rounded-lg border border-red-300 bg-red-50 dark:bg-red-950/30 px-4 py-2 text-sm text-red-700 dark:text-red-300">
              This page was changed in another tab or by another session.{' '}
              <button type="button" onClick={load} className="underline font-medium">
                Reload the latest version
              </button>
            </p>
          )}
          {saveError && saveState === 'error' && (
            <p role="alert" className="rounded-lg border border-red-300 bg-red-50 dark:bg-red-950/30 px-4 py-2 text-sm text-red-700 dark:text-red-300">
              {saveError}
            </p>
          )}
          {notice && (
            <p role="status" className="rounded-lg border border-border bg-card px-4 py-2 text-sm text-foreground flex items-center justify-between gap-3">
              <span className="break-all">{notice}</span>
              <button type="button" onClick={() => setNotice('')} className="text-xs text-muted-foreground hover:text-foreground">
                Dismiss
              </button>
            </p>
          )}
        </div>
      )}

      <div className="max-w-[1600px] mx-auto px-4 py-6">
        {tab === 'track' ? (
          <TrackPanel micrositeId={id} maxDays={entitlements?.analytics_days} />
        ) : tab === 'inbox' ? (
          <InboxPanel micrositeId={id} blocks={content.blocks} canUseForms={!!entitlements?.can_use_forms} />
        ) : (
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(320px,420px)_minmax(0,1fr)] xl:grid-cols-[minmax(340px,420px)_minmax(0,1fr)_260px]">
            {/* Left: module config */}
            <div className="order-2 lg:order-1 min-w-0">
              {tab === 'build' ? (
                <div className="grid grid-cols-1 gap-4">
                  <h2 className="text-sm font-semibold text-foreground">Blocks</h2>
                  <BlockList
                    blocks={content.blocks}
                    selectedId={selected}
                    micrositeId={id}
                    onSelect={setSelected}
                    onReorder={(blocks) => edit((c) => ({ ...c, blocks }))}
                    onUpdate={(bid, patch) =>
                      edit((c) => ({ ...c, blocks: c.blocks.map((b) => (b.id === bid ? { ...b, ...patch } : b)) }))
                    }
                    onRemove={(bid) => edit((c) => ({ ...c, blocks: c.blocks.filter((b) => b.id !== bid) }))}
                    canUseForms={!!entitlements?.can_use_forms}
                  />
                  {/* Catalog shown inline below xl, as a side panel at xl */}
                  <div className="xl:hidden">
                    <BlockCatalog onAdd={addBlock} canUseForms={!!entitlements?.can_use_forms} />
                  </div>
                </div>
              ) : (
                <DesignPanel
                  theme={content.theme}
                  settings={content.settings}
                  entitlements={entitlements}
                  micrositeId={id}
                  onTheme={(theme) => edit((c) => ({ ...c, theme }))}
                  onSettings={(settings) => edit((c) => ({ ...c, settings }))}
                />
              )}
            </div>

            {/* Center: live canvas */}
            <div className="order-1 lg:order-2 min-w-0 lg:sticky lg:top-32 lg:self-start">
              <PreviewFrame title={title} slug={site.slug} content={content} />
            </div>

            {/* Right: catalog & sharing */}
            <aside className="hidden xl:grid gap-6 order-3 self-start">
              <BlockCatalog
                canUseForms={!!entitlements?.can_use_forms}
                onAdd={(t) => {
                  if (tab !== 'build') setTab('build');
                  addBlock(t);
                }}
              />
              <section className="grid gap-2">
                <h2 className="text-sm font-semibold text-foreground">Share</h2>
                {shortUrl ? (
                  <p className="text-sm font-mono text-foreground break-all">{shortUrl.replace(/^https?:\/\//, '')}</p>
                ) : (
                  <button
                    type="button"
                    onClick={createShortLink}
                    className="inline-flex items-center justify-center gap-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground hover:bg-muted"
                  >
                    <LinkIcon className="h-4 w-4" />
                    Create short link
                  </button>
                )}
                <button
                  type="button"
                  onClick={() => setShowQR(true)}
                  className="inline-flex items-center justify-center gap-1 rounded-lg border border-border px-3 py-2 text-sm text-foreground hover:bg-muted"
                >
                  <QrCodeIcon className="h-4 w-4" />
                  QR code
                </button>
              </section>
            </aside>
          </div>
        )}
      </div>

      <QRCodeModal
        isOpen={showQR}
        onClose={() => setShowQR(false)}
        value={`${site.public_url}?src=qr`}
        title="QR code for your page"
        description="Scans are counted as “QR code” traffic in Track."
      />
    </div>
  );
}

function BlockCatalog({ onAdd, canUseForms }: { onAdd: (type: BlockType) => void; canUseForms: boolean }) {
  return (
    <section className="grid gap-3">
      <h2 className="text-sm font-semibold text-foreground">Add a block</h2>
      {BLOCK_GROUPS.map((g) => {
        const entries = BLOCK_CATALOG.filter((b) => b.group === g.id);
        return (
          <div key={g.id} className="grid gap-1.5">
            <h3 className="text-xs font-medium text-muted-foreground">{g.name}</h3>
            <ul className="grid grid-cols-2 xl:grid-cols-1 gap-2">
              {entries.map((b) => {
                const locked = b.requiresForms && !canUseForms;
                return (
                  <li key={b.type}>
                    {locked ? (
                      <Link
                        href="/billing"
                        className="block w-full h-full text-left rounded-lg border border-dashed border-border bg-card px-3 py-2 hover:bg-muted"
                      >
                        <span className="flex items-center gap-1 text-sm font-medium text-muted-foreground">
                          <LockClosedIcon className="h-4 w-4" />
                          {b.name}
                        </span>
                        <span className="block text-xs text-muted-foreground">Pro and Ultra plans. Upgrade to add.</span>
                      </Link>
                    ) : (
                      <button
                        type="button"
                        onClick={() => onAdd(b.type)}
                        className="w-full h-full text-left rounded-lg border border-border bg-card px-3 py-2 hover:border-primary hover:bg-primary/5"
                      >
                        <span className="flex items-center gap-1 text-sm font-medium text-foreground">
                          <PlusIcon className="h-4 w-4 text-primary" />
                          {b.name}
                        </span>
                        <span className="block text-xs text-muted-foreground">{b.description}</span>
                      </button>
                    )}
                  </li>
                );
              })}
            </ul>
          </div>
        );
      })}
    </section>
  );
}
