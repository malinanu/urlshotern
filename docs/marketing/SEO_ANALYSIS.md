# Trunc SEO Analysis

**Site:** https://trunc.site · **Prepared:** 18 Nov 2026 · **Market:** Global / US English

## Data sources and limits

| Source | What it gave us | Date |
|---|---|---|
| Ahrefs API (`public-domain-rating-free`) | Domain Rating for trunc.site and 13 competitors | 18 Nov 2026 |
| Semrush keyword export (`SEO folder/trunc_list_2026-04-09.csv`) | 109 URL-shortener keywords with volume, KD and intent (US) | 9 Apr 2026 |
| Semrush Site Audit PDF (`SEO folder/`) | Technical issues on trunc.site | 8 Apr 2026 |
| Manual SERP review (US web search) | Who ranks for the Trunc Pages keyword set | 18 Nov 2026 |
| Codebase review | Sitemap, robots, metadata, schema, page indexing | 18 Nov 2026 |

> **Limitation:** The connected Ahrefs account is on the **Free plan (0 API units)**. Keyword Explorer, Site Explorer traffic, backlinks and organic keyword endpoints all return "Insufficient plan". Volumes and difficulty for the Trunc Pages keyword set are therefore **not measured here**. Before committing the full 12-week calendar, run the seed list in §4 through Ahrefs/Semrush Keyword Explorer (one Semrush keyword export is enough) and update the priority column.

---

## 1. Baseline

| Metric | trunc.site | Note |
|---|---|---|
| Ahrefs Domain Rating | **0** | Ahrefs has no meaningful referring domains for the site. |
| Indexed blog posts | 9 (now 13 with this sprint) | All 9 earlier posts target "URL shortener" queries |
| Semrush site audit errors | 2 × 4XX, 2 broken internal links (Apr 2026) | Fixed status unknown; re-crawl |
| Semrush warnings | 248 uncached JS/CSS, 20 low text-HTML ratio, 8 low word count, 5 missing H1, 1 page blocked, 1 non-descriptive anchor | See §5 |

**The single biggest constraint is authority, not content.** At DR 0, Trunc will not rank for any head term in either niche, however good the article. Everything below is sequenced around that: target low-competition long-tail first, and build referring domains in parallel (see Marketing Plan §5).

## 2. Competitive landscape (Ahrefs DR, 18 Nov 2026)

| Competitor | DR | Competes on |
|---|---|---|
| linktr.ee | 94 | Link in bio (category leader) |
| tinyurl.com | 93 | URL shortener |
| beacons.ai | 91 | Link in bio + creator store |
| carrd.co | 91 | One-page sites, used as link in bio |
| bitly.com | 90 | Shortener, QR, link-in-bio pages, digital business card QR |
| linkin.bio (Later) | 83 | Instagram link in bio |
| dub.co | 82 | Developer-focused shortener (open source) |
| blinq.me | 81 | Digital business cards |
| qrcode-monkey.com | 79 | Free QR generator |
| taplink.at | 78 | Link in bio for small business / WhatsApp sellers |
| hihello.com | 77 | Digital business cards |
| popl.co | 77 | Digital + NFC business cards |
| wave.co | 55 | (Checked as a candidate for wavecnct.com; wrong domain, ignore) |

### What the SERPs show (manual review)

- **"free digital business card with QR code":** Results are vendor product pages (Uniqode, Wave Connect, KADO, QRCodeChimp, atom.bio, Bitly) plus a few vendor blog posts (QRStuff, Scanova). The intent is transactional: searchers want a tool page, not an essay. **Implication:** a dedicated `/digital-business-card` landing page is the real ranking asset. The blog post supports it.
- **"best linktree alternatives 2026":** Dominated by listicles, most written by vendors that rank themselves first (Bitly, QRStuff, Limey, Linkie, Replug, Zeeg), plus Capterra and affiliate blogs. **Implication:** an honest, criteria-led comparison can compete on quality. It needs Reddit/community distribution and backlinks to move.
- **"vCard QR code":** Mostly QR-generator tool pages (QR Code Tiger, Uniqode, QRCodeChimp, Adobe Express, Linkly) and thin how-to pages. **Implication:** the clear differentiator is the *static vCard QR vs. QR-to-page* trade-off: static codes can't be edited or tracked. Explaining that honestly is a good informational angle.
- **"link in bio for small business whatsapp":** Smaller niche, mostly small vendors (atom.bio, UniLink, Taplink, Take App, link.boo). Authority is lower, so this is the **most winnable SERP for a DR-0 site**.

## 3. Existing keyword universe (URL shortener, Semrush US)

| Tier | Keywords (vol / KD) | Verdict |
|---|---|---|
| Head | url shortener (60.5k / 97), free url shortener (8.1k / 100), url shortener free (6.6k / 100), website url shortener (2.4k / 100), custom url shortener (1k / 100) | Not winnable for 12+ months. Keep the homepage optimised, but don't write more posts for these. |
| Mid | best url shortener (1.3k / 54), highest paying url shortener (390 / 45), amazon url shortener (390 / 41), what is a url shortener (260 / 41), url shortener for instagram (210 / 49) | Already covered by existing posts. Refresh, don't duplicate. |
| Long-tail (winnable) | url shortener api (210 / 18), dropbox url shortener (110 / 10), owly url shortener (70 / 16), best url shortener to earn money (70 / 22), wordpress url shortener (50 / 19), url shortener images (70 / 25) | Low KD. One short post each, or FAQ sections inside existing posts. |
| Bridge to Pages | url shortener with qr code (40 / 72), qr code and url shortener (30 / 52), url shortener with analytics dashboard (40 / 96) | Use as secondary terms on the QR/Pages content. |

