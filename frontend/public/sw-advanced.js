/**
 * Advanced Service Worker with Workbox
 * Optimized for mobile performance and offline functionality
 */

importScripts('https://storage.googleapis.com/workbox-cdn/releases/7.0.0/workbox-sw.js');

const { registerRoute } = workbox.routing;
const { CacheFirst, NetworkFirst, StaleWhileRevalidate, NetworkOnly } = workbox.strategies;
const { CacheableResponsePlugin } = workbox.cacheableResponse;
const { ExpirationPlugin } = workbox.expiration;
const { BackgroundSyncPlugin } = workbox.backgroundSync;
const { precacheAndRoute, cleanupOutdatedCaches } = workbox.precaching;

// Service worker version
const SW_VERSION = '2.0.1';
const CACHE_PREFIX = 'urlshortener';

// Configure workbox
workbox.core.setCacheNameDetails({
  prefix: CACHE_PREFIX,
  suffix: SW_VERSION,
  precache: 'precache',
  runtime: 'runtime',
});

// Clean up outdated caches
cleanupOutdatedCaches();

// Precache static assets (will be populated during build)
precacheAndRoute(self.__WB_MANIFEST || []);

// Filter out non-HTTP(S) requests (chrome-extension, data, blob, etc.)
self.addEventListener('fetch', (event) => {
  const url = new URL(event.request.url);
  // Only handle http and https requests
  if (!url.protocol.startsWith('http')) {
    return;
  }
});

/**
 * Cache Strategy 1: Static Assets (JS, CSS, Fonts)
 * Strategy: Cache First with long expiration
 */
registerRoute(
  ({ request }) => {
    return (
      request.destination === 'script' ||
      request.destination === 'style' ||
      request.destination === 'font'
    );
  },
  new CacheFirst({
    cacheName: `${CACHE_PREFIX}-static-${SW_VERSION}`,
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200],
      }),
      new ExpirationPlugin({
        maxEntries: 60,
        maxAgeSeconds: 30 * 24 * 60 * 60, // 30 days
        purgeOnQuotaError: true,
      }),
    ],
  })
);

/**
 * Cache Strategy 2: Images
 * Strategy: Cache First with size limits
 */
registerRoute(
  ({ request }) => request.destination === 'image',
  new CacheFirst({
    cacheName: `${CACHE_PREFIX}-images-${SW_VERSION}`,
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200],
      }),
      new ExpirationPlugin({
        maxEntries: 50,
        maxAgeSeconds: 7 * 24 * 60 * 60, // 7 days
        purgeOnQuotaError: true,
      }),
    ],
  })
);

/**
 * Cache Strategy 3: API GET Requests (Read-only)
 * Strategy: Network First with short cache fallback
 */
registerRoute(
  ({ url, request }) => {
    return (
      url.pathname.startsWith('/api/') &&
      request.method === 'GET' &&
      !url.pathname.includes('/realtime')
    );
  },
  new NetworkFirst({
    cacheName: `${CACHE_PREFIX}-api-${SW_VERSION}`,
    networkTimeoutSeconds: 5,
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200],
      }),
      new ExpirationPlugin({
        maxEntries: 50,
        maxAgeSeconds: 5 * 60, // 5 minutes
        purgeOnQuotaError: true,
      }),
    ],
  })
);

/**
 * Cache Strategy 4: Analytics Data
 * Strategy: Stale While Revalidate for better performance
 */
registerRoute(
  ({ url }) => url.pathname.includes('/analytics'),
  new StaleWhileRevalidate({
    cacheName: `${CACHE_PREFIX}-analytics-${SW_VERSION}`,
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200],
      }),
      new ExpirationPlugin({
        maxEntries: 30,
        maxAgeSeconds: 10 * 60, // 10 minutes
        purgeOnQuotaError: true,
      }),
    ],
  })
);

/**
 * Cache Strategy 5: HTML Pages
 * Strategy: Network First with cache fallback
 */
registerRoute(
  ({ request }) => request.mode === 'navigate',
  new NetworkFirst({
    cacheName: `${CACHE_PREFIX}-pages-${SW_VERSION}`,
    networkTimeoutSeconds: 3,
    plugins: [
      new CacheableResponsePlugin({
        statuses: [0, 200],
      }),
      new ExpirationPlugin({
        maxEntries: 20,
        maxAgeSeconds: 24 * 60 * 60, // 1 day
        purgeOnQuotaError: true,
      }),
    ],
  })
);

/**
 * Background Sync for POST/PUT/DELETE requests
 * Queues failed mutations for retry when online
 */
const bgSyncPlugin = new BackgroundSyncPlugin('url-mutations-queue', {
  maxRetentionTime: 24 * 60, // Retry for max of 24 hours
  onSync: async ({ queue }) => {
    let entry;
    while ((entry = await queue.shiftRequest())) {
      try {
        await fetch(entry.request);
        console.log('Background sync successful:', entry.request.url);
      } catch (error) {
        console.error('Background sync failed:', error);
        await queue.unshiftRequest(entry);
        throw error;
      }
    }
  },
});

