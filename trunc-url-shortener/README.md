# Trunc URL Shortener — WordPress Plugin

Shorten URLs and track click analytics with [Trunc](https://trunc.site) directly from your WordPress dashboard.

## Features

- Shorten any URL from the WordPress admin panel
- View click counts synced from Trunc analytics
- Copy short links to clipboard in one click
- Insert short links with the `[trunc]` shortcode
- Add short links using the **Trunc Short Link** Gutenberg block
- Stores all links locally in your WordPress database

## Requirements

- WordPress 6.0+
- PHP 7.4+
- A free [Trunc](https://trunc.site) account and API key

## Installation

1. In WordPress, go to **Plugins → Add New → Upload Plugin** and upload `trunc-wordpress-plugin.zip` from trunc.site/integrations (built by `scripts/build-plugins.sh`), or copy this folder to `/wp-content/plugins/trunc-url-shortener/`.
2. Activate **Trunc URL Shortener**.
3. Go to **Trunc Links → Settings** and choose **Connect with Trunc**, or paste an API key from trunc.site/api-keys.

No build step: the block is plain JavaScript against the `wp.*` globals (`blocks/index.js` with a hand-written `index.asset.php`).

## Connect with Trunc

OAuth 2.0 authorization code with PKCE, client id `trunc-wordpress`.

1. **Connect with Trunc** posts to `admin-post.php?action=trunc_connect` (with a nonce and the `manage_options` capability). The plugin stores the PKCE verifier in a 10-minute transient keyed by `state`, then sends the admin to `<server>/connect`.
2. Trunc sends the admin back to `wp-admin/admin.php?page=trunc-settings&code=…&state=…`.
3. `Trunc_Connect::maybe_finish()` checks the state belongs to the same admin, then exchanges the code server side at `/api/v1/oauth/token`.
4. The key is stored in the `trunc_api_key` option.

**Disconnect** revokes a key that came from Connect with Trunc. The Trunc server only accepts redirects that contain `/wp-admin/` for this client, and requires https (or localhost).

## Files

- `trunc-url-shortener.php`: bootstrap, `trunc_server()`, and the activation hook (creates `wp_trunc_links`).
- `includes/class-trunc-connect.php`: Connect with Trunc and Disconnect.
- `includes/class-trunc-admin.php`: the admin pages, settings and AJAX endpoints.
- `includes/class-trunc-api.php`: the Trunc REST API client.
- `includes/class-trunc-block.php`: the block and the `[trunc url="…" label="…"]` shortcode.
- `uninstall.php`: removes the options and the table.
