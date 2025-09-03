package services

import (
	"crypto/sha256"
	"database/sql"
	"encoding/hex"
	"fmt"
	"net"
	"net/url"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/configs"
	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/URLshorter/url-shortener/internal/utils"
)

type ShortenerService struct {
	db          *storage.PostgresStorage
	redis       *storage.RedisStorage
	config      *configs.Config
	analytics   *AnalyticsService
	realtime    *RealtimeAnalyticsService
	attribution *AttributionService
}

// NewShortenerService creates a new shortener service
func NewShortenerService(db *storage.PostgresStorage, redis *storage.RedisStorage, config *configs.Config) *ShortenerService {
	// Initialize Snowflake ID generator
	utils.InitializeSnowflake(config.NodeID)
	
	service := &ShortenerService{
		db:     db,
		redis:  redis,
		config: config,
	}
	
	// Initialize analytics service
	service.analytics = NewAnalyticsService(db)
	service.analytics.SetRedis(redis)
	
	return service
}

// SetRealtimeService sets the real-time analytics service
func (s *ShortenerService) SetRealtimeService(realtime *RealtimeAnalyticsService) {
	s.realtime = realtime
}

// SetAttributionService sets the attribution service
func (s *ShortenerService) SetAttributionService(attribution *AttributionService) {
	s.attribution = attribution
}

// ShortenURL creates a new short URL
func (s *ShortenerService) ShortenURL(request *models.ShortenRequest, clientIP string, userID *int64) (*models.ShortenResponse, error) {
	// Validate URL
	if err := s.validateURL(request.URL); err != nil {
		return nil, err
	}

	// Generate unique ID
	id, err := utils.GenerateID()
	if err != nil {
		return nil, fmt.Errorf("failed to generate unique ID: %w", err)
	}

	// Generate short code
	var shortCode string
	if request.CustomCode != "" {
		// Validate custom code
		if err := s.validateCustomCode(request.CustomCode); err != nil {
			return nil, err
		}
		shortCode = request.CustomCode
		
		// Check if custom code already exists
		exists, err := s.db.ShortCodeExists(shortCode)
		if err != nil {
			return nil, fmt.Errorf("failed to check custom code availability: %w", err)
		}
		if exists {
			return nil, ErrCustomCodeAlreadyExists
		}
	} else {
		shortCode = utils.EncodeBase62(id)
		
		// Ensure uniqueness (very unlikely collision with Snowflake IDs)
		for attempts := 0; attempts < 3; attempts++ {
			exists, err := s.db.ShortCodeExists(shortCode)
			if err != nil {
				return nil, fmt.Errorf("failed to check short code uniqueness: %w", err)
			}
			if !exists {
				break
			}
			
			// Generate new ID and encode (collision handling)
			id, err = utils.GenerateID()
			if err != nil {
				return nil, fmt.Errorf("failed to generate unique ID: %w", err)
			}
			shortCode = utils.EncodeBase62(id)
		}
	}

	// Create URL mapping
	mapping := &models.URLMapping{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: request.URL,
		CreatedAt:   time.Now(),
		ExpiresAt:   request.ExpiresAt,
		IsActive:    true,
		CreatedByIP: clientIP,
		UserID:      userID,
	}

	// Save to database
	if err := s.db.SaveURLMapping(mapping); err != nil {
		return nil, fmt.Errorf("failed to save URL mapping: %w", err)
	}

	// Cache in Redis with appropriate TTL
	cacheTTL := 24 * time.Hour // Default cache TTL
	if mapping.ExpiresAt != nil {
		timeUntilExpiry := time.Until(*mapping.ExpiresAt)
		if timeUntilExpiry < cacheTTL {
			cacheTTL = timeUntilExpiry
		}
	}
	
	s.redis.SetURLMapping(shortCode, mapping, cacheTTL)

	// Create response
	response := &models.ShortenResponse{
		ShortCode:   shortCode,
		ShortURL:    fmt.Sprintf("%s/%s", s.config.BaseURL, shortCode),
		OriginalURL: request.URL,
		CreatedAt:   mapping.CreatedAt,
		ExpiresAt:   mapping.ExpiresAt,
	}

	return response, nil
}

