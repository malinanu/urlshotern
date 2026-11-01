package observability

import (
	"context"
	"fmt"

	"go.opentelemetry.io/otel"
	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/exporters/jaeger"
	"go.opentelemetry.io/otel/propagation"
	"go.opentelemetry.io/otel/sdk/resource"
	sdktrace "go.opentelemetry.io/otel/sdk/trace"
	semconv "go.opentelemetry.io/otel/semconv/v1.17.0"
	"go.opentelemetry.io/otel/trace"
	"go.uber.org/zap"
)

var (
	tracer trace.Tracer
)

// TracingConfig holds the tracing configuration
type TracingConfig struct {
	ServiceName    string
	ServiceVersion string
	Environment    string
	JaegerEndpoint string
	SamplingRate   float64
	Enabled        bool
}

// InitTracing initializes OpenTelemetry tracing
func InitTracing(config TracingConfig) (*sdktrace.TracerProvider, error) {
	if !config.Enabled {
		Info("Tracing is disabled")
		return nil, nil
	}

	// Create Jaeger exporter
	exporter, err := jaeger.New(
		jaeger.WithCollectorEndpoint(jaeger.WithEndpoint(config.JaegerEndpoint)),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create Jaeger exporter: %w", err)
	}

	// Create resource with service information
	res, err := resource.Merge(
		resource.Default(),
		resource.NewWithAttributes(
			semconv.SchemaURL,
			semconv.ServiceName(config.ServiceName),
			semconv.ServiceVersion(config.ServiceVersion),
			semconv.DeploymentEnvironment(config.Environment),
		),
	)
	if err != nil {
		return nil, fmt.Errorf("failed to create resource: %w", err)
	}

	// Create tracer provider with sampling
	tp := sdktrace.NewTracerProvider(
		sdktrace.WithBatcher(exporter),
		sdktrace.WithResource(res),
		sdktrace.WithSampler(sdktrace.ParentBased(
			sdktrace.TraceIDRatioBased(config.SamplingRate),
		)),
	)

	// Set global tracer provider and propagator
	otel.SetTracerProvider(tp)
	otel.SetTextMapPropagator(propagation.NewCompositeTextMapPropagator(
		propagation.TraceContext{},
		propagation.Baggage{},
	))

	// Initialize tracer
	tracer = tp.Tracer(config.ServiceName)

	Info("Tracing initialized successfully",
		zap.String("service", config.ServiceName),
		zap.String("jaeger_endpoint", config.JaegerEndpoint),
		zap.Float64("sampling_rate", config.SamplingRate),
	)

	return tp, nil
}

// GetTracer returns the global tracer instance
func GetTracer() trace.Tracer {
	return tracer
}

// StartSpan starts a new span with the given name
func StartSpan(ctx context.Context, name string, opts ...trace.SpanStartOption) (context.Context, trace.Span) {
	if tracer == nil {
		return ctx, trace.SpanFromContext(ctx)
	}
	return tracer.Start(ctx, name, opts...)
}

// AddSpanAttributes adds attributes to the current span
func AddSpanAttributes(ctx context.Context, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.SetAttributes(attrs...)
	}
}

// AddSpanEvent adds an event to the current span
func AddSpanEvent(ctx context.Context, name string, attrs ...attribute.KeyValue) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() {
		span.AddEvent(name, trace.WithAttributes(attrs...))
	}
}

// RecordSpanError records an error on the current span
func RecordSpanError(ctx context.Context, err error) {
	span := trace.SpanFromContext(ctx)
	if span.IsRecording() && err != nil {
		span.RecordError(err)
	}
}

// SpanFromContext returns the current span from the context
func SpanFromContext(ctx context.Context) trace.Span {
	return trace.SpanFromContext(ctx)
}

// ContextWithSpan returns a new context with the given span
func ContextWithSpan(ctx context.Context, span trace.Span) context.Context {
	return trace.ContextWithSpan(ctx, span)
}

// Helper functions for common trace attributes

// HTTPAttributes returns common HTTP attributes for tracing
func HTTPAttributes(method, path, clientIP string, statusCode int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("http.method", method),
		attribute.String("http.path", path),
		attribute.String("http.client_ip", clientIP),
		attribute.Int("http.status_code", statusCode),
	}
}

// DBAttributes returns common database attributes for tracing
func DBAttributes(operation, table string, rowCount int) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("db.operation", operation),
		attribute.String("db.table", table),
		attribute.Int("db.row_count", rowCount),
	}
}

// CacheAttributes returns common cache attributes for tracing
func CacheAttributes(operation, key string, hit bool) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("cache.operation", operation),
		attribute.String("cache.key", key),
		attribute.Bool("cache.hit", hit),
	}
}

// URLAttributes returns URL shortening specific attributes
func URLAttributes(shortCode, originalURL string, customCode bool) []attribute.KeyValue {
	return []attribute.KeyValue{
		attribute.String("url.short_code", shortCode),
		attribute.String("url.original_url", originalURL),
		attribute.Bool("url.custom_code", customCode),
	}
}

// Helper functions to create zap fields (for compatibility with logger)
func String(key, val string) attribute.KeyValue {
	return attribute.String(key, val)
}

func Int(key string, val int) attribute.KeyValue {
	return attribute.Int(key, val)
}

func Int64(key string, val int64) attribute.KeyValue {
	return attribute.Int64(key, val)
}

func Float64(key string, val float64) attribute.KeyValue {
	return attribute.Float64(key, val)
}

func Bool(key string, val bool) attribute.KeyValue {
	return attribute.Bool(key, val)
}
