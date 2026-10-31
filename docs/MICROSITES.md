# Microsites — micro-landing pages & digital vCards

User-built mobile pages served at `/m/<slug>`: a digital business card with a
one-tap **Save Contact** (`.vcf`) button, link-in-bio, product showcases and
event pages, with video and audio embeds, newsletter and contact forms, and
first-party analytics. Built in the dashboard at `/microsites` (labelled **Pages**),
with **Build**, **Design**, **Track** and **Inbox** tabs.

## Enabling

The feature is **off by default**. With `ENABLE_MICROSITES=false` none of its
routes are registered and the app behaves exactly as before.

1. Apply migrations `configs/migrations/046_microsites_up.sql` (creates
   `microsites`, `microsite_assets`, `microsite_events`) and
   `047_microsite_submissions_up.sql` (creates `microsite_submissions` for
   forms). No existing tables are changed.
2. Set `ENABLE_MICROSITES=true`. For Docker builds the same variable is passed
   to the frontend build as `NEXT_PUBLIC_ENABLE_MICROSITES`, which shows the
   **Pages** link in the header; rebuild the image after changing it.
3. Mount a persistent volume at `MICROSITE_UPLOAD_PATH` (the compose files use
   `/app/uploads`). Optional: `MICROSITE_MAX_UPLOAD_BYTES` (default 2 MB).

If the flag is on but a migration is missing, the server logs a warning and
the affected endpoints return `503` instead of failing.

The app sends each page's Content-Security-Policy itself. If you put the repo's
nginx in front (`docker/nginx.prod.conf` or `docker/nginx.conf`), both have a
`location ^~ /m/` block that keeps their site-wide policy off pages and repeats
the other security headers. Without it, that site-wide policy would block the
embedded players.

### Behind CloudPanel and Cloudflare

The production path is Cloudflare → the CloudPanel reverse-proxy vhost → the
container (`docker-compose.oracle.yml`, port 8082). Neither of the repo's nginx
files is involved, and **nothing needs to change** in CloudPanel or Cloudflare:

- CloudPanel's vhost adds X-Frame-Options, nosniff, X-XSS-Protection,
  X-Permitted-Cross-Domain-Policies, `Referrer-Policy: same-origin` and
  `Cache-Control: no-transform`, but **no Content-Security-Policy**. Pages
  therefore run under the app's own policy, which lets the players and map load.
  Each player iframe sets its own `referrerpolicy`, so YouTube still receives
  the origin it requires.
- Form rate limits and visitor counts use Cloudflare's `CF-Connecting-IP`, and
  the visitor's country comes from `CF-IPCountry`.
- Cloudflare doesn't cache the pages (they send `private, no-cache`), and
  `no-transform` keeps Rocket Loader and email obfuscation off. Both would break
  pages, which allow no inline script.

If you ever add a Content-Security-Policy in the CloudPanel vhost, exempt `/m/`
so the app's policy stays the only one. Add this to the vhost (Sites → your
site → Vhost), above its `location /` block:

```nginx
location ^~ /m/ {
    proxy_pass http://127.0.0.1:8082;
    proxy_http_version 1.1;
    proxy_set_header Host $host;
    proxy_set_header X-Real-IP $remote_addr;
    proxy_set_header X-Forwarded-For $proxy_add_x_forwarded_for;
    proxy_set_header X-Forwarded-Proto $scheme;
    # no add_header Content-Security-Policy here
}
```

`docker/nginx.prod.conf` and `docker/nginx.conf` already have an equivalent
`/m/` location for deployments that use them.

## Plans

| Plan (DB name) | Pages | Forms (newsletter / contact) | Remove “Made with Trunc” badge |
|---|---|---|---|
| Free (`free`) | 1 | — | — |
| Pro (`premium`) | 10 | ✓ | — |
| Ultra (`enterprise`) | unlimited | ✓ | ✓ |

Limits live in `PlanLimits.MaxPages` (`internal/services/billing.go`). Forms are
the `microsite_forms` feature and badge removal the `remove_branding` feature,
both in `internal/middleware/plan_gating.go`. All are enforced server-side.
A downgrade never unpublishes or deletes pages: it blocks creating new pages,
the badge reappears, and form blocks are hidden while the submit endpoint
refuses posts. Earlier submissions stay readable and exportable.

## Templates and blocks

| Template | Style | What it is |
|---|---|---|
| Digital Business Card | `letterpress` | A card with painted edges; Save Contact is the card's foot |
| Link-in-Bio | `board` | A departure board; the highlighted link is “now boarding” |
| Product Showcase | `flyer` | A weekly flyer: big prices, deal bursts, buy buttons |
| Event | `programme` | A pocket programme: date, venue, countdown, timetable, tickets |
| Blank | `classic` | The original centered card (all pre-existing pages use it) |

The style is `theme.style`; a missing value means `classic`, which renders
byte-for-byte as before (guarded by golden files in
`internal/services/testdata/classic/`). Each style is one stylesheet in
`internal/services/microsite_web/styles/` plus one self-hosted OFL font
(`microsite_web/fonts/`, served from `/m-static/fonts/`). The design rules are in
`DESIGN.md`, product context in `PRODUCT.md`, and each style's direction in
`.impeccable/surfaces/`.

