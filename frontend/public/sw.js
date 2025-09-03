// Service Worker for URL Shortener PWA
const CACHE_NAME = 'urlshortener-v1.0.0';
const API_CACHE_NAME = 'urlshortener-api-v1.0.0';

// Static assets to cache
const STATIC_ASSETS = [
  '/',
  '/dashboard',
  '/analytics',
  '/login',
  '/register',
  '/manifest.json',
  '/icons/icon-192x192.png',
  '/icons/icon-512x512.png'
];

// API endpoints that should be cached
const API_ENDPOINTS = [
  '/api/v1/user/profile',
  '/api/v1/user/urls',
  '/api/v1/analytics'
];

// Install event - cache static assets
self.addEventListener('install', (event) => {
  console.log('Service Worker: Install event');
  event.waitUntil(
    caches.open(CACHE_NAME)
      .then((cache) => {
        console.log('Service Worker: Caching static assets');
        return cache.addAll(STATIC_ASSETS);
      })
      .catch((error) => {
        console.error('Service Worker: Cache installation failed:', error);
      })
  );
  // Force the new service worker to activate immediately
  self.skipWaiting();
});

// Activate event - cleanup old caches
self.addEventListener('activate', (event) => {
  console.log('Service Worker: Activate event');
  event.waitUntil(
    caches.keys().then((cacheNames) => {
      return Promise.all(
        cacheNames.map((cacheName) => {
          if (cacheName !== CACHE_NAME && cacheName !== API_CACHE_NAME) {
            console.log('Service Worker: Deleting old cache:', cacheName);
            return caches.delete(cacheName);
          }
        })
      );
    })
  );
  // Take control of all clients immediately
  self.clients.claim();
});

// Fetch event - implement caching strategy
self.addEventListener('fetch', (event) => {
  const { request } = event;
  const url = new URL(request.url);

  // Handle different types of requests
  if (request.method === 'GET') {
    if (url.pathname.startsWith('/api/')) {
      // API requests - Network First with API cache
      event.respondWith(networkFirstAPI(request));
    } else if (url.pathname.startsWith('/_next/static/') || 
               url.pathname.endsWith('.js') || 
               url.pathname.endsWith('.css')) {
      // Static assets - Cache First
      event.respondWith(cacheFirstStatic(request));
    } else {
      // HTML pages - Network First with fallback
      event.respondWith(networkFirstHTML(request));
    }
  }
});

// Network First strategy for API requests
async function networkFirstAPI(request) {
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok && request.method === 'GET') {
      const cache = await caches.open(API_CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('Service Worker: Network failed, trying cache:', error);
    const cachedResponse = await caches.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline response for API calls
    return new Response(
      JSON.stringify({ 
        error: 'Network unavailable', 
        offline: true,
        timestamp: Date.now()
      }),
      {
        status: 503,
        statusText: 'Service Unavailable',
        headers: { 'Content-Type': 'application/json' }
      }
    );
  }
}

// Cache First strategy for static assets
async function cacheFirstStatic(request) {
  const cachedResponse = await caches.match(request);
  
  if (cachedResponse) {
    return cachedResponse;
  }
  
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.error('Service Worker: Failed to fetch static asset:', error);
    throw error;
  }
}

