package services

import (
	"crypto/md5"
	"fmt"
	"image"
	_ "image/gif"
	_ "image/jpeg"
	_ "image/png"
	"io"
	"mime/multipart"
	"os"
	"path/filepath"
	"strings"
	"time"

	"github.com/URLshorter/url-shortener/internal/models"
	"github.com/URLshorter/url-shortener/internal/storage"
	"github.com/google/uuid"
)

type MediaService struct {
	db         *storage.PostgresStorage
	uploadPath string
	baseURL    string
}

func NewMediaService(db *storage.PostgresStorage, uploadPath, baseURL string) *MediaService {
	return &MediaService{
		db:         db,
		uploadPath: uploadPath,
		baseURL:    baseURL,
	}
}

// UploadFile handles file upload and creates media record
func (s *MediaService) UploadFile(userID int64, fileHeader *multipart.FileHeader, req *models.MediaUploadRequest) (*models.MediaFile, error) {
	// Open the uploaded file
	file, err := fileHeader.Open()
	if err != nil {
		return nil, fmt.Errorf("failed to open uploaded file: %w", err)
	}
	defer file.Close()

	// Generate unique filename
	ext := filepath.Ext(fileHeader.Filename)
	filename := fmt.Sprintf("%s_%s%s", uuid.New().String(), generateMD5Hash(fileHeader.Filename), ext)
	
	// Determine file type
	fileType := s.getFileType(fileHeader.Header.Get("Content-Type"))
	
	// Create full file path
	fullPath := filepath.Join(s.uploadPath, fileType, filename)
	
	// Ensure directory exists
	if err := os.MkdirAll(filepath.Dir(fullPath), 0755); err != nil {
		return nil, fmt.Errorf("failed to create upload directory: %w", err)
	}

	// Create the destination file
	dst, err := os.Create(fullPath)
	if err != nil {
		return nil, fmt.Errorf("failed to create destination file: %w", err)
	}
	defer dst.Close()

	// Copy file content
	fileSize, err := io.Copy(dst, file)
	if err != nil {
		return nil, fmt.Errorf("failed to copy file content: %w", err)
	}

	// Get image dimensions if it's an image
	var width, height *int
	if fileType == "image" {
		if w, h, err := s.getImageDimensions(fullPath); err == nil {
			width = &w
			height = &h
		}
	}

	// Generate file URL
	relativePathparts := []string{fileType, filename}
	relativePath := strings.Join(relativePathparts, "/")
	fileURL := fmt.Sprintf("%s/uploads/%s", s.baseURL, relativePath)

	// Create media record
	mediaFile := &models.MediaFile{
		UserID:       userID,
		Filename:     filename,
		OriginalName: fileHeader.Filename,
		FilePath:     fullPath,
		FileURL:      fileURL,
		FileSize:     fileSize,
		MimeType:     fileHeader.Header.Get("Content-Type"),
		FileType:     fileType,
		Width:        width,
		Height:       height,
		Alt:          req.Alt,
		Caption:      req.Caption,
		IsPublic:     req.IsPublic,
	}

	// Save to database
	query := `
		INSERT INTO media_files (user_id, filename, original_name, file_path, file_url, 
			file_size, mime_type, file_type, width, height, alt, caption, is_public, created_at, updated_at)
		VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11, $12, $13, $14, $15)
		RETURNING id, created_at, updated_at`

	now := time.Now()
	err = s.db.QueryRow(query,
		mediaFile.UserID, mediaFile.Filename, mediaFile.OriginalName, mediaFile.FilePath,
		mediaFile.FileURL, mediaFile.FileSize, mediaFile.MimeType, mediaFile.FileType,
		mediaFile.Width, mediaFile.Height, mediaFile.Alt, mediaFile.Caption,
		mediaFile.IsPublic, now, now,
	).Scan(&mediaFile.ID, &mediaFile.CreatedAt, &mediaFile.UpdatedAt)

	if err != nil {
		// Clean up uploaded file if database insert fails
		os.Remove(fullPath)
		return nil, fmt.Errorf("failed to save media file record: %w", err)
	}

	return mediaFile, nil
}

