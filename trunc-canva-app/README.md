# Trunc for Canva

A Canva app. In the Canva editor, people make a short link with [Trunc](https://trunc.site), then add its QR code (and the link as text) to their design. The scans and clicks show up in their Trunc dashboard.

Built from Canva's [Apps SDK starter kit](https://github.com/canva-sdks/canva-apps-sdk-starter-kit). Canva's licence for the kit is in `LICENSE.md`.

## How it works

- **Sign in with Trunc** uses Canva's OAuth support (`auth.initOauth()` from `@canva/user`). Canva runs the authorization-code flow against Trunc as the confidential client `trunc-canva`, and keeps and refreshes the token. The app never sees a secret, and never asks for a pasted key: Canva apps can't keep one between sessions.
- **Create short link** calls `POST <Trunc>/api/v1/shorten` from the app's iframe with that token. Trunc allows CORS from `https://app-*.canva-apps.com`.
- **Add QR code to design** makes a 1024 px PNG in the browser (`qrcode`), uploads it with `@canva/asset`, and adds it as an image with alt text.
- **Add link as text** adds the short link as a text element.
- All visible text goes through `react-intl`, so the app is ready for translation (`dist/messages_en.json`).

## Set up in the Canva Developer Portal

1. Create an app and copy its ID into `.env` (`CANVA_APP_ID`).
2. **Authentication → OAuth**:

   | Setting | Value |
   |---|---|
   | Authorization server URL | `https://trunc.site/connect` |
   | Token exchange URL | `https://trunc.site/api/v1/oauth/token` |
   | Revocation URL | `https://trunc.site/api/v1/oauth/revoke` |
   | Client ID | `trunc-canva` |
   | Client secret | the value of `CANVA_OAUTH_CLIENT_SECRET` on the Trunc server |
   | Scopes | `links` |
   | Redirect | Canva's fixed `https://www.canva.com/apps/oauth/authorized` (the only one Trunc accepts for this client) |

3. Set `CANVA_BACKEND_HOST` in `.env` to the Trunc server (`https://trunc.site`). The app has no backend of its own.

Trunc keeps the `trunc-canva` client disabled until `CANVA_OAUTH_CLIENT_SECRET` is set on the server.

## Develop

```bash
npm install        # also copies .env.template to .env
npm start          # dev server; preview the app in Canva from the Developer Portal
npm test           # jest, with Canva's SDKs mocked
npm run lint:types
npm run lint
npm run build      # dist/app.js, uploaded in the Developer Portal
```

## Publishing

See `docs/INTEGRATIONS.md` in the Trunc repository for the Canva review checklist.
