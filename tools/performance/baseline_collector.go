package main

import (
	"context"
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"os"
	"time"

	"github.com/prometheus/client_golang/api"
	v1 "github.com/prometheus/client_golang/api/prometheus/v1"
	"github.com/prometheus/common/model"
)

// PerformanceBaseline represents a snapshot of performance metrics
type PerformanceBaseline struct {
	Timestamp      time.Time                  `json:"timestamp"`
	Period         string                     `json:"period"`
	Metrics        map[string]MetricStats     `json:"metrics"`
	Endpoints      map[string]EndpointMetrics `json:"endpoints"`
	Infrastructure InfraMetrics               `json:"infrastructure"`
	Cost           CostMetrics                `json:"cost"`
}

// MetricStats contains statistical information about a metric
type MetricStats struct {
	P50   float64 `json:"p50"`
	P95   float64 `json:"p95"`
	P99   float64 `json:"p99"`
	Mean  float64 `json:"mean"`
	Min   float64 `json:"min"`
	Max   float64 `json:"max"`
	Count int64   `json:"count"`
}

// EndpointMetrics contains performance metrics for a specific endpoint
type EndpointMetrics struct {
	Latency      MetricStats `json:"latency"`
	Throughput   float64     `json:"throughput"`
	ErrorRate    float64     `json:"error_rate"`
	CacheHitRate float64     `json:"cache_hit_rate"`
}

// InfraMetrics contains infrastructure-level metrics
type InfraMetrics struct {
	CPUUtilization    float64 `json:"cpu_utilization"`
	MemoryUtilization float64 `json:"memory_utilization"`
	NetworkInbound    float64 `json:"network_inbound_mbps"`
	NetworkOutbound   float64 `json:"network_outbound_mbps"`
	ActiveConnections int64   `json:"active_connections"`
	ActiveGoroutines  int64   `json:"active_goroutines"`
}

// CostMetrics contains cost-related information
type CostMetrics struct {
	TotalRequestsPerDay    int64   `json:"total_requests_per_day"`
	CostPerMillionRequests float64 `json:"cost_per_million_requests"`
	EstimatedDailyCost     float64 `json:"estimated_daily_cost"`
}

type Config struct {
	PrometheusURL string
	LookbackHours int
	OutputFormat  string
	OutputFile    string
}

func main() {
	config := parseFlags()

	log.Printf("Collecting performance baseline from Prometheus: %s", config.PrometheusURL)
	log.Printf("Lookback period: %d hours", config.LookbackHours)

	baseline, err := collectBaseline(config)
	if err != nil {
		log.Fatalf("Failed to collect baseline: %v", err)
	}

	if err := saveBaseline(baseline, config); err != nil {
		log.Fatalf("Failed to save baseline: %v", err)
	}

	log.Printf("Baseline collected successfully and saved to: %s", config.OutputFile)
	printSummary(baseline)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.PrometheusURL, "prometheus-url", "http://localhost:9090", "Prometheus server URL")
	flag.IntVar(&config.LookbackHours, "lookback", 24, "Hours to look back for baseline")
	flag.StringVar(&config.OutputFormat, "format", "json", "Output format (json, csv)")
	flag.StringVar(&config.OutputFile, "output", "", "Output file (default: baseline_YYYYMMDD_HHMMSS.json)")
	flag.Parse()

	if config.OutputFile == "" {
		config.OutputFile = fmt.Sprintf("baseline_%s.%s",
			time.Now().Format("20060102_150405"),
			config.OutputFormat)
	}

	return config
}

func collectBaseline(config *Config) (*PerformanceBaseline, error) {
	client, err := api.NewClient(api.Config{
		Address: config.PrometheusURL,
	})
	if err != nil {
		return nil, fmt.Errorf("error creating Prometheus client: %w", err)
	}

	v1api := v1.NewAPI(client)
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	baseline := &PerformanceBaseline{
		Timestamp: time.Now(),
		Period:    fmt.Sprintf("%dh", config.LookbackHours),
		Metrics:   make(map[string]MetricStats),
		Endpoints: make(map[string]EndpointMetrics),
	}

	// Collect endpoint metrics
	endpoints := []string{"shorten", "redirect", "analytics"}
	for _, endpoint := range endpoints {
		metrics, err := collectEndpointMetrics(ctx, v1api, endpoint, config.LookbackHours)
		if err != nil {
			log.Printf("Warning: Failed to collect metrics for endpoint %s: %v", endpoint, err)
			continue
		}
		baseline.Endpoints[endpoint] = metrics
	}

	// Collect infrastructure metrics
	infraMetrics, err := collectInfraMetrics(ctx, v1api, config.LookbackHours)
	if err != nil {
		log.Printf("Warning: Failed to collect infrastructure metrics: %v", err)
	} else {
		baseline.Infrastructure = infraMetrics
	}

	// Collect cost metrics
	costMetrics, err := collectCostMetrics(ctx, v1api, config.LookbackHours)
	if err != nil {
		log.Printf("Warning: Failed to collect cost metrics: %v", err)
	} else {
		baseline.Cost = costMetrics
	}

	return baseline, nil
}

