'use client';

import { useEffect } from 'react';
import { CONSENT_UPDATED_EVENT, getConsent, type ConsentState } from '@/lib/consent';

const GA_ID = 'G-KERP09FSLV';
const AHREFS_KEY = 'wh1b55CvCXjzUDz3SMYYWw';
const CONTENTSQUARE_SRC = 'https://t.contentsquare.net/uxa/0ac352c074f83.js';

// Everything a visitor can type is masked in session recordings (passwords,
// emails, card details, free text), plus anything tagged data-cs-mask.
const REPLAY_MASK_SELECTORS = [
  'input',
  'textarea',
  'select',
  '[contenteditable]',
  '[data-cs-mask]',
];

// Cookies set by the gated tools, removed when consent is withdrawn
const TRACKER_COOKIE = /^(_ga|_gid|_gat|_cs_)/;

type TrackerWindow = Window & {
  dataLayer?: unknown[];
  gtag?: (...args: unknown[]) => void;
  _uxa?: unknown[];
};

function injectScript(id: string, src: string, attrs: Record<string, string> = {}) {
  if (document.getElementById(id)) return;
  const script = document.createElement('script');
  script.id = id;
  script.src = src;
  script.async = true;
  Object.entries(attrs).forEach(([k, v]) => script.setAttribute(k, v));
  document.head.appendChild(script);
}

function loadAnalytics() {
  const w = window as TrackerWindow;
  w.dataLayer = w.dataLayer || [];
  w.gtag = function gtag() {
    // gtag.js expects the Arguments object itself
    // eslint-disable-next-line prefer-rest-params
    w.dataLayer!.push(arguments);
  };
  w.gtag('js', new Date());
  w.gtag('config', GA_ID);
  injectScript('ga4-gtag', `https://www.googletagmanager.com/gtag/js?id=${GA_ID}`);
  injectScript('ahrefs-analytics', 'https://analytics.ahrefs.com/analytics.js', { 'data-key': AHREFS_KEY });
}

function loadReplay() {
  const w = window as TrackerWindow;
  w._uxa = w._uxa || [];
  // Must be queued before the Contentsquare tag executes
  w._uxa.push(['setPIISelectors', { PIISelectors: REPLAY_MASK_SELECTORS }]);
  injectScript('contentsquare', CONTENTSQUARE_SRC);
}

function clearTrackerCookies() {
  const host = window.location.hostname;
  const domains = ['', host, `.${host}`, `.${host.split('.').slice(-2).join('.')}`];
  document.cookie.split(';').forEach((cookie) => {
    const name = cookie.split('=')[0].trim();
    if (!TRACKER_COOKIE.test(name)) return;
    domains.forEach((domain) => {
      document.cookie = `${name}=; expires=Thu, 01 Jan 1970 00:00:00 GMT; path=/${domain ? `; domain=${domain}` : ''}`;
    });
  });
}

// Third-party tags stay off the critical path, as with next/script lazyOnload
function whenIdle(fn: () => void) {
  const run = () => ('requestIdleCallback' in window ? window.requestIdleCallback(fn) : setTimeout(fn, 1));
  if (document.readyState === 'complete') run();
  else window.addEventListener('load', run, { once: true });
}

function apply(consent: ConsentState | null) {
  if (consent?.analytics) whenIdle(loadAnalytics);
  if (consent?.replay) whenIdle(loadReplay);
}

/**
 * Loads analytics and session replay only after the visitor has opted in
 * through the cookie banner (CIPA / GDPR). Nothing loads by default.
 */
export function ConsentScripts() {
  useEffect(() => {
    let current = getConsent();
    apply(current);

    const onUpdate = (e: Event) => {
      const next = (e as CustomEvent<ConsentState>).detail;
      const withdrawn =
        (current?.analytics && !next.analytics) || (current?.replay && !next.replay);
      current = next;
      if (withdrawn) {
        // A running tag can't be unloaded: drop its cookies and reload without it
        clearTrackerCookies();
        window.location.reload();
        return;
      }
      apply(next);
    };

    window.addEventListener(CONSENT_UPDATED_EVENT, onUpdate);
    return () => window.removeEventListener(CONSENT_UPDATED_EVENT, onUpdate);
  }, []);

  return null;
}
