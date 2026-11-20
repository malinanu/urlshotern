import { useEffect, useState } from "react";
import type { ActionFunctionArgs, HeadersFunction, LoaderFunctionArgs } from "react-router";
import { Form, redirect, useActionData, useLoaderData, useNavigation } from "react-router";
import { useAppBridge } from "@shopify/app-bridge-react";
import { boundary } from "@shopify/shopify-app-react-router/server";
import { authenticate } from "../shopify.server";
import db from "../db.server";
import { getConnection } from "../lib/shop.server";
import { getUsage, shortenUrl, TruncError, type UsageResult } from "../lib/trunc.server";

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const { session, admin } = await authenticate.admin(request);
  const { apiKey } = await getConnection(db, session.shop);
  if (!apiKey) throw redirect("/app/settings");

  const links = await db.link.findMany({ where: { shop: session.shop }, orderBy: { createdAt: "desc" }, take: 50 });

  let usage: UsageResult | null = null;
  try {
    usage = await getUsage(apiKey);
  } catch {
    // Show the links even if Trunc is unreachable.
  }

  // The storefront's own domain, so product links point at it.
  let storefront = `https://${session.shop}`;
  try {
    const res = await admin.graphql(`#graphql
      query TruncShopDomain { shop { primaryDomain { url } } }`);
    const json = await res.json();
    storefront = json.data?.shop?.primaryDomain?.url || storefront;
  } catch {
    // Fall back to the myshopify.com domain.
  }

  return {
    links: links.map((l) => ({ ...l, createdAt: l.createdAt.toISOString() })),
    usage,
    storefront: storefront.replace(/\/+$/, ""),
  };
};

export const action = async ({ request }: ActionFunctionArgs) => {
  const { session } = await authenticate.admin(request);
  const { apiKey } = await getConnection(db, session.shop);
  if (!apiKey) throw redirect("/app/settings");

  const form = await request.formData();
  const url = String(form.get("url") ?? "").trim();
  const customCode = String(form.get("customCode") ?? "").trim();
  if (!/^https?:\/\//i.test(url)) return { error: "Enter a full web address, starting with https://." };

  try {
    const result = await shortenUrl(apiKey, url, customCode || undefined);
    await db.link.upsert({
      where: { shop_shortCode: { shop: session.shop, shortCode: result.short_code } },
      create: { shop: session.shop, originalUrl: url, shortCode: result.short_code, shortUrl: result.short_url },
      update: {},
    });
    return { shortUrl: result.short_url };
  } catch (err) {
    return { error: err instanceof TruncError ? err.message : "Trunc could not shorten that link." };
  }
};

function limitText(u: UsageResult): string {
  return u.maxUrls < 0 ? `${u.currentUrls} short links · unlimited` : `${u.currentUrls} of ${u.maxUrls} short links`;
}

export default function Links() {
  const { links, usage, storefront } = useLoaderData<typeof loader>();
  const result = useActionData<typeof action>();
  const nav = useNavigation();
  const shopify = useAppBridge();
  const [url, setUrl] = useState("");
  const busy = nav.state === "submitting";

  useEffect(() => {
    if (result && "shortUrl" in result && result.shortUrl) {
      shopify.toast.show("Short link created");
    }
  }, [result, shopify]);

  const pickProduct = async () => {
    const picked = await shopify.resourcePicker({ type: "product", action: "select", multiple: false });
    const handle = (picked?.[0] as { handle?: string } | undefined)?.handle;
    if (handle) setUrl(`${storefront}/products/${handle}`);
  };

  const copy = async (text: string) => {
    await navigator.clipboard.writeText(text).catch(() => undefined);
    shopify.toast.show("Copied");
  };

  return (
    <s-page heading="Short links">
      {usage && (
        <s-badge slot="primary-action" tone={usage.urlLimitReached ? "critical" : "info"}>
          {usage.planName} · {limitText(usage)}
        </s-badge>
      )}

      <s-section heading="Shorten a link">
        <Form method="post">
          <s-stack gap="base">
            <s-url-field
              label="Web address"
              name="url"
              value={url}
              placeholder={`${storefront}/products/…`}
              onInput={(e: Event) => setUrl((e.currentTarget as HTMLInputElement).value)}
            />
            <s-text-field
              label="Custom back-half (optional)"
              name="customCode"
              details="Letters, numbers and dashes. Available on Premium and above."
            />
            {result && "error" in result && result.error && <s-banner tone="critical">{result.error}</s-banner>}
            {result && "shortUrl" in result && result.shortUrl && (
              <s-banner tone="success" heading="Short link ready">
                <s-stack direction="inline" gap="base">
                  <s-link href={result.shortUrl} target="_blank">
                    {result.shortUrl}
                  </s-link>
                  <s-button onClick={() => copy(result.shortUrl!)}>Copy</s-button>
                </s-stack>
              </s-banner>
            )}
            <s-stack direction="inline" gap="base">
              <s-button type="submit" variant="primary" {...(busy ? { loading: true } : {})}>
                Shorten
              </s-button>
              <s-button onClick={pickProduct}>Pick a product</s-button>
            </s-stack>
          </s-stack>
        </Form>
      </s-section>

      <s-section heading="Your links">
        {links.length === 0 ? (
          <s-paragraph>No links yet. Shorten a product or page above.</s-paragraph>
        ) : (
          <s-table>
            <s-table-header-row>
              <s-table-header>Short link</s-table-header>
              <s-table-header>Goes to</s-table-header>
              <s-table-header>Created</s-table-header>
              <s-table-header></s-table-header>
            </s-table-header-row>
            <s-table-body>
              {links.map((l) => (
                <s-table-row key={l.id}>
                  <s-table-cell>
                    <s-link href={l.shortUrl} target="_blank">
                      {l.shortUrl.replace(/^https?:\/\//, "")}
                    </s-link>
                  </s-table-cell>
                  <s-table-cell>{l.originalUrl}</s-table-cell>
                  <s-table-cell>{new Date(l.createdAt).toLocaleDateString()}</s-table-cell>
                  <s-table-cell>
                    <s-stack direction="inline" gap="small-200">
                      <s-button variant="tertiary" onClick={() => copy(l.shortUrl)}>
                        Copy
                      </s-button>
                      <s-link href={`/app/analytics/${encodeURIComponent(l.shortCode)}`}>Clicks</s-link>
                    </s-stack>
                  </s-table-cell>
                </s-table-row>
              ))}
            </s-table-body>
          </s-table>
        )}
      </s-section>
    </s-page>
  );
}

export const headers: HeadersFunction = (headersArgs) => boundary.headers(headersArgs);
