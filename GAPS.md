# URLshortn — Development Gap Analysis
> **March 2026 — What Still Needs to Be Built**
> 
> Read this file before starting any development session. Follow the build order in Section 12.
> Cross off items as they are completed.

---

## 1. Current System Snapshot

| Layer | Technology | Status |
|---|---|---|
| Backend | Go + Gin + sqlx + PostgreSQL | ✅ Production-ready |
| Frontend | Next.js 16, React 19, TypeScript | ✅ Mostly complete |
| Auth | JWT + Google OAuth + Phone OTP | ✅ Complete |
| Real-time | WebSocket (gorilla/websocket) | ✅ Complete |
| Billing | Polar.sh (webhook + SDK) | ✅ Complete |
| Analytics DB | 5 dedicated analytics tables | ✅ Complete — data collecting |
| Analytics UI | Recharts + Google Maps | ⚠️ Partially wired |
| AI Insights | OpenRouter (google/gemma-3-4b-it:free) | ✅ Complete |
| Bot Detection | BotDetectionService + Redis rate check | ✅ Complete |
| Custom Domains | Real DB backend + frontend management page | ✅ Complete |
| White-label Branding | BrandingService + settings tab (Enterprise) | ✅ Complete |
| Redis Cluster | UniversalClient — cluster or single-node | ✅ Complete |
| DB Migrations | cmd/migrate runner for configs/migrations/ | ✅ Complete |

> **Important:** Pricing is managed entirely through Polar.sh. Never hardcode prices or plan names. Always fetch from `/api/v1/products`. Plan names in DB are `free`, `premium`, `enterprise`. Map to display names Free / Pro / Ultra in the frontend only.

---

## 2. What Is Already Built — Do Not Rebuild

### Backend
- URL shortening, redirects, batch shortening, custom codes, expiry
- Full click event capture per click: IP, UA, referrer, country
- 5 analytics tables already collecting data on every click:
  - `geographic_analytics` — country, region, city, lat/lon
  - `device_analytics` — device type, brand, OS, browser, screen resolution
  - `time_analytics` — hour of day, day of week, date breakdowns
  - `referrer_analytics` — referrer domain, UTM source/medium/campaign/term/content
  - `click_events` — raw event log
- WebSocket hub — broadcasts every click to subscribed dashboard clients
- A/B testing — full implementation with z-score, p-value, confidence intervals
- Multi-touch attribution — first-click, last-click, linear models
- Conversion goal tracking and reporting
- JWT auth + Google OAuth + phone OTP + email verification
- RBAC — 7 system roles + 4 team roles + fine-grained permissions
- API key management with per-endpoint usage tracking
- Polar.sh billing — checkout, webhooks, subscription lifecycle, product sync worker
- Admin panel — users, RBAC, coupons, CMS, SEO, webhook retry
- Rate limiting — Free: 1k req/hr, Premium: 5k req/hr, Enterprise: 20k req/hr
- Prometheus metrics + OpenTelemetry/Jaeger tracing + zap structured logging

### Frontend
- Landing page with URL shortener widget
- Dashboard with URL list and quick stats
- Analytics page shell with Recharts + Google Maps components
- Billing page — subscription card, usage meters, plan selector
- Pricing page — live from Polar API
- Admin panel — tabbed (Users, RBAC, Coupons, CMS, SEO, Settings)
- Settings — profile, preferences, password change
- API keys management page
- Bulk URL shortener page
- Conversion goals manager
- Dark mode — full CSS variable system via ThemeContext
- PWA — service worker, offline sync via IndexedDB, install prompt
- Mobile optimized — MobileAppShell, virtual scroll, touch gestures
- QR code generation component (qrcode library already installed)

---

## 3. Gap Overview

