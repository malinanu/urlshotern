package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

// AttributionService handles multi-touch attribution modeling
type AttributionService struct {
	storage           *storage.PostgresStorage
	conversionService *ConversionTrackingService
}

// AttributionModel represents different attribution models
type AttributionModel string

const (
	FirstTouchAttribution  AttributionModel = "first_touch"
	LastTouchAttribution   AttributionModel = "last_touch"
	LinearAttribution      AttributionModel = "linear"
	TimeDecayAttribution   AttributionModel = "time_decay"
	PositionBasedAttribution AttributionModel = "position_based"
	DataDrivenAttribution  AttributionModel = "data_driven"
)

// TouchpointValue represents the attribution value assigned to a touchpoint
type TouchpointValue struct {
	TouchpointID     int64   `json:"touchpoint_id"`
	ShortCode        string  `json:"short_code"`
	AttributionValue float64 `json:"attribution_value"`
	AttributionModel string  `json:"attribution_model"`
	Weight           float64 `json:"weight"`
}

// AttributionReport represents a comprehensive attribution analysis
type AttributionReport struct {
	ConversionID     string                        `json:"conversion_id"`
	TotalValue       float64                       `json:"total_value"`
	Journey          models.ConversionJourney      `json:"journey"`
	AttributionBreakdown map[string][]TouchpointValue `json:"attribution_breakdown"` // model -> touchpoints
	ModelComparison  map[string]float64             `json:"model_comparison"`       // model -> total attributed value
	RecommendedModel string                         `json:"recommended_model"`
}

// ChannelAttribution represents attribution by marketing channel
type ChannelAttribution struct {
	Channel          string                         `json:"channel"`
	Source           string                         `json:"source"`
	Medium           string                         `json:"medium"`
	Touchpoints      int                           `json:"touchpoints"`
	Conversions      int                           `json:"conversions"`
	AttributionValue map[string]float64            `json:"attribution_value"` // model -> value
	ROI              float64                       `json:"roi"`
	ConversionRate   float64                       `json:"conversion_rate"`
}

// NewAttributionService creates a new attribution service
func NewAttributionService(storage *storage.PostgresStorage, conversionService *ConversionTrackingService) *AttributionService {
	return &AttributionService{
		storage:           storage,
		conversionService: conversionService,
	}
}

// RecordTouchpoint records a touchpoint in the customer journey
func (a *AttributionService) RecordTouchpoint(touchpoint *models.AttributionTouchpoint) error {
	// Start transaction
	tx, err := a.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Insert the touchpoint
	query := `
		INSERT INTO attribution_touchpoints (
			session_id, short_code, user_ip, user_agent, referrer,
			campaign_source, campaign_medium, campaign_name,
			touchpoint_order, touchpoint_time, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, NOW())
		RETURNING id
	`

	err = tx.QueryRow(query,
		touchpoint.SessionID, touchpoint.ShortCode, touchpoint.UserIP,
		touchpoint.UserAgent, touchpoint.Referrer, touchpoint.CampaignSource,
		touchpoint.CampaignMedium, touchpoint.CampaignName,
		touchpoint.TouchpointOrder, touchpoint.TouchpointTime,
	).Scan(&touchpoint.ID)

	if err != nil {
		return fmt.Errorf("failed to insert touchpoint: %w", err)
	}

	return tx.Commit()
}

// CalculateAttribution calculates attribution for a conversion using specified model
func (a *AttributionService) CalculateAttribution(conversionID string, model AttributionModel) ([]TouchpointValue, error) {
	// Get conversion journey
	journey, err := a.GetConversionJourney(conversionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion journey: %w", err)
	}

	if len(journey.Touchpoints) == 0 {
		return nil, fmt.Errorf("no touchpoints found for conversion %s", conversionID)
	}

	// Apply attribution model
	var touchpointValues []TouchpointValue
	switch model {
	case FirstTouchAttribution:
		touchpointValues = a.applyFirstTouchAttribution(journey)
	case LastTouchAttribution:
		touchpointValues = a.applyLastTouchAttribution(journey)
	case LinearAttribution:
		touchpointValues = a.applyLinearAttribution(journey)
	case TimeDecayAttribution:
		touchpointValues = a.applyTimeDecayAttribution(journey)
	case PositionBasedAttribution:
		touchpointValues = a.applyPositionBasedAttribution(journey)
	case DataDrivenAttribution:
		touchpointValues = a.applyDataDrivenAttribution(journey)
	default:
		return nil, fmt.Errorf("unsupported attribution model: %s", model)
	}

	// Store attribution values in database
	err = a.storeAttributionValues(touchpointValues, string(model))
	if err != nil {
		return nil, fmt.Errorf("failed to store attribution values: %w", err)
	}

	return touchpointValues, nil
}

