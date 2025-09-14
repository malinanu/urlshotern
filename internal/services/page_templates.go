package services

import (
	"database/sql"
	"encoding/json"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type PageTemplateService struct {
	db                    *storage.PostgresStorage
	contentBlockService   *ContentBlockService
	globalSettingsService *GlobalSettingsService
}

func NewPageTemplateService(db *storage.PostgresStorage, contentBlockService *ContentBlockService, globalSettingsService *GlobalSettingsService) *PageTemplateService {
	return &PageTemplateService{
		db:                    db,
		contentBlockService:   contentBlockService,
		globalSettingsService: globalSettingsService,
	}
}

// GetPageTemplate retrieves a page template by slug
func (s *PageTemplateService) GetPageTemplate(slug string) (*models.PageTemplate, error) {
	query := `
		SELECT id, name, slug, description, content, fields, preview_url, is_active, sort_order, created_at, updated_at
		FROM page_templates WHERE slug = $1`

	var template models.PageTemplate
	var description sql.NullString
	var fields sql.NullString
	var previewURL sql.NullString
	
	err := s.db.QueryRow(query, slug).Scan(
		&template.ID, &template.Name, &template.Slug, &description,
		&template.Content, &fields, &previewURL, &template.IsActive,
		&template.SortOrder, &template.CreatedAt, &template.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("page template '%s' not found", slug)
		}
		return nil, fmt.Errorf("failed to get page template: %w", err)
	}

	if description.Valid {
		template.Description = &description.String
	}
	if fields.Valid {
		template.Fields = &fields.String
	}
	if previewURL.Valid {
		template.PreviewURL = &previewURL.String
	}

	return &template, nil
}

// GetPageTemplateByID retrieves a page template by ID
func (s *PageTemplateService) GetPageTemplateByID(id int64) (*models.PageTemplate, error) {
	query := `
		SELECT id, name, slug, description, content, fields, preview_url, is_active, sort_order, created_at, updated_at
		FROM page_templates WHERE id = $1`

	var template models.PageTemplate
	var description sql.NullString
	var fields sql.NullString
	var previewURL sql.NullString
	
	err := s.db.QueryRow(query, id).Scan(
		&template.ID, &template.Name, &template.Slug, &description,
		&template.Content, &fields, &previewURL, &template.IsActive,
		&template.SortOrder, &template.CreatedAt, &template.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("page template with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get page template: %w", err)
	}

	if description.Valid {
		template.Description = &description.String
	}
	if fields.Valid {
		template.Fields = &fields.String
	}
	if previewURL.Valid {
		template.PreviewURL = &previewURL.String
	}

	return &template, nil
}

// ListPageTemplates lists all page templates
func (s *PageTemplateService) ListPageTemplates(activeOnly bool) ([]*models.PageTemplate, error) {
	query := `
		SELECT id, name, slug, description, content, fields, preview_url, is_active, sort_order, created_at, updated_at
		FROM page_templates`
	
	if activeOnly {
		query += " WHERE is_active = true"
	}
	query += " ORDER BY sort_order, name"

	rows, err := s.db.Query(query)
	if err != nil {
		return nil, fmt.Errorf("failed to list page templates: %w", err)
	}
	defer rows.Close()

	var templates []*models.PageTemplate
	for rows.Next() {
		var template models.PageTemplate
		var description sql.NullString
		var fields sql.NullString
		var previewURL sql.NullString
		
		err := rows.Scan(
			&template.ID, &template.Name, &template.Slug, &description,
			&template.Content, &fields, &previewURL, &template.IsActive,
			&template.SortOrder, &template.CreatedAt, &template.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan page template: %w", err)
		}

		if description.Valid {
			template.Description = &description.String
		}
		if fields.Valid {
			template.Fields = &fields.String
		}
		if previewURL.Valid {
			template.PreviewURL = &previewURL.String
		}

		templates = append(templates, &template)
	}

	return templates, nil
}

