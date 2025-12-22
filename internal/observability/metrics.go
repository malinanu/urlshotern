package observability

import (
	"strconv"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
	dto "github.com/prometheus/client_model/go"
)

var (
	// HTTP Metrics
	HTTPRequestsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "http_requests_total",
			Help: "Total number of HTTP requests",
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_duration_seconds",
			Help:    "Duration of HTTP requests in seconds",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1, 2.5, 5, 10},
		},
		[]string{"method", "endpoint", "status"},
	)

	HTTPRequestSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_request_size_bytes",
			Help:    "Size of HTTP requests in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	HTTPResponseSize = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "http_response_size_bytes",
			Help:    "Size of HTTP responses in bytes",
			Buckets: prometheus.ExponentialBuckets(100, 10, 8),
		},
		[]string{"method", "endpoint"},
	)

	// URL Shortening Business Metrics
	URLsCreatedTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "urls_created_total",
			Help: "Total number of URLs created",
		},
		[]string{"custom_code", "with_expiry"},
	)

	URLClicksTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "url_clicks_total",
			Help: "Total number of URL clicks/redirects",
		},
		[]string{"short_code"},
	)

	URLCreationDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "url_creation_duration_seconds",
			Help:    "Duration of URL creation operations",
			Buckets: []float64{.001, .005, .01, .025, .05, .1, .25, .5, 1},
		},
	)

	URLLookupDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "url_lookup_duration_seconds",
			Help:    "Duration of URL lookup operations",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
	)

	// Database Metrics
	DBConnectionsActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_active",
			Help: "Number of active database connections",
		},
	)

	DBConnectionsIdle = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "db_connections_idle",
			Help: "Number of idle database connections",
		},
	)

	DBQueryDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "db_query_duration_seconds",
			Help:    "Duration of database queries",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .05, .1, .5, 1, 5},
		},
		[]string{"operation", "table"},
	)

	DBQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_queries_total",
			Help: "Total number of database queries",
		},
		[]string{"operation", "table", "status"},
	)

	DBSlowQueriesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "db_slow_queries_total",
			Help: "Total number of slow database queries (>1s)",
		},
		[]string{"operation", "table"},
	)

	// Cache Metrics
	CacheHitsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_hits_total",
			Help: "Total number of cache hits",
		},
		[]string{"cache_key_type"},
	)

	CacheMissesTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "cache_misses_total",
			Help: "Total number of cache misses",
		},
		[]string{"cache_key_type"},
	)

	CacheOperationDuration = promauto.NewHistogramVec(
		prometheus.HistogramOpts{
			Name:    "cache_operation_duration_seconds",
			Help:    "Duration of cache operations",
			Buckets: []float64{.0001, .0005, .001, .005, .01, .025, .05, .1},
		},
		[]string{"operation"},
	)

	CacheHitRatio = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cache_hit_ratio",
			Help: "Cache hit ratio (hits / (hits + misses))",
		},
		[]string{"cache_key_type"},
	)

	// Business Metrics
	ActiveUsersGauge = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "active_users_total",
			Help: "Number of currently active users",
		},
	)

	RevenueTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "revenue_total_cents",
			Help: "Total revenue in cents",
		},
		[]string{"plan_type"},
	)

	SubscriptionsActive = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "subscriptions_active",
			Help: "Number of active subscriptions",
		},
		[]string{"plan_type"},
	)

	// Error Metrics
	ErrorsTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "errors_total",
			Help: "Total number of errors",
		},
		[]string{"type", "service"},
	)

	// System Metrics
	GoroutinesActive = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "goroutines_active",
			Help: "Number of active goroutines",
		},
	)

	MemoryAllocatedBytes = promauto.NewGauge(
		prometheus.GaugeOpts{
			Name: "memory_allocated_bytes",
			Help: "Bytes of allocated memory",
		},
	)

	GCDuration = promauto.NewHistogram(
		prometheus.HistogramOpts{
			Name:    "gc_duration_seconds",
			Help:    "Duration of garbage collection",
			Buckets: []float64{.00001, .00005, .0001, .0005, .001, .005, .01, .05},
		},
	)
)

