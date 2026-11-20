import { useEffect } from "react";
import type { ActionFunctionArgs, HeadersFunction, LoaderFunctionArgs } from "react-router";
import { Form, useActionData, useLoaderData, useNavigation, useSearchParams } from "react-router";
import { boundary } from "@shopify/shopify-app-react-router/server";
import { authenticate } from "../shopify.server";
import db from "../db.server";
import { disconnect, getConnection, savePastedKey, startConnect } from "../lib/shop.server";
import { getUsage, truncBaseUrl, TruncError, type UsageResult } from "../lib/trunc.server";

export const loader = async ({ request }: LoaderFunctionArgs) => {
  const { session } = await authenticate.admin(request);
  const { apiKey, connectedVia } = await getConnection(db, session.shop);
  let usage: UsageResult | null = null;
  let usageError: string | null = null;
  if (apiKey) {
    try {
      usage = await getUsage(apiKey);
    } catch (err) {
      usageError = err instanceof TruncError ? err.message : "Could not reach Trunc.";
    }
  }
  return { connected: Boolean(apiKey), connectedVia, usage, usageError, host: new URL(truncBaseUrl()).host, base: truncBaseUrl() };
};

export const action = async ({ request }: ActionFunctionArgs) => {
  const { session } = await authenticate.admin(request);
  const form = await request.formData();
  const intent = String(form.get("intent") || "");
  try {
    if (intent === "connect") {
      return { redirectTo: await startConnect(db, session.shop) };
    }
    if (intent === "apikey") {
      await savePastedKey(db, session.shop, String(form.get("apiKey") || ""));
      return { notice: "Connected with your API key." };
    }
    if (intent === "disconnect") {
      await disconnect(db, session.shop);
      return { notice: "Disconnected from Trunc." };
    }
  } catch (err) {
    return { error: err instanceof TruncError ? err.message : "Something went wrong." };
  }
  return { error: "Unknown action." };
};

const NOTICES: Record<string, { tone: "success" | "warning" | "critical"; text: string }> = {
  connected: { tone: "success", text: "Your store is connected to Trunc." },
  cancelled: { tone: "warning", text: "Connecting was cancelled." },
  expired: { tone: "critical", text: "That connection attempt expired. Please try again." },
  failed: { tone: "critical", text: "Trunc could not finish connecting your store. Try again, or paste an API key." },
};

export default function Settings() {
  const { connected, connectedVia, usage, usageError, host, base } = useLoaderData<typeof loader>();
  const result = useActionData<typeof action>();
  const nav = useNavigation();
  const [params] = useSearchParams();
  const returned = NOTICES[params.get("trunc") || ""];
  const busy = nav.state === "submitting";

  // Trunc's sign-in page can't open inside the Shopify admin frame.
  useEffect(() => {
    if (result && "redirectTo" in result && result.redirectTo) window.open(result.redirectTo, "_top");
  }, [result]);

  return (
    <s-page heading="Settings">
      {returned && <s-banner tone={returned.tone}>{returned.text}</s-banner>}
      {result && "notice" in result && result.notice && <s-banner tone="success">{result.notice}</s-banner>}
      {result && "error" in result && result.error && <s-banner tone="critical">{result.error}</s-banner>}

      <s-section heading="Trunc account">
        {connected ? (
          <s-stack gap="base">
            <s-paragraph>
              Connected to {host} {connectedVia === "signin" ? "with Sign in with Trunc" : "with an API key"}.
            </s-paragraph>
            {usage && (
              <s-paragraph>
                Plan: {usage.planName}.{" "}
                {usage.maxUrls < 0 ? `${usage.currentUrls} short links (unlimited).` : `${usage.currentUrls} of ${usage.maxUrls} short links used.`}
              </s-paragraph>
            )}
            {usageError && <s-banner tone="warning">{usageError}</s-banner>}
            <Form method="post">
              <input type="hidden" name="intent" value="disconnect" />
              <s-button type="submit" tone="critical" {...(busy ? { loading: true } : {})}>
                Disconnect
              </s-button>
            </Form>
          </s-stack>
        ) : (
          <s-stack gap="large">
            <s-paragraph>Connect your Trunc account to shorten product and page links from Shopify.</s-paragraph>
            <Form method="post">
              <input type="hidden" name="intent" value="connect" />
              <s-button type="submit" variant="primary" {...(busy ? { loading: true } : {})}>
                Sign in with Trunc
              </s-button>
            </Form>
            <s-divider />
            <Form method="post">
              <input type="hidden" name="intent" value="apikey" />
              <s-stack gap="base">
                <s-password-field
                  label="Or paste an API key"
                  name="apiKey"
                  placeholder="us_…"
                  details={`Create one on your API keys page at ${host}/api-keys.`}
                />
                <s-button type="submit">Save key</s-button>
              </s-stack>
            </Form>
            <s-paragraph>
              No account yet?{" "}
              <s-link href={`${base}/register`} target="_blank">
                Create a free Trunc account
              </s-link>
              .
            </s-paragraph>
          </s-stack>
        )}
      </s-section>
    </s-page>
  );
}

export const headers: HeadersFunction = (headersArgs) => boundary.headers(headersArgs);
