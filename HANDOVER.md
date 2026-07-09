# Handover — Trunc (URL shortener) work session

_Last updated: 2026-07-08_

This document summarizes a working session covering a **security audit + fixes**, a **homepage redesign**,
**real homepage stats**, a **Docker local-run setup against shared DBs**, an **SEO audit + supporting fixes**,
and **mobile/responsive + dev-stability fixes**.

> ⚠️ **Nothing in this session was committed or deployed.** All changes are in the working tree only. The
> **live site (https://trunc.site) is still the OLD build** — the redesign and fixes below ship only when you
> build + deploy. See “Deploy” below.

---

## 1. How it's running right now (local Docker)

Full stack runs in Docker against the developer's **pre-existing shared Postgres/Redis** (on the external
`shared-network`), not its own DB containers.

- Frontend (Next.js dev): http://localhost:3000
- Backend (Go API): http://localhost:8090  (health: `/health`, public stats: `/api/v1/public/stats`)
- DB: `shared-postgres:5432` (user `postgres` / `admin123`), database **`urlshortener_db`** (created + schema
  + all 46 migrations loaded this session)
- Cache: `shared-redis:6379`

**Start / stop:**
```bash
docker compose -f docker/docker-compose.sharedb.yml up -d          # start
docker compose -f docker/docker-compose.sharedb.yml logs -f        # logs
docker compose -f docker/docker-compose.sharedb.yml down           # stop
```

**Gotcha — backend cold builds:** the backend uses `air`. On a container restart/recreate the first Go build
is slow (cold cache); until it finishes, `/api/v1/public/stats` returns empty and the homepage trust strip
shows `—`. If that happens, warm it:
```bash
docker exec urlshortener-backend-dev sh -c "cd /app && go build -o ./tmp/main ./cmd/server"
```
If the frontend stays in `Created` after a rebuild, it's waiting on backend health — run
`docker compose -f docker/docker-compose.sharedb.yml up -d frontend` once the backend is healthy.

---

## 2. Security fixes (backend, Go)

- **IDOR on analytics/insights (critical):** any authenticated user could read any short code's analytics.
  Added `internal/middleware/ownership.go` (`RequireURLOwnership`) and wired it into every
  `/api/v1/analytics/:shortCode*` and `/api/v1/insights/:shortCode/*` route in `internal/routes/routes.go`.
  Admins bypass; API-key callers matched by owning user.
