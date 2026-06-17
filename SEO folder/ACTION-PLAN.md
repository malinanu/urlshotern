# SEO Action Plan — trunc.site
**Generated:** June 14, 2026  
**Overall Score:** 64/100  
**Target Score (after all fixes):** 78–82/100

---

## CRITICAL — Fix Immediately

### 1. Remove AggregateRating Schema or Add Visible Reviews
**Impact:** Eliminates Google spam/manual action risk  
**Effort:** Medium (2–4 hours)  
**Category:** Schema / Trustworthiness

The `SoftwareApplication` schema on every page declares `ratingValue: 4.8` from `ratingCount: 1247` reviews. There are zero visible reviews on the site. This is high-risk — Google can suppress schema rich results or issue a manual action for deceptive structured data.

**Option A (Recommended):** Add a visible reviews section to the homepage with 3–5 real customer quotes (name, company, job title). Then either link to a G2 or Capterra profile that backs the aggregate count, or update `ratingCount` to match actual collected reviews.

**Option B:** Remove `aggregateRating` from the schema entirely until backed by an external review platform.

Also fix: the `author` field in `SoftwareApplication` review is `"Sarah M."` — this bleeds into `BlogPosting` schema and misidentifies the article author. Separate the review object from the article author attribution.

---

### 2. Add HSTS and CSP Headers
**Impact:** Critical security signals; HSTS is a Google ranking trust factor  
**Effort:** Low (30 minutes)  
**Category:** Security / Technical SEO

```nginx
# In nginx.conf or site config:
add_header Strict-Transport-Security "max-age=31536000; includeSubDomains; preload" always;
add_header Content-Security-Policy "default-src 'self'; script-src 'self' 'unsafe-inline' https://www.googletagmanager.com https://analytics.ahrefs.com https://t.contentsquare.net; font-src 'self' https://fonts.gstatic.com; style-src 'self' 'unsafe-inline' https://fonts.googleapis.com; img-src 'self' data: https:; connect-src 'self'; frame-ancestors 'none';" always;
```

Start CSP in `Report-Only` mode first to catch violations without breaking the site.

---

### 3. Fix CLS from Auth Skeleton Loaders
**Impact:** Likely CLS failure → ranking signal loss (CWV is a ranking factor)  
**Effort:** Medium (2–3 hours)  
**Category:** Performance / CWV

Three `animate-pulse` nav skeletons replace static auth buttons on hydration. If dimensions mismatch, CLS fails the 0.1 threshold on every page load.

**Fix:** Change the SSR default to render the logged-out state (Login + Get Started buttons) directly. Defer only authenticated personalization (avatar, "My Dashboard" link). This eliminates the skeleton-to-button shift entirely.

---

### 4. Expand Blog Posts to 1,500+ Words Each
**Impact:** This is the single biggest lever for organic traffic growth  
**Effort:** High (4–8 hours per post)  
**Category:** Content Quality

Priority order:

