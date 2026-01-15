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
	"github.com/URLshorter/url-shortener/internal/handlers/admin"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/routes"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
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

	// Initialize GORM database for billing and Polar integration
	sslMode := "disable"
	if config.Environment == "production" {
		sslMode = "require"
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

	// Initialize DatabaseManager for admin services
	dbManager, err := storage.NewDatabaseManager(config)
	if err != nil {
		log.Fatalf("Failed to initialize DatabaseManager: %v", err)
	}
	defer dbManager.Close()

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
	cmsService := services.NewCMSService(db)
	mediaService := services.NewMediaService(db, "./uploads", config.BaseURL)
	apiKeyService := services.NewAPIKeyService(db)
	globalSettingsService := services.NewGlobalSettingsService(db)
	contentBlockService := services.NewContentBlockService(db, globalSettingsService)

	// Set real-time service on shortener for click broadcasting
	shortenerService.SetRealtimeService(realtimeAnalyticsService)
	
	// Set attribution service on shortener for attribution tracking
	shortenerService.SetAttributionService(attributionService)

	// Initialize auth middleware
	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService, rbacService, redis)

	// Initialize billing handler (needed for admin handlers)
	billingHandler := handlers.NewBillingHandler(gormDB, config)

	// Initialize admin services (needed before admin middleware)
	snowflakeGen, err := utils.NewSnowflakeGenerator(config.NodeID)
	if err != nil {
		log.Fatalf("Failed to create snowflake generator: %v", err)
	}
	adminAuditService := services.NewAdminAuditService(dbManager, snowflakeGen)

	// Initialize admin middleware
	adminMiddleware := middleware.NewAdminMiddleware(dbManager, adminAuditService, rbacService)

	// Initialize remaining admin services
	adminUserService := services.NewAdminUserService(dbManager, emailService, userService, jwtService)
	adminURLService := services.NewAdminURLService(dbManager)
	domainBlocklistService := services.NewDomainBlocklistService(dbManager)
	adminAnalyticsService := services.NewAdminAnalyticsService(dbManager)
	monitoringService := services.NewMonitoringService(dbManager, redis)

	// Wire services to monitoring for health checks
	monitoringService.SetPolarService(billingHandler.GetPolarService())
	monitoringService.SetEmailProvider(emailService.GetProvider())
	monitoringService.SetSMSService(smsService)

	// Initialize admin handlers
	adminHandlers := admin.NewAdminHandlers(
		dbManager,
		redis,
		adminUserService,
		adminURLService,
		domainBlocklistService,
		adminAnalyticsService,
		monitoringService,
		adminAuditService,
		rbacService,
		billingHandler.GetBillingService(),
		globalSettingsService,
		cmsService,
		mediaService,
		contentBlockService,
		apiKeyService,
	)

	// Initialize handlers
	authHandlers := handlers.NewAuthHandlers(authService, smsService, emailService)
	analyticsHandlers := handlers.NewAnalyticsHandlers(userAnalyticsService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db, gormDB, config, adminHandlers)

	// Initialize admin auth handlers
	adminAuthHandlers := admin.NewAdminAuthHandlers(authService, adminAuditService, userService)

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
	routes.SetupRoutes(router, handler, authMiddleware, adminMiddleware)

	// Setup webhook admin routes
	routes.SetupWebhookAdminRoutes(router, webhookAdminHandler, authMiddleware)

	// Setup admin auth routes (login, logout, refresh)
	routes.SetupAdminAuthRoutes(router, handler.AdminHandlers, adminAuthHandlers, authMiddleware, adminMiddleware)

	// Setup admin panel routes (dashboard, users, analytics, etc.)
	routes.SetupAdminRoutes(router, handler, authMiddleware, adminMiddleware)

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