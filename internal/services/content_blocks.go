package services

import (
	"database/sql"
	"fmt"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
)

type ContentBlockService struct {
	db *storage.PostgresStorage
	globalSettingsService *GlobalSettingsService
}

func NewContentBlockService(db *storage.PostgresStorage, globalSettingsService *GlobalSettingsService) *ContentBlockService {
	return &ContentBlockService{
		db: db,
		globalSettingsService: globalSettingsService,
	}
}

// GetContentBlock retrieves a content block by identifier
func (s *ContentBlockService) GetContentBlock(identifier string) (*models.ContentBlock, error) {
	query := `
		SELECT id, name, identifier, content, type, is_active, sort_order, created_at, updated_at
		FROM content_blocks WHERE identifier = $1`

	var block models.ContentBlock
	err := s.db.QueryRow(query, identifier).Scan(
		&block.ID, &block.Name, &block.Identifier, &block.Content,
		&block.Type, &block.IsActive, &block.SortOrder, &block.CreatedAt, &block.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content block '%s' not found", identifier)
		}
		return nil, fmt.Errorf("failed to get content block: %w", err)
	}

	return &block, nil
}

// GetContentBlockByID retrieves a content block by ID
func (s *ContentBlockService) GetContentBlockByID(id int64) (*models.ContentBlock, error) {
	query := `
		SELECT id, name, identifier, content, type, is_active, sort_order, created_at, updated_at
		FROM content_blocks WHERE id = $1`

	var block models.ContentBlock
	err := s.db.QueryRow(query, id).Scan(
		&block.ID, &block.Name, &block.Identifier, &block.Content,
		&block.Type, &block.IsActive, &block.SortOrder, &block.CreatedAt, &block.UpdatedAt,
	)
	
	if err != nil {
		if err == sql.ErrNoRows {
			return nil, fmt.Errorf("content block with ID %d not found", id)
		}
		return nil, fmt.Errorf("failed to get content block: %w", err)
	}

	return &block, nil
}

// ListContentBlocks lists all content blocks with optional filtering
func (s *ContentBlockService) ListContentBlocks(blockType string, activeOnly bool) ([]*models.ContentBlock, error) {
	baseQuery := `
		SELECT id, name, identifier, content, type, is_active, sort_order, created_at, updated_at
		FROM content_blocks`
	
	var conditions []string
	var args []interface{}
	argIndex := 1

	if blockType != "" {
		conditions = append(conditions, fmt.Sprintf("type = $%d", argIndex))
		args = append(args, blockType)
		argIndex++
	}

	if activeOnly {
		conditions = append(conditions, fmt.Sprintf("is_active = $%d", argIndex))
		args = append(args, true)
		argIndex++
	}

	query := baseQuery
	if len(conditions) > 0 {
		query += " WHERE " + strings.Join(conditions, " AND ")
	}
	query += " ORDER BY type, sort_order, name"

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to list content blocks: %w", err)
	}
	defer rows.Close()

	var blocks []*models.ContentBlock
	for rows.Next() {
		var block models.ContentBlock
		err := rows.Scan(
			&block.ID, &block.Name, &block.Identifier, &block.Content,
			&block.Type, &block.IsActive, &block.SortOrder, &block.CreatedAt, &block.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan content block: %w", err)
		}
		blocks = append(blocks, &block)
	}

	return blocks, nil
}

