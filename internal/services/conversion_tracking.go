package services

import (
	"crypto/md5"
	"database/sql"
	"fmt"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

// ConversionTrackingService handles conversion tracking operations
type ConversionTrackingService struct {
	storage *storage.PostgresStorage
}

// NewConversionTrackingService creates a new conversion tracking service
func NewConversionTrackingService(storage *storage.PostgresStorage) *ConversionTrackingService {
	return &ConversionTrackingService{
		storage: storage,
	}
}

// Conversion Goals Management

// CreateConversionGoal creates a new conversion goal for a user
func (c *ConversionTrackingService) CreateConversionGoal(userID int64, request *models.CreateConversionGoalRequest) (*models.ConversionGoal, error) {
	// Generate ID for the goal
	id, err := utils.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate goal ID: %w", err)
	}

	goal := &models.ConversionGoal{
		ID:                id,
		UserID:            userID,
		GoalName:          request.GoalName,
		GoalType:          request.GoalType,
		TargetURL:         request.TargetURL,
		CustomEventName:   request.CustomEventName,
		GoalValue:         request.GoalValue,
		AttributionWindow: request.AttributionWindow,
		IsActive:          true,
		CreatedAt:         time.Now(),
		UpdatedAt:         time.Now(),
	}

	// Validate goal configuration
	if err := c.validateConversionGoal(goal); err != nil {
		return nil, fmt.Errorf("invalid goal configuration: %w", err)
	}

	// Save to database
	query := `
		INSERT INTO conversion_goals (
			id, user_id, goal_name, goal_type, target_url, custom_event_name, 
			goal_value, attribution_window, is_active, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)
	`
	
	_, err = c.storage.Exec(query, goal.ID, goal.UserID, goal.GoalName, goal.GoalType,
		goal.TargetURL, goal.CustomEventName, goal.GoalValue, goal.AttributionWindow,
		goal.IsActive, goal.CreatedAt, goal.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to save conversion goal: %w", err)
	}

	return goal, nil
}

// GetConversionGoals retrieves all conversion goals for a user
func (c *ConversionTrackingService) GetConversionGoals(userID int64) ([]models.ConversionGoal, error) {
	query := `
		SELECT id, user_id, goal_name, goal_type, target_url, custom_event_name,
		       goal_value, attribution_window, is_active, created_at, updated_at
		FROM conversion_goals
		WHERE user_id = $1
		ORDER BY created_at DESC
	`
	
	rows, err := c.storage.Query(query, userID)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion goals: %w", err)
	}
	defer rows.Close()

	var goals []models.ConversionGoal
	for rows.Next() {
		var goal models.ConversionGoal
		var targetURL, customEventName sql.NullString
		
		err := rows.Scan(&goal.ID, &goal.UserID, &goal.GoalName, &goal.GoalType,
			&targetURL, &customEventName, &goal.GoalValue, &goal.AttributionWindow,
			&goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversion goal: %w", err)
		}

		if targetURL.Valid {
			goal.TargetURL = &targetURL.String
		}
		if customEventName.Valid {
			goal.CustomEventName = &customEventName.String
		}

		goals = append(goals, goal)
	}

	return goals, nil
}

// UpdateConversionGoal updates an existing conversion goal
func (c *ConversionTrackingService) UpdateConversionGoal(userID, goalID int64, request *models.UpdateConversionGoalRequest) (*models.ConversionGoal, error) {
	// Get existing goal
	goal, err := c.getConversionGoalByID(goalID, userID)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if request.GoalName != nil {
		goal.GoalName = *request.GoalName
	}
	if request.GoalType != nil {
		goal.GoalType = *request.GoalType
	}
	if request.TargetURL != nil {
		goal.TargetURL = request.TargetURL
	}
	if request.CustomEventName != nil {
		goal.CustomEventName = request.CustomEventName
	}
	if request.GoalValue != nil {
		goal.GoalValue = *request.GoalValue
	}
	if request.AttributionWindow != nil {
		goal.AttributionWindow = *request.AttributionWindow
	}
	if request.IsActive != nil {
		goal.IsActive = *request.IsActive
	}
	goal.UpdatedAt = time.Now()

	// Validate updated goal
	if err := c.validateConversionGoal(goal); err != nil {
		return nil, fmt.Errorf("invalid goal configuration: %w", err)
	}

	// Update in database
	query := `
		UPDATE conversion_goals
		SET goal_name = $1, goal_type = $2, target_url = $3, custom_event_name = $4,
		    goal_value = $5, attribution_window = $6, is_active = $7, updated_at = $8
		WHERE id = $9 AND user_id = $10
	`
	
	_, err = c.storage.Exec(query, goal.GoalName, goal.GoalType, goal.TargetURL,
		goal.CustomEventName, goal.GoalValue, goal.AttributionWindow, goal.IsActive,
		goal.UpdatedAt, goalID, userID)
	
	if err != nil {
		return nil, fmt.Errorf("failed to update conversion goal: %w", err)
	}

	return goal, nil
}

