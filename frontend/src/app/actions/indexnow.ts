'use server';

import { submitToIndexNow } from '@/lib/indexnow';

const SITE_URL = 'https://trunc.site';

/**
 * Notify IndexNow (Bing, Yandex, Naver, …) that one or more URLs were published or
 * updated, so they re-crawl immediately instead of waiting for a scheduled crawl.
 *
 * Runs server-side (a Server Action) so the request to api.indexnow.org is not
 * subject to browser CORS. Best-effort: failures are returned, never thrown, so a
 * publish action is never blocked by an indexing hiccup.
 *
 * @param paths Absolute URLs, or site-relative paths like "/pages/my-slug".
 */
export async function notifyIndexNow(
  paths: string[],
): Promise<{ ok: boolean; submitted: number; error?: string }> {
  const urls = paths
    .filter(Boolean)
    .map((p) =>
      p.startsWith('http') ? p : `${SITE_URL}${p.startsWith('/') ? '' : '/'}${p}`,
    );

  if (urls.length === 0) return { ok: true, submitted: 0 };

  try {
    await submitToIndexNow(urls);
    return { ok: true, submitted: urls.length };
  } catch (e) {
    return {
      ok: false,
      submitted: 0,
      error: e instanceof Error ? e.message : 'IndexNow submission failed',
    };
  }
}
