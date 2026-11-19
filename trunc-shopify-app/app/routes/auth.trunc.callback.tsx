import type { LoaderFunctionArgs } from "react-router";
import { redirect } from "react-router";
import db from "../db.server";
import { finishConnect } from "../lib/shop.server";

/**
 * Trunc sends the merchant here after Sign in with Trunc (outside the admin
 * frame). The one-time state identifies the shop that started; then the
 * merchant goes back into the app's settings inside the Shopify admin.
 */
export const loader = async ({ request }: LoaderFunctionArgs) => {
  const q = new URL(request.url).searchParams;
  const result = await finishConnect(db, { state: q.get("state"), code: q.get("code"), error: q.get("error") });
  const status = result.ok ? "connected" : result.reason;
  if (!result.shop) {
    return new Response("This connection link has expired. Open the Trunc app in Shopify and try again.", {
      status: 400,
      headers: { "Content-Type": "text/plain; charset=utf-8" },
    });
  }
  // eslint-disable-next-line no-undef
  return redirect(`https://${result.shop}/admin/apps/${process.env.SHOPIFY_API_KEY}/app/settings?trunc=${status}`);
};
