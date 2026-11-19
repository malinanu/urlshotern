# Trunc for Chrome

Shorten the page you are on with [Trunc](https://trunc.site): one click on the toolbar button (or `Alt+Shift+S`), or right-click a page or link and choose **Shorten with Trunc**. The link is copied straight away; the popup also shows its QR code (downloadable as a PNG) and your five most recent links with their clicks.

## Install

- **From a file:** download `trunc-chrome-extension.zip` from trunc.site/integrations, unzip it, open `chrome://extensions`, turn on **Developer mode**, choose **Load unpacked** and select the folder.
- **From the Chrome Web Store:** once published (see `docs/INTEGRATIONS.md`).

Then open the extension and choose **Sign in with Trunc**, or paste an API key from trunc.site/api-keys in **Settings**.

## How it works

- No build step: plain Manifest V3, ES modules. `vendor/qrcode.mjs` is qrcode-generator 2.0.4 (MIT).
- **Sign in with Trunc** uses `chrome.identity.launchWebAuthFlow` with PKCE (`client_id=trunc-chrome`). The redirect is `https://<extension-id>.chromiumapp.org/trunc`, and the server only accepts that host pattern for this client. The flow runs in the background worker, because the popup closes when the sign-in window opens.
- The key is stored in `chrome.storage.local`. **Disconnect** revokes a key obtained through sign-in (`/api/v1/oauth/revoke`) and forgets it.
- Short links are cached in `chrome.storage.session` per URL, so reopening the popup on the same page doesn't create another link.
- Permissions:
  - `activeTab`: read the current tab's URL when you click the button.
  - `contextMenus`: the right-click item.
  - `identity`: sign-in.
  - `storage`: your key and settings.
  - `offscreen` + `clipboardWrite`: copy from the right-click menu.
  - Host access is `https://trunc.site/*` only. A self-hosted server is requested at runtime from Settings.
