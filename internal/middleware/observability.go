package middleware

import (
	"time"

	"github.com/URLshorter/url-shortener/internal/observability"
	"github.com/gin-gonic/gin"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

// ObservabilityMiddleware provides comprehensive observability for HTTP requests
func ObservabilityMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		// Start distributed trace
		ctx, span := observability.StartSpan(c.Request.Context(), "http.request",
			trace.WithSpanKind(trace.SpanKindServer),
		)
		defer span.End()

		// Update context with span
		c.Request = c.Request.WithContext(ctx)

		// Add trace attributes
		observability.AddSpanAttributes(ctx,
			attribute.String("http.method", c.Request.Method),
			attribute.String("http.path", c.Request.URL.Path),
			attribute.String("http.client_ip", c.ClientIP()),
			attribute.String("http.user_agent", c.Request.UserAgent()),
		)

		// Get request ID from context (if request ID middleware ran before)
		requestID := c.GetString("request_id")
		if requestID != "" {
			observability.AddSpanAttributes(ctx, attribute.String("request.id", requestID))
		}

		// Process request
		c.Next()

		// Calculate latency
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		// Add response attributes to span
		observability.AddSpanAttributes(ctx,
			attribute.Int("http.status_code", statusCode),
			attribute.Int64("http.response_size", int64(c.Writer.Size())),
			attribute.Int64("http.latency_ms", latency.Milliseconds()),
		)

		// Record span error if status code indicates error
		if statusCode >= 500 {
			observability.RecordSpanError(ctx, nil)
		}

		// Record Prometheus metrics
		observability.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(), // Use route pattern, not actual path
			statusCode,
			latency,
			c.Request.ContentLength,
			int64(c.Writer.Size()),
		)

		// Structured logging
		fields := []zap.Field{
			zap.String("method", c.Request.Method),
			zap.String("path", c.Request.URL.Path),
			zap.String("client_ip", c.ClientIP()),
			zap.String("user_agent", c.Request.UserAgent()),
			zap.Int("status", statusCode),
			zap.Duration("latency", latency),
			zap.Int("response_size", c.Writer.Size()),
		}

		if requestID != "" {
			fields = append(fields, zap.String("request_id", requestID))
		}

		// Add error details if present
		if len(c.Errors) > 0 {
			fields = append(fields, zap.String("error", c.Errors.String()))
			observability.Error("HTTP request completed with errors", fields...)
		} else if statusCode >= 500 {
			observability.Error("HTTP server error", fields...)
		} else if statusCode >= 400 {
			observability.Warn("HTTP client error", fields...)
		} else {
			observability.Info("HTTP request completed", fields...)
		}
	}
}

// MetricsEndpointMiddleware provides a Prometheus metrics endpoint
func MetricsEndpointMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		// Metrics are served by promhttp.Handler() separately
		// This is a placeholder for custom metrics logic if needed
		c.Next()
	}
}

// HealthCheckObservability adds observability to health check endpoints
func HealthCheckObservability() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()

		c.Next()

		// Only record basic metrics for health checks (don't create traces)
		latency := time.Since(startTime)
		statusCode := c.Writer.Status()

		observability.RecordHTTPRequest(
			c.Request.Method,
			c.FullPath(),
			statusCode,
			latency,
			0,
			int64(c.Writer.Size()),
		)
	}
}

// DatabaseObservabilityWrapper wraps database operations with observability
func DatabaseObservabilityWrapper(operation, table string, fn func() error) error {
	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	// Record metrics
	observability.RecordDBQuery(operation, table, duration, err)

	// Log slow queries
	if duration > 1*time.Second {
		observability.Warn("Slow database query detected",
			zap.String("operation", operation),
			zap.String("table", table),
			zap.Duration("duration", duration),
		)
	}

	return err
}

// CacheObservabilityWrapper wraps cache operations with observability
func CacheObservabilityWrapper(operation, keyType, key string, hit bool, fn func() error) error {
	startTime := time.Now()
	err := fn()
	duration := time.Since(startTime)

	// Record metrics
	observability.RecordCacheOperation(operation, keyType, hit, duration)

	// Log cache operations
	observability.LogCacheOperation(operation, key, hit, duration)

	return err
}

// RecordBusinessEvent records custom business events for observability
func RecordBusinessEvent(ctx *gin.Context, eventName string, attributes map[string]interface{}) {
	// Add event to trace span
	if ctx != nil {
		var attrs []attribute.KeyValue
		for k, v := range attributes {
			switch val := v.(type) {
			case string:
				attrs = append(attrs, attribute.String(k, val))
			case int:
				attrs = append(attrs, attribute.Int(k, val))
			case int64:
				attrs = append(attrs, attribute.Int64(k, val))
			case float64:
				attrs = append(attrs, attribute.Float64(k, val))
			case bool:
				attrs = append(attrs, attribute.Bool(k, val))
			}
		}
		observability.AddSpanEvent(ctx.Request.Context(), eventName, attrs...)
	}

	// Log business event
	fields := []zap.Field{zap.String("event", eventName)}
	for k, v := range attributes {
		fields = append(fields, zap.Any(k, v))
	}
	observability.Info("Business event", fields...)
}

// SystemMetricsCollector periodically collects system metrics
type SystemMetricsCollector struct {
	stopChan chan struct{}
}

// NewSystemMetricsCollector creates a new system metrics collector
func NewSystemMetricsCollector() *SystemMetricsCollector {
	return &SystemMetricsCollector{
		stopChan: make(chan struct{}),
	}
}

// Start begins collecting system metrics
func (s *SystemMetricsCollector) Start(interval time.Duration) {
	ticker := time.NewTicker(interval)
	go func() {
		for {
			select {
			case <-ticker.C:
				s.collectMetrics()
			case <-s.stopChan:
				ticker.Stop()
				return
			}
		}
	}()
}

// Stop stops the metrics collector
func (s *SystemMetricsCollector) Stop() {
	close(s.stopChan)
}

// collectMetrics collects and reports system metrics
func (s *SystemMetricsCollector) collectMetrics() {
	// This would typically collect runtime metrics
	// For now, we'll use a placeholder
	// In production, integrate with runtime.NumGoroutine(), runtime.MemStats, etc.
	observability.UpdateSystemMetrics(0, 0)
}
