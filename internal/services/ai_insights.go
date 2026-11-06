package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"log"
	"math/rand"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

const (
	openRouterURL = "https://openrouter.ai/api/v1/chat/completions"
	// google/gemma-3-4b-it:free — free tier model.
	// NOTE: Gemma does NOT support the system role; system prompt is merged into user message.
	openRouterModel          = "google/gemma-3-4b-it:free"
	insightCacheTTL          = time.Hour
	insightMaxTokens         = 512
	openRouterMaxRetries     = 3
	openRouterRetryBaseDelay = 3 * time.Second
)

// AIInsightsService generates AI-powered insights for short links using OpenRouter.
type AIInsightsService struct {
	apiKey     string
	db         *storage.PostgresStorage
	redis      *storage.RedisStorage
	analytics  *AdvancedAnalyticsService
	httpClient *http.Client
	// sem serialises OpenRouter calls to 1-at-a-time to avoid rate-limiting when
	// the dashboard fires 5 insight requests simultaneously.
	sem chan struct{}
}

// NewAIInsightsService creates a new AIInsightsService.
func NewAIInsightsService(apiKey string, db *storage.PostgresStorage, redis *storage.RedisStorage, analytics *AdvancedAnalyticsService) *AIInsightsService {
	return &AIInsightsService{
		apiKey:     apiKey,
		db:         db,
		redis:      redis,
		analytics:  analytics,
		httpClient: &http.Client{Timeout: 60 * time.Second},
		sem:        make(chan struct{}, 1), // only 1 concurrent OpenRouter call
	}
}

// ─── Public Methods ───────────────────────────────────────────────────────────

func (s *AIInsightsService) GetSummary(shortCode string, userID int64) (*models.AIInsight, error) {
	const insightType = "summary"
	if cached, ok := s.getCached(shortCode, insightType); ok {
		return cached, nil
	}
	analytics, err := s.analytics.GetAdvancedAnalytics(shortCode, 30)
	if err != nil {
		return nil, fmt.Errorf("analytics fetch failed: %w", err)
	}
	prompt := s.buildSummaryPrompt(analytics)
	system := "You are an analytics assistant. Provide a concise 2-3 sentence performance summary using the specific numbers provided. Be direct and insightful."
	return s.generateAndSave(shortCode, userID, insightType, system, prompt, analytics.TotalClicks)
}

func (s *AIInsightsService) GetAnomalies(shortCode string, userID int64) (*models.AIInsight, error) {
	const insightType = "anomaly"
	if cached, ok := s.getCached(shortCode, insightType); ok {
		return cached, nil
	}
	analytics, err := s.analytics.GetAdvancedAnalytics(shortCode, 30)
	if err != nil {
		return nil, fmt.Errorf("analytics fetch failed: %w", err)
	}
	prompt := s.buildAnomalyPrompt(analytics)
	system := `You are an anomaly detector for web analytics. Respond ONLY with valid JSON, no markdown, no explanation outside JSON.
Format: {"detected": true/false, "severity": "low|medium|high", "explanation": "...", "possible_cause": "..."}`
	return s.generateAndSave(shortCode, userID, insightType, system, prompt, analytics.TotalClicks)
}

func (s *AIInsightsService) GetRecommendations(shortCode string, userID int64) (*models.AIInsight, error) {
	const insightType = "recommendation"
	if cached, ok := s.getCached(shortCode, insightType); ok {
		return cached, nil
	}
	analytics, err := s.analytics.GetAdvancedAnalytics(shortCode, 30)
	if err != nil {
		return nil, fmt.Errorf("analytics fetch failed: %w", err)
	}
	prompt := s.buildRecommendationsPrompt(analytics)
	system := "You are a link performance coach. Give 2-3 actionable, specific tips to improve click-through rate based on the data."
	return s.generateAndSave(shortCode, userID, insightType, system, prompt, analytics.TotalClicks)
}

