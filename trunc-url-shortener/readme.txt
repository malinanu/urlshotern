=== Trunc URL Shortener ===
Contributors: trunc
Tags: url shortener, link management, analytics, short links
Requires at least: 6.0
Tested up to: 6.7
Requires PHP: 8.0
Stable tag: 1.0.0
License: GPL-2.0+
License URI: https://www.gnu.org/licenses/gpl-2.0.html

Shorten URLs and track click analytics with Trunc — right inside your WordPress dashboard.

== Description ==

Connect your WordPress site to your Trunc account to:

* Shorten any URL from the admin dashboard
* View click counts synced from Trunc analytics
* Copy short links to clipboard in one click
* Insert short links using the `[trunc]` shortcode
* Use the **Trunc Short Link** Gutenberg block

**This plugin requires a Trunc account.** Sign up for free at [trunc.site](https://trunc.site). The free plan includes 10 short links per month. Premium plans unlock more links, custom domains, and extended analytics.

== External Services ==

This plugin connects to the Trunc URL shortening service (https://trunc.site) to perform the following actions:

* **Create short links** — When you shorten a URL, the original URL is sent to Trunc's API (`https://trunc.site/api/v1/shorten`) to generate a short code. The short link resolves through Trunc's servers.
* **Fetch analytics** — When you view click counts, the plugin queries Trunc's analytics API (`https://trunc.site/api/v1/analytics/`) to retrieve click data.
* **Fetch account/plan info** — The plugin calls (`https://trunc.site/api/v1/billing/usage`) to display your current plan and usage limits in the admin dashboard.

These requests are made only when you have configured an API key and actively use the plugin. No data is sent automatically or without your input.

Trunc's Terms of Service: https://trunc.site/legal/terms
Trunc's Privacy Policy: https://trunc.site/legal/privacy

== Installation ==

1. Upload the `trunc-url-shortener` folder to `/wp-content/plugins/`
2. Activate the plugin through the **Plugins** menu in WordPress
3. Build the Gutenberg block assets: `npm install && npm run build` (requires Node.js)
4. Go to **Trunc Links → Settings** and paste your Trunc API key
5. Start shortening URLs from the **Trunc Links** menu item

Alternatively, download a pre-built release from the plugin's GitHub repository which includes compiled block assets.

== Frequently Asked Questions ==

= Where do I get an API key? =

Sign up for free at [trunc.site](https://trunc.site) and go to **Dashboard → API Keys**. The free account gives you 10 short links per month at no cost.

= Does this plugin work without a Trunc account? =

No. This plugin requires a Trunc account and API key to function. The plugin itself is free and open source (GPL-2.0+); Trunc is a separate third-party service with its own pricing.

= What plan do I need? =

The free Trunc plan (10 links/month, API access included) is sufficient for basic use. Premium (1,000 links/month, custom domains, 365-day analytics) and Enterprise (unlimited) plans are available at trunc.site/pricing.

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
* Your Trunc API key (used only for authentication in request headers, never stored on Trunc's servers beyond authentication)
* Standard HTTP request metadata (IP address, user agent) sent automatically by WordPress's HTTP API

**What data is NOT sent:**

* WordPress user data, post content, or any other site data
* Data from your site visitors
* Any data beyond what you explicitly submit when shortening a link

**Data storage:**

Short links (original URL, short code, short URL, click count) are stored locally in your WordPress database in the `wp_trunc_links` table. This data is retained until you delete the plugin's database table.

For full details on how Trunc handles data, see Trunc's Privacy Policy: https://trunc.site/legal/privacy

== Screenshots ==

1. Admin link management page with shorten form and link table
2. Settings page for entering your Trunc API key and viewing plan info
3. Gutenberg block editor showing the Trunc Short Link block

== Changelog ==

= 1.0.0 =
* Initial release: admin link table, AJAX shortening, analytics sync, plan-aware UI, Gutenberg block, shortcode

== Upgrade Notice ==

= 1.0.0 =
Initial release.
