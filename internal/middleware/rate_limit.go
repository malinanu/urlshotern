package middleware

import (
	"fmt"
	"net/http"
	"sync"
	"time"

	"github.com/gin-gonic/gin"
	"golang.org/x/time/rate"
)

// RateLimitConfig defines rate limiting configuration
type RateLimitConfig struct {
	RequestsPerSecond float64
	Burst             int
	Duration          time.Duration
}

// DefaultRateLimit provides default rate limiting
var DefaultRateLimit = RateLimitConfig{
	RequestsPerSecond: 10, // 10 requests per second
	Burst:             20, // Allow burst of 20
	Duration:          time.Hour,
}

// AuthRateLimit for authentication endpoints
var AuthRateLimit = RateLimitConfig{
	RequestsPerSecond: 5,  // 5 requests per second
	Burst:             10, // Allow burst of 10
	Duration:          time.Hour,
}

// OTPRateLimit for OTP endpoints (more restrictive)
var OTPRateLimit = RateLimitConfig{
	RequestsPerSecond: 1, // 1 request per second
	Burst:             3, // Allow burst of 3
	Duration:          time.Minute * 5,
}

type visitor struct {
	limiter  *rate.Limiter
	lastSeen time.Time
}

var (
	visitors   = make(map[string]*visitor)
	visitorsMu sync.Mutex
)

// RateLimitMiddleware creates rate limiting middleware
func RateLimitMiddleware(config RateLimitConfig) gin.HandlerFunc {
	// Cleanup old visitors periodically
	go cleanupVisitors()

	return func(c *gin.Context) {
		// Prefer CF-Connecting-IP (cannot be spoofed through Cloudflare).
		// Fallback to Gin's ClientIP which respects SetTrustedProxies.
		ip := c.GetHeader("CF-Connecting-IP")
		if ip == "" {
			ip = c.ClientIP()
		}

		// Guard the shared visitors map — it is read/written from every request
		// goroutine and the cleanup goroutine concurrently.
		visitorsMu.Lock()
		v, exists := visitors[ip]
		if !exists {
			limiter := rate.NewLimiter(rate.Limit(config.RequestsPerSecond), config.Burst)
			v = &visitor{limiter, time.Now()}
			visitors[ip] = v
		}
		v.lastSeen = time.Now()
		visitorsMu.Unlock()

		if !v.limiter.Allow() {
			c.JSON(http.StatusTooManyRequests, gin.H{
				"error":   "Rate limit exceeded",
				"message": fmt.Sprintf("Too many requests. Limit: %.0f requests per second", config.RequestsPerSecond),
			})
			c.Abort()
			return
		}

		c.Next()
	}
}

// cleanupVisitors removes old visitors to prevent memory leaks
func cleanupVisitors() {
	for {
		time.Sleep(time.Minute)
		visitorsMu.Lock()
		for ip, v := range visitors {
			if time.Since(v.lastSeen) > time.Hour {
				delete(visitors, ip)
			}
		}
		visitorsMu.Unlock()
	}
}
