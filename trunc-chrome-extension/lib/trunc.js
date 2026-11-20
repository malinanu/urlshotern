// Trunc API client and "Sign in with Trunc" for the extension.
// Shared by the popup, the options page and the background service worker.

export const DEFAULT_BASE_URL = 'https://trunc.site';
export const CLIENT_ID = 'trunc-chrome';

export class NotConnectedError extends Error {
  constructor() {
    super('Connect your Trunc account first.');
    this.name = 'NotConnectedError';
  }
}

/** Removes a trailing slash and anything after the origin. */
export function normalizeBaseUrl(value) {
  try {
    const u = new URL(String(value || '').trim());
    if (u.protocol !== 'https:' && !(u.protocol === 'http:' && ['localhost', '127.0.0.1'].includes(u.hostname))) return null;
    return u.origin;
  } catch {
    return null;
  }
}

export async function getSettings() {
  const s = await chrome.storage.local.get(['apiKey', 'baseUrl', 'connectedVia']);
  return {
    apiKey: s.apiKey || '',
    baseUrl: normalizeBaseUrl(s.baseUrl) || DEFAULT_BASE_URL,
    connectedVia: s.connectedVia || '',
  };
}

export async function saveConnection(apiKey, connectedVia) {
  await chrome.storage.local.set({ apiKey, connectedVia });
}

export async function clearConnection() {
  await chrome.storage.local.remove(['apiKey', 'connectedVia']);
  await chrome.storage.session.clear();
}

/** Calls the Trunc API with the saved key. Throws Error(message) on failure. */
export async function api(path, { method = 'GET', body } = {}) {
  const { apiKey, baseUrl } = await getSettings();
  if (!apiKey) throw new NotConnectedError();
  let res;
  try {
    res = await fetch(baseUrl + path, {
      method,
      headers: { Authorization: `Bearer ${apiKey}`, 'Content-Type': 'application/json' },
      body: body ? JSON.stringify(body) : undefined,
    });
  } catch {
    throw new Error(`Can't reach ${new URL(baseUrl).host}. Check your connection.`);
  }
  const data = await res.json().catch(() => ({}));
  if (res.status === 401) throw new Error('Your key is no longer valid. Connect again in the extension settings.');
  if (!res.ok) throw new Error(data.message || data.error_description || data.error || `Request failed (${res.status}).`);
  return data;
}

export function isShortenable(url) {
  return /^https?:\/\//i.test(url || '');
}

/**
 * Shortens url, reusing the result from earlier in this browser session so
 * reopening the popup on the same page doesn't create a second link.
 */
export async function shorten(url) {
  const cacheKey = `short:${url}`;
  const cached = (await chrome.storage.session.get(cacheKey))[cacheKey];
  if (cached) return cached;
  const data = await api('/api/v1/shorten', { method: 'POST', body: { url } });
  const result = { shortUrl: data.short_url, shortCode: data.short_code, originalUrl: data.original_url || url };
  await chrome.storage.session.set({ [cacheKey]: result });
  return result;
}

export async function recentLinks(limit = 5) {
  const data = await api(`/api/v1/my-urls?page=1&limit=${limit}`);
  return (data.urls || []).map((u) => ({
    shortUrl: u.short_url,
    shortCode: u.short_code,
    originalUrl: u.original_url,
    clicks: u.click_count || 0,
  }));
}

// --- Sign in with Trunc (OAuth 2.0 authorization code + PKCE) ---

function base64url(bytes) {
  let s = '';
  bytes.forEach((b) => (s += String.fromCharCode(b)));
  return btoa(s).replace(/\+/g, '-').replace(/\//g, '_').replace(/=+$/, '');
}

export function randomString(byteLength = 32) {
  const bytes = new Uint8Array(byteLength);
  crypto.getRandomValues(bytes);
  return base64url(bytes);
}

export async function pkceChallenge(verifier) {
  const digest = await crypto.subtle.digest('SHA-256', new TextEncoder().encode(verifier));
  return base64url(new Uint8Array(digest));
}

/** Builds the /connect URL. Exposed for tests. */
export function connectUrl(baseUrl, redirectUri, state, challenge) {
  const q = new URLSearchParams({
    response_type: 'code',
    client_id: CLIENT_ID,
    redirect_uri: redirectUri,
    state,
    scope: 'links',
    code_challenge: challenge,
    code_challenge_method: 'S256',
  });
  return `${baseUrl}/connect?${q}`;
}

/** Exchanges a code for this extension's API key. Exposed for tests. */
export async function exchangeCode(baseUrl, code, redirectUri, verifier) {
  const res = await fetch(`${baseUrl}/api/v1/oauth/token`, {
    method: 'POST',
    headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
    body: new URLSearchParams({
      grant_type: 'authorization_code',
      code,
      redirect_uri: redirectUri,
      client_id: CLIENT_ID,
      code_verifier: verifier,
    }),
  });
  const data = await res.json().catch(() => ({}));
  if (!res.ok || !data.access_token) throw new Error(data.error_description || 'Could not finish signing in.');
  return data.access_token;
}

/**
 * Runs the whole sign-in in a Chrome-managed window. Must run in the
 * background service worker: the popup closes when that window opens.
 */
export async function signInWithTrunc() {
  const { baseUrl } = await getSettings();
  const redirectUri = chrome.identity.getRedirectURL('trunc');
  const verifier = randomString(48);
  const state = randomString(16);
  const challenge = await pkceChallenge(verifier);
  const result = await chrome.identity.launchWebAuthFlow({
    url: connectUrl(baseUrl, redirectUri, state, challenge),
    interactive: true,
  });
  const back = new URL(result);
  if (back.searchParams.get('state') !== state) throw new Error('Sign-in was interrupted. Please try again.');
  const err = back.searchParams.get('error');
  if (err) throw new Error(err === 'access_denied' ? 'You cancelled signing in.' : back.searchParams.get('error_description') || err);
  const token = await exchangeCode(baseUrl, back.searchParams.get('code') || '', redirectUri, verifier);
  await saveConnection(token, 'signin');
  return true;
}

/** Revokes a key that came from Sign in with Trunc, then forgets it. */
export async function disconnect() {
  const { apiKey, baseUrl, connectedVia } = await getSettings();
  if (apiKey && connectedVia === 'signin') {
    await fetch(`${baseUrl}/api/v1/oauth/revoke`, {
      method: 'POST',
      headers: { 'Content-Type': 'application/x-www-form-urlencoded' },
      body: new URLSearchParams({ token: apiKey, client_id: CLIENT_ID }),
    }).catch(() => {});
  }
  await clearConnection();
}

/** Checks that a pasted key works before saving it. */
export async function savePastedKey(key) {
  const k = String(key || '').trim();
  if (!/^us_[A-Za-z0-9]{16,}$/.test(k)) throw new Error('That does not look like a Trunc API key (it starts with us_).');
  const { baseUrl } = await getSettings();
  const res = await fetch(`${baseUrl}/api/v1/my-urls?page=1&limit=1`, { headers: { Authorization: `Bearer ${k}` } }).catch(() => null);
  if (!res) throw new Error(`Can't reach ${new URL(baseUrl).host}.`);
  if (!res.ok) throw new Error('Trunc did not accept that key.');
  await saveConnection(k, 'apikey');
}