// GetMediaFile retrieves a media file by ID
func (s *MediaService) GetMediaFile(id int64, userID int64) (*models.MediaFile, error) {
	query := `
		SELECT id, user_id, filename, original_name, file_path, file_url, file_size, 
			mime_type, file_type, width, height, alt, caption, is_public, created_at, updated_at
		FROM media_files 
		WHERE id = $1 AND (user_id = $2 OR is_public = true)`

	var media models.MediaFile
	err := s.db.QueryRow(query, id, userID).Scan(
		&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
		&media.FilePath, &media.FileURL, &media.FileSize, &media.MimeType,
		&media.FileType, &media.Width, &media.Height, &media.Alt,
		&media.Caption, &media.IsPublic, &media.CreatedAt, &media.UpdatedAt,
	)

	if err != nil {
		return nil, fmt.Errorf("failed to get media file: %w", err)
	}

	return &media, nil
}

// ListMediaFiles lists media files for a user
func (s *MediaService) ListMediaFiles(userID int64, fileType string, limit, offset int) ([]*models.MediaFile, int64, error) {
	// Build query with optional file type filter
	baseQuery := `FROM media_files WHERE user_id = $1`
	args := []interface{}{userID}
	argIndex := 2

	if fileType != "" {
		baseQuery += fmt.Sprintf(" AND file_type = $%d", argIndex)
		args = append(args, fileType)
		argIndex++
	}

	// Get total count
	countQuery := "SELECT COUNT(*) " + baseQuery
	var total int64
	err := s.db.QueryRow(countQuery, args...).Scan(&total)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to count media files: %w", err)
	}

	// Get files
	selectQuery := fmt.Sprintf(`
		SELECT id, user_id, filename, original_name, file_path, file_url, file_size, 
			mime_type, file_type, width, height, alt, caption, is_public, created_at, updated_at
		%s ORDER BY created_at DESC LIMIT $%d OFFSET $%d`, 
		baseQuery, argIndex, argIndex+1)
	
	args = append(args, limit, offset)

	rows, err := s.db.Query(selectQuery, args...)
	if err != nil {
		return nil, 0, fmt.Errorf("failed to list media files: %w", err)
	}
	defer rows.Close()

	var files []*models.MediaFile
	for rows.Next() {
		var media models.MediaFile
		err := rows.Scan(
			&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
			&media.FilePath, &media.FileURL, &media.FileSize, &media.MimeType,
			&media.FileType, &media.Width, &media.Height, &media.Alt,
			&media.Caption, &media.IsPublic, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, 0, fmt.Errorf("failed to scan media file: %w", err)
		}
		files = append(files, &media)
	}

	return files, total, nil
}

// UpdateMediaFile updates media file metadata
func (s *MediaService) UpdateMediaFile(id, userID int64, alt, caption *string, isPublic *bool) error {
	setParts := []string{"updated_at = $1"}
	args := []interface{}{time.Now()}
	argIndex := 2

	if alt != nil {
		setParts = append(setParts, fmt.Sprintf("alt = $%d", argIndex))
		args = append(args, *alt)
		argIndex++
	}
	if caption != nil {
		setParts = append(setParts, fmt.Sprintf("caption = $%d", argIndex))
		args = append(args, *caption)
		argIndex++
	}
	if isPublic != nil {
		setParts = append(setParts, fmt.Sprintf("is_public = $%d", argIndex))
		args = append(args, *isPublic)
		argIndex++
	}

	args = append(args, id, userID)
	query := fmt.Sprintf("UPDATE media_files SET %s WHERE id = $%d AND user_id = $%d", 
		strings.Join(setParts, ", "), argIndex, argIndex+1)

	result, err := s.db.Exec(query, args...)
	if err != nil {
		return fmt.Errorf("failed to update media file: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media file not found or access denied")
	}

	return nil
}

