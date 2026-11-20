import type { HeadersFunction, LoaderFunctionArgs } from "react-router";
import { redirect, useLoaderData } from "react-router";
import { boundary } from "@shopify/shopify-app-react-router/server";
import { authenticate } from "../shopify.server";
import db from "../db.server";
import { getConnection } from "../lib/shop.server";
import { getAnalytics, truncBaseUrl, TruncError } from "../lib/trunc.server";

export const loader = async ({ request, params }: LoaderFunctionArgs) => {
  const { session } = await authenticate.admin(request);
  const { apiKey } = await getConnection(db, session.shop);
  if (!apiKey) throw redirect("/app/settings");

  const code = params.code || "";
  // Only links this shop created are shown here.
  const link = await db.link.findUnique({ where: { shop_shortCode: { shop: session.shop, shortCode: code } } });
  if (!link) throw redirect("/app");

  try {
    const stats = await getAnalytics(apiKey, code);
    return { link: { shortUrl: link.shortUrl, originalUrl: link.originalUrl }, stats, error: null, fullUrl: `${truncBaseUrl()}/analytics?shortCode=${encodeURIComponent(code)}` };
  } catch (err) {
    return {
      link: { shortUrl: link.shortUrl, originalUrl: link.originalUrl },
      stats: null,
      error: err instanceof TruncError ? err.message : "Could not load clicks.",
      fullUrl: `${truncBaseUrl()}/analytics?shortCode=${encodeURIComponent(code)}`,
    };
  }
};

export default function LinkAnalytics() {
  const { link, stats, error, fullUrl } = useLoaderData<typeof loader>();
  const last14 = stats ? stats.daily.slice(-14) : [];
  const max = Math.max(1, ...last14.map((d) => d.clicks));

  return (
    <s-page heading={link.shortUrl.replace(/^https?:\/\//, "")}>
      <s-link slot="breadcrumb-actions" href="/app">
        Short links
      </s-link>
      <s-button slot="primary-action" href={fullUrl} target="_blank">
        Full analytics on Trunc
      </s-button>

      <s-section>
        <s-paragraph>Goes to {link.originalUrl}</s-paragraph>
        {error && <s-banner tone="critical">{error}</s-banner>}
        {stats && (
          <s-stack gap="base">
            <s-heading>{stats.totalClicks.toLocaleString()} clicks</s-heading>
            {stats.lastClickAt && <s-text color="subdued">Last click {new Date(stats.lastClickAt).toLocaleString()}</s-text>}
          </s-stack>
        )}
      </s-section>

      {stats && last14.length > 0 && (
        <s-section heading="Last 14 days">
          <s-stack gap="small-200">
            {last14.map((d) => (
              <s-stack key={d.date} direction="inline" gap="base" alignItems="center">
                <s-text>{d.date}</s-text>
                <div
                  aria-hidden
                  style={{ height: 8, borderRadius: 4, background: "#2c6ecb", width: `${Math.max(2, (d.clicks / max) * 240)}px` }}
                />
                <s-text>{d.clicks}</s-text>
              </s-stack>
            ))}
          </s-stack>
        </s-section>
      )}

      {stats && stats.countries.length > 0 && (
        <s-section heading="Top countries">
          <s-table>
            <s-table-header-row>
              <s-table-header>Country</s-table-header>
              <s-table-header>Clicks</s-table-header>
            </s-table-header-row>
            <s-table-body>
              {stats.countries.slice(0, 10).map((c) => (
                <s-table-row key={c.country_code}>
                  <s-table-cell>{c.country_code || "Unknown"}</s-table-cell>
                  <s-table-cell>{c.clicks}</s-table-cell>
                </s-table-row>
              ))}
            </s-table-body>
          </s-table>
        </s-section>
      )}
    </s-page>
  );
}

export const headers: HeadersFunction = (headersArgs) => boundary.headers(headersArgs);
