package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type GlobalSettingsService struct {
	db *storage.PostgresStorage
}

func NewGlobalSettingsService(db *storage.PostgresStorage) *GlobalSettingsService {
	return &GlobalSettingsService{
		db: db,
	}
}

// GetSetting retrieves a single setting by key
func (s *GlobalSettingsService) GetSetting(key string) (*models.GlobalSetting, error) {
	query := `
		SELECT id, key, value, type, category, display_name, description, 
			is_public, sort_order, created_at, updated_at
		FROM global_settings WHERE key = $1`

	var setting models.GlobalSetting
	var description sql.NullString
	
	err := s.db.QueryRow(query, key).Scan(
		&setting.ID, &setting.Key, &setting.Value, &setting.Type,
		&setting.Category, &setting.DisplayName, &description,
		&setting.IsPublic, &setting.SortOrder, &setting.CreatedAt, &setting.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("setting '%s' not found", key)
		}
		return nil, fmt.Errorf("failed to get setting: %w", err)
	}

	if description.Valid {
		setting.Description = &description.String
	}

	return &setting, nil
}

// GetSettings retrieves multiple settings with optional filtering
func (s *GlobalSettingsService) GetSettings(category string, publicOnly bool) ([]*models.GlobalSetting, error) {
	baseQuery := `
		SELECT id, key, value, type, category, display_name, description, 
			is_public, sort_order, created_at, updated_at
		FROM global_settings`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if category != "" {
		conditions = append(conditions, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, category)
		argIndex++
	}

	if publicOnly {
		conditions = append(conditions, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY category, sort_order, display_name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get settings: %w", err)
	}
	defer rows.Close()

	var settings []*models.GlobalSetting
	for rows.Next() {
		var setting models.GlobalSetting
		var description sql.NullString

		err := rows.Scan(
			&setting.ID, &setting.Key, &setting.Value, &setting.Type,
			&setting.Category, &setting.DisplayName, &description,
			&setting.IsPublic, &setting.SortOrder, &setting.CreatedAt, &setting.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan setting: %w", err)
		}

		if description.Valid {
			setting.Description = &description.String
		}

		settings = append(settings, &setting)
	}

	return settings, nil
}

// GetPublicSettings retrieves all public settings (for frontend use)
func (s *GlobalSettingsService) GetPublicSettings() (map[string]string, error) {
	settings, err := s.GetSettings("", true)
	if err != nil {
		return nil, err
	}

	result := make(map[string]string)
	for _, setting := range settings {
		result[setting.Key] = setting.Value
	}

	return result, nil
}

// GetSettingsByCategory retrieves settings grouped by category
func (s *GlobalSettingsService) GetSettingsByCategory() (map[string][]*models.GlobalSetting, error) {
	settings, err := s.GetSettings("", false)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*models.GlobalSetting)
	for _, setting := range settings {
		result[setting.Category] = append(result[setting.Category], setting)
	}

	return result, nil
}

