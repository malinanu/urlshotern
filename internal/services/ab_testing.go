package services

import (
	"database/sql"
	"fmt"
	"math"
	"math/rand"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

type ABTestingService struct {
	db    *storage.PostgresStorage
	cache *storage.RedisStorage
	rand  *rand.Rand
}

func NewABTestingService(db *storage.PostgresStorage, cache *storage.RedisStorage) *ABTestingService {
	return &ABTestingService{
		db:    db,
		cache: cache,
		rand:  rand.New(rand.NewSource(time.Now().UnixNano())),
	}
}

// CreateABTest creates a new A/B test
func (a *ABTestingService) CreateABTest(userID int64, request *models.CreateABTestRequest) (*models.ABTest, error) {
	id, err := utils.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate test ID: %w", err)
	}

	// Validate traffic allocation adds up to 100%
	totalTraffic := 0
	for _, variant := range request.Variants {
		totalTraffic += variant.TrafficAllocation
	}
	if totalTraffic != 100 {
		return nil, fmt.Errorf("traffic allocation must sum to 100%%, got %d%%", totalTraffic)
	}

	test := &models.ABTest{
		ID:             id,
		UserID:         userID,
		TestName:       request.TestName,
		TestType:       request.TestType,
		Status:         "draft",
		TrafficSplit:   request.TrafficSplit,
		StartDate:      request.StartDate,
		EndDate:        request.EndDate,
		SampleSize:     request.SampleSize,
		Confidence:     request.Confidence,
		IsActive:       false,
		CreatedAt:      time.Now(),
		UpdatedAt:      time.Now(),
	}

	// Insert the A/B test
	query := `
		INSERT INTO ab_tests (
			id, user_id, test_name, test_type, status, traffic_split, 
			start_date, end_date, sample_size, confidence, is_active, 
			created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13)
		RETURNING id
	`

	err = a.db.QueryRow(query,
		test.ID, test.UserID, test.TestName, test.TestType, test.Status,
		test.TrafficSplit, test.StartDate, test.EndDate, test.SampleSize,
		test.Confidence, test.IsActive, test.CreatedAt, test.UpdatedAt,
	).Scan(&test.ID)

	if err != nil {
		return nil, fmt.Errorf("failed to create A/B test: %w", err)
	}

	// Create test variants
	for _, variant := range request.Variants {
		variantID, err := utils.GenerateID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate variant ID: %w", err)
		}

		variantQuery := `
			INSERT INTO ab_test_variants (
				id, test_id, variant_name, short_code, traffic_allocation, 
				is_control, created_at
			) VALUES ($1, $2, $3, $4, $5, $6, $7)
		`

		_, err = a.db.Exec(variantQuery,
			variantID, test.ID, variant.VariantName, variant.ShortCode,
			variant.TrafficAllocation, variant.IsControl, time.Now(),
		)

		if err != nil {
			return nil, fmt.Errorf("failed to create test variant: %w", err)
		}
	}

	return test, nil
}