| Gap | Area | Priority | Effort |
|---|---|---|---|
| ~~Uncomment & wire analytics frontend routes~~ | ~~Frontend~~ | ~~🔴 CRITICAL~~ | ~~Low~~ | ✅ Done |
| ~~Fix `userAnalyticsService` nil panic~~ | ~~Backend~~ | ~~🔴 CRITICAL~~ | ~~Low~~ | ✅ Done |
| ~~Re-enable `cmsService` & `apiKeyService`~~ | ~~Backend~~ | ~~🔴 CRITICAL~~ | ~~Low~~ | ✅ Done |
| ~~MaxMind GeoIP .mmdb file setup~~ | ~~Infrastructure~~ | ~~🔴 CRITICAL~~ | ~~Low~~ | ✅ Documented in .env.example |
| ~~Bot detection & filtering~~ | ~~Backend~~ | ~~🟠 High~~ | ~~Medium~~ | ✅ Done |
| ~~Analytics plan gating (Free/Pro/Ultra)~~ | ~~Both~~ | ~~🟠 High~~ | ~~Medium~~ | ✅ Done |
| ~~Unregistered user teaser (blurred analytics)~~ | ~~Frontend~~ | ~~🟠 High~~ | ~~Low~~ | ✅ Done |
| ~~`ai_insights` DB table + schema~~ | ~~Backend~~ | ~~🟠 High~~ | ~~Low~~ | ✅ Done |
| ~~AI performance summary (OpenRouter)~~ | ~~Backend + Frontend~~ | ~~🟠 High~~ | ~~Medium~~ | ✅ Done |
| ~~AI anomaly detection~~ | ~~Backend + Frontend~~ | ~~🟠 High~~ | ~~Medium~~ | ✅ Done |
| ~~AI smart recommendations~~ | ~~Backend + Frontend~~ | ~~🟡 Medium~~ | ~~Medium~~ | ✅ Done |
| ~~Predictive click forecasting~~ | ~~Backend + Frontend~~ | ~~🟡 Medium~~ | ~~High~~ | ✅ Done |
| ~~Audience intelligence UI~~ | ~~Frontend~~ | ~~🟡 Medium~~ | ~~High~~ | ✅ Done |
| ~~Privacy: IP hashing~~ | ~~Backend~~ | ~~🟡 Medium~~ | ~~Medium~~ | ✅ Done (ip_hash in RecordClick) |
| ~~Privacy: data retention job~~ | ~~Backend~~ | ~~🟡 Medium~~ | ~~Medium~~ | ✅ Done |
| ~~Custom domains (currently stub)~~ | ~~Backend + Frontend~~ | ~~🟡 Medium~~ | ~~High~~ | ✅ Done |
| /teams page & team management UI | Frontend | 🟡 Medium | Medium |
| ~~/ ab-tests dedicated page~~ | ~~Frontend~~ | ~~🟢 Low~~ | ~~Low~~ | ✅ Done |
| ~~Database migration tool~~ | ~~Infrastructure~~ | ~~🟡 Medium~~ | ~~Medium~~ | ✅ Done (cmd/migrate) |

---

## 4. Critical Fixes — ✅ ALL COMPLETE

These were broken wires preventing existing features from working. All resolved.

### 4.1 MaxMind GeoIP Database
GeoIP enrichment (city, region, lat/lon) silently fails without the `.mmdb` file. All geographic analytics will be empty.

```bash
# Download free GeoLite2 City database from MaxMind:
# https://dev.maxmind.com/geoip/geolite2-free-geolocation-data

# Add to .env:
GEOIP_DB_PATH=/path/to/GeoLite2-City.mmdb
```

- Add error logging if file not found so it fails loudly instead of silently.

### 4.2 Fix `userAnalyticsService` Nil Panic
In `main.go`, `userAnalyticsService` is instantiated as nil. Handlers will panic on any request to `/api/v1/user-analytics/*`.

```go
// main.go — find this pattern:
var userAnalyticsService *services.UserAnalyticsService // nil placeholder

// Replace with:
userAnalyticsService := services.NewUserAnalyticsService(db, redisClient)

// Then pass to handler:
analyticsHandler := handlers.NewAnalyticsHandler(..., userAnalyticsService)
```

### 4.3 Re-enable Commented Services
`cmsService` and `apiKeyService` are commented out in `main.go`. This breaks the CMS and API key pages.

```go
// main.go — find and uncomment:
// cmsService := services.NewCMSService(db)
// apiKeyService := services.NewAPIKeyService(db, redisClient)
```

### 4.4 Uncomment Analytics Routes
In `routes.go`, most advanced analytics routes are commented out. The backend services and data exist — the routes just need to be re-enabled.

```go
// routes.go — uncomment this block:
// protected.GET("/analytics/:shortCode/geographic", h.GetGeographicAnalytics)
// protected.GET("/analytics/:shortCode/device",     h.GetDeviceAnalytics)
// protected.GET("/analytics/:shortCode/time",       h.GetTimeAnalytics)
// protected.GET("/analytics/:shortCode/referrers",  h.GetReferrerAnalytics)
// protected.GET("/analytics/:shortCode/heatmap",    h.GetHeatmapData)
// protected.GET("/analytics/:shortCode/map",        h.GetMapData)
```