// Network First strategy for HTML pages
async function networkFirstHTML(request) {
  try {
    const networkResponse = await fetch(request);
    
    if (networkResponse.ok) {
      const cache = await caches.open(CACHE_NAME);
      cache.put(request, networkResponse.clone());
    }
    
    return networkResponse;
  } catch (error) {
    console.log('Service Worker: Network failed, trying cache:', error);
    const cachedResponse = await caches.match(request);
    
    if (cachedResponse) {
      return cachedResponse;
    }
    
    // Return offline page
    return caches.match('/') || new Response(
      `<!DOCTYPE html>
      <html>
        <head>
          <title>URL Shortener - Offline</title>
          <meta name="viewport" content="width=device-width, initial-scale=1">
          <style>
            body { 
              font-family: system-ui, -apple-system, sans-serif; 
              display: flex; 
              justify-content: center; 
              align-items: center; 
              min-height: 100vh; 
              margin: 0; 
              background: #f8fafc;
              text-align: center;
            }
            .offline-container {
              max-width: 400px;
              padding: 2rem;
              background: white;
              border-radius: 8px;
              box-shadow: 0 4px 6px -1px rgba(0, 0, 0, 0.1);
            }
            .offline-icon {
              font-size: 4rem;
              margin-bottom: 1rem;
            }
            .offline-title {
              font-size: 1.5rem;
              font-weight: 600;
              margin-bottom: 0.5rem;
              color: #1f2937;
            }
            .offline-message {
              color: #6b7280;
              margin-bottom: 2rem;
            }
            .retry-button {
              background: #2563eb;
              color: white;
              border: none;
              padding: 0.75rem 1.5rem;
              border-radius: 0.375rem;
              cursor: pointer;
              font-weight: 500;
            }
          </style>
        </head>
        <body>
          <div class="offline-container">
            <div class="offline-icon">📡</div>
            <h1 class="offline-title">You're offline</h1>
            <p class="offline-message">
              Check your internet connection and try again.
            </p>
            <button class="retry-button" onclick="window.location.reload()">
              Retry
            </button>
          </div>
        </body>
      </html>`,
      {
        headers: { 'Content-Type': 'text/html' }
      }
    );
  }
}

// Background sync for failed URL creations
self.addEventListener('sync', (event) => {
  if (event.tag === 'background-sync-url-creation') {
    event.waitUntil(processOfflineURLCreations());
  }
});

// Process offline URL creations when back online
async function processOfflineURLCreations() {
  try {
    const db = await openDB();
    const tx = db.transaction(['offline_urls'], 'readonly');
    const store = tx.objectStore('offline_urls');
    const offlineURLs = await store.getAll();

    for (const urlData of offlineURLs) {
      try {
        const response = await fetch('/api/v1/urls', {
          method: 'POST',
          headers: {
            'Content-Type': 'application/json',
            'Authorization': `Bearer ${urlData.token}`
          },
          body: JSON.stringify(urlData.data)
        });

        if (response.ok) {
          // Remove from offline storage
          const deleteTx = db.transaction(['offline_urls'], 'readwrite');
          const deleteStore = deleteTx.objectStore('offline_urls');
          await deleteStore.delete(urlData.id);
        }
      } catch (error) {
        console.error('Failed to sync offline URL creation:', error);
      }
    }
  } catch (error) {
    console.error('Background sync failed:', error);
  }
}

// Push notification handling
self.addEventListener('push', (event) => {
  if (!event.data) return;

  const data = event.data.json();
  const options = {
    body: data.body,
    icon: '/icons/icon-192x192.png',
    badge: '/icons/badge-72x72.png',
    data: data.data,
    actions: [
      {
        action: 'view',
        title: 'View Details'
      },
      {
        action: 'dismiss',
        title: 'Dismiss'
      }
    ]
  };

  event.waitUntil(
    self.registration.showNotification(data.title, options)
  );
});

// Notification click handling
self.addEventListener('notificationclick', (event) => {
  event.notification.close();

  if (event.action === 'view') {
    const urlToOpen = event.notification.data?.url || '/dashboard';
    
    event.waitUntil(
      clients.matchAll().then((clientList) => {
        for (const client of clientList) {
          if (client.url === urlToOpen && 'focus' in client) {
            return client.focus();
          }
        }
        if (clients.openWindow) {
          return clients.openWindow(urlToOpen);
        }
      })
    );
  }
});

// Helper function to open IndexedDB
function openDB() {
  return new Promise((resolve, reject) => {
    const request = indexedDB.open('URLShortenerDB', 1);
    
    request.onerror = () => reject(request.error);
    request.onsuccess = () => resolve(request.result);
    
    request.onupgradeneeded = (event) => {
      const db = event.target.result;
      if (!db.objectStoreNames.contains('offline_urls')) {
        db.createObjectStore('offline_urls', { keyPath: 'id' });
      }
    };
  });
}