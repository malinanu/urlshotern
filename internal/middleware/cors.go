package middleware

import (
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORSMiddleware configures CORS for the API
func CORSMiddleware() gin.HandlerFunc {
	config := cors.DefaultConfig()
	config.AllowOrigins = []string{
		"http://localhost:3000",    // Next.js development server
		"http://localhost:3001",    // Alternative port
		"https://trunc.site",       // Production frontend
		"https://www.trunc.site",   // Production frontend with www
		"https://api.trunc.site",   // Production API subdomain
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

	return cors.New(config)
}