// GetConversionJourney retrieves the complete customer journey for a conversion
func (a *AttributionService) GetConversionJourney(conversionID string) (*models.ConversionJourney, error) {
	// For now, we'll work directly with touchpoints since conversion details are complex
	// In a production system, you'd want to fetch conversion details first

	// Get all touchpoints for this session leading to the conversion
	query := `
		SELECT 
			id, session_id, short_code, user_ip, user_agent, referrer,
			campaign_source, campaign_medium, campaign_name,
			touchpoint_order, touchpoint_time, created_at
		FROM attribution_touchpoints 
		WHERE session_id = $1 
		  AND touchpoint_time <= $2
		ORDER BY touchpoint_time ASC
	`

	// For now, we'll get all touchpoints for this conversionID (simplified approach)
	rows, err := a.storage.Query(query, conversionID, time.Now())
	if err != nil {
		return nil, fmt.Errorf("failed to query touchpoints: %w", err)
	}
	defer rows.Close()

	var touchpoints []models.AttributionTouchpoint
	for rows.Next() {
		var touchpoint models.AttributionTouchpoint
		var campaignSource, campaignMedium, campaignName sql.NullString

		err := rows.Scan(
			&touchpoint.ID, &touchpoint.SessionID, &touchpoint.ShortCode,
			&touchpoint.UserIP, &touchpoint.UserAgent, &touchpoint.Referrer,
			&campaignSource, &campaignMedium, &campaignName,
			&touchpoint.TouchpointOrder, &touchpoint.TouchpointTime,
			&touchpoint.CreatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan touchpoint: %w", err)
		}

		touchpoint.CampaignSource = campaignSource.String
		touchpoint.CampaignMedium = campaignMedium.String
		touchpoint.CampaignName = campaignName.String

		touchpoints = append(touchpoints, touchpoint)
	}

	// Calculate journey time in minutes (simplified)
	var journeyTime int
	var sessionID string
	if len(touchpoints) > 0 {
		firstTouch := touchpoints[0].TouchpointTime
		lastTouch := touchpoints[len(touchpoints)-1].TouchpointTime
		journeyTime = int(lastTouch.Sub(firstTouch).Minutes())
		sessionID = touchpoints[0].SessionID
	}

	journey := &models.ConversionJourney{
		ConversionID: conversionID,
		SessionID:    sessionID,
		Touchpoints:  touchpoints,
		TotalTouches: len(touchpoints),
		JourneyTime:  journeyTime,
	}

	return journey, nil
}

// Attribution Model Implementations

// applyFirstTouchAttribution gives 100% credit to the first touchpoint
func (a *AttributionService) applyFirstTouchAttribution(journey *models.ConversionJourney) []TouchpointValue {
	if len(journey.Touchpoints) == 0 {
		return nil
	}

	firstTouchpoint := journey.Touchpoints[0]
	return []TouchpointValue{
		{
			TouchpointID:     firstTouchpoint.ID,
			ShortCode:        firstTouchpoint.ShortCode,
			AttributionValue: journey.Conversion.ConversionValue,
			AttributionModel: string(FirstTouchAttribution),
			Weight:           1.0,
		},
	}
}

// applyLastTouchAttribution gives 100% credit to the last touchpoint
func (a *AttributionService) applyLastTouchAttribution(journey *models.ConversionJourney) []TouchpointValue {
	if len(journey.Touchpoints) == 0 {
		return nil
	}

	lastTouchpoint := journey.Touchpoints[len(journey.Touchpoints)-1]
	return []TouchpointValue{
		{
			TouchpointID:     lastTouchpoint.ID,
			ShortCode:        lastTouchpoint.ShortCode,
			AttributionValue: journey.Conversion.ConversionValue,
			AttributionModel: string(LastTouchAttribution),
			Weight:           1.0,
		},
	}
}

