'use client';

/*
 * Rendered-HTML loader for page and template thumbnails. Fetches run through a
 * small queue so a long list never bursts past the API rate limit, retry on
 * 429, and are cached in memory and in sessionStorage (when available) by key.
 */

import { apiFetch, isApiError } from '@/lib/microsites';

const MAX_CONCURRENT = 3;
const memory = new Map<string, string>();
const pending = new Map<string, Promise<string>>();
const waiting: (() => void)[] = [];
let running = 0;

export function cachedHtml(key: string): string | undefined {
  const hit = memory.get(key);
  if (hit !== undefined) return hit;
  try {
    const v = sessionStorage.getItem(key);
    if (v !== null) {
      memory.set(key, v);
      return v;
    }
  } catch {
    // storage unavailable: memory only
  }
  return undefined;
}

function store(key: string, html: string) {
  memory.set(key, html);
  try {
    sessionStorage.setItem(key, html);
  } catch {
    // quota exceeded or storage unavailable: memory only
  }
}

async function queued<T>(task: () => Promise<T>): Promise<T> {
  if (running >= MAX_CONCURRENT) await new Promise<void>((r) => waiting.push(r));
  running++;
  try {
    return await task();
  } finally {
    running--;
    waiting.shift()?.();
  }
}

const sleep = (ms: number) => new Promise((r) => setTimeout(r, ms));

/** Loads `{ html }` from url once per key. */
export function loadHtml(key: string, url: string, token: string | null): Promise<string> {
  const hit = cachedHtml(key);
  if (hit !== undefined) return Promise.resolve(hit);
  const inflight = pending.get(key);
  if (inflight) return inflight;
  const p = queued(async () => {
    for (let attempt = 0; ; attempt++) {
      try {
        const res = await apiFetch<{ html: string }>(url, token);
        store(key, res.html);
        return res.html;
      } catch (e) {
        if (isApiError(e) && e.status === 429 && attempt < 3) {
          await sleep(1000 * (attempt + 1));
          continue;
        }
        throw e;
      }
    }
  }).finally(() => pending.delete(key));
  pending.set(key, p);
  return p;
}
