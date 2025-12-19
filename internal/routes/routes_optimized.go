package routes

import (
	"github.com/gin-gonic/gin"

	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
)

// SetupOptimizedRoutes configures routes with optimized middleware order
// Middleware execution order is critical for performance:
// 1. Recovery (prevents panics from affecting other middleware)
// 2. RequestID (needed for logging correlation)
// 3. Compression (reduces response size early)
// 4. CORS (authentication-independent, can be checked early)
// 5. Logging (logs with request ID and compressed sizes)
// 6. RateLimit (prevents excessive work on rate-limited requests)
// 7. Auth (only authenticate valid, non-rate-limited requests)
func SetupOptimizedRoutes(
	router *gin.Engine,
	handler *handlers.Handler,
	authMiddleware *middleware.AuthMiddleware,
) {
	// 1. Recovery middleware (built-in, should be first)
	router.Use(gin.Recovery())

	// 2. Request ID middleware (for tracing/correlation)
	router.Use(middleware.RequestIDMiddleware())

	// 3. Compression middleware (reduces bandwidth early)
	router.Use(middleware.CompressionMiddleware())

	// 4. CORS middleware (allows cross-origin requests)
	router.Use(middleware.CORSMiddleware())

	// 5. Logging middleware (logs with request ID)
	router.Use(middleware.LoggingMiddleware())

	// Health check (no additional middleware needed)
	router.GET("/health", handler.HealthCheck)

	// Public routes (no authentication required)
	setupPublicRoutes(router, handler, authMiddleware)

	// Authentication routes
	setupAuthRoutes(router, handler.AuthHandlers, authMiddleware)

	// Protected routes (authentication required)
	setupProtectedRoutes(router, handler, authMiddleware)

	// Static file serving for frontend (must be last)
	setupStaticRoutes(router)
}

// Performance notes:
// - Middleware is executed in order for every request
// - Place cheap middleware (RequestID, Compression) before expensive ones (Auth, DB queries)
// - Use rate limiting before authentication to save CPU on invalid requests
// - Compression should be early to benefit all subsequent responses
// - CORS is lightweight and should be checked before heavier operations
