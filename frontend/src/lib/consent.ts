/**
 * Cookie / tracking consent shared by the consent banner and the script loader.
 *
 * Every optional category defaults to "not granted": analytics, session replay
 * and marketing tags only load after the visitor opts in.
 */

export interface ConsentState {
  version: 2;
  essential: true;
  analytics: boolean;
  /** Session analytics & replay (Contentsquare): records page interactions */
  replay: boolean;
  marketing: boolean;
  savedAt: string;
}

export type ConsentChoice = Pick<ConsentState, 'analytics' | 'replay' | 'marketing'>;

// v2 added the session-replay category, so v1 choices are not carried over
const STORAGE_KEY = 'cookie_consent_v2';
const LEGACY_STORAGE_KEY = 'cookie_consent';

export const CONSENT_UPDATED_EVENT = 'cookie-consent-updated';
export const CONSENT_OPEN_EVENT = 'cookie-consent-open';

export function getConsent(): ConsentState | null {
  if (typeof window === 'undefined') return null;
  try {
    const raw = localStorage.getItem(STORAGE_KEY);
    if (!raw) return null;
    const parsed = JSON.parse(raw) as Partial<ConsentState>;
    if (parsed.version !== 2) return null;
    return {
      version: 2,
      essential: true,
      analytics: parsed.analytics === true,
      replay: parsed.replay === true,
      marketing: parsed.marketing === true,
      savedAt: parsed.savedAt ?? '',
    };
  } catch {
    return null;
  }
}

export function saveConsent(choice: ConsentChoice): ConsentState {
  const state: ConsentState = {
    version: 2,
    essential: true,
    ...choice,
    savedAt: new Date().toISOString(),
  };
  try {
    localStorage.setItem(STORAGE_KEY, JSON.stringify(state));
    localStorage.removeItem(LEGACY_STORAGE_KEY);
  } catch {
    // storage unavailable: the choice still applies for this page view
  }
  window.dispatchEvent(new CustomEvent<ConsentState>(CONSENT_UPDATED_EVENT, { detail: state }));
  return state;
}

/** Re-open the consent banner so the visitor can change or withdraw consent. */
export function openConsentSettings(): void {
  window.dispatchEvent(new Event(CONSENT_OPEN_EVENT));
}