// CreatePageTemplate creates a new page template
func (s *PageTemplateService) CreatePageTemplate(name, slug, content string, description *string, fields *string, previewURL *string) (*models.PageTemplate, error) {
	// Check if slug already exists
	if _, err := s.GetPageTemplate(slug); err == nil {
		return nil, fmt.Errorf("page template with slug '%s' already exists", slug)
	}

	query := `
		INSERT INTO page_templates (name, slug, description, content, fields, preview_url, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	var template models.PageTemplate
	
	err := s.db.QueryRow(query,
		name, slug, description, content, fields, previewURL,
		true, 0, now, now,
	).Scan(&template.ID, &template.CreatedAt, &template.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create page template: %w", err)
	}

	// Populate the template with request data
	template.Name = name
	template.Slug = slug
	template.Description = description
	template.Content = content
	template.Fields = fields
	template.PreviewURL = previewURL
	template.IsActive = true
	template.SortOrder = 0

	return &template, nil
}

// UpdatePageTemplate updates an existing page template
func (s *PageTemplateService) UpdatePageTemplate(id int64, name, slug, content string, description *string, fields *string, previewURL *string, isActive bool, sortOrder int) (*models.PageTemplate, error) {
	// Check if template exists
	existing, err := s.GetPageTemplateByID(id)
	if err != nil {
		return nil, err
	}

	// Check if slug is being changed and if new slug already exists
	if slug != existing.Slug {
		if _, err := s.GetPageTemplate(slug); err == nil {
			return nil, fmt.Errorf("page template with slug '%s' already exists", slug)
		}
	}

	query := `
		UPDATE page_templates 
		SET name = $1, slug = $2, description = $3, content = $4, fields = $5, 
			preview_url = $6, is_active = $7, sort_order = $8, updated_at = $9
		WHERE id = $10`

	_, err = s.db.Exec(query,
		name, slug, description, content, fields, previewURL,
		isActive, sortOrder, time.Now(), id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update page template: %w", err)
	}

	// Return updated template
	return s.GetPageTemplateByID(id)
}

// DeletePageTemplate deletes a page template
func (s *PageTemplateService) DeletePageTemplate(id int64) error {
	query := `DELETE FROM page_templates WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete page template: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("page template with ID %d not found", id)
	}

	return nil
}

// RenderPageTemplate renders a page template with content and variables
func (s *PageTemplateService) RenderPageTemplate(templateSlug string, pageContent string, pageData map[string]interface{}) (string, error) {
	template, err := s.GetPageTemplate(templateSlug)
	if err != nil {
		return "", err
	}

	if !template.IsActive {
		return "", fmt.Errorf("page template '%s' is not active", templateSlug)
	}

	// Get global settings for variable substitution
	settings, err := s.globalSettingsService.GetPublicSettings()
	if err != nil {
		settings = make(map[string]string)
	}

	// Start with the template content
	result := template.Content

	// Replace {{content}} with page content
	result = strings.ReplaceAll(result, "{{content}}", pageContent)

	// Replace global setting variables
	for key, value := range settings {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}

	// Replace page-specific variables
	for key, value := range pageData {
		placeholder := fmt.Sprintf("{{%s}}", key)
		if strValue, ok := value.(string); ok {
			result = strings.ReplaceAll(result, placeholder, strValue)
		} else if value != nil {
			result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", value))
		}
	}

	// Process content blocks
	result = s.processContentBlocks(result)

	return result, nil
}

// processContentBlocks processes {{block:identifier}} patterns
func (s *PageTemplateService) processContentBlocks(content string) string {
	// Find all {{block:identifier}} patterns
	for {
		start := strings.Index(content, "{{block:")
		if start == -1 {
			break
		}

		end := strings.Index(content[start:], "}}")
		if end == -1 {
			break
		}
		end += start + 2

		// Extract block identifier
		blockPattern := content[start:end]
		identifier := strings.TrimPrefix(strings.TrimSuffix(blockPattern, "}}"), "{{block:")

		// Get block content
		blockContent, err := s.contentBlockService.RenderContentBlock(identifier)
		if err != nil {
			// Replace with empty string if block not found or error
			blockContent = ""
		}

		// Replace the pattern with block content
		content = strings.ReplaceAll(content, blockPattern, blockContent)
	}

	return content
}

// GetTemplateFields parses and returns template fields as a map
func (s *PageTemplateService) GetTemplateFields(templateSlug string) (map[string]interface{}, error) {
	template, err := s.GetPageTemplate(templateSlug)
	if err != nil {
		return nil, err
	}

	if template.Fields == nil || *template.Fields == "" {
		return make(map[string]interface{}), nil
	}

	var fields map[string]interface{}
	err = json.Unmarshal([]byte(*template.Fields), &fields)
	if err != nil {
		return nil, fmt.Errorf("failed to parse template fields: %w", err)
	}

	return fields, nil
}

// DuplicatePageTemplate creates a copy of an existing page template
func (s *PageTemplateService) DuplicatePageTemplate(id int64, newSlug, newName string) (*models.PageTemplate, error) {
	// Get original template
	original, err := s.GetPageTemplateByID(id)
	if err != nil {
		return nil, err
	}

	// Create copy
	return s.CreatePageTemplate(
		newName,
		newSlug,
		original.Content,
		original.Description,
		original.Fields,
		original.PreviewURL,
	)
}

// ValidateTemplate validates template content for common issues
func (s *PageTemplateService) ValidateTemplate(content string) []string {
	var warnings []string

	// Check for common issues
	if !strings.Contains(content, "{{content}}") {
		warnings = append(warnings, "Template doesn't contain {{content}} placeholder")
	}

	// Check for unclosed template tags
	openTags := strings.Count(content, "{{")
	closeTags := strings.Count(content, "}}")
	if openTags != closeTags {
		warnings = append(warnings, "Mismatched template tags ({{ and }})")
	}

	// Check for common HTML issues
	if strings.Contains(content, "<script>") {
		warnings = append(warnings, "Template contains script tags - ensure they are safe")
	}

	return warnings
}