// DeleteConversionGoal deletes a conversion goal
func (c *ConversionTrackingService) DeleteConversionGoal(userID, goalID int64) error {
	query := `DELETE FROM conversion_goals WHERE id = $1 AND user_id = $2`
	result, err := c.storage.Exec(query, goalID, userID)
	if err != nil {
		return fmt.Errorf("failed to delete conversion goal: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get rows affected: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("conversion goal not found")
	}

	return nil
}

// Conversion Tracking

// TrackConversion records a conversion event
func (c *ConversionTrackingService) TrackConversion(shortCode, userIP, userAgent, referrer string, request *models.ConversionTrackingRequest) (*models.Conversion, error) {
	// Get the goal to verify it exists and is active
	goal, err := c.getActiveConversionGoal(request.GoalID)
	if err != nil {
		return nil, fmt.Errorf("invalid conversion goal: %w", err)
	}

	// Generate unique conversion ID
	id, err := utils.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate conversion ID: %w", err)
	}

	// Check for duplicate conversions (same conversion_id)
	existingConversion, _ := c.getConversionByID(request.ConversionID)
	if existingConversion != nil {
		return existingConversion, nil // Return existing conversion, don't create duplicate
	}

	// Calculate time to conversion if we can find the original click
	timeToConversion := 0
	var clickID *int64
	
	if request.SessionID != "" {
		if originalClick, err := c.getOriginalClickBySession(shortCode, request.SessionID); err == nil {
			timeToConversion = int(time.Since(originalClick.ClickedAt).Minutes())
			clickID = &originalClick.ID
		}
	}

	conversion := &models.Conversion{
		ID:               id,
		ShortCode:        shortCode,
		GoalID:           request.GoalID,
		ConversionID:     request.ConversionID,
		ConversionType:   goal.GoalType,
		ConversionValue:  request.ConversionValue,
		UserIP:           userIP,
		UserAgent:        userAgent,
		Referrer:         referrer,
		SessionID:        request.SessionID,
		ClickID:          clickID,
		ConversionTime:   time.Now(),
		AttributionModel: "last_click", // Default attribution model
		TimeToConversion: timeToConversion,
		CreatedAt:        time.Now(),
	}

	// Save conversion to database
	query := `
		INSERT INTO conversions (
			id, short_code, goal_id, conversion_id, conversion_type, conversion_value,
			user_ip, user_agent, referrer, session_id, click_id, conversion_time,
			attribution_model, time_to_conversion, created_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
	`
	
	_, err = c.storage.Exec(query, conversion.ID, conversion.ShortCode, conversion.GoalID,
		conversion.ConversionID, conversion.ConversionType, conversion.ConversionValue,
		conversion.UserIP, conversion.UserAgent, conversion.Referrer, conversion.SessionID,
		conversion.ClickID, conversion.ConversionTime, conversion.AttributionModel,
		conversion.TimeToConversion, conversion.CreatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to save conversion: %w", err)
	}

	// Update conversion analytics in referrer_analytics table
	go c.updateConversionAnalytics(shortCode, conversion.ConversionValue)

	return conversion, nil
}

// GetConversionStats retrieves conversion statistics for a user's URLs
func (c *ConversionTrackingService) GetConversionStats(userID int64, shortCode string, days int) ([]models.ConversionStats, error) {
	query := `
		SELECT 
			cg.id,
			cg.goal_name,
			COUNT(c.id) as total_conversions,
			COALESCE(SUM(c.conversion_value), 0) as total_value,
			COALESCE(AVG(c.conversion_value), 0) as avg_value,
			COALESCE(AVG(c.time_to_conversion), 0) as avg_time_to_convert
		FROM conversion_goals cg
		LEFT JOIN conversions c ON cg.id = c.goal_id 
			AND c.short_code = $2 
			AND c.conversion_time >= NOW() - INTERVAL '%d days'
		WHERE cg.user_id = $1 AND cg.is_active = true
		GROUP BY cg.id, cg.goal_name
		ORDER BY total_conversions DESC
	`
	
	rows, err := c.storage.Query(fmt.Sprintf(query, days), userID, shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to get conversion stats: %w", err)
	}
	defer rows.Close()

	var stats []models.ConversionStats
	for rows.Next() {
		var stat models.ConversionStats
		err := rows.Scan(&stat.GoalID, &stat.GoalName, &stat.TotalConversions,
			&stat.TotalValue, &stat.AvgValue, &stat.AvgTimeToConvert)
		
		if err != nil {
			return nil, fmt.Errorf("failed to scan conversion stats: %w", err)
		}

		// Calculate conversion rate
		if totalClicks := c.getTotalClicksForURL(shortCode, days); totalClicks > 0 {
			stat.ConversionRate = (float64(stat.TotalConversions) / float64(totalClicks)) * 100
		}

		stats = append(stats, stat)
	}

	return stats, nil
}

