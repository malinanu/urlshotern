package services

import (
	"fmt"
	"net"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/utils"
)

// CreateInternalShortURL creates a short link that points at one of this
// site's own microsite pages (/m/<slug>). The public ShortenURL path rejects
// own-host destinations (ErrSelfReferentialURL) to prevent redirect loops;
// this separate path accepts only BaseURL + "/m/<slug>", so that protection
// stays intact for everything else.
func (s *ShortenerService) CreateInternalShortURL(targetURL string, userID int64, clientIP string) (*models.URLMapping, error) {
	if net.ParseIP(clientIP) == nil {
		clientIP = "127.0.0.1" // created_by_ip is INET
	}
	prefix := strings.TrimRight(s.config.BaseURL, "/") + "/m/"
	slug := strings.TrimPrefix(targetURL, prefix)
	if !strings.HasPrefix(targetURL, prefix) || !slugPattern.MatchString(slug) {
		return nil, ErrInvalidURL
	}

	var shortCode string
	var id int64
	for attempts := 0; attempts < 3; attempts++ {
		newID, err := utils.GenerateID()
		if err != nil {
			return nil, fmt.Errorf("failed to generate ID: %w", err)
		}
		code := utils.EncodeBase62(newID)
		exists, err := s.db.ShortCodeExists(code)
		if err != nil {
			return nil, fmt.Errorf("failed to check short code uniqueness: %w", err)
		}
		if !exists {
			id, shortCode = newID, code
			break
		}
	}
	if shortCode == "" {
		return nil, fmt.Errorf("failed to generate a unique short code")
	}

	uid := userID
	mapping := &models.URLMapping{
		ID:          id,
		ShortCode:   shortCode,
		OriginalURL: targetURL,
		CreatedAt:   time.Now(),
		IsActive:    true,
		CreatedByIP: clientIP,
		UserID:      &uid,
	}
	if err := s.db.SaveURLMapping(mapping); err != nil {
		return nil, fmt.Errorf("failed to save URL mapping: %w", err)
	}
	if s.redis != nil {
		_ = s.redis.SetURLMapping(shortCode, mapping, 24*time.Hour)
	}
	return mapping, nil
}
