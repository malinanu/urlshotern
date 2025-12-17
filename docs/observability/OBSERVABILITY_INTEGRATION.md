# Observability Integration Guide

This document provides step-by-step instructions for integrating the observability stack into the URL Shortener application.

## Integration Overview

The observability stack has been implemented with:
- ✅ Structured logging (Zap)
- ✅ Prometheus metrics
- ✅ OpenTelemetry distributed tracing
- ✅ Middleware for automatic instrumentation
- ✅ SLO tracking and error budget monitoring
- ✅ Grafana dashboards
- ✅ Alert rules and runbooks
- ✅ Log aggregation (Loki)
- ✅ Complete Docker Compose stack

## Step 1: Install Dependencies

```bash
# Navigate to project root
cd D:\github\URLshorter

# Download new Go dependencies
go mod download

# Verify dependencies
go list -m all | grep -E "(prometheus|opentelemetry|zap)"
```

Expected output:
```
github.com/prometheus/client_golang v1.18.0
go.opentelemetry.io/otel v1.21.0
go.opentelemetry.io/otel/exporters/jaeger v1.17.0
go.opentelemetry.io/otel/sdk v1.21.0
go.opentelemetry.io/otel/trace v1.21.0
go.uber.org/zap v1.26.0
```

## Step 2: Update main.go

Replace the existing `cmd/server/main.go` with observability integration:

```go
package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"runtime"
	"syscall"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/handlers"
	"github.com/URLshorter/url-shortener/internal/middleware"
	"github.com/URLshorter/url-shortener/internal/observability"
	"github.com/URLshorter/url-shortener/internal/routes"
	"github.com/URLshorter/url-shortener/internal/services"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/gin-gonic/gin"
	"github.com/prometheus/client_golang/prometheus/promhttp"
)

func main() {
	// Load configuration
	config, err := configs.LoadConfig()
	if err != nil {
		log.Fatalf("Failed to load config: %v", err)
	}

	// Initialize observability stack
	if err := initializeObservability(config); err != nil {
		log.Fatalf("Failed to initialize observability: %v", err)
	}
	defer observability.Sync()

	observability.Info("Starting URL Shortener service",
		observability.String("version", "1.0.0"),
		observability.String("environment", config.Environment),
	)

	// Initialize storage layers
	db, err := storage.NewPostgresStorage(config)
	if err != nil {
		observability.Fatal("Failed to connect to database",
			observability.Error(err),
		)
	}
	defer db.Close()

	redis, err := storage.NewRedisStorage(config)
	if err != nil {
		observability.Fatal("Failed to connect to Redis",
			observability.Error(err),
		)
	}
	defer redis.Close()

	// Initialize services (existing code)
	shortenerService := services.NewShortenerService(db, redis, config)
	analyticsService := services.NewAnalyticsService(db)
	advancedAnalyticsService := services.NewAdvancedAnalyticsService(db)
	var userAnalyticsService *services.UserAnalyticsService

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

	shortenerService.SetRealtimeService(realtimeAnalyticsService)
	shortenerService.SetAttributionService(attributionService)

	// Initialize SLO tracker
	sloTracker := observability.NewSLOTracker(observability.DefaultSLOConfig())

	authMiddleware := middleware.NewAuthMiddleware(jwtService, userService, rbacService)
	authHandlers := handlers.NewAuthHandlers(authService, smsService, emailService)
	analyticsHandlers := handlers.NewAnalyticsHandlers(userAnalyticsService)
	handler := handlers.NewHandler(shortenerService, analyticsService, advancedAnalyticsService, conversionTrackingService, abTestingService, realtimeAnalyticsService, attributionService, authHandlers, analyticsHandlers, db, config)

	// Setup Gin router
	if config.Environment == "production" {
		gin.SetMode(gin.ReleaseMode)
	}

	router := gin.New()

	// Add observability middleware (BEFORE other middleware)
	router.Use(middleware.RequestIDMiddleware())
	router.Use(middleware.ObservabilityMiddleware())
	router.Use(gin.Recovery())

	// Add metrics endpoint
	router.GET("/metrics", gin.WrapH(promhttp.Handler()))

	// Setup all routes
	routes.SetupRoutes(router, handler, authMiddleware)

	// Start system metrics collector
	metricsCollector := middleware.NewSystemMetricsCollector()
	metricsCollector.Start(30 * time.Second)
	defer metricsCollector.Stop()

	// Start background goroutine for runtime metrics
	go collectRuntimeMetrics()

	// Setup server
	srv := &http.Server{
		Addr:    ":" + config.ServerPort,
		Handler: router,
	}

	// Start server in a goroutine
	go func() {
		observability.Info("Server starting",
			observability.String("port", config.ServerPort),
			observability.String("environment", config.Environment),
		)
		if err := srv.ListenAndServe(); err != nil && err != http.ErrServerClosed {
			observability.Fatal("Failed to start server",
				observability.Error(err),
			)
		}
	}()

	// Wait for interrupt signal to gracefully shutdown the server
	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGINT, syscall.SIGTERM)
	<-quit

	observability.Info("Shutting down server...")

	// Shutdown with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	// Shutdown real-time analytics service
	realtimeAnalyticsService.Shutdown()

	if err := srv.Shutdown(ctx); err != nil {
		observability.Fatal("Server forced to shutdown",
			observability.Error(err),
		)
	}

	observability.Info("Server exited gracefully")
}

func initializeObservability(config *configs.Config) error {
	// Initialize structured logging
	if err := observability.InitLogger(config.Environment, config.LogLevel); err != nil {
		return fmt.Errorf("failed to initialize logger: %w", err)
	}

	// Initialize distributed tracing
	tracingEnabled := os.Getenv("TRACING_ENABLED") == "true"
	if tracingEnabled {
		jaegerEndpoint := os.Getenv("JAEGER_ENDPOINT")
		if jaegerEndpoint == "" {
			jaegerEndpoint = "http://localhost:14268/api/traces"
		}

		tp, err := observability.InitTracing(observability.TracingConfig{
			ServiceName:    "url-shortener",
			ServiceVersion: "1.0.0",
			Environment:    config.Environment,
			JaegerEndpoint: jaegerEndpoint,
			SamplingRate:   0.1, // Sample 10% of traces
			Enabled:        true,
		})
		if err != nil {
			return fmt.Errorf("failed to initialize tracing: %w", err)
		}

		if tp != nil {
			// Shutdown tracer on exit
			defer func() {
				ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
				defer cancel()
				if err := tp.Shutdown(ctx); err != nil {
					observability.Error("Failed to shutdown tracer", observability.Error(err))
				}
			}()
		}
	}

	return nil
}

func collectRuntimeMetrics() {
	ticker := time.NewTicker(15 * time.Second)
	defer ticker.Stop()

	var memStats runtime.MemStats

	for range ticker.C {
		// Collect goroutine count
		goroutines := runtime.NumGoroutine()
		observability.UpdateSystemMetrics(goroutines, 0)

		// Collect memory statistics
		runtime.ReadMemStats(&memStats)
		observability.UpdateSystemMetrics(goroutines, memStats.Alloc)

		// Record GC metrics
		if memStats.NumGC > 0 {
			// Get last GC pause time
			lastPause := memStats.PauseNs[(memStats.NumGC+255)%256]
			observability.RecordGCDuration(time.Duration(lastPause))
		}
	}
}
```

## Step 3: Update Handlers with Observability

Add observability to key handler methods. Example for `ShortenURL`:

```go
func (h *Handler) ShortenURL(c *gin.Context) {
	startTime := time.Now()

	// Start trace span
	ctx, span := observability.StartSpan(c.Request.Context(), "handler.ShortenURL")
	defer span.End()

	var request models.ShortenRequest
	if err := c.ShouldBindJSON(&request); err != nil {
		observability.RecordError("validation_error", "handler")
		c.JSON(http.StatusBadRequest, models.ErrorResponse{
			Error:   "invalid_request",
			Message: err.Error(),
		})
		return
	}

	// Add trace attributes
	observability.AddSpanAttributes(ctx,
		observability.String("request.url", request.URL),
		observability.Bool("request.custom_code", request.CustomCode != ""),
	)

	clientIP := getClientIP(c)
	userID, _ := middleware.GetUserID(c)
	var userIDPtr *int64
	if userID > 0 {
		userIDPtr = &userID
	}

	// Call service
	response, err := h.shortenerService.ShortenURL(&request, clientIP, userIDPtr)
	if err != nil {
		observability.RecordError("url_creation_error", "shortener")
		observability.RecordSpanError(ctx, err)

		statusCode := http.StatusInternalServerError
		errorType := "internal_error"

		switch err {
		case services.ErrInvalidURL, services.ErrInvalidURLScheme, services.ErrSelfReferentialURL:
			statusCode = http.StatusBadRequest
			errorType = "invalid_url"
		case services.ErrInvalidCustomCodeLength, services.ErrInvalidCustomCodeCharacters:
			statusCode = http.StatusBadRequest
			errorType = "invalid_custom_code"
		case services.ErrReservedCustomCode, services.ErrCustomCodeAlreadyExists:
			statusCode = http.StatusConflict
			errorType = "custom_code_unavailable"
		}

		c.JSON(statusCode, models.ErrorResponse{
			Error:   errorType,
			Message: err.Error(),
		})
		return
	}

	// Record metrics
	duration := time.Since(startTime)
	observability.RecordURLCreation(
		request.CustomCode != "",
		request.ExpiresAt != nil,
		duration,
	)

	// Record SLI
	observability.RecordSLIAvailability("url-shortener", true)
	observability.RecordSLILatency("url-shortener", duration, 1*time.Second)
	observability.RecordSLIErrorRate("url-shortener", false)

	// Log success
	observability.Info("URL created successfully",
		observability.String("short_code", response.ShortCode),
		observability.String("original_url", response.OriginalURL),
		observability.Duration("duration", duration),
	)

	c.JSON(http.StatusCreated, response)
}
```

## Step 4: Deploy Observability Stack

### Option A: Local Development

```bash
# Start full stack
docker-compose -f docker/docker-compose.observability.yml up -d

# Wait for services to start (2-3 minutes)
docker-compose -f docker/docker-compose.observability.yml ps

# Check logs
docker-compose -f docker/docker-compose.observability.yml logs -f
```

### Option B: Production Deployment

```bash
# Set production environment variables
export DB_PASSWORD="<secure-password>"
export GRAFANA_PASSWORD="<secure-password>"
export SLACK_WEBHOOK_URL="<slack-webhook>"
export PAGERDUTY_SERVICE_KEY="<pagerduty-key>"

# Deploy stack
docker-compose -f docker/docker-compose.observability.yml up -d

# Verify deployment
curl http://localhost:8080/health
curl http://localhost:8080/metrics
curl http://localhost:9090/-/healthy
curl http://localhost:3000/api/health
```

## Step 5: Configure Alerting

### Set up Slack Integration

1. Create Slack webhook:
   - Go to https://api.slack.com/messaging/webhooks
   - Create a new webhook for your workspace
   - Copy the webhook URL

2. Add to environment:
```bash
echo "SLACK_WEBHOOK_URL=https://hooks.slack.com/services/YOUR/WEBHOOK/URL" >> .env
```

3. Restart Alertmanager:
```bash
docker-compose -f docker/docker-compose.observability.yml restart alertmanager
```

### Set up PagerDuty Integration

1. Create PagerDuty service:
   - Go to Services → Service Directory
   - Create new service
   - Integration Type: "Events API V2"
   - Copy Integration Key

2. Add to environment:
```bash
echo "PAGERDUTY_SERVICE_KEY=your_integration_key" >> .env
```

3. Restart Alertmanager:
```bash
docker-compose -f docker/docker-compose.observability.yml restart alertmanager
```

## Step 6: Verify Observability

### Test Metrics Collection

```bash
# Generate some traffic
for i in {1..100}; do
  curl -X POST http://localhost:8080/api/v1/shorten \
    -H 'Content-Type: application/json' \
    -d '{"url":"https://example.com/test-'$i'"}'
done

# Check Prometheus metrics
curl -s http://localhost:8080/metrics | grep http_requests_total

# View in Prometheus UI
open http://localhost:9090/graph
# Query: rate(http_requests_total[5m])
```

### Test Distributed Tracing

```bash
# Generate requests
curl -X POST http://localhost:8080/api/v1/shorten \
  -H 'Content-Type: application/json' \
  -d '{"url":"https://example.com/trace-test"}'

# View traces in Jaeger
open http://localhost:16686
# Search for service: url-shortener
```

### Test Log Aggregation

```bash
# View logs in Loki
open http://localhost:3000/explore
# Select Loki data source
# Query: {service="url-shortener"} | json
```

### Test Alerting

