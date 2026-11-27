'use client';

import { useCallback, useEffect, useRef, useState } from 'react';
import Link from 'next/link';
import { useRouter, useSearchParams } from 'next/navigation';
import { Menu, MenuButton, MenuItem, MenuItems, Popover, PopoverButton, PopoverPanel } from '@headlessui/react';
import {
  ArrowTopRightOnSquareIcon,
  CheckIcon,
  ChevronRightIcon,
  ClipboardDocumentIcon,
  EllipsisHorizontalIcon,
  EyeIcon,
  EyeSlashIcon,
  LinkIcon,
  LockClosedIcon,
  MagnifyingGlassIcon,
  PencilSquareIcon,
  PlusIcon,
  QrCodeIcon,
  ShareIcon,
  Squares2X2Icon,
} from '@heroicons/react/24/outline';
import { QRCodeModal } from '@/components/QRCode';
import ConfirmDialog from '@/components/workspace/ConfirmDialog';
import { Badge, Button, IconButton, Segmented, StatusDot, cx, focusRing } from '@/components/workspace/ui';
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
  PAGE_KINDS,
  pageKind,
  pageState,
} from '@/lib/microsites';
import BlockList from './BlockList';
import DesignPanel from './DesignPanel';
import InboxPanel from './InboxPanel';
import PreviewFrame from './PreviewFrame';
import TrackPanel from './TrackPanel';

type Tab = 'build' | 'design' | 'analytics' | 'inbox';
const TABS: Tab[] = ['build', 'design', 'analytics', 'inbox'];
const TAB_LABELS: Record<Tab, string> = { build: 'Build', design: 'Design', analytics: 'Analytics', inbox: 'Inbox' };