/**
 * Handle API mutations (POST, PUT, DELETE)
 * Use Network Only with Background Sync
 */
registerRoute(
  ({ url, request }) => {
    return (
      url.pathname.startsWith('/api/') &&
      (request.method === 'POST' ||
        request.method === 'PUT' ||
        request.method === 'DELETE')
    );
  },
  new NetworkOnly({
    plugins: [bgSyncPlugin],
  }),
  'POST'
);

registerRoute(
  ({ url, request }) => {
    return url.pathname.startsWith('/api/') && request.method === 'PUT';
  },
  new NetworkOnly({
    plugins: [bgSyncPlugin],
  }),
  'PUT'
);

registerRoute(
  ({ url, request }) => {
    return url.pathname.startsWith('/api/') && request.method === 'DELETE';
  },
  new NetworkOnly({
    plugins: [bgSyncPlugin],
  }),
  'DELETE'
);

/**
 * Install Event
 */
self.addEventListener('install', (event) => {
  console.log(`[SW ${SW_VERSION}] Installing...`);
  self.skipWaiting();
});

/**
 * Activate Event
 */
self.addEventListener('activate', (event) => {
  console.log(`[SW ${SW_VERSION}] Activating...`);
  event.waitUntil(
    (async () => {
      // Clean up old caches
      const cacheNames = await caches.keys();
      await Promise.all(
        cacheNames.map(async (cacheName) => {
          if (cacheName.startsWith(CACHE_PREFIX) && !cacheName.includes(SW_VERSION)) {
            console.log(`[SW ${SW_VERSION}] Deleting old cache:`, cacheName);
            return caches.delete(cacheName);
          }
        })
      );

      // Take control of all clients
      await self.clients.claim();
    })()
  );
});

/**
 * Message Event - Handle commands from client
 */
self.addEventListener('message', (event) => {
  if (event.data && event.data.type === 'SKIP_WAITING') {
    self.skipWaiting();
  }

  if (event.data && event.data.type === 'CLEAR_CACHE') {
    event.waitUntil(
      caches.keys().then((cacheNames) => {
        return Promise.all(
          cacheNames.map((cacheName) => {
            if (cacheName.startsWith(CACHE_PREFIX)) {
              return caches.delete(cacheName);
            }
          })
        );
      })
    );
  }

  if (event.data && event.data.type === 'GET_VERSION') {
    event.ports[0].postMessage({ version: SW_VERSION });
  }
});

/**
 * Push Notification Handler
 */
self.addEventListener('push', (event) => {
  if (!event.data) return;

  const data = event.data.json();
  const title = data.title || 'URL Shortener';
  const options = {
    body: data.body || 'You have a new notification',
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    vibrate: [200, 100, 200],
    data: {
      url: data.url || '/',
      ...data.data,
    },
    actions: [
      {
        action: 'open',
        title: 'Open',
      },
      {
        action: 'close',
        title: 'Close',
      },
    ],
    requireInteraction: false,
  };

  event.waitUntil(self.registration.showNotification(title, options));
});

/**
 * Notification Click Handler
 */
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  if (event.action === 'close') {
    return;
  }

  const urlToOpen = event.notification.data?.url || '/';

  event.waitUntil(
    clients
      .matchAll({
        type: 'window',
        includeUncontrolled: true,
      })
      .then((clientList) => {
        // Check if there's already a window open
        for (const client of clientList) {
          if (client.url === urlToOpen && 'focus' in client) {
            return client.focus();
          }
        }

        // Open a new window
        if (clients.openWindow) {
          return clients.openWindow(urlToOpen);
        }
      })
  );
});

/**
 * Periodic Background Sync (if supported)
 */
self.addEventListener('periodicsync', (event) => {
  if (event.tag === 'update-analytics') {
    event.waitUntil(updateAnalyticsCache());
  }
});

/**
 * Update analytics cache in background
 */
async function updateAnalyticsCache() {
  try {
    const cache = await caches.open(`${CACHE_PREFIX}-analytics-${SW_VERSION}`);
    const requests = await cache.keys();

    for (const request of requests) {
      try {
        const response = await fetch(request);
        if (response.ok) {
          await cache.put(request, response);
        }
      } catch (error) {
        console.error('Failed to update analytics cache:', error);
      }
    }
  } catch (error) {
    console.error('Periodic sync failed:', error);
  }
}

/**
 * Network status monitoring
 */
self.addEventListener('online', () => {
  console.log('[SW] Back online, syncing data...');
  self.clients.matchAll().then((clients) => {
    clients.forEach((client) => {
      client.postMessage({
        type: 'ONLINE',
        message: 'Back online',
      });
    });
  });
});

self.addEventListener('offline', () => {
  console.log('[SW] Offline mode activated');
  self.clients.matchAll().then((clients) => {
    clients.forEach((client) => {
      client.postMessage({
        type: 'OFFLINE',
        message: 'Offline mode',
      });
    });
  });
});

console.log(`[SW ${SW_VERSION}] Service Worker loaded successfully`);
