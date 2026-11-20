package services

import (
	"crypto/sha256"
	"encoding/hex"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strconv"
)

// maxCardImagesPerUser caps stored photos and logos per account.
const maxCardImagesPerUser = 60

// CardImageService stores the profile photos and company logos uploaded for
// MicroPages digital cards. It reuses the hardened ProcessImage pipeline
// (content sniffing, dimension limits, EXIF orientation, re-encode) but needs no
// database: files are content-addressed under mp/<userID>/<sha256>.<ext>, so an
// identical re-upload is a no-op.
type CardImageService struct {
	store    *LocalAssetStore
	maxBytes int64
}

// NewCardImageService creates the service. maxBytes <= 0 means 2 MB.
func NewCardImageService(store *LocalAssetStore, maxBytes int64) *CardImageService {
	if maxBytes <= 0 {
		maxBytes = 2 * 1024 * 1024
	}
	return &CardImageService{store: store, maxBytes: maxBytes}
}

// MaxBytes is the upload size limit.
func (s *CardImageService) MaxBytes() int64 { return s.maxBytes }

// Upload validates, re-encodes and stores an image, returning its public URL path.
func (s *CardImageService) Upload(userID int64, r io.Reader) (string, error) {
	img, err := ProcessImage(r, s.maxBytes)
	if err != nil {
		return "", err
	}
	sum := sha256.Sum256(img.Data)
	rel := fmt.Sprintf("mp/%d/%s.%s", userID, hex.EncodeToString(sum[:]), img.Ext)

	dir := filepath.Join(s.store.Root(), "mp", strconv.FormatInt(userID, 10))
	if _, err := os.Stat(filepath.Join(s.store.Root(), filepath.FromSlash(rel))); err == nil {
		return "/uploads/" + rel, nil // already stored
	}
	if entries, err := os.ReadDir(dir); err == nil && len(entries) >= maxCardImagesPerUser {
		return "", ErrAssetQuota
	}
	if err := s.store.Put(rel, img.Data); err != nil {
		return "", fmt.Errorf("store image: %w", err)
	}
	return "/uploads/" + rel, nil
}