// applyLinearAttribution distributes credit equally among all touchpoints
func (a *AttributionService) applyLinearAttribution(journey *models.ConversionJourney) []TouchpointValue {
	touchpoints := journey.Touchpoints
	if len(touchpoints) == 0 {
		return nil
	}

	weight := 1.0 / float64(len(touchpoints))
	attributionValue := journey.Conversion.ConversionValue * weight

	var values []TouchpointValue
	for _, touchpoint := range touchpoints {
		values = append(values, TouchpointValue{
			TouchpointID:     touchpoint.ID,
			ShortCode:        touchpoint.ShortCode,
			AttributionValue: attributionValue,
			AttributionModel: string(LinearAttribution),
			Weight:           weight,
		})
	}

	return values
}

// applyTimeDecayAttribution gives more credit to touchpoints closer to conversion
func (a *AttributionService) applyTimeDecayAttribution(journey *models.ConversionJourney) []TouchpointValue {
	touchpoints := journey.Touchpoints
	if len(touchpoints) == 0 {
		return nil
	}

	conversionTime := journey.Conversion.ConversionTime
	halfLife := 7.0 * 24 * 60 // 7 days in minutes

	// Calculate decay weights
	var totalWeight float64
	weights := make([]float64, len(touchpoints))

	for i, touchpoint := range touchpoints {
		// Time difference in minutes
		timeDiff := conversionTime.Sub(touchpoint.TouchpointTime).Minutes()
		// Exponential decay: weight = 2^(-timeDiff/halfLife)
		weight := 1.0
		if timeDiff > 0 {
			weight = 1.0 / (1.0 + timeDiff/halfLife)
		}
		weights[i] = weight
		totalWeight += weight
	}

	// Normalize weights and calculate attribution values
	var values []TouchpointValue
	for i, touchpoint := range touchpoints {
		normalizedWeight := weights[i] / totalWeight
		attributionValue := journey.Conversion.ConversionValue * normalizedWeight

		values = append(values, TouchpointValue{
			TouchpointID:     touchpoint.ID,
			ShortCode:        touchpoint.ShortCode,
			AttributionValue: attributionValue,
			AttributionModel: string(TimeDecayAttribution),
			Weight:           normalizedWeight,
		})
	}

	return values
}

// applyPositionBasedAttribution gives 40% to first, 40% to last, 20% distributed among middle
func (a *AttributionService) applyPositionBasedAttribution(journey *models.ConversionJourney) []TouchpointValue {
	touchpoints := journey.Touchpoints
	if len(touchpoints) == 0 {
		return nil
	}

	if len(touchpoints) == 1 {
		// Single touchpoint gets all credit
		return []TouchpointValue{
			{
				TouchpointID:     touchpoints[0].ID,
				ShortCode:        touchpoints[0].ShortCode,
				AttributionValue: journey.Conversion.ConversionValue,
				AttributionModel: string(PositionBasedAttribution),
				Weight:           1.0,
			},
		}
	}

	var values []TouchpointValue
	totalValue := journey.Conversion.ConversionValue

	if len(touchpoints) == 2 {
		// Two touchpoints: 50% each
		for _, touchpoint := range touchpoints {
			values = append(values, TouchpointValue{
				TouchpointID:     touchpoint.ID,
				ShortCode:        touchpoint.ShortCode,
				AttributionValue: totalValue * 0.5,
				AttributionModel: string(PositionBasedAttribution),
				Weight:           0.5,
			})
		}
		return values
	}

	// More than 2 touchpoints: 40% first, 40% last, 20% distributed among middle
	middleCount := len(touchpoints) - 2
	middleWeight := 0.2 / float64(middleCount)

	for i, touchpoint := range touchpoints {
		var weight float64
		var attributionValue float64

		if i == 0 {
			// First touchpoint gets 40%
			weight = 0.4
		} else if i == len(touchpoints)-1 {
			// Last touchpoint gets 40%
			weight = 0.4
		} else {
			// Middle touchpoints share 20%
			weight = middleWeight
		}

		attributionValue = totalValue * weight

		values = append(values, TouchpointValue{
			TouchpointID:     touchpoint.ID,
			ShortCode:        touchpoint.ShortCode,
			AttributionValue: attributionValue,
			AttributionModel: string(PositionBasedAttribution),
			Weight:           weight,
		})
	}

	return values
}