func (s *AIInsightsService) GetForecast(shortCode string, userID int64) (*models.AIInsight, error) {
	const insightType = "forecast"
	if cached, ok := s.getCached(shortCode, insightType); ok {
		return cached, nil
	}
	analytics, err := s.analytics.GetAdvancedAnalytics(shortCode, 30)
	if err != nil {
		return nil, fmt.Errorf("analytics fetch failed: %w", err)
	}
	dailyClicks, err := s.getDailyClicks(shortCode, 30)
	if err != nil || len(dailyClicks) == 0 {
		dailyClicks = s.extractMonthlyAsDaily(analytics)
	}
	prompt := s.buildForecastPrompt(analytics, dailyClicks)
	system := "You are a data analyst. Write a brief 2-3 sentence narrative forecast based on the click trend data. Be specific with numbers."
	return s.generateAndSave(shortCode, userID, insightType, system, prompt, analytics.TotalClicks)
}

func (s *AIInsightsService) GetAudienceProfile(shortCode string, userID int64) (*models.AIInsight, error) {
	const insightType = "audience"
	if cached, ok := s.getCached(shortCode, insightType); ok {
		return cached, nil
	}
	analytics, err := s.analytics.GetAdvancedAnalytics(shortCode, 30)
	if err != nil {
		return nil, fmt.Errorf("analytics fetch failed: %w", err)
	}
	referrers, _ := s.analytics.GetReferrerAnalytics(shortCode, 30)
	prompt := s.buildAudiencePrompt(analytics, referrers)
	system := `You are an audience analyst. Respond ONLY with valid JSON, no markdown.
Format: [{"category": "...", "percentage": 0, "top_domains": ["..."]}]`
	return s.generateAndSave(shortCode, userID, insightType, system, prompt, analytics.TotalClicks)
}

// ─── Cache ────────────────────────────────────────────────────────────────────

func (s *AIInsightsService) getCached(shortCode, insightType string) (*models.AIInsight, bool) {
	key := fmt.Sprintf("ai:%s:%s", insightType, shortCode)

	// Try Redis first
	if data, err := s.redis.Get(key); err == nil {
		var insight models.AIInsight
		if json.Unmarshal([]byte(data), &insight) == nil {
			if !s.shouldRegenerate(&insight, shortCode) {
				return &insight, true
			}
		}
	}

	// Fallback to DB
	dbInsight, err := s.db.GetAIInsight(shortCode, insightType)
	if err != nil || dbInsight == nil {
		return nil, false
	}
	if s.shouldRegenerate(dbInsight, shortCode) {
		return nil, false
	}
	s.cacheInsight(key, dbInsight)
	return dbInsight, true
}

func (s *AIInsightsService) shouldRegenerate(insight *models.AIInsight, shortCode string) bool {
	var snapshot map[string]int64
	if err := json.Unmarshal(insight.DataSnapshot, &snapshot); err != nil {
		return false
	}
	stored := snapshot["click_count"]
	if stored == 0 {
		return false
	}
	mapping, err := s.db.GetURLMappingByShortCode(shortCode)
	if err != nil {
		return false
	}
	diff := mapping.ClickCount - stored
	return diff >= 100 || float64(diff)/float64(stored) >= 0.10
}

func (s *AIInsightsService) cacheInsight(key string, insight *models.AIInsight) {
	if data, err := json.Marshal(insight); err == nil {
		_ = s.redis.Set(key, string(data), insightCacheTTL)
	}
}

// ─── Core generation ──────────────────────────────────────────────────────────

// stripCodeFences removes markdown code-fence wrappers (```json ... ```) that
// LLMs sometimes add even when told not to.
func stripCodeFences(s string) string {
	s = strings.TrimSpace(s)
	// Remove leading ```json or ``` fence
	for _, prefix := range []string{"```json", "```JSON", "```"} {
		if strings.HasPrefix(s, prefix) {
			s = strings.TrimPrefix(s, prefix)
			break
		}
	}
	// Remove trailing ```
	if strings.HasSuffix(s, "```") {
		s = strings.TrimSuffix(s, "```")
	}
	return strings.TrimSpace(s)
}

