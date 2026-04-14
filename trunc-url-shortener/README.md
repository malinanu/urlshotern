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

1. Download or clone this repository
2. Build the Gutenberg block:
   ```bash
   npm install
   npm run build
   ```
3. Copy the `trunc-url-shortener` folder to your WordPress plugins directory:
   ```
   /wp-content/plugins/trunc-url-shortener/
   ```
4. Go to **WordPress Admin → Plugins** and activate **Trunc URL Shortener**
5. Go to **Trunc Links → Settings** and paste your Trunc API key

Get your API key at [trunc.site/dashboard/api-keys](https://trunc.site/dashboard/api-keys).

## Usage

### Admin Dashboard

Go to **Trunc Links** in the WordPress sidebar to:
- Shorten a URL by entering it in the form at the top
- Optionally set a custom short code (e.g. `my-product`)
- View all your short links with click counts
- Copy any short link to clipboard
- Open the full Trunc analytics page for any link

### Shortcode

Insert a short link anywhere in a post or page:

```
[trunc url="https://trunc.site/abc123" label="Click here"]
```

| Attribute | Required | Description |
|-----------|----------|-------------|
| `url` | Yes | The Trunc short URL |
| `label` | No | Link text (defaults to the URL if omitted) |

### Gutenberg Block

1. In the block editor, add a new block and search for **Trunc Short Link**
2. Enter the short URL and label in the sidebar panel
3. The link renders on the frontend via server-side render

## File Structure

```
trunc-url-shortener/
├── trunc-url-shortener.php       # Plugin entry point, activation hook
├── includes/
│   ├── class-trunc-api.php       # Trunc API client (shorten + analytics)
│   ├── class-trunc-admin.php     # Admin menu, settings, AJAX handlers
│   └── class-trunc-block.php     # Gutenberg block + shortcode
├── admin/
│   ├── admin.css                 # Admin page styles
│   └── admin.js                  # Admin AJAX (shorten form, copy button)
├── blocks/
│   ├── block.json                # Block manifest
│   └── index.js                  # Block editor UI (JSX)
├── package.json                  # npm build config (@wordpress/scripts)
└── readme.txt                    # WordPress.org plugin readme
```

## API

The plugin communicates with the Trunc API at `https://trunc.site/api/v1`.

| Action | Endpoint | Method |
|--------|----------|--------|
| Shorten URL | `/shorten` | POST |
| Get analytics | `/analytics/{shortCode}` | GET |

Authentication uses your API key as a Bearer token:
```
Authorization: Bearer us_your_api_key_here
```

## Development

```bash
# Install dependencies
npm install

# Build Gutenberg block (production)
npm run build

# Watch mode during development
npm start
```

## Distribution

To create a distributable zip (excludes node_modules and dev files):

```bash
cd /path/to/plugins
zip -r trunc-url-shortener-v1.0.0.zip trunc-url-shortener/ \
  --exclude "*/node_modules/*" --exclude "*/.git/*"
```

Upload the zip via **WordPress Admin → Plugins → Add New → Upload Plugin**.