// StartABTest activates an A/B test
func (a *ABTestingService) StartABTest(testID, userID int64) error {
	query := `
		UPDATE ab_tests 
		SET status = 'running', is_active = true, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND status = 'draft'
	`

	result, err := a.db.Exec(query, time.Now(), testID, userID)
	if err != nil {
		return fmt.Errorf("failed to start A/B test: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("A/B test not found or already running")
	}

	return nil
}

// StopABTest stops an active A/B test
func (a *ABTestingService) StopABTest(testID, userID int64) error {
	query := `
		UPDATE ab_tests 
		SET status = 'completed', is_active = false, updated_at = $1
		WHERE id = $2 AND user_id = $3 AND status = 'running'
	`

	result, err := a.db.Exec(query, time.Now(), testID, userID)
	if err != nil {
		return fmt.Errorf("failed to stop A/B test: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("A/B test not found or not running")
	}

	return nil
}

// GetABTestVariant determines which variant a user should see
func (a *ABTestingService) GetABTestVariant(testID int64, sessionID string) (*models.ABTestVariant, error) {
	// Check if user already has an assigned variant (cache first)
	cacheKey := fmt.Sprintf("ab_test:%d:session:%s", testID, sessionID)
	if cachedVariant, err := a.cache.Get(cacheKey); err == nil && cachedVariant != "" {
		// Return cached variant
		return a.getVariantByID(cachedVariant)
	}

	// Get test variants with traffic allocation
	variants, err := a.getTestVariants(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get test variants: %w", err)
	}

	if len(variants) == 0 {
		return nil, fmt.Errorf("no variants found for test")
	}

	// Determine variant using traffic allocation
	selectedVariant := a.selectVariantByTraffic(variants, sessionID)

	// Cache the assignment for 24 hours
	err = a.cache.Set(cacheKey, fmt.Sprintf("%d", selectedVariant.ID), 24*time.Hour)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to cache A/B test assignment: %v\n", err)
	}

	// Record the assignment
	err = a.recordTestAssignment(testID, selectedVariant.ID, sessionID)
	if err != nil {
		// Log error but don't fail the request
		fmt.Printf("Warning: failed to record A/B test assignment: %v\n", err)
	}

	return selectedVariant, nil
}

// RecordConversion records a conversion for an A/B test
func (a *ABTestingService) RecordConversion(testID, variantID int64, sessionID string, conversionValue float64) error {
	// Check if this session already converted for this test (prevent double counting)
	existingQuery := `
		SELECT id FROM ab_test_results 
		WHERE test_id = $1 AND session_id = $2 AND event_type = 'conversion'
	`

	var existingID int64
	err := a.db.QueryRow(existingQuery, testID, sessionID).Scan(&existingID)
	if err == nil {
		// Conversion already exists, don't double count
		return nil
	} else if err != sql.ErrNoRows {
		return fmt.Errorf("failed to check existing conversion: %w", err)
	}

	// Record the conversion
	id, err := utils.GenerateID()
	if err != nil {
		return fmt.Errorf("failed to generate result ID: %w", err)
	}

	query := `
		INSERT INTO ab_test_results (
			id, test_id, variant_id, session_id, event_type, 
			conversion_value, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6, $7)
	`

	_, err = a.db.Exec(query,
		id, testID, variantID, sessionID, "conversion",
		conversionValue, time.Now(),
	)

	if err != nil {
		return fmt.Errorf("failed to record conversion: %w", err)
	}

	return nil
}

// GetABTestResults gets statistical results for an A/B test
func (a *ABTestingService) GetABTestResults(testID, userID int64) (*models.ABTestResults, error) {
	// Get test details
	test, err := a.getABTest(testID, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get A/B test: %w", err)
	}

	// Get variant results
	variantResults, err := a.getVariantResults(testID)
	if err != nil {
		return nil, fmt.Errorf("failed to get variant results: %w", err)
	}

	// Calculate statistical significance
	significance := a.calculateStatisticalSignificance(variantResults, test.Confidence)

	// Determine winner
	winner := a.determineWinner(variantResults, significance)

	results := &models.ABTestResults{
		TestID:           testID,
		TestName:         test.TestName,
		Status:           test.Status,
		StartDate:        test.StartDate,
		EndDate:          test.EndDate,
		TotalSessions:    a.sumSessions(variantResults),
		TotalConversions: a.sumConversions(variantResults),
		OverallCR:        a.calculateOverallCR(variantResults),
		VariantResults:   variantResults,
		IsSignificant:    significance.IsSignificant,
		PValue:           significance.PValue,
		ConfidenceLevel:  test.Confidence,
		Winner:           winner,
		Recommendation:   a.generateRecommendation(variantResults, significance, winner),
	}

	return results, nil
}

// Helper methods

func (a *ABTestingService) getTestVariants(testID int64) ([]*models.ABTestVariant, error) {
	query := `
		SELECT id, test_id, variant_name, short_code, traffic_allocation, 
		       is_control, created_at
		FROM ab_test_variants 
		WHERE test_id = $1 
		ORDER BY traffic_allocation DESC
	`

	rows, err := a.db.Query(query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var variants []*models.ABTestVariant
	for rows.Next() {
		variant := &models.ABTestVariant{}
		err := rows.Scan(
			&variant.ID, &variant.TestID, &variant.VariantName,
			&variant.ShortCode, &variant.TrafficAllocation,
			&variant.IsControl, &variant.CreatedAt,
		)
		if err != nil {
			return nil, err
		}
		variants = append(variants, variant)
	}

	return variants, nil
}

func (a *ABTestingService) selectVariantByTraffic(variants []*models.ABTestVariant, sessionID string) *models.ABTestVariant {
	// Use session ID for consistent assignment
	hash := a.hashString(sessionID)
	bucket := hash % 100

	currentBucket := 0
	for _, variant := range variants {
		currentBucket += variant.TrafficAllocation
		if bucket < currentBucket {
			return variant
		}
	}

	// Fallback to first variant (should not happen)
	return variants[0]
}

func (a *ABTestingService) hashString(s string) int {
	h := 0
	for i := 0; i < len(s); i++ {
		h = 31*h + int(s[i])
	}
	if h < 0 {
		h = -h
	}
	return h
}

func (a *ABTestingService) getVariantByID(variantIDStr string) (*models.ABTestVariant, error) {
	query := `
		SELECT id, test_id, variant_name, short_code, traffic_allocation, 
		       is_control, created_at
		FROM ab_test_variants 
		WHERE id = $1
	`

	variant := &models.ABTestVariant{}
	err := a.db.QueryRow(query, variantIDStr).Scan(
		&variant.ID, &variant.TestID, &variant.VariantName,
		&variant.ShortCode, &variant.TrafficAllocation,
		&variant.IsControl, &variant.CreatedAt,
	)

	if err != nil {
		return nil, err
	}

	return variant, nil
}

func (a *ABTestingService) recordTestAssignment(testID, variantID int64, sessionID string) error {
	id, err := utils.GenerateID()
	if err != nil {
		return err
	}

	query := `
		INSERT INTO ab_test_results (
			id, test_id, variant_id, session_id, event_type, timestamp
		) VALUES ($1, $2, $3, $4, $5, $6)
		ON CONFLICT (test_id, session_id, event_type) DO NOTHING
	`

	_, err = a.db.Exec(query, id, testID, variantID, sessionID, "assignment", time.Now())
	return err
}

func (a *ABTestingService) getABTest(testID, userID int64) (*models.ABTest, error) {
	query := `
		SELECT id, user_id, test_name, test_type, status, traffic_split,
		       start_date, end_date, sample_size, confidence, is_active,
		       created_at, updated_at
		FROM ab_tests
		WHERE id = $1 AND user_id = $2
	`

	test := &models.ABTest{}
	err := a.db.QueryRow(query, testID, userID).Scan(
		&test.ID, &test.UserID, &test.TestName, &test.TestType,
		&test.Status, &test.TrafficSplit, &test.StartDate, &test.EndDate,
		&test.SampleSize, &test.Confidence, &test.IsActive,
		&test.CreatedAt, &test.UpdatedAt,
	)

	return test, err
}

func (a *ABTestingService) getVariantResults(testID int64) ([]*models.VariantResult, error) {
	query := `
		SELECT 
			v.id, v.variant_name, v.is_control,
			COUNT(CASE WHEN r.event_type = 'assignment' THEN 1 END) as sessions,
			COUNT(CASE WHEN r.event_type = 'conversion' THEN 1 END) as conversions,
			COALESCE(SUM(CASE WHEN r.event_type = 'conversion' THEN r.conversion_value ELSE 0 END), 0) as revenue
		FROM ab_test_variants v
		LEFT JOIN ab_test_results r ON v.id = r.variant_id
		WHERE v.test_id = $1
		GROUP BY v.id, v.variant_name, v.is_control
		ORDER BY v.is_control DESC, v.variant_name
	`

	rows, err := a.db.Query(query, testID)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var results []*models.VariantResult
	for rows.Next() {
		result := &models.VariantResult{}
		err := rows.Scan(
			&result.VariantID, &result.VariantName, &result.IsControl,
			&result.Sessions, &result.Conversions, &result.Revenue,
		)
		if err != nil {
			return nil, err
		}

		// Calculate conversion rate
		if result.Sessions > 0 {
			result.ConversionRate = float64(result.Conversions) / float64(result.Sessions)
		}

		// Calculate average order value
		if result.Conversions > 0 {
			result.AOV = result.Revenue / float64(result.Conversions)
		}

		results = append(results, result)
	}

	return results, nil
}

// Statistical analysis methods
func (a *ABTestingService) calculateStatisticalSignificance(results []*models.VariantResult, confidence float64) *models.SignificanceResult {
	if len(results) < 2 {
		return &models.SignificanceResult{IsSignificant: false}
	}

	// Find control and best variant
	var control *models.VariantResult
	var bestVariant *models.VariantResult

	for _, result := range results {
		if result.IsControl {
			control = result
		} else if bestVariant == nil || result.ConversionRate > bestVariant.ConversionRate {
			bestVariant = result
		}
	}

	if control == nil || bestVariant == nil {
		return &models.SignificanceResult{IsSignificant: false}
	}

	// Calculate z-score using two-proportion z-test
	p1 := control.ConversionRate
	p2 := bestVariant.ConversionRate
	n1 := control.Sessions
	n2 := bestVariant.Sessions

	if n1 == 0 || n2 == 0 {
		return &models.SignificanceResult{IsSignificant: false}
	}

	// Pooled proportion
	pPooled := (float64(control.Conversions) + float64(bestVariant.Conversions)) / (float64(n1) + float64(n2))
	
	// Standard error
	se := math.Sqrt(pPooled * (1 - pPooled) * (1/float64(n1) + 1/float64(n2)))
	
	if se == 0 {
		return &models.SignificanceResult{IsSignificant: false}
	}

	// Z-score
	zScore := (p2 - p1) / se
	
	// Convert to p-value (two-tailed test)
	pValue := 2 * (1 - a.normalCDF(math.Abs(zScore)))
	
	// Check significance
	alpha := 1 - (confidence / 100)
	isSignificant := pValue < alpha

	// Calculate confidence interval for difference
	confidenceInterval := a.calculateConfidenceInterval(p1, p2, int64(n1), int64(n2), confidence)
	
	// Calculate effect size (Cohen's h for proportions)
	effectSize := a.calculateCohenH(p1, p2)
	
	// Calculate minimum detectable effect
	mde := a.calculateMDE(int64(n1), int64(n2), confidence, 80.0) // 80% power
	
	return &models.SignificanceResult{
		IsSignificant:       isSignificant,
		PValue:             pValue,
		ZScore:             zScore,
		ControlCR:          p1,
		VariantCR:          p2,
		Improvement:        ((p2 - p1) / p1) * 100, // Percentage improvement
		EffectSize:         effectSize,
		ConfidenceInterval: confidenceInterval,
		MinDetectableEffect: mde,
		SampleSizeRecommendation: a.RecommendSampleSize(p1, 0.05, confidence, 80.0), // 5% relative improvement
	}
}

// Approximate normal CDF using error function approximation
func (a *ABTestingService) normalCDF(z float64) float64 {
	return 0.5 * (1 + a.erf(z/math.Sqrt(2)))
}

// Error function approximation
func (a *ABTestingService) erf(x float64) float64 {
	// Abramowitz and Stegun approximation
	a1 := 0.254829592
	a2 := -0.284496736
	a3 := 1.421413741
	a4 := -1.453152027
	a5 := 1.061405429
	p := 0.3275911

	sign := 1.0
	if x < 0 {
		sign = -1.0
	}
	x = math.Abs(x)

	t := 1.0 / (1.0 + p*x)
	y := 1.0 - (((((a5*t+a4)*t)+a3)*t+a2)*t+a1)*t*math.Exp(-x*x)

	return sign * y
}

func (a *ABTestingService) determineWinner(results []*models.VariantResult, significance *models.SignificanceResult) *string {
	if !significance.IsSignificant {
		return nil // No clear winner
	}

	var bestVariant *models.VariantResult
	for _, result := range results {
		if bestVariant == nil || result.ConversionRate > bestVariant.ConversionRate {
			bestVariant = result
		}
	}

	if bestVariant != nil {
		return &bestVariant.VariantName
	}

	return nil
}

func (a *ABTestingService) generateRecommendation(results []*models.VariantResult, significance *models.SignificanceResult, winner *string) string {
	if winner == nil {
		return "Continue running the test. No statistically significant winner has been determined yet."
	}

	improvement := significance.Improvement
	if improvement > 0 {
		return fmt.Sprintf("Implement %s. It shows a %.2f%% improvement over control with statistical significance.", *winner, improvement)
	}

	return "Consider running the test longer or with more traffic to achieve statistical significance."
}

func (a *ABTestingService) sumSessions(results []*models.VariantResult) int {
	total := 0
	for _, result := range results {
		total += result.Sessions
	}
	return total
}

func (a *ABTestingService) sumConversions(results []*models.VariantResult) int {
	total := 0
	for _, result := range results {
		total += result.Conversions
	}
	return total
}

func (a *ABTestingService) calculateOverallCR(results []*models.VariantResult) float64 {
	totalSessions := a.sumSessions(results)
	totalConversions := a.sumConversions(results)
	
	if totalSessions == 0 {
		return 0
	}
	
	return float64(totalConversions) / float64(totalSessions)
}

// calculateConfidenceInterval calculates the confidence interval for the difference in conversion rates
func (a *ABTestingService) calculateConfidenceInterval(p1, p2 float64, n1, n2 int64, confidence float64) [2]float64 {
	diff := p2 - p1
	
	// Standard error for difference of proportions
	se1 := math.Sqrt((p1 * (1 - p1)) / float64(n1))
	se2 := math.Sqrt((p2 * (1 - p2)) / float64(n2))
	seDiff := math.Sqrt(se1*se1 + se2*se2)
	
	// Calculate z-score for confidence level
	alpha := 1 - (confidence / 100)
	zAlpha := a.normalInverse(1 - alpha/2)
	
	margin := zAlpha * seDiff
	
	return [2]float64{diff - margin, diff + margin}
}

// calculateCohenH calculates Cohen's h effect size for proportions
func (a *ABTestingService) calculateCohenH(p1, p2 float64) float64 {
	if p1 <= 0 || p1 >= 1 || p2 <= 0 || p2 >= 1 {
		return 0 // Invalid proportions
	}
	
	phi1 := 2 * math.Asin(math.Sqrt(p1))
	phi2 := 2 * math.Asin(math.Sqrt(p2))
	
	return phi2 - phi1
}

// calculateMDE calculates the Minimum Detectable Effect given sample sizes
func (a *ABTestingService) calculateMDE(n1, n2 int64, confidence, power float64) float64 {
	if n1 == 0 || n2 == 0 {
		return 0
	}
	
	alpha := 1 - (confidence / 100)
	beta := 1 - (power / 100)
	
	zAlpha := a.normalInverse(1 - alpha/2)
	zBeta := a.normalInverse(1 - beta)
	
	// Assuming equal sample sizes and baseline conversion rate around 0.1
	p := 0.1
	pooledSE := math.Sqrt(2 * p * (1 - p) / float64(n1))
	
	return (zAlpha + zBeta) * pooledSE
}

// RecommendSampleSize calculates recommended sample size per variant
func (a *ABTestingService) RecommendSampleSize(baselineRate, minEffect, confidence, power float64) int64 {
	if baselineRate <= 0 || baselineRate >= 1 {
		return 0
	}
	
	alpha := 1 - (confidence / 100)
	beta := 1 - (power / 100)
	
	zAlpha := a.normalInverse(1 - alpha/2)
	zBeta := a.normalInverse(1 - beta)
	
	p1 := baselineRate
	p2 := baselineRate * (1 + minEffect)
	
	if p2 >= 1 {
		p2 = 0.99 // Cap at 99%
	}
	
	pooledP := (p1 + p2) / 2
	
	numerator := math.Pow(zAlpha*math.Sqrt(2*pooledP*(1-pooledP)) + zBeta*math.Sqrt(p1*(1-p1)+p2*(1-p2)), 2)
	denominator := math.Pow(p2-p1, 2)
	
	if denominator == 0 {
		return 0
	}
	
	sampleSize := numerator / denominator
	
	return int64(math.Ceil(sampleSize))
}

// normalInverse calculates the inverse normal distribution (approximate)
func (a *ABTestingService) normalInverse(p float64) float64 {
	if p <= 0.5 {
		return -a.normalInverse(1 - p)
	}
	
	// Beasley-Springer-Moro approximation
	c0 := 2.515517
	c1 := 0.802853
	c2 := 0.010328
	d1 := 1.432788
	d2 := 0.189269
	d3 := 0.001308
	
	t := math.Sqrt(-2 * math.Log(1-p))
	
	return t - (c0+c1*t+c2*t*t)/(1+d1*t+d2*t*t+d3*t*t*t)
}

// CalculateSequentialTest implements sequential testing for early stopping
func (a *ABTestingService) CalculateSequentialTest(testID, userID int64) (*models.SequentialTestResult, error) {
	results, err := a.GetABTestResults(testID, userID)
	if err != nil {
		return nil, err
	}
	
	// Find control and test variants
	var control, test *models.VariantResult
	for _, variant := range results.VariantResults {
		if variant.IsControl {
			control = variant
		} else {
			test = variant
		}
	}
	
	if control == nil || test == nil {
		return nil, fmt.Errorf("control or test variant not found")
	}
	
	// Sequential probability ratio test boundaries
	alpha := 0.05  // Type I error
	beta := 0.20   // Type II error (80% power)
	
	logA := math.Log(beta / (1 - alpha))
	logB := math.Log((1 - beta) / alpha)
	
	// Current log-likelihood ratio
	n1, x1 := control.Sessions, control.Conversions
	n2, x2 := test.Sessions, test.Conversions
	
	if n1 == 0 || n2 == 0 {
		return &models.SequentialTestResult{
			CanStop:      false,
			Decision:     "continue",
			Confidence:   0,
			Reason:       "Insufficient data",
		}, nil
	}
	
	p1 := float64(x1) / float64(n1)
	p2 := float64(x2) / float64(n2)
	
	// Calculate log-likelihood ratio
	logLR := a.calculateLogLikelihoodRatio(x1, n1, x2, n2, p1, p2)
	
	var decision string
	var confidence float64
	var canStop bool
	
	if logLR >= logB {
		decision = "test_wins"
		confidence = 95.0
		canStop = true
	} else if logLR <= logA {
		decision = "control_wins"
		confidence = 95.0
		canStop = true
	} else {
		decision = "continue"
		confidence = 0
		canStop = false
	}
	
	return &models.SequentialTestResult{
		CanStop:     canStop,
		Decision:    decision,
		Confidence:  confidence,
		LogLR:       logLR,
		UpperBound:  logB,
		LowerBound:  logA,
		Reason:      a.getSequentialTestReason(decision, logLR, logA, logB),
	}, nil
}

// calculateLogLikelihoodRatio calculates the log-likelihood ratio for sequential testing
func (a *ABTestingService) calculateLogLikelihoodRatio(x1, n1, x2, n2 int, p1, p2 float64) float64 {
	if p1 == 0 || p1 == 1 || p2 == 0 || p2 == 1 {
		return 0
	}
	
	// Under H0: p1 = p2 = p_pooled
	pPooled := float64(x1+x2) / float64(n1+n2)
	
	// Under H1: p1 != p2
	logL1 := float64(x1)*math.Log(p1) + float64(n1-x1)*math.Log(1-p1) +
			 float64(x2)*math.Log(p2) + float64(n2-x2)*math.Log(1-p2)
	
	logL0 := float64(x1)*math.Log(pPooled) + float64(n1-x1)*math.Log(1-pPooled) +
			 float64(x2)*math.Log(pPooled) + float64(n2-x2)*math.Log(1-pPooled)
	
	return logL1 - logL0
}

// getSequentialTestReason provides human-readable reason for sequential test decision
func (a *ABTestingService) getSequentialTestReason(decision string, logLR, logA, logB float64) string {
	switch decision {
	case "test_wins":
		return fmt.Sprintf("Test variant shows significant improvement (LR: %.3f > %.3f)", logLR, logB)
	case "control_wins":
		return fmt.Sprintf("Control performs significantly better (LR: %.3f < %.3f)", logLR, logA)
	default:
		progress := (logLR - logA) / (logB - logA) * 100
		return fmt.Sprintf("Continue testing - %.1f%% progress toward decision boundary", math.Max(0, progress))
	}
}