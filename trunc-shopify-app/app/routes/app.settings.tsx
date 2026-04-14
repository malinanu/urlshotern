import {
  json,
  type ActionFunctionArgs,
  type LoaderFunctionArgs,
} from "@remix-run/node";
import { useLoaderData, useActionData, Form, useNavigation } from "@remix-run/react";
import {
  Page,
  Layout,
  Card,
  TextField,
  Button,
  Banner,
  BlockStack,
  InlineStack,
  Text,
  Badge,
  ProgressBar,
  Divider,
  Box,
  InlineGrid,
  Icon,
  CalloutCard,
} from "@shopify/polaris";
import { CheckIcon, XIcon } from "@shopify/polaris-icons";
import { useState, useCallback } from "react";
import { authenticate } from "../shopify.server";
import { validateAndGetUsage, type UsageResult } from "../lib/trunc.server";
import { prisma } from "../db.server";

export async function loader({ request }: LoaderFunctionArgs) {
  const { session } = await authenticate.admin(request);

  const settings = await prisma.shopSettings.findUnique({
    where: { shop: session.shop },
  });

  let usage: UsageResult | null = null;
  let usageError = "";

  if (settings?.apiKey) {
    try {
      usage = await validateAndGetUsage(settings.apiKey);
    } catch (err: unknown) {
      usageError = err instanceof Error ? err.message : "Could not fetch account info";
    }
  }

  return json({
    apiKey: settings?.apiKey ?? "",
    usage,
    usageError,
  });
}

export async function action({ request }: ActionFunctionArgs) {
  const { session } = await authenticate.admin(request);
  const form = await request.formData();
  const apiKey = String(form.get("apiKey") ?? "").trim();

  if (!apiKey) {
    return json({ error: "API key is required", success: false }, { status: 400 });
  }

  // Validate the key against Trunc before saving
  let usage: UsageResult;
  try {
    usage = await validateAndGetUsage(apiKey);
  } catch (err: unknown) {
    const msg = err instanceof Error ? err.message : "Invalid API key";
    return json({ error: msg, success: false }, { status: 422 });
  }

  await prisma.shopSettings.upsert({
    where: { shop: session.shop },
    update: { apiKey, planType: usage.plan_type, planName: usage.plan_name },
    create: { shop: session.shop, apiKey, planType: usage.plan_type, planName: usage.plan_name },
  });

  return json({ error: null, success: true, usage });
}

const PLAN_BADGE: Record<string, { tone: "info" | "success" | "warning"; label: string }> = {
  free:               { tone: "info",    label: "Free" },
  premium:            { tone: "success", label: "Premium" },
  premium_yearly:     { tone: "success", label: "Premium" },
  enterprise:         { tone: "warning", label: "Enterprise" },
  enterprise_yearly:  { tone: "warning", label: "Enterprise" },
};

function PlanBadge({ planType }: { planType: string }) {
  const b = PLAN_BADGE[planType] ?? { tone: "info" as const, label: planType };
  return <Badge tone={b.tone}>{b.label}</Badge>;
}

function UsageCard({ usage }: { usage: UsageResult }) {
  const urlPct     = usage.max_urls === -1 ? 0 : Math.min(100, usage.url_pct);
  const maxUrlsStr = usage.max_urls === -1 ? "Unlimited" : usage.max_urls.toLocaleString();
  const maxClkStr  = usage.max_clicks === -1 ? "Unlimited" : usage.max_clicks.toLocaleString();

  const features = [
    { label: "API Access",     enabled: true },
    { label: "Custom Domain",  enabled: usage.custom_domain },
    { label: "Team Members",   enabled: usage.team_members > 0 },
    { label: `${usage.analytics_days}-Day Analytics`, enabled: true },
  ];

  return (
    <Card>
      <BlockStack gap="400">
        <InlineStack align="space-between" blockAlign="center">
          <Text as="h2" variant="headingMd">Account</Text>
          <PlanBadge planType={usage.plan_type} />
        </InlineStack>

        <Divider />

        {/* URL Usage */}
        <BlockStack gap="200">
          <InlineStack align="space-between">
            <Text as="span" variant="bodyMd">Short URLs this month</Text>
            <Text as="span" variant="bodyMd" fontWeight="semibold">
              {usage.current_urls.toLocaleString()} / {maxUrlsStr}
            </Text>
          </InlineStack>
          {usage.max_urls !== -1 && (
            <ProgressBar
              progress={urlPct}
              tone={urlPct >= 90 ? "critical" : urlPct >= 70 ? "highlight" : "success"}
              size="small"
            />
          )}
        </BlockStack>

        {/* Click Usage */}
        <BlockStack gap="100">
          <InlineStack align="space-between">
            <Text as="span" variant="bodyMd">Clicks this month</Text>
            <Text as="span" variant="bodyMd" fontWeight="semibold">
              {usage.monthly_clicks.toLocaleString()} / {maxClkStr}
            </Text>
          </InlineStack>
        </BlockStack>

        <Divider />

        {/* Feature grid */}
        <InlineGrid columns={2} gap="200">
          {features.map((f) => (
            <InlineStack key={f.label} gap="150" blockAlign="center">
              <Box>
                <Icon
                  source={f.enabled ? CheckIcon : XIcon}
                  tone={f.enabled ? "success" : "subdued"}
                />
              </Box>
              <Text as="span" variant="bodySm" tone={f.enabled ? undefined : "subdued"}>
                {f.label}
              </Text>
            </InlineStack>
          ))}
        </InlineGrid>

        {/* Upgrade prompt for free */}
        {usage.plan_type === "free" && (
          <>
            <Divider />
            <CalloutCard
              title="Upgrade to Premium"
              illustration=""
              primaryAction={{
                content: "View plans",
                url: "https://trunc.site/pricing",
                external: true,
              }}
            >
              <Text as="p" variant="bodyMd">
                Get 1,000 URLs/month, 365-day analytics, API access, and custom domains.
              </Text>
            </CalloutCard>
          </>
        )}
      </BlockStack>
    </Card>
  );
}