**A. /blog/best-url-shorteners-2026 (872 words → 2,500+ target)**
- Add methodology section ("how we tested")
- Expand each tool to 250–300 words
- Add Rebrandly and ClickMagick as additional tools
- Source or reframe the "39% higher CTR" stat (cite Rebrandly's original study)
- Add a full comparison table spanning all tools
- Add a 5-question FAQ block at the bottom

**B. /blog/trunc-vs-bitly-tinyurl (757 words → 2,000+ target)**
- Add Rebrandly as a 4th comparison (searchers expect it)
- Add a methodology section
- Acknowledge at least one Trunc limitation (builds trustworthiness)
- Link to Bitly's and TinyURL's own pricing pages (outbound editorial links)
- Add conflict-of-interest disclosure sentence
- Add 5-question FAQ block

**C. /blog/how-to-shorten-a-url (682 words → 1,800+ target)**
- Cover platform-agnostic methods first (browser extensions, Google Docs add-ons, Gmail)
- Then position Trunc as the recommended tool within each method
- Add FAQ block with PAA-style questions

**D. /blog/url-shortener-api-guide (747 words → 2,000+ target)**
- Add error codes/responses table
- Add webhook documentation section
- Add SDK/library examples (beyond cURL, JS, Python)
- Link to the OpenAPI spec

---

### 5. Expand Homepage to 500+ Words
**Impact:** Topical coverage for branded queries; closes the homepage content gap  
**Effort:** Medium (3–4 hours)  
**Category:** Content Quality

At 150 words the homepage has no "how it works" section, no named user personas, and no free tier details.

Add:
- **How it works** (4-step flow: paste URL → customize → share → track)
- **3 use case sections** by ICP: marketers (campaign tracking), developers (API integration), creators (Instagram bio links)
- **Free tier details** explicitly (10 links, basic analytics, QR codes — no account required)
- **Customer testimonials** section (minimum 3 named quotes with company/title — also satisfies AggregateRating fix)

---

## HIGH — Fix Within 1 Week

### 6. Add FAQPage Schema to Top 4 Blog Posts
**Impact:** Eligibility for Google AI Overviews FAQ carousels + People Also Ask boxes  
**Effort:** Low (2–3 hours total)  
**Category:** Schema / AI Search Readiness

Add a `faqs?: {question: string; answer: string}[]` field to the post type in `posts.ts`. Render `FAQPage` JSON-LD in `blog/[slug]/layout.tsx` when the field is present. The `JsonLd` component is already wired.

Add 5-question FAQ blocks to: `how-to-shorten-a-url`, `trunc-vs-bitly-tinyurl`, `best-url-shorteners-2026`, `url-shortener-api-guide`.

Example questions for `how-to-shorten-a-url`:
- "How do I shorten a URL for free?"
- "Does shortening a URL change where it goes?"
- "How do I track who clicks a shortened link?"
- "What is the shortest a URL can be?"
- "Can I customize a shortened URL?"

---

### 7. Add og:image to 4 Missing Pages
**Impact:** Social sharing generates blank cards for pricing, features, about, blog index  
**Effort:** Low (30 minutes)  
**Category:** On-Page SEO / Visual

Add to `/features`, `/pricing`, `/about`, `/blog`:
```html
<meta property="og:image" content="https://trunc.site/og-image.webp" />
<meta property="og:image:width" content="1200" />
<meta property="og:image:height" content="630" />
<meta name="twitter:image" content="https://trunc.site/og-image.webp" />
```
Ideally create page-specific OG images for /pricing and /features (use the `seo-image-gen` skill for prompts).

---

### 8. Fix Features Page Title and Meta Description
**Impact:** Features page title truncated in SERPs; meta description invisible after 160 chars  
**Effort:** Low (15 minutes)  
**Category:** On-Page SEO

**Current title (79 chars):** "URL Shortener Features — Analytics, Custom Domains, QR Codes & More | Trunc"  
**Suggested (58 chars):** "URL Shortener Features: Analytics, Custom Domains | Trunc"

**Current meta (199 chars):** Trim to ≤160 chars — remove "and more" filler clauses, keep primary features named.

---

### 9. Fix Auth Skeleton → Remove Redundant Google Fonts + Add ContentSquare Preconnect
**Impact:** LCP improvement; removes wasted cross-origin connections in LCP window  
**Effort:** Low (1 hour)  
**Category:** Performance

Three changes in `_document.tsx` or root layout:
1. Remove all `fonts.googleapis.com` and `fonts.gstatic.com` `<link>` tags — Inter is already self-hosted
2. Add `<link rel="preconnect" href="https://t.contentsquare.net">` 
3. Remove the GTM and Ahrefs `<link rel="preload" as="script">` tags from `<head>` (they're `afterInteractive` — preloading them wastes LCP bandwidth)

---

### 10. Fix BlogPosting Author Schema + Add @id and sameAs
**Impact:** E-E-A-T author entity linking; required for Article rich results  
**Effort:** Low (1 hour)  
**Category:** Schema / E-E-A-T

Update the `author` block in all `BlogPosting` schemas:
```json
"author": {
  "@type": "Person",
  "@id": "https://trunc.site/blog/authors/malin-jayaweera",
  "name": "Malin Jayaweera",
  "url": "https://trunc.site/blog/authors/malin-jayaweera",
  "sameAs": [
    "https://www.linkedin.com/in/malinjayaweera",
    "https://twitter.com/malinjayaweera"
  ]
}
```
The `@id` must match exactly the `@id` used in the author page's `Person` block.

---

### 11. Upgrade Author Page: ProfilePage Schema + Add Photo + LinkedIn
**Impact:** Profile page rich result eligibility; E-E-A-T authority  
**Effort:** Medium (2–3 hours)  
**Category:** Schema / E-E-A-T

- Replace bare `Person` schema with `ProfilePage + mainEntity: Person` pattern
- Add real author photo (not initials avatar) — required for author rich result
- Add LinkedIn URL to both the page and `sameAs` field
- Add conflict-of-interest disclosure: "Malin is the founder of Trunc. Posts comparing Trunc to competitors reflect founder expertise and disclosed product bias."

See schema snippet in FULL-AUDIT-REPORT.md §4 for the full JSON-LD.

---

### 12. Fix Trailing-Slash Redirect
**Impact:** Closes duplicate content issue; consolidates link equity  
**Effort:** Low (30 minutes)  
**Category:** Technical SEO

```nginx
location ~ ^(.+)/$ {
    return 301 $1;
}
```

---

### 13. Fix Duplicate JSON-LD on Blog Post Pages
**Impact:** Removes schema ambiguity; eliminates Rich Results Test warnings  
**Effort:** Low (1 hour)  
**Category:** Technical SEO / Schema

Move `WebSite`, `Organization`, and `SoftwareApplication` JSON-LD to root layout only. Blog post template should inject only `BlogPosting` + `BreadcrumbList`.

---

### 14. Add blog post links to Homepage
**Impact:** PageRank flow from highest-authority page to top content  
**Effort:** Low (1–2 hours)  
**Category:** Internal Linking

Add a "Latest from the Blog" section to the homepage with 3–4 links. Prioritize: `best-url-shorteners-2026`, `trunc-vs-bitly-tinyurl`, `how-to-shorten-a-url`.

---

### 15. Add Priority Prop to Blog Hero Images
**Impact:** Fixes LCP fetchpriority for blog posts  
**Effort:** Low (30 minutes)  
**Category:** Performance / Images

In the blog post `<Image>` component rendering the hero:
```tsx
<Image
  src={post.heroImage}
  alt={post.heroAlt}
  width={1200}
  height={630}
  priority  // ← add this
  sizes="(max-width: 768px) 100vw, (max-width: 1200px) 800px, 1200px"  // ← add this
/>
```

---

## MEDIUM — Fix Within 1 Month

### 16. Create Real llms-full.txt
**Impact:** Full-content file for AI crawlers; Perplexity and Claude use this when present  
**Effort:** Medium (3–5 hours)  
**Category:** AI Search Readiness

The existing `llms-full.txt` in the project root is Polar.sh dependency docs. Do not serve it.

Create `/home/malinjayaweera/dev/personal/urlshotern/frontend/public/llms-full.txt` with:
- Full text of all blog posts
- Features page expanded descriptions
- Pricing tier details
- API reference summary
Target: 3,000–8,000 words of actual Trunc content.

---

### 17. Implement IndexNow
**Impact:** Instant URL indexing on Bing/Yandex/Naver; no waiting for crawl schedule  
**Effort:** Medium (2–3 hours)  
**Category:** Technical SEO

1. Generate a GUID key
2. Place key file at `https://trunc.site/<key>.txt`
3. Call IndexNow API on every blog post publish and update
4. Integrate into the Next.js CMS publish pipeline

---

### 18. Fix Pricing Page Meta Description + Title
**Impact:** CTR improvement for high-intent "url shortener pricing" queries  
**Effort:** Low (15 minutes)  
**Category:** On-Page SEO

**Title (42 chars → target 58):** "URL Shortener Pricing: Free, Premium & Enterprise | Trunc"  
**Meta desc:** Already OK at 161 chars — minor trim to 155.

---

### 19. Add HowTo Schema to "How to Shorten a URL" Post
**Impact:** AI Overviews procedural list eligibility  
**Effort:** Low (1 hour)  
**Category:** Schema / AI Search Readiness

Supplement the existing `BlogPosting` with a `HowTo` object using the 3-step process already in the post content. Use the `HowTo` step text from the existing `ol` block in `posts.ts`.

---

### 20. Add Pricing Per-Plan Offer Schema
**Impact:** AI models can answer "how much does Trunc cost" with structured data  
**Effort:** Medium (2 hours)  
**Category:** Schema

Add page-specific `SoftwareApplication` override on `/pricing` with three `Offer` objects for Free, Premium ($9.99/mo), and Enterprise ($39.99/mo). See FULL-AUDIT-REPORT.md §4 for the full JSON-LD snippet. Requires server-side rendering of prices (currently client-side).

---

### 21. Add Organization sameAs Social Profiles
**Impact:** Brand entity disambiguation in Google's Knowledge Graph  
**Effort:** Low (30 minutes)  
**Category:** Schema / E-E-A-T

```json
"sameAs": [
  "https://twitter.com/trunc",
  "https://www.linkedin.com/company/trunc",
  "https://github.com/trunc"
]
```
Add actual verified social URLs.

---

### 22. Update Stale Performance Config Files
**Impact:** Monitoring accuracy; CI performance gates test the right metrics  
**Effort:** Low (30 minutes)  
**Category:** Performance

- `performance-budgets.json`: Replace `fid: { max_value: 100 }` with `inp: { max_value: 200 }`
- `lighthouserc.js`: Remove TTI assertion; TTI removed from Lighthouse 12.0 scoring (October 2025). Keep TBT ≤ 300ms.

---

### 23. Fix Referrer-Policy Header
**Impact:** Fixes analytics attribution — organic/referral traffic currently reported as "direct"  
**Effort:** Low (5 minutes)  
**Category:** Technical SEO

Change `Referrer-Policy: same-origin` to `Referrer-Policy: strict-origin-when-cross-origin`

---

### 24. Add BreadcrumbList to /pricing and Author Page
**Impact:** Breadcrumb URL display in SERPs  
**Effort:** Low (30 minutes)  
**Category:** Schema

Both pages are navigable sub-pages that should show the breadcrumb path in search results but currently lack `BreadcrumbList` schema.

---

### 25. Add Conflict-of-Interest Disclosure to Comparison Posts
**Impact:** Trustworthiness under Sept 2025 QRG  
**Effort:** Low (30 minutes)  
**Category:** Content / E-E-A-T

Add a one-sentence disclosure to `/blog/trunc-vs-bitly-tinyurl` and `/blog/best-url-shorteners-2026`:
> "Disclosure: This post is written by Malin Jayaweera, founder of Trunc. Our evaluations reflect direct product expertise and disclosed founder bias."

---

## LOW — Backlog

| # | Issue | Effort | Category |
|---|---|---|---|
| 26 | Remove deprecated `X-XSS-Protection` header | Low | Security |
| 27 | Remove `X-Permitted-Cross-Domain-Policies` (Flash era, irrelevant) | Low | Security |
| 28 | Add `TechArticle` schema to /api-docs | Low | Schema |
| 29 | Add `ItemList` / `CollectionPage` schema to /blog index | Low | Schema |
| 30 | Move ContentSquare to `strategy="afterInteractive"` | Low | Performance |
| 31 | Add `wordCount`, `keywords`, `inLanguage` to BlogPosting schemas | Low | Schema |
| 32 | Lazy-load PWAInstallPrompt and CookieConsent JS (not needed in critical path) | Medium | Performance |
| 33 | Add AVIF format alongside WebP for blog hero images | Medium | Images |
| 34 | Audit and increase sub-16px font sizes (10px, 12px occurrences) | Low | Mobile |
| 35 | Update sitemap lastmod dynamically on content change | Low | Crawlability |
| 36 | Source or remove the "39% higher CTR with branded domains" statistic | Low | Content |
| 37 | Add 5 named customer quotes to about page | Medium | E-E-A-T |
| 38 | Create a YouTube video (even a screen recording walkthrough) | High | AI Search |
| 39 | Consider a Wikipedia stub for Trunc | High | AI Search / Brand |
| 40 | Add Rebrandly to the "vs" comparison post | Medium | Content |

---

## Quick Reference — Prioritized by Effort × Impact

| Priority | Fix | Time | Score Impact |
|---|---|---|---|
| 🔴 1 | Fix AggregateRating (reviews or remove) | 2–4h | Eliminates spam risk |
| 🔴 2 | Add HSTS + CSP headers | 30m | +3 pts Technical |
| 🔴 3 | Fix CLS from auth skeleton hydration | 2–3h | +2 pts Performance |
| 🔴 4 | Expand blog posts to 1,500+ words | 4–8h/post | +8–10 pts Content |
| 🔴 5 | Expand homepage to 500+ words | 3–4h | +3 pts Content |
| 🟠 6 | Add FAQPage schema to 4 blog posts | 2–3h | +2 pts Schema + AI |
| 🟠 7 | Add og:image to 4 missing pages | 30m | +1 pt Images |
| 🟠 8 | Fix Features page title + meta desc | 15m | +1 pt On-Page |
| 🟠 9 | Remove redundant Google Fonts + add ContentSquare preconnect | 1h | +1 pt Performance |
| 🟠 10 | Fix BlogPosting author @id + sameAs | 1h | +1 pt Schema |
| 🟠 11 | Fix trailing-slash redirect | 30m | +1 pt Technical |
| 🟠 12 | Fix duplicate JSON-LD on blog posts | 1h | +1 pt Technical |
| 🟠 13 | Add blog links to homepage | 1–2h | +1 pt Technical |
| 🟠 14 | Add priority prop to blog hero images | 30m | +1 pt Performance |
| 🟡 15 | Create real llms-full.txt | 3–5h | +1 pt AI Search |
| 🟡 16 | Implement IndexNow | 2–3h | Bing indexing speed |
| 🟡 17 | Upgrade author page (ProfilePage + photo + LinkedIn) | 2–3h | +1 pt E-E-A-T |
| 🟡 18 | Add HowTo schema | 1h | +1 pt AI Search |
| 🟡 19 | Add per-plan Offer schema on pricing | 2h | +1 pt Schema |
| 🟡 20 | Update performance-budgets.json (FID→INP) | 30m | Monitoring accuracy |