// GetOriginalURL retrieves the original URL for a short code
func (s *ShortenerService) GetOriginalURL(shortCode string) (*models.URLMapping, error) {
	// Try cache first
	mapping, err := s.redis.GetURLMapping(shortCode)
	if err != nil && err != storage.ErrCacheKeyNotFound {
		// Log cache error but continue with database lookup
		fmt.Printf("Cache lookup failed for %s: %v\n", shortCode, err)
	}

	// If not in cache or cache error, check database
	if mapping == nil {
		mapping, err = s.db.GetURLMappingByShortCode(shortCode)
		if err != nil {
			return nil, err
		}

		// Cache the result for faster future lookups
		cacheTTL := 24 * time.Hour
		if mapping.ExpiresAt != nil {
			timeUntilExpiry := time.Until(*mapping.ExpiresAt)
			if timeUntilExpiry < cacheTTL {
				cacheTTL = timeUntilExpiry
			}
		}
		s.redis.SetURLMapping(shortCode, mapping, cacheTTL)
	}

	return mapping, nil
}

// GetURLByShortCode retrieves a URL by its short code (for dashboard)
func (s *ShortenerService) GetURLByShortCode(shortCode string) (*models.URL, error) {
	// Implementation would query database
	// For now, return simulated URL
	now := time.Now()
	userID := int64(1) // Simulated user ID
	
	return &models.URL{
		ID:          1,
		ShortCode:   shortCode,
		OriginalURL: "https://example.com/original-url",
		Title:       stringPtr("Example Website"),
		Description: stringPtr("An example website for demonstration"),
		CreatedBy:   &userID,
		ClickCount:  25,
		IsActive:    true,
		IsPublic:    true,
		CreatedAt:   now.Add(-7 * 24 * time.Hour),
		UpdatedAt:   now.Add(-1 * time.Hour),
		ExpiresAt:   nil,
	}, nil
}

// UpdateURL updates a URL's properties
func (s *ShortenerService) UpdateURL(shortCode string, req *models.UpdateURLRequest) (*models.URL, error) {
	// Get current URL
	url, err := s.GetURLByShortCode(shortCode)
	if err != nil {
		return nil, err
	}

	// Update fields if provided
	if req.Title != nil {
		url.Title = req.Title
	}
	if req.Description != nil {
		url.Description = req.Description
	}
	if req.ExpiresAt != nil {
		url.ExpiresAt = req.ExpiresAt
	}
	if req.IsPublic != nil {
		url.IsPublic = *req.IsPublic
	}

	url.UpdatedAt = time.Now()

	// Implementation would save to database
	return url, nil
}

// DeleteURL deletes a URL (soft delete)
func (s *ShortenerService) DeleteURL(shortCode string) error {
	// Implementation would mark URL as inactive or delete from database
	// For now, simulate success
	return nil
}

// Helper function to create string pointers
func stringPtr(s string) *string {
	return &s
}

// RecordClick records a click event for analytics
func (s *ShortenerService) RecordClick(shortCode, clientIP, userAgent, referrer string) error {
	// Generate ID for click event
	id, err := utils.GenerateID()
	if err != nil {
		return fmt.Errorf("failed to generate click event ID: %w", err)
	}

	// Create click event
	event := &models.ClickEvent{
		ID:        id,
		ShortCode: shortCode,
		ClickedAt: time.Now(),
		IPAddress: clientIP,
		UserAgent: userAgent,
		Referrer:  referrer,
		CountryCode: s.getCountryFromIP(clientIP), // Simple implementation
	}

	// Save to database (async operation for better performance)
	go func() {
		if err := s.db.SaveClickEvent(event); err != nil {
			fmt.Printf("Failed to save click event: %v\n", err)
		}
	}()

	// Process enhanced analytics (async operation)
	go func() {
		if err := s.analytics.ProcessEnhancedClickEvent(event); err != nil {
			fmt.Printf("Failed to process enhanced click analytics: %v\n", err)
		}
	}()

	// Increment click count in database
	go func() {
		if err := s.db.IncrementClickCount(shortCode); err != nil {
			fmt.Printf("Failed to increment click count in DB: %v\n", err)
		}
	}()

	// Increment cached click count for faster analytics
	s.redis.IncrementClickCount(shortCode)
	
	// Broadcast real-time click event if real-time service is available
	if s.realtime != nil {
		s.realtime.BroadcastClick(shortCode, clientIP, userAgent, referrer)
	}
	
	// Record attribution touchpoint if attribution service is available
	if s.attribution != nil {
		go func() {
			s.recordAttributionTouchpoint(shortCode, clientIP, userAgent, referrer)
		}()
	}

	return nil
}

