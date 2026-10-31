---
name: Trunc Pages templates
description: Four owner-branded page worlds (card, board, flyer, programme) plus the classic card, rendered server-side from validated tokens.
colors:
  press-blue: "#2340d6"
  press-ground: "#e5e8fa"
  card-stock: "#ffffff"
  ink: "#14161c"
  station-blue: "#1b2a6b"
  boarding-yellow: "#ffd400"
  board-white: "#ffffff"
  deal-red: "#d7261e"
  deal-red-deep: "#961b15"
  flyer-ink: "#17181c"
  track-green: "#0e7c66"
  programme-ink: "#111318"
  paper: "#ffffff"
  # Derived from the defaults by the renderer (accent-12/-30/-deep, track tones,
  # hairlines and muted text); listed so pages can be checked against them.
  press-blue-30: "#bdc6f3"
  press-badge: "rgb(25, 45, 150)"
  card-hairline: "rgb(220, 220, 221)"
  boarding-yellow-30: "#fff2b3"
  boarding-yellow-deep: "#b39400"
  board-hairline: "rgb(61, 74, 129)"
  board-muted: "rgb(182, 187, 208)"
  deal-red-30: "#f3bebc"
  deal-red-track: "#6d1e1d"
  track-green-30: "#b7d8d1"
  track-green-track: "#10423b"
  programme-hairline: "rgb(219, 220, 220)"
  alert-ink: "#8a1c1c"
  alert-wash: "#fde8e8"
  shadow-ink: "#000000"
typography:
  card-name:
    fontFamily: "Schibsted Grotesk, system-ui, -apple-system, Segoe UI, Roboto, Noto Sans, sans-serif"
    fontSize: "clamp(34px, 10vw, 44px)"
    fontWeight: 800
    lineHeight: 1.02
    letterSpacing: "-0.025em"
  card-label:
    fontFamily: "Schibsted Grotesk, system-ui, sans-serif"
    fontSize: "14px"
    fontWeight: 700
    letterSpacing: "0.14em"
  board-plate:
    fontFamily: "Martian Mono, ui-monospace, Menlo, Consolas, Noto Sans Mono, monospace"
    fontSize: "clamp(20px, 6vw, 26px)"
    fontWeight: 700
    lineHeight: 1.1
    letterSpacing: "0.02em"
  board-row:
    fontFamily: "Martian Mono, ui-monospace, monospace"
    fontSize: "15px"
    fontWeight: 600
    letterSpacing: "0.02em"
  board-boarding:
    fontFamily: "Martian Mono, ui-monospace, monospace"
    fontSize: "19px"
    fontWeight: 600
  flyer-masthead:
    fontFamily: "Barlow Condensed, Arial Narrow, system-ui, Noto Sans, sans-serif"
    fontSize: "clamp(40px, 13vw, 64px)"
    fontWeight: 800
    lineHeight: 0.9
    letterSpacing: "-0.01em"
  flyer-price-hero:
    fontFamily: "Barlow Condensed, Arial Narrow, sans-serif"
    fontSize: "76px"
    fontWeight: 800
    lineHeight: 0.85
    letterSpacing: "-0.02em"
  flyer-price:
    fontFamily: "Barlow Condensed, Arial Narrow, sans-serif"
    fontSize: "46px"
    fontWeight: 800
    lineHeight: 0.85
  programme-title:
    fontFamily: "Hanken Grotesk, system-ui, -apple-system, Segoe UI, Roboto, Noto Sans, sans-serif"
    fontSize: "clamp(34px, 10vw, 52px)"
    fontWeight: 800
    lineHeight: 0.98
    letterSpacing: "-0.03em"
  programme-section:
    fontFamily: "Hanken Grotesk, system-ui, sans-serif"
    fontSize: "22px"
    fontWeight: 800
    letterSpacing: "-0.01em"
  programme-body:
    fontFamily: "Hanken Grotesk, system-ui, sans-serif"
    fontSize: "15px"
    fontWeight: 400
  programme-countdown:
    fontFamily: "Hanken Grotesk, system-ui, sans-serif"
    fontSize: "clamp(40px, 12vw, 56px)"
    fontWeight: 800
    lineHeight: 1
  # The shared ramp every style draws its remaining sizes from (in the style's face).
  micro:
    fontSize: "11px"
  caption:
    fontSize: "12px"
  meta:
    fontSize: "13px"
  body:
    fontSize: "16px"
  body-strong:
    fontSize: "17px"
  lead:
    fontSize: "18px"
  heading-sm:
    fontSize: "24px"
  heading:
    fontSize: "28px"
  heading-md:
    fontSize: "30px"
  heading-lg:
    fontSize: "32px"
  display-sm:
    fontSize: "36px"
  display-md:
    fontSize: "38px"