// applyDataDrivenAttribution uses machine learning-like approach based on historical data
func (a *AttributionService) applyDataDrivenAttribution(journey *models.ConversionJourney) []TouchpointValue {
	touchpoints := journey.Touchpoints
	if len(touchpoints) == 0 {
		return nil
	}

	// For now, implement a simplified data-driven model
	// In a full implementation, this would use ML algorithms to analyze conversion patterns

	// Get historical conversion rates for each channel/source combination
	channelWeights := make(map[string]float64)
	totalWeight := 0.0

	for _, touchpoint := range touchpoints {
		channel := a.getChannelKey(touchpoint)
		
		// Get historical conversion rate for this channel
		conversionRate := a.getChannelConversionRate(channel)
		channelWeights[channel] = conversionRate
		totalWeight += conversionRate
	}

	// Normalize weights and calculate attribution
	var values []TouchpointValue
	for _, touchpoint := range touchpoints {
		channel := a.getChannelKey(touchpoint)
		weight := channelWeights[channel] / totalWeight
		attributionValue := journey.Conversion.ConversionValue * weight

		values = append(values, TouchpointValue{
			TouchpointID:     touchpoint.ID,
			ShortCode:        touchpoint.ShortCode,
			AttributionValue: attributionValue,
			AttributionModel: string(DataDrivenAttribution),
			Weight:           weight,
		})
	}

	return values
}

// Helper methods

// getChannelKey creates a unique key for a touchpoint's channel
func (a *AttributionService) getChannelKey(touchpoint models.AttributionTouchpoint) string {
	if touchpoint.CampaignSource != "" {
		return fmt.Sprintf("%s_%s", touchpoint.CampaignSource, touchpoint.CampaignMedium)
	}
	// Use referrer as fallback
	return touchpoint.Referrer
}

// getChannelConversionRate gets historical conversion rate for a channel
func (a *AttributionService) getChannelConversionRate(channel string) float64 {
	// This is a simplified implementation
	// In practice, you'd query historical data to calculate actual conversion rates
	query := `
		SELECT 
			COUNT(DISTINCT at.session_id) as touchpoints,
			COUNT(DISTINCT c.conversion_id) as conversions
		FROM attribution_touchpoints at
		LEFT JOIN conversions c ON at.session_id = c.session_id
		WHERE CONCAT(at.campaign_source, '_', at.campaign_medium) = $1
		   OR (at.campaign_source = '' AND at.referrer = $1)
		  AND at.created_at >= NOW() - INTERVAL '30 days'
	`

	var touchpoints, conversions int64
	err := a.storage.QueryRow(query, channel).Scan(&touchpoints, &conversions)
	if err != nil || touchpoints == 0 {
		// Return default rate if no data
		return 0.1 // 10% default conversion rate
	}

	return float64(conversions) / float64(touchpoints)
}

// storeAttributionValues stores calculated attribution values in the database
func (a *AttributionService) storeAttributionValues(values []TouchpointValue, model string) error {
	if len(values) == 0 {
		return nil
	}

	tx, err := a.storage.Begin()
	if err != nil {
		return fmt.Errorf("failed to begin transaction: %w", err)
	}
	defer tx.Rollback()

	// Delete existing attribution values for this model and touchpoints
	touchpointIDs := make([]interface{}, len(values))
	for i, v := range values {
		touchpointIDs[i] = v.TouchpointID
	}

	// Create placeholder string for IN clause
	placeholders := make([]string, len(touchpointIDs))
	for i := range placeholders {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
	}

	deleteQuery := fmt.Sprintf(`
		DELETE FROM touchpoint_attributions 
		WHERE attribution_model = $1 
		AND touchpoint_id IN (%s)
	`, strings.Join(placeholders, ", "))

	args := append([]interface{}{model}, touchpointIDs...)
	_, err = tx.Exec(deleteQuery, args...)
	if err != nil {
		return fmt.Errorf("failed to delete old attribution values: %w", err)
	}

	// Insert new attribution values
	insertQuery := `
		INSERT INTO touchpoint_attributions (
			touchpoint_id, attribution_model, attribution_value, weight, created_at
		) VALUES ($1, $2, $3, $4, NOW())
	`

	for _, value := range values {
		_, err = tx.Exec(insertQuery, value.TouchpointID, value.AttributionModel, 
			value.AttributionValue, value.Weight)
		if err != nil {
			return fmt.Errorf("failed to insert attribution value: %w", err)
		}
	}

	return tx.Commit()
}