// CreateSetting creates a new global setting
func (s *GlobalSettingsService) CreateSetting(req *models.GlobalSettingRequest) (*models.GlobalSetting, error) {
	// Check if key already exists
	if _, err := s.GetSetting(req.Key); err == nil {
		return nil, fmt.Errorf("setting with key '%s' already exists", req.Key)
	}

	query := `
		INSERT INTO global_settings (key, value, type, category, display_name, 
			description, is_public, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	var setting models.GlobalSetting
	
	err := s.db.QueryRow(query,
		req.Key, req.Value, req.Type, req.Category, req.DisplayName,
		req.Description, req.IsPublic, req.SortOrder, now, now,
	).Scan(&setting.ID, &setting.CreatedAt, &setting.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create setting: %w", err)
	}

	// Populate the setting with request data
	setting.Key = req.Key
	setting.Value = req.Value
	setting.Type = req.Type
	setting.Category = req.Category
	setting.DisplayName = req.DisplayName
	setting.Description = req.Description
	setting.IsPublic = req.IsPublic
	setting.SortOrder = req.SortOrder

	return &setting, nil
}

// UpdateSetting updates an existing global setting
func (s *GlobalSettingsService) UpdateSetting(key string, req *models.GlobalSettingRequest) (*models.GlobalSetting, error) {
	// Build update query dynamically
	setParts := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	// Only update fields that are provided (not empty)
	if req.Value != "" {
		setParts = append(setParts, fmt.Sprintf("value = $%d", argIndex))
		args = append(args, req.Value)
		argIndex++
	}
	if req.Type != "" {
		setParts = append(setParts, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, req.Type)
		argIndex++
	}
	if req.Category != "" {
		setParts = append(setParts, fmt.Sprintf("category = $%d", argIndex))
		args = append(args, req.Category)
		argIndex++
	}
	if req.DisplayName != "" {
		setParts = append(setParts, fmt.Sprintf("display_name = $%d", argIndex))
		args = append(args, req.DisplayName)
		argIndex++
	}
	if req.Description != nil {
		setParts = append(setParts, fmt.Sprintf("description = $%d", argIndex))
		args = append(args, *req.Description)
		argIndex++
	}
	// Always update these boolean/int fields
	setParts = append(setParts, fmt.Sprintf("is_public = $%d", argIndex))
	args = append(args, req.IsPublic)
	argIndex++
	
	setParts = append(setParts, fmt.Sprintf("sort_order = $%d", argIndex))
	args = append(args, req.SortOrder)
	argIndex++

	// Add WHERE clause
	args = append(args, key)
	whereClause := fmt.Sprintf("WHERE key = $%d", argIndex)

	query := fmt.Sprintf("UPDATE global_settings SET %s %s", strings.Join(setParts, ", "), whereClause)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to update setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return nil, fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return nil, fmt.Errorf("setting with key '%s' not found", key)
	}

	// Return updated setting
	return s.GetSetting(key)
}

// DeleteSetting deletes a global setting
func (s *GlobalSettingsService) DeleteSetting(key string) error {
	query := `DELETE FROM global_settings WHERE key = $1`
	result, err := s.db.Exec(query, key)
	if err != nil {
		return fmt.Errorf("failed to delete setting: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("setting with key '%s' not found", key)
	}

	return nil
}

// BulkUpdateSettings updates multiple settings at once
func (s *GlobalSettingsService) BulkUpdateSettings(settings map[string]string) error {
	if len(settings) == 0 {
		return nil
	}

	// Start transaction
	tx, err := s.db.Begin()
	if err != nil {
		return fmt.Errorf("failed to start transaction: %w", err)
	}
	defer tx.Rollback()

	// Update each setting
	for key, value := range settings {
		query := `UPDATE global_settings SET value = $1, updated_at = $2 WHERE key = $3`
		_, err := tx.Exec(query, value, time.Now(), key)
		if err != nil {
			return fmt.Errorf("failed to update setting '%s': %w", key, err)
		}
	}

	// Commit transaction
	if err := tx.Commit(); err != nil {
		return fmt.Errorf("failed to commit transaction: %w", err)
	}

	return nil
}

// GetSettingValue is a helper to get just the value of a setting
func (s *GlobalSettingsService) GetSettingValue(key string, defaultValue string) string {
	setting, err := s.GetSetting(key)
	if err != nil {
		return defaultValue
	}
	return setting.Value
}

// IsMaintenanceMode checks if the site is in maintenance mode
func (s *GlobalSettingsService) IsMaintenanceMode() bool {
	return s.GetSettingValue("maintenance_mode", "false") == "true"
}

// GetSiteInfo returns basic site information
func (s *GlobalSettingsService) GetSiteInfo() map[string]string {
	result := make(map[string]string)
	
	// Get basic site info
	keys := []string{"site_name", "site_tagline", "company_email", "company_phone", "company_address", "meta_description"}
	for _, key := range keys {
		if setting, err := s.GetSetting(key); err == nil {
			result[key] = setting.Value
		}
	}

	return result
}

// GetSocialLinks returns social media links
func (s *GlobalSettingsService) GetSocialLinks() map[string]string {
	settings, err := s.GetSettings("social", true)
	if err != nil {
		return make(map[string]string)
	}

	result := make(map[string]string)
	for _, setting := range settings {
		if setting.Value != "" {
			result[setting.Key] = setting.Value
		}
	}

	return result
}