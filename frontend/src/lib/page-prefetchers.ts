import { prefetch } from './prefetch-cache';

function authedFetch(url: string, token: string) {
  return fetch(url, { headers: { Authorization: `Bearer ${token}` } })
    .then(r => r.ok ? r.json() : Promise.reject(r.status));
}

export function prefetchDashboard(token: string, apiBase: string) {
  prefetch('dashboard:urls',  () => authedFetch(`${apiBase}/api/v1/my-urls`, token));
  prefetch('dashboard:stats', () => authedFetch(`${apiBase}/api/v1/dashboard/stats`, token));
}

export function prefetchAnalytics(token: string, apiBase: string) {
  prefetch('analytics:urls', () => authedFetch(`${apiBase}/api/v1/my-urls`, token));
}

export function prefetchBilling(token: string, apiBase: string) {
  prefetch('billing:sub',      () => authedFetch(`${apiBase}/api/v1/billing/usage`, token));
  prefetch('billing:products', () => authedFetch(`${apiBase}/api/v1/billing/plans`, token));
}

export function prefetchSettings(token: string, apiBase: string) {
  prefetch('settings:profile', () => authedFetch(`${apiBase}/api/v1/auth/profile`, token));
}

export function prefetchApiKeys(token: string, apiBase: string) {
  prefetch('apikeys:list', () => authedFetch(`${apiBase}/api/v1/api-keys`, token));
}
