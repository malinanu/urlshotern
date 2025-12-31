/**
 * Offline Sync Manager with IndexedDB
 * Handles offline data synchronization for URL shortener
 */

import { openDB, DBSchema, IDBPDatabase } from 'idb';

// Define the database schema
interface URLShortenerDB extends DBSchema {
  'offline-urls': {
    key: string;
    value: {
      id: string;
      url: string;
      customAlias?: string;
      expiresAt?: string;
      timestamp: number;
      token: string;
      synced: boolean;
    };
  };
  'cached-urls': {
    key: string;
    value: {
      id: string;
      shortCode: string;
      originalUrl: string;
      shortUrl: string;
      createdAt: string;
      clickCount: number;
      lastAccessed: number;
    };
  };
  'analytics-cache': {
    key: string;
    value: {
      shortCode: string;
      data: any;
      timestamp: number;
      expiresAt: number;
    };
  };
  'pending-actions': {
    key: string;
    value: {
      id: string;
      type: 'create' | 'update' | 'delete';
      endpoint: string;
      method: string;
      body: any;
      headers: Record<string, string>;
      timestamp: number;
      retryCount: number;
    };
  };
}

const DB_NAME = 'URLShortenerDB';
const DB_VERSION = 2;

// Singleton database instance
let dbInstance: IDBPDatabase<URLShortenerDB> | null = null;

/**
 * Initialize and get the IndexedDB instance
 */
export async function getDB(): Promise<IDBPDatabase<URLShortenerDB>> {
  if (dbInstance) {
    return dbInstance;
  }

  dbInstance = await openDB<URLShortenerDB>(DB_NAME, DB_VERSION, {
    upgrade(db, oldVersion, newVersion, transaction) {
      // Create object stores if they don't exist
      if (!db.objectStoreNames.contains('offline-urls')) {
        db.createObjectStore('offline-urls', { keyPath: 'id' });
      }

      if (!db.objectStoreNames.contains('cached-urls')) {
        const urlStore = db.createObjectStore('cached-urls', { keyPath: 'id' });
        (urlStore as any).createIndex('shortCode', 'shortCode', { unique: false });
        (urlStore as any).createIndex('lastAccessed', 'lastAccessed', { unique: false });
      }

      if (!db.objectStoreNames.contains('analytics-cache')) {
        const analyticsStore = db.createObjectStore('analytics-cache', {
          keyPath: 'shortCode',
        });
        (analyticsStore as any).createIndex('timestamp', 'timestamp', { unique: false });
        (analyticsStore as any).createIndex('expiresAt', 'expiresAt', { unique: false });
      }

      if (!db.objectStoreNames.contains('pending-actions')) {
        const actionsStore = db.createObjectStore('pending-actions', {
          keyPath: 'id',
        });
        (actionsStore as any).createIndex('timestamp', 'timestamp', { unique: false });
        (actionsStore as any).createIndex('type', 'type', { unique: false });
      }
    },
  });

  return dbInstance;
}

/**
 * Save URL creation request for offline sync
 */