// validateURL validates if the provided URL is valid
func (s *ShortenerService) validateURL(rawURL string) error {
	if rawURL == "" {
		return ErrInvalidURL
	}

	// Parse URL
	parsedURL, err := url.Parse(rawURL)
	if err != nil {
		return ErrInvalidURL
	}

	// Check if URL has a valid scheme
	if parsedURL.Scheme != "http" && parsedURL.Scheme != "https" {
		return ErrInvalidURLScheme
	}

	// Check if URL has a host
	if parsedURL.Host == "" {
		return ErrInvalidURL
	}

	// Additional validation: check if it's not our own domain to prevent loops
	if strings.Contains(parsedURL.Host, s.config.ServerHost) {
		return ErrSelfReferentialURL
	}

	return nil
}

// validateCustomCode validates custom short codes
func (s *ShortenerService) validateCustomCode(code string) error {
	if len(code) < 3 || len(code) > 10 {
		return ErrInvalidCustomCodeLength
	}

	if !utils.IsValidBase62(code) {
		return ErrInvalidCustomCodeCharacters
	}

	// Check for reserved words
	reservedWords := []string{"api", "admin", "www", "health", "analytics"}
	for _, reserved := range reservedWords {
		if strings.ToLower(code) == reserved {
			return ErrReservedCustomCode
		}
	}

	return nil
}

// getCountryFromIP is a simple implementation for demo purposes
// In production, you'd use a proper GeoIP service
func (s *ShortenerService) getCountryFromIP(ipAddress string) string {
	ip := net.ParseIP(ipAddress)
	if ip == nil {
		return ""
	}

	// Simple check for local IPs
	if ip.IsLoopback() || ip.IsPrivate() {
		return "LOCAL"
	}

	// In production, integrate with MaxMind GeoIP2 or similar service
	return "US" // Placeholder
}

// Custom errors
var (
	ErrInvalidURL                    = &ServiceError{Message: "invalid URL"}
	ErrInvalidURLScheme             = &ServiceError{Message: "URL must use http or https scheme"}
	ErrSelfReferentialURL           = &ServiceError{Message: "cannot shorten URLs to this domain"}
	ErrInvalidCustomCodeLength      = &ServiceError{Message: "custom code must be 3-10 characters long"}
	ErrInvalidCustomCodeCharacters  = &ServiceError{Message: "custom code can only contain letters, numbers"}
	ErrReservedCustomCode          = &ServiceError{Message: "custom code is reserved"}
	ErrCustomCodeAlreadyExists     = &ServiceError{Message: "custom code already exists"}
)

type ServiceError struct {
	Message string
}

func (e *ServiceError) Error() string {
	return e.Message
}