func (s *AIInsightsService) generateAndSave(shortCode string, userID int64, insightType, system, userPrompt string, clickCount int64) (*models.AIInsight, error) {
	// 120s to account for semaphore queue (up to 4 prior requests × ~20s each)
	ctx, cancel := context.WithTimeout(context.Background(), 120*time.Second)
	defer cancel()

	raw, err := s.callOpenRouter(ctx, system, userPrompt)
	if err != nil {
		return nil, err
	}
	content := stripCodeFences(raw)

	now := time.Now()
	expires := now.Add(24 * time.Hour)
	snapshot, _ := json.Marshal(map[string]int64{"click_count": clickCount})

	insight := &models.AIInsight{
		ShortCode:    shortCode,
		UserID:       &userID,
		InsightType:  insightType,
		Content:      content,
		DataSnapshot: snapshot,
		GeneratedAt:  now,
		ExpiresAt:    &expires,
	}

	// Persist to DB (best-effort)
	_ = s.db.SaveAIInsight(insight)

	// Cache in Redis
	cacheKey := fmt.Sprintf("ai:%s:%s", insightType, shortCode)
	s.cacheInsight(cacheKey, insight)

	return insight, nil
}

// ─── OpenRouter API ───────────────────────────────────────────────────────────

type openRouterRequest struct {
	Model     string        `json:"model"`
	Messages  []chatMessage `json:"messages"`
	MaxTokens int           `json:"max_tokens"`
}

type chatMessage struct {
	Role    string `json:"role"`
	Content string `json:"content"`
}

type openRouterResponse struct {
	Choices []struct {
		Message struct {
			Content string `json:"content"`
		} `json:"message"`
	} `json:"choices"`
	Error *struct {
		Message string `json:"message"`
	} `json:"error,omitempty"`
}

