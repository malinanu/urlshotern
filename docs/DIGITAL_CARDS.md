# Digital visiting cards

Twelve editable digital visiting card templates, available in both card products:

| Product | Where | Renders | Needs |
|---|---|---|---|
| **MicroPages** ("Trunc vCard") | `/micropages`, public page at `/m/<slug>` | React, in the browser | nothing extra |
| **Microsites** (Pages) | `/microsites`, public page at `/m/<slug>` | Go templates, on the server | `ENABLE_MICROSITES=true` |

A card looks the same in both, because they share the markup contract, the 12 stylesheets, the
template catalog and the fonts.

## The templates

| # | Template | Collection | Style | Default call to action |
|---|---|---|---|---|
| 01 | Zenith | Premium & Minimal | Executive luxury | LET'S CONNECT |
| 02 | Nova | Premium & Minimal | Contemporary minimal | CONNECT WITH ME |
| 03 | Atlas | Premium & Minimal | Warm premium corporate | LET'S CONNECT |
| 04 | Axiom | Modern & Futuristic | High-tech executive | EXPLORE MY WORK |
| 05 | Vectra | Modern & Futuristic | Futuristic minimalism | CONNECT |
| 06 | Pulse | Modern & Futuristic | Dynamic digital identity | EXPLORE |
| 07 | Horizon | Elegant & Corporate | Professional corporate | CONNECT |
| 08 | Verdant | Elegant & Corporate | Premium natural | JOIN OUR MISSION |
| 09 | Skyway | Elegant & Corporate | International business | PLAN YOUR JOURNEY |
| 10 | Creo | Creative & Colourful | Creative premium | LET'S COLLABORATE |
| 11 | Bravo | Creative & Colourful | Bold business personality | LET'S GROW TOGETHER |
| 12 | Imagine | Creative & Colourful | Contemporary creative | LET'S CREATE |

Each template has three curated colour palettes and accepts one custom accent colour. Layout and
typography belong to the template and are not editable, so a card can't be customised into an
inconsistent one.

## What is editable

Everything on the card is real text and real elements, never an image:

| Card element | MicroPages | Microsites |
|---|---|---|
| Profile photo | Profile tab (upload or link) | Profile block, Avatar |
| Company logo | Digital Card tab (upload or link) | Profile block, Company logo |
| Name, job title, company | Profile / vCard tabs | Profile block |
| Company description | Digital Card tab (`config.card.about`, falls back to the bio) | Profile block, Bio |
| Phone, email, website, address | vCard tab | Save Contact block |
| WhatsApp | Digital Card tab (`config.card.whatsapp`) | Contact buttons block |
| Social links | Links & Social tab | Social icons block |
| Call to action text | Digital Card tab (and an optional link) | Save Contact block, Button text |
| QR destination | Digital Card tab (`config.card.qr_url`) | QR code block |
| Specialisation labels | Digital Card tab (up to 4) | Profile block, labels |
| Logo background | Digital Card tab | Profile block |
| Palette and accent | Digital Card tab | Design tab |

The QR code is generated from the card's real address (or the chosen destination) as a vector
image, level M error correction, dark modules on a light plate, with the 4-module quiet zone the
QR specification asks for. It is never an image of a QR code.

## Output

- **Live card** at `/m/<slug>`, with Save Contact (`.vcf`).
- **Shareable URL**, copy button on the card page and in the builder.
- **PNG and JPG download**, 1200 x 2100 px (the card is 600 x 1050 design pixels, exported at 2x).
  The export is rendered in the visitor's browser (`html-to-image`, vendored for Microsites) from a
  copy of the card mounted off-screen at exactly 600 px wide, so it looks the same whatever size the
  card was on screen.
- **QR code download**, 1024 x 1024 px PNG.

Images that are uploaded to Trunc always export correctly. An image linked from another site
exports only if that site allows cross-origin use; otherwise it is left out of the export.

## How it works

### One markup contract

Every template is the same markup, styled by its own stylesheet:

```html
<article class="vc vc-zenith" data-vc="zenith" data-nl="m" data-logo="none" style="--vc-accent:#...">
  <div class="vc-in">
    <i class="vc-d vc-d1"></i> ... <i class="vc-d vc-d4"></i>   <!-- decoration, styled per template -->
    <header class="vc-top"> logo, company name, digital ID </header>
    <figure class="vc-photo"> photo, or initials </figure>
    <div class="vc-who"> name, role, company </div>
    <p class="vc-about"> ... </p>   <ul class="vc-tags"> ... </ul>
    <ul class="vc-list"> phone, WhatsApp, email, web, address </ul>
    <ul class="vc-soc"> social icons </ul>
    <footer class="vc-foot"> QR code, "scan" hint, call to action </footer>
  </div>
</article>
```