// CreateContentBlock creates a new content block
func (s *ContentBlockService) CreateContentBlock(req *models.ContentBlockRequest) (*models.ContentBlock, error) {
	// Check if identifier already exists
	if _, err := s.GetContentBlock(req.Identifier); err == nil {
		return nil, fmt.Errorf("content block with identifier '%s' already exists", req.Identifier)
	}

	query := `
		INSERT INTO content_blocks (name, identifier, content, type, is_active, sort_order, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	var block models.ContentBlock
	
	err := s.db.QueryRow(query,
		req.Name, req.Identifier, req.Content, req.Type,
		req.IsActive, req.SortOrder, now, now,
	).Scan(&block.ID, &block.CreatedAt, &block.UpdatedAt)
	
	if err != nil {
		return nil, fmt.Errorf("failed to create content block: %w", err)
	}

	// Populate the block with request data
	block.Name = req.Name
	block.Identifier = req.Identifier
	block.Content = req.Content
	block.Type = req.Type
	block.IsActive = req.IsActive
	block.SortOrder = req.SortOrder

	return &block, nil
}

// UpdateContentBlock updates an existing content block
func (s *ContentBlockService) UpdateContentBlock(id int64, req *models.ContentBlockRequest) (*models.ContentBlock, error) {
	// Check if block exists
	existing, err := s.GetContentBlockByID(id)
	if err != nil {
		return nil, err
	}

	// Check if identifier is being changed and if new identifier already exists
	if req.Identifier != existing.Identifier {
		if _, err := s.GetContentBlock(req.Identifier); err == nil {
			return nil, fmt.Errorf("content block with identifier '%s' already exists", req.Identifier)
		}
	}

	query := `
		UPDATE content_blocks 
		SET name = $1, identifier = $2, content = $3, type = $4, is_active = $5, sort_order = $6, updated_at = $7
		WHERE id = $8`

	_, err = s.db.Exec(query,
		req.Name, req.Identifier, req.Content, req.Type,
		req.IsActive, req.SortOrder, time.Now(), id,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to update content block: %w", err)
	}

	// Return updated block
	return s.GetContentBlockByID(id)
}

// DeleteContentBlock deletes a content block
func (s *ContentBlockService) DeleteContentBlock(id int64) error {
	query := `DELETE FROM content_blocks WHERE id = $1`
	result, err := s.db.Exec(query, id)
	if err != nil {
		return fmt.Errorf("failed to delete content block: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("content block with ID %d not found", id)
	}

	return nil
}

// RenderContentBlock renders a content block with variable substitution
func (s *ContentBlockService) RenderContentBlock(identifier string) (string, error) {
	block, err := s.GetContentBlock(identifier)
	if err != nil {
		return "", err
	}

	if !block.IsActive {
		return "", fmt.Errorf("content block '%s' is not active", identifier)
	}

	// Get global settings for variable substitution
	settings, err := s.globalSettingsService.GetPublicSettings()
	if err != nil {
		// Continue without settings if there's an error
		settings = make(map[string]string)
	}

	// Render content with variable substitution
	return s.renderTemplate(block.Content, settings), nil
}

// renderTemplate performs simple template variable substitution
func (s *ContentBlockService) renderTemplate(content string, variables map[string]string) string {
	result := content
	
	// Replace {{variable}} patterns with actual values
	for key, value := range variables {
		placeholder := fmt.Sprintf("{{%s}}", key)
		result = strings.ReplaceAll(result, placeholder, value)
	}
	
	return result
}

// GetContentBlocksByType returns content blocks grouped by type
func (s *ContentBlockService) GetContentBlocksByType() (map[string][]*models.ContentBlock, error) {
	blocks, err := s.ListContentBlocks("", false)
	if err != nil {
		return nil, err
	}

	result := make(map[string][]*models.ContentBlock)
	for _, block := range blocks {
		result[block.Type] = append(result[block.Type], block)
	}

	return result, nil
}

// RenderMultipleBlocks renders multiple content blocks at once
func (s *ContentBlockService) RenderMultipleBlocks(identifiers []string) (map[string]string, error) {
	if len(identifiers) == 0 {
		return make(map[string]string), nil
	}

	// Get global settings once for all blocks
	settings, err := s.globalSettingsService.GetPublicSettings()
	if err != nil {
		settings = make(map[string]string)
	}

	result := make(map[string]string)
	
	for _, identifier := range identifiers {
		block, err := s.GetContentBlock(identifier)
		if err != nil {
			continue // Skip missing blocks
		}

		if !block.IsActive {
			continue // Skip inactive blocks
		}

		result[identifier] = s.renderTemplate(block.Content, settings)
	}

	return result, nil
}

// DuplicateContentBlock creates a copy of an existing content block
func (s *ContentBlockService) DuplicateContentBlock(id int64, newIdentifier, newName string) (*models.ContentBlock, error) {
	// Get original block
	original, err := s.GetContentBlockByID(id)
	if err != nil {
		return nil, err
	}

	// Create copy request
	req := &models.ContentBlockRequest{
		Name:       newName,
		Identifier: newIdentifier,
		Content:    original.Content,
		Type:       original.Type,
		IsActive:   false, // Start as inactive
		SortOrder:  original.SortOrder,
	}

	return s.CreateContentBlock(req)
}