// RecordHTTPRequest records HTTP request metrics
func RecordHTTPRequest(method, endpoint string, status int, duration time.Duration, requestSize, responseSize int64) {
	statusStr := strconv.Itoa(status)

	HTTPRequestsTotal.WithLabelValues(method, endpoint, statusStr).Inc()
	HTTPRequestDuration.WithLabelValues(method, endpoint, statusStr).Observe(duration.Seconds())

	if requestSize > 0 {
		HTTPRequestSize.WithLabelValues(method, endpoint).Observe(float64(requestSize))
	}
	if responseSize > 0 {
		HTTPResponseSize.WithLabelValues(method, endpoint).Observe(float64(responseSize))
	}
}

// RecordURLCreation records URL creation metrics
func RecordURLCreation(customCode bool, withExpiry bool, duration time.Duration) {
	URLsCreatedTotal.WithLabelValues(
		strconv.FormatBool(customCode),
		strconv.FormatBool(withExpiry),
	).Inc()
	URLCreationDuration.Observe(duration.Seconds())
}

// RecordURLClick records URL click metrics
func RecordURLClick(shortCode string) {
	URLClicksTotal.WithLabelValues(shortCode).Inc()
}

// RecordURLLookup records URL lookup metrics
func RecordURLLookup(duration time.Duration) {
	URLLookupDuration.Observe(duration.Seconds())
}

// RecordDBQuery records database query metrics
func RecordDBQuery(operation, table string, duration time.Duration, err error) {
	status := "success"
	if err != nil {
		status = "error"
	}

	DBQueriesTotal.WithLabelValues(operation, table, status).Inc()
	DBQueryDuration.WithLabelValues(operation, table).Observe(duration.Seconds())

	if duration > 1*time.Second {
		DBSlowQueriesTotal.WithLabelValues(operation, table).Inc()
	}
}

// UpdateDBConnectionStats updates database connection pool metrics
func UpdateDBConnectionStats(active, idle int) {
	DBConnectionsActive.Set(float64(active))
	DBConnectionsIdle.Set(float64(idle))
}

// RecordCacheOperation records cache operation metrics
func RecordCacheOperation(operation, keyType string, hit bool, duration time.Duration) {
	CacheOperationDuration.WithLabelValues(operation).Observe(duration.Seconds())

	if operation == "get" {
		if hit {
			CacheHitsTotal.WithLabelValues(keyType).Inc()
		} else {
			CacheMissesTotal.WithLabelValues(keyType).Inc()
		}

		// Update cache hit ratio
		updateCacheHitRatio(keyType)
	}
}

// updateCacheHitRatio calculates and updates cache hit ratio
func updateCacheHitRatio(keyType string) {
	hits, _ := CacheHitsTotal.GetMetricWithLabelValues(keyType)
	misses, _ := CacheMissesTotal.GetMetricWithLabelValues(keyType)

	var hitsVal, missesVal float64
	if m, ok := hits.(prometheus.Metric); ok {
		metric := &dto.Metric{}
		if err := m.Write(metric); err == nil {
			hitsVal = metric.Counter.GetValue()
		}
	}
	if m, ok := misses.(prometheus.Metric); ok {
		metric := &dto.Metric{}
		if err := m.Write(metric); err == nil {
			missesVal = metric.Counter.GetValue()
		}
	}

	total := hitsVal + missesVal
	if total > 0 {
		ratio := hitsVal / total
		CacheHitRatio.WithLabelValues(keyType).Set(ratio)
	}
}

// RecordError records error metrics
func RecordError(errorType, service string) {
	ErrorsTotal.WithLabelValues(errorType, service).Inc()
}

// UpdateActiveUsers updates active user count
func UpdateActiveUsers(count int) {
	ActiveUsersGauge.Set(float64(count))
}

// RecordRevenue records revenue metrics
func RecordRevenue(planType string, amountCents int) {
	RevenueTotal.WithLabelValues(planType).Add(float64(amountCents))
}

// UpdateActiveSubscriptions updates active subscription count
func UpdateActiveSubscriptions(planType string, count int) {
	SubscriptionsActive.WithLabelValues(planType).Set(float64(count))
}

// UpdateSystemMetrics updates system-level metrics
func UpdateSystemMetrics(goroutines int, memAllocated uint64) {
	GoroutinesActive.Set(float64(goroutines))
	MemoryAllocatedBytes.Set(float64(memAllocated))
}

// RecordGCDuration records garbage collection duration
func RecordGCDuration(duration time.Duration) {
	GCDuration.Observe(duration.Seconds())
}