---

## 5. Bot Detection — ✅ COMPLETE

### Schema Changes
```sql
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS is_bot     BOOLEAN DEFAULT FALSE;
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS is_unique  BOOLEAN DEFAULT FALSE;
ALTER TABLE click_events ADD COLUMN IF NOT EXISTS ip_hash    VARCHAR(64);

CREATE INDEX IF NOT EXISTS idx_click_events_is_bot ON click_events(is_bot);
```

### Detection Logic
Create `services/bot_detection.go`:

```go
// Detection signals — any one match = bot:
// 1. UA matches known bot strings
// 2. Missing Accept-Language header
// 3. Missing Accept header
// 4. Same IP + same short_code > 10 clicks in 60 seconds (Redis check)
// 5. Empty user agent

var botSignals = []string{
    "bot", "crawler", "spider", "scraper", "curl", "wget",
    "python-requests", "go-http", "java/", "libwww", "Jakarta",
    "Googlebot", "Bingbot", "Slurp", "DuckDuckBot", "Baiduspider",
}

func IsBot(ua string, headers map[string]string, ip string, shortCode string, redis *RedisClient) bool
```

### Integration Rules
- Call `BotDetector.IsBot()` inside the redirect handler before writing to `click_events`
- Set `is_bot = true` — **never discard bot clicks, just flag them**
- Generate `ip_hash = SHA-256(ip + daily_salt)` for unique visitor detection
- All analytics queries must add `WHERE is_bot = FALSE` unless explicitly showing bot stats

---

## 6. Analytics Plan Gating — ✅ COMPLETE

### Feature Matrix

| Analytics Feature | Free | Pro | Ultra |
|---|---|---|---|
| Click count (total) | ✅ | ✅ | ✅ |
| Traffic sources (type) | ✅ | ✅ | ✅ |
| Country-level geo | ✅ | ✅ | ✅ |
| Device & browser breakdown | ✅ | ✅ | ✅ |
| Data retention | 30 days | 1 year | Unlimited |
| City-level geo | ❌ | ✅ | ✅ |
| Time-series chart | ❌ | ✅ | ✅ |
| Heatmap grid (hour x day) | ❌ | ✅ | ✅ |
| UTM breakdown | ❌ | ✅ | ✅ |
| Real-time click feed (WS) | ❌ | ✅ | ✅ |
| Live world map | ❌ | ✅ | ✅ |
| Bot vs human split | ❌ | ✅ | ✅ |
| CSV export | ❌ | ✅ | ✅ |
| AI performance summary | ❌ | ❌ | ✅ |
| Anomaly detection | ❌ | ❌ | ✅ |
| Smart recommendations | ❌ | ❌ | ✅ |
| Predictive forecast | ❌ | ❌ | ✅ |
| Audience intelligence | ❌ | ❌ | ✅ |

### Implementation
Add `checkAnalyticsAccess(userPlan, feature)` helper in `middleware/auth.go`:

```go
const (
    FeatureCityGeo     = "city_geo"
    FeatureTimeSeries  = "time_series"
    FeatureHeatmap     = "heatmap"
    FeatureUTM         = "utm"
    FeatureRealtime    = "realtime"
    FeatureLiveMap     = "live_map"
    FeatureBotFilter   = "bot_filter"
    FeatureCSVExport   = "csv_export"
    FeatureAISummary   = "ai_summary"      // Ultra only
    FeatureAIAnomaly   = "ai_anomaly"      // Ultra only
    FeatureAIRecommend = "ai_recommend"    // Ultra only
    FeatureForecast    = "ai_forecast"     // Ultra only
    FeatureAudience    = "ai_audience"     // Ultra only
)

var planFeatures = map[string][]string{
    "free":       {/* basic only */},
    "premium":    {FeatureCityGeo, FeatureTimeSeries, FeatureHeatmap, FeatureUTM,
                   FeatureRealtime, FeatureLiveMap, FeatureBotFilter, FeatureCSVExport},
    "enterprise": {/* all features */},
}
```

> **Note:** Plan names in DB are `free`, `premium`, `enterprise` (from Polar). Map to display names **Free / Pro / Ultra** only in the frontend.

---

## 7. Unregistered User Analytics Teaser — ✅ COMPLETE

Unregistered users see **zero analytics** — but instead of a blank page, show a blurred teaser to drive signups.