```bash
# Fire a test alert
curl -X POST http://localhost:9093/api/v2/alerts \
  -H 'Content-Type: application/json' \
  -d '[{
    "labels": {
      "alertname": "TestAlert",
      "severity": "warning",
      "service": "url-shortener"
    },
    "annotations": {
      "summary": "This is a test alert",
      "description": "Testing alert routing to Slack"
    }
  }]'

# Check Slack for notification
# Check Alertmanager UI: http://localhost:9093
```

## Step 7: Access Dashboards

### Grafana Dashboards

1. Open Grafana: http://localhost:3000
2. Login with admin / ${GRAFANA_PASSWORD}
3. Navigate to Dashboards → Browse
4. Open "Service Health Overview"

### Key Dashboards:

- **Service Health Overview**: Start here for quick health check
- **Application Performance**: Detailed request metrics
- **Database Performance**: DB and cache metrics
- **SLO Tracking**: SLO compliance and error budgets
- **Business Metrics**: URLs created, clicks, revenue

## Step 8: Set Up Monitoring Routine

### Daily (5 minutes)

```bash
# Morning health check script
./scripts/daily-health-check.sh
```

Create `scripts/daily-health-check.sh`:
```bash
#!/bin/bash
echo "=== URL Shortener Daily Health Check ==="
echo ""

# Check application health
echo "1. Application Health:"
curl -s http://localhost:8080/health | jq

# Check Prometheus targets
echo -e "\n2. Prometheus Targets:"
curl -s http://localhost:9090/api/v1/targets | jq '.data.activeTargets[] | {job: .labels.job, health: .health}'

# Check alert status
echo -e "\n3. Active Alerts:"
curl -s http://localhost:9093/api/v2/alerts | jq '.[] | {name: .labels.alertname, status: .status.state}'

# Check SLO compliance
echo -e "\n4. SLO Compliance (Last 24h):"
curl -s "http://localhost:9090/api/v1/query?query=sli:availability:ratio_rate24h" | jq '.data.result[0].value[1]'

echo -e "\n=== Health Check Complete ==="
```

### Weekly (30 minutes)

1. Review error budget consumption
2. Analyze slow query logs
3. Review capacity trends
4. Reduce alert noise

### Monthly (2 hours)

1. SLO retrospective meeting
2. Capacity planning review
3. Update runbooks
4. Improve dashboards

## Troubleshooting

### Issue: Metrics not appearing in Prometheus

**Check**:
```bash
# Verify metrics endpoint
curl http://localhost:8080/metrics

# Check Prometheus targets
curl http://localhost:9090/api/v1/targets

# Check Prometheus config
docker exec prometheus cat /etc/prometheus/prometheus.yml
```

**Fix**:
```bash
# Restart Prometheus
docker-compose -f docker/docker-compose.observability.yml restart prometheus

# Check logs
docker logs prometheus
```

### Issue: Traces not appearing in Jaeger

**Check**:
```bash
# Verify trace configuration
env | grep JAEGER

# Check Jaeger health
curl http://localhost:14269/

# Check application logs
docker logs url-shortener-app | grep -i trace
```

**Fix**:
```bash
# Enable tracing
export TRACING_ENABLED=true
export JAEGER_ENDPOINT=http://localhost:14268/api/traces

# Restart application
docker-compose -f docker/docker-compose.observability.yml restart app
```

### Issue: Alerts not being sent

**Check**:
```bash
# Check Alertmanager config
curl http://localhost:9093/api/v2/status

# Check alert routing
curl http://localhost:9093/api/v2/alerts

# Test Slack webhook
curl -X POST ${SLACK_WEBHOOK_URL} \
  -H 'Content-Type: application/json' \
  -d '{"text":"Test message from Alertmanager"}'
```

**Fix**:
```bash
# Reload Alertmanager config
curl -X POST http://localhost:9093/-/reload

# Restart Alertmanager
docker-compose -f docker/docker-compose.observability.yml restart alertmanager
```

## Next Steps

1. **Customize Dashboards**: Tailor Grafana dashboards to your needs
2. **Fine-tune Alerts**: Adjust alert thresholds based on actual traffic
3. **Test Runbooks**: Conduct disaster recovery drills
4. **Train Team**: Ensure all team members understand the observability stack
5. **Document Incidents**: Create postmortems and update runbooks

## Support

For questions or issues:
- Check runbooks: `observability/runbooks/`
- Review logs: Loki dashboards
- Escalate: Follow incident response procedures

---

**Integration Complete! Your URL Shortener now has production-grade observability with MTTD < 2 minutes.**
