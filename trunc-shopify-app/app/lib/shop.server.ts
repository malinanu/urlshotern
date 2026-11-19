// A shop's connection to Trunc, kept in ShopSettings.
import type { PrismaClient } from "@prisma/client";
import { connectUrl, createPkce, exchangeCode, getUsage, revokeKey, TruncError } from "./trunc.server";

const CONNECT_TTL_MS = 10 * 60 * 1000;

export async function getConnection(db: PrismaClient, shop: string) {
  const s = await db.shopSettings.findUnique({ where: { shop } });
  return { apiKey: s?.apiKey || "", connectedVia: s?.connectedVia || "" };
}

/** Starts Sign in with Trunc and returns the URL to send the merchant to. */
export async function startConnect(db: PrismaClient, shop: string): Promise<string> {
  const { state, verifier, challenge } = createPkce();
  const pkce = { pkceState: state, pkceVerifier: verifier, pkceExpiresAt: new Date(Date.now() + CONNECT_TTL_MS) };
  await db.shopSettings.upsert({ where: { shop }, create: { shop, ...pkce }, update: pkce });
  return connectUrl(state, challenge);
}

export type FinishResult = { shop: string; ok: true } | { shop: string | null; ok: false; reason: "expired" | "cancelled" | "failed" };

/**
 * Handles the return from Trunc. The state is single-use and identifies the
 * shop that started; the code only works with that shop's PKCE verifier.
 */
export async function finishConnect(
  db: PrismaClient,
  params: { state?: string | null; code?: string | null; error?: string | null },
): Promise<FinishResult> {
  if (!params.state) return { shop: null, ok: false, reason: "expired" };
  const s = await db.shopSettings.findUnique({ where: { pkceState: params.state } });
  if (!s) return { shop: null, ok: false, reason: "expired" };
  await db.shopSettings.update({
    where: { shop: s.shop },
    data: { pkceState: null, pkceVerifier: null, pkceExpiresAt: null },
  });
  if (!s.pkceVerifier || !s.pkceExpiresAt || s.pkceExpiresAt.getTime() < Date.now()) {
    return { shop: s.shop, ok: false, reason: "expired" };
  }
  if (params.error) return { shop: s.shop, ok: false, reason: params.error === "access_denied" ? "cancelled" : "failed" };
  if (!params.code) return { shop: s.shop, ok: false, reason: "failed" };
  try {
    const key = await exchangeCode(params.code, s.pkceVerifier);
    await db.shopSettings.update({ where: { shop: s.shop }, data: { apiKey: key, connectedVia: "signin" } });
    return { shop: s.shop, ok: true };
  } catch {
    return { shop: s.shop, ok: false, reason: "failed" };
  }
}

/** Saves a pasted API key after checking that Trunc accepts it. */
export async function savePastedKey(db: PrismaClient, shop: string, rawKey: string): Promise<void> {
  const key = rawKey.trim();
  if (!/^us_[A-Za-z0-9]{16,}$/.test(key)) {
    throw new TruncError("That does not look like a Trunc API key (it starts with us_).", 400);
  }
  await getUsage(key);
  await db.shopSettings.upsert({
    where: { shop },
    create: { shop, apiKey: key, connectedVia: "apikey" },
    update: { apiKey: key, connectedVia: "apikey" },
  });
}

/** Forgets the key; a key from Sign in with Trunc is also revoked on Trunc. */
export async function disconnect(db: PrismaClient, shop: string): Promise<void> {
  const { apiKey, connectedVia } = await getConnection(db, shop);
  if (apiKey && connectedVia === "signin") await revokeKey(apiKey);
  await db.shopSettings.updateMany({ where: { shop }, data: { apiKey: "", connectedVia: "" } });
}

/** shop/redact: remove everything stored for the shop. */
export async function deleteShopData(db: PrismaClient, shop: string): Promise<void> {
  await disconnect(db, shop).catch(() => undefined);
  await db.link.deleteMany({ where: { shop } });
  await db.shopSettings.deleteMany({ where: { shop } });
  await db.session.deleteMany({ where: { shop } });
}
