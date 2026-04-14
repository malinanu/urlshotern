import {
  json,
  redirect,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
} from "@remix-run/node";
import {
  useLoaderData,
  useActionData,
  Form,
  useNavigation,
  Link,
} from "@remix-run/react";
import {
  Page,
  Layout,
  Card,
  TextField,
  Button,
  DataTable,
  Banner,
  InlineStack,
  BlockStack,
  Text,
  Badge,
  ProgressBar,
  Divider,
  Tooltip,
  EmptyState,
} from "@shopify/polaris";
import { shortenUrl, validateAndGetUsage, type UsageResult } from "../lib/trunc.server";
import { prisma } from "../db.server";
import { authenticate } from "../shopify.server";

export async function loader({ request }: LoaderFunctionArgs) {
  const { session } = await authenticate.admin(request);

  const settings = await prisma.shopSettings.findUnique({
    where: { shop: session.shop },
  });

  // No API key — send to settings
  if (!settings?.apiKey) {
    return redirect("/app/settings");
  }

  const links = await prisma.link.findMany({
    where: { shop: session.shop },
    orderBy: { createdAt: "desc" },
    take: 50,
  });

  let usage: UsageResult | null = null;
  try {
    usage = await validateAndGetUsage(settings.apiKey);
  } catch {
    // Non-fatal — show links without plan info
  }

  return json({ links, usage, apiKey: settings.apiKey });
}

export async function action({ request }: ActionFunctionArgs) {
  const { session } = await authenticate.admin(request);

  const settings = await prisma.shopSettings.findUnique({
    where: { shop: session.shop },
  });
  if (!settings?.apiKey) return redirect("/app/settings");

  const form = await request.formData();
  const url        = String(form.get("url") ?? "").trim();
  const customCode = String(form.get("customCode") ?? "").trim();

  if (!url) return json({ error: "URL is required" }, { status: 400 });

  // Check limit before hitting the API
  try {
    const usage = await validateAndGetUsage(settings.apiKey);
    if (usage.url_limit_reached) {
      return json(
        { error: `Monthly limit of ${usage.max_urls} URLs reached. Upgrade at trunc.site/pricing.` },
        { status: 422 }
      );
    }
  } catch {
    // Let the shorten call itself return any auth/limit errors
  }

  try {
    const result = await shortenUrl(settings.apiKey, url, customCode || undefined);
    await prisma.link.create({
      data: {
        shop: session.shop,
        originalUrl: url,
        shortCode: result.short_code,
        shortUrl: result.short_url,
        clicks: 0,
      },
    });
    return redirect("/app");
  } catch (err: unknown) {
    const message = err instanceof Error ? err.message : "Unknown error";
    return json({ error: message }, { status: 500 });
  }
}

const PLAN_BADGE: Record<string, { tone: "info" | "success" | "warning"; label: string }> = {
  free:               { tone: "info",    label: "Free" },
  premium:            { tone: "success", label: "Premium" },
  premium_yearly:     { tone: "success", label: "Premium" },
  enterprise:         { tone: "warning", label: "Enterprise" },
  enterprise_yearly:  { tone: "warning", label: "Enterprise" },
};

function UsageSummary({ usage }: { usage: UsageResult }) {
  const b       = PLAN_BADGE[usage.plan_type] ?? { tone: "info" as const, label: usage.plan_type };
  const urlPct  = usage.max_urls === -1 ? 0 : Math.min(100, usage.url_pct);
  const maxStr  = usage.max_urls === -1 ? "∞" : usage.max_urls.toLocaleString();

  return (
    <InlineStack gap="400" blockAlign="center" wrap={false}>
      <Badge tone={b.tone}>{b.label}</Badge>
      <BlockStack gap="050">
        <InlineStack gap="100" blockAlign="center">
          <Text as="span" variant="bodySm" tone="subdued">URLs:</Text>
          <Text as="span" variant="bodySm" fontWeight="semibold">
            {usage.current_urls} / {maxStr}
          </Text>
        </InlineStack>
        {usage.max_urls !== -1 && (
          <div style={{ width: 120 }}>
            <ProgressBar
              progress={urlPct}
              tone={urlPct >= 90 ? "critical" : urlPct >= 70 ? "highlight" : "success"}
              size="small"
            />
          </div>
        )}
      </BlockStack>
      <Link to="/app/settings">
        <Button variant="plain" size="slim">Settings</Button>
      </Link>
    </InlineStack>
  );
}