export async function saveOfflineURL(
  url: string,
  customAlias: string | undefined,
  expiresAt: string | undefined,
  token: string
): Promise<string> {
  const db = await getDB();
  const id = `offline-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

  await db.add('offline-urls', {
    id,
    url,
    customAlias,
    expiresAt,
    timestamp: Date.now(),
    token,
    synced: false,
  });

  return id;
}

/**
 * Get all pending offline URLs
 */
export async function getPendingOfflineURLs() {
  const db = await getDB();
  const urls = await db.getAll('offline-urls');
  return urls.filter((url) => !url.synced);
}

/**
 * Mark offline URL as synced
 */
export async function markURLAsSynced(id: string): Promise<void> {
  const db = await getDB();
  const url = await db.get('offline-urls', id);

  if (url) {
    url.synced = true;
    await db.put('offline-urls', url);
  }
}

/**
 * Delete synced offline URL
 */
export async function deleteSyncedURL(id: string): Promise<void> {
  const db = await getDB();
  await db.delete('offline-urls', id);
}

/**
 * Cache URL data for offline access
 */
export async function cacheURL(urlData: {
  id: string;
  shortCode: string;
  originalUrl: string;
  shortUrl: string;
  createdAt: string;
  clickCount: number;
}): Promise<void> {
  const db = await getDB();

  await db.put('cached-urls', {
    ...urlData,
    lastAccessed: Date.now(),
  });
}

/**
 * Get cached URLs
 */
export async function getCachedURLs(limit: number = 50) {
  const db = await getDB();
  const tx = db.transaction('cached-urls', 'readonly');
  const index = (tx.store as any).index('lastAccessed');

  const urls = await index.getAll();
  return urls.sort((a: any, b: any) => b.lastAccessed - a.lastAccessed).slice(0, limit);
}

/**
 * Get specific cached URL by short code
 */
export async function getCachedURL(shortCode: string) {
  const db = await getDB();
  const tx = db.transaction('cached-urls', 'readonly');
  const index = (tx.store as any).index('shortCode');

  const urls = await index.getAll(shortCode);
  return urls[0] || null;
}

/**
 * Cache analytics data with expiration
 */
export async function cacheAnalytics(
  shortCode: string,
  data: any,
  ttlMinutes: number = 30
): Promise<void> {
  const db = await getDB();

  await db.put('analytics-cache', {
    shortCode,
    data,
    timestamp: Date.now(),
    expiresAt: Date.now() + ttlMinutes * 60 * 1000,
  });
}

/**
 * Get cached analytics data if not expired
 */
export async function getCachedAnalytics(shortCode: string) {
  const db = await getDB();
  const cached = await db.get('analytics-cache', shortCode);

  if (!cached) {
    return null;
  }

  // Check if expired
  if (cached.expiresAt < Date.now()) {
    await db.delete('analytics-cache', shortCode);
    return null;
  }

  return cached.data;
}

/**
 * Clear expired analytics cache
 */
export async function clearExpiredAnalyticsCache(): Promise<void> {
  const db = await getDB();
  const tx = db.transaction('analytics-cache', 'readwrite');
  const index = (tx.store as any).index('expiresAt');

  const now = Date.now();
  const cursor = await index.openCursor();

  if (cursor) {
    do {
      if (cursor.value.expiresAt < now) {
        await cursor.delete();
      }
    } while (await cursor.continue());
  }
}

/**
 * Queue an action for later sync
 */
export async function queueAction(action: {
  type: 'create' | 'update' | 'delete';
  endpoint: string;
  method: string;
  body: any;
  headers: Record<string, string>;
}): Promise<string> {
  const db = await getDB();
  const id = `action-${Date.now()}-${Math.random().toString(36).substr(2, 9)}`;

  await db.add('pending-actions', {
    id,
    ...action,
    timestamp: Date.now(),
    retryCount: 0,
  });

  return id;
}

/**
 * Get all pending actions
 */
export async function getPendingActions() {
  const db = await getDB();
  const actions = await db.getAll('pending-actions');
  return actions.sort((a, b) => a.timestamp - b.timestamp);
}

/**
 * Delete completed action
 */
export async function deleteAction(id: string): Promise<void> {
  const db = await getDB();
  await db.delete('pending-actions', id);
}

/**
 * Increment retry count for an action
 */
export async function incrementActionRetry(id: string): Promise<void> {
  const db = await getDB();
  const action = await db.get('pending-actions', id);

  if (action) {
    action.retryCount += 1;
    await db.put('pending-actions', action);
  }
}

/**
 * Sync all pending offline data
 */
export async function syncOfflineData(): Promise<{
  success: number;
  failed: number;
  errors: string[];
}> {
  const results = {
    success: 0,
    failed: 0,
    errors: [] as string[],
  };

  // Sync pending actions
  const actions = await getPendingActions();

  for (const action of actions) {
    try {
      const response = await fetch(action.endpoint, {
        method: action.method,
        headers: action.headers,
        body: JSON.stringify(action.body),
      });

      if (response.ok) {
        await deleteAction(action.id);
        results.success += 1;
      } else {
        await incrementActionRetry(action.id);
        results.failed += 1;
        results.errors.push(
          `Action ${action.id} failed: ${response.statusText}`
        );

        // Delete after 3 retries
        if (action.retryCount >= 3) {
          await deleteAction(action.id);
        }
      }
    } catch (error) {
      await incrementActionRetry(action.id);
      results.failed += 1;
      results.errors.push(`Action ${action.id} failed: ${error}`);

      // Delete after 3 retries
      if (action.retryCount >= 3) {
        await deleteAction(action.id);
      }
    }
  }

  // Sync offline URLs
  const offlineURLs = await getPendingOfflineURLs();

  for (const urlData of offlineURLs) {
    try {
      const response = await fetch('/api/v1/shorten', {
        method: 'POST',
        headers: {
          'Content-Type': 'application/json',
          Authorization: `Bearer ${urlData.token}`,
        },
        body: JSON.stringify({
          url: urlData.url,
          custom_alias: urlData.customAlias,
          expires_at: urlData.expiresAt,
        }),
      });

      if (response.ok) {
        await deleteSyncedURL(urlData.id);
        results.success += 1;
      } else {
        results.failed += 1;
        results.errors.push(
          `URL ${urlData.id} sync failed: ${response.statusText}`
        );
      }
    } catch (error) {
      results.failed += 1;
      results.errors.push(`URL ${urlData.id} sync failed: ${error}`);
    }
  }

  return results;
}

/**
 * Clear all offline data (use with caution)
 */
export async function clearAllOfflineData(): Promise<void> {
  const db = await getDB();

  await db.clear('offline-urls');
  await db.clear('cached-urls');
  await db.clear('analytics-cache');
  await db.clear('pending-actions');
}

/**
 * Get database statistics
 */
export async function getDBStats() {
  const db = await getDB();

  const [offlineURLs, cachedURLs, analyticsCache, pendingActions] =
    await Promise.all([
      db.count('offline-urls'),
      db.count('cached-urls'),
      db.count('analytics-cache'),
      db.count('pending-actions'),
    ]);

  return {
    offlineURLs,
    cachedURLs,
    analyticsCache,
    pendingActions,
    total: offlineURLs + cachedURLs + analyticsCache + pendingActions,
  };
}
