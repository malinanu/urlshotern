# Trunc Marketing Plan: Q4 2026 – Q1 2027

**Prepared:** 18 Nov 2026 · **Horizon:** 90 days (to mid-Feb 2027) · **Market:** Global / US English
**Companion doc:** [SEO_ANALYSIS.md](./SEO_ANALYSIS.md)

---

## 1. Situation

- Trunc combines a URL shortener (first-party analytics, custom domains, QR codes) with **Trunc Pages**: digital business cards, link-in-bio, product and event pages at `/m/<slug>`.
- The shortener market is saturated (Bitly DR 90, TinyURL DR 93) and its head terms run KD 97–100. trunc.site has **DR 0**.
- Pages competes in markets where vendor-written content dominates (Linktree, Beacons, Blinq, HiHello, Popl), but the long tail (small business, WhatsApp sellers, vCard how-tos, event pages) has weaker incumbents.
- No customer proof exists yet (PRODUCT.md: "No real customer pages, testimonials or metrics"). All marketing must stay honest: no invented users, logos or numbers.

## 2. Positioning

**One-liner:** *Your page, its short link, its QR code and its analytics, all in one place.*

**Differentiators to repeat everywhere**
1. **One link, one QR, one dashboard.** Competitors split these across products or plans.
2. **Real "Save Contact".** It produces a genuine vCard file, so the contact lands in the phone in one tap. No app is needed on either side.
3. **Honest, first-party analytics.** There are no third-party trackers on the visitor's page. Owners see views, saves, taps and QR scans.
4. **Fast and works without JavaScript.** It suits slow mobile networks and any language or script.
5. **Generous free tier.** One page, free forever, plus the full shortener.

**Pricing anchor (from `/pricing`):** Free · Pro $12/mo ($9 annual) · Growth $24/mo ($19 annual) · Ultra $49/mo ($39 annual). Before each campaign, check the prices against Polar.

## 3. Ideal customer profiles

| ICP | Job to be done | Where they are | Hook | Primary keyword cluster |
|---|---|---|---|---|
| **Professionals / sales** | Share contact after meeting someone and get saved | LinkedIn, events, conferences | "Your business card, saved in one tap" | digital business card, vCard QR |
| **Creators** | Route social traffic to latest content | Instagram, TikTok, YouTube | "A link in bio that shows what's working" | link in bio, Linktree alternative |
| **Small shops / sellers** | Turn followers into WhatsApp orders | Instagram, Facebook, WhatsApp Business | "Your shop on one link, orders on WhatsApp" | link in bio for small business, WhatsApp shop link |
| **Event organizers** | Get people to register and show up | Eventbrite, Meetup, posters, flyers | "One page and one QR for every event" | event landing page, event QR code |

## 4. Goals and KPIs (90 days)

| Goal | KPI | Baseline | 90-day target |
|---|---|---|---|
| Build authority | Ahrefs DR / referring domains | 0 / ~0 | DR 15+ / 40+ RDs |
| Organic visibility | GSC impressions per month | measure week 1 | 3× baseline |
| Organic traffic | GSC clicks per month to `/blog/*` + landing pages | measure week 1 | 1,000 |
| Rankings | Keywords in top 20 (Pages clusters) | 0 | 25 |
| Activation | New sign-ups that publish a page | measure | 150 published pages |
| Conversion | Free → paid | measure | 3% of activated |

These targets are deliberately modest for a DR-0 domain. Review them at day 45.

**Measurement**
- Google Search Console: impressions, clicks, queries per post.
- Trunc's own analytics: use **Trunc short links for every off-site placement** (directories, Reddit, social) so each source is measured first-party. That also demonstrates the product.
- UTM convention: `utm_source=<site>&utm_medium=<referral|social|community|directory>&utm_campaign=<pages-launch|blog-<slug>>`.
- Registration events: record `?ref=` and UTM on sign-up, and report published-page rate by source.

## 5. Channels

### 5.1 SEO content (primary, compounding)
- Cadence: **1–2 posts a week**, alternating Pages clusters with refreshes of existing shortener posts.
- Every post links to one money page (landing page or `/register`) and two related posts.
- Format rules: answer first in the opening paragraph, a comparison table where relevant, an FAQ block (add `FAQPage` schema, see SEO T7), and no fabricated statistics.

### 5.2 Landing pages (the actual ranking assets)
| Page | Target | Contents |
|---|---|---|
| `/digital-business-card` | digital business card, free digital business card | Live demo card, Save Contact demo, QR, 3-step how-to, FAQ, CTA |
| `/link-in-bio` | link in bio, free link in bio | Template preview, analytics screenshot, comparison table, FAQ |
| `/qr-code-generator` | qr code generator with analytics, dynamic qr code | Working free generator (link magnet) |
| `/vs/linktree`, `/vs/blinq`, `/vs/bitly` | brand + alternative | Honest feature/price tables, dated |
| `/templates/*` | "digital business card template", "event page template" | Indexable sample pages |

### 5.3 Link building (needed to move off DR 0)
- **Directories (week 1–2):** AlternativeTo (list as an alternative to Linktree, Bitly, Blinq), SaaSHub, G2, Capterra, Product Hunt (upcoming page), BetaList, Indie Hackers products, There's An AI For That (if AI Insights qualifies), Startup Stash, Uneed, Microlaunch.
- **Free tool as a link magnet:** the `/qr-code-generator` and a "vCard QR generator" with no sign-up get organic links from how-to articles.
- **Open-source angle:** publish a small Go package (for example, vCard 3.0 builder or Base62 shortcode) on GitHub that links back.
- **Guest posts and podcasts:** indie SaaS and Go/Next.js engineering topics ("Serving zero-JS pages behind a strict CSP"). Dev audiences link generously.
- **HARO-style platforms** (Qwoted, Featured.com): answer queries about networking, QR codes and small-business marketing.

