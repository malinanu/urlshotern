package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"net/http/pprof"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/routes"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

// ProfilingServer runs the application with profiling enabled
func main() {
	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize storage layers
	db, err := storage.NewPostgresStorage(config)
	if err != nil {
		log.Fatalf("Failed to connect to database: %v", err)
	}
	defer db.Close()

	// Initialize GORM database for billing
	sslMode := config.DBSSLMode
	if sslMode == "" {
		sslMode = "disable"
	}
	dsn := fmt.Sprintf("host=%s port=%s user=%s password=%s dbname=%s sslmode=%s",
		config.DBHost, config.DBPort, config.DBUser, config.DBPassword, config.DBName, sslMode)
	gormDB, err := gorm.Open(postgres.Open(dsn), &gorm.Config{})
	if err != nil {
		log.Fatalf("Failed to connect to GORM database: %v", err)
	}

	redis, err := storage.NewRedisStorage(config)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Initialize services
	shortenerService := services.NewShortenerService(db, redis, config)
	analyticsService := services.NewAnalyticsService(db)
	advancedAnalyticsService := services.NewAdvancedAnalyticsService(db)
	var userAnalyticsService *services.UserAnalyticsService

	// Initialize auth-related services
	smsService := services.NewSMSService(db, config)
	emailService := services.NewEmailService(db, config)
	jwtService := services.NewJWTService(config.JWTSecret, config.JWTIssuer, time.Hour*24, time.Hour*24*7)
	userService := services.NewUserService(db, redis, jwtService, smsService, emailService, nil)
	authService := services.NewAuthService(userService, jwtService, smsService, emailService, db, redis, config)
	rbacService := services.NewRBACServiceImpl(db, redis)
	conversionTrackingService := services.NewConversionTrackingService(db)
	abTestingService := services.NewABTestingService(db, redis)
	realtimeAnalyticsService := services.NewRealtimeAnalyticsService(db, redis, analyticsService)
	attributionService := services.NewAttributionService(db, conversionTrackingService)
	adminService := services.NewAdminService(db, rbacService)
	botDetectionService := services.NewBotDetectionService(redis)
	aiInsightsService := services.NewAIInsightsService(config.OpenRouterAPIKey, db, redis, advancedAnalyticsService)
	brandingService := services.NewBrandingService(db)
	domainService := services.NewDomainService(db, redis, config, analyticsService, userService)

	shortenerService.SetRealtimeService(realtimeAnalyticsService)
	shortenerService.SetAttributionService(attributionService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService, rbacService, redis)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, adminService, smsService, emailService, config, nil, nil)
	analyticsHandlers := handlers.NewAnalyticsHandlers(userAnalyticsService)
	adminRBACHandlers := handlers.NewAdminRBACHandlers(adminService, rbacService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db, gormDB, config, adminRBACHandlers, botDetectionService, aiInsightsService, domainService, brandingService, nil, nil)

	// Setup Gin router with metrics middleware
	gin.SetMode(gin.ReleaseMode)
	router := gin.New()
	router.Use(gin.Recovery())
	router.Use(middleware.MetricsMiddleware()) // Add metrics collection

	// Setup all routes
	routes.SetupRoutes(router, handler, authMiddleware, config)

	// Add profiling endpoints
	router.GET("/debug/pprof/", gin.WrapF(pprof.Index))
	router.GET("/debug/pprof/cmdline", gin.WrapF(pprof.Cmdline))
	router.GET("/debug/pprof/profile", gin.WrapF(pprof.Profile))
	router.GET("/debug/pprof/symbol", gin.WrapF(pprof.Symbol))
	router.GET("/debug/pprof/trace", gin.WrapF(pprof.Trace))
	router.GET("/debug/pprof/heap", gin.WrapF(pprof.Handler("heap").ServeHTTP))
	router.GET("/debug/pprof/goroutine", gin.WrapF(pprof.Handler("goroutine").ServeHTTP))
	router.GET("/debug/pprof/block", gin.WrapF(pprof.Handler("block").ServeHTTP))
	router.GET("/debug/pprof/mutex", gin.WrapF(pprof.Handler("mutex").ServeHTTP))
	router.GET("/debug/pprof/allocs", gin.WrapF(pprof.Handler("allocs").ServeHTTP))

	// Add metrics endpoint
	router.GET("/debug/metrics", func(c *gin.Context) {
		collector := middleware.GetMetricsCollector()
		stats := collector.GetStats()

		// Add runtime metrics
		var m runtime.MemStats
		runtime.ReadMemStats(&m)

		stats["runtime"] = map[string]interface{}{
			"goroutines":    runtime.NumGoroutine(),
			"heap_alloc_mb": float64(m.Alloc) / 1024 / 1024,
			"heap_sys_mb":   float64(m.HeapSys) / 1024 / 1024,
			"heap_objects":  m.HeapObjects,
			"gc_runs":       m.NumGC,
			"last_gc":       time.Unix(0, int64(m.LastGC)), //nolint:gosec // ns since epoch; fits int64 until 2262
			"gc_pause_ns":   m.PauseNs[(m.NumGC+255)%256],
		}

		c.JSON(http.StatusOK, stats)
	})

	// Setup server
	srv := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	// Start goroutine counter
	go func() {
		ticker := time.NewTicker(5 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			count := runtime.NumGoroutine()
			log.Printf("[PROFILING] Active goroutines: %d", count)
		}
	}()

	// Start memory stats logger
	go func() {
		ticker := time.NewTicker(10 * time.Second)
		defer ticker.Stop()
		for range ticker.C {
			var m runtime.MemStats
			runtime.ReadMemStats(&m)
			log.Printf("[PROFILING] Heap Alloc: %.2f MB, Sys: %.2f MB, NumGC: %d",
				float64(m.Alloc)/1024/1024,
				float64(m.Sys)/1024/1024,
				m.NumGC,
			)
		}
	}()

	// Start server in a goroutine
	go func() {
		log.Printf("Profiling server starting on port %s", config.ServerPort)
		log.Printf("pprof endpoints available at http://localhost:%s/debug/pprof/", config.ServerPort)
		log.Printf("Metrics endpoint available at http://localhost:%s/debug/metrics", config.ServerPort)
		log.Println("")
		log.Println("Profiling commands:")
		log.Printf("  CPU:       curl http://localhost:%s/debug/pprof/profile?seconds=30 -o cpu.prof", config.ServerPort)
		log.Printf("  Heap:      curl http://localhost:%s/debug/pprof/heap -o heap.prof", config.ServerPort)
		log.Printf("  Goroutine: curl http://localhost:%s/debug/pprof/goroutine -o goroutine.prof", config.ServerPort)
		log.Printf("  Metrics:   curl http://localhost:%s/debug/metrics", config.ServerPort)
		log.Println("")

		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down profiling server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown real-time analytics service
	realtimeAnalyticsService.Shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Profiling server exited")
}