- The card is a size container and everything inside is sized in `em`, where 1em is 1/60 of the
  card's width. The card scales with its container and keeps its 4:7 proportions on phones, tablets
  and desktops. It is never stretched.
- `data-nl` (s, m, l, xl) is the length of the name and steps the name's size down instead of
  clipping it. Long text is clamped with an ellipsis.
- `data-logo` (light, dark, none) is the background plate behind the company logo, so a white or
  dark logo stays readable.
- Palette values are CSS custom properties (`--vc-bg`, `--vc-accent`, ...), set inline.

### Shared files (frontend is the source)

| File | Purpose |
|---|---|
| `frontend/src/components/micropages/cards/css/*.css` | `base.css` plus one stylesheet per template |
| `frontend/src/components/micropages/cards/catalog.json` | Collections, names, descriptions, default CTA, fonts, palettes |
| `frontend/public/card-fonts/*.woff2` | Self-hosted OFL fonts (licences in `licenses/`) |

`scripts/sync-card-assets.sh` copies them to `internal/services/microsite_web/` for the Go renderer.
A Go test (`TestCardAssetsMatchFrontendSource`) fails if the copies drift, so run the script after
editing a stylesheet or the catalog.

Only the faces a template uses are fetched, and the Go renderer preloads at most two.

### MicroPages (React)

- `cards/CardRenderer.tsx` renders the markup for any template.
- `cards/cardData.ts` maps a micro-page to the card (`toCardData`). Template-specific settings are
  stored in the page's free-form `config.card` object, so no migration is needed:
  `logo_url`, `about`, `whatsapp`, `cta_text`, `cta_url`, `qr_url`, `tags`, `palette`, `accent`,
  `logo_plate`.
- `TemplateGallery.tsx` is the picker: four collections, thumbnails drawn by the real renderer,
  Preview (with palette switching) and Use Template.
- `cards/CardSettingsPanel.tsx` is the builder's Digital Card tab; `cards/CardActions.tsx` and
  `cards/export.ts` do the downloads.
- Photos and logos upload to `POST /api/v1/micropages/assets` (below).

### Microsites (Go)

- `internal/services/microsite_cards.go` builds the card from the page's blocks: the first
  Profile, Save Contact, Contact buttons, Social icons and QR code blocks. All other blocks render
  under the card, so nothing the owner added is lost.
- A card style is a `theme.style` value (`zenith`, `nova`, ...) with an optional `theme.palette`.
  The starter templates are `card_zenith` ... `card_imagine` in `GET /api/v1/microsites/templates`.
- `qr_code` is a new block type (`{"url": ""}`, empty means the page's own address). Profile blocks
  gained `tags` and `logo_plate`. Existing pages are unchanged.
- The QR code is generated by `github.com/skip2/go-qrcode` and drawn as inline SVG.
- Downloads run from `/m-static/card-export.js`, which loads `/m-static/html-to-image.js` only when
  a button is tapped. The builder preview runs no scripts.

## Image upload for MicroPages

`POST /api/v1/micropages/assets` (authenticated, multipart field `file`) stores a profile photo or
logo and returns `{"url": "/uploads/mp/<userID>/<sha256>.<jpg|png>"}`.

It reuses the same hardened pipeline as Microsites uploads: the real file type is sniffed (the
client's file name and type are ignored), only JPEG and PNG are accepted, dimensions are limited,
EXIF orientation is applied, and the image is re-encoded, which removes metadata and anything
appended to the file. Files are content-addressed, so an identical re-upload is a no-op. There is a
limit of 60 images per account and 2 MB per file (`MICROSITE_MAX_UPLOAD_BYTES`). Images are served
from `GET /uploads/mp/*path` with `nosniff` and a locked-down CSP, and are removed when the account
is deleted. It does not depend on `ENABLE_MICROSITES`; files use `MICROSITE_UPLOAD_PATH`, which
needs a persistent volume as for Microsites.

## Adding or changing a template

1. Edit or add `css/<id>.css` and its entry in `catalog.json` (frontend).
2. Add the id to `CardTemplateId` in `cards/types.ts` and the defaults in `cards/catalog.ts`.
3. Run `scripts/sync-card-assets.sh`, and add the id to `cardDefaults` / `cardSamples` in
   `internal/services/microsite_cards.go`.
4. Run `go test ./internal/services/` and the frontend type-check.

Keep em-based lengths out of `clip-path` on elements whose font size differs from the card's base
(use `cqw`), and don't put vertical padding on a line-clamped element: the clamped line shows in
the padding.