### 5.4 Launches
- **Product Hunt launch** of Trunc Pages in week 6, once the landing pages and 8+ posts are live. Prepare a gallery of real (owner-made) pages, a 60-second demo video, and a maker comment.
- **Hacker News "Show HN"**: the engineering story (zero-JS pages, first-party analytics, vCard generation in Go).

### 5.5 Community (manual, value-first)
- Reddit: r/smallbusiness, r/Entrepreneur, r/sales, r/realtors, r/Instagram, r/EventProduction, r/SideProject. Answer questions first and mention Trunc only when it's relevant. Never astroturf.
- LinkedIn: weekly founder posts (building in public, before/after of paper vs digital cards).
- Short-form video (TikTok/Reels/Shorts): "Scan → contact saved in 2 seconds" demos. These are cheap to produce and fit the visual product.

### 5.6 Partnerships
- Event organizers and coworking spaces: give attendees a free Pro month for event pages.
- NFC card and sticker sellers on Etsy/Amazon: Trunc as the page behind their tags.
- Print shops: a QR code on the back of printed cards that points to a Trunc page.

### 5.7 Lifecycle and in-product
- "Made with Trunc" badge on Free pages links to `trunc.site/?ref=m`. Track sign-ups from `ref=m` as the viral coefficient.
- Onboarding emails (3-step): publish your page → share the QR → see your first scan in Track.
- Upgrade prompts at real limits: second page, forms, badge removal.

## 6. 12-week content calendar

✅ = shipped in this PR (dated 18 Nov 2026).

| Wk | Date (Mon) | Piece | Type | Primary keyword | Links to |
|---|---|---|---|---|---|
| 1 | 17 Nov | ✅ How to Create a Free Digital Business Card (with QR Code) | Blog | how to make a digital business card | /register, vCard QR post |
| 1 | 17 Nov | ✅ vCard QR Code: What It Is and How to Make One | Blog | vcard qr code | digital card post |
| 2 | 24 Nov | ✅ Best Linktree Alternatives in 2026 | Blog | linktree alternatives | Instagram post, small biz post |
| 2 | 24 Nov | ✅ Link in Bio for Small Business (WhatsApp Orders) | Blog | link in bio for small business | Linktree alts post |
| 3 | 1 Dec | `/digital-business-card` landing page | Page | digital business card | all Pages posts |
| 4 | 8 Dec | Event Landing Page: How to Build One People Register On | Blog | event landing page | /register |
| 4 | 8 Dec | Refresh: URL Shortener for Instagram (link to Pages) | Refresh | url shortener for instagram | Linktree alts |
| 5 | 15 Dec | `/link-in-bio` landing page | Page | link in bio | creator posts |
| 5 | 15 Dec | Static vs Dynamic QR Codes: Which Should You Print? | Blog | dynamic vs static qr code | QR generator |
| 6 | 22 Dec | QR Code on Business Cards: Design and Placement Guide | Blog | qr code business card | digital card post |
| 7 | 5 Jan | **Product Hunt launch** + `/qr-code-generator` | Launch/tool | qr code generator with analytics | — |
| 7 | 5 Jan | How to Add a WhatsApp Link to Instagram Bio | Blog | whatsapp link in bio | small biz post |
| 8 | 12 Jan | Trunc vs Linktree (`/vs/linktree`) | Page | linktree vs | Linktree alts |
| 8 | 12 Jan | Digital Business Card for Real Estate Agents | Blog | digital business card for realtors | digital card LP |
| 9 | 19 Jan | Trunc vs Blinq (`/vs/blinq`) | Page | blinq alternative | digital card LP |
| 9 | 19 Jan | How to Track QR Code Scans (Free) | Blog | track qr code scans | QR generator |
| 10 | 26 Jan | Networking Event Follow-Up: Get Saved, Not Forgotten | Blog | networking follow up | digital card LP |
| 11 | 2 Feb | Link in Bio Analytics: What to Measure | Blog | link in bio analytics | /link-in-bio |
| 11 | 2 Feb | Refresh: Best URL Shorteners 2026 → 2027 edition | Refresh | best url shortener | — |
| 12 | 9 Feb | 90-day review: rankings, DR, sign-ups → next-quarter plan | Review | — | — |

## 7. Budget (solo-founder, lean)

| Item | Monthly (USD) | Note |
|---|---|---|
| Ahrefs Lite or Semrush Pro (1–2 months for validation and rank tracking) | ~$130–140 | Cancel after the keyword validation sprint if needed |
| Product Hunt launch assets (video, gallery) | $0–150 one-off | DIY with screen recording |
| Directory paid listings | $0 | Free tiers only |
| Optional paid social test (Meta, small-business ICP) | $100–200 | Only after the landing pages are live. Track with Trunc links. |

## 8. Risks

- **Authority gap:** without links, content won't rank. Treat link building as weekly, non-optional work.
- **Persona author:** fix it before scaling content (SEO T1).
- **Feature flag:** Pages requires `ENABLE_MICROSITES=true` in production. Confirm it's on before driving traffic to Pages content.
- **Claim accuracy:** every post must match shipped features and live pricing. Re-check quarterly and update the `updatedAt` dates.

## 9. Owner checklist (week 1)

- [ ] Decide on the author byline (real founder vs "Trunc Team")
- [ ] Merge this PR, deploy, and request indexing for the 4 new URLs in GSC
- [ ] Submit to 10 directories using Trunc-tracked links
- [ ] Validate the §4 keyword clusters in SEO_ANALYSIS with a paid tool
- [ ] Brief and build `/digital-business-card`
