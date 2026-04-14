import { json, redirect, type LoaderFunctionArgs } from "@remix-run/node";
import { useLoaderData } from "@remix-run/react";
import {
  Page,
  Layout,
  Card,
  DataTable,
  Text,
  BlockStack,
  InlineStack,
  Button,
  Badge,
  Divider,
  Box,
} from "@shopify/polaris";
import { authenticate } from "../shopify.server";
import { getAnalytics } from "../lib/trunc.server";
import { prisma } from "../db.server";

export async function loader({ request, params }: LoaderFunctionArgs) {
  const { session } = await authenticate.admin(request);
  const code = params.code!;

  const settings = await prisma.shopSettings.findUnique({
    where: { shop: session.shop },
  });
  if (!settings?.apiKey) return redirect("/app/settings");

  const link = await prisma.link.findFirst({
    where: { shop: session.shop, shortCode: code },
  });
  if (!link) throw new Response("Not Found", { status: 404 });

  const analytics = await getAnalytics(settings.apiKey, code);

  await prisma.link.update({
    where: { id: link.id },
    data: { clicks: analytics.total_clicks },
  });

  return json({ link, analytics });
}

export default function AnalyticsPage() {
  const { link, analytics } = useLoaderData<typeof loader>();

  const countryRows = [...analytics.countries]
    .sort((a, b) => b.clicks - a.clicks)
    .slice(0, 10)
    .map((c) => [c.country, c.clicks.toLocaleString()]);

  const dailyRows = analytics.daily
    .slice(-14)
    .map((d) => [d.date, d.clicks.toLocaleString()]);

  const maxDailyClicks = Math.max(...analytics.daily.map((d) => d.clicks), 1);

  return (
    <Page
      title="Link Analytics"
      subtitle={link.shortUrl}
      backAction={{ content: "All Links", url: "/app" }}
      primaryAction={{
        content: "View on Trunc",
        url: `https://trunc.site/analytics/${link.shortCode}`,
        external: true,
      }}
    >
      <Layout>
        {/* Stats Row */}
        <Layout.Section>
          <InlineStack gap="400" wrap>
            {[
              { label: "Total Clicks", value: analytics.total_clicks.toLocaleString() },
              { label: "Countries",    value: analytics.countries.length.toString() },
              { label: "Last 14 Days", value: analytics.daily.slice(-14).reduce((s, d) => s + d.clicks, 0).toLocaleString() },
            ].map((stat) => (
              <Card key={stat.label}>
                <Box minWidth="140px">
                  <BlockStack gap="100">
                    <Text as="p" variant="bodySm" tone="subdued">{stat.label}</Text>
                    <Text as="p" variant="heading2xl" fontWeight="bold">{stat.value}</Text>
                  </BlockStack>
                </Box>
              </Card>
            ))}
          </InlineStack>
        </Layout.Section>

        {/* Daily Chart (simple table) */}
        <Layout.Section>
          <Card>
            <BlockStack gap="300">
              <Text as="h2" variant="headingMd">Daily Clicks — Last 14 Days</Text>
              <Divider />
              {dailyRows.length > 0 ? (
                <BlockStack gap="100">
                  {analytics.daily.slice(-14).map((d) => {
                    const barPct = Math.round((d.clicks / maxDailyClicks) * 100);
                    return (
                      <InlineStack key={d.date} gap="300" blockAlign="center" wrap={false}>
                        <Text as="span" variant="bodySm" tone="subdued" alignment="end">
                          <span style={{ display: "inline-block", minWidth: 80 }}>{d.date}</span>
                        </Text>
                        <div style={{ flex: 1, background: "#f0f0f1", borderRadius: 4, height: 10 }}>
                          <div
                            style={{
                              width: `${barPct}%`,
                              height: "100%",
                              background: "#2271b1",
                              borderRadius: 4,
                              minWidth: barPct > 0 ? 4 : 0,
                            }}
                          />
                        </div>
                        <Text as="span" variant="bodySm" fontWeight="semibold">
                          <span style={{ display: "inline-block", minWidth: 40, textAlign: "right" }}>
                            {d.clicks.toLocaleString()}
                          </span>
                        </Text>
                      </InlineStack>
                    );
                  })}
                </BlockStack>
              ) : (
                <Text as="p" tone="subdued">No daily data yet.</Text>
              )}
            </BlockStack>
          </Card>
        </Layout.Section>

        {/* Countries */}
        <Layout.Section variant="oneThird">
          <Card>
            <BlockStack gap="300">
              <Text as="h2" variant="headingMd">Top Countries</Text>
              <Divider />
              {countryRows.length > 0 ? (
                <DataTable
                  columnContentTypes={["text", "numeric"]}
                  headings={["Country", "Clicks"]}
                  rows={countryRows}
                />
              ) : (
                <Text as="p" tone="subdued">No country data yet.</Text>
              )}
            </BlockStack>
          </Card>
        </Layout.Section>

        {/* Link Details */}
        <Layout.Section variant="oneThird">
          <Card>
            <BlockStack gap="300">
              <Text as="h2" variant="headingMd">Link Details</Text>
              <Divider />
              <BlockStack gap="200">
                <BlockStack gap="050">
                  <Text as="p" variant="bodySm" tone="subdued">Short URL</Text>
                  <a href={link.shortUrl} target="_blank" rel="noopener noreferrer">
                    <Text as="p" variant="bodyMd" fontWeight="semibold">{link.shortUrl}</Text>
                  </a>
                </BlockStack>
                <BlockStack gap="050">
                  <Text as="p" variant="bodySm" tone="subdued">Destination</Text>
                  <Text as="p" variant="bodyMd">{link.originalUrl}</Text>
                </BlockStack>
                <BlockStack gap="050">
                  <Text as="p" variant="bodySm" tone="subdued">Created</Text>
                  <Text as="p" variant="bodyMd">
                    {new Date(link.createdAt).toLocaleDateString("en-US", {
                      year: "numeric", month: "long", day: "numeric",
                    })}
                  </Text>
                </BlockStack>
                <InlineStack>
                  <Badge>{link.clicks.toLocaleString()} total clicks</Badge>
                </InlineStack>
              </BlockStack>
            </BlockStack>
          </Card>
        </Layout.Section>
      </Layout>
    </Page>
  );
}
