package main

import (
	"encoding/json"
	"flag"
	"fmt"
	"log"
	"math"
	"os"
	"time"
)

// CapacityPlan represents a capacity planning forecast
type CapacityPlan struct {
	GeneratedAt       time.Time          `json:"generated_at"`
	LookbackDays      int                `json:"lookback_days"`
	ForecastDays      int                `json:"forecast_days"`
	Scenario          string             `json:"scenario"`
	CurrentState      ResourceState      `json:"current_state"`
	Forecast          []ForecastPeriod   `json:"forecast"`
	Recommendations   []Recommendation   `json:"recommendations"`
	CostProjection    CostProjection     `json:"cost_projection"`
}

// ResourceState represents current resource utilization
type ResourceState struct {
	Timestamp         time.Time `json:"timestamp"`
	RequestsPerSecond float64   `json:"requests_per_second"`
	CPUUtilization    float64   `json:"cpu_utilization"`
	MemoryUtilization float64   `json:"memory_utilization"`
	StorageGB         float64   `json:"storage_gb"`
	NetworkMbps       float64   `json:"network_mbps"`
	InstanceCount     int       `json:"instance_count"`
}

// ForecastPeriod represents projected capacity needs for a time period
type ForecastPeriod struct {
	Date                  time.Time `json:"date"`
	ProjectedRPS          float64   `json:"projected_rps"`
	ProjectedCPU          float64   `json:"projected_cpu"`
	ProjectedMemory       float64   `json:"projected_memory"`
	ProjectedStorageGB    float64   `json:"projected_storage_gb"`
	RequiredInstances     int       `json:"required_instances"`
	CapacityMargin        float64   `json:"capacity_margin"`
	ScalingActionNeeded   bool      `json:"scaling_action_needed"`
}

// Recommendation represents a capacity recommendation
type Recommendation struct {
	Date        time.Time `json:"date"`
	Priority    string    `json:"priority"`
	Action      string    `json:"action"`
	Description string    `json:"description"`
	Impact      string    `json:"impact"`
	EstimatedCost float64 `json:"estimated_cost"`
}

// CostProjection represents projected costs
type CostProjection struct {
	CurrentMonthlyCost    float64            `json:"current_monthly_cost"`
	ProjectedMonthlyCost  map[string]float64 `json:"projected_monthly_cost"`
	CostIncrease          float64            `json:"cost_increase"`
	OptimizationPotential float64            `json:"optimization_potential"`
}

// GrowthScenario defines traffic growth assumptions
type GrowthScenario struct {
	Name               string
	MonthlyGrowthRate  float64 // Percentage
	StorageGrowthRateGB float64 // GB per month
}

var scenarios = map[string]GrowthScenario{
	"conservative": {
		Name:               "Conservative",
		MonthlyGrowthRate:  10.0,
		StorageGrowthRateGB: 5.0,
	},
	"expected": {
		Name:               "Expected",
		MonthlyGrowthRate:  20.0,
		StorageGrowthRateGB: 10.0,
	},
	"optimistic": {
		Name:               "Optimistic",
		MonthlyGrowthRate:  35.0,
		StorageGrowthRateGB: 20.0,
	},
}

type Config struct {
	InputFile     string
	OutputFile    string
	ForecastDays  int
	Scenario      string
	SafetyMargin  float64
	CPUThreshold  float64
	MemThreshold  float64
	CostPerInstance float64
}

func main() {
	config := parseFlags()

	log.Printf("Starting capacity planning analysis")
	log.Printf("Scenario: %s", config.Scenario)
	log.Printf("Forecast period: %d days", config.ForecastDays)

	currentState, err := loadCurrentState(config.InputFile)
	if err != nil {
		log.Fatalf("Failed to load current state: %v", err)
	}

	plan := generateCapacityPlan(currentState, config)

	if err := savePlan(plan, config.OutputFile); err != nil {
		log.Fatalf("Failed to save plan: %v", err)
	}

	printPlanSummary(plan)
	log.Printf("Capacity plan saved to: %s", config.OutputFile)
}