func collectEndpointMetrics(ctx context.Context, api v1.API, endpoint string, hours int) (EndpointMetrics, error) {
	metrics := EndpointMetrics{}

	timeRange := v1.Range{
		Start: time.Now().Add(-time.Duration(hours) * time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}

	// Query latency percentiles
	latencyQuery := fmt.Sprintf(`histogram_quantile(0.95,
		sum(rate(http_request_duration_seconds_bucket{endpoint="%s"}[5m])) by (le))`, endpoint)

	result, _, err := api.QueryRange(ctx, latencyQuery, timeRange)
	if err != nil {
		return metrics, fmt.Errorf("error querying latency: %w", err)
	}

	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		metrics.Latency = calculateStats(matrix[0].Values)
	}

	// Query throughput
	throughputQuery := fmt.Sprintf(`sum(rate(http_requests_total{endpoint="%s"}[5m]))`, endpoint)
	result, _, err = api.QueryRange(ctx, throughputQuery, timeRange)
	if err != nil {
		return metrics, fmt.Errorf("error querying throughput: %w", err)
	}

	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		metrics.Throughput = calculateAverage(matrix[0].Values)
	}

	// Query error rate
	errorQuery := fmt.Sprintf(`sum(rate(http_requests_total{endpoint="%s",status=~"5.."}[5m])) /
		sum(rate(http_requests_total{endpoint="%s"}[5m]))`, endpoint, endpoint)

	result, _, err = api.QueryRange(ctx, errorQuery, timeRange)
	if err != nil {
		return metrics, fmt.Errorf("error querying error rate: %w", err)
	}

	if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
		metrics.ErrorRate = calculateAverage(matrix[0].Values) * 100 // Convert to percentage
	}

	// Query cache hit rate
	cacheQuery := fmt.Sprintf(`sum(rate(cache_hits_total{endpoint="%s"}[5m])) /
		(sum(rate(cache_hits_total{endpoint="%s"}[5m])) +
		 sum(rate(cache_misses_total{endpoint="%s"}[5m])))`, endpoint, endpoint, endpoint)

	result, _, err = api.QueryRange(ctx, cacheQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.CacheHitRate = calculateAverage(matrix[0].Values) * 100 // Convert to percentage
		}
	}

	return metrics, nil
}

func collectInfraMetrics(ctx context.Context, api v1.API, hours int) (InfraMetrics, error) {
	metrics := InfraMetrics{}

	timeRange := v1.Range{
		Start: time.Now().Add(-time.Duration(hours) * time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}

	// CPU utilization
	cpuQuery := `avg(rate(process_cpu_seconds_total[5m])) * 100`
	result, _, err := api.QueryRange(ctx, cpuQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.CPUUtilization = calculateAverage(matrix[0].Values)
		}
	}

	// Memory utilization
	memQuery := `avg(process_resident_memory_bytes / 1024 / 1024)`
	result, _, err = api.QueryRange(ctx, memQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.MemoryUtilization = calculateAverage(matrix[0].Values)
		}
	}

	// Network metrics (if available)
	// Active connections
	connQuery := `sum(http_active_connections)`
	result, _, err = api.QueryRange(ctx, connQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.ActiveConnections = int64(calculateAverage(matrix[0].Values))
		}
	}

	// Goroutines
	goroutineQuery := `sum(go_goroutines)`
	result, _, err = api.QueryRange(ctx, goroutineQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.ActiveGoroutines = int64(calculateAverage(matrix[0].Values))
		}
	}

	return metrics, nil
}

