'use client';

import Link from 'next/link';
import { ArrowRightIcon } from '@heroicons/react/24/outline';
import { API_ENDPOINTS } from '@/lib/api-config';
import { KIND_TEMPLATE, PAGE_KINDS, browserTimeZone, type PageKind } from '@/lib/microsites';
import PhoneFrame from '../PhoneFrame';
import { cx, focusRing, panelClass } from '@/components/workspace/ui';
import { KindIcon } from './kind';

const STARTERS: PageKind[] = ['card', 'links', 'shop', 'event'];

/** First-run empty state: one starter per page type, each showing its real template. */
export default function PagesStarter({ getToken, name }: { getToken: () => string | null; name: string }) {
  return (
    <section className={cx(panelClass, 'p-6 sm:p-8')} aria-labelledby="starter-title">
      <div className="max-w-2xl">
        <h2 id="starter-title" className="font-hanken text-xl font-bold text-foreground">
          Make your first page
        </h2>
        <p className="mt-1.5 text-sm text-muted-foreground">
          Each page gets its own link and QR code, and you can change every word, colour and block later. Pick what it is for:
        </p>
      </div>

      <ul className="mt-6 grid gap-4 sm:grid-cols-2 xl:grid-cols-4">
        {STARTERS.map((kind) => {
          const info = PAGE_KINDS.find((k) => k.id === kind)!;
          const template = KIND_TEMPLATE[kind];
          const q = new URLSearchParams({ name, tz: browserTimeZone() });
          return (
            <li key={kind}>
              <Link
                href={`/microsites/new?kind=${kind}`}
                className={cx(
                  'group flex h-full flex-col overflow-hidden rounded-2xl border border-outline-variant/30 bg-surface-container/50 dark:bg-zinc-900/40 transition-colors hover:border-primary/50 hover:bg-primary/5',
                  focusRing
                )}
              >
                <div className="flex h-[210px] justify-center overflow-hidden pt-5">
                  <PhoneFrame
                    cacheKey={`trunc:tpl:${template}:${name}`}
                    url={`${API_ENDPOINTS.microsites.templatePreview(template)}?${q}`}
                    getToken={getToken}
                    width={140}
                    fallback={<KindIcon kind={kind} className="h-6 w-6" />}
                    className="transition-transform duration-200 group-hover:-translate-y-1"
                  />
                </div>
                <div className="flex flex-1 flex-col gap-1 border-t border-outline-variant/20 p-4">
                  <span className="flex items-center gap-2 font-hanken text-sm font-bold text-foreground">
                    <KindIcon kind={kind} className="h-4 w-4 text-primary" />
                    {info.name}
                  </span>
                  <span className="text-xs text-muted-foreground">{info.description}</span>
                  <span className="mt-auto inline-flex items-center gap-1 pt-2 text-xs font-semibold text-primary">
                    Start <ArrowRightIcon className="h-3.5 w-3.5 transition-transform group-hover:translate-x-0.5" />
                  </span>
                </div>
              </Link>
            </li>
          );
        })}
      </ul>

      <p className="mt-5 text-sm text-muted-foreground">
        Or{' '}
        <Link href="/microsites/new?kind=other" className={cx('rounded font-medium text-primary hover:underline', focusRing)}>
          start from a blank page
        </Link>{' '}
        and add any blocks you like.
      </p>
    </section>
  );
}
