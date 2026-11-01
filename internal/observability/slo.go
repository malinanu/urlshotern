package observability

import (
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/promauto"
)

// SLI (Service Level Indicator) metrics for SLO tracking

var (
	// Availability SLI - tracks successful requests vs total requests
	SLIAvailabilityTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sli_availability_requests_total",
			Help: "Total number of requests for availability SLI",
		},
		[]string{"service", "success"},
	)

	// Latency SLI - tracks requests meeting latency target
	SLILatencyTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sli_latency_requests_total",
			Help: "Total number of requests for latency SLI",
		},
		[]string{"service", "within_target"},
	)

	// Error Rate SLI - tracks error rate over time
	SLIErrorRateTotal = promauto.NewCounterVec(
		prometheus.CounterOpts{
			Name: "sli_error_rate_requests_total",
			Help: "Total number of requests for error rate SLI",
		},
		[]string{"service", "error"},
	)

	// SLO Compliance - current SLO compliance ratio
	SLOCompliance = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "slo_compliance_ratio",
			Help: "Current SLO compliance ratio (0-1)",
		},
		[]string{"service", "slo_type"},
	)

	// Error Budget - remaining error budget
	ErrorBudgetRemaining = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "error_budget_remaining_ratio",
			Help: "Remaining error budget as a ratio (0-1)",
		},
		[]string{"service", "slo_type"},
	)

	// Error Budget Burn Rate - rate at which error budget is being consumed
	ErrorBudgetBurnRate = promauto.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "error_budget_burn_rate",
			Help: "Rate at which error budget is being consumed (1.0 = normal, >1.0 = faster)",
		},
		[]string{"service", "slo_type", "window"},
	)
)

// SLOConfig defines Service Level Objectives
type SLOConfig struct {
	Service                string
	AvailabilityTarget     float64 // e.g., 0.9995 for 99.95%
	LatencyP95Target       time.Duration
	LatencyP99Target       time.Duration
	ErrorRateTarget        float64 // e.g., 0.001 for 0.1%
	ErrorBudgetWindow      time.Duration
	AlertBurnRateThreshold float64
}

// DefaultSLOConfig returns the default SLO configuration for the URL shortener
func DefaultSLOConfig() SLOConfig {
	return SLOConfig{
		Service:                "url-shortener",
		AvailabilityTarget:     0.9995, // 99.95% availability
		LatencyP95Target:       1 * time.Second,
		LatencyP99Target:       2 * time.Second,
		ErrorRateTarget:        0.001,               // 0.1% error rate
		ErrorBudgetWindow:      30 * 24 * time.Hour, // 30 days
		AlertBurnRateThreshold: 2.0,                 // Alert if burning error budget 2x faster than normal
	}
}

// RecordSLIAvailability records an availability SLI measurement
func RecordSLIAvailability(service string, success bool) {
	SLIAvailabilityTotal.WithLabelValues(service, boolToString(success)).Inc()
}

// RecordSLILatency records a latency SLI measurement
func RecordSLILatency(service string, latency, target time.Duration) {
	withinTarget := latency <= target
	SLILatencyTotal.WithLabelValues(service, boolToString(withinTarget)).Inc()
}

// RecordSLIErrorRate records an error rate SLI measurement
func RecordSLIErrorRate(service string, isError bool) {
	SLIErrorRateTotal.WithLabelValues(service, boolToString(isError)).Inc()
}

// UpdateSLOCompliance updates the current SLO compliance ratio
func UpdateSLOCompliance(service, sloType string, compliance float64) {
	SLOCompliance.WithLabelValues(service, sloType).Set(compliance)
}

// UpdateErrorBudget updates the remaining error budget
func UpdateErrorBudget(service, sloType string, remaining float64) {
	ErrorBudgetRemaining.WithLabelValues(service, sloType).Set(remaining)
}

// UpdateErrorBudgetBurnRate updates the error budget burn rate
func UpdateErrorBudgetBurnRate(service, sloType, window string, burnRate float64) {
	ErrorBudgetBurnRate.WithLabelValues(service, sloType, window).Set(burnRate)
}

// SLOTracker tracks SLO compliance and error budget
type SLOTracker struct {
	config SLOConfig
}

// NewSLOTracker creates a new SLO tracker
func NewSLOTracker(config SLOConfig) *SLOTracker {
	return &SLOTracker{
		config: config,
	}
}