- **Stored XSS in OG page (critical):** `buildOGPage` in `internal/handlers/handlers.go` injected the
  destination URL into an inline `<script>` with only quotes escaped. Now embedded via `json.Marshal`
  (escapes `< > & " \`), closing the breakout.
- **JWT secret startup guard:** `cmd/server/main.go` now refuses to boot in `ENVIRONMENT=production` with an
  empty or default `JWT_SECRET` (warns elsewhere). **Prod must set a real `JWT_SECRET` or the server won't
  start.**
- **Password-check brute force:** `/api/v1/check-password/:shortCode` now rate-limited (`AuthRateLimit`).
- **CMS stored-XSS hardening:** added `frontend/src/utils/sanitize.ts` (`isomorphic-dompurify`) applied to the
  public CMS render sinks (`pages/[slug]/_client.tsx`, `privacy`, `terms`).
- **Rate-limiter data race:** added a mutex around the shared `visitors` map in
  `internal/middleware/rate_limit.go`.
- **API-key latency metric bug:** fixed `time.Since(time.Now())` (always ~0) in `internal/middleware/api_auth.go`.

**Deliberately NOT done (needs its own tested migration):** moving OAuth tokens out of the URL query string
and JWTs out of `localStorage` into `httpOnly` cookies. Flagged; not bundled here.

---

## 3. Homepage redesign (frontend)

Kept `src/app/page.tsx` as a server component composing new sections; motion is isolated in client leaves.

**New:** `src/components/motion/{Reveal,Stagger,MotionProvider}.tsx`,
`src/components/marketing/{Hero,HeroAnalyticsPreview,TrustStrip,FeaturesBento,HowItWorks,FromTheBlog,CtaBand}.tsx`.
Added dependency **`framer-motion`**.

Key decisions/fixes along the way:
- **Reduced motion / hydration:** never branch the tree on `useReducedMotion` (caused a hydration mismatch);
  reduced motion is handled globally by `<MotionConfig reducedMotion="user">` in `MotionProvider`.
- **Font:** the redesign broke the font to serif (Tailwind applied `var(--font-geist-sans)` on `<html>` where
  the var was undefined). **User chose to keep the new look but restore the OLD font.** `tailwind.config.ts`
  `fontFamily.sans` is back to the original `['Inter','system-ui',…]` and `globals.css` `--font-*` reverted.
  Net: renders the original system sans (Segoe UI), no serif.
- **FeaturesBento:** the original bento left empty grid cells + dead space; rewritten as an even
  `grid-cols-1 sm:2 lg:3` of 6 equal cards (2 tinted).
- **Fixed frontend audit issues:** removed fabricated `aggregateRating`/review JSON-LD (Google penalty risk)
  in `layout.tsx`; `URLShortener.tsx` dead `primary-500` focus ring → tokens, hardcoded green success colors →
  `--success` token, token-refresh retry now resends full body, removed noisy console; `Header.tsx` ARIA +
  Escape-to-close; `ui/card.tsx` `CardTitle` accepts `as`; `Footer.tsx` dead social placeholder removed;
  `ThemeToggle.tsx` dropped `as any`; `sitemap.ts` homepage `lastModified` dynamic; `Layout.tsx` toast on
  resend failure + `next/link`.

---

## 4. Real homepage stats (trust strip)

Replaced fabricated `500K+ / 50M+ / 99.9% uptime` with **real data** (uptime replaced by registered users):

- Backend: `models.PublicStats` (`internal/models/url.go`); `AnalyticsService.GetPublicStats()`
  (`internal/services/analytics.go`) — `COUNT(url_mappings)`, `SUM(click_count)`, `COUNT(users WHERE
  deleted_at IS NULL)`, wrapped in a 5-min in-memory cache; `Handler.GetPublicStats`
  (`internal/handlers/handlers.go`); public route `GET /api/v1/public/stats`.
- Frontend: `TrustStrip.tsx` is a client component fetching the endpoint on mount (client-side so it stays
  live in the static-export prod), formatting via `@/utils/formatters` `formatNumber`, labels
  `URLs shortened / Clicks tracked / Users`, `—` placeholder until loaded.
- **Mobile fix:** trust strip is a compact **3-across** row on all breakpoints (`grid-cols-3`, responsive
  text) instead of a tall stacked column.

Dev DB shows small real numbers (e.g. `1 / 1 / 0`) — correct, not a bug. Prod shows real large numbers.

---

## 5. Docker / assets / infra

- **`docker/docker-compose.sharedb.yml`** (new): backend + frontend against the shared DBs.
  `docker/docker-compose.local.ports.yml` (new): unused port-remap helper kept for reference.
- **`Dockerfile.frontend.dev`:** base `node:20-alpine` → **`node:22-bookworm-slim`** to fix the dev-only
  `controller[kState].transformAlgorithm is not a function` SSR streaming error (Node-20/alpine web-streams
  bug); healthcheck now a Node `fetch` to `/favicon.ico` (non-streaming) instead of `wget` on `/`.
- **Icons/PWA:** generated a branded icon set (sky-blue link glyph) into `public/` + `public/icons/` via
  `frontend/scripts/{icon-source.svg,generate-icons.mjs}` (run through a one-off `sharp` container — sharp is
  NOT a project dependency). Removed the `screenshots` array from `public/manifest.json` (can't fabricate real
  screenshots; they only 404).
- **Server-side backend URL:** `layout.tsx` + `sitemap.ts` now prefer `BACKEND_URL` (in-container
  `http://backend:8080`) for SSR fetches instead of the browser `NEXT_PUBLIC_API_URL`.
- **`.air.toml`:** `include_dir = ["cmd","internal","configs"]` so `air` watches ~18 dirs instead of ~740 —
  stops spurious cold rebuilds that kept taking the API down.

---

## 6. SEO audit — why trunc.site has no traffic

**It's not a technical bug.** The site is crawlable, has correct robots/sitemap, good meta/canonical/JSON-LD,
and solid ~1,200-word blog posts. The problem is **authority + indexing + age**:

- **Domain Rating = 0** (Ahrefs) → no backlinks.
- **Barely indexed** — `site:trunc.site` returns only the homepage; the blog posts are **not indexed at all**.
- ~3–4 month-old domain, still in Google's sandbox window.
- Homepage targets head terms ("free/best url shortener") owned by DR-90 incumbents (Bitly, TinyURL).

**Supporting code shipped (Part A):** `src/app/feed.xml/route.ts` (RSS), `FromTheBlog` homepage section
(1-hop internal links to posts), IndexNow key `public/7f3c…30.txt` + `scripts/indexnow-ping.mjs`,
`layout.tsx` third-party scripts → `strategy="lazyOnload"`, RSS auto-discovery `<link>` in metadata.

**The real levers (Part B — owner must do):**
1. **Google Search Console:** submit `sitemap.xml`, then **Request Indexing** for the homepage + all blog
   posts (they're unindexed — highest priority).
2. **Bing Webmaster Tools:** add site + submit sitemap.
3. **Backlinks:** Product Hunt, AlternativeTo, SaaSHub, Slant, listicle outreach, social profiles.
4. **Keep targeting long-tail** (not head terms).
5. **Give it 3–6+ months.**

---

## 7. Outstanding / next steps

- [ ] **Deploy** the redesign + all fixes (live site is still the old build).
- [ ] **Commit** the working-tree changes (nothing committed this session).
- [ ] Ensure **`JWT_SECRET`** is set in production (new startup guard).
- [ ] After deploy: run `node frontend/scripts/indexnow-ping.mjs`; submit sitemap + request indexing in GSC.
- [ ] (Optional, security) migrate auth tokens to `httpOnly` cookies (out of URL/localStorage).
- [ ] (Optional) consent-gate GA/Ahrefs/ContentSquare for GDPR.
- [ ] One pre-existing migration (`003_performance_optimization_up.sql`) has a non-`IMMUTABLE` partial-index
      predicate that Postgres 16 rejects — only a single perf index is skipped; fix if desired.

---

## 8. Verify quickly

```bash
# backend + frontend
curl -s http://localhost:8090/api/v1/public/stats           # {"total_urls":..,"total_clicks":..,"total_users":..}
curl -s -o /dev/null -w "%{http_code}\n" http://localhost:3000/   # 200

# frontend build (static export) sanity
cd frontend && npx tsc --noEmit && npm run build            # /feed.xml, icons, blog all generate
```