function parseTab(raw: string | null): Tab {
  if (raw === 'track') return 'analytics'; // old links
  return TABS.includes(raw as Tab) ? (raw as Tab) : 'build';
}
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
  const tab = parseTab(params.get('tab'));
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
  const [mobileView, setMobileView] = useState<'edit' | 'preview'>('edit');
  const [confirmUnpublish, setConfirmUnpublish] = useState(false);

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
        setSite((s) =>
          s
            ? {
                ...s,
                version: res.version,
                title: res.title,
                updated_at: res.updated_at,
                has_unpublished_changes: res.has_unpublished_changes,
              }
            : s
        );
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
      setSite((s) =>
        s
          ? {
              ...s,
              status: res.status,
              published_at: res.published_at,
              published: res.published,
              published_version: res.published_version,
              has_unpublished_changes: res.has_unpublished_changes,
            }
          : s
      );
      setNotice('Published. Your page is live.');
    } catch (e) {
      setNotice(
        isApiError(e)
          ? `Can't publish yet: ${describeValidation(e.field, e.message, contentRef.current?.blocks ?? [])}`
          : 'Could not publish.'
      );
    } finally {
      setPublishing(false);
    }
  };

  const unpublish = async () => {
    try {
      const res = await apiFetch<Microsite>(API_ENDPOINTS.microsites.unpublish(id), getAccessToken(), { method: 'POST' });
      setSite((s) =>
        s ? { ...s, status: res.status, published_version: undefined, has_unpublished_changes: false } : s
      );
      setNotice('Page is now offline.');
    } catch (e) {
      setNotice(isApiError(e) ? e.message : 'Could not unpublish.');
    } finally {
      setConfirmUnpublish(false);
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
    return <EditorSkeleton />;
  }

  const published = site.status === 'published';
  const state = pageState(site);
  const shortUrl = site.short_code ? site.public_url.replace(/\/m\/.*$/, `/${site.short_code}`) : '';
  const shownUrl = site.public_url.replace(/^https?:\/\//, '');
  const kind = pageKind(site.template);
  const kindName = PAGE_KINDS.find((k) => k.id === kind)?.name ?? 'Page';
  const canUseForms = !!entitlements?.can_use_forms;
  const editing = tab === 'build' || tab === 'design';

  return (
    <div className="-mx-4 -mt-6 sm:-mx-6 lg:-mx-8">
      {/* Editor header: sits under the workspace header */}
      <div className="sticky top-16 z-20 border-b border-outline-variant/20 bg-surface-container-lowest/95 backdrop-blur-md dark:bg-[#07090e]/95">
        <div className="flex flex-wrap items-center gap-x-3 gap-y-2 px-4 pt-3 sm:px-6 lg:px-8">
          <nav aria-label="Breadcrumb" className="flex min-w-0 items-center gap-1 text-sm">
            <Link href="/microsites" className={cx('shrink-0 rounded text-muted-foreground hover:text-foreground', focusRing)}>
              Pages
            </Link>
            <ChevronRightIcon className="h-3.5 w-3.5 shrink-0 text-muted-foreground" />
            <input
              value={title}
              onChange={(e) => editTitle(e.target.value)}
              maxLength={120}
              aria-label="Page name"
              className="min-w-0 w-40 sm:w-56 rounded-lg bg-transparent px-1.5 py-0.5 font-hanken text-base font-bold text-foreground hover:bg-surface-container focus:bg-surface-container focus:outline-none focus:ring-2 focus:ring-primary/40 dark:hover:bg-zinc-900 dark:focus:bg-zinc-900"
            />
          </nav>
          <StatusDot state={state} className="shrink-0" />
          <Badge tone="muted" className="hidden md:inline-flex">
            {kindName}
          </Badge>

          <div className="ml-auto flex items-center gap-2">
            <span
              role="status"
              className={cx(
                'hidden text-xs sm:inline',
                saveState === 'error' || saveState === 'conflict' ? 'text-destructive' : 'text-muted-foreground'
              )}
            >
              {SAVE_LABELS[saveState]}
            </span>

            <div className="hidden items-center rounded-xl border border-outline-variant/30 bg-surface-container dark:bg-zinc-900 xl:flex">
              <span className="max-w-[16rem] truncate px-3 font-mono text-xs text-muted-foreground">{shownUrl}</span>
              <IconButton size="sm" label={copied ? 'Link copied' : 'Copy page link'} onClick={copyLink}>
                {copied ? <CheckIcon className="h-4 w-4 text-emerald-500" /> : <ClipboardDocumentIcon className="h-4 w-4" />}
              </IconButton>
            </div>

            <SharePopover
              publicUrl={site.public_url}
              shortUrl={shortUrl}
              published={published}
              copied={copied}
              onCopy={copyLink}
              onQR={() => setShowQR(true)}
              onCreateShortLink={createShortLink}
            />

            {published && state === 'published' && saveState === 'saved' && !publishing ? (
              <span
                className="inline-flex h-9 items-center gap-1.5 rounded-xl border border-emerald-500/25 bg-emerald-500/10 px-3 text-sm font-semibold text-emerald-700 dark:text-emerald-400"
                title="The live page matches your latest changes"
              >
                <CheckIcon className="h-4 w-4" />
                <span className="hidden sm:inline">Up to date</span>
              </span>
            ) : (
              <Button variant="primary" onClick={publish} disabled={publishing || saveState === 'conflict'}>
                {publishing ? 'Publishing…' : !published ? 'Publish' : 'Update live page'}
              </Button>
            )}

            <Menu as="div">
              <MenuButton
                aria-label="More page actions"
                className={cx(
                  'inline-flex h-9 w-9 items-center justify-center rounded-xl text-muted-foreground hover:bg-surface-container hover:text-foreground dark:hover:bg-zinc-900',
                  focusRing
                )}
              >
                <EllipsisHorizontalIcon className="h-5 w-5" />
              </MenuButton>
              <MenuItems
                anchor="bottom end"
                className="z-50 w-56 rounded-xl border border-outline-variant/30 bg-surface-container-lowest p-1 text-sm shadow-xl [--anchor-gap:6px] focus:outline-none dark:bg-zinc-950"
              >
                {published && (
                  <MenuItem>
                    <a href={site.public_url} target="_blank" rel="noopener noreferrer" className={menuItemClass}>
                      <ArrowTopRightOnSquareIcon className="h-4 w-4 text-muted-foreground" /> Open live page
                    </a>
                  </MenuItem>
                )}
                <MenuItem>
                  <button type="button" onClick={() => setShowQR(true)} className={menuItemClass}>
                    <QrCodeIcon className="h-4 w-4 text-muted-foreground" /> QR code
                  </button>
                </MenuItem>
                {published && (
                  <MenuItem>
                    <button type="button" onClick={() => setConfirmUnpublish(true)} className={menuItemClass}>
                      <EyeSlashIcon className="h-4 w-4 text-muted-foreground" /> Unpublish
                    </button>
                  </MenuItem>
                )}
                <div className="my-1 h-px bg-outline-variant/20" />
                <MenuItem>
                  <Link href="/microsites" className={menuItemClass}>
                    <Squares2X2Icon className="h-4 w-4 text-muted-foreground" /> All pages
                  </Link>
                </MenuItem>
              </MenuItems>
            </Menu>
          </div>
        </div>

        <div className="flex items-center justify-between gap-3 px-4 py-2.5 sm:px-6 lg:px-8">
          <Segmented<Tab>
            label="Editor sections"
            value={tab}
            onChange={setTab}
            options={TABS.map((t) => ({ value: t, label: TAB_LABELS[t] }))}
          />
          {editing && (
            <Segmented<'edit' | 'preview'>
              label="Show"
              value={mobileView}
              onChange={setMobileView}
              className="lg:hidden"
              options={[
                {
                  value: 'edit',
                  label: (
                    <>
                      <PencilSquareIcon className="h-4 w-4" />
                      <span className="sr-only sm:not-sr-only">Edit</span>
                    </>
                  ),
                },
                {
                  value: 'preview',
                  label: (
                    <>
                      <EyeIcon className="h-4 w-4" />
                      <span className="sr-only sm:not-sr-only">Preview</span>
                    </>
                  ),
                },
              ]}
            />
          )}
        </div>
      </div>

      {(saveError || notice || saveState === 'conflict') && (
        <div className="grid gap-2 px-4 pt-4 sm:px-6 lg:px-8">
          {saveState === 'conflict' && (
            <p role="alert" className="rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-2.5 text-sm text-destructive">
              This page was changed in another tab or by another session.{' '}
              <button type="button" onClick={load} className="font-semibold underline underline-offset-2">
                Reload the latest version
              </button>
            </p>
          )}
          {saveError && saveState === 'error' && (
            <p role="alert" className="rounded-xl border border-destructive/20 bg-destructive/10 px-4 py-2.5 text-sm text-destructive">
              {saveError}
            </p>
          )}
          {notice && (
            <p role="status" className="flex items-center justify-between gap-3 rounded-xl border border-outline-variant/30 bg-surface-container px-4 py-2.5 text-sm text-foreground dark:bg-zinc-900">
              <span className="break-all">{notice}</span>
              <button type="button" onClick={() => setNotice('')} className={cx('rounded text-xs text-muted-foreground hover:text-foreground', focusRing)}>
                Dismiss
              </button>
            </p>
          )}
        </div>
      )}

      <div className="px-4 py-6 sm:px-6 lg:px-8">
        {tab === 'analytics' ? (
          <TrackPanel micrositeId={id} maxDays={entitlements?.analytics_days} />
        ) : tab === 'inbox' ? (
          <InboxPanel micrositeId={id} blocks={content.blocks} canUseForms={canUseForms} />
        ) : (
          <div className="grid grid-cols-1 gap-6 lg:grid-cols-[minmax(340px,440px)_minmax(0,1fr)] xl:gap-8">
            {/* Settings column */}
            <div className={cx('min-w-0', mobileView === 'preview' && 'hidden lg:block')}>
              {tab === 'build' ? (
                <div className="grid grid-cols-1 gap-4">
                  <div className="flex items-center justify-between gap-3">
                    <h2 className="text-sm font-semibold text-foreground">
                      Blocks <span className="font-normal tabular-nums text-muted-foreground">{content.blocks.length}</span>
                    </h2>
                    <BlockPicker onAdd={addBlock} canUseForms={canUseForms} />
                  </div>
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
                    canUseForms={canUseForms}
                  />
                  {content.blocks.length > 3 && (
                    <div className="flex justify-center">
                      <BlockPicker onAdd={addBlock} canUseForms={canUseForms} />
                    </div>
                  )}
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

            {/* Live preview stage */}
            <div className={cx('min-w-0 lg:sticky lg:top-[11.5rem] lg:self-start', mobileView === 'edit' && 'hidden lg:block')}>
              <div className="rounded-2xl border border-outline-variant/30 bg-surface-container/60 px-3 py-4 [background-image:radial-gradient(hsl(var(--outline-variant)/0.45)_1px,transparent_1px)] [background-size:14px_14px] dark:bg-zinc-900/50">
                <PreviewFrame title={title} slug={site.slug} content={content} reserve={300} />
              </div>
            </div>
          </div>
        )}
      </div>

      <QRCodeModal
        isOpen={showQR}
        onClose={() => setShowQR(false)}
        value={`${site.public_url}?src=qr`}
        title="QR code for your page"
        description="Scans are counted as “QR code” traffic in Analytics."
      />

      <ConfirmDialog
        open={confirmUnpublish}
        title="Take this page offline?"
        confirmLabel="Unpublish"
        onConfirm={unpublish}
        onClose={() => setConfirmUnpublish(false)}
      >
        <p>
          <span className="font-mono text-foreground">{shownUrl}</span> will show “not found” until you publish again. Your draft and
          analytics stay.
        </p>
      </ConfirmDialog>
    </div>
  );
}

const menuItemClass =
  'flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-[13px] text-foreground data-[focus]:bg-surface-container dark:data-[focus]:bg-zinc-900';

function EditorSkeleton() {
  return (
    <div className="-mx-4 -mt-6 sm:-mx-6 lg:-mx-8" aria-busy="true" aria-label="Loading the editor">
      <div className="border-b border-outline-variant/20 px-4 py-4 sm:px-6 lg:px-8">
        <div className="h-5 w-64 animate-pulse rounded-lg bg-surface-container dark:bg-zinc-900" />
        <div className="mt-3 h-8 w-80 animate-pulse rounded-xl bg-surface-container dark:bg-zinc-900" />
      </div>
      <div className="grid gap-6 px-4 py-6 sm:px-6 lg:grid-cols-[440px_1fr] lg:px-8">
        <div className="space-y-3">
          {[0, 1, 2].map((i) => (
            <div key={i} className="h-16 animate-pulse rounded-2xl bg-surface-container dark:bg-zinc-900" />
          ))}
        </div>
        <div className="hidden h-[640px] animate-pulse rounded-2xl bg-surface-container dark:bg-zinc-900 lg:block" />
      </div>
    </div>
  );
}

function SharePopover({
  publicUrl,
  shortUrl,
  published,
  copied,
  onCopy,
  onQR,
  onCreateShortLink,
}: {
  publicUrl: string;
  shortUrl: string;
  published: boolean;
  copied: boolean;
  onCopy: () => void;
  onQR: () => void;
  onCreateShortLink: () => void;
}) {
  const [shortCopied, setShortCopied] = useState(false);
  const copyShort = () =>
    navigator.clipboard.writeText(shortUrl).then(() => {
      setShortCopied(true);
      setTimeout(() => setShortCopied(false), 1800);
    });
  return (
    <Popover>
      <PopoverButton
        className={cx(
          'inline-flex h-9 items-center gap-2 rounded-xl border border-outline-variant/30 bg-surface-container-lowest px-3 text-sm font-semibold text-foreground hover:bg-surface-container dark:bg-zinc-900 dark:hover:bg-zinc-800',
          focusRing
        )}
      >
        <ShareIcon className="h-4 w-4" />
        <span className="hidden sm:inline">Share</span>
      </PopoverButton>
      <PopoverPanel
        anchor="bottom end"
        className="z-50 w-80 rounded-2xl border border-outline-variant/30 bg-surface-container-lowest p-4 shadow-2xl [--anchor-gap:8px] dark:bg-zinc-950"
      >
        <p className="text-sm font-semibold text-foreground">Share this page</p>
        {!published && <p className="mt-1 text-xs text-muted-foreground">Publish first: until then the link shows “not found”.</p>}

        <label className="mt-3 block text-xs font-semibold text-muted-foreground" htmlFor="share-page-link">
          Page link
        </label>
        <div className="mt-1 flex items-center gap-1 rounded-xl border border-outline-variant/30 bg-surface-container pl-3 dark:bg-zinc-900">
          <input id="share-page-link" readOnly value={publicUrl.replace(/^https?:\/\//, '')} className="min-w-0 flex-1 bg-transparent py-2 font-mono text-xs text-foreground outline-none" />
          <IconButton size="sm" label={copied ? 'Copied' : 'Copy page link'} onClick={onCopy}>
            {copied ? <CheckIcon className="h-4 w-4 text-emerald-500" /> : <ClipboardDocumentIcon className="h-4 w-4" />}
          </IconButton>
        </div>

        <p className="mt-3 text-xs font-semibold text-muted-foreground">Short link</p>
        {shortUrl ? (
          <div className="mt-1 flex items-center gap-1 rounded-xl border border-outline-variant/30 bg-surface-container pl-3 dark:bg-zinc-900">
            <span className="min-w-0 flex-1 truncate py-2 font-mono text-xs text-foreground">{shortUrl.replace(/^https?:\/\//, '')}</span>
            <IconButton size="sm" label={shortCopied ? 'Copied' : 'Copy short link'} onClick={copyShort}>
              {shortCopied ? <CheckIcon className="h-4 w-4 text-emerald-500" /> : <ClipboardDocumentIcon className="h-4 w-4" />}
            </IconButton>
          </div>
        ) : (
          <Button size="sm" className="mt-1 w-full" onClick={onCreateShortLink}>
            <LinkIcon className="h-4 w-4" />
            Create a short link
          </Button>
        )}

        <div className="mt-3 grid grid-cols-2 gap-2">
          <Button size="sm" onClick={onQR}>
            <QrCodeIcon className="h-4 w-4" />
            QR code
          </Button>
          {published ? (
            <a
              href={publicUrl}
              target="_blank"
              rel="noopener noreferrer"
              className={cx(
                'inline-flex h-8 items-center justify-center gap-1.5 rounded-xl border border-outline-variant/30 bg-surface-container-lowest px-3 text-xs font-semibold text-foreground hover:bg-surface-container dark:bg-zinc-900 dark:hover:bg-zinc-800',
                focusRing
              )}
            >
              <ArrowTopRightOnSquareIcon className="h-4 w-4" />
              Open
            </a>
          ) : (
            <span className="inline-flex h-8 items-center justify-center rounded-xl text-xs text-muted-foreground">Not live yet</span>
          )}
        </div>
      </PopoverPanel>
    </Popover>
  );
}

function BlockPicker({ onAdd, canUseForms }: { onAdd: (type: BlockType) => void; canUseForms: boolean }) {
  const [q, setQ] = useState('');
  const query = q.trim().toLowerCase();
  return (
    <Popover>
      <PopoverButton
        className={cx(
          'inline-flex h-8 items-center gap-1.5 rounded-xl bg-primary px-3 text-xs font-semibold text-white shadow-sm hover:bg-primary-600',
          focusRing
        )}
      >
        <PlusIcon className="h-4 w-4 stroke-[2.5]" />
        Add block
      </PopoverButton>
      <PopoverPanel
        anchor="bottom end"
        className="z-50 flex max-h-[min(70vh,560px)] w-[min(92vw,380px)] flex-col overflow-hidden rounded-2xl border border-outline-variant/30 bg-surface-container-lowest shadow-2xl [--anchor-gap:8px] dark:bg-zinc-950"
      >
        {({ close }) => (
          <>
            <div className="border-b border-outline-variant/20 p-3">
              <div className="relative">
                <MagnifyingGlassIcon className="pointer-events-none absolute left-3 top-1/2 h-4 w-4 -translate-y-1/2 text-muted-foreground" />
                <input
                  autoFocus
                  value={q}
                  onChange={(e) => setQ(e.target.value)}
                  placeholder="Find a block"
                  aria-label="Find a block"
                  className="w-full rounded-xl border border-outline-variant/30 bg-surface-container py-2 pl-9 pr-3 text-sm text-foreground placeholder:text-muted-foreground focus:outline-none focus:ring-2 focus:ring-primary/40 dark:bg-zinc-900"
                />
              </div>
            </div>
            <div className="overflow-y-auto p-2">
              {BLOCK_GROUPS.map((g) => {
                const entries = BLOCK_CATALOG.filter(
                  (b) => b.group === g.id && (!query || b.name.toLowerCase().includes(query) || b.description.toLowerCase().includes(query))
                );
                if (entries.length === 0) return null;
                return (
                  <div key={g.id} className="mb-2">
                    <p className="px-2 pb-1 pt-1.5 text-[11px] font-semibold text-muted-foreground">{g.name}</p>
                    <ul className="grid gap-0.5">
                      {entries.map((b) => {
                        const locked = b.requiresForms && !canUseForms;
                        return (
                          <li key={b.type}>
                            {locked ? (
                              <Link
                                href="/billing"
                                className={cx('flex items-start gap-2.5 rounded-xl px-2.5 py-2 hover:bg-surface-container dark:hover:bg-zinc-900', focusRing)}
                              >
                                <LockClosedIcon className="mt-0.5 h-4 w-4 shrink-0 text-muted-foreground" />
                                <span>
                                  <span className="block text-sm font-medium text-muted-foreground">{b.name}</span>
                                  <span className="block text-xs text-muted-foreground">Premium, Growth and Enterprise. Upgrade to add.</span>
                                </span>
                              </Link>
                            ) : (
                              <button
                                type="button"
                                onClick={() => {
                                  onAdd(b.type);
                                  setQ('');
                                  close();
                                }}
                                className={cx('flex w-full items-start gap-2.5 rounded-xl px-2.5 py-2 text-left hover:bg-primary/5', focusRing)}
                              >
                                <PlusIcon className="mt-0.5 h-4 w-4 shrink-0 text-primary" />
                                <span>
                                  <span className="block text-sm font-medium text-foreground">{b.name}</span>
                                  <span className="block text-xs text-muted-foreground">{b.description}</span>
                                </span>
                              </button>
                            )}
                          </li>
                        );
                      })}
                    </ul>
                  </div>
                );
              })}
              {BLOCK_CATALOG.every((b) => query && !b.name.toLowerCase().includes(query) && !b.description.toLowerCase().includes(query)) && (
                <p className="px-2 py-6 text-center text-sm text-muted-foreground">No block matches “{q.trim()}”.</p>
              )}
            </div>
          </>
        )}
      </PopoverPanel>
    </Popover>
  );
}