func parseFlags() *Config {
	config := &Config{}

	flag.StringVar(&config.InputFile, "input", "baseline_latest.json", "Input baseline file")
	flag.StringVar(&config.OutputFile, "output", "", "Output capacity plan file")
	flag.IntVar(&config.ForecastDays, "forecast", 90, "Days to forecast")
	flag.StringVar(&config.Scenario, "scenario", "expected", "Growth scenario (conservative/expected/optimistic)")
	flag.Float64Var(&config.SafetyMargin, "safety-margin", 1.5, "Safety margin multiplier")
	flag.Float64Var(&config.CPUThreshold, "cpu-threshold", 70.0, "CPU utilization threshold for scaling")
	flag.Float64Var(&config.MemThreshold, "mem-threshold", 80.0, "Memory utilization threshold for scaling")
	flag.Float64Var(&config.CostPerInstance, "cost-per-instance", 50.0, "Monthly cost per instance")
	flag.Parse()

	if config.OutputFile == "" {
		config.OutputFile = fmt.Sprintf("capacity_plan_%s_%s.json",
			config.Scenario,
			time.Now().Format("20060102"))
	}

	return config
}

func loadCurrentState(filename string) (*ResourceState, error) {
	// For now, use dummy data. In production, load from baseline file
	state := &ResourceState{
		Timestamp:         time.Now(),
		RequestsPerSecond: 1000.0,
		CPUUtilization:    50.0,
		MemoryUtilization: 60.0,
		StorageGB:         100.0,
		NetworkMbps:       50.0,
		InstanceCount:     4,
	}

	// TODO: Load actual data from file
	// data, err := os.ReadFile(filename)
	// if err != nil {
	//     return nil, err
	// }
	// err = json.Unmarshal(data, &state)

	return state, nil
}

func generateCapacityPlan(currentState *ResourceState, config *Config) *CapacityPlan {
	scenario := scenarios[config.Scenario]

	plan := &CapacityPlan{
		GeneratedAt:  time.Now(),
		ForecastDays: config.ForecastDays,
		Scenario:     scenario.Name,
		CurrentState: *currentState,
		Forecast:     make([]ForecastPeriod, 0),
		Recommendations: make([]Recommendation, 0),
	}

	// Generate daily forecasts
	for day := 1; day <= config.ForecastDays; day++ {
		forecastDate := time.Now().AddDate(0, 0, day)

		// Calculate growth factor
		dailyGrowthRate := scenario.MonthlyGrowthRate / 30.0 / 100.0
		growthFactor := math.Pow(1+dailyGrowthRate, float64(day))

		// Project resources
		projectedRPS := currentState.RequestsPerSecond * growthFactor
		projectedCPU := currentState.CPUUtilization * growthFactor
		projectedMem := currentState.MemoryUtilization * growthFactor
		projectedStorage := currentState.StorageGB + (scenario.StorageGrowthRateGB / 30.0 * float64(day))

		// Calculate required instances
		cpuInstances := int(math.Ceil(projectedCPU / config.CPUThreshold))
		memInstances := int(math.Ceil(projectedMem / config.MemThreshold))
		requiredInstances := int(math.Max(float64(cpuInstances), float64(memInstances)))

		// Apply safety margin
		requiredInstances = int(math.Ceil(float64(requiredInstances) * config.SafetyMargin))

		// Calculate capacity margin
		capacityMargin := float64(requiredInstances) / float64(currentState.InstanceCount) - 1.0

		// Determine if scaling needed
		scalingNeeded := requiredInstances > currentState.InstanceCount

		period := ForecastPeriod{
			Date:                forecastDate,
			ProjectedRPS:        projectedRPS,
			ProjectedCPU:        projectedCPU / float64(requiredInstances), // Per instance
			ProjectedMemory:     projectedMem / float64(requiredInstances), // Per instance
			ProjectedStorageGB:  projectedStorage,
			RequiredInstances:   requiredInstances,
			CapacityMargin:      capacityMargin,
			ScalingActionNeeded: scalingNeeded,
		}

		plan.Forecast = append(plan.Forecast, period)

		// Generate recommendations at key thresholds
		if day%30 == 0 && scalingNeeded { // Monthly check
			recommendation := Recommendation{
				Date:     forecastDate,
				Priority: determinePriority(capacityMargin),
				Action:   "Scale Up",
				Description: fmt.Sprintf("Increase instance count from %d to %d to handle projected load",
					currentState.InstanceCount, requiredInstances),
				Impact: fmt.Sprintf("CPU: %.1f%%, Memory: %.1f%%",
					period.ProjectedCPU, period.ProjectedMemory),
				EstimatedCost: float64(requiredInstances-currentState.InstanceCount) * config.CostPerInstance,
			}
			plan.Recommendations = append(plan.Recommendations, recommendation)
		}
	}

	// Generate cost projection
	plan.CostProjection = generateCostProjection(currentState, plan.Forecast, config)

	return plan
}

func determinePriority(capacityMargin float64) string {
	if capacityMargin > 0.5 {
		return "P0-Critical"
	} else if capacityMargin > 0.3 {
		return "P1-High"
	} else if capacityMargin > 0.1 {
		return "P2-Medium"
	}
	return "P3-Low"
}