export default function Index() {
  const { links, usage } = useLoaderData<typeof loader>();
  const actionData       = useActionData<typeof action>();
  const nav              = useNavigation();
  const isSubmitting     = nav.state === "submitting";

  const limitReached = usage?.url_limit_reached ?? false;
  const approaching  = usage?.url_approaching   ?? false;

  const rows = links.map((l) => [
    <Tooltip content={l.shortUrl} key={`url-${l.id}`}>
      <a href={l.shortUrl} target="_blank" rel="noopener noreferrer">
        {l.shortUrl}
      </a>
    </Tooltip>,
    <Tooltip content={l.originalUrl} key={`orig-${l.id}`}>
      <Text as="span" variant="bodySm" tone="subdued">
        {l.originalUrl.length > 50 ? l.originalUrl.slice(0, 50) + "…" : l.originalUrl}
      </Text>
    </Tooltip>,
    <Badge key={`clicks-${l.id}`}>{l.clicks.toLocaleString()}</Badge>,
    <Text as="span" variant="bodySm" key={`date-${l.id}`}>
      {new Date(l.createdAt).toLocaleDateString()}
    </Text>,
    <Link to={`/app/analytics/${l.shortCode}`} key={`action-${l.id}`}>
      <Button size="slim" variant="plain">Analytics</Button>
    </Link>,
  ]);

  return (
    <Page
      title="Trunc URL Shortener"
      titleMetadata={usage ? <UsageSummary usage={usage} /> : undefined}
    >
      <Layout>
        <Layout.Section>
          <BlockStack gap="400">
            {/* Limit alerts */}
            {limitReached && (
              <Banner
                tone="critical"
                title="Monthly URL limit reached"
                action={{ content: "Upgrade plan", url: "https://trunc.site/pricing", external: true }}
              >
                <Text as="p">
                  You have used all {usage!.max_urls} URLs for this month.
                  Upgrade to continue shortening links.
                </Text>
              </Banner>
            )}

            {approaching && !limitReached && (
              <Banner
                tone="warning"
                title="Approaching URL limit"
                action={{ content: "View plans", url: "https://trunc.site/pricing", external: true }}
              >
                <Text as="p">
                  {usage!.current_urls} of {usage!.max_urls} URLs used this month.
                </Text>
              </Banner>
            )}

            {/* Error from action */}
            {actionData?.error && (
              <Banner tone="critical" title="Could not shorten URL">
                <Text as="p">{actionData.error}</Text>
              </Banner>
            )}

            {/* Shorten Form */}
            <Card>
              <BlockStack gap="400">
                <Text as="h2" variant="headingMd">Shorten a URL</Text>
                <Form method="post">
                  <BlockStack gap="300">
                    <TextField
                      label="URL to shorten"
                      name="url"
                      type="url"
                      placeholder="https://mystore.myshopify.com/products/summer-sale"
                      autoComplete="off"
                      disabled={limitReached}
                    />
                    <TextField
                      label="Custom code (optional)"
                      name="customCode"
                      placeholder="summer-sale"
                      autoComplete="off"
                      helpText="Leave blank for an auto-generated code"
                      disabled={limitReached}
                    />
                    <InlineStack>
                      <Button
                        submit
                        variant="primary"
                        loading={isSubmitting}
                        disabled={limitReached}
                      >
                        Shorten URL
                      </Button>
                    </InlineStack>
                  </BlockStack>
                </Form>
              </BlockStack>
            </Card>

            {/* Links Table */}
            <Card>
              <BlockStack gap="300">
                <InlineStack align="space-between">
                  <Text as="h2" variant="headingMd">Your Short Links</Text>
                  <Text as="span" variant="bodySm" tone="subdued">
                    {links.length} link{links.length !== 1 ? "s" : ""}
                  </Text>
                </InlineStack>
                <Divider />
                {links.length > 0 ? (
                  <DataTable
                    columnContentTypes={["text", "text", "text", "text", "text"]}
                    headings={["Short URL", "Destination", "Clicks", "Created", ""]}
                    rows={rows}
                    truncate
                  />
                ) : (
                  <EmptyState
                    heading="No links yet"
                    image=""
                  >
                    <Text as="p">Shorten your first URL above to get started.</Text>
                  </EmptyState>
                )}
              </BlockStack>
            </Card>
          </BlockStack>
        </Layout.Section>
      </Layout>
    </Page>
  );
}