rounded:
  card: "12px"
  bar: "0px"
  board: "3px"
  flyer: "2px"
  programme: "4px"
  chip: "6px"
  pill: "999px"
spacing:
  page-gutter: "18px"
  card-pad: "28px 22px 24px"
  row: "14px"
components:
  save-contact-bar:
    backgroundColor: "{colors.press-blue}"
    textColor: "{colors.card-stock}"
    typography: "{typography.card-label}"
    rounded: "{rounded.bar}"
    height: "66px"
  board-row-boarding:
    backgroundColor: "{colors.boarding-yellow}"
    textColor: "{colors.station-blue}"
    typography: "{typography.board-boarding}"
    rounded: "{rounded.bar}"
    padding: "18px"
  board-row:
    backgroundColor: "{colors.station-blue}"
    textColor: "{colors.board-white}"
    typography: "{typography.board-row}"
    padding: "13px 18px"
  flyer-buy:
    backgroundColor: "{colors.deal-red}"
    textColor: "{colors.card-stock}"
    rounded: "{rounded.flyer}"
  flyer-burst:
    backgroundColor: "{colors.deal-red-deep}"
    textColor: "{colors.card-stock}"
    size: "76px"
  programme-tickets:
    backgroundColor: "{colors.track-green}"
    textColor: "{colors.paper}"
    rounded: "{rounded.programme}"
    padding: "12px 10px"
  programme-calendar:
    backgroundColor: "{colors.paper}"
    textColor: "{colors.programme-ink}"
    rounded: "{rounded.programme}"
    padding: "12px 10px"
---

# Design System: Trunc Pages templates

## Overview

Trunc Pages are one-link pages that a business card holder, a creator, a small shop or an event organiser shares from a QR code or a bio. The visitor is on a phone, often on a slow connection, and has one job: save the contact, catch the newest link, see the price, or know when and where. The owner's brand leads every page; Trunc appears only in the removable "Made with Trunc" badge.

The system is not one look in four colors. Each template is its own world, chosen against the category default it replaces:

- **Letterpress card** (Digital Business Card). One thick card held in the hand, not a profile screen. A white card stock with 6px edges painted in the owner's accent on a ground of the accent at 12%. The name is debossed. The card's foot is a painted bar, **Save Contact**. Contact rows and socials sit on the card's back.
- **Departure board** (Link-in-Bio). Links are departures, and the newest one is boarding now. A station board in deep blue, monospace in fixed columns: destination, the host it goes *via*, and a status chip. The highlighted link is the yellow **NOW BOARDING** row, revealed once with a split-flap step animation.
- **Weekly flyer** (Product Showcase). The page sells like a printed weekly flyer: a saturated masthead band, condensed caps, huge price numerals with superscript cents, deal bursts, a dense two-column grid and a "this week" carousel.
- **Conference programme** (Event). The printed programme you'd want in your pocket on the day: a ruled cover (title, date, time, venue, countdown), **Get tickets** and **Add to calendar** together, then a timetable with track bars, then a framed map.
- **Classic** (Blank and every page created before templates had styles). The original centered card on the owner's background. It renders byte-for-byte as it always has.

Everything is rendered server-side through `html/template` from validated tokens; owners never supply CSS or HTML. Each style is one stylesheet in `internal/services/microsite_web/styles/`, loaded only by pages using it, together with the one self-hosted face it needs.

## Colors

The frontmatter lists each template's **defaults**. On a live page, four owner tokens drive every color: `accent`, `card_fill`, `text` and `button_text`. The renderer derives the rest from them (all validated hex): `--accent-12` (accent at 12% over white, the letterpress ground), `--accent-30`, `--accent-deep` (accent at 70% over black, the flyer burst), and `--trk-0..3` (programme track bars: accent, accent toward text, accent toward white, text).

### Primary
- **Press Blue** `#2340d6`: letterpress edges, avatar tile and the Save Contact bar.
- **Boarding Yellow** `#ffd400`: the NOW BOARDING row and the board's chips.
- **Deal Red** `#d7261e`: the flyer masthead, rules and buy buttons; **Deal Red Deep** `#961b15` for the burst so it separates from the band.
- **Track Green** `#0e7c66`: programme primary action and the first track.