func collectCostMetrics(ctx context.Context, api v1.API, hours int) (CostMetrics, error) {
	metrics := CostMetrics{}

	timeRange := v1.Range{
		Start: time.Now().Add(-time.Duration(hours) * time.Hour),
		End:   time.Now(),
		Step:  time.Minute,
	}

	// Total requests per day
	requestQuery := `sum(increase(http_requests_total[24h]))`
	result, _, err := api.QueryRange(ctx, requestQuery, timeRange)
	if err == nil {
		if matrix, ok := result.(model.Matrix); ok && len(matrix) > 0 {
			metrics.TotalRequestsPerDay = int64(calculateAverage(matrix[0].Values))
		}
	}

	// Note: Cost per million requests would need to be calculated from infrastructure costs
	// This is a placeholder - implement based on your cost tracking
	metrics.CostPerMillionRequests = 0.0
	metrics.EstimatedDailyCost = 0.0

	return metrics, nil
}

func calculateStats(values []model.SamplePair) MetricStats {
	if len(values) == 0 {
		return MetricStats{}
	}

	sorted := make([]float64, len(values))
	var sum float64
	min := float64(values[0].Value)
	max := float64(values[0].Value)

	for i, v := range values {
		val := float64(v.Value)
		sorted[i] = val
		sum += val
		if val < min {
			min = val
		}
		if val > max {
			max = val
		}
	}

	// Simple bubble sort for percentile calculation
	for i := 0; i < len(sorted); i++ {
		for j := i + 1; j < len(sorted); j++ {
			if sorted[i] > sorted[j] {
				sorted[i], sorted[j] = sorted[j], sorted[i]
			}
		}
	}

	return MetricStats{
		P50:   sorted[len(sorted)*50/100],
		P95:   sorted[len(sorted)*95/100],
		P99:   sorted[len(sorted)*99/100],
		Mean:  sum / float64(len(values)),
		Min:   min,
		Max:   max,
		Count: int64(len(values)),
	}
}

func calculateAverage(values []model.SamplePair) float64 {
	if len(values) == 0 {
		return 0
	}

	var sum float64
	for _, v := range values {
		sum += float64(v.Value)
	}

	return sum / float64(len(values))
}

func saveBaseline(baseline *PerformanceBaseline, config *Config) error {
	var data []byte
	var err error

	switch config.OutputFormat {
	case "json":
		data, err = json.MarshalIndent(baseline, "", "  ")
		if err != nil {
			return fmt.Errorf("error marshaling to JSON: %w", err)
		}
	default:
		return fmt.Errorf("unsupported output format: %s", config.OutputFormat)
	}

	if err := os.WriteFile(config.OutputFile, data, 0644); err != nil {
		return fmt.Errorf("error writing to file: %w", err)
	}

	return nil
}

func printSummary(baseline *PerformanceBaseline) {
	fmt.Println("\n=== Performance Baseline Summary ===")
	fmt.Printf("Timestamp: %s\n", baseline.Timestamp.Format(time.RFC3339))
	fmt.Printf("Period: %s\n\n", baseline.Period)

	fmt.Println("Endpoint Performance:")
	for endpoint, metrics := range baseline.Endpoints {
		fmt.Printf("\n  %s:\n", endpoint)
		fmt.Printf("    Latency p95: %.2fms\n", metrics.Latency.P95*1000)
		fmt.Printf("    Throughput:  %.2f req/s\n", metrics.Throughput)
		fmt.Printf("    Error Rate:  %.3f%%\n", metrics.ErrorRate)
		fmt.Printf("    Cache Hit:   %.2f%%\n", metrics.CacheHitRate)
	}

	fmt.Println("\nInfrastructure:")
	fmt.Printf("  CPU Utilization:  %.2f%%\n", baseline.Infrastructure.CPUUtilization)
	fmt.Printf("  Memory Usage:     %.2f MB\n", baseline.Infrastructure.MemoryUtilization)
	fmt.Printf("  Active Conns:     %d\n", baseline.Infrastructure.ActiveConnections)
	fmt.Printf("  Goroutines:       %d\n", baseline.Infrastructure.ActiveGoroutines)

	if baseline.Cost.TotalRequestsPerDay > 0 {
		fmt.Println("\nCost Metrics:")
		fmt.Printf("  Requests/Day:     %d\n", baseline.Cost.TotalRequestsPerDay)
		if baseline.Cost.CostPerMillionRequests > 0 {
			fmt.Printf("  Cost/Million:     $%.2f\n", baseline.Cost.CostPerMillionRequests)
			fmt.Printf("  Est. Daily Cost:  $%.2f\n", baseline.Cost.EstimatedDailyCost)
		}
	}

	fmt.Println("\n====================================")
}