// GetUserURLs retrieves URLs for a specific user with pagination
func (s *ShortenerService) GetUserURLs(userID int64, offset, limit int) ([]*models.UserURLResponse, int64, error) {
	query := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active,
		       COALESCE(user_id, 0), COALESCE(is_public, true), COALESCE(custom_alias, ''), 
		       COALESCE(title, ''), COALESCE(description, '')
		FROM url_mappings 
		WHERE user_id = $1 AND is_active = TRUE
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3
	`
	
	rows, err := s.db.Query(query, userID, limit, offset)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to query user URLs: %w", err)
	}
	defer rows.Close()

	var urls []*models.UserURLResponse
	for rows.Next() {
		var url models.UserURLResponse
		var userIDDB int64
		var isPublicDB bool
		var customAlias, title, description string
		var expiresAt sql.NullTime
		
		err := rows.Scan(
			&url.ID, &url.ShortCode, &url.OriginalURL, &url.CreatedAt,
			&expiresAt, &url.ClickCount, &url.IsActive,
			&userIDDB, &isPublicDB, &customAlias, &title, &description,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan URL row: %w", err)
		}

		// Set nullable fields
		if expiresAt.Valid {
			url.ExpiresAt = &expiresAt.Time
		}
		if title != "" {
			url.Title = &title
		}
		if description != "" {
			url.Description = &description
		}
		url.IsPublic = isPublicDB

		// Build short URL
		url.ShortURL = fmt.Sprintf("%s/%s", s.config.BaseURL, url.ShortCode)
		
		urls = append(urls, &url)
	}

	// Get total count
	var total int64
	err = s.db.QueryRow("SELECT COUNT(*) FROM url_mappings WHERE user_id = $1 AND is_active = TRUE", userID).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count user URLs: %w", err)
	}

	return urls, total, nil
}

// DeleteUserURL deletes a URL belonging to a specific user
func (s *ShortenerService) DeleteUserURL(userID int64, shortCode string) error {
	query := `UPDATE url_mappings SET is_active = FALSE WHERE short_code = $1 AND user_id = $2 AND is_active = TRUE`
	
	result, err := s.db.Exec(query, shortCode, userID)
	if err != nil {
		return fmt.Errorf("failed to delete URL: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to check affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return storage.ErrURLNotFound
	}

	// Remove from cache
	if s.redis != nil {
		s.redis.DeleteURLMapping(shortCode)
	}

	return nil
}

// UpdateUserURL updates a URL belonging to a specific user
func (s *ShortenerService) UpdateUserURL(userID int64, shortCode string, req *models.UpdateURLRequest) (*models.UserURLResponse, error) {
	// First check if the URL exists and belongs to the user
	var existingURL models.UserURLResponse
	var userIDDB sql.NullInt64
	var isPublicDB sql.NullBool
	var customAlias, title, description sql.NullString
	var expiresAt sql.NullTime

	checkQuery := `
		SELECT id, short_code, original_url, created_at, expires_at, click_count, is_active,
		       user_id, is_public, custom_alias, title, description
		FROM url_mappings 
		WHERE short_code = $1 AND is_active = TRUE
	`
	
	err := s.db.QueryRow(checkQuery, shortCode).Scan(
		&existingURL.ID, &existingURL.ShortCode, &existingURL.OriginalURL, &existingURL.CreatedAt,
		&expiresAt, &existingURL.ClickCount, &existingURL.IsActive,
		&userIDDB, &isPublicDB, &customAlias, &title, &description,
	)
	
	if err == sql.ErrNoRows {
		return nil, storage.ErrURLNotFound
	}
	if err != nil {
		return nil, fmt.Errorf("failed to fetch URL: %w", err)
	}

	// Check ownership
	if !userIDDB.Valid || userIDDB.Int64 != userID {
		return nil, storage.ErrUnauthorized
	}

	// Build update query dynamically
	updates := []string{}
	args := []interface{}{}
	argIndex := 1

	if req.Title != nil {
		updates = append(updates, fmt.Sprintf("title = $%d", argIndex))
		args = append(args, *req.Title)
		argIndex++
	}

	if req.Description != nil {
		updates = append(updates, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}

	if req.ExpiresAt != nil {
		updates = append(updates, fmt.Sprintf("expires_at = $%d", argIndex))
		args = append(args, *req.ExpiresAt)
		argIndex++
	}

	if req.IsPublic != nil {
		updates = append(updates, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *req.IsPublic)
		argIndex++
	}

	if len(updates) == 0 {
		// No updates requested, return current data
		if expiresAt.Valid {
			existingURL.ExpiresAt = &expiresAt.Time
		}
		if title.Valid {
			existingURL.Title = &title.String
		}
		if description.Valid {
			existingURL.Description = &description.String
		}
		if isPublicDB.Valid {
			existingURL.IsPublic = isPublicDB.Bool
		} else {
			existingURL.IsPublic = true
		}
		existingURL.ShortURL = fmt.Sprintf("%s/%s", s.config.BaseURL, existingURL.ShortCode)
		return &existingURL, nil
	}

	// Add WHERE clause parameters
	args = append(args, shortCode, userID)
	
	updateQuery := fmt.Sprintf(
		"UPDATE url_mappings SET %s WHERE short_code = $%d AND user_id = $%d AND is_active = TRUE",
		strings.Join(updates, ", "), argIndex, argIndex+1,
	)

	_, err = s.db.Exec(updateQuery, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update URL: %w", err)
	}

	// Remove from cache to force refresh
	if s.redis != nil {
		s.redis.DeleteURLMapping(shortCode)
	}

	// Fetch updated record
	err = s.db.QueryRow(checkQuery, shortCode).Scan(
		&existingURL.ID, &existingURL.ShortCode, &existingURL.OriginalURL, &existingURL.CreatedAt,
		&expiresAt, &existingURL.ClickCount, &existingURL.IsActive,
		&userIDDB, &isPublicDB, &customAlias, &title, &description,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to fetch updated URL: %w", err)
	}

	// Set nullable fields
	if expiresAt.Valid {
		existingURL.ExpiresAt = &expiresAt.Time
	}
	if title.Valid {
		existingURL.Title = &title.String
	}
	if description.Valid {
		existingURL.Description = &description.String
	}
	if isPublicDB.Valid {
		existingURL.IsPublic = isPublicDB.Bool
	} else {
		existingURL.IsPublic = true
	}
	existingURL.ShortURL = fmt.Sprintf("%s/%s", s.config.BaseURL, existingURL.ShortCode)

	return &existingURL, nil
}

// recordAttributionTouchpoint records a touchpoint for attribution analysis
func (s *ShortenerService) recordAttributionTouchpoint(shortCode, clientIP, userAgent, referrer string) {
	if s.attribution == nil {
		return
	}

	// For now, we'll just record basic touchpoint info
	// Device/browser parsing can be added later if needed

	// Create touchpoint data
	touchpoint := &models.AttributionTouchpoint{
		SessionID:       generateSessionID(clientIP, userAgent), // Generate session ID from IP and UA
		ShortCode:      shortCode,
		UserIP:         clientIP,
		UserAgent:      userAgent,
		Referrer:       referrer,
		CampaignSource: extractSource(referrer),
		CampaignMedium: extractMedium(referrer),
		CampaignName:   extractCampaign(referrer),
		TouchpointTime: time.Now(),
	}

	// Record the touchpoint asynchronously
	go func() {
		if err := s.attribution.RecordTouchpoint(touchpoint); err != nil {
			// Log error but don't fail the request
			// In production, use proper logging
			fmt.Printf("Failed to record attribution touchpoint: %v\n", err)
		}
	}()
}

// generateSessionID creates a session ID from IP and user agent
func generateSessionID(ip, userAgent string) string {
	h := sha256.Sum256([]byte(ip + userAgent + fmt.Sprintf("%d", time.Now().Unix()/3600))) // 1-hour sessions
	return hex.EncodeToString(h[:])[:16] // Use first 16 characters
}

// determineChannel categorizes the traffic channel from referrer
func determineChannel(referrer string) string {
	if referrer == "" {
		return "direct"
	}
	
	lowerRef := strings.ToLower(referrer)
	
	// Social media channels
	socialDomains := []string{"facebook", "twitter", "linkedin", "instagram", "youtube", "pinterest", "tiktok", "snapchat"}
	for _, domain := range socialDomains {
		if strings.Contains(lowerRef, domain) {
			return "social"
		}
	}
	
	// Search engines
	searchDomains := []string{"google", "bing", "yahoo", "duckduckgo", "baidu"}
	for _, domain := range searchDomains {
		if strings.Contains(lowerRef, domain) {
			return "search"
		}
	}
	
	// Email
	if strings.Contains(lowerRef, "mail") || strings.Contains(lowerRef, "email") {
		return "email"
	}
	
	return "referral"
}

// extractSource extracts the source parameter from referrer URL
func extractSource(referrer string) string {
	return extractUTMParam(referrer, "utm_source")
}

// extractMedium extracts the medium parameter from referrer URL
func extractMedium(referrer string) string {
	return extractUTMParam(referrer, "utm_medium")
}

// extractCampaign extracts the campaign parameter from referrer URL
func extractCampaign(referrer string) string {
	return extractUTMParam(referrer, "utm_campaign")
}

// extractContent extracts the content parameter from referrer URL
func extractContent(referrer string) string {
	return extractUTMParam(referrer, "utm_content")
}

// extractTerm extracts the term parameter from referrer URL
func extractTerm(referrer string) string {
	return extractUTMParam(referrer, "utm_term")
}

// extractUTMParam extracts a specific UTM parameter from URL
func extractUTMParam(referrer, param string) string {
	if referrer == "" {
		return ""
	}
	
	parsedURL, err := url.Parse(referrer)
	if err != nil {
		return ""
	}
	
	return parsedURL.Query().Get(param)
}