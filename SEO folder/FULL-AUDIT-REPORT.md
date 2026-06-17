# Full SEO Audit — trunc.site
**Audit Date:** June 14, 2026  
**Site:** https://trunc.site  
**Business Type:** SaaS — URL Shortener  
**Pages in Index:** 17 (sitemap) / 12 (Google Search Console)  
**Auditors:** 7 parallel specialist agents

---

## SEO Health Score: 64 / 100

| Category | Weight | Score | Weighted |
|---|---|---|---|
| Technical SEO | 22% | 74/100 | 16.3 |
| Content Quality | 23% | 48/100 | 11.0 |
| On-Page SEO | 20% | 72/100 | 14.4 |
| Schema / Structured Data | 10% | 58/100 | 5.8 |
| Performance (CWV) | 10% | 65/100 | 6.5 |
| AI Search Readiness | 10% | 68/100 | 6.8 |
| Images | 5% | 72/100 | 3.6 |
| **OVERALL** | 100% | **64/100** | |

**Verdict:** Solid foundation — Next.js SSR, sitemap, robots.txt, AI crawler access, and schema infrastructure are all in place. The dominant problem is content thinness (all blog posts <900 words, homepage 150 words) combined with two security header gaps (HSTS, CSP). Fixing content depth alone would push this score above 75.

---

## Executive Summary

### Top 5 Critical Issues
1. **All blog posts are thin content** — 682–872 words each (minimum 1,500 for competitive SERPs)
2. **Homepage has 150 words** — far below the 500-word floor for a SaaS homepage
3. **AggregateRating schema (4.8★ / 1,247 reviews) is not backed by any on-page reviews** — Google spam risk
4. **Missing HSTS and Content Security Policy headers** — critical security gaps for a SaaS product
5. **CLS failure risk from auth skeleton loaders in the nav** — 3 `animate-pulse` elements that shift on hydration

### Top 5 Quick Wins
1. Add `og:image` to pricing page — 30-minute fix, prevents blank social card when shared
2. Fix Features page title (79 chars → 55-58 chars) — immediate SERP display fix
3. Add `preconnect` for `t.contentsquare.net` — single HTML line, removes cold DNS on LCP path
4. Add `Strict-Transport-Security` header — single nginx config line, critical security signal
5. Add `priority` prop to blog hero `<Image>` component — Next.js one-liner, fixes LCP fetchpriority

---

## 1. Technical SEO (74/100)

### CRITICAL

#### C1 — Missing HSTS Header
Every HTTP visit incurs one unprotected round-trip before the 301 to HTTPS. SSL stripping attacks are possible. Google uses HSTS as a ranking trust signal.

