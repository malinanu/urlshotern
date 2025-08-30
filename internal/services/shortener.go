package services

import (
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
	db     *storage.PostgresStorage
	redis  *storage.RedisStorage
	config *configs.Config
}

// NewShortenerService creates a new shortener service
func NewShortenerService(db *storage.PostgresStorage, redis *storage.RedisStorage, config *configs.Config) *ShortenerService {
	// Initialize Snowflake ID generator
	utils.InitializeSnowflake(config.NodeID)
	
	return &ShortenerService{
		db:     db,
		redis:  redis,
		config: config,
	}
}

// ShortenURL creates a new short URL
func (s *ShortenerService) ShortenURL(request *models.ShortenRequest, clientIP string) (*models.ShortenResponse, error) {
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

	// Increment click count in database
	go func() {
		if err := s.db.IncrementClickCount(shortCode); err != nil {
			fmt.Printf("Failed to increment click count in DB: %v\n", err)
		}
	}()

	// Increment cached click count for faster analytics
	s.redis.IncrementClickCount(shortCode)

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