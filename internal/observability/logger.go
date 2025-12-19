package observability

import (
	"time"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

var (
	// Logger is the global structured logger instance
	Logger *zap.Logger
	// Sugar provides a more ergonomic API for structured logging
	Sugar *zap.SugaredLogger
)

// InitLogger initializes the structured logger based on environment
func InitLogger(environment string, logLevel string) error {
	var config zap.Config

	if environment == "production" {
		config = zap.NewProductionConfig()
		config.EncoderConfig.TimeKey = "timestamp"
		config.EncoderConfig.EncodeTime = zapcore.ISO8601TimeEncoder
	} else {
		config = zap.NewDevelopmentConfig()
		config.EncoderConfig.EncodeLevel = zapcore.CapitalColorLevelEncoder
	}

	// Set log level
	level, err := zapcore.ParseLevel(logLevel)
	if err != nil {
		level = zapcore.InfoLevel
	}
	config.Level = zap.NewAtomicLevelAt(level)

	// Add caller information
	config.EncoderConfig.CallerKey = "caller"
	config.EncoderConfig.EncodeCaller = zapcore.ShortCallerEncoder

	// Initialize logger
	logger, err := config.Build(
		zap.AddCaller(),
		zap.AddCallerSkip(1),
		zap.AddStacktrace(zapcore.ErrorLevel),
	)
	if err != nil {
		return err
	}

	Logger = logger
	Sugar = logger.Sugar()

	// Replace global logger
	zap.ReplaceGlobals(logger)

	return nil
}

// Sync flushes any buffered log entries
func Sync() {
	if Logger != nil {
		_ = Logger.Sync()
	}
}

// Info logs an info message with structured fields
func Info(msg string, fields ...zap.Field) {
	Logger.Info(msg, fields...)
}

// Debug logs a debug message with structured fields
func Debug(msg string, fields ...zap.Field) {
	Logger.Debug(msg, fields...)
}

// Warn logs a warning message with structured fields
func Warn(msg string, fields ...zap.Field) {
	Logger.Warn(msg, fields...)
}

// Error logs an error message with structured fields
func Error(msg string, fields ...zap.Field) {
	Logger.Error(msg, fields...)
}

// Fatal logs a fatal message with structured fields and exits
func Fatal(msg string, fields ...zap.Field) {
	Logger.Fatal(msg, fields...)
}

// WithContext creates a logger with contextual fields
func WithContext(fields ...zap.Field) *zap.Logger {
	return Logger.With(fields...)
}

// HTTPLogger returns a logger specifically for HTTP requests
func HTTPLogger() *HTTPRequestLogger {
	return &HTTPRequestLogger{
		logger: Logger,
	}
}

// HTTPRequestLogger wraps zap.Logger for HTTP request logging
type HTTPRequestLogger struct {
	logger *zap.Logger
}

// LogRequest logs an HTTP request with relevant fields
func (l *HTTPRequestLogger) LogRequest(method, path, clientIP, userAgent string, statusCode, latencyMs int, err error) {
	fields := []zap.Field{
		zap.String("method", method),
		zap.String("path", path),
		zap.String("client_ip", clientIP),
		zap.String("user_agent", userAgent),
		zap.Int("status_code", statusCode),
		zap.Int("latency_ms", latencyMs),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		l.logger.Error("HTTP request failed", fields...)
	} else if statusCode >= 500 {
		l.logger.Error("HTTP server error", fields...)
	} else if statusCode >= 400 {
		l.logger.Warn("HTTP client error", fields...)
	} else {
		l.logger.Info("HTTP request completed", fields...)
	}
}

// LogDatabaseQuery logs database queries for observability
func LogDatabaseQuery(query string, duration time.Duration, err error) {
	fields := []zap.Field{
		zap.String("query", query),
		zap.Duration("duration", duration),
	}

	if err != nil {
		fields = append(fields, zap.Error(err))
		Logger.Error("Database query failed", fields...)
	} else if duration > 1*time.Second {
		Logger.Warn("Slow database query detected", fields...)
	} else {
		Logger.Debug("Database query executed", fields...)
	}
}

// LogCacheOperation logs cache operations
func LogCacheOperation(operation, key string, hit bool, duration time.Duration) {
	fields := []zap.Field{
		zap.String("operation", operation),
		zap.String("key", key),
		zap.Bool("cache_hit", hit),
		zap.Duration("duration", duration),
	}

	Logger.Debug("Cache operation", fields...)
}

// LogBusinessMetric logs business-level metrics for tracking
func LogBusinessMetric(metric string, value interface{}, labels map[string]string) {
	fields := []zap.Field{
		zap.String("metric", metric),
		zap.Any("value", value),
	}

	for k, v := range labels {
		fields = append(fields, zap.String(k, v))
	}

	Logger.Info("Business metric", fields...)
}