**Fix:**
```
Strict-Transport-Security: max-age=31536000; includeSubDomains; preload
```
Submit to [hstspreload.org](https://hstspreload.org) after verifying all subdomains are HTTPS.

#### C2 — Missing Content Security Policy
The site loads scripts from 5 external origins (ContentSquare, Google Fonts, GTM, Ahrefs analytics) with no CSP. XSS vulnerabilities would allow arbitrary script execution across all origins — high risk for a SaaS handling auth and billing.

**Fix:** Start with `Content-Security-Policy-Report-Only` to identify violations, then enforce:
```
Content-Security-Policy: default-src 'self'; script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://analytics.ahrefs.com https://t.contentsquare.net; font-src 'self' https://fonts.gstatic.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; connect-src 'self'; frame-ancestors 'none';
```

### HIGH

#### H1 — Trailing-Slash Duplicate Content
`/features/` returns 200 with canonical pointing to `/features`. No server-level 301. Crawl budget wasted; link equity split.

**Fix (nginx):**
```nginx
location ~ ^(.+)/$ {
    return 301 $1;
}
```

#### H2 — Duplicate JSON-LD Schema Blocks on Blog Posts
`WebSite`, `Organization`, and `SoftwareApplication` schemas each appear **twice** on blog post pages — injected by both the root layout and the blog post template. This creates entity resolution ambiguity and Rich Results Test warnings.

**Fix:** Move `WebSite`, `Organization`, and `SoftwareApplication` to root `_app` layout only. Blog post components inject only `BlogPosting` + `BreadcrumbList`.

#### H3 — WebSite SearchAction Wrong Target
`potentialAction.target` is `https://trunc.site/?url={search_term_string}`. The `?url=` parameter is the URL shortener input, not a site search. This misleads Google's Sitelinks Searchbox.

**Fix:** Either update to a real site search endpoint (`?q={search_term_string}`) or remove `potentialAction` entirely.

#### H4 — No Blog Post Links from Homepage
Homepage links to `/blog` (index) but not to any individual posts. The blog posts are the primary organic traffic opportunity. PageRank from the homepage's highest-authority page does not flow to them.

**Fix:** Add a "Latest from the Blog" section with 3–4 direct links to top posts:
- `/blog/best-url-shorteners-2026`
- `/blog/trunc-vs-bitly-tinyurl`  
- `/blog/how-to-shorten-a-url`

#### H5 — Redundant Google Fonts Stylesheet (Local WOFF2 Already Loaded)
Inter is already self-hosted via `/_next/static/media/*.woff2` with `<link rel="preload">`. The `fonts.googleapis.com` stylesheet is also preloaded — adding 2 unnecessary cross-origin connections that waste bandwidth in the LCP window.

**Fix:** Remove all `fonts.googleapis.com` and `fonts.gstatic.com` `<link>` tags. The local Next.js font files are sufficient.

### MEDIUM

| Issue | Fix | Effort |
|---|---|---|
| Features title 79 chars (truncated in SERPs) | Trim to ~58 chars | Low |
| Features meta desc 199 chars | Trim to ≤160 chars | Low |
| Homepage meta desc 168 chars | Trim to ~113 chars | Low |
| Pricing title 42 chars (too short) | Expand to 58 chars with keywords | Low |
| Author page title 38 chars | Expand to 52 chars | Low |
| /api-docs lacks TechArticle schema | Add TechArticle JSON-LD | Low |
| X-XSS-Protection deprecated | Remove from nginx config | Low |
| IndexNow not implemented | Generate GUID key, add to nginx, call API on publish | Medium |
| Blog hero images missing explicit dimensions | Add width/height attrs | Low |
| Referrer-Policy: same-origin breaks analytics attribution | Change to strict-origin-when-cross-origin | Low |

### WORKING WELL
- robots.txt correctly disallows all app routes; explicitly allows GPTBot, OAI-SearchBot, ClaudeBot, PerplexityBot
- Sitemap at /sitemap.xml with correct lastmod, changefreq, priority values
- HTTP→HTTPS and www→non-www are clean 301s
- All pages have self-referencing canonicals with no conflicts
- Next.js SSR — full HTML content in initial response, no JS execution needed to index
- No noindex found on any public page
- Open Graph and Twitter card meta complete on most pages
- BreadcrumbList schema present on all blog posts

---

## 2. Content Quality & E-E-A-T (48/100)

**E-E-A-T Score: 45.7/100**

| Factor | Score | Key Gap |
|---|---|---|
| Experience (20%) | 52/100 | "We tested" claims unsubstantiated; no data exports or screenshots as evidence |
| Expertise (25%) | 55/100 | Technical signals strong in API guide; prior work history vague on author page |
| Authoritativeness (25%) | 38/100 | Zero external citations; no third-party recognition; no social profile links |
| Trustworthiness (30%) | 40/100 | AggregateRating not backed by visible reviews; "39% CTR" stat unsourced; no CoI disclosure |

### CRITICAL — Thin Content

| Page | Net Words | Gap | Priority |
|---|---|---|---|
| /blog/best-url-shorteners-2026 | 872 | -628 vs 1,500 min | Critical |
| /blog/trunc-vs-bitly-tinyurl | 757 | -743 | Critical |
| /blog/url-shortener-api-guide | 747 | -753 | High |
| /blog/how-to-shorten-a-url | 682 | -818 | High |
| Homepage | 150 | -350 vs 500 min | Critical |

The 5 unaudited blog posts (`what-is-a-url-shortener`, `how-to-shorten-amazon-affiliate-links`, `url-shortener-for-instagram`, `wordpress-url-shortener`, `url-shortener-to-earn-money`) follow the same 700–900 word pattern — assume thin until audited.

Top ranking competitors for "best url shorteners" and "trunc vs bitly" average **2,800–3,500 words**. Trunc's posts are 3–4× shorter.

### CRITICAL — AggregateRating Not Substantiated
`SoftwareApplication` schema declares `ratingValue: 4.8, ratingCount: 1247` on every page. There are **zero visible reviews** on the site. The single review object in schema lists "Sarah M." — not matched by any on-page testimonial. Google's QRG and structured data guidelines treat this as potentially deceptive. Risk: manual action or schema suppression.

### HIGH — Comparison Pages Have Trustworthiness Issues
- `/blog/trunc-vs-bitly-tinyurl`: Trunc wins every category with no acknowledged weakness; zero links to Bitly or TinyURL pricing pages; Rebrandly (a direct competitor searchers expect) is absent entirely
- `/blog/best-url-shorteners-2026`: Trunc ranked #1 by its own founder with no conflict-of-interest disclosure; "39% higher CTR with Rebrandly branded domains" cited as fact with no source

### HIGH — Author E-E-A-T Signals Missing
- No real photo (initials avatar only)
- No LinkedIn URL anywhere on site
- Prior work described as "several early-stage startups" — deliberately vague
- No conflict-of-interest disclosure despite founder reviewing own product
- Person schema has no `image` or `sameAs` fields

### HIGH — Content Pattern Signals AI Generation
Three signals flagged under the Sept 2025 QRG:
1. Identical verbatim author bio block across all 4 audited posts
2. "No credit card required" + "enterprise-grade" + "AI Insights" phrases appear mid-article on 5/6 pages
3. Opening paragraph structure is formula-driven across all posts

### AI Citation Readiness: 52/100
**Strengths:** `what-is-a-url-shortener` and `how-to-shorten-a-url` have direct-answer first sentences and 100% question-format H2 headings — these are close to citation-ready.  
**Gaps:** Zero FAQPage schema; body paragraphs at 60–80 words (GEO optimal is 134–167 words); zero outbound editorial citations; no FAQ blocks.

---

## 3. On-Page SEO (72/100)

| Page | Title (chars) | Meta Desc (chars) | H1 | H2s | OG:image | Status |
|---|---|---|---|---|---|---|
| / | 61 (OK) | 168 (LONG) | ✓ | 2 | ✓ | FAIR |
| /features | 79 (LONG) | 199 (LONG) | ✓ | 3 | **MISSING** | POOR |
| /pricing | 42 (SHORT) | 161 (OK) | ✓ | 3 | **MISSING** | FAIR |
| /about | 64 (LONG) | 160 (OK) | ✓ | 5 | **MISSING** | GOOD |
| /blog | 47 (SHORT) | 148 (SHORT) | ✓ | 5 | **MISSING** | GOOD |
| /blog/how-to-shorten-a-url | 55 | 151 | ✓ | 5 | ✓ | GOOD |
| /blog/trunc-vs-bitly-tinyurl | 57 | 168 (LONG) | ✓ | 5 | ✓ | GOOD |
| /blog/best-url-shorteners-2026 | 56 | 148 (SHORT) | ✓ | 5 | ✓ | GOOD |
| /blog/what-is-a-url-shortener | 52 | 169 (LONG) | ✓ | 5 | ✓ | GOOD |

**Strengths:** All pages have unique H1s. Canonical tags correct throughout. Blog posts have strong heading structure (5+ H2s each). No orphan pages.

**Gaps:**
- 4 pages missing `og:image` — blank social preview when shared
- Features and Homepage have meta descriptions over 160 chars (truncated in SERPs)
- Features title at 79 chars is well over the 50-60 char ideal
- Homepage and main pages have only 2-3 H2s (thin structure)
- Missing contextual internal links: Features→blog posts, Pricing→comparison post

---

## 4. Schema / Structured Data (58/100)

### Schema Inventory

| Page | Types Present |
|---|---|
| All pages | WebSite, Organization, SoftwareApplication (duplicated on blog posts) |
| /pricing | + FAQPage |
| Blog posts | + BlogPosting, BreadcrumbList |
| /blog/authors/malin-jayaweera | + Person |

### CRITICAL Issues

**AggregateRating not substantiated** — 4.8★/1,247 reviews in `SoftwareApplication` schema with zero on-page reviews. High risk of Google spam action. Fix: add visible reviews section or remove `aggregateRating` until backed by G2/Capterra data.

**SearchAction target is URL shortener form** — `?url={search_term_string}` is the shortening input, not site search. Remove `potentialAction` or fix to a real search endpoint.

### HIGH Issues

**Pricing page lacks per-plan Offer schema** — only the free tier at `price: "0"` is structured. Premium ($9.99/mo) and Enterprise ($39.99/mo) are loaded client-side, invisible to crawlers and AI models answering "how much does Trunc cost."

**BlogPosting author not entity-linked** — all blog posts have `author.url` but no `author.@id` and no `author.sameAs`. Google cannot definitively associate articles with a real person entity.

**Author page: bare `Person` instead of `ProfilePage + mainEntity`** — current pattern misses the Profile page rich result. Also missing `image` and `sameAs` fields.

**BlogPosting schema lists "Sarah M." as author** (from the SoftwareApplication review object) — should be Malin Jayaweera matching the visible byline.

### MEDIUM Issues
- `Organization` missing `sameAs` (no social profile links → weak Knowledge Graph entity)
- No `BreadcrumbList` on /pricing or /blog/authors/malin-jayaweera
- No `FAQPage` on any blog post (highest-value missing schema for AI Overviews)
- `FAQPage` on /pricing will not produce Google rich results (commercial site; restricted Aug 2023) — harmless for AI citation but note the limitation
- `wordCount`, `keywords`, `inLanguage` null on all BlogPosting schemas

---

## 5. Performance / Core Web Vitals (65/100)

*Field data not available. Assessment based on static HTML analysis.*

| Metric | Risk | Confidence | Primary Cause |
|---|---|---|---|
| LCP | AT RISK | Medium | Font FOUT; 3 third-party scripts competing in LCP window |
| INP | AT RISK | High | 13 JS chunks + ContentSquare MutationObserver on every interaction |
| CLS | FAILING | High | 3 `animate-pulse` nav skeleton loaders shift on hydration |

### CLS — CRITICAL (Expected to Fail)
Three auth placeholder skeletons in the nav (`w-14/h-8` for Login button, `w-28/h-9` for Get Started) are SSR'd then replaced on hydration. If hydrated button dimensions differ from skeleton dimensions by even 1px, this produces a measurable CLS shift on every page load. The `lighthouserc.js` CLS assertion is `maxNumericValue: 0.1` — these skeletons are the primary driver of any violation.

**Fix:** Render the logged-out state (Login + Get Started) as the default SSR output. Only defer auth-dependent personalization (user avatar, "My Dashboard"), not the static buttons.

### ContentSquare — HIGH INP Risk
ContentSquare (`t.contentsquare.net/uxa/0ac352c074f83.js`) is injected as a raw `<script async>` in the RSC payload with no `preconnect`. Session replay scripts attach MutationObserver + event listeners to every interaction — direct INP overhead. Cold DNS+TLS adds 100–300ms to its first load.

**Fix:** 
1. Add `<link rel="preconnect" href="https://t.contentsquare.net">` to `<head>`
2. Change ContentSquare to `<Script strategy="afterInteractive">` or `strategy="lazyOnload"`

### GTM + Ahrefs Preloaded Incorrectly
Both are `strategy="afterInteractive"` (correct) but also `<link rel="preload" as="script">` in the HTML head. Preloading `afterInteractive` scripts wastes LCP-window bandwidth. Remove the preload tags.

### JS Bundle
13 first-party async chunks + 3 third-party scripts = 16 concurrent JS network requests in the critical load window. All first-party chunks are `async` (not blocking). Reduce to 8–9 chunks by lazy-loading: `PWAInstallPrompt`, `CookieConsent`, and heavy dashboard providers not needed on marketing pages.

### Stale Config Files
- `performance-budgets.json` references `fid` — FID was removed from Chrome in September 2024. Replace with `inp: { max_value: 200 }`.
- `lighthouserc.js` references TTI — removed from Lighthouse scoring in v12.0 (October 2025). Remove TTI assertion; keep TBT (300ms) as the correct INP proxy.

---

## 6. AI Search Readiness / GEO (68/100)

| Platform | Score | Key Factor |
|---|---|---|
| Google AI Overviews | 6.5/10 | BlogPosting schema present; FAQPage missing; paragraphs 60–80 words (below 134-167 optimal) |
| Perplexity | 7.5/10 | SSR content, llms.txt accurate, direct-answer blog openings |
| ChatGPT (web) | 6.0/10 | No Wikipedia entity; no YouTube presence; brand entity signals weak |
| Bing Copilot | 7.0/10 | BingSiteAuth verified; sitemap indexed; OG metadata complete |

### Strengths
- llms.txt at `/llms.txt` — present, well-formed, includes pricing tier details in machine-readable format
- All AI crawlers explicitly allowed: GPTBot, OAI-SearchBot, ClaudeBot, PerplexityBot
- SSR — full content in HTML response; AI crawlers don't need JavaScript
- `what-is-a-url-shortener` and `how-to-shorten-a-url` have direct-answer first sentences and 100% question-format H2s

### Critical Gaps

**No FAQPage schema on any blog post** — 6 question-format H2 headings already exist on `what-is-a-url-shortener` and `how-to-shorten-a-url`. Adding FAQPage JSON-LD is a data change, not infrastructure — the `JsonLd` component is already wired.

**No HowTo schema on how-to post** — the 3-step process (Go to trunc.site → paste URL → click Shorten) is exactly the HowTo schema pattern. AI Overviews weight HowTo schema heavily for procedural queries.

**llms-full.txt not served** — `llms-full.txt` in the project root is Polar.sh dependency documentation (11,415 lines). Do not serve it. Create a real `llms-full.txt` in `frontend/public/` with actual Trunc blog post content (3,000–8,000 words).

**Blog paragraphs at 60–80 words** — GEO optimal extraction window is 134–167 words per section. Sections are too short to be cleanly quoted by AI models; they may be combined or skipped.

**No Wikipedia entity, no YouTube presence** — YouTube mention correlation with AI citation is 0.737 (highest measured signal). Wikipedia entity absence means Trunc is unknown to AI models weighting encyclopedic sources.

---

## 7. Images (72/100)

| Page | img count | WebP | AVIF | Missing alt | Missing dims | fetchpriority |
|---|---|---|---|---|---|---|
| Homepage | 0 | — | — | — | — | — |
| Pricing | 0 | — | — | — | — | — |
| Blog posts | 1 each | ✓ | ✗ | 0 | ✗ | MISSING |

**Blog hero image issues:**
- No `priority` prop on Next.js `<Image>` → outputs `loading="lazy"` instead of `fetchpriority="high"`. The image is preloaded via `<link rel="preload">` but without matching `fetchpriority="high"` on the element, the browser may still deprioritize rendering.
- No `sizes` attribute → browser fetches 1200px-wide image on a 375px mobile screen. Wastes mobile bandwidth.
- No AVIF format served alongside WebP.

**OG Images:**
- Homepage: `/og-image.webp` 1200×630 — present and correct
- Pricing: **MISSING** — blank social card when shared on Slack, LinkedIn, Twitter
- Features, About, Blog: **MISSING**

**Mobile text sizes:**
- `font-size: 10px` found in compiled CSS — below the 16px mobile readability threshold, will appear in Google Search Console Mobile Usability report
- Multiple `0.75rem` (12px) and `0.875rem` (14px) uses — audit and increase badge/label text to minimum 14px

---

## What Is Working Well

| Area | Signal |
|---|---|
| Technical foundation | Next.js SSR, sitemap, robots.txt all correct |
| AI crawler access | GPTBot, ClaudeBot, OAI-SearchBot, PerplexityBot all explicitly allowed |
| Schema infrastructure | WebSite, Organization, SoftwareApplication, BlogPosting, BreadcrumbList all present |
| Blog H2 structure | All blog posts have 5+ H2s, majority question-format — ideal for PAA and AI Overviews |
| llms.txt | Accurate, well-formatted, includes pricing tier details |
| SSR rendering | All content in HTML; no JS required for crawlers |
| HTTPS redirect | Clean 301 HTTP→HTTPS and www→non-www |
| Canonical tags | Correct self-referencing on all pages |
| Open Graph | Complete on blog posts; twitter:card = summary_large_image |
| Sitemap | All 17 pages correctly listed with lastmod |
| Internal blog linking | Blog posts link to each other; author page links to all 9 posts |
| font-display: swap | Active on self-hosted Inter WOFF2 files |
| Hero image preloading | Blog hero images correctly preloaded in `<head>` |