func generateCostProjection(currentState *ResourceState, forecast []ForecastPeriod, config *Config) CostProjection {
	currentCost := float64(currentState.InstanceCount) * config.CostPerInstance

	projection := CostProjection{
		CurrentMonthlyCost:   currentCost,
		ProjectedMonthlyCost: make(map[string]float64),
	}

	// Calculate costs at 30, 60, 90 days
	for _, day := range []int{30, 60, 90} {
		if day <= len(forecast) {
			period := forecast[day-1]
			monthLabel := fmt.Sprintf("month_%d", day/30)
			projection.ProjectedMonthlyCost[monthLabel] = float64(period.RequiredInstances) * config.CostPerInstance
		}
	}

	// Calculate total increase
	if len(forecast) > 0 {
		finalPeriod := forecast[len(forecast)-1]
		finalCost := float64(finalPeriod.RequiredInstances) * config.CostPerInstance
		projection.CostIncrease = ((finalCost - currentCost) / currentCost) * 100
	}

	// Estimate optimization potential (conservative 20%)
	projection.OptimizationPotential = currentCost * 0.20

	return projection
}

func savePlan(plan *CapacityPlan, filename string) error {
	data, err := json.MarshalIndent(plan, "", "  ")
	if err != nil {
		return fmt.Errorf("error marshaling plan: %w", err)
	}

	if err := os.WriteFile(filename, data, 0644); err != nil {
		return fmt.Errorf("error writing file: %w", err)
	}

	return nil
}

func printPlanSummary(plan *CapacityPlan) {
	fmt.Println("\n=== Capacity Plan Summary ===")
	fmt.Printf("Scenario: %s\n", plan.Scenario)
	fmt.Printf("Forecast Period: %d days\n\n", plan.ForecastDays)

	fmt.Println("Current State:")
	fmt.Printf("  Instances:      %d\n", plan.CurrentState.InstanceCount)
	fmt.Printf("  RPS:            %.0f req/s\n", plan.CurrentState.RequestsPerSecond)
	fmt.Printf("  CPU:            %.1f%%\n", plan.CurrentState.CPUUtilization)
	fmt.Printf("  Memory:         %.1f%%\n", plan.CurrentState.MemoryUtilization)
	fmt.Printf("  Storage:        %.1f GB\n\n", plan.CurrentState.StorageGB)

	// Print key forecast points
	fmt.Println("Forecast Milestones:")
	for i, period := range plan.Forecast {
		if (i+1)%30 == 0 { // Print every 30 days
			fmt.Printf("\n  Day %d (%s):\n", i+1, period.Date.Format("2006-01-02"))
			fmt.Printf("    Required Instances: %d (current: %d)\n",
				period.RequiredInstances, plan.CurrentState.InstanceCount)
			fmt.Printf("    Projected RPS:      %.0f req/s\n", period.ProjectedRPS)
			fmt.Printf("    CPU per instance:   %.1f%%\n", period.ProjectedCPU)
			fmt.Printf("    Mem per instance:   %.1f%%\n", period.ProjectedMemory)
			fmt.Printf("    Storage:            %.1f GB\n", period.ProjectedStorageGB)
		}
	}

	fmt.Println("\nRecommendations:")
	if len(plan.Recommendations) == 0 {
		fmt.Println("  No scaling actions needed in forecast period")
	} else {
		for i, rec := range plan.Recommendations {
			fmt.Printf("\n  %d. [%s] %s (by %s)\n", i+1, rec.Priority, rec.Action, rec.Date.Format("2006-01-02"))
			fmt.Printf("     %s\n", rec.Description)
			fmt.Printf("     Impact: %s\n", rec.Impact)
			if rec.EstimatedCost > 0 {
				fmt.Printf("     Additional Cost: $%.2f/month\n", rec.EstimatedCost)
			}
		}
	}

	fmt.Println("\nCost Projection:")
	fmt.Printf("  Current Monthly Cost:    $%.2f\n", plan.CostProjection.CurrentMonthlyCost)
	for period, cost := range plan.CostProjection.ProjectedMonthlyCost {
		fmt.Printf("  Projected Cost (%s): $%.2f\n", period, cost)
	}
	fmt.Printf("  Total Increase:          %.1f%%\n", plan.CostProjection.CostIncrease)
	fmt.Printf("  Optimization Potential:  $%.2f/month\n", plan.CostProjection.OptimizationPotential)

	fmt.Println("\n=============================")
}