Existing coverage: `how-to-shorten-a-url`, `trunc-vs-bitly-tinyurl`, `best-url-shorteners-2026`, `what-is-a-url-shortener`, `url-shortener-api-guide`, `how-to-shorten-amazon-affiliate-links`, `url-shortener-for-instagram`, `wordpress-url-shortener`, `url-shortener-to-earn-money`.

## 4. New keyword clusters: Trunc Pages (to be validated)

Priority reflects intent fit and SERP competition seen manually, not measured volume.

| Cluster | Primary keyword | Secondary / long-tail | Intent | Target page | Priority |
|---|---|---|---|---|---|
| Digital business card | digital business card | free digital business card, digital business card with qr code, how to make a digital business card, electronic business card | Transactional / informational | Landing: `/digital-business-card` · Blog: **how-to-create-a-free-digital-business-card** ✅ | High |
| vCard / contact QR | vcard qr code | contact qr code, qr code for contact info, vcard vs digital business card, save contact qr | Informational | Blog: **vcard-qr-code** ✅ | High |
| Link in bio | link in bio | free link in bio, link in bio tool, best link in bio | Commercial | Landing: `/link-in-bio` | Medium (KD expected very high) |
| Linktree alternatives | linktree alternative | free linktree alternative, linktree alternatives 2026, linktree vs … | Commercial | Blog: **best-linktree-alternatives** ✅ · later `/vs/linktree` | High |
| Small business / WhatsApp | link in bio for small business | whatsapp link in bio, whatsapp shop link, wa.me link, instagram shop link | Informational / commercial | Blog: **link-in-bio-for-small-business** ✅ | High (most winnable) |
| Events | event landing page | event qr code, event page builder, add to calendar link | Informational | Blog (Q1) + template page | Medium |
| QR for print | qr code business card | qr code on flyer, dynamic vs static qr code, qr code tracking | Informational | Blog (Q1) | Medium |
| Competitor alternatives | blinq alternative, hihello alternative, popl alternative, beacons alternative | Commercial | `/vs/*` pages | Medium |

**Validation step:** export these seeds from a paid keyword tool and drop any cluster where the primary keyword has KD > 60 *and* no long-tail under KD 30.

## 5. Technical SEO findings (code + Semrush audit)

| # | Finding | Impact | Fix |
|---|---|---|---|
| T1 | **Every blog post is credited to "Alex Rivera", a persona** (`frontend/src/app/blog/posts.ts`, `/blog/authors/alex-rivera`). | E-E-A-T risk for Google. Also conflicts with PRODUCT.md's rule "no invented customers… or claims presented as real". If it's ever exposed, it damages trust. | Credit a real person (founder) with a real bio and `sameAs` links (LinkedIn/GitHub), or use "Trunc Team" as an Organization author. Keep a 301 from the old author URL. |
| T2 | No public landing pages for Pages use cases. The only coverage is a section of `/features`. | Transactional SERPs (§2) need dedicated pages, so blog posts alone won't rank. | Build `/digital-business-card`, `/link-in-bio`, `/qr-code-generator` (static marketing pages with a demo, FAQ and `SoftwareApplication` + `FAQPage` schema). Add them to `sitemap.ts`. |
| T3 | User pages at `/m/<slug>` are `noindex` unless the owner opts in (`allow_indexing`, `internal/services/microsite_render.go:182`). The "Made with Trunc" badge links to `trunc.site/?ref=m`. | Correct default for privacy. Because pages live on trunc.site, badge links are internal and pass no external authority, unlike Linktree's model. | Keep it. Consider a public **template gallery** (`/templates/*`) of indexable sample pages, the SEO surface Linktree gets from user pages. |
| T4 | Semrush: 248 uncached JS/CSS resources | Core Web Vitals and crawl budget | Confirm `Cache-Control: public, max-age=31536000, immutable` for `/_next/static/*` at the CloudPanel vhost and Cloudflare (Next sets it; a proxy may be stripping it). |
| T5 | Semrush: 5 pages missing H1, 8 low word count, 20 low text-HTML ratio | Thin-page signals | Likely auth/utility pages (`/login`, `/register`, `/forgot-password`…). Add `robots: { index: false }` metadata to them rather than padding content. Check `/contact` and `/about` H1s. |
| T6 | Semrush: 2 × 4XX and 2 broken internal links (Apr 2026) | Minor | Re-crawl and fix. |
| T7 | Blog post schema: `BlogPosting` and `BreadcrumbList` are present. `FAQPage` is not. | Missed rich results / AI-answer extraction | Add an optional `faq` field to `BlogPost` and emit `FAQPage` JSON-LD in `blog/[slug]/layout.tsx`. |
| T8 | `robots.ts` explicitly allows GPTBot, ClaudeBot, PerplexityBot, and `llms-full.txt` exists. | Positive for AI search visibility | Keep it. Add the Pages product to `llms-full.txt`. |
| T9 | Hero images: some posts ship both `.svg` and `.webp`; OG uses the `.webp`. | Fine | New posts follow the same pattern (WebP for OG). |

## 6. Quick wins (next 2 weeks)

1. Decide on the author fix (T1) before publishing more posts.
2. Ship the 4 Pages posts in this PR, and submit them in Google Search Console with URL Inspection → Request indexing.
3. Build `/digital-business-card` as the first landing page and link all 4 posts to it.
4. Fix cache headers (T4) and noindex utility pages (T5).
5. Start link acquisition: list Trunc on 15–20 SaaS directories and alternatives sites (AlternativeTo, SaaSHub, Product Hunt, G2, Capterra free listing, BetaList, Indie Hackers). This is the only way to move DR off 0.
6. Upgrade Ahrefs (Lite) for one month, or use Semrush, to validate §4 volumes and set up rank tracking.
