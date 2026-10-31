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
	"github.com/jmoiron/sqlx"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
)

func main() {
	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Guard against booting with a missing or default JWT signing secret. With
	// the default secret, anyone can forge tokens (including admin), so refuse to
	// start in production. In non-production we only warn to keep local dev easy.
	const defaultJWTSecret = "your-secret-key-change-this-in-production"
	if config.JWTSecret == "" || config.JWTSecret == defaultJWTSecret {
		if config.Environment == "production" {
			log.Fatalf("SECURITY: JWT_SECRET is unset or using the insecure default. Set a strong, random JWT_SECRET before starting in production.")
		}
		log.Printf("WARNING: JWT_SECRET is unset or using the insecure default. This is unsafe outside local development — set JWT_SECRET.")
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

	// Reconcile stale is_admin flags — clears the legacy flag for any user whose
	// admin/super_admin RBAC role was removed without the flag being updated.
	if err := db.ReconcileAdminFlags(); err != nil {
		log.Printf("Warning: failed to reconcile admin flags: %v", err)
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
	sqlxDB := sqlx.NewDb(db.GetRawDB(), "postgres")
	userAnalyticsService := services.NewUserAnalyticsService(sqlxDB)
	botDetectionService := services.NewBotDetectionService(redis)
	aiInsightsService := services.NewAIInsightsService(config.OpenRouterAPIKey, db, redis, advancedAnalyticsService)
	brandingService := services.NewBrandingService(db)

	// Initialize auth-related services
	smsService := services.NewSMSService(db, config)
	emailService := services.NewEmailService(db, config)
	jwtService := services.NewJWTService(config.JWTSecret, config.JWTIssuer, time.Hour*24, time.Hour*24*7)
	userService := services.NewUserService(db, redis, jwtService, smsService, emailService, nil)
	domainService := services.NewDomainService(db, redis, config, analyticsService, userService)
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
	rbacServiceBase := services.NewRBACService(db, redis)
	teamService := services.NewTeamService(db, redis, rbacServiceBase, userService, emailService)
	teamHandlers := handlers.NewTeamHandlers(teamService, rbacServiceBase, userService, db)
	campaignService := services.NewCampaignService(db, config, emailService, smsService)
	authService.SetCampaignService(campaignService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db, gormDB, config, adminRBACHandlers, botDetectionService, aiInsightsService, domainService, brandingService, teamHandlers, globalSettingsService)
	handler.CampaignHandlers = handlers.NewCampaignHandlers(campaignService)

	// Microsites (/m/<slug> landing pages + digital vCards) — off unless enabled.
	// Nothing below runs, and no routes are registered, when the flag is false.
	if config.EnableMicrosites {
		assetStore, err := services.NewLocalAssetStore(config.MicrositeUploadPath)
		if err != nil {
			log.Printf("Warning: microsites disabled — upload directory unavailable: %v", err)
		} else {
			micrositeService := services.NewMicrositeService(db, redis)
			handler.MicrositeHandlers = handlers.NewMicrositeHandlers(
				micrositeService,
				services.NewMicrositeAnalyticsService(db, redis, config.JWTSecret),
				services.NewMicrositeAssetService(db, assetStore, config.MicrositeMaxUploadBytes),
				handler.GetBillingService(),
				shortenerService,
				botDetectionService,
				config.BaseURL,
				assetStore.Root(),
			)
			authService.OnAccountDeleted(func(userID int64) {
				if err := assetStore.DeleteUserFiles(userID); err != nil {
					log.Printf("Warning: failed to remove microsite uploads for user %d: %v", userID, err)
				}
			})
			if err := micrositeService.Ready(); err != nil {
				log.Printf("Warning: microsites enabled but not ready (apply migration 046): %v", err)
			} else {
				log.Println("✅ Microsites enabled")
			}
		}
	}

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

	// Initialize and start VirusTotal worker
	vtWorker := workers.NewVirusTotalWorker(redis, db, emailService, config.VirusTotalAPIKey)
	go vtWorker.Start(workerCtx)
	log.Println("✅ VirusTotal worker started")

	// Start product sync worker in background
	syncInterval := time.Duration(config.ProductSyncIntervalMinutes) * time.Minute
	polarSvc := handler.BillingHandlers.GetPolarService()
	productSyncWorker := workers.NewProductSyncWorker(polarSvc, syncInterval)
	go productSyncWorker.Start(workerCtx)
	log.Printf("✅ Product sync worker started (interval: %v)", syncInterval)

	// Start data retention worker in background (nightly at 02:00)
	retentionWorker := workers.NewRetentionWorker(db)
	go retentionWorker.Start(workerCtx)
	log.Println("✅ Data retention worker started")

	// Setup Gin router
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.Default()

	// Trust only the nginx/Cloudflare proxy layer — never raw client headers.
	// Cloudflare IPv4 ranges + loopback for local Docker networking.
	router.SetTrustedProxies([]string{
		"127.0.0.1", "::1", "172.16.0.0/12", // Docker / loopback
		"103.21.244.0/22", "103.22.200.0/22", "103.31.4.0/22",
		"104.16.0.0/13", "104.24.0.0/14", "108.162.192.0/18",
		"131.0.72.0/22", "141.101.64.0/18", "162.158.0.0/15",
		"172.64.0.0/13", "173.245.48.0/20", "188.114.96.0/20",
		"190.93.240.0/20", "197.234.240.0/22", "198.41.128.0/17",
		// Cloudflare IPv6
		"2400:cb00::/32", "2606:4700::/32", "2803:f800::/32",
		"2405:b500::/32", "2405:8100::/32", "2a06:98c0::/29", "2c0f:f248::/32",
	})

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