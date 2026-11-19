=== Trunc URL Shortener ===
Contributors: trunc
Tags: url shortener, link management, analytics, short links
Requires at least: 6.0
Tested up to: 7.1
Requires PHP: 8.0
Stable tag: 1.1.0
License: GPL-2.0-or-later
License URI: https://www.gnu.org/licenses/gpl-2.0.html

Shorten URLs and track click analytics with Trunc — right inside your WordPress dashboard.

== Description ==

Connect your WordPress site to your Trunc account in one click with **Connect with Trunc** (or paste an API key), then:

* Shorten any URL from the admin dashboard
* View click counts synced from Trunc analytics
* Copy short links to clipboard in one click
* Insert short links using the `[trunc]` shortcode
* Use the **Trunc Short Link** Gutenberg block

**This plugin requires a Trunc account.** Sign up for free at [trunc.site](https://trunc.site). The free plan includes 10 active short links. Paid plans raise the limit and add custom domains and longer analytics history.

== External Services ==

This plugin connects to the Trunc URL shortening service (https://trunc.site) to perform the following actions:

* **Create short links** — When you shorten a URL, the original URL is sent to Trunc's API (`https://trunc.site/api/v1/shorten`) to generate a short code. The short link resolves through Trunc's servers.
* **Fetch analytics** — When you view click counts, the plugin queries Trunc's analytics API (`https://trunc.site/api/v1/analytics/`) to retrieve click data.
* **Fetch account/plan info** — The plugin calls (`https://trunc.site/api/v1/billing/usage`) to display your current plan and usage limits in the admin dashboard.
* **Connect with Trunc** — When you choose Connect with Trunc, your browser opens `https://trunc.site/connect` to approve the site, and the plugin exchanges the one-time code at `https://trunc.site/api/v1/oauth/token` for an API key. Disconnect revokes that key at `https://trunc.site/api/v1/oauth/revoke`.

These requests are made only when you connect the plugin and actively use it. No data is sent automatically or without your input. If you run your own Trunc server, the same requests go to the server you set under **Trunc Links → Settings → Server**.

Trunc's Terms of Service: https://trunc.site/terms
Trunc's Privacy Policy: https://trunc.site/privacy

== Installation ==

1. In WordPress, go to **Plugins → Add New → Upload Plugin** and upload `trunc-wordpress-plugin.zip` (from trunc.site/integrations), or upload the `trunc-url-shortener` folder to `/wp-content/plugins/`
2. Activate the plugin through the **Plugins** menu
3. Go to **Trunc Links → Settings** and choose **Connect with Trunc** (or paste an API key)
4. Start shortening URLs from the **Trunc Links** menu item

No build step is needed.

== Frequently Asked Questions ==

= Where do I get an API key? =

You usually don't need one: choose **Connect with Trunc** and approve the site. To use a key instead, sign up at [trunc.site](https://trunc.site) and create one on the **API keys** page (trunc.site/api-keys).

= Does this plugin work without a Trunc account? =

No. This plugin requires a Trunc account and API key to function. The plugin itself is free and open source (GPL-2.0+); Trunc is a separate third-party service with its own pricing.

= What plan do I need? =

The free Trunc plan (10 active short links, API access included) is enough to start. Premium (1,000 links, custom domains, a year of analytics), Growth and Enterprise (unlimited links) are listed at trunc.site/pricing.

= Does this work on multisite? =

Yes. Each site in the network uses its own API key setting.

= What shortcode syntax does this plugin use? =

`[trunc url="https://trunc.site/abc123" label="Click here"]`

Both `url` and `label` attributes are supported. If `label` is omitted, the short URL is used as the link text.

= Does this plugin store my links locally? =

Yes. Shortened links are also stored in your WordPress database (`wp_trunc_links`) so you can view and manage them from the admin panel without leaving WordPress.

= What PHP version is required? =

PHP 8.0 or higher is required. The plugin uses PHP 8.0 features (union types, `str_contains()`).

== Privacy ==

This plugin transmits data to Trunc's servers (https://trunc.site) as described in the **External Services** section above.

**What data is sent:**

* The original URLs you choose to shorten
* Your Trunc API key (used only for authentication in request headers)
* Standard HTTP request metadata (IP address, user agent) sent automatically by WordPress's HTTP API

**What data is NOT sent:**

* WordPress user data, post content, or any other site data
* Data from your site visitors
* Any data beyond what you explicitly submit when shortening a link

**Data storage:**

Short links (original URL, short code, short URL, click count) are stored locally in your WordPress database in the `wp_trunc_links` table. Deleting the plugin removes this table and the plugin's settings.

For full details on how Trunc handles data, see Trunc's Privacy Policy: https://trunc.site/privacy

== Screenshots ==

1. Admin link management page with shorten form and link table
2. Settings page for entering your Trunc API key and viewing plan info
3. Gutenberg block editor showing the Trunc Short Link block

== Changelog ==

= 1.1.0 =
* Connect with Trunc: connect the site without copying an API key (OAuth with PKCE); Disconnect revokes the key
* Server setting for self-hosted Trunc
* The block no longer needs a build step
* Clearer error messages from the Trunc API, corrected links and limit wording
* Uninstall removes the plugin's settings and table

= 1.0.0 =
* Initial release: admin link table, AJAX shortening, analytics sync, plan-aware UI, Gutenberg block, shortcode

== Upgrade Notice ==

= 1.1.0 =
Adds Connect with Trunc. Existing API keys keep working.

= 1.0.0 =
Initial release.
