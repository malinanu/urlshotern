# Integrations: Chrome, WordPress, Shopify and Canva

Four plugins let people use Trunc where they already work. Each one connects to a Trunc account with **Sign in with Trunc** (OAuth 2.0), or with an API key where the platform allows it.

| Plugin | Folder | How people install it today | Sign-in |
|---|---|---|---|
| Chrome extension | `trunc-chrome-extension/` | Download the .zip from `/integrations` ("Load unpacked"), or the Chrome Web Store once published | Sign in with Trunc, or paste a key |
| WordPress plugin | `trunc-url-shortener/` | Download the .zip from `/integrations` ("Upload Plugin"), or WordPress.org once published | Connect with Trunc, or paste a key |
| Shopify app | `trunc-shopify-app/` | Shopify App Store only (needs your Partner app) | Sign in with Trunc, or paste a key |
| Canva app | `trunc-canva-app/` | Canva's app directory only (needs your Canva developer app) | Sign in with Trunc (Canva-managed OAuth) |

Each folder has a README with build and test steps.

## Sign in with Trunc

A small OAuth 2.0 authorization-code server in the Trunc backend: `internal/services/oauth_connect.go` and `internal/handlers/oauth_connect.go`.

1. The plugin sends the user to `https://trunc.site/connect` with:
   - `response_type=code`, `client_id` and `redirect_uri`;
   - `state` and `scope=links`;
   - `code_challenge` and `code_challenge_method=S256` (PKCE).
2. `/connect` (`frontend/src/app/connect/page.tsx`) makes the user sign in. It then names the app and the host it will return to, and asks them to **Allow** or **Cancel**.
3. **Allow** calls `POST /api/v1/oauth/authorize`. This works with a signed-in session only: an API key cannot approve new keys. The browser goes back to `redirect_uri?code=…&state=…`.
4. The plugin posts the code to `POST /api/v1/oauth/token`, with its `code_verifier` or its client secret. It receives `access_token`, which is a normal Trunc API key used as `Authorization: Bearer us_…`.

Details:

