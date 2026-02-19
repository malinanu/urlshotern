package main

import (
	"context"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"strings"
	"syscall"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/routes"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/workers"
	"github.com/gin-gonic/gin"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
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

	if err := db.SeedDefaultStaticPages(); err != nil {
		log.Printf("Warning: failed to seed default static pages: %v", err)
	}

	// Initialize GORM database for billing and Polar integration
	sslMode := config.DBSSLMode
	if sslMode == "" {
		sslMode = "disable" // Default to disable if not specified
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
	// userAnalyticsService := services.NewUserAnalyticsService(db)  // Temporarily disabled
	var userAnalyticsService *services.UserAnalyticsService // Placeholder

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
	globalSettingsService := services.NewGlobalSettingsService(db)
	contentBlockService := services.NewContentBlockService(db, globalSettingsService)
	pageTemplateService := services.NewPageTemplateService(db, contentBlockService, globalSettingsService)
	// cmsService := services.NewCMSService(db)  // Temporarily disabled
	// apiKeyService := services.NewAPIKeyService(db)  // Temporarily disabled

	// Set real-time service on shortener for click broadcasting
	shortenerService.SetRealtimeService(realtimeAnalyticsService)
	
	// Set attribution service on shortener for attribution tracking
	shortenerService.SetAttributionService(attributionService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService, rbacService, redis)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, adminService, smsService, emailService, config, contentBlockService, pageTemplateService)
	analyticsHandlers := handlers.NewAnalyticsHandlers(userAnalyticsService)
	adminRBACHandlers := handlers.NewAdminRBACHandlers(adminService, rbacService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db, gormDB, config, adminRBACHandlers)

	// Initialize webhook retry worker with email alerting
	webhookProcessor := handlers.NewWebhookEventProcessor(handler.PolarWebhookHandler)

	// Parse alert emails from config (comma-separated list)
	var alertEmails []string
	if config.WebhookAlertEmails != "" {
		emails := strings.Split(config.WebhookAlertEmails, ",")
		for _, email := range emails {
			trimmedEmail := strings.TrimSpace(email)
			if trimmedEmail != "" {
				alertEmails = append(alertEmails, trimmedEmail)
			}
		}
	}

	webhookRetryWorker := workers.NewWebhookRetryWorker(gormDB, webhookProcessor, emailService, alertEmails)
	webhookAdminHandler := handlers.NewWebhookAdminHandler(webhookRetryWorker)

	// Start webhook retry worker in background
	workerCtx, workerCancel := context.WithCancel(context.Background())
	defer workerCancel()
	go webhookRetryWorker.Start(workerCtx)
	log.Println("✅ Webhook retry worker started")

	// Start product sync worker in background
	syncInterval := time.Duration(config.ProductSyncIntervalMinutes) * time.Minute
	polarSvc := handler.BillingHandlers.GetPolarService()
	productSyncWorker := workers.NewProductSyncWorker(polarSvc, syncInterval)
	go productSyncWorker.Start(workerCtx)
	log.Printf("✅ Product sync worker started (interval: %v)", syncInterval)

	// Setup Gin router
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Setup all routes
	routes.SetupRoutes(router, handler, authMiddleware, config)

	// Setup webhook admin routes
	routes.SetupWebhookAdminRoutes(router, webhookAdminHandler, authMiddleware)

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

	// Shutdown webhook retry worker
	workerCancel()
	log.Println("✅ Webhook retry worker stopped")

	// Shutdown real-time analytics service
	realtimeAnalyticsService.Shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		log.Fatalf("Server forced to shutdown: %v", err)
	}

	log.Println("Server exited")
}