export default function SettingsPage() {
  const { apiKey, usage, usageError } = useLoaderData<typeof loader>();
  const actionData = useActionData<typeof action>();
  const nav = useNavigation();
  const isSubmitting = nav.state === "submitting";

  const [showKey, setShowKey] = useState(false);
  const [keyValue, setKeyValue] = useState(apiKey);

  const handleToggle = useCallback(() => setShowKey((v) => !v), []);

  const hasKey    = Boolean(apiKey);
  const saveError = actionData?.error;
  const saved     = actionData?.success;

  return (
    <Page
      title="Trunc Settings"
      subtitle="Connect your Trunc account to start shortening links"
    >
      <Layout>
        <Layout.Section>
          <BlockStack gap="400">
            {saveError && (
              <Banner tone="critical" title="Connection failed">
                <Text as="p">{saveError}</Text>
              </Banner>
            )}

            {saved && (
              <Banner tone="success" title="API key saved">
                <Text as="p">Your Trunc account is connected and ready to use.</Text>
              </Banner>
            )}

            {usageError && hasKey && (
              <Banner tone="warning" title="Could not reach Trunc">
                <Text as="p">{usageError}. Check your API key or try again.</Text>
              </Banner>
            )}

            {/* API Key Card */}
            <Card>
              <BlockStack gap="400">
                <BlockStack gap="100">
                  <Text as="h2" variant="headingMd">API Key</Text>
                  <Text as="p" variant="bodyMd" tone="subdued">
                    Get your free API key from{" "}
                    <a
                      href="https://trunc.site/dashboard/api-keys"
                      target="_blank"
                      rel="noopener noreferrer"
                    >
                      trunc.site → Dashboard → API Keys
                    </a>
                    .
                  </Text>
                </BlockStack>

                <Form method="post">
                  <BlockStack gap="300">
                    <TextField
                      label="Trunc API Key"
                      name="apiKey"
                      type={showKey ? "text" : "password"}
                      value={keyValue}
                      onChange={setKeyValue}
                      placeholder="us_••••••••••••••••"
                      autoComplete="off"
                      connectedRight={
                        <Button onClick={handleToggle} variant="tertiary">
                          {showKey ? "Hide" : "Show"}
                        </Button>
                      }
                      helpText={
                        hasKey
                          ? "Key saved. Enter a new key to replace it."
                          : "Your API key starts with us_"
                      }
                    />
                    <InlineStack gap="200">
                      <Button
                        submit
                        variant="primary"
                        loading={isSubmitting}
                        disabled={!keyValue.trim()}
                      >
                        {hasKey ? "Update Key" : "Connect Trunc"}
                      </Button>
                      {hasKey && !isSubmitting && (
                        <Text as="span" variant="bodySm" tone="subdued">
                          ✓ Connected
                        </Text>
                      )}
                    </InlineStack>
                  </BlockStack>
                </Form>
              </BlockStack>
            </Card>

            {/* Onboarding if no key yet */}
            {!hasKey && (
              <Card>
                <BlockStack gap="300">
                  <Text as="h2" variant="headingMd">Get started in 2 steps</Text>
                  <BlockStack gap="200">
                    <InlineStack gap="200" blockAlign="start">
                      <Badge>1</Badge>
                      <Text as="p" variant="bodyMd">
                        Create a free account at{" "}
                        <a href="https://trunc.site/register" target="_blank" rel="noopener noreferrer">
                          trunc.site
                        </a>
                      </Text>
                    </InlineStack>
                    <InlineStack gap="200" blockAlign="start">
                      <Badge>2</Badge>
                      <Text as="p" variant="bodyMd">
                        Copy your API key from{" "}
                        <a
                          href="https://trunc.site/dashboard/api-keys"
                          target="_blank"
                          rel="noopener noreferrer"
                        >
                          Dashboard → API Keys
                        </a>{" "}
                        and paste it above
                      </Text>
                    </InlineStack>
                  </BlockStack>
                </BlockStack>
              </Card>
            )}
          </BlockStack>
        </Layout.Section>

        {/* Account sidebar */}
        <Layout.Section variant="oneThird">
          {usage ? (
            <UsageCard usage={usage} />
          ) : hasKey ? (
            <Card>
              <Text as="p" variant="bodyMd" tone="subdued">
                Account info unavailable.
              </Text>
            </Card>
          ) : (
            <Card>
              <BlockStack gap="200">
                <Text as="h2" variant="headingMd">Free plan includes</Text>
                <BlockStack gap="100">
                  {["10 short links/month", "Basic click analytics", "QR codes", "Link expiration"].map((f) => (
                    <InlineStack key={f} gap="150" blockAlign="center">
                      <Icon source={CheckIcon} tone="success" />
                      <Text as="span" variant="bodySm">{f}</Text>
                    </InlineStack>
                  ))}
                </BlockStack>
              </BlockStack>
            </Card>
          )}
        </Layout.Section>
      </Layout>
    </Page>
  );
}
