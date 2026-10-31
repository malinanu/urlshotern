# Microsites — micro-landing pages & digital vCards

User-built mobile pages served at `/m/<slug>`: a digital business card with a
one-tap **Save Contact** (`.vcf`) button, link-in-bio blocks, and first-party
analytics. Built in the dashboard at `/microsites` (labelled **Pages**).

## Enabling

The feature is **off by default**. With `ENABLE_MICROSITES=false` none of its
routes are registered and the app behaves exactly as before.

1. Apply migration `configs/migrations/046_microsites_up.sql` (creates
   `microsites`, `microsite_assets`, `microsite_events`; no existing tables
   are changed).
2. Set `ENABLE_MICROSITES=true`. For Docker builds the same variable is passed
   to the frontend build as `NEXT_PUBLIC_ENABLE_MICROSITES`, which shows the
   **Pages** link in the header; rebuild the image after changing it.
3. Mount a persistent volume at `MICROSITE_UPLOAD_PATH` (the compose files use
   `/app/uploads`). Optional: `MICROSITE_MAX_UPLOAD_BYTES` (default 2 MB).

If the flag is on but the migration is missing, the server logs a warning and
the endpoints return `503` instead of failing.

## Plans

| Plan (DB name) | Pages | Remove “Made with Trunc” badge |
|---|---|---|
| Free (`free`) | 1 | — |
| Pro (`premium`) | 10 | — |
| Ultra (`enterprise`) | unlimited | ✓ |

Limits live in `PlanLimits.MaxPages` (`internal/services/billing.go`); badge
removal is the `remove_branding` feature in `internal/middleware/plan_gating.go`.
Both are enforced server-side. A downgrade never unpublishes or deletes pages;
it only blocks creating new ones (and the badge reappears).

## Routes

Public (no auth):

| Route | Purpose |
|---|---|
| `GET /m/:slug` | Published page (server-rendered HTML; records a view) |
| `GET /m/:slug/vcard?b=<blockId>` | vCard 3.0 download (records a save) |
| `POST /api/v1/m/:slug/event` | Click beacon for links/contact buttons (rate-limited) |
| `GET /m-static/beacon.js` | The beacon script (< 1 KB) |
| `GET /uploads/m/:user/:hash.(jpg\|png)` | Uploaded images |

Authenticated (`/api/v1/microsites`, JWT or API key):

| Route | Purpose |
|---|---|
| `GET /` · `POST /` | List (with plan usage) · create from a template |
| `GET /templates` · `GET /slug-available?slug=` | Templates · slug check |
| `POST /preview` | Render unsaved content for the builder preview |
| `GET /:id` · `PUT /:id` · `DELETE /:id` | Read · save draft (needs `version`; 409 on conflict) · soft delete |
| `POST /:id/publish` · `POST /:id/unpublish` | Publish draft snapshot · take offline |
| `GET /:id/analytics?days=` | Track-tab report |
| `POST /:id/assets` · `GET /assets` · `DELETE /assets/:assetId` | Image upload (multipart `file`) · list · delete |
| `POST /:id/short-link` | Create (once) a short link pointing at the page |

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
- Page events follow the click retention windows (free 30 days, premium 1 year).

## Known limits / next steps

- Uploads use local disk (`LocalAssetStore`); run one backend instance or add an
  object-storage `AssetStore` before scaling out (the k8s manifests use an HPA).
- Custom domains are not supported for pages yet (needs real SSL provisioning
  and a catch-all host config).
- Changing a slug breaks printed QR codes; there is no slug history yet.
- Not in v1: product/event templates, video embeds (would need a CSP change for
  `/m/`), forms, team-owned pages.