Blocks: profile, Save Contact, contact buttons, social icons, links, text,
image, divider, **product card**, **product carousel**, **event details**
(with `.ics` “Add to calendar”), **countdown**, **agenda**, **map & directions**
(tap-to-load OpenStreetMap when coordinates are set), **video** (YouTube, Vimeo),
**audio** (Spotify), and **form** (newsletter or contact).

## Routes

Public (no auth):

| Route | Purpose |
|---|---|
| `GET /m/:slug` | Published page (server-rendered HTML; records a view) |
| `GET /m/:slug/vcard?b=<blockId>` | vCard 3.0 download (records a save) |
| `GET /m/:slug/ics?b=<blockId>` | iCalendar file for an event block (records a calendar add) |
| `POST /m/:slug/submit` | Form post; redirects back with `?sent=` or `?form_error=` (rate-limited) |
| `POST /api/v1/m/:slug/event` | Click beacon for links, contact buttons and plays (rate-limited) |
| `GET /m-static/:file` | `beacon.js`, `embed.js`, `countdown.js` (each < 1 KB, loaded only when needed) |
| `GET /m-static/fonts/:file` | Self-hosted template fonts (immutable cache) |
| `GET /uploads/m/:user/:hash.(jpg\|png)` | Uploaded images |

Authenticated (`/api/v1/microsites`, JWT or API key):

| Route | Purpose |
|---|---|
| `GET /` · `POST /` | List (with plan usage) · create from a template |
| `GET /templates` · `GET /templates/:template/preview` · `GET /slug-available?slug=` | Templates · sample render for the create dialog · slug check |
| `POST /preview` | Render unsaved content for the builder preview |
| `GET /:id` · `PUT /:id` · `DELETE /:id` | Read · save draft (needs `version`; 409 on conflict) · soft delete |
| `POST /:id/publish` · `POST /:id/unpublish` | Publish draft snapshot · take offline |
| `GET /:id/analytics?days=` | Track-tab report |
| `POST /:id/assets` · `GET /assets` · `DELETE /assets/:assetId` | Image upload (multipart `file`) · list · delete |
| `POST /:id/short-link` | Create (once) a short link pointing at the page |
| `GET /:id/submissions?limit=&offset=` · `GET /:id/submissions/export` · `DELETE /:id/submissions/:sid` | Inbox list · CSV export · delete one |

`/:id` routes require the caller to own the page (`RequireMicrositeOwnership`).

## How it fits the existing system

- **Separate from short links.** Page telemetry goes to `microsite_events`;
  `click_events`, the four aggregate analytics tables and `RecordClick` are
  untouched. Only stateless helpers (bot detection, UA/referrer parsing) are reused.
- **Rendered by Go, not Next.js.** The frontend is a static export, so per-page
  `<title>`/OG tags can't come from it. `internal/services/microsite_render.go`
  renders pages with `html/template`; the builder preview uses the same renderer
  inside a sandboxed `srcdoc` iframe.
- **Drafts vs published.** Edits save to `draft`; publishing validates strictly
  and copies a snapshot to `published`, which is what `/m/<slug>` serves.

## Safety

- Every block, URL and theme value is validated on save, publish and render
  (`internal/services/microsite_blocks.go`). Links allow only `https`/`http`,
  `mailto:`, `tel:`, `sms:`; there is no raw-HTML block; themes are structured
  data and CSS is generated server-side.
- Uploads: size cap, content sniffing (client type and filename ignored),
  JPEG/PNG only (no SVG), dimension limit, EXIF orientation applied, then
  decoded and re-encoded (metadata and appended payloads dropped), stored under
  a content hash.
- No raw IPs are stored for page analytics; visitors are counted with a keyed,
  daily-rotating HMAC. Country comes from Cloudflare's `CF-IPCountry` header.
- Page events and form submissions follow the click retention windows (free 30
  days, premium 1 year).
- Pages send their own Content-Security-Policy: no inline script, scripts only
  from this site, and frames only from `www.youtube-nocookie.com`,
  `player.vimeo.com`, `open.spotify.com` and `www.openstreetmap.org`.
- Embeds are click-to-load facades. The server parses a strict video, track or
  coordinate ID and builds the player URL itself; owners never supply an iframe
  `src`. Nothing third-party loads until the visitor taps.
- Forms work without JavaScript and carry a hidden honeypot and an HMAC-signed
  render-time token (rejected if sent within 3 s or after 24 h). Limits: 5 posts
  per visitor per 10 minutes and 500 per page per day. Bodies are capped at 8 KB.
  Emails are validated, text is length-capped with control characters stripped,
  and the CSV export prefixes cells starting with `= + - @` so spreadsheets don't
  run them as formulas. Owner alerts are escaped and throttled to one email a
  minute per form (later ones are counted into the next alert).

## Known limits / next steps

- Uploads use local disk (`LocalAssetStore`); run one backend instance or add an
  object-storage `AssetStore` before scaling out (the k8s manifests use an HPA).
- Custom domains are not supported for pages yet (needs real SSL provisioning
  and a catch-all host config).
- Changing a slug breaks printed QR codes; there is no slug history yet.
- Not yet: team-owned pages, form fields beyond newsletter/contact, and
  per-language template copy. Template fonts cover Latin scripts; other scripts
  use the visitor's system fonts.