// TrackRequest tracks a request for SLO compliance
func (t *SLOTracker) TrackRequest(success bool, latency time.Duration, isError bool) {
	service := t.config.Service

	// Track availability SLI
	RecordSLIAvailability(service, success)

	// Track latency SLI (P95 target)
	RecordSLILatency(service, latency, t.config.LatencyP95Target)

	// Track error rate SLI
	RecordSLIErrorRate(service, isError)

	// Calculate and update compliance metrics
	// This is a simplified calculation - in production, you'd use time-windowed calculations
	t.updateCompliance()
}

// updateCompliance calculates and updates SLO compliance metrics
func (t *SLOTracker) updateCompliance() {
	// This is a placeholder for compliance calculation
	// In production, you would:
	// 1. Query Prometheus for historical data over the SLO window
	// 2. Calculate actual availability, latency, and error rates
	// 3. Compare against targets
	// 4. Calculate error budget remaining
	// 5. Calculate burn rate over multiple time windows (1h, 6h, 24h, 72h)

	// For now, we'll just set placeholder values
	service := t.config.Service

	// These would be calculated from actual metrics
	UpdateSLOCompliance(service, "availability", 0.9997)
	UpdateSLOCompliance(service, "latency_p95", 0.9985)
	UpdateSLOCompliance(service, "error_rate", 0.9990)

	UpdateErrorBudget(service, "availability", 0.85)
	UpdateErrorBudget(service, "latency", 0.90)
	UpdateErrorBudget(service, "error_rate", 0.88)

	UpdateErrorBudgetBurnRate(service, "availability", "1h", 0.8)
	UpdateErrorBudgetBurnRate(service, "availability", "6h", 1.2)
	UpdateErrorBudgetBurnRate(service, "availability", "24h", 1.0)
}

// CheckErrorBudget checks if error budget is being burned too quickly
func (t *SLOTracker) CheckErrorBudget() map[string]bool {
	// Returns map of SLO types to whether they're violating burn rate threshold
	// In production, this would query Prometheus and calculate actual burn rates
	return map[string]bool{
		"availability": false,
		"latency":      false,
		"error_rate":   false,
	}
}

// GetSLOStatus returns current SLO status
func (t *SLOTracker) GetSLOStatus() map[string]interface{} {
	return map[string]interface{}{
		"service": t.config.Service,
		"slos": map[string]interface{}{
			"availability": map[string]interface{}{
				"target":                 t.config.AvailabilityTarget * 100,
				"current":                99.97, // Would be calculated from metrics
				"compliant":              true,
				"error_budget_remaining": 85.0,
			},
			"latency_p95": map[string]interface{}{
				"target":                 t.config.LatencyP95Target.Milliseconds(),
				"current":                850, // Would be calculated from metrics
				"compliant":              true,
				"error_budget_remaining": 90.0,
			},
			"error_rate": map[string]interface{}{
				"target":                 t.config.ErrorRateTarget * 100,
				"current":                0.08, // Would be calculated from metrics
				"compliant":              true,
				"error_budget_remaining": 88.0,
			},
		},
	}
}

// Helper function to convert bool to string
func boolToString(b bool) string {
	if b {
		return "true"
	}
	return "false"
}

// SLO-related Prometheus recording rules (to be added to Prometheus config)
/*
Recording rules for efficient SLO calculation:

groups:
  - name: slo_rules
    interval: 30s
    rules:
      # Availability SLI
      - record: sli:availability:ratio_rate5m
        expr: |
          sum(rate(sli_availability_requests_total{success="true"}[5m]))
          /
          sum(rate(sli_availability_requests_total[5m]))

      - record: sli:availability:ratio_rate30m
        expr: |
          sum(rate(sli_availability_requests_total{success="true"}[30m]))
          /
          sum(rate(sli_availability_requests_total[30m]))

      # Latency SLI
      - record: sli:latency:ratio_rate5m
        expr: |
          sum(rate(sli_latency_requests_total{within_target="true"}[5m]))
          /
          sum(rate(sli_latency_requests_total[5m]))

      # Error Rate SLI
      - record: sli:error_rate:ratio_rate5m
        expr: |
          sum(rate(sli_error_rate_requests_total{error="false"}[5m]))
          /
          sum(rate(sli_error_rate_requests_total[5m]))

      # Error Budget Burn Rate (1h window)
      - record: slo:error_budget:burn_rate_1h
        expr: |
          (1 - sli:availability:ratio_rate1h) / (1 - 0.9995)

      # Error Budget Burn Rate (6h window)
      - record: slo:error_budget:burn_rate_6h
        expr: |
          (1 - sli:availability:ratio_rate6h) / (1 - 0.9995)
*/