func (s *AIInsightsService) callOpenRouter(ctx context.Context, system, userPrompt string) (string, error) {
	if s.apiKey == "" {
		log.Println("[AI Insights] ERROR: OPENROUTER_API_KEY is empty")
		return "", fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	// ── Acquire semaphore ────────────────────────────────────────────────────
	// Serialises all OpenRouter calls to 1-at-a-time.  The dashboard fires 5
	// requests concurrently; without this they all hit the free-tier rate limit.
	select {
	case s.sem <- struct{}{}:
		defer func() { <-s.sem }()
	case <-ctx.Done():
		return "", ctx.Err()
	}

	// ── Build payload ────────────────────────────────────────────────────────
	// Gemma models do not support the "system" role (400 INVALID_ARGUMENT).
	// Merge system instructions into the user message instead.
	mergedUser := system + "\n\n" + userPrompt

	payload := openRouterRequest{
		Model:     openRouterModel,
		MaxTokens: insightMaxTokens,
		Messages: []chatMessage{
			{Role: "user", Content: mergedUser},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
	}

	log.Printf("[AI Insights] Calling OpenRouter model=%s", openRouterModel)

	var lastErr error
	for attempt := 0; attempt < openRouterMaxRetries; attempt++ {
		if attempt > 0 {
			// Exponential backoff with jitter to help when 5 requests hit at once
			jitter := time.Duration(rand.Intn(1000)) * time.Millisecond
			delay := (openRouterRetryBaseDelay * time.Duration(attempt)) + jitter
			log.Printf("[AI Insights] Retry %d after rate-limit... waiting %v", attempt, delay)
			select {
			case <-time.After(delay):
			case <-ctx.Done():
				return "", ctx.Err()
			}
		}

		req, err := http.NewRequestWithContext(ctx, http.MethodPost, openRouterURL, bytes.NewReader(body))
		if err != nil {
			return "", fmt.Errorf("failed to build request: %w", err)
		}
		req.Header.Set("Content-Type", "application/json")
		req.Header.Set("Authorization", "Bearer "+s.apiKey)
		req.Header.Set("HTTP-Referer", "https://trunc.site")
		req.Header.Set("X-Title", "URLshortn AI Insights")

		resp, err := s.httpClient.Do(req)
		if err != nil {
			log.Printf("[AI Insights] HTTP error: %v", err)
			lastErr = fmt.Errorf("openrouter request failed: %w", err)
			continue
		}

		respBody, err := io.ReadAll(resp.Body)
		resp.Body.Close()
		if err != nil {
			return "", fmt.Errorf("failed to read response: %w", err)
		}

		log.Printf("[AI Insights] OpenRouter HTTP status=%d", resp.StatusCode)

		// Retry on 429 rate-limit
		if resp.StatusCode == http.StatusTooManyRequests {
			log.Printf("[AI Insights] Rate-limited (429), will retry. Body: %s", string(respBody))
			lastErr = fmt.Errorf("openrouter rate-limited")
			continue
		}

		var result openRouterResponse
		if err := json.Unmarshal(respBody, &result); err != nil {
			return "", fmt.Errorf("failed to parse response: %w\nbody: %s", err, string(respBody))
		}

		if result.Error != nil {
			log.Printf("[AI Insights] OpenRouter API error (status=%d): %s | raw: %s", resp.StatusCode, result.Error.Message, string(respBody))
			return "", fmt.Errorf("openrouter error: %s", result.Error.Message)
		}

		if len(result.Choices) == 0 {
			log.Printf("[AI Insights] No choices returned. Full body: %s", string(respBody))
			return "", fmt.Errorf("no choices in openrouter response")
		}

		return strings.TrimSpace(result.Choices[0].Message.Content), nil
	}

	return "", lastErr
}

// ─── Prompt builders ──────────────────────────────────────────────────────────

func (s *AIInsightsService) buildSummaryPrompt(a *models.AdvancedAnalyticsResponse) string {
	topCountry := "Unknown"
	if len(a.Geographic.Countries) > 0 {
		topCountry = a.Geographic.Countries[0].CountryName
	}
	topDevice := "Unknown"
	if len(a.DeviceData.DeviceTypes) > 0 {
		topDevice = a.DeviceData.DeviceTypes[0].DeviceType
	}
	topReferrer := "direct"
	if len(a.Referrers) > 0 && a.Referrers[0].Referrer != "" {
		topReferrer = a.Referrers[0].Referrer
	}

	var lines []string
	lines = append(lines, "Link performance data (last 30 days):")
	lines = append(lines, fmt.Sprintf("- Total clicks: %d", a.TotalClicks))
	lines = append(lines, fmt.Sprintf("- Top country: %s", topCountry))
	lines = append(lines, fmt.Sprintf("- Top device: %s", topDevice))
	lines = append(lines, fmt.Sprintf("- Top referrer: %s", topReferrer))
	if a.TimeData.PeakTimes.MaxClicks > 0 {
		lines = append(lines, fmt.Sprintf("- Best hour: %d:00", a.TimeData.PeakTimes.PeakHour))
		lines = append(lines, fmt.Sprintf("- Best day: %s", a.TimeData.PeakTimes.PeakDay))
	}
	return strings.Join(lines, "\n")
}

func (s *AIInsightsService) buildAnomalyPrompt(a *models.AdvancedAnalyticsResponse) string {
	avgPerHour := 0.0
	if a.TotalClicks > 0 {
		avgPerHour = float64(a.TotalClicks) / (30 * 24)
	}
	topReferrer := "direct"
	if len(a.Referrers) > 0 {
		topReferrer = a.Referrers[0].Referrer
	}
	topCountry := "Unknown"
	if len(a.Geographic.Countries) > 0 {
		topCountry = a.Geographic.Countries[0].CountryName
	}
	return fmt.Sprintf(`Link metrics (last 30 days):
- Total clicks: %d
- Avg clicks/hour: %.2f
- Top country: %s
- Top referrer: %s

Output ONLY a raw JSON object (no markdown, no explanation outside JSON):
{"detected":true/false,"severity":"low|medium|high","explanation":"one sentence","possible_cause":"one sentence"}`,
		a.TotalClicks, avgPerHour, topCountry, topReferrer,
	)
}

func (s *AIInsightsService) buildRecommendationsPrompt(a *models.AdvancedAnalyticsResponse) string {
	topCountry := "Unknown"
	if len(a.Geographic.Countries) > 0 {
		topCountry = a.Geographic.Countries[0].CountryName
	}
	mobilePct, desktopPct := 0.0, 0.0
	for _, d := range a.DeviceData.DeviceTypes {
		switch d.DeviceType {
		case "mobile":
			mobilePct = d.Percentage
		case "desktop":
			desktopPct = d.Percentage
		}
	}
	topReferrer := "direct"
	if len(a.Referrers) > 0 {
		topReferrer = a.Referrers[0].Referrer
	}
	peakInfo := ""
	if a.TimeData.PeakTimes.MaxClicks > 0 {
		peakInfo = fmt.Sprintf("\n- Best time: %s at %d:00", a.TimeData.PeakTimes.PeakDay, a.TimeData.PeakTimes.PeakHour)
	}
	return fmt.Sprintf(`Link stats (30 days): clicks=%d, country=%s, mobile=%.0f%%, desktop=%.0f%%, referrer=%s%s

Give EXACTLY 3 short bullet points (one sentence each, no sub-bullets) to improve click-through rate. Use "- " prefix. No intro, no conclusion.`,
		a.TotalClicks, topCountry, mobilePct, desktopPct, topReferrer, peakInfo)
}

func (s *AIInsightsService) buildForecastPrompt(a *models.AdvancedAnalyticsResponse, dailyClicks []int64) string {
	slope, intercept := linearRegression(dailyClicks)
	n := len(dailyClicks)

	projected := make([]int64, 7)
	for i := range projected {
		v := slope*float64(n+i) + intercept
		if v < 0 {
			v = 0
		}
		projected[i] = int64(v)
	}

	var projected30 int64
	for i := 0; i < 30; i++ {
		v := slope*float64(n+i) + intercept
		if v > 0 {
			projected30 += int64(v)
		}
	}

	return fmt.Sprintf(`Traffic data: total clicks=%d over 30 days.
Linear projection: next 7 days=%v, next 30-day total=%d.

Write 2 sentences: one summarising the trend, one giving the 30-day projection. Use specific numbers. No bullet points, no headers.`,
		a.TotalClicks, projected, projected30)
}

func (s *AIInsightsService) buildAudiencePrompt(a *models.AdvancedAnalyticsResponse, referrers []models.ReferrerStat) string {
	var sb strings.Builder
	sb.WriteString("Referrer domains by clicks:\n")
	limit := 8
	if len(referrers) < limit {
		limit = len(referrers)
	}
	for _, r := range referrers[:limit] {
		sb.WriteString(fmt.Sprintf("- %s: %d clicks\n", r.Referrer, r.Clicks))
	}
	if limit == 0 {
		sb.WriteString("- (no referrer data)\n")
	}
	sb.WriteString(`
Output ONLY a raw JSON array (no markdown fences, no explanation):
[{"category":"...","percentage":0,"top_domains":["..."]}]
Use 2-4 segments. percentages must sum to 100.`)
	return sb.String()
}

// ─── Helpers ──────────────────────────────────────────────────────────────────

func (s *AIInsightsService) getDailyClicks(shortCode string, days int) ([]int64, error) {
	query := fmt.Sprintf(`
		SELECT COALESCE(SUM(clicks), 0)
		FROM time_analytics
		WHERE short_code = $1
		  AND click_date >= CURRENT_DATE - INTERVAL '%d days'
		GROUP BY click_date
		ORDER BY click_date ASC
	`, days)

	rows, err := s.db.Query(query, shortCode)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var result []int64
	for rows.Next() {
		var c int64
		if rows.Scan(&c) == nil {
			result = append(result, c)
		}
	}
	return result, nil
}

func (s *AIInsightsService) extractMonthlyAsDaily(a *models.AdvancedAnalyticsResponse) []int64 {
	var result []int64
	for _, m := range a.TimeData.MonthlyTrend {
		result = append(result, m.Clicks)
	}
	return result
}

// linearRegression returns slope and intercept for a simple OLS fit.
func linearRegression(data []int64) (slope, intercept float64) {
	n := float64(len(data))
	if n < 2 {
		return 0, 0
	}
	var sumX, sumY, sumXY, sumX2 float64
	for i, v := range data {
		x := float64(i)
		y := float64(v)
		sumX += x
		sumY += y
		sumXY += x * y
		sumX2 += x * x
	}
	denom := n*sumX2 - sumX*sumX
	if denom == 0 {
		return 0, sumY / n
	}
	slope = (n*sumXY - sumX*sumY) / denom
	intercept = (sumY - slope*sumX) / n
	return
}