// GetAttributionReport generates a comprehensive attribution report for a conversion
func (a *AttributionService) GetAttributionReport(conversionID string) (*AttributionReport, error) {
	journey, err := a.GetConversionJourney(conversionID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion journey: %w", err)
	}

	report := &AttributionReport{
		ConversionID:         conversionID,
		TotalValue:           journey.Conversion.ConversionValue,
		Journey:              *journey,
		AttributionBreakdown: make(map[string][]TouchpointValue),
		ModelComparison:      make(map[string]float64),
	}

	// Calculate attribution for all models
	models := []AttributionModel{
		FirstTouchAttribution,
		LastTouchAttribution,
		LinearAttribution,
		TimeDecayAttribution,
		PositionBasedAttribution,
		DataDrivenAttribution,
	}

	for _, model := range models {
		values, err := a.CalculateAttribution(conversionID, model)
		if err != nil {
			continue // Skip models that fail
		}

		modelStr := string(model)
		report.AttributionBreakdown[modelStr] = values

		// Calculate total attributed value for model comparison
		total := 0.0
		for _, value := range values {
			total += value.AttributionValue
		}
		report.ModelComparison[modelStr] = total
	}

	// Recommend best model (simplified logic)
	report.RecommendedModel = a.recommendAttributionModel(journey)

	return report, nil
}

// recommendAttributionModel recommends the best attribution model for the journey
func (a *AttributionService) recommendAttributionModel(journey *models.ConversionJourney) string {
	touchpointCount := len(journey.Touchpoints)
	journeyDays := float64(journey.JourneyTime) / (24 * 60) // Convert minutes to days

	// Simple recommendation logic
	if touchpointCount == 1 {
		return string(FirstTouchAttribution)
	} else if touchpointCount == 2 {
		return string(LinearAttribution)
	} else if journeyDays > 7 {
		return string(TimeDecayAttribution)
	} else {
		return string(PositionBasedAttribution)
	}
}

// GetChannelAttribution provides attribution analysis by marketing channel
func (a *AttributionService) GetChannelAttribution(shortCode string, days int, model AttributionModel) ([]ChannelAttribution, error) {
	// This would implement channel-level attribution analysis
	// For now, return a simplified version
	query := `
		SELECT 
			COALESCE(at.campaign_source, 'Direct') as source,
			COALESCE(at.campaign_medium, 'None') as medium,
			COUNT(DISTINCT at.id) as touchpoints,
			COUNT(DISTINCT c.conversion_id) as conversions,
			COALESCE(SUM(ta.attribution_value), 0) as attribution_value
		FROM attribution_touchpoints at
		LEFT JOIN conversions c ON at.session_id = c.session_id
		LEFT JOIN touchpoint_attributions ta ON at.id = ta.touchpoint_id 
			AND ta.attribution_model = $1
		WHERE at.short_code = $2
		  AND at.created_at >= NOW() - INTERVAL '%d days'
		GROUP BY at.campaign_source, at.campaign_medium
		ORDER BY attribution_value DESC
	`

	rows, err := a.storage.Query(fmt.Sprintf(query, days), string(model), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to query channel attribution: %w", err)
	}
	defer rows.Close()

	var channels []ChannelAttribution
	for rows.Next() {
		var source, medium string
		var touchpoints, conversions int64
		var attributionValue float64

		err := rows.Scan(&source, &medium, &touchpoints, &conversions, &attributionValue)
		if err != nil {
			return nil, fmt.Errorf("failed to scan channel attribution: %w", err)
		}

		channel := fmt.Sprintf("%s/%s", source, medium)
		conversionRate := 0.0
		if touchpoints > 0 {
			conversionRate = float64(conversions) / float64(touchpoints) * 100
		}

		channels = append(channels, ChannelAttribution{
			Channel:        channel,
			Source:         source,
			Medium:         medium,
			Touchpoints:    int(touchpoints),
			Conversions:    int(conversions),
			AttributionValue: map[string]float64{string(model): attributionValue},
			ConversionRate: conversionRate,
		})
	}

	return channels, nil
}