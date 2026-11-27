'use client';

import { useState } from 'react';
import Link from 'next/link';
import { Menu, MenuButton, MenuItem, MenuItems } from '@headlessui/react';
import {
  ArrowTopRightOnSquareIcon,
  ChartBarIcon,
  CheckIcon,
  ClipboardDocumentIcon,
  DocumentDuplicateIcon,
  EllipsisHorizontalIcon,
  InboxIcon,
  PencilSquareIcon,
  QrCodeIcon,
  RocketLaunchIcon,
  TrashIcon,
  EyeSlashIcon,
} from '@heroicons/react/24/outline';
import { API_ENDPOINTS } from '@/lib/api-config';
import { PAGE_KINDS, pageKind, pageState, timeAgo, type MicrositeSummary } from '@/lib/microsites';
import PhoneFrame from '../PhoneFrame';
import { Badge, IconButton, Sparkline, StatusDot, cx, focusRing, panelClass } from '@/components/workspace/ui';
import { KindIcon } from './kind';

export interface PageCardActions {
  onQR: (page: MicrositeSummary) => void;
  onDuplicate: (page: MicrositeSummary) => void;
  onTogglePublish: (page: MicrositeSummary) => void;
  onDelete: (page: MicrositeSummary) => void;
}

export default function PageCard({
  page,
  baseUrl,
  getToken,
  canDuplicate,
  actions,
}: {
  page: MicrositeSummary;
  baseUrl: string;
  getToken: () => string | null;
  canDuplicate: boolean;
  actions: PageCardActions;
}) {
  const [copied, setCopied] = useState(false);
  const kind = pageKind(page.template);
  const kindName = PAGE_KINDS.find((k) => k.id === kind)?.name ?? 'Page';
  const state = pageState(page);
  const url = `${baseUrl}${page.slug}`;
  const shownUrl = url.replace(/^https?:\/\//, '');
  const editHref = `/microsites/edit?id=${page.id}`;
  const published = page.status === 'published';

  const copy = async () => {
    try {
      await navigator.clipboard.writeText(url);
      setCopied(true);
      setTimeout(() => setCopied(false), 1800);
    } catch {
      // clipboard blocked: nothing to do
    }
  };

  return (
    <li className={cx(panelClass, 'group relative flex flex-col overflow-hidden transition-colors hover:border-outline-variant/60 focus-within:border-primary/50')}>
      {/* Stage: the page itself */}
      <div className="relative flex justify-center pt-5 h-[232px] overflow-hidden bg-surface-container/60 dark:bg-zinc-900/50 border-b border-outline-variant/20 [background-image:radial-gradient(hsl(var(--outline-variant)/0.45)_1px,transparent_1px)] [background-size:14px_14px]">
        <PhoneFrame
          cacheKey={`trunc:thumb:${page.id}:${page.version}:${page.status}:${page.published_at ?? ''}`}
          url={API_ENDPOINTS.microsites.thumbnail(page.id)}
          getToken={getToken}
          width={150}
          fallback={
            <>
              <KindIcon kind={kind} className="h-6 w-6" />
              <span className="px-3 text-center text-[11px] font-medium leading-tight line-clamp-2 text-foreground/80">{page.title}</span>
            </>
          }
        />
        {/* Quick actions: always shown on touch screens, on hover or focus elsewhere */}
        <div className="absolute top-2.5 right-2.5 z-10 flex items-center gap-1 rounded-xl border border-outline-variant/30 bg-surface-container-lowest/95 dark:bg-zinc-950/95 p-0.5 shadow-sm transition-opacity [@media(hover:hover)]:opacity-0 group-hover:opacity-100 group-focus-within:opacity-100">
          <IconButton size="sm" label={copied ? 'Link copied' : `Copy link ${shownUrl}`} onClick={copy}>
            {copied ? <CheckIcon className="h-4 w-4 text-emerald-500" /> : <ClipboardDocumentIcon className="h-4 w-4" />}
          </IconButton>
          <IconButton size="sm" label={`QR code for ${page.title}`} onClick={() => actions.onQR(page)}>
            <QrCodeIcon className="h-4 w-4" />
          </IconButton>
          {published && (
            <a
              href={url}
              target="_blank"
              rel="noopener noreferrer"
              aria-label={`Open the live page ${shownUrl}`}
              title="Open live page"
              className={cx('inline-flex h-8 w-8 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900', focusRing)}
            >
              <ArrowTopRightOnSquareIcon className="h-4 w-4" />
            </a>
          )}
        </div>
      </div>

      {/* Details */}
      <div className="flex flex-1 flex-col gap-3 p-4">
        <div className="flex items-center justify-between gap-2">
          <StatusDot state={state} />
          <Badge tone="muted">
            <KindIcon kind={kind} className="h-3 w-3" />
            {kindName}
          </Badge>
        </div>

        <div className="min-w-0">
          <h3 className="font-hanken font-bold text-[15px] leading-snug text-foreground truncate">
            <Link href={editHref} className={cx('rounded-md after:absolute after:inset-0 after:content-[""]', focusRing)}>
              {page.title}
            </Link>
          </h3>
          <button
            type="button"
            onClick={copy}
            className={cx('relative z-10 mt-0.5 max-w-full truncate rounded-md text-left font-mono text-xs text-muted-foreground hover:text-foreground', focusRing)}
            title="Copy link"
          >
            {copied ? 'Copied to clipboard' : shownUrl}
          </button>
        </div>

        <div className="mt-auto flex items-end justify-between gap-3">
          <div className="min-w-0">
            <div className="flex items-baseline gap-1.5">
              <span className="text-lg font-bold tabular-nums text-foreground">{page.views_30d.toLocaleString()}</span>
              <span className="text-xs text-muted-foreground">views · 30 days</span>
            </div>
            <p className="text-xs text-muted-foreground">
              {page.submissions_30d > 0 && (
                <>
                  <span className="font-semibold text-foreground tabular-nums">{page.submissions_30d}</span>{' '}
                  {page.submissions_30d === 1 ? 'lead' : 'leads'} ·{' '}
                </>
              )}
              Edited {timeAgo(page.updated_at)}
            </p>
          </div>
          <Sparkline
            values={page.views_14d?.length ? page.views_14d : new Array(14).fill(0)}
            label={`Views over the last 14 days: ${(page.views_14d ?? []).reduce((a, b) => a + b, 0)}`}
            className="shrink-0"
          />
        </div>
      </div>

      {/* Footer actions */}
      <div className="relative z-10 flex items-center gap-1.5 border-t border-outline-variant/20 px-3 py-2.5">
        <Link
          href={editHref}
          className={cx(
            'inline-flex h-8 items-center gap-1.5 rounded-xl bg-primary px-3 text-xs font-semibold text-white shadow-sm hover:bg-primary-600',
            focusRing
          )}
        >
          <PencilSquareIcon className="h-4 w-4" />
          Edit
        </Link>
        <Link
          href={`${editHref}&tab=analytics`}
          className={cx(
            'inline-flex h-8 items-center gap-1.5 rounded-xl px-3 text-xs font-semibold text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900',
            focusRing
          )}
        >
          <ChartBarIcon className="h-4 w-4" />
          Analytics
        </Link>

        <Menu as="div" className="ml-auto">
          <MenuButton
            aria-label={`More actions for ${page.title}`}
            className={cx(
              'inline-flex h-8 w-8 items-center justify-center rounded-xl text-muted-foreground hover:text-foreground hover:bg-surface-container dark:hover:bg-zinc-900 data-[open]:bg-surface-container dark:data-[open]:bg-zinc-900',
              focusRing
            )}
          >
            <EllipsisHorizontalIcon className="h-5 w-5" />
          </MenuButton>
          <MenuItems
            anchor="bottom end"
            className="z-50 w-52 rounded-xl border border-outline-variant/30 bg-surface-container-lowest dark:bg-zinc-950 p-1 text-sm shadow-xl [--anchor-gap:6px] focus:outline-none"
          >
            <MenuLink href={`${editHref}&tab=inbox`} icon={InboxIcon}>
              Inbox{page.submissions_30d > 0 ? ` (${page.submissions_30d})` : ''}
            </MenuLink>
            <MenuAction icon={ClipboardDocumentIcon} onClick={copy}>
              Copy link
            </MenuAction>
            <MenuAction icon={QrCodeIcon} onClick={() => actions.onQR(page)}>
              QR code
            </MenuAction>
            <MenuAction
              icon={DocumentDuplicateIcon}
              onClick={() => actions.onDuplicate(page)}
              disabled={!canDuplicate}
              hint={canDuplicate ? undefined : 'Page limit reached'}
            >
              Duplicate
            </MenuAction>
            <MenuAction icon={published ? EyeSlashIcon : RocketLaunchIcon} onClick={() => actions.onTogglePublish(page)}>
              {published ? 'Unpublish' : 'Publish'}
            </MenuAction>
            <div className="my-1 h-px bg-outline-variant/20" />
            <MenuAction icon={TrashIcon} onClick={() => actions.onDelete(page)} danger>
              Delete…
            </MenuAction>
          </MenuItems>
        </Menu>
      </div>
    </li>
  );
}

const itemClass =
  'flex w-full items-center gap-2.5 rounded-lg px-2.5 py-2 text-left text-[13px] data-[focus]:bg-surface-container dark:data-[focus]:bg-zinc-900 data-[disabled]:opacity-50';

function MenuAction({
  icon: Icon,
  onClick,
  children,
  danger,
  disabled,
  hint,
}: {
  icon: typeof TrashIcon;
  onClick: () => void;
  children: React.ReactNode;
  danger?: boolean;
  disabled?: boolean;
  hint?: string;
}) {
  return (
    <MenuItem disabled={disabled}>
      <button type="button" onClick={onClick} className={cx(itemClass, danger ? 'text-destructive' : 'text-foreground')}>
        <Icon className={cx('h-4 w-4 shrink-0', danger ? 'text-destructive' : 'text-muted-foreground')} />
        <span className="flex-1">{children}</span>
        {hint && <span className="text-[11px] text-muted-foreground">{hint}</span>}
      </button>
    </MenuItem>
  );
}

function MenuLink({ href, icon: Icon, children }: { href: string; icon: typeof TrashIcon; children: React.ReactNode }) {
  return (
    <MenuItem>
      <Link href={href} className={cx(itemClass, 'text-foreground')}>
        <Icon className="h-4 w-4 shrink-0 text-muted-foreground" />
        <span className="flex-1">{children}</span>
      </Link>
    </MenuItem>
  );
}