### Frontend Component
```tsx
// components/AnalyticsTeaser.tsx
export function AnalyticsTeaser({ children }: { children: React.ReactNode }) {
  return (
    <div className="relative">
      <div className="blur-sm pointer-events-none select-none">
        {children}
      </div>
      <div className="absolute inset-0 flex items-center justify-center">
        <Card className="p-8 text-center shadow-xl">
          <h3>See who is clicking your links</h3>
          <p>Sign up free — no credit card required</p>
          <Button href="/register">Create free account</Button>
        </Card>
      </div>
    </div>
  )
}
```

### Rules
- Unregistered users: full blur + "Sign up free" CTA
- Free users on Pro features: lighter blur + "Upgrade to Pro" CTA
- Pro users on Ultra AI features: lighter blur + "Upgrade to Ultra" CTA
- Always render mock/placeholder data behind the blur so the value is obvious

---

## 8. AI-Powered Insights (Ultra Plan Only)

All AI features use the **Anthropic Claude API**. Results are cached in `ai_insights` table.

### 8.1 New Environment Variable
```env
# Uses OpenRouter (free tier) — NOT Anthropic directly
OPENROUTER_API_KEY=sk-or-...
```

### 8.2 New Database Table
```sql
CREATE TABLE IF NOT EXISTS ai_insights (
    id            BIGSERIAL PRIMARY KEY,
    short_code    VARCHAR(10) NOT NULL REFERENCES url_mappings(short_code) ON DELETE CASCADE,
    user_id       BIGINT REFERENCES users(id),
    insight_type  VARCHAR(50) NOT NULL,  -- summary|anomaly|recommendation|forecast|audience
    content       TEXT NOT NULL,
    data_snapshot JSONB,
    generated_at  TIMESTAMPTZ DEFAULT NOW(),
    expires_at    TIMESTAMPTZ,
    is_read       BOOLEAN DEFAULT FALSE
);

CREATE INDEX idx_ai_insights_short_code ON ai_insights(short_code);
CREATE INDEX idx_ai_insights_type ON ai_insights(insight_type);
```

### 8.3 AI Service Structure
Create `services/ai_insights.go`:

```go
type AIInsightsService struct {
    anthropicKey string
    db           *sqlx.DB
    redis        *RedisClient
}

// Methods to implement:
func (s *AIInsightsService) GetSummary(shortCode string) (*AIInsight, error)
func (s *AIInsightsService) GetAnomalies(shortCode string) (*AIInsight, error)
func (s *AIInsightsService) GetRecommendations(shortCode string) (*AIInsight, error)
func (s *AIInsightsService) GetForecast(shortCode string) (*AIInsight, error)
func (s *AIInsightsService) GetAudienceProfile(shortCode string) (*AIInsight, error)

// Always check cache before calling Claude API:
func (s *AIInsightsService) getCached(shortCode, insightType string) (*AIInsight, bool)
func (s *AIInsightsService) saveToCache(insight *AIInsight)

// Regenerate when click count increases by 10% or 100 new clicks:
func (s *AIInsightsService) ShouldRegenerate(shortCode, insightType string) bool
```

### 8.4 Claude API Prompts

**Performance Summary**
```
system: You are a concise analytics assistant. Respond in 2-3 sentences only.
        Be specific with numbers. Professional but friendly tone.

user:   Summarize this link's performance. Highlight what is notable.
        Data: {total_clicks, unique_clicks, top_country, top_device,
               top_source, trend_vs_last_week, best_day, best_hour}
```

**Anomaly Detection**
```
system: You are an analytics anomaly detector.
        Respond with JSON: {detected: bool, severity: low|medium|high,
        explanation: string, possible_cause: string}

user:   Analyze this click pattern for anomalies.
        Rolling 7-day average: {avg_clicks_per_hour}
        Last hour: {current_clicks}
        Ratio: {ratio}
        Recent referrers: {top_referrers_last_hour}
```

**Smart Recommendations**
```
system: You are a link performance coach. Give 2-3 specific, actionable tips.
        Base recommendations only on the data provided.

user:   Based on this data, when and how should this link be shared?
        Heatmap peak: {best_hour} on {best_day}
        Top platform: {top_social_platform}
        Top country: {top_country}
        Device split: {mobile_pct}% mobile, {desktop_pct}% desktop
```

**Predictive Forecast**
```
# Forecast numbers computed server-side (linear regression + 7-day seasonality)
# Claude only generates the narrative explanation:

system: You are a data analyst. Explain this 30-day click forecast in 2 sentences.
        Be specific about the projected numbers and trend direction.

user:   Last 30 days: {daily_clicks_array}
        Projected next 7 days: {forecast_array}
        Projected next 30 days total: {forecast_total}
```

