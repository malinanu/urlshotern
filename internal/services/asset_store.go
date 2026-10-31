package services

import (
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strings"
)

// AssetStore persists uploaded files. LocalAssetStore is the only
// implementation today; an object-storage backend (R2/S3) can implement the
// same interface when the backend runs on more than one instance.
type AssetStore interface {
	Put(relPath string, data []byte) error
	Get(relPath string) ([]byte, error)
	Delete(relPath string) error
}

// LocalAssetStore stores files under a root directory on local disk.
type LocalAssetStore struct {
	root string
}

// NewLocalAssetStore creates a store rooted at dir (created if missing).
func NewLocalAssetStore(dir string) (*LocalAssetStore, error) {
	abs, err := filepath.Abs(dir)
	if err != nil {
		return nil, err
	}
	if err := os.MkdirAll(abs, 0o750); err != nil {
		return nil, fmt.Errorf("create upload dir: %w", err)
	}
	return &LocalAssetStore{root: abs}, nil
}

// Root returns the absolute storage directory.
func (s *LocalAssetStore) Root() string { return s.root }

// resolve maps a relative path to an absolute path inside root, rejecting
// anything that would escape it.
func (s *LocalAssetStore) resolve(relPath string) (string, error) {
	clean := filepath.Clean("/" + relPath)
	full := filepath.Join(s.root, clean)
	if full != s.root && !strings.HasPrefix(full, s.root+string(os.PathSeparator)) {
		return "", errors.New("invalid asset path")
	}
	return full, nil
}

// Put writes data atomically (temp file + rename).
func (s *LocalAssetStore) Put(relPath string, data []byte) error {
	full, err := s.resolve(relPath)
	if err != nil {
		return err
	}
	if err := os.MkdirAll(filepath.Dir(full), 0o750); err != nil {
		return err
	}
	tmp, err := os.CreateTemp(filepath.Dir(full), ".upload-*")
	if err != nil {
		return err
	}
	defer os.Remove(tmp.Name()) // no-op after a successful rename
	if _, err := tmp.Write(data); err != nil {
		_ = tmp.Close()
		return err
	}
	if err := tmp.Close(); err != nil {
		return err
	}
	if err := os.Chmod(tmp.Name(), 0o600); err != nil {
		return err
	}
	return os.Rename(tmp.Name(), full)
}

// Get reads a stored file.
func (s *LocalAssetStore) Get(relPath string) ([]byte, error) {
	full, err := s.resolve(relPath)
	if err != nil {
		return nil, err
	}
	return os.ReadFile(full) // #nosec G304 -- path confined to the store root by resolve
}

// Delete removes a stored file (missing files are not an error).
func (s *LocalAssetStore) Delete(relPath string) error {
	full, err := s.resolve(relPath)
	if err != nil {
		return err
	}
	if err := os.Remove(full); err != nil && !os.IsNotExist(err) {
		return err
	}
	return nil
}

// DeleteUserFiles removes every file stored for a user (m/<userID>/).
func (s *LocalAssetStore) DeleteUserFiles(userID int64) error {
	dir, err := s.resolve(fmt.Sprintf("m/%d", userID))
	if err != nil {
		return err
	}
	return os.RemoveAll(dir)
}
