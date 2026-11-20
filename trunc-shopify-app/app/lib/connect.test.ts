// Integration tests: the app's Trunc connection against a real Trunc server
// (TRUNC_BASE_URL) and the app's own SQLite database.
//
//   TRUNC_BASE_URL=http://localhost:8099 TRUNC_TEST_EMAIL=… TRUNC_TEST_PASSWORD=… npm test
//
// Skipped when TRUNC_TEST_EMAIL is not set.
import { beforeAll, beforeEach, describe, expect, it } from "vitest";

const EMAIL = process.env.TRUNC_TEST_EMAIL;
const PASSWORD = process.env.TRUNC_TEST_PASSWORD;
process.env.SHOPIFY_APP_URL = "https://trunc-app.example.com";
process.env.SHOPIFY_API_KEY = "test-client-id";

const SHOP = "trunc-test.myshopify.com";

describe.skipIf(!EMAIL)("Trunc connection", () => {
  let db: typeof import("../db.server").default;
  let shop: typeof import("./shop.server");
  let trunc: typeof import("./trunc.server");
  let jwt = "";
  const base = () => trunc.truncBaseUrl();

  /** Plays the merchant approving the app on trunc.site/connect. */
  async function approve(connectUrl: string, decision: "allow" | "deny" = "allow") {
    const u = new URL(connectUrl);
    const body = Object.fromEntries(u.searchParams.entries());
    const res = await fetch(`${base()}/api/v1/oauth/authorize`, {
      method: "POST",
      headers: { Authorization: `Bearer ${jwt}`, "Content-Type": "application/json" },
      body: JSON.stringify({ ...body, decision }),
    });
    const data = (await res.json()) as { redirect: string };
    return new URL(data.redirect);
  }

  async function newApiKey() {
    const res = await fetch(`${base()}/api/v1/api-keys`, {
      method: "POST",
      headers: { Authorization: `Bearer ${jwt}`, "Content-Type": "application/json" },
      body: JSON.stringify({ name: "Shopify test", permissions: ["urls:create", "urls:read", "analytics:read"] }),
    });
    return ((await res.json()) as { plain_key: string }).plain_key;
  }

  beforeAll(async () => {
    db = (await import("../db.server")).default;
    shop = await import("./shop.server");
    trunc = await import("./trunc.server");
    const login = await fetch(`${base()}/api/v1/auth/login`, {
      method: "POST",
      headers: { "Content-Type": "application/json" },
      body: JSON.stringify({ email: EMAIL, password: PASSWORD }),
    });
    jwt = ((await login.json()) as { access_token: string }).access_token;
    expect(jwt).toBeTruthy();
  });

  beforeEach(async () => {
    await db.link.deleteMany({});
    await db.shopSettings.deleteMany({});
  });

  it("builds a PKCE sign-in URL and remembers the state", async () => {
    const url = new URL(await shop.startConnect(db, SHOP));
    expect(url.origin + url.pathname).toBe(`${base()}/connect`);
    expect(url.searchParams.get("client_id")).toBe("trunc-shopify");
    expect(url.searchParams.get("redirect_uri")).toBe("https://trunc-app.example.com/auth/trunc/callback");
    expect(url.searchParams.get("code_challenge_method")).toBe("S256");
    const row = await db.shopSettings.findUnique({ where: { shop: SHOP } });
    expect(row?.pkceState).toBe(url.searchParams.get("state"));
    expect(row?.pkceVerifier).toBeTruthy();
  });

  it("signs in, shortens, reads clicks, and revokes on disconnect", async () => {
    const back = await approve(await shop.startConnect(db, SHOP));
    expect(back.origin + back.pathname).toBe("https://trunc-app.example.com/auth/trunc/callback");

    const result = await shop.finishConnect(db, { state: back.searchParams.get("state"), code: back.searchParams.get("code") });
    expect(result).toEqual({ shop: SHOP, ok: true });
    const { apiKey, connectedVia } = await shop.getConnection(db, SHOP);
    expect(apiKey).toMatch(/^us_/);
    expect(connectedVia).toBe("signin");
    expect((await trunc.getUsage(apiKey)).planName).toBe("Enterprise");

    const link = await trunc.shortenUrl(apiKey, `http://127.0.0.1:8765/menu.html?shopify=${Date.now()}`);
    expect(link.short_url).toContain(link.short_code);
    const stats = await trunc.getAnalytics(apiKey, link.short_code);
    expect(stats.totalClicks).toBe(0);

    // The same state can't be used twice.
    const again = await shop.finishConnect(db, { state: back.searchParams.get("state"), code: back.searchParams.get("code") });
    expect(again.ok).toBe(false);

    await shop.disconnect(db, SHOP);
    expect((await shop.getConnection(db, SHOP)).apiKey).toBe("");
    await expect(trunc.getUsage(apiKey)).rejects.toMatchObject({ status: 401 });
  });

  it("reports a cancelled sign-in and clears the state", async () => {
    const back = await approve(await shop.startConnect(db, SHOP), "deny");
    expect(back.searchParams.get("error")).toBe("access_denied");
    const result = await shop.finishConnect(db, { state: back.searchParams.get("state"), error: "access_denied" });
    expect(result).toEqual({ shop: SHOP, ok: false, reason: "cancelled" });
    expect((await db.shopSettings.findUnique({ where: { shop: SHOP } }))?.pkceState).toBeNull();
  });

  it("refuses an expired or unknown state", async () => {
    const back = await approve(await shop.startConnect(db, SHOP));
    await db.shopSettings.update({ where: { shop: SHOP }, data: { pkceExpiresAt: new Date(Date.now() - 1000) } });
    expect(await shop.finishConnect(db, { state: back.searchParams.get("state"), code: back.searchParams.get("code") })).toEqual({
      shop: SHOP,
      ok: false,
      reason: "expired",
    });
    expect(await shop.finishConnect(db, { state: "made-up", code: "x" })).toEqual({ shop: null, ok: false, reason: "expired" });
  });

  it("accepts a working pasted key and rejects others", async () => {
    await expect(shop.savePastedKey(db, SHOP, "not-a-key")).rejects.toThrow(/starts with us_/);
    await expect(shop.savePastedKey(db, SHOP, "us_0000000000000000000000")).rejects.toMatchObject({ status: 401 });
    const key = await newApiKey();
    await shop.savePastedKey(db, SHOP, key);
    expect(await shop.getConnection(db, SHOP)).toEqual({ apiKey: key, connectedVia: "apikey" });
    // A pasted key is the merchant's own: disconnect forgets it without revoking it.
    await shop.disconnect(db, SHOP);
    expect((await trunc.getUsage(key)).planType).toBe("enterprise");
  });

  it("shop/redact deletes everything for the shop", async () => {
    await shop.savePastedKey(db, SHOP, await newApiKey());
    await db.link.create({ data: { shop: SHOP, originalUrl: "https://x.example", shortCode: "abc", shortUrl: "https://t/abc" } });
    await shop.deleteShopData(db, SHOP);
    expect(await db.shopSettings.count({ where: { shop: SHOP } })).toBe(0);
    expect(await db.link.count({ where: { shop: SHOP } })).toBe(0);
  });

  it("the callback route finishes sign-in and returns to the admin", async () => {
    const { loader } = await import("../routes/auth.trunc.callback");
    const back = await approve(await shop.startConnect(db, SHOP));
    const req = new Request(`https://trunc-app.example.com/auth/trunc/callback${back.search}`);
    const res = (await loader({ request: req, params: {}, context: {} } as never)) as Response;
    expect(res.status).toBe(302);
    expect(res.headers.get("Location")).toBe(`https://${SHOP}/admin/apps/test-client-id/app/settings?trunc=connected`);
    const bad = (await loader({ request: new Request("https://trunc-app.example.com/auth/trunc/callback?state=nope&code=x"), params: {}, context: {} } as never)) as Response;
    expect(bad.status).toBe(400);
  });
});