// Helper methods

func (c *ConversionTrackingService) validateConversionGoal(goal *models.ConversionGoal) error {
	switch goal.GoalType {
	case "url_visit":
		if goal.TargetURL == nil || *goal.TargetURL == "" {
			return fmt.Errorf("target_url is required for url_visit goals")
		}
	case "custom_event":
		if goal.CustomEventName == nil || *goal.CustomEventName == "" {
			return fmt.Errorf("custom_event_name is required for custom_event goals")
		}
	case "form_submit", "purchase":
		// These are valid without additional requirements
	default:
		return fmt.Errorf("invalid goal_type: %s", goal.GoalType)
	}

	if goal.AttributionWindow < 1 || goal.AttributionWindow > 365 {
		return fmt.Errorf("attribution_window must be between 1 and 365 days")
	}

	return nil
}

func (c *ConversionTrackingService) getConversionGoalByID(goalID, userID int64) (*models.ConversionGoal, error) {
	query := `
		SELECT id, user_id, goal_name, goal_type, target_url, custom_event_name,
		       goal_value, attribution_window, is_active, created_at, updated_at
		FROM conversion_goals
		WHERE id = $1 AND user_id = $2
	`
	
	var goal models.ConversionGoal
	var targetURL, customEventName sql.NullString
	
	err := c.storage.QueryRow(query, goalID, userID).Scan(
		&goal.ID, &goal.UserID, &goal.GoalName, &goal.GoalType,
		&targetURL, &customEventName, &goal.GoalValue, &goal.AttributionWindow,
		&goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversion goal not found")
		}
		return nil, fmt.Errorf("failed to get conversion goal: %w", err)
	}

	if targetURL.Valid {
		goal.TargetURL = &targetURL.String
	}
	if customEventName.Valid {
		goal.CustomEventName = &customEventName.String
	}

	return &goal, nil
}

func (c *ConversionTrackingService) getActiveConversionGoal(goalID int64) (*models.ConversionGoal, error) {
	query := `
		SELECT id, user_id, goal_name, goal_type, target_url, custom_event_name,
		       goal_value, attribution_window, is_active, created_at, updated_at
		FROM conversion_goals
		WHERE id = $1 AND is_active = true
	`
	
	var goal models.ConversionGoal
	var targetURL, customEventName sql.NullString
	
	err := c.storage.QueryRow(query, goalID).Scan(
		&goal.ID, &goal.UserID, &goal.GoalName, &goal.GoalType,
		&targetURL, &customEventName, &goal.GoalValue, &goal.AttributionWindow,
		&goal.IsActive, &goal.CreatedAt, &goal.UpdatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("conversion goal not found or inactive")
		}
		return nil, fmt.Errorf("failed to get conversion goal: %w", err)
	}

	if targetURL.Valid {
		goal.TargetURL = &targetURL.String
	}
	if customEventName.Valid {
		goal.CustomEventName = &customEventName.String
	}

	return &goal, nil
}

func (c *ConversionTrackingService) getConversionByID(conversionID string) (*models.Conversion, error) {
	query := `
		SELECT id, short_code, goal_id, conversion_id, conversion_type, conversion_value,
		       user_ip, user_agent, referrer, session_id, click_id, conversion_time,
		       attribution_model, time_to_conversion, created_at
		FROM conversions
		WHERE conversion_id = $1
	`
	
	var conversion models.Conversion
	var userIP, userAgent, referrer, sessionID sql.NullString
	var clickID sql.NullInt64
	
	err := c.storage.QueryRow(query, conversionID).Scan(
		&conversion.ID, &conversion.ShortCode, &conversion.GoalID, &conversion.ConversionID,
		&conversion.ConversionType, &conversion.ConversionValue, &userIP, &userAgent,
		&referrer, &sessionID, &clickID, &conversion.ConversionTime,
		&conversion.AttributionModel, &conversion.TimeToConversion, &conversion.CreatedAt)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, nil // Conversion not found, return nil without error
		}
		return nil, fmt.Errorf("failed to get conversion: %w", err)
	}

	if userIP.Valid {
		conversion.UserIP = userIP.String
	}
	if userAgent.Valid {
		conversion.UserAgent = userAgent.String
	}
	if referrer.Valid {
		conversion.Referrer = referrer.String
	}
	if sessionID.Valid {
		conversion.SessionID = sessionID.String
	}
	if clickID.Valid {
		conversion.ClickID = &clickID.Int64
	}

	return &conversion, nil
}