### 8.5 New API Endpoints
```go
// Add to routes.go — Ultra-gated group:
ultra := protected.Group("/", RequireEnterprise())
ultra.GET("/insights/:shortCode/summary",         h.GetAISummary)
ultra.GET("/insights/:shortCode/anomalies",        h.GetAIAnomalies)
ultra.GET("/insights/:shortCode/recommendations",  h.GetAIRecommendations)
ultra.GET("/insights/:shortCode/forecast",         h.GetAIForecast)
ultra.GET("/insights/:shortCode/audience",         h.GetAudienceProfile)
```

---

## 9. Audience Intelligence (Ultra Only)

Attribution backend already exists. What is missing is the audience profiling layer and frontend UI.

### 9.1 Visitor Journey Mapping
```sql
-- New query: find the same visitor across multiple of the user's links
SELECT ce.short_code, ce.clicked_at, um.original_url
FROM click_events ce
JOIN url_mappings um ON um.short_code = ce.short_code
WHERE ce.ip_hash = $1
  AND um.user_id = $2
  AND ce.is_bot = FALSE
ORDER BY ce.clicked_at ASC;
```

### 9.2 Interest / Category Profiling
```
# Use Claude API to classify referrer domains:

system: Classify these referrer domains into 3-5 audience interest categories.
        Return JSON only: [{category: string, percentage: number, top_domains: string[]}]

user:   Top referrer domains for this link: {domain_list_with_click_counts}
```

### 9.3 Frontend Pages to Build
- `/dashboard/audience/[shortCode]` — new route
  - Visitor journey timeline visualization (Recharts)
  - Interest category donut chart
  - Cohort retention grid (similar to heatmap grid already built)
  - Attribution model selector (first-touch / last-touch / linear) — backend already supports all three

---

## 10. Privacy & Data Retention

### 10.1 IP Hashing — ✅ COMPLETE
Implemented in `RecordClick` in `services/shortener.go`. SHA-256(ip + daily_salt) written to `ip_hash` column on every click.

```go
dailySalt := time.Now().Format("2006-01-02") // rotates daily
ipHash := sha256.Sum256([]byte(ip + dailySalt))
click.IPHash = hex.EncodeToString(ipHash[:])

// Keep raw ip_address for 24h for abuse detection, then null it out
```

### 10.2 Data Retention Worker
Create `workers/retention_worker.go` — runs nightly at 2:00 AM:

```go
func (w *RetentionWorker) Run() {
    // 1. Null out raw IPs older than 24h
    // UPDATE click_events SET ip_address = NULL
    // WHERE clicked_at < NOW() - INTERVAL '24 hours' AND ip_address IS NOT NULL

    // 2. Delete click_events for free users older than 30 days
    // DELETE FROM click_events ce JOIN url_mappings um ...
    // WHERE u.account_type = 'free' AND ce.clicked_at < NOW() - INTERVAL '30 days'

    // 3. Delete for premium users older than 1 year
    // (enterprise = never delete)
}
```

---

## 11. Missing Frontend Pages

### 11.1 Analytics Sub-Pages
Routes are now live (Section 4 complete). Wire existing frontend components to these endpoints:

| Component | Endpoint |
|---|---|
| Clicks over time (LineChart) | `GET /analytics/:shortCode/trends` |
| Geographic map (Google Maps) | `GET /analytics/:shortCode/map` |
| Device breakdown (PieChart) | `GET /analytics/:shortCode` (device fields) |
| Referrer breakdown | `GET /analytics/:shortCode/referrers` |
| Hour × Day heatmap grid | `GET /analytics/:shortCode/heatmap` |
| UTM campaign table | `GET /analytics/:shortCode/referrers` (utm fields) |
| Real-time feed | `WS /realtime/ws` |
| AI insight cards | `GET /insights/:shortCode/*` (Ultra — blurred for others) |

### 11.2 Teams Page
Backend is complete (`teams`, `team_members`, `team_invitations` tables + RBAC handlers). Build:
- `/dashboard/team` — list members, invite by email, assign roles
- Team invitation email (email service exists, needs template)

### 11.3 A/B Tests Page
Backend fully implemented. Build:
- `/dashboard/ab-tests` — list tests, create new test, view results
- Results view with confidence level and p-value display (already calculated by backend)

