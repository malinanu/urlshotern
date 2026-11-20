// Trunc API client and "Sign in with Trunc" (OAuth 2.0 code + PKCE) for the
// Shopify app. Server-only: API keys never reach the browser.
import { createHash, randomBytes } from "node:crypto";

export const CLIENT_ID = "trunc-shopify";

/** The Trunc server: trunc.site, or TRUNC_BASE_URL for a self-hosted one. */
export function truncBaseUrl(): string {
  return (process.env.TRUNC_BASE_URL || "https://trunc.site").replace(/\/+$/, "");
}

/** Where Trunc sends the merchant back to. Must match on both steps. */
export function connectRedirectUri(): string {
  return `${(process.env.SHOPIFY_APP_URL || "").replace(/\/+$/, "")}/auth/trunc/callback`;
}

export class TruncError extends Error {
  constructor(
    message: string,
    public status: number,
  ) {
    super(message);
    this.name = "TruncError";
  }
}

async function call<T>(path: string, apiKey: string, init: RequestInit = {}): Promise<T> {
  let res: Response;
  try {
    res = await fetch(`${truncBaseUrl()}${path}`, {
      ...init,
      headers: {
        Authorization: `Bearer ${apiKey}`,
        "Content-Type": "application/json",
        ...(init.headers || {}),
      },
    });
  } catch {
    throw new TruncError(`Can't reach ${new URL(truncBaseUrl()).host}.`, 0);
  }
  const data = (await res.json().catch(() => ({}))) as Record<string, unknown>;
  if (res.status === 401) {
    throw new TruncError("Trunc did not accept the key. Connect again in Settings.", 401);
  }
  if (!res.ok) {
    const msg = [data.message, data.error_description, data.error].find((v) => typeof v === "string" && v) as string | undefined;
    throw new TruncError(msg || `Trunc returned ${res.status}.`, res.status);
  }
  return data as T;
}

export interface ShortLink {
  short_code: string;
  short_url: string;
}

export function shortenUrl(apiKey: string, url: string, customCode?: string): Promise<ShortLink> {
  return call<ShortLink>("/api/v1/shorten", apiKey, {
    method: "POST",
    body: JSON.stringify(customCode ? { url, custom_code: customCode } : { url }),
  });
}

export interface UsageResult {
  planType: string;
  planName: string;
  currentUrls: number;
  maxUrls: number; // -1 = unlimited
  monthlyClicks: number;
  maxClicks: number; // -1 = unlimited
  urlLimitReached: boolean;
}

const PLAN_NAMES: Record<string, string> = {
  free: "Free",
  premium: "Premium",
  growth: "Growth",
  enterprise: "Enterprise",
};

export function planName(planType: string): string {
  return PLAN_NAMES[planType.replace(/_yearly$/, "")] || planType;
}

/** Plan and usage for the key's owner. Also confirms the key works. */
export async function getUsage(apiKey: string): Promise<UsageResult> {
  type Raw = {
    subscription?: { plan_type?: string };
    usage?: { current_urls?: number; monthly_clicks?: number };
    limits?: { max_urls?: number; max_clicks_month?: number };
    warnings?: { url_limit_reached?: boolean };
  };
  const d = await call<Raw>("/api/v1/billing/usage", apiKey);
  const planType = d.subscription?.plan_type || "free";
  return {
    planType,
    planName: planName(planType),
    currentUrls: d.usage?.current_urls ?? 0,
    maxUrls: d.limits?.max_urls ?? 0,
    monthlyClicks: d.usage?.monthly_clicks ?? 0,
    maxClicks: d.limits?.max_clicks_month ?? 0,
    urlLimitReached: Boolean(d.warnings?.url_limit_reached),
  };
}

export interface LinkAnalytics {
  totalClicks: number;
  lastClickAt?: string;
  daily: { date: string; clicks: number }[];
  countries: { country_code: string; clicks: number }[];
}

export async function getAnalytics(apiKey: string, shortCode: string): Promise<LinkAnalytics> {
  type Raw = {
    total_clicks?: number;
    last_click_at?: string;
    daily_clicks?: { date: string; clicks: number }[];
    country_stats?: { country_code: string; clicks: number }[];
  };
  const d = await call<Raw>(`/api/v1/analytics/${encodeURIComponent(shortCode)}`, apiKey);
  return {
    totalClicks: d.total_clicks ?? 0,
    lastClickAt: d.last_click_at,
    daily: d.daily_clicks ?? [],
    countries: d.country_stats ?? [],
  };
}

// --- Sign in with Trunc ---

const b64url = (buf: Buffer) => buf.toString("base64url");

export function createPkce(): { state: string; verifier: string; challenge: string } {
  const verifier = b64url(randomBytes(48));
  return {
    state: b64url(randomBytes(24)),
    verifier,
    challenge: b64url(createHash("sha256").update(verifier).digest()),
  };
}

export function connectUrl(state: string, challenge: string): string {
  const q = new URLSearchParams({
    response_type: "code",
    client_id: CLIENT_ID,
    redirect_uri: connectRedirectUri(),
    state,
    scope: "links",
    code_challenge: challenge,
    code_challenge_method: "S256",
  });
  return `${truncBaseUrl()}/connect?${q}`;
}

export async function exchangeCode(code: string, verifier: string): Promise<string> {
  const res = await fetch(`${truncBaseUrl()}/api/v1/oauth/token`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({
      grant_type: "authorization_code",
      code,
      redirect_uri: connectRedirectUri(),
      client_id: CLIENT_ID,
      code_verifier: verifier,
    }),
  });
  const data = (await res.json().catch(() => ({}))) as { access_token?: string; error_description?: string };
  if (!res.ok || !data.access_token) {
    throw new TruncError(data.error_description || "Trunc could not finish connecting.", res.status);
  }
  return data.access_token;
}

/** Revokes a key that came from Sign in with Trunc. Never throws. */
export async function revokeKey(apiKey: string): Promise<void> {
  await fetch(`${truncBaseUrl()}/api/v1/oauth/revoke`, {
    method: "POST",
    headers: { "Content-Type": "application/x-www-form-urlencoded" },
    body: new URLSearchParams({ token: apiKey, client_id: CLIENT_ID }),
  }).catch(() => undefined);
}