func (c *ConversionTrackingService) getOriginalClickBySession(shortCode, sessionID string) (*models.ClickEvent, error) {
	// This would need to be implemented based on how you track sessions
	// For now, returning a placeholder implementation
	return nil, fmt.Errorf("session tracking not implemented")
}

func (c *ConversionTrackingService) getTotalClicksForURL(shortCode string, days int) int64 {
	query := `
		SELECT COALESCE(SUM(clicks), 0)
		FROM click_events
		WHERE short_code = $1 AND clicked_at >= NOW() - INTERVAL '%d days'
	`
	
	var totalClicks int64
	err := c.storage.QueryRow(fmt.Sprintf(query, days), shortCode).Scan(&totalClicks)
	if err != nil {
		return 0
	}
	
	return totalClicks
}

func (c *ConversionTrackingService) updateConversionAnalytics(shortCode string, conversionValue float64) {
	// Update conversion count in referrer_analytics table
	query := `
		UPDATE referrer_analytics 
		SET conversions = conversions + 1
		WHERE short_code = $1
	`
	
	c.storage.Exec(query, shortCode)
	
	// Could also update other analytics tables with conversion data
}

// GenerateSessionID generates a unique session ID for tracking user sessions
func GenerateSessionID(userIP, userAgent string) string {
	data := fmt.Sprintf("%s_%s_%d", userIP, userAgent, time.Now().UnixNano())
	hash := md5.Sum([]byte(data))
	return fmt.Sprintf("%x", hash)
}

// Attribution Modeling

// CalculateAttributedConversions calculates conversions with different attribution models
func (c *ConversionTrackingService) CalculateAttributedConversions(shortCode string, attributionModel string, days int) (map[string]float64, error) {
	// This is a simplified implementation
	// In a full implementation, you would analyze the complete customer journey
	
	conversions := make(map[string]float64)
	
	query := `
		SELECT referrer_domain, COUNT(*) as conversion_count
		FROM conversions c
		JOIN referrer_analytics ra ON c.short_code = ra.short_code
		WHERE c.short_code = $1 
		AND c.conversion_time >= NOW() - INTERVAL '%d days'
		GROUP BY referrer_domain
	`
	
	rows, err := c.storage.Query(fmt.Sprintf(query, days), shortCode)
	if err != nil {
		return nil, fmt.Errorf("failed to calculate attributed conversions: %w", err)
	}
	defer rows.Close()
	
	for rows.Next() {
		var domain string
		var count float64
		if err := rows.Scan(&domain, &count); err != nil {
			continue
		}
		
		switch attributionModel {
		case "first_click":
			// Attribute 100% to first touch
			conversions[domain] = count
		case "last_click":
			// Attribute 100% to last touch  
			conversions[domain] = count
		case "linear":
			// Distribute evenly across all touches
			conversions[domain] = count // Simplified - would need journey analysis
		default:
			conversions[domain] = count
		}
	}
	
	return conversions, nil
}

// GetUserConversionGoals is an alias for GetConversionGoals to match handler expectations
func (c *ConversionTrackingService) GetUserConversionGoals(userID int64) ([]models.ConversionGoal, error) {
	return c.GetConversionGoals(userID)
}

// GetConversionGoal retrieves a specific conversion goal by ID and user ID
func (c *ConversionTrackingService) GetConversionGoal(goalID, userID int64) (*models.ConversionGoal, error) {
	return c.getConversionGoalByID(goalID, userID)
}

// TrackConversionFromRequest tracks a conversion from a request object
func (c *ConversionTrackingService) TrackConversionFromRequest(request *models.ConversionTrackingRequest) (*models.Conversion, error) {
	// Call the existing method with empty strings for IP, user agent, and referrer
	return c.TrackConversion("", "", "", "", request)
}

// GetConversionStatsByGoal gets conversion stats by goal ID  
func (c *ConversionTrackingService) GetConversionStatsByGoal(goalID, userID int64, days int) ([]models.ConversionStats, error) {
	// Placeholder implementation - in reality you'd query by goal ID
	// For now, return empty stats
	return []models.ConversionStats{}, nil
}

// GetAttributionReport generates an attribution report for a conversion goal
func (c *ConversionTrackingService) GetAttributionReport(goalID, userID int64, model string, days int) (interface{}, error) {
	// Placeholder implementation - would generate comprehensive attribution analysis
	return map[string]interface{}{
		"goal_id":           goalID,
		"attribution_model": model,
		"period_days":       days,
		"touchpoints":       []interface{}{},
		"conversions":       []interface{}{},
	}, nil
}