### Neutral
- **Card Stock / Paper** `#ffffff`, **Ink** `#14161c`, **Flyer Ink** `#17181c`, **Programme Ink** `#111318`.
- **Station Blue** `#1b2a6b`: the board itself (the owner's `card_fill`). It is deep blue on purpose, not black.

### Named Rules
- **The Owner's Brand Rule.** No template hard-codes its signature color. Every accent on the page is the owner's `accent`; the defaults above are only where a new page starts.
- **The Contrast Floor.** Text sits on a flat field whose contrast a checker can read. Layered paint (the letterpress foot, the flyer burst) is drawn as separate layers, never as a gradient behind text.

## Typography

One face per template, self-hosted as OFL woff2 (Latin and Latin-ext subsets) and preloaded only when the page uses it. Other scripts fall back to the system stack, so pages work in many languages.

- **Schibsted Grotesk** (letterpress): the name at 800 weight, debossed (inset highlight and shadow, no glow). Title and company in tracked caps.
- **Martian Mono** (board): everything uppercase in fixed columns; the via column in lowercase hostnames.
- **Barlow Condensed** (flyer): 800 weight in caps, prices at 46px (76px on the hero) with superscript cents and a struck-through "was" price.
- **Hanken Grotesk** (programme): tabular figures throughout, so times and the countdown align.

### Hierarchy
Exactly one `h1` per page: the owner's name, or the shop name on the flyer. The event title is an `h2` inside the cover. Section titles (Programme, Also this week, Getting there, form headings) are `h2`. A standalone product's name is an `h2`; product names inside a carousel are `h3`.

## Layout

Mobile first at 390px with an 18px gutter; the same column centers on desktop (max 620px on the flyer, a card width on letterpress). Styled pages group blocks into a **lead** (the first viewport's promise) and a **body**:

- Letterpress: the front of the card is the profile run up to Save Contact; everything after is the back.
- Board: the header plate is the profile; the link list is the board.
- Flyer: the masthead band holds the profile; products form the sheet.
- Programme: the cover holds the event details and countdown; the timetable starts at the fold.

No page scrolls horizontally at any width (the carousel scrolls inside itself, with scroll-snap and keyboard focus).

## Elevation & Depth

Flat by default. Letterpress alone has depth, because it is a held object: `0 2px 3px rgba(0,0,0,.06), 0 22px 44px -22px rgba(0,0,0,.4)` under each face of the card. Nothing else casts shadows.

## Shapes

Corners follow the medium: the card has a 12px radius and a square foot bar; the board is 3px; the flyer 2px; the programme 4px. The flyer burst is a **twelve-point star built from three equal squares** (0°, 30°, 60°) turned −8°: plain geometry, never a traced contour.

## Components

### Buttons
- **Save Contact bar** (letterpress): full card width, 66px, square, tracked caps, a contact icon. It is the card's painted foot, not a floating button.
- **Buy** (flyer): full-width Deal Red blocks with condensed caps. "Buy now" on the hero, "Buy" in the grid.
- **Get tickets / Add to calendar** (programme): a primary-and-secondary pair at equal width; the calendar button downloads an `.ics` file.

### Rows
- **Contact row** (letterpress): a small line icon, label, chevron, hairline rule. No icon tiles.
- **Board row**: destination, via host, optional status chip; the boarding row is larger and yellow.
- **Timetable row** (programme): time column, a 4px track bar cell, session title, speaker and track in small caps. A row with details opens with native `<details>`, no script.

### Media and maps
Videos, Spotify and the OpenStreetMap embed render as **click-to-load facades**: a poster or label plus a play affordance. The player iframe (youtube-nocookie, player.vimeo.com, open.spotify.com, www.openstreetmap.org) is created only when tapped. The page's Content-Security-Policy allows exactly those four frame origins.

### Forms
Newsletter and contact forms post without JavaScript and come back with an inline thank-you or a `role="alert"` message. They have a hidden honeypot and a signed render-time token. Pro and Ultra plans only.

## Do's and Don'ts

### Do:
- Do let the owner's four tokens paint the page; derive, never invent, extra colors.
- Do keep each template's world intact when adding a block: a new block gets a rule in each style sheet, or inherits the shared `blocks.css` look.
- Do preload only the face the page uses, and load `embed.js` or `countdown.js` only when the page has that block.
- Do keep functional text at 11px or larger, the "Made with Trunc" badge included.

### Don't:
- Don't put text on a gradient; split the paint into layers.
- Don't use eyebrow labels, colored side-borders as accents, or icon tiles behind contact actions.
- Don't draw organic shapes with many-vertex `clip-path`; use geometry.
- Don't change the classic style's output: existing pages must render byte-identical (guarded by golden tests in `internal/services/testdata/classic/`).

## Deviations from the direction contracts

Recorded where the build departs from a surface brief, with the reason:

1. **Board "№" column became a *via* column** (the destination's host). A number carried no meaning for a link, and the craft floor forbids decorative data.
2. **Programme track colour is its own 4px cell**, not a colored left border, per the craft floor's ban on colored side-borders.
3. **The flyer burst is three rotated squares, not a ≤12-point `clip-path` polygon.** The detector flags many-vertex clip-paths as organic contours. The square construction gives the same twelve points as geometry, and in the deep accent it stays visible where it overlaps the masthead.
4. **The letterpress foot is a separate painted layer, not a background gradient,** so every text color sits on a flat field (Contrast Floor).
5. **Templates ship without a form block.** Forms are Pro and Ultra only, so a Free owner's template must publish as-is.