### 11.4 Admin Analytics Tab
Add to existing admin panel:
- Total platform clicks over time
- Top links globally
- Bot traffic percentage trend
- User growth and plan distribution chart

---

## 12. Recommended Build Order

Work through these phases in order. Do not start a phase until the previous one is verified working.

### ✅ Fix Phase — COMPLETE
- [x] Download & configure MaxMind GeoIP `.mmdb` file — documented in `.env.example`
- [x] Fix `userAnalyticsService` nil initialization in `main.go` — wired via sqlx
- [x] Re-enable `cmsService` and `apiKeyService` in `main.go` — confirmed live inside `NewHandler()`
- [x] Uncomment analytics routes in `routes.go` — 10 routes live + `AdvancedAnalyticsHandler` created
- [ ] Verify all analytics endpoints return real data with test requests *(manual — requires prod env)*

### ✅ P1 — Bot Detection + Plan Gating — COMPLETE
- [x] Add `is_bot`, `is_unique`, `ip_hash` columns to `click_events` — schema migration in `initTables()`
- [x] Build `BotDetector` service in `services/bot_detection.go` — UA match, missing headers, Redis rate check, empty UA
- [x] Wire bot detection into redirect handler — `handlers.go` `RedirectURL`; bots flagged, never discarded
- [x] IP hash generated in `RecordClick` — SHA-256(ip + daily_salt), unique visitor tracked via Redis
- [x] Skip `ProcessEnhancedClickEvent` for bot clicks — aggregated tables stay clean
- [x] Implement `planFeatures` map and `CheckPlanFeature` middleware — `middleware/plan_gating.go`
- [x] Apply plan gating to analytics endpoints — geographic, time, heatmap, UTM, map gated to Pro+
- [x] Build `AnalyticsTeaser` blur component — 3 variants: unauthenticated / upgrade-pro / upgrade-ultra
- [x] `is_bot = FALSE` added to all `click_events` analytics queries

### ✅ P2 — AI Insights — COMPLETE
- [x] Add `OPENROUTER_API_KEY` to `.env.example` — uses OpenRouter free tier (`google/gemma-3-4b-it:free`)
- [x] Create `ai_insights` table with unique index on `(short_code, insight_type)` — upsert pattern
- [x] Build `AIInsightsService` in `services/ai_insights.go` — OpenRouter HTTP client, Redis + DB caching, regeneration on 100 clicks or 10% growth
- [x] Implement all 5 insight types: summary, anomaly, recommendations, forecast (linear regression), audience
- [x] Add `/insights/:shortCode/*` API endpoints — enterprise-gated via `CheckPlanFeature`
- [x] Build `AIInsightCard` + `AIInsightsPanel` frontend components
- [x] Add "AI Insights" third tab to analytics page; blur teaser for non-Ultra users

### ✅ P3 — Privacy + Missing Pages — COMPLETE
- [x] IP hashing in `RecordClick` — SHA-256(ip + daily_salt) written to `ip_hash` column
- [x] Build data retention worker (`workers/retention_worker.go`) — nightly at 02:00; nulls IPs >24h, purges free >30d, premium >1yr
- [x] Wire Referrers tab in `AdvancedAnalytics.tsx` — 4th tab using `analytics.referrers` from advanced endpoint
- [x] Build `/ab-tests` page — list/create/start/stop tests; results modal with p-value + confidence
- [x] Build `/audience/[shortCode]` page — referrer distribution + AI audience categories (Ultra gated)
- [x] Add "Analytics" tab to admin panel — platform KPIs, today's activity, plan distribution, system health
- [ ] `/dashboard/team` page — skipped; team service is stub-only (no real DB queries)

### ✅ P4 — Custom Domains + Infrastructure — COMPLETE
- [x] Implement custom domain backend — real SQL via `storage/postgres_domains.go`, `services/domain.go` wired to DB
- [x] Add database migration runner — `cmd/migrate/main.go` (up/down/status, reads `configs/migrations/*_up.sql`)
- [x] Configure Redis Cluster support — `redis.UniversalClient` in `storage/redis.go`; set `REDIS_CLUSTER_ADDRS` to enable
- [x] White-label branding for Enterprise — `BrandingService`, `BrandingHandler`, `/api/v1/branding` (GET+PUT), Branding tab in settings
- [x] Custom domains frontend — `/custom-domains` page with add/verify/delete flow
- [ ] SSO integration — too complex; deferred to separate feature branch

---

*URLshortn — Gap Analysis — March 2026*
