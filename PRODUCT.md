# Product

<!-- impeccable:product-schema 1 -->

## Platform

web

## Users

Trunc account holders who publish **Trunc Pages**: mobile micro-landing pages and digital business cards at `/m/<slug>`. Four owner groups matter equally (confirmed):

- **Professionals and sales people** share a digital business card after meeting someone, often by QR code at an event or meeting. Their visitor's job: save the contact to the phone, then call, message or email.
- **Creators and personal brands** use a link-in-bio. Their visitor arrives from Instagram, TikTok or YouTube on a phone and taps through to the latest content.
- **Small shops and sellers** show a few products with buy links. Their visitor browses and taps through to checkout or WhatsApp.
- **Event organizers** share one page per event. Their visitor checks time, place and agenda, then registers or buys a ticket.

Visitors are almost always on a phone, often arriving from a QR scan or a social profile, with a few seconds of attention.

## Product Purpose

Trunc is a link shortener with first-party analytics, custom domains and QR codes. Trunc Pages extends it: an owner builds a page from blocks in a Build / Design / Track workspace, publishes it, and shares one link or QR code. Success for a page is the visitor doing the one thing the owner wants (saving the contact, tapping a link, buying, registering, subscribing). The owner then sees that in Track.

## Positioning

The page, its short link, its QR code and its analytics live in the same place. Saves, taps and sources (including QR scans) are measured first-party, with no third-party trackers on the page. The one-tap "Save Contact" produces a real vCard file. *(Inferred from the codebase.)*

## Operating Context

- Owners edit in the Trunc dashboard (desktop or mobile); a live preview renders the exact public page.
- Pages are served as server-rendered HTML with near-zero JavaScript, behind Cloudflare and nginx, under a strict Content Security Policy.
- Owners supply all content: names, photos, logos, products, event details. Templates are starting points that owners overwrite.

## Capabilities and Constraints

- Blocks are structured data, never raw HTML or CSS. Themes are validated tokens (colors, enumerated fonts, shapes) rendered server-side.
- Fonts must be self-hosted (no third-party font services). Any language may appear: owners write in many scripts, so every face needs robust fallbacks for non-Latin text.
- Plans: Free (1 page, Trunc badge), Pro (10 pages), Ultra (unlimited, badge removable). Forms are Pro and Ultra only.
- Pages must stay fast on mobile networks and work without JavaScript; scripts are progressive enhancement.

## Brand Commitments

- **The owner's brand leads** (confirmed). Templates are neutral starting points that the owner makes their own.
- Trunc appears only as the small "Made with Trunc" badge, which Ultra can remove.

## Evidence on Hand

- No real customer pages, testimonials or metrics exist yet.
- Template content is sample material and must read as placeholder: no invented customers, prices or claims presented as real.

## Product Principles

1. **One page, one action.** Every template makes the owner's primary action unmistakable within the first screen on a phone.
2. **The owner's identity, not ours.** Templates give owners a strong starting point that survives their own photo, colors and words.
3. **Works everywhere.** Any script, any phone, slow networks, no JavaScript required.
4. **Honest measurement.** Analytics are first-party and privacy-respecting; nothing is tracked that the owner can't see.

## Accessibility & Inclusion

- Readable text contrast on every theme, tap targets suited to thumbs, and visible keyboard focus.
- Reduced-motion respected.
- Must support many languages and scripts (confirmed), so template typography cannot depend on Latin-only faces without fallbacks.