// DeleteMediaFile deletes a media file
func (s *MediaService) DeleteMediaFile(id, userID int64) error {
	// First get file info to delete physical file
	media, err := s.GetMediaFile(id, userID)
	if err != nil {
		return err
	}

	// Delete from database
	query := `DELETE FROM media_files WHERE id = $1 AND user_id = $2`
	result, err := s.db.Exec(query, id, userID)
	if err != nil {
		return fmt.Errorf("failed to delete media file record: %w", err)
	}

	rowsAffected, err := result.RowsAffected()
	if err != nil {
		return fmt.Errorf("failed to get affected rows: %w", err)
	}

	if rowsAffected == 0 {
		return fmt.Errorf("media file not found or access denied")
	}

	// Delete physical file
	if err := os.Remove(media.FilePath); err != nil && !os.IsNotExist(err) {
		// Log error but don't fail the operation
		fmt.Printf("Warning: failed to delete physical file %s: %v\n", media.FilePath, err)
	}

	return nil
}

// Helper functions

func (s *MediaService) getFileType(mimeType string) string {
	switch {
	case strings.HasPrefix(mimeType, "image/"):
		return "image"
	case strings.HasPrefix(mimeType, "video/"):
		return "video"
	case strings.HasPrefix(mimeType, "audio/"):
		return "audio"
	case mimeType == "application/pdf":
		return "document"
	case strings.Contains(mimeType, "document") || strings.Contains(mimeType, "text"):
		return "document"
	default:
		return "other"
	}
}

func (s *MediaService) getImageDimensions(filePath string) (int, int, error) {
	file, err := os.Open(filePath)
	if err != nil {
		return 0, 0, err
	}
	defer file.Close()

	img, _, err := image.DecodeConfig(file)
	if err != nil {
		return 0, 0, err
	}

	return img.Width, img.Height, nil
}

func generateMD5Hash(input string) string {
	hash := md5.Sum([]byte(input + time.Now().String()))
	return fmt.Sprintf("%x", hash)[:8]
}

// ValidateFile validates uploaded file
func (s *MediaService) ValidateFile(fileHeader *multipart.FileHeader) error {
	// Check file size (10MB limit)
	maxSize := int64(10 << 20) // 10MB
	if fileHeader.Size > maxSize {
		return fmt.Errorf("file size exceeds 10MB limit")
	}

	// Check file extension
	ext := strings.ToLower(filepath.Ext(fileHeader.Filename))
	allowedExts := map[string]bool{
		".jpg":  true, ".jpeg": true, ".png": true, ".gif": true, ".svg": true,
		".pdf":  true, ".doc": true, ".docx": true, ".txt": true,
		".mp4":  true, ".avi": true, ".mov": true,
		".mp3":  true, ".wav": true,
		".zip":  true, ".rar": true,
	}

	if !allowedExts[ext] {
		return fmt.Errorf("file type %s is not allowed", ext)
	}

	return nil
}

// GetFilesByIDs retrieves multiple files by their IDs
func (s *MediaService) GetFilesByIDs(ids []int64, userID int64) ([]*models.MediaFile, error) {
	if len(ids) == 0 {
		return []*models.MediaFile{}, nil
	}

	// Build query with placeholders for IDs
	placeholders := make([]string, len(ids))
	args := []interface{}{userID}
	for i, id := range ids {
		placeholders[i] = fmt.Sprintf("$%d", i+2)
		args = append(args, id)
	}

	query := fmt.Sprintf(`
		SELECT id, user_id, filename, original_name, file_path, file_url, file_size, 
			mime_type, file_type, width, height, alt, caption, is_public, created_at, updated_at
		FROM media_files 
		WHERE (user_id = $1 OR is_public = true) AND id IN (%s)
		ORDER BY created_at DESC`, strings.Join(placeholders, ","))

	rows, err := s.db.Query(query, args...)
	if err != nil {
		return nil, fmt.Errorf("failed to get media files: %w", err)
	}
	defer rows.Close()

	var files []*models.MediaFile
	for rows.Next() {
		var media models.MediaFile
		err := rows.Scan(
			&media.ID, &media.UserID, &media.Filename, &media.OriginalName,
			&media.FilePath, &media.FileURL, &media.FileSize, &media.MimeType,
			&media.FileType, &media.Width, &media.Height, &media.Alt,
			&media.Caption, &media.IsPublic, &media.CreatedAt, &media.UpdatedAt,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to scan media file: %w", err)
		}
		files = append(files, &media)
	}

	return files, nil
}