- **Codes** are 32 random bytes, kept 10 minutes in Redis and usable once (GETDEL). Each is bound to the user, client, redirect and PKCE challenge.
- **Keys** are created through `APIKeyService` with the default permissions (create, read, update and delete links; read analytics). They are named "Trunc for Chrome", "Trunc for WordPress" and so on, with the host and date in the description. They show on the user's **API keys** page, where they can be revoked.
- `POST /api/v1/oauth/revoke` (RFC 7009) deletes a key. The plugins call it on Disconnect.
- `GET /api/v1/oauth/client` checks a client and redirect for the consent page.
- **CORS** also allows `https://app-*.canva-apps.com` (the Canva app's iframe) and `chrome-extension://<id>` pages. Both send Bearer keys, not cookies.

### Clients

The clients are registered in code, so the redirect rules are reviewable.

| client_id | Type | Allowed `redirect_uri` |
|---|---|---|
| `trunc-chrome` | public, PKCE required | `https://<32 letters a–p>.chromiumapp.org/…` (`chrome.identity`) |
| `trunc-wordpress` | public, PKCE required | any https URL whose path contains `/wp-admin/`, and `http://localhost` for development |
| `trunc-shopify` | public, PKCE required | exactly the URL(s) in `SHOPIFY_CONNECT_REDIRECT_URI` (`<SHOPIFY_APP_URL>/auth/trunc/callback`); off while it is unset |
| `trunc-canva` | confidential: client secret, PKCE optional | exactly `https://www.canva.com/apps/oauth/authorized` |

For WordPress, every site has its own host, so any `…/wp-admin/…` URL is accepted. The consent page shows the host with "Only continue if this is your site". This is a deliberate trade-off: a phishing link could name an attacker's site, and the warning is the defence.

The Shopify app is one deployment, so its callback is pinned to the configured URL.

Redirects with user info, fragments or non-http(s) schemes are always refused. Plain http is only accepted for localhost.

## Configuration

| Variable | Where | Purpose |
|---|---|---|
| `CANVA_OAUTH_CLIENT_SECRET` | server runtime | The secret you also enter in the Canva Developer Portal. The `trunc-canva` client is disabled while it is empty. |
| `SHOPIFY_CONNECT_REDIRECT_URI` | server runtime | The exact callback of your Shopify app (`<SHOPIFY_APP_URL>/auth/trunc/callback`), comma separated if you run more than one. The `trunc-shopify` client is disabled while it is empty; pasted keys still work. |
| `CHROME_EXTENSION_URL`, `WORDPRESS_PLUGIN_URL`, `SHOPIFY_APP_URL`, `CANVA_APP_URL` | build args, passed as `NEXT_PUBLIC_*` | Store listings shown on `/integrations` and the dashboard. Leave them empty until published. Chrome and WordPress then offer the .zip, and Shopify and Canva say "coming". Only `https://` values are used. |

`docker-compose.oracle.yml` and `Dockerfile.dev` pass all of these through.

## Downloads

`scripts/build-plugins.sh` writes:
- `frontend/public/downloads/trunc-chrome-extension.zip`
- `frontend/public/downloads/trunc-wordpress-plugin.zip`

`Dockerfile.dev` runs it before the frontend build, so the site serves `/downloads/*.zip`. The zips are not committed.

## Publishing checklists

### Chrome Web Store
1. Register a developer account (one-time fee) and create an item. Upload `trunc-chrome-extension.zip`.
2. **Listing:** description and screenshots (popup, options page), 128 px icon (`icons/icon-128.png`), and the privacy policy URL `https://trunc.site/privacy`.
3. **Permission justifications:**
   - `activeTab`: the current page's URL when the user clicks.
   - `contextMenus`: the right-click item.
   - `identity`: Sign in with Trunc.
   - `storage`: the key and settings.
   - `offscreen` and `clipboardWrite`: copy the link.
   - Host access: `trunc.site` only.
4. **Data use:** the extension sends only the URL being shortened to Trunc. It does not sell data and does not use it for unrelated purposes.
5. After approval, set `CHROME_EXTENSION_URL` and rebuild. The extension ID changes from the unpacked one, and nothing on the server needs updating, because any `*.chromiumapp.org` ID is accepted.

### WordPress.org
1. Submit the zip at wordpress.org/plugins/developers/add. The slug requested is `trunc-url-shortener`.
2. Already in place:
   - `readme.txt` with "Tested up to: 7.1", External Services and Privacy sections;
   - GPL-2.0-or-later;
   - nonces and capability checks on every action, and sanitized and escaped output;
   - `uninstall.php`;
   - no build step.
   PHPCS (WordPress-Extra) reports formatting only.
3. After approval, SVN-commit to the plugin repository, then set `WORDPRESS_PLUGIN_URL` and rebuild.

### Shopify App Store
1. Create the app in your Partner dashboard, then `npm run config:link` and `npm run deploy` (pushes `shopify.app.toml`, including the compliance webhooks).
2. Set `SHOPIFY_CONNECT_REDIRECT_URI=<SHOPIFY_APP_URL>/auth/trunc/callback` on the Trunc server.
3. Host the app (the template's `Dockerfile`, or any Node 22 host), with:
   - `SHOPIFY_API_KEY`, `SHOPIFY_API_SECRET`, `SHOPIFY_APP_URL`, `SCOPES=read_products`;
   - `TRUNC_BASE_URL=https://trunc.site`.
4. **Listing:** app icon, screenshots, a demo screencast, and privacy policy and support URLs. The app uses only `read_products`, for the product picker.
5. **Review** checks the install flow, the three compliance webhooks (implemented in `app/routes/webhooks.*`), and that the embedded app loads in the admin.
6. After approval, set `SHOPIFY_APP_URL` (the App Store listing) for Trunc's build and rebuild.

### Canva
1. Create an app in the Canva Developer Portal. Under **Authentication**, fill in the OAuth settings from `trunc-canva-app/README.md` (client `trunc-canva`, scope `links`).
2. Set `CANVA_OAUTH_CLIENT_SECRET` on the Trunc server to the same secret.
3. `npm run build` and upload `dist/app.js`. Add the app's listing details, then preview it in the editor.
4. Submit for review.
5. After approval, set `CANVA_APP_URL` and rebuild.
