package middleware

import (
	"sync"
	"time"

	"github.com/gin-gonic/gin"
)

// MetricsCollector collects application metrics
type MetricsCollector struct {
	mu                sync.RWMutex
	requestDurations  map[string][]time.Duration
	requestCounts     map[string]int64
	errorCounts       map[string]int64
	cacheHits         int64
	cacheMisses       int64
	activeGoroutines  int64
	databaseQueries   int64
	slowQueries       int64
}

var globalMetrics = &MetricsCollector{
	requestDurations: make(map[string][]time.Duration),
	requestCounts:    make(map[string]int64),
	errorCounts:      make(map[string]int64),
}

// GetMetricsCollector returns the global metrics collector
func GetMetricsCollector() *MetricsCollector {
	return globalMetrics
}

// MetricsMiddleware tracks request metrics
func MetricsMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		start := time.Now()
		path := c.FullPath()
		if path == "" {
			path = c.Request.URL.Path
		}

		c.Next()

		duration := time.Since(start)
		status := c.Writer.Status()

		// Record metrics
		globalMetrics.mu.Lock()
		key := path + ":" + c.Request.Method
		globalMetrics.requestDurations[key] = append(globalMetrics.requestDurations[key], duration)
		globalMetrics.requestCounts[key]++

		if status >= 400 {
			globalMetrics.errorCounts[key]++
		}

		// Keep only last 1000 measurements per endpoint
		if len(globalMetrics.requestDurations[key]) > 1000 {
			globalMetrics.requestDurations[key] = globalMetrics.requestDurations[key][1:]
		}
		globalMetrics.mu.Unlock()

		// Log slow requests
		if duration > 100*time.Millisecond {
			c.Writer.Header().Set("X-Slow-Request", "true")
		}
	}
}

// RecordCacheHit records a cache hit
func (m *MetricsCollector) RecordCacheHit() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheHits++
}

// RecordCacheMiss records a cache miss
func (m *MetricsCollector) RecordCacheMiss() {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.cacheMisses++
}

// RecordDatabaseQuery records a database query
func (m *MetricsCollector) RecordDatabaseQuery(duration time.Duration) {
	m.mu.Lock()
	defer m.mu.Unlock()
	m.databaseQueries++
	if duration > 10*time.Millisecond {
		m.slowQueries++
	}
}

// GetStats returns current metrics statistics
func (m *MetricsCollector) GetStats() map[string]interface{} {
	m.mu.RLock()
	defer m.mu.RUnlock()

	stats := make(map[string]interface{})
	endpointStats := make(map[string]interface{})

	for endpoint, durations := range m.requestDurations {
		if len(durations) == 0 {
			continue
		}

		// Calculate percentiles
		sorted := make([]time.Duration, len(durations))
		copy(sorted, durations)
		sortDurations(sorted)

		p50 := sorted[len(sorted)*50/100]
		p95 := sorted[len(sorted)*95/100]
		p99 := sorted[len(sorted)*99/100]

		// Calculate average
		var sum time.Duration
		for _, d := range durations {
			sum += d
		}
		avg := sum / time.Duration(len(durations))

		endpointStats[endpoint] = map[string]interface{}{
			"count":    m.requestCounts[endpoint],
			"errors":   m.errorCounts[endpoint],
			"avg_ms":   avg.Milliseconds(),
			"p50_ms":   p50.Milliseconds(),
			"p95_ms":   p95.Milliseconds(),
			"p99_ms":   p99.Milliseconds(),
			"error_rate": float64(m.errorCounts[endpoint]) / float64(m.requestCounts[endpoint]) * 100,
		}
	}

	stats["endpoints"] = endpointStats
	stats["cache_hit_rate"] = 0.0
	if m.cacheHits+m.cacheMisses > 0 {
		stats["cache_hit_rate"] = float64(m.cacheHits) / float64(m.cacheHits+m.cacheMisses) * 100
	}
	stats["cache_hits"] = m.cacheHits
	stats["cache_misses"] = m.cacheMisses
	stats["database_queries"] = m.databaseQueries
	stats["slow_queries"] = m.slowQueries
	stats["active_goroutines"] = m.activeGoroutines

	return stats
}

// ResetStats resets all metrics (for testing)
func (m *MetricsCollector) ResetStats() {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.requestDurations = make(map[string][]time.Duration)
	m.requestCounts = make(map[string]int64)
	m.errorCounts = make(map[string]int64)
	m.cacheHits = 0
	m.cacheMisses = 0
	m.databaseQueries = 0
	m.slowQueries = 0
}

// sortDurations sorts durations in place (simple bubble sort for small arrays)
func sortDurations(durations []time.Duration) {
	n := len(durations)
	for i := 0; i < n-1; i++ {
		for j := 0; j < n-i-1; j++ {
			if durations[j] > durations[j+1] {
				durations[j], durations[j+1] = durations[j+1], durations[j]
			}
		}
	}
}
