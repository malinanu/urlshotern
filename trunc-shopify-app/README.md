# Trunc URL Shortener for Shopify

An embedded Shopify app: merchants make short links to their products, collections and pages with [Trunc](https://trunc.site), and see the clicks behind each one, without leaving the Shopify admin.

Built on Shopify's [React Router app template](https://github.com/Shopify/shopify-app-template-react-router). The older Remix template is archived. The UI uses Polaris web components.

## What it does

- **Short links:**
  - Shorten any URL, with an optional custom back-half.
  - **Pick a product** builds the product's URL on the shop's primary storefront domain.
  - Copy links, and see the shop's links in a table.
- **Clicks:** total clicks, the last 14 days, and top countries for each link, with a link to the full analytics on Trunc.
- **Settings:**
  - **Sign in with Trunc**: OAuth 2.0 with PKCE, client id `trunc-shopify`.
  - Or paste an API key; it is checked before it is saved.
  - Shows the plan and usage.
  - **Disconnect** revokes a signed-in key.
- **Privacy webhooks:**
  - `customers/data_request` and `customers/redact`: the app stores no customer data.
  - `shop/redact`: deletes the shop's settings, links and sessions.

## How Sign in with Trunc works here

1. In Settings, **Sign in with Trunc** stores a one-time `state` and PKCE verifier on the shop's `ShopSettings` row (10 minutes). It then opens `<TRUNC_BASE_URL>/connect` in the top window, because Trunc's page can't run inside the admin iframe.
2. Trunc returns the merchant to `<SHOPIFY_APP_URL>/auth/trunc/callback`. The state finds the shop, and is cleared before anything else.
3. The code is exchanged server side for the shop's own API key, and the merchant goes back to `https://<shop>/admin/apps/<api key>/app/settings`.

The Trunc server only accepts this client's redirect if it is listed in its `SHOPIFY_CONNECT_REDIRECT_URI` setting, and the client stays off until it is set. Set it to `<SHOPIFY_APP_URL>/auth/trunc/callback`. During development, the CLI's tunnel URL changes each run, so pasting an API key is simpler. Keys never reach the browser.

## Run it

You need Node 22.12+, a [Shopify Partner](https://partners.shopify.com) account and a development store.

```bash
npm install
npm run config:link   # pick or create the app in your Partner account
npm run dev           # Shopify CLI: tunnel, env vars, install link
```

Environment:

| Variable | Purpose |
| --- | --- |
| `TRUNC_BASE_URL` | Trunc server. Defaults to `https://trunc.site`. |
| `SHOPIFY_API_KEY`, `SHOPIFY_API_SECRET`, `SHOPIFY_APP_URL`, `SCOPES` | Set by the Shopify CLI in development; set them yourself in production. |

The database is Prisma with SQLite (`prisma/dev.sqlite`). `npm run setup` runs `prisma generate` and `migrate deploy`.

## Checks

```bash
npm run typecheck
npm run lint
npm run build
npx prisma migrate deploy
TRUNC_BASE_URL=http://localhost:8099 TRUNC_TEST_EMAIL=you@example.com TRUNC_TEST_PASSWORD=… npm test
```

`npm test` runs `app/lib/connect.test.ts` against a running Trunc server started with `SHOPIFY_CONNECT_REDIRECT_URI=https://trunc-app.example.com/auth/trunc/callback`, using a real account. It covers:
- the full sign-in, including the authorize step;
- a reused, expired, cancelled or unknown state;
- shortening, clicks and plan usage;
- revoke on disconnect, and pasted keys;
- `shop/redact`;
- the callback route's redirect.

## Publishing

See `docs/INTEGRATIONS.md` in the Trunc repository for the App Store checklist.
