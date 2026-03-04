package services

import (
	"crypto/sha256"
	"encoding/hex"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/storage"
)

// botSignals contains UA substrings that identify known bots/crawlers.
var botSignals = []string{
	"bot", "crawler", "spider", "scraper",
	"curl", "wget", "python-requests", "go-http-client",
	"java/", "libwww", "jakarta", "googlebot", "bingbot",
	"slurp", "duckduckbot", "baiduspider", "yandexbot",
	"facebookexternalhit", "twitterbot", "linkedinbot",
	"semrushbot", "ahrefsbot", "mj12bot", "dotbot",
}

type BotDetectionService struct {
	redis *storage.RedisStorage
}

func NewBotDetectionService(redis *storage.RedisStorage) *BotDetectionService {
	return &BotDetectionService{redis: redis}
}

// IsBot returns true if the request looks like a bot.
// Signals checked: known UA strings, empty UA, missing Accept-Language,
// missing Accept header, and same-IP rate > 10 clicks/60s on same short code.
func (b *BotDetectionService) IsBot(userAgent, acceptLang, accept, ip, shortCode string) bool {
	// Empty user agent
	if strings.TrimSpace(userAgent) == "" {
		return true
	}

	// UA matches known bot strings
	ua := strings.ToLower(userAgent)
	for _, sig := range botSignals {
		if strings.Contains(ua, sig) {
			return true
		}
	}

	// Missing both Accept-Language and Accept headers (browsers always send these)
	if acceptLang == "" && accept == "" {
		return true
	}

	// Rate check: > 10 clicks from same IP on same short code in 60 seconds
	if b.redis != nil {
		count, err := b.redis.IncrBotRateCheck(ip, shortCode)
		if err == nil && count > 10 {
			return true
		}
	}

	return false
}

// GenerateIPHash returns SHA-256(ip + daily_salt) for privacy-preserving unique visitor tracking.
func (b *BotDetectionService) GenerateIPHash(ip string) string {
	dailySalt := time.Now().UTC().Format("2006-01-02")
	hash := sha256.Sum256([]byte(ip + dailySalt))
	return hex.EncodeToString(hash[:])
}

// IsUniqueVisitor returns true if this is the first visit for ipHash+shortCode today.
func (b *BotDetectionService) IsUniqueVisitor(ipHash, shortCode string) bool {
	if b.redis == nil {
		return false
	}
	unique, err := b.redis.CheckAndSetUniqueVisitor(ipHash, shortCode)
	if err != nil {
		return false
	}
	return unique
}
