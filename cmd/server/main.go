package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/routes"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
)

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

	redis, err := storage.NewRedisStorage(config)
	if err != nil {
		log.Fatalf("Failed to connect to Redis: %v", err)
	}
	defer redis.Close()

	// Initialize services
	shortenerService := services.NewShortenerService(db, redis, config)
	analyticsService := services.NewAnalyticsService(db)
	advancedAnalyticsService := services.NewAdvancedAnalyticsService(db)
	// userAnalyticsService := services.NewUserAnalyticsService(db)  // Temporarily disabled
	var userAnalyticsService *services.UserAnalyticsService // Placeholder

	// Initialize auth-related services
	smsService := services.NewSMSService(db, config)
	emailService := services.NewEmailService(db, config)
	jwtService := services.NewJWTService(config.JWTSecret, config.JWTIssuer, time.Hour*24, time.Hour*24*7)
	userService := services.NewUserService(db, redis, jwtService, smsService, emailService, nil)
	authService := services.NewAuthService(userService, jwtService, smsService, emailService, db, redis, config)
	rbacService := services.NewRBACService(db, redis)
	conversionTrackingService := services.NewConversionTrackingService(db)
	abTestingService := services.NewABTestingService(db, redis)
	realtimeAnalyticsService := services.NewRealtimeAnalyticsService(db, redis, analyticsService)
	attributionService := services.NewAttributionService(db, conversionTrackingService)
	// cmsService := services.NewCMSService(db)  // Temporarily disabled
	// apiKeyService := services.NewAPIKeyService(db)  // Temporarily disabled
	
	// Set real-time service on shortener for click broadcasting
	shortenerService.SetRealtimeService(realtimeAnalyticsService)
	
	// Set attribution service on shortener for attribution tracking
	shortenerService.SetAttributionService(attributionService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService, rbacService)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, smsService, emailService)
	analyticsHandlers := handlers.NewAnalyticsHandlers(userAnalyticsService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db)

	// Setup Gin router
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup all routes
	routes.SetupRoutes(router, handler, authMiddleware)

	// Setup server
	srv := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		log.Printf("Server starting on port %s", config.ServerPort)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			log.Fatalf("Failed to start server: %v", err)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit
	log.Println("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown real-time analytics service
	realtimeAnalyticsService.Shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}