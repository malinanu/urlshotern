package middleware

import (
	"strings"

	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS for the API
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",  // Next.js development server
		"http://localhost:3001",  // Alternative port
		"http://localhost:8080",  // Static frontend served from Go backend
		"https://trunc.site",     // Production frontend
		"https://www.trunc.site", // Production frontend with www
		"https://api.trunc.site", // Production API subdomain
	}
	config.AllowMethods = []string{"GET", "POST", "PUT", "PATCH", "DELETE", "OPTIONS"}
	config.AllowHeaders = []string{
		"Origin",
		"Content-Type",
		"Accept",
		"Authorization",
		"Cache-Control",
		"X-Requested-With",
	}
	config.AllowCredentials = true

	handler := cors.New(config)
	return func(c *gin.Context) {
		// Public page assets (self-hosted fonts) are fetched in CORS mode by any
		// page, including the builder's sandboxed preview (Origin: null). They
		// carry no credentials and set Access-Control-Allow-Origin: * themselves.
		if strings.HasPrefix(c.Request.URL.Path, "/m-static/") {
			c.Next()
			return
		}
		handler(c)
	}
}

// SecurityHeadersMiddleware adds security headers to all responses, including
// HSTS to enforce HTTPS on the main domain and all subdomains.
func SecurityHeadersMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		c.Header("Strict-Transport-Security", "max-age=31536000; includeSubDomains")
		c.Next()
	}
}
