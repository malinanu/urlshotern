type CacheEntry = { data: unknown; ts: number };
const cache = new Map<string, CacheEntry>();
const TTL = 60_000; // 60 seconds

export function getCached<T>(key: string): T | null {
  const entry = cache.get(key);
  if (!entry) return null;
  if (Date.now() - entry.ts > TTL) { cache.delete(key); return null; }
  return entry.data as T;
}

export function setCached(key: string, data: unknown): void {
  cache.set(key, { data, ts: Date.now() });
}

export function prefetch(key: string, fetchFn: () => Promise<unknown>): void {
  if (getCached(key) !== null) return; // already fresh
  fetchFn().then(data => setCached(key, data)).catch(() => {});
}
