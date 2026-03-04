package services

import (
	"bytes"
	"context"
	"encoding/json"
	"fmt"
	"io"
	"net/http"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

const (
	openRouterURL    = "https://openrouter.ai/api/v1/chat/completions"
	openRouterModel  = "google/gemma-3-4b-it:free"
	insightCacheTTL  = time.Hour
	insightMaxTokens = 512
)

// AIInsightsService generates AI-powered insights for short links using OpenRouter.
type AIInsightsService struct {
	apiKey    string
	db        *storage.PostgresStorage
	redis     *storage.RedisStorage
	analytics *AdvancedAnalyticsService
	httpClient *http.Client
}

// NewAIInsightsService creates a new AIInsightsService.
func NewAIInsightsService(apiKey string, db *storage.PostgresStorage, redis *storage.RedisStorage, analytics *AdvancedAnalyticsService) *AIInsightsService {
	return &AIInsightsService{
		apiKey:    apiKey,
		db:        db,
		redis:     redis,
		analytics: analytics,
		httpClient: &http.Client{Timeout: 35 * time.Second},
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

func (s *AIInsightsService) generateAndSave(shortCode string, userID int64, insightType, system, userPrompt string, clickCount int64) (*models.AIInsight, error) {
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	content, err := s.callOpenRouter(ctx, system, userPrompt)
	if err != nil {
		return nil, err
	}

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
	Model     string          `json:"model"`
	Messages  []chatMessage   `json:"messages"`
	MaxTokens int             `json:"max_tokens"`
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
		return "", fmt.Errorf("OPENROUTER_API_KEY not configured")
	}

	payload := openRouterRequest{
		Model:     openRouterModel,
		MaxTokens: insightMaxTokens,
		Messages: []chatMessage{
			{Role: "system", Content: system},
			{Role: "user", Content: userPrompt},
		},
	}

	body, err := json.Marshal(payload)
	if err != nil {
		return "", fmt.Errorf("failed to marshal request: %w", err)
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
		return "", fmt.Errorf("openrouter request failed: %w", err)
	}
	defer resp.Body.Close()

	respBody, err := io.ReadAll(resp.Body)
	if err != nil {
		return "", fmt.Errorf("failed to read response: %w", err)
	}

	var result openRouterResponse
	if err := json.Unmarshal(respBody, &result); err != nil {
		return "", fmt.Errorf("failed to parse response: %w", err)
	}

	if result.Error != nil {
		return "", fmt.Errorf("openrouter error: %s", result.Error.Message)
	}

	if len(result.Choices) == 0 {
		return "", fmt.Errorf("no choices in openrouter response")
	}

	return strings.TrimSpace(result.Choices[0].Message.Content), nil
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
	return fmt.Sprintf(`Analyze these link metrics for anomalies:
- Total clicks (30 days): %d
- Average clicks/hour (30-day rolling): %.2f
- Top country: %s
- Top referrer: %s

Respond with JSON only.`,
		a.TotalClicks, avgPerHour,
		func() string {
			if len(a.Geographic.Countries) > 0 {
				return a.Geographic.Countries[0].CountryName
			}
			return "Unknown"
		}(),
		topReferrer,
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

	var lines []string
	lines = append(lines, "Link performance data for optimisation advice:")
	lines = append(lines, fmt.Sprintf("- Total clicks: %d", a.TotalClicks))
	lines = append(lines, fmt.Sprintf("- Top country: %s", topCountry))
	lines = append(lines, fmt.Sprintf("- Mobile visitors: %.1f%%", mobilePct))
	lines = append(lines, fmt.Sprintf("- Desktop visitors: %.1f%%", desktopPct))
	lines = append(lines, fmt.Sprintf("- Top referrer: %s", topReferrer))
	if a.TimeData.PeakTimes.MaxClicks > 0 {
		lines = append(lines, fmt.Sprintf("- Best posting hour: %d:00", a.TimeData.PeakTimes.PeakHour))
		lines = append(lines, fmt.Sprintf("- Best posting day: %s", a.TimeData.PeakTimes.PeakDay))
	}
	return strings.Join(lines, "\n")
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

	// Format last 14 daily values for the prompt
	show := dailyClicks
	if len(show) > 14 {
		show = show[len(show)-14:]
	}
	var dataStr strings.Builder
	for i, c := range show {
		dataStr.WriteString(fmt.Sprintf("Day %d: %d clicks\n", len(dailyClicks)-len(show)+i+1, c))
	}

	return fmt.Sprintf(`Click trend data for the last 30 days (last 14 shown):
%s
Server-computed linear projection:
- Next 7 days: %v
- Projected 30-day total: %d

Write a brief forecast narrative with specific numbers.`,
		dataStr.String(), projected, projected30)
}

func (s *AIInsightsService) buildAudiencePrompt(a *models.AdvancedAnalyticsResponse, referrers []models.ReferrerStat) string {
	var sb strings.Builder
	sb.WriteString("Top referrer domains by clicks:\n")
	limit := 10
	if len(referrers) < limit {
		limit = len(referrers)
	}
	for _, r := range referrers[:limit] {
		sb.WriteString(fmt.Sprintf("- %s: %d clicks\n", r.Referrer, r.Clicks))
	}
	if limit == 0 {
		sb.WriteString("- (no referrer data yet)\n")
	}
	sb.WriteString("\nClassify these domains into audience segments. Respond with JSON only.")
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
